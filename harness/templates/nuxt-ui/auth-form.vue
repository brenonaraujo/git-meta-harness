<script setup lang="ts">
// Auth form template — copy-paste starter for /auth/login and /auth/register.
// Supports role preselect via ?role=leader|resident|supplier (per SPEC #16).
definePageMeta({ layout: 'auth' })

const props = defineProps<{
  mode: 'login' | 'register'
  isSubmitting: boolean
}>()

const emit = defineEmits<{
  (e: 'submit', payload: Record<string, unknown>): void
}>()

const route = useRoute()
const { t } = useI18n()

const preselectedRole = computed(() => {
  const r = route.query.role
  return typeof r === 'string' ? r : null
})

const form = reactive({
  email: '',
  password: '',
  name: '',
  role: preselectedRole.value ?? 'resident',
})

const roleOptions = [
  { label: t('auth.role.resident'), value: 'resident' },
  { label: t('auth.role.leader'),   value: 'leader' },
  { label: t('auth.role.supplier'), value: 'supplier' },
]

function onSubmit(event: Event) {
  event.preventDefault()
  emit('submit', { ...form })
}
</script>

<template>
  <div class="flex items-center justify-center min-h-screen p-6 bg-elevated">
    <UCard class="w-full max-w-md p-8">
      <template #header>
        <h1 class="text-3xl font-bold leading-tight text-highlighted m-0 mb-2">
          {{ $t(`auth.${props.mode}.title`) }}
        </h1>
        <p class="text-base leading-relaxed text-muted m-0">
          {{ $t(`auth.${props.mode}.subtitle`) }}
        </p>
      </template>

      <form class="flex flex-col gap-4 mt-6" @submit="onSubmit">
        <UFormField
          v-if="props.mode === 'register'"
          :label="$t('auth.field.name')"
          name="name"
          required
        >
          <UInput
            v-model="form.name"
            autocomplete="name"
            :placeholder="$t('auth.field.name.placeholder')"
          />
        </UFormField>

        <UFormField :label="$t('auth.field.email')" name="email" required>
          <UInput
            v-model="form.email"
            type="email"
            autocomplete="email"
            :placeholder="$t('auth.field.email.placeholder')"
          />
        </UFormField>

        <UFormField :label="$t('auth.field.password')" name="password" required>
          <UInput
            v-model="form.password"
            type="password"
            :autocomplete="props.mode === 'login' ? 'current-password' : 'new-password'"
            :placeholder="$t('auth.field.password.placeholder')"
          />
        </UFormField>

        <UFormField
          v-if="props.mode === 'register'"
          :label="$t('auth.field.role')"
          name="role"
          required
        >
          <USelect v-model="form.role" :items="roleOptions" value-key="value" />
        </UFormField>

        <UButton
          type="submit"
          color="primary"
          size="lg"
          block
          :loading="props.isSubmitting"
        >
          {{ $t(`auth.${props.mode}.submit`) }}
        </UButton>
      </form>

      <template #footer>
        <p class="text-center text-sm text-muted m-0">
          <NuxtLink
            v-if="props.mode === 'login'"
            to="/auth/register"
            class="text-primary font-medium no-underline hover:underline"
          >
            {{ $t('auth.login.switch_to_register') }}
          </NuxtLink>
          <NuxtLink
            v-else
            to="/auth/login"
            class="text-primary font-medium no-underline hover:underline"
          >
            {{ $t('auth.register.switch_to_login') }}
          </NuxtLink>
        </p>
      </template>
    </UCard>
  </div>
</template>
