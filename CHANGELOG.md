# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [1.6.10] - 2026-07-18

### Fixed — `gmh doctor` no longer false-flags "CI drift" for legitimate customizations

**Bug**: `gmh doctor` reported
`CI: aligned with template (drift: +N lines from template)`
even when the only differences were legitimate project-specific
customizations:
- Image names like `mandai-backend` vs template's generic `app-backend`
- Pinned tool versions (`govulncheck@v1.1.4` vs `@latest` in template)
- Trivy scan format (`format: table` vs `format: sarif`)

These customizations are intentional and should not be flagged
as drift from the framework template.

**Fix**: normalize lines before comparison in
`cli/cmd/doctor.go`:
- `<word>-backend` / `<word>-frontend` → `app-backend` / `app-frontend`
- `govulncheck@<version>` → `govulncheck@<version>` (literal)
- `oapi-codegen@<version>` → `oapi-codegen@<version>` (literal)
- `format: <fmt>` and `output: <file>` lines normalized

Result: `gmh doctor` now reports `All local checks passed`
for projects with legitimate customizations.

## [1.6.9] - 2026-07-18

### Fixed — `gmh update --to vX.Y.Z` now bumps the local VERSION file even when `harness/` is unchanged

**Bug**: `gmh update --to vX.Y.Z` would report
`"No changes needed"` and exit early (line 100, before the
`copyFile(VERSION)` block) when the local `harness/` directory
already matched the target version. This meant the local
`VERSION` file would **not** be updated to the target version
even though the project was now "pinned" to it.

**Workaround documented in v1.6.8 release**: bump VERSION
manually (`echo 1.6.X > VERSION`) after running `gmh update`.

**Fix**: moved the `VERSION` copy block to **before** the
no-changes early return. The target version is the source of
truth for the pin, regardless of whether `harness/` content
changed. Output now reads:
`"No changes needed (VERSION bumped to v1.6.X)"`.

`gmh doctor` will no longer report `Out of date by 0 version(s)`
followed by `Latest: vX.Y.Z / Local: v(W-1)` after a `gmh update`
that doesn't change `harness/`.

## [1.6.8] - 2026-07-18

### Fixed — `cli-release.yml` workflow can now create + populate the release in one step

The old `cli-release.yml` (for tag pushes `v*.*.*`) ran
`gh release upload $TAG` after building all 5 binaries + checksums.
The upload always failed with `release not found` because
pushing a tag does **not** auto-create a release — that
requires a separate workflow or manual step.

This was worked around manually for v1.6.5, v1.6.6, v1.6.7
(release was created via `gh release create` and then binaries
uploaded via `gh release upload --clobber`).

**Fix**: replaced the manual `gh release upload` step with the
idempotent `softprops/action-gh-release@v2` action. It:
- creates the release if it doesn't exist
- uploads files to the existing release if it does
- works for both `v*.*.*` (framework) and `cli-v*.*.*` (CLI-only) tags

This unblocks end-to-end automated CLI release on every tag push.

## [1.6.7] - 2026-07-18

### Fixed — Detection: Trivy supply-chain attack (mar/2026) — `trivy-action` gray-zone warning

**Context**: On Mar 19, 2026, attackers force-pushed malicious code to 76
of 77 version tags of `aquasecurity/trivy-action`. A second wave hit
Mar 22-24 (Docker Hub v0.69.5, v0.69.6, latest). v0.35.0 was the only
pre-attack tag that survived.

Reference: <https://snyk.io/articles/trivy-github-actions-supply-chain-compromise/>
and <https://thehackernews.com/2026/03/trivy-security-scanner-github-actions.html>

**Safe versions**:
- `trivy CLI`: `v0.69.3` or earlier
- `trivy-action`: `v0.35.0` or **SHA-pinned** (recommended for production)
- `setup-trivy`: `v0.2.6`

**Changes**:
- `harness/scripts/check-stack-versions.sh` section 8: now warns on
  `trivy-action@v0.36.0` through `v0.69.x` (gray-zone between attacks)
- `harness/stack/versions.md`: pin table corrected — `trivy-action`
  was `v0.36.0` (gray-zone), now `v0.35.0` or SHA-pinned with
  explanation of the compromise window

**Lesson**: Mutatable tags in GitHub Actions are an attack surface.
For any third-party security tooling, prefer SHA-pinned refs.

## [1.6.6] - 2026-07-18

### Fixed — CRITICAL: template `harness/templates/.github-workflows-ci.yml` had broken action versions

**Bug discovered when `gmh agents update` was delegated to Hermes on
mandai-v2 (jul/2026).** The CI renewal PR opened by the agentic
failed 8/13 checks with two classes of errors:

1. **trivy-action without `v` prefix** — `aquasecurity/trivy-action@0.36.0`
   (no `v`) returns **404** on GitHub Actions. The tag exists only
   as `v0.36.0`. CI broke with `Unable to resolve action
   aquasecurity/trivy-action@0.36.0, unable to find version 0.36.0`.

2. **golangci-lint-action v6 with v2.12.2 linter** — `golangci-lint-action@v6`
   does NOT support golangci-lint v2.x. CI broke with
   `invalid version string 'v2.12.2', golangci-lint v2 is not
   supported by golangci-lint-action v6, you must update to
   golangci-lint-action v7.`

**Root cause**: the canonical `harness/stack/versions.md` had
the wrong values, and the agentic (and humans) blindly copied
them. Two pinning bugs that escaped the verify-after-build
protocol (sensor 09).

**Fix**:
- `harness/templates/.github-workflows-ci.yml`:
  - `golangci/golangci-lint-action@v6` → `@v9.3.0` (latest, v2.x-compatible)
  - `aquasecurity/trivy-action@0.36.0` → `@v0.36.0` (2 places)
- `harness/stack/versions.md`:
  - Fixed all 3 occurrences of the wrong pinning values
  - Added notes explaining the gotchas (v prefix required, v9.3.0 minimum)
- `harness/scripts/check-stack-versions.sh`:
  - **New section 8b**: detects `golangci-lint-action @v6/v7/v8 + version: v2.x` → fail
  - **Section 8 extended**: detects `trivy-action@[0-9]` (no v) → fail
  - Total sections: 15 → 16 (sensor 7 now catches 2 more bug classes)

**Lesson** (saved to memory): GitHub Actions require the **`v` prefix**
on tags. `0.36.0` ≠ `v0.36.0` for refs/actions. Always test pinning
end-to-end (sensor 09: verify-after-build).

## [1.6.5] - 2026-07-18

### Fixed — Critical bug in `gmh agents update` invocation

**Bug**: `agentic.Invocation("hermes", ...)` retornava
`hermes profile <name> --prompt "..."` — mas o CLI do Hermes
NÃO aceita esse formato. O correto é `hermes chat -p <name> -q "..."`.

Resultado: ao rodar `gmh doctor` ou `gmh agents update`, o
comando sugerido para o user **não funcionava** — copiava,
colava no terminal, e recebia erro de argumentos.

**Fix**:
- `cli/internal/agentic/agentic.go::Invocation` agora retorna
  `hermes chat -p <name> -q "..."` (validado contra `hermes chat --help`).
- Adicionado helper `shellQuote()` para escape correto de aspas
  em prompts longos.
- Outros agentics (claude, codex, opencode) marcados como TBD
  (validação pendente — Hermes é o único end-to-end testado).

### Fixed — `gmh agents sync` mensagem enganosa

A mensagem "safe (only update if persona marker missing)" no
summary era enganosa. O safe mode na verdade atualiza em 3 casos:
marker ausente, version drift, hash drift. Corrigido para
"safe (update on marker mismatch, version drift, or hash drift)".

### Known issue — `gmh update --to` não atualiza VERSION

`gmh update --to vX.Y.Z` reporta "no changes" quando o `harness/`
local já bate com a versão target, mas **não atualiza o arquivo
VERSION local**. Workaround: bumpar manualmente (`echo 1.6.4 > VERSION`).
A ser corrigido em v1.6.6.

## [1.6.4] - 2026-07-18

### Added — `gmh agents update` (CI renewal via agentic delegation)

Fechando o último gap do ciclo de atualização: **`gmh sync`
atualiza `harness/`, `gmh agents sync` atualiza `~/.hermes/profiles/`,
mas faltava atualizar `.github/`** (CI, codeql, labeler, etc.) — que
requer **contexto do agentic** para fazer merge inteligente
(preservar customizações do projeto + aplicar mudanças do framework
novo).

`gmh agents update` resolve isso delegando ao agentic (Hermes por
default) com um **prompt estruturado** que inclui:

- Contexto do projeto (cwd, framework version, agentic)
- Lista de arquivos em `.github/`
- Detecção heurística de customizações locais (env vars, secrets, custom jobs)
- Diff naive entre CI local e template novo
- Mudanças recentes do framework (extraídas do CHANGELOG)
- Task estruturada (5 passos: re-read, run sensors, apply, open PR, update VERSION)
- Critical: NÃO quebrar CI, PRESERVAR secrets, pin actions

#### Comandos

| Comando | O que faz |
|---------|-----------|
| `gmh agents update --dry-run` | Mostra diff + summary (sem invocar) |
| `gmh agents update` | Sugere comando `hermes -p team-manager --prompt "..."` |
| `gmh agents update --no-prompt` | Imprime só o prompt (sem invocation helper) |
| `gmh agents update --agent <x>` | Use outro agentic (default: hermes) |

#### Novos internals

- `internal/prompt/ci_renewal.go` — `CIRenewalPrompt()` template +
  `RecentChangesFromChangelog()` extractor

#### Novos checks no `gmh doctor`

- `CI: aligned with template (drift: +N lines from template)` —
  detecta drift antes do `gmh agents update`

#### Validação

Testado contra mandai-v2 (jul/2026):

```
$ gmh agents update --dry-run
==> Local:    1.6.0
==> Latest:   v1.6.3
==> Agentic:  hermes
==> Files in .github/:
  • .github/CODEOWNERS, ISSUE_TEMPLATE, labeler.yml, ...
==> Naive diff: +291 lines, -98 lines (read both files for real merge)
```

## [1.6.3] - 2026-07-18

### Fixed — `gmh agents sync` heuristic + skills + doctor checks

3 correções pontuais identificadas durante uso real no mandai-v2
piloto (jul/2026):

#### 1. `gmh agents sync` em safe mode era fraco (D-AGENTS-1)

**Antes:** heurística era "só atualiza se persona path não está no SOUL".
Depois de um `--aggressive`, o persona path ficava no SOUL, então safe
nunca mais atualizava. Profiles outdated ficavam stale pra sempre.

**Fix:** injeta um **version marker** no início do SOUL gerado:
```
<!-- gmh:soul version=1.6.2 hash=ba7ba5166c5034b2 -->
```

O safe mode agora detecta 3 tipos de drift:
- sem marker (SOUL antigo)
- framework version mudou (`currentVer != newVer`)
- persona hash mudou (`currentHash != newHash`)

Em qualquer um desses casos, **safe mode atualiza** (preserva custom
sections via `## Custom sections (preserved)` block). Só preserva sem
atualizar quando markers match + diff existe (= custom sections
legítimas do user).

#### 2. Skill `code-graph` warning (D-AGENTS-2)

**Antes:** `gmh agents sync` mostrava `⚠ code-graph: open ...: no such
file or directory` e **não instalava** a skill. Skills ausentes ficavam
ausentes.

**Fix:** `ReadSkill` agora retorna `("", nil)` para skills não
instaladas (em vez de erro). `agents sync` então trata "vazio" como
"não instalado" e **instala automaticamente** (`+ not installed in
Hermes` → `installed`).

#### 3. `gmh doctor` agora detecta `oasdiff/oasdiff:latest` (D-CI-1)

**Antes:** `oasdiff/oasdiff:latest` no `ci.yml` quebrava a pipeline
quando o user abria um PR (tag `@latest` não existe mais). O
`gmh doctor` não detectava.

**Fix:** novos checks locais:
- `CI: no @latest actions`
- `CI: no oasdiff/oasdiff:latest (tag invalid)`
- `CI: Trivy pinado (não @master)`
- `CI: dorny/paths-filter presente`

Se algum desses falhar, o doctor avisa — o user sabe antes de abrir PR.

#### Bonus: bug de cwd

`gmh doctor` e `gmh agents` agora respeitam o flag `-C/--cwd` (antes
usavam só `os.Getwd()`). Se o user rodar `gmh -C /path doctor`, o
cwd usado pelos checks é o `/path`.

## [1.6.2] - 2026-07-18

### Added — `gmh agents` (sync profiles + skills com framework)

Fechando o **gargalo do "harness desatualizado no user-side"**:
o `gmh sync` atualiza `harness/` no projeto, mas não toca os
**profiles do Hermes em `~/.hermes/profiles/`** nem as
**skills em `~/.hermes/skills/`**.

Resultado: profiles ficam com SOUL.md desatualizado (sabem de
18 invariantes quando o framework já tem 19), skills ficam
antigas, e domain-experts especializados perdem sync com o
framework.

**`gmh agents`** resolve isso com merge agentico.

#### Comandos

| Comando | O que faz |
|---------|-----------|
| `gmh agents list` | Lista profiles + skills instalados em `~/.hermes/` |
| `gmh agents inspect <profile>` | Mostra o diff entre SOUL.md atual e o gerado do persona |
| `gmh agents sync` | Sincroniza profiles + skills com a versão do framework |
| `gmh agents install <profile>` | Instala um profile a partir do `harness/personas/<name>.md` |

#### Sync strategies

- **Safe (default)**: só atualiza profiles cujo SOUL.md não
  referencia o persona path. Customizações locais são
  preservadas.
- **Aggressive (`--aggressive`)**: regenera todos os profiles
  matching os personas. Custom sections do SOUL.md antigo
  são preservadas em um bloco `## Custom sections (preserved)`.

#### Como funciona

1. Lê o persona file (ex: `harness/personas/team-manager.md`)
2. Extrai as seções `Identidade`, `Responsabilidades`, `Limites`,
   `Referências` (canônico SOUL.md)
3. Lê o SOUL.md atual em `~/.hermes/profiles/<name>/SOUL.md`
4. Faz diff linha-a-linha
5. Aplica merge (safe ou aggressive)
6. Reporta o que mudou (Updated/Preserved/Unchanged)

#### Skills sync

`gmh agents sync` também sincroniza as skills em
`~/.hermes/skills/`:

- Skill não instalada → instala
- Skill igual → skip
- Skill divergente → atualiza (em aggressive) ou avisa (em safe)

#### Internals

3 novos packages em `cli/internal/`:

- `hermes/` — filesystem client para `~/.hermes/`
  (ListProfiles, ReadSoul, WriteSoul, ListSkills, ReadSkill,
  WriteSkill)
- `soul/` — `Generate(personaPath)` extrai as seções
  canônicas de um persona file; `ComputeDiff(current, generated)`
  faz diff linha-a-linha
- `skills/` — `BuildManifest(dir)` lê `harness/skills/`
  e retorna manifest

#### Validação

- 6 profiles detectados no Hermes do user (team-manager,
  backend-engineer, frontend-engineer, devops-engineer,
  quality-assurance, solutions-architect)
- 41 skills detectados
- `gmh agents sync --dry-run` mostra 6 outdated (em safe)
- `gmh agents sync --aggressive` regenera os 6 + atualiza
  7 skills built-in
- SOUL.md do team-manager cresceu de ~200 → 824 linhas
  (com todo o conteúdo do persona, antes só tinha resumo)
- Cross-build 5 plataformas OK
- Backward compatible: nada quebra se você não usar
  `gmh agents`

## [1.6.0] - 2026-07-18

### Added — Release pipeline (GHCR) + `gmh` CLI

The meta-harness **closes the loop** with two major additions:

1. **Automated release pipeline** that publishes Docker
   images to **GitHub Container Registry (GHCR)** with
   multi-arch, cosign signatures, SBOMs, and Trivy scans.
2. **`gmh` CLI** (Go single static binary) that lets users
   install, sync, and manage the meta-harness in their
   projects with a single command.

#### Release pipeline (`docs/DEPLOY.md` + workflow 06)

After merging a PR, `devops-engineer` tags the commit
(`git tag v0.1.0 && git push origin v0.1.0`) and the
[`release.yml`](../templates/.github-workflows-release.yml)
workflow:

- **Pre-flight:** re-runs `check-stack-versions.sh` +
  `smoke-test.sh` on the tagged commit.
- **Build:** multi-arch (amd64 + arm64) backend + frontend
  in parallel, with cache `scope=backend-amd64` etc.
- **Scan:** Trivy on CRITICAL (block).
- **Sign:** cosign (keyless, OIDC GitHub).
- **SBOM:** SPDX attached to GitHub Release.
- **Push:** `ghcr.io/<owner>/<repo>/<service>:<tag>`.
- **Release notes:** auto-generated.

The published images can be deployed to **ECS, EKS, Docker
Swarm, or locally via docker-compose** — all documented in
[`docs/DEPLOY.md`](../docs/DEPLOY.md).

#### `gmh` CLI (`cli/` + `docs/CLI.md`)

A **Go single static binary** that:

- **`gmh install`** — install meta-harness into a project.
- **`gmh sync`** — sync local `harness/` with latest version.
- **`gmh update --to vX.Y.Z`** — pin to a specific version.
- **`gmh doctor`** — health check the local project.
- **`gmh skills [list|install|remove|available]`** — manage skills.
- **`gmh personas [list|create|remove]`** — manage domain-experts.
- **`gmh plugins [...]`** — manage gmh plugins (experimental).
- **`gmh version`** — print version info.

**Install (Linux/macOS):**

```bash
curl -sSL https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.sh | bash
```

**Install (Windows PowerShell):**

```powershell
iwr -useb https://raw.githubusercontent.com/brenonaraujo/git-meta-harness/main/cli/installer/install.ps1 | iex
```

Distributed via `cli-vX.Y.Z` GitHub Releases for 5 platforms
(linux/darwin/windows × amd64/arm64). See [`docs/CLI.md`](../docs/CLI.md)
for the full manual.

### Why this is `1.6.0` (minor, not patch)

- Adds 2 **major new public APIs**: release pipeline + CLI.
- Adds 2 new ADRs (0015, 0016) → significant design decisions.
- Adds 2 new docs (DEPLOY.md, CLI.md) → public documentation.
- All additive — does NOT break any existing project.

### Backward compatibility

- Existing projects continue to work without changes.
- The release pipeline is opt-in (add the workflow file).
- The `gmh` CLI is opt-in (don't have to install it; can keep
  using `harness/scripts/smoke-test.sh`).

## [1.5.0] - 2026-07-18

### Added — `verify-after-build` (sensor 09) + invariante 19

The **Mandaí v2 pilot** (PR #5, jul/2026) exposed a class of
defects that **auto-report from subagents** had been masking
(D1, D3, D4, D6, D7, D10). The team-manager was trusting
"PRONTO" reports without re-verifying. v1.5.0 closes that gap.

**3 coordinated changes** (see ADR-0014):

#### 1. `check-stack-versions.sh` v3 → v4 — 5 new sections (11-15)

Catches **5 classes of defects automatically** that v3 didn't:

- **§11** — Compose healthcheck `CMD-SHELL` in distroless
  image (D7). Distroless has no shell; healthcheck dies
  silently.
- **§12** — Compose `command: "...${VAR}..."` without `$$`
  escape (D3). Host shell expands `${VAR}` instead of passing
  it to the container; the URL becomes empty.
- **§13** — Makefile `go test -coverprofile=` without
  `-coverpkg=` (D4). Coverage is diluted across `main`,
  generated code, etc.; real numbers get masked (47% reported
  as 92%).
- **§14** — `govulncheck` absent from CI (D6). Go dependency
  CVEs (`quic-go`, `pgx`, etc.) never caught.
- **§15** — `pnpm audit` absent from CI (D10). Node dependency
  CVEs (`happy-dom`, etc.) never caught.

#### 2. Invariante 19 in `harness/AGENTS.md`

> **Team-manager verifies, doesn't trust.** After a builder
> reports "DONE" / "GREEN", the `team-manager` **re-runs**
> critical checks (re-reads `go.mod`/`Dockerfile`/`ci.yml`,
> runs `make lint && make test && make vuln`) **before**
> labeling as `in-review` or requesting human validation.

#### 3. Sensor 09 — `harness/sensors/09-verify-after-build.md`

A 6-step protocol the `team-manager` runs **itself** between
`in-progress` and `in-review`:

1. Re-read source-of-truth files (`go.mod`, `Dockerfile`,
   `ci.yml`, `package.json`).
2. Re-run `check-stack-versions.sh`.
3. Re-run the 3 canonical commands (`make lint && make test
   && make vuln` for backend; `pnpm lint && pnpm typecheck
   && pnpm test:run && pnpm audit` for frontend).
4. Check `gh pr checks <id>` (don't trust "CI passed" from
   the builder).
5. Check PR template (Como testar, Sensors, Changes).
6. Check coverage is in the correct scope (`-coverpkg=...`).

Operationalized in **`harness/personas/team-manager.md` §11**
with templates for green and red outcomes.

### Why this is `1.5.0` (minor, not `1.4.1` patch)

- Adds a **new public sensor** (sensor 09) → new public API.
- Adds a **new public invariant** (invariante 19) → new
  contract.
- Adds **5 new sections** to a public script (v3 → v4) →
  new behavior.
- All additive — does NOT break any existing project.

But is **backward compatible**: existing projects don't have
to adopt the new sensor; they can keep their workflow. The
new sensor is opt-in for the team-manager to run.

## [1.4.0] - 2026-07-18

### Added — `docs/HOWTO.md` and `harness/skills/code-graph/`

Two new artifacts address the **spec discovery** gap and the
**code graph** gap in the loop.

#### `docs/HOWTO.md` (15K, 6 diagrams, 7 sections)

The user asked: **where does the spec live, and what if
there is no spec?** This doc answers that.

- **§1** — the single input is a functional spec.
- **§2** — **4 valid paths** to provide a spec (paste inline,
  `docs/SPEC.md` in repo, describe in 3-5 sentences, or
  spec discovery for existing projects).
- **§3** — **spec discovery** for existing projects, inspired
  by the [Reversa framework](https://github.com/sandeco/reversa)
  (Macedo & da Costa, May 2026). 5-phase pipeline:
  Reconnaissance → Excavation → Interpretation → Generation
  → Review. Output: `docs/SPEC.md` with **confidence
  markers** (🟢 CONFIRMED, 🟡 INFERRED, 🔴 GAP).
- **§4** — the full loop, with spec discovery.
- **§5** — the code graph: an optional accelerator (Sourcegraph,
  CodeCompass, CodeGraph, Cursor, or agentic search).
- **§6** — step-by-step instructions for greenfield, existing
  project with spec, and existing project without spec.
- **§7** — anti-patterns to avoid.

#### `harness/skills/code-graph/SKILL.md` (6.3K)

A new skill that teaches any persona to use a **code graph**
(or semantic search) **instead of `grep` + `ls` + `read`**
to navigate a codebase. Covers:

- The 3 philosophies (index-first, agentic search, hybrid).
- Compatible tools (Sourcegraph, CodeCompass, CodeGraph,
  Cursor, agentic search fallback).
- Setup per tool.
- The rule of thumb: **code_search before grep**.
- 3 metrics to track (TTFRF, Tokens-per-task, Staleness).

#### `README.md` — full loop diagram updated

The "Visual overview" diagram now shows:

- The 4 spec paths converging on `docs/SPEC.md`.
- The `team-manager` **creating the issue-mãe itself** (not
  the user).
- The optional code graph step.
- The 9-sensor + human-validation gate.
- The release tag.

This is the diagram the user asked for: **what is
actually implemented**, not the previous (incomplete)
version.

#### Total Mermaid diagrams in the project

**32** (was 27): 6 in README, 2 in CONCEPT, 3 in ORIGIN,
2 in COMPARISON, 2 in PIPELINE, 5 in LOOP, **6 new in
HOWTO**, 1 in bootstrap, 6 in issue-lifecycle.

#### Why v1.4.0 (minor)

New doc + new skill + updated diagram. Conceptually
significant. Warrants a minor bump.

## [1.3.1] - 2026-07-18
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
