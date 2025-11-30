<script setup lang="ts">
const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)

const { data: article, error, status } = await useAsyncData(
  `article-${slug.value}`,
  () => api.getArticleBySlug(slug.value)
)

const { data: trendingArticles } = await useAsyncData(
  'article-trending',
  () => api.getTrendingArticles()
)

// Track article view (only on client-side)
onMounted(() => {
  if (article.value) {
    api.trackArticleView(slug.value)
  }
})

function formatDate(dateString?: string): string {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// SEO
useSeoMeta({
  title: () => article.value?.title || 'Article',
  ogTitle: () => article.value?.title,
  description: () => article.value?.summary || '',
  ogDescription: () => article.value?.summary || '',
  ogImage: () => article.value?.featured_image,
  ogType: 'article',
  articlePublishedTime: () => article.value?.published_at,
  articleAuthor: () => article.value?.author?.name
})

// Schema.org structured data
useHead({
  script: [
    {
      type: 'application/ld+json',
      children: computed(() => JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'NewsArticle',
        headline: article.value?.title,
        description: article.value?.summary,
        image: article.value?.featured_image,
        datePublished: article.value?.published_at,
        dateModified: article.value?.updated_at,
        author: article.value?.author ? {
          '@type': 'Person',
          name: article.value.author.name
        } : undefined,
        publisher: {
          '@type': 'Organization',
          name: 'Pulpulitiko',
          logo: {
            '@type': 'ImageObject',
            url: '/logo.png'
          }
        }
      }))
    }
  ]
})
</script>

<template>
  <div>
    <!-- Loading State -->
    <UContainer v-if="status === 'pending'" class="py-8">
      <div class="max-w-4xl mx-auto animate-pulse">
        <div class="h-8 bg-gray-200 dark:bg-gray-800 rounded w-3/4 mb-4" />
        <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-1/4 mb-8" />
        <div class="h-96 bg-gray-200 dark:bg-gray-800 rounded mb-8" />
        <div class="space-y-4">
          <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded" />
          <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded" />
          <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-3/4" />
        </div>
      </div>
    </UContainer>

    <!-- Error State -->
    <UContainer v-else-if="error" class="py-8">
      <div class="max-w-4xl mx-auto">
        <UAlert
          color="error"
          icon="i-heroicons-exclamation-triangle"
          title="Article not found"
          description="The article you're looking for doesn't exist or has been removed."
        />
        <div class="mt-6">
          <UButton to="/" variant="outline">
            Back to Home
          </UButton>
        </div>
      </div>
    </UContainer>

    <!-- Article Content -->
    <article v-else-if="article">
      <!-- Hero Image -->
      <div v-if="article.featured_image" class="relative h-[400px] md:h-[500px] w-full">
        <NuxtImg
          :src="article.featured_image"
          :alt="article.title"
          class="w-full h-full object-cover"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent" />
      </div>

      <UContainer class="py-8">
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <!-- Main Content -->
          <div class="lg:col-span-2">
            <div class="max-w-none">
              <!-- Category -->
              <NuxtLink
                v-if="article.category"
                :to="`/category/${article.category.slug}`"
                class="inline-block mb-4"
              >
                <UBadge color="primary" variant="subtle" size="lg">
                  {{ article.category.name }}
                </UBadge>
              </NuxtLink>

              <!-- Title -->
              <h1 class="text-3xl md:text-4xl lg:text-5xl font-bold text-gray-900 dark:text-white leading-tight">
                {{ article.title }}
              </h1>

              <!-- Meta -->
              <div class="flex flex-wrap items-center gap-4 mt-6 text-gray-500 dark:text-gray-400">
                <div v-if="article.author" class="flex items-center gap-2">
                  <NuxtImg
                    v-if="article.author.avatar"
                    :src="article.author.avatar"
                    :alt="article.author.name"
                    class="w-10 h-10 rounded-full object-cover"
                  />
                  <div v-else class="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                    <UIcon name="i-heroicons-user" class="w-5 h-5 text-primary" />
                  </div>
                  <span class="font-medium text-gray-900 dark:text-white">
                    {{ article.author.name }}
                  </span>
                </div>
                <span v-if="article.published_at" class="flex items-center gap-1">
                  <UIcon name="i-heroicons-calendar" class="w-4 h-4" />
                  {{ formatDate(article.published_at) }}
                </span>
              </div>

              <!-- Share Buttons -->
              <div class="mt-6 pb-6 border-b border-gray-200 dark:border-gray-800">
                <ShareButtons :title="article.title" />
              </div>

              <!-- Summary -->
              <p
                v-if="article.summary"
                class="mt-6 text-xl text-gray-600 dark:text-gray-300 leading-relaxed"
              >
                {{ article.summary }}
              </p>

              <!-- Content -->
              <div
                class="mt-8 prose prose-lg dark:prose-invert max-w-none prose-headings:font-bold prose-a:text-primary"
                v-html="article.content"
              />

              <!-- Tags -->
              <div v-if="article.tags?.length" class="mt-8 pt-8 border-t border-gray-200 dark:border-gray-800">
                <h3 class="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-3">
                  Tags
                </h3>
                <div class="flex flex-wrap gap-2">
                  <NuxtLink
                    v-for="tag in article.tags"
                    :key="tag.id"
                    :to="`/tag/${tag.slug}`"
                  >
                    <UBadge variant="outline">
                      {{ tag.name }}
                    </UBadge>
                  </NuxtLink>
                </div>
              </div>

              <!-- Share Buttons (Bottom) -->
              <div class="mt-8 pt-8 border-t border-gray-200 dark:border-gray-800">
                <p class="text-sm text-gray-500 dark:text-gray-400 mb-3">
                  Enjoyed this article? Share it!
                </p>
                <ShareButtons :title="article.title" />
              </div>
            </div>
          </div>

          <!-- Sidebar -->
          <aside class="lg:col-span-1">
            <div class="sticky top-24">
              <!-- Author Bio -->
              <div v-if="article.author" class="bg-gray-50 dark:bg-gray-900 rounded-lg p-6 mb-6">
                <h3 class="font-semibold text-gray-900 dark:text-white mb-4">About the Author</h3>
                <div class="flex items-center gap-4">
                  <NuxtImg
                    v-if="article.author.avatar"
                    :src="article.author.avatar"
                    :alt="article.author.name"
                    class="w-16 h-16 rounded-full object-cover"
                  />
                  <div v-else class="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center">
                    <UIcon name="i-heroicons-user" class="w-8 h-8 text-primary" />
                  </div>
                  <div>
                    <p class="font-medium text-gray-900 dark:text-white">
                      {{ article.author.name }}
                    </p>
                  </div>
                </div>
                <p v-if="article.author.bio" class="mt-4 text-sm text-gray-600 dark:text-gray-400">
                  {{ article.author.bio }}
                </p>
              </div>

              <!-- Trending Articles -->
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white mb-4">
                  Trending Articles
                </h3>
                <div v-if="trendingArticles?.length" class="space-y-4">
                  <NuxtLink
                    v-for="trending in trendingArticles.filter(a => a.slug !== article.slug).slice(0, 5)"
                    :key="trending.id"
                    :to="`/article/${trending.slug}`"
                    class="block group"
                  >
                    <h4 class="font-medium text-gray-900 dark:text-white group-hover:text-primary transition-colors line-clamp-2">
                      {{ trending.title }}
                    </h4>
                    <p v-if="trending.category_name" class="text-sm text-gray-500 mt-1">
                      {{ trending.category_name }}
                    </p>
                  </NuxtLink>
                </div>
              </div>
            </div>
          </aside>
        </div>
      </UContainer>
    </article>
  </div>
</template>
