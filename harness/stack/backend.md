# Stack — Backend (Go + Gin + GORM + PostgreSQL + OpenAPI)

> **Padrão canônico** do meta-harness para microsserviços backend.
> Mudanças aqui exigem ADR em `contrib/design-decisions.md` e
> aprovação do `team-manager`.

---

## Linguagem & runtime

- **Go** 1.22+ (mínimo; preferir a LTS mais recente).
- **Módulos** Go (não GOPATH). Todo repo tem `go.mod` + `go.sum`.
- Compilar com `CGO_ENABLED=0` para imagens estáticas (scratch/distroless).

---

## Frameworks & libs (padrão)

| Camada               | Lib                                                  | Versão alvo |
|----------------------|------------------------------------------------------|-------------|
| HTTP framework       | `github.com/gin-gonic/gin`                          | v1.10+      |
| ORM                  | `gorm.io/gorm` + `gorm.io/driver/postgres`          | v1.25+ / v1.5+ |
| Migrations           | `github.com/golang-migrate/migrate/v4`              | v4.17+      |
| Validação            | `github.com/go-playground/validator/v10`            | v10+        |
| Config               | `github.com/kelseyhightower/envconfig`              | v1.4+       |
| Logging              | `log/slog` (stdlib Go 1.21+)                        | stdlib      |
| Metrics              | `github.com/prometheus/client_golang`               | v1.19+      |
| Tracing              | `go.opentelemetry.io/otel` + `go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin` | latest |
| Testes               | `testing` (stdlib) + `github.com/stretchr/testify`  | v1.9+       |
| Mocks                | `go.uber.org/mock` + `mockgen`, ou `mockery`         | latest      |
| HTTP test            | `net/http/httptest` (stdlib)                        | stdlib      |
| **i18n**             | `github.com/nicksnyder/go-i18n/v2/i18n`             | v2.5+       |
| Locale parsing       | `golang.org/x/text/language`                        | latest      |

> **Não usar** libs que não estão nesta lista sem approval do
> `solutions-architect` (e atualização desta lista com ADR).

---

## Estrutura de pastas

```
my-service/
├── cmd/
│   ├── server/main.go          # binário principal (apenas wiring)
│   ├── migrate/main.go         # binário de migration
│   └── seed/main.go            # binário de seed (one-off)
├── internal/
│   ├── app/
│   │   ├── config.go           # envconfig
│   │   ├── logger.go           # slog JSON setup
│   │   ├── metrics.go          # Prometheus registry + handler
│   │   ├── db.go               # GORM setup
│   │   └── server.go           # gin engine + middlewares
│   ├── api/
│   │   └── openapi.gen.go      # gerado de api/openapi.yaml
│   ├── handler/                # 1 endpoint = 1 func (max 25 linhas)
│   ├── service/                # regras de negócio (puro, sem gin/gorm)
│   ├── repository/             # interfaces + GORM impls
│   ├── domain/                 # entidades, erros, value objects
│   └── platform/               # adapters: postgres, httpclient, ...
├── api/
│   └── openapi.yaml            # CONTRATO — fonte da verdade
├── migrations/                 # *.sql versionados (golang-migrate)
├── deploy/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── .env.example
├── test/
│   ├── integration/            # build tag `integration`
│   ├── smoke/
│   └── load/                   # gatling
├── scripts/
│   ├── check-twelve-factor.sh
│   └── ...
├── .golangci.yml
├── go.mod
├── go.sum
└── README.md
```

---

## Camadas e dependências

```
handler  →  service  →  repository  →  DB
  ↓           ↓            ↓
  └───────────┴────────────┴──→ platform (httpclient, broker, ...)
```

- **Handler** não tem regra de negócio; só binding (request → DTO),
  chamada ao service, render (response).
- **Service** é puro: recebe interfaces, retorna DTOs/erros. Não
  importa `gin` nem `gorm`.
- **Repository** define interface em `internal/repository` e a
  implementação GORM em `internal/platform/postgres`. Permite trocar
  o DB ou mockar em testes.
- **Domain** é o núcleo: entidades, value objects, erros de domínio.
  Zero dependências externas.

> **Testes de service** usam **fakes in-memory** do repository
> (map), não mocks do GORM. Testes de repository usam **testcontainers
> com Postgres real** (build tag `integration`).

---

## Configuração (env)

Toda config via `envconfig`. Nada hardcoded. Nada em YAML/JSON em
runtime.

```go
// internal/app/config.go
package app

import "github.com/kelseyhightower/envconfig"

type Config struct {
    Port        string `envconfig:"PORT" default:"8080"`
    DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
    LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
    GinMode     string `envconfig:"GIN_MODE" default:"release"`
    MetricsPath string `envconfig:"METRICS_PATH" default:"/metrics"`
    OtelEndpoint string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" required:"false"`
}

func LoadConfig() (*Config, error) {
    var c Config
    if err := envconfig.Process("", &c); err != nil {
        return nil, err
    }
    return &c, nil
}
```

---

## API contract (OpenAPI spec-first)

1. `solutions-architect` edita `api/openapi.yaml` **antes** do
   `backend-engineer` codificar.
2. `backend-engineer` regenera: `make oas` (roda `oapi-codegen`).
3. Implementa contra os tipos/handlers gerados.
4. **Não** anota com comments (não usamos swag); o **spec vem
   primeiro**.

```yaml
# api/openapi.yaml
openapi: 3.1.0
info:
  title: Auth Service
  version: 0.1.0
paths:
  /api/v1/auth/login:
    post:
      operationId: Login
      tags: [auth]
      summary: Login with email and password
      description: |
        Authenticates a user and returns a JWT token.
        Returns 401 on invalid credentials.

        **i18n:** error responses are localized based on the
        `Accept-Language` header. Supported: `en`, `pt-BR`, `es`.
        Default: `en`. The `message` field is human-readable; the
        `message_key` field is a stable key clients can map to their
        own strings.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
components:
  schemas:
    LoginRequest:
      type: object
      required: [email, password]
      properties:
        email: { type: string, format: email }
        password: { type: string, minLength: 8 }
    LoginResponse:
      type: object
      required: [token, expires_at]
      properties:
        token: { type: string }
        expires_at: { type: string, format: date-time }

    Error:
      type: object
      required: [code, message, message_key]
      properties:
        code:        { type: string, description: 'Machine-readable code' }
        message:     { type: string, description: 'Localized human message' }
        message_key: { type: string, description: 'Stable i18n key' }
        details:     { type: object, additionalProperties: true }
```

---

## Logging (slog JSON)

```go
// internal/app/logger.go
package app

import (
    "log/slog"
    "os"
)

func NewLogger(level string) *slog.Logger {
    var lvl slog.Level
    _ = lvl.UnmarshalText([]byte(level))
    h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level:     lvl,
        AddSource: false,
    })
    return slog.New(h)
}
```

Uso:

```go
slog.Info("user logged in",
    "user_id", user.ID,
    "request_id", reqID,
    "duration_ms", time.Since(start).Milliseconds(),
)
```

> Nunca `fmt.Println` em código de produção. Nunca log em arquivo
> (XII-factor XI). Sempre JSON em stdout.

---

## i18n (internacionalização)

> **Toda string visível ao usuário** (mensagens de erro da API,
> subject/body de e-mail, SMS, push) **deve** passar por
> `internal/i18n`. Idiomas obrigatórios: **en**, **pt-BR**, **es**.
> Ver skill [`../skills/i18n.md`](../skills/i18n.md) e sensor
> [`../sensors/08-i18n-audit.md`](../sensors/08-i18n-audit.md).

### Estrutura

```
internal/i18n/
├── bundle.go          # carrega os bundles
├── translate.go       # helper de tradução
└── locales/
    ├── en.json
    ├── pt-BR.json
    └── es.json
```

### Helper de tradução (uso)

```go
// Em qualquer handler, depois do middleware de localizer:
c.JSON(http.StatusUnauthorized, api.Error{
    Code:    "invalid_credentials",
    Message: i18n.T(c, "auth.invalid_credentials"),
})

// Com interpolação:
c.JSON(http.StatusOK, api.Success{
    Message: i18n.T(c, "user.created", "name", user.Name),
})
```

### Convenção de chaves

`<domínio>.<ação>.<contexto>` (ex.: `auth.invalid_credentials`,
`user.delete_blocked`, `validation.invalid_email`). **NUNCA** chaves
planas (`error1`, `msg`, `unauthorized`).

### Interpolação (Go)

Use `{{.variavel}}` (sintaxe do `go-i18n`).

```json
// en.json
"login_success": "Welcome back, {{.name}}!"
```

```json
// pt-BR.json
"login_success": "Bem-vindo de volta, {{.name}}!"
```

> **O nome da variável deve ser idêntico** em todos os idiomas.

### Seleção de idioma

1. Header `Accept-Language` da request (RFC 7231).
2. Fallback: env `DEFAULT_LOCALE` (default: `en`).
3. Cache: client pode enviar `?lang=pt-BR` para override explícito.

```go
// parseAcceptLanguage usando golang.org/x/text/language
tags, _, _ := language.ParseAcceptLanguage(c.GetHeader("Accept-Language"))
if len(tags) == 0 {
    return defaultLang
}
return tags[0].String()
```

### Testes (obrigatório)

Cada handler que retorna mensagem i18n **deve** ter pelo menos 1
teste para `en` e 1 para `pt-BR` (ver `sensors/08-i18n-audit.md`).

```go
func TestLogin_InvalidCredentials_PTBR(t *testing.T) {
    c.Request.Header.Set("Accept-Language", "pt-BR")
    i18n.WithLocalizer(c, "pt-BR")
    // ...
    require.Contains(t, w.Body.String(), "E-mail ou senha inválidos")
}
```

### Proibições

- ❌ `c.JSON(..., gin.H{"message": "Email inválido"})` — string
  hardcoded.
- ❌ `fmt.Errorf("user %s not found", id)` retornado em handler — o
  erro vai para o usuário; precisa ser chave i18n.
- ❌ Chave plana: `error1`, `unauthorized` (sem namespace).
- ❌ Variável de interpolação diferente entre idiomas.
- ❌ Idioma faltando em `locales/`.

---

## Metrics (Prometheus)

Métricas **obrigatórias** (registradas em `internal/app/metrics.go`):

```go
package app

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    DBQueriesTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_queries_total",
            Help: "Total DB queries",
        },
        []string{"operation", "table", "status"},
    )

    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "DB query duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation", "table"},
    )

    AppInfo = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "app_info",
            Help: "App info (always 1)",
        },
        []string{"version", "commit", "go_version"},
    )
)
```

Métricas expostas em `:8080/metrics`.

---

## Health & readiness

```go
// internal/app/server.go
r.GET("/healthz", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})

r.GET("/readyz", func(c *gin.Context) {
    if err := db.Ping(); err != nil {
        c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "ready"})
})
```

---

## Graceful shutdown

```go
ctx, stop := signal.NotifyContext(context.Background(),
    syscall.SIGINT, syscall.SIGTERM)
defer stop()

go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        slog.Error("server failed", "error", err)
        os.Exit(1)
    }
}()

<-ctx.Done()
slog.Info("shutting down")

shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := srv.Shutdown(shutdownCtx); err != nil {
    slog.Error("forced shutdown", "error", err)
}
```

---

## Testes (TDD)

- **Testes unitários** no mesmo package, cobrindo **bordas**.
- **Table-driven** com `testify/require` (precondições) e
  `testify/assert` (validações).
- **Mocks** apenas nas fronteiras externas (DB, HTTP client,
  broker). Use fakes in-memory quando possível.
- **Testes de integração** com build tag `integration` (testcontainers).
- **Coverage** ≥ 80% de branch nos pacotes alterados.

```go
// internal/service/auth_test.go
package service_test

import (
    "context"
    "errors"
    "testing"

    "github.com/stretchr/testify/require"

    "my-service/internal/repository"
    "my-service/internal/service"
)

type fakeUserRepo struct {
    users map[string]repository.User
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (repository.User, error) {
    u, ok := f.users[email]
    if !ok {
        return repository.User{}, repository.ErrNotFound
    }
    return u, nil
}

func TestAuthService_Login(t *testing.T) {
    tests := []struct {
        name      string
        repo      repository.UserRepository
        email     string
        password  string
        wantErr   error
    }{
        {
            name: "valid credentials",
            repo: &fakeUserRepo{users: map[string]repository.User{
                "user@example.com": {Email: "user@example.com", PasswordHash: "$2a$10$..."},
            }},
            email:    "user@example.com",
            password: "secret123",
            wantErr:  nil,
        },
        {
            name:    "user not found",
            repo:    &fakeUserRepo{users: map[string]repository.User{}},
            email:   "missing@example.com",
            password: "secret123",
            wantErr:  service.ErrInvalidCredentials,
        },
        // ... mais casos
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := service.NewAuthService(tt.repo, nil)
            _, err := svc.Login(context.Background(), tt.email, tt.password)
            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## Migrations (golang-migrate)

```
migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_add_user_status.up.sql
└── 000002_add_user_status.down.sql
```

Cada migration tem `up` e `down`. CI roda `migrate up` em ambiente
isolado para validar.

---

## Comandos canônicos (Makefile)

```makefile
.PHONY: tidy build test lint vuln oas migrate-up run docker compose-up

tidy:
	go mod tidy

build:
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate

test:
	go test -race -shuffle=on -coverprofile=coverage.out -covermode=atomic ./...

lint:
	golangci-lint run --timeout=5m ./...

vuln:
	govulncheck ./...

oas:
	oapi-codegen --config=codegen.yaml api/openapi.yaml

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

run:
	go run ./cmd/server

docker:
	docker build -f deploy/Dockerfile -t my-service:dev .

compose-up:
	docker compose -f deploy/docker-compose.yml up -d --build

compose-down:
	docker compose -f deploy/docker-compose.yml down
```

---

## Anti-padrões (proibidos)

- ❌ `gin.SetMode(gin.ReleaseMode)` no código (ler de env).
- ❌ `log` ou `fmt.Println` em produção (use `slog`).
- ❌ `os.Exit(1)` no meio de handler.
- ❌ Comentários redundantes em código.
- ❌ Funções > 25 linhas.
- ❌ Arquivos > 150 linhas.
- ❌ `panic` em código de produção (exceto init).
- ❌ `interface{}` sem justificativa (use generics ou `any`).
- ❌ Config em YAML/JSON (só env).
- ❌ Libs fora da lista sem ADR.
- ❌ ORM sem interface no repository (não dá pra mockar).
- ❌ `swag` (anotações em comments) — usar spec-first com `oapi-codegen` ou `ogen`.
- ❌ Commits diretos na main.
- ❌ PR sem "Como testar localmente".
