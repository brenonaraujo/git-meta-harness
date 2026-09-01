package seed

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brenonaraujo/git-meta-harness/cli/internal/loop"
)

func TestPlan_PersonasAndIssue0(t *testing.T) {
	p, err := Plan(Options{
		Name:     "booking-saas",
		Describe: "SaaS de agendamento online para clinicas. Assinatura mensal, horarios, pacientes.",
		Domain:   "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Slug != "booking-saas" {
		t.Fatalf("slug=%q", p.Slug)
	}
	if p.Domain == "" || p.Domain == "domain-expert" {
		t.Fatalf("domain=%q", p.Domain)
	}
	foundDE := false
	for _, n := range p.Personas {
		if n == "domain-expert" {
			t.Fatal("generic domain-expert")
		}
		if strings.HasPrefix(n, "domain-expert-") {
			foundDE = true
		}
	}
	if !foundDE {
		t.Fatalf("missing specialized domain-expert: %v", p.Personas)
	}
	if !strings.Contains(strings.ToLower(p.Issue0Title), "spec") && !strings.Contains(strings.ToLower(p.Issue0Title), "bootstrap") {
		t.Fatalf("issue0 title=%q", p.Issue0Title)
	}
	if len(p.CronJobs) != 2 {
		t.Fatalf("cron jobs=%d", len(p.CronJobs))
	}
	if p.CronJobs[0].Monitor || !p.CronJobs[0].NoAgent {
		t.Fatalf("tm job: %+v", p.CronJobs[0])
	}
}

func TestFindHarnessSrc_FrameworkRepo(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "harness", "bootstrap.md"), "# bootstrap\n")
	mustWrite(t, filepath.Join(dir, "cli", "go.mod"), "module github.com/brenonaraujo/git-meta-harness/cli\n")
	got := FindHarnessSrc("", dir)
	want := filepath.Join(dir, "harness")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFindHarnessSrc_ExplicitWins(t *testing.T) {
	got := FindHarnessSrc("/explicit/harness", "/other")
	if got != "/explicit/harness" {
		t.Fatalf("got %q", got)
	}
}

func TestApply_WritesSpecLoopAndContractsCI(t *testing.T) {
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "bootstrap.md"), "# bootstrap\n")
	mustWrite(t, filepath.Join(src, "personas", "team-manager.md"), "Você não implementa código de feature.\n")
	mustWrite(t, filepath.Join(src, "templates", ".github-workflows-ci.yml"), "name: CI\n# go+nuxt full\n")
	mustWrite(t, filepath.Join(src, "templates", ".github-workflows-ci-contracts.yml"), "name: Contracts\n")
	mustWrite(t, filepath.Join(src, "templates", ".github-workflows-release.yml"), "name: release\n")
	mustWrite(t, filepath.Join(src, "templates", "pr-description.md"), "# PR\n")

	dst := t.TempDir()
	res, err := Apply(Options{
		Name:       "booking-saas",
		Describe:   "SaaS de agendamento online",
		Domain:     "saas",
		TargetDir:  dst,
		HarnessSrc: src,
		Stack:      "vue,swarm",
		SkipAgents: true,
		SkipGitHub: true,
		SkipCron:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.CIKind != "contracts" {
		t.Fatalf("CIKind=%q", res.CIKind)
	}
	mustExist(t, filepath.Join(dst, "docs", "SPEC.md"))
	mustExist(t, filepath.Join(dst, "harness", "bootstrap.md"))
	mustExist(t, filepath.Join(dst, "harness", "loop.env"))
	mustExist(t, filepath.Join(dst, "harness", "scripts", "loop", "spawn-tm.sh"))
	mustExist(t, filepath.Join(dst, "harness", "PROJECT.md"))
	mustExist(t, filepath.Join(dst, ".github", "workflows", "ci.yml"))
	mustExist(t, filepath.Join(dst, ".github", "ISSUE_0.md"))
	ci, _ := os.ReadFile(filepath.Join(dst, ".github", "workflows", "ci.yml"))
	if !strings.Contains(string(ci), "Contracts") {
		t.Fatalf("expected contracts CI, got:\n%s", ci)
	}
	rep := loop.Doctor(dst)
	if !rep.Pass {
		t.Fatalf("loop doctor failed: %v", rep.Failed)
	}
}

func TestApply_FullCIForGoNuxt(t *testing.T) {
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "bootstrap.md"), "# b\n")
	mustWrite(t, filepath.Join(src, "personas", "team-manager.md"), "Você não implementa código de feature.\n")
	mustWrite(t, filepath.Join(src, "templates", ".github-workflows-ci.yml"), "name: CI\ndorny/paths-filter\n")
	mustWrite(t, filepath.Join(src, "templates", ".github-workflows-ci-contracts.yml"), "name: Contracts\n")
	mustWrite(t, filepath.Join(src, "templates", ".github-workflows-release.yml"), "name: release\n")

	dst := t.TempDir()
	res, err := Apply(Options{
		Name:       "app",
		Describe:   "Go API",
		Domain:     "internal",
		TargetDir:  dst,
		HarnessSrc: src,
		Stack:      "go,nuxt,postgresql",
		SkipAgents: true,
		SkipGitHub: true,
		SkipCron:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.CIKind != "full" {
		t.Fatalf("CIKind=%q", res.CIKind)
	}
	ci, _ := os.ReadFile(filepath.Join(dst, ".github", "workflows", "ci.yml"))
	if !strings.Contains(string(ci), "dorny/paths-filter") {
		t.Fatalf("expected full CI, got:\n%s", ci)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("missing %s: %v", path, err)
	}
}
