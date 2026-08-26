# Persona evolve (v1.15.0) Implementation Plan

> **For Hermes:** Use subagent-driven-development. Tasks A/B/C/D are independent (disjoint files). Parent wires `root.go`, VERSION, CHANGELOG, PR.

**Goal:** Make domain-expert profiles actually instantiable from project context, persist generated-harness memory, and add a Stanford-style outer loop that improves personas/skills from GitHub issue+comment history — with Hermes as the OS/tool harness underneath.

**Architecture:** Three new CLI packages (`personas`, `harnessmem`, `evolve`) plus docs. No new Go deps. Tests use `t.TempDir()` fixtures. `gmh evolve` never calls a live LLM in unit tests; it writes a proposer prompt that Hermes (or any agentic) can run.

**Tech Stack:** Go 1.26.5, cobra, stdlib tests. Module `github.com/brenonaraujo/git-meta-harness/cli`.

---

### Task A — `gmh personas` for real (dynamic domain-expert)

**Files:**
- Create: `cli/internal/personas/personas.go`
- Create: `cli/internal/personas/personas_test.go`
- Modify: `cli/cmd/personas.go` (replace TODOs)

**Behavior:**
- `Slug(domain) string` — kebab-case, ASCII fold, reject empty / `generic` / `domain-expert`
- `Create(harnessDir, domain, context string) (path, error)` — copy `harness/personas/domain-expert.template.md` if present, else a short built-in skeleton; replace `<domínio>` and `<seu-dominio>`; append `## Project context` when context != ""; write `harness/personas/domain-expert-<slug>.md`; refuse overwrite
- `List(harnessDir) []string` — persona basenames
- `Remove(harnessDir, name)` — only `domain-expert-*`, never template/adopter
- Wire cobra: list / create `--domain` `--context` `--from-spec` (read file into context) / remove

**Tests:** slug, refuse generic, create writes file with domain name, context section present, list sees it, remove deletes, second create errors.

**Do not** edit `cli/cmd/root.go`, VERSION, CHANGELOG.

---

### Task B — generated harness memory

**Files:**
- Create: `cli/internal/harnessmem/snapshot.go`
- Create: `cli/internal/harnessmem/snapshot_test.go`
- Create: `cli/cmd/memory.go`

**Behavior:**
- `Snapshot` JSON: `schema_version` (`git-meta-harness/memory/v1`), `framework_version`, `generated_at` RFC3339, `runtime` (hermes|claude-code|none), `project`, `personas []string`, `skills []string`, `hermes_profiles []string`, `notes` explaining Hermes = OS/tool harness, this snapshot = delivery-harness memory
- `Write(harnessDir, snap) error` → `harness/memory/snapshot.json` (mkdir)
- `Read(harnessDir) (*Snapshot, error)`
- `Build(harnessDir, hermesHome, runtime, version string) Snapshot` — scan personas + skills dirs; if hermesHome set, list profiles
- cobra `gmh memory write|show`

**Do not** edit `root.go`.

---

### Task C — Stanford-style evolve loop over personas/skills

**Files:**
- Create: `cli/internal/evolve/evolve.go`
- Create: `cli/internal/evolve/evolve_test.go`
- Create: `cli/cmd/evolve.go`

**Behavior (deterministic, no LLM in tests):**
- `Comment` struct: Issue, Author, Body, CreatedAt
- `LoadComments(dir string)` — each `*.json` is `{issue, author, body, created_at}` OR a `comments.json` array
- `WriteTrace(memoryDir string, comments []Comment, proposal string) (traceDir, error)` — `harness/memory/traces/<utc>/comments.json` + `proposal.md` + `PROMPT.md`
- `Propose(comments []Comment, personaNames []string) string` — markdown proposal: which persona/skill to patch, quoting comment snippets that mention AC/DoD/skill gaps; always include a section "Run this prompt in Hermes team-manager (OS/tool harness)"
- cobra `gmh evolve --from-dir DIR [--apply]` — `--apply` writes `harness/memory/traces/...` only (does not overwrite personas without explicit later command). `--from-dir` required for v1 (gh harvest can be a follow-up; if `gh` exists, optional `--from-github` may shell out, but tests use `--from-dir`)

**Do not** edit `root.go`.

---

### Task D — docs + ADR (markdown only)

**Files:**
- Create: `docs/EVOLVE.md`
- Create: `harness/workflow/08-persona-evolve.md`
- Modify: `harness/contrib/design-decisions.md` — append **ADR-0030** (do not rewrite earlier ADRs)
- Modify: `docs/LOOP.md` — short section pointing at EVOLVE.md
- Modify: `docs/ECOSYSTEM.md` — Stanford IRIS bridge is no longer "v2 idea only"; `gmh evolve` is the governance-side analog

**Content must state:**
1. Domain-expert profiles are created dynamically from project context (`gmh personas create --domain X --context/--from-spec`).
2. Hermes is the runtime/OS-tool harness; git-meta-harness emits project-specific agents, skills, tools, memory.
3. Personas/skills self-improve from issue+comment history via `gmh evolve` (filesystem of traces, like the paper).
4. Human still validates before persona files are edited in anger.

---

### Parent after A–D

- Register `MemoryCmd()` and `EvolveCmd()` in `cli/cmd/root.go`
- VERSION `1.15.0`, CHANGELOG entry, README "what's new"
- `cd cli && go test ./... && go vet ./...`
- Commit remaining, push, PR
- Update brenon.cloud blog EN+PT
