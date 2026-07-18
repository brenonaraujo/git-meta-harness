## What & Why

<!-- What is changing? Why? Reference the issue with `Closes #N` or
     `Refs #N`. -->

Closes #

## Type

<!-- Check the box that applies. Remove others. -->

- [ ] `type/feature` — new functionality
- [ ] `type/technical` — technical refactor
- [ ] `type/infra` — infra / DevOps
- [ ] `type/bug` — bug fix
- [ ] `type/tech-debt` — refactor / cleanup
- [ ] `type/docs` — documentation only
- [ ] `type/spike` — research / spike

## Affected scope

<!-- List the files / sections affected. -->

- [ ] `harness/personas/...`
- [ ] `harness/sensors/...`
- [ ] `harness/workflow/...`
- [ ] `harness/stack/...`
- [ ] `harness/templates/...`
- [ ] `harness/skills/...`
- [ ] `harness/contrib/design-decisions.md` (ADR)
- [ ] `harness/AGENTS.md` (invariantes)
- [ ] `harness/bootstrap.md` (princípios)
- [ ] `harness/seed/meta-harness-seed.md` (instanciação)
- [ ] `harness/scripts/...`

## Pre-flight (mandatory, run BEFORE `gh pr create`)

- [ ] `./harness/scripts/smoke-test.sh .` — passa (zero fails)
- [ ] `./harness/scripts/check-stack-versions.sh --check-latest` — passa (zero fails)
- [ ] (Go) `cd backend && make lint && make test && make vuln` — verde
- [ ] (Node) `cd web && pnpm lint && pnpm typecheck && pnpm test:run && pnpm audit` — verde

## How to test locally

<!-- Steps a reviewer (or you) can run to verify. -->

```bash
# paste commands here
```

## Risk & rollback

<!-- What's the blast radius? How to undo if needed? -->

## ADR / invariant changes

<!-- If you added/changed an invariant or an ADR, list them here. -->

- [ ] ADR: (none / link)
- [ ] Invariant added/changed: (#N, briefly)

## Checklist

- [ ] Commits reference the issue (`Refs #N` or `Closes #N`)
- [ ] PR title: `<type>(<scope>): <subject>`
- [ ] No new strings hardcoded (i18n preserved)
- [ ] Coverage ≥ 80% in changed packages
- [ ] CHANGELOG.md updated (if user-facing)
- [ ] `VERSION` bumped (if release-worthy)
