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

---

## ADR-0018 — Cercas (técnica + design) + skill `domain-refinement` (v1.8.0)

> **Status:** Aceito (2026-07-18).
> **Contexto:** Mandaí v2 (jul/2026), validação do meta-harness.
> **Decisão:** `domain-expert` fala em **comportamento de
> domínio**, **NUNCA** em UI nem em tecnologia. **2 cercas
> simétricas** (Design v1.7.0 + Técnica v1.8.0) + skill
> dedicada `domain-refinement` que codifica o teste
> "e se a stack mudar?".

### Contexto

Em jul/2026, durante calibração do meta-harness com Mandaí v2,
identificamos **2 modos de falha complementares** do
`domain-expert`:

1. **Vazamento de design** (já tratado em v1.7.0, ADR-0017):
   o `domain-expert` direcionava UI no refinamento ("clicar
   no modal de confirmação", "drop-down com X, Y, Z",
   "toast verde"), prendendo o frontend-engineer num padrão
   ruim antes do design pensar. Cerca de Design adicionada
   em v1.7.0.

2. **Vazamento de tecnologia** (este ADR, v1.8.0): o
   `domain-expert` foi acionado para refinar issues
   **puramente técnicas** e direcionou implementação:
   - Issue #A ("configurar Helm chart de staging") —
     domain-expert escreveu "Helm chart com 3 réplicas
     e HPA 70% CPU" em vez de "suportar X usuários
     simultâneos no checkout".
   - Issue #B ("criar índice composto no PostgreSQL") —
     domain-expert escreveu "índice composto
     (tenant_id, created_at DESC)" em vez de "listagem
     eficiente para 10k pedidos, p95 ≤ 200ms".
   - Issue #C ("atualizar Trivy action para SHA pinned") —
     domain-expert escreveu "Trivy action SHA-pinned em
     `aquasecurity/trivy-action@<sha-de-v0.35.0>`" em
     vez de "scan de vulnerabilidades antes do merge".

3. **Routing errado**: o `team-manager` estava acionando
   `domain-expert` para `type/technical` / `type/infra` /
   `type/tech-debt` / `type/docs` / `type/ui`. Não há
   regra de negócio a refinar nessas issues.

### Decisão

**3 movimentos coordenados**:

#### 1. Cerca Técnica no `domain-expert.template.md`

Espelha a Cerca de Design (v1.7.0) com regra simétrica:

> **Se a frase que você está escrevendo tem nome de tecnologia
> (linguagem, framework, ORM, banco, fila, protocolo, action
> de CI) → reformule para descrever o COMPORTAMENTO de
> domínio ou o SLO/SLA esperado, não a implementação.**

**Tabela de transformação (resumo)**:

| ❌ Vazou (tech) | ✅ Certo (comportamento) |
|---|---|
| "Endpoint POST /api/v1/users" | "Permitir criar usuário" |
| "PostgreSQL com `gorm.Model`" | "Persistir o usuário" |
| "Redis TTL 5min" | "Resultados consistentes por 5min" |
| "OAuth2 + PKCE" | "Login seguro sem expor credenciais" |
| "Índice composto (a, b DESC)" | "Listagem eficiente para 10k (p95 ≤ 200ms)" |
| "3 réplicas + HPA 70% CPU" | "Suportar 1k usuários simultâneos" |
| "Trivy action SHA-pinned" | "Scan de vulnerabilidades antes do merge" |

#### 2. Skill `harness/skills/domain-refinement/SKILL.md`

Skill dedicada para o `domain-expert` (9.7KB) que codifica:

- **Cerca #0** — Você é o POR QUÊ (não o COMO). Camadas:
  Negócio (você) → Design (frontend) → Arquitetura
  (solutions-architect) → Implementação (builders).
- **Cerca #1** — Domínio fala em comportamento, técnico
  fala em mecanismo. Tabela completa de transformação.
- **Cerca #2** — Quando o tipo é `type/technical`,
  `type/infra`, `type/tech-debt`, `type/docs`, `type/ui`
  você **NÃO** é acionado (sinalizar ao `team-manager`).
- **Cerca #3** — Não mencione personas pelo nome (você
  descreve **o que precisa acontecer**, não **quem faz**).
- **Cerca #4** — Não feche issues, não crie branches, não
  escreva código.
- **Teste do "e se a stack mudar?"** — toda AC deve
  sobreviver à troca de stack (Go → Rust, Nuxt → React,
  PostgreSQL → MongoDB, REST → GraphQL).
- **Checklist pré-post** com 9 itens (ACs em
  comportamento, sem personas, sem tech, etc).

#### 3. §4.1.2 no `team-manager.md` — detecção + rerouting

Team-manager agora detecta tech vazando em DOIS eixos:

- **(a) Tipo errado da issue**: `type/technical`,
  `type/infra`, `type/tech-debt`, `type/docs`, `type/ui` →
  reroute imediato (script bash pronto).
- **(b) Tech vazando em ACs de domínio**: devolve para
  `domain-expert` reformular em comportamento de
  domínio ou SLO/SLA.

**Sinais de violação (em ambos eixos)**:

- Endpoints, payloads, JSON schema
- Frameworks (Vue, Pinia, Nuxt UI, Go, Gin, FastAPI)
- ORM/banco/queue (`gorm.Model`, PostgreSQL, Redis, SQS)
- Auth (OAuth2, JWT, mTLS, HMAC-SHA256)
- CI (Trivy action SHA-pinned, CODEQL, golangci-lint)
- Performance (índices compostos, réplicas, HPA)
- Resiliência (circuit breaker, DLQ, retry exponencial)
  — pode descrever **comportamento** de resiliência, mas
  não a tech.

#### 4. Invariante 20 na AGENTS.md

Codifica as 2 cercas + teste de stack-agnostic como
invariante não-violável do meta-harness.

### Consequências

- **+** `domain-expert` mantém foco no domínio. Não
  compete com `frontend-engineer` (UI) nem com
  `solutions-architect` (tech).
- **+** ACs viram **promessas stack-agnostic** que
  sobrevivem a migrações (PostgreSQL → MongoDB, REST →
  GraphQL, etc).
- **+** Histórico de issues não fica desatualizado quando
  a stack muda.
- **+** `team-manager` tem ferramenta de detecção em 2
  eixos (tipo errado + tech vazando).
- **+** Routing mais enxuto: `type/technical`,
  `type/infra`, `type/tech-debt`, `type/docs`, `type/ui`
  pulam `domain-expert` (sem overhead).
- **+** Skill dedicada dá referência canônica (não
  precisa decorar — basta consultar).
- **−** `domain-expert` precisa aprender a falar em
  comportamento (curva de aprendizado inicial,
  mitigada pela skill `domain-refinement` + tabela
  de transformação).
- **−** `team-manager` precisa detectar tech vazando
  (mais um sinal para ficar atento, mitigado pelo
  template de resposta em §4.1.2).

### Reversibilidade

- Reverter = remover a skill `domain-refinement`,
  reverter a Cerca Técnica do `domain-expert.template.md`,
  reverter §4.1.2 do `team-manager.md`, reverter
  invariante 20. ~45 min.
- Manter a Cerca de Design (v1.7.0) mesmo se reverter
  esta cerca (são independentes — uma foca em UI, outra
  em tech).
- Migrar a skill para outro framework = atualizar
  apenas os exemplos da tabela de transformação (as
  regras de "stack-agnostic" permanecem).

### Lições do Mandaí v2 (jul/2026)

| Sintoma observado | Correção |
|---|---|
| Domain-expert escreveu "Helm chart 3 réplicas" | "Suportar 1k usuários simultâneos" |
| Domain-expert escreveu "índice composto (a, b DESC)" | "Listagem eficiente p95 ≤ 200ms" |
| Domain-expert escreveu "OAuth2 + PKCE" | "Login seguro sem expor credenciais" |
| Domain-expert escreveu "Trivy action SHA-pinned" | "Scan de vulnerabilidades antes do merge" |
| Team-manager roteou `type/technical` para domain-expert | Reroute para `solutions-architect` |
| Domain-expert mencionou "@solutions-architect, valida X" | "A próxima etapa é validar o DoD técnico" |

**Custo evitado**: ~3 retrocessos por mês onde o
`domain-expert` reescrevia ACs de domínio toda vez que
a stack mudava (Go 1.22 → 1.25, SQS → Kafka,
monolith → microservice) — ~4h/mês de retrabalho
economizado.

**Validação**: aplicado em Mandaí v2 via
`gmh update --to v1.8.0 --force` (após release).
`gmh agents sync` deve instalar `domain-refinement`
skill em `~/.hermes/skills/` e propagar a Cerca Técnica
para o profile `domain-expert-mandai` (já customizado,
só atualiza, não sobrescreve).

---

## ADR-0019 — Decomposition Safety (path-scope + depends-on) (v1.9.0)

> **Status:** Aceito (2026-07-18).
> **Contexto:** Mandaí v2, Épico #12 (autenticação + role
> switching), jul/2026.
> **Decisão:** `team-manager` NUNCA dispara 2+ builders em
> paralelo sem antes validar que seus `path-scope` são
> disjuntos (ou têm `depends-on` explícito). 4 mudanças
> coordenadas: sensor 10 + script de detecção + DoD
> obrigatório + invariante não-violável.

### Contexto

Em jul/2026, durante a decomposição do Épico #12 (autenticação
+ role switching) do Mandaí v2 em 6 sub-issues (#13–#18), o
`team-manager` disparou 6 builders em paralelo **no mesmo
`cwd`** sem checagem de overlap de paths:

| Sub-issue | Builder | path-scope? |
|---|---|---|
| #13 backend auth-api | backend-engineer | ❌ |
| #15 backend user-role | backend-engineer | ❌ |
| #14 frontend homepage | frontend-engineer | ❌ |
| #16 frontend cadastro/login | frontend-engineer | ❌ |
| #17 frontend home auth | frontend-engineer | ❌ |
| #18 infra migrations+seed | devops-engineer | ❌ |

**Resultado (1ª tentativa)**:
- **#13 (auth-api) e #15 (user-role) ambos declararam
  interface `UserRepository`** no mesmo pacote
  (`internal/repository/`) → conflito de compilação.
- **#15 (user-role) referenciou tipos do auditlog** (que
  estava criando) antes de #15 terminar → erro de tipo
  indefinido.
- **Nenhum dos 6 builders chegou a commitar** — o trabalho
  foi perdido (working tree volátil dos processos Hermes
  que terminaram por "limite de iterações" + conflito).
- **Custo**: ~4h de orquestração desperdiçada, retrabalho
  manual necessário para consolidar o que sobrou.

**Causa raiz**: o meta-harness (até v1.8.0) **não tinha
mecanismo para detectar overlap de paths** entre sub-issues
em paralelo. O `team-manager` confiou em "backend e
frontend em arquivos separados" (`workflow/05-orchestration.md`
§2), mas **2 backends no mesmo package não são "arquivos
separados"** — eles compartilham imports e tipos.

### Decisão

**4 mudanças coordenadas** + invariante não-violável:

#### 1. Path-scope obrigatório no DoD (solutions-architect)

`solutions-architect` declara `path-scope: <glob>` (1+) no
DoD de cada sub-issue criada via decomposição. Glob syntax
mesma do `.gitignore` / `find -path`.

Sem path-scope = sub-issue não vai pra `ready`
(DoD rejeitado, `needs-info` para solutions-architect).

#### 2. Sensor 10-decomposition-safety + script automatizado

Novo
[`harness/sensors/10-decomposition-safety.md`](../sensors/10-decomposition-safety.md)
+ script
[`harness/scripts/check-parallel-builders.sh`](../scripts/check-parallel-builders.sh).

**Protocolo** (3 passos, ~2 min total):
1. Ler path-scope de cada sub-issue em `ready`
2. Calcular overlap dos globs (heurística: regex-based
   test paths matching)
3. Bloquear (exit 1) ou aceitar (exit 0)

Se overlap detectado E sem `depends-on` explícito → **exit 1**.

#### 3. Labels canônicas no repo

- `path-scope: <glob>` (1+ por sub-issue) — declarada no DoD
- `depends-on: #X` (1+ por sub-issue) — serialização explícita

GitHub renderiza `depends-on` nativamente (com app
[Blocked PRs](https://github.com/settings/blocked_prs)).

#### 4. Invariante 21 na AGENTS.md

Codifica path-scope + depends-on + bloqueio como
não-violável. Ver `harness/AGENTS.md` invariante 21.

#### 5. §6 nova no team-manager.md

`team-manager` ganha seção dedicada "Decomposition Safety"
com protocolo passo-a-passo, exemplos concretos do Épico
#12 (antes/depois), e comportamento esperado (BOM vs RUIM).

#### 6. workflow/05-orchestration.md §2 expandido

Texto de "Paralelizar o que dá" agora inclui referência
explícita ao sensor 10 e à lição do Épico #12.

### Consequências

- **+** `team-manager` tem ferramenta automática para
  detectar overlap antes de desperdiçar trabalho.
- **+** Sub-issues sem path-scope explícito **não passam**
  — força rigor na decomposição.
- **+** `depends-on` permite serialização explícita sem
  ambiguidade.
- **+** Custo do sensor é baixo (~2 min para rodar) e
  automatizado.
- **−** `solutions-architect` precisa adicionar path-scope
  em toda sub-issue (curva de aprendizado inicial, mitigada
  pela tabela de exemplos e regras de ouro).
- **−** `team-manager` precisa rodar sensor antes de cada
  batch de dispatch (mais 2 min, mitigado por ser
  automatizado e bloqueante).
- **−** Overlap detectado por glob heuristic pode ter
  **falso positivo** (2 globs que se sobrepõem em teoria
  mas não na prática) ou **falso negativo** (2 arquivos
  que conflitam mas os globs não foram declarados com
  granularidade suficiente). Mitigação: revisar manualmente
  o output do sensor antes de bloquear.

### Reversibilidade

- Reverter = remover sensor 10, script, §6 do
  `team-manager.md`, invariante 21, e regra de path-scope
  no `solutions-architect.md`. ~30 min.
- **Forward-compatible com v2.0.0**: v2.0.0 adiciona
  **worktree isolation** (cada builder em worktree
  separado) + **WIP commits** incrementais. Path-scope
  permanece útil mesmo com worktree (é a forma de
  declarar intenção, independente de onde o trabalho
  acontece).

### Worktree isolation (v2.0.0, PLANEJADO, NÃO NESTE ADR)

Limitação conhecida do v1.9.0: mesmo com path-scope
disjunto declarado, 2 builders no **mesmo filesystem**
podem ter race conditions em arquivos fora do path-scope
(ex.: `.git/`, `go.sum` se ambos rodam `go mod tidy`,
lock files, IDE state). Worktree isolation (cada builder
em `git worktree add` separado) elimina isso.

**Quando**: v2.0.0 (próximo major). Roadmap:
- Cada builder: `git worktree add ../projeto-#13 feature/13-...`
- `cwd` distinto por builder → sem conflito de filesystem
- Merge no fim via `git merge --no-ff` ou rebase
- Custo: ~5x mais código (precisa sync de go.mod, package.json,
  lock files)

### Lições do Épico #12 do Mandaí v2 (jul/2026)

| Sintoma observado | Correção (v1.9.0) |
|---|---|
| 6 builders em paralelo, mesmo `cwd` | Sensor 10 detecta overlap antes |
| Backend #13 e #15 ambos criaram `UserRepository` | path-scope `repository/auth.go` vs `repository/user.go` (disjunto) |
| Auditlog (#15) referenciou tipos não-existentes | `depends-on: #X` serializa ordem |
| Trabalho perdido (nenhum commit) | v2.0.0: WIP commits incrementais |
| Race conditions em `.git/`, `go.sum` | v2.0.0: worktree isolation |

**Custo evitado**: ~4h/epic de retrabalho manual
(consolidação de conflitos, debug de compilação, recriação
de branches). Em 4 épicos por mês = ~16h/mês economizado.

**Validação**: aplicado em Mandaí v2 via
`gmh update --to v1.9.0 --force` (após release).
Recriação do Épico #12 (em issue de exemplo) deve passar
o sensor 10 com path-scope corretamente declarado.

---

## ADR-0020 — Limite de função 25→35 + skill `pre-implementation-design` (v1.10.0)

> **Status:** Aceito (2026-07-18).
> **Contexto:** Mandaí v2, Épico #12 (auth + role switching),
> jul/2026, refinamento de golangci-lint.
> **Decisão:** limite duro de função 25 → 35 linhas (recomendado
> 25 mantido), + skill `pre-implementation-design` que força o
> builder a **pensar em abstração ANTES de codar**.

### Contexto

Durante o refinamento do Épico #12 (backend #15 user-role,
jul/2026), o `golangci-lint` reportou:

```
1 issue — OnboardRole tem 34 linhas (limite 25). Vou ver e
refatorar quebrando em função auxiliar.
```

**Resultado da "refatoração"** (relatado pelo builder):
> "Vou extrair a parte de criação da UserRole para uma função
> auxiliar."

Essa é uma **má prática clássica** (anti-pattern "split for
compliance"): uma função coesa de 34 linhas foi quebrada em
**4 funções coesas + 1 helper que só delega**, adicionando
**glue code** sem ganho real de leitura, e **mais código,
mais arquivos, mais imports**.

**Causa raiz**: o limite rígido de 25 linhas (v1.0.0) força
o builder a fazer **escolhas mecânicas** sobre decomposição
quando o natural seria manter a função coesa. O builder
não tem uma heurística para **decidir** entre "uma função
de 30 linhas" vs "duas de 15 linhas" — ele só tenta caber
no limite.

**Quem mais sofre com isso**:
- Builders que estão implementando **transações atômicas**
  (validate → create → persist → audit) que naturalmente
  têm 25-35 linhas.
- Builders que estão implementando **pipelines coesos** onde
  cada etapa é 1-2 linhas e o todo é legível de cima a
  baixo.
- Code review fica mais lento (4 funções pra navegar onde
  1 bastaria).

### Decisão

**2 mudanças coordenadas**:

#### 1. Limite duro 25 → 35 (recomendado: 25)

`templates/.golangci.yml` (e `sensors/00-static-analysis.md`):

```yaml
settings:
  funlen:
    lines: 35  # era 25
    statements: 30  # era 20
    ignore-comments: true
```

`stack/code-style.md` §"Funções / Tamanho":

| Faixa | Status | Ação |
|---|---|---|
| 0-25 linhas | ✅ Ideal | Manter assim |
| 26-35 linhas | ⚠️ Aceitável | Skill aplicada (justificativa no commit) ou decompor |
| 36+ linhas | ❌ Erro | `funlen` falha. Refatorar. |

**Recomendação mantida em 25** (não 35) — 35 é o **teto**,
não o **ideal**.

#### 2. Skill `pre-implementation-design` (NOVA)

[`harness/skills/pre-implementation-design/SKILL.md`](../skills/pre-implementation-design/SKILL.md)
(8.3KB) força o builder a **listar 2-3 decomposições possíveis
ANTES de implementar** e **justificar a escolha**:

```markdown
## Decomposição de `OnboardRole(user, role, tenantID)`

### Opção A — Função única (32 linhas) [escolhida]
- Coesa: pipeline de 1 transação atômica
- Glue mínimo
- Contra: harder to mock intermediate

### Opção B — 4 helpers (rejected)
- `validateOnboarding()` / `createUserWithRole()` / `auditOnboarding()`
- Contra: 4 funções pra navegar, glue explícito

### Escolha: A (32 linhas)
**Por quê**: transação atômica. Helpers de B fragmentariam
a leitura sem ganho real.

**Quando reverteria**: se `audit` virar compliance de outro
time (LGPD, BACEN), aí extrair faz sentido.
```

**Quem aplica**:
- `backend-engineer` (Go) — sempre que implementar função
  que pode passar de 25 linhas
- `frontend-engineer` (Vue/TypeScript) — sempre que implementar
  composable, helper, ou componente com lógica não-trivial

**Quem valida**:
- `team-manager` (sensor 09 verify-after-build): re-executa
  `make lint` e verifica que funções > 25 têm documentação
  da decisão.
- `quality-assurance` (sensor 02 unit tests): verifica que
  testes cobrem pelo menos 1 caso de borda por função > 25
  linhas.

### Consequências

- **+** Elimina o anti-pattern "split for compliance" —
  função coesa de 32 linhas pode ficar como está, com
  justificativa.
- **+** Força o builder a **pensar em trade-off** (coesão
  vs granularidade) ANTES de escrever código, evitando
  ambos os extremos (mega-função E split artificial).
- **+** Documentação da decisão (commit message com
  decomposição considerada) é explícita e code-reviewable.
- **+** Skill é genérica — funciona pra qualquer builder
  (backend, frontend, futuro mobile, etc).
- **−** Builder precisa gastar 2-3 min ANTES de implementar
  função não-trivial. Curva de aprendizado inicial.
- **−** Commit message fica mais verboso (3 opções + escolha
  + justificação).
- **−** Risco: builder pode abusar do limite 35 e criar
  funções de 35 linhas sem pensar (mitigado pela skill
  obrigatória, mas não 100%).

### Reversibilidade

- Reverter = baixar funlen pra 25, deletar skill, remover
  invariante 9a. ~10 min.
- Skill pode ser **adaptada** (não deletada) se o trade-off
  for refinado (ex.: split em 2 skills: "pre-impl" e
  "post-impl review").

### Worktree isolation (v2.0.0, PLANEJADO, NÃO NESTE ADR)

Limite 35 é teto. Casos de funções > 35 que **realmente**
precisam ser decompostas:
- Pipeline > 35 linhas → decompor por responsabilidade
- Múltiplas chamadas sequenciais (glue) → extrair helper
- Lógica com complexidade > 15 → decompor

### Lições do Mandaí v2 (jul/2026)

| Sintoma observado | Correção (v1.10.0) |
|---|---|
| `OnboardRole` 34 linhas → builder "quebrou em helper" (split for compliance) | Skill obriga listar 2-3 decomposições + justificar |
| Função coesa artificialmente quebrada em 4+ | Limite 35 permite manter coesa se for justificada |
| Builders sem heurística de decomposição | Skill fornece heurísticas (quando 1 função vs helpers) |
| Code review lento por navegar 4 funções | Reduz pra 1 função coesa (quando apropriado) |

**Custo evitado**: ~10-15 min/epic de "refatorar
funções que não precisavam ser refatoradas". Em 4
épicos por mês = ~40-60 min/mês economizado.

**Validação**: aplicado em Mandaí v2 via
`gmh update --to v1.10.0 --force` (após release).
Épico #12 (issues #13-#18 com `OnboardRole` em #15)
deve:
- `golangci-lint` passar com funlen=35
- Commit message do #15 deve incluir "Decomposition
  considered: A/B/C + Escolha: A (32 linhas)" se
  função ficar > 25
- QA deve revisar e aprovar com base na skill
  aplicada, não na forma da função

---

## ADR-0021 — Scope Discipline (PILARES vs BLUEPRINTS) (v1.11.0)

> **Status:** Aceito (2026-07-19).
> **Contexto:** Mandaí v2, Épico F4+F5 (Ciclos + Pedidos),
> jul/2026.
> **Decisão:** `domain-expert` e `solutions-architect`
> entregam **PILARES** (o que + por quê), **NÃO** BLUEPRINTS
> (o como). Builder tem **autonomia total** pra implementar.
> **Sensor 11 `scope-discipline` (NOVA) detecta vazamento
> de camada com regex heurística e emite RECOMENDAÇÃO
> (warning, não bloqueia) pra encurtar na próxima iteração.

### Contexto

Em jul/2026, durante o refinamento do Épico F4+F5 (Ciclos
de Compra + Pedidos) do Mandaí v2, o `domain-expert` e o
`solutions-architect` entregaram **blueprints detalhados**
em vez de **pilares arquiteturais**:

**`domain-expert` entregou**:
- 19 ACs (OK, é papel dele)
- 18 edge cases (OK)
- **Mas também mencionou**:
  - "snapshot em `cycle_products.price_cents INTEGER NOT NULL`"
  - "auto-join via `INSERT em workspace_members (ON CONFLICT
    DO NOTHING)`"
  - "máquina de estado (state machine)"
  - "Idempotência de webhooks (compliance BACEN)"

**`solutions-architect` entregou**:
- DoD de **150+ linhas**
- **Code-level decisions** (que são do builder):
  - Função `MustGenerateCycleSlug(name string) string` com
    `nanoid(12)` em `internal/service/nanoid.go`
  - `func (c *Cycle) CanTransition(to CycleStatus) error` em
    `internal/domain/cycle.go`
  - `INSERT INTO order_items (order_id, product_id, quantity,
    unit_price_cents) VALUES (..., cycle_product.price_cents, ...)`
  - Migrations `000009_cycles.up.sql`, `000010_cycle_products.up.sql`,
    `000011_orders.up.sql`, `000012_order_items.up.sql`,
    `000013_audit_log.up.sql`
  - Counter `orders_created_total{status}`,
    `cycle_transitions_total{from, to}`,
    `order_auto_join_total{workspace_id}`,
    `order_limit_rejections_total{reason}`

**Resultado**:
- O `backend-engineer` virou **executor cego** — só seguiu
  o blueprint, sem questionar, sem otimizar, sem ownership
  técnica.
- Decisões blueprinted (e.g., "retry de slug com max 5",
  "snapshot em 2 colunas cycle_products + order_items")
  estavam **engessadas** — quando o builder percebeu que
  algumas eram subótimas, ~3-5h de retrabalho.
- Domínio + arquitetura + implementação ficaram
  **misturados** num só documento. Impossível evoluir
  uma camada sem reescrever a outra.

**Causa raiz**:
1. `domain-expert` e `solutions-architect` não tinham
   **limites explícitos** de output
2. Eram incentivados a escrever mais (mais detalhe = "mais
   completo" do ponto de vista do modelo)
3. As cercas anteriores (Design v1.7.0, Técnica v1.8.0)
   focavam em **UI** e **linguagem/banco**, mas não em
   **pilares vs blueprints**

### Decisão

**3 mudanças coordenadas**:

#### 1. Princípio canônico: PILARES vs BLUEPRINTS

| Camada | Quem | Entrega |
|---|---|---|
| **Negócio** | `domain-expert-<x>` | Comportamento + regras + edge cases (≤ 12 ACs, ≤ 8 E-C) |
| **Arquitetura** | `solutions-architect` | 3-5 **pilares** (alto nível) + DoD macro (≤ 80 linhas) + 12-factor audit |
| **Implementação** | `backend-engineer` / `frontend-engineer` | **Tudo** o que precisar (escolha livre de linguagem, ORM, schema, query) |

> **PILAR** = decisão de arquitetura em alto nível (ex.:
> "consistência de preço via snapshot do momento de
> inclusão"). **BLUEPRINT** = instrução de implementação
> (ex.: "tabela `cycle_products.price_cents` copia valor
> para `order_items.unit_price_cents` no INSERT").

#### 2. Skill `solution-scoping` (NOVA, ~12KB)

[`harness/skills/solution-scoping/SKILL.md`](../skills/solution-scoping/SKILL.md)
codifica:
- Princípio central (PILARES vs BLUEPRINTS)
- 6 categorias com exemplos bons (pilares) vs ruins (blueprints):
  pricing, limites, state machine, idempotência, compliance,
  slug uniqueness
- Regras por persona (FAZ vs NÃO FAZ)
- Detector de vazamento (regex heurística)
- Limites recomendados (não-bloqueantes)
- Checklist pré-postar (domain-expert + solutions-architect)
- Quem detecta / valida

#### 3. Sensor 11 `scope-discipline` (NOVA) + script

[`harness/sensors/11-scope-discipline.md`](../sensors/11-scope-discipline.md)
+ [`harness/scripts/check-scope-discipline.sh`](../scripts/check-scope-discipline.sh)
(3.7KB).

**Comportamento**:
- **NÃO bloqueia** (diferente dos sensors 04-verify e
  10-decomposition-safety que bloqueiam)
- Detecta 10 padrões via regex:
  - `sql_keywords` (SELECT, INSERT, UPDATE, DELETE, WHERE,
    FROM)
  - `orm_names` (gorm, pgx, sqlx, sqlc, gin, echo, chi,
    fiber, nestjs, express)
  - `typeorm_nestjs` (TypeORM, GORM, PGx, etc)
  - `go_files` (`*.go`)
  - `internal_paths` (`internal/...`)
  - `migrations` (`00000N_*.up.sql`)
  - `endpoints` (`(GET|POST|PUT|PATCH|DELETE) /api`)
  - `func_names` (PascalCase function declarations)
  - `prometheus` (`prometheus.NewCounter`, etc)
  - `tokens` (> 75k chars = ~30k tokens)
- Thresholds **diferenciados**:
  - `domain-expert`: ≥ 1 (zero tolerância a tech)
  - `solutions-architect`: ≥ 2-5 (mais permissivo, pode
    mencionar stack pinada)
- **Emite recomendação** (warning) com template de
  reformulação. Builder **segue o que está escrito** mesmo
  se passar dos limites.

#### 4. `team-manager.md` §12 (NOVA)

Adiciona §"Scope discipline" com:
- Princípio (PILARES vs BLUEPRINTS)
- Protocolo de 3 passos (rodar sensor → interpretar →
  decidir)
- Template de reformulação
- Quando PULAR (output é só checklist ou ADR)
- Limites recomendados
- Quem detecta / aplica

#### 5. AGENTS.md invariante 22 (NOVA)

Codifica scope discipline como **não-violável** (mas
**não-bloqueante**) — alinhado com a invariante 20/21 que
são não-violáveis + bloqueantes.

### Consequências

- **+** Builder tem **autonomia total** pra implementar —
  escolhe linguagem, ORM, schema, queries
- **+** `domain-expert` foca em comportamento (não se perde
  em SQL/migrations)
- **+** `solutions-architect` foca em pilares (não se perde
  em funções específicas)
- **+** Documentos de refinamento são **menores e mais
  objetivos** (≤ 5k tokens vs ~10k+ anteriormente)
- **+** Mudanças de stack não invalidam ACs nem pilares
  (são stack-agnostic)
- **+** `team-manager` tem ferramenta pra detectar e
  recomendar encurtamento (sensor 11)
- **−** Builder recebe **menos guidance** (precisa pensar
  mais, questionar mais, ter ownership técnica)
- **−** `domain-expert` e `solutions-architect` precisam
  aprender a calibrar (curva de aprendizado)
- **−** Sensor 11 é **recomendação**, não bloqueante — pode
  ser ignorado se o team-manager não rodar

### Reversibilidade

- Reverter = remover skill `solution-scoping`, sensor 11,
  §12 do `team-manager.md`, invariante 22, reforçar cercas
  do `domain-expert` e `solutions-architect`. ~30 min.
- **Forward-compatible com v2.0.0**: worktree isolation
  (cap. 2.0.0) + WIP commits. Builder com autonomia total
  vai apreciar worktree isolation (trabalha sem afetar
  outros).

### Worktree isolation (v2.0.0, PLANEJADO, NÃO NESTE ADR)

Limitação conhecida do v1.11.0: builder tem autonomia
total, mas **compartilha filesystem** com outros builders
(mesmo `cwd`). Se 2 builders rodam em paralelo (com
path-scope disjoint), ainda há race condition em `.git/`,
`go.sum` (se ambos rodam `go mod tidy`), lock files. Worktree
isolation (`git worktree add ../projeto-#13 feature/13-...`)
elimina isso.

**Quando**: v2.0.0 (próximo major).

### Lições do Mandaí v2 (jul/2026)

| Sintoma observado | Correção (v1.11.0) |
|---|---|
| `domain-expert` escreveu 19 ACs + 18 E-C com SQL/ORMs | Skill obriga comportamento puro (sem tech) |
| `solutions-architect` escreveu DoD de 150+ linhas com code-level decisions | DoD ≤ 80 linhas + 3-5 pilares |
| `backend-engineer` virou executor cego | Builder tem autonomia total, é dono da implementação |
| ~3-5h de retrabalho por blueprint errado | Pilares stack-agnostic não envelhecem |

**Custo evitado**: ~3-5h/épico × 4 épicos/mês = ~12-20h/mês
de retrabalho evitado.

**Validação**: aplicado em Mandaí v2 via
`gmh update --to v1.11.0 --force` (após release). Próximo
épico (F6+) deve ter refinamento + DoD com ≤ 5k tokens
cada, sem nomes de funções/SQL/paths no output. Sensor 11
recomenda encurtar se passar.

---

## 0022 — Frontend Public Skills + Cold-Start Polish

**Status:** Accepted
**Data:** 2026-07-19
**Decisor:** Brenon Araujo via PR #23 do Mandaí v2 + discussão sobre
          "frontend-engineer não segue skills públicas"
**Relacionado:** issue #42 (Mandaí v2 cold-start), invariante 23 do
              AGENTS.md, sensor 12 `frontend-polish`, 3 templates Nuxt
              UI em `harness/templates/nuxt-ui/`

### Contexto

O PR #23 do Mandaí v2 ("Redesign Landing + i18n + Auth wizard",
jul/2026) expôs 5 problemas sistêmicos do `frontend-engineer`:

1. **Cores hex hardcoded** em `<style scoped>` (e.g., `#10b981`,
   `#064e3b`, `#ecfdf5`) em vez de usar os design tokens já
   definidos em `app.config.ts` (`primary: 'green'`).
2. **CSS BEM** (`.home-hero__title`, `.home-hero__ctas`) misturado
   com classes Tailwind — confuso e anti-pattern.
3. **Comentários redundantes** explicando o que o componente faz
   (e.g., `// HomeHero — top of the public landing page. Carries
   the one-liner tagline and the two primary CTAs...`).
4. **Emojis decorativos** sem função semântica.
5. **Zero uso do registry público de skills** (`https://www.skills.sh`,
   `npx skills`). A skill oficial `nuxt/ui@nuxt-ui` (15.2K installs)
   existe, é mantida pela Nuxt team, e tem padrões prontos — mas
   o builder não consultou.

**Resultado**: landing page com cara de "W3Schools 2018" em vez
de marketplace profissional. Custo: 1 iteração de retrabalho
polish + tempo de QA explicando o que estava errado.

**Causa raiz tripla**:
- O meta-harness **não documentava** o registry público (skill
  `nuxt-ui-patterns` só tinha 1 frase sobre templates oficiais).
- O meta-harness **não tinha mecanismo** de bloquear anti-patterns
  visuais (só sensor 08-i18n-audit, e esse não cobre UI quality).
- O `frontend-engineer` **não tinha regra explícita** "consultar
  skills públicas ANTES de implementar UI".

### Decisão

Esta versão (v1.12.0) ataca **as 3 causas raiz simultaneamente**:

1. **Documenta o registry público** via 3 skills:
   - **`frontend-public-skills`** (NOVA, 10.5KB): workflow
     canônico de `npx skills find/add/use`, lista curada por
     stack, MCP do Nuxt UI, validação de segurança.
   - **`tailwind-only-patterns`** (NOVA, 9KB): pra projetos sem
     Nuxt UI (shadcn-vue, Reka UI standalone, etc).
   - **`visual-polish`** (NOVA, 12.3KB): técnicas de polish
     (hierarchy, whitespace, contrast WCAG AA, consistency,
     motion, touch targets ≥ 44×44px).

2. **Detecta + bloqueia anti-patterns** via sensor 12
   `frontend-polish` (NOVA, 11.7KB) + `check-frontend-polish.sh`
   (14KB, Python-backed, 369 linhas, 10 categorias):
   - `hardcoded_colors`, `bem_naming`, `redundant_comment`,
     `emojis_excessive`, `spacing_off_scale`, `inline_color_style`,
     `off_stack_imports`, `img_no_alt`, `button_no_text`,
     `no_design_system`.
   - **BLOQUEANTE** (exit 1) — diferente do sensor 11 (recomendação).
   - **Razão de ser bloqueante**: refactor é trivial (< 5min) mas
     cold-start ruim custa caro (retrabalho + tempo de QA).

3. **Força a consulta ao registry** via 2 regras não-violáveis na
   persona `frontend-engineer.md`:
   - **Responsabilidade #0** (v1.12.0): "Consultar registry público
     de skills ANTES de implementar UI". Não escrever a primeira
     linha de `.vue`/`.css` sem antes ter rodado
     `npx skills find <seu-stack>`.
   - **Responsabilidade #13** (v1.12.0): "Screenshot local ANTES
     de abrir PR". Cold-start visual é feature, não polish step
     depois. Roda
     `harness/scripts/visual/playwright-screenshot.mjs` contra a
     rota nova.

4. **Fornece cold-start visual pronto** via 3 templates Nuxt UI
   em `harness/templates/nuxt-ui/`:
   - **`landing.vue`** (5.7KB) — hero + features + CTA + footer,
     tokens semânticos, hierarchy correta, zero emojis.
   - **`dashboard.vue`** (3.6KB) — admin panel com stats cards +
     `UDashboardPage` + `UDashboardNavbar`, sidebar via layout.
   - **`auth-form.vue`** (4.3KB) — login/signup reutilizável,
     preselect role via `?role=`, i18n pronto.

5. **Adiciona Visual Report** na persona `quality-assurance.md`:
   - Screenshot Playwright (3 viewports: 375/768/1440) por rota
     nova/alterada.
   - Checklist visual (hierarchy, whitespace, contrast, consistency,
     responsive).
   - Salvo em `qa/visual-report-<pr>.md` no repo do projeto.
   - Bloqueia se sensor 12 vermelho.

6. **Adiciona 2 scripts Playwright** em
   `harness/scripts/visual/`:
   - **`playwright-screenshot.mjs`** (3.7KB) — script Node
     standalone, `pnpm add -D playwright` + `pnpm exec playwright
     install chromium` (1x) + `pnpm screenshot` (a cada vez).
   - **`setup-playwright-screenshot.sh`** (2.6KB) — instala
     Playwright + adiciona `package.json` scripts (`screenshot`,
     `screenshot:setup`).

7. **Atualiza skill `nuxt-ui-patterns` pra v2.0.0** (253 → ~340
   linhas):
   - Frontmatter stack: `nuxt-ui-v3` → `nuxt-ui-v4` (Mandaí v2 usa
     `@nuxt/ui@^4.10.0`).
   - Seção "Public Skills Registry" (npx skills, MCP).
   - Seção "Anti-patterns" (5 sub-seções com exemplos bom/ruim).
   - Self-check expandido (npx skills, sensor 12, screenshot, etc).

8. **Adiciona invariante 23 ao AGENTS.md** (não-violável +
   bloqueante): "Frontend polish (cold-start visual)".

### Princípios

| Camada | Quem | Entrega |
|---|---|---|
| **Pre-flight** | `frontend-engineer` | `npx skills find <stack>` + `npx skills add <oficial>` + ler skill |
| **Implementação** | `frontend-engineer` | Código com tokens semânticos, hierarchy correta, zero hex hardcoded |
| **Pre-PR** | `frontend-engineer` | `pnpm screenshot` + sensor 12 verde |
| **PR review** | `team-manager` (você) | roda sensor 12 + Visual Report do QA |
| **Bloqueio** | sensor 12 + `team-manager` | exit 1, devolve com `in-review` + lista de violações |

### Custo evitado (estimativa)

- **~1 iteração de retrabalho polish por feature de UI**: ~30min
  por iteração × 4 features/mês = ~2h/mês.
- **~10-20min de QA explicando o que está errado** (quando builder
  empurra 3x com sensor vermelho): ~30min × 1x/mês = ~30min/mês.
- **Total**: ~2.5h/mês × 12 meses = ~30h/ano economizadas.

### Validação

- **Sensor 12 validado** com 4 casos:
  1. Componentes "limpos" (templates Nuxt UI) → exit 0 ✅
  2. Componentes com hex hardcoded (Mandaí v2 PR #23) → exit 1
     com lista + recovery ✅
  3. Componentes com BEM + emojis → exit 1 ✅
  4. Componentes com spacing off-scale (e.g., `p-3`, `gap-5`) →
     exit 1 ✅

- **Skill `frontend-public-skills` validada** com workflow real:
  1. `npx skills find nuxt-ui` → 6 skills listadas
     (`nuxt/ui@nuxt-ui` é oficial, 15.2K installs).
  2. `npx skills add nuxt/ui@nuxt-ui` → instala SKILL.md.
  3. Skill fica disponível pro agent (via `external_dirs` do
     Hermes).

- **Playwright screenshot validado** com Mandaí v2:
  1. `pnpm dev` em `localhost:3000`.
  2. `node harness/scripts/visual/playwright-screenshot.mjs
     --routes /,/auth/login`.
  3. Gera `qa/screenshots/landing-desktop.png` + `auth_login-desktop.png`.

- **Templates Nuxt UI validados** com copi-colar em Mandaí v2:
  1. Copiou `harness/templates/nuxt-ui/landing.vue` para
     `web/app/pages/index.vue`.
  2. Adicionou chaves i18n em `i18n/locales/{en,pt-BR,es}.json`.
  3. Rodou `pnpm dev` → tela renderiza sem erros.
  4. Rodou sensor 12 → exit 0.

### Quem detecta / Quem corrige

- **`frontend-engineer`**: roda sensor 12 local ANTES de PR. Se
  vermelho, **corrige antes de abrir PR** (não empurra pro QA).
- **`team-manager` (você)**: roda no PR review. Se vermelho,
  **devolve com** `in-review` + comentário listando violações.
- **`quality-assurance`**: roda + Playwright + Visual Report.
  Bloqueia se vermelho.
- **`solutions-architect`**: define tokens em `app.config.ts` e
  linka esta seção no DoD.

### Lições

1. **Cold-start visual é uma feature, não polish step depois.**
   Primeira renderização define percepção do produto. Não dá pra
   "consertar depois".
2. **Skills públicas existem — use o registry.** A Nuxt team
   mantém `nuxt/ui@nuxt-ui` (15.2K installs). É mantida,
   auditada, e tem padrões prontos. Reinventar é caro.
3. **Anti-patterns visuais são detectáveis automaticamente.**
   Cores hex hardcoded, BEM, emojis excessivos, spacing off-scale
   — tudo via regex. Sensor bloqueia em < 1s, refactor em < 5min.
4. **BLOQUEANTE > recomendação quando refactor é trivial.**
   Sensor 11 (scope) é recomendação porque reformular issue
   comment é caro. Sensor 12 (polish) é bloqueante porque
   refactor é trivial. Mesma arquitetura, threshold diferente.

### Custo

- **3 skills novas** (~32KB total) + skill atualizada
  (`nuxt-ui-patterns` v2.0.0, +10KB) = ~42KB de docs/skills.
- **1 sensor novo** (11.7KB) + script (14KB) = ~26KB de tooling.
- **3 templates** (~14KB total) = cold-start visual pronto.
- **2 scripts Playwright** (~6.4KB) = visual regression ready.
- **1 invariante nova** (12 linhas no AGENTS.md).

**Total**: ~100KB de framework. **Custo evitado**: ~30h/ano.


---

## 0023 — Agent config preservation (v1.12.1 HOTFIX)

**Status:** Accepted
**Data:** 2026-07-19
**Decisor:** Brenon Araujo (v1.12.0 pós-release, BRT)
**Relacionado:** v1.10.2 (gmh agents sync writes config.yaml), v1.12.0
              release, hotfix v1.12.1

### Contexto

A v1.10.2 introduziu `gmh agents sync` que escreve
`skills.external_dirs: ["~/.hermes/skills"]` no `config.yaml` de
cada profile. A v1.12.0 reaproveitou esse mecanismo pra propagar
o mesmo em todos os caminhos (`fresh/stale/preserved/skipped`).

A implementação original:

```go
type ProfileConfig struct {
    Skills *ProfileSkills `yaml:"skills,omitempty"`
}

func (c *Client) WriteConfig(profileName string, externalDirs []string) error {
    cfg := &ProfileConfig{}
    data, _ := os.ReadFile(path)
    yaml.Unmarshal(data, cfg)               // só lê campos conhecidos
    cfg.Skills = &ProfileSkills{ExternalDirs: externalDirs}
    out, _ := yaml.Marshal(cfg)             // só escreve campos conhecidos
    os.WriteFile(path, out, 0o644)
}
```

**Bug**: o `ProfileConfig` struct só conhece `skills`. Quando o
YAML é unmarshalado nele, `model.default`, `model.provider`,
`agent.reasoning_effort`, e qualquer outra chave custom são
**silenciosamente descartadas**. O `yaml.Marshal` então escreve
só `skills` de volta, **apagando** o resto.

**Sintoma observado** (Mandaí v2, jul/2026, pós-v1.12.0):

| Profile | Estado antes | Estado depois de `gmh agents sync` |
|---|---|---|
| `team-manager` | tinha `model` + `agent` | **ambos apagados** ❌ |
| `solutions-architect` | tinha `model` + `agent` | **ambos apagados** ❌ |
| `quality-assurance` | tinha `model` + `agent` | **ambos apagados** ❌ |
| `devops-engineer` | tinha `model` + `agent` | **ambos apagados** ❌ |
| `backend-engineer` | tinha `model` + `agent` | preservados (não re-sincronizado) |
| `frontend-engineer` | tinha `model` + `agent` | preservados (não re-sincronizado) |

O Hermes CLI recusava iniciar os 4 profiles afetados com erro
"missing required field: model.default".

**Custo**: ~30min para diagnosticar + restaurar manualmente 4
profiles. Risco de repetição em qualquer projeto que tenha
sofrido `gmh update` → v1.12.0 (todos os 5+ projetos com
meta-harness instalado).

### Causa raiz

A classe de bug é "**struct round-trip erasure**":

1. Ferramenta lê arquivo do usuário (config.yaml, package.json,
   .env, etc).
2. Ferramenta usa struct tipada (Go) / dataclass (Python) /
   interface (TS) com **subconjunto** dos campos.
3. Ferramenta escreve de volta.
4. **Qualquer campo fora do subset é apagado**.

Isso é diferente de "ler todos os campos e preservar" — que é o
que o usuário espera de uma ferramenta "não-destrutiva".

### Decisão (v1.12.1 HOTFIX)

`WriteConfig` agora usa `map[string]interface{}` (Go) / `dict`
(Python) em vez de struct tipada. O resultado é:

1. **Read**: unmarshal em `map[string]interface{}` → todos os
   campos são preservados em memória.
2. **Modify**: altere **apenas** `skills.external_dirs`.
3. **Write**: marshal de volta → todos os campos originais
   (incluindo futuros que o agent adicionar) sobrevivem.

```go
// v1.12.1 — fix
func (c *Client) writeConfigPreserveAll(profileName string, externalDirs []string) error {
    root := map[string]interface{}{}
    data, _ := os.ReadFile(path)
    if err := yaml.Unmarshal(data, &root); err != nil {
        // Não perder dados em parse error — backup + erro
        backup := path + ".bak-" + fmt.Sprintf("%d", os.Getpid())
        os.WriteFile(backup, data, 0o644)
        return fmt.Errorf("parse config %s: %w (backup at %s)", path, err, backup)
    }

    skills, _ := root["skills"].(map[string]interface{})
    if skills == nil {
        skills = map[string]interface{}{}
    }
    skills["external_dirs"] = mergeUniqueGeneric(skills["external_dirs"], externalDirs)
    root["skills"] = skills

    out, _ := yaml.Marshal(root)
    os.WriteFile(path, out, 0o644)
}
```

**Garantias**:

1. **Preserva `model`, `agent`, e qualquer outra chave**.
2. **Não duplica** `external_dirs` (mergeUnique dedupa).
3. **Backup em parse error** (não sobrescreve config corrompido).
4. **Teste de regressão**:
   [`cli/internal/hermes/hermes_test.go`](../cli/internal/hermes/hermes_test.go)
   `TestWriteConfigPreservesModelAgent` (com 4 fields + custom key
   + 3 external_dirs, verifica que TODOS sobrevivem).

### Princípio (pra todos os tools)

> **"A ferramenta nunca deve apagar dados do usuário ao
> re-escrever arquivo que ela não possui."**

Aplicações concretas:

| Ferramenta | Arquivo | Implementação |
|---|---|---|
| `gmh` (Go) | `~/.hermes/profiles/*/config.yaml` | `map[string]interface{}` (v1.12.1) |
| `gmh` (Go) | `package.json` (futuro) | `map[string]interface{}` + known-keys merge |
| `gmh` (Go) | `.github/workflows/*.yml` | `map[string]interface{}` |
| Edit de `pyproject.toml` | python | `tomllib` + `tomli_w` com merge key-by-key |
| Edit de `Cargo.toml` | rust | `toml` crate com merge key-by-key |

### Quem detecta / Quem corrige

- **`gmh`**: usa `writeConfigPreserveAll` internamente.
- **`team-manager`**: roda `gmh doctor` no PR review — se um
  profile perdeu `model.default`, **bloqueia** o merge.
- **`quality-assurance`**: roda `gmh doctor --strict` antes de
  aprovar qualquer update de framework.
- **User**: se notar profile quebrado, **backup** primeiro
  (`cp ~/.hermes/profiles/<name>/config.yaml{,.bak}`), depois
  investigar.

### Custo evitado

- **Por ocorrência**: ~30min (diagnóstico + restore manual).
- **Por release** (que toca o agent): ~5-10 projetos × 30min =
  ~2.5-5h.
- **Por ano** (12 releases): ~30-60h/ano.

### Lições

1. **Nunca use struct tipada pra round-trip de arquivo do
   usuário.** Use `map`/`dict`/`interface{}` (Go), `dict`
   (Python), `Record<string, unknown>` (TS).
2. **Parse error = backup + error, não silent overwrite.**
   O `writeConfigPreserveAll` faz `path + ".bak-" + pid` antes
   de retornar erro.
3. **Documente o invariante**: "framework não mexe em config
   do agent sem preservar campos do usuário". Esse é o tipo de
   coisa que precisa estar no `AGENTS.md` invariante
   (candidata a #24 em v1.13+).
4. **Teste de regressão é mandatório pra round-trip.** Não
   confie em "deve funcionar" — escreva o teste que reproduz
   o cenário do usuário (config com model+agent+custom,
   rodar WriteConfig, ler de volta, verificar tudo).


---

## 0024 — Hermes CLI invocation correctness (v1.12.2 HOTFIX)

**Status:** Accepted
**Data:** 2026-07-19
**Decisor:** Brenon Araujo (v1.12.1 pós-release, BRT)
**Relacionado:** v1.6.5 (primeira tentativa de agentic.Invocation),
              v1.12.2 (correção)

### Contexto

`cli/internal/agentic/agentic.go:Invocation(Hermes, profile, prompt)`
é usado pelo `gmh` para delegar trabalho ao Hermes. O comentário
do código (e a implementação) dizia:

```
// Hermes: `hermes chat -p <profile> -q "<prompt>"`
return fmt.Sprintf("hermes chat -p %s -q %s", profile, ...)
```

**Estava errado.** A flag `-p` no Hermes **NÃO é do subcomando
`chat`** — é uma **flag global** do `hermes` raiz, que seta o
profile para toda a sessão.

A forma correta (validada empiricamente em jul/2026):

```bash
$ hermes -p team-manager chat -q "echo test"
Query: echo test
Initializing agent...
[executa como team-manager]
```

A forma errada falha com:

```bash
$ hermes chat -p team-manager -q "hello"
hermes: error: argument command: invalid choice:
"hello" (choose from 'chat', 'model', 'moa', ...)
```

O parser do `chat` não conhece `-p` e trata `-p` + valor como o
argumento posicional (o subcommand).

**Sintoma** (Brenon, jul/2026, BRT): "o team manager nao esta
delegando corretamente para os profiles as issues via hermes".
Sempre que `gmh` tentava invocar o agentic, o comando falhava
silenciosamente ou usava o profile errado.

### Causa raiz

A função `Invocation()` foi escrita baseada em **documentação
de cabeçalho** (o próprio comentário) em vez de **execução real**.
O bug existia desde v1.6.5 (out/2025) e só foi pego quando o
Brenon tentou usar a delegação em produção.

Pior: o **mesmo tipo de bug** já tinha acontecido antes. Em
v1.6.5, a forma documentada era `hermes profile <name> --prompt`,
também errada. Foi corrigido pra `hermes chat -p <name> -q ...`
— mas essa correção também estava errada (v1.6.5 introduziu o
problema que v1.12.2 corrige).

### Decisão (v1.12.2 HOTFIX)

1. **Corretude da invocação**:

```go
// Antes (v1.6.5+ — bug):
return fmt.Sprintf("hermes chat -p %s -q %s", profile, shellQuote(prompt)), nil

// Depois (v1.12.2 — fix):
return fmt.Sprintf("hermes -p %s chat -q %s", profile, shellQuote(prompt)), nil
```

2. **Exemplos em docs corrigidos** (`team-manager.md` §6.6) — 2
   exemplos mostravam a forma errada, agora atualizados.

3. **Teste de regressão** (3 casos em
   `cli/internal/agentic/agentic_test.go`):

```go
func TestInvocation_Hermes_ProfileFlagBeforeSubcommand(t *testing.T) {
    got, _ := Invocation(Hermes, "backend-engineer", "...")
    if strings.HasPrefix(got, "hermes chat -p ") {
        t.Errorf("BUGGY form; Hermes `-p` is global (must come before `chat`)")
    }
    want := "hermes -p backend-engineer chat -q '...'"
    if got != want { t.Errorf(...) }

    // Bonus: if hermes is on PATH, run `hermes --help` and verify
    // the CLI shape (catches future breakage of the assumption).
    if path, err := exec.LookPath("hermes"); err == nil {
        out, _ := exec.Command("hermes", "--help").CombinedOutput()
        if !strings.Contains(string(out), "{chat,") {
            t.Errorf("hermes --help missing {chat,...}")
        }
    }
}
```

O CI automaticamente roda esse teste. Se `hermes` está no PATH
do runner, valida a forma do CLI; se não está, valida só a
string output. **Não confiamos só em comentário — validamos
execução.**

### Princípio (pra todos os agentics)

> **"Documentação em comentário não é validação."** Quando uma
> função produz um comando CLI (especialmente pra um binário
> externo cuja sintaxe pode evoluir), um unit test que asserta
> o formato do output é o **mínimo**. Melhor: live test que
> roda o comando e checa se o binário aceita.

**Aplicação concreta**:

| Agentic | Forma validada (jul/2026) | Onde está documentada |
|---|---|---|
| Hermes | `hermes -p <profile> chat -q "<prompt>"` | `agentic.Invocation()` + teste live |
| Claude Code | `claude -p "<prompt>"` (TBD) | `agentic.Invocation()` — falta live test |
| Codex | `codex -p "<prompt>"` (TBD) | idem |
| OpenCode | `opencode -p "<prompt>"` (TBD) | idem |

**Próximos passos** (candidatos a v1.13+):
- Adicionar live test pra `claude`, `codex`, `opencode`
  (quando Brenon usar esses tools).
- Adicionar sensor "agentic-invocation-validity" que roda
  `gmh doctor` e verifica que o comando produzido é
  executável (se o agentic está no PATH).

### Quem detecta / Quem corrige

- **`gmh`**: usa `Invocation()` que produz a forma correta.
- **CI (Go test)**: 3 testes em
  `cli/internal/agentic/agentic_test.go` rodam em todo PR
  (cobre o caso de regressão).
- **User**: se notar profile errado sendo invocado, rodar
  `gmh doctor` e verificar o comando sugerido (deve começar
  com `hermes -p <profile> chat`).

### Custo evitado

- ~5-15min/ocorrência (descobrir + corrigir invocação errada
  em workflow).
- ~30min-1h/épico se o team-manager não detecta e empurra
  com o profile errado.

### Lição histórica

Esse é o **segundo** "agentic invocation bug" do meta-harness:

| Versão | Forma documentada | Problema |
|---|---|---|
| v1.6.5 | `hermes profile <name> --prompt ...` | Não existe `--prompt` |
| v1.6.5–v1.12.1 | `hermes chat -p <name> -q ...` | `-p` não é do `chat` |
| **v1.12.2** | `hermes -p <name> chat -q ...` | ✅ validado |

**Padrão**: comentário desatualizado → implementação copiou o
comentário → CI não pegou (não havia live test) → user descobre
em produção.

**Fix sistemico** (proposto pra v1.13+): todo `agentic.Invocation()`
deve ter um live test que rode o binário com `--help` e valide
a forma. O test em
`TestInvocation_Hermes_ProfileFlagBeforeSubcommand` é o modelo.

