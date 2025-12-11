<script setup lang="ts">
import type { ArticleListItem, PaginatedArticles, ApiResponse } from '~/types'
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const page = ref(1)
const loading = ref(false)
const articles = ref<ArticleListItem[]>([])
const totalPages = ref(1)
const error = ref('')

const columns: TableColumn<ArticleListItem>[] = [
  { accessorKey: 'title', header: 'Title' },
  { accessorKey: 'category_name', header: 'Category' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'published_at', header: 'Published' },
  { id: 'actions', header: '' }
]

async function fetchArticles() {
  loading.value = true
  error.value = ''

  try {
    const response = await $fetch<ApiResponse<PaginatedArticles>>(`${baseUrl}/admin/articles?page=${page.value}&per_page=20`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      articles.value = response.data.articles
      totalPages.value = response.data.total_pages
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load articles'
  }

  loading.value = false
}

async function deleteArticle(id: string) {
  if (!confirm('Are you sure you want to delete this article?')) return

  try {
    await $fetch(`${baseUrl}/admin/articles/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchArticles()
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    alert(err?.data?.error?.message || 'Failed to delete article')
  }
}

function getStatusColor(status: string) {
  switch (status) {
    case 'published': return 'success'
    case 'draft': return 'warning'
    case 'archived': return 'neutral'
    default: return 'neutral'
  }
}
onMounted(fetchArticles)
watch(page, fetchArticles)

useSeoMeta({
  title: 'Articles - Pulpulitiko Admin'
})

</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Articles</h1>
      <UButton to="/admin/articles/new" icon="i-heroicons-plus">
        New Article
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <UCard>
      <div v-if="loading" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      </div>

      <UTable
        v-else-if="articles.length"
        :data="articles"
        :columns="columns"
      >
        <template #title-cell="{ row }">
          <NuxtLink :to="`/admin/articles/${row.original.id}`" class="block min-w-0">
            <p class="font-medium text-gray-900 dark:text-white truncate hover:text-primary">{{ row.original.title }}</p>
            <p class="text-sm text-gray-500 truncate">{{ row.original.slug }}</p>
          </NuxtLink>
        </template>

        <template #category_name-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400">
            {{ row.original.category_name || '-' }}
          </span>
        </template>

        <template #status-cell="{ row }">
          <UBadge :color="getStatusColor(row.original.status)" variant="subtle">
            {{ row.original.status }}
          </UBadge>
        </template>

        <template #published_at-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400 text-sm whitespace-nowrap">
            {{ row.original.published_at ? new Date(row.original.published_at).toLocaleDateString() : '-' }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-2 justify-end">
            <UButton
              :to="`/admin/articles/${row.original.id}`"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
            />
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-trash"
              @click="deleteArticle(row.original.id)"
            />
          </div>
        </template>
      </UTable>

      <div v-else class="py-8 text-center text-gray-500">
        No articles yet.
        <NuxtLink to="/admin/articles/new" class="text-primary hover:underline">Create one</NuxtLink>
      </div>

      <template v-if="totalPages > 1" #footer>
        <div class="flex justify-center">
          <UPagination v-model:page="page" :total="totalPages * 20" :items-per-page="20" />
        </div>
      </template>
    </UCard>
  </div>
</template>
