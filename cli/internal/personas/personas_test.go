package personas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "simple", in: "banking", want: "banking"},
		{name: "uppercase", in: "Retail", want: "retail"},
		{name: "spaces", in: "Open Banking", want: "open-banking"},
		{name: "underscores", in: "open_banking", want: "open-banking"},
		{name: "trim", in: "  healthcare  ", want: "healthcare"},
		{name: "accents", in: "domínio", want: "dominio"},
		{name: "collapse hyphens", in: "foo--bar", want: "foo-bar"},
		{name: "empty", in: "", wantErr: true},
		{name: "whitespace only", in: "   ", wantErr: true},
		{name: "generic", in: "generic", wantErr: true},
		{name: "GENERIC", in: "GENERIC", wantErr: true},
		{name: "domain-expert", in: "domain-expert", wantErr: true},
		{name: "Domain Expert", in: "Domain Expert", wantErr: true},
		{name: "invalid chars", in: "bank/ing", wantErr: true},
		{name: "punctuation", in: "banking!", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Slug(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Slug(%q) = %q, want error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Slug(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("Slug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCreateWritesPersona(t *testing.T) {
	harnessDir := t.TempDir()

	path, err := Create(harnessDir, "banking", "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	wantPath := filepath.Join(harnessDir, "personas", "domain-expert-banking.md")
	if path != wantPath {
		t.Fatalf("Create path = %q, want %q", path, wantPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read created file: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "banking") {
		t.Fatalf("created file missing domain name:\n%s", got)
	}
	if strings.Contains(got, "<domínio>") || strings.Contains(got, "<seu-dominio>") {
		t.Fatalf("placeholders were not replaced:\n%s", got)
	}
	if strings.Contains(got, "## Project context") {
		t.Fatalf("empty context should not add Project context section:\n%s", got)
	}
}

func TestCreateUsesTemplateAndContext(t *testing.T) {
	harnessDir := t.TempDir()
	personasDir := filepath.Join(harnessDir, "personas")
	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	tmpl := "# Persona — Domain Expert `<domínio>`\n\nCopy to domain-expert-<seu-dominio>.md\n"
	if err := os.WriteFile(filepath.Join(personasDir, "domain-expert.template.md"), []byte(tmpl), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	path, err := Create(harnessDir, "Retail", "Pix + Open Banking")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "Domain Expert `retail`") {
		t.Fatalf("template <domínio> not replaced:\n%s", got)
	}
	if !strings.Contains(got, "domain-expert-retail.md") {
		t.Fatalf("template <seu-dominio> not replaced:\n%s", got)
	}
	if !strings.Contains(got, "## Project context") {
		t.Fatalf("missing Project context heading:\n%s", got)
	}
	if !strings.Contains(got, "Pix + Open Banking") {
		t.Fatalf("missing project context body:\n%s", got)
	}
}

func TestCreateRefusesOverwrite(t *testing.T) {
	harnessDir := t.TempDir()
	if _, err := Create(harnessDir, "banking", ""); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if _, err := Create(harnessDir, "banking", "again"); err == nil {
		t.Fatal("second Create: want overwrite error")
	}
}

func TestList(t *testing.T) {
	t.Run("missing dir", func(t *testing.T) {
		got, err := List(t.TempDir())
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("List missing dir = %v, want empty", got)
		}
	})

	t.Run("lists basenames after create", func(t *testing.T) {
		harnessDir := t.TempDir()
		if _, err := Create(harnessDir, "banking", ""); err != nil {
			t.Fatalf("Create banking: %v", err)
		}
		if _, err := Create(harnessDir, "retail", ""); err != nil {
			t.Fatalf("Create retail: %v", err)
		}
		// non-md files must be ignored
		if err := os.WriteFile(filepath.Join(harnessDir, "personas", "notes.txt"), []byte("x"), 0o644); err != nil {
			t.Fatalf("write notes: %v", err)
		}

		got, err := List(harnessDir)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		want := []string{"domain-expert-banking.md", "domain-expert-retail.md"}
		if len(got) != len(want) {
			t.Fatalf("List = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("List = %v, want %v", got, want)
			}
		}
	})
}

func TestRemove(t *testing.T) {
	harnessDir := t.TempDir()
	path, err := Create(harnessDir, "banking", "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := Remove(harnessDir, "domain-expert-banking"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file still exists after Remove: %v", err)
	}

	names, err := List(harnessDir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("List after Remove = %v, want empty", names)
	}
}

func TestRemoveRefusesAdopterAndProtected(t *testing.T) {
	harnessDir := t.TempDir()
	personasDir := filepath.Join(harnessDir, "personas")
	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	protected := []string{
		"domain-expert-adopter.md",
		"domain-expert.template.md",
		"team-manager.md",
	}
	for _, name := range protected {
		if err := os.WriteFile(filepath.Join(personasDir, name), []byte("keep"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	tests := []struct {
		name string
		arg  string
	}{
		{name: "adopter", arg: "domain-expert-adopter"},
		{name: "adopter with ext", arg: "domain-expert-adopter.md"},
		{name: "template", arg: "domain-expert.template.md"},
		{name: "template stem", arg: "template"},
		{name: "non domain-expert", arg: "team-manager"},
		{name: "empty", arg: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Remove(harnessDir, tt.arg); err == nil {
				t.Fatalf("Remove(%q): want error", tt.arg)
			}
		})
	}

	for _, name := range protected {
		if _, err := os.Stat(filepath.Join(personasDir, name)); err != nil {
			t.Fatalf("protected %s was deleted: %v", name, err)
		}
	}
}
