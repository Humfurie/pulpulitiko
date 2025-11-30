<script setup lang="ts">
const api = useApi()
const route = useRoute()

const currentPage = computed(() => Number(route.query.page) || 1)

const { data: articlesData, error, status } = await useAsyncData(
  `all-articles-page-${currentPage.value}`,
  () => api.getArticles(currentPage.value, 12),
  { watch: [currentPage] }
)

function handlePageChange(page: number) {
  navigateTo({ query: { page: page > 1 ? page : undefined } })
}

useSeoMeta({
  title: 'All Articles - Pulpulitiko',
  ogTitle: 'All Articles - Pulpulitiko',
  description: 'Browse all Philippine political news articles and commentary',
  ogDescription: 'Browse all Philippine political news articles and commentary'
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950">
    <UContainer class="py-8 lg:py-12">
      <!-- Header -->
      <div class="mb-8">
        <h1 class="text-4xl md:text-5xl font-black text-gray-900 dark:text-white tracking-tight mb-2">
          All Articles
        </h1>
        <p class="text-gray-600 dark:text-gray-400">
          Browse all our political news and commentary
        </p>
      </div>

      <!-- Loading State -->
      <div v-if="status === 'pending'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div v-for="i in 12" :key="i" class="animate-pulse">
          <div class="bg-gray-200 dark:bg-gray-800 rounded-2xl aspect-[4/3] mb-4" />
          <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-1/4 mb-2" />
          <div class="h-6 bg-gray-200 dark:bg-gray-800 rounded w-3/4 mb-2" />
          <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-full" />
        </div>
      </div>

      <!-- Error State -->
      <UAlert
        v-else-if="error"
        color="error"
        icon="i-heroicons-exclamation-triangle"
        title="Error loading articles"
        :description="error.message"
        class="mb-8"
      />

      <!-- Articles Grid -->
      <div v-else-if="articlesData?.articles?.length">
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <ArticleCard
            v-for="article in articlesData.articles"
            :key="article.id"
            :article="article"
          />
        </div>

        <!-- Pagination -->
        <div v-if="articlesData.total_pages > 1" class="mt-12 flex justify-center">
          <AppPagination
            :current-page="articlesData.page"
            :total-pages="articlesData.total_pages"
            @change="handlePageChange"
          />
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-20">
        <div class="w-20 h-20 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center mx-auto mb-6">
          <UIcon name="i-heroicons-document-text" class="w-10 h-10 text-gray-400" />
        </div>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">No articles yet</h3>
        <p class="text-gray-500 dark:text-gray-400">Check back soon for the latest political news.</p>
      </div>
    </UContainer>
  </div>
</template>
