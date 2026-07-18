// Package apply materializes a diff.Result from the remote harness
// into the local project, respecting --keep-local and --dry-run.
package apply

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/diff"
)

// Options controls how a diff is applied.
type Options struct {
	// LocalDir is the local harness/ directory.
	LocalDir string
	// RemoteDir is the freshly-downloaded remote harness/ directory.
	RemoteDir string
	// KeepLocal: if true, files that exist locally and are not
	// identical to the remote baseline are NOT overwritten. The
	// "locally modified" detection is done by comparing against
	// the local file as-is (any difference is a modification).
	//
	// In our case, since we don't have a "local baseline" snapshot,
	// we treat "local differs from remote" as "locally modified",
	// which means we PRESERVE the local version by default. This
	// is the safe default for `gmh sync`.
	KeepLocal bool
	// DryRun: if true, no files are written/deleted. Only logs
	// what would happen.
	DryRun bool
	// Force: if true, overwrite even locally-modified files
	// (skip KeepLocal).
	Force bool
	// Logger receives human-readable progress lines.
	Logger func(format string, args ...interface{})
}

// Result summarizes what was applied.
type Result struct {
	Added     int
	Modified  int
	Skipped   int
	Deleted   int
	Conflicts []string // paths where KeepLocal preserved the local file
}

// Apply materializes the diff into LocalDir.
func Apply(d *diff.Result, opts Options) (*Result, error) {
	if opts.Logger == nil {
		opts.Logger = func(format string, args ...interface{}) {}
	}

	res := &Result{}

	for _, c := range d.Changes {
		switch c.Type {
		case diff.Added:
			if err := copyFile(c.RemotePath, filepath.Join(opts.LocalDir, filepath.FromSlash(c.Path)), opts.DryRun, opts.Logger); err != nil {
				return res, err
			}
			res.Added++

		case diff.Modified:
			// "LocalMod" heuristic: any modification is treated as local-mod.
			// We can't distinguish "I edited this" from "framework changed this"
			// without a baseline. So we preserve the local file unless --force.
			overwrite := opts.Force
			if !opts.KeepLocal {
				overwrite = true
			}
			if !overwrite {
				opts.Logger("  ⚠  %s — locally modified, preserving (use --force to overwrite)", c.Path)
				res.Skipped++
				res.Conflicts = append(res.Conflicts, c.Path)
				continue
			}
			if err := copyFile(c.RemotePath, c.LocalPath, opts.DryRun, opts.Logger); err != nil {
				return res, err
			}
			res.Modified++

		case diff.Deleted:
			if opts.DryRun {
				opts.Logger("  - %s", c.Path)
				res.Deleted++
				continue
			}
			if err := os.Remove(c.LocalPath); err != nil && !os.IsNotExist(err) {
				return res, fmt.Errorf("remove %s: %w", c.LocalPath, err)
			}
			opts.Logger("  - %s", c.Path)
			res.Deleted++

		case diff.Unchanged:
			// No-op
		}
	}

	// Also copy top-level files that live next to harness/ (VERSION, etc.)
	// We don't track those in the diff (since they aren't in harness/),
	// but they should be in sync. For now we skip this — the user can
	// `gmh update` to do a full replace.
	return res, nil
}

func copyFile(src, dst string, dryRun bool, log func(string, ...interface{})) error {
	if dryRun {
		log("  + %s", relPath(dst))
		return nil
	}
	// Ensure parent dir
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	log("  + %s", relPath(dst))
	return nil
}

func relPath(p string) string {
	// Try to make it relative to cwd for nicer output
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, p); err == nil && !startsWith(rel, "..") {
			return rel
		}
	}
	return p
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
