# Workflow 05 — Orchestration (como o team-manager coordena)

> Como o `team-manager` orquestra o fluxo de issues, atribui
> personas, e enforce os princípios. Pensado para ser o "playbook"
> que o team-manager segue em todo turno.

---

## Loop principal (pseudocódigo)

```python
def orchestrate():
    # 1. Listar issues ativas
    issues = gh.issue_list(state="open", labels=["triage","refined","ready","in-progress","in-review","qa","needs-info"])

    for issue in issues:
        label = issue.primary_label  # assume single primary label do conjunto canônico

        match label:
            case "triage":
                handle_triage(issue)
            case "refined":
                # domain-expert já postou; só aguarda
                wait_for_solutions_architect(issue)
            case "ready":
                handle_ready(issue)
            case "in-progress":
                wait_for_pr(issue)
            case "in-review":
                wait_for_qa(issue)
            case "qa":
                wait_for_user_validation(issue)
            case "needs-info":
                wait_for_author(issue)
            case "blocked":
                monitor_blocked(issue)
            case _:
                log_warning(f"Issue #{issue.id} com label inesperada: {label}")

def handle_triage(issue):
    """Triagem inicial."""
    if needs_more_info(issue):
        gh.issue_edit(issue.id, add_labels=["needs-info"], remove_labels=["triage"])
        gh.issue_comment(issue.id, "🤖 team-manager: preciso de mais informações. @<autor>, pode detalhar ...")
    else:
        gh.issue_edit(issue.id, add_labels=["refined"], remove_labels=["triage"])
        gh.issue_edit(issue.id, add_assignee="<domain-expert>")

def handle_ready(issue):
    """DoD pronto; criar branch e atribuir."""
    slug = slugify(issue.title)
    branch = f"feature/{issue.id}-{slug}"
    create_branch(branch, base="main")
    gh.issue_comment(issue.id, f"🤖 team-manager: branch `{branch}` criada. Atribuindo ao builder.")
    gh.issue_edit(issue.id, add_labels=["in-progress"], remove_labels=["ready"])
    if is_backend(issue):
        gh.issue_edit(issue.id, add_assignee="<backend-engineer>")
    if is_frontend(issue):
        gh.issue_edit(issue.id, add_assignee="<frontend-engineer>")
    if is_infra(issue):
        gh.issue_edit(issue.id, add_assignee="<devops-engineer>")

def wait_for_pr(issue):
    """Espera PR abrir e virar in-review."""
    prs = gh.pr_list(linked_issue=issue.id)
    if not prs:
        return  # sem PR ainda; só espera
    pr = prs[0]
    if pr.state == "open":
        # PR existe; verificar se draft ou ready
        if pr.draft:
            return
        # PR ready; mover para in-review
        if "in-review" not in pr.labels:
            gh.issue_edit(issue.id, remove_labels=["in-progress"], add_labels=["in-review"])
            gh.issue_edit(issue.id, add_assignee="<quality-assurance>")
            gh.issue_comment(issue.id, "🤖 team-manager: PR #<pr> pronto. Atribuindo ao QA.")

def wait_for_qa(issue):
    """Aguarda QA postar relatório."""
    if qa_approved(issue):
        gh.issue_edit(issue.id, remove_labels=["in-review"], add_labels=["qa"])
        gh.issue_edit(issue.id, add_assignee="<user-validator>")
        pr = linked_pr(issue)
        gh.pr_comment(pr.id, "🤖 team-manager: QA aprovou. @<user>, pode validar localmente e responder 'validado' aqui?")
    elif qa_rejected(issue):
        bugs = parse_bugs_from_last_comment(issue)
        gh.issue_edit(issue.id, remove_labels=["in-review"], add_labels=["in-progress"])
        gh.issue_edit(issue.id, add_assignee="<builder>")
        gh.issue_comment(issue.id, f"🤖 team-manager: QA reprovou. Bugs: {bugs}. Corrigir e reabrir PR.")

def wait_for_user_validation(issue):
    """Aguarda 'validado' do usuário no PR."""
    pr = linked_pr(issue)
    if user_validated(pr):
        # Merge + release
        gh.pr_merge(pr.id, squash=True)
        gh.issue_edit(issue.id, remove_labels=["qa"], add_labels=["done"])
        gh.issue_close(issue.id, comment="✅ Mergeado. Release será disparado pelo devops-engineer.")
        gh.issue_edit(issue.id, add_assignee="<devops-engineer>")
        # Disparar release workflow
        gh.workflow_run("release.yml")
    elif user_rejected(pr):
        # Volta para in-progress
        gh.issue_edit(issue.id, remove_labels=["qa"], add_labels=["in-progress"])
        gh.issue_edit(issue.id, add_assignee="<builder>")

def monitor_blocked(issue):
    """Issue parada; perguntar progresso a cada 1 dia útil."""
    if days_since_last_comment(issue) >= 1:
        gh.issue_comment(issue.id, "🤖 team-manager: ainda bloqueado? Tem novidade?")
```

---

## Princípios de orquestração

### 1. Não pular etapas

Sempre: `triage` → `refined` → `ready` → `in-progress` → `in-review`
→ `qa` → validação → `done`.

Se tentar pular, o `team-manager` recusa e devolve.

### 2. Paralelizar o que dá

Backend e frontend podem trabalhar **na mesma branch** (em arquivos
separados). O `team-manager` atribui os dois e o PR é único.

### 3. Enforce invariantes

Antes de cada transição, o `team-manager` valida:

- [ ] Labels canônicos (sem `enhancement`/`bug` se já temos
  `backend`/`frontend`).
- [ ] Branch segue o padrão `feature/<id>-<slug>`.
- [ ] PR referencia a issue.
- [ ] PR tem "Como testar localmente".
- [ ] Sensores passaram (CI verde).
- [ ] QA aprovou (se chegou em `qa`).
- [ ] Usuário validou (se chegou em `done`).

### 4. Não fechar issues sem validação

Mesmo com QA aprovando, **a issue só fecha com "validado" do
usuário**. Exceção: issues triviais (typo, doc) que não precisam
de snapshot, com waiver registrado.

### 5. Comunicação

- Toda transição tem **comentário de status** na issue.
- Toda decisão tem **comentário datado** (ADR-lite).
- Toda exceção (waiver) tem **motivo + plano + prazo**.

### 6. Visibilidade

- Atualizar o **GitHub Project** (board kanban) a cada transição.
- Resumir o **status semanal** em issue `weekly-status` (criada
  automaticamente pelo `team-manager` na segunda-feira).

---

## Comportamento em exceções

### CI falhou no PR

1. `team-manager` comenta na issue: "CI falhou em `<job>`. Ver log
   em <link>."
2. Remove label de transição (mantém `in-progress` ou `in-review`).
3. Atribui ao builder responsável.
4. Builder corrige, push, CI re-roda.

### PR abriu com sensor falhando

1. Bloquear merge (status check required).
2. Comentar na issue: "PR #X tem sensor Y falhando. Corrigir antes
   do merge."
3. Builder corrige.

### Builder sumiu (1+ dia útil sem commit)

1. `team-manager` comenta: "Sem movimento há X dias. Status?"
2. Se não responder em 1 dia, escalona (atribui a outro builder ou
   avisa o time).

### Usuário não valida (5+ dias úteis)

1. `team-manager` comenta 1 follow-up.
2. Se mais 5 dias, **escalona** ou **assume validação** (com waiver
   registrado: "validado por @team-manager; usuário sem resposta
   há X dias").

### Múltiplas pessoas trabalhando na mesma branch

- Coordenado pelo `team-manager` na criação da branch.
- Cada builder cuida de **paths separados** (backend em `internal/`,
  frontend em `app/`).
- Conflitos = `team-manager` resolve com ajuda do `solutions-architect`.

---

## Métricas de orquestração

O `team-manager` é avaliado por:

- **Lead time** (issue criada → done). Meta: ≤ 5 dias úteis.
- **Cycle time** (in-progress → done). Meta: ≤ 3 dias úteis.
- **% issues com sensor falhando no PR**: meta: ≤ 5%.
- **% issues sem "validado" do usuário antes de fechar**: meta: 0%
  (sem waiver).
- **Tempo de resposta do `team-manager` a transições**: meta: ≤ 1
  dia útil.
