<script setup lang="ts">
import type { Conversation } from '~/types'

const auth = useAuth()
const messaging = useMessaging()

const isOpen = ref(false)
const showNewChat = ref(false)
const newMessage = ref('')
const newSubject = ref('')

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
    await messaging.fetchMessages(conv.id)
    messaging.currentConversation.value = conv
  }
})

async function handleSelectConversation(conv: Conversation) {
  showNewChat.value = false
  messaging.currentConversation.value = conv
}

function isActiveConversation(convId: string): boolean {
  return selectedConversation.value?.id === convId
}

async function handleStartNewChat() {
  if (!newMessage.value.trim()) return

  try {
    await messaging.createConversation({
      subject: newSubject.value || undefined,
      message: newMessage.value.trim()
    })
    showNewChat.value = false
    newMessage.value = ''
    newSubject.value = ''
  } catch (error) {
    console.error('Failed to start chat:', error)
  }
}

async function handleSendMessage(content: string) {
  if (!messaging.currentConversation.value) return
  await messaging.sendMessage(messaging.currentConversation.value.id, { content })
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
</script>

<template>
  <div class="fixed bottom-4 right-4 z-50">
    <!-- Chat Button -->
    <UButton
      v-if="!isOpen && auth.isAuthenticated.value"
      icon="i-heroicons-chat-bubble-left-right"
      color="primary"
      size="xl"
      class="rounded-full shadow-lg w-14 h-14"
      @click="isOpen = true"
    >
      <!-- Unread badge -->
      <span
        v-if="messaging.unreadCounts.value.total > 0"
        class="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center"
      >
        {{ messaging.unreadCounts.value.total > 9 ? '9+' : messaging.unreadCounts.value.total }}
      </span>
    </UButton>

    <!-- Chat Window -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0 scale-95 translate-y-4"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100 scale-100 translate-y-0"
      leave-to-class="opacity-0 scale-95 translate-y-4"
    >
      <div
        v-if="isOpen && auth.isAuthenticated.value"
        class="absolute bottom-0 right-0 w-96 h-[500px] bg-white dark:bg-gray-900 rounded-lg shadow-2xl border border-gray-200 dark:border-gray-700 flex flex-col overflow-hidden"
      >
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 bg-primary-500 text-white">
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-chat-bubble-left-right" class="w-5 h-5" />
            <span class="font-medium">Support Chat</span>
          </div>
          <div class="flex items-center gap-1">
            <UButton
              icon="i-heroicons-plus"
              variant="ghost"
              color="neutral"
              size="xs"
              class="text-white hover:bg-white/20"
              @click="showNewChat = true; messaging.currentConversation.value = null"
            />
            <UButton
              icon="i-heroicons-x-mark"
              variant="ghost"
              color="neutral"
              size="xs"
              class="text-white hover:bg-white/20"
              @click="isOpen = false"
            />
          </div>
        </div>

        <!-- New Chat Form -->
        <div
          v-if="showNewChat"
          class="flex-1 flex flex-col p-4"
        >
          <h3 class="font-medium text-gray-900 dark:text-white mb-4">Start a new conversation</h3>

          <UFormGroup label="Subject (optional)" class="mb-3">
            <UInput
              v-model="newSubject"
              placeholder="What's this about?"
            />
          </UFormGroup>

          <UFormGroup label="Message" class="mb-4">
            <UTextarea
              v-model="newMessage"
              placeholder="How can we help you?"
              :rows="4"
            />
          </UFormGroup>

          <div class="flex gap-2 mt-auto">
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
              Send
            </UButton>
          </div>
        </div>

        <!-- Conversations List & Chat -->
        <template v-else>
          <!-- Conversations sidebar (shown when no conversation selected) -->
          <div
            v-if="!selectedConversation"
            class="flex-1 overflow-y-auto"
          >
            <div
              v-if="conversations.length === 0"
              class="flex flex-col items-center justify-center h-full p-4 text-center"
            >
              <UIcon name="i-heroicons-inbox" class="w-12 h-12 text-gray-400 mb-2" />
              <p class="text-gray-500 mb-4">No conversations yet</p>
              <UButton
                color="primary"
                @click="showNewChat = true"
              >
                Start a conversation
              </UButton>
            </div>

            <div v-else class="divide-y divide-gray-100 dark:divide-gray-800">
              <ConversationItem
                v-for="conv in conversations"
                :key="conv.id"
                :conversation="conv"
                :active="isActiveConversation(conv.id)"
                @click="handleSelectConversation(conv)"
              />
            </div>
          </div>

          <!-- Chat View -->
          <template v-else>
            <!-- Back button -->
            <div class="px-3 py-2 border-b border-gray-200 dark:border-gray-700">
              <UButton
                icon="i-heroicons-arrow-left"
                variant="ghost"
                size="xs"
                @click="messaging.currentConversation.value = null"
              >
                Back to conversations
              </UButton>
            </div>

            <ChatWindow
              :conversation="selectedConversation"
              :messages="messaging.messages.value"
              :loading="messaging.loading.value"
              :typing-users="messaging.typingUsers.value"
              class="flex-1"
              @send="handleSendMessage"
              @typing="handleTyping"
              @mark-as-read="handleMarkAsRead"
              @close="messaging.currentConversation.value = null"
            />
          </template>
        </template>
      </div>
    </Transition>

    <!-- Login prompt -->
    <UButton
      v-if="!auth.isAuthenticated.value"
      icon="i-heroicons-chat-bubble-left-right"
      color="primary"
      size="xl"
      class="rounded-full shadow-lg w-14 h-14"
      @click="navigateTo('/login')"
    />
  </div>
</template>
