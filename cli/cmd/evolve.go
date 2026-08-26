package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/evolve"
)

// EvolveCmd creates the `gmh evolve` command.
//
// Stanford-style outer loop over personas/skills using issue
// comment traces. v1 is deterministic (no live LLM): it writes a
// proposal the team-manager can run in Hermes. --apply stores the
// trace under harness/memory/traces/ and never overwrites persona
// markdown.
func EvolveCmd() *cobra.Command {
	var (
		fromDir string
		apply   bool
	)

	cmd := &cobra.Command{
		Use:   "evolve",
		Short: "Propose persona/skill patches from issue comment traces",
		Long: `Stanford-style outer loop over delivery personas and skills.

Reads issue comment traces from --from-dir and prints a
deterministic proposal (no live LLM). With --apply, stores the
trace under harness/memory/traces/<utc>/ — never overwrites
persona markdown in v1.

Hermes is the OS/tool harness (gh, filesystem). git-meta-harness
only stores the delivery context.

Examples:
  gmh evolve --from-dir ./comments
  gmh evolve --from-dir ./comments --apply`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			comments, err := evolve.LoadComments(fromDir)
			if err != nil {
				return err
			}

			cwd := getCwd(cmd)
			personas := listPersonaNames(filepath.Join(cwd, "harness", "personas"))
			proposal := evolve.Propose(comments, personas)
			fmt.Fprintln(cmd.OutOrStdout(), proposal)

			if apply {
				memoryDir := filepath.Join(cwd, "harness", "memory")
				traceDir, err := evolve.WriteTrace(memoryDir, comments, proposal)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), traceDir)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&fromDir, "from-dir", "",
		"Directory of comment JSON traces (required in v1)")
	cmd.Flags().BoolVar(&apply, "apply", false,
		"Write harness/memory/traces/<utc>/ (does not overwrite personas)")
	_ = cmd.MarkFlagRequired("from-dir")

	return cmd
}

// listPersonaNames returns basenames of *.md files in dir.
// Missing dir is ignored (v1: personas listing is best-effort).
func listPersonaNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.EqualFold(filepath.Ext(name), ".md") {
			continue
		}
		names = append(names, strings.TrimSuffix(name, filepath.Ext(name)))
	}
	return names
}
