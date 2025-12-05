<script setup lang="ts">
import type { PoliticianListItem, PaginatedPoliticians, ApiResponse } from '~/types'
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
const politicians = ref<PoliticianListItem[]>([])
const total = ref(0)
const totalPages = ref(1)
const error = ref('')
const search = ref('')

const columns: TableColumn<PoliticianListItem>[] = [
  { accessorKey: 'photo', header: '' },
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'position', header: 'Position' },
  { accessorKey: 'party', header: 'Party' },
  { accessorKey: 'article_count', header: 'Articles' },
  { id: 'actions', header: '' }
]

async function fetchPoliticians() {
  loading.value = true
  error.value = ''

  try {
    const params = new URLSearchParams({
      page: String(page.value),
      per_page: '20'
    })
    if (search.value) {
      params.append('search', search.value)
    }

    const response = await $fetch<ApiResponse<PaginatedPoliticians>>(`${baseUrl}/admin/politicians?${params}`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      politicians.value = response.data.politicians
      total.value = response.data.total
      totalPages.value = response.data.total_pages
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load politicians'
  }

  loading.value = false
}

async function deletePolitician(id: string) {
  if (!confirm('Are you sure you want to delete this politician?')) return

  try {
    await $fetch(`${baseUrl}/admin/politicians/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchPoliticians()
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    alert(err?.data?.error?.message || 'Failed to delete politician')
  }
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchPoliticians()
}, 300)

onMounted(fetchPoliticians)
watch(page, fetchPoliticians)
watch(search, debouncedSearch)

useSeoMeta({
  title: 'Politicians - Pulpulitiko Admin'
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Politicians</h1>
      <UButton to="/admin/politicians/new" icon="i-heroicons-plus">
        New Politician
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <UCard>
      <template #header>
        <div class="flex items-center gap-4">
          <UInput
            v-model="search"
            placeholder="Search politicians..."
            icon="i-heroicons-magnifying-glass"
            class="w-64"
          />
          <span class="text-sm text-gray-500">{{ total }} politicians</span>
        </div>
      </template>

      <div v-if="loading" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      </div>

      <UTable
        v-else-if="politicians.length"
        :data="politicians"
        :columns="columns"
      >
        <template #photo-cell="{ row }">
          <UAvatar
            :src="row.original.photo"
            :alt="row.original.name"
            size="sm"
          />
        </template>

        <template #name-cell="{ row }">
          <div>
            <p class="font-medium text-gray-900 dark:text-white">{{ row.original.name }}</p>
            <p class="text-sm text-gray-500">{{ row.original.slug }}</p>
          </div>
        </template>

        <template #position-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400 text-sm">
            {{ row.original.position || '-' }}
          </span>
        </template>

        <template #party-cell="{ row }">
          <UBadge v-if="row.original.party" variant="subtle">
            {{ row.original.party }}
          </UBadge>
          <span v-else class="text-gray-400">-</span>
        </template>

        <template #article_count-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400">
            {{ row.original.article_count }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-2 justify-end">
            <UButton
              :to="`/admin/politicians/${row.original.id}`"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
            />
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-trash"
              @click="deletePolitician(row.original.id)"
            />
          </div>
        </template>
      </UTable>

      <div v-else class="py-8 text-center text-gray-500">
        No politicians yet.
        <NuxtLink to="/admin/politicians/new" class="text-primary hover:underline">Create one</NuxtLink>
      </div>

      <template v-if="totalPages > 1" #footer>
        <div class="flex justify-center">
          <UPagination v-model:page="page" :total="total" :items-per-page="20" />
        </div>
      </template>
    </UCard>
  </div>
</template>
