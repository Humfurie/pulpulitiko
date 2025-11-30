<script setup lang="ts">
import type { Category, PaginatedCategories, ApiResponse } from '~/types'
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const page = ref(1)
const loading = ref(false)
const categories = ref<Category[]>([])
const totalPages = ref(1)
const error = ref('')

const columns: TableColumn<Category>[] = [
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'slug', header: 'Slug' },
  { accessorKey: 'description', header: 'Description' },
  { id: 'actions', header: '' }
]

async function fetchCategories() {
  loading.value = true
  error.value = ''

  try {
    const response = await $fetch<ApiResponse<PaginatedCategories>>(`${baseUrl}/admin/categories?page=${page.value}&per_page=20`, {
      headers: auth.getAuthHeaders()
    })

    console.log('Categories API Response:', response)
    if (response.success) {
      console.log('Categories data:', response.data.categories)
      categories.value = response.data.categories
      totalPages.value = response.data.total_pages
    }
  } catch (e: any) {
    error.value = e?.data?.error?.message || 'Failed to load categories'
  }

  loading.value = false
}

async function deleteCategory(id: string) {
  if (!confirm('Are you sure you want to delete this category?')) return

  try {
    await $fetch(`${baseUrl}/admin/categories/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchCategories()
  } catch (e: any) {
    alert(e?.data?.error?.message || 'Failed to delete category')
  }
}

onMounted(fetchCategories)
watch(page, fetchCategories)

useSeoMeta({
  title: 'Categories - Pulpulitiko Admin'
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Categories</h1>
      <UButton to="/admin/categories/new" icon="i-heroicons-plus">
        New Category
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <UCard>
      <div v-if="loading" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      </div>

      <UTable
        v-else-if="categories.length"
        :data="categories"
        :columns="columns"
      >
        <template #name-cell="{ row }">
          <p class="font-medium text-gray-900 dark:text-white">{{ row.original.name }}</p>
        </template>

        <template #slug-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400 text-sm">
            {{ row.original.slug }}
          </span>
        </template>

        <template #description-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400 text-sm">
            {{ row.original.description || '-' }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-2 justify-end">
            <UButton
              :to="`/admin/categories/${row.original.id}`"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
            />
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-trash"
              @click="deleteCategory(row.original.id)"
            />
          </div>
        </template>
      </UTable>

      <div v-else class="py-8 text-center text-gray-500">
        No categories yet.
        <NuxtLink to="/admin/categories/new" class="text-primary hover:underline">Create one</NuxtLink>
      </div>

      <template #footer v-if="totalPages > 1">
        <div class="flex justify-center">
          <UPagination v-model:page="page" :total="totalPages * 20" :items-per-page="20" />
        </div>
      </template>
    </UCard>
  </div>
</template>
