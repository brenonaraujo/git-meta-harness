# Stack — Observability (Prometheus + slog + OpenTelemetry)

> Padrão de observability do meta-harness. Todo microsserviço Go
> **deve** ter Prometheus + slog JSON. Tracing é recomendado.

---

## Os 3 pilares

| Pilar     | Ferramenta           | Obrigatório?       |
|-----------|----------------------|--------------------|
| Métricas  | Prometheus           | **SIM**            |
| Logs      | `slog` (JSON, stdout)| **SIM**            |
| Tracing   | OpenTelemetry + OTLP | recomendado        |
| Health    | `/healthz` + `/readyz`| **SIM**            |

---

## 1. Métricas (Prometheus)

### Endpoint

- Path: `/metrics`
- Porta: mesma do serviço (configurável via `METRICS_PATH`).
- Formato: text/plain (Prometheus exposition).

```go
// internal/app/metrics.go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/gin-gonic/gin"
)

func RegisterMetrics(r *gin.Engine) {
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
```

### Métricas obrigatórias (por microsserviço)

| Métrica                              | Tipo      | Labels                       | Descrição                          |
|--------------------------------------|-----------|------------------------------|------------------------------------|
| `app_info`                           | Gauge     | `version, commit, go_version`| Sempre 1; info da build            |
| `http_requests_total`                | Counter   | `method, path, status`       | Total de requests                  |
| `http_request_duration_seconds`      | Histogram | `method, path`               | Latência                           |
| `db_queries_total`                   | Counter   | `operation, table, status`   | Total de queries no DB             |
| `db_query_duration_seconds`          | Histogram | `operation, table`           | Latência de query                  |

### Métricas recomendadas (por feature)

- `auth_login_total{status}` (success, invalid_password, user_not_found)
- `auth_token_validations_total{result}`
- `cache_hits_total{cache}` / `cache_misses_total{cache}`
- `outbox_pending_total`
- `queue_messages_total{queue, result}`

### Convenção de nomes

- `snake_case`
- `<unidade>_<nome>_<sufixo>` (ex.: `http_request_duration_seconds`)
- Sem prefixo de nome do serviço (use `app_info` para info; escopo por label)

---

## 2. Logs (slog JSON)

### Setup

```go
// internal/app/logger.go
package app

import (
    "log/slog"
    "os"
)

func NewLogger(level, service, version string) *slog.Logger {
    var lvl slog.Level
    _ = lvl.UnmarshalText([]byte(level))
    h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: lvl,
    })
    return slog.New(h).With(
        "service", service,
        "version", version,
    )
}
```

### Campos obrigatórios

Todo log deve ter (no root ou via `.With`):

- `ts` (ISO 8601, automático)
- `level` (info, warn, error, debug)
- `msg` (string curta)
- `service` (nome do microsserviço)
- `version` (versão semver)

Quando aplicável:

- `request_id`
- `trace_id`
- `span_id`
- `user_id`
- `error` (string do erro, com `%v` ou `err.Error()`)

### Uso

```go
slog.Info("user logged in",
    "user_id", user.ID,
    "request_id", reqID,
    "duration_ms", time.Since(start).Milliseconds(),
)

slog.Error("failed to query DB",
    "operation", "GetUser",
    "table", "users",
    "error", err.Error(),
)
```

### Proibições

- ❌ `fmt.Println` em produção.
- ❌ `log.Printf` (stdlib legacy) — use `slog`.
- ❌ Log em arquivo (XII-factor XI).
- ❌ Mensagens multilinhas (`\n` no `msg`).
- ❌ Dados sensíveis (senha, token, cartão) em log.

### Boas práticas

- Mensagem curta, contexto nos campos.
- `Info` para eventos de negócio; `Debug` para debugging; `Warn` para
  situações inesperadas mas recuperáveis; `Error` para falhas.
- Não log em loop apertado (rate limit ou agregue).

---

## 3. Health & Readiness

```go
// internal/app/server.go
r.GET("/healthz", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})

r.GET("/readyz", func(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "ready"})
})
```

- `/healthz` (liveness): o processo está vivo? 200 sempre.
- `/readyz` (readiness): pode receber tráfego? 200 só quando
  dependências estão OK.

---

## 4. Tracing (OpenTelemetry, recomendado)

### Setup

```go
// internal/app/tracing.go
package app

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InitTracer(ctx context.Context, service, version, endpoint string) (func(context.Context) error, error) {
    exporter, err := otlptrace.New(ctx,
        otlptracehttp.NewClient(
            otlptracehttp.WithEndpoint(endpoint),
            otlptracehttp.WithInsecure(),
        ),
    )
    if err != nil {
        return nil, err
    }
    res, _ := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(service),
            semconv.ServiceVersion(version),
        ),
    )
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
    )
    otel.SetTracerProvider(tp)
    return tp.Shutdown, nil
}
```

### Middleware Gin

```go
import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

r.Use(otelgin.Middleware("my-service"))
```

### Correlação com logs

Use `slog-otel` (`github.com/remychantenay/slog-otel`) para injetar
`trace_id` e `span_id` em todo log automaticamente.

```go
import slogotel "github.com/remychantenay/slog-otel"

slog.SetDefault(slog.New(slogotel.OtelHandler{
    Next: slog.NewJSONHandler(os.Stdout, nil),
}))
```

---

## 5. Variáveis de ambiente

| Env                          | Default                      | Descrição                          |
|------------------------------|------------------------------|------------------------------------|
| `LOG_LEVEL`                  | `info`                       | `debug`, `info`, `warn`, `error`   |
| `METRICS_PATH`               | `/metrics`                   | Path do endpoint de metrics        |
| `OTEL_EXPORTER_OTLP_ENDPOINT`| (vazio = tracing desabilitado)| Host:port do collector OTLP        |
| `OTEL_SERVICE_NAME`          | (nome do serviço)            | Service name em spans              |
| `APP_VERSION`                | (do build)                   | Versão em `app_info`               |
| `APP_COMMIT`                 | (do build)                   | Commit SHA em `app_info`           |

---

## 6. Dashboards (Grafana)

Dashboards **mínimos** (criar como JSON no repo, ex.: `deploy/grafana/`):

- **Service Overview**: req/s, p50/p95/p99, error rate, app_info.
- **DB**: queries/s, query p95, errors.
- **Logs**: últimos erros, top 10 mensagens de erro.

Provisionar via Helm/dashboards-as-code (ex.: `grafonnet` ou JSON
com `gz`/`jsonnet`).

---

## 7. Alertas (Alertmanager)

Alertas **mínimos** (em `deploy/alerts/`):

- `HighErrorRate`: 5xx > 5% por 5min.
- `HighLatency`: p95 > 1s por 5min.
- `ServiceDown`: `up{job="my-service"} == 0` por 1min.
- `DBPoolExhausted`: conexões em uso > 90% por 5min.

---

## Anti-padrões

- ❌ Endpoint `/metrics` sem auth em produção (usar network policy ou
  sidecar).
- ❌ Métricas de alta cardinalidade (`user_id` como label).
- ❌ Logs em texto puro.
- ❌ Logs com secrets.
- ❌ `/healthz` que checa dependências (deve ser `/readyz`).
- ❌ Tracing sem propagação de contexto entre serviços.
- ❌ `panic` em código de produção (capturar com `recover` no middleware).
