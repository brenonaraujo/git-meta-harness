# Sensor 05 — Smoke Tests

> **Objetivo:** garantir que, **após o deploy do snapshot local**,
> o serviço está vivo, saudável, e responde aos fluxos mais críticos.
> **Quando roda:** QA (após `docker compose up`), em todo PR com label
> `in-review`.
> **Falha → ação:** **devolve o PR** com log do erro.

---

## O que é smoke

Smoke = **teste mínimo viável** pós-deploy. Não é teste funcional
completo; é o "está vivo e responde ao happy path?".

Deve ser:
- **Rápido** (≤ 2 min total).
- **Confiável** (sem flakiness — sem timing frágil).
- **Específico** (health, ready, 1-3 fluxos críticos).

---

## Comandos exatos

### Subir o ambiente

```bash
docker compose -f deploy/docker-compose.yml up -d
# Esperar o serviço estar pronto (com timeout, não sleep cego)
docker compose -f deploy/docker-compose.yml exec backend \
  bash -c 'until wget -q -O- http://localhost:8080/healthz; do sleep 1; done'
```

### Health checks obrigatórios

```bash
# Liveness
curl -fsS http://localhost:8080/healthz
# Resposta esperada: 200 + {"status":"ok"}

# Readiness
curl -fsS http://localhost:8080/readyz
# Resposta esperada: 200 + {"status":"ready"} (200 só quando DB está conectado)

# Metrics endpoint (sanity)
curl -fsS http://localhost:8080/metrics | grep -E "^app_info"
```

### Fluxos críticos (exemplos)

```bash
# Backend: criar + ler recurso
curl -fsS -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com"}'

curl -fsS http://localhost:8080/api/v1/users/1

# Frontend: página principal responde
curl -fsS http://localhost:3000/ | head -50
```

### Playwright (smoke E2E)

```ts
// tests/e2e/smoke.spec.ts
import { test, expect } from '@playwright/test'

test('homepage loads and shows hero', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('h1')).toBeVisible()
})

test('login flow works', async ({ page }) => {
  await page.goto('/login')
  await page.fill('input[name="email"]', 'user@example.com')
  await page.fill('input[name="password"]', 'secret')
  await page.click('button[type="submit"]')
  await expect(page).toHaveURL('/dashboard')
})
```

---

## Estrutura recomendada

```
test/
├── smoke/
│   ├── backend.sh        # curl-based smoke
│   ├── frontend.sh
│   └── e2e/
│       ├── playwright.config.ts
│       └── smoke.spec.ts
├── integration/
│   └── ...
└── load/
    └── gatling/
```

---

## Thresholds

| Métrica                      | Limite              |
|------------------------------|---------------------|
| Tempo total do smoke         | **≤ 2 min**         |
| Latência `/healthz`          | **< 100ms** (local) |
| Latência fluxo crítico       | **< 500ms** (local) |
| Taxa de sucesso              | **100%** (sem retry)|
| Flakiness                    | **0**               |

---

## Onde pluga no pipeline

### Local (QA agent)

```bash
# Roda local
make smoke
# (Make target que sobe compose, espera ready, roda smoke, derruba)
```

### CI (opcional, mas recomendado para release)

```yaml
smoke:
  name: Smoke (compose)
  runs-on: ubuntu-latest
  needs: [build-and-scan]
  steps:
    - uses: actions/checkout@v4
    - name: Bring up stack
      run: docker compose -f deploy/docker-compose.yml up -d
    - name: Wait for ready
      run: |
        timeout 60 bash -c 'until curl -fsS http://localhost:8080/healthz; do sleep 1; done'
    - name: Run smoke
      run: bash test/smoke/backend.sh
    - name: Run E2E
      run: pnpm exec playwright test test/smoke/e2e
    - name: Teardown
      if: always()
      run: docker compose -f deploy/docker-compose.yml down
```

---

## Falha típica & remediação

| Falha                                          | Como corrigir                                      |
|------------------------------------------------|----------------------------------------------------|
| `/healthz` 404                                 | Registrar rota no router; verificar middleware.    |
| `/healthz` 500                                 | Log do servidor; provável bug de boot.             |
| `/readyz` 503 (DB não conecta)                 | Verificar DATABASE_URL; verificar migrations.      |
| Fluxo crítico 4xx                              | Investigar handler; pode ser auth/validation.      |
| Compose up timeout                             | Aumentar timeout ou investigar saúde dos containers.|
| Playwright timeout                             | Aumentar timeout do test OU investigar lentidão.   |

---

## Quem roda

- **Local:** `quality-assurance` após builders terminarem.
- **CI:** opcional no PR; obrigatório antes do release.
- **Falha:** devolve ao builder.
