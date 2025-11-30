<script setup lang="ts">
import type { Category } from '~/types'

const api = useApi()
const searchQuery = ref('')
const mobileMenuOpen = ref(false)

const { data: categories } = await useAsyncData('categories', () => api.getCategories())

function handleSearch() {
  if (searchQuery.value.trim()) {
    navigateTo({ path: '/search', query: { q: searchQuery.value.trim() } })
    searchQuery.value = ''
    mobileMenuOpen.value = false
  }
}
</script>

<template>
  <header class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 sticky top-0 z-50">
    <UContainer>
      <div class="flex items-center justify-between h-16">
        <!-- Logo -->
        <NuxtLink to="/" class="flex items-center gap-2">
          <span class="text-2xl font-bold text-primary">Pulpulitiko</span>
        </NuxtLink>

        <!-- Desktop Navigation -->
        <nav class="hidden md:flex items-center gap-6">
          <NuxtLink
            to="/"
            class="text-gray-600 dark:text-gray-300 hover:text-primary transition-colors"
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
            <UButton variant="ghost" trailing-icon="i-heroicons-chevron-down">
              Categories
            </UButton>
          </UDropdownMenu>
        </nav>

        <!-- Search -->
        <div class="hidden md:flex items-center gap-4">
          <form @submit.prevent="handleSearch" class="relative">
            <UInput
              v-model="searchQuery"
              placeholder="Search articles..."
              icon="i-heroicons-magnifying-glass"
              class="w-64"
            />
          </form>
          <UColorModeButton />
        </div>

        <!-- Mobile Menu Button -->
        <UButton
          class="md:hidden"
          variant="ghost"
          :icon="mobileMenuOpen ? 'i-heroicons-x-mark' : 'i-heroicons-bars-3'"
          @click="mobileMenuOpen = !mobileMenuOpen"
        />
      </div>

      <!-- Mobile Menu -->
      <div v-if="mobileMenuOpen" class="md:hidden py-4 border-t border-gray-200 dark:border-gray-800">
        <form @submit.prevent="handleSearch" class="mb-4">
          <UInput
            v-model="searchQuery"
            placeholder="Search articles..."
            icon="i-heroicons-magnifying-glass"
            class="w-full"
          />
        </form>
        <nav class="flex flex-col gap-2">
          <NuxtLink
            to="/"
            class="py-2 text-gray-600 dark:text-gray-300"
            @click="mobileMenuOpen = false"
          >
            Home
          </NuxtLink>
          <template v-if="categories?.length">
            <p class="py-2 text-sm font-semibold text-gray-500">Categories</p>
            <NuxtLink
              v-for="cat in categories"
              :key="cat.id"
              :to="`/category/${cat.slug}`"
              class="py-2 pl-4 text-gray-600 dark:text-gray-300"
              @click="mobileMenuOpen = false"
            >
              {{ cat.name }}
            </NuxtLink>
          </template>
        </nav>
        <div class="mt-4 flex justify-end">
          <UColorModeButton />
        </div>
      </div>
    </UContainer>
  </header>
</template>
