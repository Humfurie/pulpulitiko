<script setup lang="ts">
import type { Notification } from '~/types'

definePageMeta({
  middleware: 'auth'
})

const api = useApi()
const auth = useAuth()

const notifications = ref<Notification[]>([])
const loading = ref(true)
const error = ref('')
const page = ref(1)
const totalPages = ref(1)
const unreadCount = ref(0)
const loadingMore = ref(false)

// Load notifications
async function loadNotifications(pageNum = 1, append = false) {
  if (pageNum === 1) {
    loading.value = true
  } else {
    loadingMore.value = true
  }
  error.value = ''

  try {
    const result = await api.getNotifications(auth.getAuthHeaders(), pageNum, 20)
    if (append) {
      notifications.value = [...notifications.value, ...(result.notifications || [])]
    } else {
      notifications.value = result.notifications || []
    }
    page.value = pageNum
    totalPages.value = result.total_pages
    unreadCount.value = result.unread_count
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load notifications'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

// Load more
async function loadMore() {
  if (loadingMore.value || page.value >= totalPages.value) return
  await loadNotifications(page.value + 1, true)
}

// Mark as read
async function markAsRead(notification: Notification) {
  if (notification.is_read) return
  try {
    await api.markNotificationAsRead(notification.id, auth.getAuthHeaders())
    notification.is_read = true
    unreadCount.value = Math.max(0, unreadCount.value - 1)
  } catch (e) {
    console.error('Failed to mark as read:', e)
  }
}

// Mark all as read
async function markAllAsRead() {
  try {
    await api.markAllNotificationsAsRead(auth.getAuthHeaders())
    notifications.value.forEach(n => n.is_read = true)
    unreadCount.value = 0
  } catch (e) {
    console.error('Failed to mark all as read:', e)
  }
}

// Delete notification
async function deleteNotification(id: string) {
  try {
    await api.deleteNotification(id, auth.getAuthHeaders())
    notifications.value = notifications.value.filter(n => n.id !== id)
  } catch (e) {
    console.error('Failed to delete notification:', e)
  }
}

// Get notification link
function getNotificationLink(notification: Notification): string {
  if (notification.politician) {
    return `/politician/${notification.politician.slug}`
  }
  if (notification.article) {
    return `/article/${notification.article.slug}`
  }
  return '/'
}

// Format time ago
function timeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000)

  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 7) return `${days}d ago`
  return date.toLocaleDateString()
}

// Load on mount
onMounted(() => {
  loadNotifications()
})

useHead({
  title: 'Notifications - Pulpulitiko'
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950">
    <UContainer class="py-8">
      <div class="max-w-2xl mx-auto">
        <!-- Header -->
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Notifications</h1>
            <p v-if="unreadCount > 0" class="text-sm text-gray-500 dark:text-gray-400 mt-1">
              {{ unreadCount }} unread
            </p>
          </div>
          <button
            v-if="unreadCount > 0"
            class="text-sm text-primary hover:text-primary/80 font-medium"
            @click="markAllAsRead"
          >
            Mark all as read
          </button>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="py-12 text-center">
          <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 text-gray-400 mx-auto animate-spin mb-2" />
          <p class="text-gray-500 dark:text-gray-400">Loading notifications...</p>
        </div>

        <!-- Error -->
        <UAlert
          v-else-if="error"
          color="error"
          icon="i-heroicons-exclamation-triangle"
          :description="error"
          class="mb-4"
        />

        <!-- Notifications list -->
        <div v-else-if="notifications.length > 0" class="space-y-2">
          <div
            v-for="notification in notifications"
            :key="notification.id"
            class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden"
            :class="{ 'bg-orange-50/50 dark:bg-orange-950/20 border-orange-200 dark:border-orange-800/50': !notification.is_read }"
          >
            <NuxtLink
              :to="getNotificationLink(notification)"
              class="block p-4 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              @click="markAsRead(notification)"
            >
              <div class="flex gap-4">
                <!-- Avatar -->
                <div class="flex-shrink-0">
                  <img
                    v-if="notification.actor?.avatar"
                    :src="notification.actor.avatar"
                    :alt="notification.actor.name"
                    class="w-10 h-10 rounded-full object-cover"
                  >
                  <div v-else class="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                    <UIcon name="i-heroicons-bell" class="w-5 h-5 text-primary" />
                  </div>
                </div>

                <!-- Content -->
                <div class="flex-1 min-w-0">
                  <p class="font-medium text-gray-900 dark:text-white">
                    {{ notification.title }}
                  </p>
                  <p v-if="notification.message" class="text-sm text-gray-600 dark:text-gray-400 mt-1">
                    {{ notification.message }}
                  </p>
                  <div class="flex items-center gap-3 mt-2 text-xs text-gray-500 dark:text-gray-400">
                    <span>{{ timeAgo(notification.created_at) }}</span>
                    <span v-if="notification.politician" class="text-primary">
                      {{ notification.politician.name }}
                    </span>
                    <span v-else-if="notification.article" class="text-primary">
                      {{ notification.article.name }}
                    </span>
                  </div>
                </div>

                <!-- Unread indicator -->
                <div class="flex-shrink-0 flex items-start gap-2">
                  <span
                    v-if="!notification.is_read"
                    class="w-2 h-2 bg-orange-500 rounded-full mt-2"
                  />
                </div>
              </div>
            </NuxtLink>

            <!-- Actions -->
            <div class="px-4 pb-3 flex justify-end gap-2">
              <button
                v-if="!notification.is_read"
                class="text-xs text-gray-500 hover:text-primary"
                @click="markAsRead(notification)"
              >
                Mark as read
              </button>
              <button
                class="text-xs text-gray-500 hover:text-red-500"
                @click="deleteNotification(notification.id)"
              >
                Delete
              </button>
            </div>
          </div>

          <!-- Load more -->
          <div v-if="page < totalPages" class="pt-4 text-center">
            <button
              class="text-primary hover:text-primary/80 font-medium disabled:opacity-50"
              :disabled="loadingMore"
              @click="loadMore"
            >
              <template v-if="loadingMore">
                <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin inline mr-2" />
                Loading...
              </template>
              <template v-else>
                Load more
              </template>
            </button>
          </div>
        </div>

        <!-- Empty state -->
        <div v-else class="py-16 text-center">
          <UIcon name="i-heroicons-bell-slash" class="w-16 h-16 text-gray-300 dark:text-gray-700 mx-auto mb-4" />
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">No notifications</h2>
          <p class="text-gray-500 dark:text-gray-400">
            You're all caught up! Check back later for new notifications.
          </p>
        </div>
      </div>
    </UContainer>
  </div>
</template>
