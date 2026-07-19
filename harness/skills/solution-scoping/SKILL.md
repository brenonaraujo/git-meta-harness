---
name: solution-scoping
version: 1.0.0
type: orchestration
applies-to: domain-expert, solutions-architect, team-manager
---

# Solution Scoping — Pilares, não Blueprints

Skill for **domain-expert** and **solutions-architect** personas.
Forces them to stay in their layer (POR QUÊ / PILARES) and **not**
drift into implementation details (COMO).

> **Lição do Mandaí v2 (jul/2026, Épico F4+F5 — Ciclos + Pedidos):**
> o `domain-expert` e o `solutions-architect` escreveram
> **19+ ACs**, **18 edge cases**, e um DoD de **150+ linhas**
> com nomes de funções específicas (`MustGenerateCycleSlug`,
> `CheckCycleLimits`), SQL (`SELECT FOR UPDATE`), ORMs
> (`gorm`, `pgx`), paths de arquivos (`backend/internal/...`)
> e métricas Prometheus. O `backend-engineer` virou **executor
> cego** — só seguiu o blueprint, sem questionar, sem
> otimizar, sem ownership técnica.
>
> **Princípio v1.11.0 (ADR-0021)**: domain-expert e
> solutions-architect entregam **PILARES** (o que + por quê +
> invariantes). O **builder** entrega **BLUEPRINTS** (como).

---

## 🚦 Princípio central: PILARES vs BLUEPRINTS

| Camada | Quem | Entrega | Tamanho recomendado | Tokens máx (~recomendado) |
|---|---|---|---|---|
| **Negócio** | `domain-expert-<x>` | Comportamento + regras + edge cases | ≤ 12 ACs, ≤ 8 edge cases | ~3k tokens |
| **Arquitetura** | `solutions-architect` | 3-5 **pilares** + DoD macro + 12-factor audit | DoD ≤ 80 linhas, pilares ≤ 5 | ~5k tokens |
| **Implementação** | `backend-engineer` / `frontend-engineer` | **Tudo** o que for necessário | (sem limite razoável) | (sem limite) |

> **Recomendação (não-bloqueante)**: outputs acima de 30k tokens
> (~75k chars) devem ser encurtados. O sensor 11 (`scope-discipline`)
> **NÃO bloqueia** — apenas recomenda objetividade.

---

## ✅ vs ❌ — exemplos por categoria

### 1. Pricing (snapshot de preço)

✅ **PILAR (domain-expert)**:
> "O preço cobrado é o do momento em que o produto foi incluído
> no ciclo. Alterações posteriores no catálogo não se propagam
> para pedidos já feitos."

❌ **BLUEPRINT (vazou de camada)**:
> "O preço fica em `cycle_products.price_cents INTEGER NOT NULL`
> e é copiado para `order_items.unit_price_cents` no momento
> do INSERT em orders. O OrderService.CreateOrder faz
> `INSERT INTO order_items (order_id, product_id, quantity, unit_price_cents) VALUES (..., cycle_product.price_cents, ...)`."

### 2. Limite de pedidos (anti-fraude)

✅ **PILAR (domain-expert)**:
> "Morador não pode ter mais que R$ 500 em pedidos ativos+pagos
> no mesmo ciclo. Ciclo não pode ter mais que 200 pedidos no
> total."

❌ **BLUEPRINT (vazou de camada)**:
> "Em `order_service.go`, função `CheckCycleLimits(cycleID, residentID, totalCents)`:
> ```go
> sum := repo.SumActiveByResidentCycle(...)
> if sum + totalCents > 50000 { return ErrOrderLimitExceeded }
> count := repo.CountByCycle(cycleID)
> if count >= 200 { return ErrCycleOrderLimitExceeded }
> ```"

### 3. State machine (transições de ciclo)

✅ **PILAR (solutions-architect)**:
> "Pilar: **transições de estado explícitas e validadas**. Um
> ciclo tem estados discretos; certas transições são
> inválidas (ex.: fulfilled → draft) e devem ser rejeitadas
> com mensagem clara."

❌ **BLUEPRINT (vazou de camada)**:
> "Em `internal/domain/cycle.go`:
> ```go
> func (c *Cycle) CanTransition(to CycleStatus) error {
>   if c.Status == Draft && to == Open { return nil }
>   if c.Status == Open && to == Closed { return nil }
>   // ... etc
>   return ErrCycleTransitionInvalid
> }
> ```"

### 4. Idempotência (webhook Pix)

✅ **PILAR (domain-expert)**:
> "Confirmação de pagamento que chega mais de uma vez para o
> mesmo pedido é tratada como no-op: nenhum efeito duplicado,
> nenhum repasse duplicado."

❌ **BLUEPRINT (vazou de camada)**:
> "`orderRepo.UpdateStatus(id, Paid)` deve verificar
> `WHERE status='pending'` antes de UPDATE. Se já for
> 'paid', retorna 0 rows affected e o handler retorna 200
> OK (idempotente)."

### 5. Compliance (LGPD / BACEN)

✅ **PILAR (domain-expert)**:
> "Pedidos vinculam morador e ciclo. O histórico detalhado
> segue o limite de 90 dias (não-objetivo do domínio)."

❌ **BLUEPRINT (vazou de camada)**:
> "Tabela `orders` tem FK para `users.id` e `cycles.id` com
> `ON DELETE SET NULL` no user_id. Cron job diário em
> `cmd/anonymize.go` remove `resident_id` de orders > 90
> dias, mantendo `total_cents` para fins contábeis."

### 6. Slug uniqueness

✅ **PILAR (solutions-architect)**:
> "Pilar: **slug único** por ciclo, sem 5xx em colisão. Gera
> versão alternativa se a base colidir."

❌ **BLUEPRINT (vazou de camada)**:
> "Em `cycle_service.go`, `MustGenerateCycleSlug(name string)`
> faz retry com sufixo `-2`, `-3`, ..., `-5`. Usa
> `nanoid(12)` da lib `internal/service/nanoid.go`. Após 5
> retries, retorna `ErrCycleSlugsExhausted`."

---

## 🚧 Regras por persona

### `domain-expert`

**FAZ** (comportamento + regra de negócio + edge cases):
- ✅ "O usuário precisa **confirmar a exclusão** antes de executá-la"
- ✅ "Total de pedidos ativos+pagos por morador ≤ R$ 500 por ciclo"
- ✅ "Estado do pedido: pending → paid → fulfilled/cancelled"
- ✅ "Confirmação de Pix duplicada é no-op (idempotência)"

**NÃO FAZ** (é camada de outro):
- ❌ Nomes de tabelas (`cycle_products`, `order_items`)
- ❌ Nomes de migrations (`000009_cycles.up.sql`)
- ❌ Nomes de funções (`MustGenerateCycleSlug`, `CheckCycleLimits`)
- ❌ Paths de arquivo (`internal/service/cycle_service.go`)
- ❌ Nomes de ORMs/bancos (`gorm`, `pgx`, `PostgreSQL`)
- ❌ HTTP endpoints (`POST /api/v1/cycles`)
- ❌ Schemas JSON / payload shapes
- ❌ Métricas Prometheus específicas
- ❌ SQL / queries / índices

> **Se o domain-expert precisa mencionar tecnologia pra
> cumprir regulamentação** (ex.: "carimbo de tempo ICP-Brasil
> para assinatura digital"), pode — mas o quê/por quê, não
> como implementar.

### `solutions-architect`

**FAZ** (pilares + DoD macro + trade-offs):
- ✅ 3-5 **pilares arquiteturais** em alto nível
  (ex.: "consistência de preço via snapshot do momento de
  inclusão")
- ✅ DoD macro: "Cobertura ≥ 80% nos pacotes alterados",
  "Migrations idempotentes", "Métricas operacionais adicionadas"
- ✅ Trade-offs de alto nível: "Snapshot em 2 lugares vs.
> complexidade de JOIN temporal — aceito: snapshot é barato,
> atende LGPD"
- ✅ 12-factor audit (12 checks)
- ✅ Decisões em ADR-lite (formato: contexto, decisão,
> trade-off, reverter quando)

**NÃO FAZ** (é camada do builder):
- ❌ Nomes de funções específicas
- ❌ Implementações em pseudocódigo/Go/Vue
- ❌ Paths de arquivo
- ❌ HTTP endpoints específicos (paths, query params)
- ❌ Schemas OpenAPI completos
- ❌ Métricas Prometheus específicas
- ❌ Decisões de query/index/schema detalhadas

> **O builder tem autonomia** pra escolher:
> - Linguagem (Go, Rust, Node, etc — desde que siga a stack)
> - Framework web (gin, echo, chi, fiber)
> - ORM (gorm, sqlx, sqlc, raw pgx)
> - Schema SQL exato (nomes de colunas, tipos, índices)
> - Estrutura de arquivos (camadas, packages)

---

## 🛑 Detector de vazamento (sensor 11)

O `team-manager` roda o sensor 11 (`scope-discipline`) depois
de cada output de `domain-expert` e `solutions-architect`. O
sensor detecta menções técnicas e **emite recomendação** (não
bloqueia) para encurtar.

**Sinais de vazamento** (regex heurística):

| Pattern | Quem | Por quê |
|---|---|---|
| `\b(SELECT|INSERT|UPDATE|DELETE|WHERE|FROM)\b` em maiúsculas | domain-expert | SQL é implementação |
| `\b(gorm|pgx|sqlx|sqlc|gin|echo|chi|fiber|nestjs|express)\b` | domain-expert | ORM/framework é tech |
| `\b(GORM|PGx|Sqlx|Gin|...)\b` (PascalCase) | domain-expert | Type/struct name |
| `internal/[a-z]+/` ou `backend/[a-z]+/` | domain-expert | Path de arquivo |
| `\.go:\d+` ou `\.go\b` | domain-expert | Linguagem Go |
| `00000[0-9]_.*\.up\.sql` | domain-expert | Nome de migration |
| `(GET|POST|PUT|PATCH|DELETE) /api` | domain-expert | Endpoint HTTP |
| `camelCase` ou `PascalCase` (funções Go) | domain-expert | Nome de função |
| `metrics.NewCounter\|prometheus.NewCounter` | domain-expert | Métrica específica |
| Output > 30k tokens (~75k chars) | ambos | Falta objetividade |

**Quando o sensor dispara** (warning não-bloqueante):
```
⚠️  scope-discipline: output do domain-expert tem 47k tokens
    (recomendado: ≤ 30k). Sugestão: reformule em comportamento
    puro, sem nomes de tabelas/funções/paths. Ver
    harness/skills/solution-scoping/SKILL.md.
```

> O builder **segue o que está escrito** mesmo se o sensor
> recomendar encurtar. A recomendação é pra **próxima iteração**.

---

## 📏 Limites recomendados (não-bloqueantes)

| Persona | Output máx (recomendado) | Output tokens (recomendado) | Se passar |
|---|---|---|---|
| `domain-expert` | ~150 linhas / ~10k chars | ~3k tokens | ⚠️ warning, encurte próximo |
| `solutions-architect` | ~200 linhas / ~13k chars | ~5k tokens | ⚠️ warning, encurte próximo |
| `team-manager` | sem limite | sem limite | (orquestrador) |
| `backend-engineer` | sem limite | sem limite | (builder — autonomia total) |
| `frontend-engineer` | sem limite | sem limite | (builder — autonomia total) |

> **Por que limitamos domain-expert e solutions-architect?**
> Porque eles **não decidem como** — limitá-los força
> objetividade. Builders têm **autonomia total** porque eles
> **decidem como**.

---

## 📋 Checklist pré-postar (domain-expert)

Antes de postar o refinamento, verifique:

- [ ] **ACs descrevem comportamento** (não UI, não tech)
- [ ] **ACs passam o teste "e se a stack mudar?"** (stack-agnostic)
- [ ] **ACs ≤ 12** (se mais, está detalhando demais)
- [ ] **Edge cases ≤ 8** (cada um descreve 1 cenário de domínio)
- [ ] **Não menciono personas pelo nome**
- [ ] **Não escolho tecnologia** (linguagem, framework, banco, API, ORM)
- [ ] **Não escrevo código, SQL, OpenAPI yaml**
- [ ] **Não menciono nomes de tabelas/migrations/funções/paths**
- [ ] **Output total ≤ 30k tokens** (se mais, encurte — sensor avisa)

## 📋 Checklist pré-postar (solutions-architect)

- [ ] **DoD ≤ 80 linhas no comentário** (pilares + 12-factor)
- [ ] **Pilares ≤ 5** (cada um em alto nível, sem SQL/paths/funções)
- [ ] **Trade-offs explícitos** (decisão + por quê + reverter quando)
- [ ] **12-factor audit** (12 checks, não pula nenhum)
- [ ] **Não escrevo código Go/Vue/SQL/pseudocódigo**
- [ ] **Não dou nomes de funções/paths/migrations**
- [ ] **Não dou schemas OpenAPI completos** (apenas referencio que B vai fazer)
- [ ] **Output total ≤ 30k tokens** (se mais, encurte — sensor avisa)

---

## 🔗 Quem detecta / Quem valida

| Persona | Papel |
|---|---|
| `domain-expert` | Aplica esta skill antes de postar refinamento |
| `solutions-architect` | Aplica esta skill antes de postar DoD |
| `team-manager` | Roda **sensor 11** (scope-discipline) após cada output; emite **recomendação** (não bloqueia) |
| `quality-assurance` | Verifica no code review se as decisões seguem os pilares (sem se perder em detalhes) |

## 🔗 Quem é afetado

- **`builder` (backend/frontend)**: **ganha autonomia** — não tem
  que seguir blueprint cego, decide o como dentro dos pilares.
- **`user` (Brenon)**: recebe DoD mais limpo, focado em
  comportamento + pilares, sem se perder em SQL/functions.

---

## Referências

- [`../domain-refinement/SKILL.md`](../domain-refinement/SKILL.md) — cercas do domain-expert (UI + tech)
- [`../personas/domain-expert.template.md`](../personas/domain-expert.template.md) — cerca de Design + cerca Técnica
- [`../personas/solutions-architect.md`](../personas/solutions-architect.md) — DoD + 12-factor
- [`../sensors/11-scope-discipline.md`](../sensors/11-scope-discipline.md) — sensor não-bloqueante
- [`../contrib/design-decisions.md`](../contrib/design-decisions.md) ADR-0021 — decisão
- AGENTS.md invariante 22 — scope discipline (recomendação, não-bloqueante)
