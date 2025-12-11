<script setup lang="ts">
import type { Category, Notification } from '~/types'

const api = useApi()
const auth = useAuth()
const route = useRoute()
const searchQuery = ref('')
const mobileMenuOpen = ref(false)
const isScrolled = ref(false)
const searchFocused = ref(false)

// Notification state
const unreadCount = ref(0)
const notifications = ref<Notification[]>([])
const notificationDropdownOpen = ref(false)
const loadingNotifications = ref(false)

const { data: categories } = await useAsyncData('categories', () => api.getCategories())

// Active state for navigation
const isHomeActive = computed(() => route.path === '/')
const isCategoryActive = computed(() => route.path.startsWith('/category'))
const isPoliticianActive = computed(() => route.path.startsWith('/politician'))

// Check auth on mount
onMounted(async () => {
  if (auth.token.value && !auth.user.value) {
    await auth.fetchCurrentUser()
  }
})

// Fetch unread notification count when authenticated
async function fetchUnreadCount() {
  if (!auth.isAuthenticated.value || !auth.token.value) return
  try {
    const result = await api.getUnreadNotificationCount({ Authorization: `Bearer ${auth.token.value}` })
    unreadCount.value = result.count
  } catch {
    // Silently fail
  }
}

// Fetch recent notifications
async function fetchNotifications() {
  if (!auth.isAuthenticated.value || !auth.token.value) return
  loadingNotifications.value = true
  try {
    const result = await api.getNotifications({ Authorization: `Bearer ${auth.token.value}` }, 1, 5)
    notifications.value = result.notifications || []
    unreadCount.value = result.unread_count
  } catch {
    // Silently fail
  } finally {
    loadingNotifications.value = false
  }
}

// Mark notification as read
async function markAsRead(notification: Notification) {
  if (!auth.token.value || notification.is_read) return
  try {
    await api.markNotificationAsRead(notification.id, { Authorization: `Bearer ${auth.token.value}` })
    notification.is_read = true
    unreadCount.value = Math.max(0, unreadCount.value - 1)
  } catch {
    // Silently fail
  }
}

// Mark all as read
async function markAllAsRead() {
  if (!auth.token.value) return
  try {
    await api.markAllNotificationsAsRead({ Authorization: `Bearer ${auth.token.value}` })
    notifications.value.forEach(n => n.is_read = true)
    unreadCount.value = 0
  } catch {
    // Silently fail
  }
}

// Get notification link
function getNotificationLink(notification: Notification): string {
  if (notification.politician) {
    return `/politician/${notification.politician.slug}`
  }
  if (notification.article) {
    return `/${notification.article.slug}`
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

// Fetch notifications when dropdown opens
watch(notificationDropdownOpen, (isOpen) => {
  if (isOpen) {
    fetchNotifications()
  }
})

// Initial fetch of unread count
watch(() => auth.isAuthenticated.value, (isAuth) => {
  if (isAuth) {
    fetchUnreadCount()
  } else {
    unreadCount.value = 0
    notifications.value = []
  }
}, { immediate: true })

// Poll for new notifications every 10 seconds for near real-time updates
let notificationInterval: ReturnType<typeof setInterval> | null = null
onMounted(() => {
  if (auth.isAuthenticated.value) {
    fetchUnreadCount()
  }
  notificationInterval = setInterval(() => {
    if (auth.isAuthenticated.value) {
      fetchUnreadCount()
    }
  }, 10000) // Check every 10 seconds
})

onUnmounted(() => {
  if (notificationInterval) {
    clearInterval(notificationInterval)
  }
})

function handleSearch() {
  if (searchQuery.value.trim()) {
    navigateTo({ path: '/search', query: { q: searchQuery.value.trim() } })
    searchQuery.value = ''
    mobileMenuOpen.value = false
  }
}

// Track scroll position for header styling
onMounted(() => {
  const handleScroll = () => {
    isScrolled.value = window.scrollY > 20
  }
  window.addEventListener('scroll', handleScroll, { passive: true })
  onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll)
  })
})
</script>

<template>
  <header
    class="sticky top-0 z-50 transition-all duration-300"
    :class="[
      isScrolled
        ? 'bg-white/80 dark:bg-stone-900/80 backdrop-blur-lg shadow-sm shadow-stone-200/50 dark:shadow-stone-900/50'
        : 'bg-white dark:bg-stone-900',
      'border-b border-stone-200/80 dark:border-stone-800/80'
    ]"
  >
    <UContainer>
      <div class="flex items-center justify-between h-16">
        <!-- Logo -->
        <NuxtLink to="/" class="flex items-center gap-2 group">
          <img
            src="/pulpulitiko.png"
            alt="Pulpulitiko"
            class="h-8 w-auto dark:hidden"
          >
          <img
            src="/pulpulitiko_dark.png"
            alt="Pulpulitiko"
            class="h-8 w-auto hidden dark:block"
          >
        </NuxtLink>

        <!-- Desktop Navigation -->
        <nav class="hidden md:flex items-center gap-1">
          <NuxtLink
            to="/"
            class="px-4 py-2 rounded-full transition-all duration-300 font-medium"
            :class="isHomeActive
              ? 'text-orange-500 bg-orange-50 dark:bg-orange-950/30'
              : 'text-stone-600 dark:text-stone-300 hover:text-orange-500 hover:bg-orange-50 dark:hover:bg-orange-950/30'"
          >
            Home
          </NuxtLink>
          <UDropdownMenu
            v-if="categories?.length"
            :items="categories.map((cat: Category) => ({
              label: cat.name,
              to: `/category/${cat.slug}`
            }))"
          >
            <UButton
              color="neutral"
              variant="ghost"
              trailing-icon="i-heroicons-chevron-down"
              class="rounded-full transition-colors duration-300 font-medium px-4 py-2"
              :class="isCategoryActive
                ? 'text-orange-500 bg-orange-50 dark:bg-orange-950/30'
                : 'text-stone-600 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500'"
              :ui="{ trailingIcon: 'size-4' }"
            >
              Categories
            </UButton>
          </UDropdownMenu>
          <NuxtLink
            to="/politicians"
            class="px-4 py-2 rounded-full transition-all duration-300 font-medium"
            :class="isPoliticianActive
              ? 'text-orange-500 bg-orange-50 dark:bg-orange-950/30'
              : 'text-stone-600 dark:text-stone-300 hover:text-orange-500 hover:bg-orange-50 dark:hover:bg-orange-950/30'"
          >
            Politicians
          </NuxtLink>
          <UDropdownMenu
            :items="[
              { label: 'Polls', to: '/polls', icon: 'i-heroicons-chart-bar' },
              { label: 'Elections', to: '/elections', icon: 'i-heroicons-clipboard-document-check' },
              { label: 'Voter Education', to: '/voter-education', icon: 'i-heroicons-academic-cap' },
              { label: 'Legislation Tracker', to: '/legislation', icon: 'i-heroicons-document-text' },
              { label: 'Political Parties', to: '/parties', icon: 'i-heroicons-flag' },
              { label: 'Find My Representatives', to: '/my-representatives', icon: 'i-heroicons-user-group' },
              { label: 'Philippine Locations', to: '/locations', icon: 'i-heroicons-map-pin' }
            ]"
          >
            <UButton
              color="neutral"
              variant="ghost"
              trailing-icon="i-heroicons-chevron-down"
              class="rounded-full transition-colors duration-300 font-medium px-4 py-2 text-stone-600 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500"
              :ui="{ trailingIcon: 'size-4' }"
            >
              More
            </UButton>
          </UDropdownMenu>
        </nav>

        <!-- Search & Actions -->
        <div class="hidden md:flex items-center gap-2">
          <form class="relative" @submit.prevent="handleSearch">
            <div
              class="relative transition-all duration-300"
              :class="searchFocused ? 'w-72' : 'w-64'"
            >
              <UInput
                v-model="searchQuery"
                placeholder="Search articles..."
                icon="i-heroicons-magnifying-glass"
                class="rounded-full bg-stone-100 dark:bg-stone-800 border-transparent focus:border-orange-500 focus:ring-orange-500/20 transition-all duration-300"
                @focus="searchFocused = true"
                @blur="searchFocused = false"
              />
            </div>
          </form>
          <UColorModeButton class="rounded-full hover:bg-orange-50 dark:hover:bg-orange-950/30 transition-colors duration-300" />

          <!-- Notification Bell (logged in) -->
          <UPopover v-if="auth.isAuthenticated.value" v-model:open="notificationDropdownOpen">
            <UButton
              variant="ghost"
              class="rounded-full hover:bg-orange-50 dark:hover:bg-orange-950/30 transition-colors duration-300 relative"
            >
              <UIcon name="i-heroicons-bell" class="w-5 h-5" />
              <span
                v-if="unreadCount > 0"
                class="absolute -top-1 -right-1 bg-red-500 text-white text-xs font-bold rounded-full min-w-[18px] h-[18px] flex items-center justify-center px-1"
              >
                {{ unreadCount > 99 ? '99+' : unreadCount }}
              </span>
            </UButton>

            <template #content>
              <div class="w-[calc(100vw-2rem)] sm:w-80 max-h-96 overflow-hidden">
                <!-- Header -->
                <div class="flex items-center justify-between px-4 py-3 border-b border-stone-200 dark:border-stone-700">
                  <h3 class="font-semibold text-stone-800 dark:text-stone-200">Notifications</h3>
                  <button
                    v-if="unreadCount > 0"
                    class="text-xs text-orange-500 hover:text-orange-600 font-medium"
                    @click="markAllAsRead"
                  >
                    Mark all as read
                  </button>
                </div>

                <!-- Loading -->
                <div v-if="loadingNotifications" class="py-8 flex justify-center">
                  <UIcon name="i-heroicons-arrow-path" class="w-6 h-6 animate-spin text-stone-400" />
                </div>

                <!-- Notifications List -->
                <div v-else-if="notifications.length > 0" class="overflow-y-auto max-h-72">
                  <NuxtLink
                    v-for="notification in notifications"
                    :key="notification.id"
                    :to="getNotificationLink(notification)"
                    class="block px-4 py-3 hover:bg-stone-50 dark:hover:bg-stone-800 transition-colors border-b border-stone-100 dark:border-stone-800 last:border-b-0"
                    :class="{ 'bg-orange-50/50 dark:bg-orange-950/20': !notification.is_read }"
                    @click="markAsRead(notification); notificationDropdownOpen = false"
                  >
                    <div class="flex gap-3">
                      <UAvatar
                        v-if="notification.actor"
                        :src="notification.actor.avatar"
                        :alt="notification.actor.name"
                        size="sm"
                      />
                      <UIcon
                        v-else
                        name="i-heroicons-bell"
                        class="w-8 h-8 text-stone-400"
                      />
                      <div class="flex-1 min-w-0">
                        <p class="text-sm font-medium text-stone-800 dark:text-stone-200">
                          {{ notification.title }}
                        </p>
                        <p v-if="notification.message" class="text-xs text-stone-500 dark:text-stone-400 mt-0.5 line-clamp-2">
                          {{ notification.message }}
                        </p>
                        <p class="text-xs text-stone-400 dark:text-stone-500 mt-1">
                          {{ timeAgo(notification.created_at) }}
                          <span v-if="notification.politician" class="text-orange-500">• {{ notification.politician.name }}</span>
                          <span v-else-if="notification.article" class="text-orange-500">• {{ notification.article.name }}</span>
                        </p>
                      </div>
                      <span
                        v-if="!notification.is_read"
                        class="w-2 h-2 bg-orange-500 rounded-full flex-shrink-0 mt-2"
                      />
                    </div>
                  </NuxtLink>
                </div>

                <!-- Empty State -->
                <div v-else class="py-8 px-4 text-center">
                  <UIcon name="i-heroicons-bell-slash" class="w-10 h-10 mx-auto text-stone-300 dark:text-stone-600 mb-2" />
                  <p class="text-sm text-stone-500 dark:text-stone-400">No notifications yet</p>
                </div>

                <!-- Footer -->
                <div v-if="notifications.length > 0" class="px-4 py-3 border-t border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50">
                  <NuxtLink
                    to="/notifications"
                    class="text-sm text-orange-500 hover:text-orange-600 font-medium"
                    @click="notificationDropdownOpen = false"
                  >
                    View all notifications
                  </NuxtLink>
                </div>
              </div>
            </template>
          </UPopover>

          <!-- User Menu (logged in) -->
          <UDropdownMenu
            v-if="auth.isAuthenticated.value"
            :items="[
              { label: 'My Account', to: '/account', icon: 'i-heroicons-user' },
              ...(auth.isAdmin.value || auth.isAuthor.value ? [{ label: 'Admin Panel', to: '/admin', icon: 'i-heroicons-cog-6-tooth' }] : []),
              { type: 'separator' as const },
              { label: 'Logout', icon: 'i-heroicons-arrow-right-on-rectangle', onSelect: () => auth.logout() }
            ]"
          >
            <UButton
              variant="ghost"
              class="rounded-full hover:bg-orange-50 dark:hover:bg-orange-950/30 transition-colors duration-300 gap-2"
            >
              <UAvatar :src="auth.user.value?.avatar" :alt="auth.user.value?.name" size="xs" />
              <span class="text-sm font-medium">{{ auth.user.value?.name }}</span>
              <UIcon name="i-heroicons-chevron-down" class="w-4 h-4" />
            </UButton>
          </UDropdownMenu>
          <!-- Login Button (not logged in) -->
          <NuxtLink v-else to="/login">
            <UButton
              variant="ghost"
              icon="i-heroicons-user-circle"
              class="rounded-full hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-colors duration-300"
            >
              Sign In
            </UButton>
          </NuxtLink>
        </div>

        <!-- Mobile Menu Button -->
        <UButton
          class="md:hidden rounded-full"
          variant="ghost"
          :icon="mobileMenuOpen ? 'i-heroicons-x-mark' : 'i-heroicons-bars-3'"
          @click="mobileMenuOpen = !mobileMenuOpen"
        />
      </div>

      <!-- Mobile Menu -->
      <Transition
        enter-active-class="transition-all duration-300 ease-[cubic-bezier(0.19,1,0.22,1)]"
        leave-active-class="transition-all duration-200 ease-in"
        enter-from-class="opacity-0 -translate-y-4"
        leave-to-class="opacity-0 -translate-y-4"
      >
        <div v-if="mobileMenuOpen" class="md:hidden py-4 border-t border-stone-200 dark:border-stone-800">
          <form class="mb-4" @submit.prevent="handleSearch">
            <UInput
              v-model="searchQuery"
              placeholder="Search articles..."
              icon="i-heroicons-magnifying-glass"
              class="w-full rounded-full bg-stone-100 dark:bg-stone-800"
            />
          </form>
          <nav class="flex flex-col gap-1">
            <NuxtLink
              to="/"
              class="py-3 px-4 rounded-xl transition-all duration-300 font-medium"
              :class="isHomeActive
                ? 'text-orange-500 bg-orange-50 dark:bg-orange-950/30'
                : 'text-stone-700 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500'"
              @click="mobileMenuOpen = false"
            >
              Home
            </NuxtLink>
            <template v-if="categories?.length">
              <p class="py-2 px-4 text-xs font-bold text-stone-400 dark:text-stone-500 uppercase tracking-wider">Categories</p>
              <NuxtLink
                v-for="cat in categories"
                :key="cat.id"
                :to="`/category/${cat.slug}`"
                class="py-3 px-4 rounded-xl transition-all duration-300"
                :class="route.path === `/category/${cat.slug}`
                  ? 'text-orange-500 bg-orange-50 dark:bg-orange-950/30 font-medium'
                  : 'text-stone-600 dark:text-stone-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500'"
                @click="mobileMenuOpen = false"
              >
                {{ cat.name }}
              </NuxtLink>
            </template>
            <NuxtLink
              to="/politicians"
              class="py-3 px-4 rounded-xl transition-all duration-300 font-medium"
              :class="isPoliticianActive
                ? 'text-orange-500 bg-orange-50 dark:bg-orange-950/30'
                : 'text-stone-700 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500'"
              @click="mobileMenuOpen = false"
            >
              Politicians
            </NuxtLink>
            <p class="py-2 px-4 text-xs font-bold text-stone-400 dark:text-stone-500 uppercase tracking-wider mt-2">More</p>
            <NuxtLink
              to="/polls"
              class="py-3 px-4 rounded-xl transition-all duration-300 text-stone-600 dark:text-stone-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 flex items-center gap-2"
              @click="mobileMenuOpen = false"
            >
              <UIcon name="i-heroicons-chart-bar" class="w-5 h-5" />
              Polls
            </NuxtLink>
            <NuxtLink
              to="/elections"
              class="py-3 px-4 rounded-xl transition-all duration-300 text-stone-600 dark:text-stone-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 flex items-center gap-2"
              @click="mobileMenuOpen = false"
            >
              <UIcon name="i-heroicons-clipboard-document-check" class="w-5 h-5" />
              Elections
            </NuxtLink>
            <NuxtLink
              to="/voter-education"
              class="py-3 px-4 rounded-xl transition-all duration-300 text-stone-600 dark:text-stone-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 flex items-center gap-2"
              @click="mobileMenuOpen = false"
            >
              <UIcon name="i-heroicons-academic-cap" class="w-5 h-5" />
              Voter Education
            </NuxtLink>
            <NuxtLink
              to="/legislation"
              class="py-3 px-4 rounded-xl transition-all duration-300 text-stone-600 dark:text-stone-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 flex items-center gap-2"
              @click="mobileMenuOpen = false"
            >
              <UIcon name="i-heroicons-document-text" class="w-5 h-5" />
              Legislation Tracker
            </NuxtLink>
            <NuxtLink
              to="/parties"
              class="py-3 px-4 rounded-xl transition-all duration-300 text-stone-600 dark:text-stone-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 flex items-center gap-2"
              @click="mobileMenuOpen = false"
            >
              <UIcon name="i-heroicons-flag" class="w-5 h-5" />
              Political Parties
            </NuxtLink>
            <NuxtLink
              to="/my-representatives"
              class="py-3 px-4 rounded-xl transition-all duration-300 text-stone-600 dark:text-stone-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 flex items-center gap-2"
              @click="mobileMenuOpen = false"
            >
              <UIcon name="i-heroicons-user-group" class="w-5 h-5" />
              Find My Representatives
            </NuxtLink>
            <NuxtLink
              to="/locations"
              class="py-3 px-4 rounded-xl transition-all duration-300 text-stone-600 dark:text-stone-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 flex items-center gap-2"
              @click="mobileMenuOpen = false"
            >
              <UIcon name="i-heroicons-map-pin" class="w-5 h-5" />
              Philippine Locations
            </NuxtLink>
          </nav>
          <!-- User Section (Mobile) -->
          <div class="mt-4 pt-4 border-t border-stone-200 dark:border-stone-800">
            <template v-if="auth.isAuthenticated.value">
              <!-- User Info -->
              <div class="flex items-center gap-3 px-4 py-2 mb-2">
                <UAvatar :src="auth.user.value?.avatar" :alt="auth.user.value?.name" size="sm" />
                <div>
                  <p class="font-medium text-stone-800 dark:text-stone-200">{{ auth.user.value?.name }}</p>
                  <p class="text-xs text-stone-500 dark:text-stone-400">{{ auth.user.value?.email }}</p>
                </div>
              </div>
              <NuxtLink
                to="/notifications"
                class="py-3 px-4 rounded-xl text-stone-700 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-all duration-300 font-medium flex items-center gap-2"
                @click="mobileMenuOpen = false"
              >
                <div class="relative">
                  <UIcon name="i-heroicons-bell" class="w-5 h-5" />
                  <span
                    v-if="unreadCount > 0"
                    class="absolute -top-1 -right-1 bg-red-500 text-white text-[10px] font-bold rounded-full min-w-[14px] h-[14px] flex items-center justify-center px-0.5"
                  >
                    {{ unreadCount > 9 ? '9+' : unreadCount }}
                  </span>
                </div>
                Notifications
              </NuxtLink>
              <NuxtLink
                to="/account"
                class="py-3 px-4 rounded-xl text-stone-700 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-all duration-300 font-medium flex items-center gap-2"
                @click="mobileMenuOpen = false"
              >
                <UIcon name="i-heroicons-user" class="w-5 h-5" />
                My Account
              </NuxtLink>
              <NuxtLink
                v-if="auth.isAdmin.value || auth.isAuthor.value"
                to="/admin"
                class="py-3 px-4 rounded-xl text-stone-700 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-all duration-300 font-medium flex items-center gap-2"
                @click="mobileMenuOpen = false"
              >
                <UIcon name="i-heroicons-cog-6-tooth" class="w-5 h-5" />
                Admin Panel
              </NuxtLink>
              <button
                class="w-full py-3 px-4 rounded-xl text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/30 transition-all duration-300 font-medium flex items-center gap-2"
                @click="auth.logout(); mobileMenuOpen = false"
              >
                <UIcon name="i-heroicons-arrow-right-on-rectangle" class="w-5 h-5" />
                Logout
              </button>
            </template>
            <template v-else>
              <NuxtLink
                to="/login"
                class="py-3 px-4 rounded-xl text-stone-700 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-all duration-300 font-medium flex items-center gap-2"
                @click="mobileMenuOpen = false"
              >
                <UIcon name="i-heroicons-user-circle" class="w-5 h-5" />
                Sign In
              </NuxtLink>
              <NuxtLink
                to="/register"
                class="py-3 px-4 rounded-xl text-orange-600 dark:text-orange-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 transition-all duration-300 font-medium flex items-center gap-2"
                @click="mobileMenuOpen = false"
              >
                <UIcon name="i-heroicons-user-plus" class="w-5 h-5" />
                Create Account
              </NuxtLink>
            </template>
          </div>
          <div class="mt-4 pt-4 border-t border-stone-200 dark:border-stone-800 flex justify-between items-center">
            <span class="text-sm text-stone-500 dark:text-stone-400">Theme</span>
            <UColorModeButton class="rounded-full" />
          </div>
        </div>
      </Transition>
    </UContainer>
  </header>
</template>
