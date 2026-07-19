# Stack — Versões Pinadas (canônica)

> **O QUÊ:** tabela canônica de **versões pinadas** de cada
> componente da stack. Todo template, Dockerfile, workflow e
> persona DEVE referenciar **exatamente** estas versões.
>
> **POR QUÊ:** aprendemos com o piloto Mandaí v2 (jul/2026) que
> **versões inconsistentes** entre `go.mod`, Dockerfile, e
> workflow causam bugs em cascata:
> - `go.mod` declarava `go 1.26.5`, mas Dockerfile usava
>   `golang:1.22-alpine` → conflito na build.
> - `.golangci.yml` schema v1, linter instalado era v2 → lint quebrava.
> - GitHub Action `oasdiff/oasdiff-action@v1` (tag inválida) → CI quebrava.
> - Node.js v24 vs v26 (LTS vs Current) → instabilidade.
> - **`go mod tidy` com `GOTOOLCHAIN=auto` reescreve o `go` directive
>   do `go.mod`** sem avisar (de 1.23 → 1.25 porque gin, kin-openapi e
>   golang.org/x/text requerem 1.25). Dockerfile fica desatualizado.
> - **Custom migrate builder é frágil**: o `go.mod` do backend não
>   inclui os subpackages de `golang-migrate/migrate/v4/cmd/migrate`,
>   `/database/postgres`, `/source/file` que o `main.go` custom
>   importa. Solução: usar imagem oficial `migrate/migrate`.
>
> **REGRA:** se você precisa de uma versão que não está aqui, **atualize
> este arquivo PRIMEIRO** e gere/atualize o ADR. Não use
> versões "a Latest" em produção.

---

## ⚠️ Gotcha #1 — `go mod tidy` reescreve o `go` directive

Se você rodar `go mod tidy` com Go 1.21+ e o `GOTOOLCHAIN=auto`
(default), o Go **automaticamente** atualiza o `go` directive no
`go.mod` para a versão mínima requerida pelas dependências. Isso
pode quebrar o Dockerfile (que pode estar usando uma versão mais
antiga).

**Exemplo real (Mandaí v2):**

```bash
# Antes
# go.mod:  go 1.23.0
# Dockerfile: golang:1.22-alpine  (já desatualizado!)

# Depois de rodar go mod tidy localmente com Go 1.26.5
# go.mod:  go 1.25.0  (gin + kin-openapi + x/text requerem 1.25)
# Dockerfile: golang:1.22-alpine  ← QUEBRA
```

**Como evitar:**

1. **Pinar `GOTOOLCHAIN=local`** no CI para que o Go **não**
   atualize o `go.mod` automaticamente. Exemplo de CI:

   ```yaml
   - name: Run tests
     working-directory: backend
     env:
       GOTOOLCHAIN: local
     run: go test ./...
   ```

2. **Alinhar sempre `go.mod` e Dockerfile** — se o `go.mod` diz
   `go 1.25.0`, a base image do Dockerfile DEVE ser
   `golang:1.25.x-alpine` ou superior (nunca inferior).

3. **Usar o smoke test + `check-stack-versions.sh --check-latest`**
   detecta inconsistências entre `go.mod`, Dockerfile, CI e
   versões latest estáveis (online).

4. **Atualizar a tabela de versões** quando o `go mod tidy`
   autalizar o `go` directive.

---

## ⚠️ Gotcha #2 — `migrate` custom é frágil

**NÃO** construa uma imagem custom para o `migrate`. Use a imagem
**oficial** `migrate/migrate:v4.19.1` que já embute todos os drivers
(postgres, mysql, sqlite, etc.).

**O que NÃO fazer (causou 1h de debug no Mandaí v2):**

```dockerfile
# ❌ ERRADO — custom migrate builder
FROM golang:1.26.5-alpine AS migrate-builder
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY deploy/migrate-main.go ./cmd/migrate/main.go
RUN go build -o /out/migrate ./cmd/migrate
# Erro: "no required module provides package
# github.com/golang-migrate/migrate/v4/database/postgres"
```

**O que fazer (correto):**

```yaml
# ✅ CERTO — imagem oficial
migrate:
  image: migrate/migrate:v4.19.1
  volumes:
    - ./backend/migrations:/migrations:ro
  command: ["-path", "/migrations", "-database", "${DATABASE_URL}", "up"]
```

Vantagens: sem custom build, sem dependência de go.mod,
atualizações de segurança automáticas, drivers todos disponíveis.

---

## ⚠️ Gotcha #3 — CI Go version drift

O Dockerfile e o `.github/workflows/ci.yml` DEVEM ter o **mesmo**
Go version, e esse Go version DEVE ser ≥ o `go` directive do
`go.mod` **e** ≥ o **bootstrap requirement** daquela versão.

**Exemplo real (Mandaí v2):** o Dockerfile foi atualizado para
`golang:1.25-alpine`, mas o CI ainda usava `GO_VERSION: "1.22"` no
`ci.yml`. Resultado: build local passava, CI falhava.

**Como evitar:** o smoke test + `check-stack-versions.sh --check-latest`
checa essa consistência automaticamente.

---

## ⚠️ Gotcha #4 — Go bootstrap requirement (NOVO, jul/2026)

**Go 1.26+ exige Go ≥ 1.24.6 como bootstrap** para compilar a si
mesmo. Dockerfile com `golang:1.23-alpine` **não compila** código
que tenha `toolchain go1.26.x` ou `go 1.25+` com deps que exigem
1.25+. Ver release notes:
https://go.dev/doc/go1.26

| Go version   | Bootstrap mínimo |
|--------------|------------------|
| 1.26.x       | **1.24.6**       |
| 1.25.x       | 1.22.x           |
| 1.24.x       | 1.21.x           |
| 1.23.x       | 1.20.x           |

**Regra:** se seu `go.mod` exige `go 1.25+`, **TODAS** as
imagens envolvidas (Dockerfile, CI, dev local) DEVEM ser
**≥ 1.24.6** (1.25 ideal). Senão o build falha com
"go.mod requires go >= X (running go Y; GOTOOLCHAIN=local)".

---

## ⚠️ Gotcha #5 — `golangci-lint` v2 schema migration (NOVO)

**`golangci-lint` v2** introduziu um schema **totalmente diferente**.
Um binary v2 lendo `.golangci.yml` em schema v1 **falha com
"unsupported version"** sem lintar nada.

**O que mudou:**
- `issues:` continua existindo no top level, mas a maioria dos
  campos foi renomeada.
- `exclusions:` (v2) substitui `exclude-rules:` (v1) — vai
  **dentro de `linters:`**, não no top level.
- `settings:` (v1, top level) → `linters.settings:` (v2).
- **gofmt e goimports** saíram de `linters:` e viraram
  `formatters:` (seção separada no v2).

**Como migrar:**

```bash
# 1. Backup
cp .golangci.yml .golangci.v1.bck.yml

# 2. Migrar automaticamente
golangci-lint migrate

# 3. Revisar (migration não copia comentários)
# 4. Validar
golangci-lint run ./...
```

**Pinar até migrar:** se ainda usa v1 syntax, fixe a action em
`version: v1.64.x` (última v1) até fazer a migração.

**Lição Mandaí v2 (jul/2026):** o template tinha `version: "2"`
no top mas `settings:` no top level (v1) E `exclusions:` em
`linters:` (v2) misturados. Resultado: lint quebrava com
"additional properties 'issues' not allowed".

**Como evitar:** usar o template `harness/templates/.golangci.yml`
que já está em v2 puro. Não copiar de fontes v1.

---

## ⚠️ Gotcha #6 — distroless `debianX` suffix é OBRIGATÓRIO (NOVO, jun/2026)

A partir de jun/2026, as tags **sem sufixo `-debianX`** do
distroless (`gcr.io/distroless/static`, `gcr.io/distroless/base`,
`gcr.io/distroless/cc`) foram marcadas como **deprecated** e
apontam para `debian13`. Quem usar tag sem sufixo está em canal
deprecated e pode quebrar silenciosamente.

**O que fazer:**

```dockerfile
# ❌ ERRADO — tag sem sufixo (deprecated, aponta para debian13 anyway)
FROM gcr.io/distroless/static:nonroot
FROM gcr.io/distroless/base:nonroot

# ✅ CERTO — sempre com sufixo debianX explícito
FROM gcr.io/distroless/static-debian13:nonroot  # Go (statically linked)
FROM gcr.io/distroless/base-debian13:nonroot   # Node (precisa libc)
```

**Regra:**
- **Go** (CGO_ENABLED=0, binário estático) → `static-debian13`.
- **Node** (precisa de `libc`) → `base-debian13` (NUNCA `static`).
- **UID 65532:65532** numérico (não `nonroot:nonroot` string —
  k8s `runAsNonRoot: true` não valida string).

---

## ⚠️ Gotcha #7 — Trivy supply-chain attack (mar/2026)

Em **19/mar/2026**, a Aqua sofreu um supply-chain attack que
comprometeu **trivy v0.69.4** e a maioria das tags de
`trivy-action` (76 de 77 tags, exceto `0.35.0` que estava pinada
por SHA). Versões seguras:

- **trivy CLI:** `v0.69.2`, `v0.69.3`, e **tudo a partir de
  v0.69.5 (pós-incidente)**.
- **trivy-action:** `v0.35.0` SHA-pinned, ou qualquer versão
  a partir de `v0.32.0` (jul/2026, pós-incidente).
- **setup-trivy:** `v0.2.6` SHA-pinned.

**Lição Mandaí v2:** pinamos `trivy-action@0.32.0` (jul/2026) —
versão segura e atual. Mas a recomendação forte é **SHA-pinned**:

```yaml
# Recomendado para produção
- uses: aquasecurity/trivy-action@57a97c7b4f9c4f95c8c8e0d2e6c4f5a8b9c0d1e2
- uses: aquasecurity/trivy-action@0.32.0  # OK se atualizado mensalmente
```

**Como evitar:** verificar periodicamente
https://github.com/aquasecurity/trivy/releases e
https://github.com/aquasecurity/trivy-action/releases. Usar
`check-stack-versions.sh --check-latest` que alerta se a versão
pinada está 6+ meses desatualizada.

---

## ⚠️ Gotcha #8 — Nuxt 3 EOL 31/jul/2026 (NOVO)

A Nuxt 3 chegou ao **end-of-life em 31/jul/2026** (essa semana,
no momento em que este doc é escrito). Apenas **Nuxt 4** recebe
updates daqui pra frente.

**Regra:** todo projeto novo (e migração) usa **Nuxt 4.3+**. Não
usar Nuxt 3.

---

## ⚠️ Gotcha #9 — Node.js 26 ainda NÃO é LTS (NOVO)

Node.js 26 (released 2026-05-05) é "Current", **NÃO LTS** até
out/2026. Só vira LTS em outubro/2026. **Não usar em produção**
até lá.

**Regra:** Node 24 LTS (Krypton, EOL abril/2028) é a escolha
default. Node 22 LTS (Jod, EOL abril/2027) é fallback aceitável.

---

## Política de pinning (resumo)

1. **MAJOR version é fixo** (ex.: Go 1.26, não `latest`).
2. **MINOR/PATCH é fixo** quando houver risco de regressão
   (ex.: golangci-lint, Nuxt, Trivy).
3. **Imagens Docker**: usar **digest SHA256** em produção, tag
   semver no dev/CI.
4. **GitHub Actions**: usar **SHA-pinned** em produção, tag semver
   no dev/CI.
5. **Atualizar esta tabela** quando uma major version nova for
   lançada e estabilizar (≥ 3 meses no mercado).
6. **Quebrar o pinning só via ADR** registrado em
   `contrib/design-decisions.md`.

---

## Como pesquisar a latest estável (workflow)

Sempre que uma nova versão pinada for necessária, **pesquise
online** e cite a fonte na tabela abaixo.

**1. Linguagens & runtimes:**

| Componente | Fonte canônica                       | Comando CLI                          |
|------------|--------------------------------------|--------------------------------------|
| Go         | https://go.dev/dl/                    | `go version` (local)                 |
| Node.js    | https://nodejs.org/en/about/previous-releases | `node --version`               |
| PostgreSQL | https://hub.docker.com/_/postgres/tags | `docker run postgres --version`     |

**2. Libs Go:**

```bash
# Latest version de uma lib
go list -m -versions github.com/gin-gonic/gin | tr ' ' '\n' | tail -1

# Checar deprecation
go list -m -versions github.com/deepmap/oapi-codegen  # path antigo
go list -m -versions github.com/oapi-codegen/oapi-codegen/v2  # path novo
```

**3. Libs JS/TS:**

```bash
# Latest version
npm view nuxt version
npm view @nuxt/ui version
npm view @nuxtjs/i18n version
```

**4. GitHub Actions (versão + SHA):**

```bash
# Latest release tag
gh release list --repo aquasecurity/trivy-action --limit 1

# SHA de uma tag específica (para pin)
gh api repos/aquasecurity/trivy-action/git/refs/tags/0.32.0 --jq '.object.sha'
```

**5. Docker images (latest tag + digest):**

```bash
# Latest tag
docker manifest inspect golang:latest | jq '.tags[0]'

# Digest SHA256
docker manifest inspect golang:1.26.5-alpine3.22 | jq '.manifest.digest'
```

**6. Validar automaticamente:** `check-stack-versions.sh --check-latest`
faz as chamadas acima e alerta drift.

---

## Linguagens & runtimes

| Componente    | Versão pinada  | Última estável (jul/2026) | Fonte                              | Notas                              |
|---------------|---------------|--------------------------|------------------------------------|------------------------------------|
| **Go**        | **1.26.5**     | 1.26.5 (2026-07-07)      | https://go.dev/dl/                 | Bootstrap: **1.24.6+**. 1.27 ainda em beta |
| **Node.js**   | **24 LTS**     | 24.18.0 (LTS Krypton)    | https://nodejs.org/en/about/previous-releases | 26 é "Current" (não-LTS até Out/2026). Use 24 LTS em produção. EOL: abr/2028 |
| **TypeScript** | **5.x**        | 5.9.x                    | https://www.typescriptlang.org/    | Required by Pinia 3, Nuxt 4        |

> ⚠️ **Node.js 26** (released 2026-05-05) é "Current" mas **NÃO** é
> LTS ainda. Volta a ser LTS em Out/2026. **Use Node 24 LTS em
> produção até lá.**

> ⚠️ **Node 20 LTS** entrou em EOL em 30/abr/2026. Não usar mais.

---

## Backend (Go) — libs

| Lib                                           | Versão pinada  | Última estável (jul/2026) | Import path                                                | Notas |
|----------------------------------------------|---------------|--------------------------|------------------------------------------------------------|-------|
| `gin`                                          | **v1.10+**     | v1.10.1 (nov/2025)       | `github.com/gin-gonic/gin`                                  | Requer `go 1.25+` |
| `gorm`                                         | **v1.31.2**    | v1.31.2 (jul/2026)       | `gorm.io/gorm`                                             |       |
| `gorm.io/driver/postgres`                      | **v1.6+**      | v1.6.0 (2025)            | `gorm.io/driver/postgres`                                  |       |
| `golang-migrate`                               | **v4.19.1**    | v4.19.1 (29/nov/2025)    | `github.com/golang-migrate/migrate/v4`                     | **Usar imagem oficial** `migrate/migrate:v4.19.1` |
| `go-playground/validator`                      | **v10+**       | v10.22+ (2025)           | `github.com/go-playground/validator/v10`                   |       |
| `envconfig`                                    | **v1.4+**      | v1.4.0                   | `github.com/kelseyhightower/envconfig`                     |       |
| `slog`                                         | **stdlib**      | stdlib Go                | `log/slog` (Go 1.21+)                                     |       |
| `prometheus/client_golang`                     | **v1.19+**     | v1.20+ (2025)            | `github.com/prometheus/client_golang`                       |       |
| `nicksnyder/go-i18n/v2`                        | **v2.5+**      | v2.6+ (2025)             | `github.com/nicksnyder/go-i18n/v2`                          |       |
| `oapi-codegen/v2`                              | **v2.8.0**     | v2.8.0 (jul/2026)        | `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen` | Path NOVO (não `deepmap/oapi-codegen`) |
| `testify`                                      | **v1.11.1**    | v1.11.1 (2025)           | `github.com/stretchr/testify`                              |       |
| `go.uber.org/mock`                             | **v0.5+**      | v0.5+ (2025)             | `go.uber.org/mock`                                         |       |
| `google/uuid`                                  | **v1.6+**      | v1.6.0                   | `github.com/google/uuid`                                   |       |
| `golang.org/x/text/language`                   | **v0.x latest**| v0.x (atualizar)         | `golang.org/x/text/language`                               | Requer `go 1.25+` |
| `otelgin` (OpenTelemetry)                      | **v0.x latest**| v0.x (atualizar)         | `go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin` |       |
| `slog-otel`                                    | **v0.x latest**| v0.x (atualizar)         | `github.com/remychantenay/slog-otel`                       |       |

> ⚠️ **`oapi-codegen` mudou de path em v2.3.0.** O path antigo
> `github.com/deepmap/oapi-codegen/v2` foi descontinuado. Use
> **sempre** `github.com/oapi-codegen/oapi-codegen/v2`.

> ⚠️ **gin + kin-openapi + x/text** exigem `go 1.25+`. Se aparecer
> "go.mod requires go >= 1.25 (running go 1.X)" em build, é o
> Dockerfile/ci com versão antiga — alinhe para 1.25+.

---

## Frontend (Node/JS) — libs

| Lib                              | Versão pinada  | Última estável (jul/2026) | Package                              | Notas |
|----------------------------------|---------------|--------------------------|--------------------------------------|-------|
| `nuxt`                            | **v4.5.0**    | v4.5.0 (jul/2026)        | `nuxt`                               | **Nuxt 3 EOL 31/jul/2026** — só usar 4.x |
| `@nuxt/ui`                        | **v3.3.6**    | v3.3.6 (2026)            | `@nuxt/ui`                           | v3 estável; v4.10.0 (unificou com Pro) também OK |
| `pinia`                           | **v3.0.3**    | v3.0.3 (2025)            | `pinia`                              | Requires Vue 3, TS 5+ |
| `@pinia/nuxt`                     | **v0.5+**     | v0.5+ (2025)             | `@pinia/nuxt`                        |       |
| `@nuxtjs/i18n`                    | **v10.4.1**   | v10.4.1 (2025)           | `@nuxtjs/i18n`                       | v10 = Nuxt 4 support |
| `vueuse/core`, `vueuse/nuxt`      | **latest**    | (atualizar)              | `@vueuse/core`, `@vueuse/nuxt`       |       |
| `vitest`                          | **latest**    | (atualizar)              | `vitest`                             |       |
| `@vue/test-utils`                 | **latest**    | (atualizar)              | `@vue/test-utils`                    |       |
| `@nuxt/test-utils`                | **latest**    | (atualizar)              | `@nuxt/test-utils`                   |       |
| `@playwright/test`                | **latest**    | (atualizar)              | `@playwright/test`                   |       |
| `zod`                             | **v3+**       | v3.x (2025)              | `zod`                                |       |
| `typescript`                      | **v5.x**      | v5.9.x                   | `typescript`                         |       |

> ⚠️ **Nuxt 3 chegou ao EOL em 31/jul/2026.** Não usar mais.
> Migrar para Nuxt 4.3+ se ainda estiver em 3.x.

> ⚠️ **Nuxt UI v4 (v4.10.0)** unificou com Pro e está open-source
> 100%. Pode ser usado, mas v3.3.6 é a linha mais estável e tem
> mais material na comunidade. Decida com o `solutions-architect`.

---

## Imagens Docker (base)

| Componente    | Tag pinada                   | Última estável (jul/2026) | Fonte                              | Notas                                |
|---------------|------------------------------|--------------------------|------------------------------------|--------------------------------------|
| **Go build**  | `golang:1.26.5-alpine3.22`   | 1.26.5-alpine3.22        | https://hub.docker.com/_/golang/tags | Pin TUA exato de Go + Alpine        |
| **Node build**| `node:24-alpine3.22`         | 24.18.0-alpine3.22       | https://hub.docker.com/_/node/tags  | LTS + Alpine 3.22 (latest stable)   |
| **Go runtime**| `gcr.io/distroless/static-debian13:nonroot` | nonroot (UID 65532) | https://github.com/GoogleContainerTools/distroless | ~2 MB, non-root (UID 65532) |
| **Node runtime**| `gcr.io/distroless/base-debian13:nonroot` | nonroot (UID 65532) | https://github.com/GoogleContainerTools/distroless | Precisa de libc; use `base` não `static` |
| **PostgreSQL**| `postgres:18.4-alpine`       | 18.4-alpine              | https://hub.docker.com/_/postgres/tags | 18.4 é current; 17 também suportado  |
| **migrate**   | `migrate/migrate:v4.19.1`    | v4.19.1 (29/nov/2025)    | https://github.com/golang-migrate/migrate | **Imagem OFICIAL** (não construir custom) |
| **Trivy**     | `aquasec/trivy:0.72.0`       | v0.72.0 (30/jun/2026)    | https://github.com/aquasecurity/trivy/releases | CLI standalone. v0.69.4 comprometido (mar/2026) — usar 0.72.0+ |

> ⚠️ **distroless `debian13` é o default atual** (jun/2026).
> Tags sem sufixo (`gcr.io/distroless/static:nonroot`) estão
> **deprecated** e podem quebrar. **Sempre** usar com sufixo
> `-debian13` explícito.

> ⚠️ **Para Node use `base` (precisa de libc)**, **não** `static`
> (não tem libc).

> ⚠️ **Sempre** use `USER 65532:65532` (não `USER nonroot:nonroot`)
> em distroless — Kubernetes `runAsNonRoot: true` não valida
> string username.

> ⚠️ **Use a imagem OFICIAL `migrate/migrate:v4.19.1`** em vez de
> construir uma custom. A imagem oficial já embute os drivers
> postgres, mysql, sqlite, etc. Construir um migrate custom (ex.:
> `deploy/migrate-main.go`) **sempre dá problema** porque o `go.mod`
> do backend não inclui os subpackages `cmd/migrate`, `database/postgres`,
> `source/file` que o `main.go` custom importa. **Lição do Mandaí v2
> (jul/2026):** perdemos 1h debugando isso até descobrir que o
> jeito certo é usar a imagem oficial.

---

## Security & quality tools

| Ferramenta        | Versão pinada  | Última estável (jul/2026) | Como rodar                           |
|-------------------|---------------|--------------------------|--------------------------------------|
| **golangci-lint**  | **v2.12.2**   | v2.12.2 (6/mai/2026)     | `golangci/golangci-lint-action@v9.3.0` | **Schema v2 puro** — `golangci-lint migrate` para converter de v1. **v9.3.0 é o MÍNIMO que aceita `version: v2.x`** (v6 não suporta v2 → "invalid version string"). |
| **govulncheck**    | **latest**     | (atualizar)              | `go install golang.org/x/vuln/cmd/govulncheck@latest` | |
| **trivy CLI**      | **v0.72.0**   | v0.72.0 (30/jun/2026)    | `aquasec/trivy:0.72.0`                | **NÃO** usar v0.69.4 (comprometido) |
| **trivy-action**   | **v0.36.0**   | v0.36.0 (jul/2026)       | `aquasecurity/trivy-action@v0.36.0`   | **SEMPRE usar prefixo `v`** — `0.36.0` (sem `v`) retorna 404. Pós-incidente supply-chain (mar/2026). Para produção, SHA-pinned |
| **oasdiff**        | **v1.7.0**    | (atualizar)              | `oasdiff/oasdiff-action@v1.7.0`      | |
| **oasdiff-action** | **v1.7.0**    | (atualizar)              | (pinada, sem `latest`)                | Tag `@v1` é inválida (não existe) — sempre `@v1.7.0` |
| **Gitleaks**       | **v8.18+**    | (atualizar)              | `gitleaks/gitleaks-action@v2`        | |
| **Spectral**       | **v6+**       | (atualizar)              | `stoplightio/spectral-action@v0`     | |

> ⚠️ **Trivy teve incidente de supply-chain em mar/2026** (v0.69.4
> comprometido). Use **sempre** versões pinadas e **NÃO** v0.69.4.
> Trivy seguro: **v0.69.2–v0.69.3** ou **≥ v0.69.5** (pós-incidente).
> Recomendamos **v0.72.0** (atual, jul/2026) ou mais novo.

> ⚠️ **`golangci-lint` v2 schema é totalmente diferente de v1.**
> Use SEMPRE o template `harness/templates/.golangci.yml` que está
> em v2 puro. Para migrar de v1: `golangci-lint migrate`.

---

## GitHub Actions (pinned versions)

| Action                                | Versão pinada  | Última estável (jul/2026) | Uso                                    |
|---------------------------------------|---------------|--------------------------|----------------------------------------|
| `actions/checkout`                     | **@v4**       | v4.2.2                   | Checkout do repo                        |
| `actions/setup-go`                     | **@v5**       | v5.0.0                   | Setup Go 1.26.5                         |
| `actions/setup-node`                   | **@v4**       | v4.1.0                   | Setup Node 24 LTS                       |
| `actions/setup-python`                 | **@v5**       | v5.3.0                   | Setup Python (Trivy, Spectral, etc.)    |
| `actions/upload-artifact`              | **@v4**       | v4.4.0                   | Upload de artifacts                     |
| `actions/download-artifact`            | **@v4**       | v4.1.2                   | Download de artifacts                    |
| `golangci/golangci-lint-action`       | **@v9.3.0**   | v9.3.0 (jun/2026)        | Lint Go (v2.12.2) — v9.3.0 é o MÍNIMO que aceita `version: v2.x` |
| `docker/setup-buildx-action`           | **@v3**       | v3.7.1                   | Setup Buildx                            |
| `docker/build-push-action`             | **@v6**       | v6.7.0                   | Build + push Docker image               |
| `docker/login-action`                  | **@v3**       | v3.2.0                   | Login em registry                       |
| `aquasecurity/trivy-action`            | **@v0.35.0** ou SHA-pinned | v0.35.0 (mar/2026, pré-ataque) | Trivy image scan — **v0.36.0-v0.69.x é ZONA CINZENTA** (comprometimento mar/2026 poisonou 76/77 tags). Usar `@v0.35.0` (última validada) ou SHA-pinned. **SEMPRE com prefixo `v`** (`0.36.0` retorna 404). |
| `aquasecurity/setup-trivy`             | **@0.2.6**    | 0.2.6 (pós-incidente)    | Setup Trivy CLI                          |
| `oasdiff/oasdiff-action`               | **@v1.7.0**   | (atualizar)              | OpenAPI contract diff                   |
| `github/codeql-action/upload-sarif`    | **@v3**       | v3.27.0                  | Upload SARIF                             |
| `golang/govulncheck-action`            | **@v1**       | v1.0.0                   | govulncheck                              |
| `gitleaks/gitleaks-action`              | **@v2**       | v2.3.0                   | Gitleaks (secrets scan)                  |
| `stoplightio/spectral-action`          | **@v0**       | v0.0.2                   | Spectral (OpenAPI lint)                  |

> ⚠️ **NUNCA** use `@latest` ou `@main` em GitHub Actions em
> produção. **Sempre** pinar a versão (semver ou SHA).

> ⚠️ **Para produção crítica, prefira SHA-pinned**:
> ```yaml
> - uses: aquasecurity/trivy-action@57a97c7b4f9c4f95c8c8e0d2e6c4f5a8b9c0d1e2
> ```
> Tag semver pode ser movida por supply-chain attack (vimos isso
> com Trivy em mar/2026).

---

## Como adicionar uma nova versão

1. **Pesquisar latest** (ver seção "Como pesquisar a latest
   estável" acima). Citar a **fonte/URL** na tabela.
2. **Validar com `check-stack-versions.sh --check-latest`** no
   projeto piloto (Mandaí v2 ou próximo).
3. **Atualizar este arquivo** (`harness/stack/versions.md`).
4. **Criar/estender ADR** em `contrib/design-decisions.md`
   justificando a mudança (especialmente se for major version).
5. **Atualizar todos os templates** que referenciam a versão:
   - `templates/Dockerfile.template` (se mudou Go/Node).
   - `templates/.github-workflows-ci.yml` (se mudou action).
   - `templates/.golangci.yml` (se mudou lint).
   - `templates/docker-compose.template.yml` (se mudou imagem).
6. **Sincronizar** o meta-harness no `mandai-v2` (ou próximo
   projeto) com a nova versão.
7. **Validar com smoke test + rodar CI completo** (todos os jobs).

---

## Como bumpar automaticamente

> Use **Renovate** (recomendado) ou **Dependabot** (built-in do
> GitHub) para monitorar novas versões e abrir PRs automáticos.
> Configurar em `.github/renovate.json` (não incluso por default;
> ver exemplo em `templates/` quando necessário).
>
> **Recomendação Renovate:**
> - Agrupar por domínio (Go, Node, Docker, Actions).
> - Pinning semanal para libs, mensal para Actions/imagens.
> - Auto-merge apenas para patch updates.
> - Major updates exigem revisão manual.
