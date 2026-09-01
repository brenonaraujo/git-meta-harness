# SEED — what an agent does with git-meta-harness

> **TL;DR** — `gmh seed` is the install. The seed *prompt* is a
> fallback, not the procedure. If you only read markdown and
> then "become team-manager" in the parent chat, the loop is
> not running. That is the home.cloud failure mode.

---

## 1. One command

```bash
# from a checkout of git-meta-harness (this repo)
make -C cli build
./cli/bin/gmh seed ../my-app \
  --describe "what the product does in one paragraph" \
  --github owner/my-app \
  --from harness
```

`--spec FILE` instead of `--describe` if a spec already exists.

`--stack go,nuxt,postgresql` selects the full modular CI.
Any other stack (vue, swarm, docs) gets
`harness/templates/.github-workflows-ci-contracts.yml`.

`--no-agents` / `--no-github` / `--no-cron` skip those sides
(useful in CI of the framework itself).

## 2. What seed actually materializes

| Piece | Where |
|---|---|
| Framework | `harness/` copied from this tree (not GitHub Latest if `--from` / local checkout) |
| Spec | `docs/SPEC.md` |
| Overlay | `harness/PROJECT.md` |
| Specialized persona | `harness/personas/domain-expert-<domain>.md` |
| Loop scripts | `harness/scripts/loop/*`, `harness/loop.env` |
| CI | `.github/workflows/ci.yml` + `release.yml` |
| Issue 0 body | `.github/ISSUE_0.md` then `gh issue create` |
| Labels | canonical set from `harness/AGENTS.md` §4 |
| Hermes profiles | `hermes profile create <p> --no-skills` **then** SOUL.md, copy `.env`, wipe profile `skills/` |
| Cron | `<slug>-loop` (2m, no_agent, **no monitor**) + `<slug>-orchestrator` (5m, supervisor) |

## 3. How pooling events trigger team-manager

There is no GitHub webhook to localhost. Hermes cron **is** the
hook. Every 2 minutes `spawn-tm.sh`:

1. If `hermes -p team-manager` is already running → print `busy` (so
   the cron UI is not "sem execução") and exit.
2. Else `nohup hermes -p team-manager chat --oneshot --in <repo>
   --query-file harness/scripts/loop/team-manager-tick.md`.

team-manager **reads** `gh issue list` / `gh pr list` and treats
label changes as events. It moves labels and spawns personas with
`spawn-persona.sh`. It does **not** implement.

The supervisor cron (5m, default profile, `bot-chat:default`)
checks that this heartbeat is alive. It always prints 3–6 lines.

Details: [`harness/workflow/07-hermes-loop.md`](../harness/workflow/07-hermes-loop.md).

## 4. Guardrails already in the tick

- NÃO escreva código.
- Worker death ≠ license to open the PR yourself.
- No GitHub comment on `idle`.
- No `monitor` on the TM job (unchanged triage snapshot skips the agent).
- Spawn detached (`nohup`), never the parent tracker.

## 5. Verify (this is the Definition of Ready for the loop)

```bash
gmh loop doctor -C ../my-app          # exit 0
bash ../my-app/harness/scripts/check-loop.sh
gh issue list -R owner/my-app         # issue 0 open
hermes profile list                   # 7 personas, specialized domain-expert
hermes cron list                      # <slug>-loop without --monitor
```

If any of those fail, **fix seed/loop**, do not start coding the product.

## 6. After seed

The parent/default chat is finished. Further work is issue-driven:

`type/feature` → domain-expert → architect → builder → qa → human → merge.

`gmh new --spec` still exists (TODO list only). Prefer `gmh seed`.
