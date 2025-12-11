<script setup lang="ts">
import type { ArticleListItem } from '~/types'

const props = defineProps<{
  article: ArticleListItem
  featured?: boolean
  variant?: 'default' | 'overlay' | 'horizontal'
}>()

const cardRef = ref<HTMLElement | null>(null)
const isHovered = ref(false)

function formatDate(dateString?: string): string {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

// Parallax effect on hover for featured card
function handleMouseMove(e: MouseEvent) {
  if (!cardRef.value || !props.featured) return

  const rect = cardRef.value.getBoundingClientRect()
  const x = (e.clientX - rect.left) / rect.width - 0.5
  const y = (e.clientY - rect.top) / rect.height - 0.5

  const img = cardRef.value.querySelector('img')
  if (img) {
    img.style.transform = `scale(1.05) translate(${x * 10}px, ${y * 10}px)`
  }
}

function handleMouseLeave() {
  if (!cardRef.value) return
  const img = cardRef.value.querySelector('img')
  if (img) {
    img.style.transform = ''
  }
  isHovered.value = false
}
</script>

<template>
  <!-- Featured/Hero Card -->
  <article
    v-if="featured"
    ref="cardRef"
    class="group relative rounded-3xl overflow-hidden bg-stone-100 dark:bg-stone-800 aspect-[16/10] md:aspect-[21/9] shine-effect"
    @mousemove="handleMouseMove"
    @mouseleave="handleMouseLeave"
    @mouseenter="isHovered = true"
  >
    <NuxtLink :to="`/article/${article.slug}`" class="absolute inset-0">
      <NuxtImg
        v-if="article.featured_image"
        :src="article.featured_image"
        :alt="article.title"
        class="w-full h-full object-cover transition-transform duration-700 ease-[cubic-bezier(0.19,1,0.22,1)]"
        loading="lazy"
      />
      <div v-else class="w-full h-full flex items-center justify-center bg-gradient-to-br from-orange-100 to-amber-50 dark:from-stone-800 dark:to-stone-900">
        <UIcon name="i-heroicons-newspaper" class="w-20 h-20 text-orange-300 dark:text-orange-700 animate-float-slow" />
      </div>
      <!-- Gradient overlay with warm tones -->
      <div class="absolute inset-0 bg-gradient-to-t from-stone-900/95 via-stone-900/60 to-stone-900/20 group-hover:from-orange-950/95 transition-colors duration-500" />
    </NuxtLink>

    <!-- Decorative elements -->
    <div class="absolute top-6 left-6 md:top-10 md:left-10 opacity-0 group-hover:opacity-100 transition-opacity duration-500">
      <div class="w-2 h-2 bg-orange-500 rounded-full animate-pulse" />
    </div>

    <!-- Content overlay -->
    <div class="absolute bottom-0 left-0 right-0 p-6 md:p-10 transform group-hover:translate-y-0 translate-y-2 transition-transform duration-500 ease-[cubic-bezier(0.19,1,0.22,1)]">
      <div class="flex items-center gap-3 mb-4">
        <span v-if="article.published_at" class="px-4 py-1.5 bg-stone-800/70 backdrop-blur-md rounded-full text-white text-sm font-medium border border-stone-700/50">
          {{ formatDate(article.published_at) }}
        </span>
        <NuxtLink
          v-if="article.category_slug && article.category_name"
          :to="`/category/${article.category_slug}`"
          class="px-4 py-1.5 bg-orange-500 rounded-full text-white text-sm font-semibold hover:bg-orange-400 transition-all duration-300 hover:scale-105"
        >
          {{ article.category_name }}
        </NuxtLink>
      </div>
      <NuxtLink :to="`/article/${article.slug}`">
        <h2 class="text-2xl sm:text-3xl md:text-4xl lg:text-5xl font-extrabold text-white leading-tight mb-3 tracking-tight">
          <span class="bg-gradient-to-r from-white to-white bg-[length:0%_2px] bg-no-repeat bg-left-bottom group-hover:bg-[length:100%_2px] transition-all duration-500">
            {{ article.title }}
          </span>
        </h2>
      </NuxtLink>
      <p v-if="article.summary" class="text-white/90 text-base md:text-lg line-clamp-2 max-w-3xl group-hover:text-white transition-colors duration-300">
        {{ article.summary }}
      </p>
    </div>

    <!-- Arrow icon with enhanced animation -->
    <NuxtLink
      :to="`/article/${article.slug}`"
      class="absolute bottom-6 right-6 md:bottom-10 md:right-10 w-14 h-14 bg-white/90 backdrop-blur-sm rounded-full flex items-center justify-center group-hover:bg-orange-500 group-hover:scale-110 transition-all duration-300 shadow-lg group-hover:shadow-orange-500/30"
    >
      <UIcon name="i-heroicons-arrow-up-right" class="w-6 h-6 text-stone-900 group-hover:text-white group-hover:rotate-45 transition-all duration-300" />
    </NuxtLink>
  </article>

  <!-- Overlay Card (for grid layout) -->
  <article
    v-else-if="variant === 'overlay'"
    class="group relative rounded-2xl overflow-hidden bg-stone-100 dark:bg-stone-800 aspect-[4/5] card-lift"
  >
    <NuxtLink :to="`/article/${article.slug}`" class="absolute inset-0 image-reveal">
      <NuxtImg
        v-if="article.featured_image"
        :src="article.featured_image"
        :alt="article.title"
        class="w-full h-full object-cover"
        loading="lazy"
      />
      <div v-else class="w-full h-full flex items-center justify-center bg-gradient-to-br from-orange-50 to-amber-100 dark:from-stone-800 dark:to-stone-900">
        <UIcon name="i-heroicons-newspaper" class="w-12 h-12 text-orange-300 dark:text-orange-700" />
      </div>
      <!-- Gradient overlay -->
      <div class="absolute inset-0 bg-gradient-to-t from-stone-900/80 via-stone-900/20 to-transparent group-hover:from-orange-950/80 transition-colors duration-500" />
    </NuxtLink>

    <!-- Content overlay -->
    <div class="absolute bottom-0 left-0 right-0 p-5 transform translate-y-2 group-hover:translate-y-0 transition-transform duration-400 ease-[cubic-bezier(0.19,1,0.22,1)]">
      <NuxtLink
        v-if="article.category_slug && article.category_name"
        :to="`/category/${article.category_slug}`"
        class="inline-block mb-3 px-3 py-1 bg-orange-500/90 backdrop-blur-sm rounded-full text-white text-xs font-semibold hover:bg-orange-400 transition-all duration-300"
      >
        {{ article.category_name }}
      </NuxtLink>
      <NuxtLink :to="`/article/${article.slug}`">
        <h3 class="text-lg font-bold text-white leading-snug line-clamp-3">
          <span class="bg-gradient-to-r from-white to-white bg-[length:0%_1px] bg-no-repeat bg-left-bottom group-hover:bg-[length:100%_1px] transition-all duration-400">
            {{ article.title }}
          </span>
        </h3>
      </NuxtLink>
    </div>

    <!-- Arrow icon -->
    <NuxtLink
      :to="`/article/${article.slug}`"
      class="absolute top-4 right-4 w-10 h-10 bg-white/90 backdrop-blur-sm rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transform -translate-y-2 group-hover:translate-y-0 transition-all duration-300 shadow-lg"
    >
      <UIcon name="i-heroicons-arrow-up-right" class="w-4 h-4 text-stone-900 group-hover:rotate-45 transition-transform duration-300" />
    </NuxtLink>
  </article>

  <!-- Horizontal Card -->
  <article
    v-else-if="variant === 'horizontal'"
    class="group flex gap-4 items-start p-2 -m-2 rounded-xl hover:bg-orange-50/50 dark:hover:bg-orange-950/20 transition-colors duration-300"
  >
    <NuxtLink :to="`/article/${article.slug}`" class="shrink-0 w-20 h-20 rounded-xl overflow-hidden bg-stone-100 dark:bg-stone-800 image-reveal shadow-sm">
      <NuxtImg
        v-if="article.featured_image"
        :src="article.featured_image"
        :alt="article.title"
        class="w-full h-full object-cover"
        loading="lazy"
      />
      <div v-else class="w-full h-full flex items-center justify-center bg-gradient-to-br from-orange-50 to-amber-50 dark:from-stone-800 dark:to-stone-900">
        <UIcon name="i-heroicons-newspaper" class="w-6 h-6 text-orange-300 dark:text-orange-700" />
      </div>
    </NuxtLink>
    <div class="flex-1 min-w-0">
      <NuxtLink
        v-if="article.category_slug && article.category_name"
        :to="`/category/${article.category_slug}`"
        class="text-xs font-semibold text-orange-600 dark:text-orange-400 hover:text-orange-500 transition-colors"
      >
        {{ article.category_name }}
      </NuxtLink>
      <NuxtLink :to="`/article/${article.slug}`">
        <h4 class="font-semibold text-stone-800 dark:text-stone-100 leading-snug line-clamp-2 group-hover:text-orange-600 dark:group-hover:text-orange-400 transition-colors duration-300 mt-0.5">
          {{ article.title }}
        </h4>
      </NuxtLink>
      <span v-if="article.published_at" class="text-xs text-stone-500 dark:text-stone-400 mt-1 block">
        {{ formatDate(article.published_at) }}
      </span>
    </div>
  </article>

  <!-- Default Card -->
  <article
    v-else
    class="group bg-white dark:bg-stone-900 rounded-2xl overflow-hidden border border-stone-200 dark:border-stone-800 card-lift"
  >
    <!-- Image -->
    <NuxtLink :to="`/article/${article.slug}`" class="block overflow-hidden aspect-[16/10] image-reveal relative">
      <NuxtImg
        v-if="article.featured_image"
        :src="article.featured_image"
        :alt="article.title"
        class="w-full h-full object-cover"
        loading="lazy"
      />
      <div v-else class="w-full h-full bg-gradient-to-br from-orange-50 to-amber-50 dark:from-stone-800 dark:to-stone-900 flex items-center justify-center">
        <UIcon name="i-heroicons-newspaper" class="w-12 h-12 text-orange-300 dark:text-orange-700" />
      </div>
      <!-- Subtle overlay on hover -->
      <div class="absolute inset-0 bg-orange-500/0 group-hover:bg-orange-500/10 transition-colors duration-300" />
    </NuxtLink>

    <!-- Content -->
    <div class="p-5">
      <div class="flex items-center gap-3 mb-3">
        <NuxtLink
          v-if="article.category_slug && article.category_name"
          :to="`/category/${article.category_slug}`"
          class="text-xs font-semibold text-orange-600 dark:text-orange-400 hover:text-orange-500 uppercase tracking-wide transition-colors"
        >
          {{ article.category_name }}
        </NuxtLink>
        <span v-if="article.category_name && article.published_at" class="text-stone-300 dark:text-stone-600">•</span>
        <span v-if="article.published_at" class="text-xs text-stone-500 dark:text-stone-400">
          {{ formatDate(article.published_at) }}
        </span>
      </div>

      <NuxtLink :to="`/article/${article.slug}`">
        <h3 class="text-xl font-bold text-stone-800 dark:text-stone-100 leading-snug line-clamp-2 group-hover:text-orange-600 dark:group-hover:text-orange-400 transition-colors duration-300">
          {{ article.title }}
        </h3>
      </NuxtLink>

      <p v-if="article.summary" class="mt-3 text-stone-600 dark:text-stone-400 text-sm line-clamp-2">
        {{ article.summary }}
      </p>

      <div v-if="article.author_name" class="mt-4 pt-4 border-t border-stone-100 dark:border-stone-800 flex items-center justify-between">
        <NuxtLink v-if="article.author_slug" :to="`/author/${article.author_slug}`" class="flex items-center gap-2 hover:opacity-80 transition-opacity">
          <UAvatar
            v-if="article.author_avatar"
            :src="article.author_avatar"
            :alt="article.author_name"
            size="sm"
          />
          <div v-else class="w-8 h-8 rounded-full bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center">
            <UIcon name="i-heroicons-user" class="w-4 h-4 text-orange-500" />
          </div>
          <span class="text-sm font-medium text-stone-700 dark:text-stone-300">{{ article.author_name }}</span>
        </NuxtLink>
        <div v-else class="flex items-center gap-2">
          <UAvatar
            v-if="article.author_avatar"
            :src="article.author_avatar"
            :alt="article.author_name"
            size="sm"
          />
          <div v-else class="w-8 h-8 rounded-full bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center">
            <UIcon name="i-heroicons-user" class="w-4 h-4 text-orange-500" />
          </div>
          <span class="text-sm font-medium text-stone-700 dark:text-stone-300">{{ article.author_name }}</span>
        </div>
        <div class="w-8 h-8 rounded-full bg-stone-100 dark:bg-stone-800 flex items-center justify-center opacity-0 group-hover:opacity-100 transform translate-x-2 group-hover:translate-x-0 transition-all duration-300">
          <UIcon name="i-heroicons-arrow-right" class="w-4 h-4 text-orange-500" />
        </div>
      </div>
    </div>
  </article>
</template>
