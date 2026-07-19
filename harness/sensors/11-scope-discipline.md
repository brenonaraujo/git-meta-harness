# Sensor 11 — Scope Discipline (PILARES vs BLUEPRINTS)

> **Objetivo:** detectar quando `domain-expert` ou
> `solutions-architect` **vazam** pra camada de
> implementação (escrevem SQL, dão nomes de funções, paths
> de arquivo, etc). **Quem roda:** `team-manager`.
> **Quando:** depois de cada output de domain-expert ou
> solutions-architect, ANTES de mover label pra `refined`
> ou `ready`. **Falha → ação:** emite **recomendação** (NÃO
> bloqueia) pra encurtar/reescrever na próxima iteração.

---

## Por que este sensor existe

**Lição do Mandaí v2 (jul/2026, Épico F4+F5 — Ciclos +
Pedidos):** o `domain-expert` escreveu 19 ACs + 18 edge cases,
e o `solutions-architect` escreveu um DoD de **150+ linhas**
com:
- Nomes de funções: `MustGenerateCycleSlug`, `CheckCycleLimits`,
  `CanTransition`
- SQL: `SELECT FOR UPDATE no cycles WHERE workspace_id=$1`
- Paths: `backend/internal/service/cycle_service.go`,
  `internal/repository/cycle_repo.go`
- ORMs: `gorm`, `pgx`
- Migrations: `000009_cycles.up.sql`, `000010_cycle_products`
- Métricas: `orders_created_total{status}`,
  `cycle_transitions_total{from, to}`

O `backend-engineer` virou **executor cego** — só seguiu o
blueprint, sem questionar, sem otimizar, sem ownership
técnica. **Custo:** ~3-5h de retrabalho quando o builder
percebe que algumas decisões estavam erradas (e.g.,
retry de slug não precisava de nanoid lib, snapshot podia
ser um campo a mais em order_items em vez de 2 colunas).

**Causa raiz**: o `domain-expert` e o `solutions-architect`
não tinham **limites explícitos** nem **cerca de output**.
Eram incentivados a escrever mais (mais detalhe = mais
"completo"), e o resultado era **vazamento de camada**.

**Princípio v1.11.0 (ADR-0021)**: domain-expert e
solutions-architect entregam **PILARES** (o que + por quê +
invariantes). O **builder** entrega **BLUEPRINTS** (como).

---

## Como rodar (3 passos)

### Passo 1 — Detectar persona origem

O `team-manager` identifica se o output é de `domain-expert`
ou `solutions-architect` (baseado em label ou autor do
comentário).

### Passo 2 — Aplicar regex heurística

Pra cada padrão proibido (ver tabela abaixo), contar
ocorrências. Se **algum padrão** for detectado OU se o output
passar de **30k tokens** (~75k chars), emitir **recomendação**.

```bash
# Pseudo-código (implementado em gmh scope-check ou
# manualmente pelo team-manager via skill)
output=$(cat /path/to/comment.md)
persona="domain-expert"  # ou "solutions-architect"

# Sinais de vazamento
signals=$(echo "$output" | python3 -c '
import sys, re
text = sys.stdin.read()
signals = []
checks = [
  ("sql_keywords", r"\b(SELECT|INSERT|UPDATE|DELETE|WHERE|FROM)\b", 5),
  ("orm_names",    r"\b(gorm|pgx|sqlx|sqlc|gin|echo|chi|fiber|nestjs|express)\b", 3),
  ("go_files",     r"\b[a-z_/]+\.go\b", 3),
  ("internal_paths", r"\binternal/[a-z_/]+", 2),
  ("migrations",   r"\b00000[0-9]_.*\.up\.sql", 1),
  ("endpoints",    r"\b(GET|POST|PUT|PATCH|DELETE) /api", 3),
  ("func_names",   r"\b[A-Z][a-zA-Z]+\([^)]*\)\s*\{", 10),  # PascalCase func()
  ("prometheus",   r"\b(prometheus|metrics)\.New(Counter|Histogram|Gauge)", 2),
  ("tokens",       r".{75000,}", 1),  # > 75k chars
]
for name, pattern, threshold in checks:
  count = len(re.findall(pattern, text, re.MULTILINE))
  if count >= threshold:
    signals.append(f"{name}:{count}")
print(" ".join(signals) if signals else "OK")
')
```

### Passo 3 — Emitir recomendação (NÃO bloquear)

**Sempre passa o output para o próximo passo** (label
`refined` ou `ready` aplicados). Apenas **emite aviso**:

```markdown
⚠️  scope-discipline (sensor 11): output do `<persona>` tem
sinais de vazamento de camada:
- `sql_keywords: 7` (SELECT, INSERT, UPDATE mencionados)
- `orm_names: 4` (gorm, pgx mencionados)
- `tokens: 47k` (acima de 30k recomendado)

**Recomendação** (não bloqueia):
- Reformule em PILARES (o que + por quê), não BLUEPRINTS
  (o como)
- Remova nomes de funções, paths, SQL, ORMs específicos
- Aplique a skill `solution-scoping` na próxima iteração
- Detalhes: `harness/skills/solution-scoping/SKILL.md`

> Esta recomendação é **pra próxima iteração**, não pra
> reescrever agora. O builder segue o que está escrito.
```

---

## Padrões detectados (regex heurística)

| Pattern | Regex | Threshold | Quem |
|---|---|---|---|
| `sql_keywords` | `\b(SELECT\|INSERT\|UPDATE\|DELETE\|WHERE\|FROM)\b` | ≥ 5 ocorrências | domain-expert, solutions-architect |
| `orm_names` | `\b(gorm\|pgx\|sqlx\|sqlc\|gin\|echo\|chi\|fiber\|nestjs\|express)\b` | ≥ 3 | domain-expert, solutions-architect |
| `typeorm_nestjs` | `\b(TypeORM\|GORM\|PGx\|Sqlx\|Gin\|Echo\|NestJS)\b` (PascalCase) | ≥ 2 | domain-expert, solutions-architect |
| `go_files` | `\b[a-z_/]+\.go\b` (excluindo SKILL.md etc) | ≥ 3 | domain-expert |
| `internal_paths` | `\binternal/[a-z_/]+` | ≥ 2 | domain-expert |
| `migrations` | `\b00000[0-9]_.*\.up\.sql` | ≥ 1 | domain-expert, solutions-architect |
| `endpoints` | `\b(GET\|POST\|PUT\|PATCH\|DELETE) /api` | ≥ 3 | domain-expert |
| `func_names` | `\b[A-Z][a-zA-Z]+\([^)]*\)\s*\{` (PascalCase func decl) | ≥ 10 | domain-expert |
| `prometheus` | `\b(prometheus\|metrics)\.New(Counter\|Histogram\|Gauge)` | ≥ 2 | domain-expert, solutions-architect |
| `tokens` | `.{75000,}` (output > 75k chars = ~30k tokens) | ≥ 1 | domain-expert, solutions-architect |

> **Por que não bloqueia**: outputs com vazamento leve
> (1-2 sinais abaixo do threshold) podem ser aceitáveis
> (ex.: domain-expert menciona "Pix" como contrato de
> pagamento, mesmo "Pix" não sendo tecnologia). O
> `team-manager` decide caso a caso se reformula ou aceita.

---

## Limites recomendados (não-bloqueantes)

| Persona | Output máx (recomendado) | Output tokens (recomendado) | Se passar |
|---|---|---|---|
| `domain-expert` | ~150 linhas / ~10k chars | ~3k tokens | ⚠️ warning |
| `solutions-architect` | ~200 linhas / ~13k chars | ~5k tokens | ⚠️ warning |

> **Por que limitamos?** Porque domain-expert e
> solutions-architect **não decidem como** — limitá-los
> força objetividade. Builders têm **autonomia total**.

---

## Implementação

**`harness/scripts/check-scope-discipline.sh`** (NOVA, opcional):
- Recebe o output de um comentário (via stdin ou arg)
- Aplica a regex heurística
- Imprime sinais detectados + recomendação
- Exit 0 (warning, não bloqueia)

**`team-manager.md` §7 (NOVA)**: rodar após cada output de
domain-expert / solutions-architect.

**`gmh agents update`** (futuro): integrar via skill
`agentic.Invocation` (não bloqueia).

---

## Quem faz o quê

| Persona | Papel |
|---|---|
| `domain-expert` | Aplica a skill `solution-scoping` ANTES de postar refinamento (checklist pré-postar) |
| `solutions-architect` | Aplica a skill `solution-scoping` ANTES de postar DoD (checklist pré-postar) |
| `team-manager` | Roda **sensor 11** depois do output; emite **recomendação** (NÃO bloqueia) |
| `quality-assurance` | Verifica no code review se as decisões seguem os pilares (sem se perder em detalhes) |

---

## Edge cases

### Domain-expert menciona "Pix"

- Pix é **contrato de pagamento** (não é tech). Pode mencionar.
- Mas "POST /api/v1/payments com payload X" é **tech**. Não pode.

### Solutions-architect menciona "Vue" ou "Nuxt"

- "Frontend em Nuxt" (stack pinada) pode mencionar.
- "Componente `CycleWizardStep1.vue` em
  `web/app/components/feature/cycles/`" (path + nome) é **tech**.
  Não pode.

### Output grande mas bem estruturado

- Se for **12 ACs** com 3-4 linhas cada + **8 edge cases** com
  5-6 linhas cada, o total fica em ~150 linhas. **OK**.
- Se for **19 ACs** com 5-6 linhas cada + 18 edge cases, o
  total fica em ~250+ linhas. **Warning**.

### Builder "reclama" que o DoD está vago

- **Bom sinal** — significa que o solutions-architect
  entregou **pilares**, não **blueprints**.
- O builder tem autonomia pra escolher **como** implementar
  dentro dos pilares.

---

## Referências

- [`../skills/solution-scoping/SKILL.md`](../skills/solution-scoping/SKILL.md) — princípio + exemplos
- [`../personas/domain-expert.template.md`](../personas/domain-expert.template.md) §"Cerca de Solução"
- [`../personas/solutions-architect.md`](../personas/solutions-architect.md) §"DoD — PILARES"
- [`../personas/team-manager.md`](../personas/team-manager.md) §7 (NOVA)
- AGENTS.md invariante 22 (NOVA)
- `harness/contrib/design-decisions.md` ADR-0021 (NOVA)
- `harness/scripts/check-scope-discipline.sh` (opcional, NOVA)
