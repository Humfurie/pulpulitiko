<script setup lang="ts">
import type { ArticleListItem } from '~/types'

const props = defineProps<{
  article: ArticleListItem
  featured?: boolean
  variant?: 'default' | 'overlay' | 'horizontal'
}>()

function formatDate(dateString?: string): string {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}
</script>

<template>
  <!-- Featured/Hero Card -->
  <article
    v-if="featured"
    class="group relative rounded-3xl overflow-hidden bg-gray-100 dark:bg-gray-800 aspect-[16/10] md:aspect-[21/9]"
  >
    <NuxtLink :to="`/article/${article.slug}`" class="absolute inset-0">
      <NuxtImg
        v-if="article.featured_image"
        :src="article.featured_image"
        :alt="article.title"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
        loading="lazy"
      />
      <div v-else class="w-full h-full flex items-center justify-center">
        <UIcon name="i-heroicons-photo" class="w-20 h-20 text-gray-400" />
      </div>
      <!-- Gradient overlay -->
      <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/40 to-transparent" />
    </NuxtLink>

    <!-- Content overlay -->
    <div class="absolute bottom-0 left-0 right-0 p-6 md:p-10">
      <div class="flex items-center gap-3 mb-4">
        <span v-if="article.published_at" class="px-3 py-1 bg-white/20 backdrop-blur-sm rounded-full text-white text-sm">
          {{ formatDate(article.published_at) }}
        </span>
        <NuxtLink
          v-if="article.category_slug && article.category_name"
          :to="`/category/${article.category_slug}`"
          class="px-3 py-1 bg-white rounded-full text-gray-900 text-sm font-medium hover:bg-gray-100 transition-colors"
        >
          {{ article.category_name }}
        </NuxtLink>
      </div>
      <NuxtLink :to="`/article/${article.slug}`">
        <h2 class="text-2xl md:text-4xl lg:text-5xl font-bold text-white leading-tight mb-3 group-hover:underline decoration-2 underline-offset-4">
          {{ article.title }}
        </h2>
      </NuxtLink>
      <p v-if="article.summary" class="text-white/80 text-base md:text-lg line-clamp-2 max-w-3xl">
        {{ article.summary }}
      </p>
    </div>

    <!-- Arrow icon -->
    <NuxtLink
      :to="`/article/${article.slug}`"
      class="absolute bottom-6 right-6 md:bottom-10 md:right-10 w-12 h-12 bg-white rounded-full flex items-center justify-center group-hover:bg-primary transition-colors"
    >
      <UIcon name="i-heroicons-arrow-up-right" class="w-5 h-5 text-gray-900 group-hover:text-white" />
    </NuxtLink>
  </article>

  <!-- Overlay Card (for grid layout) -->
  <article
    v-else-if="variant === 'overlay'"
    class="group relative rounded-2xl overflow-hidden bg-gray-100 dark:bg-gray-800 aspect-[4/5]"
  >
    <NuxtLink :to="`/article/${article.slug}`" class="absolute inset-0">
      <NuxtImg
        v-if="article.featured_image"
        :src="article.featured_image"
        :alt="article.title"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
        loading="lazy"
      />
      <div v-else class="w-full h-full flex items-center justify-center">
        <UIcon name="i-heroicons-photo" class="w-12 h-12 text-gray-400" />
      </div>
      <!-- Gradient overlay -->
      <div class="absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent" />
    </NuxtLink>

    <!-- Content overlay -->
    <div class="absolute bottom-0 left-0 right-0 p-5">
      <NuxtLink
        v-if="article.category_slug && article.category_name"
        :to="`/category/${article.category_slug}`"
        class="inline-block mb-3 px-3 py-1 bg-white/20 backdrop-blur-sm rounded-full text-white text-xs font-medium hover:bg-white/30 transition-colors"
      >
        {{ article.category_name }}
      </NuxtLink>
      <NuxtLink :to="`/article/${article.slug}`">
        <h3 class="text-lg font-bold text-white leading-snug line-clamp-3 group-hover:underline decoration-1 underline-offset-2">
          {{ article.title }}
        </h3>
      </NuxtLink>
    </div>

    <!-- Arrow icon -->
    <NuxtLink
      :to="`/article/${article.slug}`"
      class="absolute top-4 right-4 w-10 h-10 bg-white rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
    >
      <UIcon name="i-heroicons-arrow-up-right" class="w-4 h-4 text-gray-900" />
    </NuxtLink>
  </article>

  <!-- Horizontal Card -->
  <article
    v-else-if="variant === 'horizontal'"
    class="group flex gap-4 items-start"
  >
    <NuxtLink :to="`/article/${article.slug}`" class="shrink-0 w-24 h-24 rounded-xl overflow-hidden bg-gray-100 dark:bg-gray-800">
      <NuxtImg
        v-if="article.featured_image"
        :src="article.featured_image"
        :alt="article.title"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
        loading="lazy"
      />
      <div v-else class="w-full h-full flex items-center justify-center">
        <UIcon name="i-heroicons-photo" class="w-8 h-8 text-gray-400" />
      </div>
    </NuxtLink>
    <div class="flex-1 min-w-0">
      <NuxtLink
        v-if="article.category_slug && article.category_name"
        :to="`/category/${article.category_slug}`"
        class="text-xs font-medium text-primary hover:underline"
      >
        {{ article.category_name }}
      </NuxtLink>
      <NuxtLink :to="`/article/${article.slug}`">
        <h4 class="font-semibold text-gray-900 dark:text-white leading-snug line-clamp-2 group-hover:text-primary transition-colors mt-1">
          {{ article.title }}
        </h4>
      </NuxtLink>
      <span v-if="article.published_at" class="text-xs text-gray-500 mt-1 block">
        {{ formatDate(article.published_at) }}
      </span>
    </div>
  </article>

  <!-- Default Card -->
  <article
    v-else
    class="group bg-white dark:bg-gray-900 rounded-2xl overflow-hidden border border-gray-200 dark:border-gray-800 hover:shadow-xl hover:shadow-gray-200/50 dark:hover:shadow-gray-900/50 transition-all duration-300"
  >
    <!-- Image -->
    <NuxtLink :to="`/article/${article.slug}`" class="block overflow-hidden aspect-[16/10]">
      <NuxtImg
        v-if="article.featured_image"
        :src="article.featured_image"
        :alt="article.title"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
        loading="lazy"
      />
      <div v-else class="w-full h-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
        <UIcon name="i-heroicons-photo" class="w-12 h-12 text-gray-400" />
      </div>
    </NuxtLink>

    <!-- Content -->
    <div class="p-5">
      <div class="flex items-center gap-3 mb-3">
        <NuxtLink
          v-if="article.category_slug && article.category_name"
          :to="`/category/${article.category_slug}`"
          class="text-xs font-semibold text-primary hover:underline uppercase tracking-wide"
        >
          {{ article.category_name }}
        </NuxtLink>
        <span v-if="article.category_name && article.published_at" class="text-gray-300 dark:text-gray-600">•</span>
        <span v-if="article.published_at" class="text-xs text-gray-500">
          {{ formatDate(article.published_at) }}
        </span>
      </div>

      <NuxtLink :to="`/article/${article.slug}`">
        <h3 class="text-xl font-bold text-gray-900 dark:text-white leading-snug line-clamp-2 group-hover:text-primary transition-colors">
          {{ article.title }}
        </h3>
      </NuxtLink>

      <p v-if="article.summary" class="mt-3 text-gray-600 dark:text-gray-400 text-sm line-clamp-2">
        {{ article.summary }}
      </p>

      <div v-if="article.author_name" class="mt-4 flex items-center gap-2">
        <div class="w-8 h-8 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
          <UIcon name="i-heroicons-user" class="w-4 h-4 text-gray-500" />
        </div>
        <span class="text-sm text-gray-600 dark:text-gray-400">{{ article.author_name }}</span>
      </div>
    </div>
  </article>
</template>
