# Contributing to git-meta-harness

Thank you for your interest in contributing to **git-meta-harness**! This
document covers how to propose changes, file issues, and submit pull requests.

## Code of conduct

Be respectful. Disagree on technical merit, not on people. We follow the
[Contributor Covenant](https://www.contributor-covenant.org/) (abbreviated).

## How to contribute

### Reporting bugs / proposing changes

1. **Check existing issues** first — your bug may already be tracked.
2. **Use the issue templates** in `.github/ISSUE_TEMPLATE/`:
   - `bug.md` — for defects
   - `feature.md` — for new functionality
   - `tech-debt.md` — for refactors / cleanups
3. **Label the issue**:
   - `type/bug`, `type/feature`, `type/tech-debt`, `type/docs`, `type/spike`
   - `domain/<x>` (if applicable)

### Proposing an ADR (Architectural Decision Record)

Major decisions go in `harness/contrib/design-decisions.md` as an ADR
(ADR-XXXX). Use the template at the top of that file.

**ADR triggers:**
- Adding/removing a persona, sensor, or workflow doc
- Changing a pinned version in `versions.md` (major version)
- Changing an invariant in `AGENTS.md` §8
- Adding a new orchestrator pattern (smart routing, etc.)

### Submitting a pull request

1. **Branch from `main`**: `git checkout -b type/feature-<short-name>`
2. **Run the local pre-flight BEFORE committing**:
   ```bash
   ./harness/scripts/smoke-test.sh .
   ./harness/scripts/check-stack-versions.sh --check-latest
   ```
   Both must pass (zero fails).
3. **Write a clear PR title and body**:
   - Title: `<type>(<scope>): <subject>` (e.g., `feat(personas): add
     domain-expert-logistics`).
   - Body: reference the issue (`Closes #N`); explain the "what" and "why";
     list breaking changes.
4. **All commits must reference an issue**: `Refs #N` or `Closes #N`.
5. **The team-manager creates branches for features** (ADR-0006). If
   you're a contributor, just open a PR — the team-manager bot will
   rename the branch if needed.

### Local pre-flight (CI gate)

This is the same gate the CI uses (faster, no GitHub Actions minutes
spent). Run it before every commit:

```bash
cd /path/to/git-meta-harness
./harness/scripts/smoke-test.sh .
./harness/scripts/check-stack-versions.sh --check-latest
```

Both must report `✅ OK` (or `✅ OK (com warns)` for the second).
Any `❌` blocks the PR.

## Project structure

```
git-meta-harness/
├── README.md                 # overview + quickstart
├── CHANGELOG.md              # version history
├── LICENSE                   # MIT
├── CONTRIBUTING.md           # this file
├── VERSION                   # semver (1.0.0)
├── bootstrap.md              # 13 princípios
├── AGENTS.md                 # contrato multi-tool + 18 invariantes
├── CLAUDE.md                 # atalho para Claude Code
├── smoke-test.md             # spec do smoke test
├── harness/
│   ├── personas/             # 7 personas + examples/
│   ├── sensors/              # 9 sensors (00-08)
│   ├── workflow/             # 6 workflow docs (00-05)
│   ├── stack/                # backend, frontend, observability, docker, code-style, versions
│   ├── templates/            # 13 templates
│   ├── skills/               # 7 skills
│   ├── contrib/              # design-decisions.md (ADRs)
│   ├── examples/             # domain-expert-* + README
│   ├── seed/                 # meta-harness-seed.md (instanciação)
│   └── scripts/              # smoke-test.sh, check-stack-versions.sh
└── .github/
    ├── ISSUE_TEMPLATE/       # bug, feature, tech-debt
    ├── PULL_REQUEST_TEMPLATE.md
    └── CODEOWNERS
```

## Style guide

- **Markdown**: ATX headers (`#`), fenced code blocks, single blank line
  between sections.
- **YAML**: 2-space indent; never `latest`, always pinned.
- **Shell**: `set -e`; `bash -n` before commit; `shellcheck` if available.
- **English** for code/comments; **pt-BR** for narrative docs is OK but
  English preferred for international accessibility.

## Release process

1. Update `VERSION` (semver).
2. Update `CHANGELOG.md` with the new version section.
3. Create an ADR if any architectural change is included.
4. Run the full local pre-flight.
5. Open a PR titled `release: vX.Y.Z`.
6. After merge, tag the commit: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`.
7. Push the tag: `git push origin vX.Y.Z`.
8. Create a GitHub Release from the tag with the CHANGELOG excerpt.

## License

By contributing, you agree that your contributions will be licensed under
the project's [MIT License](./LICENSE).
