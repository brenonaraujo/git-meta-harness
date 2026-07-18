// Package harness provides read/write access to the local
// harness/ directory of a meta-harness project.
package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Harness represents a meta-harness project on disk.
type Harness struct {
	// Root is the absolute path to the project root (where
	// harness/ lives or will live).
	Root string
	// HarnessDir is the absolute path to the harness/ directory.
	HarnessDir string
	// Version is the version of the meta-harness framework
	// currently installed (read from harness/VERSION).
	Version string
}

// New constructs a Harness rooted at the given directory.
// It does NOT check that the directory exists.
func New(root string) *Harness {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	return &Harness{
		Root:       abs,
		HarnessDir: filepath.Join(abs, "harness"),
	}
}

// Exists returns true if the harness/ directory exists.
func (h *Harness) Exists() bool {
	info, err := os.Stat(h.HarnessDir)
	return err == nil && info.IsDir()
}

// ReadVersion reads the version from harness/VERSION.
// Returns empty string if VERSION does not exist.
func (h *Harness) ReadVersion() (string, error) {
	path := filepath.Join(h.HarnessDir, "..", "VERSION")
	if !h.Exists() {
		return "", fmt.Errorf("harness directory does not exist: %s", h.HarnessDir)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read VERSION: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// RequiredFiles lists the files that must exist in harness/
// for the meta-harness to be considered installed.
var RequiredFiles = []string{
	"harness/AGENTS.md",
	"harness/bootstrap.md",
	"harness/seed/meta-harness-seed.md",
	"harness/personas/team-manager.md",
	"harness/personas/domain-expert.template.md",
	"harness/sensors/00-static-analysis.md",
	"harness/sensors/09-verify-after-build.md",
	"harness/scripts/smoke-test.sh",
	"harness/scripts/check-stack-versions.sh",
}

// CheckRequiredFiles verifies all RequiredFiles are present.
// Returns a list of missing files (empty if all present).
func (h *Harness) CheckRequiredFiles() []string {
	var missing []string
	for _, rel := range RequiredFiles {
		path := filepath.Join(h.Root, rel)
		if _, err := os.Stat(path); err != nil {
			missing = append(missing, rel)
		}
	}
	return missing
}
