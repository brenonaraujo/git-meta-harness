# Stack — Code Style (KISS, DRY, clean, ≤25/≤150)

> Padrão de estilo e limites rígidos. **Inegociáveis**.

---

## Princípios (ordem de prioridade)

1. **KISS** — simples > clever. Padrões conhecidos > exóticos.
2. **DRY** — uma única fonte da verdade. Nada duplicado.
3. **Clean code** — legibilidade > esperteza.
4. **Twelve-factor** — ver `bootstrap.md` §7 e
   `sensors/07-twelve-factor-audit.md`.
5. **YAGNI** — You Aren't Gonna Need It. Não implemente o que não foi
   pedido.

---

## Limites rígidos

| Métrica                       | Limite            | Enforcement                              |
|-------------------------------|-------------------|------------------------------------------|
| Linhas por função             | **≤ 35 (max) / 25 (recomendado)** | `funlen` no golangci-lint (v1.10.0: limite duro subiu de 25 → 35) |
| Linhas por arquivo `.go`      | **≤ 150**         | convenção + script `wc -l` no CI         |
| Linhas por componente Vue     | **≤ 200**         | convenção + `vue-tsc` + reviewer         |
| Complexidade ciclomática      | **≤ 15**          | `gocyclo`                                |
| Parâmetros por função         | **≤ 5**           | `revive` (`argument-limit`)              |
| Aninhamento de blocos         | **≤ 4**           | `gocritic` (`nestreturn`)                |
| Cobertura de testes (branch)  | **≥ 80%**         | CI                                      |
| Comentários por arquivo       | **0 redundantes** | revisão + `gocritic` (`commentedOutCode`) |
| `any` em TS                   | **0 sem justificativa** | ESLint + revisão                    |
| `panic` em produção           | **0**             | `gocritic` + revisão                     |
| `fmt.Println` em produção     | **0**             | `gocritic` + revisão                     |

---

## Comentários

### O que **NÃO** comentar

```go
// ❌ ERRADO — comentário redundante
// Incrementa o contador
counter++

// ❌ ERRADO — explica o que o código faz (deveria ser óbvio)
// Loop que itera sobre os usuários
for _, u := range users {
    // ...
}
```

### O que **PODE** comentar

```go
// ✅ CORRETO — explica o porquê (decisão não-óbvia)
// We retry up to 3 times because the upstream API has
// transient failures during peak hours (see incident #42).
for i := 0; i < 3; i++ {
    // ...
}

// ✅ CORRETO — TODO com dono e contexto
// TODO(@backend-engineer): migrate to OpenTelemetry
// once the platform team finishes the OTLP collector setup.

// ✅ CORRETO — godoc em exports (documentação pública)
// Login authenticates a user and returns a JWT token.
func Login(ctx context.Context, email, password string) (string, error) {
    // ...
}
```

### Regras

- **Nada de "WHAT"** (o código diz o que).
- **"WHY" quando relevante** (decisão não-óbvia, link para issue/ADR).
- **TODO com dono**: `// TODO(@<username>): <o que> — <contexto>`.
- **Godoc em exports**: comentário acima de func/type público em Go,
  JSDoc/TSDoc em funções/componentes exportados em TS/Vue.
- **Nada de "seção"** (`// ============ Helpers ===========`).

---

## Nomenclatura

### Go

| Elemento         | Convenção                  | Exemplo                       |
|------------------|----------------------------|-------------------------------|
| Package          | lowercase, sem underscore  | `package auth`                |
| Função pública   | PascalCase                 | `Login`, `GetUserByID`        |
| Função privada   | camelCase                  | `validateEmail`               |
| Variável         | camelCase                  | `userID`, `retryCount`        |
| Constante        | PascalCase ou camelCase    | `MaxRetries` ou `maxRetries`  |
| Erro             | `Err<Contexto>`            | `ErrInvalidCredentials`       |
| Interface        | sufixo `-er` (1 método) ou nome descritivo | `UserRepository` |
| Struct           | PascalCase                 | `AuthService`                 |
| Acronyms         | todos em mesma case        | `http` → `HTTP` (ID → ID)     |

### TypeScript / Vue

| Elemento         | Convenção                  | Exemplo                       |
|------------------|----------------------------|-------------------------------|
| Arquivo .vue     | PascalCase                 | `LoginForm.vue`               |
| Arquivo .ts      | camelCase                  | `useAuth.ts`, `formatDate.ts` |
| Componente       | PascalCase                 | `LoginForm`                   |
| Composable       | `use<Name>`                | `useAuth`, `useFetch`         |
| Store            | `use<Name>Store`           | `useAuthStore`                |
| Tipo             | PascalCase                 | `LoginRequest`, `UserInfo`    |
| Interface        | sem prefixo `I`            | `UserRepository`              |
| Constante        | UPPER_SNAKE ou camelCase   | `MAX_RETRIES` ou `maxRetries` |
| Enum             | PascalCase; valores PascalCase | `OrderStatus.Pending`     |

---

## Funções

### Tamanho (v1.10.0)

**Recomendado: ≤ 25 linhas. Máximo: ≤ 35 linhas.**

| Faixa | Status | Ação |
|---|---|---|
| 0-25 linhas | ✅ Ideal | Manter assim |
| 26-35 linhas | ⚠️ Aceitável | Considere se a função é coesa; se for, documente o porquê; se não, decomponha |
| 36+ linhas | ❌ Erro | `funlen` falha. Refatore OBRIGATORIAMENTE (extraia helpers, decompose por responsabilidade) |

**Antes de implementar** (skill `pre-implementation-design`): liste
2-3 decomposições possíveis da função que vai implementar e justifique
a escolha. **Pular essa etapa leva a funções que ficam em 26-35 linhas
"por acidente" e a abstrações desnecessárias só pra caber em 25.**

**Quando 26-35 é OK**:
- Função é coesa (1 responsabilidade clara)
- Helpers extraídos tornariam o código mais difícil de seguir
- Exemplos: `OnboardRole` (pipeline de criação de user+role+audit),
  `ProcessPayment` (múltiplas etapas que juntas formam 1 transação)

**Quando 26-35 NÃO é OK**:
- Função faz 2+ coisas distintas → decompor
- Tem `for` aninhado, múltiplos `if err != nil` em sequência → extrair
- Linhas "de glue" (chamadas sequenciais sem lógica) → extrair para helper

### Parâmetros

≤ 5. Se passar, agrupe em struct.

```go
// ❌ Ruim
func CreateUser(name, email, password, role, tenantID string, active bool) error

// ✅ Bom
type CreateUserParams struct {
    Name     string
    Email    string
    Password string
    Role     string
    TenantID string
    Active   bool
}
func CreateUser(ctx context.Context, p CreateUserParams) error
```

### Retorno

- **Sempre** retorne `error` por último.
- Use `errors.Is` / `errors.As` (não comparação direta com `==`).
- Erros de domínio: declare em `internal/domain/errors.go`.

```go
var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserNotFound       = errors.New("user not found")
)
```

### Pureza

- **Service** é puro: recebe `context.Context` e entradas, retorna
  resultado. Não importa `gin`, `gorm`, `*sql.DB`.
- **Side effects** (DB, HTTP, log) ficam no **handler** (input/output)
  ou no **repository** (persistência).

---

## Estrutura de função (early return)

```go
// ✅ Preferir early return
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
    if email == "" {
        return "", ErrInvalidEmail
    }
    if password == "" {
        return "", ErrInvalidPassword
    }

    user, err := s.repo.GetByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return "", ErrInvalidCredentials
        }
        return "", fmt.Errorf("get user: %w", err)
    }

    if !checkPasswordHash(password, user.PasswordHash) {
        return "", ErrInvalidCredentials
    }

    token, err := s.jwt.Sign(user.ID)
    if err != nil {
        return "", fmt.Errorf("sign jwt: %w", err)
    }
    return token, nil
}

// ❌ Evitar
func (s *AuthService) Login(...) (string, error) {
    if email != "" {
        if password != "" {
            user, err := s.repo.GetByEmail(...)
            if err == nil {
                if checkPasswordHash(...) {
                    // ... happy path no fundo do aninhamento
                }
            }
        }
    }
    return "", ErrInvalidCredentials
}
```

---

## Erros

### Wrap

```go
// ✅ Wrap com contexto
if err := db.Query(...); err != nil {
    return fmt.Errorf("query users: %w", err)
}

// ❌ Perder o erro original
if err := db.Query(...); err != nil {
    return errors.New("query failed")
}
```

### Logging

```go
// ✅ Log com contexto
slog.Error("failed to query users",
    "operation", "GetByEmail",
    "email", email,
    "error", err.Error(),
)

// ❌ Log sem contexto
log.Println(err)
```

---

## Imports

### Go

```go
import (
    // stdlib
    "context"
    "errors"
    "fmt"

    // 3rd party
    "github.com/gin-gonic/gin"
    "github.com/kelseyhightower/envconfig"

    // internal
    "my-service/internal/domain"
    "my-service/internal/repository"
)
```

Use `goimports -w .` para organizar automaticamente.

### TypeScript

- Use auto-imports do Nuxt (não importe manualmente).
- Para `~/`, `@/`, etc., configure em `tsconfig.json`.
- ESLint com `@nuxt/eslint` resolve ordem automaticamente.

---

## Testes

### Padrão

```go
// ✅ Table-driven
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {name: "valid", input: "user@example.com", wantErr: false},
        {name: "empty", input: "", wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### require vs assert

```go
// require: para precondições (parar o teste se falhar)
require.NoError(t, err)
require.NotNil(t, user)

// assert: para validações (rodar todas, ver todas as falhas)
assert.Equal(t, expected, actual)
assert.Len(t, users, 3)
```

### Fakes vs Mocks

```go
// ✅ Preferir fakes in-memory para repository
type fakeUserRepo struct {
    users map[string]User
}
func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (User, error) {
    u, ok := f.users[email]
    if !ok {
        return User{}, ErrNotFound
    }
    return u, nil
}

// ⚠️ Mock apenas para fronteiras externas
// (HTTP client, broker, ...)
```

---

## Componentes Vue

### Ordem

1. `<script setup>`
2. `<template>`
3. `<style>` (se houver; scoped, sem `<style global>`)

### Props tipadas

```ts
// ✅ Sempre tipar
const props = defineProps<{
  user: User
  isLoading?: boolean
}>()
const emit = defineEmits<{
  submit: [data: LoginInput]
}>()
```

### Composables para lógica

```ts
// ❌ Lógica no componente
<script setup>
const email = ref('')
const password = ref('')
const error = ref<string | null>(null)
const isLoading = ref(false)

async function submit() {
  if (!email.value) {
    error.value = 'Email required'
    return
  }
  // ...
}
</script>

// ✅ Lógica no composable
<script setup>
const { email, password, error, isLoading, submit } = useLoginForm()
</script>
```

---

## Git

- Mensagens em **inglês**, formato **Conventional Commits**.
- 1 commit por mudança lógica (squash antes do PR).
- Body explica o **porquê** (não o quê).

```
feat(auth): implementa login com JWT (Refs #42)

- Adiciona endpoint POST /api/v1/auth/login
- Gera token JWT com TTL configurável via env
- Métricas auth_login_total e auth_login_duration_seconds
- Cobertura 92% no pacote auth/
- Audit 12-factor: F1..F12 OK
```

---

## Anti-padrões (resumo)

- ❌ Comentários redundantes.
- ❌ Funções > 35 linhas (v1.10.0: limite duro subiu de 25 → 35; recomendado: 25).
- ❌ Arquivos > 150 linhas.
- ❌ Complexidade > 15.
- ❌ `panic` em produção.
- ❌ `fmt.Println` em produção.
- ❌ `any` sem justificativa.
- ❌ Config hardcoded.
- ❌ Log em arquivo.
- ❌ Duplicação de código.
- ❌ Singleton global.
- ❌ Mutação direta de state fora de actions (Pinia).
- ❌ `swag` (anotações em comments) — spec-first.
- ❌ Commits direto na main.
