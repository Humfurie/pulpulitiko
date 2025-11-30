<script setup lang="ts">
const api = useApi()
const route = useRoute()

const currentPage = computed(() => Number(route.query.page) || 1)

const { data: articlesData, error, status } = await useAsyncData(
  `articles-page-${currentPage.value}`,
  () => api.getArticles(currentPage.value, 10),
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
  <div>
    <UContainer class="py-8">
      <!-- Hero Section with Featured Article -->
      <section v-if="articlesData?.articles?.length" class="mb-12">
        <ArticleCard :article="articlesData.articles[0]" featured />
      </section>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <!-- Main Content -->
        <div class="lg:col-span-2">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">
            Latest News
          </h2>

          <!-- Loading State -->
          <div v-if="status === 'pending'" class="space-y-6">
            <div v-for="i in 5" :key="i" class="animate-pulse">
              <div class="bg-gray-200 dark:bg-gray-800 h-48 rounded-lg mb-4" />
              <div class="h-6 bg-gray-200 dark:bg-gray-800 rounded w-3/4 mb-2" />
              <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-1/2" />
            </div>
          </div>

          <!-- Error State -->
          <UAlert
            v-else-if="error"
            color="error"
            icon="i-heroicons-exclamation-triangle"
            title="Error loading articles"
            :description="error.message"
          />

          <!-- Articles -->
          <div v-else-if="articlesData?.articles?.length" class="space-y-6">
            <ArticleCard
              v-for="article in articlesData.articles.slice(1)"
              :key="article.id"
              :article="article"
            />

            <!-- Pagination -->
            <Pagination
              v-if="articlesData.total_pages > 1"
              :current-page="articlesData.page"
              :total-pages="articlesData.total_pages"
              class="mt-8"
              @change="handlePageChange"
            />
          </div>

          <!-- Empty State -->
          <UAlert
            v-else
            icon="i-heroicons-document-text"
            title="No articles yet"
            description="Check back soon for the latest political news."
          />
        </div>

        <!-- Sidebar -->
        <aside class="lg:col-span-1">
          <!-- Trending Articles -->
          <div class="sticky top-24">
            <h3 class="text-xl font-bold text-gray-900 dark:text-white mb-4">
              Trending
            </h3>
            <div v-if="trendingArticles?.length" class="space-y-4">
              <NuxtLink
                v-for="(article, index) in trendingArticles.slice(0, 5)"
                :key="article.id"
                :to="`/article/${article.slug}`"
                class="flex gap-4 group"
              >
                <span class="text-3xl font-bold text-gray-300 dark:text-gray-700">
                  {{ String(index + 1).padStart(2, '0') }}
                </span>
                <div>
                  <h4 class="font-medium text-gray-900 dark:text-white group-hover:text-primary transition-colors line-clamp-2">
                    {{ article.title }}
                  </h4>
                  <p v-if="article.category_name" class="text-sm text-gray-500 mt-1">
                    {{ article.category_name }}
                  </p>
                </div>
              </NuxtLink>
            </div>
            <p v-else class="text-gray-500 dark:text-gray-400">
              No trending articles yet.
            </p>
          </div>
        </aside>
      </div>
    </UContainer>
  </div>
</template>
