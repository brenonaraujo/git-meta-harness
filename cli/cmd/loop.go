package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/loop"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

// LoopCmd installs and validates the Hermes cron pooling loop.
func LoopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loop",
		Short: "Install / doctor the Hermes cron pooling loop",
		Long: `GitHub cannot webhook a local Hermes process. The loop is a
Hermes cron heartbeat that wakes team-manager. team-manager
moves labels and spawns personas; it never implements.

Subcommands:
  install   Write loop scripts + install the two cron jobs
  doctor    Filesystem liveness (sensor 14)
  status    hermes cron list filtered to this project`,
	}
	cmd.AddCommand(loopInstallCmd())
	cmd.AddCommand(loopDoctorCmd())
	cmd.AddCommand(loopStatusCmd())
	return cmd
}

func loopInstallCmd() *cobra.Command {
	var (
		slug   string
		github string
		domain string
		noCron bool
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Write loop scripts and install Hermes cron jobs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd := getCwd(cmd)
			if slug == "" {
				slug = filepath.Base(cwd)
			}
			cfg := loop.Config{
				ProjectRoot: cwd,
				ProjectSlug: slug,
				ProjectName: slug,
				GitHubRepo:  github,
				Domain:      domain,
			}
			if err := loop.WriteProjectFiles(cfg, ""); err != nil {
				return err
			}
			ui.OK("loop scripts written under harness/scripts/loop/")

			if noCron {
				return nil
			}
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			scriptsDir := filepath.Join(home, ".hermes", "scripts")
			if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
				return err
			}
			wrapper := loop.HermesScriptWrapper(cwd, slug)
			scriptName := slug + "-spawn-tm.sh"
			if err := os.WriteFile(filepath.Join(scriptsDir, scriptName), []byte(wrapper), 0o755); err != nil {
				return err
			}
			for _, job := range loop.CronCommands(cfg) {
				argv := job.Argv()
				out, err := exec.Command(argv[0], argv[1:]...).CombinedOutput()
				if err != nil {
					ui.Warn("%s: %v (%s)", job.Name, err, strings.TrimSpace(string(out)))
					continue
				}
				ui.OK("cron %s (%s, monitor=false)", job.Name, job.Schedule)
			}
			ui.Info("gateway: hermes gateway install --start-now")
			return nil
		},
	}
	cmd.Flags().StringVar(&slug, "slug", "", "Project slug (default: directory name)")
	cmd.Flags().StringVar(&github, "github", "", "owner/repo")
	cmd.Flags().StringVar(&domain, "domain", "", "domain slug")
	cmd.Flags().BoolVar(&noCron, "no-cron", false, "Only write scripts, do not call hermes cron")
	return cmd
}

func loopDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Validate that the harness loop is wired (sensor 14)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd := getCwd(cmd)
			rep := loop.Doctor(cwd)
			ui.Header("gmh loop doctor")
			for _, c := range rep.Checks {
				if c.Pass {
					ui.OK("%s", c.Name)
				} else {
					ui.Fail("%s %s", c.Name, c.Detail)
				}
			}
			if !rep.Pass {
				return fmt.Errorf("%d check(s) failed: %s", len(rep.Failed), strings.Join(rep.Failed, ", "))
			}
			ui.OK("loop liveness OK")
			return nil
		},
	}
}

func loopStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Hermes cron jobs for this project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd := getCwd(cmd)
			slug := filepath.Base(cwd)
			out, err := exec.Command("hermes", "cron", "list").CombinedOutput()
			if err != nil {
				return fmt.Errorf("hermes cron list: %w (%s)", err, strings.TrimSpace(string(out)))
			}
			ui.Header("hermes cron (%s)", slug)
			fmt.Print(string(out))
			if strings.Contains(string(out), "--monitor") {
				ui.Fail("a job still has monitor — remove it (triage snapshots would freeze the loop)")
			}
			return nil
		},
	}
}
