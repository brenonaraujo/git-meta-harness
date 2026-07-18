# Deploy — usando imagens publicadas do meta-harness

> **O QUÊ:** como usar as imagens Docker publicadas pelo
> **release pipeline** (v1.6.0+) do meta-harness. As imagens vão
> para o **GitHub Container Registry (GHCR)** e podem ser usadas
> em **ECS, EKS, Docker Swarm, ou localmente via docker-compose**.
>
> **POR QUÊ:** o release pipeline é o último passo do
> meta-harness — fecha o ciclo de "código → release → artefato
> utilizável em produção". Sem ele, o time-manager marca "done"
> mas não há artefato consumível fora do repo.
>
> **QUEM:** `devops-engineer` (cria a tag) + CI (builda + push)
> + humano (deploya em produção via ECS/EKS/Swarm).

---

## 1. O que é publicado

Quando você cria uma tag `vX.Y.Z` na `main`, o
[release workflow](../templates/.github-workflows-release.yml)
roda e publica:

| Imagem                                                  | Conteúdo                  | Tamanho típico |
|---------------------------------------------------------|---------------------------|----------------|
| `ghcr.io/<owner>/<repo>/backend:X.Y.Z`                  | Binário Go + distroless   | ~20 MB         |
| `ghcr.io/<repo-owner>/<repo>/backend:X.Y.Z-amd64`       | Só amd64                  | ~20 MB         |
| `ghcr.io/<repo-owner>/<repo>/backend:X.Y.Z-arm64`       | Só arm64                  | ~20 MB         |
| `ghcr.io/<repo-owner>/<repo>/frontend:X.Y.Z`            | Bundle Nuxt + distroless  | ~80 MB         |
| `ghcr.io/<repo-owner>/<repo>/frontend:X.Y.Z-amd64`     | Só amd64                  | ~80 MB         |
| `ghcr.io/<repo-owner>/<repo>/frontend:X.Y.Z-arm64`     | Só arm64                  | ~80 MB         |

**Adicionalmente:**

- **SBOM** (Software Bill of Materials) em formato SPDX — anexado
  à release.
- **Cosign signature** — assinado com `cosign sign` (keyless, via
  OIDC GitHub).
- **Trivy scan** — CRITICAL/HIGH bloqueia o push.
- **Provenance** (`--provenance=true`) — attestation SLSA L3.

**Tags canônicas:**

- `X.Y.Z` — versão exata (ex.: `0.1.0`)
- `X.Y` — minor flutuante (ex.: `0.1`)
- `X` — major flutuante (ex.: `0`)
- `latest` — última tag em `main` (NÃO em PRs)
- `sha-<short>` — referência exata por commit (imutável)

---

## 2. Como criar uma release

### 2.1. Via tag (recomendado)

```bash
# 1. Garantir que a main está atualizada e CI verde
git checkout main
git pull origin main
gh pr checks  # (se houver PRs abertos)

# 2. Criar tag (semver)
git tag -a v0.1.0 -m "Release v0.1.0: bootstrap + first feature"
git push origin v0.1.0

# 3. CI automaticamente:
#    - Roda todos os sensors (lint, test, vuln, contract, image, 12-factor, i18n)
#    - Builda imagens backend + frontend (multi-arch: amd64 + arm64)
#    - Roda Trivy (block em CRITICAL)
#    - Assina com cosign
#    - Push para ghcr.io/<owner>/<repo>/<service>:<tag>
#    - Cria GitHub Release com notas auto-geradas
#    - Faz upload dos SBOMs
```

### 2.2. Via workflow_dispatch (manual)

```bash
# Use quando precisar de uma release sem tag (ex.: hotfix pré-release)
gh workflow run release.yml \
  --repo <owner>/<repo> \
  -f tag=v0.1.0-rc.1
```

---

## 3. Como usar as imagens

### 3.1. Local (docker-compose)

```bash
# Pull
docker pull ghcr.io/<owner>/<repo>/backend:0.1.0
docker pull ghcr.io/<owner>/<repo>/frontend:0.1.0

# Ou use o template do docker-compose direto:
curl -sSL https://raw.githubusercontent.com/<owner>/<repo>/v0.1.0/deploy/docker-compose.yml -o docker-compose.yml
DATABASE_URL=postgres://app:app@localhost:5432/app?sslmode=disable \
  docker compose -f docker-compose.yml up -d
```

### 3.2. ECS (Fargate)

```bash
# 1. Crie um ECR mirror ou use GHCR direto
#    (ECS suporta GHCR via task definition com registryCredentials)

# 2. task-definition.json
cat > task-definition.json <<EOF
{
  "family": "my-app-backend",
  "containerDefinitions": [{
    "name": "backend",
    "image": "ghcr.io/<owner>/<repo>/backend:0.1.0",
    "essential": true,
    "portMappings": [{ "containerPort": 8080 }],
    "environment": [
      { "name": "PORT", "value": "8080" },
      { "name": "DATABASE_URL", "value": "..." }
    ]
  }]
}
EOF

# 3. Register + update service
aws ecs register-task-definition --cli-input-json file://task-definition.json
aws ecs update-service --cluster my-cluster --service backend --task-definition my-app-backend:1
```

### 3.3. EKS (Kubernetes)

```yaml
# k8s/backend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      # Importante: GHCR com GITHUB_TOKEN (imagePullSecret)
      imagePullSecrets:
        - name: ghcr-secret
      containers:
        - name: backend
          image: ghcr.io/<owner>/<repo>/backend:0.1.0
          ports:
            - containerPort: 8080
          env:
            - name: PORT
              value: "8080"
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: database-url
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8080
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8080
          securityContext:
            runAsNonRoot: true
            runAsUser: 65532
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 256Mi
```

```bash
# Criar secret GHCR
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=<github-user> \
  --docker-password=<github-pat-with-read:packages> \
  --namespace=production

# Deploy
kubectl apply -f k8s/backend-deployment.yaml
```

### 3.4. Docker Swarm

```bash
# 1. Initialize swarm (1 manager + N workers)
docker swarm init

# 2. docker-stack.yml
cat > docker-stack.yml <<EOF
version: "3.8"
services:
  backend:
    image: ghcr.io/<owner>/<repo>/backend:0.1.0
    deploy:
      replicas: 3
      restart_policy:
        condition: on-failure
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      DATABASE_URL: postgres://...
    healthcheck:
      test: ["CMD", "/server", "-healthcheck"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - app-net

  frontend:
    image: ghcr.io/<owner>/<repo>/frontend:0.1.0
    deploy:
      replicas: 2
    ports:
      - "3000:3000"
    environment:
      NUXT_PUBLIC_API_BASE: http://backend:8080/api/v1
    depends_on:
      - backend
    networks:
      - app-net

networks:
  app-net:
    driver: overlay
EOF

# 3. Deploy
docker stack deploy -c docker-stack.yml myapp
```

---

## 4. Verificação pós-deploy

### 4.1. Health checks obrigatórios

```bash
# Liveness
curl -fsS https://my-app.example.com/healthz
# → 200 {"status":"ok"}

# Readiness (verifica DB)
curl -fsS https://my-app.example.com/readyz
# → 200 {"status":"ready","checks":{"db":"ok"}}

# Metrics (Prometheus)
curl -fsS https://my-app.example.com/metrics | grep app_info
# → app_info{commit="...",go_version="1.25.0",service="my-app",version="0.1.0"}
```

### 4.2. Verificar imagem

```bash
# SBOM
cosign download sbom ghcr.io/<owner>/<repo>/backend:0.1.0

# Signature
cosign verify ghcr.io/<owner>/<repo>/backend:0.1.0 \
  --certificate-identity-regexp "https://github.com/<owner>/<repo>" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"

# Trivy (se quiser re-rodar)
trivy image ghcr.io/<owner>/<repo>/backend:0.1.0
```

---

## 5. Rollback

```bash
# ECS
aws ecs update-service --cluster my-cluster --service backend \
  --task-definition my-app-backend:<previous-version>

# EKS
kubectl set image deployment/backend backend=ghcr.io/<owner>/<repo>/backend:<previous-tag>
kubectl rollout status deployment/backend

# Swarm
docker service update --image ghcr.io/<owner>/<repo>/backend:<previous-tag> myapp_backend

# Local
docker compose down
docker pull ghcr.io/<owner>/<repo>/backend:<previous-tag>
docker compose up -d
```

---

## 6. Anti-patterns (NÃO faça)

| ❌ Errado                                                    | ✅ Certo                                                  |
|--------------------------------------------------------------|-----------------------------------------------------------|
| Buildar a imagem localmente e push manual para ECR/GHCR      | Deixar o release pipeline do CI buildar + pushar         |
| Usar tag `latest` em produção                                | Usar tag exata (`v0.1.0`) — `latest` é só conveniência dev |
| Deploy sem verificar `/readyz`                               | Esperar `/readyz=200` antes de rotear tráfego            |
| Rodar container como root (sem securityContext)              | `runAsNonRoot: true` + `runAsUser: 65532`                |
| Hardcodar DATABASE_URL no task definition                    | Usar Secrets Manager / k8s Secrets / Docker secrets      |
| Imagem sem healthcheck                                       | Healthcheck via `["CMD", "/server", "-healthcheck"]`      |
| Sem resource limits                                          | `requests: {cpu, mem}` + `limits: {cpu, mem}`             |
| Imagem sem SBOM/signature                                    | Pipeline já assina com cosign + gera SBOM                |

---

## 7. Quem usa este doc

- **`devops-engineer`** — configura o release pipeline,
  valida que imagens foram publicadas, dispara deploy.
- **`team-manager`** — fecha a issue com referência à release
  tag.
- **Time de plataforma** — consome as imagens em ECS/EKS/Swarm
  para staging + produção.

---

## 8. Referências

- [templates/.github-workflows-release.yml](../templates/.github-workflows-release.yml) — o workflow em si
- [harness/workflow/06-release-pipeline.md](../harness/workflow/06-release-pipeline.md) — o lado do orquestrador
- [harness/stack/versions.md](../harness/stack/versions.md) — versões das imagens base
- [GHCR docs](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [cosign](https://docs.sigstore.dev/cosign/overview/)
