<script setup lang="ts">
import type { Comment, UserProfile, ArticleListItem, AuthorWithArticles } from '~/types'

const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)

// State
const profile = ref<UserProfile | null>(null)
const comments = ref<Comment[]>([])
const replies = ref<Comment[]>([])
const articles = ref<ArticleListItem[]>([])
const totalArticles = ref(0)
const loading = ref(true)
const error = ref('')
const activeTab = ref<'posts' | 'comments' | 'replies'>('posts')
const commentsPage = ref(1)
const repliesPage = ref(1)
const articlesPage = ref(1)
const loadingMore = ref(false)
const hasMoreComments = ref(true)
const hasMoreReplies = ref(true)
const hasMoreArticles = ref(true)

// Load profile and initial data
async function loadProfile() {
  loading.value = true
  error.value = ''

  try {
    // Load profile and comments first
    const [profileData, commentsData, repliesData] = await Promise.all([
      api.getUserProfile(slug.value),
      api.getUserComments(slug.value, 1, 10),
      api.getUserReplies(slug.value, 1, 10)
    ])

    profile.value = profileData
    comments.value = commentsData || []
    replies.value = repliesData || []

    hasMoreComments.value = (commentsData?.length || 0) >= 10
    hasMoreReplies.value = (repliesData?.length || 0) >= 10

    // Try to load articles (user might also be an author)
    try {
      const authorData = await api.getAuthorArticles(slug.value, 1, 10) as AuthorWithArticles
      articles.value = authorData.articles?.articles || []
      totalArticles.value = authorData.articles?.total || 0
      hasMoreArticles.value = (authorData.articles?.articles?.length || 0) >= 10
    } catch {
      // User is not an author or has no articles - that's fine
      articles.value = []
      totalArticles.value = 0
      hasMoreArticles.value = false
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load user profile'
  } finally {
    loading.value = false
  }
}

// Load more comments
async function loadMoreComments() {
  if (loadingMore.value || !hasMoreComments.value) return

  loadingMore.value = true
  commentsPage.value++

  try {
    const moreComments = await api.getUserComments(slug.value, commentsPage.value, 10)
    comments.value = [...comments.value, ...(moreComments || [])]
    hasMoreComments.value = (moreComments?.length || 0) >= 10
  } catch (e) {
    commentsPage.value--
    console.error('Failed to load more comments:', e)
  } finally {
    loadingMore.value = false
  }
}

// Load more replies
async function loadMoreReplies() {
  if (loadingMore.value || !hasMoreReplies.value) return

  loadingMore.value = true
  repliesPage.value++

  try {
    const moreReplies = await api.getUserReplies(slug.value, repliesPage.value, 10)
    replies.value = [...replies.value, ...(moreReplies || [])]
    hasMoreReplies.value = (moreReplies?.length || 0) >= 10
  } catch (e) {
    repliesPage.value--
    console.error('Failed to load more replies:', e)
  } finally {
    loadingMore.value = false
  }
}

// Load more articles
async function loadMoreArticles() {
  if (loadingMore.value || !hasMoreArticles.value) return

  loadingMore.value = true
  articlesPage.value++

  try {
    const authorData = await api.getAuthorArticles(slug.value, articlesPage.value, 10) as AuthorWithArticles
    const moreArticles = authorData.articles?.articles || []
    articles.value = [...articles.value, ...moreArticles]
    hasMoreArticles.value = moreArticles.length >= 10
  } catch (e) {
    articlesPage.value--
    console.error('Failed to load more articles:', e)
  } finally {
    loadingMore.value = false
  }
}

// Format date
function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-PH', {
    month: 'long',
    day: 'numeric',
    year: 'numeric'
  })
}

// Format relative time
function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString('en-PH', {
    month: 'short',
    day: 'numeric',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
  })
}

// Format comment content (simplified, no HTML)
function formatContent(content: string): string {
  // Strip markdown and truncate
  return content
    .replace(/\*\*(.*?)\*\*/g, '$1')
    .replace(/\*(.*?)\*/g, '$1')
    .replace(/~~(.*?)~~/g, '$1')
    .replace(/^>\s?/gm, '')
    .replace(/@(\w+)/g, '@$1')
    .substring(0, 200) + (content.length > 200 ? '...' : '')
}

// Watch for slug changes
watch(slug, () => {
  loadProfile()
}, { immediate: true })

// SEO
useHead(() => ({
  title: profile.value ? `${profile.value.name} - User Profile` : 'User Profile',
  meta: [
    {
      name: 'description',
      content: `View ${profile.value?.name || 'user'}'s comments and activity`
    }
  ]
}))
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Loading state -->
      <div v-if="loading" class="animate-pulse">
        <div class="flex items-center gap-6 mb-8">
          <div class="w-24 h-24 rounded-full bg-gray-200 dark:bg-gray-800" />
          <div class="flex-1">
            <div class="h-8 bg-gray-200 dark:bg-gray-800 rounded w-48 mb-2" />
            <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-32" />
          </div>
        </div>
      </div>

      <!-- Error state -->
      <UAlert
        v-else-if="error"
        color="error"
        icon="i-heroicons-exclamation-triangle"
        :description="error"
        class="mb-6"
      />

      <!-- Profile content -->
      <template v-else-if="profile">
        <!-- Profile header -->
        <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-800 p-6 mb-6">
          <div class="flex flex-col sm:flex-row items-center sm:items-start gap-6">
            <!-- Avatar -->
            <div class="flex-shrink-0">
              <NuxtImg
                v-if="profile.avatar"
                :src="profile.avatar"
                :alt="profile.name"
                class="w-24 h-24 rounded-full object-cover ring-4 ring-gray-100 dark:ring-gray-800"
              />
              <div v-else class="w-24 h-24 rounded-full bg-primary/10 flex items-center justify-center ring-4 ring-gray-100 dark:ring-gray-800">
                <UIcon name="i-heroicons-user" class="w-12 h-12 text-primary" />
              </div>
            </div>

            <!-- User info -->
            <div class="flex-1 text-center sm:text-left">
              <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-1">
                {{ profile.name }}
              </h1>
              <p class="text-gray-500 dark:text-gray-400 text-sm">
                Member since {{ formatDate(profile.created_at) }}
              </p>
            </div>

            <!-- Stats -->
            <div class="flex gap-6 sm:gap-8">
              <div class="text-center">
                <div class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ profile.comment_count }}
                </div>
                <div class="text-sm text-gray-500 dark:text-gray-400">
                  Comments
                </div>
              </div>
              <div class="text-center">
                <div class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ profile.reply_count }}
                </div>
                <div class="text-sm text-gray-500 dark:text-gray-400">
                  Replies
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Tabs -->
        <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-800 overflow-hidden">
          <!-- Tab buttons -->
          <div class="flex border-b border-gray-200 dark:border-gray-800">
            <button
              :class="[
                'flex-1 px-4 py-3 text-sm font-medium transition-colors',
                activeTab === 'posts'
                  ? 'text-primary border-b-2 border-primary bg-primary/5'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              ]"
              @click="activeTab = 'posts'"
            >
              Posts ({{ totalArticles }})
            </button>
            <button
              :class="[
                'flex-1 px-4 py-3 text-sm font-medium transition-colors',
                activeTab === 'comments'
                  ? 'text-primary border-b-2 border-primary bg-primary/5'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              ]"
              @click="activeTab = 'comments'"
            >
              Comments ({{ profile.comment_count }})
            </button>
            <button
              :class="[
                'flex-1 px-4 py-3 text-sm font-medium transition-colors',
                activeTab === 'replies'
                  ? 'text-primary border-b-2 border-primary bg-primary/5'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              ]"
              @click="activeTab = 'replies'"
            >
              Replies ({{ profile.reply_count }})
            </button>
          </div>

          <!-- Posts tab -->
          <div v-if="activeTab === 'posts'" class="p-4">
            <template v-if="articles.length > 0">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <NuxtLink
                  v-for="article in articles"
                  :key="article.id"
                  :to="`/article/${article.slug}`"
                  class="group block bg-gray-50 dark:bg-gray-800/50 rounded-xl overflow-hidden hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                >
                  <div v-if="article.featured_image" class="aspect-video overflow-hidden">
                    <img
                      :src="article.featured_image"
                      :alt="article.title"
                      class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                    >
                  </div>
                  <div class="p-4">
                    <h3 class="font-semibold text-gray-900 dark:text-white group-hover:text-primary transition-colors line-clamp-2">
                      {{ article.title }}
                    </h3>
                    <p v-if="article.summary" class="mt-1 text-sm text-gray-600 dark:text-gray-400 line-clamp-2">
                      {{ article.summary }}
                    </p>
                    <div class="mt-2 flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
                      <span v-if="article.category_name" class="bg-primary/10 text-primary px-2 py-0.5 rounded">
                        {{ article.category_name }}
                      </span>
                      <span>{{ formatRelativeTime(article.published_at || article.created_at) }}</span>
                      <span class="flex items-center gap-1">
                        <UIcon name="i-heroicons-eye" class="w-3 h-3" />
                        {{ article.view_count }}
                      </span>
                    </div>
                  </div>
                </NuxtLink>
              </div>

              <!-- Load more -->
              <div v-if="hasMoreArticles" class="mt-4 text-center">
                <button
                  class="text-primary hover:text-primary/80 font-medium disabled:opacity-50"
                  :disabled="loadingMore"
                  @click="loadMoreArticles"
                >
                  <template v-if="loadingMore">
                    <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin inline mr-2" />
                    Loading...
                  </template>
                  <template v-else>
                    Load more posts
                  </template>
                </button>
              </div>
            </template>

            <!-- Empty state -->
            <div v-else class="py-8 text-center">
              <UIcon name="i-heroicons-document-text" class="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <p class="text-gray-500 dark:text-gray-400">No posts yet</p>
            </div>
          </div>

          <!-- Comments tab -->
          <div v-if="activeTab === 'comments'" class="divide-y divide-gray-100 dark:divide-gray-800">
            <template v-if="comments.length > 0">
              <div
                v-for="comment in comments"
                :key="comment.id"
                class="p-4 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              >
                <div class="flex items-start gap-3">
                  <div class="flex-1 min-w-0">
                    <!-- Article link -->
                    <NuxtLink
                      v-if="comment.article_slug"
                      :to="`/article/${comment.article_slug}`"
                      class="text-sm text-primary hover:underline font-medium"
                    >
                      View article
                    </NuxtLink>
                    <!-- Comment content -->
                    <p class="text-gray-700 dark:text-gray-300 mt-1">
                      {{ formatContent(comment.content) }}
                    </p>
                    <!-- Meta -->
                    <div class="flex items-center gap-4 mt-2 text-sm text-gray-500 dark:text-gray-400">
                      <span>{{ formatRelativeTime(comment.created_at) }}</span>
                      <span v-if="comment.reactions?.length" class="flex items-center gap-1">
                        <UIcon name="i-heroicons-hand-thumb-up" class="w-4 h-4" />
                        {{ comment.reactions.find(r => r.reaction === 'thumbsup')?.count || 0 }}
                      </span>
                      <span v-if="comment.reply_count" class="flex items-center gap-1">
                        <UIcon name="i-heroicons-chat-bubble-left" class="w-4 h-4" />
                        {{ comment.reply_count }} {{ comment.reply_count === 1 ? 'reply' : 'replies' }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Load more -->
              <div v-if="hasMoreComments" class="p-4 text-center">
                <button
                  class="text-primary hover:text-primary/80 font-medium disabled:opacity-50"
                  :disabled="loadingMore"
                  @click="loadMoreComments"
                >
                  <template v-if="loadingMore">
                    <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin inline mr-2" />
                    Loading...
                  </template>
                  <template v-else>
                    Load more comments
                  </template>
                </button>
              </div>
            </template>

            <!-- Empty state -->
            <div v-else class="p-8 text-center">
              <UIcon name="i-heroicons-chat-bubble-left-right" class="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <p class="text-gray-500 dark:text-gray-400">No comments yet</p>
            </div>
          </div>

          <!-- Replies tab -->
          <div v-if="activeTab === 'replies'" class="divide-y divide-gray-100 dark:divide-gray-800">
            <template v-if="replies.length > 0">
              <div
                v-for="reply in replies"
                :key="reply.id"
                class="p-4 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              >
                <div class="flex items-start gap-3">
                  <div class="flex-1 min-w-0">
                    <!-- Replying to indicator -->
                    <div class="text-sm text-gray-500 dark:text-gray-400 mb-1">
                      <span>Replied to a comment</span>
                      <NuxtLink
                        v-if="reply.article_slug"
                        :to="`/article/${reply.article_slug}`"
                        class="text-primary hover:underline ml-1"
                      >
                        View article
                      </NuxtLink>
                    </div>
                    <!-- Reply content -->
                    <p class="text-gray-700 dark:text-gray-300">
                      {{ formatContent(reply.content) }}
                    </p>
                    <!-- Meta -->
                    <div class="flex items-center gap-4 mt-2 text-sm text-gray-500 dark:text-gray-400">
                      <span>{{ formatRelativeTime(reply.created_at) }}</span>
                      <span v-if="reply.reactions?.length" class="flex items-center gap-1">
                        <UIcon name="i-heroicons-hand-thumb-up" class="w-4 h-4" />
                        {{ reply.reactions.find(r => r.reaction === 'thumbsup')?.count || 0 }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Load more -->
              <div v-if="hasMoreReplies" class="p-4 text-center">
                <button
                  class="text-primary hover:text-primary/80 font-medium disabled:opacity-50"
                  :disabled="loadingMore"
                  @click="loadMoreReplies"
                >
                  <template v-if="loadingMore">
                    <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin inline mr-2" />
                    Loading...
                  </template>
                  <template v-else>
                    Load more replies
                  </template>
                </button>
              </div>
            </template>

            <!-- Empty state -->
            <div v-else class="p-8 text-center">
              <UIcon name="i-heroicons-chat-bubble-left" class="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <p class="text-gray-500 dark:text-gray-400">No replies yet</p>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
