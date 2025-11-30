<script setup lang="ts">
const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)
const currentPage = computed(() => Number(route.query.page) || 1)

const { data, error, status } = await useAsyncData(
  `tag-${slug.value}-page-${currentPage.value}`,
  () => api.getTagArticles(slug.value, currentPage.value, 10),
  { watch: [currentPage] }
)

function handlePageChange(page: number) {
  navigateTo({
    params: { slug: slug.value },
    query: { page: page > 1 ? page : undefined }
  })
}

useSeoMeta({
  title: () => data.value?.tag?.name ? `#${data.value.tag.name} - Pulpulitiko` : 'Tag',
  ogTitle: () => data.value?.tag?.name ? `#${data.value.tag.name}` : undefined,
  description: () => `Articles tagged with ${data.value?.tag?.name}`,
  ogDescription: () => `Articles tagged with ${data.value?.tag?.name}`
})
</script>

<template>
  <div>
    <UContainer class="py-8">
      <!-- Loading State -->
      <div v-if="status === 'pending'" class="animate-pulse">
        <div class="h-10 bg-gray-200 dark:bg-gray-800 rounded w-1/3 mb-8" />
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
          title="Tag not found"
          description="The tag you're looking for doesn't exist."
        />
        <div class="mt-6">
          <UButton to="/" variant="outline">
            Back to Home
          </UButton>
        </div>
      </div>

      <!-- Tag Content -->
      <div v-else-if="data">
        <!-- Header -->
        <div class="mb-8">
          <div class="flex items-center gap-3">
            <UIcon name="i-heroicons-hashtag" class="w-8 h-8 text-primary" />
            <h1 class="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white">
              {{ data.tag.name }}
            </h1>
          </div>
          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
            {{ data.articles.total }} article{{ data.articles.total !== 1 ? 's' : '' }}
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
          description="There are no articles with this tag yet."
        />

        <!-- Pagination -->
        <Pagination
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
