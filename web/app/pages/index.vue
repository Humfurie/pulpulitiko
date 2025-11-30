<script setup lang="ts">
const api = useApi()
const route = useRoute()

const currentPage = computed(() => Number(route.query.page) || 1)

const { data: articlesData, error, status } = await useAsyncData(
  `articles-page-${currentPage.value}`,
  () => api.getArticles(currentPage.value, 12),
  { watch: [currentPage] }
)

const { data: trendingArticles } = await useAsyncData(
  'trending-articles',
  () => api.getTrendingArticles()
)

function handlePageChange(page: number) {
  navigateTo({ query: { page: page > 1 ? page : undefined } })
}

useSeoMeta({
  title: 'Pulpulitiko - Philippine Politics News',
  ogTitle: 'Pulpulitiko - Philippine Politics News',
  description: 'Your trusted source for Philippine political news and commentary',
  ogDescription: 'Your trusted source for Philippine political news and commentary'
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950">
    <UContainer class="py-8 lg:py-12">
      <!-- Hero Section with Featured Article -->
      <section v-if="articlesData?.articles?.length" class="mb-12 lg:mb-16">
        <div class="flex items-center justify-between mb-8">
          <h1 class="text-4xl md:text-5xl lg:text-6xl font-black text-gray-900 dark:text-white tracking-tight">
            Latest News
          </h1>
          <NuxtLink
            to="/articles"
            class="hidden sm:flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-primary transition-colors font-medium"
          >
            See all posts
            <UIcon name="i-heroicons-arrow-right" class="w-4 h-4" />
          </NuxtLink>
        </div>

        <!-- Hero Grid: Featured + Side cards -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <!-- Main Featured Article -->
          <div class="lg:col-span-2">
            <ArticleCard :article="articlesData.articles[0]" featured />
          </div>

          <!-- Side Cards -->
          <div class="flex flex-col gap-6">
            <!-- Trending/Editor's Pick Card -->
            <div class="bg-amber-50 dark:bg-amber-950/30 rounded-2xl p-6 flex-1">
              <div class="flex items-center gap-2 mb-4">
                <span class="w-2 h-2 bg-amber-500 rounded-full animate-pulse" />
                <span class="text-sm font-semibold text-amber-700 dark:text-amber-400 uppercase tracking-wide">Trending</span>
              </div>
              <div v-if="trendingArticles?.length" class="space-y-4">
                <ArticleCard
                  v-for="article in trendingArticles.slice(0, 3)"
                  :key="article.id"
                  :article="article"
                  variant="horizontal"
                />
              </div>
              <p v-else class="text-gray-500 dark:text-gray-400 text-sm">
                No trending articles yet.
              </p>
            </div>

            <!-- Second Article Card (if exists) -->
            <ArticleCard
              v-if="articlesData.articles[1]"
              :article="articlesData.articles[1]"
              variant="overlay"
            />
          </div>
        </div>
      </section>

      <!-- Loading State -->
      <div v-if="status === 'pending'" class="space-y-8">
        <div class="animate-pulse">
          <div class="h-12 bg-gray-200 dark:bg-gray-800 rounded-lg w-1/3 mb-8" />
          <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div class="lg:col-span-2 bg-gray-200 dark:bg-gray-800 rounded-3xl aspect-[21/9]" />
            <div class="space-y-6">
              <div class="bg-gray-200 dark:bg-gray-800 rounded-2xl h-48" />
              <div class="bg-gray-200 dark:bg-gray-800 rounded-2xl aspect-[4/5]" />
            </div>
          </div>
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

      <!-- More Articles Section -->
      <section v-if="articlesData?.articles?.length && articlesData.articles.length > 2" class="mb-12">
        <div class="flex items-center justify-between mb-8">
          <h2 class="text-2xl md:text-3xl font-bold text-gray-900 dark:text-white">
            More Stories
          </h2>
        </div>

        <!-- Articles Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <ArticleCard
            v-for="article in articlesData.articles.slice(2)"
            :key="article.id"
            :article="article"
          />
        </div>

        <!-- Pagination -->
        <div v-if="articlesData.total_pages > 1" class="mt-12 flex justify-center">
          <Pagination
            :current-page="articlesData.page"
            :total-pages="articlesData.total_pages"
            @change="handlePageChange"
          />
        </div>
      </section>

      <!-- Empty State -->
      <div
        v-if="!status && !error && !articlesData?.articles?.length"
        class="text-center py-20"
      >
        <div class="w-20 h-20 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center mx-auto mb-6">
          <UIcon name="i-heroicons-document-text" class="w-10 h-10 text-gray-400" />
        </div>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">No articles yet</h3>
        <p class="text-gray-500 dark:text-gray-400">Check back soon for the latest political news.</p>
      </div>
    </UContainer>
  </div>
</template>
