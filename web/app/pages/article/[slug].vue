<script setup lang="ts">
const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)

// Helper function to sanitize text for Schema.org JSON-LD
// Strips HTML tags and ensures safe content for JSON serialization
function sanitizeForSchema(html: string | undefined): string | undefined {
  if (!html) return undefined

  // Strip HTML tags
  const withoutTags = html.replace(/<[^>]*>/g, '')

  // Decode common HTML entities to prevent double-encoding
  const decoded = withoutTags
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'")
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&amp;/g, '&')

  // Trim and normalize whitespace
  const normalized = decoded.replace(/\s+/g, ' ').trim()

  // Truncate to reasonable length (Google recommends max 5000 characters for articleBody)
  return normalized.length > 5000 ? normalized.substring(0, 4997) + '...' : normalized
}

const { data: article, error, status } = await useAsyncData(
  `article-${slug.value}`,
  () => api.getArticleBySlug(slug.value)
)

const { data: relatedArticles } = await useAsyncData(
  `article-related-${slug.value}`,
  () => api.getRelatedArticles(slug.value),
  { watch: [slug] }
)

// Calculate reading time (average 200 words per minute)
const readingTime = computed(() => {
  if (!article.value?.content) return 0
  // Strip HTML tags and count words
  const text = article.value.content.replace(/<[^>]*>/g, '')
  const words = text.trim().split(/\s+/).length
  return Math.max(1, Math.ceil(words / 200))
})

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
    day: 'numeric'
  })
}

// SEO
const config = useRuntimeConfig()
const siteUrl = config.public.siteUrl

// Ensure image URL is absolute for social media
const ogImageUrl = computed(() => {
  const img = article.value?.featured_image
  if (!img) return undefined
  if (img.startsWith('http')) return img
  return `${siteUrl}${img.startsWith('/') ? '' : '/'}${img}`
})

// Generate keywords from tags and category
const keywords = computed(() => {
  const tags = article.value?.tags?.map(t => t.name) || []
  const category = article.value?.category?.name
  const allKeywords = category ? [category, ...tags] : tags
  return allKeywords.slice(0, 5).join(', ')
})

useSeoMeta({
  title: () => article.value?.title || 'Article',
  ogTitle: () => article.value?.title,
  description: () => article.value?.summary || '',
  ogDescription: () => article.value?.summary || '',
  keywords: () => keywords.value,
  ogImage: () => ogImageUrl.value,
  ogImageWidth: 1200,
  ogImageHeight: 630,
  ogType: 'article',
  ogUrl: () => `${siteUrl}/article/${slug.value}`,
  ogLocale: 'en_PH',
  articlePublishedTime: computed(() => article.value?.published_at),
  articleAuthor: computed(() => article.value?.author?.name ? [article.value.author.name] : undefined),
  twitterCard: 'summary_large_image',
  twitterTitle: () => article.value?.title,
  twitterDescription: () => article.value?.summary || '',
  twitterImage: () => ogImageUrl.value
})

// Canonical URL
useHead({
  link: [
    { rel: 'canonical', href: `${siteUrl}/article/${slug.value}` }
  ]
})

// Schema.org structured data
useHead({
  script: [
    {
      type: 'application/ld+json',
      innerHTML: computed(() => {
        // Sanitize content for articleBody (strips HTML, decodes entities, truncates)
        const articleBody = sanitizeForSchema(article.value?.content)

        const schema: Record<string, unknown> = {
          '@context': 'https://schema.org',
          '@type': 'NewsArticle',
          mainEntityOfPage: {
            '@type': 'WebPage',
            '@id': `${siteUrl}/article/${slug.value}`
          },
          headline: article.value?.title,
          description: article.value?.summary,
          image: ogImageUrl.value ? [ogImageUrl.value] : undefined,
          datePublished: article.value?.published_at,
          dateModified: article.value?.updated_at || article.value?.published_at,
          articleBody,
          author: article.value?.author ? {
            '@type': 'Person',
            name: article.value.author.name,
            url: article.value.author.slug ? `${siteUrl}/user/${article.value.author.slug}` : undefined,
            image: article.value.author.avatar ? {
              '@type': 'ImageObject',
              url: article.value.author.avatar
            } : undefined
          } : undefined,
          publisher: {
            '@type': 'Organization',
            name: 'Pulpulitiko',
            url: siteUrl,
            logo: {
              '@type': 'ImageObject',
              url: `${siteUrl}/pulpulitiko.png`,
              width: 512,
              height: 512
            }
          }
        }

        // Remove undefined fields
        Object.keys(schema).forEach(key => {
          if (schema[key] === undefined) {
            delete schema[key]
          }
        })

        return JSON.stringify(schema)
      })
    },
    {
      type: 'application/ld+json',
      innerHTML: computed(() => {
        const breadcrumbItems = [
          {
            '@type': 'ListItem',
            position: 1,
            name: 'Home',
            item: siteUrl
          }
        ]

        if (article.value?.category) {
          breadcrumbItems.push({
            '@type': 'ListItem',
            position: 2,
            name: article.value.category.name,
            item: `${siteUrl}/category/${article.value.category.slug}`
          })
        }

        breadcrumbItems.push({
          '@type': 'ListItem',
          position: article.value?.category ? 3 : 2,
          name: article.value?.title || 'Article',
          item: `${siteUrl}/article/${slug.value}`
        })

        return JSON.stringify({
          '@context': 'https://schema.org',
          '@type': 'BreadcrumbList',
          itemListElement: breadcrumbItems
        })
      })
    }
  ]
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950">
    <!-- Loading State -->
    <div v-if="status === 'pending'" class="max-w-3xl mx-auto px-4 py-16">
      <div class="animate-pulse">
        <div class="h-6 bg-gray-200 dark:bg-gray-800 rounded-full w-24 mx-auto mb-6" />
        <div class="h-12 bg-gray-200 dark:bg-gray-800 rounded w-3/4 mx-auto mb-4" />
        <div class="h-12 bg-gray-200 dark:bg-gray-800 rounded w-1/2 mx-auto mb-8" />
        <div class="h-80 bg-gray-200 dark:bg-gray-800 rounded-2xl mb-8" />
        <div class="space-y-4">
          <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded" />
          <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded" />
          <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-3/4" />
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="max-w-3xl mx-auto px-4 py-16">
      <div class="text-center">
        <div class="w-20 h-20 mx-auto mb-6 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
          <UIcon name="i-heroicons-exclamation-triangle" class="w-10 h-10 text-red-500" />
        </div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">Article not found</h1>
        <p class="text-gray-500 dark:text-gray-400 mb-6">
          The article you're looking for doesn't exist or has been removed.
        </p>
        <UButton to="/" color="primary" size="lg">
          Back to Home
        </UButton>
      </div>
    </div>

    <!-- Article Content -->
    <article v-else-if="article" class="pb-16">
      <!-- Breadcrumbs -->
      <nav class="max-w-3xl mx-auto px-4 pt-8 pb-4" aria-label="Breadcrumb">
        <ol class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <li>
            <NuxtLink to="/" class="hover:text-primary transition-colors">
              <UIcon name="i-heroicons-home" class="w-4 h-4" />
            </NuxtLink>
          </li>
          <li>
            <UIcon name="i-heroicons-chevron-right" class="w-3 h-3" />
          </li>
          <li v-if="article.category">
            <NuxtLink
              :to="`/category/${article.category.slug}`"
              class="hover:text-primary transition-colors"
            >
              {{ article.category.name }}
            </NuxtLink>
          </li>
          <li v-if="article.category">
            <UIcon name="i-heroicons-chevron-right" class="w-3 h-3" />
          </li>
          <li class="text-gray-900 dark:text-white font-medium truncate">
            {{ article.title }}
          </li>
        </ol>
      </nav>

      <!-- Header Section -->
      <header class="max-w-3xl mx-auto px-4 pt-4 pb-8 text-center">
        <!-- Category Tags -->
        <div class="flex items-center justify-center gap-2 mb-6">
          <NuxtLink
            v-if="article.category"
            :to="`/category/${article.category.slug}`"
            class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-primary/10 text-primary hover:bg-primary/20 transition-colors"
          >
            {{ article.category.name }}
          </NuxtLink>
          <template v-if="article.tags?.length">
            <NuxtLink
              v-for="tag in article.tags.slice(0, 2)"
              :key="tag.id"
              :to="`/tag/${tag.slug}`"
              class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
            >
              {{ tag.name }}
            </NuxtLink>
          </template>
        </div>

        <!-- Title -->
        <h1 class="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900 dark:text-white leading-tight mb-8">
          {{ article.title }}
        </h1>

        <!-- Author & Meta -->
        <div class="flex flex-col sm:flex-row items-center justify-center gap-4 sm:gap-6">
          <!-- Author -->
          <NuxtLink
            v-if="article.author"
            :to="`/user/${article.author.slug}`"
            class="flex items-center gap-3 hover:opacity-80 transition-opacity"
          >
            <NuxtImg
              v-if="article.author.avatar"
              :src="article.author.avatar"
              :alt="article.author.name"
              class="w-12 h-12 rounded-full object-cover ring-2 ring-white dark:ring-gray-900 shadow-md"
            />
            <div v-else class="w-12 h-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center ring-2 ring-white dark:ring-gray-900 shadow-md">
              <span class="text-white font-semibold text-lg">{{ article.author.name.charAt(0) }}</span>
            </div>
            <div class="text-left">
              <p class="font-semibold text-gray-900 dark:text-white hover:text-primary transition-colors">
                {{ article.author.name }}
              </p>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                Contributor
              </p>
            </div>
          </NuxtLink>

          <!-- Divider (visible on larger screens) -->
          <div class="hidden sm:block w-px h-10 bg-gray-200 dark:bg-gray-700" />

          <!-- Date & Reading Time -->
          <div class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
            <div class="flex items-center gap-1.5">
              <UIcon name="i-heroicons-calendar" class="w-4 h-4" />
              <span>{{ formatDate(article.published_at) }}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <UIcon name="i-heroicons-clock" class="w-4 h-4" />
              <span>{{ readingTime }} min read</span>
            </div>
          </div>
        </div>
      </header>

      <!-- Featured Image -->
      <div v-if="article.featured_image" class="max-w-4xl mx-auto px-4 mb-12">
        <div class="relative overflow-hidden rounded-2xl shadow-2xl">
          <NuxtImg
            :src="article.featured_image"
            :alt="article.title"
            class="w-full h-auto object-cover"
          />
        </div>
      </div>

      <!-- Article Body -->
      <div class="max-w-3xl mx-auto px-4">
        <!-- Summary/Lead -->
        <p
          v-if="article.summary"
          class="text-xl sm:text-2xl text-gray-600 dark:text-gray-300 leading-relaxed mb-10 font-medium"
        >
          {{ article.summary }}
        </p>

        <!-- Content -->
        <!-- eslint-disable-next-line vue/no-v-html -->
        <div
          class="article-content prose prose-lg dark:prose-invert max-w-none"
          v-html="article.content"
        />

        <!-- Tags Section -->
        <div v-if="article.tags?.length" class="mt-12 pt-8 border-t border-gray-200 dark:border-gray-800">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-sm font-medium text-gray-500 dark:text-gray-400 mr-2">Tags:</span>
            <NuxtLink
              v-for="tag in article.tags"
              :key="tag.id"
              :to="`/tag/${tag.slug}`"
              class="inline-flex items-center px-3 py-1.5 rounded-full text-sm font-medium bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-primary/10 hover:text-primary transition-colors"
            >
              #{{ tag.name }}
            </NuxtLink>
          </div>
        </div>

        <!-- Share Section -->
        <div class="mt-10 pt-8 border-t border-gray-200 dark:border-gray-800">
          <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
            <p class="text-gray-600 dark:text-gray-400 font-medium">
              Enjoyed this article? Share it with others!
            </p>
            <ShareButtons :title="article.title" />
          </div>
        </div>

        <!-- Author Card -->
        <NuxtLink
          v-if="article.author"
          :to="`/user/${article.author.slug}`"
          class="block mt-10 p-6 bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm hover:border-primary/30 hover:shadow-md transition-all"
        >
          <div class="flex items-start gap-4">
            <NuxtImg
              v-if="article.author.avatar"
              :src="article.author.avatar"
              :alt="article.author.name"
              class="w-16 h-16 rounded-full object-cover flex-shrink-0"
            />
            <div v-else class="w-16 h-16 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center flex-shrink-0">
              <span class="text-white font-bold text-2xl">{{ article.author.name.charAt(0) }}</span>
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1">
                <h3 class="font-bold text-gray-900 dark:text-white text-lg group-hover:text-primary transition-colors">
                  {{ article.author.name }}
                </h3>
                <span class="px-2 py-0.5 text-xs font-medium bg-primary/10 text-primary rounded-full">
                  Author
                </span>
              </div>
              <p v-if="article.author.bio" class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed">
                {{ article.author.bio }}
              </p>
              <p v-else class="text-gray-500 dark:text-gray-500 text-sm">
                Contributing writer at Pulpulitiko
              </p>
            </div>
            <UIcon name="i-heroicons-arrow-right" class="w-5 h-5 text-gray-400 flex-shrink-0" />
          </div>
        </NuxtLink>

        <!-- Related Articles -->
        <div v-if="relatedArticles?.length" class="mt-12 pt-8 border-t border-gray-200 dark:border-gray-800">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">
            Related Articles
          </h2>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <NuxtLink
              v-for="related in relatedArticles.slice(0, 4)"
              :key="related.id"
              :to="`/article/${related.slug}`"
              class="group block"
            >
              <div class="bg-white dark:bg-gray-900 rounded-xl overflow-hidden border border-gray-200 dark:border-gray-800 hover:shadow-lg hover:border-primary/30 transition-all duration-300">
                <div class="relative h-40 overflow-hidden">
                  <NuxtImg
                    v-if="related.featured_image"
                    :src="related.featured_image"
                    :alt="related.title"
                    class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                  />
                  <div v-else class="w-full h-full bg-gradient-to-br from-gray-100 to-gray-200 dark:from-gray-800 dark:to-gray-700" />
                </div>
                <div class="p-4">
                  <p v-if="related.category_name" class="text-xs font-semibold text-primary mb-2">
                    {{ related.category_name }}
                  </p>
                  <h3 class="font-semibold text-gray-900 dark:text-white group-hover:text-primary transition-colors line-clamp-2">
                    {{ related.title }}
                  </h3>
                </div>
              </div>
            </NuxtLink>
          </div>
        </div>

        <!-- Comments Section -->
        <CommentSection :article-slug="slug" />
      </div>
    </article>
  </div>
</template>

<style>
/* Article Content Styling - Improved contrast */
.article-content {
  color: #374151;
}

.dark .article-content {
  color: #d1d5db !important;
}

/* Force override inline styles in dark mode */
.dark .article-content * {
  color: inherit !important;
}

.dark .article-content a {
  color: var(--ui-primary) !important;
}

.article-content h1,
.article-content h2,
.article-content h3,
.article-content h4,
.article-content h5,
.article-content h6 {
  font-weight: 700;
  color: #111827;
  margin-top: 2.5rem;
  margin-bottom: 1rem;
}

.dark .article-content h1,
.dark .article-content h2,
.dark .article-content h3,
.dark .article-content h4,
.dark .article-content h5,
.dark .article-content h6 {
  color: #f9fafb;
}

.article-content h2 {
  font-size: 1.875rem;
  line-height: 2.25rem;
}

.article-content h3 {
  font-size: 1.5rem;
  line-height: 2rem;
}

.article-content p {
  font-size: 1.125rem;
  line-height: 1.85;
  margin-bottom: 1.5rem;
}

.article-content a {
  color: var(--ui-primary);
  font-weight: 500;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.article-content a:hover {
  opacity: 0.8;
}

.article-content strong {
  font-weight: 600;
  color: #111827;
}

.dark .article-content strong {
  color: #f9fafb;
}

.article-content em {
  font-style: italic;
}

.article-content ul,
.article-content ol {
  margin-top: 1.5rem;
  margin-bottom: 1.5rem;
  padding-left: 1.5rem;
}

.article-content li {
  margin-bottom: 0.75rem;
  font-size: 1.125rem;
  line-height: 1.75;
}

.article-content ul li {
  list-style-type: disc;
}

.article-content ol li {
  list-style-type: decimal;
}

/* Blockquote - Better contrast in dark mode */
.article-content blockquote {
  position: relative;
  margin: 2rem 0;
  padding: 1.5rem 2rem;
  background-color: #f3f4f6;
  border-radius: 0.75rem;
  border-left: 4px solid var(--ui-primary);
  font-style: italic;
  color: #374151;
}

.dark .article-content blockquote {
  background-color: #1f2937;
  color: #e5e7eb;
}

.article-content blockquote p {
  font-size: 1.25rem;
  line-height: 1.75rem;
  margin-bottom: 0;
  color: inherit;
}

.article-content blockquote::before {
  content: '"';
  position: absolute;
  top: -0.5rem;
  left: 1rem;
  font-size: 3.75rem;
  color: var(--ui-primary);
  opacity: 0.3;
  font-family: serif;
}

/* Images */
.article-content img {
  border-radius: 0.75rem;
  margin: 2rem 0;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.article-content figure {
  margin: 2rem 0;
}

.article-content figcaption {
  text-align: center;
  font-size: 0.875rem;
  color: #6b7280;
  margin-top: 0.75rem;
  font-style: italic;
}

.dark .article-content figcaption {
  color: #d1d5db;
}

/* Code blocks */
.article-content pre {
  margin: 1.5rem 0;
  padding: 1rem;
  background-color: #1f2937;
  border-radius: 0.75rem;
  overflow-x: auto;
  font-size: 0.875rem;
}

.dark .article-content pre {
  background-color: #111827;
  border: 1px solid #374151;
}

.article-content code {
  font-size: 0.875rem;
  background-color: #f3f4f6;
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  color: #dc2626;
}

.dark .article-content code {
  background-color: #374151;
  color: #fbbf24;
}

.article-content pre code {
  background-color: transparent;
  padding: 0;
  color: #e5e7eb;
}

/* Horizontal rule */
.article-content hr {
  margin: 2.5rem 0;
  border-color: #e5e7eb;
}

.dark .article-content hr {
  border-color: #374151;
}

/* Tables */
.article-content table {
  width: 100%;
  margin: 1.5rem 0;
  font-size: 0.875rem;
  border-collapse: collapse;
}

.article-content th {
  background-color: #f3f4f6;
  padding: 0.75rem 1rem;
  text-align: left;
  font-weight: 600;
  color: #111827;
  border-bottom: 2px solid #e5e7eb;
}

.dark .article-content th {
  background-color: #1f2937;
  color: #f9fafb;
  border-bottom-color: #374151;
}

.article-content td {
  border-bottom: 1px solid #e5e7eb;
  padding: 0.75rem 1rem;
  color: #374151;
}

.dark .article-content td {
  border-bottom-color: #374151;
  color: #e5e7eb;
}

.article-content tr:hover td {
  background-color: #f9fafb;
}

.dark .article-content tr:hover td {
  background-color: #1f2937;
}
</style>
