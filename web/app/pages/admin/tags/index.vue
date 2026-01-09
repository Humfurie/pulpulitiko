<script setup lang="ts">
import type { Tag, PaginatedTags, ApiResponse } from '~/types'
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
const tags = ref<Tag[]>([])
const total = ref(0)
const totalPages = ref(1)
const error = ref('')
const search = ref('')
const sortBy = ref('name')
const sortOrder = ref('asc')

async function fetchTags() {
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

    const response = await $fetch<ApiResponse<PaginatedTags>>(`${baseUrl}/admin/tags?${params}`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      tags.value = response.data.tags
      total.value = response.data.total
      totalPages.value = response.data.total_pages
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load tags'
  }

  loading.value = false
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchTags()
}, 300)

async function deleteTag(id: string) {
  if (!confirm('Are you sure you want to delete this tag?')) return

  try {
    await $fetch(`${baseUrl}/admin/tags/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchTags()
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    alert(err?.data?.error?.message || 'Failed to delete tag')
  }
}

onMounted(fetchTags)
watch(page, fetchTags)
watch(search, debouncedSearch)
watch(sortBy, fetchTags)
watch(sortOrder, fetchTags)

useSeoMeta({
  title: 'Tags - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-4xl mx-auto">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Tags</h1>
      <UButton to="/admin/tags/new" icon="i-heroicons-plus">
        New Tag
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <UCard>
      <template #header>
        <div class="flex flex-col sm:flex-row gap-4">
          <UInput
            v-model="search"
            placeholder="Search tags..."
            icon="i-heroicons-magnifying-glass"
            class="flex-1"
          />
          <div class="flex gap-2">
            <USelect
              v-model="sortBy"
              :options="[
                { label: 'Name', value: 'name' },
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
          {{ total }} tag{{ total !== 1 ? 's' : '' }} total
        </div>
      </template>
      <div v-if="loading" class="py-12 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
        <p class="mt-2 text-gray-500">Loading tags...</p>
      </div>

      <div v-else-if="tags.length" class="divide-y divide-gray-200 dark:divide-gray-800">
        <div
          v-for="tag in tags"
          :key="tag.id"
          class="flex items-center justify-between py-4 first:pt-0 last:pb-0"
        >
          <div class="flex items-center gap-3">
            <div class="flex items-center justify-center w-10 h-10 rounded-lg bg-primary-50 dark:bg-primary-900/20">
              <UIcon name="i-heroicons-tag" class="text-primary-500" />
            </div>
            <NuxtLink :to="`/admin/tags/${tag.id}`" class="block">
              <p class="font-medium text-gray-900 dark:text-white hover:text-primary">{{ tag.name }}</p>
              <p class="text-sm text-gray-500">{{ tag.slug }}</p>
            </NuxtLink>
          </div>
          <div class="flex items-center gap-2">
            <UButton
              :to="`/admin/tags/${tag.id}`"
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
              @click="deleteTag(tag.id)"
            >
              Delete
            </UButton>
          </div>
        </div>
      </div>

      <div v-else class="py-12 text-center">
        <UIcon name="i-heroicons-tag" class="size-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
        <p class="text-gray-500 dark:text-gray-400 mb-4">No tags yet.</p>
        <UButton to="/admin/tags/new" variant="outline">
          Create your first tag
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
