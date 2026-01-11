<script setup lang="ts">
import type {Comment, CommentAuthor} from '~/types'

const props = defineProps<{
  articleSlug: string
  parentId?: string
  replyToAuthor?: string
  editingComment?: Comment | null
}>()

const emit = defineEmits<{
  submitted: [comment: Comment]
  cancelled: []
}>()

const api = useApi()
const auth = useAuth()
const { sanitizeComment } = useSanitizedHtml()

const content = ref('')
const isSubmitting = ref(false)
const error = ref('')
const showPreview = ref(false)
const showEmojiPicker = ref(false)
const showMentionDropdown = ref(false)
const mentionSearch = ref('')
const mentionUsers = ref<CommentAuthor[]>([])
const mentionPosition = ref({ top: 0, left: 0 })
const selectedMentionIndex = ref(0)
const textareaRef = ref<HTMLTextAreaElement | null>(null)

// Emoji categories with comprehensive emoji lists
const emojiCategories = [
  {
    name: 'Smileys',
    emojis: ['😀', '😃', '😄', '😁', '😆', '😅', '🤣', '😂', '🙂', '🙃', '😉', '😊', '😇', '🥰', '😍', '🤩', '😘', '😗', '😚', '😙', '🥲', '😋', '😛', '😜', '🤪', '😝', '🤑', '🤗', '🤭', '🤫', '🤔', '🤐', '🤨', '😐', '😑', '😶', '😏', '😒', '🙄', '😬', '🤥', '😌', '😔', '😪', '🤤', '😴', '😷', '🤒', '🤕', '🤢', '🤮', '🤧', '🥵', '🥶', '🥴', '😵', '🤯', '🤠', '🥳', '🥸', '😎', '🤓', '🧐']
  },
  {
    name: 'Gestures',
    emojis: ['👋', '🤚', '🖐️', '✋', '🖖', '👌', '🤌', '🤏', '✌️', '🤞', '🤟', '🤘', '🤙', '👈', '👉', '👆', '🖕', '👇', '☝️', '👍', '👎', '✊', '👊', '🤛', '🤜', '👏', '🙌', '👐', '🤲', '🤝', '🙏', '✍️', '💪', '🦾', '🦿']
  },
  {
    name: 'Hearts',
    emojis: ['❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔', '❤️‍🔥', '❤️‍🩹', '❣️', '💕', '💞', '💓', '💗', '💖', '💘', '💝', '💟']
  },
  {
    name: 'Symbols',
    emojis: ['✅', '❌', '❓', '❗', '💯', '🔥', '⭐', '🌟', '✨', '💫', '💥', '💢', '💤', '💬', '💭', '🗯️', '♠️', '♣️', '♥️', '♦️', '🎵', '🎶', '➕', '➖', '➗', '✖️', '♾️', '💲', '💱']
  },
  {
    name: 'Objects',
    emojis: ['📱', '💻', '🖥️', '🖨️', '⌨️', '🖱️', '💾', '📷', '📹', '🎥', '📺', '📻', '🎙️', '🎚️', '🎛️', '⏰', '📡', '🔋', '🔌', '💡', '🔦', '📚', '📖', '📰', '🗞️', '📄', '📃', '📑', '🔖', '🏷️']
  },
  {
    name: 'Nature',
    emojis: ['🌸', '🌺', '🌻', '🌼', '🌷', '🌹', '🥀', '🌾', '🍀', '🍁', '🍂', '🍃', '🌿', '☘️', '🌱', '🌲', '🌳', '🌴', '🌵', '🌊', '🌈', '☀️', '🌤️', '⛅', '🌥️', '☁️', '🌦️', '🌧️', '⛈️', '🌩️', '🌪️', '🌫️']
  },
  {
    name: 'Food',
    emojis: ['🍎', '🍊', '🍋', '🍌', '🍉', '🍇', '🍓', '🫐', '🍈', '🍒', '🍑', '🥭', '🍍', '🥥', '🥝', '🍅', '🥑', '🍆', '🥦', '🥬', '🥒', '🌶️', '🫑', '🌽', '🥕', '🧄', '🧅', '🥔', '🍠', '🥐', '🥯', '🍞', '🥖', '🥨', '🧀', '🥚', '🍳', '🧈', '🥞', '🧇', '🥓', '🥩', '🍗', '🍖', '🦴', '🌭', '🍔', '🍟', '🍕', '🫓', '🥪', '🥙', '🧆', '🌮', '🌯', '🫔', '🥗', '🥘']
  },
  {
    name: 'Activities',
    emojis: ['⚽', '🏀', '🏈', '⚾', '🥎', '🎾', '🏐', '🏉', '🥏', '🎱', '🪀', '🏓', '🏸', '🏒', '🏑', '🥍', '🏏', '🪃', '🥅', '⛳', '🪁', '🏹', '🎣', '🤿', '🥊', '🥋', '🎽', '🛹', '🛼', '🛷', '⛸️', '🥌', '🎿', '⛷️', '🏂', '🎯', '🎮', '🎲', '🧩', '🎰', '🎳']
  }
]

const selectedEmojiCategory = ref(0)

// Initialize content when editing
watch(() => props.editingComment, (comment) => {
  if (comment) {
    content.value = comment.content
  } else {
    content.value = ''
  }
}, { immediate: true })

// Add @mention when replying
watch(() => props.replyToAuthor, (author) => {
  if (author && !props.editingComment) {
    const mention = `@${author.toLowerCase().replace(/\s+/g, '-')} `
    if (!content.value.startsWith(mention)) {
      content.value = mention + content.value
    }
  }
}, { immediate: true })

const formattedPreview = computed(() => {
  return content.value
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
      // @mentions - make them clickable links
      .replace(/@([a-zA-Z0-9_-]+)/g, '<a href="/user/$1" class="text-primary font-medium hover:underline">@$1</a>')
      // Line breaks
      .replace(/\n/g, '<br>')
})

// Sanitize preview for safe display (XSS protection)
const sanitizedPreview = computed(() => sanitizeComment(formattedPreview.value))

// Filter users for mention dropdown
const filteredMentionUsers = computed(() => {
  if (!mentionSearch.value) return mentionUsers.value
  const search = mentionSearch.value.toLowerCase()
  return mentionUsers.value.filter(user =>
    user.name.toLowerCase().includes(search)
  )
})

// Insert emoji at cursor position
function insertEmoji(emoji: string) {
  const textarea = textareaRef.value
  if (!textarea) {
    content.value += emoji
    showEmojiPicker.value = false
    return
  }

  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  content.value = content.value.substring(0, start) + emoji + content.value.substring(end)
  showEmojiPicker.value = false

  nextTick(() => {
    textarea.focus()
    const newCursorPos = start + emoji.length
    textarea.setSelectionRange(newCursorPos, newCursorPos)
  })
}

// Handle @ mention trigger
function handleInput(event: Event) {
  const textarea = event.target as HTMLTextAreaElement
  textareaRef.value = textarea
  const cursorPos = textarea.selectionStart
  const textBeforeCursor = content.value.substring(0, cursorPos)

  // Check if we're in a mention context (@ followed by letters)
  const mentionMatch = textBeforeCursor.match(/@([a-zA-Z0-9_-]*)$/)

  if (mentionMatch) {
    mentionSearch.value = mentionMatch[1] || ''
    showMentionDropdown.value = true
    selectedMentionIndex.value = 0

    // Position the dropdown near the cursor
    const rect = textarea.getBoundingClientRect()
    mentionPosition.value = {
      top: rect.bottom + window.scrollY,
      left: rect.left + window.scrollX
    }

    // Load users if not already loaded
    if (mentionUsers.value.length === 0) {
      loadMentionUsers()
    }
  } else {
    showMentionDropdown.value = false
  }
}

// Load users that can be mentioned
async function loadMentionUsers() {
  try {
    mentionUsers.value = await api.getMentionableUsers()
  } catch (e) {
    console.error('Failed to load mentionable users:', e)
  }
}

// Insert selected mention
function insertMention(user: CommentAuthor) {
  const textarea = textareaRef.value
  if (!textarea) return

  const cursorPos = textarea.selectionStart
  const textBeforeCursor = content.value.substring(0, cursorPos)
  const textAfterCursor = content.value.substring(cursorPos)

  // Find the @ symbol position
  const atIndex = textBeforeCursor.lastIndexOf('@')
  if (atIndex === -1) return

  // Replace @search with @username
  const slug = user.name.toLowerCase().replace(/\s+/g, '-')
  content.value = textBeforeCursor.substring(0, atIndex) + '@' + slug + ' ' + textAfterCursor
  showMentionDropdown.value = false

  nextTick(() => {
    textarea.focus()
    const newCursorPos = atIndex + slug.length + 2 // +2 for @ and space
    textarea.setSelectionRange(newCursorPos, newCursorPos)
  })
}

// Handle keyboard navigation in mention dropdown
function handleKeydown(event: KeyboardEvent) {
  if (!showMentionDropdown.value) return

  const users = filteredMentionUsers.value
  if (users.length === 0) return

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    selectedMentionIndex.value = (selectedMentionIndex.value + 1) % users.length
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    selectedMentionIndex.value = (selectedMentionIndex.value - 1 + users.length) % users.length
  } else if (event.key === 'Enter' && showMentionDropdown.value) {
    event.preventDefault()
    const selectedUser = users[selectedMentionIndex.value]
    if (selectedUser) {
      insertMention(selectedUser)
    }
  } else if (event.key === 'Escape') {
    showMentionDropdown.value = false
  }
}

// Toggle emoji picker
function toggleEmojiPicker() {
  showEmojiPicker.value = !showEmojiPicker.value
}

// Open mention dropdown manually
function openMentionPicker() {
  const textarea = textareaRef.value || document.querySelector('textarea')
  if (!textarea) return

  textareaRef.value = textarea as HTMLTextAreaElement
  const cursorPos = textarea.selectionStart
  content.value = content.value.substring(0, cursorPos) + '@' + content.value.substring(cursorPos)

  nextTick(() => {
    ;(textarea as HTMLTextAreaElement).focus()
    ;(textarea as HTMLTextAreaElement).setSelectionRange(cursorPos + 1, cursorPos + 1)
    // Trigger the input handler to show dropdown
    handleInput({ target: textarea } as unknown as Event)
  })
}

// Close dropdowns when clicking outside
function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('.emoji-picker-container')) {
    showEmojiPicker.value = false
  }
  if (!target.closest('.mention-dropdown')) {
    showMentionDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

const placeholder = computed(() => {
  if (props.editingComment) return 'Edit your comment...'
  if (props.parentId) return `Reply to ${props.replyToAuthor || 'comment'}...`
  return 'Add comment...'
})

const _submitLabel = computed(() => {
  if (isSubmitting.value) return 'Submitting...'
  if (props.editingComment) return 'Update'
  return props.parentId ? 'Reply' : 'Comment'
})

function insertFormat(format: string) {
  const textarea = document.querySelector('textarea')
  if (!textarea) return

  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const selected = content.value.substring(start, end)

  let before = ''
  let after = ''

  switch (format) {
    case 'bold':
      before = '**'
      after = '**'
      break
    case 'italic':
      before = '*'
      after = '*'
      break
    case 'strikethrough':
      before = '~~'
      after = '~~'
      break
    case 'quote':
      before = '> '
      after = ''
      break
  }

  content.value = content.value.substring(0, start) + before + selected + after + content.value.substring(end)

  // Restore cursor position
  nextTick(() => {
    textarea.focus()
    const newCursorPos = start + before.length + selected.length + after.length
    textarea.setSelectionRange(newCursorPos, newCursorPos)
  })
}

async function handleSubmit() {
  if (!content.value.trim() || isSubmitting.value) return

  isSubmitting.value = true
  error.value = ''

  try {
    const authHeaders = auth.getAuthHeaders()
    let result: Comment

    if (props.editingComment) {
      result = await api.updateComment(props.editingComment.id, content.value, authHeaders)
    } else {
      result = await api.createComment(props.articleSlug, {
        content: content.value,
        parent_id: props.parentId
      }, authHeaders)
    }

    content.value = ''
    showPreview.value = false
    emit('submitted', result)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to submit comment'
  } finally {
    isSubmitting.value = false
  }
}

function handleCancel() {
  content.value = ''
  showPreview.value = false
  emit('cancelled')
}
</script>

<template>
  <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
    <!-- Login prompt for unauthenticated users -->
    <div v-if="!auth.isAuthenticated.value" class="text-center py-8 px-4">
      <p class="text-gray-500 dark:text-gray-400 mb-4">
        Sign in to join the discussion
      </p>
      <UButton to="/login" color="primary">
        Sign In
      </UButton>
    </div>

    <!-- Comment form for authenticated users -->
    <form v-else @submit.prevent="handleSubmit">
      <!-- Header with reply info or edit info -->
      <div v-if="parentId || editingComment" class="flex items-center justify-between px-4 py-2 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
        <span class="text-sm text-gray-600 dark:text-gray-400">
          <template v-if="editingComment">Editing comment</template>
          <template v-else>Replying to <span class="font-medium text-gray-900 dark:text-white">{{ replyToAuthor }}</span></template>
        </span>
        <button
          type="button"
          class="text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
          @click="handleCancel"
        >
          Cancel
        </button>
      </div>

      <!-- Textarea area -->
      <div class="p-4 relative">
        <div v-if="!showPreview">
          <UTextarea
            v-model="content"
            :placeholder="placeholder"
            :rows="2"
            autoresize
            variant="none"
            class="w-full text-gray-900 dark:text-white placeholder-gray-400 resize-none border-0 focus:ring-0 p-0 bg-transparent"
            :ui="{ base: 'border-0 focus:ring-0 bg-transparent' }"
            @input="handleInput"
            @keydown="handleKeydown"
          />
        </div>

        <!-- Mention dropdown -->
        <div
          v-if="showMentionDropdown && filteredMentionUsers.length > 0"
          class="mention-dropdown absolute z-50 mt-1 w-full sm:w-64 max-w-[calc(100vw-2rem)] bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 max-h-48 overflow-y-auto"
        >
          <button
            v-for="(user, index) in filteredMentionUsers"
            :key="user.id"
            type="button"
            :class="[
              'w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors',
              index === selectedMentionIndex ? 'bg-gray-100 dark:bg-gray-700' : ''
            ]"
            @click="insertMention(user)"
          >
            <NuxtImg
              v-if="user.avatar"
              :src="user.avatar"
              :alt="user.name"
              class="w-8 h-8 rounded-full object-cover"
            />
            <span v-else class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
              <UIcon name="i-heroicons-user" class="w-4 h-4 text-primary" />
            </span>
            <span class="text-sm font-medium text-gray-900 dark:text-white">{{ user.name }}</span>
          </button>
        </div>
        <div v-else class="min-h-[52px]">
          <div
            v-if="content.trim()"
            class="text-gray-700 dark:text-gray-300"
            v-html="sanitizedPreview"
          />
          <span v-else class="text-gray-400">Nothing to preview</span>
        </div>
      </div>

      <!-- Error message -->
      <div v-if="error" class="mx-4 mb-2">
        <UAlert
          color="error"
          variant="subtle"
          :description="error"
        />
        <!-- Show logout button for session errors -->
        <div v-if="error.includes('session invalid') || error.includes('log out')" class="mt-2 text-center">
          <UButton
            color="primary"
            size="sm"
            @click="auth.logout(); navigateTo('/login')"
          >
            Log out and sign in again
          </UButton>
        </div>
      </div>

      <!-- Toolbar and submit -->
      <div class="flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-800">
        <!-- Formatting toolbar -->
        <div class="flex items-center gap-1">
          <UButton
            type="button"
            color="neutral"
            variant="ghost"
            size="xs"
            title="Bold"
            class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            @click="insertFormat('bold')"
          >
            <span class="font-bold text-sm">B</span>
          </UButton>
          <UButton
            type="button"
            color="neutral"
            variant="ghost"
            size="xs"
            title="Italic"
            class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            @click="insertFormat('italic')"
          >
            <span class="italic text-sm">I</span>
          </UButton>
          <UButton
            type="button"
            color="neutral"
            variant="ghost"
            size="xs"
            title="Strikethrough"
            class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            @click="insertFormat('strikethrough')"
          >
            <span class="line-through text-sm">S</span>
          </UButton>

          <div class="w-px h-4 bg-gray-200 dark:bg-gray-700 mx-1" />

          <!-- Emoji picker -->
          <div class="relative emoji-picker-container">
            <UButton
              type="button"
              color="neutral"
              variant="ghost"
              size="xs"
              icon="i-heroicons-face-smile"
              title="Add emoji"
              class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              @click.stop="toggleEmojiPicker"
            />
            <!-- Emoji picker dropdown -->
            <div
              v-if="showEmojiPicker"
              class="absolute bottom-full left-0 mb-2 w-full sm:w-80 max-w-[calc(100vw-2rem)] bg-white dark:bg-gray-800 rounded-lg shadow-xl border border-gray-200 dark:border-gray-700 z-50"
              @click.stop
            >
              <!-- Category tabs -->
              <div class="flex gap-1 p-2 border-b border-gray-200 dark:border-gray-700 overflow-x-auto">
                <button
                  v-for="(category, index) in emojiCategories"
                  :key="category.name"
                  type="button"
                  :class="[
                    'px-2 py-1 text-xs rounded whitespace-nowrap transition-colors',
                    selectedEmojiCategory === index
                      ? 'bg-primary text-white'
                      : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
                  ]"
                  @click="selectedEmojiCategory = index"
                >
                  {{ category.name }}
                </button>
              </div>
              <!-- Emoji grid -->
              <div class="p-2 max-h-48 overflow-y-auto">
                <div class="grid grid-cols-8 gap-1">
                  <button
                    v-for="emoji in emojiCategories[selectedEmojiCategory]?.emojis ?? []"
                    :key="emoji"
                    type="button"
                    class="w-8 h-8 flex items-center justify-center text-xl hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
                    @click="insertEmoji(emoji)"
                  >
                    {{ emoji }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Mention button -->
          <UButton
            type="button"
            color="neutral"
            variant="ghost"
            size="xs"
            icon="i-heroicons-at-symbol"
            title="Mention someone"
            class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            @click="openMentionPicker"
          />
        </div>

        <!-- Submit button -->
        <UButton
          type="submit"
          color="primary"
          :disabled="!content.trim() || isSubmitting"
          :loading="isSubmitting"
          class="px-6"
        >
          Submit
        </UButton>
      </div>
    </form>
  </div>
</template>
