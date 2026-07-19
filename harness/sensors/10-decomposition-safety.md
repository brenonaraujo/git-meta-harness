# Sensor 10 — Decomposition Safety

> **Objetivo:** impedir que o `team-manager` dispare 2+ builders
> em paralelo cujas sub-issues têm **path-scope overlap** (ou
> sem path-scope declarado). **Quem roda:** `team-manager`.
> **Quando:** ANTES de delegar builders, na transição
> `ready` → `in-progress`. **Falha → ação:** bloquear e
> pedir resolução (`depends-on` ou refatoração de path-scope).

---

## Por que este sensor existe

**Lição do Mandaí v2 (jul/2026, Épico #12):** durante a
decomposição do épico #12 (autenticação + role switching) em
6 sub-issues (#13–#18), o `team-manager` disparou 6 builders
em paralelo **no mesmo `cwd`**:

| Sub-issue | Builder | path-scope declarado? |
|---|---|---|
| #13 backend auth-api | backend-engineer | ❌ (sem label) |
| #15 backend user-role | backend-engineer | ❌ (sem label) |
| #14 frontend homepage | frontend-engineer | ❌ |
| #16 frontend cadastro/login | frontend-engineer | ❌ |
| #17 frontend home auth | frontend-engineer | ❌ |
| #18 infra migrations+seed | devops-engineer | ❌ |

**Resultado**:
- **#13 (auth-api) e #15 (user-role) ambos declararam interface
  `UserRepository`** no mesmo pacote (`internal/repository/`) →
  conflito de compilação.
- **#15 referenciou tipos do auditlog** (que estava criando)
  antes de #15 terminar → erro de tipo indefinido.
- **Nenhum dos 6 builders chegou a commitar** — o trabalho foi
  perdido (working tree volátil dos processos Hermes que
  terminaram por "limite de iterações" + conflito).
- **Custo**: ~4h de orquestração desperdiçada, retrabalho
  manual necessário para consolidar o que sobrou.

**Causa raiz**: o meta-harness **não tinha mecanismo para
detectar overlap de paths** entre sub-issues em paralelo. O
`team-manager` confiou em "backend e frontend em arquivos
separados" (workflow/05-orchestration.md §2), mas 2 backends
no mesmo pacote **não são "arquivos separados"**.

**Princípio**: o `team-manager` **nunca dispara 2+ builders
em paralelo sem antes validar que seus path-scopes são
disjuntos** (ou têm dependência explícita). Esta validação
é **automática** (sensor 10 + script
`harness/scripts/check-parallel-builders.sh`) e
**bloqueante** (sub-issue não vai pra `in-progress` se
falhar).

---

## Como rodar (3 passos)

### Passo 1 — `team-manager` lê path-scope de cada sub-issue em `ready`

```bash
# Lista sub-issues com path-scope
gh issue list --label ready --label path-scope --state open \
  --json number,title,labels
```

Para cada issue, extrair o glob da label `path-scope: <glob>`
(pode haver múltiplas labels `path-scope:`, uma por glob).

### Passo 2 — Calcular overlap dos path-scopes

Para cada par de sub-issues em paralelo (candidatas a
`in-progress`):

```bash
# Pseudo-código (implementado em check-parallel-builders.sh):
for each pair (A, B) of in-progress or ready-to-start issues:
  if A and B have NO depends-on relationship:
    overlap = intersect(path_scope(A), path_scope(B))
    if overlap != ∅:
      FAIL with overlap details
```

**Overlap** = 1+ arquivo real seria tocado por ambos builders.
Globs usam sintaxe `.gitignore` (mesma do `find -path`).

### Passo 3 — Bloquear OU aceitar

| Condição | Ação |
|---|---|
| Todos path-scopes **disjuntos** | ✅ Aceitar paralelização |
| Path-scopes com **overlap** + `depends-on` explícito | ✅ Aceitar (serializar pela ordem da dep) |
| Path-scopes com **overlap** + **sem** `depends-on` | 🛑 **Bloquear** e pedir resolução |
| Qualquer sub-issue **sem path-scope** | 🛑 **Bloquear** e pedir DoD completo |

---

## Script automatizado

[`harness/scripts/check-parallel-builders.sh`](../scripts/check-parallel-builders.sh)
implementa este sensor:

```bash
cd seu-projeto
./harness/scripts/check-parallel-builders.sh
```

Saída típica (caso de overlap detectado):

```
==> Checking 3 ready sub-issues for path-scope overlap...
  •   #13 (backend auth-api)        path-scope: backend/internal/auth/**, repository/auth.go
  •   #15 (backend user-role)       path-scope: backend/internal/role/**, repository/user.go
  •   #14 (frontend homepage)       path-scope: web/app/pages/home/**

⚠️  OVERLAP DETECTED between:
    #13 (backend internal/auth/**, repository/auth.go)
    #15 (backend internal/role/**, repository/user.go)
    → overlapping file: backend/internal/repository/auth.go vs user.go (same package)

❌ Action required: add `depends-on: #X` to one of them, or
   refactor path-scope to be disjoint.
```

Exit codes:
- `0` = sem overlap, pode paralelizar
- `1` = overlap detectado, bloquear
- `2` = sub-issue sem path-scope, bloquear

---

## Quem faz o quê

| Persona | Papel |
|---|---|
| `solutions-architect` | **Declara `path-scope`** no DoD da sub-issue (label `path-scope: <glob>`) |
| `team-manager` | **Roda este sensor** (passo 1+2+3) antes de `in-progress` |
| `backend-engineer` / `frontend-engineer` | **Validam** que estão dentro do path-scope; se precisarem tocar fora, pedem extensão ao `team-manager` |
| `quality-assurance` | Verifica que o PR final **respeita** o path-scope declarado (sem diffs surpresa em arquivos fora do escopo) |

---

## Edge cases

### Path-scope que cobre `package.json` ou `go.mod`

**Sempre serializar**. Mudanças em lock files invalidam caches
e podem causar conflitos sutis (ex.: versão de dep X bumpada
em uma branch mas não em outra).

### Path-scope que cobre `migrations/*.sql`

**Serializar entre si** (cada migration tem que ser sequencial
no `_up.sql` order) mas pode paralelizar com código.

### Path-scope que deleta arquivo

**Bloquear** se o arquivo é referenciado por outra sub-issue.
Use `git grep` para detectar referências.

### Path-scope que cobre testes de outra feature

**Bloquear**. Se sua feature precisa de testes de outra
feature, é porque a outra feature tem que fechar primeiro
(depends-on).

### Path-scope não declarado (label ausente)

**Bloquear sempre**. Sub-issue sem `path-scope` é DoD
incompleto — voltar para `solutions-architect`.

---

## Referências

- ADR-0019 (decomposition safety — esta ADR documenta a decisão)
- `personas/solutions-architect.md` §"Path scoping" (quem declara)
- `personas/team-manager.md` §6 (quem valida)
- `workflow/05-orchestration.md` §2 (princípio atualizado)
- `AGENTS.md` invariante 21 (obrigatoriedade)
- `harness/scripts/check-parallel-builders.sh` (automação)
