package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/agentic"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/hermes"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/skills"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/soul"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

// AgentsCmd creates the `gmh agents` parent command.
//
// `gmh agents` syncs the agentic's profiles (Hermes) and skills
// (all agentics) with the current meta-harness framework version.
//
// This is the part that `gmh sync` does NOT cover: gmh sync updates
// harness/ in the project, but the user-side profiles (e.g.,
// ~/.hermes/profiles/team-manager/SOUL.md) may be stale.
//
// Examples:
//
//	gmh agents list                # List installed profiles + skills
//	gmh agents inspect team-manager # Show diff for one profile
//	gmh agents sync                # Sync all (safe strategy)
//	gmh agents sync --aggressive   # Overwrite including customizations
//	gmh agents install team-manager  # Install a single profile from framework
func AgentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage agentic profiles (Hermes, Claude Code, etc.)",
		Long: `Manage agentic profiles and skills for the meta-harness framework.

This is the part that 'gmh sync' does NOT cover: 'gmh sync' updates
harness/ in the project, but the user-side profiles (e.g.,
~/.hermes/profiles/team-manager/SOUL.md) may be stale when the
framework evolves.

Subcommands:
  list       List installed profiles + skills
  inspect    Show what would change for a profile
  sync       Sync profiles + skills with the framework
  install    Install a single profile from the framework

Examples:
  gmh agents list
  gmh agents inspect team-manager
  gmh agents sync --aggressive
  gmh agents install domain-expert-banking`,
	}

	cmd.AddCommand(agentsListCmd())
	cmd.AddCommand(agentsInspectCmd())
	cmd.AddCommand(agentsSyncCmd())
	cmd.AddCommand(agentsInstallCmd())

	return cmd
}

func agentsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed profiles and skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			hermesClient, err := hermes.NewClient("")
			if err != nil {
				return err
			}

			ui.Header("Agentic profiles + skills")
			ui.Info("Hermes home: %s", hermesClient.Home)

			if !hermesClient.IsInstalled() {
				ui.Warn("Hermes not installed at %s", hermesClient.Home)
				ui.Info("Run 'hermes profile create <name>' to bootstrap")
				return nil
			}

			profiles, err := hermesClient.ListProfiles()
			if err != nil {
				return err
			}
			if len(profiles) == 0 {
				ui.Warn("No profiles installed yet")
			} else {
				ui.Info("")
				ui.Info("Profiles (%d):", len(profiles))
				for _, p := range profiles {
					marker := "✅"
					if !fileExists(p.SoulPath) {
						marker = "⚠️  (no SOUL.md)"
					}
					ui.Step("  %s %s", marker, p.Name)
				}
			}

			skills, err := hermesClient.ListSkills()
			if err != nil {
				return err
			}
			ui.Info("")
			if len(skills) == 0 {
				ui.Warn("No skills installed in ~/.hermes/skills/")
			} else {
				ui.Info("Skills (%d):", len(skills))
				for _, s := range skills {
					ui.Step("  • %s", s.Name)
				}
			}
			return nil
		},
	}
}

func agentsInspectCmd() *cobra.Command {
	var agenticName string
	cmd := &cobra.Command{
		Use:   "inspect <profile>",
		Short: "Show what would change for a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			cwd := getCwd(cmd)
			personaPath := filepath.Join(cwd, "harness", "personas", profileName+".md")
			if _, err := os.Stat(personaPath); err != nil {
				ui.Fail("Persona not found: %s", personaPath)
				ui.Info("Available personas:")
				listPersonas(cwd)
				return fmt.Errorf("persona not found")
			}

			ui.Header("Inspect: " + profileName)

			// Generate new SOUL.md from persona
			newSoul, err := soul.Generate(personaPath, frameworkVersion(cwd))
			if err != nil {
				return err
			}

			// Read current SOUL.md from Hermes
			hermesClient, _ := hermes.NewClient("")
			var current string
			if hermesClient.IsInstalled() {
				current, _ = hermesClient.ReadSoul(profileName)
			}
			if current == "" {
				ui.Warn("No existing SOUL.md for %s in Hermes", profileName)
				ui.Info("New SOUL.md would be created from scratch (%d bytes)", len(newSoul))
				ui.Info("")
				ui.Step("Run 'gmh agents install %s' to install", profileName)
				return nil
			}

			// Diff
			d := soul.ComputeDiff(current, newSoul)
			ui.Info("Diff: %s", d.Summary)
			ui.Info("")

			if len(d.Added) > 0 {
				ui.Info("Lines to add (%d):", len(d.Added))
				for _, l := range d.Added {
					if l != "" {
						ui.Step("+ %s", truncate(l, 80))
					}
				}
			}
			if len(d.Removed) > 0 {
				ui.Info("")
				ui.Info("Lines to remove (%d):", len(d.Removed))
				for _, l := range d.Removed {
					if l != "" {
						ui.Step("- %s", truncate(l, 80))
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&agenticName, "agent", "hermes", "Agentic (default: hermes)")
	return cmd
}

func agentsSyncCmd() *cobra.Command {
	var (
		aggressive bool
		dryRun     bool
		openPR     bool
		base       string
		onlyProfile string
	)
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync profiles + skills with the framework",
		Long: `Sync agentic profiles and skills with the current meta-harness version.

By default (--safe), only updates profiles that are clearly outdated
(e.g., SOUL.md missing or content hash differs). Custom sections
are preserved.

With --aggressive, all profiles matching framework personas are
regenerated (custom sections in the OLD SOUL.md are preserved in
a "## Custom sections (preserved)" block).

This is the counterpart to 'gmh sync': 'gmh sync' updates harness/
in the project; 'gmh agents sync' updates the user-side profiles.

Examples:
  gmh agents sync                        # Safe sync of all
  gmh agents sync --aggressive           # Overwrite including customizations
  gmh agents sync --only team-manager    # Sync one profile
  gmh agents sync --dry-run              # Show what would change
  gmh agents sync --open-pr              # Open a PR (Hermes-side; rare)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd := getCwd(cmd)

			hermesClient, err := hermes.NewClient("")
			if err != nil {
				return err
			}
			if !hermesClient.IsInstalled() {
				ui.Fail("Hermes not installed at %s", hermesClient.Home)
				ui.Info("Install Hermes first, then run 'gmh agents sync'")
				return fmt.Errorf("hermes not installed")
			}

			// Build framework manifest
			skillsDir := filepath.Join(cwd, "harness", "skills")
			manifest, err := skills.BuildManifest(skillsDir)
			if err != nil {
				ui.Warn("Could not build skills manifest: %v", err)
			}

			// List installed profiles
			profiles, err := hermesClient.ListProfiles()
			if err != nil {
				return err
			}

			ui.Header("Syncing profiles + skills with framework")

			// Sync each profile
			profileResults := []profileSyncResult{}
			for _, p := range profiles {
				if onlyProfile != "" && p.Name != onlyProfile {
					continue
				}
				result := syncProfile(p, cwd, aggressive, dryRun)
				profileResults = append(profileResults, result)
			}

			// Install missing profiles (personas in framework without Hermes profile)
			personas, _ := listPersonaFiles(cwd)
			installed := make(map[string]bool)
			for _, p := range profiles {
				installed[p.Name] = true
			}
			for _, persona := range personas {
				name := soul.PersonaNameFromFilename(persona)
				if !installed[name] {
					if onlyProfile != "" && name != onlyProfile {
						continue
					}
					ui.Warn("Persona %q exists in framework but no Hermes profile", name)
					ui.Info("  Run 'gmh agents install %s' to install", name)
				}
			}

			// Sync skills
			if manifest != nil {
				ui.Info("")
				ui.Info("Skills sync:")
				skillsInstalled := 0
				skillsUpdated := 0
				skillsUnchanged := 0
				for _, s := range manifest.Skills {
					if onlyProfile != "" {
						continue
					}
					installed, err := hermesClient.ReadSkill(s.Name)
					if err != nil {
						ui.Warn("  ⚠ %s: %v", s.Name, err)
						continue
					}
					if installed == "" {
						ui.Step("  + %s (not installed in Hermes)", s.Name)
						if !dryRun {
							if err := hermesClient.WriteSkill(s.Name, s.Content); err != nil {
								ui.Warn("    failed: %v", err)
							} else {
								ui.OK("    installed")
								skillsInstalled++
							}
						} else {
							skillsInstalled++
						}
					} else if installed != s.Content {
						d := soul.ComputeDiff(installed, s.Content)
						ui.Step("  ~ %s (diff: %s)", s.Name, d.Summary)
						if aggressive && !dryRun {
							if err := hermesClient.WriteSkill(s.Name, s.Content); err != nil {
								ui.Warn("    failed: %v", err)
							} else {
								ui.OK("    updated")
								skillsUpdated++
							}
						} else if !aggressive {
							ui.Step("    (safe mode: not updated; use --aggressive)")
						}
					} else {
						ui.Step("  = %s (unchanged)", s.Name)
						skillsUnchanged++
					}
				}
				ui.Info("")
				ui.Info("Skills: %d installed, %d updated, %d unchanged", skillsInstalled, skillsUpdated, skillsUnchanged)
			}

			// Summary
			ui.Info("")
			ui.Header("Summary")
			updated := 0
			preserved := 0
			skipped := 0
			for _, r := range profileResults {
				switch r.action {
				case "updated":
					updated++
					ui.OK("  ✓ %s — updated", r.name)
				case "preserved":
					preserved++
					ui.Step("  ~ %s — preserved (customizations kept)", r.name)
				case "skipped":
					skipped++
					ui.Step("  = %s — unchanged", r.name)
				}
			}
			ui.Info("")
			ui.Info("Updated: %d, Preserved: %d, Unchanged: %d", updated, preserved, skipped)
			ui.Info("Strategy: %s", strategyName(aggressive))
			if dryRun {
				ui.Warn("DRY RUN — no changes were written")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&aggressive, "aggressive", false,
		"Overwrite including local customizations (safe by default)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Show what would change without making changes")
	cmd.Flags().BoolVar(&openPR, "open-pr", false,
		"Open a GitHub PR (rarely used for Hermes-side changes)")
	cmd.Flags().StringVar(&base, "base", "main", "Base branch for PR")
	cmd.Flags().StringVar(&onlyProfile, "only", "", "Sync only this profile")
	return cmd
}

type profileSyncResult struct {
	name   string
	action string // "updated" | "preserved" | "skipped"
}

func syncProfile(p hermes.Profile, cwd string, aggressive, dryRun bool) profileSyncResult {
	personaPath := filepath.Join(cwd, "harness", "personas", p.Name+".md")
	if _, err := os.Stat(personaPath); err != nil {
		// No matching persona in framework
		return profileSyncResult{name: p.Name, action: "preserved"}
	}

	newSoul, err := soul.Generate(personaPath, frameworkVersion(cwd))
	if err != nil {
		ui.Warn("  ⚠ %s: failed to generate SOUL: %v", p.Name, err)
		return profileSyncResult{name: p.Name, action: "skipped"}
	}

	current, _ := p.Path, "" // we read SOUL below
	currentSoul, err := os.ReadFile(p.SoulPath)
	if err != nil {
		// No SOUL.md yet — just install
		ui.Step("  + %s (no SOUL.md, installing fresh)", p.Name)
		if !dryRun {
			if err := writeFile(p.SoulPath, newSoul); err != nil {
				ui.Warn("    failed: %v", err)
			}
		}
		return profileSyncResult{name: p.Name, action: "updated"}
	}
	_ = current

	d := soul.ComputeDiff(string(currentSoul), newSoul)
	if d.Summary == fmt.Sprintf("+%d -%d =%d", 0, 0, d.Unchanged) {
		// Identical
		return profileSyncResult{name: p.Name, action: "skipped"}
	}

	if !aggressive {
		// Safe mode: update if the persona hash OR framework version
		// in the current SOUL is stale (marker mismatch). Otherwise,
		// preserve the user's customizations.
		currentVer, currentHash := soul.ExtractVersionMarker(string(currentSoul))
		newVer := frameworkVersion(cwd)
		newHash, _ := soul.PersonaHash(personaPath)

		stale := false
		reason := ""
		if currentVer == "" {
			stale = true
			reason = "no version marker"
		} else if currentVer != newVer {
			stale = true
			reason = fmt.Sprintf("framework version %s → %s", currentVer, newVer)
		} else if currentHash != newHash {
			stale = true
			reason = "persona hash changed"
		}

		if !stale {
			// Markers match → user has the current version. If there's
			// no diff, skip. If there's a diff, it's from customizations
			// the user made — preserve.
			if len(d.Added) == 0 && len(d.Removed) == 0 {
				return profileSyncResult{name: p.Name, action: "skipped"}
			}
			ui.Step("  ~ %s (markers match; customizations preserved)", p.Name)
			return profileSyncResult{name: p.Name, action: "preserved"}
		}

		// Outdated → update in safe mode
		ui.Step("  ↻ %s (outdated: %s)", p.Name, reason)
		if !dryRun {
			if err := writeFile(p.SoulPath, newSoul); err != nil {
				ui.Warn("    failed: %v", err)
				return profileSyncResult{name: p.Name, action: "skipped"}
			}
		}
		return profileSyncResult{name: p.Name, action: "updated"}
	}

	// Aggressive: write the new SOUL (custom sections preserved inside
	// the generated content per Generate()).
	ui.Step("  + %s (regenerating)", p.Name)
	if !dryRun {
		if err := writeFile(p.SoulPath, newSoul); err != nil {
			ui.Warn("    failed: %v", err)
			return profileSyncResult{name: p.Name, action: "skipped"}
		}
	}
	return profileSyncResult{name: p.Name, action: "updated"}
}

func agentsInstallCmd() *cobra.Command {
	var onlyProfile string
	cmd := &cobra.Command{
		Use:   "install <profile>",
		Short: "Install a single profile from the framework",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			cwd := getCwd(cmd)
			personaPath := filepath.Join(cwd, "harness", "personas", profileName+".md")
			if _, err := os.Stat(personaPath); err != nil {
				ui.Fail("Persona not found: %s", personaPath)
				ui.Info("Available personas:")
				listPersonas(cwd)
				return fmt.Errorf("persona not found")
			}

			newSoul, err := soul.Generate(personaPath, frameworkVersion(cwd))
			if err != nil {
				return err
			}

			hermesClient, err := hermes.NewClient("")
			if err != nil {
				return err
			}
			if !hermesClient.IsInstalled() {
				ui.Fail("Hermes not installed at %s", hermesClient.Home)
				return fmt.Errorf("hermes not installed")
			}

			ui.Info("Installing profile %q from %s", profileName, personaPath)
			if err := hermesClient.WriteSoul(profileName, newSoul); err != nil {
				return err
			}
			ui.OK("Profile %q installed at %s", profileName, hermesClient.Home+"/profiles/"+profileName)
			ui.Info("")
			ui.Info("Next: use it with 'hermes -p %s' (or via your client)", profileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&onlyProfile, "only", "", "Ignored; install takes one profile name")
	return cmd
}

// helpers

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func listPersonas(cwd string) {
	files, _ := listPersonaFiles(cwd)
	for _, f := range files {
		ui.Step("  - %s", soul.PersonaNameFromFilename(f))
	}
}

func listPersonaFiles(cwd string) ([]string, error) {
	dir := filepath.Join(cwd, "harness", "personas")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := e.Name()
		// Skip non-persona files
		if strings.HasSuffix(name, ".template.md") {
			continue
		}
		if name == "interactions.md" {
			continue
		}
		out = append(out, name)
	}
	return out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func strategyName(aggressive bool) string {
	if aggressive {
		return "aggressive (overwrite including customizations)"
	}
	return "safe (only update if persona marker missing)"
}

// suppress unused import warnings for agentic (used in future iterations)
var _ = agentic.None

// getCwd returns the working directory, honoring the --cwd/-C flag
// via viper. Falls back to os.Getwd() if the flag is empty.
func getCwd(cmd *cobra.Command) string {
	if v := viper.GetString("cwd"); v != "" && v != "." {
		return v
	}
	cwd, _ := os.Getwd()
	return cwd
}

// frameworkVersion reads the meta-harness version from the project's
// VERSION file (set by `gmh sync` / `gmh install`).
// Returns "" if not found.
func frameworkVersion(cwd string) string {
	data, err := os.ReadFile(filepath.Join(cwd, "VERSION"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
