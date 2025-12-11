<script setup lang="ts">
import type { PollCategory, PollListItem, PaginatedPolls } from '~/types'

definePageMeta({
  layout: 'default'
})

useSeoMeta({
  title: 'Polls - Pulpulitiko',
  description: 'Vote on political polls and share your opinions on current issues in Philippine politics.'
})

const api = useApi()
const route = useRoute()
const router = useRouter()

// Query params
const page = computed(() => Number(route.query.page) || 1)
const category = computed(() => (route.query.category as PollCategory) || undefined)
const search = computed(() => (route.query.search as string) || '')

// Fetch polls
const { data: pollsData, pending, error, refresh } = await useAsyncData<PaginatedPolls>(
  `polls-${page.value}-${category.value}-${search.value}`,
  () => api.getPolls({ category: category.value, search: search.value || undefined }, page.value, 12),
  { watch: [page, category, search] }
)

// Fetch featured polls
const { data: featuredPolls } = await useAsyncData<PollListItem[]>(
  'featured-polls',
  () => api.getFeaturedPolls(5)
)

const polls = computed(() => pollsData.value?.polls || [])
const totalPages = computed(() => pollsData.value?.total_pages || 1)

// Search
const searchInput = ref(search.value)
const handleSearch = () => {
  router.push({
    query: { ...route.query, search: searchInput.value || undefined, page: undefined }
  })
}

// Category filter
const categories: { value: PollCategory | ''; label: string }[] = [
  { value: '', label: 'All Categories' },
  { value: 'general', label: 'General' },
  { value: 'election', label: 'Elections' },
  { value: 'legislation', label: 'Legislation' },
  { value: 'politician', label: 'Politicians' },
  { value: 'policy', label: 'Policy' },
  { value: 'local_issue', label: 'Local Issues' },
  { value: 'national_issue', label: 'National Issues' }
]

const selectedCategory = ref<PollCategory | ''>(category.value || '')

watch(selectedCategory, (val) => {
  router.push({
    query: { ...route.query, category: val || undefined, page: undefined }
  })
})

// Pagination
const goToPage = (newPage: number) => {
  router.push({
    query: { ...route.query, page: newPage > 1 ? newPage : undefined }
  })
}

// Format helpers
const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

const formatTimeRemaining = (endsAt: string) => {
  const end = new Date(endsAt)
  const now = new Date()
  const diff = end.getTime() - now.getTime()

  if (diff <= 0) return 'Ended'

  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))

  if (days > 0) return `${days}d ${hours}h left`
  if (hours > 0) return `${hours}h left`
  return 'Ending soon'
}

const getCategoryColor = (cat: PollCategory) => {
  const colors: Record<PollCategory, string> = {
    general: 'bg-gray-100 text-gray-800',
    election: 'bg-blue-100 text-blue-800',
    legislation: 'bg-purple-100 text-purple-800',
    politician: 'bg-green-100 text-green-800',
    policy: 'bg-yellow-100 text-yellow-800',
    local_issue: 'bg-orange-100 text-orange-800',
    national_issue: 'bg-red-100 text-red-800'
  }
  return colors[cat] || 'bg-gray-100 text-gray-800'
}

const getCategoryLabel = (cat: PollCategory) => {
  return categories.find(c => c.value === cat)?.label || cat
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Hero Section -->
    <div class="bg-gradient-to-r from-blue-600 to-indigo-700 text-white py-12">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 class="text-3xl md:text-4xl font-bold mb-4">Community Polls</h1>
        <p class="text-lg text-blue-100 max-w-2xl">
          Voice your opinion on important political issues. Vote on polls and see what fellow citizens think.
        </p>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
        <!-- Main Content -->
        <div class="lg:col-span-3">
          <!-- Filters -->
          <div class="bg-white rounded-lg shadow-sm p-4 mb-6">
            <div class="flex flex-col sm:flex-row gap-4">
              <!-- Search -->
              <div class="flex-1">
                <form @submit.prevent="handleSearch" class="flex">
                  <input
                    v-model="searchInput"
                    type="text"
                    placeholder="Search polls..."
                    class="flex-1 rounded-l-lg border-gray-300 focus:ring-blue-500 focus:border-blue-500"
                  />
                  <button
                    type="submit"
                    class="px-4 py-2 bg-blue-600 text-white rounded-r-lg hover:bg-blue-700"
                  >
                    Search
                  </button>
                </form>
              </div>

              <!-- Category Filter -->
              <div class="w-full sm:w-48">
                <select
                  v-model="selectedCategory"
                  class="w-full rounded-lg border-gray-300 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option v-for="cat in categories" :key="cat.value" :value="cat.value">
                    {{ cat.label }}
                  </option>
                </select>
              </div>
            </div>
          </div>

          <!-- Loading State -->
          <div v-if="pending" class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div v-for="n in 6" :key="n" class="bg-white rounded-lg shadow-sm p-6 animate-pulse">
              <div class="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
              <div class="h-6 bg-gray-200 rounded w-3/4 mb-4"></div>
              <div class="h-4 bg-gray-200 rounded w-full mb-2"></div>
              <div class="h-4 bg-gray-200 rounded w-2/3"></div>
            </div>
          </div>

          <!-- Error State -->
          <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
            <p class="text-red-600">Failed to load polls. Please try again.</p>
            <button class="mt-4 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700" @click="() => refresh()">
              Retry
            </button>
          </div>

          <!-- Empty State -->
          <div v-else-if="polls.length === 0" class="bg-white rounded-lg shadow-sm p-12 text-center">
            <svg class="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
            </svg>
            <h3 class="text-lg font-medium text-gray-900 mb-2">No polls found</h3>
            <p class="text-gray-500">Try adjusting your search or filter criteria.</p>
          </div>

          <!-- Polls Grid -->
          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <NuxtLink
              v-for="poll in polls"
              :key="poll.id"
              :to="`/polls/${poll.slug}`"
              class="bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow p-6 block"
            >
              <div class="flex items-center justify-between mb-3">
                <span :class="['px-2 py-1 text-xs font-medium rounded-full', getCategoryColor(poll.category)]">
                  {{ getCategoryLabel(poll.category) }}
                </span>
                <span v-if="poll.is_featured" class="text-yellow-500">
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                  </svg>
                </span>
              </div>

              <h3 class="text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
                {{ poll.title }}
              </h3>

              <div class="flex items-center text-sm text-gray-500 mb-4">
                <span>{{ poll.option_count }} options</span>
                <span class="mx-2">·</span>
                <span>{{ poll.total_votes }} votes</span>
                <span v-if="poll.ends_at" class="mx-2">·</span>
                <span v-if="poll.ends_at" class="text-orange-600 font-medium">
                  {{ formatTimeRemaining(poll.ends_at) }}
                </span>
              </div>

              <div class="flex items-center justify-between text-sm">
                <div v-if="poll.author" class="flex items-center text-gray-500">
                  <img
                    v-if="poll.author.avatar"
                    :src="poll.author.avatar"
                    :alt="poll.author.name"
                    class="w-6 h-6 rounded-full mr-2"
                  />
                  <span class="w-6 h-6 rounded-full bg-gray-300 mr-2 flex items-center justify-center text-xs text-white" v-else>
                    {{ poll.author.name.charAt(0) }}
                  </span>
                  <span>{{ poll.author.name }}</span>
                </div>
                <span class="text-gray-400">{{ formatDate(poll.created_at) }}</span>
              </div>
            </NuxtLink>
          </div>

          <!-- Pagination -->
          <div v-if="totalPages > 1" class="mt-8 flex justify-center">
            <nav class="flex items-center space-x-2">
              <button
                @click="goToPage(page - 1)"
                :disabled="page <= 1"
                class="px-3 py-2 rounded-lg border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Previous
              </button>

              <template v-for="p in totalPages" :key="p">
                <button
                  v-if="p === 1 || p === totalPages || (p >= page - 1 && p <= page + 1)"
                  @click="goToPage(p)"
                  :class="[
                    'px-3 py-2 rounded-lg border',
                    p === page
                      ? 'bg-blue-600 text-white border-blue-600'
                      : 'border-gray-300 text-gray-600 hover:bg-gray-50'
                  ]"
                >
                  {{ p }}
                </button>
                <span
                  v-else-if="p === page - 2 || p === page + 2"
                  class="px-2 text-gray-400"
                >
                  ...
                </span>
              </template>

              <button
                @click="goToPage(page + 1)"
                :disabled="page >= totalPages"
                class="px-3 py-2 rounded-lg border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Next
              </button>
            </nav>
          </div>
        </div>

        <!-- Sidebar -->
        <div class="lg:col-span-1">
          <!-- Featured Polls -->
          <div v-if="featuredPolls && featuredPolls.length > 0" class="bg-white rounded-lg shadow-sm p-6 mb-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4 flex items-center">
              <svg class="w-5 h-5 text-yellow-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
              </svg>
              Featured Polls
            </h2>
            <div class="space-y-4">
              <NuxtLink
                v-for="poll in featuredPolls"
                :key="poll.id"
                :to="`/polls/${poll.slug}`"
                class="block hover:bg-gray-50 -mx-2 px-2 py-2 rounded-lg"
              >
                <h3 class="text-sm font-medium text-gray-900 line-clamp-2 mb-1">
                  {{ poll.title }}
                </h3>
                <p class="text-xs text-gray-500">
                  {{ poll.total_votes }} votes
                </p>
              </NuxtLink>
            </div>
          </div>

          <!-- Quick Stats -->
          <div class="bg-white rounded-lg shadow-sm p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Categories</h2>
            <div class="space-y-2">
              <button
                v-for="cat in categories.filter(c => c.value)"
                :key="cat.value"
                @click="selectedCategory = cat.value"
                :class="[
                  'w-full text-left px-3 py-2 rounded-lg text-sm transition-colors',
                  selectedCategory === cat.value
                    ? 'bg-blue-100 text-blue-800'
                    : 'hover:bg-gray-100 text-gray-700'
                ]"
              >
                {{ cat.label }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
