<script setup lang="ts">
const currentYear = new Date().getFullYear()

const api = useApi()
const { data: categories } = await useAsyncData('footer-categories', () => api.getCategories())
</script>

<template>
  <footer class="bg-stone-100 dark:bg-stone-900 border-t border-stone-200 dark:border-stone-800 mt-auto relative overflow-hidden">
    <!-- Decorative elements -->
    <div class="absolute top-0 right-0 w-64 h-64 bg-orange-200/20 dark:bg-orange-900/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 pointer-events-none" />

    <UContainer class="py-16 relative">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-12">
        <!-- Brand -->
        <div class="md:col-span-2">
          <NuxtLink to="/" class="inline-block group">
            <span class="text-3xl font-extrabold tracking-tight">
              <span class="text-stone-800 dark:text-white transition-colors duration-300">Pulpul</span><span class="text-orange-500 transition-colors duration-300">itiko</span>
            </span>
          </NuxtLink>
          <p class="mt-5 text-stone-600 dark:text-stone-400 max-w-md leading-relaxed">
            Your trusted source for Philippine political news and commentary.
            Covering the latest developments in national and local politics.
          </p>
          <!-- Social-like decorative elements -->
          <div class="mt-6 flex gap-3">
            <div class="w-10 h-10 rounded-full bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center hover:bg-orange-200 dark:hover:bg-orange-900/50 transition-colors cursor-pointer">
              <UIcon name="i-heroicons-rss" class="w-5 h-5 text-orange-600 dark:text-orange-400" />
            </div>
          </div>
        </div>

        <!-- Categories -->
        <div>
          <h3 class="font-bold text-stone-800 dark:text-white mb-5 flex items-center gap-2">
            <span class="w-1 h-4 bg-orange-500 rounded-full" />
            Categories
          </h3>
          <ul class="space-y-3">
            <li v-for="category in categories" :key="category.id">
              <NuxtLink
                :to="`/category/${category.slug}`"
                class="text-stone-600 dark:text-stone-400 hover:text-orange-500 transition-colors duration-300 animated-underline inline-block"
              >
                {{ category.name }}
              </NuxtLink>
            </li>
          </ul>
        </div>

        <!-- Quick Links -->
        <div>
          <h3 class="font-bold text-stone-800 dark:text-white mb-5 flex items-center gap-2">
            <span class="w-1 h-4 bg-orange-500 rounded-full" />
            Quick Links
          </h3>
          <ul class="space-y-3">
            <li>
              <NuxtLink
                to="/"
                class="text-stone-600 dark:text-stone-400 hover:text-orange-500 transition-colors duration-300 animated-underline inline-block"
              >
                Home
              </NuxtLink>
            </li>
            <li>
              <NuxtLink
                to="/search"
                class="text-stone-600 dark:text-stone-400 hover:text-orange-500 transition-colors duration-300 animated-underline inline-block"
              >
                Search
              </NuxtLink>
            </li>
          </ul>
        </div>
      </div>

      <!-- Copyright -->
      <div class="mt-16 pt-8 border-t border-stone-200 dark:border-stone-800">
        <p class="text-center text-stone-500 dark:text-stone-500 text-sm">
          &copy; {{ currentYear }} Pulpulitiko. All rights reserved.
        </p>
      </div>
    </UContainer>
  </footer>
</template>
