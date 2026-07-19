---
name: visual-polish
version: 1.0.0
type: design
applies-to: all-ui
---

# Visual Polish — Hierarchy, Whitespace, Contrast, Consistency

Skill for the **frontend-engineer** persona. **Stack-agnostic**
(Nuxt UI, Tailwind-only, React, Vue, mobile — todos). Foco nas
**técnicas** que separam "código que funciona" de "código que
parece profissional".

> **Lição do Mandaí v2 (jul/2026):** o frontend-engineer
> entregou UI com cores hardcoded, espaçamento aleatório,
> hierarquia visual quebrada, e zero consistência entre
> telas. Esta skill é o **mínimo profissional** que toda
> entrega de UI precisa atender.

---

## 🚨 Rule #0 — Visual polish is a feature, not a polish step

> **Polish visual faz parte da entrega de feature, não é
> "depois eu ajusto".** Se a primeira renderização parece
> amadora, o usuário final **já vai sentir** — não importa
> se o backend está perfeito.

O polish começa ANTES da primeira linha de código:
1. **Decidir tokens** (cores, espaçamento, tipografia, raio).
2. **Aplicar escala consistente** (spacing scale 4/8/12/16/24/32).
3. **Validar contraste** (WCAG AA: 4.5:1 texto, 3:1 UI).
4. **Screenshot local** (Playwright + Chromium) **antes** do PR.

---

## 1. Hierarchy — quem é o quê?

A página deve responder "o que é o título? o que é seção?
o que é body?" em **< 1 segundo**.

### Escala tipográfica (Tailwind v4 / Nuxt UI)

| Nível | Tailwind | Tamanho | Line-height | Peso |
|---|---|---|---|---|
| **H1 (page title)** | `text-4xl sm:text-5xl` | 36-48px | 1.1-1.2 | `font-bold` (700) |
| **H2 (section)** | `text-2xl sm:text-3xl` | 24-30px | 1.2-1.3 | `font-semibold` (600) |
| **H3 (subsection)** | `text-xl` | 20px | 1.3-1.4 | `font-semibold` (600) |
| **Body** | `text-base` | 16px | 1.5-1.6 | `font-normal` (400) |
| **Small / caption** | `text-sm` | 14px | 1.4-1.5 | `font-normal` (400) |

**Regra prática**: **cada nível é 1.2-1.5× maior que o
anterior** (modular scale ratio). Não invente (ex.: H1 32px,
H2 28px — H2 quase do tamanho de H1, quebra hierarchy).

### ❌ Anti-pattern

```vue
<!-- ERRADO: tudo text-base, sem hierarquia -->
<h1>Dashboard</h1>
<h2>Receita</h2>
<p>R$ 12.500</p>
<h2>Pedidos</h2>
<p>124</p>
```

### ✅ Correto

```vue
<h1 class="text-3xl font-bold tracking-tight">Dashboard</h1>
<section>
  <h2 class="text-xl font-semibold mb-2">Receita</h2>
  <p class="text-2xl font-semibold text-fg">R$ 12.500</p>
</section>
<section>
  <h2 class="text-xl font-semibold mb-2">Pedidos</h2>
  <p class="text-2xl font-semibold text-fg">124</p>
</section>
```

---

## 2. Whitespace — espaço é luxo

> **Whitespace é a primeira coisa que separa amador de
> profissional.** Apps cheios parecem "W3Schools 2018";
> apps com whitespace parecem "Linear/Notion/Vercel".

### Spacing scale (sempre)

Use **apenas** estes valores (em px):

| Token | px | Uso típico |
|---|---|---|
| `1` | 4px | Micro-ajustes (gap entre ícone e label) |
| `2` | 8px | Gap entre elementos tight (lista de pills) |
| `3` | 12px | — (evite — não está na escala) |
| `4` | 16px | **Default** — gap entre blocos relacionados |
| `6` | 24px | Gap entre seções |
| `8` | 32px | Margins entre blocos distintos |
| `12` | 48px | Margins top/bottom de página |
| `16` | 64px | Hero padding top/bottom |
| `24` | 96px | Sections "respira" (landing page) |

**NUNCA** use valores fora da escala (`gap-5`, `gap-7`,
`p-9`, `mt-11`). A consistência visual **depende** da
escala.

### Regra dos 8

> **Tudo** deve ser múltiplo de 4 (idealmente 8) para
> alinhamento com a grade de pixels. Sub-pixel é feio.

```vue
<!-- ❌ ERRADO: valores aleatórios -->
<div class="p-3 mt-7 gap-5">

<!-- ✅ CORRETO: scale consistente -->
<div class="p-4 mt-8 gap-4">
```

### Whitespace em formulários

- **Label → input**: 4-8px (`gap-1` ou `gap-2`).
- **Input → helper text / erro**: 4-8px.
- **Input → próximo input**: 16-24px (`gap-4` ou `gap-6`).
- **Form → próximo bloco**: 32-48px (`mt-8` ou `mt-12`).

```vue
<form class="flex flex-col gap-6 max-w-md">
  <label class="flex flex-col gap-1">
    <span class="text-sm font-medium">Email</span>
    <input class="rounded-md border px-3 py-2" />
  </label>
  <label class="flex flex-col gap-1">
    <span class="text-sm font-medium">Senha</span>
    <input type="password" class="rounded-md border px-3 py-2" />
  </label>
  <button class="rounded-md bg-primary py-2 mt-2">Entrar</button>
</form>
```

---

## 3. Contrast — WCAG AA, sempre

> **Texto**: contraste **≥ 4.5:1** com o fundo.
> **UI components** (borda, ícone): **≥ 3:1**.
> **Não-negociável** (WCAG AA).

### Tabela de combinações seguras (Nuxt UI / Tailwind v4)

| Texto | Fundo | Ratio | OK? |
|---|---|---|---|
| `text-fg` (`oklch(0.20 0 0)`) | `bg-bg` (`oklch(1 0 0)`) | ~15:1 | ✅ |
| `text-fg-muted` (`oklch(0.50 0 0)`) | `bg-bg` | ~7:1 | ✅ |
| `text-fg` | `bg-elevated` (`oklch(0.98 0 0)`) | ~14:1 | ✅ |
| `text-fg-subtle` (`oklch(0.65 0 0)`) | `bg-bg` | ~4.6:1 | ⚠️ mínimo |
| `text-fg-subtle` | `bg-elevated` | ~4.4:1 | ❌ abaixo |
| `text-white` | `bg-primary-500` | depende do shade | medir! |

### Ferramentas pra validar

- **Chrome DevTools** → Inspect → "Accessibility" tab.
- **WebAIM Contrast Checker**: [https://webaim.org/resources/contrastchecker/](https://webaim.org/resources/contrastchecker/)
- **Nuxt UI tem `text-toned` variant** que dá contraste melhor
  sobre fundos coloridos.

### Regra prática

> **Se você precisa colocar texto claro sobre fundo claro ou
> escuro sobre escuro, está errado.** Use `text-fg` (escuro)
> sobre `bg-bg` (claro) ou vice-versa. Nunca `text-fg-muted`
> sobre `bg-primary-100`.

### Cor como indicador único

❌ **Errado**: badge vermelha para erro (daltônicos não veem).
✅ **Certo**: badge vermelha + ícone de alerta + texto "Erro".

```vue
<UAlert
  color="error"
  icon="i-lucide-alert-triangle"
  title="Falha no pagamento"
  description="Tente novamente em 30s."
/>
```

**Sempre combine cor com ícone + texto**. Cor sozinha
**não é acessível**.

---

## 4. Consistency — escala em tudo

> **O usuário não percebe o que está bom, mas percebe
> imediatamente o que está inconsistente.** Aplicar a
> mesma escala em **tudo** (botões, cards, inputs, badges,
> alerts).

### Botões (sempre 3 variantes, no máximo)

```vue
<!-- Variantes fixas -->
<UButton color="primary">Primary</UButton>     <!-- ação principal -->
<UButton color="neutral" variant="outline">Secondary</UButton>
<UButton variant="ghost">Ghost</UButton>       <!-- terciária -->
<UButton color="error">Delete</UButton>        <!-- destrutiva -->
```

**Nunca** invente variante nova no meio do projeto
(`variant="subtle-blue"`).

### Tamanhos (3, no máximo)

```vue
<UButton size="sm">Small</UButton>   <!-- tabelas, dense UI -->
<UButton size="md">Default</UButton> <!-- default -->
<UButton size="lg">Large</UButton>   <!-- CTAs, hero -->
```

### Cards (sempre mesmo padding + raio)

```vue
<UCard class="p-6">      <!-- default -->
<UCard class="p-4">      <!-- dense (tabelas) -->
<UCard class="p-8">      <!-- espaçoso (hero cards) -->
```

`rounded-card` (8px ou 12px) **sempre** — nunca `rounded-md`
neste card e `rounded-lg` naquele.

### Tipografia (sempre mesma font family)

```css
/* main.css */
@theme {
  --font-sans: "Inter", system-ui, -apple-system, sans-serif;
  --font-mono: "JetBrains Mono", ui-monospace, monospace;
}
```

**Não** misture Inter (H1) + Roboto (body) + system-ui
(footer). Escolhe 1 e usa em tudo.

---

## 5. Motion — sutileza é tudo

> **Motion é pra dar feedback, não pra distrair.**
> 200-300ms é o sweet spot. Acima de 500ms o usuário
> acha que travou.

| Tipo | Duração | Easing | Exemplo |
|---|---|---|---|
| **Hover/focus** | 100-150ms | `ease-out` | Button color change |
| **Modal/drawer open** | 200-250ms | `ease-out` | Slideover enters |
| **Page transition** | 200-300ms | `ease-in-out` | Route change |
| **Toast** | 200ms in, 200ms out | `ease-out` | Notification |
| **Loading spinner** | loop infinito | `linear` | Spinner rotate |

```vue
<button class="transition-colors duration-150 hover:bg-primary-600">
  Salvar
</button>
```

**Não** anime `width`/`height`/`top`/`left` (causa reflow).
Anime `transform` (translate, scale, rotate) e `opacity`
(composite layers, 60fps garantido).

---

## 6. Touch targets — mobile-first

> **Mínimo 44×44px** (Apple HIG / Material). Abaixo disso
> é "difícil de tocar" no celular.

```vue
<!-- ❌ ERRADO: botão pequeno demais pra mobile -->
<button class="px-2 py-1 text-sm">OK</button>

<!-- ✅ CORRETO: 44px mínimo -->
<button class="px-4 py-3 min-h-[44px]">OK</button>
```

**Inputs** também: `min-h-[44px]` ou `h-11` (44px) no
default.

---

## 7. Self-check pré-PR (visual)

Antes de abrir PR de UI, valide:

### Hierarchy
- [ ] **H1 é claramente o maior** da página (≥ 1.5× body).
- [ ] **H2 vs H3** diferença visível (tamanho OU peso).
- [ ] **Body text é 16px** (não 14px — ruim pra mobile).

### Whitespace
- [ ] **Spacing scale consistente** (4/8/12/16/24/32, não 5/7/11).
- [ ] **Padding de card** é 1 valor único (`p-6`, não `p-5` em
      uns e `p-7` em outros).
- [ ] **Gap entre seções** ≥ 32px (`gap-8` ou `mt-8`).

### Contrast
- [ ] **Texto** sobre fundo: ≥ 4.5:1 (Chrome DevTools).
- [ ] **Não depende só de cor** (sempre ícone + texto).
- [ ] **Dark mode** testado (se aplicável).

### Consistency
- [ ] **Botões** usam mesmas 3-4 variantes em todo o projeto.
- [ ] **Cards** têm mesmo padding + raio.
- [ ] **Font family** é 1 (não mistura Inter + Roboto).
- [ ] **Bordas** são 1 cor (`border-border`, não `border-gray-200`
      em uns e `border-neutral-300` em outros).

### Mobile
- [ ] **Touch targets** ≥ 44×44px.
- [ ] **Responsive** em 375px (iPhone SE), 768px (tablet),
      1280px+ (desktop).

### Polish
- [ ] **Screenshot local** gerado com Playwright (não só
      "abri no navegador e parece ok").
- [ ] **Loading + empty + error states** têm o mesmo polish
      visual que o estado "happy path".

---

## 8. Self-check rápido (5 segundos)

Quando o screenshot aparecer, **respire e pergunte**:

1. **"Eu pagaria por este app?"** — Se sim, ✅.
2. **"Parece com Linear/Notion/Vercel/Stripe?"** — Esses
   são a referência de "profissional moderno".
3. **"Tem algum elemento que parece 'grudado' ou 'perdido'?"** —
   Spacing ou hierarchy ruim.
4. **"Alguma cor parece aleatória (não combinando com o resto)?"** —
   Token system quebrado.
5. **"Os botões principais são óbvios?"** — Se você precisa
   procurar o "Salvar", o hierarchy de CTA está errado.

Se 3+ respostas são "não" ou "mais ou menos", **refazer
antes de abrir PR**.

---

## Referências externas (skills públicas)

Instale estas skills via `npx skills add`:

- `leonxlnx/taste-skill@high-end-visual-design` (210.6K installs)
  — Princípios visuais, antes/depois, exemplos reais.
- `wshobson/agents@visual-design-foundations` (10.1K) —
  Foundations (hierarchy, contrast, color).
- `nuxt/ui@nuxt-ui` (15.2K) — Nuxt UI v4 design system.

---

## Anti-patterns (NÃO faça)

❌ **Emojis em UI** (exceto ícone de marca ou empty state
   explicitamente "fofinho"). Use **ícones** da coleção
   padrão (lucide no Nuxt UI).

❌ **Mais de 2-3 fontes** na mesma tela. Limita a 1
   sans-serif + 1 mono (se necessário).

❌ **Cores hex hardcoded** em componentes. Vai no
   `@theme` (CSS-first config).

❌ **Animações longas** (> 500ms). Usuário acha que travou.

❌ **Sombras pesadas** (drop-shadow-2xl em tudo). Use sombra
   **discreta** por padrão, **mais forte** só em modais/
   popovers.

❌ **Gradientes berrantes** (purple-to-pink-to-orange).
   Use gradiente **sutil** ou cor sólida.

❌ **Mix de frameworks de UI** (Nuxt UI + PrimeVue +
   Bootstrap na mesma página). Escolhe 1.

---

## Referências internas

- **Skill `frontend-public-skills`**: [../frontend-public-skills/SKILL.md](../frontend-public-skills/SKILL.md)
- **Skill `nuxt-ui-patterns`**: [../nuxt-ui-patterns/SKILL.md](../nuxt-ui-patterns/SKILL.md)
- **Skill `tailwind-only-patterns`**: [../tailwind-only-patterns/SKILL.md](../tailwind-only-patterns/SKILL.md)
- **Skill `ux-design-best-practices`**: [../ux-design-best-practices/SKILL.md](../ux-design-best-practices/SKILL.md)
- **Sensor 12 `frontend-polish`**: [../../sensors/12-frontend-polish.md](../../sensors/12-frontend-polish.md)
- **Playwright screenshot script**: [../../scripts/visual/playwright-screenshot.mjs](../../scripts/visual/playwright-screenshot.mjs)
- **ADR-0022** (esta decisão)
