package cmd

import "github.com/spf13/cobra"

// SkillsCmd creates the `gmh skills` parent command.
//
// `gmh skills` manages skills (atomic capabilities like
// code-graph, i18n, tdd-go) that can be added to a project.
func SkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage meta-harness skills",
		Long: `Manage skills for the meta-harness framework.

Skills are atomic capabilities (e.g., code-graph, i18n, tdd-go)
that can be added to a project. Each skill is a single .md file
that describes when to use the skill and how.

Subcommands:
  list       List installed skills
  install    Install a skill
  remove     Remove a skill
  available  List skills available in the registry

Examples:
  gmh skills list
  gmh skills install code-graph
  gmh skills remove i18n
  gmh skills available`,
	}

	cmd.AddCommand(skillsListCmd())
	cmd.AddCommand(skillsInstallCmd())
	cmd.AddCommand(skillsRemoveCmd())
	cmd.AddCommand(skillsAvailableCmd())

	return cmd
}

func skillsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: list harness/skills/*.md
			return nil
		},
	}
}

func skillsInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <name>",
		Short: "Install a skill from the registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: download skill from registry
			return nil
		},
	}
}

func skillsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: remove harness/skills/<name>.md
			return nil
		},
	}
}

func skillsAvailableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "available",
		Short: "List skills available in the registry",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: list from remote registry
			return nil
		},
	}
}
