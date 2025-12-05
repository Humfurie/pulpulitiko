<script setup lang="ts">
import type { ApiResponse, SearchAnalytics, TimeRange } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const loading = ref(true)
const analytics = ref<SearchAnalytics | null>(null)
const selectedTimeRange = ref<TimeRange>('1d')

const timeRangeOptions = [
  { label: 'Last Hour', value: '1h' as TimeRange },
  { label: 'Last 24 Hours', value: '1d' as TimeRange },
  { label: 'Last Week', value: '1w' as TimeRange },
  { label: 'Last Month', value: '1m' as TimeRange },
  { label: 'Last Year', value: '1y' as TimeRange },
  { label: 'Last 5 Years', value: '5y' as TimeRange },
  { label: 'All Time', value: 'lifetime' as TimeRange }
]

async function loadAnalytics() {
  loading.value = true
  try {
    const res = await $fetch<ApiResponse<SearchAnalytics>>(
      `${baseUrl}/admin/analytics/search?time_range=${selectedTimeRange.value}`,
      { headers: auth.getAuthHeaders() }
    )

    if (res.success) {
      analytics.value = res.data
    }
  } catch (e) {
    console.error('Failed to load search analytics', e)
  }
  loading.value = false
}

function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

function formatCTR(ctr: number): string {
  return ctr.toFixed(1) + '%'
}

function formatPosition(pos: number): string {
  return pos.toFixed(1)
}

watch(selectedTimeRange, () => {
  loadAnalytics()
})

onMounted(loadAnalytics)

useSeoMeta({
  title: 'Search Analytics - Pulpulitiko Admin'
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Search Analytics</h1>
      <USelect
        v-model="selectedTimeRange"
        :items="timeRangeOptions"
        value-key="value"
        class="w-48"
      />
    </div>

    <!-- Stats Overview -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <div class="flex items-center gap-4">
          <div class="p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
            <UIcon name="i-heroicons-magnifying-glass" class="size-6 text-blue-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Total Searches</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              <span v-if="loading" class="inline-block w-8 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              <span v-else>{{ formatNumber(analytics?.total_searches || 0) }}</span>
            </p>
          </div>
        </div>
      </UCard>

      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <div class="flex items-center gap-4">
          <div class="p-3 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
            <UIcon name="i-heroicons-hashtag" class="size-6 text-purple-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Unique Terms</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              <span v-if="loading" class="inline-block w-8 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              <span v-else>{{ formatNumber(analytics?.unique_search_terms || 0) }}</span>
            </p>
          </div>
        </div>
      </UCard>

      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <div class="flex items-center gap-4">
          <div class="p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
            <UIcon name="i-heroicons-cursor-arrow-rays" class="size-6 text-green-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Total Clicks</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              <span v-if="loading" class="inline-block w-8 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              <span v-else>{{ formatNumber(analytics?.total_clicks || 0) }}</span>
            </p>
          </div>
        </div>
      </UCard>

      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <div class="flex items-center gap-4">
          <div class="p-3 bg-orange-50 dark:bg-orange-900/20 rounded-lg">
            <UIcon name="i-heroicons-chart-bar" class="size-6 text-orange-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Click-Through Rate</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              <span v-if="loading" class="inline-block w-8 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              <span v-else>{{ formatCTR(analytics?.overall_ctr || 0) }}</span>
            </p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Top Search Terms & Top Clicked Articles Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
      <!-- Top Search Terms -->
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
              <UIcon name="i-heroicons-magnifying-glass" class="size-5 text-blue-500" />
            </div>
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">Top Search Terms</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">Most searched queries</p>
            </div>
          </div>
        </template>

        <div v-if="loading" class="space-y-3">
          <div v-for="i in 5" :key="i" class="flex items-center gap-3">
            <div class="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
            <div class="flex-1 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
            <div class="w-16 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
          </div>
        </div>

        <div v-else-if="analytics?.top_search_terms?.length" class="space-y-3">
          <div
            v-for="(term, index) in analytics.top_search_terms.slice(0, 10)"
            :key="term.query"
            class="flex items-center gap-3"
          >
            <span
              class="flex-shrink-0 w-6 h-6 flex items-center justify-center text-sm font-bold rounded-full"
              :class="{
                'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400': index === 0,
                'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400': index > 0
              }"
            >
              {{ index + 1 }}
            </span>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
                "{{ term.query }}"
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ term.click_count }} clicks / {{ formatCTR(term.ctr) }} CTR
              </p>
            </div>
            <div class="flex items-center gap-1 text-sm text-gray-500 dark:text-gray-400">
              <UIcon name="i-heroicons-magnifying-glass" class="size-4" />
              {{ formatNumber(term.count) }}
            </div>
          </div>
        </div>

        <div v-else class="text-center py-4 text-gray-500 dark:text-gray-400">
          No search data yet
        </div>
      </UCard>

      <!-- Top Clicked Articles -->
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-green-50 dark:bg-green-900/20">
              <UIcon name="i-heroicons-document-text" class="size-5 text-green-500" />
            </div>
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">Top Clicked Articles</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">Most clicked from search results</p>
            </div>
          </div>
        </template>

        <div v-if="loading" class="space-y-3">
          <div v-for="i in 5" :key="i" class="flex items-center gap-3">
            <div class="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
            <div class="flex-1 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
            <div class="w-16 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
          </div>
        </div>

        <div v-else-if="analytics?.top_clicked_articles?.length" class="space-y-3">
          <div
            v-for="(article, index) in analytics.top_clicked_articles.slice(0, 10)"
            :key="article.article_id"
            class="flex items-center gap-3"
          >
            <span
              class="flex-shrink-0 w-6 h-6 flex items-center justify-center text-sm font-bold rounded-full"
              :class="{
                'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400': index === 0,
                'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400': index > 0
              }"
            >
              {{ index + 1 }}
            </span>
            <div class="flex-1 min-w-0">
              <NuxtLink
                :to="`/admin/articles/${article.article_id}`"
                class="text-sm font-medium text-gray-900 dark:text-white hover:text-primary-500 truncate block"
              >
                {{ article.article_title }}
              </NuxtLink>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                Avg position: {{ formatPosition(article.avg_position) }}
              </p>
            </div>
            <div class="flex items-center gap-1 text-sm text-gray-500 dark:text-gray-400">
              <UIcon name="i-heroicons-cursor-arrow-rays" class="size-4" />
              {{ formatNumber(article.click_count) }}
            </div>
          </div>
        </div>

        <div v-else class="text-center py-4 text-gray-500 dark:text-gray-400">
          No click data yet
        </div>
      </UCard>
    </div>

    <!-- Politician Searches & Search Trends Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Politician Searches -->
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-purple-50 dark:bg-purple-900/20">
              <UIcon name="i-heroicons-user-circle" class="size-5 text-purple-500" />
            </div>
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">Politician Searches</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">Searches matched to politicians</p>
            </div>
          </div>
        </template>

        <div v-if="loading" class="space-y-3">
          <div v-for="i in 5" :key="i" class="flex items-center gap-3">
            <div class="flex-1 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
            <div class="w-16 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
          </div>
        </div>

        <div v-else-if="analytics?.politician_searches?.length" class="space-y-3">
          <div
            v-for="pol in analytics.politician_searches"
            :key="pol.politician_id"
            class="flex items-center justify-between"
          >
            <NuxtLink
              :to="`/admin/politicians/${pol.politician_id}`"
              class="text-sm font-medium text-gray-900 dark:text-white hover:text-primary-500"
            >
              {{ pol.politician_name }}
            </NuxtLink>
            <UBadge color="primary" variant="soft">
              {{ formatNumber(pol.search_count) }} searches
            </UBadge>
          </div>
        </div>

        <div v-else class="text-center py-4 text-gray-500 dark:text-gray-400">
          No politician search data yet
        </div>
      </UCard>

      <!-- Search Trends -->
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-orange-50 dark:bg-orange-900/20">
              <UIcon name="i-heroicons-arrow-trending-up" class="size-5 text-orange-500" />
            </div>
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">Search Trends</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">Search volume over time</p>
            </div>
          </div>
        </template>

        <div v-if="loading" class="space-y-2">
          <div v-for="i in 6" :key="i" class="flex items-center gap-2">
            <div class="w-24 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
            <div class="flex-1 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
          </div>
        </div>

        <div v-else-if="analytics?.search_trends?.length" class="space-y-2">
          <div
            v-for="trend in analytics.search_trends.slice(0, 10)"
            :key="trend.period"
            class="flex items-center gap-2"
          >
            <span class="w-24 text-xs text-gray-500 dark:text-gray-400 truncate">
              {{ trend.period }}
            </span>
            <div class="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-6 relative overflow-hidden">
              <div
                class="bg-orange-500 h-6 rounded-full transition-all duration-300 flex items-center justify-end pr-2"
                :style="{
                  width: `${analytics.search_trends.length > 0
                    ? Math.max(10, (trend.search_count / Math.max(...analytics.search_trends.map(t => t.search_count))) * 100)
                    : 0}%`
                }"
              >
                <span class="text-xs font-medium text-white">
                  {{ trend.search_count }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="text-center py-4 text-gray-500 dark:text-gray-400">
          No trend data yet
        </div>
      </UCard>
    </div>
  </div>
</template>
