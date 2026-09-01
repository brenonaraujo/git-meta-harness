# Sensor 14 — Loop liveness

> **Objetivo:** provar que o loop do meta-harness está **ligado**,
> não só documentado. **Quem roda:** `gmh loop doctor`, CI contracts,
> `team-manager` no tick 0. **Falha → ação:** BLOQUEAR (exit 1).
> Não começar `type/feature` enquanto isto estiver vermelho.

---

## Por que este sensor existe

**Lição do home.cloud (2026-08-31):** o seed prompt pedia para o
agente criar personas, copiar CI e abrir a issue 0. O agente leu o
prompt, não executou o CLI, e o loop nunca ligou:

1. Personas não foram criadas (`hermes profile create` depois de
   `WriteSoul` pulava `--no-skills`).
2. Pipelines Go/Nuxt foram copiados para um landing zone.
3. Sem webhook GitHub → Hermes, o board ficou em `triage` para sempre.
4. `cron monitor` no job do team-manager: snapshot inalterado =
   agente nunca acordava.
5. Worker morto (SIGINT/SIGKILL do parent) virou "então eu implemento".

O framework **descrevia** o loop. Não **instalava** o loop.

**Solução (v1.16.0, ADR-0031):** `gmh seed` materializa; este sensor
verifica o artefato.

---

## O que este sensor detecta

| Check | Bloqueante? |
|---|---|
| `harness/scripts/loop/spawn-tm.sh` existe e usa `nohup` | ✅ |
| spawn **não** contém `--monitor` | ✅ |
| `team-manager-tick.md` proíbe implementar | ✅ |
| `orchestrator-verify.md` existe | ✅ |
| `harness/loop.env` existe | ✅ |
| sem `domain-expert.md` genérico | ✅ |
| `.github/workflows/ci.yml` existe | ✅ |
| team-manager persona proíbe código de feature | ✅ (doctor Go) |

Cron jobs no Hermes (`<slug>-loop` no_agent, `<slug>-orchestrator`)
são verificados por `gmh loop status` (precisa do CLI Hermes).

---

## Como rodar

```bash
gmh loop doctor -C .
./harness/scripts/check-loop.sh
```

Exit 1 = não despachar builder.
