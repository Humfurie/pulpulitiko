<script setup lang="ts">
import type { ApiResponse, DashboardMetrics } from '~/types'

definePageMeta({
  layout: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const loading = ref(true)
const metrics = ref<DashboardMetrics | null>(null)

async function loadMetrics() {
  loading.value = true
  try {
    const res = await $fetch<ApiResponse<DashboardMetrics>>(`${baseUrl}/admin/metrics`, {
      headers: auth.getAuthHeaders()
    })

    if (res.success) {
      metrics.value = res.data
    }
  } catch (e) {
    console.error('Failed to load metrics', e)
  }
  loading.value = false
}

function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

onMounted(loadMetrics)

useSeoMeta({
  title: 'Dashboard - Pulpulitiko Admin'
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">Dashboard</h1>

    <!-- Stats Overview -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <div class="flex items-center gap-4">
          <div class="p-3 bg-primary-50 dark:bg-primary-900/20 rounded-lg">
            <UIcon name="i-heroicons-document-text" class="size-6 text-primary-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Articles</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              <span v-if="loading" class="inline-block w-8 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              <span v-else>{{ metrics?.total_articles || 0 }}</span>
            </p>
          </div>
        </div>
      </UCard>

      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <div class="flex items-center gap-4">
          <div class="p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
            <UIcon name="i-heroicons-eye" class="size-6 text-green-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Total Views</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              <span v-if="loading" class="inline-block w-8 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              <span v-else>{{ formatNumber(metrics?.total_views || 0) }}</span>
            </p>
          </div>
        </div>
      </UCard>

      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <div class="flex items-center gap-4">
          <div class="p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
            <UIcon name="i-heroicons-folder" class="size-6 text-blue-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Categories</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              <span v-if="loading" class="inline-block w-8 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              <span v-else>{{ metrics?.total_categories || 0 }}</span>
            </p>
          </div>
        </div>
      </UCard>

      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <div class="flex items-center gap-4">
          <div class="p-3 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
            <UIcon name="i-heroicons-tag" class="size-6 text-purple-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Tags</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              <span v-if="loading" class="inline-block w-8 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              <span v-else>{{ metrics?.total_tags || 0 }}</span>
            </p>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Top Articles & Quick Actions Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
      <!-- Top Articles -->
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-orange-50 dark:bg-orange-900/20">
              <UIcon name="i-heroicons-fire" class="size-5 text-orange-500" />
            </div>
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">Top Articles</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">Most viewed articles</p>
            </div>
          </div>
        </template>

        <div v-if="loading" class="space-y-3">
          <div v-for="i in 5" :key="i" class="flex items-center gap-3">
            <div class="w-6 h-6 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
            <div class="flex-1 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
            <div class="w-16 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
          </div>
        </div>

        <div v-else-if="metrics?.top_articles?.length" class="space-y-3">
          <div
            v-for="(article, index) in metrics.top_articles.slice(0, 5)"
            :key="article.id"
            class="flex items-center gap-3"
          >
            <span class="flex-shrink-0 w-6 h-6 flex items-center justify-center text-sm font-bold rounded-full"
              :class="{
                'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400': index === 0,
                'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400': index > 0
              }">
              {{ index + 1 }}
            </span>
            <div class="flex-1 min-w-0">
              <NuxtLink :to="`/admin/articles/${article.id}`" class="text-sm font-medium text-gray-900 dark:text-white hover:text-primary-500 truncate block">
                {{ article.title }}
              </NuxtLink>
              <span v-if="article.category_name" class="text-xs text-gray-500 dark:text-gray-400">
                {{ article.category_name }}
              </span>
            </div>
            <div class="flex items-center gap-1 text-sm text-gray-500 dark:text-gray-400">
              <UIcon name="i-heroicons-eye" class="size-4" />
              {{ formatNumber(article.view_count) }}
            </div>
          </div>
        </div>

        <div v-else class="text-center py-4 text-gray-500 dark:text-gray-400">
          No articles yet
        </div>
      </UCard>

      <!-- Quick Actions -->
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-emerald-50 dark:bg-emerald-900/20">
              <UIcon name="i-heroicons-rocket-launch" class="size-5 text-emerald-500" />
            </div>
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">Quick Actions</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">Common tasks and shortcuts</p>
            </div>
          </div>
        </template>

        <div class="flex flex-wrap gap-3">
          <UButton to="/admin/articles/new" icon="i-heroicons-plus">
            New Article
          </UButton>
          <UButton to="/admin/articles" variant="soft" color="neutral" icon="i-heroicons-document-text">
            Manage Articles
          </UButton>
          <UButton to="/admin/categories/new" variant="soft" color="neutral" icon="i-heroicons-folder-plus">
            New Category
          </UButton>
          <UButton to="/admin/tags/new" variant="soft" color="neutral" icon="i-heroicons-tag">
            New Tag
          </UButton>
        </div>

        <template #footer>
          <p class="text-sm text-gray-600 dark:text-gray-400">
            Logged in as <strong class="text-gray-900 dark:text-white">{{ auth.user.value?.name || 'User' }}</strong>
            <UBadge :color="auth.user.value?.role === 'admin' ? 'primary' : 'neutral'" variant="soft" size="sm" class="ml-2">
              {{ auth.user.value?.role || 'user' }}
            </UBadge>
          </p>
        </template>
      </UCard>
    </div>

    <!-- Category & Tag Metrics Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Category Metrics -->
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
              <UIcon name="i-heroicons-chart-bar" class="size-5 text-blue-500" />
            </div>
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">Categories Performance</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">Articles and views by category</p>
            </div>
          </div>
        </template>

        <div v-if="loading" class="space-y-4">
          <div v-for="i in 4" :key="i" class="space-y-2">
            <div class="flex justify-between">
              <div class="w-24 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              <div class="w-16 h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
            </div>
            <div class="h-2 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
          </div>
        </div>

        <div v-else-if="metrics?.category_metrics?.length" class="space-y-4">
          <div v-for="category in metrics.category_metrics" :key="category.id">
            <div class="flex justify-between items-center mb-1">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ category.name }}</span>
              <div class="flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
                <span>{{ category.article_count }} articles</span>
                <span class="flex items-center gap-1">
                  <UIcon name="i-heroicons-eye" class="size-3" />
                  {{ formatNumber(category.total_views) }}
                </span>
              </div>
            </div>
            <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
              <div
                class="bg-blue-500 h-2 rounded-full transition-all duration-300"
                :style="{ width: `${metrics.total_views > 0 ? (category.total_views / metrics.total_views) * 100 : 0}%` }"
              />
            </div>
          </div>
        </div>

        <div v-else class="text-center py-4 text-gray-500 dark:text-gray-400">
          No categories yet
        </div>
      </UCard>

      <!-- Tag Metrics -->
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-purple-50 dark:bg-purple-900/20">
              <UIcon name="i-heroicons-hashtag" class="size-5 text-purple-500" />
            </div>
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">Tags Performance</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">Articles and views by tag</p>
            </div>
          </div>
        </template>

        <div v-if="loading" class="flex flex-wrap gap-2">
          <div v-for="i in 8" :key="i" class="w-20 h-8 bg-gray-200 dark:bg-gray-700 rounded-full animate-pulse" />
        </div>

        <div v-else-if="metrics?.tag_metrics?.length" class="flex flex-wrap gap-2">
          <div
            v-for="tag in metrics.tag_metrics.slice(0, 12)"
            :key="tag.id"
            class="inline-flex items-center gap-2 px-3 py-1.5 bg-purple-50 dark:bg-purple-900/20 rounded-full"
          >
            <span class="text-sm font-medium text-purple-700 dark:text-purple-300">{{ tag.name }}</span>
            <span class="text-xs text-purple-500 dark:text-purple-400">
              {{ tag.article_count }} / {{ formatNumber(tag.total_views) }}
            </span>
          </div>
        </div>

        <div v-else class="text-center py-4 text-gray-500 dark:text-gray-400">
          No tags yet
        </div>

        <template #footer v-if="metrics?.tag_metrics && metrics.tag_metrics.length > 12">
          <NuxtLink to="/admin/tags" class="text-sm text-primary-500 hover:underline">
            View all {{ metrics.tag_metrics.length }} tags
          </NuxtLink>
        </template>
      </UCard>
    </div>
  </div>
</template>
