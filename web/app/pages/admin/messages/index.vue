<script setup lang="ts">
import type { Conversation, ConversationStatus } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const messaging = useMessaging()
const route = useRoute()
const router = useRouter()

// State
const selectedConversationId = ref<string | null>(null)
const statusFilter = ref<ConversationStatus | ''>('')
const searchQuery = ref('')
const page = ref(1)
const perPage = 20
const showQuickReplies = ref(false)

// Quick replies / canned responses
const quickReplies = [
  { label: 'Greeting', text: 'Hello! Thank you for reaching out to us. How can I help you today?' },
  { label: 'Processing', text: 'I\'m looking into this for you. Please give me a moment.' },
  { label: 'Need Info', text: 'Could you please provide more details about your issue?' },
  { label: 'Resolved', text: 'I\'m glad I could help! Is there anything else you need?' },
  { label: 'Follow Up', text: 'I wanted to follow up on your inquiry. Have you had a chance to review my previous message?' },
  { label: 'Escalate', text: 'I\'ll escalate this to our senior team. You\'ll receive a response within 24 hours.' },
]

// Load conversations
const { data: conversationsData, pending, refresh } = await useAsyncData(
  'admin-conversations',
  () => messaging.fetchAdminConversations(page.value, perPage, statusFilter.value || undefined),
  { watch: [page, statusFilter] }
)

const conversations = computed(() => {
  const convs = conversationsData.value?.conversations || []
  if (!searchQuery.value) return convs
  const query = searchQuery.value.toLowerCase()
  return convs.filter(c =>
    c.user?.name?.toLowerCase().includes(query) ||
    c.subject?.toLowerCase().includes(query) ||
    c.last_message?.content?.toLowerCase().includes(query)
  )
})

const totalPages = computed(() => conversationsData.value?.total_pages || 1)

// Load selected conversation from route
watch(
  () => route.query.id,
  async (id) => {
    if (id && typeof id === 'string') {
      selectedConversationId.value = id
      await messaging.fetchConversation(id)
      await messaging.fetchMessages(id)
      // Mark as read immediately when opening conversation
      await messaging.markAsRead(id)
      await refresh()
    } else {
      selectedConversationId.value = null
      messaging.currentConversation.value = null
    }
  },
  { immediate: true }
)

function selectConversation(conv: Conversation) {
  router.push({ query: { id: conv.id } })
}

function closeConversation() {
  router.push({ query: {} })
}

async function handleSendMessage(content: string) {
  if (!selectedConversationId.value) return
  await messaging.sendMessage(selectedConversationId.value, { content })
  showQuickReplies.value = false
}

function handleTyping() {
  if (selectedConversationId.value) {
    messaging.startTyping(selectedConversationId.value)
  }
}

async function handleMarkAsRead() {
  if (selectedConversationId.value) {
    await messaging.markAsRead(selectedConversationId.value)
    await refresh()
  }
}

async function handleUpdateStatus(status: ConversationStatus) {
  if (!selectedConversationId.value) return
  await messaging.updateConversationStatus(selectedConversationId.value, { status })
  await refresh()
}

function insertQuickReply(text: string) {
  handleSendMessage(text)
}

const statusOptions = [
  { value: '', label: 'All Chats' },
  { value: 'open', label: 'Open' },
  { value: 'closed', label: 'Closed' },
  { value: 'archived', label: 'Archived' }
]

// Stats
const openCount = computed(() =>
  (conversationsData.value?.conversations || []).filter(c => c.status === 'open').length
)
const totalUnread = computed(() =>
  (conversationsData.value?.conversations || []).reduce((sum, c) => sum + (c.unread_count || 0), 0)
)
</script>

<template>
  <div class="h-[calc(100vh-4rem)] -m-8 flex">
    <!-- Left Sidebar - Conversations List -->
    <div class="w-80 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col">
      <!-- Header -->
      <div class="p-4 border-b border-gray-200 dark:border-gray-800">
        <div class="flex items-center justify-between mb-4">
          <h1 class="text-xl font-bold text-gray-900 dark:text-white">Messages</h1>
          <div class="flex items-center gap-2">
            <UButton
              icon="i-heroicons-arrow-path"
              variant="ghost"
              size="sm"
              :loading="pending"
              @click="refresh()"
            />
          </div>
        </div>

        <!-- Search -->
        <UInput
          v-model="searchQuery"
          icon="i-heroicons-magnifying-glass"
          placeholder="Search conversations..."
          size="sm"
          class="mb-3"
        />

        <!-- Filter tabs -->
        <div class="flex gap-1">
          <button
            v-for="option in statusOptions"
            :key="option.value"
            class="px-3 py-1.5 text-xs font-medium rounded-full transition-colors"
            :class="statusFilter === option.value
              ? 'bg-primary-500 text-white'
              : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'"
            @click="statusFilter = option.value as ConversationStatus | ''"
          >
            {{ option.label }}
          </button>
        </div>
      </div>

      <!-- Stats bar -->
      <div class="px-4 py-2 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800 flex items-center gap-4 text-xs">
        <div class="flex items-center gap-1">
          <span class="w-2 h-2 bg-green-500 rounded-full"></span>
          <span class="text-gray-600 dark:text-gray-400">{{ openCount }} open</span>
        </div>
        <div v-if="totalUnread > 0" class="flex items-center gap-1">
          <span class="w-2 h-2 bg-primary-500 rounded-full"></span>
          <span class="text-gray-600 dark:text-gray-400">{{ totalUnread }} unread</span>
        </div>
        <div class="text-gray-500 dark:text-gray-500">
          {{ conversationsData?.total || 0 }} total
        </div>
      </div>

      <!-- Conversations list -->
      <div class="flex-1 overflow-y-auto">
        <div v-if="pending" class="flex justify-center py-8">
          <UIcon name="i-heroicons-arrow-path" class="w-6 h-6 animate-spin text-gray-400" />
        </div>

        <div
          v-else-if="conversations.length === 0"
          class="flex flex-col items-center justify-center h-full p-4 text-center"
        >
          <UIcon name="i-heroicons-inbox" class="w-12 h-12 text-gray-300 dark:text-gray-700 mb-2" />
          <p class="text-gray-500 dark:text-gray-400">No conversations found</p>
        </div>

        <div v-else>
          <div
            v-for="conv in conversations"
            :key="conv.id"
            class="px-3 py-2 cursor-pointer transition-all border-l-2"
            :class="[
              selectedConversationId === conv.id
                ? 'bg-primary-50 dark:bg-primary-900/20 border-primary-500'
                : 'border-transparent hover:bg-gray-50 dark:hover:bg-gray-800/50'
            ]"
            @click="selectConversation(conv)"
          >
            <div class="flex items-start gap-3">
              <!-- Avatar with online indicator -->
              <div class="relative flex-shrink-0">
                <UAvatar
                  :src="conv.user?.avatar"
                  :alt="conv.user?.name || 'User'"
                  size="md"
                />
                <!-- Online status indicator -->
                <span
                  class="absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 rounded-full border-2 border-white dark:border-gray-900"
                  :class="conv.status === 'open' ? 'bg-green-500' : 'bg-gray-400'"
                />
              </div>

              <!-- Content -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between gap-2">
                  <h4
                    class="text-sm truncate"
                    :class="(conv.unread_count || 0) > 0
                      ? 'font-semibold text-gray-900 dark:text-white'
                      : 'font-medium text-gray-700 dark:text-gray-300'"
                  >
                    {{ conv.user?.name || 'User' }}
                  </h4>
                  <span class="text-[11px] text-gray-400 flex-shrink-0">
                    {{ formatTimeAgo(conv.last_message_at || conv.created_at) }}
                  </span>
                </div>

                <p
                  v-if="conv.subject"
                  class="text-xs text-primary-600 dark:text-primary-400 truncate"
                >
                  {{ conv.subject }}
                </p>

                <div class="flex items-center justify-between gap-2 mt-0.5">
                  <p
                    class="text-xs truncate"
                    :class="(conv.unread_count || 0) > 0
                      ? 'text-gray-700 dark:text-gray-200'
                      : 'text-gray-500 dark:text-gray-400'"
                  >
                    {{ conv.last_message?.content || 'No messages yet' }}
                  </p>
                  <UBadge
                    v-if="(conv.unread_count || 0) > 0"
                    color="primary"
                    size="xs"
                    class="flex-shrink-0"
                  >
                    {{ conv.unread_count }}
                  </UBadge>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div
        v-if="totalPages > 1"
        class="p-3 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50"
      >
        <div class="flex items-center justify-between">
          <UButton
            icon="i-heroicons-chevron-left"
            variant="ghost"
            size="xs"
            :disabled="page === 1"
            @click="page--"
          />
          <span class="text-xs text-gray-500">
            {{ page }} / {{ totalPages }}
          </span>
          <UButton
            icon="i-heroicons-chevron-right"
            variant="ghost"
            size="xs"
            :disabled="page === totalPages"
            @click="page++"
          />
        </div>
      </div>
    </div>

    <!-- Right Panel - Chat Area -->
    <div class="flex-1 flex flex-col bg-gray-50 dark:bg-gray-950">
      <!-- Empty state -->
      <div
        v-if="!messaging.currentConversation.value"
        class="flex-1 flex items-center justify-center"
      >
        <div class="text-center">
          <div class="w-20 h-20 bg-gray-200 dark:bg-gray-800 rounded-full flex items-center justify-center mx-auto mb-4">
            <UIcon name="i-heroicons-chat-bubble-left-right" class="w-10 h-10 text-gray-400" />
          </div>
          <h2 class="text-lg font-medium text-gray-900 dark:text-white mb-1">Select a conversation</h2>
          <p class="text-sm text-gray-500 dark:text-gray-400">Choose from your existing conversations<br>to start messaging</p>
        </div>
      </div>

      <!-- Chat view -->
      <template v-else>
        <!-- Chat header -->
        <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-6 py-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-4">
              <div class="relative">
                <UAvatar
                  :src="messaging.currentConversation.value.user?.avatar"
                  :alt="messaging.currentConversation.value.user?.name || 'User'"
                  size="lg"
                />
                <span
                  class="absolute -bottom-0.5 -right-0.5 w-4 h-4 rounded-full border-2 border-white dark:border-gray-900"
                  :class="messaging.currentConversation.value.status === 'open' ? 'bg-green-500' : 'bg-gray-400'"
                />
              </div>
              <div>
                <h2 class="font-semibold text-gray-900 dark:text-white">
                  {{ messaging.currentConversation.value.user?.name || 'User' }}
                </h2>
                <div class="flex items-center gap-2 text-sm">
                  <span
                    class="flex items-center gap-1"
                    :class="messaging.currentConversation.value.status === 'open' ? 'text-green-600' : 'text-gray-500'"
                  >
                    <span class="w-2 h-2 rounded-full" :class="messaging.currentConversation.value.status === 'open' ? 'bg-green-500' : 'bg-gray-400'"></span>
                    {{ messaging.currentConversation.value.status === 'open' ? 'Active' : messaging.currentConversation.value.status }}
                  </span>
                  <span v-if="messaging.currentConversation.value.subject" class="text-gray-400">
                    &bull; {{ messaging.currentConversation.value.subject }}
                  </span>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-2">
              <!-- Status actions -->
              <UButton
                v-if="messaging.currentConversation.value.status === 'open'"
                size="sm"
                variant="soft"
                color="neutral"
                icon="i-heroicons-check-circle"
                @click="handleUpdateStatus('closed')"
              >
                Close
              </UButton>
              <UButton
                v-if="messaging.currentConversation.value.status === 'closed'"
                size="sm"
                variant="soft"
                color="success"
                icon="i-heroicons-arrow-path"
                @click="handleUpdateStatus('open')"
              >
                Reopen
              </UButton>
              <UButton
                v-if="messaging.currentConversation.value.status !== 'archived'"
                size="sm"
                variant="ghost"
                color="warning"
                icon="i-heroicons-archive-box"
                @click="handleUpdateStatus('archived')"
              />
              <UButton
                icon="i-heroicons-x-mark"
                variant="ghost"
                size="sm"
                @click="closeConversation"
              />
            </div>
          </div>
        </div>

        <!-- Messages area -->
        <MessengerChatArea
          :conversation="messaging.currentConversation.value"
          :messages="messaging.messages.value"
          :loading="messaging.loading.value"
          :typing-users="messaging.typingUsers.value"
          :quick-replies="quickReplies"
          :show-quick-replies="showQuickReplies"
          class="flex-1"
          @send="handleSendMessage"
          @typing="handleTyping"
          @mark-as-read="handleMarkAsRead"
          @toggle-quick-replies="showQuickReplies = !showQuickReplies"
          @insert-quick-reply="insertQuickReply"
        />
      </template>
    </div>
  </div>
</template>

<script lang="ts">
function formatTimeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return 'now'
  if (minutes < 60) return `${minutes}m`
  if (hours < 24) return `${hours}h`
  if (days < 7) return `${days}d`
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
}
</script>
