<script setup lang="ts">
import type { Category } from '~/types'

const api = useApi()
const auth = useAuth()
const route = useRoute()
const searchQuery = ref('')
const mobileMenuOpen = ref(false)
const isScrolled = ref(false)
const searchFocused = ref(false)

const { data: categories } = await useAsyncData('categories', () => api.getCategories())

// Active state for navigation
const isHomeActive = computed(() => route.path === '/')
const isCategoryActive = computed(() => route.path.startsWith('/category'))

// Check auth on mount
onMounted(async () => {
  if (auth.token.value && !auth.user.value) {
    await auth.fetchCurrentUser()
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
          <span class="text-2xl font-extrabold tracking-tight">
            <span class="text-stone-800 dark:text-white transition-colors duration-300">Pulpul</span><span class="text-orange-500 transition-colors duration-300">itiko</span>
          </span>
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
