<script setup lang="ts">
import type {CommentAuthor, PoliticianComment, PoliticianWithArticles} from '~/types'

definePageMeta({
  layout: 'politician'
})

const route = useRoute()
const api = useApi()
const auth = useAuth()
const { sanitizeComment } = useSanitizedHtml()

const slug = computed(() => route.params.slug as string)

// State
const data = ref<PoliticianWithArticles | null>(null)
const loading = ref(true)
const error = ref('')
const activeTab = ref<'articles' | 'about' | 'discussion'>('articles')
const currentPage = ref(1)
const loadingMore = ref(false)
const hasMoreArticles = ref(true)

// Comments state
const comments = ref<PoliticianComment[]>([])
const commentsLoading = ref(false)
const commentsError = ref('')
const commentsPage = ref(1)
const commentsTotal = ref(0)
const commentsTotalPages = ref(0)
const hasMoreComments = ref(false)
const loadingMoreComments = ref(false)
const newComment = ref('')
const submittingComment = ref(false)
const replyingTo = ref<string | null>(null)
const replyContent = ref('')
const submittingReply = ref(false)
const expandedReplies = ref<Set<string>>(new Set())
const repliesMap = ref<Record<string, PoliticianComment[]>>({})
const loadingReplies = ref<Set<string>>(new Set())

// Mention autocomplete state
const mentionableUsers = ref<CommentAuthor[]>([])
const showMentionDropdown = ref(false)
const mentionSearch = ref('')
const mentionAnchor = ref<'comment' | 'reply'>('comment')
const selectedMentionIndex = ref(0)
const commentTextarea = ref<HTMLTextAreaElement | null>(null)
const replyTextarea = ref<HTMLTextAreaElement | null>(null)

// Filtered users for mention dropdown
const filteredMentionUsers = computed(() => {
  if (!mentionSearch.value) return mentionableUsers.value.slice(0, 5)
  const search = mentionSearch.value.toLowerCase()
  return mentionableUsers.value
    .filter(u => u.name.toLowerCase().includes(search))
    .slice(0, 5)
})

// Load mentionable users
async function loadMentionableUsers() {
  try {
    mentionableUsers.value = await api.getMentionableUsers()
  } catch (e) {
    console.error('Failed to load mentionable users:', e)
  }
}

// Handle input for mention detection
function handleCommentInput(event: Event) {
  const textarea = event.target as HTMLTextAreaElement
  detectMention(textarea, 'comment')
}

function handleReplyInput(event: Event) {
  const textarea = event.target as HTMLTextAreaElement
  detectMention(textarea, 'reply')
}

function detectMention(textarea: HTMLTextAreaElement, anchor: 'comment' | 'reply') {
  const cursorPos = textarea.selectionStart
  const text = textarea.value.substring(0, cursorPos)

  // Find the last @ symbol
  const lastAtIndex = text.lastIndexOf('@')
  if (lastAtIndex === -1) {
    showMentionDropdown.value = false
    return
  }

  // Check if there's a space between @ and cursor (meaning mention is complete)
  const textAfterAt = text.substring(lastAtIndex + 1)
  if (textAfterAt.includes(' ') || textAfterAt.includes('\n')) {
    showMentionDropdown.value = false
    return
  }

  // Show dropdown with search
  mentionSearch.value = textAfterAt
  mentionAnchor.value = anchor
  showMentionDropdown.value = true
  selectedMentionIndex.value = 0
}

// Insert mention
function insertMention(user: CommentAuthor) {
  const textarea = mentionAnchor.value === 'comment' ? commentTextarea.value : replyTextarea.value
  const modelRef = mentionAnchor.value === 'comment' ? newComment : replyContent

  if (!textarea) return

  const text = modelRef.value
  const lastAtIndex = text.lastIndexOf('@')

  if (lastAtIndex !== -1) {
    // Get what's currently typed after @
    const currentMention = text.substring(lastAtIndex + 1).split(/[\s\n]/)[0] || ''

    // If user already typed the exact name, just add a space and close
    if (currentMention && currentMention.toLowerCase() === user.name.toLowerCase()) {
      // Just add space if not already there
      if (!text.endsWith(' ')) {
        modelRef.value = text + ' '
      }
      showMentionDropdown.value = false
      mentionSearch.value = ''
      return
    }

    // Replace @partial with @fullname
    const before = text.substring(0, lastAtIndex)
    const afterMention = text.substring(lastAtIndex + 1 + currentMention.length)
    modelRef.value = `${before}@${user.name} ${afterMention.trimStart()}`

    // Move cursor after the inserted mention
    nextTick(() => {
      const newPos = lastAtIndex + user.name.length + 2 // +2 for @ and space
      textarea.setSelectionRange(newPos, newPos)
      textarea.focus()
    })
  }

  showMentionDropdown.value = false
  mentionSearch.value = ''
}

// Handle mention blur
function handleMentionBlur() {
  globalThis.setTimeout(() => {
    showMentionDropdown.value = false
  }, 200)
}

// Handle keyboard navigation in mention dropdown
function handleMentionKeydown(event: KeyboardEvent) {
  if (!showMentionDropdown.value) return

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    selectedMentionIndex.value = Math.min(selectedMentionIndex.value + 1, filteredMentionUsers.value.length - 1)
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    selectedMentionIndex.value = Math.max(selectedMentionIndex.value - 1, 0)
  } else if (event.key === 'Enter' && filteredMentionUsers.value.length > 0) {
    event.preventDefault()
    const user = filteredMentionUsers.value[selectedMentionIndex.value]
    if (user) insertMention(user)
  } else if (event.key === 'Escape') {
    showMentionDropdown.value = false
  }
}

// Load mentionable users on mount
onMounted(() => {
  loadMentionableUsers()
})

// Load politician data
async function loadPolitician() {
  loading.value = true
  error.value = ''

  try {
    const result = await api.getPoliticianArticles(slug.value, 1, 10)
    data.value = result
    hasMoreArticles.value = (result.articles?.articles?.length || 0) >= 10
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load politician'
  } finally {
    loading.value = false
  }
}

// Load more articles
async function loadMoreArticles() {
  if (loadingMore.value || !hasMoreArticles.value || !data.value) return

  loadingMore.value = true
  currentPage.value++

  try {
    const result = await api.getPoliticianArticles(slug.value, currentPage.value, 10)
    const moreArticles = result.articles?.articles || []
    if (data.value.articles.articles) {
      data.value.articles.articles = [...data.value.articles.articles, ...moreArticles]
    }
    hasMoreArticles.value = moreArticles.length >= 10
  } catch (e) {
    currentPage.value--
    console.error('Failed to load more articles:', e)
  } finally {
    loadingMore.value = false
  }
}

// Load comments
async function loadComments(page = 1, append = false) {
  if (page === 1) {
    commentsLoading.value = true
  } else {
    loadingMoreComments.value = true
  }
  commentsError.value = ''

  try {
    const result = await api.getPoliticianComments(slug.value, auth.getAuthHeaders(), page, 10)
    if (append && page > 1) {
      comments.value = [...comments.value, ...result.comments]
    } else {
      comments.value = result.comments || []
    }
    commentsTotal.value = result.total
    commentsTotalPages.value = result.total_pages
    commentsPage.value = page
    hasMoreComments.value = page < result.total_pages
  } catch (e) {
    commentsError.value = e instanceof Error ? e.message : 'Failed to load comments'
  } finally {
    commentsLoading.value = false
    loadingMoreComments.value = false
  }
}

// Load more comments
async function loadMoreComments() {
  if (loadingMoreComments.value || !hasMoreComments.value) return
  await loadComments(commentsPage.value + 1, true)
}

// Submit new comment
async function submitComment() {
  if (!newComment.value.trim() || submittingComment.value || !auth.isAuthenticated.value) return

  submittingComment.value = true

  try {
    const comment = await api.createPoliticianComment(
      slug.value,
      { content: newComment.value.trim() },
      auth.getAuthHeaders()
    )
    comments.value = [comment, ...comments.value]
    commentsTotal.value++
    newComment.value = ''
  } catch (e) {
    console.error('Failed to submit comment:', e)
  } finally {
    submittingComment.value = false
  }
}

// Toggle replies
async function toggleReplies(commentId: string) {
  if (expandedReplies.value.has(commentId)) {
    expandedReplies.value.delete(commentId)
    expandedReplies.value = new Set(expandedReplies.value)
    return
  }

  // Load replies if not already loaded
  if (!repliesMap.value[commentId]) {
    loadingReplies.value.add(commentId)
    loadingReplies.value = new Set(loadingReplies.value)

    try {
      repliesMap.value[commentId] = await api.getPoliticianCommentReplies(commentId, auth.getAuthHeaders())
    } catch (e) {
      console.error('Failed to load replies:', e)
    } finally {
      loadingReplies.value.delete(commentId)
      loadingReplies.value = new Set(loadingReplies.value)
    }
  }

  expandedReplies.value.add(commentId)
  expandedReplies.value = new Set(expandedReplies.value)
}

// Start replying
function startReply(commentId: string) {
  replyingTo.value = commentId
  replyContent.value = ''
}

// Cancel reply
function cancelReply() {
  replyingTo.value = null
  replyContent.value = ''
}

// Submit reply
async function submitReply(parentId: string) {
  if (!replyContent.value.trim() || submittingReply.value || !auth.isAuthenticated.value) return

  submittingReply.value = true

  try {
    const reply = await api.createPoliticianComment(
      slug.value,
      { content: replyContent.value.trim(), parent_id: parentId },
      auth.getAuthHeaders()
    )

    // Add to replies map
    if (!repliesMap.value[parentId]) {
      repliesMap.value[parentId] = []
    }
    repliesMap.value[parentId].push(reply)

    // Update reply count
    const parentComment = comments.value.find(c => c.id === parentId)
    if (parentComment) {
      parentComment.reply_count = (parentComment.reply_count || 0) + 1
    }

    // Make sure replies are expanded
    expandedReplies.value.add(parentId)
    expandedReplies.value = new Set(expandedReplies.value)

    replyingTo.value = null
    replyContent.value = ''
  } catch (e) {
    console.error('Failed to submit reply:', e)
  } finally {
    submittingReply.value = false
  }
}

// Toggle reaction
async function toggleReaction(commentId: string, reaction: string, hasReacted: boolean) {
  if (!auth.isAuthenticated.value) return

  try {
    if (hasReacted) {
      await api.removePoliticianCommentReaction(commentId, reaction, auth.getAuthHeaders())
    } else {
      await api.addPoliticianCommentReaction(commentId, reaction, auth.getAuthHeaders())
    }
    // Reload comments to get updated reactions
    await loadComments(commentsPage.value, false)
  } catch (e) {
    console.error('Failed to toggle reaction:', e)
  }
}

// Format content with highlighted mentions
function formatContentWithMentions(content: string): string {
  // Replace @username with highlighted span
  return content.replace(/@(\w+(?:\s+\w+)?)/g, '<span class="text-primary font-medium cursor-pointer hover:underline">@$1</span>')
}

// Sanitize formatted content for safe display (XSS protection)
function getSanitizedCommentContent(content: string): string {
  return sanitizeComment(formatContentWithMentions(content))
}

// Format count (1000 → 1k, 1000000 → 1M)
function formatCount(count: number): string {
  if (count >= 1000000) {
    return (count / 1000000).toFixed(1).replace(/\.0$/, '') + 'M'
  }
  if (count >= 1000) {
    return (count / 1000).toFixed(1).replace(/\.0$/, '') + 'k'
  }
  return count.toString()
}

// Get reaction count for a comment
function getReactionCount(comment: PoliticianComment, reactionType: string): number {
  return comment.reactions?.find(r => r.reaction === reactionType)?.count || 0
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

// Watch for slug changes
watch(slug, () => {
  currentPage.value = 1
  hasMoreArticles.value = true
  commentsPage.value = 1
  comments.value = []
  loadPolitician()
}, { immediate: true })

// Watch for tab changes to load comments
watch(activeTab, (newTab) => {
  if (newTab === 'discussion' && comments.value.length === 0 && !commentsLoading.value) {
    loadComments()
  }
})

// SEO
useHead(() => ({
  title: data.value?.politician ? `${data.value.politician.name} - Pulpulitiko` : 'Politician Profile',
  meta: [
    {
      name: 'description',
      content: data.value?.politician?.short_bio || `News and articles about ${data.value?.politician?.name}`
    },
    {
      property: 'og:title',
      content: data.value?.politician?.name || 'Politician Profile'
    },
    {
      property: 'og:description',
      content: data.value?.politician?.short_bio || `News and articles about ${data.value?.politician?.name}`
    },
    {
      property: 'og:image',
      content: data.value?.politician?.photo || ''
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
        title="Politician not found"
        :description="error"
        class="mb-6"
      />

      <!-- Politician content -->
      <template v-else-if="data?.politician">
        <!-- Profile header -->
        <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-800 p-6 mb-6">
          <div class="flex flex-col sm:flex-row items-center sm:items-start gap-6">
            <!-- Photo -->
            <div class="flex-shrink-0">
              <NuxtImg
                v-if="data.politician.photo"
                :src="data.politician.photo"
                :alt="data.politician.name"
                class="w-24 h-24 rounded-full object-cover ring-4 ring-gray-100 dark:ring-gray-800"
              />
              <div v-else class="w-24 h-24 rounded-full bg-primary/10 flex items-center justify-center ring-4 ring-gray-100 dark:ring-gray-800">
                <UIcon name="i-heroicons-user-circle" class="w-12 h-12 text-primary" />
              </div>
            </div>

            <!-- Info -->
            <div class="flex-1 text-center sm:text-left">
              <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-1">
                {{ data.politician.name }}
              </h1>
              <p v-if="data.politician.position" class="text-gray-600 dark:text-gray-400 mb-1">
                {{ data.politician.position }}
              </p>
              <UBadge v-if="data.politician.party" variant="subtle" size="sm">
                {{ data.politician.party }}
              </UBadge>
            </div>

            <!-- Stats -->
            <div class="flex gap-6 sm:gap-8">
              <div class="text-center">
                <div class="text-2xl font-bold text-gray-900 dark:text-white">
                  {{ data.articles.total }}
                </div>
                <div class="text-sm text-gray-500 dark:text-gray-400">
                  Articles
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
                activeTab === 'articles'
                  ? 'text-primary border-b-2 border-primary bg-primary/5'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              ]"
              @click="activeTab = 'articles'"
            >
              Related Articles ({{ data.articles.total }})
            </button>
            <button
              :class="[
                'flex-1 px-4 py-3 text-sm font-medium transition-colors',
                activeTab === 'discussion'
                  ? 'text-primary border-b-2 border-primary bg-primary/5'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              ]"
              @click="activeTab = 'discussion'"
            >
              Discussion
              <span v-if="commentsTotal > 0" class="ml-1">({{ commentsTotal }})</span>
            </button>
            <button
              v-if="data.politician.short_bio"
              :class="[
                'flex-1 px-4 py-3 text-sm font-medium transition-colors',
                activeTab === 'about'
                  ? 'text-primary border-b-2 border-primary bg-primary/5'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              ]"
              @click="activeTab = 'about'"
            >
              About
            </button>
          </div>

          <!-- Articles tab -->
          <div v-if="activeTab === 'articles'" class="p-4">
            <template v-if="data.articles.articles?.length">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <NuxtLink
                  v-for="article in data.articles.articles"
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
                    Load more articles
                  </template>
                </button>
              </div>
            </template>

            <!-- Empty state -->
            <div v-else class="py-8 text-center">
              <UIcon name="i-heroicons-document-text" class="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <p class="text-gray-500 dark:text-gray-400">No articles yet</p>
            </div>
          </div>

          <!-- Discussion tab -->
          <div v-if="activeTab === 'discussion'" class="p-4">
            <!-- Comment form -->
            <div class="mb-6">
              <template v-if="auth.isAuthenticated.value">
                <div class="flex gap-3">
                  <div class="flex-shrink-0">
                    <div class="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                      <UIcon name="i-heroicons-user" class="w-5 h-5 text-primary" />
                    </div>
                  </div>
                  <div class="flex-1 relative">
                    <textarea
                      ref="commentTextarea"
                      v-model="newComment"
                      rows="3"
                      class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-4 py-2 text-sm text-gray-900 dark:text-white placeholder-gray-500 focus:border-primary focus:ring-1 focus:ring-primary resize-none"
                      placeholder="Share your thoughts... Use @ to mention users"
                      :disabled="submittingComment"
                      @input="handleCommentInput"
                      @keydown="handleMentionKeydown"
                      @blur="handleMentionBlur"
                    />
                    <!-- Mention dropdown for comment -->
                    <div
                      v-if="showMentionDropdown && mentionAnchor === 'comment' && filteredMentionUsers.length > 0"
                      class="absolute z-50 mt-1 w-64 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 overflow-hidden"
                    >
                      <div class="py-1">
                        <button
                          v-for="(user, index) in filteredMentionUsers"
                          :key="user.id"
                          class="w-full px-3 py-2 flex items-center gap-2 text-left text-sm transition-colors"
                          :class="index === selectedMentionIndex
                            ? 'bg-primary/10 text-primary'
                            : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'"
                          @mousedown.prevent="insertMention(user)"
                        >
                          <img
                            v-if="user.avatar"
                            :src="user.avatar"
                            :alt="user.name"
                            class="w-6 h-6 rounded-full object-cover"
                          >
                          <span v-else class="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center">
                            <UIcon name="i-heroicons-user" class="w-3 h-3 text-primary"/>
                          </span>
                          <span class="font-medium">{{ user.name }}</span>
                        </button>
                      </div>
                    </div>
                    <div class="mt-2 flex items-center justify-between">
                      <span class="text-xs text-gray-500 dark:text-gray-400">
                        Tip: Type @ to mention someone
                      </span>
                      <UButton
                        color="primary"
                        size="sm"
                        :loading="submittingComment"
                        :disabled="!newComment.trim() || submittingComment"
                        @click="submitComment"
                      >
                        Post Comment
                      </UButton>
                    </div>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4 text-center">
                  <p class="text-gray-600 dark:text-gray-400 text-sm mb-2">
                    Sign in to join the discussion
                  </p>
                  <NuxtLink to="/login" class="text-primary hover:text-primary/80 font-medium text-sm">
                    Log in or create an account
                  </NuxtLink>
                </div>
              </template>
            </div>

            <!-- Loading state -->
            <div v-if="commentsLoading" class="py-8 text-center">
              <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 text-gray-400 mx-auto animate-spin mb-2" />
              <p class="text-gray-500 dark:text-gray-400 text-sm">Loading comments...</p>
            </div>

            <!-- Error state -->
            <UAlert
              v-else-if="commentsError"
              color="error"
              icon="i-heroicons-exclamation-triangle"
              :description="commentsError"
              class="mb-4"
            />

            <!-- Comments list -->
            <template v-else-if="comments.length > 0">
              <div class="space-y-4">
                <div
                  v-for="comment in comments"
                  :key="comment.id"
                  class="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4"
                >
                  <!-- Comment header -->
                  <div class="flex items-start gap-3">
                    <div class="flex-shrink-0">
                      <img
                        v-if="comment.author?.avatar"
                        :src="comment.author.avatar"
                        :alt="comment.author.name"
                        class="w-8 h-8 rounded-full object-cover"
                      >
                      <div v-else class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
                        <UIcon name="i-heroicons-user" class="w-4 h-4 text-primary" />
                      </div>
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-2">
                        <span class="font-medium text-gray-900 dark:text-white text-sm">
                          {{ comment.author?.name || 'Anonymous' }}
                        </span>
                        <span class="text-xs text-gray-500 dark:text-gray-400">
                          {{ formatRelativeTime(comment.created_at) }}
                        </span>
                      </div>
                      <p
                        class="mt-1 text-gray-700 dark:text-gray-300 text-sm whitespace-pre-wrap"
                        v-html="getSanitizedCommentContent(comment.content)"
                      />

                      <!-- Comment actions -->
                      <div class="mt-2 flex items-center gap-2">
                        <!-- Like -->
                        <button
                          class="text-xs text-gray-500 dark:text-gray-400 hover:text-green-500 flex items-center gap-1 px-2 py-1 rounded-full transition-colors"
                          :class="{
                            'text-green-500 bg-green-50 dark:bg-green-950/30': comment.reactions?.find(r => r.reaction === 'thumbsup')?.has_reacted,
                            'cursor-not-allowed opacity-60': !auth.isAuthenticated.value
                          }"
                          :title="!auth.isAuthenticated.value ? 'Sign in to react' : ''"
                          @click="auth.isAuthenticated.value ? toggleReaction(comment.id, 'thumbsup', comment.reactions?.find(r => r.reaction === 'thumbsup')?.has_reacted || false) : $router.push('/login')"
                        >
                          <UIcon name="i-heroicons-hand-thumb-up" class="w-4 h-4" />
                          <span v-if="getReactionCount(comment, 'thumbsup') > 0">{{ formatCount(getReactionCount(comment, 'thumbsup')) }}</span>
                        </button>

                        <!-- Dislike -->
                        <button
                          class="text-xs text-gray-500 dark:text-gray-400 hover:text-red-500 flex items-center gap-1 px-2 py-1 rounded-full transition-colors"
                          :class="{
                            'text-red-500 bg-red-50 dark:bg-red-950/30': comment.reactions?.find(r => r.reaction === 'thumbsdown')?.has_reacted,
                            'cursor-not-allowed opacity-60': !auth.isAuthenticated.value
                          }"
                          :title="!auth.isAuthenticated.value ? 'Sign in to react' : ''"
                          @click="auth.isAuthenticated.value ? toggleReaction(comment.id, 'thumbsdown', comment.reactions?.find(r => r.reaction === 'thumbsdown')?.has_reacted || false) : $router.push('/login')"
                        >
                          <UIcon name="i-heroicons-hand-thumb-down" class="w-4 h-4" />
                          <span v-if="getReactionCount(comment, 'thumbsdown') > 0">{{ formatCount(getReactionCount(comment, 'thumbsdown')) }}</span>
                        </button>

                        <!-- Reply button -->
                        <button
                          class="text-xs text-gray-500 dark:text-gray-400 hover:text-primary flex items-center gap-1"
                          :class="{ 'cursor-not-allowed opacity-60': !auth.isAuthenticated.value }"
                          :title="!auth.isAuthenticated.value ? 'Sign in to reply' : ''"
                          @click="auth.isAuthenticated.value ? startReply(comment.id) : $router.push('/login')"
                        >
                          <UIcon name="i-heroicons-chat-bubble-left" class="w-4 h-4" />
                          Reply
                        </button>

                        <!-- Show replies button -->
                        <button
                          v-if="comment.reply_count && comment.reply_count > 0"
                          class="text-xs text-primary hover:text-primary/80 flex items-center gap-1"
                          @click="toggleReplies(comment.id)"
                        >
                          <UIcon
                            :name="expandedReplies.has(comment.id) ? 'i-heroicons-chevron-up' : 'i-heroicons-chevron-down'"
                            class="w-4 h-4"
                          />
                          {{ expandedReplies.has(comment.id) ? 'Hide' : 'Show' }} {{ comment.reply_count }} {{ comment.reply_count === 1 ? 'reply' : 'replies' }}
                        </button>
                      </div>

                      <!-- Reply form -->
                      <div v-if="replyingTo === comment.id" class="mt-3 pl-4 border-l-2 border-gray-200 dark:border-gray-700 relative">
                        <textarea
                          ref="replyTextarea"
                          v-model="replyContent"
                          rows="2"
                          class="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white placeholder-gray-500 focus:border-primary focus:ring-1 focus:ring-primary resize-none"
                          placeholder="Write a reply... Use @ to mention"
                          :disabled="submittingReply"
                          @input="handleReplyInput"
                          @keydown="handleMentionKeydown"
                          @blur="handleMentionBlur"
                        />
                        <!-- Mention dropdown for reply -->
                        <div
                          v-if="showMentionDropdown && mentionAnchor === 'reply' && filteredMentionUsers.length > 0"
                          class="absolute z-50 mt-1 w-64 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 overflow-hidden"
                        >
                          <div class="py-1">
                            <button
                              v-for="(user, index) in filteredMentionUsers"
                              :key="user.id"
                              class="w-full px-3 py-2 flex items-center gap-2 text-left text-sm transition-colors"
                              :class="index === selectedMentionIndex
                                ? 'bg-primary/10 text-primary'
                                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'"
                              @mousedown.prevent="insertMention(user)"
                            >
                              <img
                                v-if="user.avatar"
                                :src="user.avatar"
                                :alt="user.name"
                                class="w-6 h-6 rounded-full object-cover"
                              >
                              <span v-else class="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center">
                                <UIcon name="i-heroicons-user" class="w-3 h-3 text-primary" />
                              </span>
                              <span class="font-medium">{{ user.name }}</span>
                            </button>
                          </div>
                        </div>
                        <div class="mt-2 flex gap-2 justify-end">
                          <UButton
                            color="neutral"
                            variant="ghost"
                            size="xs"
                            @click="cancelReply"
                          >
                            Cancel
                          </UButton>
                          <UButton
                            color="primary"
                            size="xs"
                            :loading="submittingReply"
                            :disabled="!replyContent.trim() || submittingReply"
                            @click="submitReply(comment.id)"
                          >
                            Reply
                          </UButton>
                        </div>
                      </div>

                      <!-- Replies loading -->
                      <div v-if="loadingReplies.has(comment.id)" class="mt-3 pl-4 border-l-2 border-gray-200 dark:border-gray-700">
                        <div class="flex items-center gap-2 text-sm text-gray-500">
                          <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin" />
                          Loading replies...
                        </div>
                      </div>

                      <!-- Replies list -->
                      <div v-if="expandedReplies.has(comment.id) && repliesMap[comment.id]?.length" class="mt-3 pl-4 border-l-2 border-gray-200 dark:border-gray-700 space-y-3">
                        <div
                          v-for="reply in repliesMap[comment.id]"
                          :key="reply.id"
                          class="bg-white dark:bg-gray-900 rounded-lg p-3"
                        >
                          <div class="flex items-start gap-2">
                            <div class="shrink-0">
                              <img
                                v-if="reply.author?.avatar"
                                :src="reply.author.avatar"
                                :alt="reply.author.name"
                                class="w-6 h-6 rounded-full object-cover"
                              >
                              <div v-else class="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center">
                                <UIcon name="i-heroicons-user" class="w-3 h-3 text-primary" />
                              </div>
                            </div>
                            <div class="flex-1 min-w-0">
                              <div class="flex items-center gap-2">
                                <span class="font-medium text-gray-900 dark:text-white text-xs">
                                  {{ reply.author?.name || 'Anonymous' }}
                                </span>
                                <span class="text-xs text-gray-500 dark:text-gray-400">
                                  {{ formatRelativeTime(reply.created_at) }}
                                </span>
                              </div>
                              <p
                                class="mt-1 text-gray-700 dark:text-gray-300 text-sm whitespace-pre-wrap"
                                v-html="getSanitizedCommentContent(reply.content)"
                              />
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Load more comments -->
              <div v-if="hasMoreComments" class="mt-6 text-center">
                <button
                  class="text-primary hover:text-primary/80 font-medium disabled:opacity-50"
                  :disabled="loadingMoreComments"
                  @click="loadMoreComments"
                >
                  <template v-if="loadingMoreComments">
                    <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin inline mr-2" />
                    Loading...
                  </template>
                  <template v-else>
                    Show more comments ({{ commentsTotal - comments.length }} remaining)
                  </template>
                </button>
              </div>
            </template>

            <!-- Empty state -->
            <div v-else class="py-8 text-center">
              <UIcon name="i-heroicons-chat-bubble-left-right" class="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <p class="text-gray-500 dark:text-gray-400">No comments yet. Be the first to share your thoughts!</p>
            </div>
          </div>

          <!-- About tab -->
          <div v-if="activeTab === 'about' && data.politician.short_bio" class="p-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">
              About {{ data.politician.name }}
            </h3>
            <p class="text-gray-600 dark:text-gray-400 whitespace-pre-wrap">
              {{ data.politician.short_bio }}
            </p>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
