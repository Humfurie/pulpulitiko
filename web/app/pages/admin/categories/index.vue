<script setup lang="ts">
import type { Category, PaginatedCategories, ApiResponse } from '~/types'
import type { TableColumn } from '@nuxt/ui'
import { useDebounceFn } from '@vueuse/core'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const page = ref(1)
const loading = ref(false)
const categories = ref<Category[]>([])
const total = ref(0)
const totalPages = ref(1)
const error = ref('')
const search = ref('')
const sortBy = ref('name')
const sortOrder = ref('asc')

const columns: TableColumn<Category>[] = [
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'slug', header: 'Slug' },
  { accessorKey: 'description', header: 'Description' },
  { id: 'actions', header: 'Actions' }
]

async function fetchCategories() {
  loading.value = true
  error.value = ''

  try {
    const params = new URLSearchParams({
      page: String(page.value),
      per_page: '20',
      sort_by: sortBy.value,
      sort_order: sortOrder.value
    })

    if (search.value) {
      params.append('search', search.value)
    }

    const response = await $fetch<ApiResponse<PaginatedCategories>>(`${baseUrl}/admin/categories?${params}`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      categories.value = response.data.categories
      total.value = response.data.total
      totalPages.value = response.data.total_pages
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load categories'
  }

  loading.value = false
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchCategories()
}, 300)

async function deleteCategory(id: string) {
  if (!confirm('Are you sure you want to delete this category?')) return

  try {
    await $fetch(`${baseUrl}/admin/categories/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchCategories()
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    alert(err?.data?.error?.message || 'Failed to delete category')
  }
}

onMounted(fetchCategories)
watch(page, fetchCategories)
watch(search, debouncedSearch)
watch(sortBy, fetchCategories)
watch(sortOrder, fetchCategories)

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
      <template #header>
        <div class="flex flex-col sm:flex-row gap-4">
          <UInput
            v-model="search"
            placeholder="Search categories..."
            icon="i-heroicons-magnifying-glass"
            class="flex-1"
          />
          <div class="flex gap-2">
            <USelect
              v-model="sortBy"
              :options="[
                { label: 'Name', value: 'name' },
                { label: 'Slug', value: 'slug' },
                { label: 'Created', value: 'created_at' }
              ]"
              class="w-32"
            />
            <USelect
              v-model="sortOrder"
              :options="[
                { label: 'Asc', value: 'asc' },
                { label: 'Desc', value: 'desc' }
              ]"
              class="w-24"
            />
          </div>
        </div>
        <div v-if="total > 0" class="text-sm text-gray-500 mt-2">
          {{ total }} categor{{ total !== 1 ? 'ies' : 'y' }} total
        </div>
      </template>

      <div v-if="loading" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
        <p class="mt-2 text-gray-500">Loading categories...</p>
      </div>

      <UTable
        v-else-if="categories.length"
        :data="categories"
        :columns="columns"
      >
        <template #name-cell="{ row }">
          <NuxtLink :to="`/admin/categories/${row.original.id}`" class="block">
            <p class="font-medium text-gray-900 dark:text-white hover:text-primary">{{ row.original.name }}</p>
          </NuxtLink>
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
            >
              Edit
            </UButton>
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-trash"
              @click="deleteCategory(row.original.id)"
            >
              Delete
            </UButton>
          </div>
        </template>
      </UTable>

      <div v-else class="py-12 text-center">
        <UIcon name="i-heroicons-folder" class="size-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
        <p class="text-gray-500 dark:text-gray-400 mb-4">No categories yet.</p>
        <UButton to="/admin/categories/new" variant="outline">
          Create your first category
        </UButton>
      </div>

      <template v-if="totalPages > 1" #footer>
        <div class="flex justify-center">
          <UPagination v-model:page="page" :total="totalPages * 20" :items-per-page="20" />
        </div>
      </template>
    </UCard>
  </div>
</template>
