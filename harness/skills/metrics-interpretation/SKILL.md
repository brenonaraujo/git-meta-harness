# Skill: metrics-interpretation

> **Quando usar:** ao interpretar `gmh metrics --json` ou o
> dashboard Prometheus gerado. Usado por `team-manager` em
> weekly review e por `quality-assurance` em trend analysis.
>
> **Output típico:** ação corretiva baseada em tendência
> (ex.: "BEM é top violação, calibrar sensor 12 em warn").

---

## 1. Quando esta skill se aplica

- Review semanal de métricas (`gmh metrics --json` ou
  dashboard).
- Decisão de calibrar sensor (passar de blocking pra warn,
  ou vice-versa).
- Diagnóstico de "por que o harness tá degradando".
- Report para stakeholder (Brenon, eng manager, etc).

**NÃO se aplica** se:
- Métrica está momentânea (use `gmh doctor` direto).
- Você quer apenas confirmar health (use `gmh doctor --json`).

---

## 2. As 4 dimensões do Health Score

| Dimensão | O que mede | Como melhorar |
|----------|------------|---------------|
| **Harness** | % invariantes passando | Adicionar arquivos faltantes (AGENTS.md, scripts/, etc). |
| **Agents** | Core personas presentes + specialized | Adicionar `domain-expert-<domínio>` (não generic). |
| **Skills** | Skills declaradas vs instaladas | Rodar `gmh agents sync` pra instalar. |
| **Sensors** | Sensors 10-13 com script | Criar scripts `check-<sensor>.sh` pra sensors 10-13. |

**Score geral** = média ponderada (harness×2 + agents×1 + skills×1 + sensors×2) / 6.

---

## 3. Thresholds e ações

### Health score

| Range | Status | Ação |
|-------|--------|------|
| **90-100** | 🟢 healthy | Manter. Celebrar com time. |
| **80-89** | 🟡 needs attention | Revisar quais dimensões estão <80. |
| **70-79** | 🟠 needs work | Priorizar fixes. Rodar `gmh doctor --json`. |
| **<70** | 🔴 critical | Bloquear merges até melhorar. --strict exit 1. |

### Flow compliance (% type/feature com refined+ready+comments)

| Range | Status | Ação |
|-------|--------|------|
| **≥85%** | 🟢 healthy | Manter. |
| **70-84%** | 🟡 needs attention | Revisar quais épicos pularam flow. |
| **<70%** | 🔴 critical | team-manager deve parar e refazer. |

**Por que flow compliance é a métrica mais importante**:
se 100% das features passam pelo flow (refined+ready+comments),
o resto segue. Se <80%, **todos os outros sensors viram
teatro** (sensores detectam violações em código que nem
deveria ter sido escrito sem refinement).

### Sensor blocks (count per sensor per week)

| Range | Status | Ação |
|-------|--------|------|
| **0-2** | 🟢 normal | Esperado. |
| **3-5** | 🟡 atenção | Investigar: convenção local conflitando? Calibrar? |
| **>5** | 🔴 many | Sensor provavelmente está bloqueando coisas legítimas. Re-calibrar (warn em vez de block, ou scope). |

### Drift (skills stale, CI drift, harness files missing)

| Range | Status | Ação |
|-------|--------|------|
| **0** | 🟢 | Manter. |
| **1-3** | 🟡 | Rodar `gmh agents sync` / `gmh sync`. |
| **>3** | 🔴 | Drift acumulado. Bloquear novos features até limpar. |

---

## 4. Padrões comuns (e o que fazer)

### Padrão 1: Health score caindo 5+ pontos/semana

**Sintoma**: `gmh_health_score` era 90, agora 85, caindo
consistentemente.

**Diagnóstico**:
- Rodar `gmh doctor --json` e olhar quais dimensões caíram.
- Se `sensors` caiu: novo sensor adicionado mas script não.
- Se `agents` caiu: persona removida.
- Se `skills` caiu: skill removida do framework.

**Ação**:
- Se queda é por adição de novo sensor (esperado): escrever
  script `check-<sensor>.sh` em 1 dia.
- Se queda é por remoção acidental: re-adicionar.
- Se queda é por drift: rodar `gmh agents sync`.

### Padrão 2: Flow compliance < 80%

**Sintoma**: `gmh_flow_compliance_pct` 75% (30 de 40
type/feature com flow completo).

**Diagnóstico**:
- `gh issue list --label type/feature --state all` e ver
  quais NÃO têm `refined` + `ready`.
- Listar por épico: épico X tem 5/5 sem flow = team-manager
  pulou para esse épico inteiro.
- Listar por team-manager (se houver assignees): 1 pessoa
  pulou consistentemente = problema cultural.

**Ação**:
- Para 1 épico específico: rodar sensor 13 contra o épico
  (`./harness/scripts/check-feature-flow.sh <id>`), listar
  violações, e team-manager refaz.
- Para problema cultural: reunião 1:1 com team-manager,
  explicar sensor 13 + invariante 24 + 3-pillar pattern.

### Padrão 3: Sensor blocks crescendo em 1 sensor

**Sintoma**: `gmh_sensor_blocks_7d{sensor="frontend-polish"}`
foi de 2 pra 15 em 1 semana.

**Diagnóstico**:
- `git log --oneline | grep "frontend-polish" | head`:
  ver o que tá sendo bloqueado.
- Se são PRs com BEM misturado: convenção local (Vue 2 legado)
  vs Nuxt 4 — calibrar sensor pra "warn, não block" em
  projetos Vue 2.
- Se são PRs com hex hardcoded: time não viu o guidance —
  compartilhar skill `nuxt-ui-patterns` + rodar workshop.

**Ação**:
- Calibrar sensor (modifica `harness/sensors/<NN>.md`
  threshold).
- OU: documentar exceção em `harness/ADOPT-REPORT.md` (projetos
  com stack legado são aceitos com warn).
- Comunicar ao time: "calibramos sensor, mas a regra ainda vale
  pra greenfield".

### Padrão 4: Avg time-to-close > 7 dias

**Sintoma**: `gmh_avg_time_to_close_days` 8.5 (meta: 7).

**Diagnóstico**:
- `gh issue list --state closed --limit 30 --json closedAt,
  createdAt`: calcular distribuição.
- Se集中在 1-2 épicos: épicos grandes demais, decompor.
- Se distribuídos: time subdimensionado OU processo lento.

**Ação**:
- Decompor épicos grandes em 3-5 sub-issues (skill
  `pre-implementation-design`).
- Se time subdimensionado: report para eng manager, pedir
  headcount ou scope reduction.
- Se processo lento: identificar gargalo (espera de review?
  espera de QA? espera de devops?).

### Padrão 5: Out of date

**Sintoma**: `gmh_out_of_date 1` (local < latest).

**Diagnóstico**:
- `gmh doctor` mostra local vs latest.
- Olhar CHANGELOG.md entre local e latest: o que mudou?

**Ação**:
- `gmh update --to <latest>` (atualiza harness/).
- `gmh agents sync` (atualiza profiles).
- Revisar breaking changes (se houver).

---

## 5. Calibração de sensors (decisão)

**Quando bloquear vs warn**:

| Caso | Block | Warn |
|------|-------|------|
| **Greenfield, Nuxt 4+** | BEM, hex, off-scale spacing | (n/a, é tudo block) |
| **Legacy Vue 2** | Off-scale spacing, missing alt | BEM, hex hardcoded |
| **Quick prototype** | (n/a) | Tudo (calibrar pra warn) |
| **Compliance-heavy (fintech)** | PII handling, audit log | Hex hardcoded |
| **ML/data science** | Missing tests | BEM (não relevante) |

**Default**: começar com BLOCK em todos. Calibrar pra WARN
quando:
- 5+ violations legítimas em 1 semana.
- Convenção local é justificada (legacy stack).
- Documentar em `harness/ADOPT-REPORT.md` a calibração.

**Voltar pra BLOCK quando**:
- Stack é atualizado (Vue 2 → Vue 3).
- Time já teve workshop sobre a regra.
- 0 violações por 4+ semanas seguidas (validar que regra foi
  aprendida).

---

## 6. Relatório semanal (template)

```markdown
# Harness metrics — week of YYYY-MM-DD

## TL;DR
- Health: 88 (was 85 last week) 🟢
- Flow compliance: 87% (was 82%) 🟢
- Top violation: BEM (sensor 12, 12 occurrences)
- Out of date: no

## What changed
- 3 issues went through full flow
- 2 issues were blocked by sensor 12 (calibrating to warn)
- 1 sensor calibration (sensor 12, BEM: block → warn)

## What to do next week
- [ ] Run `gmh new --spec` for next epic
- [ ] Re-run `gmh adopt` after Vue 2 → Vue 3 migration
- [ ] Workshop on BEM (1h, all engineers)
```

---

## 7. Quem usa

- `team-manager` (principal): weekly review.
- `quality-assurance`: trend analysis, calibration proposals.
- `devops-engineer`: Prometheus alerting tuning.
- Brenon: high-level review mensal.

---

## 8. Quem mantém

- Brenon + time de plataforma.
- Atualizada quando ADRs 0026, 0029 evoluem.
