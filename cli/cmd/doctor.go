package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/agentic"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/prompt"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/source"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

// DoctorCmd creates the `gmh doctor` command.
//
// `gmh doctor` runs health checks on the local project to verify
// that the meta-harness is correctly installed and configured.
//
// With --agent (e.g., --agent hermes), it generates a prompt for
// the team-manager of that agentic to perform a deep health check
// and detect drift. Use --apply to also run sync/update.
func DoctorCmd() *cobra.Command {
	var (
		fix       bool
		verbose   bool
		agentName string
		apply     bool
		noPrompt  bool
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks on the local meta-harness project",
		Long: `Run a series of health checks to verify the local project is
correctly set up with the meta-harness framework.

By default, runs quick local checks (15+) and prints a summary.

With --agent, generates a prompt for the team-manager of the
specified agentic (e.g., 'hermes', 'claude-code', 'codex',
'opencode') to perform a deep health check.

With --apply, runs sync/update automatically if drift is detected.

Examples:
  gmh doctor                         # Quick local checks
  gmh doctor --agent hermes          # Generate deep-check prompt for Hermes
  gmh doctor --agent hermes --apply  # Run sync if drift detected
  gmh doctor --fix                   # Auto-fix common issues`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			harnessDir := filepath.Join(cwd, "harness")

			// Quick local checks
			report := runLocalChecks(cwd, harnessDir, verbose)

			// Resolve versions
			src := source.NewClient("")
			latest, err := src.ResolveVersion("latest")
			if err != nil {
				ui.Warn("Could not resolve latest version: %v", err)
				latest = ""
			}
			localVersion := readLocalVersion(cwd)
			outOfDate := localVersion != "" && latest != "" && localVersion != latest

			// Detect agentics
			detected, _ := agentic.Detect()
			all := agentic.ListAll()
			allNames := make([]string, len(all))
			for i, a := range all {
				allNames[i] = a.String()
			}

			// Print summary
			ui.Header("Meta-Harness doctor report")
			ui.Info("Project:  %s", cwd)
			ui.Info("Local:    %s", orNA(localVersion))
			ui.Info("Latest:   %s", orNA(latest))
			if outOfDate {
				ui.Warn("Out of date by %d version(s)", versionDelta(localVersion, latest))
			} else {
				ui.OK("Up to date")
			}
			ui.Info("Agentic:  %s (installed: %v)", detected, allNames)
			ui.Info("")

			// Local checks
			ui.Header("Local checks (15+)")
			fails := 0
			for _, c := range report.checks {
				if c.fail {
					fails++
				}
				if c.fail || verbose {
					if c.fail {
						ui.Fail("%s", c.name)
					} else {
						ui.OK("%s", c.name)
					}
				}
			}
			ui.Info("")
			if fails == 0 {
				ui.OK("All local checks passed")
			} else {
				ui.Fail("%d check(s) failed", fails)
			}

			// Decide if deep check is needed
			needsDeep := fails > 0 || outOfDate || apply
			if !needsDeep && !verbose {
				ui.Info("")
				ui.Info("Run with --verbose to see all checks.")
				ui.Info("Run with --agent <hermes|claude-code|codex|opencode> for deep check.")
				return nil
			}

			// Resolve agentic
			targetAgent := agentic.Agentic(agentName)
			if targetAgent == "" {
				// Default to detected, fallback to hermes
				if detected != agentic.None {
					targetAgent = detected
				} else {
					targetAgent = agentic.Hermes
				}
				ui.Info("Using default agentic: %s (use --agent to override)", targetAgent)
			}
			if !agentic.IsValid(targetAgent) {
				ui.Fail("Unknown agentic: %s", targetAgent)
				ui.Info("Valid: hermes, claude-code, codex, opencode, copilot, cursor, devin")
				return fmt.Errorf("invalid agentic")
			}

			// Build prompt
			doctorIn := prompt.DoctorInput{
				Cwd:                cwd,
				LocalVersion:       localVersion,
				LatestVersion:      latest,
				OutOfDate:          outOfDate,
				DetectedAgentic:    detected.String(),
				AllAgentics:        allNames,
				LocalChecks:        report.checkStrings(),
				Fails:              report.failsList(),
				ProjectName:        filepath.Base(cwd),
				ProjectDescription: report.description,
			}
			p := prompt.DoctorPrompt(doctorIn)

			// Print or invoke
			if noPrompt {
				fmt.Println(p)
				return nil
			}

			ui.Header("Deep check via " + targetAgent.String())
			invocation, err := agentic.Invocation(targetAgent, "team-manager", p)
			if err != nil {
				ui.Info("Agentic %q has no CLI invocation. Prompt to delegate:", targetAgent)
				ui.Info("")
				fmt.Println(p)
				ui.Info("")
				return nil
			}

			ui.Info("Run the following to delegate the deep check to %s:", targetAgent)
			ui.Info("")
			ui.Info("  %s", invocation)
			ui.Info("")

			if apply {
				ui.Warn("--apply requested: this would run gmh sync/update")
				ui.Info("Auto-applying not yet implemented. Run the command above to delegate.")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false,
		"Auto-fix common issues (destructive)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"Show all checks including passing ones")
	cmd.Flags().StringVar(&agentName, "agent", "",
		"Agentic to delegate the deep check to (hermes, claude-code, codex, opencode)")
	cmd.Flags().BoolVar(&apply, "apply", false,
		"Auto-apply sync/update if drift is detected")
	cmd.Flags().BoolVar(&noPrompt, "no-prompt", false,
		"Just print the prompt (don't suggest how to invoke)")

	return cmd
}

// localCheckResult is one doctor check.
type localCheckResult struct {
	name string
	pass bool
	fail bool
}

type doctorReport struct {
	checks      []localCheckResult
	description string
}

func (r *doctorReport) checkStrings() []string {
	out := make([]string, 0, len(r.checks))
	for _, c := range r.checks {
		prefix := "PASS"
		if c.fail {
			prefix = "FAIL"
		}
		out = append(out, fmt.Sprintf("%s: %s", prefix, c.name))
	}
	return out
}

func (r *doctorReport) failsList() []string {
	var out []string
	for _, c := range r.checks {
		if c.fail {
			out = append(out, c.name)
		}
	}
	return out
}

func runLocalChecks(cwd, harnessDir string, verbose bool) *doctorReport {
	rep := &doctorReport{}
	check := func(name string, ok bool) {
		c := localCheckResult{name: name, pass: ok}
		if !ok {
			c.fail = true
		}
		rep.checks = append(rep.checks, c)
	}

	// 1. harness/ exists
	check("harness/ directory exists", dirExists(harnessDir))

	// 2. AGENTS.md present
	check("harness/AGENTS.md present", fileExists(filepath.Join(harnessDir, "AGENTS.md")))

	// 3. 19+ invariants
	agentsPath := filepath.Join(harnessDir, "AGENTS.md")
	if data, err := os.ReadFile(agentsPath); err == nil {
		invCount := countMatches(string(data), `^[0-9]+\. \*\*`)
		check(fmt.Sprintf("≥19 invariants in AGENTS.md (found %d)", invCount), invCount >= 19)
	} else {
		check("AGENTS.md readable", false)
	}

	// 4. Sensors 00-09
	for i := 0; i <= 9; i++ {
		name := fmt.Sprintf("sensor/0%d-*", i)
		matches, _ := filepath.Glob(filepath.Join(harnessDir, "sensors", fmt.Sprintf("0%d-*", i)+".md"))
		check(name+" present", len(matches) > 0)
	}

	// 5. Domain-experts specialized (no domain-expert.md generic)
	check("No generic domain-expert.md", !fileExists(filepath.Join(harnessDir, "personas", "domain-expert.md")))

	// 6. scripts/
	check("scripts/smoke-test.sh present", fileExists(filepath.Join(harnessDir, "scripts", "smoke-test.sh")))
	check("scripts/check-stack-versions.sh present", fileExists(filepath.Join(harnessDir, "scripts", "check-stack-versions.sh")))

	// 7. ADR-0014 verify-after-build
	adrPath := filepath.Join(harnessDir, "contrib", "design-decisions.md")
	if data, err := os.ReadFile(adrPath); err == nil {
		check("ADR-0014 (verify-after-build) present", contains(string(data), "ADR-0014"))
		check("ADR-0015 (release pipeline) present", contains(string(data), "ADR-0015"))
		check("ADR-0016 (gmh CLI) present", contains(string(data), "ADR-0016"))
	} else {
		check("design-decisions.md readable", false)
	}

	// 8. bootstrap.md
	check("bootstrap.md present", fileExists(filepath.Join(harnessDir, "bootstrap.md")))

	// 9. CLAUDE.md
	check("CLAUDE.md present", fileExists(filepath.Join(harnessDir, "CLAUDE.md")))

	// 10. seed
	check("seed/meta-harness-seed.md present", fileExists(filepath.Join(harnessDir, "seed", "meta-harness-seed.md")))

	// 11. README and docs
	check("README.md at project root", fileExists(filepath.Join(cwd, "..", "README.md")) || fileExists(filepath.Join(cwd, "README.md")))

	return rep
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func countMatches(s, pattern string) int {
	// Naive line-count; the regex isn't compiled for performance
	count := 0
	line := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			if matchPattern(line, pattern) {
				count++
			}
			line = ""
		} else {
			line += string(s[i])
		}
	}
	if line != "" && matchPattern(line, pattern) {
		count++
	}
	return count
}

func matchPattern(s, pattern string) bool {
	// Tiny pattern matcher: `^[0-9]+\. \*\*`
	// We just need "starts with one or more digits, then '. **'"
	if len(s) == 0 {
		return false
	}
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return false
	}
	rest := s[i:]
	return len(rest) >= 4 && rest[0] == '.' && rest[1] == ' ' && rest[2] == '*' && rest[3] == '*'
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func versionDelta(local, latest string) int {
	if local == "" || latest == "" {
		return 0
	}
	lp := semverTuple(local)
	hp := semverTuple(latest)
	return hp[0]*10000+hp[1]*100+hp[2] - (lp[0]*10000 + lp[1]*100 + lp[2])
}
