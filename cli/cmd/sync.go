package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/apply"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/diff"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/gitutil"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/source"
	"github.com/brenonaraujo/git-meta-harness/cli/internal/ui"
)

// SyncCmd creates the `gmh sync` command.
//
// `gmh sync` updates the local project's `harness/` directory with
// the latest version from the meta-harness repo, while preserving
// any local customizations (e.g., materialized personas).
//
// With --open-pr, it creates a branch and a GitHub PR with the
// changes (useful for CI updates, new sensors, etc.).
func SyncCmd() *cobra.Command {
	var (
		dryRun    bool
		keepLocal bool
		force     bool
		openPR    bool
		base      string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync the local project with the latest meta-harness version",
		Long: `Update the local harness/ directory with the latest version
of the meta-harness framework while preserving local customizations.

By default, files that differ from the remote (i.e., locally modified
files) are PRESERVED. Use --force to overwrite them.

Examples:
  gmh sync                       # Pull latest, preserve local
  gmh sync --dry-run             # Show what would change
  gmh sync --force               # Overwrite locally-modified files
  gmh sync --open-pr             # Open a PR with the changes
  gmh sync --open-pr --base main # Open PR targeting main`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			harnessDir := filepath.Join(cwd, "harness")

			if _, err := os.Stat(harnessDir); err != nil {
				ui.Fail("harness/ not found at %s", harnessDir)
				ui.Info("Run `gmh install` first to install the framework")
				return fmt.Errorf("harness/ not found")
			}

			src := source.NewClient("")
			latest, err := src.ResolveVersion("latest")
			if err != nil {
				return fmt.Errorf("resolve latest: %w", err)
			}

			// Read local version
			localVersion := readLocalVersion(cwd)

			if localVersion == latest {
				ui.OK("Already on %s — no sync needed", localVersion)
				return nil
			}
			ui.Info("Syncing %s → %s", orNA(localVersion), latest)

			// Download latest to temp
			tmp, err := os.MkdirTemp("", "gmh-sync-*")
			if err != nil {
				return err
			}
			if !dryRun {
				defer os.RemoveAll(tmp)
			}

			remoteHarness, err := src.DownloadTarball(latest, tmp)
			if err != nil {
				return fmt.Errorf("download: %w", err)
			}

			// Diff
			d, err := diff.Compute(harnessDir, remoteHarness)
			if err != nil {
				return fmt.Errorf("diff: %w", err)
			}
			ui.Info("Diff: %s", d.Summary())

			if d.Added+d.Modified+d.Deleted == 0 {
				ui.OK("Already up to date — no changes needed")
				return nil
			}

			// Apply
			opts := apply.Options{
				LocalDir:  harnessDir,
				RemoteDir: remoteHarness,
				KeepLocal: keepLocal,
				DryRun:    dryRun,
				Force:     force,
				Logger:    ui.Step,
			}
			res, err := apply.Apply(d, opts)
			if err != nil {
				return fmt.Errorf("apply: %w", err)
			}

			if dryRun {
				ui.Info("Dry run — no changes applied")
				ui.Info("Run without --dry-run to apply")
				return nil
			}

			// Update VERSION at project root
			_ = copyFile(
				filepath.Join(source.ExtractedRoot(remoteHarness), "VERSION"),
				filepath.Join(cwd, "VERSION"),
			)

			ui.OK("Synced: +%d ~%d -%d (skipped: %d)", res.Added, res.Modified, res.Deleted, res.Skipped)
			if len(res.Conflicts) > 0 {
				ui.Warn("%d file(s) preserved (locally modified):", len(res.Conflicts))
				for _, c := range res.Conflicts {
					ui.Step("  %s", c)
				}
				ui.Info("Use --force to overwrite")
			}

			// Open PR if asked
			if openPR {
				return openSyncPR(cwd, latest, res, base)
			}

			ui.Info("")
			ui.Info("Next: review changes, commit, and (optionally) run `gmh doctor`")
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Show what would change without making changes")
	cmd.Flags().BoolVar(&keepLocal, "keep-local", true,
		"Preserve locally-modified files (default true)")
	cmd.Flags().BoolVarP(&force, "force", "f", false,
		"Overwrite locally-modified files")
	cmd.Flags().BoolVar(&openPR, "open-pr", false,
		"Open a GitHub PR with the changes")
	cmd.Flags().StringVar(&base, "base", "main",
		"Base branch for the PR (default: main)")

	return cmd
}

// openSyncPR creates a branch + commits + opens a PR for the sync.
func openSyncPR(cwd, latest string, res *apply.Result, base string) error {
	if !gitutil.IsGitRepo() {
		ui.Fail("Not a git repo — cannot open a PR")
		return fmt.Errorf("not a git repo")
	}
	if !gitutil.IsClean() {
		ui.Fail("Working tree is not clean — commit or stash your changes first")
		return fmt.Errorf("uncommitted changes")
	}
	if !gitutil.HasRemote() {
		ui.Fail("No 'origin' remote — cannot push")
		return fmt.Errorf("no origin remote")
	}
	if !gitutil.GhAvailable() {
		ui.Fail("gh CLI not available or not authenticated")
		return fmt.Errorf("gh not available")
	}

	branch := "chore/harness-sync-" + stripV(latest)
	ui.Info("Creating branch %s", branch)
	if err := gitutil.CreateBranch(branch); err != nil {
		return err
	}

	if err := gitutil.AddAll(); err != nil {
		return err
	}
	commitMsg := fmt.Sprintf("chore: harness sync to %s\n\nSynced via gmh sync.\n+ %d ~ %d - %d files.",
		latest, res.Added, res.Modified, res.Deleted)
	if err := gitutil.Commit(commitMsg); err != nil {
		return err
	}

	ui.Info("Pushing to origin/%s", branch)
	if err := gitutil.Push("origin", branch); err != nil {
		return err
	}

	title := "chore: harness sync to " + latest
	body := buildSyncPRBody(latest, res)

	url, err := gitutil.CreatePR(title, body, base, branch)
	if err != nil {
		return err
	}
	ui.OK("PR opened: %s", url)
	return nil
}

// buildSyncPRBody builds the PR body for a sync. Extracted to keep
// the RunE function readable.
func buildSyncPRBody(latest string, res *apply.Result) string {
	bt := "`" // backtick for inline code
	conflicts := len(res.Conflicts)
	return fmt.Sprintf(`## Harness sync to %s

Synced via %sgmh sync --open-pr%s.

### Changes
- **+%d** added
- **~%d** modified
- **-%d** deleted
- **%d** preserved (locally modified; use %sgmh sync --force%s to overwrite)

### What to review
1. **CI** — new sensors/scripts may require updates to %s.github/workflows/ci.yml%s.
2. **check-stack-versions.sh** — re-run to confirm.
3. **ADRs** — new ADRs (0014, 0015, 0016) added in this release.

### After merge
- %s+%s files: usually safe to accept
- %s~%s files: review — these may have framework improvements (e.g., new sensor)
- %s-%s files: usually safe to remove (framework removed them)

See %sdocs/HOWTO.md%s and %sharness/stack/versions.md%s for details.`,
		latest, bt, bt,
		res.Added, res.Modified, res.Deleted, conflicts, bt, bt,
		bt, bt,
		bt, bt, bt, bt,
		bt, bt, bt, bt,
		bt, bt)
}

func readLocalVersion(cwd string) string {
	data, err := os.ReadFile(filepath.Join(cwd, "VERSION"))
	if err != nil {
		return ""
	}
	return string(data)
}

func orNA(s string) string {
	if s == "" {
		return "_(unknown)_"
	}
	return s
}
