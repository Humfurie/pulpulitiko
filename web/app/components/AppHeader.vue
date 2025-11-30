<script setup lang="ts">
import type { Category } from '~/types'

const api = useApi()
const searchQuery = ref('')
const mobileMenuOpen = ref(false)
const isScrolled = ref(false)
const searchFocused = ref(false)

const { data: categories } = await useAsyncData('categories', () => api.getCategories())

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
            class="px-4 py-2 rounded-full text-stone-600 dark:text-stone-300 hover:text-orange-500 hover:bg-orange-50 dark:hover:bg-orange-950/30 transition-all duration-300 font-medium"
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
              variant="ghost"
              trailing-icon="i-heroicons-chevron-down"
              class="rounded-full hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-all duration-300"
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
          <!-- Admin/User Button -->
          <NuxtLink to="/admin">
            <UButton
              variant="ghost"
              icon="i-heroicons-user-circle"
              class="rounded-full hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-colors duration-300"
            />
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
              class="py-3 px-4 rounded-xl text-stone-700 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-all duration-300 font-medium"
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
                class="py-3 px-4 rounded-xl text-stone-600 dark:text-stone-400 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-all duration-300"
                @click="mobileMenuOpen = false"
              >
                {{ cat.name }}
              </NuxtLink>
            </template>
          </nav>
          <!-- Admin Link (Mobile) -->
          <NuxtLink
            to="/admin"
            class="py-3 px-4 rounded-xl text-stone-700 dark:text-stone-300 hover:bg-orange-50 dark:hover:bg-orange-950/30 hover:text-orange-500 transition-all duration-300 font-medium flex items-center gap-2 mt-2"
            @click="mobileMenuOpen = false"
          >
            <UIcon name="i-heroicons-user-circle" class="w-5 h-5" />
            Admin Panel
          </NuxtLink>
          <div class="mt-4 pt-4 border-t border-stone-200 dark:border-stone-800 flex justify-between items-center">
            <span class="text-sm text-stone-500 dark:text-stone-400">Theme</span>
            <UColorModeButton class="rounded-full" />
          </div>
        </div>
      </Transition>
    </UContainer>
  </header>
</template>
