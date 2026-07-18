package cmd

import "github.com/spf13/cobra"

// PluginsCmd creates the `gmh plugins` parent command.
//
// `gmh plugins` is a placeholder for future plugin management.
// Plugins extend the gmh CLI itself (e.g., add new subcommands).
func PluginsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Manage gmh plugins (experimental)",
		Long: `Manage plugins for the gmh CLI.

Plugins extend the gmh CLI itself with new subcommands. This is
experimental and the API is not stable yet.

Subcommands:
  list       List installed plugins
  install    Install a plugin from the registry
  remove     Remove a plugin

Examples:
  gmh plugins list
  gmh plugins install my-plugin
  gmh plugins remove my-plugin`,
	}

	cmd.AddCommand(pluginsListCmd())
	cmd.AddCommand(pluginsInstallCmd())
	cmd.AddCommand(pluginsRemoveCmd())

	return cmd
}

func pluginsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: list plugins in ~/.gmh/plugins/
			return nil
		},
	}
}

func pluginsInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <name>",
		Short: "Install a plugin from the registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: download plugin to ~/.gmh/plugins/<name>/
			return nil
		},
	}
}

func pluginsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: remove ~/.gmh/plugins/<name>/
			return nil
		},
	}
}
