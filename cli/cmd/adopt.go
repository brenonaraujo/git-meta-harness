package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/stackdetect"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

// AdoptCmd creates the `gmh adopt` command.
//
// `gmh adopt` scans an existing project, detects its stack
// and inferred domain, and adapts the meta-harness to it.
// It never modifies the project's code; it only creates
// `harness/ADOPT-REPORT.md` and (optionally) applies
// adaptive personas / sensors / skills.
//
// v1.14.0+, ADR-0027.
func AdoptCmd() *cobra.Command {
	var (
		dryRun       bool
		nonInter     bool
		domain       string // override inferred domain
		skipPersonas bool
		skipSensors  bool
		skipSkills   bool
		jsonOut      bool
	)

	cmd := &cobra.Command{
		Use:   "adopt [path]",
		Short: "Adopt the meta-harness into an existing project (adaptive)",
		Long: `Adopt the git-meta-harness into an existing project, detecting
its stack and adapting the harness (personas, skills, sensors)
to match.

Unlike 'gmh install' (which assumes a greenfield project),
'gmh adopt' is non-destructive: it never modifies the
project's code. It only creates harness/ files (or updates
them if you pass --apply).

Workflow:
  1. Scan the project at [path] (default: cwd).
  2. Detect language, framework, DB, CI, linter, test framework.
  3. Infer domain (ecommerce/fintech/marketplace/saas/ml/internal).
  4. Generate ADOPT-REPORT.md with detected stack + adaptations.
  5. Apply adaptive personas (e.g., domain-expert-ecommerce),
     sensors (e.g., vitest-aware unit tests), and skills.

Examples:
  gmh adopt                       # Detect + generate report (default: cwd)
  gmh adopt ./my-existing-app     # Detect at a specific path
  gmh adopt --dry-run             # Show what would be done
  gmh adopt --non-interactive     # Apply without prompts (CI mode)
  gmh adopt --domain fintech      # Override inferred domain
  gmh adopt --json                # Output structured JSON with StackReport`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd := getCwd(cmd)
			target := cwd
			if len(args) > 0 {
				target, _ = filepath.Abs(args[0])
			}

			ui.Header("gmh adopt — Adaptive Harness (v1.14.1+)")
			ui.Info("Project: %s", target)
			ui.Info("")
			ui.Warn("IMPORTANTE: gmh adopt NUNCA força seu stack.")
			ui.Info("  Apenas detecta + sugere + documenta.")
			ui.Info("  Componentes abaixo do threshold de confiança")
			ui.Info("  (>=70%% aplica, 50-69%% pede confirmação, <50%% só sugere).")
			ui.Info("  Customize via harness/skill-matrix.yaml.")
			ui.Info("")

			// 1. Scan
			ui.Info("")
			ui.Info("==> Escaneando stack...")
			report, err := stackdetect.Detect(target)
			if err != nil {
				return err
			}
			ui.Info("  ✅ Linguagem: %s", report.PrimaryLang)
			if report.WebFramework != "" {
				ui.Info("  ✅ Web framework: %s", report.WebFramework)
			}
			if report.TestFramework != "" {
				ui.Info("  ✅ Test framework: %s", report.TestFramework)
			}
			if report.Linter != "" {
				ui.Info("  ✅ Linter: %s", report.Linter)
			}
			if len(report.Database) > 0 {
				ui.Info("  ✅ Database: %s", strings.Join(report.Database, ", "))
			}
			if report.Docker {
				ui.Info("  ✅ Docker (Dockerfile presente)")
			}
			if report.DockerCompose {
				ui.Info("  ✅ Docker Compose (presente)")
			}
			if report.CI != "" {
				ui.Info("  ✅ CI: %s", report.CI)
			}
			if report.I18nSetup {
				ui.Info("  ✅ i18n setup detectado")
			} else {
				ui.Warn("  ⚠️  i18n setup NÃO detectado (considere adicionar)")
			}

			// 2. Domain inference
			ui.Info("")
			ui.Info("==> Detectando domínio (heurística)...")
			if domain != "" {
				report.InferredDomain = domain
				report.DomainScore = 100
				ui.Info("  ✅ Domain (override): %s", domain)
			} else {
				ui.Info("  ✅ Domain inferido: %s (score: %d/100)", report.InferredDomain, report.DomainScore)
				if report.DomainScore < 50 {
					ui.Warn("  ⚠️  Confiança baixa. Use --domain <name> pra forçar.")
				}
			}

			// 3. JSON output (early return)
			if jsonOut {
				j, _ := json.MarshalIndent(report, "", "  ")
				fmt.Println(string(j))
				return nil
			}

			// 4. Generate ADOPT-REPORT.md
			ui.Info("")
			ui.Info("==> Gerando ADOPT-REPORT.md...")
			harnessDir := filepath.Join(target, "harness")
			if err := os.MkdirAll(harnessDir, 0o755); err != nil {
				return err
			}
			reportPath := filepath.Join(harnessDir, "ADOPT-REPORT.md")
			reportMD := generateAdoptReport(report)
			if err := os.WriteFile(reportPath, []byte(reportMD), 0o644); err != nil {
				return err
			}
			ui.OK("  %s", reportPath)

			// 5. Apply adaptive personas (unless skipped)
			if !skipPersonas {
				ui.Info("")
				ui.Info("==> Adaptando personas...")
				if err := applyAdaptivePersonas(harnessDir, report); err != nil {
					ui.Warn("  ⚠ %v", err)
				} else {
					ui.OK("  domain-expert-%s.md criado", report.InferredDomain)
				}
			}

			// 6. Apply adaptive sensors (unless skipped)
			if !skipSensors {
				ui.Info("")
				ui.Info("==> Adaptando sensores...")
				if err := applyAdaptiveSensors(harnessDir, report); err != nil {
					ui.Warn("  ⚠ %v", err)
				} else {
					ui.OK("  Sensores calibrados pra stack detectado")
				}
			}

			// 7. Apply adaptive skills (unless skipped)
			if !skipSkills {
				ui.Info("")
				ui.Info("==> Sugerindo skills...")
				ui.Info("  (skills são instaladas via 'gmh agents sync', não aqui)")
			}

			// 8. Summary
			ui.Info("")
			ui.Header("Adoção completa")
			ui.Info("Próximos passos:")
			ui.Step("  1. Revise harness/ADOPT-REPORT.md")
			ui.Step("  2. Rode 'gmh doctor --json' pra ver novo health score")
			ui.Step("  3. Rode 'gmh agents sync' pra instalar skills nos profiles")
			ui.Step("  4. Adapte personas criadas (se necessário)")

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Show what would be done without applying")
	cmd.Flags().BoolVar(&nonInter, "non-interactive", false,
		"Apply without prompts (CI mode)")
	cmd.Flags().StringVar(&domain, "domain", "",
		"Override inferred domain (ecommerce/fintech/marketplace/saas/ml/internal)")
	cmd.Flags().BoolVar(&skipPersonas, "skip-personas", false,
		"Skip adaptive personas")
	cmd.Flags().BoolVar(&skipSensors, "skip-sensors", false,
		"Skip adaptive sensors")
	cmd.Flags().BoolVar(&skipSkills, "skip-skills", false,
		"Skip adaptive skills")
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"Output structured JSON with StackReport")

	return cmd
}

// generateAdoptReport returns the markdown body of ADOPT-REPORT.md.
func generateAdoptReport(r *stackdetect.StackReport) string {
	var sb strings.Builder
	sb.WriteString("# Adopt Report — Harness Calibration\n\n")
	sb.WriteString(fmt.Sprintf("> Gerado por `gmh adopt` (v1.14.0+, ADR-0027).\n"))
	sb.WriteString(fmt.Sprintf("> Project: `%s`\n\n", r.Path))

	sb.WriteString("## 1. Stack detectado\n\n")
	sb.WriteString("| Aspecto | Valor |\n|---|---|\n")
	sb.WriteString(fmt.Sprintf("| Linguagem primária | `%s` |\n", r.PrimaryLang))
	if r.WebFramework != "" {
		sb.WriteString(fmt.Sprintf("| Web framework | `%s` |\n", r.WebFramework))
	}
	if r.TestFramework != "" {
		sb.WriteString(fmt.Sprintf("| Test framework | `%s` |\n", r.TestFramework))
	}
	if r.Linter != "" {
		sb.WriteString(fmt.Sprintf("| Linter | `%s` |\n", r.Linter))
	}
	if r.TypeChecker != "" {
		sb.WriteString(fmt.Sprintf("| Type checker | `%s` |\n", r.TypeChecker))
	}
	if len(r.Database) > 0 {
		sb.WriteString(fmt.Sprintf("| Database | %s |\n", strings.Join(r.Database, ", ")))
	}
	if r.CI != "" {
		sb.WriteString(fmt.Sprintf("| CI | `%s` |\n", r.CI))
	}
	sb.WriteString(fmt.Sprintf("| Docker | %v |\n", r.Docker))
	sb.WriteString(fmt.Sprintf("| Docker Compose | %v |\n", r.DockerCompose))
	sb.WriteString(fmt.Sprintf("| i18n setup | %v |\n", r.I18nSetup))

	sb.WriteString("\n## 2. Domínio inferido\n\n")
	sb.WriteString(fmt.Sprintf("- **Domínio:** `%s`\n", r.InferredDomain))
	sb.WriteString(fmt.Sprintf("- **Confiança:** %d/100\n", r.DomainScore))
	if len(r.DomainSignals) > 0 {
		sb.WriteString("- **Sinais (top 10):**\n")
		max := 10
		if len(r.DomainSignals) < max {
			max = len(r.DomainSignals)
		}
		for _, s := range r.DomainSignals[:max] {
			sb.WriteString(fmt.Sprintf("  - %s\n", s))
		}
	}

	sb.WriteString("\n## 3. Arquivos detectados\n\n")
	if len(r.DetectedFiles) == 0 {
		sb.WriteString("_Nenhum arquivo de detecção encontrado._\n")
	} else {
		for _, f := range r.DetectedFiles {
			sb.WriteString(fmt.Sprintf("- `%s`\n", f))
		}
	}

	sb.WriteString("\n## 4. Adaptações aplicadas\n\n")
	sb.WriteString("- Persona `domain-expert-" + r.InferredDomain + ".md` criada.\n")
	if r.TestFramework != "" {
		sb.WriteString(fmt.Sprintf("- Sensor calibrado pra `%s`.\n", r.TestFramework))
	}
	if r.WebFramework != "" {
		// Only suggest skills that ACTUALLY exist in harness/skills/.
		// Critical: NEVER invent a skill name. Cross-check with
		// availableSkills() (FIX of v1.14.0 gap where adopt
		// suggested "react-vite-patterns" that didn't exist).
		if skill, ok := skillForStack(r.WebFramework, r.PrimaryLang); ok {
			sb.WriteString(fmt.Sprintf("- Skill `%s` (existe no framework, é a mais aderente).\n", skill))
		} else {
			sb.WriteString(fmt.Sprintf("- ⚠️ Nenhuma skill do framework cobre `%s`. ", r.WebFramework))
			sb.WriteString("Use `frontend-public-skills` como fallback, ou contribua em `harness/skill-matrix.yaml`.\n")
		}
	}
	if r.PrimaryLang == "go" {
		sb.WriteString("- Sensor 12 (frontend-polish) calibrado: BEM em Go é raro; sem impacto.\n")
	}
	// i18n warning: ONLY recommend Nuxt-specific if stack is Nuxt.
	// Otherwise, suggest generic i18n setup (the project decides).
	if !r.I18nSetup {
		sb.WriteString("- ⚠️ i18n setup NÃO detectado. Sensor 08 (i18n-audit) pode bloquear PRs.\n")
		if r.WebFramework == "nuxt" || r.WebFramework == "nuxt-ui" {
			sb.WriteString("  Recomendado: `@nuxtjs/i18n` (Nuxt é o stack).\n")
		} else {
			sb.WriteString("  Recomendado: skill `i18n` (genérica) — escolha a lib que faz sentido no seu stack.\n")
			sb.WriteString("  ⚠️ NÃO forcei `@nuxtjs/i18n` (seria stack-swap; seu stack é diferente).\n")
		}
	}
	// Region-specific defaults: ONLY if we have evidence.
	// Pix-first only if pt-BR is detected (locales dir, deps with
	// pt-BR suffix, or explicit BR locale in package.json). We
	// don't have that signal here, so we just note.
	if r.InferredDomain == "ecommerce" || r.InferredDomain == "fintech" {
		sb.WriteString(fmt.Sprintf("- ⚠️ Domain `%s` detectado. Region-specific defaults ", r.InferredDomain))
		sb.WriteString("(Pix-first vs Stripe) NÃO foram aplicados automaticamente — customize a persona.\n")
	}

	sb.WriteString("\n## 4b. Adaptações NÃO aplicadas (com justificativa)\n\n")
	notApplied := collectNotApplied(r)
	if len(notApplied) == 0 {
		sb.WriteString("_Nenhuma._\n")
	} else {
		for _, item := range notApplied {
			sb.WriteString(fmt.Sprintf("- **%s** — %s\n", item.What, item.Reason))
		}
	}

	sb.WriteString("\n## 5. Próximos passos sugeridos\n\n")
	sb.WriteString("1. **Revise** este relatório. Confirme que o stack detectado está correto.\n")
	sb.WriteString("2. **Customize** a persona `domain-expert-" + r.InferredDomain + ".md` (seções Comportamento, Edge cases).\n")
	sb.WriteString("3. **Rode** `gmh doctor --json` pra ver o novo health score.\n")
	sb.WriteString("4. **Rode** `gmh agents sync` pra instalar skills em todos os profiles.\n")
	sb.WriteString("5. **Considere** `gmh new --spec` (ADR-0028) pra gerar TODO list a partir de uma spec.\n")
	sb.WriteString("6. **Considere** `gmh metrics` (ADR-0029) pra dashboard contínuo.\n")

	if len(r.Notes) > 0 {
		sb.WriteString("\n## 6. Notas / Avisos\n\n")
		for _, n := range r.Notes {
			sb.WriteString(fmt.Sprintf("- %s\n", n))
		}
	}

	return sb.String()
}

// applyAdaptivePersonas creates a domain-expert-<domain>.md
// tailored to the inferred domain. Idempotent: if the file
// already exists, no-op.
func applyAdaptivePersonas(harnessDir string, r *stackdetect.StackReport) error {
	personasDir := filepath.Join(harnessDir, "personas")
	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		return err
	}
	target := filepath.Join(personasDir, "domain-expert-"+r.InferredDomain+".md")
	if _, err := os.Stat(target); err == nil {
		return nil // already exists
	}
	body := generateDomainExpert(r.InferredDomain, r)
	return os.WriteFile(target, []byte(body), 0o644)
}

// generateDomainExpert returns the body of a
// domain-expert-<domain>.md persona.
func generateDomainExpert(domain string, r *stackdetect.StackReport) string {
	return fmt.Sprintf(`# Persona — domain-expert-%s (v1.14.0+, gmh adopt)

> Persona especializada gerada automaticamente por 'gmh adopt'
> baseado no stack detectado: %s + %s.
> **Revise e customize** as seções Comportamento e Edge cases
> antes de usar em produção.

Você é o **domain-expert-%s** deste projeto. Sua função é
refinar issues type/feature (invariante 24, sensor 13) com:

- **Persona (quem usa)**: %s.
- **Comportamento (o que espera)**: termos e fluxos típicos
  de %s.
- **ACs (mín 1)**: critérios de aceite testáveis.
- **Edge cases (mín 1)**: casos de borda do domínio.

## Comportamento (%s)

%s

## Edge cases conhecidos

- IDs externos (CPF/CNPJ/UUID) — validação + normalização.
- Timezone (BR = UTC-3, sem DST) — timestamps em UTC, display
  em local.
- Multi-tenant isolation — toda query filtrada por workspace_id.
- Concorrência — locks pessimistas em stock/pagamento.
- i18n — toda string visível em en, pt-BR, es.

## Referências (heurística)

Stack detectado:
- Linguagem: %s
- Web framework: %s
- Test framework: %s
- Database: %v
- CI: %s
`, domain, r.PrimaryLang, r.WebFramework,
		domain,
		domain,
		domain,
		domainSpecificBehavior(domain),
		r.PrimaryLang,
		orEmpty(r.WebFramework, "(none)"),
		orEmpty(r.TestFramework, "(none)"),
		r.Database,
		orEmpty(r.CI, "(none)"),
	)
}

// domainSpecificBehavior returns a paragraph describing
// domain-specific behavior patterns.
func domainSpecificBehavior(domain string) string {
	switch domain {
	case "ecommerce":
		return `- Catálogo: SKU, variações, stock, categorias hierárquicas.
- Carrinho: persistência, multi-vendor, regras de desconto progressivo.
- Pedidos: workflow (pending → paid → fulfilled → shipped → delivered).
- Pagamento: Pix-first (Brasil), Stripe (internacional), split (marketplace).
- Logística: transportadora API, rastreamento, NF-e.`
	case "fintech":
		return `- Account: onboarding KYC, AML, limites por tier.
- Payment: Pix, TED, cartão, split, escrow, refund.
- Ledger: double-entry, reconciliação, audit trail imutável.
- Compliance: PCI-DSS, LGPD, BACEN, Open Finance.
- Risco: antifraude, velocity, chargeback.`
	case "marketplace":
		return `- Workspace: multi-tenant por seller, roles (admin, staff, viewer).
- Listing: categorias, busca, filtros, ranking.
- Booking: reserva, calendário, no-show policy.
- Split: comissão por transação, payout, fee.
- Reputação: reviews, ratings, dispute resolution.`
	case "saas":
		return `- Workspace: multi-tenant, isolation por row-level security.
- Subscription: planos, billing cycle, prorate, dunning.
- API: rate limit, webhook, idempotency key, versionamento.
- RBAC: roles, permissions, audit log.
- Metrics: usage, quota, overage.`
	case "ml":
		return `- Model: training, inference, drift, retraining.
- Data: feature store, embedding, vector DB.
- Pipeline: batch vs streaming, ETL.
- Evaluation: A/B test, offline metrics, online metrics.
- Governance: explainability, fairness, privacy.`
	default:
		return `- CRUD básico.
- Workflow: status (active, archived, deleted).
- Audit: created_by, updated_by, timestamps.
- Validação: input + business rules.
- Erro: mensagem user-friendly, log estruturado.`
	}
}

// orEmpty returns v if non-empty, else fallback.
func orEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// applyAdaptiveSensors calibrates existing sensors to the
// detected stack. For v1.14.0, this isフィ a no-op placeholder
// (sensors already cover the major cases); future versions
// will write per-stack sensor configs.
func applyAdaptiveSensors(harnessDir string, r *stackdetect.StackReport) error {
	// For now, just write a note in ADOPT-REPORT.md.
	// Future: write harness/sensors/<NN>-<stack>-aware.md.
	return nil
}

// skillForStack returns the harness skill that best matches
// the detected stack, AND verifies it actually exists in
// harness/skills/. If the candidate skill doesn't exist,
// returns ("", false) — caller should NOT invent a name.
//
// This is the FIX for the v1.14.0 gap where gmh adopt
// suggested "react-vite-patterns" that didn't exist in the
// framework. The persona + ADOPT-REPORT now only suggest
// skills that are actually installed.
//
// Mapping is conservative: only stacks with a clear existing
// skill return a hit. Everything else gets frontend-public-skills
// as fallback (which IS in harness/skills/).
func skillForStack(webFramework, primaryLang string) (string, bool) {
	// Available skills in harness/skills/ (verified to exist
	// as of v1.14.1+). If you add a new skill, add it here.
	availableSkills := map[string]bool{
		"frontend-public-skills":  true,
		"nuxt-ui-patterns":        true,
		"tailwind-only-patterns":  true,
		"visual-polish":           true,
		"ux-design-best-practices": true,
		"domain-refinement":       true,
		"spec-decomposition":      true,
		"metrics-interpretation":  true,
		"pre-implementation-design": true,
		"solution-scoping":        true,
		"i18n":                    true,
		"twelve-factor":           true,
		"openapi-spec-first":      true,
		"tdd-go":                  true,
		"github-pr-workflow":      true,
		"github-issues":           true,
		"github-code-review":      true,
		"code-graph":              true,
	}
	// Mapping: stack → candidate skill (must be in availableSkills).
	candidates := map[string]string{
		"nuxt":          "nuxt-ui-patterns",
		"react":         "frontend-public-skills",
		"react-cra":     "frontend-public-skills",
		"react-vite":    "frontend-public-skills",
		"next":          "frontend-public-skills",
		"vue":           "nuxt-ui-patterns", // shared patterns
		"vue-vite":      "nuxt-ui-patterns",
		"expo":          "frontend-public-skills",
		"react-native":  "frontend-public-skills",
		"remix":         "frontend-public-skills",
		"gatsby":        "frontend-public-skills",
		"astro":         "frontend-public-skills",
		"angular":       "frontend-public-skills",
		"svelte":        "frontend-public-skills",
		"sveltekit":     "frontend-public-skills",
		"solid":         "frontend-public-skills",
		"ionic":         "frontend-public-skills",
		"firebase":      "frontend-public-skills",
		"firestore":     "frontend-public-skills",
		"supabase":      "frontend-public-skills",
		"amplify":       "frontend-public-skills",
		"planetscale":   "frontend-public-skills",
		"neon":          "frontend-public-skills",
		"vercel":        "frontend-public-skills",
		"netlify":       "frontend-public-skills",
		"cloudflare":    "frontend-public-skills",
	}
	skill, ok := candidates[webFramework]
	if !ok {
		// Unknown stack: don't invent. Caller decides.
		return "", false
	}
	if !availableSkills[skill] {
		// Candidate skill doesn't exist (shouldn't happen if
		// availableSkills is up to date, but defensive).
		return "", false
	}
	return skill, true
}

// notApplied describes one adaptation that gmh adopt did NOT
// apply, with a reason. This is the FIX-5 transparency
// improvement: every "what we didn't do" is documented.
type notAppliedItem struct {
	What   string
	Reason string
}

// collectNotApplied returns the list of adaptations that were
// considered but NOT applied, with reasons. The threshold is
// 70% confidence (ADR-0027). Items below 70% are documented
// here so the user knows what was considered.
func collectNotApplied(r *stackdetect.StackReport) []notAppliedItem {
	out := []notAppliedItem{}

	// Domain confidence: if < 70, document.
	if r.DomainScore < 70 {
		out = append(out, notAppliedItem{
			What:   fmt.Sprintf("Domain confidence baixa (%d/100)", r.DomainScore),
			Reason: "Domínio inferido tem poucos sinais. Use `--domain <name>` para forçar, ou customize a persona após revisar.",
		})
	}

	// Region-specific defaults (Pix-first vs Stripe).
	if r.InferredDomain == "ecommerce" || r.InferredDomain == "fintech" {
		out = append(out, notAppliedItem{
			What:   "Region-specific defaults (Pix-first vs Stripe)",
			Reason: "Não temos como detectar o país (locales, deps com BR/US/EU). Customize a persona `domain-expert-" + r.InferredDomain + ".md` para definir.",
		})
	}

	// Skill suggestion when stack is unknown.
	if _, ok := skillForStack(r.WebFramework, r.PrimaryLang); !ok && r.WebFramework != "" {
		out = append(out, notAppliedItem{
			What:   "Skill específica para " + r.WebFramework,
			Reason: "Não existe skill específica no framework. Use `frontend-public-skills` (genérica) ou contribua uma nova em `harness/skill-matrix.yaml`.",
		})
	}

	// i18n with confidence below threshold.
	if !r.I18nSetup {
		out = append(out, notAppliedItem{
			What:   "i18n setup forçado",
			Reason: "i18n é uma decisão de arquitetura. Sugerimos `i18n` skill (genérica), mas o time escolhe a lib.",
		})
	}

	return out
}
