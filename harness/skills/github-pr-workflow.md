# Skill — github-pr-workflow

> Skill usada por todas as personas que mexem em PRs (`backend-engineer`,
> `frontend-engineer`, `quality-assurance`, `devops-engineer`).
> Padroniza o ciclo de vida de um PR.

---

## Quando carregar

- Vai abrir um PR.
- Vai revisar um PR.
- Vai mergear um PR.
- Vai disparar release após merge.

---

## Fluxo (passo a passo)

### 1. Abrir PR

```bash
# 1. Garantir que está na branch certa e atualizada
git checkout feature/<id>-<slug>
git fetch origin
git rebase origin/main
git push --force-with-lease

# 2. Garantir que CI está verde localmente
make lint
make test
make vuln
pnpm lint  # se frontend
pnpm test  # se frontend
pnpm typecheck  # se frontend

# 3. Abrir PR
gh pr create \
  --base main \
  --title "(#<id>) <título>" \
  --body-file .github/PULL_REQUEST_TEMPLATE.md \
  --reviewer <peer1>,<peer2> \
  --label "backend,frontend"  # ou o que for
```

### 2. Revisar PR (peer)

- Ler descrição.
- Ler DoD da issue.
- Conferir se o bloco "Como testar localmente" funciona:
  ```bash
  gh pr checkout <pr-number>
  docker compose -f deploy/docker-compose.yml up -d --build
  # Testar fluxos
  ```
- Comentar inline (sugestões) ou em review geral.
- Aprovar ou pedir mudanças:
  ```bash
  gh pr review <pr-number> --approve
  # ou
  gh pr review <pr-number> --request-changes --body "..."
  ```

### 3. Rebasear e atualizar

```bash
git fetch origin
git rebase origin/main
git push --force-with-lease
# CI re-roda
```

### 4. Merge (apenas `team-manager`)

```bash
# Validar que CI está verde, QA aprovou, usuário validou
gh pr view <pr-number> --json reviewDecision,statusCheckRollup
gh pr view <pr-number> --comments  # verificar "validado" do usuário

# Squash merge
gh pr merge <pr-number> --squash --delete-branch
```

### 5. Pós-merge

```bash
# Branch já foi deletada pelo GitHub
git checkout main
git pull origin main
git branch -d feature/<id>-<slug>
```

---

## Regras duras

- **1 issue = 1 PR** (mesma feature entregue por múltiplos builders
  = 1 PR único).
- **PR sem "Como testar localmente"** = bloquear.
- **PR com sensor falhando** = bloquear.
- **PR com breaking change no OpenAPI** = bloquear (sem waiver).
- **Auto-merge** = proibido.
- **Merge antes de QA aprovar** = proibido.
- **Merge antes do usuário validar** = proibido.

---

## Comandos úteis

```bash
# Listar PRs em review
gh pr list --state open --label in-review

# Ver checks de um PR
gh pr checks <pr-number>

# Comentar no PR
gh pr comment <pr-number> --body "..."

# Adicionar label
gh pr edit <pr-number> --add-label "qa"

# Fechar sem merge
gh pr close <pr-number> --comment "Motivo: ..."
```

---

## Quem carrega

- `backend-engineer`, `frontend-engineer`, `quality-assurance`,
  `devops-engineer`, `team-manager`.
