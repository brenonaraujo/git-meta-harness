package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/personas"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

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
  gmh personas create --domain banking --context "Pix + Open Banking"
  gmh personas create --domain retail --from-spec ./SPEC.md
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
			cwd := getCwd(cmd)
			harnessDir := filepath.Join(cwd, "harness")
			names, err := personas.List(harnessDir)
			if err != nil {
				return err
			}
			ui.Header("Installed personas")
			if len(names) == 0 {
				ui.Info("No personas found in %s", filepath.Join(harnessDir, "personas"))
				return nil
			}
			for _, name := range names {
				ui.Step("%s", name)
			}
			return nil
		},
	}
}

func personasCreateCmd() *cobra.Command {
	var (
		domain      string
		fromGeneric bool
		context     string
		fromSpec    string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new domain-expert-<domínio> from template",
		Long: `Create a new specialized domain-expert-<domínio> persona
from the domain-expert.template.md.

Examples:
  gmh personas create --domain banking
  gmh personas create --domain retail --from-generic
  gmh personas create --domain banking --context "Pix + Open Banking"
  gmh personas create --domain retail --from-spec ./SPEC.md`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd := getCwd(cmd)
			harnessDir := filepath.Join(cwd, "harness")

			projectContext := context
			if fromSpec != "" {
				data, err := os.ReadFile(fromSpec)
				if err != nil {
					return fmt.Errorf("read --from-spec %s: %w", fromSpec, err)
				}
				spec := string(data)
				if projectContext != "" {
					projectContext = projectContext + "\n" + spec
				} else {
					projectContext = spec
				}
			}
			if fromGeneric {
				ui.Warn("--from-generic is deprecated and ignored; create a specialized domain-expert instead")
			}

			path, err := personas.Create(harnessDir, domain, projectContext)
			if err != nil {
				return err
			}
			ui.OK("%s", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "",
		"Domain name (e.g., banking, retail, healthcare). Required.")
	cmd.Flags().BoolVar(&fromGeneric, "from-generic", false,
		"Convert an existing generic domain-expert.md (deprecated path)")
	cmd.Flags().StringVar(&context, "context", "",
		"Project context appended to the new domain-expert persona")
	cmd.Flags().StringVar(&fromSpec, "from-spec", "",
		"Read a spec file and append its contents to --context")
	_ = cmd.MarkFlagRequired("domain")

	return cmd
}

func personasRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a domain-expert (use with care)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd := getCwd(cmd)
			harnessDir := filepath.Join(cwd, "harness")
			name := strings.TrimSpace(args[0])
			if err := personas.Remove(harnessDir, name); err != nil {
				return err
			}
			ui.OK("Removed %s", name)
			return nil
		},
	}
}
