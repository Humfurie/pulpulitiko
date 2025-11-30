<script setup lang="ts">
import type { PermissionCategory, ApiResponse, CreateRoleRequest } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const router = useRouter()
const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const permissionCategories = ref<PermissionCategory[]>([])

const form = reactive({
  name: '',
  slug: '',
  description: '',
  permission_ids: [] as string[]
})

function generateSlug() {
  form.slug = form.name
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .trim()
}

async function loadPermissions() {
  loading.value = true
  try {
    const response = await $fetch<ApiResponse<PermissionCategory[]>>(`${baseUrl}/admin/roles/permissions`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      permissionCategories.value = response.data || []
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load permissions'
  }
  loading.value = false
}

function togglePermission(permId: string) {
  const index = form.permission_ids.indexOf(permId)
  if (index === -1) {
    form.permission_ids.push(permId)
  } else {
    form.permission_ids.splice(index, 1)
  }
}

function toggleCategory(category: PermissionCategory) {
  const categoryPermIds = category.permissions.map(p => p.id)
  const allSelected = categoryPermIds.every(id => form.permission_ids.includes(id))

  if (allSelected) {
    // Remove all permissions from this category
    form.permission_ids = form.permission_ids.filter(id => !categoryPermIds.includes(id))
  } else {
    // Add all permissions from this category
    categoryPermIds.forEach(id => {
      if (!form.permission_ids.includes(id)) {
        form.permission_ids.push(id)
      }
    })
  }
}

function isCategoryFullySelected(category: PermissionCategory) {
  return category.permissions.every(p => form.permission_ids.includes(p.id))
}

function isCategoryPartiallySelected(category: PermissionCategory) {
  const selected = category.permissions.filter(p => form.permission_ids.includes(p.id)).length
  return selected > 0 && selected < category.permissions.length
}

async function handleSubmit() {
  saving.value = true
  error.value = ''

  try {
    const payload: CreateRoleRequest = {
      name: form.name,
      slug: form.slug,
      description: form.description || undefined,
      permission_ids: form.permission_ids
    }

    await $fetch(`${baseUrl}/admin/roles`, {
      method: 'POST',
      headers: auth.getAuthHeaders(),
      body: payload
    })

    await router.push('/admin/roles')
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to create role'
  }

  saving.value = false
}

onMounted(loadPermissions)

useSeoMeta({
  title: 'New Role - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
      <div class="flex items-center gap-4">
        <UButton
          to="/admin/roles"
          variant="soft"
          color="neutral"
          icon="i-heroicons-arrow-left"
          size="md"
          class="shadow-sm"
        />
        <div>
          <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">New Role</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Create a new role with custom permissions</p>
        </div>
      </div>
      <div class="sm:hidden">
        <UButton type="submit" form="role-form" :loading="saving" class="w-full" size="lg">
          Create Role
        </UButton>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="py-16 text-center">
      <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary-50 dark:bg-primary-900/20 mb-4">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-primary-500" />
      </div>
      <p class="text-gray-500 dark:text-gray-400">Loading permissions...</p>
    </div>

    <template v-else>
      <UAlert v-if="error" color="error" :title="error" class="mb-8" icon="i-heroicons-exclamation-circle" />

      <form id="role-form" @submit.prevent="handleSubmit">
        <div class="space-y-8">
          <!-- Basic Information -->
          <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
            <template #header>
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-lg bg-primary-50 dark:bg-primary-900/20">
                  <UIcon name="i-heroicons-shield-check" class="size-5 text-primary-500" />
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900 dark:text-white">Basic Information</h3>
                  <p class="text-sm text-gray-500 dark:text-gray-400">Role name and description</p>
                </div>
              </div>
            </template>

            <div class="space-y-6">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <UFormField label="Name" name="name" required class="w-full">
                  <UInput
                    v-model="form.name"
                    placeholder="Editor"
                    size="lg"
                    @blur="!form.slug && generateSlug()"
                  />
                </UFormField>

                <UFormField label="Slug" name="slug" required class="w-full">
                  <template #hint>
                    <span class="text-xs text-gray-400">URL-friendly identifier</span>
                  </template>
                  <UInput v-model="form.slug" placeholder="editor" />
                </UFormField>
              </div>

              <UFormField label="Description" name="description" class="w-full">
                <UTextarea
                  v-model="form.description"
                  placeholder="Describe this role's purpose..."
                  :rows="2"
                  autoresize
                />
              </UFormField>
            </div>
          </UCard>

          <!-- Permissions -->
          <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
            <template #header>
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
                  <UIcon name="i-heroicons-key" class="size-5 text-blue-500" />
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900 dark:text-white">Permissions</h3>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    Select what this role can do ({{ form.permission_ids.length }} selected)
                  </p>
                </div>
              </div>
            </template>

            <div class="space-y-6">
              <div
                v-for="category in permissionCategories"
                :key="category.category"
                class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden"
              >
                <div
                  class="bg-gray-50 dark:bg-gray-800/50 px-4 py-3 flex items-center gap-3 cursor-pointer"
                  @click="toggleCategory(category)"
                >
                  <UCheckbox
                    :model-value="isCategoryFullySelected(category)"
                    :indeterminate="isCategoryPartiallySelected(category)"
                    @update:model-value="toggleCategory(category)"
                    @click.stop
                  />
                  <span class="font-medium text-gray-900 dark:text-white capitalize">
                    {{ category.category }}
                  </span>
                  <UBadge variant="subtle" color="neutral" size="xs">
                    {{ category.permissions.filter(p => form.permission_ids.includes(p.id)).length }}/{{ category.permissions.length }}
                  </UBadge>
                </div>
                <div class="p-4 grid grid-cols-1 md:grid-cols-2 gap-3">
                  <div
                    v-for="permission in category.permissions"
                    :key="permission.id"
                    class="flex items-start gap-3 p-2 rounded hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer"
                    @click="togglePermission(permission.id)"
                  >
                    <UCheckbox
                      :model-value="form.permission_ids.includes(permission.id)"
                      @update:model-value="togglePermission(permission.id)"
                      @click.stop
                    />
                    <div>
                      <p class="text-sm font-medium text-gray-900 dark:text-white">
                        {{ permission.name }}
                      </p>
                      <p v-if="permission.description" class="text-xs text-gray-500">
                        {{ permission.description }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </UCard>

          <!-- Submit Button -->
          <div class="hidden sm:flex justify-end">
            <UButton type="submit" :loading="saving" size="lg">
              Create Role
            </UButton>
          </div>
        </div>
      </form>
    </template>
  </div>
</template>
