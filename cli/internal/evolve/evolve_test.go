package evolve

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestLoadComments_ArrayFile(t *testing.T) {
	got, err := LoadComments("testdata")
	if err != nil {
		t.Fatalf("LoadComments: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d comments, want 2", len(got))
	}
	if got[0].Issue != 42 || got[0].Author != "alice" {
		t.Errorf("first comment = %+v", got[0])
	}
	if !strings.Contains(got[0].Body, "acceptance") {
		t.Errorf("fixture body missing acceptance: %q", got[0].Body)
	}
	if got[1].Author != "bob" {
		t.Errorf("second author = %q, want bob", got[1].Author)
	}
}

func TestLoadComments_PerFile(t *testing.T) {
	dir := t.TempDir()
	one := Comment{Issue: 7, Author: "cara", Body: "missing AC", CreatedAt: "2026-01-01T00:00:00Z"}
	raw, _ := json.Marshal(one)
	if err := os.WriteFile(filepath.Join(dir, "issue-7.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadComments(dir)
	if err != nil {
		t.Fatalf("LoadComments: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d comments, want 1", len(got))
	}
	if got[0] != one {
		t.Errorf("got %+v, want %+v", got[0], one)
	}
}

func TestPropose_QuotesSnippet(t *testing.T) {
	comments, err := LoadComments("testdata")
	if err != nil {
		t.Fatalf("LoadComments: %v", err)
	}
	got := Propose(comments, []string{"team-manager", "domain-expert-banking"})

	if !strings.Contains(got, "Persona/skill evolve proposal") {
		t.Errorf("missing title:\n%s", got)
	}
	if !strings.Contains(got, "team-manager") || !strings.Contains(got, "domain-expert-banking") {
		t.Errorf("missing personas in play:\n%s", got)
	}
	if !strings.Contains(got, "acceptance criteria have a gap") {
		t.Errorf("missing quoted snippet:\n%s", got)
	}
	if strings.Contains(got, "Looks good to me") {
		t.Errorf("non-signal comment should not be quoted:\n%s", got)
	}
	if !strings.Contains(got, "## Hermes (OS/tool harness)") {
		t.Errorf("missing Hermes section:\n%s", got)
	}
	if !strings.Contains(got, "## PROMPT") {
		t.Errorf("missing PROMPT block:\n%s", got)
	}
}

func TestPropose_AlwaysHasHermesAndPrompt(t *testing.T) {
	got := Propose(nil, nil)
	if !strings.Contains(got, "## Hermes (OS/tool harness)") {
		t.Errorf("missing Hermes section:\n%s", got)
	}
	if !strings.Contains(got, "team-manager should run this proposal") {
		t.Errorf("Hermes section missing team-manager instruction:\n%s", got)
	}
	if !strings.Contains(got, "git-meta-harness only stores the delivery context") {
		t.Errorf("Hermes section missing delivery-context note:\n%s", got)
	}
	if !strings.Contains(got, "## PROMPT") {
		t.Errorf("missing PROMPT block:\n%s", got)
	}
}

func TestPropose_SnippetAtMost200(t *testing.T) {
	long := strings.Repeat("gap ", 80) // 320 runes of "gap "
	got := Propose([]Comment{{Issue: 1, Author: "x", Body: long}}, nil)
	// Every quoted line after "> " must be <= 200 chars.
	for _, line := range strings.Split(got, "\n") {
		if !strings.HasPrefix(line, "> ") {
			continue
		}
		snip := strings.TrimPrefix(line, "> ")
		if len(snip) > 200 {
			t.Errorf("snippet len %d > 200: %q", len(snip), snip)
		}
	}
}

func TestWriteTrace_CreatesThreeFiles(t *testing.T) {
	comments, err := LoadComments("testdata")
	if err != nil {
		t.Fatalf("LoadComments: %v", err)
	}
	proposal := Propose(comments, []string{"domain-expert-banking"})

	mem := t.TempDir()
	traceDir, err := WriteTrace(mem, comments, proposal)
	if err != nil {
		t.Fatalf("WriteTrace: %v", err)
	}

	wantDir := filepath.Join(mem, "traces")
	if !strings.HasPrefix(traceDir, wantDir+string(filepath.Separator)) {
		t.Errorf("traceDir = %q, want under %q", traceDir, wantDir)
	}
	stamp := filepath.Base(traceDir)
	if !regexp.MustCompile(`^\d{8}T\d{6}Z$`).MatchString(stamp) {
		t.Errorf("trace stamp %q is not compact UTC (20060102T150405Z)", stamp)
	}

	for _, name := range []string{"comments.json", "proposal.md", "PROMPT.md"} {
		path := filepath.Join(traceDir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("missing %s: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", name)
		}
	}

	raw, err := os.ReadFile(filepath.Join(traceDir, "comments.json"))
	if err != nil {
		t.Fatalf("read comments.json: %v", err)
	}
	var got []Comment
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("comments.json: %v", err)
	}
	if len(got) != len(comments) {
		t.Errorf("comments.json has %d entries, want %d", len(got), len(comments))
	}

	prompt, err := os.ReadFile(filepath.Join(traceDir, "PROMPT.md"))
	if err != nil {
		t.Fatalf("read PROMPT.md: %v", err)
	}
	if !strings.Contains(string(prompt), "## PROMPT") {
		t.Errorf("PROMPT.md missing prompt section:\n%s", prompt)
	}

	prop, err := os.ReadFile(filepath.Join(traceDir, "proposal.md"))
	if err != nil {
		t.Fatalf("read proposal.md: %v", err)
	}
	if !strings.Contains(string(prop), "acceptance criteria have a gap") {
		t.Errorf("proposal.md missing quoted snippet")
	}
}
