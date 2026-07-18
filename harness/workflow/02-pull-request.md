# Workflow 02 — Pull Request

> Padrão de PR. **1 issue → 1 PR**. Vários builders podem commitar
> na mesma branch, mas **só existe 1 PR por feature**.

---

## Regra de ouro

> **1 issue = 1 branch = 1 PR = 1 feature entregue.**

Múltiplos builders (ex.: `backend-engineer` + `frontend-engineer`)
podem fazer commits na **mesma** branch, mas o PR é **único**,
aberto quando ambos terminam.

---

## Título

```
(#<issue-id>) <título curto da feature>
```

Exemplos:

- `(#42) Implementa login com JWT`
- `(#55) Corrige race condition no checkout`
- `(#60) Extrai validação de email para helper`

---

## Corpo (template)

> O template completo está em
> [`../templates/pr-description.md`](../templates/pr-description.md).
> Resumo:

```markdown
## Summary
(1 parágrafo do que foi feito)

## Issue
Closes #<id>

## Changes
- [ ] Mudança 1
- [ ] Mudança 2

## Sensores (todos verdes)
- [ ] `make lint` — OK
- [ ] `make test` — coverage ≥ 80% nos pacotes alterados
- [ ] `govulncheck` — sem HIGH/CRITICAL
- [ ] `trivy image` — sem CRITICAL
- [ ] `openapi-diff` — sem breaking changes
- [ ] `12-factor audit` — F1..F12 OK

## Como testar localmente
```bash
docker compose -f deploy/docker-compose.yml up -d
curl http://localhost:8080/healthz
# UI: http://localhost:3000
```

## Screenshots / curls
(anexar)

## Riscos & rollback
- Risco: ...
- Rollback: reverter merge (ou `git revert <sha>` + tag `v<X.Y.Z-1>`)
```

---

## Quem abre

**`backend-engineer` ou `frontend-engineer`** abre o PR (ou o último
a terminar, em caso de trabalho conjunto). O `team-manager` revisa
e mergeia.

```bash
gh pr create \
  --base main \
  --title "(#42) Login com JWT" \
  --body-file .github/PULL_REQUEST_TEMPLATE.md \
  --reviewer <backend-engineer-1>,<frontend-engineer-2> \
  --label "backend,frontend"
```

---

## Quem revisa

| Persona              | Revisa PRs de                                           |
|----------------------|---------------------------------------------------------|
| `backend-engineer`   | outros backends (peer review)                           |
| `frontend-engineer`  | outros frontends (peer review)                          |
| `quality-assurance`  | qualquer PR (sempre, como gate final)                   |
| `solutions-architect`| PRs com mudança arquitetural (OpenAPI, schema, lib)      |
| `devops-engineer`    | PRs com mudança em Dockerfile, workflow, infra          |
| `team-manager`       | qualquer PR (sempre, no fim)                            |

---

## Status checks obrigatórios

O GitHub deve **bloquear merge** se algum desses falhar:

- `lint` — golangci-lint + ESLint
- `test` — go test + coverage
- `vuln` — govulncheck + pnpm audit
- `build-and-scan` — docker build + trivy
- `contract` — openapi-diff + schemathesis
- `twelve-factor` — script de auditoria
- `smoke` — (em release, opcional no PR)

Configurar em **Settings → Branches → Branch protection rules**.

---

## Tamanho

- **Ideal:** < 400 linhas de diff.
- **Máximo aceitável:** 800 linhas (com justificativa).
- **> 800 linhas:** quebrar em PRs menores.

> PRs grandes são bloqueados pelo `team-manager` na triagem.

---

## Draft PR

Builders podem abrir **Draft PR** cedo para feedback contínuo:

```bash
gh pr create --draft --title "..." --body "..."
```

Para marcar como ready:

```bash
gh pr ready
```

---

## Comentários

- **Reviewers** comentam inline ou em review geral.
- **Builder** responde com commit adicional (não com novo PR).
- **`team-manager`** decide se aprova ou pede mudanças.
- **`quality-assurance`** posta o relatório final (ver
  `sensors/05-smoke-tests.md`).

---

## Merge

**Apenas `team-manager` mergeia.**

Estratégia: **squash and merge** (mantém main linear e changelog limpo).
O squash inclui a referência à issue: `(#42) feat(auth): login com JWT`.

Após merge:

1. Branch é deletada (auto).
2. Workflow de release dispara (se em main).
3. Issue é fechada (com referência ao PR via `Closes #42`).

---

## Fechar PR sem merge

Se a abordagem mudou ou a issue foi cancelada:

- **`team-manager`** fecha com comentário justificando.
- Se houver trabalho útil, **cria nova issue** e **abre novo PR**.
- Nunca force-push para "apagar" commits depois do review.

---

## Anti-padrões

- ❌ PR sem link para issue.
- ❌ PR sem "Como testar localmente".
- ❌ PR com 1500+ linhas sem justificativa.
- ❌ PR com sensor falhando.
- ❌ Auto-merge.
- ❌ Merge antes do QA aprovar.
- ❌ Merge antes do usuário validar.
