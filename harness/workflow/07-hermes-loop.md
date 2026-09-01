# Workflow 07 — Hermes cron pooling (the actual hook)

> GitHub Issues are the **memory**. They are not a running loop.
> There is no webhook from GitHub into a local Hermes process.
> The scheduler is the hook.

---

## Why pooling

A two-way bind (GitHub App / webhook → localhost) does not exist
in the Hermes-only setup. Polling is the event bus.

```
GitHub issue/PR/label change
        │  (no push to the laptop)
        ▼
Hermes cron  <slug>-loop   every 2m   no_agent
        │  spawn-tm.sh  (prints "busy" or "spawned")
        ▼
hermes -p team-manager --oneshot --query-file team-manager-tick.md
        │  reads gh issue list / gh pr list
        ▼
label move + spawn-persona.sh (nohup, detached)
        │
        ▼
persona profile  (domain-expert / architect / builder / qa / devops)
        │  comments peguei … pronto
        ▼
next TM tick sees the new state
```

A second job, `<slug>-orchestrator` every 5m (default profile,
`deliver: bot-chat:default`), is the **supervisor**. It always
prints 3–6 lines. It never implements. It never uses `[SILENT]`.

---

## How an "event" reaches team-manager

There is no event object. The tick **discovers** state:

| GitHub state | TM action |
|---|---|
| new issue, `triage` + `type/feature` | spawn `domain-expert-<domain>` |
| `refined` | spawn `solutions-architect` |
| `ready` | create `feature/<id>-<slug>`, spawn builder |
| `in-progress` + PR + CI green | spawn `quality-assurance` |
| `qa` + mergeable PR | merge / wait `validado` |
| `qa` + no PR + APROVADO | close `done` (not idle) |
| `qa` + no PR + REPROVADO | spawn builder |
| no new action | print `idle`, **no GitHub comment** |

Idle with a comment that repeats "não despachar" is a bug (looks
alive, does nothing, spams the board).

---

## Guardrails (team-manager)

Encoded in `harness/personas/team-manager.md` and in the tick file.

1. **Do not implement.** Orchestrate. A dead worker is a spawn bug.
2. **Do not wait** on the worker inside the tick. Spawn and exit
   (cron LLM interrupt is ~3 min).
3. **Do not** put `monitor` on `<slug>-loop`. Unchanged `triage`
   snapshots suppress the agent forever.
4. **Do not** spawn via the parent chat's `terminal(background=true)`
   — the parent tracker sends SIGINT then SIGKILL. Use `spawn-persona.sh`.
5. **Do** print one line from the no_agent script (`busy` / `spawned`).
6. Create Hermes profiles with `--no-skills` **before** writing
   `SOUL.md`. Copy `~/.hermes/.env`. Never `--clone-from` + `--no-skills`.
7. After `gmh agents sync`, wipe `~/.hermes/profiles/<p>/skills/*`.

---

## Install

```bash
gmh seed <name> --describe "…" --github owner/repo
# or, on an already-copied harness:
gmh loop install -C . --github owner/repo
hermes gateway install --start-now   # cron does not fire otherwise
gmh loop doctor
```

Jobs created:

- `<slug>-loop` — `2m`, `--no-agent`, `--script <slug>-spawn-tm.sh`, `--workdir <repo>`
- `<slug>-orchestrator` — `5m`, agent, `--deliver bot-chat:default`, `--workdir <repo>`

Neither job has `--monitor-script` / `--monitor-url`.
