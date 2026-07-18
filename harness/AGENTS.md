# Meta-Harness — AGENTS.md (contrato multi-tool)

> Este arquivo é o **contrato** que o `team-manager` materializa nos
> layouts nativos de cada tool suportado. Cada seção é gerada a partir
> das personas, sensors e workflow definidos em `bootstrap.md`.
>
> **Versão do meta-harness:** 1.5.0 (jul/2026) — verify-after-build.
> **Licença:** MIT. **Status:** stable.

## Para qual tool você está lendo isto?

| Tool              | Layout esperado pelo tool                                | Este arquivo vira…                       |
|-------------------|-----------------------------------------------------------|------------------------------------------|
| **Claude Code**   | `CLAUDE.md` na raiz                                       | `CLAUDE.md` (copie/conecte este arquivo) |
| **Codex CLI**     | `AGENTS.md` na raiz                                       | `AGENTS.md` (já está no path correto)    |
| **OpenCode**      | `AGENTS.md` + `.opencode/`                                | `AGENTS.md`                              |
| **Devin CLI**     | `AGENTS.md`                                               | `AGENTS.md`                              |
| **GitHub Copilot**| `.github/copilot-instructions.md`                         | gere a partir daqui                      |
| **Cursor**        | `.cursorrules`                                            | gere a partir daqui                      |
| **Hermes Agent**  | `~/.hermes/skills/<name>/SKILL.md` + `SOUL.md` por profile | gere profiles + instale as skills        |

> O **`team-manager`** é o único agente que sabe gerar todos esses
> artefatos. Os demais apenas consomem o que está materializado.

---

## 1. Source of truth

**GitHub Issues + `harness/*` são a fonte da verdade.** Nada é decidido
em chat que não vire issue ou commit. As personas documentam **tudo**
no issue correspondente (comentários de status, decisões, blockers).

---

## 2. O time (personas)

| Persona              | Quando atua                                                | O que entrega                                                                                          |
|----------------------|------------------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| **team-manager**     | Em **toda** transição de estado da issue.                  | Sub-issues, labels, assignees, branches, comments de status, merge, tag, close.                        |
| **domain-expert-`<domínio>`** | Após issue criada, antes do DoD. **Sempre especializado** (ex.: `domain-expert-banking`). | História refinada + critérios de aceite + DOR (Definition of Ready). Pode haver 1+ specialists por projeto. |
| **solutions-architect** | Após DOR, antes da implementação.                       | DoD técnico + checklist 12-factor + decisões arquiteturais (ADR-lite).                                  |
| **backend-engineer** | Quando a issue é `ready` + atribuição `backend`.           | Código Go + testes + Dockerfile + migration + commit na branch da feature.                              |
| **frontend-engineer**| Quando a issue é `ready` + atribuição `frontend`.          | Código Nuxt + testes + Dockerfile + commit na branch da feature.                                       |
| **quality-assurance**| Quando a branch está pronta (label `in-review`).           | Relatório de sensores + smoke/load + aprovação ou devolução.                                            |
| **devops-engineer**  | Quando QA aprova (label `qa`).                             | Validação de pipeline + (se skill existir) deploy + release + tag.                                      |

---

## 3. Routing rules (qual persona age em qual momento)

```yaml
issue_created:
  -> team-manager: aplica label `triage` + label de TIPO (`type/<x>`)
  # Tipos: feature | technical | infra | bug | tech-debt | docs | spike
  # Ver team-manager.md §4 e workflow/00-issue-lifecycle.md §0
  # Detecta domínio (`domain/<x>`) se type/feature ou type/bug de negócio

triage_done:
  # ==== CAMINHO FEATURE / BUG DE NEGÓCIO ====
  if type in [feature, bug] and domain/<x>:
    -> domain-expert-<domain>: refina história + critérios de aceite
    # domain-expert posta comentário com: história, AC, edge cases, dependências
    # e.g.: domain-expert-banking, domain-expert-retail, domain-expert-mandai

  # ==== CAMINHO TÉCNICO / INFRA / TECH-DEBT ====
  if type in [technical, infra, tech-debt]:
    # PULA domain-expert; vai direto para solutions-architect
    -> (skip refined, go to ready)

ready_for_tech:
  -> solutions-architect: define DoD + checklist 12-factor
  # solutions-architect posta: DoD checklist, decisões, riscos
  # Valida se a estrutura segue o meta-harness (mesmo para type/infra)

tech_approved:
  -> team-manager: cria branch `feature/<id>-<slug>`, atribui, label `in-progress`

in_progress:
  if type/infra:
    -> devops-engineer: executa (pipeline, workflow, deploy)
  else:
    -> backend-engineer: se issue toca backend
    -> frontend-engineer: se issue toca frontend
    # ambos podem trabalhar em paralelo na MESMA branch
    # ambos escrevem testes primeiro (TDD)

builders_done:
  -> quality-assurance: roda todos os sensores + sobe snapshot + smoke/load
  # QA posta relatório e: aprova (label `qa`) ou devolve (label `in-review` + bugs)

qa_approved:
  -> team-manager: pede validação ao usuário (snapshot URL no PR)
  # espera comentário "validado" do usuário
  # IMPORTANTE: team-manager ACOMPANHA ATÉ O FIM (não larga após delegar)

user_validated:
  -> devops-engineer: valida pipeline, dispara release
  -> team-manager: merge, tag, fecha issue
```

### Smart routing — qual persona entra?

> Detalhado em [`team-manager.md` §4](../personas/team-manager.md)
> e [`workflow/00-issue-lifecycle.md` §0](./00-issue-lifecycle.md).

| Tipo             | domain-expert | solutions-architect | builder | devops | qa |
|------------------|---------------|---------------------|---------|--------|----|
| `type/feature`   | ✅ sim        | ✅ sim              | ✅ sim  | ✅ sim | ✅ sim |
| `type/technical` | ❌ não        | ✅ sim              | ✅ sim  | ✅ sim | ✅ sim |
| `type/infra`     | ❌ não        | ✅ sim              | ❌ não  | ✅ sim | ✅ sim |
| `type/bug`       | ⚠️ depende    | ✅ sim              | ✅ sim  | ✅ sim | ✅ sim |
| `type/tech-debt` | ❌ não        | ✅ sim              | ✅ sim  | ✅ sim | ✅ sim |
| `type/docs`      | ❌ não        | ⚠️ revisão         | ❌ não  | ❌ não | ❌ não |
| `type/spike`     | ⚠️ depende    | ⚠️ depende         | ❌ não  | ❌ não | ❌ não |

> **Regra de ouro:** o `team-manager` decide o caminho baseado no
> **tipo** + **domínio** da issue. **Não** é "sempre passa por
> todos" — é "só passa por quem agrega valor".

---

## 4. Labels canônicas (criar no repo no bootstrap)

> **Sem prefixo** para reduzir ruído. Cores agrupam a categoria.

| Label                  | Cor     | Significado                                            |
|------------------------|---------|--------------------------------------------------------|
| `triage`               | `#cccccc` | Issue nova, ainda não avaliada.                        |
| `needs-info`           | `#fbca04` | Faltam informações do autor da issue.                  |
| `refined`              | `#0e8a16` | `domain-expert-<domínio>` refinou a história.          |
| `domain/<nome>`        | `#fef2c0` | Domínio da issue (ex.: `domain/banking`, `domain/retail`, `domain/mandai`). Usado pelo team-manager para rotear ao specialist correto. |
| `type/feature`         | `#7057ff` | Feature de negócio. Entra `domain-expert-<x>` no fluxo. |
| `type/technical`        | `#5319e7` | Setup técnico puro. **Pula** `domain-expert`. Vai direto para `solutions-architect`. |
| `type/infra`            | `#5319e7` | Infraestrutura. **Pula** `domain-expert` e builder. Direto para `solutions-architect` → `devops-engineer`. |
| `type/bug`              | `#b60205` | Bug. `domain-expert` entra se for bug de negócio. |
| `type/tech-debt`        | `#fbca04` | Dívida técnica. **Pula** `domain-expert`. |
| `type/docs`             | `#0075ca` | Documentação. Sem DoD formal. |
| `type/spike`            | `#cccccc` | Investigação/Pesquisa. Saída: ADR. Sem código de produção. |
| `ready`                | `#0e8a16` | `solutions-architect` definiu DoD.                      |
| `in-progress`          | `#1d76db` | Builder está implementando.                            |
| `in-review`            | `#1d76db` | Builder terminou, QA rodando.                          |
| `qa`                   | `#5319e7` | QA aprovou; aguardando validação do usuário.            |
| `done`                 | `#0e8a16` | Mergeado + release feito.                              |
| `blocked`              | `#b60205` | Bloqueado por dependência externa.                     |
| `wontfix`              | `#ffffff` | Não será feito.                                        |
| `duplicate`            | `#cccccc` | Duplicado de outra issue.                              |
| `backend`              | `#bfd4f2` | Componente backend.                                    |
| `frontend`             | `#bfd4f2` | Componente frontend.                                   |
| `infra`                | `#bfd4f2` | Componente infra/devops.                               |
| `breaking-change`      | `#b60205` | Mudança incompatível.                                  |
| `tech-debt`            | `#fbca04` | Dívida técnica (não é bug).                            |
| `security`             | `#b60205` | Issue de segurança.                                    |
| `documentation`        | `#0075ca` | Mudança/adição de docs.                                |

---

## 5. Convenções de branches e commits

- **Branch:** `feature/<issue-id>-<slug-em-kebab-case>` (ou `fix/`,
  `chore/`, `release/vX.Y.Z`).
- **Commits:** **Conventional Commits** (`feat:`, `fix:`, `chore:`,
  `docs:`, `refactor:`, `test:`, `ci:`). Referência à issue no rodapé:
  `Refs #42` ou `Closes #42` quando fechar.

Exemplo:

```
feat(auth): implementa login com JWT (Refs #42)
```

---

## 6. PR template (mínimo obrigatório)

```markdown
## Summary
(1 parágrafo do que foi feito)

## Issue
Closes #<id>

## Changes
- [ ] ...

## Sensors (todos verdes)
- [ ] `make lint` — OK
- [ ] `make test` — coverage ≥ 80%
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
(descrever)
```

---

## 7. Comandos canônicos (Makefile mínimo esperado)

Todo microsserviço Go expõe estes `make` targets (ver
[`templates/`](./templates/) para o Makefile completo):

```bash
make tidy        # go mod tidy
make build       # go build ./...
make test        # go test -race -coverprofile=coverage.out ./...
make lint        # golangci-lint run --timeout=5m
make vuln        # govulncheck ./...
make oas         # oapi-codegen (regenera internal/api/openapi.gen.go)
make migrate-up  # aplica migrations
make run         # roda o serviço local
make docker      # builda a imagem
make compose-up  # sobe docker-compose do deploy/
make compose-down
```

---

## 8. Invariantes do meta-harness (não-violáveis)

1. **Toda issue** tem 1+ commits que referenciam o número
   (`Refs #<id>` ou `Closes #<id>`).
2. **Todo PR** cita a issue que fecha.
3. **Todo PR** tem o bloco "Como testar localmente" preenchido.
4. **Todo microsserviço** expõe `/healthz`, `/readyz` e `/metrics`.
5. **Todo microsserviço** loga em JSON via `slog`.
6. **Nenhum microsserviço** lê config de arquivo. Só env.
7. **Nenhum microsserviço** roda como root no container.
8. **Nenhum microsserviço** entra em produção sem `govulncheck` verde.
9. **Nenhum PR** é mergeado sem coverage ≥ 80% nos pacotes alterados.
10. **Nenhuma issue** é fechada sem validação do usuário.
11. **Nenhuma string de usuário é hardcoded** — toda mensagem
    externalizada (erro de API, copy de UI, e-mail, notificação) usa
    i18n. Idiomas obrigatórios: **en, pt-BR, es**. O sensor
    `08-i18n-audit` valida paridade de chaves e ausência de hardcode
    em todo PR.
12. **Toda issue é roteada ao `domain-expert-<domínio>` correto** —
    nunca existe um `domain-expert` genérico. A label
    `domain/<nome>` é obrigatória na triagem (quando aplicável).
13. **Toda issue tem label de tipo** (`type/feature`,
    `type/technical`, `type/infra`, `type/bug`, `type/tech-debt`,
    `type/docs`, `type/spike`) na triagem. Define quem entra no
    fluxo.
14. **Issue-mãe só fecha quando todas as sub-issues estão `done`** e
    o PR foi mergeado + validado pelo usuário. O `team-manager`
    **acompanha cada sub-issue até a conclusão** (não larga após
    delegar).
15. **Branches de feature/fix/chore são criadas pelo team-manager
    e delegadas no briefing.** Quem implementa
    (`backend-engineer`/`frontend-engineer`) **recebe o nome da
    branch** no briefing e só clona. O team-manager **NÃO escreve
    código de feature** — essa é a única linha vermelha
    (orquestração inclui criar branch; engenharia é o que está
    dentro dela). Personas **não-técnicas** (`domain-expert`,
    `solutions-architect`, `quality-assurance`) **nunca** mencionam
    nome de branch nem dizem a quem atribuir. Ver
    [`personas/interactions.md`](./personas/interactions.md) e
    ADR-0006.
16. **Nenhum PR é aberto com CI local vermelho.** Builders rodam
    `make lint && make test && make vuln` (Go) ou
    `pnpm lint && pnpm typecheck && pnpm test:run && pnpm audit`
    (Node) **antes** de `gh pr create`. QA devolve IMEDIATAMENTE
    se o PR chegar com checks vermelhos. Team-manager **NÃO** pede
    validação do user com CI vermelho. Ver ADR-0008.
17. **1 Dockerfile por service, em path canônico.** Cada service do
    monorepo tem **exatamente 1** Dockerfile em path canônico
    documentado. Proibido:
    - `Dockerfile` na raiz (mover para `deploy/Dockerfile.backend`
      ou path específico do service).
    - 2+ Dockerfiles pro mesmo service (ex.: `backend/Dockerfile`
      E `deploy/Dockerfile.backend`).
    Paths canônicos por padrão:
    - **Backend Go:** `deploy/Dockerfile.backend`
    - **Frontend Node:** `web/Dockerfile`
    - **Migrate (12-factor XII):** usa imagem oficial
      `migrate/migrate:v4.19.1` no compose — **NÃO** custom build
      (gotcha #2 do `versions.md`).
    O `check-stack-versions.sh` detecta divergência. Ver
    ADR-0011.
18. **CI modular com path filters.** O workflow `.github/workflows/
    ci.yml` DEVE ter:
    - 1 job `changes` no topo com `dorny/paths-filter@v3.0.2`
      (SHA-pinned em prod) que computa 6+ outputs
      (`backend`, `frontend`, `infra`, `docs`, `workflow`,
      `contracts`).
    - Todos os outros jobs com `needs: changes` + `if: needs.
      changes.outputs.<X> == 'true'`. **Proibido** rodar lint
      de Go quando o PR só muda `web/`.
    - `concurrency` com `cancel-in-progress: ${{ github.event_name
      == 'pull_request' }}` (cancela rodadas obsoletas em PRs;
      nunca em main).
    - Cache Docker com `scope=<service>` (ex.: `scope=backend`,
      `scope=frontend`) — caches separados por service.
    - Trivy SHA-pinado (`@0.36.0` ou SHA completo) — NUNCA
      `@master` ou `@latest` (supply-chain risk comprovado em
      mar/2026).
    - `GOTOOLCHAIN= local` em todos os jobs Go (impede `go mod
      tidy` de reescrever `go.mod` no CI).
    - 12-Factor audit roda **sempre** (gate de segurança não
      pode ser pulado por path filter).
    Ver `templates/.github-workflows-ci.yml` e ADR-0011.
19. **Team-manager verifica, não confia.** Após um builder reportar
    "PRONTO" / "VERDE", o `team-manager` **re-executa** os
    checks críticos (re-lê `go.mod`/`Dockerfile`/`ci.yml`,
    roda `make lint && make test && make vuln`) **antes** de
    rotular como `in-review` ou pedir validação humana.
    Lição do Mandaí v2 (jul/2026, ADR-0014): um builder reportou
    `go.mod` com `go 1.22.0` quando o arquivo continha `go 1.25.0`
    — incoerência só foi pega pelo humano que leu o arquivo
    diretamente. **Auto-relato de subagente é evidência fraca.**
    Sensor [`09-verify-after-build`](./sensors/09-verify-after-build.md)
    codifica o protocolo. Ver §11 do
    [`personas/team-manager.md`](./personas/team-manager.md).

---

## 9. Como cada tool consome este AGENTS.md

### Claude Code

```bash
# O team-manager gera:
cp harness/AGENTS.md CLAUDE.md
mkdir -p .claude/agents .claude/skills .claude/commands
# Para cada persona em harness/personas/*.md, gera .claude/agents/<name>.md
# Para cada sensor em harness/sensors/*.md, gera .claude/skills/<name>/SKILL.md
```

### GitHub Copilot

```bash
# O team-manager gera:
mkdir -p .github/agents
# Gera .github/copilot-instructions.md a partir deste AGENTS.md
# Para cada persona, gera .github/agents/<name>.md
```

### Codex CLI / OpenCode

```bash
# Já funciona: AGENTS.md na raiz é o contrato
# Persona files podem ser copiados para .codex/agents/ ou .opencode/agents/
```

### Hermes Agent

```bash
# O team-manager gera profiles (um por persona) + skills:
hermes profile create team-manager --description "Orquestrador do meta-harness"
hermes profile create backend-engineer --description "..."
# ...
# Cada persona vira um ~/.hermes/skills/<name>/SKILL.md
hermes skills install <path-para-harness/skills/<name>>
```

### Devin CLI / Cursor

```bash
# Devin: AGENTS.md + .devin/ configurado pelo time-manager
# Cursor: .cursorrules gerado a partir deste arquivo
```

---

## 10. Como estender o meta-harness

- **Nova persona:** crie `harness/personas/<name>.md` (use
  `team-manager.md` como template), adicione à lista em
  `bootstrap.md` §4, e gere os artefatos nativos do tool.
- **Novo sensor:** crie `harness/sensors/<id>-<name>.md` com comando
  exato, exit code, thresholds e onde pluga no workflow. Adicione ao
  CI workflow.
- **Nova stack:** adapte `harness/stack/*.md` e `templates/*`. Não
  remova os princípios (§2 do `bootstrap.md`).
- **Nova regra:** adicione à §8 deste arquivo. Se for princípio
  fundamental, promova para `bootstrap.md` §2 via ADR.

---

> Este arquivo é **vivo**: o `team-manager` é responsável por mantê-lo
> sincronizado com `bootstrap.md` e os artefatos nativos do tool em uso.
