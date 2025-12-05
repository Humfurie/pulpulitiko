<script setup lang="ts">
import type { Politician } from '~/types'

const api = useApi()
const route = useRoute()

// Fetch all politicians to extract unique parties
const { data: politicians } = await useAsyncData('politicians-nav', () => api.getPoliticians())

// Extract unique parties from politicians
const parties = computed(() => {
  if (!politicians.value) return []
  const partySet = new Set<string>()
  politicians.value.forEach((p: Politician) => {
    if (p.party) partySet.add(p.party)
  })
  return Array.from(partySet).sort()
})

// Get politicians grouped by party for dropdown
const politiciansByParty = computed(() => {
  if (!politicians.value) return new Map<string, Politician[]>()
  const grouped = new Map<string, Politician[]>()
  politicians.value.forEach((p: Politician) => {
    const party = p.party || 'Independent'
    if (!grouped.has(party)) {
      grouped.set(party, [])
    }
    grouped.get(party)!.push(p)
  })
  return grouped
})

// Active state
const currentSlug = computed(() => route.params.slug as string)
</script>

<template>
  <nav class="bg-white dark:bg-stone-900 border-b border-stone-200 dark:border-stone-800">
    <UContainer>
      <div class="flex items-center gap-6 h-12 overflow-x-auto scrollbar-hide">
        <!-- All Politicians link -->
        <NuxtLink
          to="/politicians"
          class="flex-shrink-0 px-3 py-1.5 rounded-full text-sm font-medium transition-all duration-200"
          :class="route.path === '/politicians'
            ? 'text-orange-600 bg-orange-50 dark:bg-orange-950/30'
            : 'text-stone-600 dark:text-stone-400 hover:text-orange-500 hover:bg-orange-50 dark:hover:bg-orange-950/30'"
        >
          All Politicians
        </NuxtLink>

        <span class="text-stone-300 dark:text-stone-700">|</span>

        <!-- Party dropdowns -->
        <div class="flex items-center gap-1">
          <UDropdownMenu
            v-for="party in parties"
            :key="party"
            :items="(politiciansByParty.get(party) || []).map((p: Politician) => ({
              label: p.name,
              to: `/politician/${p.slug}`,
              avatar: p.photo ? { src: p.photo, alt: p.name } : undefined
            }))"
          >
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              trailing-icon="i-heroicons-chevron-down"
              class="rounded-full text-sm font-medium transition-all duration-200 flex-shrink-0"
              :class="politiciansByParty.get(party)?.some((p: Politician) => p.slug === currentSlug)
                ? 'text-orange-600 bg-orange-50 dark:bg-orange-950/30'
                : 'text-stone-600 dark:text-stone-400 hover:text-orange-500 hover:bg-orange-50 dark:hover:bg-orange-950/30'"
              :ui="{ trailingIcon: 'size-3' }"
            >
              {{ party }}
            </UButton>
          </UDropdownMenu>

          <!-- Independent politicians (no party) -->
          <UDropdownMenu
            v-if="politiciansByParty.has('Independent')"
            :items="(politiciansByParty.get('Independent') || []).map((p: Politician) => ({
              label: p.name,
              to: `/politician/${p.slug}`,
              avatar: p.photo ? { src: p.photo, alt: p.name } : undefined
            }))"
          >
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              trailing-icon="i-heroicons-chevron-down"
              class="rounded-full text-sm font-medium transition-all duration-200 flex-shrink-0"
              :class="politiciansByParty.get('Independent')?.some((p: Politician) => p.slug === currentSlug)
                ? 'text-orange-600 bg-orange-50 dark:bg-orange-950/30'
                : 'text-stone-600 dark:text-stone-400 hover:text-orange-500 hover:bg-orange-50 dark:hover:bg-orange-950/30'"
              :ui="{ trailingIcon: 'size-3' }"
            >
              Independent
            </UButton>
          </UDropdownMenu>
        </div>
      </div>
    </UContainer>
  </nav>
</template>

<style scoped>
.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
</style>
