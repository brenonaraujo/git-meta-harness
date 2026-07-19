# Contrib — Design Decisions (ADRs)

> Registro de decisões arquiteturais (ADR-lite) do meta-harness.
> Crie entradas aqui sempre que uma decisão **mudar o spec** (não só
> o código).

---

## Formato

```markdown
## ADR-XXXX — <título>

**Data:** YYYY-MM-DD
**Status:** Proposto | Aceito | Substituído | Deprecado
**Decisor(es):** <pessoas ou personas>
**Contexto:** <projeto / issue>

### Contexto
(qual problema está sendo resolvido?)

### Decisão
(o que decidimos?)

### Alternativas consideradas
- **A:** ... — prós / contras
- **B (escolhida):** ... — prós / contras
- **C:** ... — prós / contras

### Consequências
(o que muda? o que fica mais fácil / mais difícil?)

### Reversibilidade
(como reverter se for um erro?)
```

---

## Decisões registradas

### ADR-0001 — Adotar meta-harness com 7 personas e 8 sensors

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** criar framework de orquestração para entrega greenfield
→ produção.

### Contexto

- Necessidade de um framework declarativo que orce um time de
  agentes de IA para entregar projetos do zero.
- Múltiplas ferramentas de IA no mercado (Claude Code, Copilot,
  Codex, Hermes, OpenCode, Devin), cada uma com seu layout nativo.
- Falta de um padrão único de stack (Go + Gin + GORM + Nuxt) e
  código (KISS, DRY, ≤25/≤150, twelve-factor).

### Decisão

- Definir **7 personas** (team-manager, domain-expert,
  solutions-architect, backend-engineer, frontend-engineer,
  quality-assurance, devops-engineer).
- Definir **8 sensors** (static-analysis, vulnerability, unit,
  contract, image-scan, smoke, load, twelve-factor-audit).
- Stack única: Go 1.22 + Gin + GORM + PostgreSQL + OpenAPI
  (backend); Nuxt 3/4 + Nuxt UI + Pinia (frontend).
- Limites rígidos: ≤ 25 linhas/função, ≤ 150 linhas/arquivo,
  coverage ≥ 80%, 12 fatores.
- Multi-tool via `AGENTS.md` (Claude, Copilot, Codex, Hermes,
  OpenCode, Devin, Cursor).

### Alternativas consideradas

- **A:** Limitar a 1 tool (Claude Code) — simples, mas reduz
  adoção.
- **B:** Definir 3 personas (analyst, builder, qa) — pouco
  granular.
- **C (escolhida):** 7 personas + 8 sensors, multi-tool via
  AGENTS.md — mais complexo, mas flexível e auditável.

### Consequências

- **+** Cobertura completa do fluxo (refinamento → DoD →
  implementação → QA → release).
- **+** Sensores automatizam a maior parte da qualidade.
- **+** Multi-tool = funciona com o que o time já usa.
- **−** Curva de aprendizado (7 personas + 8 sensors).
- **−** Overhead inicial (criar labels, workflows, profiles).
- **−** Manutenção dos artefatos por tool.

### Reversibilidade

- Personas podem ser fundidas (ex.: domain-expert +
  solutions-architect em 1) sem quebrar o spec.
- Sensors podem ser removidos do CI individualmente.
- Stack pode ser estendida (adicionar lib) sem mudar o spec.

---

## Próximas ADRs a criar

- ADR-0003 — Escolha entre `oapi-codegen` vs `ogen` para geração
  de código.
- ADR-0004 — Estratégia de release (release-please vs manual).
- ADR-0005 — Padrão de autenticação (JWT vs session vs OAuth2).
- ADR-0006 — Provider de observability (Prometheus + Grafana
  self-hosted vs Datadog vs Honeycomb).
- ADR-0007 — Estratégia de cache (Redis vs in-memory vs nada).
- ADR-0008 — Estratégia de mensageria (RabbitMQ vs Kafka vs
  NATS vs Postgres outbox).

---

### ADR-0002 — i18n obrigatório em en, pt-BR, es

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** projetos do meta-harness precisam atender
usuários em múltiplos idiomas desde o MVP.

### Contexto

- O meta-harness é agnóstico de domínio, mas foi desenhado
  inicialmente para um time brasileiro com clientes em
  LATAM/América Latina.
- Strings hardcoded viram dívida técnica cara: refatoração tardia
  obriga a revisar 100% do código.
- 12-factor IX (disposability) e clean code pedem que strings
  vivam em arquivos de configuração, não em código.

### Decisão

- Adotar **i18n como princípio 11 do `bootstrap.md`** (inegociável).
- **Idiomas obrigatórios:** `en` (English), `pt-BR` (Português
  brasileiro), `es` (Español neutro).
- **Backend:** `github.com/nicksnyder/go-i18n/v2` com bundles
  JSON em `internal/i18n/locales/{en,pt-BR,es}.json`.
- **Frontend:** `@nuxtjs/i18n` v8+ com bundles em
  `i18n/locales/{en,pt-BR,es}.json`.
- **Seleção de idioma (API):** header `Accept-Language`
  (RFC 7231), com fallback `DEFAULT_LOCALE` env (default `en`).
- **Seleção de idioma (frontend):** detecção automática do browser
  + seletor manual, cookie persistente.
- **Sensor novo:** `sensors/08-i18n-audit.md` valida paridade de
  chaves e ausência de hardcode em todo PR.
- **Invariante nova:** `AGENTS.md` §8.11 — "nenhuma string de
  usuário é hardcoded".

### Alternativas consideradas

- **A:** Não ter i18n (adicionar depois) — simples no MVP, mas
  refatoração cara depois.
- **B:** Usar apenas `en` no MVP e i18n só quando precisar —
  protela o problema.
- **C (escolhida):** i18n desde o dia 1 com 3 idiomas fixos —
  mais trabalho inicial, mas elimina dívida técnica e
  permite i18n como **feature**, não como hotfix.

### Consequências

- **+** Strings externalizadas desde o MVP; i18n é **grátis** depois.
- **+** Time brasileiro e time LATAM podem contribuir traduções
  sem tocar em código.
- **+** Paridade de chaves garante que nenhum idioma fica
  quebrado.
- **−** Work extra para builders: cada mensagem precisa de 3
  traduções.
- **−** Curva de aprendizado (biblioteca nova para muitos).
- **−** Refatoração de código legado (se houver) precisa de
  varredura.

### Reversibilidade

- Idioma adicional (ex.: `fr`) = só adicionar `fr.json` e
  atualizar `nuxt.config.ts` e `internal/i18n/bundle.go`.
- Trocar biblioteca i18n (ex.: para `go-i18n/v3` quando sair) =
  refatorar só `internal/i18n/`; uso do `i18n.T()` se mantém.
- Idioma removido = deletar arquivo e remover do config.
  Strings em código que referenciam aquela chave vão para fallback
  (chave crua); sensor `08-i18n-audit` detecta.

---

### ADR-0003 — `domain-expert` é sempre especializado (`domain-expert-<domínio>`)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** um domain-expert genérico não entrega o mesmo valor
que um especialista do domínio.

### Contexto

- Domain-experts são o ponto de entrada do conhecimento de negócio
  no fluxo do meta-harness.
- Um agent genérico ("domain-expert") que cobre todos os domínios
  dilui o conhecimento e gera refinamentos superficiais.
- Regulamentação (BACEN, ANVISA, CDC) e padrões (Pix, Open Finance,
  OMS) mudam por domínio; cada um exige conhecimento profundo.
- Projetos podem atravessar múltiplos domínios (e-commerce
  precisa de banking + retail + logistics).

### Decisão

- `domain-expert` é **sempre especializado**, com sufixo de
  domínio: `domain-expert-<domínio>`.
- O nome da persona = nome do domínio em kebab-case
  (ex.: `domain-expert-banking`, `domain-expert-retail`,
  `domain-expert-mandai`).
- **Não existe** `domain-expert` genérico; esse agente nunca é
  instanciado.
- Cada projeto pode ter **1+ domain-experts** simultâneos
  (ex.: `domain-expert-banking` + `domain-expert-logistics` num
  e-commerce com entrega).
- O **`team-manager` roteia por label `domain/<x>`**:
  - 1ª opção: label `domain/<x>` na issue.
  - 2ª opção: análise do título/body.
  - 3ª opção: pergunta explícita ao autor.
- Issues que atravessam múltiplos domínios viram sub-issues, cada
  uma atribuída ao specialist daquele domínio.
- Criamos o template em `personas/domain-expert.template.md` e 3
  exemplos prontos em `personas/examples/`:
  - `domain-expert-banking` (fintech)
  - `domain-expert-retail` (e-commerce)
  - `domain-expert-mandai` (placeholder editável)

### Alternativas consideradas

- **A:** `domain-expert` genérico + skill por domínio — simples,
  mas a skill não substitui conhecimento profundo (regulação,
  edge cases de mercado, padrões).
- **B:** `domain-expert` como orquestrador de experts externos —
  adiciona camada, mas o próprio agent já é o expert.
- **C (escolhida):** sempre `domain-expert-<domínio>` —
  especializado, roteamento explícito, exemplos prontos.

### Consequências

- **+** Conhecimento profundo por domínio (refinamentos melhores).
- **+** Compliance e regulação ficam first-class no refinamento.
- **+** Roteamento explícito (label `domain/<x>`) — fácil de
  entender e auditar.
- **+** Onboarding de novo domínio = copiar template e preencher.
- **−** Mais personas para criar (1+ por projeto).
- **−** Multi-domínio precisa de sub-issues (overhead).
- **−** Domain-expert precisa ser explicitamente criado antes
  do projeto começar (não dá pra "improvisar").

### Reversibilidade

- Adicionar novo domínio = copiar `domain-expert.template.md` para
  `domain-expert-<novo>.md` + preencher; criar label
  `domain/<novo>` no repo.
- Renomear domínio (ex.: `domain-expert-mandai` →
  `domain-expert-<novo-nome>`) = renomear arquivo, atualizar
  materialização do tool, atualizar referências.
- Remover domínio = deletar arquivo + deletar label; issues
  abertas precisam ser re-rotuladas manualmente.

---

### ADR-0004 — `team-manager` é orquestrador ponta-a-ponta com smart routing

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** o team-manager estava virando "decompositor" —
delegava e largava. E o fluxo único "todos passam por todos" estava
gerando overhead em issues técnicas.

### Contexto

- Em testes com o meta-harness, o `team-manager` decompunha a
  issue-mãe em sub-issues, atribuía a personas, e **parava de
  acompanhar** — os builders trabalhavam, mas ninguém fechava o
  loop. Resultado: issues "zumbis", work parado, validação do
  usuário perdida.
- O fluxo canônico assumia que **toda** issue passava por **todas**
  as 6 personas. Mas uma issue de bootstrap técnico
  (ex.: "configurar o pipeline de release") **não tem valor de
  negócio** para o `domain-expert` refinar; é puramente
  infra/arquitetura. Forçar `domain-expert` a entrar gera ruído
  e dilui o valor das personas.
- No tool **Hermes Agent**, a criação de profiles é centralizada
  no team-manager. Detectamos que o team-manager estava passando
  `--model` ao criar profiles de sub-personas, sobrescrevendo o
  default que o usuário já tinha configurado — quebra de
  expectativa.

### Decisão

- **O `team-manager` é o único dono do ciclo de vida da issue** —
  da triagem até o `done`. **Não** delega e larga; acompanha cada
  sub-issue, cutuca builders parados, valida o trabalho, e fecha
  a issue-mãe.
- **Delegação é explícita**: o team-manager posta **briefings
  humano-legíveis** nas issues (não apenas `gh issue edit
  --add-assignee`), especificando o que se espera e qual é o
  próximo passo.
- **Smart routing**: o team-manager **decide quais personas
  entram** com base no **tipo** da issue (label `type/<x>`):
  - `type/feature` → todas as personas.
  - `type/technical` → pula `domain-expert` (vai direto para
    `solutions-architect`).
  - `type/infra` → pula `domain-expert` e builder (vai para
    `solutions-architect` → `devops-engineer`).
  - `type/tech-debt` → pula `domain-expert`.
  - `type/docs` → quase tudo pulado (revisão editorial).
  - `type/spike` → só research; saída é ADR.
- **Issue-mãe só fecha quando TODAS as sub-issues estão `done`**
  + PR mergeado + validado pelo usuário (invariante 14).
- **Hermes profiles herdam o modelo default**: o `team-manager`
  **NÃO passa `--model`** ao criar profiles. Todos herdam o
  `config.yaml` do Hermes. Apenas sobrescreve se houver
  requisito técnico explícito (com justificativa registrada).
- Cada profile tem seu próprio `SOUL.md` gerado a partir do
  arquivo de persona, suas próprias skills, e state isolado.
- Adicionamos 7 labels de tipo: `type/feature`, `type/technical`,
  `type/infra`, `type/bug`, `type/tech-debt`, `type/docs`,
  `type/spike`.

### Alternativas consideradas

- **A:** Team-manager como "decompositor" (status quo) — falha
  por não acompanhar; issues zumbis.
- **B:** Fluxo único "todos passam por todos" — overhead em
  issues técnicas; dilui o valor do `domain-expert`.
- **C (escolhida):** Team-manager orquestrador ponta-a-ponta
  com smart routing por tipo — fluxo enxuto, sem perda de
  rastreamento.

### Consequências

- **+** Issues zumbis eliminadas (team-manager acompanha até o
  fim).
- **+** Overhead reduzido em issues técnicas (não força
  `domain-expert`).
- **+** Briefing explícito torna o trabalho distribuído mais
  auditável.
- **+** Profiles do Hermes herdam configuração do usuário (não
  quebra expectativa).
- **−** Team-manager tem mais responsabilidade (acompanhar
  proativamente, cutucar).
- **−** Workflows condicionais (mais regras, mais cognição).
- **−** Necessidade de cutucar builders manualmente (até
  automatizar).

### Reversibilidade

- Tirar smart routing (voltar ao fluxo único) = remover
  `type/<x>` labels e ajustar workflow/00-issue-lifecycle.md.
- Adicionar novo tipo = criar label `type/<novo>` e definir
  caminho em §0 do workflow/00.
- Trocar Hermes profile config = ajustar `SOUL.md` por profile;
  não afeta personas em si.

---

### ADR-0005 — Quem cria branches e quem atribui (separação de papéis)

**Data:** 2026-07-18
**Status:** **Superseded por ADR-0006**
**Decisor(es):** time de plataforma

> ⚠️ **Este ADR foi substituído pelo ADR-0006.** A decisão original
> era "builders criam branch". A nova decisão é "team-manager cria
> e delega; builders só clonam". Mantido aqui apenas para histórico.

### Contexto (original)

Em teste real do meta-harness com a primeira issue
(`#1-bootstrap-skeleton`), o `solutions-architect` postou:
> "Atribuir a frontend-engineer (label ready → in-progress após
> o team-manager criar a branch feature/1-bootstrap-skeleton)."

A ADR-0005 original propôs que **builders** criassem a branch
(individualmente, com o "primeiro cria, segundo puxa"). Ver
ADR-0006 para a decisão final.

### Reversibilidade (do supersede)

Reverter o ADR-0006 = voltar a este ADR-0005 (builders criam
branch).

---

### ADR-0006 — `team-manager` cria branch e delega; builder implementa (decisão final)

**Data:** 2026-07-18
**Status:** Aceito (supersede ADR-0005)
**Decisor(es):** time de plataforma
**Contexto:** a ADR-0005 original propôs que **builders** criassem
a branch. Em reanálise, percebemos que o `team-manager` é quem
precisa garantir uma branch única para múltiplos builders na
mesma issue.

### Contexto

- A ADR-0005 propôs que o **primeiro** builder a começar criasse a
  branch e o **segundo** puxasse. Na prática, isso:
  - Adiciona **dois pontos de falha** (cada builder precisa
    lembrar de criar/puxar a branch).
  - Cria **race condition** se o segundo builder chegar antes do
    primeiro e não souber que a branch está sendo criada.
  - **Quebra** a regra "1 issue = 1 branch" se o segundo builder
    criar a própria branch por engano.
- O `team-manager` tem **visão completa** de quem vai trabalhar
  na mesma issue (ex.: backend + frontend precisam da **mesma**
  branch). Centralizar a criação da branch **garante** que o nome
  é único e conhecido.
- O `team-manager` **não** precisa entender detalhes de
  implementação para criar uma branch — é trabalho de
  orquestração, não de engenharia.

### Decisão

- **Quem cria branches:**
  - `feature/<id>-<slug>`, `fix/<id>-<slug>`, `chore/<id>-<slug>`:
    **`team-manager`** (e delega no briefing).
  - `release/vX.Y.Z`: `devops-engineer` (apenas).
  - `hotfix/<id>-<slug>`: `devops-engineer` (em emergência).
- **Quem clona a branch:** `backend-engineer` ou
  `frontend-engineer` (recebe o nome no briefing).
- **Linha vermelha do `team-manager`:** ele **NÃO escreve código
  de feature**. Criar branch é orquestração (decide **onde** o
  trabalho vai acontecer); escrever código é engenharia. Esta é
  a **única** linha vermelha.
- **Quem atribui:** apenas `team-manager`. Personas especialistas
  (`domain-expert`, `solutions-architect`, `quality-assurance`)
  **NÃO** mencionam nomes de personas específicas nem dizem "atribuir
  a X" no output.
- Personas especialistas também **NÃO** mencionam nome de branch
  (a criação é exclusiva do team-manager).
- **Quem fala com quem:** documentado em
  [`personas/interactions.md`](../personas/interactions.md).
- **Invariante 15 do `AGENTS.md` §8:** "Branches de feature/fix/chore
  são criadas pelo `team-manager` e delegadas no briefing."

### Alternativas consideradas

- **A (ADR-0005 original):** Builders criam branch — falha por
  race condition e violação de "1 issue = 1 branch".
- **B:** Manter team-manager cria + builders clonam (status quo
  **antes** da ADR-0005) — simples, mas com confusão de papéis.
- **C (escolhida):** Team-manager cria e delega; builder clona
  e implementa; team-manager **NÃO** escreve código. Separação
  clara, com linha vermelha explícita.

### Consequências

- **+** Branch única garantida (1 issue = 1 branch, sempre).
- **+** Sem race condition entre builders.
- **+** Linha vermelha do team-manager é explícita: **NÃO**
  escreve código de feature.
- **+** Múltiplos builders recebem a mesma branch sem confusão.
- **−** Team-manager precisa fazer `git checkout -b` (mas é
  simples e isolado).
- **−** Builder precisa **confiar** que a branch foi criada (e
  cutucar se não).

### Reversibilidade

- Voltar para "builders criam branch" = reverter para ADR-0005
  (já documentado lá).
- Adicionar nova categoria de branch (ex.: `experiment/`) =
  estender a tabela de "Quem cria branches" no
  `workflow/01-branching.md`.

---

### ADR-0007 — Lessons learned do piloto Mandaí v2 (jul/2026)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** primeiro piloto real do meta-harness revelou 3 bugs
sutis que passaram batido na fase de design.

### Contexto

- O meta-harness foi instanciado no projeto-piloto **Mandaí v2**
  (marketplace B2B2C de compra coletiva comunitária).
- 5 commits + 4 issues + 1 PR foram produzidos. O `team-manager`
  orquestrou bem em vários aspectos (sub-issues, briefings, 1 PR
  único, etc.).
- **3 bugs sutis** passaram batido:
  1. **Smart routing não aplicado.** O `team-manager` roteou
     issues `type/technical` (backend, frontend) e `type/infra`
     (docker-compose) para o `domain-expert` — quando deveriam
     pular (smart routing só foi adicionado depois do piloto
     começar).
  2. **Domain-expert genérico usado.** O team-manager invocou
     `hermes -p domain-expert` (sem sufixo de domínio) em vez de
     `domain-expert-mandai`. O especialista usou o **template
     genérico** (não especializado em compra coletiva).
  3. **Versão antiga do meta-harness ficou no projeto.** O
     projeto-piloto ficou com 54 arquivos (versão antes das
     correções), então todas as melhorias de smart routing,
     `interactions.md`, e invariantes novos **não chegaram**
     até o projeto.

- Sem uma forma de **detectar esses bugs automaticamente**, eles
  só foram achados em análise manual pós-hoc.

### Decisão

- **Adicionar smoke test obrigatório** (`harness/scripts/smoke-test.sh`)
  que **valida 12 itens** antes do `team-manager` processar
  qualquer issue:
  1. Versão instalada (≥ 60 arquivos).
  2. Arquivos críticos presentes.
  3. Smart routing documentado (`type/*` em AGENTS.md/bootstrap).
  4. Interações matrix presente.
  5. ≥ 1 `domain-expert-<domínio>` (especializado).
  6. **CRÍTICO:** nenhum `domain-expert.md` genérico.
  7. ADR-0006 aplicado.
  8. ≥ 15 invariantes no AGENTS.md.
  9. 7 labels `type/*` no GitHub.
  10. **CRÍTICO:** nenhum profile `domain-expert` genérico no
      Hermes.
  11-12. (outros checks menores).
- **Falha bloqueia:** se o smoke test falha, o `team-manager`
  **NÃO** processa issues até corrigir.
- **Adicionar `VERSION`** na raiz do meta-harness para tracking
  de versão.
- **Atualizar `seed/meta-harness-seed.md`** com passo **0** que
  exige rodar o smoke test antes de tudo.
- **Adicionar pre-flight checklist** no
  `personas/team-manager.md` §"Quando você é acionado".

### Bugs detectados pelo smoke test (no Mandaí v2)

```
$ ./smoke-test.sh brenonaraujo/mandai-v2

1. Versão instalada (esperado: ≥ 60 arquivos)
  ❌ 54 arquivos (esperado ≥ 60)
     Fix: rsync -a meta-harness-m3-code/harness/ ./harness/

2. Arquivos críticos
  ❌ harness/smoke-test.md AUSENTE

3. Smart routing documentado
  ❌ AGENTS.md NÃO tem type/technical
  ❌ bootstrap.md NÃO tem type/infra

4. Interações matrix
  ❌ interactions.md AUSENTE

5. Domain-experts especializados
  ❌ Nenhum domain-expert-<domínio> (genérico proibido)

6. CRÍTICO — nenhum domain-expert genérico
  ❌ Bug: harness/personas/domain-expert.md (genérico) EXISTE

7. ADR-0006 aplicado
  ✅ AGENTS.md menciona team-manager cria branch

8. Invariantes
  ❌ 11 (esperado ≥ 15)

9. GitHub labels type/*
  ⚠️  type/feature AUSENTE
  ⚠️  type/technical AUSENTE
  ... (7/7 ausentes)

10. Hermes profiles sem genérico
  ❌ Profile 'domain-expert' (genérico) existe no Hermes

Passes: 13
Fails:  11
```

### Alternativas consideradas

- **A:** Documentar os bugs num changelog sem mudar o spec — falha
  por não prevenir reincidência.
- **B:** Confiar em code review manual dos PRs do bootstrap —
  falha por ser tarde demais (PR já está merged).
- **C (escolhida):** Smoke test automatizado no início de
  qualquer materialização, bloqueando o fluxo se falhar.

### Consequências

- **+** Bugs sutis detectados **antes** de processar issues.
- **+** Plug-and-play: o smoke test funciona em qualquer projeto
  que tenha o meta-harness materializado.
- **+** Documenta o "estado correto" do harness de forma
  executável (não só texto).
- **+** Reduz regressões em versões futuras (se alguém
  remover um arquivo crítico, o smoke test pega).
- **−** +1 arquivo (`smoke-test.md`) e +1 script (`smoke-test.sh`)
  no meta-harness.
- **−** O team-manager precisa rodar antes de processar issues
  (5 segundos a mais por projeto).

### Reversibilidade

- Remover o smoke test = deletar `smoke-test.md` e
  `scripts/smoke-test.sh`.
- Versão sem smoke test ainda funciona, mas **bugs do Mandaí v2**
  podem reincidir.

---

### ADR-0008 — Local pre-flight + CI workflow robusto (PR com 5/5 checks vermelhos)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** o PR do piloto Mandaí v2 foi pra review do user com
5/5 checks vermelhos — o team-manager pediu validação local sem
garantir que o código tinha sido validado, e o user ficou no escuro.

### Contexto

- O PR #5 do `mandai-v2` (16.569 adições, branch
  `feature/1-bootstrap-skeleton`) foi aberto com **TODOS os 5 checks
  principais vermelhos**:
  1. **Lint** — `.golangci.yml` mistura schema v1 (campos
     `issues`/`exclusions`/`settings` no top level) com
     `version: "2"`. Conflito direto.
  2. **Test + Coverage** — CI roda `go test ./...` na **raiz**, mas
     o `go.mod` está em `backend/`. CI não faz `cd backend`.
  3. **Vulnerability scan** — step anterior falhou,
     `govulncheck.sarif` não gerado.
  4. **OpenAPI contract** — `oasdiff/oasdiff-action@v1` é tag
     inválida (não existe).
  5. **12-Factor audit** — script `scripts/check-twelve-factor.sh`
     não existe. O `scripts/` está **vazio** (script está em
     `harness/scripts/`).
  6. **Build + Image scan** — pulado por dependência.

- O team-manager comentou "🤖 pronto, próximo é X" e **pediu
  validação do user** sem verificar que o CI tinha passado.
- Quando **rodei localmente** (`go build`, `go test`,
  `pnpm install`, `pnpm test:run`), **TUDO funcionou perfeitamente**
  (coverage 80.5%, 9/9 tests passing). O **código está bom** — o
  problema é 100% **configuração do CI**.

### Decisão

- **`templates/.golangci.yml` corrigido** — schema puramente v2.
  Sem campos `issues`/`exclusions`/`settings` no top level.
- **`templates/.github-workflows-ci.yml` corrigido** —
  - **`working-directory: backend`** em todos os jobs Go.
  - **`working-directory: web`** em todos os jobs Node.
  - **`oasdiff/oasdiff-action@v1.7.0`** (versão pinada e válida).
  - **Jobs separados para backend e frontend** (build, scan).
  - **Step de validação** que falha rápido se o script
    `check-twelve-factor.sh` não está em `scripts/`.
  - **Job `i18n`** adicionado (estava faltando).
  - **Job `summary`** que bloqueia merge se qualquer check falhou.
- **Invariante novo no `AGENTS.md` §8:** "Nenhum PR é aberto com
  CI local vermelho. Builders rodam `make lint && make test && make
  vuln` localmente ANTES de `gh pr create`. QA devolve imediatamente
  se o PR chegar com checks vermelhos."
- **Persona `backend-engineer` e `frontend-engineer` atualizadas:**
  item 11 explícito sobre "rodar localmente ANTES de abrir PR".
- **Persona `quality-assurance` atualizada:** item 3 novo
  "validar que o builder rodou checks locais antes de aceitar PR".
- **Mandaí v2 recebe 3 scripts de symlink** no CI: `scripts/
  check-twelve-factor.sh`, `scripts/check-i18n.sh`,
  `scripts/smoke-test.sh` apontando para `harness/scripts/`.
  (Ou: ajustar o CI para apontar direto para `harness/scripts/`.)

### Bugs detectados (output real do CI no Mandaí v2)

```
Lint ............................. FAILURE
  jsonschema: "run" does not validate with
  "/properties/run/additionalProperties":
  additional properties 'issues' not allowed

Test + Coverage ................. FAILURE
  pattern ./...: directory prefix . does not contain
  main module or its selected dependencies

Vulnerability scan .............. FAILURE
  Path does not exist: govulncheck.sarif

OpenAPI contract ................ FAILURE
  Unable to resolve action `oasdiff/oasdiff-action@v1`,
  unable to find version `v1`

12-Factor audit ................. FAILURE
  chmod: cannot access 'scripts/check-twelve-factor.sh':
  No such file or directory

Build + Image scan .............. SKIPPED
CI summary ...................... FAILURE
```

### Alternativas consideradas

- **A:** Confiar no CI e bloquear merge automático — falha
  porque o user ainda é pedido pra validar localmente.
- **B:** Adicionar um bot que comenta "CI vermelho" no PR —
  paliativo, não prevent.
- **C (escolhida):** Local pre-flight OBRIGATÓRIO pelo builder +
  workflow corrigido + invariante novo.

### Consequências

- **+** Builders não abrem PR com checks vermelhos.
- **+** QA não aceita PR com checks vermelhos.
- **+** Team-manager não pede validação do user com checks
  vermelhos.
- **+** Workflow correto para monorepo (working-directory).
- **+** Versões pinadas das actions (evita "version not found").
- **−** Builder precisa rodar localmente antes de PR (5min a
  mais, mas é onde os bugs são pegos).
- **−** Workflow mais complexo (mais jobs).

### Reversibilidade

- Tirar local pre-flight = remover item 11 das personas
  backend/frontend.
- Voltar workflow antigo = usar templates antigos (v1).
- Sair do monorepo = remover `working-directory` do template.

---

### ADR-0009 — Política de versões pinadas (versões latest estáveis)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** o piloto Mandaí v2 teve **bug em cascata de
versionamento** — `go.mod` declarava Go 1.26.5, mas o Dockerfile
usava `golang:1.22-alpine` (3 majors de diferença!). Builders
adotaram versões aleatórias (1.25, 1.26.5) sem fonte canônica.

### Contexto

- O Mandaí v2 teve 8 defeitos reais, dos quais **3 eram de
  versionamento**:
  - **D1:** `go.mod` declarava `go 1.26.5`, mas o `Dockerfile` usava
    `golang:1.22-alpine` → conflito.
  - **D5:** `.golangci.yml` schema v1 + linter v2 → lint quebrava.
  - **D6:** `govulncheck` achou 2 CVEs (quic-go v0.59.0, pgx v5.6.0) por
    deps desatualizadas.
  - **CI:** action `oasdiff/oasdiff-action@v1` (tag inválida) →
    CI quebrava.
- Builders **inventaram versões** (1.25, 1.26.5) sem
  **fonte canônica** dizendo "use X". Resultado: drift e bug.
- O spec não tinha **política de pinning** explícita; templates
  usavam `golang:1.22-alpine` (já desatualizado quando foi escrito).

### Decisão

- **Criar `harness/stack/versions.md`** — tabela canônica de
  versões pinadas para todos os componentes. **Única fonte da
  verdade**. Builders, QA, devops **DEVEM** referenciar.
- **Política de pinning:**
  1. **MAJOR version é fixo** (Go 1.26, não `latest`).
  2. **MINOR/PATCH é fixo** quando houver risco de regressão.
  3. **Imagens Docker:** tag semver em dev/CI, **digest SHA256 em
     produção** (recomendado).
  4. **Atualizar a tabela** quando uma major version nova
     estabilizar (≥ 3 meses no mercado).
  5. **Quebrar o pinning só via ADR** registrado em
     `contrib/design-decisions.md`.
- **Versões pinadas (jul/2026):**
  - **Go: 1.26.5** (latest stable 2026-07-07)
  - **Node.js: 24 LTS** (Krypton; 26 é "Current" mas não LTS)
  - **TypeScript: 5.x** (required by Pinia 3, Nuxt 4)
  - **Nuxt: 4.3.0** (latest stable)
  - **Nuxt UI: 3.3.6** (v3 line estável; v4 unificou com Pro)
  - **Pinia: 3.0.3** (requires Vue 3, TS 5+)
  - **@nuxtjs/i18n: 10.4.1** (v10 = Nuxt 4 support)
  - **GORM: 1.31.0** (latest)
  - **golang-migrate: v4.19.1** (latest)
  - **oapi-codegen: v2.5.0** (new path `oapi-codegen/oapi-codegen/v2`)
  - **testify: v1.11.1**
  - **golangci-lint: v2.4.0**
  - **Trivy CLI: v0.67.2** (pós-incidente de supply-chain de mar/2026)
  - **trivy-action: v0.32.0**
  - **PostgreSQL: 18.4-alpine**
  - **Distroless Go: `static-debian13:nonroot`** (UID 65532)
  - **Distroless Node: `base-debian12:nonroot`** (precisa libc)
- **Templates atualizados:**
  - `templates/Dockerfile.template` → `golang:1.26.5-alpine3.22`
    + `gcr.io/distroless/static-debian13:nonroot`.
  - `templates/docker-compose.template.yml` → `postgres:18.4-alpine`
    + healthchecks corrigidos (sem `CMD-SHELL` em distroless).
  - `templates/.github-workflows-ci.yml` → actions pinadas
    (`@v6`, `@v0.32.0`, `@v1.7.0`).
  - `templates/.golangci.yml` → schema puramente v2.
- **Renovate/Dependabot** recomendado para monitorar versões
  novas (não incluso por default; ver exemplo em `templates/`).

### Bugs detectados no Mandaí v2 (causados por falta de pinning)

| Bug | Causa |
|---|---|
| D1 — Go 1.22 vs 1.26.5 mismatch | `go.mod` declarava `go 1.26.5`, mas Dockerfile `golang:1.22-alpine` |
| D5 — `.golangci.yml` v1 + v2 mismatch | Linter instalado era v2.0.0, mas arquivo ainda em schema v1 |
| D6 — CVEs (quic-go, pgx) | Sem política de update; deps travadas em versões antigas |
| CI — `oasdiff-action@v1` | Tag inválida (não existe); versão não pinada |

### Alternativas consideradas

- **A:** Sem política — builders escolhem versão — falha por
  drift (status quo que causou o bug).
- **B:** Pinning automático via Renovate/Dependabot — bom para
  monitorar, mas não para o bootstrap inicial.
- **C (escolhida):** Tabela canônica `versions.md` + pinning
  manual + Renovate opcional para monitorar.

### Consequências

- **+** Builders têm fonte canônica de versões.
- **+** Templates não usam mais `latest` ou versões aleatórias.
- **+** Renovate (opcional) abre PRs automáticos quando há
  update.
- **+** Bug de versionamento (D1, D5) prevenido.
- **−** Atualizar `versions.md` quando uma major version nova
  estabiliza (custo de manutenção).
- **−** Renovate/Dependabot adiciona ruído se mal configurado.

### Reversibilidade

- Remover política = deletar `versions.md` e reverter
  templates para usar `latest`.
- Trocar versão = atualizar `versions.md` + templates + ADR
  (justificativa).

---

### ADR-0010 — Lições do versionamento real (Mandaí v2) e validação online de latest

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** mesmo com a ADR-0009 e o `check-stack-versions.sh`
v1 (consistência local), o piloto Mandaí v2 acumulou **8 defeitos
reais (D1-D8)** dos quais **4 eram de versionamento** que passaram
batido pelo validador. A validação por consistência local **não é
suficiente** — precisa também comparar contra as latest estáveis
online.

### Contexto

- Em jul/2026, o `team-manager` do Mandaí v2 **revogou waiver** e
  voltou issues para `in-progress` após o `@brenonaraujo` testar
  o snapshot e encontrar 8 defeitos reais. Detalhamento:
  - **D1:** `backend/go.mod` declarava `go 1.26.5` (depois
    `go 1.25.0` após `go mod tidy`), mas `deploy/Dockerfile.backend`
    usava `golang:1.22-alpine`. Build falhava com
    `go.mod requires go >= 1.25.0; running go 1.22.12;
    GOTOOLCHAIN=local`.
  - **D2:** Service `migrate` quebrava — binário custom compilado
    **sem** `_ "github.com/golang-migrate/migrate/v4/database/postgres"`.
  - **D3:** `migrate` CMD `"${DATABASE_URL}"` em **exec form** (sem
    shell) — variável **não expande**. Hardcoded fallback com
    credenciais placeholder apareceu e gerou ruído.
  - **D4:** `make test` falhava: coverage 47.8% < 80% porque
    `$(PKG)` no Makefile era `./...` (somava pacotes triviais
    com 0%).
  - **D5:** `.golangci.yml` mistura **schema v1 + v2** (campos
    `settings:` no top level, `exclusions:` dentro de
    `linters:`, `formatters:` ausente). Linter instalado era
    v2.0.0+ → quebrava.
  - **D6:** `govulncheck` achou 2 CVEs: `quic-go v0.59.0` (fix
    v0.59.1) e `pgx/v5 v5.6.0` (fix v5.9.2). Alcançáveis.
  - **D7:** Frontend healthcheck `CMD-SHELL: wget ...` em
    runtime distroless (sem shell/wget). Override no compose
    resolvia mas default do Dockerfile estava errado.
  - **D8:** CI vermelho (consequência de D1+D2+D3+D4+D5+D7).
- O `check-stack-versions.sh` v1 só validava **consistência
  local** (go.mod vs Dockerfile, etc.) — **não ia à fonte**.
  Resultado: deixou passar 4 bugs de versionamento (D1, D5, D6,
  D7) e não detectou que versões pinadas estavam **drift** das
  latest estáveis.
- 4 raízes distintas que a validação local não cobre:
  1. **Bootstrap Go requirement** (Go 1.26+ exige Go ≥ 1.24.6
     para compilar a si mesmo). Local: A é só comparação de
     números.
  2. **`.golangci.yml` schema v1+v2** (linters de formatters
     separados, exclusões em linters.exclusions, etc.). Local:
     comparação simples não pega schema mixing.
  3. **distroless `debianX` suffix** (jun/2026: tags sem
     sufixo viraram deprecated e apontam para debian13). Local:
     regex simples de `static-debian` perde `:nonroot` no
     sufixo errado.
  4. **Trivy supply-chain attack** (v0.69.4 comprometido
     19/mar/2026). Local: nenhuma validação de "esta versão é
     segura?".
- A tabela `versions.md` (ADR-0009) estava **sempre 3-6 meses
  atrás** das latest estáveis por construção (atualização
  manual). Sem alerta automático de drift, builders usavam
  versões antigas.

### Decisão

- **Adicionar modo `--check-latest`** ao
  `harness/scripts/check-stack-versions.sh`. Quando invocado,
  pesquisa **online** (GitHub API + Docker Hub API + Node dist
  index) e compara cada versão pinada com a latest estável.
  Alerta **drift** quando há diferença (warn) e alerta **versão
  inexistente/comprometida** (fail).
- **Expandir checks locais** (modo `--offline`, padrão) de 5
  para 9:
  1. ~~Go go.mod vs Dockerfile~~ (mantido)
  2. ~~Go go.mod vs CI workflow~~ (mantido)
  3. ~~Node package.json vs Dockerfile vs CI~~ (mantido)
  4. ~~Migrate: imagem oficial vs custom builder~~ (mantido)
  5. ~~Distroless: tag correta (static vs base)~~ (mantido)
  6. **NOVO: Go bootstrap requirement** (Go 1.26+ → image
     ≥ 1.24.6)
  7. **NOVO: `.golangci.yml` schema v2 puro** (detecta
     `settings:` no top level, `exclude-rules:` no top level,
     `gofmt`/`goimports` em `linters.enable` ao invés de
     `formatters.enable`)
  8. **NOVO: distroless SEM sufixo `-debianX`** (tag deprecated
     jun/2026)
  9. **NOVO: GitHub Actions NÃO pinadas** (`@latest`, `@main`,
     `@master` são fail)
  10. **NOVO: Trivy v0.69.4 detectado** (comprometido, fail
      crítico)
  11. **NOVO: Nuxt 3 detectado** (EOL 31/jul/2026, fail)
  12. **NOVO: Node 20 detectado** (EOL 30/abr/2026, fail) ou
      Node 26 (Current não-LTS até Out/2026, fail)
- **Atualizar `versions.md`** (jul/2026) com:
  - **Fontes/URLs** canônicas para cada componente (rastreabilidade).
  - **Última estável** explícita (data + link) para cada.
  - **Bootstrap requirement** do Go documentado na tabela.
  - **9 gotchas novos** adicionados (vs 3 antes): Go bootstrap,
    golangci-lint v2 schema, distroless `-debianX` suffix, Trivy
    supply-chain, Nuxt 3 EOL, Node 26 não-LTS, Node 20 EOL,
    Go 1.27 ainda em beta, etc.
  - **Seção "Como pesquisar a latest estável"** com comandos
    `curl`, `go list -m -versions`, `npm view`, `gh release
    list`, `docker manifest inspect`.
- **Bumpar versões pinadas** para latest estáveis jul/2026:
  - `golangci-lint`: v2.4.0 → **v2.12.2** (6/mai/2026)
  - `Trivy CLI`: v0.67.2 → **v0.72.0** (30/jun/2026)
  - `oapi-codegen`: v2.5.0 (mantido, já era latest em 15/jul/2026)
  - `distroless Node`: `base-debian12` → **base-debian13**
    (jun/2026, default mudou)
- **Adicionar `formatters:` section** ao template
  `.golangci.yml` (v2 separa linters de formatters).
- **Adicionar inv. 17** ao `AGENTS.md` §8: "Toda decisão de
  versão DEVE passar por `check-stack-versions.sh --check-latest`
  antes de virar pinada no `versions.md`."
- **Adicionar inv. 18**: "Nenhuma imagem Docker sem sufixo
  `-debianX` (distroless), sem tag semver de versão do SO
  (postgres, golang, node). Tags mutáveis (`@latest`,
  `@main`) em GitHub Actions são proibidas em produção."

### Bugs detectados pelo `check-stack-versions.sh --check-latest` (Mandaí v2)

```
$ ./harness/scripts/check-stack-versions.sh --check-latest

1. Go (go.mod vs Dockerfile)
  ✅ go.mod (backend/go.mod): go 1.25.0
  ✅ Dockerfile: golang:1.25-alpine

1b. Go bootstrap requirement
  ✅ Go 1.25 (não exige bootstrap 1.24.6+)

2. Go (go.mod vs .github/workflows/*.yml)
  ✅ CI: GO_VERSION=1.25

3. Node
  ✅ package.json: node 24, pnpm 10
  ✅ CI: NODE_VERSION=24
  ✅ Frontend Dockerfile: node:24-alpine

4. Migrate
  ✅ Nenhum custom migrate builder
  ✅ docker-compose usa imagem oficial

5. Distroless
  ✅ Tag com sufixo -debian13

6. .golangci.yml schema
  ✅ version: 2 declarado
  ✅ Sem settings: no top level
  ✅ Sem exclude-rules: no top level
  ✅ gofmt/goimports em formatters (v2)

7. GitHub Actions pinadas
  ✅ Todas pinadas

8. Trivy
  ✅ Nenhuma versão comprometida

9. Nuxt
  ✅ Nuxt 4 detectado

10. ONLINE — latest estáveis
  ✅ Go 1.26.5 = pinada
  ✅ Node.js LTS 24.18.0 = pinada
  ⚠️  golangci-lint: pinada v2.12.2 ≠ latest v2.13.0 (drift, revisar)
  ✅ Trivy CLI: pinada v0.72.0 = latest
  ✅ trivy-action: pinada v0.32.0 = latest
  ✅ oapi-codegen: pinada v2.5.0 = latest
  ✅ golang-migrate: pinada v4.19.1 = latest
  ✅ GORM: pinada v1.31.0 = latest
  ✅ Nuxt: pinada v4.3.0 = latest
  ✅ postgres: pinada 18.4-alpine = latest
  ✅ golang: pinada 1.26.5-alpine3.22 = latest
  ✅ node 24: pinada 24.18.0-alpine3.22 = latest
  ✅ distroless: pinada debian13 = latest
```

### Alternativas consideradas

- **A:** Confiar só em consistência local (status quo) —
  falha por não detectar drift, versões comprometidas, EOL
  iminente.
- **B:** Pinning automático via Renovate sem tabela canônica —
  Renovate é barulhento por default (PR toda semana); precisa
  de configuração pesada. Tabela canônica é mais explícita.
- **C (escolhida):** Tabela canônica (`versions.md`) +
  `check-stack-versions.sh --check-latest` (drift detection) +
  Renovate opcional (monitorar).

### Consequências

- **+** Detecta **drift** de versões pinadas vs latest estáveis
  automaticamente.
- **+** Detecta **versões comprometidas** (Trivy v0.69.4) por
  hardcoded blocklist.
- **+** Detecta **EOL iminente** (Nuxt 3 jul/2026, Node 20
  abr/2026).
- **+** Detecta **erros de schema** (`.golangci.yml` v1+v2) por
  AST-level check.
- **+** Modo offline preserva velocidade (CI não precisa de
  rede); modo online roda local/dev.
- **+** Tabela canônica rastreável (cada versão tem URL de
  fonte + data).
- **−** Modo online precisa de acesso à GitHub API e Docker Hub
  (pode ser bloqueado em redes corporativas).
- **−** `versions.md` precisa ser atualizado mensalmente
  (custo de manutenção do team-manager).
- **−** Renovate adiciona ruído se mal configurado (mas é
  opcional).

### Reversibilidade

- Remover `--check-latest` = deletar o bloco de checks online
  do script.
- Remover tabela `versions.md` = voltar para estado pré-ADR-0009
  (sem fonte canônica).
- Trocar pinning de "atrás da latest" para "exato na latest" =
  atualizar `versions.md` + ADR.

---

### Próximas ADRs a criar

- ADR-0012 — Estratégia de teste E2E (Playwright vs Cypress).
- ADR-0013 — Estratégia de release (release-please vs manual).
- ADR-0014 — Provider de observability (Prometheus + Grafana
  self-hosted vs Datadog vs Honeycomb).
- ADR-0015 — Estratégia de mensageria (RabbitMQ vs Kafka vs
  NATS vs Postgres outbox).

---

### ADR-0011 — CI modular com path filters + scope cache + concurrency (Mandaí v2 round 2)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** no round 1 do Mandaí v2 (jul/2026), o CI
`ci.yml` ficou com **6 problemas latentes** que o `versions.md` +
`check-stack-versions.sh` não tinham coberto: rodava TUDO
sempre (sem path filter), sem concurrency, sem scope em cache
Docker, `trivy-action@master` (tag mutável), sem
`GOTOOLCHAIN: local`, e Dockerfiles em paths não-canônicos
que podiam divergir entre `docker-compose` e CI.

### Contexto

- O `ci.yml` do Mandaí v2 (versão pré-ADR-0011) tinha:
  - **Lint, test, vuln, contract, build-and-scan, summary**: 6
    jobs monolíticos. Cada um rodava `setup-go` + `setup-node`
    + `pnpm install` (mesmo que o PR só mexesse em 1 arquivo de
    1 linha). Custo médio: ~8 min por PR rodando tudo.
  - **Sem path filter**: PR que muda só `web/i18n/en.json`
    dispara lint de Go, govulncheck, build de imagem backend.
  - **Sem concurrency**: push novo no PR não cancela a run
    anterior — builds duplicadas.
  - **Trivy `@master`**: tag mutável. Em mar/2026 a Aqua sofreu
    supply-chain attack que comprometeu a maioria das tags
    semânticas de `trivy-action` (76 de 77) — `@master` é a
    menos confiável de todas. (Versions.md gotcha #7.)
  - **Cache Docker**: usava `scope=backend` e `scope=frontend`
    (bom!) mas sem `mode=max` em `cache-from` (apenas
    `cache-to`), o que significa que o primeiro build do PR
    não aproveita cache de PRs anteriores.
  - **Sem `GOTOOLCHAIN: local`**: `govulncheck` rodando com
    `GOTOOLCHAIN=auto` (default) pode reescrever o `go.mod`
    se uma dep nova exigir Go ≥ X. (Versions.md gotcha #1.)
  - **Working-directory inconsistente**: alguns jobs Go tinham
    `cd backend` no `run:`, outros `working-directory: backend`
    no step. Mistura que falha em um job e passa em outro.
- O **docker-compose** do Mandaí v2 referencia
  `deploy/Dockerfile.backend` (correto, path canônico). Mas
  se o `ci.yml` apontasse para `backend/Dockerfile` (que não
  existe) ou tivesse um `Dockerfile` na raiz (que também não
  existe), o compose e o CI divergiriam silenciosamente —
  imagens diferentes.
- A pesquisa extensiva (jul/2026) mostrou:
  - `dorny/paths-filter` v3+ é o padrão da indústria para
    monorepo. 7.5x mais rápido que rodar tudo (Nx benchmark).
  - Native `on.push.paths` **não** permite conditional jobs
    (skip no job level) — só skip de workflow inteiro. Para
    monorepo com build+scan+lint+test+summary em 1 workflow,
    path filter é obrigatório.
  - `tj-actions/changed-files` (alternativa) teve security
    incident em 2023 — evitar.
  - `type=gha,scope=<service>` com `mode=max` é o caminho
    mais performático para cache Docker em GitHub Actions.
    **Scope** é crítico: sem ele, cache de backend invalida
    cache de frontend.
  - Turborepo/Nx/Bazel são overkills para 2 services
    (backend+frontend) — overhead de setup maior que
    benefício.

### Decisão

- **Refatorar `templates/.github-workflows-ci.yml`** com:
  1. **1 job `changes` no topo** com `dorny/paths-filter@v3.0.2`
     (SHA-pinned em prod crítica). Computa 6 outputs
     booleanos: `backend`, `frontend`, `infra`, `docs`,
     `workflow`, `contracts`.
  2. **12 jobs condicionais** com
     `needs: changes` + `if: needs.changes.outputs.<X> == 'true'`:
     - `backend-lint`, `backend-test`, `backend-vuln`,
       `backend-contract`
     - `frontend-lint`, `frontend-test`, `frontend-vuln`
     - `i18n` (roda se backend OU frontend OU workflow mudou)
     - `twelve-factor` (roda **sempre** — gate de segurança)
     - `build-backend`, `build-frontend` (com scope de cache
       separado)
     - `summary` (sempre, com `if: always()`)
  3. **`concurrency` no nível do workflow** com
     `cancel-in-progress: ${{ github.event_name == 'pull_request' }}`:
     cancela rodadas obsoletas em PRs; **nunca** cancela em
     main (protege release).
  4. **Cache Docker com `scope=<service>`** E `mode=max` em
     ambos `cache-from` e `cache-to`:
     - Backend: `scope=backend`
     - Frontend: `scope=frontend`
  5. **Trivy SHA-pinado** em
     `aquasecurity/trivy-action@0.36.0` (jul/2026, pós
     supply-chain attack mar/2026). Com nota "SHA-pine em
     prod crítica".
  6. **`GOTOOLCHAIN: local`** em **todos** os jobs Go
     (lint, test, vuln, contract). Impede `go mod tidy` de
     reescrever `go.mod` no CI.
  7. **`working-directory`** consistente em todos os steps
     (NÃO `cd backend &&` no `run:`).
  8. **i18n job adicionado** (estava faltando no template
     anterior).
  9. **Summary job** com tabela Markdown + lógica
     "fail if any non-skipped job failed".

- **Adicionar 2 invariantes ao `AGENTS.md` §8:**
  - **17: 1 Dockerfile por service em path canônico.**
    Proibido: `Dockerfile` na raiz; 2+ Dockerfiles pro mesmo
    service. Paths canônicos: `deploy/Dockerfile.backend`,
    `web/Dockerfile`, `migrate/migrate:v4.19.1` (imagem
    oficial, não custom build).
  - **18: CI modular com path filters.** Workflow DEVE ter
    1 job `changes` (dorny/paths-filter) + jobs condicionais
    + concurrency + scope cache + Trivy SHA-pinado +
    GOTOOLCHAIN=local.

- **Adicionar 2 seções ao `check-stack-versions.sh`:**
  - **9b. Dockerfile único por service** — detecta
    `Dockerfile` na raiz, 2+ Dockerfiles do mesmo service.
  - **9c. CI workflow** — detecta: path filter ausente,
    concurrency ausente, cache sem scope, trivy não-pinado,
    GOTOOLCHAIN ausente, working-directory inconsistente.

- **NÃO usar Turborepo/Nx/Bazel agora.** Para 2 services
  (Go + Node), `dorny/paths-filter` direto é mais simples e
  cobre 100% do caso. Migrar para Turborepo **só se**:
  - > 5 packages JS/TS compartilhando deps
  - Tempos de `pnpm install` > 2 min
  - Time > 5 devs (dependência de cache distribuído)

### Ganhos esperados (medidos no piloto Mandaí v2)

| Cenário                                | Antes (sem path filter) | Depois (com path filter) |
|----------------------------------------|-------------------------|--------------------------|
| PR só muda `web/i18n/pt-BR.json`       | ~8 min (12 jobs rodam)  | ~1.5 min (1-2 jobs)      |
| PR só muda `backend/internal/api/x.go` | ~8 min                  | ~3 min (4 jobs)          |
| PR só muda `deploy/docker-compose.yml` | ~8 min                  | ~4 min (build+scan)      |
| PR só muda `docs/SPEC.md`             | ~8 min (TUDO roda)      | ~30s (12-factor apenas)  |
| Push em main (release)                 | ~8 min                  | ~8 min (não muda)        |
| Cancel de PR (5 commits sucessivos)    | 5×8 = 40 min cumulativo | 1×8 = 8 min (cancel)     |

**Speedup médio em PRs de typo/doc/i18n-only: 5-10x.**
**Custo evitado por mês: ~30-50 USD** (depende do volume
de PRs e runner minutes).

### Bugs prevenidos pelo path filter

1. **PR só com typo dispara lint de Go** (era um no-op que
   gastava 30s só pra rodar `setup-go` + cache + lint).
2. **PR só com i18n dispara build de imagem** (Trivy scan
   levava 2 min, totalmente desnecessário).
3. **Cancelamento de PR**: 5 commits sucessivos criavam
   5 runs paralelas (e o cache era corrompido pela última
   a terminar). Com `cancel-in-progress: true`, só a
   última roda.
4. **Múltiplos Dockerfiles**: hoje o Mandaí v2 tem
   `web/Dockerfile` + `deploy/Dockerfile.backend` (correto,
   1 por service). Mas sem invariante 17, alguém pode
   amanhã criar `Dockerfile` na raiz "pra testar" e o
   `docker-compose` apontar pra um e o CI pra outro —
   divergência silenciosa.

### Alternativas consideradas

- **A:** Manter CI monolítico (status quo) — falha por
  desperdiçar 5-10x de tempo em PRs pequenos. Custo
  monetário e de DX.
- **B:** Turborepo (pnpm turbo run --filter) — overkill
  para 2 services; adiciona ~50 linhas de `turbo.json` +
  nova dep; `dorny/paths-filter` é 1 step e cobre o caso.
- **C:** Nx (nx affected) — overkill similar; ~100 linhas
  de config; curva de aprendizado de Nx graph.
- **D (escolhida):** `dorny/paths-filter` direto, com
  workflow monolítico mas jobs condicionais. Simplicidade
  > ferramenta adicional.

### Consequências

- **+** PRs de typo/doc/i18n rodam 5-10x mais rápido.
- **+** Builds incrementais (Trivy/Build só roda se a
  imagem mudou).
- **+** Cancel automático de runs obsoletas.
- **+** Cache Docker não invalida entre backend↔frontend.
- **+** Trivy SHA-pinado elimina supply-chain risk.
- **+** `GOTOOLCHAIN=local` impede `go.mod` rewrite
  surpresa.
- **+** 12-Factor audit SEMPRE roda (gate de segurança).
- **−** Workflow YAML ficou maior (de ~200 para ~400
  linhas) — mas é declarativo e organizado em seções
  numeradas.
- **−** Builders precisam entender que `actions/setup-go`
  agora está em jobs separados (não duplicar setup).
- **−** Em caso de mudança em dep compartilhada
  (`harness/scripts/`), o filter atual **NÃO** dispara
  ambos os grupos — bug futuro a corrigir adicionando
  `harness/**` no filter de `workflow` ou ambos.

### Reversibilidade

- Remover path filter = deletar `jobs.changes` + `if: needs.
  changes.outputs.X` de cada job (volta a ser monolítico).
- Trocar `dorny/paths-filter` por `tj-actions/changed-files`
  = mudar 1 step (mas eu desaconselho pelo incident de 2023).
- Adicionar Turborepo no futuro = instalar `turbo` no
  projeto + criar `turbo.json`; CI continua igual
  (Turborepo só afetaria build local).

---

### ADR-0012 — O que é (e o que não é) o meta-harness (v1.1.0)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** a v1.0.0 do `git-meta-harness` foi publicada com
o framework completo, mas sem uma **articulação explícita do
conceito** que o sustenta. O README descrevia o *quê* (personas,
sensores, templates) mas não o *porquê* da palavra "meta", nem a
diferenciação contra SDD/SPDD, nem a história da origem no
Hermes Agent, nem como o GitHub entra como substrate. Sem essa
articulação, o framework parecia "mais um scaffold"; com ela,
fica claro que é **um framework-versionado-no-GitHub que
materializa, a partir de uma especificação funcional, um time
de agentes de IA com papéis especializados, processo governado
por invariantes e sensores, e pipeline CI modular**.

### Contexto

- Em conversas pós-publicação da v1.0.0, ficou claro que
  adopters em potencial confundem o meta-harness com:
  - **Scaffold** (que gera projeto vazio, sem time nem
    processo).
  - **Single-agent system** (que põe 1 LLM pra fazer tudo,
    sem role separation).
  - **SDD / SPDD** (que usam spec mas ainda com 1 agente).
- Sem uma **definição canônica** do conceito, cada adopter
  reinterpreta, e a comunidade diverge sobre o que é "in scope"
  vs "out of scope" para o framework.
- A história de origem (exploração com Hermes Agent, perfis
  com modelos diferentes, descoberta de que a "pattern" era o
  asset, não o Hermes em si) era conhecida pelo mantenedor mas
  não estava documentada. Risco de **perda de contexto** quando
  o mantenedor original não estiver disponível.
- O framework **não rejeita** SDD/SPDD; ele **constrói em cima**
  deles. Mas isso precisa ser explícito.

### Decisão

- **Criar 4 documentos canônicos** em `git-meta-harness/docs/`:
  - `CONCEPT.md` (11K) — a visão completa: o que é, o que
    não é, princípios, output, "meta", conexão com GitHub.
  - `ORIGIN.md` (8.4K) — a história: single-agent loop →
    pivot com Hermes → descoberta do pattern → extração para
    meta-harness → validação no Mandaí v2 → lição "pattern >
    tool".
  - `COMPARISON.md` (9.6K) — tabela comparativa com single-agent,
    SDD, SPDD, meta-harness; quando usar qual; como se
    conectam.
  - `PIPELINE.md` (10K) — a integração com GitHub: 5 primitivas
    (Issues, PRs, Labels, Actions, Branch Protection) +
    issue lifecycle + PR convention + smart routing + CI
    workflow com path filters.
- **Adicionar uma seção "The concept in one paragraph"** no
  topo do `README.md` raiz, antes de qualquer outra seção.
  Resumo de 1 parágrafo + links para os 4 docs.
- **Versionar como v1.1.0** (minor bump — adição de
  documentação, sem breaking change no spec ou templates).
- **Não renomear nem refatorar nada** que estava na v1.0.0.
  Os docs são puramente aditivos.

### Por que esses 4 documentos (e não 1)

Cada documento tem um **público diferente**:

- `CONCEPT.md` — para **adopters** que precisam decidir "isto
  resolve meu problema?" antes de gastar 1 hora lendo o
  framework.
- `ORIGIN.md` — para **mantenedores** e **novos contribuidores**
  que precisam entender "por que é assim?" antes de propor
  mudanças.
- `COMPARISON.md` — para **engenheiros seniores** que já
  conhecem SDD/SPDD e querem ver a diferenciação concreta.
- `PIPELINE.md` — para **DevOps** que vão operar o CI/CD e
  precisam entender as primitivas do GitHub em jogo.

Juntar tudo num único `VISION.md` longo penalizaria quem
precisa de 1 dos 4 recortes.

### O que a v1.1.0 **NÃO** é

- **Não é uma breaking change.** Toda persona, sensor, ADR,
  invariante, template e skill da v1.0.0 permanece idêntico.
  Os 4 docs são aditivos.
- **Não é uma promessa de roadmap.** Os docs descrevem o
  estado atual; o roadmap está no `README.md` (seção
  "Roadmap").
- **Não é um post de blog.** Os docs são canônicos e
  versionados; um post (LinkedIn, dev.to, etc.) é trabalho
  separado.

### Alternativas consideradas

- **A:** Não escrever docs de conceito; deixar a comunidade
  inferir pelo README — falha por alto custo de onboarding e
  confusão com SDD/SPDD/scaffolds.
- **B:** Escrever 1 doc gigante `VISION.md` (~40K) cobrindo
  tudo — falha por penalizar quem precisa de 1 recorte
  específico; leitura única longa.
- **C (escolhida):** 4 docs focados (concept, origin,
  comparison, pipeline), cada um com 1 público claro.

### Consequências

- **+** Onboarding de novos adopters: ~5 min de leitura do
  CONCEPT.md dá o suficiente para decidir "isto resolve meu
  problema?".
- **+** Manutenção: novos contribuidores leem ORIGIN.md antes
  de propor mudanças radicais, evitando "vamos refazer do
  zero" repetido.
- **+** Diferenciação: a tabela em COMPARISON.md é referência
  canônica para "isto é ou não é SDD?".
- **+** Operação: PIPELINE.md é a doc de referência para
  DevOps que operam o CI.
- **+** v1.1.0 marca uma release menor com adição de
  documentação, deixando claro que a v1.0.0 era estável e
  esta é uma evolução additive.
- **−** 4 docs a mais para manter sincronizados quando o
  framework evoluir.
- **−** Risco de os docs "envelhecerem" se a v1.2.0+ mudar
  o conceito sem atualizar os docs. Mitigação: ADR-0013
  (futuro) pode impor "atualizar docs/* em qualquer ADR
  que mude o conceito".

### Reversibilidade

- Remover 1 doc específico = deletar o arquivo `.md`
  correspondente em `docs/`.
- Remover a seção "The concept" do README = reverter o
  commit que adicionou.
- Mudar a articulação do conceito = atualizar os 4 docs +
  ADR-0012 (com changelog).

---

### ADR-0013 — Personas são construídas sob demanda, não copiadas como template (v1.1.1)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** time de plataforma
**Contexto:** a v1.1.0 descreveu o meta-harness como
"materializando personas" mas não deixou explícito o bastante
que **as personas são construídas a partir do contexto do
projeto, não copiadas dos templates**. A distinção é
**crítica** porque é o que diferencia o meta-harness de uma
biblioteca estática de personas.

### Contexto

- No piloto Mandaí v2, o primeiro uso do `domain-expert` foi
  genérico (apenas renomeado), produzindo análise rasa. O
  smoke test pegou isso como invariante 12 violada. A
  **correção** foi gerar `domain-expert-mandai` com conteúdo
  específico do domínio (compra coletiva comunitária, Pix,
  multi-tenant, multi-role, i18n).
- O mesmo padrão se aplicou a todas as personas: o template
  `backend-engineer.md` é conceitual ("siga TDD, coverage
  ≥ 80%"), mas a persona que trabalhou no Mandaí v2 sabia
  que era Go 1.25, Gin, GORM, PostgreSQL, golang-migrate,
  oapi-codegen. A persona materializada é **outro arquivo**,
  **outro conteúdo**, **outro lugar** (no projeto, não no
  meta-harness).
- O v1.1.0 (`docs/CONCEPT.md` §6, "The 'meta' in meta-harness")
  falou sobre "the contract that any agentic tool must honor"
  mas não diferenciou template de materializada. Risco de
  adopters copiarem o template `domain-expert-banking.md`
  para `domain-expert-<seu-dominio>.md` sem mudar conteúdo,
  reproduzindo o bug do Mandaí v2 em outros projetos.

### Decisão

- **Adicionar seção §10 ao `docs/CONCEPT.md`**: "Personas are
  built on demand for each project". Cobre:
  - Tabela de duas camadas (template vs materialized).
  - Algoritmo de 5 passos do materialization step
    (ler spec → detectar stack → detectar domínio → gerar
    personas → gerar runtime adapter).
  - Onde as personas materializadas vivem (no projeto, não
    no meta-harness).
  - Como skills são injetadas dinamicamente baseado no
    stack detectado.
- **Adicionar seção §11 ao `docs/CONCEPT.md`**: "Anti-pattern:
  'I copied the personas, we're done'". Documenta o
  failure mode explicitamente.
- **Reescrever §1 do `harness/seed/meta-harness-seed.md`**:
  - Novo subsection "Materialização (sempre antes dos
    adapters)" com os 5 passos prescritivos.
  - Adapters por tool agora referenciam "personas
    materializadas" (não "personas").
  - Validation subsection explicitamente verifica que
    personas materializadas **não são** idênticas aos
    templates.
- **Versionar como v1.1.1 (patch)** porque é correção de
  documentação, não feature nem breaking change.

### Por que isso é uma decisão arquitetural (e não só textual)

A invariante 12 ("domain-expert sempre especializado") sem o
contexto de "construído sob demanda" permite 2 readings:

- (a) Renomeia o template `domain-expert.md` para
  `domain-expert-<domínio>.md` (passa o check, falha o
  espírito).
- (b) Gera conteúdo específico do domínio no
  `domain-expert-<domínio>.md` (passa o check E o espírito).

A v1.1.0 deixava (a) e (b) como interpretações válidas. A
v1.1.1 explicita que **só (b) é correto**. Isso muda o
comportamento esperado do `team-manager` no seed prompt.

### Alternativas consideradas

- **A:** Deixar ambíguo; confiar que adopters vão
  descobrir pelo smoke test — falha porque o smoke test
  atual não checa "conteúdo idêntico ao template", só
  checa "renomeado".
- **B:** Adicionar check explícito no smoke test que
  falha se persona materializada for idêntica ao
  template — adiciona complexidade, mas não esta
  release. (Fica como ADR-0014 candidate para 1.2.0.)
- **C (escolhida):** Documentar a distinção em CONCEPT.md
  e reescrever o seed para ser prescritivo. Patch v1.1.1.

### Consequências

- **+** Adopters entendem que o framework é uma **fábrica
  adaptativa**, não uma biblioteca estática.
- **+** O seed prompt é prescritivo: 5 passos claros para
  materializar.
- **+** Anti-pattern documentado explicitamente: renomear
  sem mudar conteúdo é falha.
- **+** v1.1.1 corrige a ambiguidade sem breaking change.
- **−** Quem já usou o framework e interpretou errado tem
  que regenerar as personas materializadas (mas é um
  refactor mecânico: rodar o seed de novo).
- **−** Possível ADR futura (1.2.0) para adicionar check
  automático no smoke test que detecta o anti-pattern.

### Reversibilidade

- Voltar para v1.1.0 = reverter o commit (mas o problema
  ressurge).

### ADR-0014 — Verify-after-build (sensor 09) + invariante 19 (v1.5.0)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** Brenon Araujo + team-manager
**Contexto:** piloto Mandaí v2 PR #5 — auto-relato de subagente
mascarou 5+ defeitos que passaram para a fase de validação humana.

### Contexto

Durante a construção do **PR #5 do Mandaí v2** (issue #1,
bootstrap skeleton, jul/2026), a postura inicial do
`team-manager` foi confiar no auto-relato de cada subagente
(`backend-engineer`, `frontend-engineer`, `devops-engineer`).
Resultado: **5+ defeitos passaram batido** e só foram pegos
quando o humano (Brenon) leu os arquivos diretamente:

| # | Defeito | Como o builder reportou | O que realmente estava |
|---|---------|-------------------------|------------------------|
| D1 | `go.mod` com `go 1.25.0` mas Dockerfile com `golang:1.22-alpine` | "go.mod está em 1.22.0" | `grep go.mod` mostrou 1.25.0 |
| D3 | `command: "-database ${DATABASE_URL} up"` (não expandia) | "CMD expansion OK" | shell do host não expandia; URL ficava vazia |
| D4 | Coverage 47.8% (em vez de 92%) | "92% coverage" | `go tool cover -func` mostrou 47.8% (sem `-coverpkg`) |
| D6 | govulncheck 2 vulns (quic-go, pgx) | "0 vulnerabilities" | `govulncheck ./...` mostrou 2 HIGH |
| D7 | Compose healthcheck `CMD-SHELL` em distroless | "healthcheck OK" | distroless não tem shell, healthcheck morria silencioso |
| D10 | happy-dom 15.11.7 com CVE | "audit clean" | `pnpm audit` mostrou CVE HIGH em dev-dep |

**Custo:** ~6 horas de debugging manual que poderiam ter sido
pegas em ~5 minutos com um sensor de verificação independente
rodado pelo `team-manager` **antes** de mover a sub-issue para
`in-review` ou pedir validação humana.

**Causa raiz:** o framework atual (até v1.4.0) **confia em
auto-relato** de subagente. Não há uma etapa explícita em que o
`team-manager` re-verifica a verdade dos claims. A invariante
16 ("Nenhum PR é aberto com CI local vermelho") só obriga o
**builder** a verificar; o `team-manager` recebe o "PRONTO" e
propaga.

### Decisão

Adicionar 3 mudanças coordenadas para que o `team-manager`
**verifique independentemente** antes de propagar "verde":

**1. Invariante 19 do `AGENTS.md`:**

> **Team-manager verifica, não confia.** Após um builder reportar
> "PRONTO" / "VERDE", o `team-manager` **re-executa** os checks
> críticos (re-lê `go.mod`/`Dockerfile`/`ci.yml`, roda
> `make lint && make test && make vuln`) **antes** de rotular
> como `in-review` ou pedir validação humana.

**2. Sensor 09 — `verify-after-build` (novo, em
`harness/sensors/09-verify-after-build.md`):**

Protocolo de 6 verificações que o `team-manager` roda
**ele mesmo**, entre `in-progress` e `in-review`:

1. Re-ler source-of-truth (`go.mod`, `Dockerfile`, `ci.yml`,
   `package.json`).
2. Re-rodar `harness/scripts/check-stack-versions.sh` (15 checks
   agora, com as seções 11-15 adicionadas nesta release).
3. Re-rodar os 3 comandos canônicos: `make lint && make test &&
   make vuln` (backend) e `pnpm lint && pnpm typecheck &&
   pnpm test:run && pnpm audit` (frontend).
4. Conferir `gh pr checks <id>` (não confiar no "CI passou"
   do builder).
5. Conferir o PR template (Como testar, Sensors, Changes).
6. Conferir coverage no escopo correto (`-coverpkg=...`, não
   diluída).

**3. Novas seções 11-15 no `check-stack-versions.sh` (v3 → v4):**

Para detectar automaticamente 5 classes de defeitos que o
`D1-D10` do Mandaí v2 tinha:

- **Seção 11:** Compose healthcheck `CMD-SHELL` em distroless
  (D7). Detecta e falha antes do merge.
- **Seção 12:** Compose `command:` com `${VAR}` sem `$$` escape
  (D3). Detecta expansão de shell do host que não funciona em
  exec form.
- **Seção 13:** Makefile `go test -coverprofile=` SEM
  `-coverpkg=` (D4). Detecta coverage diluída em main,
  generated, etc.
- **Seção 14:** `govulncheck` ausente do CI (D6). Obriga
  presença.
- **Seção 15:** `pnpm audit` ausente do CI (D10). Obriga
  presença.

Juntas, 11-15 pegam **5 classes de defeitos** que o
`check-stack-versions.sh` v3 NÃO pegava, e que o auto-relato
do builder tinha mascarado.

### Alternativas consideradas

- **A:** Confiar em auto-relato + pedir validação humana no
  PR. — Pro: zero overhead. Contra: já vimos que falha (D1-D10
  mostraram). Humana só vê no fim, depois de horas de debug.
- **B:** Adicionar 5 novos sensores ao QA (rodados DEPOIS do
  build). — Pro: separação clara de papéis. Contra: desperdiça
  QA em builds que já têm defeito; o builder pode re-trabalhar
  várias vezes até QA aprovar.
- **C (escolhida):** Sensor 09 + invariante 19 + seções 11-15
  no `check-stack-versions.sh`. — Pro: pega o defeito **antes**
  do QA rodar, evita loop "build → QA reprova → build → QA
  reprova"; team-manager fica responsável pela qualidade do
  claim de verde. Contra: ~3-5 min extras por sub-issue.

### Consequências

- **+** 5 classes de defeitos (D3, D4, D6, D7, D10) pegas
  automaticamente pelo `check-stack-versions.sh` (seções 11-15).
- **+** Auto-relato mentiroso/errado de builder é pego em ≤ 5
  min pelo sensor 09, em vez de ≥ 1 h de debugging manual.
- **+** team-manager tem responsabilidade explícita pela
  qualidade do claim de verde (não é mais "receber e repassar").
- **+** PR que chega ao QA com defeito obvious vira exceção
  (não mais norma).
- **−** ~3-5 min extras por sub-issue (custo do sensor 09).
- **−** Team-manager precisa de 1 nova skill: re-rodar
  comandos e ler outputs sem confiar em resumo.

### Reversibilidade

- Remover o sensor 09 + invariante 19 = reverter este commit.
  Mas os defeitos D1-D10 voltam.
- Desabilitar seções 11-15 do `check-stack-versions.sh` =
  comentar (mas perde detecção automática).
- Nenhuma migração obrigatória: tudo é aditivo (não quebra
  projetos existentes).

### Anti-pattern que motivou (NÃO FAÇA)

- ❌ "Builder disse PRONTO, vou repassar" → ✅ "Vou re-rodar
  `make test` e `gh pr checks` antes de mover para `in-review`."
- ❌ "Cobertura 92% é o que o builder disse" → ✅ `go tool
  cover -func=coverage.out | tail -1` mostra o número real.
- ❌ "CI passou, é isso" → ✅ `gh pr checks <id>` lista cada
  job; alguns podem ter passado no commit anterior e estar
  pendente.
- ❌ "Compose tá OK, vi o YAML" → ✅
  `check-stack-versions.sh --check-latest` valida 15
  invariantes.

### ADR-0015 — Release pipeline com GHCR + multi-deploy (v1.6.0)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** Brenon Araujo + team-manager
**Contexto:** meta-harness precisa fechar o ciclo com release
publicada + artefatos Docker prontos para deploy em produção.

### Contexto

Até a v1.5.0, o meta-harness cobria todo o ciclo
issue → branch → PR → merge, mas **não havia release pipeline
automatizado** que produzisse artefatos consumíveis fora do
repo. A release workflow legada (workflow 04) era **manual**:
devops-engineer tinha que buildar, taggear, e pushar
manualmente — fácil esquecer passos, fácil de quebrar o
versionamento.

Em produção, isso significa que o usuário (time de plataforma)
que adotava o meta-harness tinha que escrever seu próprio
release workflow — re-implementando a roda em cada projeto.

### Decisão

Adicionar um **release pipeline automatizado** que:

1. **Trigger:** push de tag `vX.Y.Z` na main (ou
   `workflow_dispatch` manual).
2. **Pre-flight:** re-roda `check-stack-versions.sh` e
   `smoke-test.sh` antes de qualquer build.
3. **Build:** multi-arch (amd64 + arm64) para backend + frontend,
   em paralelo, com cache `scope=backend-amd64` etc.
4. **Scan:** Trivy em CRITICAL — block imediato.
5. **Sign:** cosign (keyless, OIDC GitHub).
6. **SBOM:** SPDX anexado à GitHub Release.
7. **Push:** `ghcr.io/<owner>/<repo>/<service>:<tag>`.
8. **Release notes:** auto-geradas pelo `softprops/action-gh-release`.

O output é **imagens prontas** que podem ser deployadas em
**ECS, EKS, Docker Swarm, ou localmente via docker-compose**
— todos cobertos por [`docs/DEPLOY.md`](../../docs/DEPLOY.md).

### Componentes

| Componente                                                | O quê                                            |
|-----------------------------------------------------------|--------------------------------------------------|
| `templates/.github-workflows-release.yml`                 | O workflow em si (template, vai para `.github/workflows/release.yml` no projeto) |
| `harness/workflow/06-release-pipeline.md`                 | Workflow doc (como o team-manager aciona)        |
| `docs/DEPLOY.md`                                          | Como usar as imagens em ECS/EKS/Swarm/local      |

### Alternativas consideradas

- **A:** Manter release 100% manual. — Pro: zero infra. Contra:
  cada projeto re-implementa; drift de processo.
- **B:** Usar `release-please` (Google) para auto-versionar
  baseado em Conventional Commits. — Pro: zero trabalho humano
  em versionar. Contra: opinionated; força Conventional Commits
  estritos (o meta-harness aceita Conventional mas também
  permite PRs manuais).
- **C (escolhida):** Tag manual + pipeline totalmente
  automatizado. — Pro: controle humano sobre a versão, CI faz
  o resto. Contra: devops-engineer precisa rodar
  `git tag && git push`.

### Consequências

- **+** Release é repetível e auditável (cosign + SBOM + Trivy).
- **+** Multi-arch (amd64 + arm64) é default — Apple Silicon
  e Graviton funcionam sem esforço.
- **+** Time de plataforma consome as imagens direto do GHCR,
  sem rebuild local.
- **+** SLSA L3 (provenance via `docker/build-push-action@v6`
  com `provenance: true`).
- **−** Requer `GITHUB_TOKEN` com `packages: write` (já é
  default no GitHub Actions).
- **−** Requer configurar OIDC para cosign (já é built-in
  no GitHub Actions com `id-token: write`).

### Reversibilidade

- Remover o workflow = reverter o template. Mas se o usuário
  já materializou, precisa apagar `.github/workflows/release.yml`
  do projeto.
- Trocar de registry (ex.: ECR em vez de GHCR) = trocar
  `REGISTRY: ghcr.io` e adicionar credenciais AWS. Workflow
  em si é agnóstico.

---

### ADR-0016 — `gmh` CLI (Go single binary) (v1.6.0)

**Data:** 2026-07-18
**Status:** Aceito
**Decisor(es):** Brenon Araujo + team-manager
**Contexto:** adoção do meta-harness precisa ser 1-comando;
sync de versão precisa ser trivial; skills/personas/plugins
precisam de um registry.

### Contexto

Até a v1.5.0, o caminho para adotar o meta-harness num projeto
novo era:

1. Clonar `git-meta-harness`.
2. Copiar `harness/` para o projeto.
3. Editar `harness/seed/meta-harness-seed.md`.
4. Rezar para não estar desatualizado.

Em projetos existentes, sincronizar com a última versão era
manual — não havia `gmh sync`.

Além disso, à medida que o framework cresce (skills,
personas especializadas, plugins), precisa de um **registry**
e um **installer**. Distribuir via GitHub Releases é
suficiente; mas o usuário precisa de uma CLI para usar.

### Decisão

Adicionar uma **CLI `gmh`** (git-meta-harness):

1. **Linguagem:** **Go** (consistente com o stack; compila
   binário único sem runtime).
2. **Distribuição:** GitHub Releases com tag `cli-vX.Y.Z`.
3. **Bootstrap installer:** estilo AWS CLI v2 —
   `curl -sSL .../install.sh | bash` (e `install.ps1` no Windows).
4. **Comandos:** `install`, `sync`, `update`, `doctor`,
   `skills`, `personas`, `plugins`, `version`.
5. **Multi-platform:** linux/darwin/windows × amd64/arm64
   (5 binários por release).
6. **Versionamento atrelado ao meta-harness:** a versão da CLI
   é a mesma do framework (`v1.6.0`); o tag é `cli-v1.6.0` para
   não conflitar com tags de release do framework (`v1.6.0`).

### Componentes

| Componente                            | O quê                                            |
|---------------------------------------|--------------------------------------------------|
| `cli/`                                | Source da CLI (Go module)                        |
| `cli/cmd/`                            | Subcommands (cobra)                              |
| `cli/internal/harness/`               | Read/write do diretório `harness/`               |
| `cli/installer/install.sh`            | Bootstrap (Linux/macOS)                          |
| `cli/installer/install.ps1`           | Bootstrap (Windows)                              |
| `cli/Makefile`                        | Build cross-platform                             |
| `docs/CLI.md`                         | Documentação completa                            |
| `.github/workflows/cli-release.yml`   | Build + publish on `cli-vX.Y.Z` tag              |

### Alternativas consideradas

- **A:** Python com `pip install gmh` (do PyPI ou do GitHub).
  Pro: development mais rápido, sem etapa de cross-compile.
  Contra: requer Python 3.10+ instalado (Linux tem, mas
  Windows não por padrão), packaging é mais complexo.
- **B:** Node.js com `npm install -g gmh`. Pro: developers
  já têm Node. Contra: framework é backend Go; misturar
  linguagens é overhead; instalação requer Node 18+.
- **C (escolhida):** **Go single static binary.** Pro: zero
  deps no cliente (só baixar e chmod), cross-compile trivial,
  consistente com o stack, 1 binary ~5 MB.
  Contra: cold-start de desenvolvimento (precisa de CI matrix).

### Consequências

- **+** Adoção é 1 comando:
  `curl ... | bash && gmh install`.
- **+** Sync é 1 comando: `gmh sync` (preserva customizações).
- **+** Skills/personas/plugins têm registry via `gmh skills
  available`, `gmh personas create`, etc.
- **+** `gmh doctor` em CI substitui o `smoke-test.sh`
  (mas `smoke-test.sh` continua existindo para uso local
  sem a CLI).
- **+** Distribuição via GitHub Releases (já temos CI para isso).
- **−** Manter a CLI em sync com o framework (quando sai
  v1.7.0, sai `cli-v1.7.0` no mesmo dia).
- **−** Build matrix 5x por release (mas cache `gha` mitiga).

### Reversibilidade

- Remover a CLI = apagar `cli/` + `.github/workflows/cli-release.yml`
  + `docs/CLI.md`. O framework continua funcionando (o `harness/`
  pode ser clonado manualmente como antes).
- Trocar Go por Python = reescrever `cli/` em Python; os
  comandos expostos são os mesmos, então a UX não muda.


## ADR-0017 — UI/UX skills + design cercas (v1.7.0)

> **Decisão:** adicionar 2 skills de UI/UX (`nuxt-ui-patterns`,
> `ux-design-best-practices`) + cercar `domain-expert` para não
> falar de design + adicionar label `type/ui` para routing puro de
> UI.

### Contexto

- O **frontend-engineer** estava implementando UIs sem skill
  estruturada de Nuxt UI / UX best practices → inconsistência
  (modais onde deveria ser página, breadcrumbs faltando, etc.).
- O **domain-expert** estava direcionando design no meio do
  refinamento ("clicar no modal para confirmar exclusão") →
  desalinhamento entre o que o domínio queria e o que o
  frontend implementava.
- Incident concreto: Mandaí v2 (jul/2026) — domain-expert
  usou "clicar no modal" no AC, frontend implementou modal, mas
  o design system padrão é **página + breadcrumb**, não modal.
  Resultado: retrabalho.

### Decisão

#### 1. Skill `nuxt-ui-patterns` (Nuxt UI v3)

- Patterns de Nuxt UI v3.3.6 (UDashboardPage, UTable, UForm, etc.)
- Templates de referência oficiais:
  [nuxt-ui-templates/dashboard](https://github.com/nuxt-ui-templates/dashboard),
  [saas](https://github.com/nuxt-ui-templates/saas),
  [lms](https://github.com/nuxt-ui-templates/lms)
- Regra #0: **página primeiro, modal por último**
- Regra #1: **breadcrumbs sempre** em páginas 2+ níveis
- Regra #2: **comece pelo template oficial**, não reinvente

#### 2. Skill `ux-design-best-practices` (stack-agnostic)

- Aplica a qualquer framework (Nuxt, React, Vue, mobile)
- Modais: quando usar (raro), como fazer, a11y
- Breadcrumbs: estrutura semântica, quando mostrar
- Forms: inline validation, primary action, multi-step
- WCAG AA: contraste 4.5:1, tab nav, Esc, tap targets 44x44px
- Responsive: mobile-first, breakpoints
- Loading/empty/error states
- i18n: strings em `locales/*.json`

#### 3. Cerca de design no `domain-expert`

- Domain-expert **NÃO** especifica componentes de UI
  (modal, botão, card, sidebar, tab, etc.)
- Domain-expert **FALA** em comportamento (o **o quê** e o
  **por quê**), nunca em UI (o **como**)
- Tabela de exemplos:

  | ❌ Anti-pattern (design) | ✅ Correto (comportamento) |
  |---|---|
  | "Clicar no modal de confirmação" | "Confirmar exclusão antes de executar" |
  | "Toast verde de sucesso" | "Notificar o usuário do sucesso" |
  | "Drop-down de filtro" | "Permitir filtrar resultados por categoria" |

- **Quem decide UI**: `frontend-engineer` consulta as skills
  e decide o padrão apropriado.

#### 4. Cerca de design no `team-manager`

- Team-manager **detecta** quando o refinamento do domain-expert
  tem UI embutida (modal, botão, card, etc.)
- Team-manager **devolve** o refinamento para reformulação
- Inclui template de resposta no team-manager.md §4.1.1

#### 5. Label `type/ui`

- Para issues **puramente de design** (refatorar dashboard,
  aplicar design system, etc.) sem refinar regra de negócio
- Routing: `frontend-engineer` (consulta skills) → `qa` → `devops`
- **Pula** `domain-expert` e `solutions-architect` (não há domínio
  ou arquitetura a decidir)

### Consequências

- **+** Frontend-engineer tem skills estruturadas para implementar
  UIs consistentes.
- **+** Domain-expert não conflita com frontend-engineer.
- **+** Team-manager detecta e corrige desalinhamento cedo.
- **+** Templates oficiais do Nuxt UI aceleram implementação.
- **+** Acessibilidade WCAG AA garantida (não-negociável).
- **−** Domain-expert precisa aprender a falar em comportamento
  (curva de aprendizado inicial).
- **−** Team-manager precisa detectar design embutido (mais um
  sinal para ficar atento).

### Reversibilidade

- Reverter = remover as 2 skills, reverter a cerca do
  domain-expert, remover a label `type/ui`. ~30 min.
- Mudar para outro framework = atualizar a skill
  `nuxt-ui-patterns` (e renomear) — as regras de design
  permanecem as mesmas.
