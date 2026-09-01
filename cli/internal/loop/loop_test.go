package loop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonicalPersonas_SpecializedDomainExpert(t *testing.T) {
	got := CanonicalPersonas("saas")
	want := []string{
		"team-manager",
		"domain-expert-saas",
		"solutions-architect",
		"backend-engineer",
		"frontend-engineer",
		"quality-assurance",
		"devops-engineer",
	}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("personas[%d]=%q want %q", i, got[i], want[i])
		}
	}
	for _, p := range got {
		if p == "domain-expert" {
			t.Fatal("generic domain-expert is forbidden")
		}
	}
}

func TestCronCommands_TMIsNoAgentWithoutMonitor(t *testing.T) {
	cfg := Config{
		ProjectRoot: "/tmp/booking-saas",
		ProjectSlug: "booking-saas",
		GitHubRepo:  "brenonaraujo/booking-saas",
		Domain:      "saas",
	}
	jobs := CronCommands(cfg)
	if len(jobs) != 2 {
		t.Fatalf("jobs=%d want 2", len(jobs))
	}
	tm := jobs[0]
	if tm.Name != "booking-saas-loop" {
		t.Fatalf("tm name=%q", tm.Name)
	}
	if !tm.NoAgent {
		t.Fatal("team-manager cron MUST be no_agent (script spawn, not LLM-in-cron)")
	}
	if tm.Monitor {
		t.Fatal("team-manager cron MUST NOT use monitor (unchanged triage snapshot skips the agent)")
	}
	if tm.Schedule != DefaultTMInterval {
		t.Fatalf("tm schedule=%q want %q", tm.Schedule, DefaultTMInterval)
	}
	if !strings.Contains(tm.Script, "spawn-tm.sh") {
		t.Fatalf("tm script=%q", tm.Script)
	}
	orch := jobs[1]
	if orch.NoAgent {
		t.Fatal("orchestrator cron MUST run an agent (supervisor)")
	}
	if orch.Monitor {
		t.Fatal("orchestrator MUST NOT use monitor either")
	}
	if orch.Deliver != "bot-chat:default" {
		t.Fatalf("orchestrator deliver=%q", orch.Deliver)
	}
	if orch.Schedule != DefaultOrchInterval {
		t.Fatalf("orch schedule=%q", orch.Schedule)
	}
}

func TestCronArgv_NoMonitorFlags(t *testing.T) {
	cfg := Config{ProjectRoot: "/tmp/x", ProjectSlug: "x"}
	for _, job := range CronCommands(cfg) {
		argv := job.Argv()
		joined := strings.Join(argv, " ")
		if strings.Contains(joined, "--monitor") {
			t.Fatalf("cron argv leaked monitor: %s", joined)
		}
		if argv[0] != "hermes" || argv[1] != "cron" || argv[2] != "create" {
			t.Fatalf("argv prefix: %v", argv[:3])
		}
		if job.NoAgent && !contains(argv, "--no-agent") {
			t.Fatalf("missing --no-agent: %v", argv)
		}
	}
}

func TestIssue0_IsTheSpecBootstrap(t *testing.T) {
	title := Issue0Title("booking-saas")
	if !strings.Contains(strings.ToLower(title), "spec") && !strings.Contains(strings.ToLower(title), "bootstrap") {
		t.Fatalf("issue 0 title should name spec/bootstrap, got %q", title)
	}
	body := Issue0Body(Config{ProjectName: "booking-saas", Domain: "saas", GitHubRepo: "acme/booking-saas"}, "# SaaS de agendamento\n\nUsuários marcam horários.")
	for _, needle := range []string{"agendamento", "type/feature", "team-manager", "gmh loop"} {
		if !strings.Contains(body, needle) {
			t.Fatalf("issue 0 body missing %q", needle)
		}
	}
}

func TestCanonicalLabels_IncludeFlowAndType(t *testing.T) {
	names := map[string]bool{}
	for _, l := range CanonicalLabels() {
		if l.Name == "" || l.Color == "" {
			t.Fatalf("incomplete label: %+v", l)
		}
		names[l.Name] = true
	}
	for _, need := range []string{
		"triage", "refined", "ready", "in-progress", "in-review", "qa", "done", "blocked",
		"type/feature", "type/technical", "type/infra", "type/bug",
		"backend", "frontend", "infra",
	} {
		if !names[need] {
			t.Fatalf("missing label %q", need)
		}
	}
}

func TestDoctor_EmptyDirFails(t *testing.T) {
	dir := t.TempDir()
	rep := Doctor(dir)
	if rep.Pass {
		t.Fatal("empty project must fail loop doctor")
	}
	if len(rep.Failed) == 0 {
		t.Fatal("expected failed checks")
	}
}

func TestDoctor_AfterWriteScriptsPassesCore(t *testing.T) {
	root := t.TempDir()
	cfg := Config{
		ProjectRoot: root,
		ProjectSlug: "demo",
		ProjectName: "demo",
		GitHubRepo:  "acme/demo",
		Domain:      "saas",
	}
	if err := WriteProjectFiles(cfg, "# spec\n"); err != nil {
		t.Fatal(err)
	}
	// Minimal persona contract so doctor can see the guardrail.
	personaDir := filepath.Join(root, "harness", "personas")
	if err := os.MkdirAll(personaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tm := "# Persona\nVocê **não implementa código de feature**. Você **orquestra**.\n"
	if err := os.WriteFile(filepath.Join(personaDir, "team-manager.md"), []byte(tm), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".github", "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".github", "workflows", "ci.yml"), []byte("name: Contracts\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	rep := Doctor(root)
	if !rep.Pass {
		t.Fatalf("expected pass, failed=%v checks=%v", rep.Failed, checkNames(rep))
	}
}

func TestSpawnScript_DetachedNoMonitor(t *testing.T) {
	root := t.TempDir()
	cfg := Config{ProjectRoot: root, ProjectSlug: "demo"}
	if err := WriteProjectFiles(cfg, ""); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "harness", "scripts", "loop", "spawn-tm.sh"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, need := range []string{"nohup", "hermes -p", "--oneshot", "busy"} {
		if !strings.Contains(s, need) {
			t.Fatalf("spawn-tm.sh missing %q", need)
		}
	}
	if strings.Contains(s, "--monitor") {
		t.Fatal("spawn-tm.sh must not mention --monitor")
	}
	tick, err := os.ReadFile(filepath.Join(root, "harness", "scripts", "loop", "team-manager-tick.md"))
	if err != nil {
		t.Fatal(err)
	}
	ts := string(tick)
	if !strings.Contains(ts, "NÃO escreva código") && !strings.Contains(strings.ToLower(ts), "do not write") {
		t.Fatal("tick must forbid implementing")
	}
	if !strings.Contains(ts, "idle") {
		t.Fatal("tick must define idle")
	}
}

func TestHermesScriptWrapper_ExecsProjectScript(t *testing.T) {
	w := HermesScriptWrapper("/proj", "demo")
	if !strings.Contains(w, "/proj/harness/scripts/loop/spawn-tm.sh") {
		t.Fatalf("wrapper: %s", w)
	}
	if !strings.Contains(w, "exec") {
		t.Fatal("wrapper must exec the project script")
	}
}

func TestCIKind_ContractsVsFull(t *testing.T) {
	if k := CIKind(""); k != "full" {
		t.Fatalf("default stack CI=%q", k)
	}
	if k := CIKind("go,nuxt,postgresql"); k != "full" {
		t.Fatalf("go+nuxt CI=%q", k)
	}
	if k := CIKind("vue,swarm"); k != "contracts" {
		t.Fatalf("vue swarm CI=%q want contracts", k)
	}
	if k := CIKind("docs"); k != "contracts" {
		t.Fatalf("docs CI=%q", k)
	}
}

func contains(argv []string, s string) bool {
	for _, a := range argv {
		if a == s {
			return true
		}
	}
	return false
}

func checkNames(r Report) []string {
	out := make([]string, 0, len(r.Checks))
	for _, c := range r.Checks {
		out = append(out, c.Name+":"+boolStr(c.Pass))
	}
	return out
}

func boolStr(b bool) string {
	if b {
		return "pass"
	}
	return "fail"
}
