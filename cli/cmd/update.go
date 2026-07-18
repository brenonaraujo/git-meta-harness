package cmd

import "github.com/spf13/cobra"

// UpdateCmd creates the `gmh update` command.
//
// `gmh update` is an alias for `gmh sync --to <version>`. Useful
// when you want to pin a specific version (e.g., for reproducibility).
func UpdateCmd() *cobra.Command {
	var (
		toVersion string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the meta-harness framework to a specific version",
		Long: `Update the local harness/ directory to a specific version of
the meta-harness framework. By default, this is equivalent to
'gmh sync'.

Examples:
  gmh update                    # Update to latest
  gmh update --to v1.5.0        # Pin to v1.5.0
  gmh update --to v1.4.0 --force  # Downgrade to v1.4.0 (destructive)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement
			// 1. Resolve version (latest if --to empty)
			// 2. Download tarball of meta-harness@version
			// 3. Extract harness/ into cwd
			// 4. Run gmh doctor
			return nil
		},
	}

	cmd.Flags().StringVar(&toVersion, "to", "",
		"Target version (e.g., v1.5.0). Default: latest")
	cmd.Flags().BoolVarP(&force, "force", "f", false,
		"Allow downgrades (destructive)")

	return cmd
}
