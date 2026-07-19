<script setup lang="ts">
// Admin dashboard — top stats cards + main content slot. Sidebar lives
// in layouts/default.vue; this page is the content area only.
definePageMeta({ layout: 'default' })

interface Stat {
  labelKey: string
  value: string
  delta: string
  trend: 'up' | 'down' | 'neutral'
  icon: string
}

const stats: Stat[] = [
  { labelKey: 'dashboard.stats.revenue.label', value: 'R$ 12.500', delta: '+12%', trend: 'up',   icon: 'i-lucide-trending-up' },
  { labelKey: 'dashboard.stats.orders.label',   value: '124',       delta: '+8%',  trend: 'up',   icon: 'i-lucide-shopping-cart' },
  { labelKey: 'dashboard.stats.users.label',   value: '892',       delta: '-2%',  trend: 'down', icon: 'i-lucide-users' },
]
</script>

<template>
  <UDashboardPage>
    <UDashboardPanel grow>
      <UDashboardNavbar :title="$t('dashboard.title')">
        <template #right>
          <UButton
            :to="$t('dashboard.actions.new.href')"
            color="primary"
            icon="i-lucide-plus"
          >
            {{ $t('dashboard.actions.new.label') }}
          </UButton>
        </template>
      </UDashboardNavbar>

      <UDashboardPanelContent>
        <BreadcrumbHome class="mb-6" />

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
          <UCard
            v-for="stat in stats"
            :key="stat.labelKey"
            class="flex flex-col gap-2 p-6"
          >
            <template #header>
              <div class="flex items-center justify-between w-full">
                <span class="text-sm font-medium text-muted">
                  {{ $t(stat.labelKey) }}
                </span>
                <UIcon :name="stat.icon" class="w-5 h-5 text-muted" />
              </div>
            </template>
            <div class="text-3xl font-bold leading-tight text-highlighted">
              {{ stat.value }}
            </div>
            <div
              class="text-sm font-medium"
              :class="stat.trend === 'up' ? 'text-success' : stat.trend === 'down' ? 'text-error' : 'text-muted'"
            >
              {{ stat.delta }}
            </div>
          </UCard>
        </div>

        <UCard class="mt-4">
          <template #header>
            <h2 class="text-xl font-semibold text-highlighted m-0">
              {{ $t('dashboard.main.title') }}
            </h2>
          </template>
          <slot />
        </UCard>
      </UDashboardPanelContent>
    </UDashboardPanel>
  </UDashboardPage>
</template>
