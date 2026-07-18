# Sensor 07 — Twelve-Factor Audit

> **Objetivo:** garantir que cada microsserviço adere aos 12 fatores
> do [`12factor.net`](https://12factor.net/).
> **Quando roda:** `solutions-architect` no DoD; `devops-engineer` no
> release; check automatizado em CI (script `check-twelve-factor.sh`).
> **Falha → ação:** **bloqueia merge** se algum fator ❌.

---

## Os 12 fatores (resumo executivo)

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

## Auditoria automatizada (script)

`scripts/check-twelve-factor.sh`:

```bash
#!/usr/bin/env bash
# Audita 12 fatores automaticamente.
# Usage: ./scripts/check-twelve-factor.sh [service-dir]
set -e

SERVICE="${1:-.}"
FAILS=0
echo "🔎 12-Factor audit: $SERVICE"
echo

# I. Codebase
echo -n "I. Codebase ... "
[ -d "$SERVICE/.git" ] && echo "✅" || { echo "❌ no .git"; FAILS=$((FAILS+1)); }

# II. Dependencies (go.mod + go.sum)
echo -n "II. Dependencies ... "
if [ -f "$SERVICE/go.mod" ]; then
  [ -f "$SERVICE/go.sum" ] && echo "✅" || { echo "❌ no go.sum"; FAILS=$((FAILS+1)); }
else
  echo "ℹ️  (não-Go, skip)"
fi

# III. Config — sem config em código
echo -n "III. Config (no hardcoded) ... "
if grep -rE "(DB_URL|database_url|API_KEY|secret_key)\s*[:=]\s*[\"'][^\"']+[\"']" \
     "$SERVICE/internal" "$SERVICE/cmd" 2>/dev/null | grep -v "_test.go" | grep -v "test" >/dev/null; then
  echo "❌ hardcoded config encontrado"
  FAILS=$((FAILS+1))
else
  echo "✅"
fi

# IV. Backing services via env
echo -n "IV. Backing services via env ... "
if grep -rE "DATABASE_URL|REDIS_URL|RABBITMQ_URL" "$SERVICE/internal" 2>/dev/null >/dev/null; then
  echo "✅"
else
  echo "⚠️  (nenhuma URL de backing service detectada)"
fi

# V. Build/Release/Run — Dockerfile presente
echo -n "V. Dockerfile presente ... "
[ -f "$SERVICE/Dockerfile" ] || [ -f "$SERVICE/deploy/Dockerfile" ] \
  && echo "✅" || { echo "❌"; FAILS=$((FAILS+1)); }

# VI. Processes — stateless
echo -n "VI. Stateless (sem sessions locais) ... "
if grep -rE "(sync\.Map|local cache|session\[)" "$SERVICE/internal" 2>/dev/null \
   | grep -v "_test.go" >/dev/null; then
  echo "⚠️  (verificar uso de state local)"
else
  echo "✅"
fi

# VII. Port binding — PORT env
echo -n "VII. PORT env ... "
if grep -rE "os\.Getenv\(\"PORT\"\)|envconfig.*PORT" "$SERVICE/internal" "$SERVICE/cmd" 2>/dev/null >/dev/null; then
  echo "✅"
else
  echo "❌ PORT env não detectado"
  FAILS=$((FAILS+1))
fi

# VIII. Concurrency — sem in-process worker que deveria ser externo
echo -n "VIII. Sem worker que deveria ser externo ... "
# (heurística: sem flag específica aqui; manual no DoD)
echo "ℹ️  (manual no DoD)"

# IX. Disposability — graceful shutdown (SIGTERM)
echo -n "IX. Graceful shutdown ... "
if grep -rE "signal\.Notify.*SIGTERM|context\.WithTimeout" "$SERVICE/cmd" 2>/dev/null >/dev/null; then
  echo "✅"
else
  echo "❌ sem handler SIGTERM"
  FAILS=$((FAILS+1))
fi

# X. Dev/prod parity — mesma imagem
echo -n "X. Sem scripts de dev/prod separados ... "
if [ -f "$SERVICE/scripts/dev.sh" ] && [ -f "$SERVICE/scripts/prod.sh" ]; then
  echo "⚠️  (verificar se são realmente diferentes)"
else
  echo "✅"
fi

# XI. Logs — stdout, JSON, sem arquivos
echo -n "XI. Logs em stdout JSON ... "
if grep -rE "slog\.NewJSONHandler.*os\.Stdout" "$SERVICE/internal" 2>/dev/null >/dev/null; then
  echo "✅"
else
  echo "❌ slog JSON handler não detectado"
  FAILS=$((FAILS+1))
fi
if grep -rE "log\.(Fatal|Error).*ioutil\.WriteFile|os\.Create.*\.log" "$SERVICE/internal" 2>/dev/null >/dev/null; then
  echo "❌ log.Fatal em arquivo detectado"
  FAILS=$((FAILS+1))
fi

# XII. Admin processes — cmd/migrate, cmd/seed, etc.
echo -n "XII. Admin processes ... "
if [ -d "$SERVICE/cmd/migrate" ] || [ -d "$SERVICE/cmd/seed" ] || [ -d "$SERVICE/cmd/admin" ]; then
  echo "✅"
else
  echo "ℹ️  (nenhum cmd/admin encontrado; OK se não houver admin tasks)"
fi

echo
if [ $FAILS -gt 0 ]; then
  echo "❌ $FAILS fator(es) falharam."
  exit 1
fi
echo "✅ Todos os fatores auditados."
```

---

## Checklist manual (no DoD do `solutions-architect`)

| Fator | Como validar                                                    | Bloqueia? |
|-------|-----------------------------------------------------------------|-----------|
| I     | Repo único; nenhuma pasta com código de outro serviço.          | ✅        |
| II    | `go.mod` + `go.sum` (ou `package.json` + lockfile) commitados.  | ✅        |
| III   | Zero config hardcoded; apenas env via `envconfig`.              | ✅        |
| IV    | Postgres/Redis/etc. conectados por URL env.                     | ✅        |
| V     | CI separa build (Dockerfile) / release (tag) / run (deploy).    | ✅        |
| VI    | Sem sessão local; JWT validado a cada request.                  | ✅        |
| VII   | `PORT` env; sem servidor embutido em libs.                      | ✅        |
| VIII  | `docker compose scale` ou k8s replicas funciona.                 | ⚠️ info   |
| IX    | SIGTERM handler com timeout 30s.                                | ✅        |
| X     | Mesma imagem em dev/staging/prod.                              | ✅        |
| XI    | `slog.NewJSONHandler(os.Stdout, ...)`; zero log em arquivo.      | ✅        |
| XII   | `cmd/migrate/`, `cmd/seed/` como binários separados.            | ⚠️ info   |

---

## Onde pluga no pipeline

### CI (`.github/workflows/ci.yml`)

```yaml
twelve-factor:
  name: 12-Factor audit
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - run: chmod +x scripts/check-twelve-factor.sh
    - run: ./scripts/check-twelve-factor.sh .
```

### Local (pre-commit)

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: twelve-factor
        name: 12-Factor audit
        entry: scripts/check-twelve-factor.sh
        language: system
        pass_filenames: false
```

---

## Falha típica & remediação

| Falha                                         | Como corrigir                                                |
|-----------------------------------------------|--------------------------------------------------------------|
| Config hardcoded                              | Mover para env; usar `envconfig`.                           |
| Sem `PORT` env                                | Trocar `Listen(":8080")` por `os.Getenv("PORT")`.            |
| Sem SIGTERM handler                           | Adicionar `signal.NotifyContext(ctx, syscall.SIGTERM)`.     |
| Logs em arquivo                               Trocar para `slog.NewJSONHandler(os.Stdout, ...)`.             |
| Sem `cmd/migrate/`                            | Criar `cmd/migrate/main.go` que chama `golang-migrate`.      |

---

## Quem roda

- **CI:** workflow `ci.yml` (job `twelve-factor`).
- **Manual (DoD):** `solutions-architect` na transição `refined` → `ready`.
- **Manual (release):** `devops-engineer` no release.
- **Falha:** bloqueia merge.
