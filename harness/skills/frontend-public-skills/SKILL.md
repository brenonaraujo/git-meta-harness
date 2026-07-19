---
name: frontend-public-skills
version: 1.0.0
type: tooling
applies-to: frontend-engineer
---

# Frontend Public Skills — Registry & Workflow

Skill for the **frontend-engineer** persona. **Applies to ALL
UI work** (Nuxt UI, Tailwind-only, Vue, React, mobile). Tells
where to **find and install public skills** before reinventing
the wheel.

> **Lição do Mandaí v2 (jul/2026, PR #23):** o `frontend-engineer`
> entregou a landing page com cores **hardcoded** (`#10b981`,
> `#064e3b`), CSS BEM em vez de Tailwind/Nuxt UI, e zero uso
> de skills públicas (`npx skills`). Resultado: tela com cara
> de "W3Schools 2018" em vez de marketplace profissional. Esta
> skill é o **pre-flight obrigatório** que faltou.

---

## 🚨 Rule #0 — Consult the registry BEFORE implementing

> **ANTES de escrever a primeira linha de `.vue`/`.tsx`/`.css`,
> você PRECISA ter consultado o registry público de skills**
> e instalado pelo menos a skill mais relevante pro seu stack.

A ordem **NÃO é negociável**:

```
1. Identificar stack (Nuxt UI, Tailwind-only, Vue, React, …)
2. npx skills find <stack>    → descobrir skills disponíveis
3. npx skills add <owner/repo@skill>  → instalar a mais relevante
4. LER a skill instalada (ela é a "documentação viva" do stack)
5. AGORA SIM, começar a implementar
```

**Por quê:**

- Skills públicas são mantidas **pelos criadores do framework**
  (Nuxt team mantém `nuxt/ui@nuxt-ui`, Tailwind team mantém
  `lombiq/tailwind-agent-skills@tailwind-4-docs`, etc).
- Têm **versionamento** (sempre compatíveis com a versão pinada).
- Têm **exemplos auditados** e **padrões de produção** (não
  invenção de quem nunca usou).
- O registry `skills.sh` tem **security audits** (Gen Agent
  Trust Hub, Socket, Snyk) — você sabe o que está executando.

---

## O que é o registry público

- **URL canônica**: [https://www.skills.sh](https://www.skills.sh)
- **Quem mantém**: Vercel (vercel-labs/skills no GitHub).
- **Cobertura**: 6.000+ skills, organizada por **owner/repo@skill**.
- **Como navegar**: busca por stack (`nuxt`, `vue`, `tailwind`,
  `react`, `playwright`, `shadcn`, …) ou por domínio
  (`visual`, `accessibility`, `screenshot`, `auth`).
- **Segurança**: cada skill tem `/security` com audit de
  Socket, Snyk, e Gen Agent Trust Hub. **Antes de instalar,
  conferir se está "Pass" em pelo menos 2 auditores.**

---

## Comandos essenciais

### `npx skills find <query>`

Busca skills por stack ou domínio. Exemplo:

```bash
$ npx skills find nuxt-ui
nuxt/ui@nuxt-ui                 15.2K installs  (oficial)
└ https://skills.sh/nuxt/ui/nuxt-ui

onmax/nuxt-skills@reka-ui      6.6K installs
└ https://skills.sh/onmax/nuxt-skills/reka-ui

onmax/nuxt-skills@nuxt-ui      6.1K installs
└ https://skills.sh/onmax/nuxt-skills/nuxt-ui

nuxt/ui@contributing           404 installs
secondsky/claude-skills@nuxt-ui-v4   307 installs
mui.nuxt.com@nuxt-ui           139 installs
```

### `npx skills add <owner/repo@skill>`

Instala a skill localmente. Após instalar, a skill vira um
arquivo SKILL.md que seu agente consome automaticamente
(via external_dirs do Hermes, ou similar em outros tools).

```bash
# Stack Nuxt UI
npx skills add nuxt/ui@nuxt-ui
# Stack Tailwind-only
npx skills add wshobson/agents@tailwind-design-system
# Visual design (cross-stack)
npx skills add leonxlnx/taste-skill@high-end-visual-design
# Vue genérico (Nuxt UI por baixo)
npx skills add antfu/skills@vue
```

### `npx skills use <package>@<skill>`

Gera um **prompt pronto** pra você colar no agent. Útil pra
"experimentar" sem instalar (não persiste).

### `npx skills list`

Lista skills já instaladas localmente.

### `npx skills update [skills...]`

Atualiza skills instaladas pra versão mais recente.

---

## Lista curada por stack (v1.12.0)

### Nuxt UI v4 (padrão do Mandaí v2, pin: `@nuxt/ui@^4.10.0`)

| Skill | Installs | Por quê |
|---|---|---|
| `nuxt/ui@nuxt-ui` | 15.2K | **Oficial** Nuxt team. Cobre 125+ componentes, design system, theming. **PRIMEIRA ESCOLHA**. |
| `onmax/nuxt-skills@reka-ui` | 6.6K | Reka UI (base do Nuxt UI v4). Pra entender primitivos headless. |
| `onmax/nuxt-skills@nuxt-ui` | 6.1K | Alternativa mantida pela comunidade. |

**MCP server (recomendado para agent em runtime)**:
```bash
claude mcp add --transport http nuxt-ui https://ui.nuxt.com/mcp
```

Isso dá pro seu agent acesso direto à API de componentes
do Nuxt UI v4 (props, slots, eventos, exemplos).

### Tailwind v4 (sem Nuxt UI)

| Skill | Installs | Por quê |
|---|---|---|
| `lombiq/tailwind-agent-skills@tailwind-4-docs` | 7.7K | **Oficial Tailwind v4 docs** (Lombiq mantém). |
| `wshobson/agents@tailwind-design-system` | 55K | Design system patterns (spacing scale, typography). |
| `heygen-com/hyperframes@tailwind` | 71.2K | Mais popular — bom pra layouts. |
| `giuseppe-trisciuoglio/developer-kit@tailwind-css-patterns` | 13.7K | Patterns comuns (cards, forms, navbar). |

### shadcn (Tailwind + componentes copiados)

| Skill | Installs | Por quê |
|---|---|---|
| `shadcn/ui@shadcn` | 242.4K | **A mais popular do registry**. Cobre CLI, theming, copy-paste components. |

### Vue genérico (qualquer projeto Vue/Nuxt)

| Skill | Installs | Por quê |
|---|---|---|
| `antfu/skills@vue` | 29.7K | Anthony Fu (criador VueUse). Padrões idiomáticos Vue 3. |
| `hyf0/vue-skills@vue-best-practices` | 24.8K | Time do Evan You. Conventions. |

### Visual design (cross-stack)

| Skill | Installs | Por quê |
|---|---|---|
| `leonxlnx/taste-skill@high-end-visual-design` | 210.6K | **A mais popular de visual**. Princípios de design. |
| `wshobson/agents@visual-design-foundations` | 10.1K | Foundations (hierarchy, contrast, color). |

### Playwright (testes e screenshot)

| Skill | Installs | Por quê |
|---|---|---|
| `microsoft/playwright-cli@playwright-cli` | 91.5K | **Oficial Microsoft**. CLI completa. |
| `currents-dev/playwright-best-practices-skill@playwright-best-practices` | 62.5K | Best practices + patterns. |

---

## Workflow por caso

### Caso A — Projeto Nuxt UI v4 (padrão mandai-v2)

```bash
# 1. Setup do MCP (opcional, mas recomendado)
claude mcp add --transport http nuxt-ui https://ui.nuxt.com/mcp

# 2. Instalar skill oficial
npx skills add nuxt/ui@nuxt-ui

# 3. Consultar a skill antes de implementar
# (ela vira um SKILL.md que seu agent consome automaticamente)
cat ~/.claude/skills/nuxt-ui/SKILL.md    # ou equivalente do tool
```

### Caso B — Projeto Tailwind-only (sem Nuxt UI)

```bash
# 1. Instalar design system + docs oficiais
npx skills add lombiq/tailwind-agent-skills@tailwind-4-docs
npx skills add wshobson/agents@tailwind-design-system

# 2. Se for usar shadcn
npx skills add shadcn/ui@shadcn
```

### Caso C — Adicionar polish visual (qualquer stack)

```bash
npx skills add leonxlnx/taste-skill@high-end-visual-design
# Esta skill complementa (não substitui) nossa skill
# interna visual-polish.
```

---

## Segurança: como validar uma skill antes de instalar

Skills são **código executado pelo seu agent**. Antes de
instalar:

1. **Conferir auditoria**: `https://www.skills.sh/<owner>/<repo>/<skill>/security`
   - **Pass em 2+ auditores** (Gen Agent Trust Hub + Socket é o mínimo).
   - Snyk "Warn" é aceitável se a warning for conhecida.
2. **Olhar o source**: `https://github.com/<owner>/<repo>/blob/main/skills/<skill>/SKILL.md`
   - Não tem código executável (só markdown), mas pode ter
     exemplos de código que seu agent vai seguir.
3. **Conferir o owner**: Nuxt team (`nuxt`), Tailwind
   (`lombiq`, `tailwindlabs`), Vue team (`vuejs`, `antfu`,
   `hyf0`), Microsoft (`microsoft`) são confiáveis.
4. **Evitar skills abandonadas**: `installs < 50` = provavelmente
   experimental, testar em branch separada.

---

## Quando NÃO usar public skills

- **Tweak pequeno e óbvio** (1-2 linhas de CSS, classe Tailwind
  simples): não precisa instalar skill, use a documentação
  oficial do framework direto.
- **Conhecimento já consolidado no framework**: a skill
  `nuxt-ui-patterns` (interna) já cobre 80% dos casos comuns.
  Public skills complementam, não substituem.
- **UI crítico de segurança** (PIX flow, KYC, pagamento):
  **NÃO confiar cegamente em skill pública** sem revisar.
  Use a skill como **referência**, mas o `quality-assurance`
  precisa validar o output final.

---

## Self-check pré-implementação

Antes de abrir PR de UI, confirme:

- [ ] `npx skills find <seu-stack>` rodou E o output foi
      lido/considerado
- [ ] Pelo menos 1 skill relevante foi instalada (a oficial do
      framework é o mínimo)
- [ ] MCP do Nuxt UI (se aplicável) está configurado e
      respondendo
- [ ] A skill interna `visual-polish` foi consultada
      (hierarchy/whitespace/contrast)
- [ ] O `package.json` foi lido — você sabe qual versão de
      Nuxt UI / Tailwind está pinada
- [ ] O `app.config.ts` foi lido — você sabe quais cores
      semânticas o projeto usa
- [ ] Screenshot local foi gerado com `pnpm dev` +
      Playwright (ver `harness/scripts/visual/playwright-screenshot.mjs`)

---

## Anti-patterns (NÃO faça)

❌ **Inventar design do zero** sem consultar skill oficial.
Use a skill oficial como ponto de partida.

❌ **Hardcoded colors** (`#10b981`, `rgb(...)`, `hsl(...)`).
Use **sempre** os tokens semânticos do Nuxt UI
(`primary`, `neutral`, `success`, etc) ou Tailwind v4
(`text-primary`, `bg-elevated`).

❌ **CSS BEM** misturado com Tailwind (`.foo__bar` + classes
Tailwind). Use **um** sistema de estilização por projeto.

❌ **Emojis excessivos** em UI (mais de 5% do conteúdo
visível, ou em componentes que não pedem ícone). Use
**ícones** da coleção padrão (lucide no Nuxt UI).

❌ **Pular a consulta ao registry** "porque o framework é
óbvio". O framework sempre evolui, e a skill pública
documenta os **padrões atuais**, não os de 2 anos atrás.

---

## Referências

- **Registry**: [https://www.skills.sh](https://www.skills.sh)
- **Vercel Labs GitHub**: [https://github.com/vercel-labs/skills](https://github.com/vercel-labs/skills)
- **Nuxt UI v4 docs**: [https://ui.nuxt.com](https://ui.nuxt.com)
- **Nuxt UI MCP**: [https://ui.nuxt.com/docs/getting-started/ai/mcp](https://ui.nuxt.com/docs/getting-started/ai/mcp)
- **Skill interna `nuxt-ui-patterns`**: [../nuxt-ui-patterns/SKILL.md](../nuxt-ui-patterns/SKILL.md)
- **Skill interna `tailwind-only-patterns`**: [../tailwind-only-patterns/SKILL.md](../tailwind-only-patterns/SKILL.md)
- **Skill interna `visual-polish`**: [../visual-polish/SKILL.md](../visual-polish/SKILL.md)
- **Sensor 12 `frontend-polish`**: [../../sensors/12-frontend-polish.md](../../sensors/12-frontend-polish.md)
- **ADR-0022** (esta decisão)
