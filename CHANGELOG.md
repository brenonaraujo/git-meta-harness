# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [1.3.1] - 2026-07-18

### Fixed — `docs/PIPELINE.md` Mermaid `stateDiagram-v2` was malformed

The stateDiagram had multi-line transition labels with parens
(`(skip if ...)`) that Mermaid does not parse. The diagram
would render with broken/unexpected text on `triage → refined`,
`refined → ready`, and `ready → in_progress` transitions. This
patch rewrites the diagram with `note` blocks for the
smart-routing skips (which is the correct Mermaid syntax) and
simplifies the transitions to single-line labels.

### Added — `docs/PIPELINE.md` §6: the `gh` CLI trick

New section explaining **why the meta-harness uses the `gh`
CLI instead of MCP servers** (or heavy SDKs) to talk to GitHub.
Covers:
- The actual `gh` commands used by the personas.
- A side-by-side comparison of `gh` vs MCPs (token cost,
  round-trips, setup, composability, etc.).
- Why `gh` is dramatically better for the meta-harness use
  case (one question, one answer, one decision at a time).
- When MCPs would make sense (stateful investigations, custom
  tools, multi-agent coordination) — and why the meta-harness
  doesn't need any of those.

The takeaway: **5-10x less token overhead per interaction**
vs MCP-based equivalents, with no server to run.

### Added — README "Multi-tool portability" table now shows validation status

The table now has a **Validated?** column. Status as of
v1.3.1:

- **Hermes Agent** — ✅ **Yes** (the only tool validated
  end-to-end via the mandai-v2 case).
- All other tools — ⏳ **Adapter-only** (adapters exist;
  full validation is pending).

This makes it explicit that the framework is **designed** to
be tool-agnostic, but the **practical validation** so far
covers only Hermes Agent. Adopters using other tools are
encouraged to contribute their validation back.

### Added — README "Validation and test case" section clarified

Now explicitly states that **mandai-v2 was built with Hermes
Agent** (the only tool tested with the framework so far).

### Why v1.3.1 (patch)

A diagram fix + a clarifying section + a README clarification.
No breaking change, no new features (the framework itself is
unchanged).

## [1.3.0] - 2026-07-18

### Added — `docs/LOOP.md`: how git-meta-harness fits loop engineering
and this project adheres to [Semantic Versioning](https://semver.org/).

## [1.3.0] - 2026-07-18

### Added — `docs/LOOP.md`: how git-meta-harness fits loop engineering

Loop engineering became the dominant AI agent pattern in
mid-2026 (Addy Osmani, IBM Think, AI Builder Club). The
meta-harness is a **concrete, ship-now implementation** of
loop engineering for greenfield software delivery.

This release adds a 9-section, 6-diagram canonical document
articulating the relationship:

- **§1** — what loop engineering is (the 4 loop types, the 5
  building blocks, the verifier-is-the-bottleneck insight).
- **§2** — direct mapping from loop engineering concepts to
  meta-harness artifacts (automations → GitHub Actions,
  sub-agents → 7 personas, memory → issues + ADRs + 18
  invariants, etc.).
- **§3** — where the meta-harness goes beyond generic loop
  engineering: 9 verifiers (sensors) with explicit fail
  actions, 18 invariants, testable stop conditions, durable
  auditable memory.
- **§4** — all 4 loop types (heartbeat, cron, hook, goal)
  implemented in the meta-harness, with concrete locations.
- **§5** — the verifier bottleneck, resolved (9 sensors
  + human validation as the loop's stop condition).
- **§6** — the relationship summarized: loop engineering is
  the discipline, meta-harness is the concrete instance,
  your project is the output.
- **§7** — anti-patterns the meta-harness prevents (with
  explicit defenses).
- **§8** — when to use the meta-harness vs custom loop
  engineering.
- **§9** — the takeaway.

#### Diagrams added (6 in `docs/LOOP.md`)

1. The loop: discover → plan → execute → verify.
2. Mapping loop engineering concepts to meta-harness.
3. The 4 loop types in the meta-harness.
4. The verifier bottleneck, resolved.
5. The relationship: discipline → instance → project.
6. (1 more — see file.)

#### Why v1.3.0 (minor)

The new document is **conceptually significant** (articulates
the framework's relationship to the most important AI agent
pattern of 2026) and warrants a minor bump, not a patch.

## [1.2.4] - 2026-07-18

### Changed — Rename "Validated in production" to "Validation and test case"

Per maintainer feedback, the section that documents the Mandaí v2
pilot is now called **"Validation and test case"** instead of
**"Validated in production"**. The semantics: the project that
exercised the framework end-to-end (Mandaí v2) is referred to as
the **validation and test case**, not as a "production
deployment" reference. The framework itself is what is validated
by applying it to a real project; that application is the
test case, not a production-grade system.

#### Changes

- **`README.md`** (top metadata): `Validated in production: ✅`
  → `Validation and test case: ✅`.
- **`README.md`** §section header: `## Validated in production:
  mandai-v2` → `## Validation and test case: mandai-v2`.
- **`CHANGELOG.md`** §v1.0.0 subsection: `### Validated in
  production` → `### Validation and test case`.

#### Why v1.2.4 (patch)

Purely a terminology fix. No content change, no breaking change.

and this project adheres to [Semantic Versioning](https://semver.org/).

## [1.2.3] - 2026-07-18

### Fixed — Replace remaining ASCII diagrams across the repo

The v1.2.1 and v1.2.2 patches caught the README and the
PIPELINE.md issue lifecycle, but several other docs in the
**`harness/` tree** still had ASCII art:

- `harness/workflow/00-issue-lifecycle.md` — 6 ASCII boxes
  (main flow + 5 type/* variations).
- `harness/bootstrap.md` — 1 large ASCII art in §3 (fluxo geral
  ponta-a-ponta).
- `docs/PIPELINE.md` §5 — 1 ASCII CI workflow diagram.

This patch replaces them all with proper Mermaid diagrams.

#### Changes

- **`harness/workflow/00-issue-lifecycle.md`** — replaced 6 ASCII
  art boxes with Mermaid `stateDiagram-v2` (1 for the main
  `type/feature` flow + 5 for the `type/technical|infra|tech-debt|
  docs|spike` variations).
- **`harness/bootstrap.md`** §3 — replaced the 50-line ASCII art
  with a single Mermaid `flowchart TB` showing the full pipeline
  from user → team-manager → domain-expert → solutions-architect
  → backend+frontend → QA → devops.
- **`docs/PIPELINE.md`** §5 — replaced the ASCII CI workflow
  diagram with a Mermaid `flowchart TB` showing the
  `dorny/paths-filter` changes job + 12 conditional jobs +
  always-on 12-factor and summary.

#### Final Mermaid count

| File | Count |
|---|---|
| `README.md` | 6 |
| `docs/CONCEPT.md` | 2 |
| `docs/ORIGIN.md` | 3 |
| `docs/COMPARISON.md` | 2 |
| `docs/PIPELINE.md` | 2 |
| `harness/bootstrap.md` | 1 |
| `harness/workflow/00-issue-lifecycle.md` | 6 |
| **Total** | **22** |

#### Note on tree-style characters

The `├──` `└──` `│` characters that remain in `docs/CONCEPT.md`
§10.4 are **tree-style directory structure** (a standard markdown
convention for showing filesystem layout, not flow diagrams) and
are intentionally kept.

#### Why v1.2.3 (patch)

Purely a fix to missed diagrams. No content change, no breaking
change.

## [1.2.2] - 2026-07-18

### Fixed — Replace remaining ASCII diagram in `docs/PIPELINE.md`

The v1.2.1 patch missed the issue lifecycle ASCII art in
`docs/PIPELINE.md` §2. This patch replaces it with a proper
Mermaid `stateDiagram-v2`.

#### Changes

- **`docs/PIPELINE.md`** — the issue lifecycle section now uses
  a Mermaid `stateDiagram-v2` instead of the ASCII art box. The
  diagram captures all 7 states (triage, refined, ready,
  in_progress, in_review, qa, awaiting_human, done) with the
  smart-routing skips annotated as transition labels
  (e.g., `triage → refined: type/feature (skip if
  type/technical|...)`).

#### Note on tree-style characters

`docs/CONCEPT.md` still contains `├──` `└──` `│` characters in
its §10.4 "Where the materialized personas live" diagram. These
are **tree-style directory structure** (a standard markdown
convention for showing filesystem layout, not flow diagrams),
so they are intentionally kept.

#### Why v1.2.2 (patch)

Purely a fix to a missed diagram. No content change, no breaking
change.

## [1.2.1] - 2026-07-18

### Fixed — Replace remaining ASCII diagram in README

The v1.2.0 Mermaid pass missed the "Architecture overview" section
in the README, which still contained a 25-line ASCII art box. This
patch replaces it with 3 proper Mermaid diagrams.

#### Changes

- **`README.md`** — the "Architecture overview" section now has
  3 Mermaid diagrams instead of the ASCII box:
  - **The team (7 personas) and the 9 sensors** — single diagram
    showing the team-manager + 7 personas + 9 sensors as nested
    subgraphs.
  - **Sensors (when each runs, what happens on fail)** — 9
    sensors each with their fail action (`blocks merge` vs
    `blocks deploy` vs `blocks release`).
  - **CI workflow (modular with path filters)** — the 1 `changes`
    job at the top + 12 conditional jobs, with the 12-Factor and
    summary jobs marked as always-running (security gates).

#### Why v1.2.1 (patch)

Purely a fix to a missed diagram. No content change, no breaking
change.

## [1.2.0] - 2026-07-18

### Added — Mermaid diagrams across all docs

Documentation is now enriched with **Mermaid diagrams** (rendered
natively by GitHub) that explain the concepts visually. Adds clarity
for adopters and contributors who learn better with diagrams.

#### Diagrams added

- **`README.md`** — 3 new diagrams in a "Visual overview" section:
  - The full loop: spec → team-manager → personas → PR → human → release.
  - The team: 7 personas + smart routing by `type/*` (4 paths).
  - GitHub as native substrate: 5 primitives mapped to project artifacts.
- **`docs/CONCEPT.md`** — 2 new diagrams:
  - §10.0: The two problems solved by the template-vs-materialized
    distinction.
  - §10.2: The 5-step materialization algorithm.
- **`docs/ORIGIN.md`** — 3 new diagrams:
  - §1: The single-agent loop (with `bad` color marking the failure).
  - §4: The extraction from Hermes profiles to the meta-harness.
  - §6: Pattern > tool (tools are ephemeral, pattern is durable).
- **`docs/COMPARISON.md`** — 2 new diagrams:
  - §2: The evolution single-agent → SDD → SPDD → meta-harness.
  - §5: How SDD/SPDD connect to the meta-harness.
- **`docs/PIPELINE.md`** — already has good structure; minor
  improvements in this release.

#### Why v1.2.0

Visual aids make the concepts significantly easier to grasp.
Diagrams are non-breaking (purely additive) but the change is large
enough across 4 docs to warrant a minor bump, not a patch.

## [1.1.1] - 2026-07-18

### Fixed — "Personas are built on demand, not copied" (clarification)

Documentation patch. The v1.1.0 docs described the meta-harness
as "materializing personas" but did not make it explicit enough
that **personas are built on demand from the project's context,
not copied from the templates**. This is a critical distinction:

- A **template** (`harness/personas/*.md` in this repo) is a
  conceptual persona: principles, posture, what they do and
  don't do. Stable across projects.
- A **materialized persona** (lives in the target project) is
  the same persona, plus: the detected stack, the in-context
  skills, the project name and domain knowledge, the runtime
  adapter.

A `domain-expert-banking.md` that has the same content as
`domain-expert.template.md` is a **failure** of the framework,
not a success — it means the materialization step was skipped.

#### Changes

- **`docs/CONCEPT.md`**: new sections §10 ("Personas are built
  on demand for each project") and §11 ("Anti-pattern: 'I copied
  the personas, we're done'"). Includes the two-layer table
  (template vs materialized), the materialization step
  algorithm, where materialized personas live, why the
  distinction matters, and the anti-pattern.
- **`harness/seed/meta-harness-seed.md`**: §1 "MATERIALIZAÇÃO"
  rewritten. Adds a new subsection "Materialização (sempre antes
  dos adapters)" that prescribes the 5-step materialization
  algorithm. The per-tool adapter sections now reference
  "personas materializadas" instead of just "personas". The
  validation subsection explicitly checks that materialized
  personas are not identical to the templates.

#### Why v1.1.1 (patch) and not v1.1.2 or v1.2.0

This is a **clarification of the existing concept**, not a
breaking change and not a feature addition. The behavior was
already what we wanted; the docs just didn't say it clearly
enough. Per [Keep a Changelog](https://keepachangelog.com/),
documentation corrections are patch-level.

## [1.1.0] - 2026-07-18

### Added — Concept documentation

The framework is now articulated explicitly. v1.1.0 is purely
additive: no breaking changes to personas, sensors, ADRs,
invariants, templates, or skills from v1.0.0.

#### New `docs/` directory

- **`docs/CONCEPT.md`** (11K) — The full vision. What the
  meta-harness is, what it is NOT, the "meta" in meta-harness,
  the input (functional spec), the output (a system, not a
  project), the connection with GitHub. For adopters asking
  "does this solve my problem?".
- **`docs/ORIGIN.md`** (8.4K) — The story. Single-agent loop →
  pivot with Hermes Agent → discovery of the "one model, one
  role" pattern → extraction to a tool-agnostic framework →
  validation on Mandaí v2 → the lesson "pattern > tool". For
  maintainers and new contributors.
- **`docs/COMPARISON.md`** (9.6K) — Side-by-side comparison
  with single-agent, SDD, SPDD, and the meta-harness. When to
  use which. How they connect. The meta-harness builds on
  SDD/SPDD, it does not reject them.
- **`docs/PIPELINE.md`** (10K) — How the meta-harness rides on
  GitHub as its native substrate. The 5 primitives
  (Issues, PRs, Labels, Actions, Branch Protection), issue
  lifecycle, PR convention, smart routing, CI workflow with
  path filters. For DevOps operating the pipeline.

#### README updated

- **New top section: "The concept in one paragraph"** —
  one-paragraph summary of the framework + links to the 4
  docs. Above any other section.

#### New ADR

- **ADR-0012** — "O que é (e o que não é) o meta-harness"
  (conceptual decision; documents the rationale for the 4
  docs + the section in the README).

### Why v1.1.0 and not v1.0.1

This is documentation-only, but it is **conceptual** documentation,
not just typo fixes. Bumping the minor signals that the project
has matured its public articulation, while keeping the v1.0.0
contract stable. Per [Keep a Changelog](https://keepachangelog.com/),
documentation can be a minor bump.

## [1.0.0] - 2026-07-18

### Added — First Public Release

The first stable, public, tagged release of **git-meta-harness** — a plug-and-play
multi-agent orchestration framework for greenfield → production software delivery.

#### Core framework
- **7 personas** (`harness/personas/`):
  - `team-manager` — orquestrador ponta-a-ponta com smart routing por `type/*`
  - `domain-expert` — **sempre especializado** (`domain-expert-<domínio>`),
    nunca genérico
  - `solutions-architect` — refinamento técnico, DoD, 12-factor check
  - `backend-engineer` — implementação Go + local pre-flight
  - `frontend-engineer` — implementação Node/Nuxt + local pre-flight
  - `quality-assurance` — validação com sensores, recusa PR com CI vermelho
  - `devops-engineer` — CI/CD, Docker, observability
- **8 sensors** (`harness/sensors/`) — checks automatizados:
  - `00-static-analysis`, `01-vulnerability-scan`, `02-unit-tests`,
    `03-contract-tests`, `04-image-scan`, `05-smoke-tests`,
    `06-load-tests`, `07-twelve-factor-audit`, `08-i18n-audit`
- **6 workflow docs** (`harness/workflow/`) — issue-lifecycle, branching,
  PR, snapshot-deploy, release, orchestration
- **5 stack files** (`harness/stack/`) — backend, frontend, observability,
  docker, code-style + **`versions.md`** canônica
- **13 templates** (`harness/templates/`) — Dockerfile, docker-compose,
  ci.yml, release.yml, .golangci.yml, .env.example, 3 issue templates,
  pr-description, 3 locales (en/pt-BR/es)
- **7 skills** (`harness/skills/`) — github-pr-workflow, github-issues,
  github-code-review, tdd-go, openapi-spec-first, twelve-factor, i18n
- **2 examples** (`harness/examples/`) — `domain-expert-banking`,
  `domain-expert-retail`, `domain-expert-mandai` (com README)
- **`harness/bootstrap.md`** — spec canônica com 13 princípios
- **`harness/AGENTS.md`** — contrato multi-tool (Claude Code, Copilot,
  Codex, OpenCode, Devin, Hermes Agent, Cursor) com 18 invariantes
- **`harness/CLAUDE.md`** — atalho para Claude Code
- **`harness/seed/meta-harness-seed.md`** — prompt de instanciação
  (cola o seed num agentic CLI e materializa o framework no projeto)

#### Cross-cutting principles
- **KISS + DRY**: ≤ 25 linhas/função, ≤ 150 linhas/arquivo, sem comentários
  redundantes
- **12-factor obrigatório** (auditado pelo sensor `07-twelve-factor-audit`)
- **TDD com table-driven tests + testify** (backend) / Vitest (frontend)
- **OpenAPI spec-first** (nunca `swag`)
- **i18n obrigatório** em en, pt-BR, es
- **Observability**: Prometheus + slog JSON
- **Multi-tenant ready** (workspaces + roles)
- **Pix-first** payments-ready (pt-BR)
- **Stack pinada** (sem `latest`): Go 1.26.5, Node 24 LTS, Nuxt 4.5,
  PostgreSQL 18.4, etc. — fonte canônica `harness/stack/versions.md`

#### Smart routing & multi-tool portability
- **`team-manager` com smart routing** por `type/feature|technical|
  infra|bug|tech-debt|docs|spike` (ADR-0004)
- **`team-manager` cria branch e delega**; builders clonam (ADR-0006)
- **Domain-expert sempre especializado** com label `domain/<x>` (ADR-0003)
- **`interactions.md`** — matriz "quem pode fazer o quê" entre personas
- **Hermes Agent** profiles herdam modelo default do `config.yaml`
  (nunca passa `--model`)
- **Multi-tool**: mesmo `harness/` materializado roda em Claude Code,
  Copilot, Codex, OpenCode, Devin, Hermes Agent, Cursor

#### Sensores e CI/CD
- **`smoke-test.sh`** — 12 checks que detectaram 11 bugs no piloto Mandaí v2
- **`check-stack-versions.sh`** com modo `--check-latest` — pesquisa online
  (GitHub API + Docker Hub) por drift, EOL, versões comprometidas
- **CI workflow modular** com `dorny/paths-filter` (jobs rodam só nos
  componentes que mudaram), concurrency, cache com scope por service,
  Trivy SHA-pinado, `GOTOOLCHAIN: local`
- **12 invariantes** no `AGENTS.md` §8 + 6 novas (17, 18, 19) pós-piloto

#### Governance
- **10 ADRs** registrados em `harness/contrib/design-decisions.md`
- **Política de versões pinadas** (ADR-0009) com `versions.md` fonte
  canônica e checagem online de latest
- **Smoke test + local pre-flight** (ADR-0007, ADR-0008) — gate
  obrigatório antes de processar issues

### Validation and test case
- **Piloto Mandaí v2** — marketplace B2B2C de compra coletiva
  comunitária (estilo Meituan Select / Duoduo Maicai), servindo
  como **validation and test case** end-to-end com Hermes Agent:
  - Repo: https://github.com/brenonaraujo/mandai-v2
  - 4 issues, 5 commits, 1 PR
  - Stack: Go 1.25 + Gin + GORM + PostgreSQL + Nuxt 4 + Pinia
  - i18n em en/pt-BR/es
  - 9 gotchas detectados e prevenidos pelo framework

### Notes
- This is a **v1.0.0** — backward-compatible APIs are stable but the
  format and content may evolve in 1.x with backwards compatibility.
- For projects materializing from a previous draft, see
  [Migration from 0.x](#migration-from-0x) below.

## [0.2.0] - 2026-07-18 (draft)

- Smart routing + domain-expert especializado + i18n
- Smoke test + check-stack-versions + ADRs 0001-0009
- Local pre-flight + CI workflow robusto

## [0.1.0] - 2026-07-18 (draft)

- 7 personas + 8 sensors + 6 workflow docs
- Stack files + 10 templates
- Multi-tool via `AGENTS.md`

[1.0.0]: https://github.com/brenonaraujo/git-meta-harness/releases/tag/v1.0.0
[0.2.0]: https://github.com/brenonaraujo/git-meta-harness/releases/tag/v0.2.0
[0.1.0]: https://github.com/brenonaraujo/git-meta-harness/releases/tag/v0.1.0
