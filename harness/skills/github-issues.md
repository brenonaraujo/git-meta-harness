# Skill — github-issues

> Skill para criar, triar, comentar, e mover issues.
> Usada por **todas** as personas.

---

## Quando carregar

- Criar issue.
- Comentar em issue.
- Mover label.
- Atribuir.
- Fechar.

---

## Comandos canônicos

### Criar issue

```bash
# Feature
gh issue create \
  --title "✨ Feature: <título>" \
  --body-file harness/templates/issue-feature.md \
  --label "triage,backend" \
  --milestone "v0.4.0"

# Bug
gh issue create \
  --title "🐛 Bug: <título>" \
  --body-file harness/templates/issue-bug.md \
  --label "triage,bug,severity/high" \
  --milestone "v0.4.0"

# Tech debt
gh issue create \
  --title "🧹 Tech debt: <título>" \
  --body-file harness/templates/issue-tech-debt.md \
  --label "triage,tech-debt,backend"
```

### Triar

```bash
# Ver issue
gh issue view 42

# Aplicar label inicial
gh issue edit 42 --add-label "triage" --add-assignee @me

# Se precisa de mais info
gh issue edit 42 --add-label "needs-info" --remove-label "triage"
gh issue comment 42 --body "🤖 Preciso de mais informações: ..."

# Se pode seguir
gh issue edit 42 --remove-label "triage" --add-label "refined"
gh issue edit 42 --remove-assignee @me --add-assignee <domain-expert>
```

### Mover entre estados

```bash
# refined → ready
gh issue edit 42 --remove-label "refined" --add-label "ready"
gh issue edit 42 --remove-assignee <domain-expert> --add-assignee <solutions-architect>

# ready → in-progress
gh issue edit 42 --remove-label "ready" --add-label "in-progress"
gh issue edit 42 --remove-assignee <architect> --add-assignee <builder>

# in-progress → in-review
gh issue edit 42 --remove-label "in-progress" --add-label "in-review"
gh issue edit 42 --remove-assignee <builder> --add-assignee <qa>

# in-review → qa (aprovado)
gh issue edit 42 --remove-label "in-review" --add-label "qa"
gh issue edit 42 --remove-assignee <qa> --add-assignee <user>

# qa → done (após validação)
gh issue edit 42 --remove-label "qa" --add-label "done"
gh issue close 42 --comment "✅ Entregue."
```

### Comentar

```bash
# Status
gh issue comment 42 --body "🤖 **team-manager**: PR #43 aberto e CI verde. Seguindo para QA."

# Bloqueio
gh issue comment 42 --body "🚧 **backend-engineer**: bloqueado por #45 (DB precisa de migration)."

# Decisão (ADR-lite)
gh issue comment 42 --body "📐 **solutions-architect**: Decisão — usar JWT com TTL 24h.
Motivo: balancear segurança vs UX. Reversível se necessário."
```

### Vincular PR

```bash
# No PR body: "Closes #42" (GitHub fecha automaticamente no merge)

# Ou manualmente:
gh issue edit 42 --add-assignee <pr-author>
gh pr edit 43 --add-label "closes-42"
```

### Fechar

```bash
gh issue close 42 --comment "✅ Entregue no release v0.4.0."
```

### Listar

```bash
# Issues abertas por label
gh issue list --state open --label in-review

# Issues atribuídas a mim
gh issue list --state open --assignee @me

# Issues em milestone
gh issue list --milestone "v0.4.0"
```

---

## Convenções

- **Comentários** sempre começam com emoji + persona:
  `🤖 **team-manager**`, `🤝 **domain-expert**`, `🏗️ **solutions-architect**`,
  `🛠️ **backend-engineer**`, `🎨 **frontend-engineer**`, `🔍 **quality-assurance**`,
  `🚀 **devops-engineer**`.
- **Decisões** em ADR-lite, datadas, no comentário (link para issue).
- **Bloqueios** com label `blocked` e motivo claro.
- **Waivers** em comentário datado com CVE id + plano + prazo.

---

## Quem carrega

- Todas as personas.
