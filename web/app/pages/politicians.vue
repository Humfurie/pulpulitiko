<script setup lang="ts">
import type { Politician } from '~/types'

definePageMeta({
  layout: 'politician'
})

const api = useApi()
const route = useRoute()
const router = useRouter()

// View mode from query param
const viewMode = computed({
  get: () => (route.query.view as string) || 'party',
  set: (value: string) => router.replace({ query: { ...route.query, view: value } })
})

const { data: politicians, pending, error } = await useAsyncData('all-politicians', () => api.getPoliticians())

// Position hierarchy definition (only current office holders)
const positionHierarchy = [
  {
    branch: 'Executive Branch',
    icon: 'i-heroicons-building-office-2',
    levels: [
      { title: 'President', keywords: ['President of the Philippines'], exclude: ['Vice President'] },
      { title: 'Vice President', keywords: ['Vice President'] },
      { title: 'Cabinet Secretaries', keywords: ['Secretary of', 'Executive Secretary'] },
    ]
  },
  {
    branch: 'Legislative Branch',
    icon: 'i-heroicons-building-library',
    levels: [
      { title: 'Senate Leadership', keywords: ['Senate President', 'Senate President Pro Tempore'] },
      { title: 'Senators', keywords: ['Senator'] },
      { title: 'House Leadership', keywords: ['Speaker of the House', 'Speaker'] },
      { title: 'Representatives', keywords: ['Representative', 'Party-list Representative'] },
    ]
  },
  {
    branch: 'Judicial Branch',
    icon: 'i-heroicons-scale',
    levels: [
      { title: 'Supreme Court', keywords: ['Chief Justice', 'Associate Justice'] },
    ]
  },
  {
    branch: 'Constitutional Bodies',
    icon: 'i-heroicons-shield-check',
    levels: [
      { title: 'Commission on Elections', keywords: ['COMELEC'] },
      { title: 'Commission on Audit', keywords: ['COA'] },
      { title: 'Civil Service Commission', keywords: ['CSC'] },
      { title: 'Commission on Human Rights', keywords: ['CHR'] },
      { title: 'Bangko Sentral ng Pilipinas', keywords: ['BSP Governor'] },
      { title: 'Other Bodies', keywords: ['NEDA', 'National Security Adviser'] },
    ]
  },
  ]

// Helper to check if a position indicates a former official
function isFormerOfficial(position: string | undefined): boolean {
  if (!position) return false
  return position.toLowerCase().includes('former')
}

// Group politicians by position hierarchy (only current office holders)
const politiciansByPosition = computed(() => {
  if (!politicians.value) return []

  // Filter out former officials first
  const currentOfficials = politicians.value.filter(p => !isFormerOfficial(p.position))

  const result: Array<{
    branch: string
    icon: string
    levels: Array<{
      title: string
      politicians: Politician[]
    }>
  }> = []

  for (const branchDef of positionHierarchy) {
    const branchData = {
      branch: branchDef.branch,
      icon: branchDef.icon,
      levels: [] as Array<{ title: string; politicians: Politician[] }>
    }

    for (const levelDef of branchDef.levels) {
      const levelPoliticians = currentOfficials.filter(p => {
        if (!p.position) return false
        // Check if position matches any keyword
        const matchesKeyword = levelDef.keywords.some(kw => p.position!.includes(kw))
        if (!matchesKeyword) return false
        // Check if position should be excluded
        const excludeList = (levelDef as { exclude?: string[] }).exclude || []
        const shouldExclude = excludeList.some(ex => p.position!.includes(ex))
        return !shouldExclude
      })
      if (levelPoliticians.length > 0) {
        branchData.levels.push({
          title: levelDef.title,
          politicians: levelPoliticians
        })
      }
    }

    if (branchData.levels.length > 0) {
      result.push(branchData)
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
            viewMode === 'party'
              ? 'bg-orange-500 text-white'
              : 'bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:border-orange-300 dark:hover:border-orange-700'
          ]"
          @click="viewMode = 'party'"
        >
          <UIcon name="i-heroicons-flag" class="w-4 h-4 inline mr-1.5" />
          By Party
        </button>
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

        <!-- By Position View (Hierarchy) -->
        <template v-else-if="viewMode === 'position'">
          <div v-for="branch in politiciansByPosition" :key="branch.branch" class="mb-10">
            <!-- Branch Header -->
            <div class="flex items-center gap-3 mb-4 pb-2 border-b border-gray-200 dark:border-gray-800">
              <div class="p-2 bg-orange-100 dark:bg-orange-900/30 rounded-lg">
                <UIcon :name="branch.icon" class="w-6 h-6 text-orange-600 dark:text-orange-400" />
              </div>
              <h2 class="text-xl font-bold text-gray-900 dark:text-white">
                {{ branch.branch }}
              </h2>
            </div>

            <!-- Levels within branch -->
            <div class="space-y-6 pl-4 border-l-2 border-orange-200 dark:border-orange-900/50">
              <div v-for="level in branch.levels" :key="level.title" class="relative">
                <!-- Level connector dot -->
                <div class="absolute -left-[calc(1rem+5px)] top-0 w-2.5 h-2.5 rounded-full bg-orange-400 dark:bg-orange-600" />

                <!-- Level title -->
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-3 flex items-center gap-2">
                  {{ level.title }}
                  <span class="text-sm font-normal text-gray-500 dark:text-gray-400">({{ level.politicians.length }})</span>
                </h3>

                <!-- Politicians grid -->
                <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
                  <NuxtLink
                    v-for="politician in level.politicians"
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
