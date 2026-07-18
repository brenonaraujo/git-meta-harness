package cmd

import "github.com/spf13/cobra"

// SyncCmd creates the `gmh sync` command.
//
// `gmh sync` updates the local project's `harness/` directory with
// the latest version from the meta-harness repo, while preserving
// any local customizations (e.g., materialized personas).
func SyncCmd() *cobra.Command {
	var (
		dryRun    bool
		keepLocal bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync the local project with the latest meta-harness version",
		Long: `Update the local harness/ directory with the latest version
of the meta-harness framework while preserving local customizations.

This is the typical 'bump version' operation for an existing
meta-harness project.

Examples:
  gmh sync                    # Pull latest version
  gmh sync --dry-run          # Show what would change
  gmh sync --keep-local       # Don't overwrite locally-modified files`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement
			// 1. Detect current version from harness/VERSION
			// 2. Get latest version
			// 3. Diff harness/ local vs remote
			// 4. Show changes, prompt if not --dry-run
			// 5. Update files (preserve locally-modified if --keep-local)
			// 6. Run gmh doctor
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Show what would change without making changes")
	cmd.Flags().BoolVar(&keepLocal, "keep-local", true,
		"Preserve locally-modified files (default true)")

	return cmd
}
