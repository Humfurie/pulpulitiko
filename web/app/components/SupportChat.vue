<script setup lang="ts">
import type { Conversation, Message } from '~/types'

const auth = useAuth()
const messaging = useMessaging()
const { soundEnabled, toggleSound } = useNotificationSound()

const isOpen = ref(false)
const showNewChat = ref(false)
const newMessage = ref('')
const newSubject = ref('')
const messageInput = ref('')
const messagesContainer = ref<HTMLElement | null>(null)
const showEmojiPicker = ref(false)
const skipNextMessagesFetch = ref(false)

// Common emojis
const commonEmojis = ['😊', '😂', '❤️', '👍', '🎉', '🙏', '😍', '🤔', '👋', '✨', '🔥', '💯']

// Inactivity timeout in minutes
const INACTIVITY_TIMEOUT_MINUTES = 5

// Check if a conversation is closed (status is closed/archived OR expired due to inactivity)
const isConversationClosed = computed((): boolean => {
  const conv = selectedConversation.value
  if (!conv) return false

  // Check if status is explicitly closed or archived
  if (conv.status === 'closed' || conv.status === 'archived') {
    return true
  }

  // Check if expired due to inactivity (5 minutes since last message from admin without user reply)
  if (conv.last_message_at) {
    const lastMessageTime = new Date(conv.last_message_at).getTime()
    const now = Date.now()
    const minutesSinceLastMessage = (now - lastMessageTime) / (1000 * 60)

    // Only expire if the last message was from admin and user hasn't replied in 5 minutes
    const lastMessage = messaging.messages.value[messaging.messages.value.length - 1]
    if (lastMessage && lastMessage.sender_id !== auth.user.value?.id && minutesSinceLastMessage >= INACTIVITY_TIMEOUT_MINUTES) {
      return true
    }
  }

  return false
})

// Get the reason why conversation is closed
const closedReason = computed((): string => {
  const conv = selectedConversation.value
  if (!conv) return ''

  if (conv.status === 'closed') return 'This conversation has been resolved'
  if (conv.status === 'archived') return 'This conversation has been archived'

  // Check inactivity expiration (user didn't reply to admin's message within 5 minutes)
  if (conv.last_message_at) {
    const lastMessageTime = new Date(conv.last_message_at).getTime()
    const now = Date.now()
    const minutesSinceLastMessage = (now - lastMessageTime) / (1000 * 60)
    const lastMessage = messaging.messages.value[messaging.messages.value.length - 1]

    if (lastMessage && lastMessage.sender_id !== auth.user.value?.id && minutesSinceLastMessage >= INACTIVITY_TIMEOUT_MINUTES) {
      return 'This conversation has expired due to inactivity'
    }
  }

  return ''
})

// Load conversations when opening
watch(isOpen, async (open) => {
  if (open && auth.isAuthenticated.value) {
    await messaging.fetchConversations()
    await messaging.fetchUnreadCounts()
  }
})

// Typed conversations
const conversations = computed((): Conversation[] => messaging.conversations.value as Conversation[])

// Select first conversation or show new chat form
const selectedConversation = computed((): Conversation | null => {
  if (messaging.currentConversation.value) {
    return messaging.currentConversation.value
  }
  if (conversations.value.length > 0 && !showNewChat.value) {
    return conversations.value[0] ?? null
  }
  return null
})

// Load messages when conversation changes
watch(selectedConversation, async (conv) => {
  if (conv) {
    // Skip fetching if we just created this conversation
    // (createConversation already sets messages appropriately)
    if (skipNextMessagesFetch.value) {
      skipNextMessagesFetch.value = false
    } else {
      await messaging.fetchMessages(conv.id)
    }
    messaging.currentConversation.value = conv
    // Scroll to bottom
    nextTick(() => {
      if (messagesContainer.value) {
        messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
      }
    })
  }
})

// Scroll to bottom when messages change
watch(() => messaging.messages.value.length, () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
})

async function handleSelectConversation(conv: Conversation) {
  showNewChat.value = false
  messaging.currentConversation.value = conv
}

function isActiveConversation(convId: string): boolean {
  return selectedConversation.value?.id === convId
}

// Check if a specific conversation is closed/expired
function isConvClosed(conv: Conversation): boolean {
  // Check if status is explicitly closed or archived
  if (conv.status === 'closed' || conv.status === 'archived') {
    return true
  }
  return false
}

async function handleStartNewChat() {
  if (!newMessage.value.trim()) return

  try {
    // Set flag to skip fetchMessages in the watcher
    // since createConversation already sets messages correctly
    skipNextMessagesFetch.value = true
    await messaging.createConversation({
      subject: newSubject.value || undefined,
      message: newMessage.value.trim()
    })
    showNewChat.value = false
    newMessage.value = ''
    newSubject.value = ''
  } catch (error) {
    skipNextMessagesFetch.value = false
    console.error('Failed to start chat:', error)
  }
}

async function handleSendMessage() {
  const content = messageInput.value.trim()
  if (!content || !messaging.currentConversation.value) return

  await messaging.sendMessage(messaging.currentConversation.value.id, { content })
  messageInput.value = ''
  showEmojiPicker.value = false
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    handleSendMessage()
  } else {
    handleTyping()
  }
}

function handleTyping() {
  if (messaging.currentConversation.value) {
    messaging.startTyping(messaging.currentConversation.value.id)
  }
}

async function handleMarkAsRead() {
  if (messaging.currentConversation.value) {
    await messaging.markAsRead(messaging.currentConversation.value.id)
  }
}

function insertEmoji(emoji: string) {
  messageInput.value += emoji
  showEmojiPicker.value = false
}

function isOwnMessage(message: Message): boolean {
  return message.sender_id === auth.user.value?.id
}

function formatTime(dateStr: string): string {
  return new Date(dateStr).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

// Group consecutive messages from same sender
interface MessageCluster {
  isOwn: boolean
  messages: Message[]
}

const groupedMessages = computed((): MessageCluster[] => {
  const clusters: MessageCluster[] = []
  let currentCluster: MessageCluster | null = null

  for (const message of messaging.messages.value) {
    const isOwn = isOwnMessage(message)

    if (currentCluster && currentCluster.isOwn === isOwn) {
      currentCluster.messages.push(message)
    } else {
      currentCluster = { isOwn, messages: [message] }
      clusters.push(currentCluster)
    }
  }

  return clusters
})

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
  <!-- Hide support chat for admins - they use the admin messages page -->
  <div v-if="!auth.isAdmin.value" class="fixed bottom-4 right-4 z-[9999] safe-area-bottom">
    <!-- Chat Button -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0 scale-75"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-75"
    >
      <button
        v-if="!isOpen && auth.isAuthenticated.value"
        class="relative w-14 h-14 bg-gradient-to-br from-primary-500 to-primary-600 hover:from-primary-600 hover:to-primary-700 text-white rounded-full shadow-lg hover:shadow-xl transition-all duration-200 flex items-center justify-center group"
        @click="isOpen = true"
      >
        <UIcon name="i-heroicons-chat-bubble-left-right" class="w-6 h-6 group-hover:scale-110 transition-transform" />
        <!-- Unread badge -->
        <span
          v-if="messaging.unreadCounts.value.total > 0"
          class="absolute -top-1 -right-1 min-w-5 h-5 px-1.5 bg-red-500 text-white text-xs font-bold rounded-full flex items-center justify-center animate-pulse"
        >
          {{ messaging.unreadCounts.value.total > 99 ? '99+' : messaging.unreadCounts.value.total }}
        </span>
        <!-- Pulse animation for new messages -->
        <span
          v-if="messaging.unreadCounts.value.total > 0"
          class="absolute inset-0 rounded-full bg-primary-500 animate-ping opacity-25"
        />
      </button>
    </Transition>

    <!-- Chat Panel -->
    <Transition
      enter-active-class="transition ease-out duration-300"
      enter-from-class="opacity-0 translate-y-8 scale-95"
      enter-to-class="opacity-100 translate-y-0 scale-100"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="opacity-100 translate-y-0 scale-100"
      leave-to-class="opacity-0 translate-y-8 scale-95"
    >
      <div
        v-if="isOpen && auth.isAuthenticated.value"
        class="absolute bottom-0 right-0 w-[380px] h-[520px] bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-800 flex flex-col overflow-hidden"
      >
        <!-- Header -->
        <div class="bg-gradient-to-r from-primary-500 to-primary-600 text-white px-4 py-3">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-white/20 rounded-full flex items-center justify-center">
                <UIcon name="i-heroicons-chat-bubble-left-right" class="w-5 h-5" />
              </div>
              <div>
                <h3 class="font-semibold text-sm">Support Chat</h3>
                <p class="text-xs text-white/80">We typically reply within minutes</p>
              </div>
            </div>
            <div class="flex items-center gap-1">
              <!-- Sound toggle -->
              <button
                class="p-2 hover:bg-white/20 rounded-lg transition-colors"
                :title="soundEnabled ? 'Mute notifications' : 'Unmute notifications'"
                @click="toggleSound"
              >
                <UIcon
                  :name="soundEnabled ? 'i-heroicons-speaker-wave' : 'i-heroicons-speaker-x-mark'"
                  class="w-4 h-4"
                />
              </button>
              <!-- New chat -->
              <button
                class="p-2 hover:bg-white/20 rounded-lg transition-colors"
                title="New conversation"
                @click="showNewChat = true; messaging.currentConversation.value = null"
              >
                <UIcon name="i-heroicons-plus" class="w-4 h-4" />
              </button>
              <!-- Close -->
              <button
                class="p-2 hover:bg-white/20 rounded-lg transition-colors"
                @click="isOpen = false"
              >
                <UIcon name="i-heroicons-x-mark" class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- New Chat Form -->
        <div
          v-if="showNewChat"
          class="flex-1 flex flex-col p-4 bg-gray-50 dark:bg-gray-950"
        >
          <div class="bg-white dark:bg-gray-900 rounded-xl p-4 shadow-sm border border-gray-200 dark:border-gray-800">
            <h4 class="font-medium text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <UIcon name="i-heroicons-pencil-square" class="w-5 h-5 text-primary-500" />
              Start a conversation
            </h4>

            <div class="space-y-3">
              <div>
                <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1.5">Subject (optional)</label>
                <UInput
                  v-model="newSubject"
                  placeholder="What's this about?"
                  size="sm"
                />
              </div>

              <div>
                <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1.5">Message</label>
                <UTextarea
                  v-model="newMessage"
                  placeholder="How can we help you?"
                  :rows="3"
                />
              </div>
            </div>
          </div>

          <div class="flex gap-2 mt-auto pt-4">
            <UButton
              variant="ghost"
              class="flex-1"
              @click="showNewChat = false"
            >
              Cancel
            </UButton>
            <UButton
              color="primary"
              class="flex-1"
              :disabled="!newMessage.trim()"
              :loading="messaging.loading.value"
              @click="handleStartNewChat"
            >
              <UIcon name="i-heroicons-paper-airplane" class="w-4 h-4 mr-1" />
              Send
            </UButton>
          </div>
        </div>

        <!-- Conversations & Chat -->
        <template v-else>
          <!-- Conversation tabs -->
          <div
            v-if="conversations.length > 0"
            class="flex items-center gap-1.5 px-3 py-2 bg-gray-50 dark:bg-gray-950 border-b border-gray-200 dark:border-gray-800 overflow-x-auto scrollbar-hide"
          >
            <button
              v-for="conv in conversations"
              :key="conv.id"
              class="flex-shrink-0 px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200 flex items-center gap-1.5"
              :class="[
                isActiveConversation(conv.id)
                  ? 'bg-primary-500 text-white shadow-sm'
                  : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 border border-gray-200 dark:border-gray-700',
                isConvClosed(conv) && !isActiveConversation(conv.id) ? 'opacity-60' : ''
              ]"
              @click="handleSelectConversation(conv)"
            >
              <!-- Lock icon for closed conversations -->
              <UIcon
                v-if="isConvClosed(conv)"
                name="i-heroicons-lock-closed"
                class="w-3 h-3 flex-shrink-0"
                :class="isActiveConversation(conv.id) ? 'text-white/80' : 'text-gray-400'"
              />
              <span class="truncate max-w-[80px]">{{ conv.subject || 'Support' }}</span>
              <span
                v-if="conv.unread_count && conv.unread_count > 0 && !isConvClosed(conv)"
                class="flex-shrink-0 w-4 h-4 bg-red-500 text-white text-[10px] rounded-full flex items-center justify-center"
              >
                {{ conv.unread_count > 9 ? '9+' : conv.unread_count }}
              </span>
            </button>
          </div>

          <!-- Empty state -->
          <div
            v-if="conversations.length === 0"
            class="flex-1 flex flex-col items-center justify-center p-6 text-center bg-gray-50 dark:bg-gray-950"
          >
            <div class="w-16 h-16 bg-primary-100 dark:bg-primary-900/30 rounded-full flex items-center justify-center mb-4">
              <UIcon name="i-heroicons-inbox" class="w-8 h-8 text-primary-500" />
            </div>
            <h4 class="font-medium text-gray-900 dark:text-white mb-1">No conversations yet</h4>
            <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">Start a conversation with our support team</p>
            <UButton
              color="primary"
              size="sm"
              @click="showNewChat = true"
            >
              <UIcon name="i-heroicons-chat-bubble-left" class="w-4 h-4 mr-1" />
              Start chat
            </UButton>
          </div>

          <!-- Chat messages -->
          <div
            v-else-if="selectedConversation"
            class="flex-1 flex flex-col min-h-0"
            @mouseenter="handleMarkAsRead"
          >
            <!-- Messages container -->
            <div
              ref="messagesContainer"
              class="flex-1 overflow-y-auto p-3 space-y-3 bg-gray-50 dark:bg-gray-950"
            >
              <!-- Loading -->
              <div v-if="messaging.loading.value" class="flex justify-center py-4">
                <UIcon name="i-heroicons-arrow-path" class="w-5 h-5 animate-spin text-gray-400" />
              </div>

              <!-- Message clusters -->
              <template v-else>
                <div
                  v-for="(cluster, idx) in groupedMessages"
                  :key="idx"
                  class="flex"
                  :class="cluster.isOwn ? 'justify-end' : 'justify-start'"
                >
                  <div class="max-w-[80%] space-y-0.5">
                    <div
                      v-for="(message, msgIdx) in cluster.messages"
                      :key="message.id"
                      class="flex items-end gap-1"
                      :class="cluster.isOwn ? 'flex-row-reverse' : ''"
                    >
                      <div
                        class="px-3 py-2 text-sm whitespace-pre-wrap break-words"
                        :class="[
                          cluster.isOwn
                            ? 'bg-primary-500 text-white'
                            : 'bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 shadow-sm border border-gray-100 dark:border-gray-700',
                          // Rounded corners
                          cluster.isOwn
                            ? msgIdx === 0 && cluster.messages.length > 1
                              ? 'rounded-2xl rounded-br-md'
                              : msgIdx === cluster.messages.length - 1 && cluster.messages.length > 1
                                ? 'rounded-2xl rounded-tr-md'
                                : cluster.messages.length > 1
                                  ? 'rounded-2xl rounded-r-md'
                                  : 'rounded-2xl'
                            : msgIdx === 0 && cluster.messages.length > 1
                              ? 'rounded-2xl rounded-bl-md'
                              : msgIdx === cluster.messages.length - 1 && cluster.messages.length > 1
                                ? 'rounded-2xl rounded-tl-md'
                                : cluster.messages.length > 1
                                  ? 'rounded-2xl rounded-l-md'
                                  : 'rounded-2xl'
                        ]"
                      >
                        {{ message.content }}
                      </div>
                    </div>
                    <!-- Time for last message -->
                    <p
                      class="text-[10px] text-gray-400 px-1"
                      :class="cluster.isOwn ? 'text-right' : 'text-left'"
                    >
                      {{ formatTime(cluster.messages[cluster.messages.length - 1]?.created_at || '') }}
                    </p>
                  </div>
                </div>
              </template>

              <!-- Typing indicator -->
              <div
                v-if="messaging.typingUsers.value.length > 0"
                class="flex items-center gap-2"
              >
                <div class="bg-white dark:bg-gray-800 rounded-2xl px-3 py-2 shadow-sm border border-gray-100 dark:border-gray-700">
                  <div class="flex gap-1">
                    <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0ms" />
                    <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 150ms" />
                    <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 300ms" />
                  </div>
                </div>
              </div>
            </div>

            <!-- Input area or Closed message -->
            <div class="p-3 bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-800">
              <!-- Closed conversation message -->
              <div
                v-if="isConversationClosed"
                class="text-center py-2"
              >
                <div class="flex items-center justify-center gap-2 text-gray-500 dark:text-gray-400 mb-2">
                  <UIcon name="i-heroicons-lock-closed" class="w-4 h-4" />
                  <span class="text-sm">{{ closedReason }}</span>
                </div>
                <UButton
                  size="sm"
                  color="primary"
                  variant="soft"
                  @click="showNewChat = true; messaging.currentConversation.value = null"
                >
                  <UIcon name="i-heroicons-plus" class="w-4 h-4 mr-1" />
                  Start new conversation
                </UButton>
              </div>

              <!-- Active input area -->
              <div v-else class="flex items-end gap-2">
                <!-- Emoji picker -->
                <div class="relative emoji-picker-container">
                  <button
                    class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                    @click.stop="showEmojiPicker = !showEmojiPicker"
                  >
                    <UIcon name="i-heroicons-face-smile" class="w-5 h-5" />
                  </button>

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
                      class="absolute bottom-full left-0 mb-2 p-2 bg-white dark:bg-gray-800 rounded-xl shadow-lg border border-gray-200 dark:border-gray-700 z-10"
                      @click.stop
                    >
                      <div class="grid grid-cols-6 gap-1">
                        <button
                          v-for="emoji in commonEmojis"
                          :key="emoji"
                          class="w-7 h-7 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
                          @click="insertEmoji(emoji)"
                        >
                          {{ emoji }}
                        </button>
                      </div>
                    </div>
                  </Transition>
                </div>

                <!-- Input -->
                <div class="flex-1 relative">
                  <textarea
                    v-model="messageInput"
                    placeholder="Type a message..."
                    rows="1"
                    class="w-full px-3 py-2 bg-gray-100 dark:bg-gray-800 border-0 rounded-xl text-sm resize-none focus:ring-2 focus:ring-primary-500 focus:bg-white dark:focus:bg-gray-900 transition-all"
                    @keydown="handleKeydown"
                  />
                </div>

                <!-- Send button -->
                <button
                  class="p-2 text-white bg-primary-500 hover:bg-primary-600 rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  :disabled="!messageInput.trim()"
                  @click="handleSendMessage"
                >
                  <UIcon name="i-heroicons-paper-airplane" class="w-5 h-5" />
                </button>
              </div>
            </div>
          </div>

          <!-- No conversation selected -->
          <div
            v-else
            class="flex-1 flex flex-col items-center justify-center p-4 text-center bg-gray-50 dark:bg-gray-950"
          >
            <UIcon name="i-heroicons-chat-bubble-left-right" class="w-12 h-12 text-gray-300 dark:text-gray-700 mb-2" />
            <p class="text-sm text-gray-500 dark:text-gray-400">Select a conversation above</p>
          </div>
        </template>
      </div>
    </Transition>

    <!-- Login prompt button -->
    <button
      v-if="!auth.isAuthenticated.value"
      class="w-14 h-14 bg-gradient-to-br from-primary-500 to-primary-600 hover:from-primary-600 hover:to-primary-700 text-white rounded-full shadow-lg hover:shadow-xl transition-all duration-200 flex items-center justify-center"
      @click="navigateTo('/login')"
    >
      <UIcon name="i-heroicons-chat-bubble-left-right" class="w-6 h-6" />
    </button>
  </div>
</template>

<style scoped>
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
/* Safe area handling for mobile devices with notches/gesture navigation */
.safe-area-bottom {
  padding-bottom: env(safe-area-inset-bottom, 0);
}
@supports (padding-bottom: env(safe-area-inset-bottom)) {
  .safe-area-bottom {
    bottom: calc(1rem + env(safe-area-inset-bottom, 0));
  }
}
</style>
