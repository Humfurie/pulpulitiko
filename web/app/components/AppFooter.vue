<script setup lang="ts">
const currentYear = new Date().getFullYear()

const api = useApi()
const { data: categories } = await useAsyncData('footer-categories', () => api.getCategories())
</script>

<template>
  <footer class="bg-gray-100 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-800 mt-auto">
    <UContainer class="py-12">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-8">
        <!-- Brand -->
        <div class="md:col-span-2">
          <NuxtLink to="/" class="text-2xl font-bold text-primary">
            Pulpulitiko
          </NuxtLink>
          <p class="mt-4 text-gray-600 dark:text-gray-400 max-w-md">
            Your trusted source for Philippine political news and commentary.
            Covering the latest developments in national and local politics.
          </p>
        </div>

        <!-- Categories -->
        <div>
          <h3 class="font-semibold text-gray-900 dark:text-white mb-4">Categories</h3>
          <ul class="space-y-2">
            <li v-for="category in categories" :key="category.id">
              <NuxtLink
                :to="`/category/${category.slug}`"
                class="text-gray-600 dark:text-gray-400 hover:text-primary transition-colors"
              >
                {{ category.name }}
              </NuxtLink>
            </li>
          </ul>
        </div>

        <!-- Quick Links -->
        <div>
          <h3 class="font-semibold text-gray-900 dark:text-white mb-4">Quick Links</h3>
          <ul class="space-y-2">
            <li>
              <NuxtLink
                to="/"
                class="text-gray-600 dark:text-gray-400 hover:text-primary transition-colors"
              >
                Home
              </NuxtLink>
            </li>
            <li>
              <NuxtLink
                to="/search"
                class="text-gray-600 dark:text-gray-400 hover:text-primary transition-colors"
              >
                Search
              </NuxtLink>
            </li>
          </ul>
        </div>
      </div>

      <!-- Copyright -->
      <div class="mt-12 pt-8 border-t border-gray-200 dark:border-gray-800">
        <p class="text-center text-gray-500 dark:text-gray-400 text-sm">
          &copy; {{ currentYear }} Pulpulitiko. All rights reserved.
        </p>
      </div>
    </UContainer>
  </footer>
</template>
