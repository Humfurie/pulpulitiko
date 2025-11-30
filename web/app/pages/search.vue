<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const api = useApi()

const searchQuery = ref((route.query.q as string) || '')
const currentPage = computed(() => Number(route.query.page) || 1)

const { data, error, status, refresh } = await useAsyncData(
  `search-${searchQuery.value}-page-${currentPage.value}`,
  () => {
    if (!searchQuery.value.trim()) {
      return Promise.resolve(null)
    }
    return api.searchArticles(searchQuery.value, currentPage.value, 10)
  },
  { watch: [() => route.query.q, currentPage] }
)

function handleSearch() {
  if (searchQuery.value.trim()) {
    router.push({
      query: { q: searchQuery.value.trim(), page: undefined }
    })
    refresh()
  }
}

function handlePageChange(page: number) {
  navigateTo({
    query: {
      q: searchQuery.value,
      page: page > 1 ? page : undefined
    }
  })
}

watch(() => route.query.q, (newQuery) => {
  searchQuery.value = (newQuery as string) || ''
})

useSeoMeta({
  title: () => searchQuery.value ? `Search: ${searchQuery.value} - Pulpulitiko` : 'Search - Pulpulitiko',
  description: 'Search for political articles and news on Pulpulitiko'
})
</script>

<template>
  <div>
    <UContainer class="py-8">
      <!-- Search Header -->
      <div class="mb-8">
        <h1 class="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-6">
          Search Articles
        </h1>
        <form class="max-w-xl" @submit.prevent="handleSearch">
          <div class="flex gap-2">
            <UInput
              v-model="searchQuery"
              placeholder="Search for articles..."
              icon="i-heroicons-magnifying-glass"
              size="lg"
              class="flex-1"
            />
            <UButton type="submit" size="lg">
              Search
            </UButton>
          </div>
        </form>
      </div>

      <!-- No Query State -->
      <div v-if="!route.query.q" class="text-center py-12">
        <UIcon name="i-heroicons-magnifying-glass" class="w-16 h-16 text-gray-300 dark:text-gray-700 mx-auto" />
        <p class="mt-4 text-gray-500 dark:text-gray-400">
          Enter a search term to find articles
        </p>
      </div>

      <!-- Loading State -->
      <div v-else-if="status === 'pending'" class="animate-pulse">
        <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-1/4 mb-8" />
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div v-for="i in 4" :key="i">
            <div class="h-48 bg-gray-200 dark:bg-gray-800 rounded mb-4" />
            <div class="h-6 bg-gray-200 dark:bg-gray-800 rounded w-3/4" />
          </div>
        </div>
      </div>

      <!-- Error State -->
      <UAlert
        v-else-if="error"
        color="error"
        icon="i-heroicons-exclamation-triangle"
        title="Search failed"
        :description="error.message"
      />

      <!-- Search Results -->
      <div v-else-if="data">
        <p class="mb-6 text-gray-600 dark:text-gray-400">
          Found {{ data.total }} result{{ data.total !== 1 ? 's' : '' }} for
          <span class="font-semibold text-gray-900 dark:text-white">"{{ route.query.q }}"</span>
        </p>

        <!-- Results Grid -->
        <div v-if="data.articles?.length" class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <ArticleCard
            v-for="article in data.articles"
            :key="article.id"
            :article="article"
          />
        </div>

        <!-- No Results -->
        <div v-else class="text-center py-12">
          <UIcon name="i-heroicons-document-magnifying-glass" class="w-16 h-16 text-gray-300 dark:text-gray-700 mx-auto" />
          <p class="mt-4 text-gray-500 dark:text-gray-400">
            No articles found matching your search
          </p>
          <p class="mt-2 text-sm text-gray-400 dark:text-gray-500">
            Try different keywords or browse categories
          </p>
        </div>

        <!-- Pagination -->
        <AppPagination
          v-if="data.total_pages > 1"
          :current-page="data.page"
          :total-pages="data.total_pages"
          class="mt-8"
          @change="handlePageChange"
        />
      </div>
    </UContainer>
  </div>
</template>
