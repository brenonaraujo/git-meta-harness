// Package harnessmem persists generated delivery-harness memory
// snapshots (personas, skills, project-specific agent context).
package harnessmem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Schema is the snapshot JSON schema identifier.
const Schema = "git-meta-harness/memory/v1"

// DefaultNotes distinguishes Hermes (OS/tool harness) from this
// snapshot (delivery-harness memory produced by git-meta-harness).
const DefaultNotes = "Hermes is the OS/tool harness (terminal, fs, gh, browsers). This snapshot is the delivery-harness memory (personas, skills, project-specific agent context) produced by git-meta-harness."

const snapshotRel = "memory/snapshot.json"

// Snapshot is the on-disk delivery-harness memory document.
type Snapshot struct {
	SchemaVersion    string   `json:"schema_version"` // git-meta-harness/memory/v1
	FrameworkVersion string   `json:"framework_version"`
	GeneratedAt      string   `json:"generated_at"` // RFC3339 UTC
	Runtime          string   `json:"runtime"`      // hermes | claude-code | none
	Project          string   `json:"project"`
	Personas         []string `json:"personas"`
	Skills           []string `json:"skills"`
	HermesProfiles   []string `json:"hermes_profiles"`
	Notes            string   `json:"notes"`
}

// Write writes s as pretty JSON to harnessDir/memory/snapshot.json.
func Write(harnessDir string, s Snapshot) error {
	s.Personas = nonNil(s.Personas)
	s.Skills = nonNil(s.Skills)
	s.HermesProfiles = nonNil(s.HermesProfiles)

	dir := filepath.Join(harnessDir, "memory")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	data = append(data, '\n')
	path := filepath.Join(harnessDir, snapshotRel)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Read loads harnessDir/memory/snapshot.json.
func Read(harnessDir string) (*Snapshot, error) {
	path := filepath.Join(harnessDir, snapshotRel)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}
	s.Personas = nonNil(s.Personas)
	s.Skills = nonNil(s.Skills)
	s.HermesProfiles = nonNil(s.HermesProfiles)
	return &s, nil
}

// Build scans harnessDir (personas + skills) and optional hermesHome
// profiles into a Snapshot. Missing personas/skills/profiles dirs are
// skipped.
func Build(harnessDir, hermesHome, runtime, version, project string) (Snapshot, error) {
	personas, err := listMarkdownNames(filepath.Join(harnessDir, "personas"))
	if err != nil {
		return Snapshot{}, err
	}
	skills, err := listSkillNames(filepath.Join(harnessDir, "skills"))
	if err != nil {
		return Snapshot{}, err
	}
	profiles := []string{}
	if hermesHome != "" {
		profiles, err = listDirNames(filepath.Join(hermesHome, "profiles"))
		if err != nil {
			return Snapshot{}, err
		}
	}
	return Snapshot{
		SchemaVersion:    Schema,
		FrameworkVersion: version,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		Runtime:          runtime,
		Project:          project,
		Personas:         nonNil(personas),
		Skills:           nonNil(skills),
		HermesProfiles:   nonNil(profiles),
		Notes:            DefaultNotes,
	}, nil
}

func listMarkdownNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return nonNil(names), nil
}

func listSkillNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if e.IsDir() || strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return nonNil(names), nil
}

func listDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return nonNil(names), nil
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
