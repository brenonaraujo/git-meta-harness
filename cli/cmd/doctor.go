package cmd

import "github.com/spf13/cobra"

// DoctorCmd creates the `gmh doctor` command.
//
// `gmh doctor` runs health checks on the local project to verify
// that the meta-harness is correctly installed and configured.
// It is the equivalent of `harness/scripts/smoke-test.sh` but
// written in Go and always available via the CLI.
func DoctorCmd() *cobra.Command {
	var (
		fix     bool
		verbose bool
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks on the local meta-harness project",
		Long: `Run a series of health checks to verify the local project is
correctly set up with the meta-harness framework.

Checks include:
  - harness/ directory exists and has expected structure
  - All 19 invariants from AGENTS.md are present
  - 9 sensors (00-08) + sensor 09 are present
  - Domain-experts are specialized (no domain-expert.md generic)
  - Smart routing is documented
  - check-stack-versions.sh passes
  - GitHub labels are created (type/* + domain/*)

Examples:
  gmh doctor                # Run all checks
  gmh doctor --fix          # Auto-fix common issues (e.g., missing files)
  gmh doctor --verbose      # Show all checks (including passing)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement
			// 1. Check harness/ exists
			// 2. Check 19 invariants in AGENTS.md
			// 3. Check 10 sensors (00-09)
			// 4. Check no domain-expert.md generic
			// 5. Check check-stack-versions.sh passes
			// 6. Print report
			return nil
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false,
		"Auto-fix common issues (destructive)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"Show all checks including passing ones")

	return cmd
}
