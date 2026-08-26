package harnessmem

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestWriteReadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	want := Snapshot{
		SchemaVersion:    Schema,
		FrameworkVersion: "1.15.0",
		GeneratedAt:      "2026-08-26T15:04:05Z",
		Runtime:          "hermes",
		Project:          "demo",
		Personas:         []string{"team-manager.md"},
		Skills:           []string{"tdd-go.md"},
		HermesProfiles:   []string{"team-manager"},
		Notes:            DefaultNotes,
	}
	if err := Write(dir, want); err != nil {
		t.Fatalf("Write: %v", err)
	}
	path := filepath.Join(dir, "memory", "snapshot.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot.json missing: %v", err)
	}
	got, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.SchemaVersion != want.SchemaVersion {
		t.Errorf("schema_version: got %q want %q", got.SchemaVersion, want.SchemaVersion)
	}
	if got.FrameworkVersion != want.FrameworkVersion {
		t.Errorf("framework_version: got %q want %q", got.FrameworkVersion, want.FrameworkVersion)
	}
	if got.GeneratedAt != want.GeneratedAt {
		t.Errorf("generated_at: got %q want %q", got.GeneratedAt, want.GeneratedAt)
	}
	if got.Runtime != want.Runtime {
		t.Errorf("runtime: got %q want %q", got.Runtime, want.Runtime)
	}
	if got.Project != want.Project {
		t.Errorf("project: got %q want %q", got.Project, want.Project)
	}
	if got.Notes != want.Notes {
		t.Errorf("notes: got %q want %q", got.Notes, want.Notes)
	}
	if !slices.Equal(got.Personas, want.Personas) {
		t.Errorf("personas: got %v want %v", got.Personas, want.Personas)
	}
	if !slices.Equal(got.Skills, want.Skills) {
		t.Errorf("skills: got %v want %v", got.Skills, want.Skills)
	}
	if !slices.Equal(got.HermesProfiles, want.HermesProfiles) {
		t.Errorf("hermes_profiles: got %v want %v", got.HermesProfiles, want.HermesProfiles)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read raw: %v", err)
	}
	if !json.Valid(raw) {
		t.Fatalf("snapshot.json is not valid JSON")
	}
	if !strings.Contains(string(raw), "\n  \"schema_version\"") {
		t.Errorf("expected pretty-printed JSON, got:\n%s", raw)
	}
}

func TestBuildFindsPlantedPersona(t *testing.T) {
	harnessDir := t.TempDir()
	personasDir := filepath.Join(harnessDir, "personas")
	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		t.Fatalf("mkdir personas: %v", err)
	}
	planted := filepath.Join(personasDir, "team-manager.md")
	if err := os.WriteFile(planted, []byte("# team-manager\n"), 0o644); err != nil {
		t.Fatalf("plant persona: %v", err)
	}

	got, err := Build(harnessDir, "", "none", "dev", "demo")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !slices.Contains(got.Personas, "team-manager.md") {
		t.Errorf("personas = %v, want team-manager.md", got.Personas)
	}
	if got.SchemaVersion != Schema {
		t.Errorf("schema_version: got %q want %q", got.SchemaVersion, Schema)
	}
	if got.FrameworkVersion != "dev" {
		t.Errorf("framework_version: got %q want %q", got.FrameworkVersion, "dev")
	}
	if got.Runtime != "none" {
		t.Errorf("runtime: got %q want %q", got.Runtime, "none")
	}
	if got.Project != "demo" {
		t.Errorf("project: got %q want %q", got.Project, "demo")
	}
	if got.Notes != DefaultNotes {
		t.Errorf("notes: got %q want DefaultNotes", got.Notes)
	}
	if !strings.Contains(got.Notes, "Hermes is the OS/tool harness") {
		t.Errorf("notes missing required Hermes OS/tool harness sentence: %q", got.Notes)
	}
	if _, err := time.Parse(time.RFC3339, got.GeneratedAt); err != nil {
		t.Errorf("generated_at not RFC3339: %q (%v)", got.GeneratedAt, err)
	}
	if got.Skills == nil {
		t.Errorf("skills should be non-nil empty slice")
	}
	if got.HermesProfiles == nil {
		t.Errorf("hermes_profiles should be non-nil empty slice")
	}
}

func TestBuildListsSkillsAndHermesProfiles(t *testing.T) {
	harnessDir := t.TempDir()
	skillsDir := filepath.Join(harnessDir, "skills")
	if err := os.MkdirAll(filepath.Join(skillsDir, "code-graph"), 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "tdd-go.md"), []byte("# tdd\n"), 0o644); err != nil {
		t.Fatalf("plant skill md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "README.txt"), []byte("skip me\n"), 0o644); err != nil {
		t.Fatalf("plant non-skill: %v", err)
	}

	hermesHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(hermesHome, "profiles", "team-manager"), 0o755); err != nil {
		t.Fatalf("mkdir profile: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(hermesHome, "profiles", "backend-engineer"), 0o755); err != nil {
		t.Fatalf("mkdir profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hermesHome, "profiles", "not-a-profile.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("plant file in profiles: %v", err)
	}

	got, err := Build(harnessDir, hermesHome, "hermes", "1.15.0", "acme")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !slices.Contains(got.Skills, "code-graph") {
		t.Errorf("skills = %v, want code-graph dir", got.Skills)
	}
	if !slices.Contains(got.Skills, "tdd-go.md") {
		t.Errorf("skills = %v, want tdd-go.md", got.Skills)
	}
	if slices.Contains(got.Skills, "README.txt") {
		t.Errorf("skills should skip non-.md files, got %v", got.Skills)
	}
	if !slices.Contains(got.HermesProfiles, "team-manager") {
		t.Errorf("hermes_profiles = %v, want team-manager", got.HermesProfiles)
	}
	if !slices.Contains(got.HermesProfiles, "backend-engineer") {
		t.Errorf("hermes_profiles = %v, want backend-engineer", got.HermesProfiles)
	}
	if slices.Contains(got.HermesProfiles, "not-a-profile.md") {
		t.Errorf("hermes_profiles should skip files, got %v", got.HermesProfiles)
	}
}

func TestBuildSkipsMissingSkillsDir(t *testing.T) {
	harnessDir := t.TempDir()
	got, err := Build(harnessDir, "", "none", "dev", "demo")
	if err != nil {
		t.Fatalf("Build with missing skills: %v", err)
	}
	if len(got.Skills) != 0 {
		t.Errorf("skills: got %v want empty", got.Skills)
	}
	if len(got.Personas) != 0 {
		t.Errorf("personas: got %v want empty", got.Personas)
	}
}

func TestReadMissingFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := Read(dir); err == nil {
		t.Fatal("Read missing snapshot: want error")
	}
}

func TestWriteNilSlicesMarshalAsEmptyArrays(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, Snapshot{SchemaVersion: Schema}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "memory", "snapshot.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	body := string(raw)
	for _, key := range []string{"personas", "skills", "hermes_profiles"} {
		if strings.Contains(body, `"`+key+`": null`) {
			t.Errorf("%s marshaled as null; want []\n%s", key, body)
		}
	}
}
