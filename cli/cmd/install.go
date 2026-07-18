// Package cmd implements the gmh subcommands.
package cmd

import (
	"github.com/spf13/cobra"
)

// InstallCmd creates the `gmh install` command.
//
// `gmh install` copies the `harness/` directory from a specific
// version of the git-meta-harness repo into the current project.
// This is the entry point for any new project that wants to adopt
// the meta-harness.
func InstallCmd() *cobra.Command {
	var (
		toVersion string
		force     bool
		skipCheck bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install meta-harness into the current project",
		Long: `Install the meta-harness framework into the current project.

This creates a 'harness/' directory at the project root containing
all the framework files (AGENTS.md, personas, sensors, scripts, etc.).

Examples:
  gmh install                     # Install latest version
  gmh install --to v1.5.0         # Install a specific version
  gmh install --force             # Overwrite existing harness/
  gmh install --skip-check        # Don't run doctor after install`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement
			// 1. Resolve version (latest if empty)
			// 2. Download tarball of meta-harness@version
			// 3. Extract harness/ into cwd
			// 4. Optionally run doctor
			// 5. Print success message
			return nil
		},
	}

	cmd.Flags().StringVar(&toVersion, "to", "",
		"Specific version to install (e.g., v1.5.0). Default: latest")
	cmd.Flags().BoolVarP(&force, "force", "f", false,
		"Overwrite existing harness/ directory")
	cmd.Flags().BoolVar(&skipCheck, "skip-check", false,
		"Skip running gmh doctor after install")

	return cmd
}
