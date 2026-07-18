# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

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

### Validated in production
- **Piloto Mandaí v2** — marketplace B2B2C de compra coletiva
  comunitária (estilo Meituan Select / Duoduo Maicai), validado
  end-to-end com Hermes Agent:
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
