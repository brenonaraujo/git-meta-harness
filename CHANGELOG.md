# Changelog


## [1.13.0] - 2026-07-19

### Added — Feature Flow Enforcement (ADR-0025)

**Context**: Mandaí v2 (jul/2026, Épico #48 F7+F8+F10 — Avaliações
+ Reputação + Share) exposed a systemic problem: the `team-manager`
created the epic with `type/feature`, but **zero sub-issues went
through the canonic flow** (`domain-expert` → `solutions-architect` →
builder). Result: builders received only the 1-2 paragraph issue
description (no ACs, no DoD, no edge cases), implemented "in the dark",
and cost ~30min-1h of rework per sub-issue.

The framework **documented** the flow (AGENTS.md §3.1, team-manager
§4 Smart Routing, invariante 12) but had **no enforcement** — the
team-manager could skip `refined` and `ready` without being blocked.

**Solution (v1.13.0) — 6 coordinated changes**:

#### 1. Sensor 13 `feature-flow` (BLOCKING) + script

[`harness/scripts/check-feature-flow.sh`](harness/scripts/check-feature-flow.sh)
(3.2KB shell) +
[`harness/scripts/visual/check_feature_flow.py`](harness/scripts/visual/check_feature_flow.py)
(8.1KB Python).

**Detects 5 categories** in `type/feature` issues:
- `no_refined_label` — missing `refined` label (domain-expert didn't refine)
- `no_ready_label` — missing `ready` label (architect didn't DoD)
- `no_refinement_comment` — no comment with ACs + edge cases
- `no_dod_comment` — no comment with pillars + DoD
- `dod_without_refined` — architect ran before domain-expert

**BLOCKING** (exit 1) — prevents `in-progress` without proper flow.

**Usage**:
```bash
# All type/feature issues
./harness/scripts/check-feature-flow.sh

# One issue
./harness/scripts/check-feature-flow.sh 48

# Explicit repo
./harness/scripts/check-feature-flow.sh --repo owner/name 48
```

**Validation against Mandaí v2**:
```
$ ./harness/scripts/check-feature-flow.sh 48
BLOCKING: FEATURE FLOW VIOLATIONS (sensor 13, v1.13.0):

  Issue #48: [Épico] Avaliações + Reputação + Share (F7+F8+F10)
    ❌ no_refined_label
    ❌ no_ready_label
    ❌ no_refinement_comment
    ❌ no_dod_comment
```

#### 2. Comment templates (canonical, in `harness/templates/comments/`)

- **[`domain-expert-refinement.md`](harness/templates/comments/domain-expert-refinement.md)**
  (1.6KB) — mandatory format: Persona, Comportamento, Por que importa,
  ACs (mín 1), Edge cases (mín 1), Validação. Copy-paste ready.
- **[`solutions-architect-dod.md`](harness/templates/comments/solutions-architect-dod.md)**
  (2.6KB) — mandatory format: 3-5 Pilares, DoD checklist, Decisões
  (ADR-lite), Riscos, 12-factor audit. Copy-paste ready.

Sensor 13 validates heuristically (regex AC/EC/DoD/pilar) to ensure
the comment has the canonical content.

#### 3. AGENTS.md invariante 24 (NEW)

"Feature flow enforcement" — non-violable + blocking. Every
`type/feature` issue REQUIRES:

1. `refined` label applied by `domain-expert-<x>` (after posting
   refinement comment with ACs + edge cases)
2. `ready` label applied by `solutions-architect` (after posting
   DoD comment with 3-5 pillars)
3. Builder **reads ALL comments** (not just the description) before
   implementing, and references ACs/DoD in commits

#### 4. team-manager §3.1.3 (NEW)

"Feature flow enforcement" with:
- Canonical command before delegating to builder
- Recovery table per category (no_refined_label, no_ready_label, etc.)
- Edge cases: small sub-issues (recommend changing type), partial
  refinement (wait for architect), builder complains (refuse), builder
  pushes 3x with red sensor (escalate to user)

#### 5. Builder personas updated

- **`backend-engineer.md`** (responsibility 0, NEW) — "READ ALL
  COMMENTS OF THE ISSUE before implementing" (non-violable).
  If those comments don't exist, **STOP** — report to team-manager
  that the issue needs to go through the flow.
- **`frontend-engineer.md`** (responsibility 0a, NEW) — same rule.

#### 6. PR template updated

[`harness/templates/pr-description.md`](harness/templates/pr-description.md):

- New section: **"Context from domain-expert (v1.13.0+, invariante 24)"**
  — mandatory for `type/feature`. References refinement comment
  + lists which ACs are covered.
- New section: **"DoD from solutions-architect (v1.13.0+, invariante 24)"**
  — same, for DoD pillars.

Defense in depth: sensor 13 blocks flow; PR template ensures
builder actually read it.

### Validation

- ✅ Issue #48 of Mandaí v2 (bad case) detected: 4 violations
- ✅ Mock issue with proper flow (clean case) → exit 0
- ✅ Smoke-test still passes (43 passes + 1 local expected fail)
- ✅ AGENTS.md recognizes 24 invariantes (incl. new 24)

### Breaking changes

**None** (backward compatible):
- Existing flow works as before
- Sensor 13 only **adds** a new check (doesn't change existing
  labels or workflow)
- Comment templates are **suggested** (sensor validates
  content heuristically, doesn't enforce format strictly)
- PR template additions are optional for now (team-manager
  can ask for them per project)

### Migration guide

```bash
# 1. Pull latest meta-harness
cd your-project
gmh update --to v1.13.0

# 2. (Optional) Validate your existing type/feature issues
./harness/scripts/check-feature-flow.sh

# 3. (Optional) Add sensor 13 to CI
# In .github/workflows/ci.yml:
#   - name: Feature flow
#     run: ./harness/scripts/check-feature-flow.sh

# 4. Use the comment templates in your next feature:
#    - Copy harness/templates/comments/domain-expert-refinement.md
#    - Post comment on the issue, fill in, add label `refined`
#    - solutions-architect copies solutions-architect-dod.md
#    - Post DoD, add label `ready`
#    - Now builder can implement with full context
```


## [1.12.2] - 2026-07-19 (HOTFIX)

### Fixed — `gmh` agentic invocation: Hermes `-p` is a global flag

**Severity**: MEDIUM (delegação de issue pra persona errada / silenciosa)

**Context**: v1.6.5 introduced `agentic.Invocation()` to delegate
work to the agentic (Hermes). The comment said:

> Hermes: `hermes chat -p <profile> -q "<prompt>"`

That comment was **wrong**. The actual Hermes CLI (validated in
jul/2026) has `-p <profile>` as a **global flag on the `hermes`
root command**, NOT a flag of `hermes chat`. Running the buggy
form fails:

```
$ hermes chat -p team-manager -q "hello"
hermes: error: argument command: invalid choice:
"hello" (choose from 'chat', 'model', 'moa', ...)
```

The `chat` subcommand parser doesn't recognize `-p` and treats
`-p` and the profile name as the positional command argument.

**Correct form** (validated):

```
$ hermes -p team-manager chat -q "echo test"
Query: echo test
Initializing agent...
[executes as team-manager]
Resume this session with:
  hermes --resume 20260719_174540_572a30 -p team-manager
```

**Symptom observed** (Brenon, jul/2026, BRT): "o team manager
nao esta delegando corretamente para os profiles as issues via
hermes". Whenever `gmh doctor` (or any future delegating command)
called `Invocation(Hermes, profile, prompt)`, the resulting
command failed silently or hit the wrong profile.

**Root cause**: comment in `cli/internal/agentic/agentic.go`
documented the wrong invocation. The implementation followed
the comment. No live test validated the actual `hermes` CLI
syntax.

**Fix (v1.12.2)**:

```go
// Before (bug):
return fmt.Sprintf("hermes chat -p %s -q %s", profile, shellQuote(prompt)), nil

// After (fix):
return fmt.Sprintf("hermes -p %s chat -q %s", profile, shellQuote(prompt)), nil
```

Also fixed 2 examples in `harness/personas/team-manager.md` §6.6
that had the same buggy form.

**Regression test added**:
[`cli/internal/agentic/agentic_test.go`](cli/internal/agentic/agentic_test.go)
- `TestInvocation_Hermes_ProfileFlagBeforeSubcommand` — checks
  the output starts with `hermes -p <profile> chat` and **not**
  `hermes chat -p <profile>`. If `hermes` is on PATH, also
  validates `hermes --help` lists the `chat` subcommand.
- `TestInvocation_Hermes_DefaultProfile` — empty profile
  defaults to `team-manager`, still correct order.
- `TestInvocation_LongPromptIsShellQuoted` — single quotes in
  prompt are shell-escaped (Hermes invoked via shell).

All 3 pass; CI also runs `hermes --help` (if on PATH) to confirm
the expected CLI shape.

**Lesson** (see [ADR-0024](harness/contrib/design-decisions.md)):

> **Documentation in comments is not validation.** When a function
> produces a CLI command (especially for an external binary
> whose syntax may evolve), a unit test that asserts the output
> format is the **minimum**. Better: a live test that runs the
> command and checks the binary accepts it.

This is the second "agentic invocation" bug since v1.6.5
(see also v1.6.5 lesson on `hermes chat -p` vs
`hermes profile <name> --prompt`). Pattern: every time we
change agentic syntax, add a live test that runs the binary
with `--help` and parses the result.

**Migration**:

```bash
gmh update --to v1.12.2
# No manual config change needed — fix is in the CLI binary.
# Re-run gmh doctor to verify:
gmh doctor
```


## [1.12.1] - 2026-07-19 (HOTFIX)

### Fixed — `gmh agents sync` no longer erases `model`/`agent` config

**Severity**: HIGH (configuração do agent era sobrescrita em todo sync)

**Context**: v1.12.0 added `gmh agents sync` writing
`skills.external_dirs` to each profile's `config.yaml`. The
implementation unmarshaled the existing YAML into a typed struct
(`ProfileConfig`) that only knew about `skills`, then marshaled
it back, **silently erasing any other field the user had set**
(`model.default`, `model.provider`, `agent.reasoning_effort`,
custom keys, etc).

**Symptom**: After running `gmh agents sync` (which happens on
every `gmh update` to v1.12.0+), 4 of Brenon's profiles
(`team-manager`, `solutions-architect`, `quality-assurance`,
`devops-engineer`) lost their `model` and `agent` blocks. The
Hermes CLI then refused to start the profile (`missing required
field: model.default`).

**Root cause**: `WriteConfig` in
`cli/internal/hermes/hermes.go` used a typed struct (yaml.v3
`ProfileConfig`) for unmarshal **and** marshal. Any field not in
the struct was dropped on the way out.

**Fix (v1.12.1)** — `WriteConfig` now uses a generic
`map[string]interface{}` for both read and write. Only
`skills.external_dirs` is touched; everything else (model,
agent, custom keys, future fields) is preserved.

```go
// Before (bug):
cfg := &ProfileConfig{}
yaml.Unmarshal(data, cfg)  // only reads known fields
cfg.Skills = &ProfileSkills{ExternalDirs: externalDirs}
yaml.Marshal(cfg)          // writes back ONLY known fields
                           // → silently erases model/agent!

// After (fix):
root := map[string]interface{}{}
yaml.Unmarshal(data, &root)  // reads ALL fields as map
skills := root["skills"].(map[string]interface{})
skills["external_dirs"] = merged
yaml.Marshal(root)            // writes ALL fields back
                              // → preserves model/agent!
```

**Regression test added**:
[`cli/internal/hermes/hermes_test.go`](cli/internal/hermes/hermes_test.go)
(`TestWriteConfigPreservesModelAgent`, `TestWriteConfigCreatesWhenMissing`,
`TestWriteConfigIdempotent` — all 3 pass).

**Validation**:
- ✅ 3 unit tests pass (preserves model/agent, creates when missing,
  idempotent)
- ✅ Restored 4 profiles (manual edit) on Brenon's machine — all
  preserved through a real `gmh agents sync` run
- ✅ All 6 framework-controlled profiles (`team-manager`,
  `backend-engineer`, `frontend-engineer`, `solutions-architect`,
  `quality-assurance`, `devops-engineer`) keep their `model` and
  `agent` blocks

**Migration**:

```bash
# 1. Pull hotfix
cd your-project
gmh update --to v1.12.1

# 2. (If you were on v1.12.0) restore your model/agent config
#    manually — `gmh doctor` will warn if a profile is missing
#    `model.default`. Example config:
#
#    # ~/.hermes/profiles/team-manager/config.yaml
#    agent:
#      reasoning_effort: max
#    model:
#      default: MiniMax-M3
#      provider: minimax-oauth
#    skills:
#      external_dirs:
#      - ~/.hermes/skills
#
# 3. Re-run `gmh agents sync` to confirm config is preserved
gmh agents sync
cat ~/.hermes/profiles/team-manager/config.yaml   # model still there ✅
```

**Lesson** (see [ADR-0023](harness/contrib/design-decisions.md)):

> When a tool writes back to a file the user owns (config.yaml,
> `.env`, `package.json`, `SOUL.md`, anything that was hand-edited
> or set by the tool itself), **always use a generic map
> representation**, never a typed struct with a subset of fields.
> A typed struct is fine for *reading*, but the moment you
> *write*, you've committed to knowing every field — and a new
> field added by the agent later will silently nuke user data.

This bug class has a name: **"struct round-trip erasure"**. It
also affected `package.json`, `Cargo.toml`, `pyproject.toml`,
`.env`, etc in many tools. The fix is universal: marshal via
a generic map, not a typed struct.


## [1.12.0] - 2026-07-19

### Added — Frontend Public Skills + Cold-Start Polish (ADR-0022)

**Context**: Mandaí v2's PR #23 (Redesign Landing, jul/2026)
exposed 5 systemic problems in the `frontend-engineer` delivery:

| Anti-pattern | Example in PR #23 | Detector |
|---|---|---|
| **Hex colors hardcoded** | `background: #ecfdf5`, `color: #064e3b` | regex `#([0-9a-fA-F]{3,8})\b` |
| **CSS BEM** | `.home-hero__title`, `.home-hero__cta` | regex `\.[a-z][a-z0-9-]*__[a-z]` |
| **Redundant comments** | `// HomeHero — top of the public landing page...` | heuristic: comment repeats component name |
| **Excessive emojis** | 4+ decorative emojis in form/dashboard | Unicode emoji regex, threshold-based |
| **Off-scale spacing** | `p-3`, `gap-5`, `mt-7` | regex filtering against allowed scale |

**Root cause triple**:
- The meta-harness **didn't document** the public skills registry
  (`https://www.skills.sh`, `npx skills`).
- The meta-harness **had no mechanism** to block visual anti-patterns
  (only sensor 08-i18n-audit, which doesn't cover UI quality).
- The `frontend-engineer` persona **had no explicit rule** "consult
  public skills BEFORE implementing UI".

**Solution (v1.12.0) — 7 coordinated changes**:

#### 1. Skill `frontend-public-skills` (NEW, 10.5KB)

Documents the registry workflow:

```bash
# 1. Identify stack
grep "@nuxt/ui" web/package.json

# 2. Consult registry
npx skills find nuxt-ui
# → nuxt/ui@nuxt-ui              15.2K installs (oficial)
# → onmax/nuxt-skills@reka-ui   6.6K
# → onmax/nuxt-skills@nuxt-ui   6.1K

# 3. Install official skill
npx skills add nuxt/ui@nuxt-ui

# 4. (Optional) Setup MCP for runtime component API
claude mcp add --transport http nuxt-ui https://ui.nuxt.com/mcp
```

Includes:
- Curated lists by stack (Nuxt UI, Tailwind-only, shadcn, Vue,
  Visual, Playwright)
- Security validation (how to check `/security` page)
- MCP server setup (Nuxt UI has one)
- Self-check pre-implementation
- Anti-pattern documentation (hardcoded colors, BEM, etc)

#### 2. Skill `tailwind-only-patterns` (NEW, 9KB)

For projects **without** `@nuxt/ui`:
- Tailwind v4 CSS-first config (`@theme`)
- shadcn-vue, PrimeVue, Reka UI standalone
- Decision tree: when to use which library
- Token system (`bg-bg-elevated text-fg border-border`)
- Anti-patterns (hex hardcoded, `@apply` excess, BEM mixing)

#### 3. Skill `visual-polish` (NEW, 12.3KB)

Stack-agnostic techniques:
- **Hierarchy**: H1 36-48px, H2 24-30px, H3 20px (modular scale)
- **Whitespace**: scale 4/8/12/16/24/32/48/64/96 (no `p-3` or `gap-5`)
- **Contrast (WCAG AA)**: ≥ 4.5:1 text, ≥ 3:1 UI components
- **Consistency**: same variants, sizes, padding across project
- **Motion**: 200-300ms sweet spot, animate `transform`/`opacity`
- **Touch targets**: ≥ 44×44px (Apple HIG / Material)
- **5-second self-check**: "Would I pay for this app? Looks like
  Linear/Notion/Vercel?"

#### 4. Skill `nuxt-ui-patterns` v2.0.0 (UPDATED, +10KB)

- Frontmatter: `nuxt-ui-v3` → `nuxt-ui-v4` (Mandaí v2 uses
  `@nuxt/ui@^4.10.0`)
- New section: "Public Skills Registry" (npx skills, MCP)
- New section: "Anti-patterns" (5 categories with good/bad examples)
- Expanded self-check (npx skills, sensor 12, screenshot, etc)

#### 5. Sensor 12 `frontend-polish` (NEW, 11.7KB) + Python script

[`harness/scripts/check-frontend-polish.sh`](harness/scripts/check-frontend-polish.sh)
(3KB shell) +
[`harness/scripts/visual/check_frontend_polish.py`](harness/scripts/visual/check_frontend_polish.py)
(8.5KB Python companion).

**Detects 10 categories**:
- `hardcoded_colors` (hex/rgb/hsl in components, not in tokens)
- `bem_naming` (`.foo__bar`, `.foo--bar`)
- `redundant_comment` (comment repeats component name)
- `emojis_excessive` (> 3 in any file, > 1 in serious components)
- `spacing_off_scale` (`p-3`, `gap-5`, `mt-7` — values not in scale)
- `inline_color_style` (`style="color: #..."`)
- `off_stack_imports` (bootstrap in Nuxt UI project)
- `img_no_alt` (`<img>` without `alt`)
- `button_no_text` (`<button></button>` without text/aria-label)
- `no_design_system` (component with `<style>` but no `var(--ui-*)`)

**BLOCKING** (exit 1) — different from sensor 11 (warning-only).
Rationale: refactor is trivial (< 5min) but cold-start poor
quality costs expensive rework.

**Validates**:
- ✅ Clean components (Nuxt UI templates) → exit 0
- ❌ Mandaí v2 PR #23 (hex hardcoded, BEM, redundant comments) →
   exit 1 with 22 issues + recovery

#### 6. Templates Nuxt UI (NEW, 3 files, ~10KB)

[`harness/templates/nuxt-ui/`](harness/templates/nuxt-ui/):
- **`landing.vue`** (3.6KB) — hero + features + CTA + footer, all
  with `text-highlighted text-muted bg-elevated` tokens, no BEM,
  no hex.
- **`dashboard.vue`** (2.5KB) — admin panel with stats cards
  (3 trend variants), `UDashboardPage` + `UDashboardNavbar`.
- **`auth-form.vue`** (3.6KB) — login/signup reusable, role
  preselect via `?role=leader|resident|supplier`, i18n ready.

All 3 pass sensor 12 (`OK: No frontend polish issues detected`).

#### 7. Visual Report (in `quality-assurance.md`)

New responsibility 4.1 (v1.12.0):
- Generate screenshots via Playwright (3 viewports: 375/768/1440)
- For each route new/changed in the PR
- Save in `qa/screenshots/<route>-<viewport>.png`
- Visual checklist (hierarchy, whitespace, contrast, consistency,
  responsive)
- Save report in `qa/visual-report-<pr>.md`
- **Block** if sensor 12 fails

#### 8. Playwright scripts (NEW, 2 files, ~6.4KB)

[`harness/scripts/visual/playwright-screenshot.mjs`](harness/scripts/visual/playwright-screenshot.mjs)
(3.7KB) — Node script, `pnpm screenshot` after setup.

[`harness/scripts/visual/setup-playwright-screenshot.sh`](harness/scripts/visual/setup-playwright-screenshot.sh)
(2.6KB) — installs Playwright + Chromium + adds `package.json`
scripts (`screenshot`, `screenshot:setup`). Idempotent.

#### 9. AGENTS.md invariante 23 (NEW)

> **Frontend polish (cold-start visual)** (v1.12.0, ADR-0022).
> Lição do Mandaí v2 (jul/2026, PR #23): o `frontend-engineer`
> entregou UI com cores hex hardcoded, CSS BEM misturado com
> Tailwind, comentários redundantes, emojis excessivos, e zero
> uso de skills públicas. Resultado: tela com cara de
> "W3Schools 2018" em vez de marketplace profissional.

Rule of thumb:
- **`frontend-engineer` MUST consult `npx skills find <stack>`**
  before writing any `.vue`/`.css`
- **Respect project design tokens** (zero hex, zero BEM mixed)
- **Screenshot local BEFORE PR** (Playwright)
- **Sensor 12 BLOCKS** (exit 1) on anti-patterns

#### 10. team-manager §13 (NEW)

"Frontend polish (sensor 12 — você BLOQUEIA)":
- 10 categories that block
- 4 moments to run (local frontend / CI / PR review / Visual Report)
- Recovery actions per category
- Difference from sensor 11 (blocking vs recommendation)

#### 11. ADR-0022 (NEW)

Documents the decision: 3 root causes, 7 changes, principles,
cost avoided (~30h/year), validation (4 test cases), lessons.

### Personas updated

- **`frontend-engineer.md`**: rules #0 (npx skills) and #13
  (screenshot local), rule #14 (respect design tokens), §"Skills"
  expanded to v1.12.0 with 3 new skills
- **`quality-assurance.md`**: Visual Report responsibility, §"Skills"
  includes `visual-polish` and `frontend-public-skills`
- **`team-manager.md`**: §13 (Frontend polish), §"Skills" includes
  `frontend-public-skills`

### Total: 11 coordinated changes

- 3 new skills (~32KB)
- 1 updated skill (nuxt-ui-patterns v2.0.0, +10KB)
- 1 new sensor (12) + Python companion (~20KB)
- 3 templates Nuxt UI (~10KB)
- 2 Playwright scripts (~6.4KB)
- 1 ADR (~5KB)
- 1 AGENTS.md invariante
- 3 personas updated
- 1 CHANGELOG entry

**Total**: ~100KB of framework. **Cost avoided**: ~30h/year
(rework polish + QA explanation + cold-start amortization).

### Validation

- ✅ 3 Nuxt UI templates pass sensor 12 (`OK: No frontend polish
   issues detected`)
- ✅ Mandaí v2 PR #23 components FAIL sensor 12 (22 issues:
   18 BEM + 3 redundant_comment + 1 hardcoded_colors)
- ✅ Clean test case (CleanLanding.vue) exits 0
- ✅ Smoke-test passes 43/44 (1 fail is local: Hermes has
   generic domain-expert profile, expected in Brenon's env)
- ✅ AGENTS.md recognized 23 invariantes (incl. new 23)

### Breaking changes

**None** (backward compatible):
- Existing projects that use BEM can whitelist via
  `package.json` → `meta-harness.sensors.frontend-polish.whitelist`
- Existing nuxt-ui-patterns users get v2.0.0 transparently
  (npx skills workflow is new, rest is additive)

### Migration guide

```bash
# 1. Pull latest meta-harness
cd your-project
gmh update --to v1.12.0

# 2. (Optional) Install Playwright
bash harness/scripts/visual/setup-playwright-screenshot.sh

# 3. (Optional) Install public skill for your stack
npx skills add nuxt/ui@nuxt-ui  # or wshobson/agents@tailwind-design-system

# 4. (Optional) Setup CI job for sensor 12
# Add to .github/workflows/ci.yml:
#   - name: Frontend polish
#     run: ./harness/scripts/check-frontend-polish.sh

# 5. Existing projects with BEM: whitelist
# package.json:
#   "meta-harness": {
#     "sensors": {
#       "frontend-polish": { "whitelist": ["bem_naming"] }
#     }
#   }
```

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [1.11.0] - 2026-07-19

### Added — Scope Discipline (PILARES vs BLUEPRINTS) (ADR-0021)

**Context**: Mandaí v2's Épico F4+F5 (Ciclos + Pedidos)
showed the `domain-expert` and `solutions-architect` writing
**blueprints** (function names, SQL, paths, ORMs) instead of
**pillars** (what + why). The `backend-engineer` became a
**blind executor** — followed the blueprint without
questioning, without optimizing, without technical ownership.
Cost: ~3-5h of rework per blueprint mistake.

**3 coordinated changes**:

1. **Skill `solution-scoping` (NEW, 12KB)** — codifies the
   PILARES vs BLUEPRINTS principle with:
   - 6 categories with good (pillar) vs bad (blueprint) examples:
     pricing, limits, state machine, idempotency, compliance,
     slug uniqueness
   - Per-persona rules (DO vs DON'T)
   - Detection heuristics (regex patterns)
   - Recommended limits (non-blocking)
   - Pre-post checklist (domain-expert + solutions-architect)

2. **Sensor 11 `scope-discipline` (NEW)** +
   [`harness/scripts/check-scope-discipline.sh`](harness/scripts/check-scope-discipline.sh)
   (3.7KB):
   - Detects 10 patterns via regex (SQL keywords, ORM names,
     paths, function names, migrations, endpoints, prometheus,
     tokens)
   - **Non-blocking** (different from sensors 04-verify and
     10-decomposition-safety which block) — emits
     **recommendation** (warning) to shorten on next iteration
   - Per-persona thresholds: `domain-expert` ≥ 1 (zero
     tolerance for tech), `solutions-architect` ≥ 2-5
     (more permissive, can mention pinned stack)
   - **Does NOT truncate** — output passes through unchanged
   - 4 validated test cases (clean + leaked for both personas)

3. **`team-manager.md` §12 (NEW)** — "Scope discipline" with:
   - Principle (PILARES vs BLUEPRINTS)
   - 3-step protocol (run sensor → interpret → decide)
   - Reformulation template
   - When to skip (output is just checklist or ADR)
   - Recommended limits
   - Who detects / applies

4. **AGENTS.md invariante 22 (NEW)** — scope discipline as
   non-violable (but non-blocking). Aligned with invariants
   20/21 (non-violable + blocking) but inverted: non-violable
   + non-blocking.

5. **Reforced fences** in:
   - [`harness/personas/domain-expert.template.md`](harness/personas/domain-expert.template.md):
     added "🚧 Cerca de Solução — você NÃO fala de IMPLEMENTAÇÃO"
   - [`harness/personas/solutions-architect.md`](harness/personas/solutions-architect.md):
     replaced DoD with "PILARES, não BLUEPRINTS" + 3-5 pilares + DoD macro

6. **ADR-0021** (full decision documentation with 6 lessons
   from Mandaí v2's F4+F5 incident).

7. **CHANGELOG + VERSION** (1.10.3 → 1.11.0).

**Non-blocking by design** (per user request): the
recommendation from sensor 11 is **advisory only** — the
team-manager decides whether to request reformulation or
accept. This avoids the "blocker tax" of false positives
while still surfacing the issue.

**Validated test cases** (4):
1. Clean `domain-expert` output → ✅ no issues
2. Leaked `domain-expert` (SQL, ORMs, paths) → ⚠️ 6 signals
3. Clean `solutions-architect` (4 pillars) → ✅ no issues
4. Leaked `solutions-architect` (functions, SQL, paths) → ⚠️ 6 signals

**Limits (non-blocking recommendations)**:
- `domain-expert`: ≤ 12 ACs, ≤ 8 edge cases, ≤ 3k tokens
- `solutions-architect`: ≤ 5 pillars, DoD ≤ 80 lines, ≤ 5k tokens
- **>30k tokens** (75k chars): warning (anyone)

**Cost avoided**: ~3-5h/épico × 4 épicos/month = ~12-20h/month
of rework avoided by giving builder autonomy instead of
blueprints.

## [1.10.3] - 2026-07-19

### Fixed — Hermes desktop UI shows 0 skills (HOTFIX)

**Context**: After v1.10.2, the user reported that even
though the runtime saw 115 skills (12 harness + 103 Hermes
global), the Hermes desktop UI showed 0 for most profiles
(e.g., `backend-engineer: Skills: 0`).

**Root cause**: The Hermes desktop UI's `_count_skills`
function uses `skills_dir.rglob("SKILL.md")` — it counts
ONLY physical skills in each profile's `skills/` directory.
It does **NOT** count skills from `external_dirs` (which
still work at runtime, but the UI shows 0).

**Profiles affected** (before this fix):
- `domain-expert`: 73 skills (the 73 bundled from Hermes)
- `team-manager`: 1 skill (software-development, manually copied)
- `backend-engineer`, `frontend-engineer`, `solutions-architect`,
  `quality-assurance`, `devops-engineer`, `domain-expert-mandai`:
  **0 skills** (skills/ empty after `external_dirs` migration)

**Solution (v1.10.3 hotfix)**:

1. **New function** in [`cli/internal/hermes/hermes.go`](cli/internal/hermes/hermes.go):
   - `WriteProfileSkill(profileName, skillName, content)` —
     writes to `~/.hermes/profiles/<name>/skills/<skill>/SKILL.md`
     (so the UI counts it).

2. **`gmh agents sync` now copies harness skills into each
   profile** (v1.10.3) — the 12 harness skills (not the 73+
   Hermes global catalog, which remain only in the global
   `~/.hermes/skills/` and are reachable at runtime via
   `external_dirs`).

3. **Strategy: duplication** (not replacement) — skills live
   in 2 places:
   - `~/.hermes/skills/<skill>/SKILL.md` (global, via `external_dirs`)
   - `~/.hermes/profiles/<name>/skills/<skill>/SKILL.md` (per profile, for UI)

   This is the **meio-termo** (middle-ground) approach: the
   12 skills that change with the framework are duplicated
   (so the UI works); the 73+ Hermes static skills stay
   only in the global catalog (storage optimization).

**Validado** (mandai-v2, 7 profiles):
- team-manager: 4 skills (3 harness + 1 software-development)
- backend-engineer: 9 skills (3 harness + 6 mais)
- frontend-engineer: 8 skills
- solutions-architect: 6 skills
- quality-assurance: 5 skills
- devops-engineer: 3 skills
- domain-expert-mandai: 3 skills

**Trade-off**:
- **+**: UI shows correct count, runtime still sees full
  115-skill catalog, sync is fully automatic.
- **−**: ~2× storage of harness skills (12 skills × 7 profiles
  + 12 global = 96 arquivos instead of 12). Acceptable.
- **−**: Sync is slightly slower (~5s for 7 profiles × 12 skills).

**For other projects com Hermes**: rodando `gmh agents sync`
agora popula `skills/` em cada profile automaticamente. UI
do Hermes desktop mostra a contagem correta sem edição manual.

**Cost avoided**: ~5 min × 7 profiles = ~35 min de cópia manual
por setup de projeto novo + sincronização a cada release do
framework (que era manual e propensa a drift).

## [1.10.2] - 2026-07-19

### Added — `gmh agents sync` writes `config.yaml` with `skills.external_dirs`

**Context**: After v1.10.1, the user reported that profiles
were seeing only the 12 harness skills (not the 73+ skills from
the Hermes global catalog). Root cause: Hermes profiles
created with `hermes profile create` (default) bundle 73+ skills
but **don't** include the user's `~/.hermes/skills/` directory.
Profiles that want to see both must have
`skills.external_dirs: ["~/.hermes/skills"]` in their `config.yaml`.

Previously, `gmh agents sync` only updated `SOUL.md` — it
didn't touch `config.yaml`. Users had to manually edit each
profile's `config.yaml` to add `external_dirs`, which was
error-prone and easy to forget.

**v1.10.2 (hotfix)** automates this:

1. **New functions** in [`cli/internal/hermes/hermes.go`](cli/internal/hermes/hermes.go):
   - `ReadConfig(profileName)` — reads `config.yaml` (or returns
     empty `ProfileConfig` if not found).
   - `WriteConfig(profileName, externalDirs)` — writes `config.yaml`
     preserving other fields.
   - `EnsureExternalDirs(profileName, dirs)` — adds dirs to existing
     `external_dirs` (deduped, preserved order). Idempotent.

2. **`gmh agents sync` now writes `config.yaml`** via
   `ensureProfileExternalDirs()` after every sync. This is
   called in **all** sync paths:
   - Fresh install (no SOUL.md yet)
   - Stale persona (updated)
   - Customizations preserved
   - Skipped (identical)
   - Aggressive regenerate

3. **`gmh agents install <profile>` now also**:
   - Runs `hermes profile create <name> --no-skills` (best-effort;
     warns if hermes not on PATH).
   - Writes `config.yaml` with `skills.external_dirs: ["~/.hermes/skills"]`.

### Added — `## Skills` section in every persona template

[`harness/personas/`](harness/personas/) (v1.10.2):
- `team-manager.md` — 8 skills prioritárias
- `backend-engineer.md` — 6 skills prioritárias
- `frontend-engineer.md` — 5 skills prioritárias
- `solutions-architect.md` — 6 skills prioritárias
- `quality-assurance.md` — 6 skills prioritárias
- `devops-engineer.md` — 4 skills prioritárias
- `domain-expert.template.md` — 4 skills prioritárias (adapta
  por domínio)

Each section has a 3-column table: `Skill | Quando usar | Por quê`.

### Changed — `SoulSections` includes `Skills`

[`cli/internal/soul/soul.go`](cli/internal/soul/soul.go):
- Added `"Skills"` to `SoulSections` so the new `## Skills`
  section is included in the generated `SOUL.md`.
- Existing customizations on `Skills` section are preserved
  (same as other canonical sections).

### Validated scenario (mandai-v2)

After `gmh agents sync`, every profile's `config.yaml` has
`skills.external_dirs: ["~/.hermes/skills"]`. Profiles see
**115 skills** (12 harness + ~103 Hermes global catalog).

## [1.10.1] - 2026-07-19
## [1.10.1] - 2026-07-19

### Fixed — `bin/safe-commit-harness-sync.sh` (HOTFIX)

**Context**: In v1.8.0, v1.9.0, and v1.10.0, `git add -A`
in main of `mandai-v2` (and other client projects) repeatedly
captured untracked files from feature branches that `git
checkout` does NOT remove (only modifies tracked files). This
polluted harness sync commits with feature work — 3 separate
incidents, ~15 min of recovery each.

**Root cause**: documenting the rule "don't use `git add -A`"
didn't prevent the bug (3rd time proved it). The only solution
is **automation** — a helper that does the right thing
without manual discipline.

**Added**: [`bin/safe-commit-harness-sync.sh`](bin/safe-commit-harness-sync.sh)
(6.4KB) — automates `git add` + `git commit` + `git push`
for harness syncs in main, **without `git add -A`**.

**Behavior**:
1. ALWAYS stages `harness/` and `VERSION` (framework sync).
2. Auto-detects whitelisted local customizations
   (`.golangci.yml`, `.github/workflows/*.yml`,
   `.markdownlint.json`, `Makefile`, `docker-compose*.yml`,
   `deploy/*.sh`, `docs/HOWTO*.md`).
3. **BLOCKS** the commit if there are non-whitelisted
   modifications/untracked files, with a list of the
   suspicious files and 3 recovery options (stash,
   checkout feature, re-evaluate).
4. Asks for confirmation before committing
   (`--auto` to skip for CI).
5. Shows `git diff --cached --stat` before commit.

**Exit codes**:
- `0` = success (or clean dry-run)
- `1` = blocked (suspicious files found)
- `2` = git error

**Usage**:
```bash
./bin/safe-commit-harness-sync.sh                    # add + commit + push
./bin/safe-commit-harness-sync.sh --dry-run          # preview
./bin/safe-commit-harness-sync.sh --no-push          # only add + commit
./bin/safe-commit-harness-sync.sh --message "msg"    # custom commit msg
./bin/safe-commit-harness-sync.sh --auto             # skip confirmations
```

**Validated** (3 scenarios):
1. **Happy path** (only `harness/` + `VERSION` changed) → ✅
   stages 2 paths, proceeds.
2. **Block** (untracked `web/fake_feature.ts` and
   `backend/internal/fake/`) → 🛑 exit 1 with 2 file list
   and recovery options.
3. **Whitelisted** (`.golangci.yml` modified) → ✅ stages 3
   paths, `.golangci.yml` auto-added.

**Customizing the whitelist**: edit `WHITELIST_REGEX` in the
script. Paths outside the whitelist are **blocked**, not
silently added.

### Changed

- [`harness/bootstrap.md`](harness/bootstrap.md): added §5b
  "Helper `bin/safe-commit-harness-sync.sh`" with usage and
  customization guide.
- [`bin/.gitignore`](bin/.gitignore): added (NEW, ignores
  `*.local` customizations).

## [1.10.0] - 2026-07-18

### Changed — Function limit 25 → 35 lines (recommended: 25)

**Context**: Mandaí v2's Épico #12 backend builder reported
"OnboardRole has 34 lines (limit 25). I'll refactor by
extracting to a helper function." This is the classic
**"split for compliance"** anti-pattern: a cohesive 34-line
function was mechanically decomposed into 4 functions + 1
delegating helper, adding glue code without readability
gain. v1.10.0 raises the hard limit from 25 → 35 lines
(recommended: 25 unchanged) and adds a new skill to force
builders to **think about abstraction before coding**.

**Files updated**:
- [`harness/templates/.golangci.yml`](harness/templates/.golangci.yml):
  `funlen { lines: 35, statements: 30 }` (was 25/20)
- [`harness/sensors/00-static-analysis.md`](harness/sensors/00-static-analysis.md):
  table + remediation updated
- [`harness/stack/code-style.md`](harness/stack/code-style.md):
  §"Funções / Tamanho" with 3-band table (0-25/26-35/36+)
- [`harness/stack/backend.md`](harness/stack/backend.md):
  handler/ comment + anti-pattern
- [`harness/bootstrap.md`](harness/bootstrap.md): limits + table
- [`harness/CLAUDE.md`](harness/CLAUDE.md): limits
- [`harness/personas/solutions-architect.md`](harness/personas/solutions-architect.md):
  DoD (25 recommended, 35 max, justification required)
- [`harness/personas/backend-engineer.md`](harness/personas/backend-engineer.md):
  §4 (limits) + §5 (new: pre-implementation-design)
- [`harness/personas/frontend-engineer.md`](harness/personas/frontend-engineer.md):
  §5 (limits + skill reference)
- [`harness/team-manager.md`](harness/personas/team-manager.md):
  §5.3 (hermes skills install — added pre-implementation-design)
- [`harness/AGENTS.md`](harness/AGENTS.md): invariante 9a

### Added — Skill `pre-implementation-design` (v1.10.0)

[`harness/skills/pre-implementation-design/SKILL.md`](harness/skills/pre-implementation-design/SKILL.md)
(8.3KB) — forces the builder to **list 2-3 possible
decompositions BEFORE coding** and justify the choice:

```markdown
## Decomposition of `OnboardRole(user, role, tenantID)`

### Option A — Single function (32 lines) [chosen]
### Option B — 4 helpers (rejected)
### Option C — 2 helpers (rejected)

### Choice: A
**Why**: atomic transaction. B's helpers would fragment
reading without real gain.

**When would revert**: if `audit` becomes a separate
team's compliance (LGPD, BACEN), then extract.
```

Includes:
- 🚦 Traffic-light rules (when to apply)
- 📋 "3 decompositions" protocol (4 steps)
- 🧠 Heuristics: "1 function" vs "decompose" (4 + 5 scenarios)
- 🚫 Anti-patterns the skill ELIMINATES:
  - Split for compliance
  - Empty helper
  - Mega-function without abstraction

### Forward-compatible with v2.0.0

- v2.0.0 will add **worktree isolation** (per builder) +
  **WIP commits** (incremental persistence).
- v2.0.0+ may add **post-impl review** skill (separate
  from pre-impl) if needed.

## [1.9.0] - 2026-07-18

### Added — Decomposition Safety (path-scope + depends-on + sensor 10) (ADR-0019)

**Context**: The Épico #12 (auth + role switching) of the Mandaí
v2 validation project was decomposed into 6 sub-issues (#13–#18)
that were dispatched in parallel by 6 builders sharing the same
`cwd`. Backend #13 (auth-api) and #15 (user-role) both declared
the `UserRepository` interface in the same package
(`internal/repository/`) — compile conflict. **None of the 6
builders committed** — work was lost (~4h wasted).

**4 coordinated changes**:

1. **`solutions-architect.md` §Path scoping (NEW, required)**:
   - Every sub-issue from a decomposition **MUST** declare 1+
     `path-scope: <glob>` label in the DoD.
   - Glob syntax (same as `.gitignore` / `find -path`).
   - Glob rules of thumb + edge cases (lock files, migrations).
   - "When to serialize" table with 6 scenarios.

2. **Sensor 10 — Decomposition Safety (NEW)**:
   - [`harness/sensors/10-decomposition-safety.md`](harness/sensors/10-decomposition-safety.md)
     — protocol for the `team-manager` to detect path-scope
     overlap before dispatching builders.
   - [`harness/scripts/check-parallel-builders.sh`](harness/scripts/check-parallel-builders.sh)
     — automated bash script (Python-backed glob overlap
     heuristic) with 3 exit codes (0 = OK, 1 = overlap, 2 = no
     path-scope).

3. **`team-manager.md` §6 "Decomposition Safety" (NEW)**:
   - Step-by-step protocol: read path-scopes → calculate
     overlap → block or accept.
   - Concrete example from Épico #12 (before/after).
   - "Good vs Bad" behavior examples.
   - Edge cases (lock files, migrations, file deletion).

4. **AGENTS.md invariante 21 (NEW)**:
   - `path-scope` + `depends-on` + sensor 10 = non-violable.
   - Sub-issue without path-scope = DoD incomplete.
   - Sub-issue covering `go.mod`, `package.json`, lock files
     = always serialize.

5. **`workflow/05-orchestration.md` §2 expanded**:
   - "Parallelize what fits" now references sensor 10 + Épico
     #12 lesson explicitly.

6. **Canonical labels**:
   - `path-scope: <glob>` (1+ per sub-issue) — declared in DoD.
   - `depends-on: #X` (1+ per sub-issue) — explicit
     serialization (GitHub renders as native blocker with
     [Blocked PRs](https://github.com/settings/blocked_prs) app).

### Changed

- `team-manager.md` §7 (Comportamento esperado): added reference
  to §6 "Decomposition Safety".
- `team-manager.md` §10 (Limites): added "Não dispara 2+ builders
  em paralelo sem rodar sensor 10".

### Forward-compatible with v2.0.0

- v2.0.0 will add **worktree isolation** (`git worktree add`
  per builder) + **WIP commits** (incremental persistence).
- path-scope remains useful even with worktree (declarative
  intent, regardless of where work happens).

## [1.8.0] - 2026-07-18

### Added — Cerca Técnica (espelho da Cerca de Design) + skill `domain-refinement` (ADR-0018)

**Context**: The domain-expert was being called to refine **purely
technical** issues (Helm chart config, PostgreSQL index, Trivy action
SHA) and was directing implementation ("`gorm.Model`", "PostgreSQL",
"Redis TTL 5min", "OAuth2 + PKCE"). The team-manager was also routing
`type/technical` / `type/infra` / `type/tech-debt` / `type/docs` /
`type/ui` issues to the domain-expert — issues that **should skip**
domain-expert (no business rule to refine).

**Skill added** (in `harness/skills/`):
- `domain-refinement/SKILL.md` (9.7KB) — for the `domain-expert`:
  - **Cerca #0**: Você é o **POR QUÊ**, não o **COMO** (camadas
    Negócio → Design → Arquitetura → Implementação)
  - **Cerca #1**: Domínio fala em **comportamento**, técnico fala
    em **mecanismo** (tabela de transformação completa)
  - **Cerca #2**: Quando o tipo é `type/technical`, `type/infra`,
    `type/tech-debt`, `type/docs`, `type/ui` você **NÃO** é acionado
  - **Cerca #3**: Não mencione personas pelo nome
  - **Cerca #4**: Não feche issues, não crie branches, não escreva
    código
  - **Teste "e se a stack mudar?"** — toda AC deve sobreviver à
    troca de stack (Go → Rust, Nuxt → React, PostgreSQL → MongoDB,
    REST → GraphQL)
  - **Checklist pré-post** com 9 itens verificáveis

### Added — Cerca Técnica no `domain-expert.template.md`

Domain-expert now **fences itself** from technology (symmetric to
the Design fence added in v1.7.0):

- ❌ NEVER specifies endpoints, payloads, JSON schema
- ❌ NEVER specifies frameworks (Vue, Pinia, Nuxt UI, Go, Gin)
- ❌ NEVER specifies ORM/bank/queue (`gorm.Model`, PostgreSQL,
  Redis, SQS, Kafka)
- ❌ NEVER specifies auth (OAuth2, JWT, mTLS, HMAC-SHA256)
- ❌ NEVER specifies CI actions (Trivy SHA, CODEQL, golangci-lint)
- ❌ NEVER specifies performance infra (índices compostos, réplicas,
  HPA)
- ✅ Always describes **domain behavior** ("persist the user") or
  **SLO/SLA** ("listing efficient for 10k records, p95 ≤ 200ms")

**Transformation table** (selected):

| ❌ Anti-pattern (tech) | ✅ Correct (behavior) |
|---|---|
| "POST /api/v1/users with payload `{ name, email }`" | "Create user with name and email" |
| "Save in PostgreSQL with `gorm.Model`" | "Persist the user durably and uniquely" |
| "Cache with Redis and TTL 5min" | "Results consistent for up to 5 minutes" |
| "OAuth2 + PKCE + refresh token rotation" | "Secure login without exposing credentials" |
| "Helm chart with 3 replicas and HPA 70% CPU" | "Support 1,000 concurrent users" |
| "Composite index (tenant_id, created_at DESC)" | "Listing efficient for 10k records, p95 ≤ 200ms" |
| "Trivy action SHA-pinned" | "Vulnerability scan before merge" |

### Added — §4.1.2 in `team-manager.md` — Technical fence + rerouting

Team-manager now detects tech leaking in **two axes**:

- **(a) Wrong issue type** — if `type/technical` / `type/infra` /
  `type/tech-debt` / `type/docs` / `type/ui` is routed to
  domain-expert, **immediate reroute** (bash script + reassign
  to the correct persona).
- **(b) Tech leaking inside ACs of a domain issue** — return to
  domain-expert for reformulation in **behavior** or **SLO/SLA**.

Includes a **template response** with 5 examples of
anti-pattern → correct pattern.

### Added — Invariante 20 in AGENTS.md

Codifies the **2 symmetric fences** (Design + Technical) + the
"e se a stack mudar?" test as a non-violable invariant of the
meta-harness.

### Changed — `domain-expert.template.md` frontmatter

- Now lists **5 cercas** (POR QUÊ, Comportamento, Tipo apropriado,
  Sem nome de personas, Sem ação de orquestração).
- References the new skill
  [`../skills/domain-refinement/SKILL.md`](../skills/domain-refinement/SKILL.md)
  in the "Referências" section.
- "Limites" section now includes "Não direciona implementação
  técnica" + "Não é acionado para issues puramente técnicas".

### Changed — `team-manager.md` §5.3

`hermes skills install` list now includes the 3 new skills
(`nuxt-ui-patterns`, `ux-design-best-practices`,
`domain-refinement`).

### Fixed — `cli/cmd/sync.go` pre-existing `Sprintf` arg mismatch

`buildSyncPRBody` had 21 format placeholders but only 19 args
(bug introduced in v1.6.1, undetected because `go test ./...`
was not run in CI for the CLI). `go vet` failed. Added 2 `bt`
args for the `See docs/HOWTO.md and harness/stack/versions.md`
trailing placeholders.

## [1.7.0] - 2026-07-18

### Added — UI/UX skills + design cercas (ADR-0017)

**Context**: The frontend-engineer was building UIs without
structured guidance on Nuxt UI v3 patterns or UX best practices,
and the domain-expert was directing design ("click the modal
to confirm") during refinement, causing misalignment with the
project's design system (which defaults to page + breadcrumb,
not modal).

**Skills added** (in `harness/skills/`):
- `nuxt-ui-patterns/SKILL.md` — Nuxt UI v3 patterns:
  - UDashboardPage, UDashboardNavbar, UTable, UForm
  - Reference templates:
    [nuxt-ui-templates/dashboard](https://github.com/nuxt-ui-templates/dashboard),
    [saas](https://github.com/nuxt-ui-templates/saas),
    [lms](https://github.com/nuxt-ui-templates/lms)
  - Rule #0: **Page first, modal last** (decision tree)
  - Rule #1: **Breadcrumbs always** on 2+ level pages
- `ux-design-best-practices/SKILL.md` — stack-agnostic UX:
  - When to use modal/page/slideover/drawer/toast
  - Breadcrumb patterns (semantic HTML, ARIA)
  - Form patterns (inline validation, primary action)
  - WCAG AA: contrast 4.5:1, tab nav, Esc, tap targets 44x44px
  - Responsive, loading/empty/error states, i18n

### Added — Design fence in `domain-expert`

Domain-expert now **fences itself** from design:
- ❌ NEVER specifies UI components (modal, button, card, sidebar, tab)
- ✅ Always describes **behavior** (what + why), never **UI** (how it looks)
- Reformulation table with 5+ practical examples
- Updated `Limites` section to include design

### Added — `type/ui` label + routing

- New label `type/ui` for pure UI/UX work (no business logic)
- Routing: `frontend-engineer` (consults skills) → `qa` → `devops`
- **Skips** `domain-expert` and `solutions-architect`
- Added to team-manager §4.1 routing table

### Added — Design detection in `team-manager`

Team-manager now **detects** when domain-expert refinement has
UI specifics (modal, button, card, sidebar) and returns it
for reformulation. Includes a response template in
team-manager.md §4.1.1.

### Changed

- `frontend-engineer.md`: added "Design rules (UI/UX) — invioláveis"
  section with references to both new skills
- `domain-expert.template.md`: added "Cerca de Design" section with
  examples, reformulation table, gold rule
- `team-manager.md`: added `type/ui` to routing table and §4.1.1
  Design detection with response template
- New ADR-0017 documenting the decision

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
