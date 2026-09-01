# AGENT.md — git-meta-harness (read this first)

This repository is the **seed**, not a product. If a human says
anything like:

> usa o git-meta-harness pra criar um projeto novo, vai ser um saas
> de agendamento online

you do **not** paste `harness/seed/meta-harness-seed.md` and improvise
personas. You run the CLI. Full playbook: [`docs/SEED.md`](docs/SEED.md).

## Do this

1. Build the CLI from **this tree** (PATH `gmh` is often stale):
   ```bash
   make -C cli build
   ./cli/bin/gmh version
   ```
2. Seed the project (creates dir, harness, specialized domain-expert,
   CI, issue 0 body, Hermes loop scripts, profiles, cron):
   ```bash
   ./cli/bin/gmh seed ../booking-saas \
     --describe "SaaS de agendamento online" \
     --github <owner>/<repo> \
     --from harness
   ```
3. Prove the loop is wired:
   ```bash
   ./cli/bin/gmh loop doctor -C ../booking-saas
   bash ../booking-saas/harness/scripts/check-loop.sh
   ```
4. Gateway must be up or cron never fires:
   `hermes gateway install --start-now`
5. **Stop.** Do not implement the SaaS in this session.
   `team-manager` (Hermes cron pooling) owns the board.
   You are done when issue 0 exists and `gmh loop doctor` exits 0.

## Do not

- Create Hermes profiles by hand after writing `SOUL.md` (skips `--no-skills`).
- Copy Go/Nuxt `ci.yml` / GHCR `release.yml` onto a docs or Vue landing zone.
- Put `monitor` on the team-manager cron job.
- Spawn workers with `terminal(background=true)` in the parent chat.
- Implement Vue/Go/Dockerfile because a worker died.
- Ask more than two clarifying questions. Default and seed.

## Commands that matter

| Intent | Command |
|---|---|
| New project from a sentence | `gmh seed <name> --describe "…" --github owner/repo` |
| New project from a spec file | `gmh seed <name> --spec SPEC.md --github owner/repo` |
| Existing repo | `gmh adopt` then `gmh loop install` |
| Is the loop real? | `gmh loop doctor` |
| Personas only | `gmh agents install <persona>` (create `--no-skills` **first**) |

Runtime contract after seed: [`harness/workflow/07-hermes-loop.md`](harness/workflow/07-hermes-loop.md).
