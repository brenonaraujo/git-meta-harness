// Package prompt generates the harness-health-check prompt that
// `gmh doctor` delegates to the user's agentic.
package prompt

import (
	"fmt"
	"strings"
)

// DoctorInput is what `gmh doctor` already knows about the project.
type DoctorInput struct {
	// Cwd is the working directory.
	Cwd string
	// LocalVersion is the version of the framework currently installed
	// (read from harness/..VERSION or "" if missing).
	LocalVersion string
	// LatestVersion is the latest framework version available.
	LatestVersion string
	// OutOfDate is true if LocalVersion != LatestVersion.
	OutOfDate bool
	// DetectedAgentic is the most-likely agentic the user has.
	DetectedAgentic string
	// AllAgentics is the list of all agentics detected.
	AllAgentics []string
	// LocalChecks is the list of gmh doctor checks that passed/failed.
	// Each entry is "PASS: <name>" or "FAIL: <name>".
	LocalChecks []string
	// Fails is just the failed checks (subset of LocalChecks).
	Fails []string
	// ProjectName is the basename of Cwd.
	ProjectName string
	// ProjectDescription is a one-line description if known.
	ProjectDescription string
}

// DoctorPrompt generates the prompt for the team-manager to perform
// a deep harness health check. The agent should:
//
//  1. Re-run the canonical checks (make lint, make test, make vuln)
//  2. Detect drift between local harness/ and the latest framework
//  3. Suggest concrete actions (sync, update, fix)
//  4. Optionally open a PR to apply the fixes
func DoctorPrompt(in DoctorInput) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`# Harness health check — %s

`+`
You are the **team-manager** of the meta-harness. A user just ran
`+"`gmh doctor`"+` and got a quick local report. Now I need a DEEP check.

`, in.ProjectName))

	b.WriteString("## Local context\n\n")
	b.WriteString(fmt.Sprintf("- **Project:** `%s`\n", in.Cwd))
	b.WriteString(fmt.Sprintf("- **Local meta-harness version:** %s\n", orNA(in.LocalVersion)))
	b.WriteString(fmt.Sprintf("- **Latest meta-harness version:** %s\n", in.LatestVersion))
	if in.OutOfDate {
		b.WriteString(fmt.Sprintf("- ⚠️ **Out of date** — %s behind\n",
			diffMsg(in.LocalVersion, in.LatestVersion)))
	} else {
		b.WriteString("- ✅ **Up to date**\n")
	}
	b.WriteString(fmt.Sprintf("- **Detected agentic:** %s\n", in.DetectedAgentic))
	if len(in.AllAgentics) > 0 {
		b.WriteString(fmt.Sprintf("- **All agentics installed:** %s\n", strings.Join(in.AllAgentics, ", ")))
	}
	b.WriteString("\n")

	b.WriteString("## Pre-checks (gmh doctor ran these)\n\n")
	for _, c := range in.LocalChecks {
		b.WriteString(fmt.Sprintf("- %s\n", c))
	}
	b.WriteString("\n")

	b.WriteString(`## Your task

1. **Re-read source-of-truth files yourself** (don't trust gmh doctor alone):
   ` + "`go.mod`" + `, ` + "`Dockerfile`" + `, ` + "`docker-compose.yml`" + `, ` + "`.github/workflows/ci.yml`" + `, ` + "`harness/AGENTS.md`" + `.
   Compare to the canonical in ` + "`harness/stack/versions.md`" + `.

2. **Re-run the canonical sensors** (do NOT skip):
   ` + "```bash\n" +
		"   cd " + in.Cwd + "\n" +
		"   ./harness/scripts/check-stack-versions.sh\n" +
		"   ./harness/scripts/smoke-test.sh\n" +
		"   cd backend && make lint && make test && make vuln && cd ..\n" +
		"   cd web && pnpm lint && pnpm typecheck && pnpm test:run && pnpm audit --audit-level=high && cd ..\n" +
		"   ```\n\n" +
		`   (This is the verify-after-build protocol from sensor 09 / invariante 19.)

3. **If out of date, propose a plan:**
   - If local is just a few patch versions behind: ` + "`gmh update --to " + in.LatestVersion + "`" + `
   - If local is on an old major/minor: ` + "`gmh sync`" + ` (then review the diff)
   - If CI is failing locally: fix root cause first, then sync
   - If new sensors/scripts apply: open a PR with ` + "`gmh sync --open-pr`" + `

4. **Apply if asked, or just report:**
   - If the user said ` + "`--apply`" + `, run ` + "`gmh sync --open-pr`" + ` and stop.
   - Otherwise, print a clear status with concrete next commands.

5. **Open a PR if drift is significant:**
   - Use ` + "`gmh sync --open-pr --base " + in.LatestVersion + "`" + `
   - Title: ` + "`chore: harness sync to " + in.LatestVersion + "`" + `
   - Body: list which sensors changed, what's new, what to review

## Critical

- **DO NOT trust auto-report.** Re-read files. Re-run commands. (Invariante 19 / sensor 09.)
- **Be concise.** 1 paragraph of diagnosis + 1 list of next steps.
- **If everything is green**, say so and stop. Don't invent work.
`)

	return b.String()
}

func orNA(s string) string {
	if s == "" {
		return "_(not installed)_"
	}
	return s
}

func diffMsg(local, latest string) string {
	if local == "" {
		return "framework not yet installed (or VERSION missing)"
	}
	return fmt.Sprintf("local %s, latest %s", local, latest)
}
