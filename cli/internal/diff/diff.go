// Package diff computes a file-level diff between a local harness/
// directory and a remote one (just downloaded from a tarball).
package diff

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ChangeType describes what happened to a file.
type ChangeType int

const (
	// Added: file exists in remote but not in local.
	Added ChangeType = iota
	// Modified: file exists in both, but content differs.
	Modified
	// Deleted: file exists in local but not in remote.
	Deleted
	// Unchanged: file exists in both, content identical.
	Unchanged
)

// Change represents a single file change.
type Change struct {
	// Path is the relative path inside harness/ (forward-slash separated).
	Path string
	// Type is the kind of change.
	Type ChangeType
	// LocalPath is the absolute path on disk (or "" if Deleted).
	LocalPath string
	// RemotePath is the absolute path on disk (or "" if Added).
	RemotePath string
	// LocalMod indicates whether the local file was modified vs the
	// framework baseline. Only meaningful when comparing to the
	// baseline (the previous remote version).
	LocalMod bool
}

// Result is the full diff between local and remote harness/.
type Result struct {
	// Changes is the list of file changes, sorted by Path.
	Changes []Change
	// Added, Modified, Deleted, Unchanged counts.
	Added, Modified, Deleted, Unchanged int
}

// Compute computes the diff between localDir and remoteDir.
// Both directories should be the "harness/" dirs (or any dirs).
func Compute(localDir, remoteDir string) (*Result, error) {
	localFiles, err := walkFiles(localDir)
	if err != nil {
		return nil, fmt.Errorf("walk local: %w", err)
	}
	remoteFiles, err := walkFiles(remoteDir)
	if err != nil {
		return nil, fmt.Errorf("walk remote: %w", err)
	}

	all := make(map[string]bool)
	for p := range localFiles {
		all[p] = true
	}
	for p := range remoteFiles {
		all[p] = true
	}

	paths := make([]string, 0, len(all))
	for p := range all {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	res := &Result{}
	for _, p := range paths {
		l := localFiles[p]
		r := remoteFiles[p]
		switch {
		case l == "" && r != "":
			res.Changes = append(res.Changes, Change{
				Path: p, Type: Added,
				RemotePath: filepath.Join(remoteDir, filepath.FromSlash(p)),
			})
			res.Added++
		case l != "" && r == "":
			res.Changes = append(res.Changes, Change{
				Path: p, Type: Deleted,
				LocalPath: filepath.Join(localDir, filepath.FromSlash(p)),
			})
			res.Deleted++
		case l != "" && r != "":
			if filesEqual(l, r) {
				res.Changes = append(res.Changes, Change{
					Path: p, Type: Unchanged,
					LocalPath: l, RemotePath: r,
				})
				res.Unchanged++
			} else {
				res.Changes = append(res.Changes, Change{
					Path: p, Type: Modified,
					LocalPath: l, RemotePath: r,
				})
				res.Modified++
			}
		}
	}
	return res, nil
}

// walkFiles returns a map of relative path → absolute path for all
// regular files under root. Paths are stored with forward slashes.
func walkFiles(root string) (map[string]string, error) {
	out := make(map[string]string)
	if root == "" {
		return out, nil
	}
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Skip the .git directory (just in case)
		if info.Name() == ".git" {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = path
		return nil
	})
	return out, err
}

// filesEqual returns true if two files have identical content.
func filesEqual(a, b string) bool {
	fa, err := os.Open(a)
	if err != nil {
		return false
	}
	defer fa.Close()
	fb, err := os.Open(b)
	if err != nil {
		return false
	}
	defer fb.Close()
	const bufSize = 32 * 1024
	bufA := make([]byte, bufSize)
	bufB := make([]byte, bufSize)
	for {
		nA, errA := io.ReadFull(fa, bufA)
		nB, errB := io.ReadFull(fb, bufB)
		if nA != nB {
			return false
		}
		if !equalBytes(bufA[:nA], bufB[:nB]) {
			return false
		}
		if errA == io.EOF && errB == io.EOF {
			return true
		}
		if errA == io.ErrUnexpectedEOF && errB == io.ErrUnexpectedEOF {
			return true
		}
		if errA != nil || errB != nil {
			return false
		}
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Summary returns a one-line summary of the diff.
func (r *Result) Summary() string {
	parts := []string{}
	if r.Added > 0 {
		parts = append(parts, fmt.Sprintf("+%d", r.Added))
	}
	if r.Modified > 0 {
		parts = append(parts, fmt.Sprintf("~%d", r.Modified))
	}
	if r.Deleted > 0 {
		parts = append(parts, fmt.Sprintf("-%d", r.Deleted))
	}
	if len(parts) == 0 {
		return "no changes"
	}
	return strings.Join(parts, " ") + fmt.Sprintf(" (%d unchanged)", r.Unchanged)
}
