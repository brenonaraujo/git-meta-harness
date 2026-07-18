# Workflow 03 — Snapshot Deploy (validação local)

> Antes do merge, o PR **deve** oferecer um ambiente local rodando
> para o usuário validar. Este documento define como o snapshot é
> montado, subido, e como o usuário acessa.

---

## O que é

**Snapshot** = ambiente `docker compose` que sobe o serviço (ou
serviços) + dependências (Postgres, Redis, …) na máquina do
desenvolvedor (ou do reviewer).

O PR **sempre** inclui o comando para subir e a URL local para
validar. O usuário valida **antes** do merge.

---

## Estrutura padrão

```
deploy/
├── docker-compose.yml         # compose principal
├── docker-compose.override.yml # overrides para dev (hot-reload, debug)
├── Dockerfile                 # produção (multi-stage, distroless)
├── .env.example               # template de envs
└── README.md                  # como rodar
```

### `deploy/docker-compose.yml` (mínimo)

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: app
      POSTGRES_DB: app
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U app"]
      interval: 5s
      timeout: 5s
      retries: 5

  migrate:
    build:
      context: ..
      dockerfile: deploy/Dockerfile.migrate
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DATABASE_URL: postgres://app:app@postgres:5432/app?sslmode=disable

  backend:
    build:
      context: ..
      dockerfile: deploy/Dockerfile
    ports:
      - "8080:8080"
    environment:
      PORT: "8080"
      DATABASE_URL: postgres://app:app@postgres:5432/app?sslmode=disable
      LOG_LEVEL: info
      GIN_MODE: debug
    depends_on:
      postgres:
        condition: service_healthy
      migrate:
        condition: service_completed_successfully
    healthcheck:
      test: ["CMD", "wget", "-q", "-O-", "http://localhost:8080/healthz"]
      interval: 5s
      timeout: 3s
      retries: 5

  frontend:
    build:
      context: ../frontend
      dockerfile: deploy/Dockerfile
    ports:
      - "3000:3000"
    environment:
      NUXT_PUBLIC_API_BASE: http://localhost:8080/api/v1
    depends_on:
      - backend

volumes:
  postgres_data:
```

---

## Como subir

```bash
# Subir tudo (Postgres + migrate + backend + frontend)
docker compose -f deploy/docker-compose.yml up -d --build

# Acompanhar logs
docker compose -f deploy/docker-compose.yml logs -f

# Esperar backend ficar ready
docker compose -f deploy/docker-compose.yml exec backend \
  wget -q -O- http://localhost:8080/healthz

# Derrubar
docker compose -f deploy/docker-compose.yml down
```

---

## Validação mínima (smoke)

1. **Health:** `curl -fsS http://localhost:8080/healthz` → 200 OK.
2. **Ready:** `curl -fsS http://localhost:8080/readyz` → 200 OK
   (com DB conectado).
3. **Metrics:** `curl -fsS http://localhost:8080/metrics` → texto Prometheus.
4. **Fluxo crítico** (varia por feature):
   - Login: `curl -X POST /api/v1/auth/login -d '{"email":"...","password":"..."}'`
   - Listar recurso: `curl /api/v1/<recurso>`
   - Frontend: abrir `http://localhost:3000` e ver a página principal.

---

## Onde o usuário valida

- **Se a máquina do usuário tem Docker:** ele sobe localmente e
  acessa `http://localhost:<porta>`.
- **Se o usuário não tem Docker:** `devops-engineer` ou `team-manager`
  sobe em um **staging efêmero** (ex.: fly.io, railway, ou k8s dev) e
  passa a URL. **Isso é optativo e custoso; preferir local.**

---

## Checklist de "Pronto para validação"

No comentário do PR, o builder (ou QA) **deve** ter:

- [ ] Comando `docker compose -f deploy/docker-compose.yml up -d` testado.
- [ ] URL de health: `http://localhost:8080/healthz`.
- [ ] URL do frontend (se aplicável): `http://localhost:3000`.
- [ ] Credenciais de teste (se houver): usuário/senha demo.
- [ ] Lista do que validar (1-3 fluxos críticos).
- [ ] `@<username-do-usuário>` mencionando explicitamente.

Exemplo de comentário:

```markdown
## 🧪 Pronto para validação

Sobe o ambiente completo com:
```bash
docker compose -f deploy/docker-compose.yml up -d
```

Endpoints:
- Backend: http://localhost:8080
- Frontend: http://localhost:3000
- Postgres: localhost:5432 (user: app, pass: app)

Fluxos para validar:
1. Fazer login com `user@example.com` / `secret`
2. Criar um recurso via POST `/api/v1/<recurso>`
3. Verificar `/metrics` expor `app_info{version="..."}`

@<username> pode validar?
```

---

## Validação do usuário

Usuário responde no PR:

- **✅ "validado"** → `team-manager` move para merge.
- **❌ "bug encontrado: ..."** → `team-manager` move para `in-progress`
  (com label `in-review`) e pede fix.
- **💬 dúvida** → responde e aguarda.

Se em **5 dias úteis** sem resposta, `team-manager` pergunta
ativamente (1 follow-up); se mais 5 dias, **escalona** ou **assume
validação por conta** (registrado em ADR).

---

## Snapshot em CI (opcional, mas útil para release)

```yaml
snapshot-smoke:
  name: Snapshot + smoke (CI)
  runs-on: ubuntu-latest
  needs: [build-and-scan]
  steps:
    - uses: actions/checkout@v4
    - name: Bring up
      run: docker compose -f deploy/docker-compose.yml up -d
    - name: Wait for ready
      run: |
        timeout 90 bash -c 'until curl -fsS http://localhost:8080/healthz; do sleep 2; done'
    - name: Smoke
      run: bash test/smoke/backend.sh
    - name: Teardown
      if: always()
      run: docker compose -f deploy/docker-compose.yml down
```

---

## Anti-padrões

- ❌ PR sem o bloco "Como testar localmente".
- ❌ PR que precisa de setup manual fora do compose (variáveis de
  ambiente extras, DB externo, …).
- ❌ Snapshot sem healthcheck.
- ❌ Migrações não versionadas (o compose deve rodar `migrate` antes
  do `backend`).
- ❌ Credenciais de produção no `.env.example`.
