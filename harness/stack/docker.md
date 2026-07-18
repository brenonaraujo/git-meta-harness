# Stack — Docker (multi-stage, distroless, non-root)

> Padrão de Dockerfiles para microsserviços do meta-harness.
> Imagens **leves, seguras, reprodutíveis**.

---

## Princípios

1. **Multi-stage**: build e runtime em stages separados.
2. **Base mínima**: `gcr.io/distroless/static` (Go) ou `alpine:3.19+`
   (Node) — `scratch` apenas se você sabe o que está fazendo.
3. **Non-root**: executar como usuário sem privilégios.
4. **Pin de versão**: nunca `latest`. Sempre `<image>:<version>`.
5. **Cache mounts**: usar BuildKit para acelerar.
6. **HEALTHCHECK**: definir no Dockerfile.
7. **Ca certificates**: incluir para chamadas HTTPS.
8. **Variáveis em build**: `VERSION`, `COMMIT` via `--build-arg`.

---

## Backend (Go) — Dockerfile

```dockerfile
# syntax=docker/dockerfile:1

# ============ Stage 1: deps ============
FROM golang:1.22.5-alpine3.20 AS deps
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# ============ Stage 2: build ============
FROM golang:1.22.5-alpine3.20 AS builder
WORKDIR /src
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY . .

ARG VERSION=dev
ARG COMMIT=unknown

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/server ./cmd/server

# ============ Stage 3: runtime ============
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/server /server

USER nonroot:nonroot
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/server", "-healthcheck"]

ENTRYPOINT ["/server"]
```

**Tamanhos típicos:**

- `scratch`: ~12 MB
- `distroless/static`: ~14 MB ✅ **recomendado**
- `alpine:3.19`: ~22 MB

---

## Frontend (Node) — Dockerfile

```dockerfile
# syntax=docker/dockerfile:1

# ============ Stage 1: deps ============
FROM node:20-alpine3.20 AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

# ============ Stage 2: build ============
FROM node:20-alpine3.20 AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ARG VERSION=dev
ENV NUXT_PUBLIC_APP_VERSION=${VERSION}
RUN pnpm build

# ============ Stage 3: runtime ============
FROM node:20-alpine3.20
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000
ENV HOST=0.0.0.0

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nuxt -u 1001

COPY --from=builder --chown=nuxt:nodejs /app/.output ./.output
COPY --from=builder --chown=nuxt:nodejs /app/public ./public

USER nuxt
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O- http://localhost:3000/ || exit 1

ENTRYPOINT ["node", ".output/server/index.mjs"]
```

---

## Variáveis de build

```bash
docker build \
  --build-arg VERSION=$(git describe --tags --abbrev=0) \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  -t my-service:0.4.0 \
  -f deploy/Dockerfile .
```

Injete no `go build` via `-ldflags`:

```go
package main

var (
    version = "dev"
    commit  = "unknown"
)
```

E no build:

```bash
go build -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT}"
```

---

## Cache de layers (BuildKit)

Use `--mount=type=cache` para `/go/pkg/mod` e `/root/.cache/go-build`:

```dockerfile
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build ...
```

E habilite BuildKit no CI:

```yaml
# .github/workflows/ci.yml
env:
  DOCKER_BUILDKIT: 1
```

---

## Multi-arch

Para suportar `linux/amd64` e `linux/arm64`:

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION=v0.4.0 \
  -t ghcr.io/org/my-service:v0.4.0 \
  --push .
```

---

## Healthcheck em Go

Adicione uma flag para healthcheck sem subir o servidor completo:

```go
// cmd/server/main.go
func main() {
    if slices.Contains(os.Args[1:], "-healthcheck") {
        // faz ping e sai
        res, err := http.Get("http://localhost:8080/healthz")
        if err != nil || res.StatusCode != 200 {
            os.Exit(1)
        }
        os.Exit(0)
    }
    // ... start server
}
```

Ou use `wget` (em distroless, você precisa de `wget` na imagem, ou
usar um healthcheck binário separado):

```dockerfile
# Adicionar wget/curl na imagem é OK; alternativa: usar healthcheck
# via netcat (ncat) em distroless:
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/server", "-healthcheck"]
```

---

## .dockerignore

```
.git
.github
.harness
.idea
.vscode
node_modules
dist
build
coverage*
*.log
.DS_Store
README.md
LICENSE
```

---

## Docker Compose (snapshot local)

Ver template em
[`../templates/docker-compose.template.yml`](../templates/docker-compose.template.yml).

Pontos:

- **Pin de versões** das imagens (`postgres:16-alpine`, não
  `postgres:latest`).
- **Healthcheck** em todos os serviços.
- **`depends_on: condition: service_healthy`** para ordem.
- **Variáveis em `.env`** (não em compose direto).
- **Volume nomeado** para Postgres.

---

## Scan de vulnerabilidade

Sempre rodar Trivy (ver
[`../sensors/04-image-scan.md`](../sensors/04-image-scan.md)):

```bash
trivy image --severity CRITICAL,HIGH my-service:0.4.0
```

---

## Anti-padrões

- ❌ `FROM golang:latest` ou `node:latest`.
- ❌ Build em stage único (imagem final tem source + Go + toolchain).
- ❌ Rodar como root.
- ❌ `apt-get install` sem `&& rm -rf /var/lib/apt/lists/*`.
- ❌ Imagem base sem atualização (deprecada).
- ❌ Sem HEALTHCHECK.
- ❌ `COPY . .` antes de `go mod download` (quebra cache).
- ❌ Secrets no Dockerfile.
- ❌ `EXPOSE 22` (SSH).
- ❌ Múltiplos processos em um container (1 processo = 1 container).
