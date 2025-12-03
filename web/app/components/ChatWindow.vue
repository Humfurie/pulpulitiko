<script setup lang="ts">
import type { Conversation, Message } from '~/types'

const props = defineProps<{
  conversation: Conversation | null
  messages: Message[]
  loading?: boolean
  typingUsers?: string[]
}>()

const emit = defineEmits<{
  send: [content: string]
  typing: []
  close: []
  markAsRead: []
}>()

const auth = useAuth()
const messageInput = ref('')
const messagesContainer = ref<HTMLElement | null>(null)

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
  return message.sender_id === auth.user.value?.id
}

// Group messages by date
const groupedMessages = computed(() => {
  const groups: { date: string; messages: Message[] }[] = []
  let currentDate = ''

  for (const message of props.messages) {
    const messageDate = new Date(message.created_at).toDateString()
    if (messageDate !== currentDate) {
      currentDate = messageDate
      groups.push({
        date: formatDateHeader(message.created_at),
        messages: [message]
      })
    } else {
      const lastGroup = groups[groups.length - 1]
      if (lastGroup) {
        lastGroup.messages.push(message)
      }
    }
  }

  return groups
})

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
</script>

<template>
  <div class="flex flex-col h-full bg-white dark:bg-gray-900">
    <!-- Header -->
    <div
      v-if="conversation"
      class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-3">
        <UAvatar
          :src="conversation.user?.avatar"
          :alt="conversation.user?.name || 'User'"
          size="sm"
        />
        <div>
          <h3 class="font-medium text-gray-900 dark:text-white">
            {{ conversation.user?.name || 'User' }}
          </h3>
          <p v-if="conversation.subject" class="text-xs text-gray-500">
            {{ conversation.subject }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <UBadge
          :color="conversation.status === 'open' ? 'success' : 'neutral'"
          variant="subtle"
          size="xs"
        >
          {{ conversation.status }}
        </UBadge>
        <UButton
          icon="i-heroicons-x-mark"
          variant="ghost"
          size="sm"
          @click="$emit('close')"
        />
      </div>
    </div>

    <!-- Empty state -->
    <div
      v-if="!conversation"
      class="flex-1 flex items-center justify-center text-gray-500"
    >
      <div class="text-center">
        <UIcon name="i-heroicons-chat-bubble-left-right" class="w-12 h-12 mx-auto mb-2" />
        <p>Select a conversation to start messaging</p>
      </div>
    </div>

    <!-- Messages -->
    <div
      v-else
      ref="messagesContainer"
      class="flex-1 overflow-y-auto p-4 space-y-4"
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
        >
          <!-- Date separator -->
          <div class="flex items-center justify-center my-4">
            <span class="px-3 py-1 text-xs text-gray-500 bg-gray-100 dark:bg-gray-800 rounded-full">
              {{ group.date }}
            </span>
          </div>

          <!-- Messages for this date -->
          <div class="space-y-3">
            <ChatBubble
              v-for="message in group.messages"
              :key="message.id"
              :message="message"
              :is-own="isOwnMessage(message)"
            />
          </div>
        </div>
      </template>

      <!-- Typing indicator -->
      <div
        v-if="typingUsers && typingUsers.length > 0"
        class="flex items-center gap-2 text-gray-500"
      >
        <div class="flex gap-1">
          <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0ms" />
          <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 150ms" />
          <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 300ms" />
        </div>
        <span class="text-xs">typing...</span>
      </div>
    </div>

    <!-- Input -->
    <div
      v-if="conversation"
      class="p-4 border-t border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-end gap-2">
        <UTextarea
          v-model="messageInput"
          placeholder="Type a message..."
          :rows="1"
          autoresize
          :maxrows="4"
          class="flex-1"
          @keydown="handleKeydown"
        />
        <UButton
          icon="i-heroicons-paper-airplane"
          color="primary"
          :disabled="!messageInput.trim()"
          @click="handleSend"
        />
      </div>
    </div>
  </div>
</template>
