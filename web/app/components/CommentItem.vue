<script setup lang="ts">
import type { Comment } from '~/types'

const props = defineProps<{
  comment: Comment
  articleSlug: string
  isReply?: boolean
  currentUserId?: string
  isExpanded?: boolean
}>()

const emit = defineEmits<{
  reply: [commentId: string, authorName: string]
  edit: [comment: Comment]
  delete: [commentId: string]
  reactionToggle: [commentId: string, reaction: string, hasReacted: boolean]
  refreshReplies: [commentId: string]
  expandToggle: [commentId: string, isExpanded: boolean]
}>()

const api = useApi()
const auth = useAuth()

const showReplies = ref(props.isExpanded || false)
const replies = ref<Comment[]>([])
const loadingReplies = ref(false)

// Sync with parent's expanded state
watch(() => props.isExpanded, (expanded) => {
  showReplies.value = expanded || false
  if (expanded && replies.value.length === 0 && !loadingReplies.value) {
    loadReplies()
  }
}, { immediate: true })

// Get thumbs up and thumbs down counts
const thumbsUpCount = computed(() => {
  const reaction = props.comment.reactions?.find(r => r.reaction === 'thumbsup')
  return reaction?.count || 0
})

const thumbsDownCount = computed(() => {
  const reaction = props.comment.reactions?.find(r => r.reaction === 'thumbsdown')
  return reaction?.count || 0
})

const hasThumbsUp = computed(() => {
  const reaction = props.comment.reactions?.find(r => r.reaction === 'thumbsup')
  return reaction?.has_reacted || false
})

const hasThumbsDown = computed(() => {
  const reaction = props.comment.reactions?.find(r => r.reaction === 'thumbsdown')
  return reaction?.has_reacted || false
})

// Check if author is staff/verified
const isVerified = computed(() => {
  return props.comment.author?.is_system || false
})

// Dropdown menu items
const menuItems = computed(() => {
  const items = []

  if (isOwner.value) {
    items.push({
      label: 'Edit',
      icon: 'i-heroicons-pencil',
      click: () => handleEdit()
    })
  }

  if (canDelete.value) {
    items.push({
      label: 'Delete',
      icon: 'i-heroicons-trash',
      click: () => handleDelete()
    })
  }

  items.push({
    label: 'Report',
    icon: 'i-heroicons-flag',
    click: () => {}
  })

  return [items]
})

const isOwner = computed(() => {
  return props.currentUserId === props.comment.user_id
})

const isAdmin = computed(() => auth.isAdmin.value)

const canDelete = computed(() => isOwner.value || isAdmin.value)

function formatDate(dateString: string): string {
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

function formatContent(content: string): string {
  // Parse markdown-like syntax
  let formatted = content
    // Bold: **text** or __text__
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/__(.*?)__/g, '<strong>$1</strong>')
    // Italic: *text* or _text_
    .replace(/\*([^*]+)\*/g, '<em>$1</em>')
    .replace(/_([^_]+)_/g, '<em>$1</em>')
    // Strikethrough: ~~text~~
    .replace(/~~(.*?)~~/g, '<del>$1</del>')
    // Blockquote: > text
    .replace(/^>\s?(.*)$/gm, '<blockquote class="border-l-4 border-gray-300 dark:border-gray-600 pl-4 my-2 text-gray-600 dark:text-gray-400 italic">$1</blockquote>')
    // @mentions - make them clickable links to user profiles
    .replace(/@([a-zA-Z0-9_-]+)/g, '<a href="/user/$1" class="text-primary font-medium hover:underline">@$1</a>')
    // Line breaks
    .replace(/\n/g, '<br>')

  return formatted
}

// Get user slug from name for profile link
function getUserSlug(name?: string): string {
  if (!name) return ''
  return name.toLowerCase().replace(/\s+/g, '-')
}

async function loadReplies() {
  if (loadingReplies.value || props.isReply) return

  loadingReplies.value = true
  try {
    const authHeaders = auth.isAuthenticated.value ? auth.getAuthHeaders() : undefined
    replies.value = await api.getCommentReplies(props.comment.id, authHeaders)
    showReplies.value = true
    emit('expandToggle', props.comment.id, true)
  } catch (error) {
    console.error('Failed to load replies:', error)
  } finally {
    loadingReplies.value = false
  }
}

function toggleReplies() {
  if (showReplies.value) {
    showReplies.value = false
    emit('expandToggle', props.comment.id, false)
  } else {
    loadReplies()
  }
}

function handleReply() {
  // Expand replies when clicking reply so user can see the thread context
  if (!showReplies.value && (props.comment.reply_count || 0) > 0) {
    loadReplies()
  }
  emit('reply', props.comment.id, props.comment.author?.name || 'User')
}

function handleEdit() {
  emit('edit', props.comment)
}

function handleDelete() {
  emit('delete', props.comment.id)
}

async function handleReactionClick(reaction: string) {
  if (!auth.isAuthenticated.value) return

  const existingReaction = props.comment.reactions?.find(r => r.reaction === reaction)
  const hasReacted = existingReaction?.has_reacted || false

  // For replies, handle locally since parent doesn't track reply data
  if (props.isReply) {
    // Optimistic update
    if (!props.comment.reactions) props.comment.reactions = []
    const reactionObj = props.comment.reactions.find(r => r.reaction === reaction)
    let removedOpposite: string | null = null

    if (hasReacted) {
      // Remove reaction
      if (reactionObj) {
        reactionObj.count = Math.max(0, reactionObj.count - 1)
        reactionObj.has_reacted = false
      }
    } else {
      // Remove opposite reaction first (like/dislike are mutually exclusive)
      const oppositeReaction = reaction === 'thumbsup' ? 'thumbsdown' : reaction === 'thumbsdown' ? 'thumbsup' : null
      if (oppositeReaction) {
        const opposite = props.comment.reactions.find(r => r.reaction === oppositeReaction)
        if (opposite?.has_reacted) {
          opposite.count = Math.max(0, opposite.count - 1)
          opposite.has_reacted = false
          removedOpposite = oppositeReaction
        }
      }

      // Add the new reaction
      if (reactionObj) {
        reactionObj.count++
        reactionObj.has_reacted = true
      } else {
        props.comment.reactions.push({ reaction, count: 1, has_reacted: true })
      }
    }

    // Make API call
    try {
      if (hasReacted) {
        await api.removeReaction(props.comment.id, reaction, auth.getAuthHeaders())
      } else {
        if (removedOpposite) {
          await api.removeReaction(props.comment.id, removedOpposite, auth.getAuthHeaders())
        }
        await api.addReaction(props.comment.id, reaction, auth.getAuthHeaders())
      }
    } catch (e) {
      // Revert on error
      const revertReaction = props.comment.reactions?.find(r => r.reaction === reaction)
      if (revertReaction) {
        if (hasReacted) {
          revertReaction.count++
          revertReaction.has_reacted = true
        } else {
          revertReaction.count = Math.max(0, revertReaction.count - 1)
          revertReaction.has_reacted = false
        }
      }
      // Revert opposite too
      if (removedOpposite) {
        const opposite = props.comment.reactions?.find(r => r.reaction === removedOpposite)
        if (opposite) {
          opposite.count++
          opposite.has_reacted = true
        }
      }
      console.error('Failed to toggle reaction:', e)
    }
  } else {
    // For root comments, emit to parent
    emit('reactionToggle', props.comment.id, reaction, hasReacted)
  }
}

// Refresh replies when parent emits refresh event
watch(() => props.comment, () => {
  if (showReplies.value && !props.isReply) {
    loadReplies()
  }
}, { deep: true })

defineExpose({ loadReplies })
</script>

<template>
  <div :class="['group', isReply ? 'relative pl-8 mt-4' : '']">
    <!-- Vertical connector line for replies -->
    <div
      v-if="isReply"
      class="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200 dark:bg-gray-700"
    />

    <div class="flex gap-3">
      <!-- Avatar -->
      <NuxtLink :to="`/user/${getUserSlug(comment.author?.name)}`" class="flex-shrink-0 relative z-10">
        <NuxtImg
          v-if="comment.author?.avatar"
          :src="comment.author.avatar"
          :alt="comment.author?.name"
          class="w-10 h-10 rounded-full object-cover hover:ring-2 hover:ring-primary transition-all"
        />
        <div v-else class="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center hover:ring-2 hover:ring-primary transition-all">
          <UIcon name="i-heroicons-user" class="w-5 h-5 text-primary" />
        </div>
      </NuxtLink>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <!-- Header -->
        <div class="flex items-center gap-2 flex-wrap">
          <NuxtLink
            :to="`/user/${getUserSlug(comment.author?.name)}`"
            class="font-semibold text-gray-900 dark:text-white hover:text-primary transition-colors"
          >
            {{ comment.author?.name || 'Anonymous' }}
          </NuxtLink>
          <!-- Verified badge -->
          <span
            v-if="isVerified"
            class="inline-flex items-center justify-center w-4 h-4 rounded-full bg-blue-500"
            title="Verified"
          >
            <UIcon name="i-heroicons-check" class="w-3 h-3 text-white" />
          </span>
          <span class="text-sm text-gray-500 dark:text-gray-400">
            {{ formatDate(comment.created_at) }}
          </span>
          <span v-if="comment.updated_at !== comment.created_at" class="text-xs text-gray-400 dark:text-gray-500">
            (edited)
          </span>
        </div>

        <!-- Comment body -->
        <!-- eslint-disable-next-line vue/no-v-html -->
        <div
          class="mt-1 text-gray-700 dark:text-gray-300 break-words"
          v-html="formatContent(comment.content)"
        />

        <!-- Actions row -->
        <div class="mt-3 flex items-center gap-4">
          <!-- Thumbs up -->
          <button
            :class="[
              'inline-flex items-center gap-1.5 text-sm transition-colors',
              hasThumbsUp
                ? 'text-primary'
                : 'text-gray-500 dark:text-gray-400 hover:text-primary'
            ]"
            :disabled="!auth.isAuthenticated.value"
            @click="handleReactionClick('thumbsup')"
          >
            <UIcon name="i-heroicons-hand-thumb-up" class="w-4 h-4" />
            <span>{{ thumbsUpCount }}</span>
          </button>

          <!-- Thumbs down -->
          <button
            :class="[
              'inline-flex items-center gap-1.5 text-sm transition-colors',
              hasThumbsDown
                ? 'text-red-500'
                : 'text-gray-500 dark:text-gray-400 hover:text-red-500'
            ]"
            :disabled="!auth.isAuthenticated.value"
            @click="handleReactionClick('thumbsdown')"
          >
            <UIcon name="i-heroicons-hand-thumb-down" class="w-4 h-4" />
            <span>{{ thumbsDownCount }}</span>
          </button>

          <!-- Reply button (only for root comments) -->
          <button
            v-if="!isReply && auth.isAuthenticated.value"
            class="inline-flex items-center gap-1.5 text-sm text-gray-500 dark:text-gray-400 hover:text-primary transition-colors"
            @click="handleReply"
          >
            <UIcon name="i-heroicons-chat-bubble-left" class="w-4 h-4" />
            <span>Reply</span>
          </button>

          <!-- More menu -->
          <UDropdownMenu :items="menuItems">
            <button class="inline-flex items-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
              <UIcon name="i-heroicons-ellipsis-horizontal" class="w-5 h-5" />
            </button>
          </UDropdownMenu>
        </div>

        <!-- View/hide replies -->
        <button
          v-if="!isReply && (comment.reply_count || 0) > 0"
          class="mt-3 text-sm text-primary hover:text-primary/80 transition-colors flex items-center gap-1"
          @click="toggleReplies"
        >
          <UIcon
            :name="showReplies ? 'i-heroicons-chevron-up' : 'i-heroicons-chevron-down'"
            class="w-4 h-4"
          />
          {{ showReplies ? 'Hide' : 'View' }} {{ comment.reply_count }} {{ comment.reply_count === 1 ? 'reply' : 'replies' }}
        </button>

        <!-- Replies -->
        <div v-if="showReplies && replies.length > 0" class="mt-4 space-y-4">
          <CommentItem
            v-for="reply in replies"
            :key="reply.id"
            :comment="reply"
            :article-slug="articleSlug"
            :is-reply="true"
            :current-user-id="currentUserId"
            @edit="emit('edit', $event)"
            @delete="emit('delete', $event)"
            @reaction-toggle="emit('reactionToggle', $event[0], $event[1], $event[2])"
          />
        </div>

        <!-- Loading replies -->
        <div v-if="loadingReplies" class="mt-4 flex items-center gap-2 text-gray-500">
          <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin" />
          <span class="text-sm">Loading replies...</span>
        </div>
      </div>
    </div>
  </div>
</template>
