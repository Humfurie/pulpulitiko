<script setup lang="ts">
import type { Author, PaginatedUsers, ApiResponse } from '~/types'
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
const users = ref<Author[]>([])
const total = ref(0)
const totalPages = ref(1)
const error = ref('')
const search = ref('')
const roleFilter = ref<string>('')
const sortBy = ref('name')
const sortOrder = ref('asc')

const columns = [
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'email', header: 'Email' },
  { accessorKey: 'phone', header: 'Phone' },
  { accessorKey: 'role', header: 'Role' },
  { accessorKey: 'created_at', header: 'Created' },
  { id: 'actions', header: 'Actions' }
]

async function fetchUsers() {
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

    if (roleFilter.value) {
      params.append('role', roleFilter.value)
    }

    const response = await $fetch<ApiResponse<PaginatedUsers>>(`${baseUrl}/admin/users?${params}`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      users.value = response.data.users
      total.value = response.data.total
      totalPages.value = response.data.total_pages
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load users'
  }

  loading.value = false
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchUsers()
}, 300)

async function deleteUser(id: string) {
  if (!confirm('Are you sure you want to delete this user?')) return

  try {
    await $fetch(`${baseUrl}/admin/users/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchUsers()
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    alert(err?.data?.error?.message || 'Failed to delete user')
  }
}

function getRoleColor(role: string) {
  switch (role) {
    case 'admin': return 'error'
    case 'author': return 'primary'
    case 'user': return 'neutral'
    default: return 'neutral'
  }
}

onMounted(fetchUsers)
watch(page, fetchUsers)
watch(search, debouncedSearch)
watch(roleFilter, () => {
  page.value = 1
  fetchUsers()
})
watch(sortBy, fetchUsers)
watch(sortOrder, fetchUsers)

useSeoMeta({
  title: 'Users - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
      <div>
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">Users</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Manage user accounts and roles</p>
      </div>
      <UButton to="/admin/users/new" icon="i-heroicons-plus" size="lg">
        New User
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-6" icon="i-heroicons-exclamation-circle" />

    <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
      <template #header>
        <div class="flex flex-col sm:flex-row gap-4">
          <UInput
            v-model="search"
            placeholder="Search users..."
            icon="i-heroicons-magnifying-glass"
            class="flex-1"
          />
          <div class="flex gap-2">
            <USelect
              v-model="roleFilter"
              :options="[
                { label: 'All Roles', value: '' },
                { label: 'Admin', value: 'admin' },
                { label: 'Author', value: 'author' },
                { label: 'User', value: 'user' }
              ]"
              class="w-32"
            />
            <USelect
              v-model="sortBy"
              :options="[
                { label: 'Name', value: 'name' },
                { label: 'Email', value: 'email' },
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
          {{ total }} user{{ total !== 1 ? 's' : '' }} total
        </div>
      </template>

      <div v-if="loading" class="py-16 text-center">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary-50 dark:bg-primary-900/20 mb-4">
          <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary-500" />
        </div>
        <p class="text-gray-500 dark:text-gray-400">Loading users...</p>
      </div>

      <UTable
        v-else-if="users.length"
        :data="users"
        :columns="columns"
        class="flex-1"
      >
        <template #name-cell="{ row }">
          <NuxtLink :to="`/admin/users/${row.original.id}`" class="flex items-center gap-3">
            <UAvatar
              v-if="row.original.avatar"
              :src="row.original.avatar"
              :alt="row.original.name"
              size="sm"
            />
            <UAvatar v-else :alt="row.original.name" size="sm" />
            <div>
              <p class="font-medium text-gray-900 dark:text-white hover:text-primary">{{ row.original.name }}</p>
              <p class="text-sm text-gray-500">{{ row.original.slug }}</p>
            </div>
          </NuxtLink>
        </template>

        <template #email-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400">
            {{ row.original.email || '-' }}
          </span>
        </template>

        <template #phone-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400">
            {{ row.original.phone || '-' }}
          </span>
        </template>

        <template #role-cell="{ row }">
          <UBadge :color="getRoleColor(row.original.role)" variant="subtle" class="capitalize">
            {{ row.original.role }}
          </UBadge>
        </template>

        <template #created_at-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400 text-sm">
            {{ new Date(row.original.created_at).toLocaleDateString() }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-2 justify-end">
            <UButton
              :to="`/admin/users/${row.original.id}`"
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
              @click="deleteUser(row.original.id)"
            >
              Delete
            </UButton>
          </div>
        </template>
      </UTable>

      <div v-else class="py-16 text-center">
        <UIcon name="i-heroicons-users" class="size-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
        <p class="text-gray-500 dark:text-gray-400 mb-4">No users found.</p>
        <UButton to="/admin/users/new" variant="soft">
          Create your first user
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
