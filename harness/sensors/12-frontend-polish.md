# Sensor 12 — Frontend Polish (cold-start visual)

> **Objetivo:** bloquear PRs de UI que tenham **anti-patterns
> visuais** óbvios (cores hardcoded, BEM, comentários
> redundantes, emojis excessivos, spacing fora da escala).
> **Quem roda:** `frontend-engineer` (e `quality-assurance`
> no Visual Report). **Quando:** local, ANTES do PR; CI,
> no job `frontend-polish`. **Falha → ação:** BLOQUEAR
> (exit 1) com lista de violações + recovery.

---

## Por que este sensor existe

**Lição do Mandaí v2 (jul/2026, PR #23 — Redesign Landing):**
o `frontend-engineer` entregou a landing page com:

| Anti-pattern | Exemplo no PR #23 | Detector |
|---|---|---|
| **Cores hex hardcoded** | `background: #ecfdf5`, `color: #064e3b` | regex `#([0-9a-fA-F]{3,8})\b` em `.vue`/`.css` |
| **CSS BEM** | `.home-hero__title`, `.home-hero__cta` | regex `\.[a-z][a-z0-9-]*__[a-z]` |
| **Comentários redundantes** | `// HomeHero — top of the public landing page. Carries the one-liner tagline…` | heurística: comentário repete nome do componente/função |
| **Emojis excessivos** | (vários no PR de PersonaPicker) | regex Unicode emoji, threshold > 1 em componente sério |
| **Spacing fora da escala** | `p-3`, `gap-5`, `mt-7` | regex `(\b[pm][xytblr]?-\|gap-)([0-9]+)\b` filtrando valores permitidos |

**Resultado do PR #23:** tela com cara de "W3Schools 2018"
em vez de marketplace profissional. Custo: 1 iteração de
retrabalho pra repolir + tempo de QA explicando o que
estava errado.

**Causa raiz:** o `frontend-engineer` **não consultou** as
skills públicas (registry `skills.sh`), **não viu** que o
`app.config.ts` já tinha `primary: 'green'`, e **não
usou** as skills internas (`nuxt-ui-patterns` e
`visual-polish`). Esta versão do framework ataca **as 3
causas raiz simultaneamente** (skill, sensor, persona rule).

**Princípio:** primeira renderização de uma feature nova
**DEVE** parecer profissional. Cold-start visual é uma
**feature**, não um "polish step depois". Se o screenshot
local não parece Linear/Notion/Vercel/Stripe, **refazer
antes de abrir PR**.

---

## Como rodar (3 passos)

### Passo 1 — Instalar o sensor (uma vez, no bootstrap do projeto)

```bash
# Já vem no git-meta-harness. Em outros projetos:
./harness/scripts/setup-frontend-polish.sh   # copia sensor + script
```

### Passo 2 — Rodar local antes do PR

```bash
# Roda contra todos os .vue/.css do diretório web/
./harness/scripts/check-frontend-polish.sh

# Ou contra um arquivo específico
./harness/scripts/check-frontend-polish.sh web/app/components/feature/home/HomeHero.vue
```

### Passo 3 — Ler o output e corrigir

```
==> Frontend polish check (sensor 12, v1.12.0)
==> Scanning 47 .vue files in web/app/...

❌ POLISH ISSUES DETECTED (BLOCKING):

  hardcoded_colors:
    web/app/components/feature/home/HomeHero.vue:42 → #ecfdf5
    web/app/components/feature/home/HomeHero.vue:43 → #064e3b
    web/app/components/feature/home/HomeHowItWorks.vue:55 → #10b981

  bem_naming:
    web/app/components/feature/home/HomeHero.vue:20 → .home-hero__title
    web/app/components/feature/home/HomeHero.vue:36 → .home-hero__ctas

  emojis_excessive:
    web/app/components/feature/auth/PersonaPicker.vue:8 → 4 emojis (threshold: 1)

🛑 Action required: corrija os itens acima antes de abrir PR.

Recovery:
  - Cores: use tokens semânticos (color="primary") ou var(--ui-bg-elevated)
  - BEM: use Nuxt UI props ou Tailwind utilities, não classes BEM
  - Emojis: remova ou troque por ícones (icon="i-lucide-...")
  - Comentários: explique POR QUÊ, não O QUÊ (ver code-style.md)
```

Exit codes:
- `0` = sem violação, pode abrir PR
- `1` = violação detectada, **bloquear**
- `2` = erro de uso (path inválido, sem arquivos .vue)

---

## Padrões detectados (10 categorias)

### 1. Cores hex hardcoded (`hardcoded_colors`)

| Padrão regex | Onde | Exemplo |
|---|---|---|
| `#([0-9a-fA-F]{3,8})\b` | `<template>` e `<style>` | `#ecfdf5`, `#064e3b`, `#fff` |
| `rgb\([^)]+\)` | `<style>` | `rgb(236, 253, 245)` |
| `hsl\([^)]+\)` | `<style>` | `hsl(150, 80%, 95%)` |
| `oklch\([^)]+\)` | `<style>` | `oklch(0.97 0.02 145)` |

**Exceções** (não conta como violação):
- `app.config.ts` (aqui é onde os tokens **devem** morar)
- `assets/css/main.css` (CSS-first config do Tailwind v4)
- Comentários (`/* */` ou `//`)
- `var(--ui-*)` (variáveis CSS do Nuxt UI — ok)

### 2. CSS BEM (`bem_naming`)

| Padrão regex | Onde | Exemplo |
|---|---|---|
| `\.[a-z][a-z0-9-]*__[a-z]` | `class="..."` ou `<style>` | `.home-hero__title`, `.card__body` |
| `\.[a-z][a-z0-9-]*--[a-z]` | modificador BEM | `.button--primary` |

**Exceção**: se o projeto é Tailwind-only sem Nuxt UI e a
convenção BEM é justificada (componentes customizados
complexos), whitelist via `package.json`:

```json
{
  "meta-harness": {
    "sensors": {
      "frontend-polish": {
        "whitelist": ["bem_naming"]
      }
    }
  }
}
```

### 3. Comentários redundantes (`redundant_comments`)

| Padrão | Onde |
|---|---|
| `// Foo — top of the page` quando há `<Foo />` no mesmo arquivo | `<script setup>` |
| `/* Bar component */` antes de `<template>` com `<Bar />` | qualquer lugar |

**Heurística**: comentário cujo subject (1ª palavra após
`//` ou `/*`) **bate com o nome do componente exportado** OU
**descreve literalmente o que o código faz** (sem "por quê").

**Não conta** (são úteis):
- `// TODO:` / `// FIXME:` / `// HACK:`
- `// ADR-XXXX: ...`
- `// why: ...` / `// because: ...` / `// rationale: ...`
- `// references #XYZ` (issue tracker)

### 4. Emojis excessivos (`emojis_excessive`)

| Threshold | Arquivo | Bloqueia? |
|---|---|---|
| > 3 emojis | qualquer | sim |
| > 1 emoji | componente "sério" (form, dashboard, list) | sim |
| > 0 emojis em copy que pode ter tom informal (empty state, success message) | empty state, 404 | NÃO (whitelist) |

**Detecção**: regex Unicode emoji (`[\U0001F300-\U0001F9FF\u2600-\u26FF\u2700-\u27BF]`).

**Exceções**:
- 404 / error pages (pode ter tom mais leve)
- Empty states com tom amigável (`Nenhum resultado 🎭`)
- Componentes explicitamente "playful" (configurável)

### 5. Spacing fora da escala (`spacing_off_scale`)

| Padrão | Detecta | Threshold |
|---|---|---|
| `(p\|m\|gap)-\d+` com valor ∈ {3, 5, 7, 9, 10, 11, 13, 14, 15} | `p-3`, `m-5`, `gap-7`, `mt-11` | qualquer ocorrência = bloqueia |

**Permitido**: `1, 2, 4, 6, 8, 12, 16, 20, 24, 32, 40, 48, 64, 80, 96`
(mapeia a 4/8/12/16/20/24/32/40/48/64/80/96 px).

**Exceção**: valores > 96 (e.g., `p-[200px]`) para casos
especiais (full-page hero) — usar bracket notation não
dispara o sensor.

### 6. Texto cru em componentes (`hardcoded_strings`)

| Padrão | Onde | Exemplo |
|---|---|---|
| Strings em PT/EN/ES hardcoded no `<template>` | `<template>` | `<h1>Bem-vindo</h1>` |

**Regra**: usar sempre `{{ $t('chave') }}` (i18n). Exceções:
- `<title>` em `app.head` (config, não template)
- Strings em `definePageMeta` (URL slug)
- Comentários

(Este padrão é parcialmente coberto pelo sensor 08-i18n-audit
mas é re-checado aqui pra garantir consistência visual com i18n.)

### 7. `style="..."` inline com cores (`inline_styles_color`)

| Padrão | Onde |
|---|---|
| `style="[^"]*color:\s*(#\|rgb\|hsl\|oklch)` | `<template>` |

**Regra**: nunca `style="color: #..."`. Vai no
`class="text-primary"` ou `:style="{ color: var(...) }"`.

### 8. Imports de libs fora do stack (`off_stack_imports`)

| Padrão | Detecta | Threshold |
|---|---|---|
| `from ['"]bootstrap` | Vue/React UI | bloqueia (a menos que o stack seja explicitamente Bootstrap) |
| `from ['"]@mui/material` | Nuxt UI project | bloqueia (use Nuxt UI) |
| `from ['"]vuetify` | Nuxt UI project | bloqueia |
| `from ['"]element-plus` | Nuxt UI project | bloqueia |

**Detecção de "Nuxt UI project"**: presença de `@nuxt/ui` em
`package.json`. Se não tem, esses imports são OK.

### 9. Acessibilidade (parcial — `a11y_quick`)

| Padrão | Onde |
|---|---|
| `<img>` sem `alt` | `<template>` |
| `<button>` sem texto acessível (sem texto, sem `aria-label`) | `<template>` |
| `<a>` sem `href` ou sem texto | `<template>` |

(O `quality-assurance` faz a11y completo. Este é só "quick
check" pra pegar os óbvios antes do PR.)

### 10. Falta de design system reference (`no_design_system`)

| Padrão | Detecta | Threshold |
|---|---|---|
| Arquivo `.vue` em `components/` mas o projeto tem `app.config.ts` e o componente não usa **nenhuma** cor/var do Nuxt UI | `app/components/**/*.vue` |

Heurística: arquivo `.vue` em `components/` que tem
`<style scoped>` mas **não** referencia `var(--ui-*)` em
**nenhum** lugar. Indica "componente que estiliza do zero"
em vez de usar o design system.

---

## Quem faz o quê

| Persona | Papel |
|---|---|
| `frontend-engineer` | **Roda este sensor local** antes de PR. Se bloquear, **corrige antes de abrir PR** (não empurra problema pro QA). |
| `quality-assurance` | **Re-roda** no Visual Report (via Playwright screenshot). Reporta violações no `qa/visual-report-<pr>.md`. |
| `team-manager` | **Não roda** este sensor diretamente, mas **bloqueia merge** se o job `frontend-polish` no CI falhar. |
| `solutions-architect` | **Define tokens** em `app.config.ts` e **proíbe anti-patterns** no DoD da feature (link pra esta skill). |

---

## Edge cases

### Componentes de domínio (form, dashboard) com cores

**Bloquear** se hex hardcoded. Use `color="primary"` etc.

### Componentes decorativos (landing hero, illustrations)

**Permitir** gradientes complexos desde que **via tokens**
(e.g., `bg-gradient-to-br from-primary-500 to-primary-700`).
Hex hardcoded em `bg-gradient-...` também bloqueia.

### Componentes third-party copiados (e.g., de shadcn)

**Whitelist** via `package.json` se for copy-paste conhecido.
**Default**: bloquear e pedir refactor (geralmente é 1-2
linhas pra trocar hex por token).

### Dark mode com `dark:` variants

**Permitido** (não é hex). Se o hex for **dentro** de uma
variante `dark:bg-...`, aí sim é problema — vai no
`@theme` e usa `dark:` no token.

### Tokens em `assets/css/main.css` (Tailwind v4 @theme)

**Permitido** (é onde devem ficar). Só bloqueia em
`<style scoped>` de componente.

---

## Auto-fix (opcional, v1.12.0+)

O script tem flag `--suggest-fix` (ainda experimental) que
sugere substituições:

```bash
./harness/scripts/check-frontend-polish.sh --suggest-fix web/app/components/feature/home/HomeHero.vue
```

Output:
```
Suggested fixes for web/app/components/feature/home/HomeHero.vue:

  Line 42:  background: #ecfdf5;
            → background: var(--ui-bg-elevated);
            ⚠ Review: --ui-bg-elevated é o token correto? Se não,
            cadastre em app.config.ts (ui.colors.elevated).

  Line 43:  color: #064e3b;
            → color: var(--ui-text);
            ✓ Token existe (ui.text).

  Line 55:  background: #10b981;
            → background: var(--ui-primary);
            ⚠ Token existe mas é o mesmo que primary-500. Use
            color="primary" via prop ao invés.

Apply suggestions? (y/N)
```

**Não recomendado aplicar automaticamente** (a substituição
pode estar errada). Use o output como guia.

---

## Referências

- **Skill `frontend-public-skills`**: [registry, npx skills, MCP](../skills/frontend-public-skills/SKILL.md)
- **Skill `visual-polish`**: [técnicas visuais](../skills/visual-polish/SKILL.md)
- **Skill `nuxt-ui-patterns` v2.0.0**: [anti-patterns](../skills/nuxt-ui-patterns/SKILL.md#-anti-patterns-sensor-12-detecta-e-bloqueia)
- **Script**: [`harness/scripts/check-frontend-polish.sh`](../scripts/check-frontend-polish.sh)
- **Templates Nuxt UI prontos**: [harness/templates/nuxt-ui/](../templates/nuxt-ui/)
- **Playwright screenshot**: [harness/scripts/visual/playwright-screenshot.mjs](../scripts/visual/playwright-screenshot.mjs)
- **ADR-0022** (esta decisão)
- **AGENTS.md invariante 23** (obrigatoriedade, bloqueante)
