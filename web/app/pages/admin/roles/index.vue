<script setup lang="ts">
import type { RoleWithPermissionCount, ApiResponse } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const loading = ref(false)
const roles = ref<RoleWithPermissionCount[]>([])
const error = ref('')

const columns = [
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'description', header: 'Description' },
  { accessorKey: 'permission_count', header: 'Permissions' },
  { accessorKey: 'is_system', header: 'Type' },
  { accessorKey: 'created_at', header: 'Created' },
  { id: 'actions', header: '' }
]

async function fetchRoles() {
  loading.value = true
  error.value = ''

  try {
    const response = await $fetch<ApiResponse<RoleWithPermissionCount[]>>(`${baseUrl}/admin/roles`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      roles.value = response.data || []
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load roles'
  }

  loading.value = false
}

async function deleteRole(id: string, isSystem: boolean) {
  if (isSystem) {
    alert('System roles cannot be deleted')
    return
  }

  if (!confirm('Are you sure you want to delete this role?')) return

  try {
    await $fetch(`${baseUrl}/admin/roles/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchRoles()
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    alert(err?.data?.error?.message || 'Failed to delete role')
  }
}

onMounted(fetchRoles)

useSeoMeta({
  title: 'Roles - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
      <div>
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">Roles</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Manage user roles and their permissions</p>
      </div>
      <UButton to="/admin/roles/new" icon="i-heroicons-plus" size="lg">
        New Role
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-6" icon="i-heroicons-exclamation-circle" />

    <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
      <div v-if="loading" class="py-16 text-center">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary-50 dark:bg-primary-900/20 mb-4">
          <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary-500" />
        </div>
        <p class="text-gray-500 dark:text-gray-400">Loading roles...</p>
      </div>

      <UTable
        v-else-if="roles.length"
        :data="roles"
        :columns="columns"
        class="flex-1"
      >
        <template #name-cell="{ row }">
          <div>
            <p class="font-medium text-gray-900 dark:text-white">{{ row.original.name }}</p>
            <p class="text-sm text-gray-500">{{ row.original.slug }}</p>
          </div>
        </template>

        <template #description-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400 text-sm">
            {{ row.original.description || '-' }}
          </span>
        </template>

        <template #permission_count-cell="{ row }">
          <UBadge color="primary" variant="subtle">
            {{ row.original.permission_count }} permissions
          </UBadge>
        </template>

        <template #is_system-cell="{ row }">
          <UBadge
            :color="row.original.is_system ? 'warning' : 'neutral'"
            variant="subtle"
          >
            {{ row.original.is_system ? 'System' : 'Custom' }}
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
              :to="`/admin/roles/${row.original.id}`"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
            />
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-trash"
              :disabled="row.original.is_system"
              @click="deleteRole(row.original.id, row.original.is_system)"
            />
          </div>
        </template>
      </UTable>

      <div v-else class="py-16 text-center">
        <UIcon name="i-heroicons-shield-check" class="size-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
        <p class="text-gray-500 dark:text-gray-400 mb-4">No roles yet.</p>
        <UButton to="/admin/roles/new" variant="soft">
          Create your first role
        </UButton>
      </div>
    </UCard>
  </div>
</template>
