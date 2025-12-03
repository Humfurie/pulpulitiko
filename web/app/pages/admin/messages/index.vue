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
const page = ref(1)
const perPage = 20

// Load conversations
const { data: conversationsData, pending, refresh } = await useAsyncData(
  'admin-conversations',
  () => messaging.fetchAdminConversations(page.value, perPage, statusFilter.value || undefined),
  { watch: [page, statusFilter] }
)

const conversations = computed(() => conversationsData.value?.conversations || [])
const totalPages = computed(() => conversationsData.value?.total_pages || 1)

// Load selected conversation from route
watch(
  () => route.query.id,
  async (id) => {
    if (id && typeof id === 'string') {
      selectedConversationId.value = id
      await messaging.fetchConversation(id)
      await messaging.fetchMessages(id)
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

const statusOptions = [
  { value: '', label: 'All' },
  { value: 'open', label: 'Open' },
  { value: 'closed', label: 'Closed' },
  { value: 'archived', label: 'Archived' }
]
</script>

<template>
  <div class="h-[calc(100vh-8rem)]">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Messages</h1>
        <p class="text-gray-500">Manage user support conversations</p>
      </div>

      <div class="flex items-center gap-3">
        <USelectMenu
          v-model="statusFilter"
          :options="statusOptions"
          value-attribute="value"
          option-attribute="label"
          placeholder="Filter by status"
          class="w-40"
        />
        <UButton
          icon="i-heroicons-arrow-path"
          variant="ghost"
          :loading="pending"
          @click="refresh()"
        />
      </div>
    </div>

    <div class="flex h-full border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden bg-white dark:bg-gray-900">
      <!-- Conversations List -->
      <div class="w-80 border-r border-gray-200 dark:border-gray-700 flex flex-col">
        <div class="p-3 border-b border-gray-200 dark:border-gray-700">
          <h2 class="font-medium text-gray-900 dark:text-white">
            Conversations
            <span class="text-gray-500 text-sm">({{ conversationsData?.total || 0 }})</span>
          </h2>
        </div>

        <div class="flex-1 overflow-y-auto">
          <div v-if="pending" class="flex justify-center py-8">
            <UIcon name="i-heroicons-arrow-path" class="w-6 h-6 animate-spin text-gray-400" />
          </div>

          <div
            v-else-if="conversations.length === 0"
            class="flex flex-col items-center justify-center h-full p-4 text-center"
          >
            <UIcon name="i-heroicons-inbox" class="w-12 h-12 text-gray-400 mb-2" />
            <p class="text-gray-500">No conversations found</p>
          </div>

          <div v-else class="divide-y divide-gray-100 dark:divide-gray-800">
            <ConversationItem
              v-for="conv in conversations"
              :key="conv.id"
              :conversation="conv"
              :active="selectedConversationId === conv.id"
              @click="selectConversation(conv)"
            />
          </div>
        </div>

        <!-- Pagination -->
        <div
          v-if="totalPages > 1"
          class="p-3 border-t border-gray-200 dark:border-gray-700"
        >
          <div class="flex items-center justify-between">
            <UButton
              icon="i-heroicons-chevron-left"
              variant="ghost"
              size="sm"
              :disabled="page === 1"
              @click="page--"
            />
            <span class="text-sm text-gray-500">
              Page {{ page }} of {{ totalPages }}
            </span>
            <UButton
              icon="i-heroicons-chevron-right"
              variant="ghost"
              size="sm"
              :disabled="page === totalPages"
              @click="page++"
            />
          </div>
        </div>
      </div>

      <!-- Chat Window -->
      <div class="flex-1 flex flex-col">
        <!-- Status actions bar (when conversation selected) -->
        <div
          v-if="messaging.currentConversation.value"
          class="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800"
        >
          <div class="flex items-center gap-2">
            <span class="text-sm text-gray-500">Status:</span>
            <UBadge
              :color="messaging.currentConversation.value.status === 'open' ? 'success' : 'neutral'"
            >
              {{ messaging.currentConversation.value.status }}
            </UBadge>
          </div>

          <div class="flex items-center gap-2">
            <UButton
              v-if="messaging.currentConversation.value.status === 'open'"
              size="xs"
              variant="soft"
              color="neutral"
              @click="handleUpdateStatus('closed')"
            >
              Close
            </UButton>
            <UButton
              v-if="messaging.currentConversation.value.status === 'closed'"
              size="xs"
              variant="soft"
              color="success"
              @click="handleUpdateStatus('open')"
            >
              Reopen
            </UButton>
            <UButton
              v-if="messaging.currentConversation.value.status !== 'archived'"
              size="xs"
              variant="soft"
              color="warning"
              @click="handleUpdateStatus('archived')"
            >
              Archive
            </UButton>
          </div>
        </div>

        <ChatWindow
          :conversation="messaging.currentConversation.value"
          :messages="messaging.messages.value"
          :loading="messaging.loading.value"
          :typing-users="messaging.typingUsers.value"
          class="flex-1"
          @send="handleSendMessage"
          @typing="handleTyping"
          @mark-as-read="handleMarkAsRead"
          @close="closeConversation"
        />
      </div>
    </div>
  </div>
</template>
