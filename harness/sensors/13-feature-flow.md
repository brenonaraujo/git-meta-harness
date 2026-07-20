# Sensor 13 — Feature Flow Enforcement

> **Objetivo:** bloquear transições `triage → in-progress` em
> issues `type/feature` que **não passaram pelo flow canônico**
> (domain-expert refinement + solutions-architect DoD + labels
> `refined` + `ready` + comentários canônicos). **Quem roda:**
> `team-manager` antes de delegar builder, e `quality-assurance`
> no PR. **Quando:** local, ANTES de `in-progress`; CI, no job
> `feature-flow`. **Falha → ação:** BLOQUEAR (exit 1) com lista
> de violações + recovery.

---

## Por que este sensor existe

**Lição do Mandaí v2 (jul/2026, Épico #48 F7+F8+F10 — Avaliações
+ Reputação + Share):** o `team-manager` criou o épico com
`type/feature`, mas **zero sub-issues passaram pelo flow canônico**
(`domain-expert` → `solutions-architect` → builder). Resultado:
builders receberam só a descrição de 1-2 parágrafos da issue (sem
ACs, sem DoD, sem edge cases), implementaram "no escuro", e o
custo foi ~30min-1h de retrabalho por sub-issue.

O framework **documentava** o flow (AGENTS.md §3.1, team-manager
§4 Smart Routing, invariante 12) mas **não tinha enforcement** —
o team-manager podia pular `refined` e `ready` sem ser bloqueado.

**Solução (v1.13.0, 3-pillar pattern, ADR-0025)** — 6 mudanças
coordenadas:

1. **Sensor 13 (este arquivo)** — BLOQUEIA transição
   `triage → in-progress` em `type/feature` sem flow completo.
2. **Templates canônicos** em `harness/templates/comments/`
   (`domain-expert-refinement.md`, `solutions-architect-dod.md`)
   — copy-paste pronto.
3. **AGENTS.md invariante 24** (NOVA) — type/feature REQUER
   flow completo antes de `in-progress`.
4. **team-manager.md §3.1.3** (NOVA) — comando canônico antes
   de delegar builder + recovery table.
5. **Builder personas** — responsabilidade 0a: LER TODOS OS
   COMENTÁRIOS antes de implementar.
6. **PR template atualizado** — seções "Context from
   domain-expert" + "DoD from architect" obrigatórias.

---

## O que este sensor detecta

Para cada issue `type/feature` (ou `type/<x>` que herda de
`type/feature`), o sensor verifica **5 categorias**:

| Categoria | Detector | Bloqueante? |
|---|---|---|
| `no_refined_label` | label `refined` ausente | ✅ SIM |
| `no_ready_label` | label `ready` ausente | ✅ SIM |
| `no_refinement_comment` | comentário de domain-expert (heurística: regex `(AC|Edge case)`) ausente | ✅ SIM |
| `no_dod_comment` | comentário de solutions-architect (heurística: regex `(Pilar|Definition of Done|DoD)`) ausente | ✅ SIM |
| `dod_without_refined` | DoD presente mas refined ausente (inverted order) | ✅ SIM |

Se QUALQUER categoria falhar → exit 1, lista de violações.

---

## Como rodar

```bash
# Local: verificar 1 issue específica
./harness/scripts/check-feature-flow.sh 48

# Local: verificar todas as type/feature abertas
./harness/scripts/check-feature-flow.sh

# CI: mesmo comando, exit 1 = bloqueia merge
```

**Auto-detecta repo** via `git remote get-url origin` (não
precisa de flag).

---

## Output (exit 1)

```
==> Feature flow check (sensor 13, v1.13.0)
==> Issue: #48

Checked 1 type/feature issue(s).

BLOCKING: FEATURE FLOW VIOLATIONS (sensor 13, v1.13.0):

  Issue #48: [Épico] Avaliações + Reputação + Share (F7+F8+F10)
    ❌ no_refined_label
    ❌ no_ready_label
    ❌ no_refinement_comment
    ❌ no_dod_comment

Total: 4 violation(s) across 1 issue(s).
```

**Recovery** (team-manager deve fazer):

1. Pedir ao `domain-expert-<domínio>` para refinar a issue:
   - Adicionar label `refined`.
   - Comentar usando `harness/templates/comments/domain-expert-refinement.md`.
2. Pedir ao `solutions-architect` para fazer o DoD:
   - Adicionar label `ready`.
   - Comentar usando `harness/templates/comments/solutions-architect-dod.md`.
3. Re-rodar o sensor: deve passar.

---

## Output (exit 0)

```
==> Feature flow check (sensor 13, v1.13.0)

Checked 1 type/feature issue(s).
OK: All type/feature issues have refined + ready + DoD.
```

---

## Casos especiais

### Sub-issue pequena (1-2 ACs, sem edge case)

Sub-issues filhas de épicos que já passaram pelo flow são
**exceptions** — não precisam de refinement próprio, o épico
já cobriu. **Mas a sub-issue precisa ter:**

- Label `refined` (herdada do épico via automation OU manual).
- Label `ready` (idem).
- Comentário linkando pro épico `#<parent>` (1 linha).

Sensor valida: se issue é `type/feature` E tem `parent` no
body, basta ter as 2 labels + 1 link. Sem `refinement_comment`
+ `dod_comment` próprios.

### Refinement parcial (só `refined`, sem `ready`)

**Bloqueia** (`no_ready_label`). team-manager deve pedir ao
`solutions-architect` para fazer o DoD antes de prosseguir.

### Builder reclama "não tenho contexto"

**Esse é o ponto.** Builder (backend/frontend-engineer) tem
responsabilidade 0a: LER TODOS OS COMENTÁRIOS. Se não há
comentários, PARE e reporte ao team-manager que a issue
precisa passar pelo flow.

### Builder empurra 3x sem refinar

team-manager deve **forçar refinement** (não delegar pro
builder). Se domain-expert não tem bandwidth, escalar.

---

## Princípio

> **"Documentar o flow não basta — tem que ENFORCE no
> sensor + fornecer templates canônicos + dar regra
> explícita ao builder."**

3-pillar pattern (v1.13.0):

1. **Sensor BLOQUEANTE** (este arquivo) — garante que
   team-manager não pula.
2. **Templates canônicos** (`harness/templates/comments/`)
   — copy-paste pronto, formato consistente.
3. **Regra explícita ao builder** (responsabilidade 0a
   em backend/frontend-engineer) — LER TODOS OS
   COMENTÁRIOS antes de codar.

Aplica-se a QUALQUER workflow multi-stage (PM → designer
→ architect → builder). Documentar é necessário, mas não
suficiente.

---

## Quem detecta / Quem corrige

- **team-manager**: roda este sensor antes de `in-progress`.
  Se vermelho: devolve com `in-progress` → `triage` + comentário
  listando violações.
- **backend-engineer / frontend-engineer**: regra 0a, LER
  TODOS OS COMENTÁRIOS antes de codar. Se faltam: PARE e
  reporte ao team-manager.
- **quality-assurance**: verifica no PR se as seções
  "Context from domain-expert" + "DoD from architect" do
  PR template estão preenchidas.
- **CI (futuro)**: pode rodar `./check-feature-flow.sh` no
  job `pre-merge` e bloquear se vermelho.
