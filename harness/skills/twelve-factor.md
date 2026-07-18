# Skill — twelve-factor

> Skill para auditar e aplicar os 12 fatores em microsserviços Go.
> Usada por `solutions-architect` (no DoD) e `devops-engineer`
> (no release).

---

## Os 12 fatores (cheat-sheet)

| #    | Fator              | Regra de ouro                                             |
|------|--------------------|-----------------------------------------------------------|
| I    | Codebase           | 1 repo = 1 app, várias deploys                            |
| II   | Dependencies       | Declare explicitamente, isole                             |
| III  | Config             | No ambiente, nunca em código                              |
| IV   | Backing services   | Recurso anexado, conectado por URL                        |
| V    | Build/Release/Run  | 3 estágios estritos                                        |
| VI   | Processes          | Stateless, estado em backing services                     |
| VII  | Port binding       | Exporte via PORT, sem servidor embutido                   |
| VIII | Concurrency        | Escale via processo (horizontal)                          |
| IX   | Disposability      | Startup rápido, shutdown gracioso                         |
| X    | Dev/prod parity    | Ambientes similares (mesma imagem, mesmas migrations)    |
| XI   | Logs               | Event stream, stdout, sem arquivos                        |
| XII  | Admin processes    | One-off processes (migrations, seed, cleanup)             |

---

## Como auditar (passo a passo)

### I. Codebase

```bash
# 1 repo = 1 app: existe apenas um módulo Go?
ls go.mod
# E um só main?
ls cmd/
# Sem código de outro serviço no repo?
find . -name "main.go" -not -path "./vendor/*"
```

✅ 1 repo, 1 go.mod, 1 cmd/server.
❌ Múltiplos serviços no mesmo repo (monorepo) sem motivo.

### II. Dependencies

```bash
# go.mod + go.sum commitados?
ls go.mod go.sum
# Pin de versão direta?
grep -E "^\trequire" go.mod
```

✅ go.mod + go.sum commitados, versões pinned.
❌ Sem go.sum; dependências globais no sistema.

### III. Config

```bash
# Zero config hardcoded
grep -rE "(password|secret|api_key|database_url)\s*=\s*[\"'][^\"']+[\"']" \
  internal/ cmd/ | grep -v _test.go

# Config via envconfig?
grep -r "envconfig" internal/
```

✅ Tudo via env (envconfig); zero hardcode.
❌ `const dbURL = "postgres://..."` em código.

### IV. Backing services

```bash
# URLs via env?
grep -rE "DATABASE_URL|REDIS_URL|RABBITMQ_URL" internal/
# Sem IP/host hardcoded?
grep -rE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b" internal/ | grep -v _test.go
```

✅ Conexões por URL env.
❌ `db.Connect("10.0.0.1:5432")`.

### V. Build / Release / Run

```bash
# 3 estágios separados no CI?
cat .github/workflows/ci.yml | grep -E "build|release|run"
# Dockerfile multi-stage?
grep -E "^FROM" deploy/Dockerfile
# Tags imutáveis?
grep "tags:" .github/workflows/release.yml
```

✅ CI tem job de build separado, release com tag.
❌ `docker run --env` em produção; sem release versionado.

### VI. Processes

```bash
# Sem sessão local?
grep -rE "(session\.|sync\.Map|local cache)" internal/ | grep -v _test.go
# JWT validado a cada request?
grep -rE "jwt\.Parse|VerifyToken" internal/
```

✅ Stateless; JWT validado a cada request.
❌ Cache de sessão em memória; cache local sem TTL.

### VII. Port binding

```bash
# Lê de PORT env?
grep -rE "os\.Getenv\(\"PORT\"\)|envconfig.*PORT" internal/ cmd/
# Servidor embutido em lib?
grep -rE "http\.ListenAndServe" internal/ | grep -v _test.go
```

✅ `os.Getenv("PORT")` no main; lib não sobe servidor.
❌ Porta hardcoded (`:8080` direto).

### VIII. Concurrency

```bash
# Stateless = escala horizontal funciona?
# (manual, mas verificar: não há estado local)
```

✅ Funciona com `docker compose scale backend=5`.
❌ Estado em memória; race condition.

### IX. Disposability

```bash
# SIGTERM handler?
grep -rE "signal\.Notify.*SIGTERM|NotifyContext" cmd/
# Timeout no shutdown?
grep -rE "WithTimeout.*shutdown" cmd/
# Startup rápido?
time go run ./cmd/server
```

✅ `signal.NotifyContext(ctx, SIGINT, SIGTERM)`; timeout 30s no
`server.Shutdown`; startup < 5s.
❌ Sem handler; `os.Exit` direto.

### X. Dev/prod parity

```bash
# Mesma imagem em dev e prod?
# (verificar que o Dockerfile é o mesmo)
# Sem dev.sh e prod.sh diferentes?
ls scripts/
# Migrations aplicadas em dev e prod da mesma forma?
grep -rE "migrate up|migrate -path" scripts/ deploy/
```

✅ Mesma imagem; mesmo compose em CI; migrations via `migrate up`.
❌ `scripts/dev.sh` roda migration, `scripts/prod.sh` não.

### XI. Logs

```bash
# stdout JSON?
grep -rE "slog\.NewJSONHandler.*os\.Stdout" internal/
# Sem log em arquivo?
grep -rE "ioutil\.WriteFile.*\.log|os\.Create.*\.log" internal/
# Sem fmt.Println?
grep -rE "fmt\.Println" internal/ cmd/ | grep -v _test.go
```

✅ `slog.NewJSONHandler(os.Stdout, ...)`; sem arquivos.
❌ `logrus.SetOutput(file)`; `fmt.Println` em produção.

### XII. Admin processes

```bash
# Binários one-off?
ls cmd/
# (esperado: cmd/server, cmd/migrate, cmd/seed, etc.)
```

✅ `cmd/migrate/`, `cmd/seed/`, `cmd/cleanup/`.
❌ Tudo no mesmo `cmd/server` com flag `--mode=migrate`.

---

## Auditoria automatizada

Use o script:

```bash
./scripts/check-twelve-factor.sh .
```

(Ver `sensors/07-twelve-factor-audit.md` para o script completo.)

---

## Quem carrega

- `solutions-architect` (no DoD, antes de `ready`).
- `devops-engineer` (no release, antes de tag).
- `backend-engineer` (auto-check antes de pedir review).
