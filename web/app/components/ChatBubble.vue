<script setup lang="ts">
import type { Message } from '~/types'

const props = defineProps<{
  message: Message
  isOwn: boolean
}>()

const formattedTime = computed(() => {
  const date = new Date(props.message.created_at)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
})

const _formattedDate = computed(() => {
  const date = new Date(props.message.created_at)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  if (date.toDateString() === today.toDateString()) {
    return 'Today'
  } else if (date.toDateString() === yesterday.toDateString()) {
    return 'Yesterday'
  } else {
    return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
  }
})
</script>

<template>
  <div
    class="flex"
    :class="isOwn ? 'justify-end' : 'justify-start'"
  >
    <div class="max-w-[75%]">
      <!-- Sender name for non-own messages -->
      <p
        v-if="!isOwn && message.sender"
        class="text-xs text-gray-500 dark:text-gray-400 mb-1 ml-3"
      >
        {{ message.sender.name }}
      </p>

      <div
        class="rounded-2xl px-4 py-2"
        :class="[
          isOwn
            ? 'bg-primary-500 text-white rounded-br-md'
            : 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-bl-md'
        ]"
      >
        <p class="text-sm whitespace-pre-wrap break-words">{{ message.content }}</p>
      </div>

      <!-- Time and read status -->
      <div
        class="flex items-center gap-1 mt-1"
        :class="isOwn ? 'justify-end mr-2' : 'ml-3'"
      >
        <span class="text-xs text-gray-400">{{ formattedTime }}</span>
        <span v-if="isOwn" class="text-xs">
          <UIcon
            v-if="message.is_read"
            name="i-heroicons-check-circle"
            class="w-3 h-3 text-primary-500"
          />
          <UIcon
            v-else
            name="i-heroicons-check"
            class="w-3 h-3 text-gray-400"
          />
        </span>
      </div>
    </div>
  </div>
</template>
