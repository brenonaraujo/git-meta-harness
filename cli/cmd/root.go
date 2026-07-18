// Package cmd implements the gmh subcommands.
package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version is set at build time via -ldflags in the parent
	// package main (in cli/main.go).
	Version = "dev"
	// Commit is set at build time via -ldflags.
	Commit = "unknown"
	// Date is set at build time via -ldflags.
	Date = "unknown"
)

// Execute builds the root command and runs it. Called from
// main.go in the parent package.
func Execute(version, commit, date string) error {
	Version = version
	Commit = commit
	Date = date

	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Error: %v", err))
		return err
	}
	return nil
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gmh",
		Short: "git-meta-harness CLI",
		Long: color.New(color.FgCyan, color.Bold).Sprint("gmh") +
			` — the official CLI for the meta-harness framework.

` + color.New(color.FgYellow).Sprint("Usage:") + ` gmh <command> [flags]

` + color.New(color.FgYellow).Sprint("Common commands:") + `
  install    Install meta-harness into a project
  sync       Sync the local project with the latest version
  update     Update to a specific version
  doctor     Health check the local project
  skills     Install/list skills
  personas   Install/list personas
  plugins    Install/list plugins

` + color.New(color.FgYellow).Sprint("Examples:") + `
  gmh install                    # Install latest version into ./harness/
  gmh install --to v1.5.0        # Install specific version
  gmh sync                       # Pull latest version into existing project
  gmh doctor                     # Check project is healthy
  gmh skills install code-graph  # Add a skill

` + color.New(color.FgYellow).Sprint("Docs:") + `  https://github.com/brenonaraujo/git-meta-harness
`,
		Version: fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date),
	}

	// Persistent flags
	rootCmd.PersistentFlags().StringP("cwd", "C", ".",
		"Working directory (where harness/ lives or will live)")
	rootCmd.PersistentFlags().String("source", "brenonaraujo/git-meta-harness",
		"Source repository (owner/repo) to pull from")
	rootCmd.PersistentFlags().Bool("dry-run", false,
		"Print what would be done without actually doing it")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false,
		"Verbose output")

	// Bind to viper for env override
	_ = viper.BindPFlag("cwd", rootCmd.PersistentFlags().Lookup("cwd"))
	_ = viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source"))
	_ = viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.SetEnvPrefix("GMH")
	viper.AutomaticEnv()

	// Add subcommands
	rootCmd.AddCommand(InstallCmd())
	rootCmd.AddCommand(SyncCmd())
	rootCmd.AddCommand(UpdateCmd())
	rootCmd.AddCommand(DoctorCmd())
	rootCmd.AddCommand(AgentsCmd())
	rootCmd.AddCommand(SkillsCmd())
	rootCmd.AddCommand(PersonasCmd())
	rootCmd.AddCommand(PluginsCmd())
	rootCmd.AddCommand(VersionCmd())

	return rootCmd
}
