<script setup lang="ts">
const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)
const currentPage = computed(() => Number(route.query.page) || 1)

const { data, error, status } = await useAsyncData(
  `author-${slug.value}-page-${currentPage.value}`,
  () => api.getAuthorArticles(slug.value, currentPage.value, 10),
  { watch: [currentPage] }
)

function handlePageChange(page: number) {
  navigateTo({
    params: { slug: slug.value },
    query: { page: page > 1 ? page : undefined }
  })
}

useSeoMeta({
  title: () => data.value?.author?.name ? `${data.value.author.name} - Pulpulitiko` : 'Author',
  ogTitle: () => data.value?.author?.name,
  description: () => data.value?.author?.bio || `Articles by ${data.value?.author?.name}`,
  ogDescription: () => data.value?.author?.bio || `Articles by ${data.value?.author?.name}`
})
</script>

<template>
  <div>
    <UContainer class="py-8">
      <!-- Loading State -->
      <div v-if="status === 'pending'" class="animate-pulse">
        <div class="flex items-center gap-4 mb-6">
          <div class="w-20 h-20 bg-gray-200 dark:bg-gray-800 rounded-full" />
          <div>
            <div class="h-8 bg-gray-200 dark:bg-gray-800 rounded w-48 mb-2" />
            <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-32" />
          </div>
        </div>
        <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-3/4 mb-8" />
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div v-for="i in 4" :key="i">
            <div class="h-48 bg-gray-200 dark:bg-gray-800 rounded mb-4" />
            <div class="h-6 bg-gray-200 dark:bg-gray-800 rounded w-3/4" />
          </div>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error">
        <UAlert
          color="error"
          icon="i-heroicons-exclamation-triangle"
          title="Author not found"
          description="The author you're looking for doesn't exist."
        />
        <div class="mt-6">
          <UButton to="/" variant="outline">
            Back to Home
          </UButton>
        </div>
      </div>

      <!-- Author Content -->
      <div v-else-if="data">
        <!-- Author Header -->
        <div class="mb-8">
          <div class="flex items-start gap-4 mb-4">
            <UAvatar
              v-if="data.author.avatar"
              :src="data.author.avatar"
              :alt="data.author.name"
              size="xl"
            />
            <UAvatar
              v-else
              :alt="data.author.name"
              size="xl"
              icon="i-heroicons-user"
            />
            <div>
              <h1 class="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white">
                {{ data.author.name }}
              </h1>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ data.articles.total }} article{{ data.articles.total !== 1 ? 's' : '' }}
              </p>
            </div>
          </div>
          <p v-if="data.author.bio" class="text-lg text-gray-600 dark:text-gray-400 max-w-3xl">
            {{ data.author.bio }}
          </p>
        </div>

        <!-- Articles Grid -->
        <div v-if="data.articles.articles?.length" class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <ArticleCard
            v-for="article in data.articles.articles"
            :key="article.id"
            :article="article"
          />
        </div>

        <!-- Empty State -->
        <UAlert
          v-else
          icon="i-heroicons-document-text"
          title="No articles yet"
          description="This author hasn't published any articles yet."
        />

        <!-- Pagination -->
        <AppPagination
          v-if="data.articles.total_pages > 1"
          :current-page="data.articles.page"
          :total-pages="data.articles.total_pages"
          class="mt-8"
          @change="handlePageChange"
        />
      </div>
    </UContainer>
  </div>
</template>
