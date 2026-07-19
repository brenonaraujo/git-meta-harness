---
name: nuxt-ui-patterns
version: 2.0.0
type: ui-patterns
stack: nuxt-ui-v4
---

# Nuxt UI v4 — Patterns & Best Practices

Skill for the **frontend-engineer** persona. Applies when building
frontends with **Nuxt UI v4** (project's pin: `@nuxt/ui@^4.10.0`).
Inspired by the official [nuxt-ui-templates](https://github.com/nuxt-ui-templates)

> **v2.0.0 (jul/2026):** atualizado pra Nuxt UI v4 (Tailwind v4
> + Reka UI), adicionado seção "Public Skills Registry", seção
> "Anti-patterns", e self-check expandido. Ver
> [CHANGELOG](../../../CHANGELOG.md) v1.12.0 e ADR-0022.
organization (Dashboard, SaaS, LMS, Minimal) — all MIT-licensed.

## 🚨 Rule #0 — Page first, modal last

> **Default to a dedicated page, not a modal.** Use modals only when
> the user MUST stop what they're doing to make a decision.

Decision tree:

```
User needs to act?
├── Yes, but they need to see the data behind it → Page (with breadcrumb)
├── Yes, but only a quick confirm (< 5s) → Modal OK
│   └── Examples: "Delete this project?", "Confirm upgrade"
├── No, just informing them → Toast / Banner / Inline
└── Long task (> 30s) or multi-step → Page (with breadcrumb)
```

**Never use modals for**:
- Long forms (use a dedicated page with breadcrumb)
- Multi-step flows (use a dedicated page with breadcrumb)
- Tasks where the user needs to compare data on the page behind (use a slideover or dedicated page)
- Tasks that take > 30 seconds
- Primary CRUD operations (create/edit) — use pages

**Use modals for**:
- Destructive confirmations: "Delete this project?", "Revoke access?"
- Login gates: "Sign in to save"
- Upsell at limit: "Storage full — upgrade?"
- Urgent system alerts: "Subscription expires today"

When in doubt: **page + breadcrumb**. Always.

## 📐 Layout — Dashboard structure

Reference: [nuxt-ui-templates/dashboard](https://github.com/nuxt-ui-templates/dashboard)

```vue
<!-- app/pages/dashboard.vue -->
<template>
  <UDashboardPage>
    <UDashboardPanel grow>
      <UDashboardNavbar title="Dashboard">
        <template #right>
          <UButton
            label="Nova ação"
            icon="i-heroicons-plus"
            @click="navigateTo('/dashboard/items/new')"
          />
        </template>
      </UDashboardNavbar>

      <UDashboardPanelContent>
        <BreadcrumbHome class="mb-4" />  <!-- always -->

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
          <UCard v-for="stat in stats" :key="stat.label">
            <template #header>
              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-500">{{ stat.label }}</span>
                <UIcon :name="stat.icon" class="text-gray-400" />
              </div>
            </template>
            <span class="text-2xl font-bold">{{ stat.value }}</span>
          </UCard>
        </div>

        <UTable
          :columns="columns"
          :rows="items"
          :loading="pending"
          :pagination="pagination"
          @select="(row) => navigateTo(`/dashboard/items/${row.id}`)"
        />
      </UDashboardPanelContent>
    </UDashboardPanel>
  </UDashboardPage>
</template>
```

## 🍞 Breadcrumbs — always

> **Every page that is 2+ levels deep must have breadcrumbs.**

Reference: [UX best practices for breadcrumbs](https://www.eleken.co/blog-posts/breadcrumbs-ux)
and [accessibility](https://accessiblyapp.com/blog/breadcrumbs-accessibility/).

```vue
<!-- app/components/BreadcrumbHome.vue -->
<script setup lang="ts">
const route = useRoute()
const links = computed(() => {
  const segments = route.path.split('/').filter(Boolean)
  return segments.map((seg, i) => ({
    label: humanize(seg),
    to: '/' + segments.slice(0, i + 1).join('/')
  }))
})
</script>

<template>
  <nav aria-label="breadcrumb" class="text-sm">
    <ol class="flex items-center gap-2 text-gray-500">
      <li>
        <NuxtLink to="/" class="hover:text-primary">Início</NuxtLink>
      </li>
      <template v-for="(link, i) in links" :key="i">
        <li aria-hidden="true">/</li>
        <li>
          <NuxtLink :to="link.to" class="hover:text-primary">
            {{ link.label }}
          </NuxtLink>
        </li>
      </template>
    </ol>
  </nav>
</template>
```

Best practices:
- `<nav aria-label="breadcrumb">` + `<ol>` (semantic HTML)
- Place **above H1** (then key controls like filters)
- Quiet visual: small text, separator chevrons
- Tap targets ≥ 44x44px (mobile)
- WCAG AA contrast (4.5:1)
- Don't show the current page if it duplicates the H1 directly below

## 📊 Data tables — `UTable` (don't roll your own)

```vue
<UTable
  :columns="columns"
  :rows="rows"
  :loading="pending"
  :pagination="pagination"
  :sort="sort"
  @select="(row) => navigateTo(`/items/${row.id}`)"
>
  <template #status-data="{ row }">
    <UBadge :color="row.status === 'active' ? 'green' : 'gray'">
      {{ row.status }}
    </UBadge>
  </template>
  <template #actions-data="{ row }">
    <UDropdown :items="actionsFor(row)">
      <UButton icon="i-heroicons-ellipsis-vertical" variant="ghost" />
    </UDropdown>
  </template>
</UTable>
```

Features built-in: sorting, filtering, pagination, row selection,
column-specific render slots. **Don't reimplement** — use what `UTable` gives you.

## ⌘ Command palette (Cmd+K) — `UCommandPalette`

Reference: [nuxt-ui-templates/dashboard](https://github.com/nuxt-ui-templates/dashboard)
includes a `commandPalette.vue` that uses `UCommandPalette`.

```vue
<!-- app/components/CommandPalette.vue -->
<UCommandPalette
  v-model="isOpen"
  :groups="[
    { key: 'actions', label: 'Ações', commands: actions },
    { key: 'pages', label: 'Páginas', commands: pages },
  ]"
  @update:model-value="onSelect"
/>
```

Use the dashboard template's `commandPalette.vue` as the reference
implementation — copy and adapt, don't rewrite from scratch.

## 📝 Forms — `UForm` + `UFormField`

```vue
<UForm :schema="schema" :state="state" @submit="onSubmit">
  <UFormGroup label="Nome" name="name" required>
    <UInput v-model="state.name" />
  </UFormGroup>
  <UFormGroup label="Email" name="email" required>
    <UInput v-model="state.email" type="email" />
  </UFormGroup>
  <UButton type="submit" label="Salvar" />
</UForm>
```

- Use **inline validation** (under each field, not in a toast)
- Show **errors after first blur** (not on every keystroke)
- Always have a **clear primary action** ("Salvar", "Criar")
- Always have an **explicit secondary** ("Cancelar")
- Use a **dedicated page**, not a modal, for create/edit forms

## 🪟 Slideover vs Modal

For non-destructive side panels (e.g., "edit this item while
seeing the list"), prefer `USlideover` over `UModal`:

```vue
<USlideover v-model="isOpen" :ui="{ width: 'w-full max-w-md' }">
  <!-- form, details, etc. -->
</USlideover>
```

Slideover:
- Preserves context (background is dimmed but visible)
- Better for: filters, item details, "quick edit"
- Modal is still appropriate for: destructive confirms, login gates

## 📚 Templates de referência (sempre consulte)

| Template | Quando usar |
|---|---|
| [nuxt-ui-templates/dashboard](https://github.com/nuxt-ui-templates/dashboard) | Admin panels, listagens, formulários com sidebar colapsável, ⌘K command palette |
| [nuxt-ui-templates/saas](https://github.com/nuxt-ui-templates/saas) | Landing, pricing, docs, blog (powered by Nuxt Content) |
| [nuxt-ui-templates/lms](https://github.com/nuxt-ui-templates/lms) | Learning management, course pages, progress tracking |
| [nuxt-ui-templates/minimal](https://github.com/nuxt-ui-templates/minimal) | Boilerplate limpo, ESLint configurado |

Para implementar um dashboard, **comece sempre pelo template oficial**:

```bash
npx nuxi@latest init my-app -t github:nuxt-ui-templates/dashboard
```

Depois adapte: renomeie `app-` → `<seu-projeto>-`, ajuste as cores
em `app.config.ts`, etc.

---

## 🌐 Public Skills Registry (v1.12.0+)

> **Lição do Mandaí v2 (jul/2026):** o `frontend-engineer`
> entregou UI com cores hardcoded e zero uso de skills
> públicas. O registry existe — é só consultar.

### Por que usar o registry

Skills públicas (Vercel `vercel-labs/skills`, [skills.sh](https://www.skills.sh))
são mantidas **pelos criadores do framework** e têm:

- **Versionamento compatível** com a versão pinada.
- **Exemplos auditados** (Socket + Snyk + Gen Agent Trust Hub).
- **Padrões de produção** (não invenção de quem usou 1x).

### Workflow pré-implementação (v1.12.0, regra não-violável)

```bash
# 1. Identificar stack (já tem @nuxt/ui no package.json?)
grep "@nuxt/ui" web/package.json

# 2. Consultar registry
npx skills find nuxt-ui

# 3. Instalar a skill oficial (mantida pela Nuxt team)
npx skills add nuxt/ui@nuxt-ui

# 4. (Opcional) Setup do MCP — agent ganha API de componentes
claude mcp add --transport http nuxt-ui https://ui.nuxt.com/mcp

# 5. AGORA SIM implementar — já com a skill oficial no contexto
```

### Skills Nuxt UI no registry (top 3, jul/2026)

| Skill | Installs | Manutenção |
|---|---|---|
| `nuxt/ui@nuxt-ui` | 15.2K | **Nuxt team** (oficial) |
| `onmax/nuxt-skills@reka-ui` | 6.6K | Comunidade (foco Reka UI) |
| `onmax/nuxt-skills@nuxt-ui` | 6.1K | Comunidade (geral) |

Ver skill
[`frontend-public-skills`](../frontend-public-skills/SKILL.md)
para a lista completa curada.

---

## 🚨 Anti-patterns (sensor 12 detecta e BLOQUEIA)

> **v1.12.0:** o `frontend-engineer` é bloqueado pelo
> `check-frontend-polish.sh` se qualquer um desses padrões
> aparecer. **Corrija antes de abrir PR.**

### 1. Cores hex hardcoded em `<template>` ou `<style>`

❌ **Errado** (o que o Mandaí v2 PR #23 tinha):

```vue
<style scoped>
.hero {
  background: linear-gradient(180deg, #ecfdf5 0%, #ffffff 100%);
}
.title {
  color: #064e3b;  <!-- ❌ hex hardcoded -->
}
</style>
```

✅ **Correto** (sempre via tokens semânticos):

```vue
<style scoped>
.hero {
  background: linear-gradient(180deg, var(--ui-bg-elevated), var(--ui-bg));
}
.title {
  color: var(--ui-text);
}
</style>
```

Ou melhor, sem `<style scoped>` — use **Nuxt UI props**:

```vue
<UButton color="primary" size="lg">Salvar</UButton>
```

**Regra:** o token semântico vive em `app.config.ts` (e.g.,
`primary: 'green'`). Componentes usam `color="primary"`, nunca
o hex literal. Sensor 12 detecta `#abcdef` em `.vue`/`.css`
(exceto em `app.config.ts`).

### 2. CSS BEM misturado com Tailwind/Nuxt UI

❌ **Errado** (mistura confusa):

```vue
<button class="home-hero__cta bg-primary-500">Salvar</button>
```

✅ **Correto** (Nuxt UI props + Tailwind utilities):

```vue
<UButton color="primary" size="lg" block>Salvar</UButton>
```

**Regra:** se você está estilizando com classes BEM (`.foo__bar`)
E Tailwind ao mesmo tempo, está errado. Escolhe **um**:
Nuxt UI props (preferido) OU Tailwind utilities OU CSS modules.

### 3. Comentários redundantes explicando o que o código faz

❌ **Errado** (comentário explica o óbvio):

```vue
<script setup lang="ts">
// HomeHero — top of the public landing page. Carries the one-liner
// tagline and the two primary CTAs (Entrar / Criar conta). Mobile-first.
</script>
```

✅ **Correto** (comentário explica **por que** ou **restrições**,
não **o que**):

```vue
<script setup lang="ts">
// CTAs land on /auth/{login,register}; preselect role from query
// (?role=leader) for the signup flow (#16).
</script>
```

**Regra:** o `frontend-engineer` está no `code-style.md` que
proíbe comentários redundantes. Sensor 12 detecta comentários
que repetem literalmente o nome do componente/função.

### 4. Emojis excessivos

❌ **Errado** (UI cheia de emojis decorativos):

```vue
<h1>🎉 Bem-vindo ao Mandaí! 💚</h1>
<p>🚀 Sua praça compra junto ✨</p>
```

✅ **Correto** (ícones ou zero):

```vue
<h1>Bem-vindo ao Mandaí</h1>
<p>Sua praça compra junto</p>
```

**Regra:** emojis em UI são permitidos **só** quando fazem
função semântica (e.g., 🎉 para "conquista") OU quando o
projeto explicitamente pede tom informal. Sensor 12 detecta
> 3 emojis por arquivo `.vue` OU > 1 emoji no `<template>` de
componente sério (form, dashboard, etc).

Use **ícones** (Nuxt UI tem `lucide` collection:
`icon="i-lucide-shopping-cart"`).

### 5. Spacing fora da escala

❌ **Errado** (valores aleatórios):

```vue
<div class="p-3 mt-7 gap-5">
```

✅ **Correto** (escala 4/8/12/16/24/32):

```vue
<div class="p-4 mt-8 gap-4">
```

**Regra:** use apenas `1, 2, 4, 6, 8, 12, 16, 24` (4/8/12/16/24/32/48/64/96px).
**Nunca** `3, 5, 7, 9, 10, 11, 13, 14, 15`.

---

## ✅ Self-check antes de mergear

Antes de abrir PR de UI, verifique:

- [ ] **`npx skills find nuxt-ui` foi rodado** ANTES de
      implementar (regra não-violável, v1.12.0)
- [ ] **Pelo menos 1 skill Nuxt UI** está instalada
      (`npx skills add nuxt/ui@nuxt-ui` é o mínimo)
- [ ] **Zero cores hex hardcoded** em `.vue`/`.css` (tudo via
      tokens semânticos do `app.config.ts`)
- [ ] **Zero modais** para tasks > 30s ou com mais de 2 campos
- [ ] **Breadcrumbs** em todas as páginas 2+ níveis
- [ ] **Tab navigation** funciona (Tab/Shift+Tab cicla, Esc fecha modais)
- [ ] **WCAG AA** — contraste 4.5:1, alt text em imagens, labels em forms
- [ ] **Dark mode** testado (usar `useColorMode()`)
- [ ] **Responsive** — testado em mobile (375px), tablet (768px), desktop (1280px+)
- [ ] **Loading + empty + error states** em cada lista/form
- [ ] **URL state** — filtros, paginação, e item selecionado são refletidos na URL
- [ ] **Emojis** ≤ 1 por componente sério (ou justificado)
- [ ] **Spacing scale** consistente (4/8/12/16/24/32)
- [ ] **Screenshot local** gerado com Playwright (ver
      `harness/scripts/visual/playwright-screenshot.mjs`)
- [ ] **Sensor 12 `frontend-polish`** roda verde (não tem
      hex hardcoded, BEM, comentários redundantes)
- [ ] **i18n** — strings extraídas para `i18n/locales/*.json`
