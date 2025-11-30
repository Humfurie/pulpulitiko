<script setup lang="ts">
const auth = useAuth()
const route = useRoute()

// Navigation items with role requirements
// roles: undefined = all authenticated users, ['admin'] = admin only, ['admin', 'author'] = admin and author
const allNavigation = [
  { name: 'Dashboard', href: '/admin', icon: 'i-heroicons-home', roles: undefined },
  { name: 'Articles', href: '/admin/articles', icon: 'i-heroicons-document-text', roles: ['admin', 'author'] },
  { name: 'Categories', href: '/admin/categories', icon: 'i-heroicons-folder', roles: ['admin', 'author'] },
  { name: 'Tags', href: '/admin/tags', icon: 'i-heroicons-tag', roles: ['admin', 'author'] },
  { name: 'Users', href: '/admin/users', icon: 'i-heroicons-users', roles: ['admin'] },
  { name: 'Roles', href: '/admin/roles', icon: 'i-heroicons-shield-check', roles: ['admin'] }
]

// Filter navigation based on user role
const navigation = computed(() => {
  const userRole = auth.user.value?.role
  if (!userRole) return []

  return allNavigation.filter(item => {
    if (!item.roles) return true // Visible to all authenticated users
    return item.roles.includes(userRole)
  })
})

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
          <NuxtLink
            to="/admin/account"
            class="flex items-center gap-3 mb-3 p-2 -m-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors cursor-pointer"
          >
            <UAvatar :alt="auth.user.value?.name" size="sm" />
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
