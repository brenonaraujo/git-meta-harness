// Package prompt — CI renewal prompt generator.
//
// The CI renewal prompt is what `gmh agents update` produces when it
// detects that the local `.github/workflows/ci.yml` (and other
// `.github/*` files) is out of sync with the framework's new
// template. The agentic (Hermes, Claude Code, etc.) reads this prompt
// and produces the actual update PR.
package prompt

import (
	"fmt"
	"os"
	"strings"
)

// CIRenewalInput is the context for generating a CI renewal prompt.
type CIRenewalInput struct {
	// Cwd is the project root.
	Cwd string
	// LocalVersion is the framework version currently installed.
	LocalVersion string
	// LatestVersion is the latest framework version.
	LatestVersion string
	// LocalCI is the current content of .github/workflows/ci.yml.
	LocalCI string
	// TemplateCI is the new content of harness/templates/.github-workflows-ci.yml.
	TemplateCI string
	// FilesToUpdate is a list of paths in .github/ that may need updating,
	// relative to the project root. e.g. ".github/workflows/ci.yml".
	FilesToUpdate []string
	// FrameworkChanges is a summary of recent changes in the framework
	// (extracted from CHANGELOG.md between LocalVersion and LatestVersion).
	FrameworkChanges []string
	// LocalCustomizations is a list of human-readable notes about what
	// the local CI does differently (extracted heuristically).
	LocalCustomizations []string
	// AgenticName is the agentic to delegate to (default: "hermes").
	AgenticName string
}

// CIRenewalPrompt generates the structured prompt for the agentic.
//
// The agentic (Hermes, Claude Code, etc.) receives this prompt and
// should:
//
//	1. Re-read .github/workflows/ci.yml and the framework's new
//	   template (harness/templates/.github-workflows-ci.yml)
//	2. Understand the project's customizations (preserve them)
//	3. Update .github/workflows/ci.yml to match the new template,
//	   merging in project-specific customizations
//	4. Open a PR with the title "ci: renew to framework v{version}"
//
// The agentic has access to: project context, framework skills
// (github-pr-workflow, twelve-factor, code-graph), the new template
// diff, and the framework's recent changes.
func CIRenewalPrompt(in CIRenewalInput) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`# CI Renewal — %s

`+`
You are the **team-manager** (or a CI-specialist persona) of the
**meta-harness** framework. A user just ran `+"`gmh agents update`"+`
because their local `+"`.github/workflows/ci.yml`"+` is out of sync
with the framework's new template.

`, filepathBase(in.Cwd)))

	b.WriteString("## Project context\n\n")
	b.WriteString(fmt.Sprintf("- **Project:** `%s`\n", in.Cwd))
	b.WriteString(fmt.Sprintf("- **Local framework version:** %s\n", orNA(in.LocalVersion)))
	b.WriteString(fmt.Sprintf("- **Latest framework version:** %s\n", orNA(in.LatestVersion)))
	if in.AgenticName != "" {
		b.WriteString(fmt.Sprintf("- **Agentic:** %s\n", in.AgenticName))
	}
	b.WriteString("\n")

	b.WriteString("## Files in `.github/` that may need updating\n\n")
	if len(in.FilesToUpdate) == 0 {
		b.WriteString("_(none detected)_\n\n")
	} else {
		for _, f := range in.FilesToUpdate {
			b.WriteString(fmt.Sprintf("- `%s`\n", f))
		}
		b.WriteString("\n")
	}

	if len(in.LocalCustomizations) > 0 {
		b.WriteString("## Local customizations (preserve these!)\n\n")
		for _, c := range in.LocalCustomizations {
			b.WriteString(fmt.Sprintf("- %s\n", c))
		}
		b.WriteString("\n")
	}

	if len(in.FrameworkChanges) > 0 {
		b.WriteString("## Framework changes since v" + in.LocalVersion + "\n\n")
		for _, c := range in.FrameworkChanges {
			b.WriteString(fmt.Sprintf("- %s\n", c))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Diff between local CI and new template\n\n")
	if in.LocalCI != "" && in.TemplateCI != "" {
		b.WriteString("```diff\n")
		// Naive line-by-line diff for human readability
		// (the agentic will re-read both files for full context)
		localLines := strings.Split(in.LocalCI, "\n")
		templateLines := strings.Split(in.TemplateCI, "\n")
		localSet := make(map[string]bool)
		for _, l := range localLines {
			localSet[l] = true
		}
		templateSet := make(map[string]bool)
		for _, l := range templateLines {
			templateSet[l] = true
		}
		// Show lines only in template (additions) and only in local (deletions)
		for _, l := range templateLines {
			if !localSet[l] && strings.TrimSpace(l) != "" {
				b.WriteString(fmt.Sprintf("+ %s\n", l))
			}
		}
		for _, l := range localLines {
			if !templateSet[l] && strings.TrimSpace(l) != "" {
				b.WriteString(fmt.Sprintf("- %s\n", l))
			}
		}
		b.WriteString("```\n\n")
		b.WriteString("_Read both files in full to make a real merge, not just this naive diff._\n\n")
	}

	b.WriteString(`## Your task

1. **Re-read source-of-truth files yourself** (don't trust this prompt):
   ` + "`" + in.Cwd + "/.github/workflows/ci.yml`" + `
   ` + "`" + in.Cwd + "/harness/templates/.github-workflows-ci.yml`" + `
   Compare and understand what changed.

2. **Run the canonical checks** (sensor 09 protocol):
   ` + "```bash\n" +
		"   cd " + in.Cwd + "\n" +
		"   ./harness/scripts/check-stack-versions.sh\n" +
		"   ./harness/scripts/smoke-test.sh\n" +
		"   ```\n\n" +
		`   (This catches any other issues that need fixing in the same PR.)

3. **Apply the update:**
   - Preserve project-specific customizations (env vars, secrets,
     additional jobs, custom paths).
   - Add the framework's new sensors: govulncheck, dorny/paths-filter,
     scope cache, Trivy SHA-pinned, cosign (if release pipeline).
   - Pin ALL action versions (no `+"`@latest`"+`, no `+"`@master`"+`).
   - Add `+"`GOTOOLCHAIN=local`"+` to all Go jobs (impedes `+"`go mod tidy`"+`
     from rewriting go.mod in CI).
   - Add 12-Factor audit as an always-running job.
   - Add i18n audit if the project has i18n.
   - Use `+"`dorny/paths-filter@v3.0.2`"+` (SHA-pinned in prod) at the top
     to skip irrelevant jobs.

4. **Open a PR:**
   - Branch: `+"`ci/renew-framework-" + strings.TrimPrefix(in.LatestVersion, "v") + "`" + `
   - Title: `+"`ci: renew to framework v" + in.LatestVersion + "`" + `
   - Body: list the new sensors/gates added, the customizations
     preserved, and the verification commands.

5. **Update the local VERSION file** if it changed.

## Critical

- **DO NOT trust auto-report.** Re-read files. Re-run commands. (Invariante 19.)
- **DO NOT break existing CI.** Test with `+"`act`"+` or in a fork before opening the PR.
- **PRESERVE secrets and env vars** (DATABASE_URL, GHCR_TOKEN, etc.).
- **Use pinned actions only** (no @latest, no @master).
- **Document each change** in the PR body — the user will review.
- **Be concise.** 1 paragraph of diagnosis + 1 list of changes.
`)

	return b.String()
}

func filepathBase(p string) string {
	if p == "" {
		return ""
	}
	// Naive: take last component
	parts := strings.Split(strings.TrimRight(p, "/"), "/")
	if len(parts) == 0 {
		return p
	}
	return parts[len(parts)-1]
}

// RecentChangesFromChangelog reads CHANGELOG.md and extracts the
// changes between oldVersion and newVersion. Returns at most `max`
// bullet points.
func RecentChangesFromChangelog(changelogPath, oldVersion, newVersion string, max int) []string {
	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return nil
	}
	content := string(data)
	// Find the section for newVersion
	idx := strings.Index(content, "## ["+newVersion+"]")
	if idx < 0 {
		return nil
	}
	// Find the next section
	end := strings.Index(content[idx+10:], "## [")
	var section string
	if end < 0 {
		section = content[idx:]
	} else {
		section = content[idx : idx+10+end]
	}
	// Extract bullet points
	var out []string
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			out = append(out, strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
			if len(out) >= max {
				break
			}
		}
	}
	return out
}
