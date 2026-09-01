// Package seed is the one-shot materializer: harness copy, spec,
// specialized domain-expert, GitHub control plane files, Hermes
// loop scripts. Live profile/cron/issue creation stays in cmd so
// this package stays filesystem-pure and testable.
package seed

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/loop"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/personas"
)

// Options is the input to Plan/Apply.
type Options struct {
	Name       string
	Describe   string
	SpecPath   string
	SpecText   string
	Domain     string
	Stack      string
	GitHubRepo string
	TargetDir  string
	HarnessSrc string
	SkipAgents bool
	SkipGitHub bool
	SkipCron   bool
}

// PlanResult is what will happen; ApplyResult is what happened.
type PlanResult struct {
	Slug        string
	Domain      string
	TargetDir   string
	HarnessSrc  string
	Personas    []string
	Issue0Title string
	Issue0Body  string
	CronJobs    []loop.CronJob
	CIKind      string
	Spec        string
	GitHubRepo  string
}

// ApplyResult records filesystem side effects.
type ApplyResult struct {
	PlanResult
	Files []string
}

// Plan infers domain, personas, issue 0, cron jobs. No IO except
// reading SpecPath if set.
func Plan(opts Options) (PlanResult, error) {
	if opts.Name == "" && opts.TargetDir == "" {
		return PlanResult{}, fmt.Errorf("name or target dir is required")
	}
	spec, err := loadSpec(opts)
	if err != nil {
		return PlanResult{}, err
	}
	name := opts.Name
	if name == "" {
		name = filepath.Base(opts.TargetDir)
	}
	slug := loopSlug(name)
	domain := opts.Domain
	if domain == "" {
		domain = inferDomain(spec + " " + opts.Describe)
	}
	target := opts.TargetDir
	if target == "" {
		target = name
	}
	cfg := loop.Config{
		ProjectRoot: target,
		ProjectSlug: slug,
		ProjectName: name,
		GitHubRepo:  opts.GitHubRepo,
		Domain:      domain,
	}
	return PlanResult{
		Slug:        slug,
		Domain:      domain,
		TargetDir:   target,
		HarnessSrc:  FindHarnessSrc(opts.HarnessSrc, ""),
		Personas:    loop.CanonicalPersonas(domain),
		Issue0Title: loop.Issue0Title(name),
		Issue0Body:  loop.Issue0Body(cfg, spec),
		CronJobs:    loop.CronCommands(cfg),
		CIKind:      loop.CIKind(opts.Stack),
		Spec:        spec,
		GitHubRepo:  opts.GitHubRepo,
	}, nil
}

// Apply copies harness, writes spec/loop/CI/issue-0. Does not talk
// to GitHub or Hermes.
func Apply(opts Options) (ApplyResult, error) {
	plan, err := Plan(opts)
	if err != nil {
		return ApplyResult{}, err
	}
	if opts.HarnessSrc != "" {
		plan.HarnessSrc = opts.HarnessSrc
	}
	if plan.HarnessSrc == "" {
		return ApplyResult{}, fmt.Errorf("harness source not found: pass --from <git-meta-harness/harness> or run from the framework repo")
	}
	dst := plan.TargetDir
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return ApplyResult{}, err
	}
	var files []string
	track := func(p string) { files = append(files, p) }

	if err := copyDir(plan.HarnessSrc, filepath.Join(dst, "harness")); err != nil {
		return ApplyResult{}, fmt.Errorf("copy harness: %w", err)
	}
	track(filepath.Join(dst, "harness"))

	if err := os.MkdirAll(filepath.Join(dst, "docs"), 0o755); err != nil {
		return ApplyResult{}, err
	}
	specPath := filepath.Join(dst, "docs", "SPEC.md")
	if err := os.WriteFile(specPath, []byte(plan.Spec+"\n"), 0o644); err != nil {
		return ApplyResult{}, err
	}
	track(specPath)

	cfg := loop.Config{
		ProjectRoot: dst,
		ProjectSlug: plan.Slug,
		ProjectName: opts.Name,
		GitHubRepo:  opts.GitHubRepo,
		Domain:      plan.Domain,
	}
	if cfg.ProjectName == "" {
		cfg.ProjectName = plan.Slug
	}
	if err := loop.WriteProjectFiles(cfg, plan.Spec); err != nil {
		return ApplyResult{}, err
	}

	if _, err := personas.Create(filepath.Join(dst, "harness"), plan.Domain, plan.Spec); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return ApplyResult{}, fmt.Errorf("domain-expert: %w", err)
		}
	}

	projectMD := filepath.Join(dst, "harness", "PROJECT.md")
	if err := os.WriteFile(projectMD, []byte(projectOverlay(cfg, plan)), 0o644); err != nil {
		return ApplyResult{}, err
	}
	track(projectMD)

	if err := os.MkdirAll(filepath.Join(dst, ".github", "workflows"), 0o755); err != nil {
		return ApplyResult{}, err
	}
	ciSrc := filepath.Join(plan.HarnessSrc, "templates", ".github-workflows-ci.yml")
	if plan.CIKind == "contracts" {
		alt := filepath.Join(plan.HarnessSrc, "templates", ".github-workflows-ci-contracts.yml")
		if _, err := os.Stat(alt); err == nil {
			ciSrc = alt
		}
	}
	ciDst := filepath.Join(dst, ".github", "workflows", "ci.yml")
	if err := copyFile(ciSrc, ciDst); err != nil {
		// contracts template may live only in the seed package fallback
		if plan.CIKind == "contracts" {
			if err := os.WriteFile(ciDst, []byte(contractsCI), 0o644); err != nil {
				return ApplyResult{}, err
			}
		} else {
			return ApplyResult{}, fmt.Errorf("copy ci.yml: %w", err)
		}
	}
	track(ciDst)

	relSrc := filepath.Join(plan.HarnessSrc, "templates", ".github-workflows-release.yml")
	relDst := filepath.Join(dst, ".github", "workflows", "release.yml")
	if err := copyFile(relSrc, relDst); err == nil {
		track(relDst)
	}

	prSrc := filepath.Join(plan.HarnessSrc, "templates", "pr-description.md")
	prDst := filepath.Join(dst, ".github", "PULL_REQUEST_TEMPLATE.md")
	if err := copyFile(prSrc, prDst); err == nil {
		track(prDst)
	}

	issue0 := filepath.Join(dst, ".github", "ISSUE_0.md")
	if err := os.WriteFile(issue0, []byte(plan.Issue0Body+"\n"), 0o644); err != nil {
		return ApplyResult{}, err
	}
	track(issue0)

	plan.TargetDir = dst
	plan.CIKind = loop.CIKind(opts.Stack)
	return ApplyResult{PlanResult: plan, Files: files}, nil
}

// FindHarnessSrc returns the directory to copy. explicit wins.
// A framework checkout (harness/bootstrap.md + cli/go.mod) is next.
func FindHarnessSrc(explicit, cwd string) string {
	if explicit != "" {
		return explicit
	}
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	if isFrameworkRepo(cwd) {
		return filepath.Join(cwd, "harness")
	}
	if env := os.Getenv("GMH_HARNESS_SRC"); env != "" {
		return env
	}
	return ""
}

func isFrameworkRepo(cwd string) bool {
	if _, err := os.Stat(filepath.Join(cwd, "harness", "bootstrap.md")); err != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(cwd, "cli", "go.mod"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "github.com/brenonaraujo/git-meta-harness/cli")
}

func loadSpec(opts Options) (string, error) {
	if opts.SpecText != "" {
		return strings.TrimSpace(opts.SpecText), nil
	}
	if opts.SpecPath != "" {
		b, err := os.ReadFile(opts.SpecPath)
		if err != nil {
			return "", fmt.Errorf("read spec: %w", err)
		}
		return strings.TrimSpace(string(b)), nil
	}
	if strings.TrimSpace(opts.Describe) != "" {
		d := strings.TrimSpace(opts.Describe)
		return "# Functional spec\n\n" + d + "\n", nil
	}
	return "", fmt.Errorf("--spec or --describe is required")
}

func inferDomain(text string) string {
	lower := strings.ToLower(text)
	patterns := map[string][]string{
		"ecommerce":   {"product", "cart", "checkout", "sku", "inventory", "catalog"},
		"fintech":     {"pix", "payment", "wallet", "kyc", "ledger", "bacen"},
		"marketplace": {"vendor", "seller", "listing", "group buying"},
		"saas":        {"subscription", "agendamento", "agendar", "booking", "horario", "horário", "clinica", "clínica", "schedule", "appointment", "billing", "workspace"},
		"ml":          {"model", "training", "inference", "embedding", "pytorch"},
		"internal":    {"admin", "internal", "tooling"},
	}
	best, bestScore := "internal", 0
	for dom, kws := range patterns {
		score := 0
		for _, kw := range kws {
			score += strings.Count(lower, kw)
		}
		if score > bestScore {
			best, bestScore = dom, score
		}
	}
	return best
}

func loopSlug(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prev := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prev = false
		case r == ' ' || r == '_' || r == '-' || r == '/' || r == '.':
			if !prev && b.Len() > 0 {
				b.WriteByte('-')
				prev = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func projectOverlay(cfg loop.Config, plan PlanResult) string {
	return fmt.Sprintf(`# Project overlay

This file is the **project** contract. `+"`"+`harness/bootstrap.md`+"`"+` stays
the framework invariant. Do not replace framework rules with stack defaults.

- **Name:** %s
- **Slug:** %s
- **Domain:** %s
- **GitHub:** %s
- **CI kind:** %s (full = Go/Nuxt modular CI; contracts = thin required-files CI)
- **Loop:** Hermes cron pooling (no GitHub webhook). See `+"`"+`harness/workflow/07-hermes-loop.md`+"`"+`.
- **team-manager:** orchestrates only. Never implements feature code.
- **Issue 0:** %s

## DoR
Spec in `+"`"+`docs/SPEC.md`+"`"+`, specialized domain-expert, labels, CI, `+"`"+`gmh loop doctor`+"`"+` green.

## DoD
Sensors green, PR with "Como testar", human `+"`"+`validado`+"`"+` unless PROJECT.md waives it for test-lab merges.
`, cfg.ProjectName, cfg.ProjectSlug, cfg.Domain, cfg.GitHubRepo, plan.CIKind, plan.Issue0Title)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

const contractsCI = `name: Contracts

# Thin CI for non-Go/Nuxt projects (docs, Vue+Swarm landing zones).
# Do NOT copy the harness Go/Nuxt + GHCR template here.

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  contracts:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Required harness files
        run: |
          test -f docs/SPEC.md
          test -f harness/PROJECT.md
          test -f harness/bootstrap.md
          test -f harness/loop.env
          test -x harness/scripts/loop/spawn-tm.sh
      - name: Loop liveness
        run: bash harness/scripts/check-loop.sh
      - name: No generic domain-expert
        run: test ! -f harness/personas/domain-expert.md
`
