<script setup lang="ts">
import type { Category, ApiResponse, UpdateCategoryRequest } from '~/types'

definePageMeta({
  layout: 'admin'
})

const route = useRoute()
const router = useRouter()
const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const categoryId = route.params.id as string

const loading = ref(false)
const saving = ref(false)
const error = ref('')

const form = reactive<UpdateCategoryRequest>({
  name: '',
  slug: '',
  description: ''
})

async function loadCategory() {
  loading.value = true
  try {
    const response = await $fetch<ApiResponse<Category>>(`${baseUrl}/admin/categories/${categoryId}`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      const category = response.data
      form.name = category.name
      form.slug = category.slug
      form.description = category.description || ''
    }
  } catch (e: any) {
    error.value = e?.data?.error?.message || 'Failed to load category'
  }
  loading.value = false
}

async function handleSubmit() {
  saving.value = true
  error.value = ''

  try {
    const payload = {
      ...form,
      description: form.description || undefined
    }

    await $fetch(`${baseUrl}/admin/categories/${categoryId}`, {
      method: 'PUT',
      headers: auth.getAuthHeaders(),
      body: payload
    })

    await router.push('/admin/categories')
  } catch (e: any) {
    error.value = e?.data?.error?.message || 'Failed to update category'
  }

  saving.value = false
}

onMounted(loadCategory)

useSeoMeta({
  title: 'Edit Category - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-8">
      <UButton
        to="/admin/categories"
        variant="soft"
        color="neutral"
        icon="i-heroicons-arrow-left"
        size="md"
        class="shadow-sm"
      />
      <div>
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">Edit Category</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Update category details</p>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="py-16 text-center">
      <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-blue-50 dark:bg-blue-900/20 mb-4">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-blue-500" />
      </div>
      <p class="text-gray-500 dark:text-gray-400">Loading category...</p>
    </div>

    <form v-else @submit.prevent="handleSubmit">
      <UAlert v-if="error" color="error" :title="error" class="mb-6" icon="i-heroicons-exclamation-circle" />

      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
              <UIcon name="i-heroicons-folder" class="size-5 text-blue-500" />
            </div>
            <div>
              <h3 class="font-semibold text-gray-900 dark:text-white">Category Details</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">Basic category information</p>
            </div>
          </div>
        </template>

        <div class="space-y-6">
          <UFormField label="Name" name="name" required class="w-full">
            <template #hint>
              <span class="text-xs text-gray-400">Display name for the category</span>
            </template>
            <UInput
              v-model="form.name"
              placeholder="Enter category name"
              size="lg"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Slug" name="slug" required class="w-full">
            <template #hint>
              <span class="text-xs text-gray-400">URL-friendly identifier</span>
            </template>
            <div class="flex items-center w-full">
              <span class="inline-flex items-center px-3 py-2 text-sm text-gray-500 bg-gray-50 dark:bg-gray-800 border border-r-0 border-gray-300 dark:border-gray-700 rounded-l-lg shrink-0">
                /category/
              </span>
              <UInput
                v-model="form.slug"
                placeholder="category-slug"
                class="rounded-l-none flex-1 min-w-0"
              />
            </div>
          </UFormField>

          <UFormField label="Description" name="description" class="w-full">
            <template #hint>
              <span class="text-xs text-gray-400">Optional description for the category</span>
            </template>
            <UTextarea
              v-model="form.description"
              placeholder="Write a brief description of this category..."
              :rows="4"
              autoresize
              class="w-full"
            />
          </UFormField>
        </div>

        <template #footer>
          <div class="flex flex-col-reverse sm:flex-row sm:justify-end gap-3">
            <UButton to="/admin/categories" variant="soft" color="neutral">
              Cancel
            </UButton>
            <UButton type="submit" :loading="saving" size="lg">
              <UIcon v-if="!saving" name="i-heroicons-check" class="size-4 mr-2" />
              {{ saving ? 'Saving...' : 'Update Category' }}
            </UButton>
          </div>
        </template>
      </UCard>
    </form>
  </div>
</template>
