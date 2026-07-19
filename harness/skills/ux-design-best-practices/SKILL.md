---
name: ux-design-best-practices
version: 1.0.0
type: design
applies-to: all-ui
---

# UX Design — Best Practices

Skill for the **frontend-engineer** and **domain-expert** personas.
Applies to **any UI work**, regardless of framework (Nuxt, React, Vue,
mobile, etc.). Companion to the `nuxt-ui-patterns` skill (which is
Nuxt-specific).

## 🚨 Rule #0 — Page first, modal last

> **Default to a dedicated page. Use a modal only when the user MUST
> stop what they're doing to make a decision.**

This rule is **stack-agnostic**. It applies whether you're building
in Nuxt, React, Vue, Flutter, or any other framework.

### When to use each pattern

| Pattern | When | Examples |
|---|---|---|
| **Page** (default) | Task takes > 30s, multi-step, user needs context | Create/edit, settings, long forms |
| **Modal** | Quick confirm (< 5s), destructive, user must stop | "Delete this project?", "Confirm upgrade" |
| **Slideover** (right panel) | Edit while seeing context | Item edit while list is visible, filters |
| **Drawer** (left panel) | Navigation, settings on demand | Sidebar menu, filter panel |
| **Toast / Banner** | Informing without blocking | "Saved successfully", "New version available" |
| **Tooltip / Popover** | Contextual help on hover | Definition of a term, hint icon |
| **Inline** | Form validation, status | Field-level error, success indicator |

### The modal decision tree

```
User needs to act?
│
├── 1. Does the user need context of the current screen?
│   ├── YES → Page (NOT modal) — modals hide context
│   └── NO ↓
│
├── 2. Is the task short (< 30s) and self-contained?
│   ├── NO → Page (NOT modal) — multi-step needs URL state
│   └── YES ↓
│
├── 3. Is the action destructive or irreversible?
│   ├── YES → Modal OK (confirmation)
│   └── NO → Page (NOT modal) — non-destructive actions don't need a halt
│
└── 4. Is the action user-initiated?
    ├── YES → Modal OK (e.g., "Sign in to continue")
    └── NO → Toast/Banner (system-initiated, not blocking)
```

**When in doubt: page.** Always.

References:
- [userpilot — Modal UX Design for SaaS in 2026](https://userpilot.com/blog/modal-ux-design/)
- [uiuxatlas — Modals, Dialogs & Overlays](https://uiuxatlas.com/lessons/ui-patterns/modals-dialogs-and-overlays/)

## 🍞 Breadcrumbs — always on 2+ levels

> **Every page 2+ levels deep must have a visible breadcrumb trail.**
> Place above the H1.

### Why breadcrumbs matter

- **Orientation** — users always know where they are
- **Escape hatch** — quick way back to a parent section
- **Discoverability** — reveals the IA (information architecture)
- **Accessibility (a11y)** — assistive tech uses breadcrumbs for navigation

### The right structure (semantic HTML)

```html
<nav aria-label="breadcrumb">
  <ol>
    <li><a href="/">Início</a></li>
    <li aria-hidden="true">/</li>
    <li><a href="/dashboard">Dashboard</a></li>
    <li aria-hidden="true">/</li>
    <li><a href="/dashboard/items">Itens</a></li>
    <li aria-hidden="true">/</li>
    <li aria-current="page">Item 42</li>
  </ol>
</nav>
```

- `<nav aria-label="breadcrumb">` — tells screen readers what it is
- `<ol>` — ordered list (the path is sequential)
- `aria-current="page"` on the last item
- `aria-hidden="true"` on the separator (so AT doesn't read it)

### When to show the current page

- **Show it** if depth is 3+ levels, or the page title is far from the breadcrumb, or accessibility is a priority.
- **Omit it** if it duplicates the H1 directly below (visual noise), or screen space is tight.

References:
- [Eleken — UX Breadcrumbs in 2026](https://www.eleken.co/blog-posts/breadcrumbs-ux/)
- [Accessibly App — Breadcrumbs Accessibility](https://accessiblyapp.com/blog/breadcrumbs-accessibility/)

## 📝 Forms

### Inline validation
- Show errors **below each field** (not in a toast at the top)
- Validate on **blur** (after user leaves the field), not on every keystroke
- Use **specific error messages** ("Email deve ter formato válido" not "Inválido")
- Mark required fields visually (asterisk + `aria-required="true"`)

### Primary action
- Use a **single, specific, action-focused label**: "Salvar alterações", "Criar projeto", "Enviar mensagem"
- **Never** use "OK", "Yes", "Submit" — they're not specific
- Place primary action **right-aligned** at the bottom (Western reading order)
- Always have an explicit **secondary** action: "Cancelar" (link or ghost button)

### Multi-step forms
- Use a **dedicated page per step** with breadcrumbs showing progress
- Or a **single page with tabs** (`UTabs`, `UStepper`)
- **Never** use a modal for multi-step — losing context between steps is bad

## ♿ Accessibility (WCAG AA minimum)

Non-negotiable. All UI must meet:

| Criterion | Requirement |
|---|---|
| **Color contrast** | 4.5:1 (text), 3:1 (large text, UI components) |
| **Keyboard nav** | All interactive elements reachable via Tab, in logical order |
| **Focus visible** | Always show focus ring (don't `outline: none` without replacement) |
| **Alt text** | All `<img>` have `alt=""` (decorative) or descriptive text |
| **Form labels** | Every input has a `<label>` (or `aria-label`/`aria-labelledby`) |
| **Error association** | Input errors linked via `aria-describedby` |
| **Modal focus** | On open: focus moves to first focusable. On close: focus returns to trigger |
| **Modal Esc** | Escape key always closes (and Tab cycles within, doesn't leak out) |
| **Color independence** | Don't rely on color alone to convey meaning (also use icons/text) |

## 📱 Responsive design

### Mobile-first breakpoints

```css
/* Default = mobile */
.element { font-size: 14px; padding: 8px; }

/* Tablet ≥ 768px */
@media (min-width: 768px) { .element { font-size: 16px; padding: 12px; } }

/* Desktop ≥ 1024px */
@media (min-width: 1024px) { .element { font-size: 16px; padding: 16px; } }
```

### Tap targets
- **Minimum 44x44px** for all interactive elements (Apple HIG)
- **Minimum 48x48dp** for Android (Material)
- Add padding if the visual element is smaller

### Touch vs mouse
- Don't rely on hover for information (touch devices don't have hover)
- Tooltips on desktop OK; on mobile, use expand-on-tap or always-visible info

## 🔄 Loading, empty, error states

Every list/form must handle all 3:

| State | Visual | Code |
|---|---|---|
| **Loading** | Skeleton or spinner | `:loading="pending"` (Nuxt UI) |
| **Empty** | Illustration + "Nada por aqui ainda" + CTA | `<UEmpty>` (Nuxt UI) |
| **Error** | Red border + "Erro ao carregar" + retry | `<UAlert color="red">` |

## 🌐 i18n (internationalization)

- **All UI strings** in `i18n/locales/{en,pt-BR,es}.json`
- **Never** hardcode user-facing strings in `.vue` files
- Use `$t('key')` or `<i18n-t>` for translation
- Date/number formatting via `Intl.DateTimeFormat` / `Intl.NumberFormat`
- Right-to-left (RTL) support: avoid hardcoded `text-left`/`text-right`, use `text-start`/`text-end` (Tailwind logical properties)

## 🎨 Visual hierarchy

- **H1** = page title (one per page, semantic)
- **H2** = section title (2-4 per page)
- **H3** = subsection / card title
- **body** = paragraph text
- **caption** = metadata (timestamps, counts)
- Color contrast: H1/H2 should be 7:1 (AAA), body 4.5:1 (AA)
- Whitespace > lines > colors for grouping (Gestalt principles)

## ✅ Pre-merge self-check (any UI change)

- [ ] **Zero modais** for tasks > 30s or with > 2 fields
- [ ] **Breadcrumbs** on all pages 2+ levels deep
- [ ] **Tab navigation** works, focus visible
- [ ] **Esc** closes modals, focus returns to trigger
- [ ] **WCAG AA**: contrast 4.5:1, alt text, labels, errors linked
- [ ] **Dark mode** tested
- [ ] **Responsive**: 375px, 768px, 1280px+ all look right
- [ ] **Loading + empty + error states** on every list/form
- [ ] **URL state** for filters, pagination, selected item
- [ ] **i18n**: strings in `locales/*.json`, no hardcoded text
- [ ] **Tap targets** ≥ 44x44px
- [ ] **Color-independent** (don't rely on color alone)
