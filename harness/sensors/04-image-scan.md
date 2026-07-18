# Sensor 04 — Image Scan (Trivy)

> **Objetivo:** garantir que a imagem Docker produzida não tem
> vulnerabilidades CRITICAL no OS ou nas libs empacotadas.
> **Quando roda:** CI (após build da imagem).
> **Falha → ação:** **bloqueia deploy** se CRITICAL sem waiver.

---

## Comandos exatos

```bash
# Instalar
brew install trivy
# ou
go install github.com/aquasecurity/trivy/cmd/trivy@latest

# Escanear imagem local
trivy image --severity CRITICAL,HIGH --exit-code 1 my-service:v0.4.0

# Escanear filesystem (em CI, antes de buildar)
trivy fs --severity CRITICAL --exit-code 1 .

# Escanear IaC (Dockerfile, compose, k8s)
trivy config --severity CRITICAL --exit-code 1 deploy/
```

### Com SARIF (para GitHub Code Scanning)

```bash
trivy image --format sarif --output trivy-image.sarif my-service:v0.4.0
```

---

## Thresholds

| Severidade   | Bloqueia merge? | Bloqueia release? | Waiver?                      |
|--------------|-----------------|-------------------|------------------------------|
| **CRITICAL** | ✅ sim          | ✅ sim            | só com ADR + prazo ≤ 7 dias  |
| **HIGH**     | ⚠️ reportar     | ✅ sim            | com ADR + prazo ≤ 30 dias    |
| **MEDIUM**   | ❌ não          | ❌ não            | reportar                     |
| **LOW**      | ❌ não          | ❌ não            | reportar                     |

---

## Onde pluga no pipeline

### CI (`.github/workflows/ci.yml`)

```yaml
build-and-scan:
  name: Build + Image scan
  runs-on: ubuntu-latest
  needs: [lint, test, contract]
  steps:
    - uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Build (and cache)
      uses: docker/build-push-action@v5
      with:
        context: .
        push: false
        tags: my-service:${{ github.sha }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

    - name: Trivy scan
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: my-service:${{ github.sha }}
        format: sarif
        output: trivy-image.sarif
        severity: CRITICAL,HIGH
        exit-code: '1'
        ignore-unfixed: true

    - name: Upload SARIF
      if: always()
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: trivy-image.sarif
```

---

## Boas práticas de imagem (para evitar findings)

- Usar **multi-stage** build (ver `harness/stack/docker.md`).
- Base **`gcr.io/distroless/static-debian12:nonroot`** ou
  **`alpine:3.19`** (atualizadas).
- **Atualizar a base regularmente** (Dependabot atualiza Go, mas a base
  Alpine/Distroless precisa de manual ou Renovate).
- **Remover pacotes não usados** no build (`apk del` no final).
- **Não incluir secrets** na imagem.
- **Pin de versão** da base (`FROM golang:1.22.5-alpine3.20`, não
  `golang:latest`).
- **Incluir `HEALTHCHECK`** no Dockerfile.
- **Rodar como non-root** (`USER nonroot:nonroot` no Distroless, ou
  criar user no Alpine).

---

## Falha típica & remediação

| Falha                                                | Como corrigir                                        |
|------------------------------------------------------|------------------------------------------------------|
| CRITICAL em `openssl`                               | Atualizar base Alpine ou `apt-get update && upgrade`. |
| CRITICAL em lib Go                                  | `go get -u lib` ou fix upstream.                     |
| HIGH em shell busybox                                | Trocar base para distroless.                         |
| CRITICAL em `musl`                                  | Trocar base para versão patchada.                    |
| Múltiplos CRITICAL em OS                             | Rebuildar com base mais recente (Dependabot).        |

---

## Quem roda

- **CI:** workflow `ci.yml` (job `build-and-scan`).
- **Falha CRITICAL:** bloqueia release (ver
  `harness/workflow/04-release.md`).
- **Falha HIGH:** reportar e abrir issue `tech-debt` se for feature,
  bloquear release.
