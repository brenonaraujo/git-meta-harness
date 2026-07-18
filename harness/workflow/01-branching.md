# Workflow 01 — Branching Strategy

> Padrão de branches. Trunk-based, com branches curtas por feature,
> sempre criadas pelo `team-manager`.

---

## Modelo

**Trunk-based development** com branches de feature curtas. A `main`
é a fonte da verdade; tudo entra por PR.

```
main ─────●────●────●────●────●──── (release tags aqui)
           \      \    \
            \      \    \─── feature/42-login ──●─(merge)─▶
             \      \────── feature/41-csp ──●─(merge)─▶
              \───── fix/40-bug-crítico ──●─(merge)─▶
```

---

## Padrão de nomenclatura

| Tipo         | Padrão                              | Exemplo                                |
|--------------|-------------------------------------|----------------------------------------|
| Feature      | `feature/<issue-id>-<slug>`         | `feature/42-login-jwt`                 |
| Bugfix       | `fix/<issue-id>-<slug>`             | `fix/55-race-condition-checkout`       |
| Tech debt    | `chore/<issue-id>-<slug>`           | `chore/60-extract-validate-email`      |
| Hotfix       | `hotfix/<issue-id>-<slug>`          | `hotfix/99-prod-down`                  |
| Release      | `release/vX.Y.Z`                    | `release/v0.4.0`                       |

> **Slug** = kebab-case curto (≤ 4 palavras) descrevendo o que a branch faz.

---

## Quem cria branches

> **Invariante 15 do `AGENTS.md` §8:** branches de feature/fix/chore
> são criadas pelo **`team-manager`** e delegadas no briefing. Quem
> implementa (`backend-engineer`/`frontend-engineer`) clona a
> branch. **Linha vermelha:** o `team-manager` **NÃO escreve código
> de feature**. Detalhes em
> [`personas/interactions.md`](../personas/interactions.md) e
> ADR-0006.

| Tipo de branch       | Quem cria                                       |
|----------------------|-------------------------------------------------|
| `feature/<id>-<slug>`| `team-manager` (e delega no briefing)             |
| `fix/<id>-<slug>`    | `team-manager` (e delega no briefing)             |
| `chore/<id>-<slug>`  | `team-manager` (e delega no briefing)             |
| `release/vX.Y.Z`     | `devops-engineer` (apenas)                       |
| `hotfix/<id>-<slug>` | `devops-engineer` (apenas, em emergência)        |
| `main`               | Ninguém (protegida — só merge via PR)            |

```bash
# team-manager (uma vez por issue):
git checkout main
git pull origin main
git checkout -b feature/42-login-jwt
git push -u origin feature/42-login-jwt
gh issue comment 42 --body "🤖 **team-manager → @backend-engineer
e/ou @frontend-engineer**

Branch: \`feature/42-login-jwt\` (criada por mim). Clonem e
implementem conforme o DoD."

# Builder (qualquer um):
git fetch origin
git checkout feature/42-login-jwt
# implementar + commit + push
```

> **Por que o team-manager cria a branch?** Porque ele é o único
> com visão completa de quem vai trabalhar na mesma issue (ex.:
> backend + frontend precisam da **mesma** branch). Criar
> localmente garante um único nome e evita duplicação.

> O `team-manager` **cria e delega**, mas **NÃO escreve código de
> feature**. Builders implementam; team-manager orquestra. Ver
> ADR-0006.

> Personas **não-técnicas** (`domain-expert`,
> `solutions-architect`, `quality-assurance`) **nunca** mencionam
> nome de branch nem dizem "crie branch X" — isso é trabalho do
> `team-manager`.

---

## Vida útil da branch

- **Feature:** 1-5 dias úteis.
- **Fix:** 0.5-2 dias úteis.
- **Hotfix:** ≤ 4 horas (cria release patch imediatamente após merge).
- **Release:** curta, usada apenas para bump de versão + tag.

Branches com mais de 7 dias sem commit → rebase na main.

---

## Proteções da main

Configurar via GitHub:

- ✅ Require PR before merging
- ✅ Require 1+ approval
- ✅ Require status checks (CI: lint, test, vuln, image-scan, contract, 12-factor)
- ✅ Require linear history (no merge commits)
- ✅ Require signed commits (recomendado)
- ✅ Include administrators (mesmo admins seguem as regras)
- ❌ Allow force push (NUNCA)
- ❌ Allow deletion (NUNCA)

---

## Rebase vs merge

- **Builders fazem rebase** da main na branch antes de pedir review.
- **`team-manager` faz merge** (squash ou merge commit) ao aceitar PR.
- **Nunca** force-push depois do review começar (a menos que combinem
  explicitamente).

```bash
# Builder rebasa:
git fetch origin
git rebase origin/main
git push --force-with-lease
```

---

## Conflitos

- Conflitos entre branches de builders = **`team-manager` resolve**
  (com consulta ao `solutions-architect` se for decisão arquitetural).
- Conflito com a main durante rebase = builder resolve.
- Conflito na hora do merge = `team-manager` resolve, pedindo ajuda se
  necessário.

---

## Limpeza

Após merge:

- Branch é **deletada automaticamente** pelo GitHub (configurar
  `Automatically delete head branches`).
- Branch local: `git branch -d feature/42-login-jwt`.
- Branch de hotfix: também deletada após release patch.

---

## Anti-padrões

- ❌ `develop` como branch de longo prazo (use main + flags).
- ❌ `master` (sempre `main`).
- ❌ Branches sem issue relacionada.
- ❌ Branches > 7 dias sem progresso.
- ❌ Force push na main.
- ❌ Commits direto na main.
- ❌ Commits sem Conventional Commits.
