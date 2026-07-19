---
name: tailwind-only-patterns
version: 1.0.0
type: ui-patterns
stack: tailwind-v4
---

# Tailwind v4 (sem Nuxt UI) — Patterns & Components

Skill for the **frontend-engineer** persona. **Applies when the
project uses Tailwind v4 WITHOUT `@nuxt/ui`** — i.e., plain Vue
3 + Vite, or React + Vite, or any setup where the UI primitives
come from another source (shadcn-vue, PrimeVue, Reka UI
standalone, headlessui, etc).

> **Quando usar esta skill:** o `package.json` NÃO tem
> `@nuxt/ui` como dependência. Se tem, use a skill
> `nuxt-ui-patterns` (vai ter Nuxt UI pronto e é mais
> produtivo).

> **Lição do Mandaí v2 (jul/2026):** o frontend-engineer
> entregou CSS BEM + hex hardcoded em vez de usar Tailwind
> utilities + tokens semânticos. Esta skill documenta o
> caminho "Tailwind puro" com qualidade profissional.

---

## 🚨 Rule #0 — Tailwind v4 + tokens semânticos, sempre

> **Default:** Tailwind v4 com **CSS-first config** + tokens
> semânticos (`bg-primary`, `text-foreground`,
> `border-default`). Zero hex hardcoded em componentes.

```css
/* app/assets/css/main.css — Tailwind v4 CSS-first */
@import "tailwindcss";

@theme {
  --color-primary-50:  oklch(0.97 0.02 145);
  --color-primary-500: oklch(0.65 0.20 145);
  --color-primary-900: oklch(0.30 0.10 145);
  /* … */
  --font-sans: "Inter", system-ui, sans-serif;
  --radius-card: 0.75rem;
}
```

Componentes:

```vue
<button class="bg-primary-500 text-white px-4 py-2 rounded-card
  hover:bg-primary-600 focus:ring-2 focus:ring-primary-300">
  Salvar
</button>
```

**NUNCA** use hex (`#abc123`) ou `rgb()` em componentes. Vai
no `main.css` via `@theme`, e os componentes usam o token.

---

## 📐 Layout patterns

### Container responsivo (auto-padding)

```vue
<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
  <h1 class="text-3xl font-bold tracking-tight">Dashboard</h1>
</div>
```

### Grid responsivo

```vue
<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
  <div v-for="item in items" :key="item.id"
       class="rounded-card border border-default bg-elevated p-6">
    {{ item.title }}
  </div>
</div>
```

### Stack (vertical) com gap consistente

```vue
<div class="flex flex-col gap-4">  <!-- 1rem = 16px -->
  <h2>Title</h2>
  <p>Body</p>
  <button>CTA</button>
</div>
```

Use a **spacing scale do Tailwind v4**: `gap-2` (8px), `gap-4`
(16px), `gap-6` (24px), `gap-8` (32px). **Nunca** `gap-3` ou
`gap-5` (não está na escala).

---

## 🧩 Componentes — escolha a base certa

### Decisão tree

```
Precisa de componentes prontos?
├── SIM, copy-paste + customizable
│   ├── Vue → shadcn-vue (https://www.shadcn-vue.com)
│   ├── React → shadcn/ui (https://ui.shadcn.com)
│   └── Customizado → Reka UI standalone (https://reka-ui.com)
├── SIM, library completa
│   ├── Vue → PrimeVue, Naive UI, Element Plus
│   └── React → Radix UI, MUI, Chakra
└── NÃO, só estilo
    └── Headless UI (https://headlessui.com) — sem estilo, só comportamento
```

### shadcn-vue (recomendado pra Vue 3 + Vite)

```bash
# Setup
pnpm dlx shadcn-vue@latest init

# Adicionar componentes
pnpm dlx shadcn-vue@latest add button card input
```

Componentes são **seus** (copiados pra `components/ui/`,
customizáveis). Tailwind v4 + Reka UI + CVA (class-variance-authority)
por baixo. **Não é uma dependência runtime.**

```vue
<!-- components/ui/button/Button.vue (gerado pelo shadcn-vue) -->
<template>
  <button :class="cn(buttonVariants({ variant, size }))">
    <slot />
  </button>
</template>
```

---

## 🎨 Theming e dark mode

### Dark mode com variante `dark:`

```vue
<div class="bg-white dark:bg-neutral-900 text-neutral-900
            dark:text-neutral-100">
```

### Variantes custom (cva)

```ts
// lib/utils.ts
import { cva, type VariantProps } from "class-variance-authority"

export const buttonVariants = cva(
  "inline-flex items-center justify-center rounded-md font-medium transition-colors focus:outline-none focus:ring-2 disabled:opacity-50",
  {
    variants: {
      variant: {
        primary:   "bg-primary-500 text-white hover:bg-primary-600",
        secondary: "bg-neutral-100 text-neutral-900 hover:bg-neutral-200",
        ghost:     "hover:bg-neutral-100 text-neutral-700",
        danger:    "bg-error-500 text-white hover:bg-error-600",
      },
      size: {
        sm: "h-8 px-3 text-sm",
        md: "h-10 px-4 text-base",
        lg: "h-12 px-6 text-lg",
      },
    },
    defaultVariants: { variant: "primary", size: "md" },
  },
)

export type ButtonProps = VariantProps<typeof buttonVariants>
```

### Tokens semânticos (vai no `main.css` via `@theme`)

```css
@theme {
  --color-bg:        oklch(1 0 0);          /* white */
  --color-bg-elevated: oklch(0.98 0 0);     /* card bg */
  --color-fg:        oklch(0.20 0 0);       /* near-black */
  --color-fg-muted:  oklch(0.50 0 0);
  --color-border:    oklch(0.90 0 0);
  --color-primary:   oklch(0.55 0.20 145);  /* brand green */
  --color-error:     oklch(0.55 0.22 25);
}
```

Componentes usam `bg-bg-elevated text-fg border-border`. Zero
hex no `<template>`.

---

## 🍞 Breadcrumb (padrão)

```vue
<nav aria-label="breadcrumb" class="text-sm">
  <ol class="flex items-center gap-2 text-fg-muted">
    <li>
      <NuxtLink to="/" class="hover:text-primary">Início</NuxtLink>
    </li>
    <li aria-hidden="true">/</li>
    <li>
      <NuxtLink to="/produtos" class="hover:text-primary">Produtos</NuxtLink>
    </li>
    <li aria-hidden="true">/</li>
    <li aria-current="page" class="text-fg">Maçãs</li>
  </ol>
</nav>
```

`<nav aria-label="breadcrumb">` + `<ol>` é **semântica**, não
opcional. Screen readers usam isso pra navegar.

---

## 📝 Forms

```vue
<form @submit.prevent="onSubmit" class="flex flex-col gap-4 max-w-md">
  <label class="flex flex-col gap-1">
    <span class="text-sm font-medium text-fg">Nome</span>
    <input type="text" required
           class="rounded-md border border-border bg-bg px-3 py-2
                  focus:border-primary focus:ring-1 focus:ring-primary" />
  </label>
  <button type="submit" class="rounded-md bg-primary text-white
                                 py-2 hover:bg-primary-600">
    Salvar
  </button>
</form>
```

- **Labels acima dos inputs** (mobile-first, melhor pra
  touchscreen).
- **Erros embaixo do input** (não em toast — inline é mais
  acessível).
- **Validação depois do primeiro blur**, não a cada keystroke.

---

## 📊 Data tables

Sem Nuxt UI, use **TanStack Table** (Vue/React agnostic,
headless) + Tailwind pra estilo:

```ts
import { useVueTable, getCoreRowModel, FlexRender } from "@tanstack/vue-table"
```

Ou, pra CRUD simples, use **shadcn-vue** `<DataTable>` (já
vem com sorting/filtering/pagination).

---

## 🚨 Anti-patterns

❌ **Hex hardcoded em componentes** — vai no `@theme` (`main.css`).
   Sensor 12 detecta isso.

❌ **`@apply` em excesso** — vira CSS disfarçado. Prefira utility
   classes inline no `class=""`. `@apply` só pra 1+ utilities
   que se repetem em 3+ lugares.

❌ **`space-y-X` em vez de `flex flex-col gap-X`** — `gap` é mais
   previsível, `space-y` quebra com elementos `display:inline`.

❌ **Misturar BEM e Tailwind** (`.foo__bar text-sm`) — escolhe
   um. Se Tailwind, fica só Tailwind.

❌ **Inline styles com `style="color: #..."`** — vai no `@theme`
   ou usa `class="text-fg"`.

❌ **CSS-in-JS (styled-components, emotion)** sem necessidade
   — Tailwind v4 já é "atomic CSS" moderno, não duplica stack.

---

## Self-check antes de PR

- [ ] **Zero hex** em `<template>` ou `<style scoped>` (tudo via
      `@theme` tokens).
- [ ] **Spacing scale consistente** (`gap-2/4/6/8/12/16`, não
      `gap-3`/`gap-5`).
- [ ] **Dark mode testado** (`dark:` variants aplicadas onde
      faz sentido).
- [ ] **Componentes** vêm de shadcn-vue (ou escolha similar),
      não inventados do zero.
- [ ] **Breadcrumbs** em páginas 2+ níveis.
- [ ] **Acessibilidade**: `aria-label`, `aria-current`, tab
      navigation funcional.
- [ ] **Responsive**: 375px, 768px, 1280px+ testados.
- [ ] **Loading + empty + error states** em listas/forms.

---

## Referências

- **Tailwind v4 docs**: [https://tailwindcss.com/docs](https://tailwindcss.com/docs)
- **shadcn-vue**: [https://www.shadcn-vue.com](https://www.shadcn-vue.com)
- **Reka UI**: [https://reka-ui.com](https://reka-ui.com)
- **Class Variance Authority**: [https://cva.style](https://cva.style)
- **Headless UI**: [https://headlessui.com](https://headlessui.com)
- **Skill pública oficial**: `lombiq/tailwind-agent-skills@tailwind-4-docs`
  (instalar via `npx skills add`)
- **Skill pública design system**: `wshobson/agents@tailwind-design-system`
- **Skill interna `frontend-public-skills`**: [../frontend-public-skills/SKILL.md](../frontend-public-skills/SKILL.md)
- **Skill interna `visual-polish`**: [../visual-polish/SKILL.md](../visual-polish/SKILL.md)
- **Skill interna `ux-design-best-practices`**: [../ux-design-best-practices/SKILL.md](../ux-design-best-practices/SKILL.md)
- **Sensor 12 `frontend-polish`**: [../../sensors/12-frontend-polish.md](../../sensors/12-frontend-polish.md)
