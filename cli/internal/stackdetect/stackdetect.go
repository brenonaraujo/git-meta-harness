// Package stackdetect analyzes an existing project and
// returns a structured StackReport describing its
// language, framework, dependencies, and inferred domain.
//
// Used by `gmh adopt` (v1.14.0+, ADR-0027) to calibrate the
// meta-harness to the project's reality.
package stackdetect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// StackReport is the result of a Detect() call.
type StackReport struct {
	Path           string   `json:"path"`
	PrimaryLang    string   `json:"primary_lang"`    // "go" | "typescript" | "python" | "rust" | "java" | "unknown"
	WebFramework   string   `json:"web_framework"`   // "nuxt" | "next" | "sveltekit" | "vite" | "" if N/A
	TestFramework  string   `json:"test_framework"`  // "go test" | "vitest" | "jest" | "pytest" | "playwright" | ""
	Database       []string `json:"database"`        // ["postgresql", "redis", "mongodb"]
	Linter         string   `json:"linter"`          // "golangci-lint" | "eslint" | "ruff" | ""
	TypeChecker    string   `json:"type_checker"`    // "tsc" | "mypy" | "" if N/A
	CI             string   `json:"ci"`              // "github-actions" | "" if N/A
	I18nSetup      bool     `json:"i18n_setup"`      // true if @nuxtjs/i18n or i18n/ dir found
	Docker         bool     `json:"docker"`          // true if Dockerfile present
	DockerCompose  bool     `json:"docker_compose"`  // true if docker-compose.yml present
	DetectedFiles  []string `json:"detected_files"`  // files that triggered detections
	InferredDomain string   `json:"inferred_domain"` // "ecommerce" | "fintech" | "marketplace" | "saas" | "ml" | "internal" | "unknown"
	DomainScore    int      `json:"domain_score"`    // 0-100 confidence
	DomainSignals  []string `json:"domain_signals"`  // signals that triggered the domain inference
	Notes          []string `json:"notes"`           // any caveats / warnings
}

// Detect analyzes the project at `root` (which should be a
// directory path) and returns a StackReport.
//
// Detection is purely filesystem-based (no network, no git
// operations) and completes in <1s for projects up to 50k
// LOC.
func Detect(root string) (*StackReport, error) {
	r := &StackReport{
		Path:          root,
		DetectedFiles: []string{},
		Notes:         []string{},
	}

	// 1. Walk top-level + common subdirs (limited depth).
	checkFiles := []string{
		"go.mod", "go.sum",
		"package.json", "pnpm-lock.yaml", "package-lock.json", "yarn.lock",
		"pyproject.toml", "requirements.txt", "Pipfile", "setup.py",
		"Cargo.toml", "pom.xml", "build.gradle",
		"Dockerfile", "docker-compose.yml", "docker-compose.yaml",
		".github/workflows/ci.yml", ".github/workflows/release.yml",
		"tsconfig.json", "mypy.ini", "pyrightconfig.json",
		".golangci.yml", ".golangci.yaml", ".eslintrc", ".eslintrc.json", "ruff.toml", ".ruff.toml",
		"i18n.config.ts", "nuxt.config.ts", "next.config.js", "next.config.ts",
	}
	for _, f := range checkFiles {
		full := filepath.Join(root, f)
		if _, err := os.Stat(full); err == nil {
			r.DetectedFiles = append(r.DetectedFiles, f)
		}
	}

	// 2. Primary language.
	if hasAny(r.DetectedFiles, "go.mod") {
		r.PrimaryLang = "go"
	} else if hasAny(r.DetectedFiles, "package.json") {
		// Check if TS or JS
		if hasAny(r.DetectedFiles, "tsconfig.json") {
			r.PrimaryLang = "typescript"
		} else {
			r.PrimaryLang = "javascript"
			if r.PrimaryLang == "javascript" {
				r.PrimaryLang = "typescript" // default to TS for modern projects
			}
		}
	} else if hasAny(r.DetectedFiles, "pyproject.toml", "requirements.txt", "Pipfile") {
		r.PrimaryLang = "python"
	} else if hasAny(r.DetectedFiles, "Cargo.toml") {
		r.PrimaryLang = "rust"
	} else if hasAny(r.DetectedFiles, "pom.xml", "build.gradle") {
		r.PrimaryLang = "java"
	} else {
		r.PrimaryLang = "unknown"
	}

	// 3. Web framework + test framework + linter + type checker
	// (read package.json if present)
	pkgJSON := readPackageJSON(filepath.Join(root, "package.json"))
	if pkgJSON != nil {
		deps := mergeMaps(pkgJSON.Dependencies, pkgJSON.DevDependencies)
		// Order matters: meta-frameworks (next/nuxt/sveltekit) win
		// over plain react/vue. This is so a Next.js project is
		// reported as "next" not "react", which is what the user
		// actually uses day-to-day.
		if _, ok := deps["nuxt"]; ok {
			r.WebFramework = "nuxt"
		} else if _, ok := deps["next"]; ok {
			r.WebFramework = "next"
		} else if _, ok := deps["@sveltejs/kit"]; ok {
			r.WebFramework = "sveltekit"
		} else if _, ok := deps["astro"]; ok {
			r.WebFramework = "astro"
		} else if _, ok := deps["@remix-run/react"]; ok || strInDeps(deps, "remix") {
			r.WebFramework = "remix"
		} else if _, ok := deps["gatsby"]; ok {
			r.WebFramework = "gatsby"
		} else if _, ok := deps["react"]; ok {
			// Plain React (no Next/Remix/Gatsby). Use vite or cra
			// as the bundler signal.
			if _, ok := deps["vite"]; ok {
				r.WebFramework = "react-vite"
			} else if _, ok := deps["react-scripts"]; ok {
				r.WebFramework = "react-cra"
			} else {
				r.WebFramework = "react"
			}
		} else if _, ok := deps["vue"]; ok {
			if _, ok := deps["vite"]; ok {
				r.WebFramework = "vue-vite"
			} else {
				r.WebFramework = "vue"
			}
		} else if _, ok := deps["@angular/core"]; ok {
			r.WebFramework = "angular"
		} else if _, ok := deps["svelte"]; ok {
			r.WebFramework = "svelte"
		} else if _, ok := deps["solid-js"]; ok {
			r.WebFramework = "solid"
		}
		// Mobile / desktop
		if _, ok := deps["react-native"]; ok {
			if _, ok := deps["expo"]; ok {
				r.WebFramework = "expo"
				r.Notes = append(r.Notes, "Mobile: Expo (React Native + tooling)")
			} else {
				r.WebFramework = "react-native"
				r.Notes = append(r.Notes, "Mobile: React Native (bare)")
			}
		} else if _, ok := deps["@ionic/angular"]; ok || strInDeps(deps, "@ionic/react") {
			r.WebFramework = "ionic"
		} else if _, ok := deps["electron"]; ok {
			r.Notes = append(r.Notes, "Desktop: Electron")
		} else if _, ok := deps["tauri"]; ok || strInDeps(deps, "@tauri-apps/api") {
			r.Notes = append(r.Notes, "Desktop: Tauri")
		}
		if _, ok := deps["vitest"]; ok {
			r.TestFramework = "vitest"
		} else if _, ok := deps["jest"]; ok {
			r.TestFramework = "jest"
		} else if _, ok := deps["@playwright/test"]; ok {
			r.TestFramework = "playwright"
		} else if _, ok := deps["cypress"]; ok {
			r.TestFramework = "cypress"
		}
		if _, ok := deps["typescript"]; ok {
			r.TypeChecker = "tsc"
		}
		if _, ok := deps["eslint"]; ok {
			r.Linter = "eslint"
		}
		if _, ok := deps["@nuxtjs/i18n"]; ok {
			r.I18nSetup = true
		}
	}

	// Go-specific
	if r.PrimaryLang == "go" {
		r.TestFramework = "go test"
		if hasAny(r.DetectedFiles, ".golangci.yml", ".golangci.yaml") {
			r.Linter = "golangci-lint"
		}
	}

	// Python-specific
	if r.PrimaryLang == "python" {
		if hasAny(r.DetectedFiles, "pyproject.toml", "requirements.txt") {
			if _, err := os.Stat(filepath.Join(root, "pytest.ini")); err == nil {
				r.TestFramework = "pytest"
			} else if _, err := os.Stat(filepath.Join(root, "pyproject.toml")); err == nil {
				// Check pyproject.toml content for pytest
				data, _ := os.ReadFile(filepath.Join(root, "pyproject.toml"))
				if strings.Contains(string(data), "pytest") {
					r.TestFramework = "pytest"
				}
			}
		}
		if hasAny(r.DetectedFiles, "mypy.ini", "pyrightconfig.json") {
			if hasAny(r.DetectedFiles, "mypy.ini") {
				r.TypeChecker = "mypy"
			} else {
				r.TypeChecker = "pyright"
			}
		}
		if hasAny(r.DetectedFiles, "ruff.toml", ".ruff.toml") {
			r.Linter = "ruff"
		}
	}

	// Docker
	if hasAny(r.DetectedFiles, "Dockerfile") {
		r.Docker = true
	}
	if hasAny(r.DetectedFiles, "docker-compose.yml", "docker-compose.yaml") {
		r.DockerCompose = true
		// Detect DBs in compose
		composeData, _ := os.ReadFile(filepath.Join(root, "docker-compose.yml"))
		if composeData == nil {
			composeData, _ = os.ReadFile(filepath.Join(root, "docker-compose.yaml"))
		}
		compose := string(composeData)
		dbPatterns := map[string]string{
			"postgres":      "postgresql",
			"mysql":         "mysql",
			"mariadb":       "mariadb",
			"mongo":         "mongodb",
			"redis":         "redis",
			"rabbitmq":      "rabbitmq",
			"kafka":         "kafka",
			"elasticsearch": "elasticsearch",
		}
		for pattern, name := range dbPatterns {
			if strings.Contains(compose, pattern) {
				r.Database = append(r.Database, name)
			}
		}
	}

	// Serverless / managed BaaS (Firebase, Supabase, Amplify, etc.)
	// These are detected via deps + config files, NOT via
	// docker-compose (serverless = no containers to compose).
	if pkgJSON != nil {
		deps := mergeMaps(pkgJSON.Dependencies, pkgJSON.DevDependencies)
		baasPatterns := map[string]string{
			"firebase":                 "firebase",
			"@firebase/app":            "firebase",
			"firebase-admin":           "firebase-admin",
			"@google-cloud/firestore":  "firestore",
			"firebase-functions":       "firebase-functions",
			"@supabase/supabase-js":    "supabase",
			"@supabase/ssr":            "supabase",
			"supabase":                 "supabase",
			"aws-amplify":              "amplify",
			"@aws-amplify/cli":         "amplify",
			"@planetscale/database":    "planetscale",
			"@neondatabase/serverless": "neon",
			"@vercel/postgres":         "vercel-postgres",
			"@vercel/kv":               "vercel-kv",
			"@upstash/redis":           "upstash-redis",
			"convex":                   "convex",
			"fauna-db":                 "fauna",
			"mongodb-atlas":            "mongodb-atlas",
			"@databases/pg":            "vercel-postgres",
		}
		for pattern, name := range baasPatterns {
			if _, ok := deps[pattern]; ok {
				r.Database = appendUnique(r.Database, name)
			}
		}
		// ORMs / query builders (these are not "databases" per se
		// but matter for the team-manager to know).
		ormPatterns := map[string]string{
			"prisma":          "prisma",
			"@prisma/client":  "prisma",
			"drizzle-orm":     "drizzle",
			"typeorm":         "typeorm",
			"sequelize":       "sequelize",
			"knex":            "knex",
			"drizzle-kit":     "drizzle",
			"@mikro-orm/core": "mikro-orm",
		}
		for pattern, name := range ormPatterns {
			if _, ok := deps[pattern]; ok {
				r.Notes = append(r.Notes, "ORM: "+name)
			}
		}
		// Hosting / deployment
		hostingPatterns := map[string]string{
			"@vercel/next":             "vercel",
			"@netlify/plugin":          "netlify",
			"wrangler":                 "cloudflare",
			"@sveltejs/adapter-vercel": "vercel",
		}
		for pattern, name := range hostingPatterns {
			if _, ok := deps[pattern]; ok {
				r.Notes = append(r.Notes, "Hosting: "+name)
			}
		}
	}

	// Firebase config files
	if hasAny(r.DetectedFiles, "firebase.json", ".firebaserc", "firestore.rules", "firestore.indexes.json") {
		r.Database = appendUnique(r.Database, "firebase")
	}
	// Supabase config
	if hasAny(r.DetectedFiles, "supabase/config.toml") {
		r.Database = appendUnique(r.Database, "supabase")
	}
	// Vercel / Netlify / Cloudflare config
	if hasAny(r.DetectedFiles, "vercel.json", ".vercelignore") {
		r.Notes = append(r.Notes, "Hosting: vercel")
	}
	if hasAny(r.DetectedFiles, "netlify.toml") {
		r.Notes = append(r.Notes, "Hosting: netlify")
	}
	if hasAny(r.DetectedFiles, "wrangler.toml", "wrangler.jsonc") {
		r.Notes = append(r.Notes, "Hosting: cloudflare")
	}
	// AWS Amplify
	if hasAny(r.DetectedFiles, "amplify.yml", "amplify/") {
		r.Notes = append(r.Notes, "Hosting: amplify")
	}

	// CI
	if hasAny(r.DetectedFiles, ".github/workflows/ci.yml", ".github/workflows/release.yml") {
		r.CI = "github-actions"
	}

	// I18n
	if !r.I18nSetup {
		if _, err := os.Stat(filepath.Join(root, "i18n")); err == nil {
			r.I18nSetup = true
		} else if _, err := os.Stat(filepath.Join(root, "locales")); err == nil {
			r.I18nSetup = true
		}
	}

	// 4. Domain inference (heuristic on filesystem).
	inferDomain(root, r)

	return r, nil
}

// inferDomain uses simple heuristics on the filesystem to
// guess the project's domain. Score 0-100. Adds signals to
// DomainSignals.
func inferDomain(root string, r *StackReport) {
	signals := map[string]int{
		"ecommerce":   0,
		"fintech":     0,
		"marketplace": 0,
		"saas":        0,
		"ml":          0,
		"internal":    0,
	}

	addSignal := func(domain, signal string, weight int) {
		signals[domain] += weight
		r.DomainSignals = append(r.DomainSignals, signal)
	}

	// Scan source files for keyword patterns.
	patterns := map[string]string{
		"ecommerce":   `(?i)\b(product|cart|checkout|sku|order|shipping|inventory|catalogue|catalog)\b`,
		"fintech":     `(?i)\b(pix|payment|transfer|wallet|account|kyc|aml|ledger|transaction)\b`,
		"marketplace": `(?i)\b(workspace|tenant|vendor|seller|listing|booking|reservation|group[\s_-]?buying)\b`,
		"saas":        `(?i)\b(workspace|tenant|subscription|plan|billing|invoice|api[\s_-]?key|webhook)\b`,
		"ml":          `(?i)\b(model|training|inference|embedding|vector|tensorflow|pytorch|sklearn)\b`,
		"internal":    `(?i)\b(admin|internal|tooling|cron|worker|job)\b`,
	}

	walkLimited(root, 3, func(path string) {
		// Skip vendor / node_modules / .git
		base := filepath.Base(path)
		if base == "node_modules" || base == "vendor" || base == ".git" || base == "dist" || base == "build" {
			return
		}
		// Only text-ish files
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".ts" && ext != ".js" && ext != ".vue" && ext != ".py" && ext != ".java" && ext != ".rs" && ext != ".md" {
			return
		}
		data, err := os.ReadFile(path)
		if err != nil || len(data) > 1<<20 { // skip files > 1MB
			return
		}
		content := string(data)
		for domain, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllString(content, -1)
			if len(matches) > 0 {
				weight := 1
				if len(matches) > 10 {
					weight = 3
				} else if len(matches) > 3 {
					weight = 2
				}
				addSignal(domain, filepath.Base(path)+": "+strconvItoa(len(matches))+" matches", weight)
			}
		}
	})

	// Pick the highest-scoring domain
	best := "unknown"
	bestScore := 0
	for domain, score := range signals {
		if score > bestScore {
			best = domain
			bestScore = score
		}
	}
	r.InferredDomain = best
	// Confidence: capped at 100, scaled by max signal count.
	r.DomainScore = bestScore * 10
	if r.DomainScore > 100 {
		r.DomainScore = 100
	}
	if bestScore < 3 {
		r.InferredDomain = "internal" // safe default
		r.DomainScore = 10
	}
}

// walkLimited walks the directory tree starting at root, up
// to maxDepth levels, calling fn for each file. Skips
// common skip-dirs.
func walkLimited(root string, maxDepth int, fn func(path string)) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == "vendor" || base == ".git" || base == "dist" || base == "build" || base == "target" || strings.HasPrefix(base, ".") {
				if path != root {
					return filepath.SkipDir
				}
			}
			rel, _ := filepath.Rel(root, path)
			depth := strings.Count(rel, string(os.PathSeparator))
			if depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		fn(path)
		return nil
	})
}

func hasAny(list []string, targets ...string) bool {
	for _, l := range list {
		for _, target := range targets {
			if l == target {
				return true
			}
		}
	}
	return false
}

// strInDeps returns true if the given key is in the deps map.
// Helper for cleaner code than `if _, ok := deps["foo"]; ok`.
func strInDeps(deps map[string]string, key string) bool {
	_, ok := deps[key]
	return ok
}

// appendUnique appends s to list if not already present.
// Used to avoid duplicates in Database (e.g., "firebase" from
// deps + "firebase" from firebase.json should appear once).
func appendUnique(list []string, s string) []string {
	for _, x := range list {
		if x == s {
			return list
		}
	}
	return append(list, s)
}

func mergeMaps(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// pkgJSON is a minimal subset of package.json.
type pkgJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func readPackageJSON(path string) *pkgJSON {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var p pkgJSON
	if err := json.Unmarshal(data, &p); err != nil {
		return nil
	}
	return &p
}

// JSON returns the report as a JSON string (for --json output).
func (r *StackReport) JSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// strconvItoa is a tiny helper to avoid pulling strconv into
// the hot path of the regex loop. (Actually, we can use
// strconv — leave this as a placeholder for symmetry.)
func strconvItoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
