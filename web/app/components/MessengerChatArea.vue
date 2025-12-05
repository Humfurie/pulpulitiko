<script setup lang="ts">
import type { Conversation, Message } from '~/types'

interface QuickReply {
  label: string
  text: string
}

const props = defineProps<{
  conversation: Conversation | null
  messages: Message[]
  loading?: boolean
  typingUsers?: string[]
  quickReplies?: QuickReply[]
  showQuickReplies?: boolean
}>()

const emit = defineEmits<{
  send: [content: string]
  typing: []
  markAsRead: []
  toggleQuickReplies: []
  insertQuickReply: [text: string]
}>()

const auth = useAuth()
const messageInput = ref('')
const messagesContainer = ref<HTMLElement | null>(null)
const showEmojiPicker = ref(false)

// Common emojis for quick access
const commonEmojis = [
  '😊', '😂', '❤️', '👍', '🎉', '🙏', '😍', '🤔',
  '👋', '✨', '🔥', '💯', '😎', '🤝', '👏', '💪',
  '😅', '🥰', '😁', '🙌', '💡', '✅', '⭐', '🌟'
]

// Scroll to bottom when messages change
watch(() => props.messages.length, () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
})

// Mark as read when viewing
watch(() => props.conversation?.id, () => {
  if (props.conversation?.id) {
    emit('markAsRead')
  }
})

function handleSend() {
  const content = messageInput.value.trim()
  if (!content) return

  emit('send', content)
  messageInput.value = ''
  showEmojiPicker.value = false
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    handleSend()
  } else {
    emit('typing')
  }
}

function isOwnMessage(message: Message): boolean {
  const userId = auth.user.value?.id
  console.log('isOwnMessage check:', { userId, senderId: message.sender_id, isOwn: message.sender_id === userId })
  if (!userId) return false
  return message.sender_id === userId
}

function insertEmoji(emoji: string) {
  messageInput.value += emoji
  showEmojiPicker.value = false
}

// Group messages by date and consecutive sender
interface MessageGroup {
  date: string
  clusters: MessageCluster[]
}

interface MessageCluster {
  senderId: string
  senderName: string
  senderAvatar?: string
  isOwn: boolean
  messages: Message[]
}

const groupedMessages = computed((): MessageGroup[] => {
  const groups: MessageGroup[] = []
  let currentDate = ''
  let currentCluster: MessageCluster | null = null

  for (const message of props.messages) {
    const messageDate = new Date(message.created_at).toDateString()
    const isOwn = isOwnMessage(message)

    // New date group
    if (messageDate !== currentDate) {
      currentDate = messageDate
      currentCluster = null
      groups.push({
        date: formatDateHeader(message.created_at),
        clusters: []
      })
    }

    const lastGroup = groups[groups.length - 1]
    if (!lastGroup) continue

    // Check if we can add to current cluster (same sender within 5 minutes)
    const canAddToCluster = currentCluster &&
      currentCluster.senderId === message.sender_id &&
      lastGroup.clusters[lastGroup.clusters.length - 1] === currentCluster &&
      isWithinMinutes(currentCluster.messages[currentCluster.messages.length - 1]?.created_at, message.created_at, 5)

    if (canAddToCluster && currentCluster) {
      currentCluster.messages.push(message)
    } else {
      // Start new cluster
      currentCluster = {
        senderId: message.sender_id,
        senderName: message.sender?.name || 'User',
        senderAvatar: message.sender?.avatar,
        isOwn,
        messages: [message]
      }
      lastGroup.clusters.push(currentCluster)
    }
  }

  return groups
})

function isWithinMinutes(date1?: string, date2?: string, minutes: number = 5): boolean {
  if (!date1 || !date2) return false
  const diff = Math.abs(new Date(date1).getTime() - new Date(date2).getTime())
  return diff < minutes * 60 * 1000
}

function formatDateHeader(dateStr: string): string {
  const date = new Date(dateStr)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  if (date.toDateString() === today.toDateString()) {
    return 'Today'
  } else if (date.toDateString() === yesterday.toDateString()) {
    return 'Yesterday'
  } else {
    return date.toLocaleDateString([], {
      weekday: 'long',
      month: 'long',
      day: 'numeric'
    })
  }
}

function formatTime(dateStr: string): string {
  return new Date(dateStr).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

// Close emoji picker on click outside
function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('.emoji-picker-container')) {
    showEmojiPicker.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Quick Replies Panel -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0 -translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-2"
    >
      <div
        v-if="showQuickReplies && quickReplies?.length"
        class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 p-3"
      >
        <div class="flex items-center justify-between mb-2">
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Quick Replies</span>
          <button
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            @click="$emit('toggleQuickReplies')"
          >
            <UIcon name="i-heroicons-x-mark" class="w-4 h-4" />
          </button>
        </div>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="reply in quickReplies"
            :key="reply.label"
            class="px-3 py-1.5 text-xs font-medium bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-full hover:bg-primary-100 dark:hover:bg-primary-900/30 hover:text-primary-700 dark:hover:text-primary-300 transition-colors"
            :title="reply.text"
            @click="$emit('insertQuickReply', reply.text)"
          >
            {{ reply.label }}
          </button>
        </div>
      </div>
    </Transition>

    <!-- Messages Area -->
    <div
      ref="messagesContainer"
      class="flex-1 overflow-y-auto p-4"
    >
      <!-- Loading state -->
      <div v-if="loading" class="flex justify-center py-4">
        <UIcon name="i-heroicons-arrow-path" class="w-6 h-6 animate-spin text-gray-400" />
      </div>

      <!-- Message groups -->
      <template v-else>
        <div
          v-for="group in groupedMessages"
          :key="group.date"
          class="mb-6"
        >
          <!-- Date separator -->
          <div class="flex items-center justify-center my-4">
            <span class="px-3 py-1 text-xs text-gray-500 bg-white dark:bg-gray-800 rounded-full shadow-sm border border-gray-200 dark:border-gray-700">
              {{ group.date }}
            </span>
          </div>

          <!-- Message clusters -->
          <div class="space-y-4">
            <div
              v-for="(cluster, clusterIndex) in group.clusters"
              :key="`${cluster.senderId}-${clusterIndex}`"
              class="flex"
              :class="cluster.isOwn ? 'justify-end' : 'justify-start'"
            >
              <!-- Avatar for non-own messages -->
              <div
                v-if="!cluster.isOwn"
                class="flex-shrink-0 mr-2 self-end"
              >
                <UAvatar
                  :src="cluster.senderAvatar"
                  :alt="cluster.senderName"
                  size="sm"
                />
              </div>

              <!-- Messages bubble group -->
              <div class="max-w-[70%] space-y-0.5">
                <!-- Sender name for first message in cluster -->
                <p
                  v-if="!cluster.isOwn"
                  class="text-xs text-gray-500 dark:text-gray-400 mb-1 ml-1"
                >
                  {{ cluster.senderName }}
                </p>
                <p
                  v-else
                  class="text-xs text-gray-500 dark:text-gray-400 mb-1 mr-1 text-right"
                >
                  You
                </p>

                <!-- Individual messages in cluster -->
                <div
                  v-for="(message, msgIndex) in cluster.messages"
                  :key="message.id"
                  class="group flex items-end gap-2"
                  :class="cluster.isOwn ? 'flex-row-reverse' : ''"
                >
                  <div
                    class="px-4 py-2 text-sm whitespace-pre-wrap break-words"
                    :class="[
                      cluster.isOwn
                        ? 'bg-primary-500 text-white'
                        : 'bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 shadow-sm border border-gray-100 dark:border-gray-700',
                      // Rounded corners based on position
                      cluster.isOwn
                        ? msgIndex === 0 && cluster.messages.length > 1
                          ? 'rounded-2xl rounded-br-md'
                          : msgIndex === cluster.messages.length - 1 && cluster.messages.length > 1
                            ? 'rounded-2xl rounded-tr-md'
                            : cluster.messages.length > 1
                              ? 'rounded-2xl rounded-r-md'
                              : 'rounded-2xl'
                        : msgIndex === 0 && cluster.messages.length > 1
                          ? 'rounded-2xl rounded-bl-md'
                          : msgIndex === cluster.messages.length - 1 && cluster.messages.length > 1
                            ? 'rounded-2xl rounded-tl-md'
                            : cluster.messages.length > 1
                              ? 'rounded-2xl rounded-l-md'
                              : 'rounded-2xl'
                    ]"
                  >
                    {{ message.content }}
                  </div>

                  <!-- Time tooltip on hover (only for last message or on hover) -->
                  <span
                    class="text-[10px] text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
                    :class="msgIndex === cluster.messages.length - 1 ? 'opacity-100' : ''"
                  >
                    {{ formatTime(message.created_at) }}
                    <template v-if="cluster.isOwn && msgIndex === cluster.messages.length - 1">
                      <UIcon
                        v-if="message.is_read"
                        name="i-heroicons-check-circle"
                        class="w-3 h-3 text-primary-400 inline ml-0.5"
                      />
                      <UIcon
                        v-else
                        name="i-heroicons-check"
                        class="w-3 h-3 text-gray-400 inline ml-0.5"
                      />
                    </template>
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Typing indicator -->
      <div
        v-if="typingUsers && typingUsers.length > 0"
        class="flex items-center gap-2 text-gray-500 mt-4"
      >
        <div class="bg-gray-100 dark:bg-gray-800 rounded-2xl px-4 py-2 flex items-center gap-2">
          <div class="flex gap-1">
            <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0ms" />
            <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 150ms" />
            <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 300ms" />
          </div>
        </div>
      </div>
    </div>

    <!-- Input Area -->
    <div
      v-if="conversation"
      class="bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-800 p-4"
    >
      <div class="flex items-end gap-2">
        <!-- Quick replies toggle -->
        <UButton
          icon="i-heroicons-bolt"
          variant="ghost"
          size="sm"
          :color="showQuickReplies ? 'primary' : 'neutral'"
          class="flex-shrink-0"
          title="Quick Replies"
          @click="$emit('toggleQuickReplies')"
        />

        <!-- Emoji picker toggle -->
        <div class="relative emoji-picker-container">
          <UButton
            icon="i-heroicons-face-smile"
            variant="ghost"
            size="sm"
            :color="showEmojiPicker ? 'primary' : 'neutral'"
            class="flex-shrink-0"
            @click.stop="showEmojiPicker = !showEmojiPicker"
          />

          <!-- Emoji picker dropdown -->
          <Transition
            enter-active-class="transition ease-out duration-100"
            enter-from-class="opacity-0 scale-95"
            enter-to-class="opacity-100 scale-100"
            leave-active-class="transition ease-in duration-75"
            leave-from-class="opacity-100 scale-100"
            leave-to-class="opacity-0 scale-95"
          >
            <div
              v-if="showEmojiPicker"
              class="absolute bottom-full left-0 mb-2 p-3 bg-white dark:bg-gray-800 rounded-xl shadow-lg border border-gray-200 dark:border-gray-700 z-10"
              @click.stop
            >
              <div class="grid grid-cols-8 gap-1">
                <button
                  v-for="emoji in commonEmojis"
                  :key="emoji"
                  class="w-8 h-8 flex items-center justify-center text-lg hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
                  @click="insertEmoji(emoji)"
                >
                  {{ emoji }}
                </button>
              </div>
            </div>
          </Transition>
        </div>

        <!-- Message input -->
        <UTextarea
          v-model="messageInput"
          placeholder="Type a message..."
          :rows="1"
          autoresize
          :maxrows="4"
          class="flex-1"
          @keydown="handleKeydown"
        />

        <!-- Send button -->
        <UButton
          icon="i-heroicons-paper-airplane"
          color="primary"
          size="sm"
          class="flex-shrink-0"
          :disabled="!messageInput.trim()"
          @click="handleSend"
        />
      </div>
    </div>
  </div>
</template>
