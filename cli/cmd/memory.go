package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/harnessmem"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

// MemoryCmd creates the `gmh memory` parent command.
//
// `gmh memory` writes and shows the delivery-harness memory
// snapshot (personas, skills, project-specific agent context).
func MemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Write and show delivery-harness memory snapshots",
		Long: `Persist generated delivery-harness memory.

The snapshot at harness/memory/snapshot.json records personas,
skills, Hermes profiles, and project context. Hermes is the
OS/tool harness (terminal, fs, gh, browsers); this snapshot is
the delivery-harness memory produced by git-meta-harness.

Subcommands:
  write   Scan the project and write harness/memory/snapshot.json
  show    Print the current snapshot as JSON

Examples:
  gmh memory write
  gmh memory show`,
	}

	cmd.AddCommand(memoryWriteCmd())
	cmd.AddCommand(memoryShowCmd())
	return cmd
}

func memoryWriteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "write",
		Short: "Build and write harness/memory/snapshot.json",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd := getCwd(cmd)
			harnessDir := filepath.Join(cwd, "harness")

			version := Version
			if version == "" {
				version = "dev"
			}

			runtime := "none"
			hermesHome := ""
			if home, err := os.UserHomeDir(); err == nil {
				candidate := filepath.Join(home, ".hermes")
				if info, err := os.Stat(candidate); err == nil && info.IsDir() {
					runtime = "hermes"
					hermesHome = candidate
				}
			}

			snap, err := harnessmem.Build(harnessDir, hermesHome, runtime, version, filepath.Base(cwd))
			if err != nil {
				return err
			}
			if err := harnessmem.Write(harnessDir, snap); err != nil {
				return err
			}
			ui.OK("Wrote %s", filepath.Join(harnessDir, "memory", "snapshot.json"))
			return nil
		},
	}
}

func memoryShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print harness/memory/snapshot.json",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd := getCwd(cmd)
			harnessDir := filepath.Join(cwd, "harness")
			snap, err := harnessmem.Read(harnessDir)
			if err != nil {
				return err
			}
			b, err := json.MarshalIndent(snap, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal snapshot: %w", err)
			}
			fmt.Println(string(b))
			return nil
		},
	}
}
