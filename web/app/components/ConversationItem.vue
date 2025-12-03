<script setup lang="ts">
import type { Conversation } from '~/types'

const props = defineProps<{
  conversation: Conversation
  active?: boolean
}>()

defineEmits<{
  click: []
}>()

const timeAgo = computed(() => {
  const date = props.conversation.last_message_at
    ? new Date(props.conversation.last_message_at)
    : new Date(props.conversation.created_at)

  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m`
  if (hours < 24) return `${hours}h`
  if (days < 7) return `${days}d`
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
})

const statusColor = computed(() => {
  switch (props.conversation.status) {
    case 'open':
      return 'bg-green-500'
    case 'closed':
      return 'bg-gray-400'
    case 'archived':
      return 'bg-yellow-500'
    default:
      return 'bg-gray-400'
  }
})

const displayName = computed(() => {
  return props.conversation.user?.name || 'User'
})

const lastMessagePreview = computed(() => {
  const msg = props.conversation.last_message
  if (!msg) return 'No messages yet'
  const content = msg.content
  return content.length > 50 ? content.substring(0, 50) + '...' : content
})
</script>

<template>
  <div
    class="flex items-start gap-3 p-3 rounded-lg cursor-pointer transition-colors"
    :class="[
      active
        ? 'bg-primary-50 dark:bg-primary-900/20'
        : 'hover:bg-gray-50 dark:hover:bg-gray-800'
    ]"
    @click="$emit('click')"
  >
    <!-- Avatar -->
    <div class="relative flex-shrink-0">
      <UAvatar
        :src="conversation.user?.avatar"
        :alt="displayName"
        size="md"
      />
      <span
        :class="statusColor"
        class="absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white dark:border-gray-900"
      />
    </div>

    <!-- Content -->
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between gap-2">
        <h4
          class="font-medium text-gray-900 dark:text-white truncate"
          :class="{ 'font-semibold': (conversation.unread_count || 0) > 0 }"
        >
          {{ displayName }}
        </h4>
        <span class="text-xs text-gray-500 flex-shrink-0">{{ timeAgo }}</span>
      </div>

      <p
        v-if="conversation.subject"
        class="text-sm text-gray-600 dark:text-gray-300 truncate"
      >
        {{ conversation.subject }}
      </p>

      <p
        class="text-sm truncate"
        :class="[
          (conversation.unread_count || 0) > 0
            ? 'text-gray-700 dark:text-gray-200 font-medium'
            : 'text-gray-500 dark:text-gray-400'
        ]"
      >
        {{ lastMessagePreview }}
      </p>
    </div>

    <!-- Unread badge -->
    <UBadge
      v-if="(conversation.unread_count || 0) > 0"
      color="primary"
      size="xs"
      class="flex-shrink-0"
    >
      {{ conversation.unread_count }}
    </UBadge>
  </div>
</template>
