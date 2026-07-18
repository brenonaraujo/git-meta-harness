# Stack — Frontend (Nuxt 3/4 + Nuxt UI + Pinia)

> **Padrão canônico** do meta-harness para frontends. Mudanças aqui
> exigem ADR e aprovação do `team-manager`.

---

## Framework & libs

| Camada               | Lib                                                | Versão alvo |
|----------------------|----------------------------------------------------|-------------|
| Framework            | `nuxt`                                             | v3.x / v4.x |
| UI                   | `@nuxt/ui`                                         | v3.x        |
| State                | `pinia` + `@pinia/nuxt`                            | v3.x / v0.5+ |
| Composables          | `@vueuse/core`, `@vueuse/nuxt`                     | latest      |
| TypeScript           | nativo (strict mode)                               | v5+         |
| Lint/Format          | `@nuxt/eslint` + `prettier`                        | latest      |
| Testes unit          | `vitest` + `@vue/test-utils` + `@nuxt/test-utils`  | latest      |
| E2E (QA)             | `@playwright/test`                                 | latest      |
| HTTP client          | `$fetch` (Nitro) ou `ofetch`                       | nativo      |
| Validação            | `zod`                                              | v3+         |
| **i18n**             | `@nuxtjs/i18n`                                     | v8+         |
| Estilo               | `@nuxt/ui` (Tailwind v4 + Reka UI)                | v3+         |
| Ícones               | `@nuxt/ui` (iconify via `@iconify-json/lucide`)    | v3+         |

---

## Estrutura de pastas (Nuxt 4)

```
my-app/
├── app/
│   ├── app.vue
│   ├── error.vue
│   ├── pages/                       # roteamento automático
│   │   ├── index.vue
│   │   └── login.vue
│   ├── layouts/
│   │   └── default.vue
│   ├── components/
│   │   ├── common/                  # 100% reutilizáveis
│   │   │   ├── AppHeader.vue
│   │   │   └── AppFooter.vue
│   │   ├── feature/                 # uma feature de negócio
│   │   │   └── auth/
│   │   │       ├── LoginForm.vue
│   │   │       └── RegisterForm.vue
│   │   └── ui/                      # wrappers sobre Nuxt UI
│   │       └── AppButton.vue
│   ├── composables/                 # useFoo()
│   │   ├── useAuth.ts
│   │   └── useApi.ts
│   ├── stores/                      # Pinia setup stores
│   │   ├── auth.ts
│   │   └── user.ts
│   ├── middleware/
│   │   └── auth.ts
│   ├── plugins/
│   │   └── api.ts
│   ├── assets/
│   │   └── css/
│   │       └── main.css
│   └── utils/                       # pure functions
│       └── format.ts
├── server/                          # rotas Nitro (se houver)
├── shared/                          # tipos compartilhados client/server
├── public/
├── tests/
│   ├── unit/                        # vitest
│   │   ├── components/
│   │   └── composables/
│   └── e2e/                         # playwright
│       ├── playwright.config.ts
│       └── smoke.spec.ts
├── nuxt.config.ts
├── package.json
├── tsconfig.json
└── .env.example
```

---

## Convenções de código

### `<script setup>` sempre

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '~/stores/auth'

const email = ref('')
const password = ref('')
const auth = useAuthStore()

const isLoading = computed(() => auth.isLoading)

async function handleSubmit() {
  await auth.login({ email: email.value, password: password.value })
}
</script>

<template>
  <UForm @submit="handleSubmit">
    <UFormField label="Email" name="email">
      <UInput v-model="email" type="email" required />
    </UFormField>
    <UFormField label="Senha" name="password">
      <UInput v-model="password" type="password" required />
    </UFormField>
    <UButton type="submit" :loading="isLoading">Entrar</UButton>
  </UForm>
</template>
```

### Composables (lógica reutilizável)

```ts
// app/composables/useAuth.ts
import { useAuthStore } from '~/stores/auth'

export function useAuth() {
  const auth = useAuthStore()
  return {
    user: computed(() => auth.user),
    isLoggedIn: computed(() => auth.isLoggedIn),
    login: auth.login.bind(auth),
    logout: auth.logout.bind(auth),
  }
}
```

### Pinia Setup Stores

```ts
// app/stores/auth.ts
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', () => {
  // state
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const isLoading = ref(false)

  // getters
  const isLoggedIn = computed(() => !!token.value)

  // actions
  async function login(creds: { email: string; password: string }) {
    isLoading.value = true
    try {
      const res = await $fetch<{ token: string; user: User }>('/api/v1/auth/login', {
        method: 'POST',
        body: creds,
      })
      token.value = res.token
      user.value = res.user
    } finally {
      isLoading.value = false
    }
  }

  function logout() {
    token.value = null
    user.value = null
  }

  return { user, token, isLoading, isLoggedIn, login, logout }
})
```

> **Não usar Options Stores** (`defineStore('auth', { state: ..., actions: ... })`).
> Setup Stores são o padrão oficial do Pinia 3 e Vue 3.

### Nuxt UI v3

```vue
<UButton color="primary" variant="solid">Salvar</UButton>
<UInput v-model="email" type="email" placeholder="voce@exemplo.com" />
<UCard>
  <template #header>Título</template>
  Conteúdo
</UCard>
<UTable :rows="rows" :columns="cols" />
```

Para customizar, use `app.config.ts` com Tailwind Variants API:

```ts
// app.config.ts
export default defineAppConfig({
  ui: {
    button: {
      slots: {
        base: 'font-medium',
      },
      defaultVariants: {
        size: 'md',
        color: 'primary',
      },
    },
  },
})
```

### Acesso a API

**Não** chame `$fetch` direto nos componentes. Use composables:

```ts
// app/composables/useApi.ts
import type { $Fetch } from 'ofetch'

export function useApi() {
  const config = useRuntimeConfig()
  return $fetch.create({
    baseURL: config.public.apiBase,
    onRequest({ request, options }) {
      const auth = useAuthStore()
      if (auth.token) {
        options.headers = { ...options.headers, Authorization: `Bearer ${auth.token}` }
      }
    },
  })
}
```

### TypeScript estrito

```json
// tsconfig.json (Nuxt gera, mas reforça)
{
  "extends": "./.nuxt/tsconfig.json",
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true
  }
}
```

---

## Validação (Zod)

```ts
import { z } from 'zod'

const LoginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
})

type LoginInput = z.infer<typeof LoginSchema>

// Em componente:
const result = LoginSchema.safeParse({ email: email.value, password: password.value })
if (!result.success) {
  // mostrar erros
}
```

---

## i18n (internacionalização)

> **Toda copy de UI** (textos, labels, mensagens de erro, e-mails,
> notificações) **deve** passar por `@nuxtjs/i18n`. Idiomas
> obrigatórios: **en**, **pt-BR**, **es**. Ver skill
> [`../skills/i18n.md`](../skills/i18n.md) e sensor
> [`../sensors/08-i18n-audit.md`](../sensors/08-i18n-audit.md).

### Configuração (`nuxt.config.ts`)

```ts
export default defineNuxtConfig({
  modules: ['@nuxt/ui', '@pinia/nuxt', '@nuxtjs/i18n'],
  i18n: {
    strategy: 'no_prefix', // ou 'prefix_except_default' se quiser /en/pt-BR/es
    defaultLocale: 'en',
    locales: [
      { code: 'en',    name: 'English',     file: 'en.json' },
      { code: 'pt-BR', name: 'Português',   file: 'pt-BR.json' },
      { code: 'es',    name: 'Español',     file: 'es.json' },
    ],
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: 'i18n_redirected',
      redirectOn: 'root',
      fallbackLocale: 'en',
    },
  },
})
```

### Estrutura

```
i18n/
└── locales/
    ├── en.json
    ├── pt-BR.json
    └── es.json
```

### Uso em componentes

```vue
<script setup lang="ts">
const { t, locale, locales, setLocale } = useI18n()
</script>

<template>
  <div>
    <h1>{{ t('home.welcome') }}</h1>

    <UForm @submit="onSubmit">
      <UFormField :label="t('auth.email')" name="email">
        <UInput v-model="email" type="email" required />
      </UFormField>
      <UButton type="submit">{{ t('auth.login') }}</UButton>
    </UForm>

    <USelect
      :model-value="locale"
      :items="locales.map(l => ({ label: l.name, value: l.code }))"
      @update:model-value="(v: string) => setLocale(v as any)"
    />
  </div>
</template>
```

### Convenção de chaves

`<domínio>.<ação>.<contexto>` (ex.: `auth.invalid_credentials`,
`home.welcome`, `error.generic`). Igual ao backend.

### Interpolação (Nuxt)

Use `{{ var }}` (sintaxe Vue I18n):

```json
"login_success": "Welcome back, {name}!"
```

```vue
<p>{{ t('auth.login_success', { name: user.name }) }}</p>
```

> **Diferente do Go**: o frontend usa `{name}`, o backend usa `{{.name}}`.

### Pluralização

```json
{
  "cart": {
    "items_count": "{count} item | {count} items"
  }
}
```

```vue
<p>{{ t('cart.items_count', cart.items.length, { count: cart.items.length }) }}</p>
```

### Proibições

- ❌ `<h1>Bem-vindo</h1>` — texto hardcoded.
- ❌ Strings em `data()` ou `const` do componente (ex.: `const welcome
  = 'Bem-vindo'`).
- ❌ Chave plana: `error1`, `unauthorized`.
- ❌ Idioma faltando em `i18n/locales/`.

---

## Auto-imports

Nuxt auto-importa:

- Composables de `app/composables/`
- Componentes de `app/components/`
- Utils de `app/utils/`
- APIs Nuxt (`ref`, `computed`, `useFetch`, etc.)

Pinia (`@pinia/nuxt`):

- `defineStore`
- `storeToRefs`
- Stores de `app/stores/`

**Não** importe manualmente o que já é auto-importado.

---

## Testes (Vitest)

```ts
// tests/unit/components/LoginForm.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { setActivePinia, createPinia } from 'pinia'
import LoginForm from '~/components/feature/auth/LoginForm.vue'
import { useAuthStore } from '~/stores/auth'

describe('LoginForm', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('submits credenciais e chama auth.login', async () => {
    const wrapper = await mountSuspended(LoginForm)
    const auth = useAuthStore()
    const spy = vi.spyOn(auth, 'login').mockResolvedValue()

    await wrapper.find('input[name="email"]').setValue('user@example.com')
    await wrapper.find('input[name="password"]').setValue('secret123')
    await wrapper.find('form').trigger('submit.prevent')

    expect(spy).toHaveBeenCalledWith({
      email: 'user@example.com',
      password: 'secret123',
    })
  })
})
```

---

## Comandos canônicos (package.json scripts)

```json
{
  "scripts": {
    "dev": "nuxt dev",
    "build": "nuxt build",
    "preview": "nuxt preview",
    "lint": "eslint .",
    "lint:fix": "eslint . --fix",
    "format": "prettier --write .",
    "format:check": "prettier --check .",
    "typecheck": "nuxt typecheck",
    "test": "vitest",
    "test:run": "vitest --run",
    "test:coverage": "vitest --run --coverage",
    "audit": "pnpm audit",
    "docker:build": "docker build -f deploy/Dockerfile -t my-app:dev ."
  }
}
```

---

## Dockerfile (multi-stage, node alpine)

Ver [`docker.md`](./docker.md) e o template em
[`../templates/Dockerfile.template`](../templates/Dockerfile.template)
(adapte para Node).

---

## Anti-padrões

- ❌ Options API (`data()`, `methods: {}`) — sempre `<script setup>`.
- ❌ Options Stores do Pinia — sempre Setup Stores.
- ❌ Acessar API direto do componente — sempre via composable.
- ❌ `any` sem justificativa.
- ❌ Lógica de negócio em `pages/` (extraia para composable ou store).
- ❌ Componentes sem `name` (sempre defina).
- ❌ Componentes > 150 linhas (quebre em subcomponentes).
- ❌ `console.log` em produção (use logger ou remova).
- ❌ CSS inline (`style=""`) sem justificativa.
- ❌ Dependências fora da lista sem ADR.
- ❌ Mutação direta de state fora de actions do Pinia.
- ❌ Destructuring de store sem `storeToRefs()` (perde reatividade).
