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
  <div class="min-h-screen bg-stone-50 dark:bg-stone-950 relative">
    <!-- Decorative background elements (contained within viewport) -->
    <div class="absolute top-0 right-0 w-96 h-96 bg-orange-200/20 dark:bg-orange-900/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/4 pointer-events-none" />
    <div class="absolute bottom-1/3 left-0 w-72 h-72 bg-amber-200/15 dark:bg-amber-900/10 rounded-full blur-3xl -translate-x-1/4 pointer-events-none" />

    <UContainer class="py-8 lg:py-12 relative">
      <!-- Hero Section with Featured Article -->
      <section v-if="articlesData?.articles?.length" class="mb-16 lg:mb-20">
        <div class="flex items-center justify-between mb-10">
          <div>
            <h1 class="text-4xl md:text-5xl lg:text-6xl font-extrabold text-stone-900 dark:text-white tracking-tight">
              <span class="inline-block">Latest</span>
              <span class="inline-block text-orange-500 ml-2">News</span>
            </h1>
          </div>
          <NuxtLink
            to="/articles"
            class="hidden sm:flex items-center gap-2 text-stone-600 dark:text-stone-400 hover:text-orange-500 transition-colors font-medium group"
          >
            See all posts
            <UIcon name="i-heroicons-arrow-right" class="w-4 h-4 transform group-hover:translate-x-1 transition-transform" />
          </NuxtLink>
        </div>

        <!-- Hero Grid: Featured + Side cards -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 lg:gap-8">
          <!-- Main Featured Article -->
          <div class="lg:col-span-2">
            <ArticleCard :article="articlesData.articles[0]" featured />
          </div>

          <!-- Side Cards -->
          <div class="flex flex-col gap-6">
            <!-- Trending/Editor's Pick Card -->
            <div class="bg-gradient-to-br from-orange-50 to-amber-50 dark:from-orange-950/40 dark:to-amber-950/30 rounded-2xl p-6 flex-1 border border-orange-100 dark:border-orange-900/30">
              <div class="flex items-center gap-2 mb-5">
                <span class="relative w-2 h-2">
                  <span class="absolute inset-0 bg-orange-500 rounded-full animate-ping opacity-75" />
                  <span class="relative block w-2 h-2 bg-orange-500 rounded-full" />
                </span>
                <span class="text-sm font-bold text-orange-700 dark:text-orange-400 uppercase tracking-wider">Trending</span>
              </div>
              <div v-if="trendingArticles?.length" class="space-y-3">
                <ArticleCard
                  v-for="article in trendingArticles.slice(0, 3)"
                  :key="article.id"
                  :article="article"
                  variant="horizontal"
                />
              </div>
              <p v-else class="text-stone-500 dark:text-stone-400 text-sm">
                No trending articles yet.
              </p>
            </div>

            <!-- Second Article Card (if exists) -->
            <div>
              <ArticleCard
                v-if="articlesData.articles[1]"
                :article="articlesData.articles[1]"
                variant="overlay"
              />
            </div>
          </div>
        </div>
      </section>

      <!-- Loading State -->
      <div v-if="status === 'pending'" class="space-y-8">
        <div class="animate-pulse">
          <div class="h-12 bg-stone-200 dark:bg-stone-800 rounded-lg w-1/3 mb-8" />
          <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div class="lg:col-span-2 bg-stone-200 dark:bg-stone-800 rounded-3xl aspect-[21/9]" />
            <div class="space-y-6">
              <div class="bg-stone-200 dark:bg-stone-800 rounded-2xl h-48" />
              <div class="bg-stone-200 dark:bg-stone-800 rounded-2xl aspect-[4/5]" />
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
        <div class="flex items-center justify-between mb-10">
          <div class="flex items-center gap-4">
            <div class="w-1 h-8 bg-gradient-to-b from-orange-500 to-amber-500 rounded-full" />
            <h2 class="text-2xl md:text-3xl font-bold text-stone-800 dark:text-white">
              More Stories
            </h2>
          </div>
        </div>

        <!-- Articles Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 lg:gap-8">
          <div
            v-for="article in articlesData.articles.slice(2)"
            :key="article.id"
          >
            <ArticleCard :article="article" />
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="articlesData.total_pages > 1" class="mt-16 flex justify-center">
          <AppPagination
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
        <div class="w-24 h-24 bg-gradient-to-br from-orange-100 to-amber-100 dark:from-orange-900/30 dark:to-amber-900/30 rounded-full flex items-center justify-center mx-auto mb-6 animate-float">
          <UIcon name="i-heroicons-newspaper" class="w-12 h-12 text-orange-400" />
        </div>
        <h3 class="text-xl font-bold text-stone-800 dark:text-white mb-2">No articles yet</h3>
        <p class="text-stone-500 dark:text-stone-400">Check back soon for the latest political news.</p>
      </div>
    </UContainer>
  </div>
</template>
