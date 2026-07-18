# Exemplo — Hello Service (Go + Gin + GORM + PostgreSQL)

> Exemplo **end-to-end** mínimo de um microsserviço seguindo o
> meta-harness. Use como referência para o seu projeto.

---

## O que este exemplo entrega

- ✅ Estrutura Standard Go Project Layout.
- ✅ Go + Gin + GORM + PostgreSQL + OpenAPI (spec-first).
- ✅ Migrações com `golang-migrate`.
- ✅ Observability (slog JSON + Prometheus + health).
- ✅ Testes (table-driven, testify, fakes in-memory).
- ✅ 12-Factor audit script.
- ✅ Dockerfile multi-stage (distroless, non-root, healthcheck).
- ✅ docker-compose para snapshot local.
- ✅ GitHub Actions (lint, test, vuln, contract, 12-factor,
  build-and-scan).

> **Escopo:** 1 endpoint `POST /api/v1/greetings` que cria uma
> saudação em `greetings (id, name, message, created_at)` e retorna
> o registro criado. Endpoint de health + metrics + ready.

---

## Estrutura

```
hello-service/
├── cmd/
│   ├── server/main.go
│   └── migrate/main.go
├── internal/
│   ├── app/
│   │   ├── config.go
│   │   ├── logger.go
│   │   ├── metrics.go
│   │   ├── db.go
│   │   └── server.go
│   ├── api/
│   │   └── openapi.gen.go       # gerado
│   ├── domain/
│   │   └── greeting.go
│   ├── repository/
│   │   └── greeting.go          # interface
│   ├── service/
│   │   ├── greeting.go
│   │   └── greeting_test.go
│   ├── handler/
│   │   └── greeting.go
│   └── platform/
│       └── postgres/
│           └── greeting.go      # impl GORM
├── api/
│   └── openapi.yaml
├── migrations/
│   ├── 000001_create_greetings_table.up.sql
│   └── 000001_create_greetings_table.down.sql
├── deploy/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── .env.example
├── test/
│   └── smoke/
│       └── backend.sh
├── scripts/
│   └── check-twelve-factor.sh
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── .golangci.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

> **Este exemplo é um esqueleto**, não o código completo. Para o
> código completo, abra o `hello-service/` (a ser preenchido em
> iteração futura do meta-harness). Por enquanto, use os snippets
> em [`harness/stack/backend.md`](../../stack/backend.md) e os
> templates em [`harness/templates/`](../../templates/) para montar
> o seu.

---

## Como rodar este exemplo (quando existir)

```bash
cd examples/hello-service

# 1. Subir ambiente
docker compose -f deploy/docker-compose.yml up -d --build

# 2. Testar
curl -fsS http://localhost:8080/healthz
curl -fsS http://localhost:8080/readyz
curl -fsS http://localhost:8080/metrics | head -5

curl -X POST http://localhost:8080/api/v1/greetings \
  -H "Content-Type: application/json" \
  -d '{"name":"World","message":"Hello, World!"}'

# 3. Derrubar
docker compose -f deploy/docker-compose.yml down
```

---

## Issues canônicas de exemplo

Para exercitar o meta-harness do zero, use estas issues:

1. **(#1) Bootstrap hello-service** — feat — meta-issue de bootstrap.
2. **(#2) Endpoint POST /api/v1/greetings** — feat — primeira feature.
3. **(#3) Health/ready/metrics endpoints** — feat — observability.
4. **(#4) Migrations (golang-migrate)** — feat — schema.
5. **(#5) Dockerfile multi-stage** — feat — container.
6. **(#6) docker-compose local** — feat — snapshot.
7. **(#7) CI (lint + test + vuln + 12-factor)** — feat — pipeline.
8. **(#8) Release workflow (tag + GHCR)** — feat — release.

---

> **Status:** em construção. Contribua abrindo PR no meta-harness.
