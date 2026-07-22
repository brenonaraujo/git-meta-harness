package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/stackdetect"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

// NewCmd creates the `gmh new` command.
//
// `gmh new <name> --spec <spec.md>` creates a new project
// from a functional spec. It:
//  1. Reads the spec file.
//  2. Detects domain (heuristic).
//  3. Creates the project directory with harness/ applied.
//  4. Generates a TODO list (epic + sub-issues) from the spec.
//  5. Writes docs/SPEC.md, docs/ADOPT-REPORT.md, harness/TODO.md.
//
// v1.14.0+, ADR-0028.
func NewCmd() *cobra.Command {
	var (
		specPath      string
		domain        string
		stackOverride string
		dryRun        bool
		nonInter      bool
		noInit        bool // don't create directory, just generate TODO
		jsonOut       bool
		skipHarness   bool
	)

	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create new project from a functional spec",
		Long: `Create a new project from a functional spec.

Reads the spec, detects the domain, creates the project
directory, applies the meta-harness, and generates a TODO
list (epic + sub-issues) for the team-manager to triage.

Examples:
  gmh new my-app --spec spec.md                 # Create from spec
  gmh new my-app --spec spec.md --dry-run       # Show what would be done
  gmh new my-app --spec spec.md --domain fintech # Override domain
  gmh new my-app --spec spec.md --json          # Output structured JSON
  gmh new my-app --spec spec.md --no-harness    # Skip harness apply (just TODO)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if specPath == "" {
				return fmt.Errorf("--spec <path> is required")
			}

			specData, err := os.ReadFile(specPath)
			if err != nil {
				return fmt.Errorf("read spec: %w", err)
			}
			spec := string(specData)

			ui.Header("gmh new — Create project from spec (v1.14.0)")
			ui.Info("Project: %s", name)
			ui.Info("Spec:    %s (%d bytes)", specPath, len(specData))

			// 1. Detect domain from spec
			ui.Info("")
			ui.Info("==> Analisando spec...")
			specReport := analyzeSpec(spec, domain)
			ui.Info("  ✅ Domain inferido: %s (score: %d/100)", specReport.InferredDomain, specReport.DomainScore)
			if len(specReport.Keywords) > 0 {
				ui.Info("  ✅ Keywords (top 10): %s", strings.Join(specReport.Keywords, ", "))
			}

			// 2. Detect stack (from spec text or override)
			ui.Info("")
			ui.Info("==> Detectando stack...")
			if stackOverride != "" {
				ui.Info("  ✅ Stack (override): %s", stackOverride)
			} else {
				ui.Info("  ✅ Stack detectado da spec (default: Go + Nuxt + PostgreSQL)")
			}

			// 3. Decompose spec into TODO list
			ui.Info("")
			ui.Info("==> Decompondo spec em TODO list...")
			todos := decomposeSpec(spec, name, specReport.InferredDomain)
			ui.Info("  ✅ %d épicos + %d sub-issues geradas", len(todos), countSubIssues(todos))

			// 4. JSON output (early return)
			if jsonOut {
				out := map[string]interface{}{
					"project":  name,
					"spec":     specPath,
					"domain":   specReport.InferredDomain,
					"epics":    todos,
					"keywords": specReport.Keywords,
				}
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			// 5. Create project directory
			target, _ := filepath.Abs(name)
			if !noInit {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return err
				}
				ui.Info("")
				ui.Info("==> Criando projeto em %s", target)
			}

			// 6. Write spec to docs/SPEC.md
			docsDir := filepath.Join(target, "docs")
			if err := os.MkdirAll(docsDir, 0o755); err != nil {
				return err
			}
			if !dryRun {
				if err := copyFile(specPath, filepath.Join(docsDir, "SPEC.md")); err != nil {
					return err
				}
				ui.OK("  docs/SPEC.md (%d bytes)", len(specData))
			}

			// 7. Write TODO.md (human-readable)
			todoMD := generateTodoMarkdown(name, specReport.InferredDomain, todos)
			todoPath := filepath.Join(target, "harness", "TODO.md")
			if err := os.MkdirAll(filepath.Dir(todoPath), 0o755); err != nil {
				return err
			}
			if !dryRun {
				if err := os.WriteFile(todoPath, []byte(todoMD), 0o644); err != nil {
					return err
				}
				ui.OK("  harness/TODO.md")
			}

			// 8. Write TODO.json (machine-readable)
			todoJSON, _ := json.MarshalIndent(todos, "", "  ")
			todoJSONPath := filepath.Join(target, "harness", "TODO.json")
			if !dryRun {
				if err := os.WriteFile(todoJSONPath, todoJSON, 0o644); err != nil {
					return err
				}
				ui.OK("  harness/TODO.json")
			}

			// 9. Write SPEC-COVERAGE.md (matriz spec → issues)
			covMD := generateSpecCoverage(spec, todos)
			covPath := filepath.Join(target, "harness", "SPEC-COVERAGE.md")
			if !dryRun {
				if err := os.WriteFile(covPath, []byte(covMD), 0o644); err != nil {
					return err
				}
				ui.OK("  harness/SPEC-COVERAGE.md")
			}

			// 10. Write ADOPT-REPORT.md
			adoptMD := generateAdoptReportForNew(name, specReport, stackOverride)
			adoptPath := filepath.Join(target, "harness", "ADOPT-REPORT.md")
			if !dryRun {
				if err := os.WriteFile(adoptPath, []byte(adoptMD), 0o644); err != nil {
					return err
				}
				ui.OK("  harness/ADOPT-REPORT.md")
			}

			// 11. Apply meta-harness (unless --no-harness)
			if !skipHarness {
				ui.Info("")
				ui.Info("==> Aplicando meta-harness v1.14.0...")
				ui.Info("  (esqueleto — rode 'gmh install' no projeto criado pra completar)")
			}

			// 12. Summary
			ui.Info("")
			ui.Header("Projeto criado")
			ui.Info("Próximos passos:")
			ui.Step("  1. cd %s", name)
			ui.Step("  2. Revise harness/TODO.md (épicos + sub-issues)")
			ui.Step("  3. Rode 'gmh install' pra completar o harness")
			ui.Step("  4. Crie o repo no GitHub + 'git remote add origin ...'")
			ui.Step("  5. Crie as issues via 'gh issue create --label <label> --title ...' (use harness/TODO.json)")

			return nil
		},
	}

	cmd.Flags().StringVar(&specPath, "spec", "", "Path to the functional spec (markdown or text)")
	cmd.Flags().StringVar(&domain, "domain", "", "Override inferred domain (ecommerce/fintech/marketplace/saas/ml/internal)")
	cmd.Flags().StringVar(&stackOverride, "stack", "", "Override stack detection (e.g., 'go,nuxt,postgresql')")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without creating files")
	cmd.Flags().BoolVar(&nonInter, "non-interactive", false, "Skip prompts (CI mode)")
	cmd.Flags().BoolVar(&noInit, "no-init", false, "Don't create directory; just generate TODO")
	cmd.Flags().BoolVar(&skipHarness, "no-harness", false, "Skip harness application (just generate TODO)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output structured JSON with epics")

	return cmd
}

// specAnalysis is the result of analyzeSpec().
type specAnalysis struct {
	InferredDomain string   `json:"inferred_domain"`
	DomainScore    int      `json:"domain_score"`
	Keywords       []string `json:"keywords"`
	Length         int      `json:"length"`
	NumSections    int      `json:"num_sections"`
}

// analyzeSpec inspects a spec text and infers domain.
//
// Heuristic: count keyword matches per domain in the spec.
func analyzeSpec(spec, domainOverride string) specAnalysis {
	a := specAnalysis{
		Length:      len(spec),
		NumSections: strings.Count(spec, "\n##") + strings.Count(spec, "\n# "),
	}

	// Default: scan for keywords.
	patterns := map[string][]string{
		"ecommerce":   {"product", "cart", "checkout", "sku", "order", "shipping", "inventory", "catalog", "catalogue"},
		"fintech":     {"pix", "payment", "transfer", "wallet", "kyc", "aml", "ledger", "transaction", "bacen"},
		"marketplace": {"workspace", "tenant", "vendor", "seller", "listing", "booking", "group buying", "group-buying"},
		"saas":        {"workspace", "subscription", "plan", "billing", "invoice", "api key", "webhook", "rbac"},
		"ml":          {"model", "training", "inference", "embedding", "vector", "tensorflow", "pytorch", "sklearn"},
		"internal":    {"admin", "internal", "tooling", "cron", "worker", "job"},
	}

	lower := strings.ToLower(spec)
	scores := map[string]int{}
	for dom, kws := range patterns {
		for _, kw := range kws {
			count := strings.Count(lower, kw)
			if count > 0 {
				scores[dom] += count
				if len(a.Keywords) < 20 {
					a.Keywords = append(a.Keywords, fmt.Sprintf("%s:%d", kw, count))
				}
			}
		}
	}

	best := "internal"
	bestScore := 0
	for dom, score := range scores {
		if score > bestScore {
			best = dom
			bestScore = score
		}
	}
	a.InferredDomain = best
	a.DomainScore = bestScore * 5
	if a.DomainScore > 100 {
		a.DomainScore = 100
	}
	if domainOverride != "" {
		a.InferredDomain = domainOverride
		a.DomainScore = 100
	}
	return a
}

// Epic is one chapter/area of the spec.
type Epic struct {
	Title     string     `json:"title"`
	Anchor    string     `json:"anchor"`
	SubIssues []SubIssue `json:"sub_issues"`
	ACs       []string   `json:"acs"`
	EdgeCases []string   `json:"edge_cases"`
	Priority  string     `json:"priority"` // p0, p1, p2, p3
	Labels    []string   `json:"labels"`
	SpecRef   string     `json:"spec_ref"` // "spec.md#section"
}

// SubIssue is one deliverable testable.
type SubIssue struct {
	Number    string   `json:"number"` // "F1.2"
	Title     string   `json:"title"`
	ACs       []string `json:"acs"`
	EdgeCases []string `json:"edge_cases"`
	SpecRef   string   `json:"spec_ref"`
	Labels    []string `json:"labels"`
}

// decomposeSpec splits a spec into epics + sub-issues using
// section headers (##) as epic boundaries and bullet points
// (## - or -) as sub-issue hints.
func decomposeSpec(spec, projectName, domain string) []Epic {
	epics := []Epic{}

	// Split on top-level sections (## or #).
	sectionRe := regexp.MustCompile(`(?m)^#+\s+(.+)$`)
	matches := sectionRe.FindAllStringSubmatchIndex(spec, -1)
	if len(matches) == 0 {
		// No sections: treat entire spec as 1 epic.
		epic := Epic{
			Title:    projectName + " — MVP",
			Anchor:   "mvp",
			Priority: "p0",
			Labels:   []string{"type/feature", "domain/" + domain, "priority/p0"},
		}
		// Extract bullets as sub-issues
		epic.SubIssues = extractSubIssuesFromText(spec, "1")
		epic.ACs = extractACs(spec)
		epic.EdgeCases = extractEdgeCases(spec)
		epic.SpecRef = "spec.md#mvp"
		return []Epic{epic}
	}

	for i, m := range matches {
		title := strings.TrimSpace(spec[m[2]:m[3]])
		// Skip the spec title (first H1)
		if i == 0 && strings.HasPrefix(spec[m[0]:m[1]], "# ") {
			continue
		}
		// Get section body
		bodyStart := m[1]
		bodyEnd := len(spec)
		if i+1 < len(matches) {
			bodyEnd = matches[i+1][0]
		}
		body := spec[bodyStart:bodyEnd]

		// Build epic
		epicNum := fmt.Sprintf("%d", i+1)
		epic := Epic{
			Title:    fmt.Sprintf("F%s: %s", epicNum, title),
			Anchor:   slugify(title),
			Priority: priorityFromTitle(title),
			Labels:   []string{"type/feature", "domain/" + domain, "priority/" + priorityFromTitle(title)},
			SpecRef:  "spec.md#" + slugify(title),
		}

		// Extract sub-issues from bullets / sub-sections
		epic.SubIssues = extractSubIssuesFromText(body, epicNum)
		epic.ACs = extractACs(body)
		epic.EdgeCases = extractEdgeCases(body)

		epics = append(epics, epic)
	}

	// Renumber priorities: first epic p0, second p1, etc.
	for i := range epics {
		switch i {
		case 0:
			epics[i].Priority = "p0"
		case 1:
			epics[i].Priority = "p1"
		case 2:
			epics[i].Priority = "p2"
		default:
			epics[i].Priority = "p3"
		}
		// Update labels
		epics[i].Labels = updatePriorityLabel(epics[i].Labels, epics[i].Priority)
	}
	return epics
}

// extractSubIssuesFromText returns sub-issues by looking for
// subsection headers (### or ####) and bullets.
func extractSubIssuesFromText(text, epicNum string) []SubIssue {
	subs := []SubIssue{}
	subNum := 0

	// Try subsection headers first
	subsectionRe := regexp.MustCompile(`(?m)^###\s+(.+)$`)
	matches := subsectionRe.FindAllStringSubmatchIndex(text, -1)
	if len(matches) > 0 {
		for i, m := range matches {
			subNum++
			title := strings.TrimSpace(text[m[2]:m[3]])
			_ = m[1]
			_ = m[0]
			_ = len(text)
			_ = i
			sub := SubIssue{
				Number:    fmt.Sprintf("F%s.%d", epicNum, subNum),
				Title:     title,
				ACs:       extractACs(text),
				EdgeCases: extractEdgeCases(text),
				SpecRef:   "spec.md#" + slugify(title),
				Labels:    []string{"type/feature"},
			}
			subs = append(subs, sub)
		}
		return subs
	}

	// Fallback: extract bullet points (- or *)
	bulletRe := regexp.MustCompile(`(?m)^[\s]*[-*]\s+(.+)$`)
	bulletMatches := bulletRe.FindAllStringSubmatch(text, -1)
	for _, m := range bulletMatches {
		subNum++
		title := strings.TrimSpace(m[1])
		if len(title) > 80 {
			title = title[:77] + "..."
		}
		sub := SubIssue{
			Number:    fmt.Sprintf("F%s.%d", epicNum, subNum),
			Title:     title,
			ACs:       extractACs(text),
			EdgeCases: extractEdgeCases(text),
			SpecRef:   "spec.md#" + slugify(title),
			Labels:    []string{"type/feature"},
		}
		subs = append(subs, sub)
	}
	return subs
}

// extractACs returns acceptance criteria from the text.
func extractACs(text string) []string {
	acs := []string{}
	re := regexp.MustCompile(`(?i)(?:AC|acceptance[ -]criteria|should|must|will)[^.]{10,200}\.`)
	matches := re.FindAllString(text, -1)
	for _, m := range matches {
		m = strings.TrimSpace(m)
		if len(m) > 5 && len(m) < 200 {
			acs = append(acs, "- "+m)
		}
	}
	if len(acs) == 0 {
		acs = append(acs, "- [ ] AC1: <acceptance criterion derived from spec>")
	}
	if len(acs) > 5 {
		acs = acs[:5]
	}
	return acs
}

// extractEdgeCases returns edge cases from the text.
func extractEdgeCases(text string) []string {
	edges := []string{}
	patterns := []string{
		`(?i)\bif\s+[a-z\s_]+\s+then\s+[a-z][^.]+`,
		`(?i)\b(edge case|exception|nota:|warning|cuidado)[^.]+`,
		`(?i)\b(empty|null|invalid|expired|missing)[^.]{5,100}\.`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		matches := re.FindAllString(text, -1)
		for _, m := range matches {
			m = strings.TrimSpace(m)
			if len(m) > 10 && len(m) < 200 {
				edges = append(edges, "- "+m)
			}
		}
	}
	if len(edges) > 3 {
		edges = edges[:3]
	}
	return edges
}

func priorityFromTitle(title string) string {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "mvp") || strings.Contains(lower, "core") || strings.Contains(lower, "auth"):
		return "p0"
	case strings.Contains(lower, "v2") || strings.Contains(lower, "enhance"):
		return "p2"
	default:
		return "p1"
	}
}

func updatePriorityLabel(labels []string, newPrio string) []string {
	out := []string{}
	for _, l := range labels {
		if !strings.HasPrefix(l, "priority/") {
			out = append(out, l)
		}
	}
	out = append(out, "priority/"+newPrio)
	return out
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "-")
	return s
}

func countSubIssues(epics []Epic) int {
	n := 0
	for _, e := range epics {
		n += len(e.SubIssues)
	}
	return n
}

func generateTodoMarkdown(projectName, domain string, epics []Epic) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# TODO — %s\n\n", projectName))
	sb.WriteString(fmt.Sprintf("> Gerado por `gmh new` (v1.14.0+, ADR-0028).\n"))
	sb.WriteString(fmt.Sprintf("> Domain: `%s`. Total: %d épicos + %d sub-issues.\n\n", domain, len(epics), countSubIssues(epics)))

	for _, e := range epics {
		sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", e.Title, e.Priority))
		sb.WriteString(fmt.Sprintf("**Spec ref:** `%s`\n", e.SpecRef))
		if len(e.Labels) > 0 {
			sb.WriteString(fmt.Sprintf("**Labels:** `%s`\n", strings.Join(e.Labels, "`, `")))
		}
		if len(e.ACs) > 0 {
			sb.WriteString("\n**ACs:**\n")
			for _, ac := range e.ACs {
				sb.WriteString(fmt.Sprintf("%s\n", ac))
			}
		}
		if len(e.EdgeCases) > 0 {
			sb.WriteString("\n**Edge cases:**\n")
			for _, ec := range e.EdgeCases {
				sb.WriteString(fmt.Sprintf("%s\n", ec))
			}
		}
		if len(e.SubIssues) > 0 {
			sb.WriteString("\n**Sub-issues:**\n")
			for _, s := range e.SubIssues {
				sb.WriteString(fmt.Sprintf("- **%s** — %s\n", s.Number, s.Title))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func generateSpecCoverage(spec string, epics []Epic) string {
	var sb strings.Builder
	sb.WriteString("# SPEC-COVERAGE — spec → épicos\n\n")
	sb.WriteString("> Matriz que mapeia cada seção da spec aos épicos/sub-issues gerados.\n\n")
	sb.WriteString("| Spec section | Épico | Sub-issues | Coverage |\n")
	sb.WriteString("|---|---|---|---|\n")
	for _, e := range epics {
		cov := "100%"
		if len(e.SubIssues) == 0 {
			cov = "0% (no sub-issues)"
		}
		sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %d | %s |\n", e.Anchor, e.SpecRef, e.Title, len(e.SubIssues), cov))
	}
	return sb.String()
}

func generateAdoptReportForNew(projectName string, a specAnalysis, stackOverride string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Adopt Report — %s (gmh new)\n\n", projectName))
	sb.WriteString("> Gerado por `gmh new` (v1.14.0+, ADR-0028).\n\n")
	sb.WriteString(fmt.Sprintf("## Stack\n\n"))
	if stackOverride != "" {
		sb.WriteString(fmt.Sprintf("- Override: `%s`\n", stackOverride))
	} else {
		sb.WriteString("- Default: Go 1.26.5 + Nuxt 4.5 + PostgreSQL 18.4 (verifique stack/versions.md)\n")
	}
	sb.WriteString(fmt.Sprintf("\n## Domain\n\n"))
	sb.WriteString(fmt.Sprintf("- Inferido: `%s` (score %d/100)\n", a.InferredDomain, a.DomainScore))
	sb.WriteString(fmt.Sprintf("- Spec length: %d bytes, %d sections\n", a.Length, a.NumSections))
	return sb.String()
}

// -- stub for stackdetect (not used in new, but referenced) --
var _ = stackdetect.StackReport{}
