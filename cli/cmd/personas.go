package cmd

import "github.com/spf13/cobra"

// PersonasCmd creates the `gmh personas` parent command.
//
// `gmh personas` manages domain-expert specializations. Each
// domain-expert-<domínio> is a specialized persona for a specific
// business domain (e.g., banking, retail, mandai).
func PersonasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "personas",
		Short: "Manage meta-harness domain-experts",
		Long: `Manage domain-expert personas for the meta-harness framework.

Domain-experts are ALWAYS specialized (per invariant 12). A generic
'domain-expert.md' is forbidden. Each domain-expert-<domínio> is
a persona for a specific business domain.

Subcommands:
  list       List installed personas
  create     Create a new domain-expert-<domínio> from template
  remove     Remove a domain-expert
  available  List personas available in the registry

Examples:
  gmh personas list
  gmh personas create --domain banking
  gmh personas remove domain-expert-banking
  gmh personas available`,
	}

	cmd.AddCommand(personasListCmd())
	cmd.AddCommand(personasCreateCmd())
	cmd.AddCommand(personasRemoveCmd())

	return cmd
}

func personasListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed personas (including domain-experts)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: list harness/personas/*.md
			return nil
		},
	}
}

func personasCreateCmd() *cobra.Command {
	var (
		domain      string
		fromGeneric bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new domain-expert-<domínio> from template",
		Long: `Create a new specialized domain-expert-<domínio> persona
from the domain-expert.template.md.

Examples:
  gmh personas create --domain banking
  gmh personas create --domain retail --from-generic`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: copy template to harness/personas/domain-expert-<domain>.md
			return nil
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "",
		"Domain name (e.g., banking, retail, healthcare). Required.")
	cmd.Flags().BoolVar(&fromGeneric, "from-generic", false,
		"Convert an existing generic domain-expert.md (deprecated path)")
	_ = cmd.MarkFlagRequired("domain")

	return cmd
}

func personasRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a domain-expert (use with care)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: remove harness/personas/<name>.md
			return nil
		},
	}
}
