<script setup lang="ts">
const auth = useAuth()
const route = useRoute()

// Mobile sidebar state
const sidebarOpen = ref(false)

// Close sidebar on route change (mobile)
watch(() => route.path, () => {
  sidebarOpen.value = false
})

// Navigation items with role requirements
// roles: ['admin'] = admin only, ['admin', 'author'] = admin and author
const allNavigation = [
  { name: 'Dashboard', href: '/admin', icon: 'i-heroicons-home', roles: ['admin', 'author'] },
  { name: 'Articles', href: '/admin/articles', icon: 'i-heroicons-document-text', roles: ['admin', 'author'] },
  { name: 'Categories', href: '/admin/categories', icon: 'i-heroicons-folder', roles: ['admin', 'author'] },
  { name: 'Tags', href: '/admin/tags', icon: 'i-heroicons-tag', roles: ['admin', 'author'] },
  { name: 'Politicians', href: '/admin/politicians', icon: 'i-heroicons-user-circle', roles: ['admin', 'author'] },
  { name: 'Polls', href: '/admin/polls', icon: 'i-heroicons-chart-pie', roles: ['admin'] },
  { name: 'Elections', href: '/admin/elections', icon: 'i-heroicons-clipboard-document-check', roles: ['admin'] },
  { name: 'Legislation', href: '/admin/legislation', icon: 'i-heroicons-scale', roles: ['admin'] },
  { name: 'Locations', href: '/admin/locations', icon: 'i-heroicons-map-pin', roles: ['admin'] },
  { name: 'Analytics', href: '/admin/analytics', icon: 'i-heroicons-chart-bar', roles: ['admin'] },
  { name: 'Users', href: '/admin/users', icon: 'i-heroicons-users', roles: ['admin'] },
  { name: 'Roles', href: '/admin/roles', icon: 'i-heroicons-shield-check', roles: ['admin'] },
  { name: 'Messages', href: '/admin/messages', icon: 'i-heroicons-chat-bubble-left-right', roles: ['admin'] }
]

// Filter navigation based on user role
const navigation = computed(() => {
  const userRole = auth.user.value?.role
  if (!userRole) return []

  return allNavigation.filter(item => {
    return item.roles.includes(userRole)
  })
})

// Check if user is a regular user (not admin/author)
const isRegularUser = computed(() => auth.user.value?.role === 'user')

function isActive(href: string) {
  if (href === '/admin') {
    return route.path === '/admin'
  }
  // Check exact match or match with subpath (e.g., /admin/articles matches /admin/articles/123)
  const result = route.path === href || route.path.startsWith(href + '/')
  if (result) {
    console.log('isActive:', { href, routePath: route.path, result })
  }
  return result
}
</script>

<template>
  <div class="min-h-screen bg-gray-100 dark:bg-gray-950">
    <!-- Mobile header -->
    <header v-if="!isRegularUser" class="lg:hidden fixed top-0 left-0 right-0 z-40 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800">
      <div class="flex items-center justify-between px-4 py-3">
        <UButton
          variant="ghost"
          size="sm"
          icon="i-heroicons-bars-3"
          class="-ml-2 text-gray-700 dark:text-gray-300"
          @click="sidebarOpen = true"
        />
        <NuxtLink to="/admin" class="flex items-center gap-2">
          <img src="/pulpulitiko.png" alt="Pulpulitiko" class="h-6 w-auto dark:hidden">
          <img src="/pulpulitiko_dark.png" alt="Pulpulitiko" class="h-6 w-auto hidden dark:block">
        </NuxtLink>
        <div class="flex items-center gap-2">
          <UColorModeButton variant="ghost" size="xs" />
          <UAvatar :src="auth.user.value?.avatar" :alt="auth.user.value?.name" size="xs" />
        </div>
      </div>
    </header>

    <!-- Mobile sidebar overlay + sidebar (in same stacking context) -->
    <Teleport to="body">
      <!-- Overlay -->
      <Transition
        enter-active-class="transition-opacity duration-300"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-300"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="sidebarOpen && !isRegularUser"
          class="fixed inset-0 z-[55] bg-black/50 lg:hidden"
          @click="sidebarOpen = false"
        />
      </Transition>

      <!-- Mobile Sidebar -->
      <aside
        v-if="!isRegularUser"
        :class="[
          'fixed inset-y-0 left-0 z-[60] w-64 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 transition-transform duration-300 lg:hidden',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        ]"
      >
      <div class="flex flex-col h-full">
        <!-- Logo -->
        <div class="p-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
          <NuxtLink to="/admin" class="flex items-center gap-2">
            <img
              src="/pulpulitiko.png"
              alt="Pulpulitiko"
              class="h-7 w-auto dark:hidden"
            >
            <img
              src="/pulpulitiko_dark.png"
              alt="Pulpulitiko"
              class="h-7 w-auto hidden dark:block"
            >
            <span class="text-sm font-medium text-gray-500 dark:text-gray-400">Admin</span>
          </NuxtLink>
          <!-- Close button for mobile -->
          <UButton
            variant="ghost"
            size="xs"
            icon="i-heroicons-x-mark"
            class="lg:hidden text-gray-700 dark:text-gray-300"
            @click="sidebarOpen = false"
          />
        </div>

        <!-- Navigation -->
        <nav class="flex-1 p-4 space-y-1 overflow-y-auto">
          <NuxtLink
            v-for="item in navigation"
            :key="item.name"
            :to="item.href"
            :class="[
              'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
              isActive(item.href)
                ? 'bg-primary/10 text-primary'
                : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
            ]"
          >
            <span :class="[item.icon, 'size-5']" />
            {{ item.name }}
          </NuxtLink>
        </nav>

        <!-- User section -->
        <div class="p-4 border-t border-gray-200 dark:border-gray-800">
          <NuxtLink
            to="/account"
            class="flex items-center gap-3 mb-3 p-2 -m-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors cursor-pointer"
          >
            <UAvatar :src="auth.user.value?.avatar" :alt="auth.user.value?.name" size="sm" />
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
                  {{ auth.user.value?.name }}
                </p>
                <UBadge
                  :color="auth.user.value?.role === 'admin' ? 'error' : auth.user.value?.role === 'author' ? 'primary' : 'neutral'"
                  variant="subtle"
                  size="xs"
                  class="capitalize"
                >
                  {{ auth.user.value?.role }}
                </UBadge>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400 truncate">
                {{ auth.user.value?.email }}
              </p>
            </div>
          </NuxtLink>
          <div class="flex gap-2">
            <UColorModeButton
              variant="ghost"
              size="sm"
              class="flex-shrink-0 hidden lg:flex"
            />
            <UButton
              to="/"
              variant="ghost"
              size="sm"
              icon="i-heroicons-arrow-left"
              class="flex-1"
            >
              Site
            </UButton>
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-arrow-right-on-rectangle"
              @click="auth.logout()"
            >
              Logout
            </UButton>
          </div>
        </div>
      </div>
    </aside>
    </Teleport>

    <!-- Desktop Sidebar (always visible on lg+) -->
    <aside
      v-if="!isRegularUser"
      class="hidden lg:block fixed inset-y-0 left-0 z-30 w-64 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800"
    >
      <div class="flex flex-col h-full">
        <!-- Logo -->
        <div class="p-4 border-b border-gray-200 dark:border-gray-800">
          <NuxtLink to="/admin" class="flex items-center gap-2">
            <img
              src="/pulpulitiko.png"
              alt="Pulpulitiko"
              class="h-7 w-auto dark:hidden"
            >
            <img
              src="/pulpulitiko_dark.png"
              alt="Pulpulitiko"
              class="h-7 w-auto hidden dark:block"
            >
            <span class="text-sm font-medium text-gray-500 dark:text-gray-400">Admin</span>
          </NuxtLink>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 p-4 space-y-1 overflow-y-auto">
          <NuxtLink
            v-for="item in navigation"
            :key="item.name"
            :to="item.href"
            :class="[
              'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
              isActive(item.href)
                ? 'bg-primary/10 text-primary'
                : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
            ]"
          >
            <span :class="[item.icon, 'size-5']" />
            {{ item.name }}
          </NuxtLink>
        </nav>

        <!-- User section -->
        <div class="p-4 border-t border-gray-200 dark:border-gray-800">
          <NuxtLink
            to="/account"
            class="flex items-center gap-3 mb-3 p-2 -m-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors cursor-pointer"
          >
            <UAvatar :src="auth.user.value?.avatar" :alt="auth.user.value?.name" size="sm" />
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
                  {{ auth.user.value?.name }}
                </p>
                <UBadge
                  :color="auth.user.value?.role === 'admin' ? 'error' : auth.user.value?.role === 'author' ? 'primary' : 'neutral'"
                  variant="subtle"
                  size="xs"
                  class="capitalize"
                >
                  {{ auth.user.value?.role }}
                </UBadge>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400 truncate">
                {{ auth.user.value?.email }}
              </p>
            </div>
          </NuxtLink>
          <div class="flex gap-2">
            <UColorModeButton
              variant="ghost"
              size="sm"
              class="flex-shrink-0"
            />
            <UButton
              to="/"
              variant="ghost"
              size="sm"
              icon="i-heroicons-arrow-left"
              class="flex-1"
            >
              Site
            </UButton>
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-arrow-right-on-rectangle"
              @click="auth.logout()"
            >
              Logout
            </UButton>
          </div>
        </div>
      </div>
    </aside>

    <!-- Simple header for regular users -->
    <header v-else class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800">
      <div class="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
        <NuxtLink to="/">
          <img
            src="/pulpulitiko.png"
            alt="Pulpulitiko"
            class="h-8 w-auto dark:hidden"
          />
          <img
            src="/pulpulitiko_dark.png"
            alt="Pulpulitiko"
            class="h-8 w-auto hidden dark:block"
          />
        </NuxtLink>
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2">
            <UAvatar :src="auth.user.value?.avatar" :alt="auth.user.value?.name" size="sm" />
            <span class="text-sm font-medium text-gray-900 dark:text-white">
              {{ auth.user.value?.name }}
            </span>
          </div>
          <UButton
            variant="ghost"
            size="sm"
            color="error"
            icon="i-heroicons-arrow-right-on-rectangle"
            @click="auth.logout()"
          >
            Logout
          </UButton>
        </div>
      </div>
    </header>

    <!-- Main content -->
    <main :class="isRegularUser ? '' : 'lg:pl-64'">
      <div :class="isRegularUser ? 'max-w-4xl mx-auto px-4 py-8' : 'p-4 pt-16 lg:p-8 lg:pt-8'">
        <slot />
      </div>
    </main>
  </div>
</template>
