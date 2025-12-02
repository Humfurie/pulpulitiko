<script setup lang="ts">
import type { Comment } from '~/types'

const props = defineProps<{
  articleSlug: string
}>()

const api = useApi()
const auth = useAuth()

const comments = ref<Comment[]>([])
const commentCount = ref(0)
const loading = ref(true)
const loadingMore = ref(false)
const error = ref('')
const sortBy = ref<'recent' | 'liked' | 'oldest'>('recent')
const page = ref(1)
const pageSize = 10

// Reply state
const replyingTo = ref<{ id: string; authorName: string } | null>(null)

// Track which comments have expanded replies
const expandedCommentIds = ref<Set<string>>(new Set())

// Edit state
const editingComment = ref<Comment | null>(null)

// Delete confirmation
const showDeleteModal = ref(false)
const deletingCommentId = ref<string | null>(null)
const isDeleting = ref(false)

// Current user ID for ownership checks
const currentUserId = computed(() => {
  // This would need to be fetched from the auth system
  return auth.user.value?.id
})

async function loadComments(reset = true) {
  if (reset) {
    loading.value = true
    page.value = 1
    comments.value = []
  }
  error.value = ''

  try {
    // Always get auth headers - getAuthHeaders() returns empty object if no token
    const authHeaders = auth.getAuthHeaders()
    const hasAuth = Object.keys(authHeaders).length > 0
    const [commentsData, countData] = await Promise.all([
      api.getArticleComments(props.articleSlug, hasAuth ? authHeaders : undefined, 1, pageSize, sortBy.value),
      api.getCommentCount(props.articleSlug)
    ])
    comments.value = commentsData || []
    commentCount.value = countData.count
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load comments'
  } finally {
    loading.value = false
  }
}

// Reload when sort changes
watch(sortBy, () => {
  loadComments()
})

async function loadMoreComments() {
  loadingMore.value = true
  page.value++

  try {
    // Always get auth headers - getAuthHeaders() returns empty object if no token
    const authHeaders = auth.getAuthHeaders()
    const hasAuth = Object.keys(authHeaders).length > 0
    const moreComments = await api.getArticleComments(props.articleSlug, hasAuth ? authHeaders : undefined, page.value, pageSize, sortBy.value)
    comments.value = [...comments.value, ...(moreComments || [])]
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load more comments'
    page.value-- // Revert page on error
  } finally {
    loadingMore.value = false
  }
}

function handleReply(commentId: string, authorName: string) {
  replyingTo.value = { id: commentId, authorName }
  editingComment.value = null
  // Mark this comment as expanded since we're replying to it
  expandedCommentIds.value.add(commentId)
  // Scroll to the reply form
  nextTick(() => {
    const form = document.querySelector(`[data-reply-to="${commentId}"]`)
    form?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}

function handleExpandToggle(commentId: string, isExpanded: boolean) {
  if (isExpanded) {
    expandedCommentIds.value.add(commentId)
  } else {
    expandedCommentIds.value.delete(commentId)
  }
}

function handleEdit(comment: Comment) {
  editingComment.value = comment
  replyingTo.value = null
}

function handleDeleteClick(commentId: string) {
  deletingCommentId.value = commentId
  showDeleteModal.value = true
}

async function confirmDelete() {
  if (!deletingCommentId.value) return

  isDeleting.value = true
  try {
    await api.deleteComment(deletingCommentId.value, auth.getAuthHeaders())
    // Refresh comments
    await loadComments()
    showDeleteModal.value = false
    deletingCommentId.value = null
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to delete comment'
  } finally {
    isDeleting.value = false
  }
}

function updateCommentReaction(commentList: Comment[], commentId: string, reaction: string, hasReacted: boolean): { updated: boolean; removedOpposite: string | null } {
  for (const comment of commentList) {
    if (comment.id === commentId) {
      if (!comment.reactions) comment.reactions = []
      const existingReaction = comment.reactions.find(r => r.reaction === reaction)
      let removedOpposite: string | null = null

      if (hasReacted) {
        // Remove reaction
        if (existingReaction) {
          existingReaction.count = Math.max(0, existingReaction.count - 1)
          existingReaction.has_reacted = false
        }
      } else {
        // Add reaction - but first remove opposite reaction (like/dislike are mutually exclusive)
        const oppositeReaction = reaction === 'thumbsup' ? 'thumbsdown' : reaction === 'thumbsdown' ? 'thumbsup' : null
        if (oppositeReaction) {
          const opposite = comment.reactions.find(r => r.reaction === oppositeReaction)
          if (opposite?.has_reacted) {
            opposite.count = Math.max(0, opposite.count - 1)
            opposite.has_reacted = false
            removedOpposite = oppositeReaction
          }
        }

        // Now add the new reaction
        if (existingReaction) {
          existingReaction.count++
          existingReaction.has_reacted = true
        } else {
          comment.reactions.push({ reaction, count: 1, has_reacted: true })
        }
      }
      return { updated: true, removedOpposite }
    }
  }
  return { updated: false, removedOpposite: null }
}

async function handleReactionToggle(commentId: string, reaction: string, hasReacted: boolean) {
  if (!auth.isAuthenticated.value) return

  // Optimistically update UI
  const { removedOpposite } = updateCommentReaction(comments.value, commentId, reaction, hasReacted)

  try {
    if (hasReacted) {
      await api.removeReaction(commentId, reaction, auth.getAuthHeaders())
    } else {
      // If we removed an opposite reaction, remove it on server too
      if (removedOpposite) {
        await api.removeReaction(commentId, removedOpposite, auth.getAuthHeaders())
      }
      await api.addReaction(commentId, reaction, auth.getAuthHeaders())
    }
  } catch (e) {
    // Revert on error
    updateCommentReaction(comments.value, commentId, reaction, !hasReacted)
    // Also revert the opposite if we removed it
    if (removedOpposite) {
      updateCommentReaction(comments.value, commentId, removedOpposite, false)
    }
    console.error('Failed to toggle reaction:', e)
  }
}

function handleCommentSubmitted(comment: Comment) {
  // If it was a reply, close the reply form
  if (replyingTo.value) {
    replyingTo.value = null
  }
  // If it was an edit, close the edit form
  if (editingComment.value) {
    editingComment.value = null
  }
  // Refresh comments
  loadComments()
}

function handleFormCancelled() {
  replyingTo.value = null
  editingComment.value = null
}

// Initial load and reload when auth state changes
// Use onMounted to ensure we're on client side where cookie is available
onMounted(() => {
  // immediate: true triggers initial load, then watch triggers on token changes
  watch(() => auth.token.value, () => {
    loadComments()
  }, { immediate: true })
})
</script>

<template>
  <div class="mt-16">
    <!-- Prominent Comments Header -->
    <div class="relative mb-8">
      <!-- Decorative background -->
      <div class="absolute inset-0 bg-gradient-to-r from-primary/5 via-primary/10 to-primary/5 rounded-2xl" />

      <div class="relative px-6 py-8 text-center">
        <!-- Icon -->
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 mb-4">
          <UIcon name="i-heroicons-chat-bubble-left-right" class="w-8 h-8 text-primary" />
        </div>

        <!-- Title -->
        <h2 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Join the Conversation
        </h2>

        <!-- Subtitle with count -->
        <p class="text-gray-600 dark:text-gray-400">
          <template v-if="commentCount > 0">
            <span class="font-semibold text-primary">{{ commentCount }}</span>
            {{ commentCount === 1 ? 'comment' : 'comments' }} so far — share your thoughts below
          </template>
          <template v-else>
            Be the first to share your thoughts on this article
          </template>
        </p>
      </div>
    </div>

    <!-- Comment Form Card -->
    <div class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden mb-8">
      <!-- Main comment form (only show when not editing) -->
      <div v-if="!editingComment">
        <CommentForm
          :article-slug="articleSlug"
          @submitted="handleCommentSubmitted"
        />
      </div>
    </div>

    <!-- Comments List Header -->
    <div v-if="commentCount > 0" class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          Comments
        </h3>
        <span class="inline-flex items-center justify-center min-w-[28px] h-7 px-2 text-sm font-medium text-white bg-primary rounded-full">
          {{ commentCount }}
        </span>
      </div>
      <UDropdownMenu
        :items="[
          [{ label: 'Most recent', icon: 'i-heroicons-arrow-down', click: () => sortBy = 'recent' }],
          [{ label: 'Most liked', icon: 'i-heroicons-heart', click: () => sortBy = 'liked' }],
          [{ label: 'Oldest first', icon: 'i-heroicons-arrow-up', click: () => sortBy = 'oldest' }]
        ]"
      >
        <UButton color="neutral" variant="ghost" trailing-icon="i-heroicons-chevron-down" size="sm">
          {{ sortBy === 'recent' ? 'Most recent' : sortBy === 'liked' ? 'Most liked' : 'Oldest first' }}
        </UButton>
      </UDropdownMenu>
    </div>

    <!-- Edit form -->
    <div v-if="editingComment" class="mb-8">
      <CommentForm
        :article-slug="articleSlug"
        :editing-comment="editingComment"
        @submitted="handleCommentSubmitted"
        @cancelled="handleFormCancelled"
      />
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="space-y-6">
      <div v-for="i in 3" :key="i" class="flex gap-3 animate-pulse">
        <div class="w-10 h-10 rounded-full bg-gray-200 dark:bg-gray-700" />
        <div class="flex-1 space-y-2">
          <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4" />
          <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4" />
          <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2" />
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

    <!-- Comments list -->
    <div v-else-if="comments.length > 0" class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
      <div class="divide-y divide-gray-100 dark:divide-gray-800">
        <template v-for="(comment, index) in comments" :key="comment.id">
          <div class="p-4 sm:p-6">
            <CommentItem
              :comment="comment"
              :article-slug="articleSlug"
              :current-user-id="currentUserId"
              :is-expanded="expandedCommentIds.has(comment.id)"
              @reply="handleReply"
              @edit="handleEdit"
              @delete="handleDeleteClick"
              @reaction-toggle="handleReactionToggle"
              @expand-toggle="handleExpandToggle"
            />

            <!-- Reply form for this comment -->
            <div
              v-if="replyingTo?.id === comment.id"
              :data-reply-to="comment.id"
              class="ml-13 mt-4"
            >
              <CommentForm
                :article-slug="articleSlug"
                :parent-id="comment.id"
                :reply-to-author="replyingTo.authorName"
                @submitted="handleCommentSubmitted"
                @cancelled="handleFormCancelled"
              />
            </div>
          </div>
        </template>
      </div>

      <!-- Show more link -->
      <div v-if="comments.length < commentCount" class="border-t border-gray-100 dark:border-gray-800 p-4 text-center bg-gray-50 dark:bg-gray-800/50">
        <button
          class="inline-flex items-center gap-2 text-primary hover:text-primary/80 transition-colors font-medium disabled:opacity-50"
          :disabled="loadingMore"
          @click="loadMoreComments"
        >
          <template v-if="loadingMore">
            <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin" />
            Loading...
          </template>
          <template v-else>
            Show more comments
            <UIcon name="i-heroicons-chevron-down" class="w-4 h-4" />
          </template>
        </button>
      </div>
    </div>

    <!-- Empty state (no longer needed since we show invitation in header) -->
    <div v-else-if="!loading" class="hidden" />

    <!-- Delete confirmation modal -->
    <UModal v-model:open="showDeleteModal">
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center gap-3">
              <UIcon name="i-heroicons-exclamation-triangle" class="w-6 h-6 text-red-500" />
              <h3 class="text-lg font-semibold">Delete Comment</h3>
            </div>
          </template>

          <p class="text-gray-600 dark:text-gray-400">
            Are you sure you want to delete this comment? This action cannot be undone.
          </p>

          <template #footer>
            <div class="flex justify-end gap-3">
              <UButton
                color="neutral"
                variant="outline"
                :disabled="isDeleting"
                @click="showDeleteModal = false"
              >
                Cancel
              </UButton>
              <UButton
                color="error"
                :loading="isDeleting"
                @click="confirmDelete"
              >
                Delete
              </UButton>
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </div>
</template>
