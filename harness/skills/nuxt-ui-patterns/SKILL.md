---
name: nuxt-ui-patterns
version: 1.0.0
type: ui-patterns
stack: nuxt-ui-v3
---

# Nuxt UI v3 — Patterns & Best Practices

Skill for the **frontend-engineer** persona. Applies when building
frontends with **Nuxt UI v3** (project's pin: `@nuxt/ui@3.3.6`).
Inspired by the official [nuxt-ui-templates](https://github.com/nuxt-ui-templates)
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

## ✅ Self-check antes de mergear

Antes de abrir PR de UI, verifique:

- [ ] **Zero modais** para tasks > 30s ou com mais de 2 campos
- [ ] **Breadcrumbs** em todas as páginas 2+ níveis
- [ ] **Tab navigation** funciona (Tab/Shift+Tab cicla, Esc fecha modais)
- [ ] **WCAG AA** — contraste 4.5:1, alt text em imagens, labels em forms
- [ ] **Dark mode** testado (usar `useColorMode()`)
- [ ] **Responsive** — testado em mobile (375px), tablet (768px), desktop (1280px+)
- [ ] **Loading + empty + error states** em cada lista/form
- [ ] **URL state** — filtros, paginação, e item selecionado são refletidos na URL
- [ ] **i18n** — strings extraídas para `i18n/locales/*.json`
