<script setup lang="ts">
import type { Politician } from '~/types'

definePageMeta({
  layout: 'politician'
})

const api = useApi()
const route = useRoute()
const router = useRouter()

// View mode from query param (default: position)
const viewMode = computed({
  get: () => (route.query.view as string) || 'position',
  set: (value: string) => router.replace({ query: { ...route.query, view: value } })
})

const { data: politicians, pending, error } = await useAsyncData('all-politicians', () => api.getPoliticians())

// Position hierarchy definition (flat list, ordered by importance)
const positionList = [
  { title: 'President', keywords: ['President of the Philippines'], exclude: ['Vice President'], icon: 'i-heroicons-building-office-2' },
  { title: 'Vice President', keywords: ['Vice President'], icon: 'i-heroicons-building-office-2' },
  { title: 'Senate President', keywords: ['Senate President'], exclude: ['Pro Tempore'], icon: 'i-heroicons-building-library' },
  { title: 'House Speaker', keywords: ['Speaker of the House', 'Speaker'], icon: 'i-heroicons-building-library' },
  { title: 'Senators', keywords: ['Senator'], icon: 'i-heroicons-building-library' },
  { title: 'Representatives', keywords: ['Representative', 'Party-list Representative'], icon: 'i-heroicons-building-library' },
  { title: 'Cabinet Secretaries', keywords: ['Secretary of', 'Executive Secretary'], icon: 'i-heroicons-building-office-2' },
  { title: 'Supreme Court', keywords: ['Chief Justice', 'Associate Justice'], icon: 'i-heroicons-scale' },
  { title: 'Constitutional Commissions', keywords: ['COMELEC', 'COA', 'CSC', 'CHR'], icon: 'i-heroicons-shield-check' },
  { title: 'Other Officials', keywords: ['BSP Governor', 'NEDA', 'National Security Adviser'], icon: 'i-heroicons-briefcase' },
]

// Helper to check if a position indicates a former official
function isFormerOfficial(position: string | undefined): boolean {
  if (!position) return false
  return position.toLowerCase().includes('former')
}

// Group politicians by position (flat list, only current office holders)
const politiciansByPosition = computed(() => {
  if (!politicians.value) return []

  // Filter out former officials first
  const currentOfficials = politicians.value.filter(p => !isFormerOfficial(p.position))

  const result: Array<{
    title: string
    icon: string
    politicians: Politician[]
  }> = []

  for (const positionDef of positionList) {
    const positionPoliticians = currentOfficials.filter(p => {
      if (!p.position) return false
      // Check if position matches any keyword
      const matchesKeyword = positionDef.keywords.some(kw => p.position!.includes(kw))
      if (!matchesKeyword) return false
      // Check if position should be excluded
      const excludeList = positionDef.exclude || []
      const shouldExclude = excludeList.some(ex => p.position!.includes(ex))
      return !shouldExclude
    })

    if (positionPoliticians.length > 0) {
      result.push({
        title: positionDef.title,
        icon: positionDef.icon,
        politicians: positionPoliticians
      })
    }
  }

  return result
})

// Group politicians by party
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
  // Sort parties alphabetically, but put Independent at the end
  const sortedMap = new Map<string, Politician[]>()
  const keys = Array.from(grouped.keys()).sort((a, b) => {
    if (a === 'Independent') return 1
    if (b === 'Independent') return -1
    return a.localeCompare(b)
  })
  keys.forEach(key => sortedMap.set(key, grouped.get(key)!))
  return sortedMap
})

// SEO
useHead({
  title: 'Politicians - Pulpulitiko',
  meta: [
    { name: 'description', content: 'Browse all politicians covered by Pulpulitiko' }
  ]
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Header -->
      <div class="mb-6">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          Politicians
        </h1>
        <p class="mt-2 text-gray-600 dark:text-gray-400">
          Browse all politicians covered in our articles
        </p>
      </div>

      <!-- View Toggle -->
      <div class="mb-6 flex gap-2">
        <button
          :class="[
            'px-4 py-2 rounded-full text-sm font-medium transition-all duration-200',
            viewMode === 'position'
              ? 'bg-orange-500 text-white'
              : 'bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:border-orange-300 dark:hover:border-orange-700'
          ]"
          @click="viewMode = 'position'"
        >
          <UIcon name="i-heroicons-building-office-2" class="w-4 h-4 inline mr-1.5" />
          By Position
        </button>
        <button
          :class="[
            'px-4 py-2 rounded-full text-sm font-medium transition-all duration-200',
            viewMode === 'party'
              ? 'bg-orange-500 text-white'
              : 'bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:border-orange-300 dark:hover:border-orange-700'
          ]"
          @click="viewMode = 'party'"
        >
          <UIcon name="i-heroicons-flag" class="w-4 h-4 inline mr-1.5" />
          By Party
        </button>
      </div>

      <!-- Loading state -->
      <div v-if="pending" class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
        <div v-for="i in 10" :key="i" class="animate-pulse">
          <div class="bg-white dark:bg-gray-900 rounded-xl p-4 border border-gray-200 dark:border-gray-800">
            <div class="w-16 h-16 rounded-full bg-gray-200 dark:bg-gray-800 mx-auto mb-3" />
            <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-3/4 mx-auto mb-2" />
            <div class="h-3 bg-gray-200 dark:bg-gray-800 rounded w-1/2 mx-auto" />
          </div>
        </div>
      </div>

      <!-- Error state -->
      <UAlert
        v-else-if="error"
        color="error"
        icon="i-heroicons-exclamation-triangle"
        title="Failed to load politicians"
        :description="error.message"
        class="mb-6"
      />

      <!-- Content -->
      <template v-else-if="politicians?.length">
        <!-- By Party View -->
        <template v-if="viewMode === 'party'">
          <div v-for="[party, partyPoliticians] in politiciansByParty" :key="party" class="mb-10">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <UIcon name="i-heroicons-flag" class="w-5 h-5 text-orange-500" />
              {{ party }}
              <span class="text-sm font-normal text-gray-500 dark:text-gray-400">({{ partyPoliticians.length }})</span>
            </h2>
            <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              <NuxtLink
                v-for="politician in partyPoliticians"
                :key="politician.id"
                :to="`/politician/${politician.slug}`"
                class="group bg-white dark:bg-gray-900 rounded-xl p-4 border border-gray-200 dark:border-gray-800 hover:border-orange-300 dark:hover:border-orange-700 hover:shadow-md transition-all duration-200"
              >
                <div class="flex justify-center mb-3">
                  <NuxtImg
                    v-if="politician.photo"
                    :src="politician.photo"
                    :alt="politician.name"
                    class="w-16 h-16 rounded-full object-cover ring-2 ring-gray-100 dark:ring-gray-800 group-hover:ring-orange-200 dark:group-hover:ring-orange-900 transition-all"
                  />
                  <div v-else class="w-16 h-16 rounded-full bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center ring-2 ring-gray-100 dark:ring-gray-800 group-hover:ring-orange-200 dark:group-hover:ring-orange-900 transition-all">
                    <UIcon name="i-heroicons-user" class="w-8 h-8 text-orange-500" />
                  </div>
                </div>
                <h3 class="text-sm font-medium text-gray-900 dark:text-white text-center group-hover:text-orange-500 transition-colors line-clamp-2">
                  {{ politician.name }}
                </h3>
                <p v-if="politician.position" class="text-xs text-gray-500 dark:text-gray-400 text-center mt-1 line-clamp-1">
                  {{ politician.position }}
                </p>
                <p v-if="politician.article_count" class="text-xs text-orange-600 dark:text-orange-400 text-center mt-2">
                  {{ politician.article_count }} {{ politician.article_count === 1 ? 'article' : 'articles' }}
                </p>
              </NuxtLink>
            </div>
          </div>
        </template>

        <!-- By Position View (Flat list) -->
        <template v-else-if="viewMode === 'position'">
          <div v-for="position in politiciansByPosition" :key="position.title" class="mb-10">
            <!-- Position Header -->
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <UIcon :name="position.icon" class="w-5 h-5 text-orange-500" />
              {{ position.title }}
              <span class="text-sm font-normal text-gray-500 dark:text-gray-400">({{ position.politicians.length }})</span>
            </h2>

            <!-- Politicians grid -->
            <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
              <NuxtLink
                v-for="politician in position.politicians"
                :key="politician.id"
                :to="`/politician/${politician.slug}`"
                class="group bg-white dark:bg-gray-900 rounded-xl p-4 border border-gray-200 dark:border-gray-800 hover:border-orange-300 dark:hover:border-orange-700 hover:shadow-md transition-all duration-200"
              >
                <div class="flex justify-center mb-3">
                  <NuxtImg
                    v-if="politician.photo"
                    :src="politician.photo"
                    :alt="politician.name"
                    class="w-16 h-16 rounded-full object-cover ring-2 ring-gray-100 dark:ring-gray-800 group-hover:ring-orange-200 dark:group-hover:ring-orange-900 transition-all"
                  />
                  <div v-else class="w-16 h-16 rounded-full bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center ring-2 ring-gray-100 dark:ring-gray-800 group-hover:ring-orange-200 dark:group-hover:ring-orange-900 transition-all">
                    <UIcon name="i-heroicons-user" class="w-8 h-8 text-orange-500" />
                  </div>
                </div>
                <h3 class="text-sm font-medium text-gray-900 dark:text-white text-center group-hover:text-orange-500 transition-colors line-clamp-2">
                  {{ politician.name }}
                </h3>
                <p v-if="politician.party" class="text-xs text-gray-500 dark:text-gray-400 text-center mt-1 line-clamp-1">
                  {{ politician.party }}
                </p>
                <p v-if="politician.article_count" class="text-xs text-orange-600 dark:text-orange-400 text-center mt-2">
                  {{ politician.article_count }} {{ politician.article_count === 1 ? 'article' : 'articles' }}
                </p>
              </NuxtLink>
            </div>
          </div>
        </template>
      </template>

      <!-- Empty state -->
      <div v-else class="text-center py-12">
        <UIcon name="i-heroicons-user-group" class="w-16 h-16 text-gray-400 mx-auto mb-4" />
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No politicians yet</h3>
        <p class="text-gray-500 dark:text-gray-400">Politicians will appear here as articles are published.</p>
      </div>
    </div>
  </div>
</template>
