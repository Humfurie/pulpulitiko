<script setup lang="ts">
import type { ArticleListItem } from '~/types'

defineProps<{
  article: ArticleListItem
  featured?: boolean
}>()

function formatDate(dateString?: string): string {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}
</script>

<template>
  <article
    class="group bg-white dark:bg-gray-900 rounded-lg overflow-hidden border border-gray-200 dark:border-gray-800 hover:shadow-lg transition-shadow"
    :class="featured ? 'md:flex' : ''"
  >
    <!-- Image -->
    <NuxtLink
      :to="`/article/${article.slug}`"
      class="block overflow-hidden"
      :class="featured ? 'md:w-1/2' : ''"
    >
      <NuxtImg
        v-if="article.featured_image"
        :src="article.featured_image"
        :alt="article.title"
        class="w-full object-cover group-hover:scale-105 transition-transform duration-300"
        :class="featured ? 'h-64 md:h-full' : 'h-48'"
        loading="lazy"
      />
      <div
        v-else
        class="w-full bg-gray-200 dark:bg-gray-800 flex items-center justify-center"
        :class="featured ? 'h-64 md:h-full' : 'h-48'"
      >
        <UIcon name="i-heroicons-photo" class="w-12 h-12 text-gray-400" />
      </div>
    </NuxtLink>

    <!-- Content -->
    <div class="p-6" :class="featured ? 'md:w-1/2 md:flex md:flex-col md:justify-center' : ''">
      <!-- Category -->
      <NuxtLink
        v-if="article.category_slug && article.category_name"
        :to="`/category/${article.category_slug}`"
        class="inline-block mb-2"
      >
        <UBadge color="primary" variant="subtle">
          {{ article.category_name }}
        </UBadge>
      </NuxtLink>

      <!-- Title -->
      <NuxtLink :to="`/article/${article.slug}`">
        <h2
          class="font-bold text-gray-900 dark:text-white group-hover:text-primary transition-colors line-clamp-2"
          :class="featured ? 'text-2xl md:text-3xl' : 'text-xl'"
        >
          {{ article.title }}
        </h2>
      </NuxtLink>

      <!-- Summary -->
      <p
        v-if="article.summary"
        class="mt-3 text-gray-600 dark:text-gray-400 line-clamp-3"
        :class="featured ? 'text-base' : 'text-sm'"
      >
        {{ article.summary }}
      </p>

      <!-- Meta -->
      <div class="mt-4 flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
        <span v-if="article.author_name" class="flex items-center gap-1">
          <UIcon name="i-heroicons-user" class="w-4 h-4" />
          {{ article.author_name }}
        </span>
        <span v-if="article.published_at" class="flex items-center gap-1">
          <UIcon name="i-heroicons-calendar" class="w-4 h-4" />
          {{ formatDate(article.published_at) }}
        </span>
      </div>
    </div>
  </article>
</template>
