<script setup lang="ts">
const api = useApi()

const page = ref(1)
const perPage = ref(12)
const category = ref<string | undefined>(undefined)

const categories = [
  { value: '', label: 'All Categories' },
  { value: 'registration', label: 'Registration' },
  { value: 'voting_day', label: 'Voting Day' },
  { value: 'requirements', label: 'Requirements' },
  { value: 'general', label: 'General Information' }
]

const { data, status } = await useAsyncData(
  () => api.getVoterEducation(undefined, category.value, page.value, perPage.value),
  { watch: [page, category] }
)

function getContentTypeIcon(type: string): string {
  switch (type) {
    case 'guide':
      return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
    case 'faq':
      return 'M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
    case 'video':
      return 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z'
    default:
      return 'M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z'
  }
}

function getContentTypeLabel(type: string): string {
  switch (type) {
    case 'guide':
      return 'Guide'
    case 'faq':
      return 'FAQ'
    case 'video':
      return 'Video'
    default:
      return 'Article'
  }
}

function getCategoryLabel(cat: string | undefined): string {
  if (!cat) return ''
  const found = categories.find(c => c.value === cat)
  return found?.label || cat.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

useSeoMeta({
  title: 'Voter Education - Pulpulitiko',
  description: 'Learn about voter registration, voting day procedures, and your rights as a Filipino voter'
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <!-- Hero Section -->
    <div class="bg-gradient-to-r from-green-600 to-teal-700 text-white py-12">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 class="text-4xl font-bold mb-4">Voter Education</h1>
        <p class="text-xl text-green-100">
          Everything you need to know about voting in the Philippines
        </p>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Category Filter -->
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-8">
        <div class="flex flex-wrap gap-2">
          <button
            v-for="cat in categories"
            :key="cat.value"
            :class="[
              'px-4 py-2 rounded-full text-sm font-medium transition-colors',
              category === cat.value || (!category && !cat.value)
                ? 'bg-green-600 text-white'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
            ]"
            @click="category = cat.value || undefined; page = 1"
          >
            {{ cat.label }}
          </button>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="status === 'pending'" class="flex justify-center py-12">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-green-600" />
      </div>

      <!-- Content Grid -->
      <div v-else-if="data?.items.length" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <NuxtLink
          v-for="item in data.items"
          :key="item.id"
          :to="`/voter-education/${item.slug}`"
          class="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow"
        >
          <div class="p-6">
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="getContentTypeIcon(item.content_type)" />
                </svg>
                <span class="text-sm font-medium text-green-600 dark:text-green-400">
                  {{ getContentTypeLabel(item.content_type) }}
                </span>
              </div>
              <span
                v-if="item.is_featured"
                class="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 rounded-full"
              >
                Featured
              </span>
            </div>

            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              {{ item.title }}
            </h3>

            <div v-if="item.category" class="mb-3">
              <span class="text-sm px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded">
                {{ getCategoryLabel(item.category) }}
              </span>
            </div>

            <div class="flex items-center justify-between text-sm text-gray-500 dark:text-gray-400 mt-4">
              <div class="flex items-center gap-1">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
                <span>{{ item.view_count.toLocaleString() }} views</span>
              </div>
              <span v-if="item.published_at">
                {{ new Date(item.published_at).toLocaleDateString('en-PH', { month: 'short', day: 'numeric', year: 'numeric' }) }}
              </span>
            </div>
          </div>
        </NuxtLink>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-12">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No content found</h3>
        <p class="mt-1 text-sm text-gray-500">Try selecting a different category.</p>
      </div>

      <!-- Pagination -->
      <div v-if="data && data.total_pages > 1" class="mt-8 flex justify-center">
        <nav class="flex items-center gap-2">
          <button
            :disabled="page === 1"
            class="px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
            @click="page--"
          >
            Previous
          </button>
          <span class="text-sm text-gray-700 dark:text-gray-300">
            Page {{ page }} of {{ data.total_pages }}
          </span>
          <button
            :disabled="page === data.total_pages"
            class="px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
            @click="page++"
          >
            Next
          </button>
        </nav>
      </div>
    </div>
  </div>
</template>
