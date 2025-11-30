<script setup lang="ts">
const auth = useAuth()
const route = useRoute()

// Protect admin routes
onMounted(async () => {
  if (!auth.token.value) {
    await navigateTo('/login')
    return
  }

  await auth.fetchCurrentUser()

  if (!auth.isAuthenticated.value) {
    await navigateTo('/login')
  }
})

const navigation = [
  { name: 'Dashboard', href: '/admin', icon: 'i-heroicons-home' },
  { name: 'Articles', href: '/admin/articles', icon: 'i-heroicons-document-text' },
  { name: 'Categories', href: '/admin/categories', icon: 'i-heroicons-folder' },
  { name: 'Tags', href: '/admin/tags', icon: 'i-heroicons-tag' }
]

function isActive(href: string) {
  if (href === '/admin') {
    return route.path === '/admin'
  }
  return route.path.startsWith(href)
}
</script>

<template>
  <div class="min-h-screen bg-gray-100 dark:bg-gray-950">
    <!-- Sidebar -->
    <aside class="fixed inset-y-0 left-0 w-64 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800">
      <div class="flex flex-col h-full">
        <!-- Logo -->
        <div class="p-4 border-b border-gray-200 dark:border-gray-800">
          <NuxtLink to="/admin" class="text-xl font-bold text-primary">
            Pulpulitiko Admin
          </NuxtLink>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 p-4 space-y-1">
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
          <div class="flex items-center gap-3 mb-3">
            <UAvatar :alt="auth.user.value?.name" size="sm" />
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
                {{ auth.user.value?.name }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400 truncate">
                {{ auth.user.value?.email }}
              </p>
            </div>
          </div>
          <div class="flex gap-2">
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

    <!-- Main content -->
    <main class="pl-64">
      <div class="p-8">
        <slot />
      </div>
    </main>
  </div>
</template>
