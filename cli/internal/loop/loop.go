// Package loop is the Hermes-cron pooling control plane for a
// meta-harness project. GitHub has no webhook into a local Hermes
// process; the scheduler is the hook. team-manager is woken on a
// heartbeat, reads GitHub Issues, and dispatches personas. It MUST
// NOT implement feature code.
//
// Lessons encoded here (home.cloud, 2026-08-31):
//   - Create Hermes profiles with --no-skills BEFORE writing SOUL.md.
//   - Do not put cron `monitor` on the team-manager job (unchanged
//     triage snapshots suppress the agent forever).
//   - Spawn workers with nohup, never the parent chat's background
//     terminal (parent tracker SIGINT/SIGKILL).
//   - no_agent spawn scripts must print a line (empty stdout = UI
//     "sem execução").
//   - Orchestrator cron must never be [SILENT].
package loop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultTMInterval   = "2m"
	DefaultOrchInterval = "5m"
)

// Config describes one project's loop.
type Config struct {
	ProjectRoot  string
	ProjectSlug  string
	ProjectName  string
	GitHubRepo   string // owner/repo, optional
	Domain       string
	TMInterval   string
	OrchInterval string
	TMProfile    string
}

func (c Config) withDefaults() Config {
	if c.ProjectSlug == "" {
		c.ProjectSlug = slugify(c.ProjectName)
	}
	if c.ProjectName == "" {
		c.ProjectName = c.ProjectSlug
	}
	if c.TMInterval == "" {
		c.TMInterval = DefaultTMInterval
	}
	if c.OrchInterval == "" {
		c.OrchInterval = DefaultOrchInterval
	}
	if c.TMProfile == "" {
		c.TMProfile = "team-manager"
	}
	if c.Domain == "" {
		c.Domain = "internal"
	}
	return c
}

// CanonicalPersonas is the set gmh seed materializes. domain-expert
// is ALWAYS specialized (invariant 12).
func CanonicalPersonas(domain string) []string {
	slug := slugify(domain)
	if slug == "" {
		slug = "internal"
	}
	return []string{
		"team-manager",
		"domain-expert-" + slug,
		"solutions-architect",
		"backend-engineer",
		"frontend-engineer",
		"quality-assurance",
		"devops-engineer",
	}
}

// Label is a GitHub label to create at seed.
type Label struct {
	Name        string
	Color       string
	Description string
}

// CanonicalLabels is the control-plane label set from harness/AGENTS.md §4.
func CanonicalLabels() []Label {
	return []Label{
		{"triage", "cccccc", "Issue nova, ainda nao avaliada"},
		{"needs-info", "fbca04", "Faltam informacoes do autor"},
		{"refined", "0e8a16", "domain-expert refinou a historia"},
		{"ready", "0e8a16", "solutions-architect definiu DoD"},
		{"in-progress", "1d76db", "Builder implementando"},
		{"in-review", "1d76db", "Builder terminou, QA rodando"},
		{"qa", "5319e7", "QA aprovou; aguarda validacao"},
		{"done", "0e8a16", "Mergeado + release feito"},
		{"blocked", "b60205", "Bloqueado por dependencia externa"},
		{"duplicate", "cccccc", "Duplicado"},
		{"type/feature", "7057ff", "Feature de negocio"},
		{"type/technical", "5319e7", "Setup tecnico puro"},
		{"type/infra", "5319e7", "Infraestrutura"},
		{"type/bug", "b60205", "Bug"},
		{"type/tech-debt", "fbca04", "Divida tecnica"},
		{"type/docs", "0075ca", "Documentacao"},
		{"type/spike", "c5def5", "Spike / pesquisa"},
		{"backend", "bfd4f2", "Componente backend"},
		{"frontend", "bfd4f2", "Componente frontend"},
		{"infra", "bfd4f2", "Componente infra/devops"},
		{"breaking-change", "b60205", "Mudanca incompativel"},
		{"tech-debt", "fbca04", "Divida tecnica"},
		{"security", "b60205", "Issue de seguranca"},
		{"documentation", "0075ca", "Docs"},
	}
}

// CronJob is one hermes cron create invocation.
type CronJob struct {
	Name     string
	Schedule string
	NoAgent  bool
	Monitor  bool // MUST stay false for TM and orchestrator
	Script   string
	Workdir  string
	Deliver  string
	Prompt   string
}

// Argv is the `hermes cron create` command. Never includes --monitor-*.
func (j CronJob) Argv() []string {
	argv := []string{"hermes", "cron", "create", "--name", j.Name, "--workdir", j.Workdir}
	if j.NoAgent {
		argv = append(argv, "--no-agent")
	}
	if j.Script != "" {
		argv = append(argv, "--script", j.Script)
	}
	if j.Deliver != "" {
		argv = append(argv, "--deliver", j.Deliver)
	}
	argv = append(argv, j.Schedule)
	if j.Prompt != "" && !j.NoAgent {
		argv = append(argv, j.Prompt)
	}
	return argv
}

// CronCommands returns the two jobs that keep the loop alive.
func CronCommands(cfg Config) []CronJob {
	cfg = cfg.withDefaults()
	script := cfg.ProjectSlug + "-spawn-tm.sh"
	return []CronJob{
		{
			Name:     cfg.ProjectSlug + "-loop",
			Schedule: cfg.TMInterval,
			NoAgent:  true,
			Monitor:  false,
			Script:   script,
			Workdir:  cfg.ProjectRoot,
		},
		{
			Name:     cfg.ProjectSlug + "-orchestrator",
			Schedule: cfg.OrchInterval,
			NoAgent:  false,
			Monitor:  false,
			Workdir:  cfg.ProjectRoot,
			Deliver:  "bot-chat:default",
			Prompt:   orchPrompt(cfg),
		},
	}
}

func orchPrompt(cfg Config) string {
	return strings.Join([]string{
		"You are the meta-harness loop supervisor for " + cfg.ProjectName + ".",
		"Repo: " + cfg.GitHubRepo + "  Root: " + cfg.ProjectRoot,
		"Read harness/scripts/loop/orchestrator-verify.md and follow it.",
		"Always print 3-6 lines. Never [SILENT]. Never implement feature code.",
		"Check: gateway up, TM cron last_run, open issues/PRs, dead workers.",
		"If the board is stuck (triage with no dispatch, qa without PR, worker dead), unblock.",
		"Do NOT put monitor on the TM job. Do NOT spawn if team-manager is already running.",
	}, " ")
}

// Issue0Title is the first GitHub issue — the spec that starts the loop.
func Issue0Title(name string) string {
	if name == "" {
		name = "project"
	}
	return "Bootstrap: functional spec + harness loop (" + name + ")"
}

// Issue0Body is the body of the seed issue. Labels applied by the CLI.
func Issue0Body(cfg Config, spec string) string {
	cfg = cfg.withDefaults()
	spec = strings.TrimSpace(spec)
	if spec == "" {
		spec = "(spec not provided — team-manager must ask the human and write docs/SPEC.md before any type/feature work)"
	}
	if len(spec) > 4000 {
		spec = spec[:4000] + "\n\n…(truncated; full text in docs/SPEC.md)"
	}
	return fmt.Sprintf(`## Issue 0 — specification that starts the loop

This issue is the **principal** of the first delivery cycle.
The host agent that created the repo MUST NOT implement the product.
`+"`"+`team-manager`+"`"+` (Hermes cron pooling) owns the flow.

### Project
- Name: %s
- Domain: `+"`"+`domain/%s`+"`"+`
- GitHub: %s

### Spec (source of truth: `+"`"+`docs/SPEC.md`+"`"+`)

%s

### Acceptance (this issue)
- [ ] `+"`"+`docs/SPEC.md`+"`"+` committed
- [ ] Personas materialized (`+"`"+`gmh agents list`+"`"+` shows 7 profiles, specialized domain-expert)
- [ ] `+"`"+`.github/workflows/ci.yml`+"`"+` and `+"`"+`release.yml`+"`"+` present
- [ ] Canonical labels exist
- [ ] `+"`"+`gmh loop doctor`+"`"+` exits 0
- [ ] Hermes cron `+"`"+`%s-loop`+"`"+` (no_agent, no monitor) and `+"`"+`%s-orchestrator`+"`"+` installed
- [ ] team-manager tick does **not** write feature code
- [ ] First type/feature sub-issue opened from the spec (not implemented in the parent chat)

### Routing
`+"`"+`type/feature`+"`"+` → domain-expert-%s → solutions-architect → builders → qa → human → merge.

### Guardrail
If a worker process dies, **fix spawn/profile**. Do not implement in the orchestrator session.
`, cfg.ProjectName, cfg.Domain, cfg.GitHubRepo, spec, cfg.ProjectSlug, cfg.ProjectSlug, cfg.Domain)
}

// CIKind chooses the GitHub Actions template. Go+Nuxt gets the full
// modular CI; everything else gets a thin contracts workflow so we
// never copy GHCR/cosign onto a landing zone.
func CIKind(stack string) string {
	s := strings.ToLower(stack)
	if s == "" {
		return "full"
	}
	hasGo := strings.Contains(s, "go")
	hasNuxt := strings.Contains(s, "nuxt")
	if hasGo && hasNuxt {
		return "full"
	}
	if hasGo && (strings.Contains(s, "node") || strings.Contains(s, "vue") || strings.Contains(s, "react")) {
		return "full"
	}
	return "contracts"
}

// Check is one loop-doctor row.
type Check struct {
	Name   string
	Pass   bool
	Detail string
}

// Report is the result of Doctor.
type Report struct {
	Pass   bool
	Checks []Check
	Failed []string
}

// Doctor validates that the harness loop is actually wired on disk.
// It does not require a live Hermes gateway (cron presence is a
// warning emitted by `gmh loop status`).
func Doctor(root string) Report {
	var checks []Check
	add := func(name string, pass bool, detail string) {
		checks = append(checks, Check{Name: name, Pass: pass, Detail: detail})
	}
	file := func(rel, name string) {
		p := filepath.Join(root, rel)
		_, err := os.Stat(p)
		add(name, err == nil, p)
	}

	file("harness/scripts/loop/spawn-tm.sh", "spawn-tm.sh")
	file("harness/scripts/loop/spawn-persona.sh", "spawn-persona.sh")
	file("harness/scripts/loop/team-manager-tick.md", "team-manager-tick.md")
	file("harness/scripts/loop/orchestrator-verify.md", "orchestrator-verify.md")
	file("harness/scripts/check-loop.sh", "check-loop.sh")
	file("harness/loop.env", "loop.env")

	if data, err := os.ReadFile(filepath.Join(root, "harness", "scripts", "loop", "spawn-tm.sh")); err == nil {
		s := string(data)
		add("spawn-tm uses nohup", strings.Contains(s, "nohup"), "")
		add("spawn-tm has busy-exit", strings.Contains(s, "busy"), "")
		add("spawn-tm has no monitor", !strings.Contains(s, "--monitor"), "")
		add("spawn-tm prints a line", strings.Contains(s, "echo "), "")
	}
	if data, err := os.ReadFile(filepath.Join(root, "harness", "scripts", "loop", "team-manager-tick.md")); err == nil {
		s := string(data)
		forbid := strings.Contains(s, "NÃO escreva código") || strings.Contains(strings.ToLower(s), "do not write")
		add("tick forbids implementing", forbid, "")
		add("tick defines idle", strings.Contains(s, "idle"), "")
	}
	if data, err := os.ReadFile(filepath.Join(root, "harness", "personas", "team-manager.md")); err == nil {
		s := strings.ToLower(string(data))
		add("team-manager persona forbids feature code",
			strings.Contains(s, "não implementa") || strings.Contains(s, "nao implementa") || strings.Contains(s, "does not implement"),
			"")
	} else {
		add("team-manager persona present", false, err.Error())
	}
	ci := filepath.Join(root, ".github", "workflows", "ci.yml")
	_, errCI := os.Stat(ci)
	add(".github/workflows/ci.yml", errCI == nil, ci)

	generic := filepath.Join(root, "harness", "personas", "domain-expert.md")
	_, errGen := os.Stat(generic)
	add("no generic domain-expert.md", errGen != nil, generic)

	rep := Report{Pass: true, Checks: checks}
	for _, c := range checks {
		if !c.Pass {
			rep.Pass = false
			rep.Failed = append(rep.Failed, c.Name)
		}
	}
	return rep
}

// WriteProjectFiles writes the generic loop scripts + loop.env +
// check-loop.sh into the project. Idempotent overwrites of the
// framework-owned files; loop.env is created only if missing.
func WriteProjectFiles(cfg Config, spec string) error {
	cfg = cfg.withDefaults()
	root := cfg.ProjectRoot
	dir := filepath.Join(root, "harness", "scripts", "loop")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	files := map[string]string{
		filepath.Join(dir, "spawn-tm.sh"):                          spawnTMScript,
		filepath.Join(dir, "spawn-persona.sh"):                     spawnPersonaScript,
		filepath.Join(dir, "team-manager-tick.md"):                 tickMarkdown,
		filepath.Join(dir, "orchestrator-verify.md"):               orchMarkdown,
		filepath.Join(root, "harness", "scripts", "check-loop.sh"): checkLoopScript,
	}
	for path, body := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(path, ".sh") {
			mode = 0o755
		}
		if err := os.WriteFile(path, []byte(body), mode); err != nil {
			return err
		}
	}
	envPath := filepath.Join(root, "harness", "loop.env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		env := fmt.Sprintf("GMH_PROJECT_SLUG=%s\nGMH_PROJECT_NAME=%s\nGMH_GITHUB_REPO=%s\nGMH_DOMAIN=%s\nGMH_TM_PROFILE=%s\n",
			cfg.ProjectSlug, cfg.ProjectName, cfg.GitHubRepo, cfg.Domain, cfg.TMProfile)
		if err := os.WriteFile(envPath, []byte(env), 0o644); err != nil {
			return err
		}
	}
	_ = spec
	return nil
}

// HermesScriptWrapper is copied to ~/.hermes/scripts/<slug>-spawn-tm.sh
// because `hermes cron --script` resolves under ~/.hermes/scripts/.
func HermesScriptWrapper(projectRoot, slug string) string {
	_ = slug
	return fmt.Sprintf("#!/usr/bin/env bash\nset -euo pipefail\nexec bash %q\n",
		filepath.Join(projectRoot, "harness", "scripts", "loop", "spawn-tm.sh"))
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prev := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prev = false
		case r == ' ' || r == '_' || r == '-' || r == '/' || r == '.':
			if !prev && b.Len() > 0 {
				b.WriteByte('-')
				prev = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// Embedded scripts. Keep in sync with harness/scripts/loop/* — the
// CLI writes these so `gmh seed` does not depend on a later rsync.

const spawnTMScript = `#!/usr/bin/env bash
# Cron (no_agent): wake team-manager if it is not already running.
# Always print one line so the cron UI is not "no execution".
# Do NOT attach a change-detector gate to this job.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
if [[ -f "$ROOT/harness/loop.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/harness/loop.env"
  set +a
fi
PROFILE="${GMH_TM_PROFILE:-team-manager}"
TICK="${GMH_TICK_FILE:-$ROOT/harness/scripts/loop/team-manager-tick.md}"
SLUG="${GMH_PROJECT_SLUG:-$(basename "$ROOT")}"
LOG="${GMH_TM_LOG:-/tmp/gmh-${SLUG}-tm.log}"

if pgrep -fl "hermes -p ${PROFILE}" 2>/dev/null | grep -v pgrep >/dev/null; then
  echo "busy ${PROFILE}"
  exit 0
fi
nohup hermes -p "$PROFILE" chat --oneshot --in "$ROOT" --query-file "$TICK" >>"$LOG" 2>&1 &
echo "spawned ${PROFILE} pid $! log=$LOG"
`

const spawnPersonaScript = `#!/usr/bin/env bash
# Detached persona spawn (not tracked by the parent Hermes chat).
# Usage: spawn-persona.sh <profile> <brief-file> [repo]
set -euo pipefail
profile="${1:?profile}"
brief="${2:?brief file}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
if [[ -f "$ROOT/harness/loop.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/harness/loop.env"
  set +a
fi
repo="${3:-$ROOT}"
id="$(basename "$brief" .md)"
log="/tmp/gmh-${id}.log"
if pgrep -fl "hermes -p ${profile} " 2>/dev/null | grep -v pgrep >/dev/null; then
  echo "busy: $profile"
  exit 0
fi
nohup hermes -p "$profile" chat --oneshot --in "$repo" --query-file "$brief" >>"$log" 2>&1 &
echo "spawned $profile pid $! log=$log"
`

const tickMarkdown = `# team-manager tick

You are the **team-manager**. This tick is short.

Read, in order: ` + "`harness/loop.env`" + `, ` + "`harness/PROJECT.md`" + ` (if present), ` + "`harness/personas/team-manager.md`" + `, ` + "`harness/workflow/05-orchestration.md`" + `, ` + "`harness/workflow/07-hermes-loop.md`" + `.

## Guardrails (non-negotiable)

- **NÃO escreva código** de feature, Dockerfile, netlify, Vue, Go de produto.
- **NÃO** espere o worker. Spawn and exit.
- Worker crash (SIGINT/SIGKILL) is a **spawn/profile bug**. Fix spawn. Do not implement.
- If there is **no new action** (spawn, merge, label move, open PR): print ` + "`idle`" + ` and **do not comment on GitHub**.
- Do not repeat the same status comment.

## Board

1. ` + "`gh issue list --state open`" + ` and ` + "`gh pr list --state open`" + `.
2. blocked-by OPEN → do not start the child.
3. **Move labels**. A comment without a label change does not count.
   - triage + type/feature → spawn domain-expert-<domain>
   - triage + type/infra|technical → spawn solutions-architect
   - refined → spawn solutions-architect
   - ready → create branch ` + "`feature/<id>-<slug>`" + ` + spawn builder
   - in-progress + PR + CI green → spawn quality-assurance
   - qa + mergeable PR → squash merge, ` + "`done`" + `, close (or wait for human ` + "`validado`" + ` if PROJECT.md says so)
   - qa + no PR + last report APROVADO → ` + "`done`" + `, close. Not idle.
   - qa + no PR + last report REPROVADO → spawn builder. Not idle.
4. Spawn with ` + "`harness/scripts/loop/spawn-persona.sh <profile> <brief>`" + ` (` + "`nohup`" + `). Never ` + "`terminal(background=true)`" + ` in this process.
5. Brief starts with **peguei**. Worker does not merge or close.
6. Max 3 workers. Same-branch FE+BE only if path-scope is disjoint (sensor 10).

Print 3–6 lines. Then exit.
`

const orchMarkdown = `# Loop supervisor

You check that the meta-harness loop is alive. You do **not** implement.

Always print 3–6 lines. Never ` + "`[SILENT]`" + `. Empty stdout looks dead.

## Checks

1. Hermes gateway running (` + "`hermes cron status`" + ` / gateway). If down: say so, do not pretend the board is idle.
2. Job ` + "`<slug>-loop`" + ` last_run. If missing: ` + "`gmh loop install`" + `.
3. That job is ` + "`no_agent`" + ` and has **no monitor**.
4. Open issues: if anything sits in ` + "`triage`" + ` / ` + "`qa`" + ` with no worker, the TM tick is stuck — do not "wait". Unblock (restart TM spawn, or report the spawn bug).
5. Dead workers: ` + "`pgrep -fl 'hermes -p'`" + `. A dead builder is not a license for you or TM to write the PR.

Unblock without asking "e aí?".
`

const checkLoopScript = `#!/usr/bin/env bash
# Sensor 14 — loop liveness (filesystem). Exit 1 on any fail.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
if [[ -d "$ROOT/harness" ]]; then
  :
else
  ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
fi
FAILS=0
pass() { echo "  ✅ $1"; }
fail() { echo "  ❌ $1"; FAILS=$((FAILS+1)); }

echo "🔎 Sensor 14 — loop liveness"
echo "Root: $ROOT"

need() {
  if [[ -f "$1" ]]; then pass "$2"; else fail "$2 ($1 missing)"; fi
}
need "$ROOT/harness/scripts/loop/spawn-tm.sh" "spawn-tm.sh"
need "$ROOT/harness/scripts/loop/spawn-persona.sh" "spawn-persona.sh"
need "$ROOT/harness/scripts/loop/team-manager-tick.md" "team-manager-tick.md"
need "$ROOT/harness/scripts/loop/orchestrator-verify.md" "orchestrator-verify.md"
need "$ROOT/harness/loop.env" "loop.env"

if grep -q 'nohup' "$ROOT/harness/scripts/loop/spawn-tm.sh" 2>/dev/null; then
  pass "spawn-tm uses nohup"
else
  fail "spawn-tm missing nohup"
fi
if grep -q -- '--monitor' "$ROOT/harness/scripts/loop/spawn-tm.sh" 2>/dev/null; then
  fail "spawn-tm mentions --monitor (forbidden)"
else
  pass "spawn-tm has no --monitor"
fi
if grep -Eqi 'NÃO escreva código|do not write' "$ROOT/harness/scripts/loop/team-manager-tick.md" 2>/dev/null; then
  pass "tick forbids implementing"
else
  fail "tick does not forbid implementing"
fi
if [[ -f "$ROOT/harness/personas/domain-expert.md" ]]; then
  fail "generic domain-expert.md present (invariant 12)"
else
  pass "no generic domain-expert.md"
fi
if [[ -f "$ROOT/.github/workflows/ci.yml" ]]; then
  pass "ci.yml present"
else
  fail "ci.yml missing"
fi

echo
if [[ "$FAILS" -gt 0 ]]; then
  echo "❌ $FAILS fail(s)"
  exit 1
fi
echo "✅ loop liveness OK"
exit 0
`
