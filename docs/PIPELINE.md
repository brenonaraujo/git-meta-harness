# Pipeline — How the meta-harness rides on GitHub

> **TL;DR** — The meta-harness does not introduce a new
> platform. It uses GitHub Issues, Pull Requests, Labels, GitHub
> Actions, Branch Protection, CODEOWNERS, and Releases as its
> native substrate. Every pattern in the meta-harness is a
> pattern that already exists in GitHub, applied with
> discipline. This is what makes it low-friction to adopt.

---

## 1. The five GitHub primitives the meta-harness relies on

| Primitive                  | Role in the meta-harness                          | Without it, the meta-harness... |
|----------------------------|--------------------------------------------------|---------------------------------|
| **Issues**                 | Work queue of the team; per-issue briefing lives here. | Cannot decompose the spec.       |
| **Pull Requests**          | Unit of delivery; 1 issue = 1 PR.                | No way to validate before merge. |
| **Labels** (`type/*`, `domain/*`) | Routing mechanism (smart routing).            | Routing becomes ad-hoc.          |
| **GitHub Actions**         | CI/CD substrate; modular workflow with path filters. | No automated gates.              |
| **Branch Protection Rules** | The "human validation" gate, enforced at GitHub.  | Humans can merge without review. |

Optional but recommended:

- **CODEOWNERS** — assigns each `harness/` path to an owner
  (the persona responsible).
- **Releases + Tags** — versioned, auditable output.
- **Discussions / Projects** — roadmap and sprint planning.

---

## 2. The issue lifecycle in the meta-harness

The full lifecycle is defined in `harness/workflow/00-issue-lifecycle.md`.
Summary:

```mermaid
stateDiagram-v2
    [*] --> triage
    triage --> refined : type/feature
                      (skip if type/technical|infra|bug|tech-debt|spike)
    refined --> ready : solutions-architect
                      (skip if type/docs|spike)
    ready --> in_progress : builder
                      (skip if type/infra)
    in_progress --> in_review : PR opened
    in_review --> qa : sensors run
    qa --> awaiting_human : 9 sensors green
    qa --> in_progress : sensor failed (return to builder)
    awaiting_human --> done : user ✅
    awaiting_human --> in_progress : user rejected
    done --> [*] : tag + release
```

| State         | Owner               | What happens                                                            |
|---------------|---------------------|-------------------------------------------------------------------------|
| `triage`      | `team-manager`      | Detect type, domain, decompose into sub-issues, add labels.             |
| `refined`     | `domain-expert-<x>` (or skip if `type/technical|infra`) | Refine ACs from the spec. |
| `ready`       | `solutions-architect` (or skip if `type/docs|spike`)  | Produce DoD, identify open questions, propose ADRs. |
| `in-progress` | `backend-engineer` / `frontend-engineer` / `devops-engineer` | Code, tests, local pre-flight. Branch created by team-manager. |
| `in-review`   | `team-manager`      | Open PR, post the briefing as a comment.                                |
| `qa`          | `quality-assurance` | Run all 9 sensors. Approve or return.                                  |
| `awaiting human` | `team-manager`   | Block on the human. Branch protection enforces.                         |
| `done`        | `team-manager`      | Merge, tag, release, close issue.                                       |

The `team-manager` is the **only** persona that moves an issue
between states. Sub-personas do not self-advance; they
**report back** via a comment on the issue.

---

## 3. The PR convention

Every PR in a meta-harness project follows this structure:

1. **Title**: `<type>(<scope>): <subject>`
   - e.g., `feat(backend): add /users endpoint with phone + password_hash`
2. **Body** (from the `harness/templates/pr-description.md`):
   ```markdown
   ## What & Why
   <!-- 1-3 sentences referencing the issue. -->

   Closes #N

   ## How to test locally
   <!-- Commands a reviewer can run. -->

   ## Definition of done
   - [ ] acceptance criteria from #N met
   - [ ] tests added/changed
   - [ ] 12-factor audit passed
   - [ ] i18n parity (en/pt-BR/es) maintained
   - [ ] no hardcoded strings
   - [ ] no coverage regression
   ```
3. **Linked issue** is closed automatically when merged.
4. **Required checks** (set in Branch Protection):
   - Smoke test
   - Stack version check
   - Lint
   - Test + coverage
   - Vulnerability scan
   - OpenAPI contract
   - 12-Factor audit
   - i18n audit
5. **CODEOWNERS** automatically assigns reviewers based on
   the changed files.

---

## 4. Labels: the routing mechanism

The meta-harness defines two label namespaces.

### `type/*` — what kind of work

| Label             | Routed personas (smart routing)                                        |
|-------------------|-------------------------------------------------------------------------|
| `type/feature`    | `domain-expert` → `solutions-architect` → builder → `qa`               |
| `type/technical`  | `solutions-architect` → builder → `qa` (skip `domain-expert`)          |
| `type/infra`      | `solutions-architect` → `devops-engineer` (skip `domain-expert`, builder) |
| `type/bug`        | `solutions-architect` → builder → `qa` (skip `domain-expert`)          |
| `type/tech-debt`  | builder → `qa` (skip `domain-expert`)                                  |
| `type/docs`       | `team-manager` only (editorial review)                                 |
| `type/spike`      | `solutions-architect` only (output = ADR)                              |

### `domain/*` — which domain-expert

| Label             | Routed persona                          |
|-------------------|------------------------------------------|
| `domain/banking`  | `domain-expert-banking`                  |
| `domain/retail`   | `domain-expert-retail`                   |
| `domain/mandai`   | `domain-expert-mandai`                   |
| `domain/<x>`      | `domain-expert-<x>`                      |

The `team-manager` reads both namespaces and dispatches. A
generic `domain-expert` is **a hard invariant violation** (no
such profile is created).

---

## 5. The CI workflow

The `harness/templates/.github-workflows-ci.yml` is a single
file with 12+ jobs orchestrated by a `changes` job at the top
that uses `dorny/paths-filter` to detect which components
changed.

```
                    changes (dorny/paths-filter)
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
  backend-* jobs         frontend-* jobs       infra/12-factor
  (lint, test,           (lint, typecheck,     (always-on gate)
   vuln, contract)        test, vuln)              │
                                                    ▼
                                            build-backend / build-frontend
                                            (cache scope per service)
                                                    │
                                                    ▼
                                            summary (always-on)
```

Speed comparison (from `harness/contrib/design-decisions.md`
ADR-0011):

| Scenario                              | Without path filter | With path filter |
|---------------------------------------|---------------------|-------------------|
| PR only changes `web/i18n/pt-BR.json` | ~8 min              | ~30s              |
| PR only changes `backend/internal/...`| ~8 min              | ~3 min            |
| PR only changes `docs/SPEC.md`        | ~8 min              | ~30s              |
| 5 commits pushed to same PR           | 5 × 8 min = 40 min  | 1 × 8 min (cancel-in-progress) |

The path filter is **not optional**. It is the difference between
a CI that scales and a CI that costs more as the team grows.

---

## 6. Why GitHub specifically (and not GitLab, Bitbucket, etc.)

The meta-harness is **tool-portable at the agent level**
(Hermes, Claude Code, Codex, etc. all work) but the pipeline is
currently **GitHub-specific** for the following reasons:

1. **Issues + PRs + Actions + Releases is the most-used
   combination** in open source and at work.
2. **CODEOWNERS and branch protection** are first-class and
   well-documented.
3. **GitHub Actions** has the largest ecosystem of reusable
   actions (including `dorny/paths-filter` and
   `aquasecurity/trivy-action`).
4. **GitHub Releases** is the standard way to publish
   versioned artifacts.

Porting to GitLab or Bitbucket would require replacing the
CI template and the issue template, but the **personas,
sensors, workflow, stack, and templates** would carry over
unchanged. The harness-of-harnesses is independent of the
hosting platform.

A future ADR-0014 (or similar) could formalize a GitLab
adapter if there is demand.

---

## 7. The "low-friction" promise, summarized

When a new project adopts the meta-harness, here is the entire
friction budget:

| Step                                           | Time    | Who        |
|------------------------------------------------|---------|------------|
| `git clone` the meta-harness                  | 5s      | human      |
| `cp -R harness ./harness` in the new project   | 5s      | human      |
| `cp templates/.github-workflows-ci.yml`         | 5s      | human      |
| `cp templates/.golangci.yml`                   | 5s      | human      |
| `cp templates/.env.example`                    | 5s      | human      |
| Paste the spec into the agentic CLI             | 60s     | human      |
| Wait for the team-manager to materialize       | 5-30 min| agent      |
| Review the first PR                            | 15 min  | human      |
| Merge when the sensors are green               | 5s      | human      |
| **Total to a working PR:**                     | **~1 hour** | —    |

Without the meta-harness, the same project would be 1-2 days
of stack decision, Dockerfile writing, CI configuration,
issue template creation, persona definition, and ADR setup —
**before** any business code is written.
