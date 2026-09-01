package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/hermes"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/loop"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/seed"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/soul"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

// SeedCmd is the one-shot: harness + personas + CI + issue 0 + Hermes loop.
func SeedCmd() *cobra.Command {
	var (
		specPath   string
		describe   string
		domain     string
		stack      string
		githubRepo string
		from       string
		skipAgents bool
		skipGitHub bool
		skipCron   bool
		skipLoop   bool
		jsonOut    bool
		inDir      string
	)

	cmd := &cobra.Command{
		Use:   "seed [name]",
		Short: "Materialize a project: harness, personas, CI, issue 0, Hermes loop",
		Long: `gmh seed is the command an agent runs when the user says
"use git-meta-harness to create a project".

It copies harness/, writes docs/SPEC.md, creates a specialized
domain-expert, copies CI (full Go/Nuxt or thin contracts), writes
the Hermes loop scripts, materializes --no-skills profiles,
creates GitHub labels + issue 0, and installs the two Hermes cron
jobs (team-manager heartbeat + supervisor). team-manager does not
implement feature code.

Examples:
  gmh seed booking-saas --describe "SaaS de agendamento online" --github acme/booking-saas
  gmh seed booking-saas --spec ./SPEC.md --from ~/Projects/git-meta-harness/harness
  gmh seed --in . --spec docs/SPEC.md --github acme/existing`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			cwd := getCwd(cmd)
			if cwd == "" || cwd == "." {
				cwd, _ = os.Getwd()
			}
			cwd, err := filepath.Abs(cwd)
			if err != nil {
				return err
			}
			target := cwd
			if inDir != "" {
				target = inDir
			} else if name != "" {
				target = name
			}
			if !filepath.IsAbs(target) {
				target = filepath.Join(cwd, target)
			}
			target, err = filepath.Abs(target)
			if err != nil {
				return err
			}
			if from != "" && !filepath.IsAbs(from) {
				from = filepath.Join(cwd, from)
			}
			harnessSrc := seed.FindHarnessSrc(from, cwd)
			if githubRepo == "" {
				skipGitHub = true
			}
			opts := seed.Options{
				Name:       name,
				Describe:   describe,
				SpecPath:   specPath,
				Domain:     domain,
				Stack:      stack,
				GitHubRepo: githubRepo,
				TargetDir:  target,
				HarnessSrc: harnessSrc,
				SkipAgents: skipAgents,
				SkipGitHub: skipGitHub,
				SkipCron:   skipCron || skipLoop,
			}
			if opts.Name == "" {
				opts.Name = filepath.Base(target)
			} else {
				opts.Name = filepath.Base(opts.Name)
			}

			ui.Header("gmh seed — materialize meta-harness (v1.16.0)")
			res, err := seed.Apply(opts)
			if err != nil {
				return err
			}
			ui.OK("project %s  domain=%s  ci=%s", res.Slug, res.Domain, res.CIKind)
			ui.Info("root: %s", res.TargetDir)

			if !skipAgents {
				if err := materializePersonas(res.TargetDir, res.Personas, res.Slug); err != nil {
					ui.Warn("personas: %v", err)
				}
			} else {
				ui.Info("skipped Hermes profiles (--no-agents)")
			}

			if !skipGitHub {
				if err := seedGitHub(res, githubRepo); err != nil {
					ui.Warn("github: %v", err)
					ui.Info("issue 0 body is at %s", filepath.Join(res.TargetDir, ".github", "ISSUE_0.md"))
				}
			} else if githubRepo == "" {
				ui.Info("skipped GitHub labels/issue 0 (pass --github owner/repo)")
			}

			if !skipCron && !skipLoop {
				if err := installLoopCrons(res); err != nil {
					ui.Warn("cron: %v", err)
					ui.Info("install later with: gmh loop install -C %s", res.TargetDir)
				}
			}

			rep := loop.Doctor(res.TargetDir)
			if !rep.Pass {
				ui.Fail("loop doctor failed: %s", strings.Join(rep.Failed, ", "))
			} else {
				ui.OK("gmh loop doctor: pass")
			}

			ui.Info("")
			ui.Header("Next")
			ui.Step("cd %s", res.TargetDir)
			ui.Step("git init && git add -A && git commit -m 'chore: seed meta-harness'")
			ui.Step("team-manager owns the board — do not implement the product in this chat")
			ui.Step("gateway must be up: hermes gateway install --start-now")
			if jsonOut {
				fmt.Printf("{\"slug\":%q,\"domain\":%q,\"root\":%q,\"ci\":%q,\"issue0\":%q}\n",
					res.Slug, res.Domain, res.TargetDir, res.CIKind, res.Issue0Title)
			}
			if !rep.Pass {
				return fmt.Errorf("loop doctor failed")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&specPath, "spec", "", "Path to functional spec (markdown)")
	cmd.Flags().StringVar(&describe, "describe", "", "Short product description (written to docs/SPEC.md)")
	cmd.Flags().StringVar(&domain, "domain", "", "Override domain slug (creates domain-expert-<slug>)")
	cmd.Flags().StringVar(&stack, "stack", "", "Stack hint (go,nuxt,postgresql → full CI; otherwise contracts CI)")
	cmd.Flags().StringVar(&githubRepo, "github", "", "owner/repo for labels + issue 0")
	cmd.Flags().StringVar(&from, "from", "", "Path to git-meta-harness/harness (default: detect framework repo)")
	cmd.Flags().BoolVar(&skipAgents, "no-agents", false, "Do not create Hermes profiles")
	cmd.Flags().BoolVar(&skipGitHub, "no-github", false, "Do not create labels/issue 0 via gh")
	cmd.Flags().BoolVar(&skipCron, "no-cron", false, "Do not install Hermes cron jobs")
	cmd.Flags().BoolVar(&skipLoop, "no-loop", false, "Alias of --no-cron")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print a JSON summary line")
	cmd.Flags().StringVar(&inDir, "in", "", "Project directory (in-place seed; default: [name] or cwd)")
	return cmd
}

func materializePersonas(projectRoot string, names []string, slug string) error {
	client, err := hermes.NewClient("")
	if err != nil {
		return err
	}
	if !client.IsInstalled() {
		return fmt.Errorf("hermes not installed at %s", client.Home)
	}
	ver := frameworkVersion(projectRoot)
	for _, name := range names {
		personaPath := filepath.Join(projectRoot, "harness", "personas", name+".md")
		if _, err := os.Stat(personaPath); err != nil {
			ui.Warn("  skip %s (no persona file)", name)
			continue
		}
		desc := fmt.Sprintf("meta-harness %s for %s", name, slug)
		if err := ensureHermesProfile(client, name, desc); err != nil {
			ui.Warn("  %s: %v", name, err)
			continue
		}
		body, err := soul.Generate(personaPath, ver)
		if err != nil {
			ui.Warn("  soul %s: %v", name, err)
			continue
		}
		if err := client.WriteSoul(name, body); err != nil {
			return err
		}
		home, _ := os.UserHomeDir()
		_, _ = client.EnsureExternalDirs(name, []string{filepath.Join(home, ".hermes", "skills")})
		_, _ = client.WipeProfileLocalSkills(name)
		ui.OK("  profile %s", name)
	}
	return nil
}

func seedGitHub(res seed.ApplyResult, repo string) error {
	if repo == "" {
		repo = res.GitHubRepo
	}
	if repo == "" {
		return fmt.Errorf("no --github owner/repo")
	}
	for _, l := range loop.CanonicalLabels() {
		cmd := exec.Command("gh", "label", "create", l.Name,
			"--repo", repo,
			"--color", l.Color,
			"--description", l.Description,
			"--force")
		_ = cmd.Run()
	}
	body := filepath.Join(res.TargetDir, ".github", "ISSUE_0.md")
	cmd := exec.Command("gh", "issue", "create",
		"--repo", repo,
		"--title", res.Issue0Title,
		"--body-file", body,
		"--label", "triage",
		"--label", "type/feature")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh issue create: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	ui.OK("issue 0: %s", strings.TrimSpace(string(out)))
	return nil
}

func installLoopCrons(res seed.ApplyResult) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	scriptsDir := filepath.Join(home, ".hermes", "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		return err
	}
	wrapper := loop.HermesScriptWrapper(res.TargetDir, res.Slug)
	scriptName := res.Slug + "-spawn-tm.sh"
	scriptPath := filepath.Join(scriptsDir, scriptName)
	if err := os.WriteFile(scriptPath, []byte(wrapper), 0o755); err != nil {
		return err
	}
	ui.OK("cron script %s", scriptPath)

	for _, job := range res.CronJobs {
		argv := job.Argv()
		cmd := exec.Command(argv[0], argv[1:]...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			ui.Warn("  %s: %v (%s)", job.Name, err, strings.TrimSpace(string(out)))
			continue
		}
		ui.OK("  cron %s", job.Name)
	}
	ui.Info("gateway must be running or cron never fires: hermes gateway install --start-now")
	return nil
}
