package hermes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteConfigPreservesModelAgent is a regression test for the
// v1.12.1 hotfix. The previous WriteConfig implementation erased
// `model`, `agent`, and any other field not in the typed
// ProfileConfig struct whenever `gmh agents sync` ran.
//
// This test verifies that WriteConfig / EnsureExternalDirs update
// only `skills.external_dirs` and preserve every other field
// (including unknown custom keys).
func TestWriteConfigPreservesModelAgent(t *testing.T) {
	// Set up a temp Hermes home.
	tmp := t.TempDir()
	profile := "frontend-engineer"
	profileDir := filepath.Join(tmp, "profiles", profile)
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// User-configured config.yaml with model, agent, and a custom key.
	original := `agent:
  reasoning_effort: max
model:
  default: MiniMax-M3
  provider: minimax-oauth
skills:
  external_dirs:
  - ~/.hermes/skills
  - /Users/araujo/.hermes/skills
custom_user_key: keep-me
`
	cfgPath := filepath.Join(profileDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(original), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	c, err := NewClient(tmp)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Add a new external_dir.
	merged, err := c.EnsureExternalDirs(profile, []string{"/opt/new-skills"})
	if err != nil {
		t.Fatalf("EnsureExternalDirs: %v", err)
	}

	// Read back the file and check every important field.
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	got := string(data)

	// model
	if !strings.Contains(got, "default: MiniMax-M3") {
		t.Errorf("model.default was erased!\n---\n%s\n---", got)
	}
	if !strings.Contains(got, "provider: minimax-oauth") {
		t.Errorf("model.provider was erased!\n---\n%s\n---", got)
	}
	// agent
	if !strings.Contains(got, "reasoning_effort: max") {
		t.Errorf("agent.reasoning_effort was erased!\n---\n%s\n---", got)
	}
	// custom user key
	if !strings.Contains(got, "custom_user_key: keep-me") {
		t.Errorf("custom user key was erased!\n---\n%s\n---", got)
	}
	// skills.external_dirs should still have the original entries
	if !strings.Contains(got, "~/.hermes/skills") {
		t.Errorf("~/.hermes/skills was lost from external_dirs!\n---\n%s\n---", got)
	}
	if !strings.Contains(got, "/Users/araujo/.hermes/skills") {
		t.Errorf("/Users/araujo/.hermes/skills was lost from external_dirs!\n---\n%s\n---", got)
	}
	// ...and the new one
	if !strings.Contains(got, "/opt/new-skills") {
		t.Errorf("new dir was not added to external_dirs!\n---\n%s\n---", got)
	}
	// merged result should reflect the new dir
	found := false
	for _, d := range merged {
		if d == "/opt/new-skills" {
			found = true
		}
	}
	if !found {
		t.Errorf("EnsureExternalDirs did not return the merged list (got: %v)", merged)
	}
}

// TestWriteConfigCreatesWhenMissing verifies that WriteConfig can
// create a fresh config.yaml when none exists, without breaking
// any other code path.
func TestWriteConfigCreatesWhenMissing(t *testing.T) {
	tmp := t.TempDir()
	profile := "new-profile"
	profileDir := filepath.Join(tmp, "profiles", profile)
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	c, err := NewClient(tmp)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// WriteConfig on a non-existent config.yaml.
	if err := c.WriteConfig(profile, []string{"~/.hermes/skills"}); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	// Verify file exists and contains the dir.
	cfgPath := filepath.Join(profileDir, "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "~/.hermes/skills") {
		t.Errorf("WriteConfig did not write the dir: %s", got)
	}
}

// TestWriteConfigIdempotent ensures that running WriteConfig twice
// with the same input does not duplicate the external_dirs or
// rewrite the file unnecessarily.
func TestWriteConfigIdempotent(t *testing.T) {
	tmp := t.TempDir()
	profile := "idempotent"
	profileDir := filepath.Join(tmp, "profiles", profile)
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	c, err := NewClient(tmp)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	dirs := []string{"~/.hermes/skills", "/opt/x"}
	if err := c.WriteConfig(profile, dirs); err != nil {
		t.Fatalf("WriteConfig 1: %v", err)
	}
	if err := c.WriteConfig(profile, dirs); err != nil {
		t.Fatalf("WriteConfig 2: %v", err)
	}

	cfgPath := filepath.Join(profileDir, "config.yaml")
	data, _ := os.ReadFile(cfgPath)
	got := string(data)

	// Count occurrences of each dir (should be 1 each, not 2).
	for _, d := range dirs {
		count := strings.Count(got, d)
		if count != 1 {
			t.Errorf("dir %q appears %d times (expected 1):\n%s", d, count, got)
		}
	}
}
