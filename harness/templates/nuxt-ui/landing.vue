<script setup lang="ts">
// Public landing page — composes Hero + Features + CTA + Footer.
// CTA buttons read query (?intent=signup) so the auth page can preselect role.
definePageMeta({ layout: 'public' })

const route = useRoute()
const signupIntent = computed(() =>
  typeof route.query.intent === 'string' ? route.query.intent : null,
)

interface Feature {
  icon: string
  titleKey: string
  descKey: string
}

const features: Feature[] = [
  { icon: 'i-lucide-shopping-bag', titleKey: 'landing.features.shop.title',     descKey: 'landing.features.shop.desc' },
  { icon: 'i-lucide-users',        titleKey: 'landing.features.community.title', descKey: 'landing.features.community.desc' },
  { icon: 'i-lucide-truck',        titleKey: 'landing.features.delivery.title',  descKey: 'landing.features.delivery.desc' },
]
</script>

<template>
  <div>
    <section class="py-24 bg-gradient-to-b from-elevated to-bg">
      <UContainer>
        <div class="flex flex-col items-center text-center gap-6 max-w-3xl mx-auto">
          <UBadge color="primary" variant="subtle" :label="$t('landing.hero.badge')" class="text-sm font-medium" />
          <h1 class="text-4xl sm:text-5xl lg:text-6xl font-extrabold leading-tight tracking-tight text-highlighted m-0">
            {{ $t('landing.hero.title') }}
          </h1>
          <p class="text-lg sm:text-xl leading-relaxed text-muted m-0 max-w-2xl">
            {{ $t('landing.hero.tagline') }}
          </p>
          <div class="flex flex-wrap gap-4 justify-center mt-2">
            <UButton
              :to="`/auth/register${signupIntent ? `?intent=${signupIntent}` : ''}`"
              color="primary"
              size="xl"
              icon="i-lucide-arrow-right"
              trailing
            >
              {{ $t('landing.hero.cta.primary') }}
            </UButton>
            <UButton to="/auth/login" color="neutral" variant="ghost" size="xl">
              {{ $t('landing.hero.cta.secondary') }}
            </UButton>
          </div>
        </div>
      </UContainer>
    </section>

    <section class="py-24">
      <UContainer>
        <h2 class="text-3xl sm:text-4xl font-bold leading-tight tracking-tight text-center text-highlighted m-0 mb-12">
          {{ $t('landing.features.title') }}
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <UCard v-for="feature in features" :key="feature.titleKey" class="flex flex-col gap-4 p-8">
            <UIcon :name="feature.icon" class="w-8 h-8 text-primary" />
            <h3 class="text-xl font-semibold text-highlighted m-0">
              {{ $t(feature.titleKey) }}
            </h3>
            <p class="text-base leading-relaxed text-muted m-0">
              {{ $t(feature.descKey) }}
            </p>
          </UCard>
        </div>
      </UContainer>
    </section>

    <section class="pb-24">
      <UContainer>
        <UCard class="flex flex-col items-center text-center gap-4 p-16 bg-elevated">
          <h2 class="text-2xl sm:text-3xl font-bold text-highlighted m-0">
            {{ $t('landing.cta.title') }}
          </h2>
          <p class="text-lg leading-relaxed text-muted m-0 mb-4 max-w-xl">
            {{ $t('landing.cta.body') }}
          </p>
          <UButton
            :to="`/auth/register${signupIntent ? `?intent=${signupIntent}` : ''}`"
            color="primary"
            size="xl"
            icon="i-lucide-arrow-right"
            trailing
          >
            {{ $t('landing.cta.button') }}
          </UButton>
        </UCard>
      </UContainer>
    </section>
  </div>
</template>
