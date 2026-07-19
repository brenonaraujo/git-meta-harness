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

// UpdateCmd creates the `gmh update` command.
//
// `gmh update` pins the project to a specific version of the framework.
// Unlike `gmh sync` (which always pulls the latest), `gmh update --to`
// is for version pinning (e.g., to lock CI to a specific version for
// reproducibility).
//
// If --to is not specified, behaves like `gmh sync` (latest).
func UpdateCmd() *cobra.Command {
	var (
		toVersion string
		force     bool
		dryRun    bool
		openPR    bool
		base      string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the meta-harness framework to a specific version",
		Long: `Update the local harness/ directory to a specific version of
the meta-harness framework.

If --to is not specified, this is equivalent to 'gmh sync' (latest).
Use --to to pin to a specific version (e.g., v1.5.0).

Examples:
  gmh update                       # Update to latest (= gmh sync)
  gmh update --to v1.5.0           # Pin to v1.5.0
  gmh update --to v1.4.0 --force   # Downgrade to v1.4.0 (destructive)
  gmh update --to v1.5.0 --open-pr # Open a PR with the pin`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			harnessDir := filepath.Join(cwd, "harness")

			if _, err := os.Stat(harnessDir); err != nil {
				ui.Fail("harness/ not found at %s", harnessDir)
				ui.Info("Run `gmh install` first")
				return fmt.Errorf("harness/ not found")
			}

			src := source.NewClient("")
			target, err := src.ResolveVersion(toVersion)
			if err != nil {
				return fmt.Errorf("resolve version: %w", err)
			}

			localVersion := readLocalVersion(cwd)
			if localVersion == target {
				ui.OK("Already on %s — nothing to do", target)
				return nil
			}

			// Detect downgrade
			if localVersion != "" && isOlder(target, localVersion) && !force {
				ui.Warn("Requested %s is older than current %s — use --force to downgrade", target, localVersion)
				return fmt.Errorf("refusing to downgrade without --force")
			}

			ui.Info("Updating %s → %s", orNA(localVersion), target)

			tmp, err := os.MkdirTemp("", "gmh-update-*")
			if err != nil {
				return err
			}
			if !dryRun {
				defer os.RemoveAll(tmp)
			}

			remoteHarness, err := src.DownloadTarball(target, tmp)
			if err != nil {
				return fmt.Errorf("download: %w", err)
			}

			d, err := diff.Compute(harnessDir, remoteHarness)
			if err != nil {
				return fmt.Errorf("diff: %w", err)
			}
			ui.Info("Diff: %s", d.Summary())

			// Bump VERSION file FIRST — before the no-changes early return.
			// The target version is the source of truth, regardless of whether
			// the harness/ directory content already matches.
			if !dryRun {
				if err := copyFile(
					filepath.Join(source.ExtractedRoot(remoteHarness), "VERSION"),
					filepath.Join(cwd, "VERSION"),
				); err != nil {
					ui.Warn("Could not update VERSION file: %v", err)
				}
			}

			if d.Added+d.Modified+d.Deleted == 0 {
				ui.OK("No changes needed (VERSION bumped to %s)", target)
				return nil
			}

			opts := apply.Options{
				LocalDir:  harnessDir,
				RemoteDir: remoteHarness,
				KeepLocal: true, // update keeps local unless --force
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
				return nil
			}

			ui.OK("Updated: +%d ~%d -%d (skipped: %d)", res.Added, res.Modified, res.Deleted, res.Skipped)
			if len(res.Conflicts) > 0 {
				ui.Warn("%d file(s) preserved (use --force to overwrite)", len(res.Conflicts))
			}

			if openPR {
				return openUpdatePR(cwd, target, localVersion, res, base)
			}

			ui.Info("")
			ui.Info("Next: review, commit, and optionally run `gmh doctor`")
			return nil
		},
	}

	cmd.Flags().StringVar(&toVersion, "to", "",
		"Target version (e.g., v1.5.0). Default: latest")
	cmd.Flags().BoolVarP(&force, "force", "f", false,
		"Allow downgrades and overwrite locally-modified files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"Show what would change without making changes")
	cmd.Flags().BoolVar(&openPR, "open-pr", false,
		"Open a GitHub PR with the changes")
	cmd.Flags().StringVar(&base, "base", "main",
		"Base branch for the PR (default: main)")

	return cmd
}

func openUpdatePR(cwd, target, from string, res *apply.Result, base string) error {
	if !gitutil.IsGitRepo() {
		return fmt.Errorf("not a git repo")
	}
	if !gitutil.IsClean() {
		return fmt.Errorf("working tree not clean")
	}
	branch := "chore/harness-update-" + stripV(target)
	if err := gitutil.CreateBranch(branch); err != nil {
		return err
	}
	if err := gitutil.AddAll(); err != nil {
		return err
	}
	commitMsg := fmt.Sprintf("chore: harness update to %s\n\nFrom %s via gmh update.\n+ %d ~ %d - %d files.",
		target, orNA(from), res.Added, res.Modified, res.Deleted)
	if err := gitutil.Commit(commitMsg); err != nil {
		return err
	}
	if err := gitutil.Push("origin", branch); err != nil {
		return err
	}
	title := "chore: harness update to " + target
	body := fmt.Sprintf(`## Harness update to %s (from %s)

%s

### Changes
- **+ %d** added, **~ %d** modified, **- %d** deleted
- **%d** preserved locally

%s`, target, orNA(from),
		fromMsg(target, from),
		res.Added, res.Modified, res.Deleted, len(res.Conflicts),
		nextStepsMsg(target))
	url, err := gitutil.CreatePR(title, body, base, branch)
	if err != nil {
		return err
	}
	ui.OK("PR opened: %s", url)
	return nil
}

func fromMsg(target, from string) string {
	if from == "" {
		return fmt.Sprintf("First install at %s.", target)
	}
	return fmt.Sprintf("Updated from %s to %s.", from, target)
}

func nextStepsMsg(version string) string {
	return fmt.Sprintf(`### After merge
1. Run ` + "`./harness/scripts/check-stack-versions.sh`" + ` to confirm.
2. Run ` + "`./harness/scripts/smoke-test.sh`" + ` to validate.
3. See ` + "`docs/CLI.md`" + ` for usage of the new version.`)
}

// isOlder returns true if a is semver-older than b.
// Both inputs are in the form "vX.Y.Z" or "X.Y.Z".
func isOlder(a, b string) bool {
	ap := semverTuple(a)
	bp := semverTuple(b)
	for i := 0; i < 3; i++ {
		if ap[i] < bp[i] {
			return true
		}
		if ap[i] > bp[i] {
			return false
		}
	}
	return false
}

func semverTuple(v string) [3]int {
	var t [3]int
	if len(v) > 0 && v[0] == 'v' {
		v = v[1:]
	}
	// Best-effort: parse "X.Y.Z" (ignore pre-release tags)
	for i, c := range v {
		if c == '.' {
			continue
		}
		if c < '0' || c > '9' {
			// Stop at first non-digit (e.g., "-rc.1")
			v = v[:i]
			break
		}
	}
	parts := []int{}
	current := 0
	hasDigit := false
	for _, c := range v {
		if c >= '0' && c <= '9' {
			current = current*10 + int(c-'0')
			hasDigit = true
		} else if c == '.' {
			parts = append(parts, current)
			current = 0
			hasDigit = false
		} else {
			break
		}
	}
	if hasDigit {
		parts = append(parts, current)
	}
	for i := 0; i < 3 && i < len(parts); i++ {
		t[i] = parts[i]
	}
	return t
}
