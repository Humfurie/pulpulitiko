<script setup lang="ts">
import type { Tag, ApiResponse, UpdateTagRequest } from '~/types'

definePageMeta({
  layout: 'admin'
})

const route = useRoute()
const router = useRouter()
const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const tagId = route.params.id as string

const loading = ref(false)
const saving = ref(false)
const error = ref('')

const form = reactive<UpdateTagRequest>({
  name: '',
  slug: ''
})

async function loadTag() {
  loading.value = true
  try {
    const response = await $fetch<ApiResponse<Tag>>(`${baseUrl}/admin/tags/${tagId}`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      const tag = response.data
      form.name = tag.name
      form.slug = tag.slug
    }
  } catch (e: any) {
    error.value = e?.data?.error?.message || 'Failed to load tag'
  }
  loading.value = false
}

async function handleSubmit() {
  saving.value = true
  error.value = ''

  try {
    await $fetch(`${baseUrl}/admin/tags/${tagId}`, {
      method: 'PUT',
      headers: auth.getAuthHeaders(),
      body: form
    })

    await router.push('/admin/tags')
  } catch (e: any) {
    error.value = e?.data?.error?.message || 'Failed to update tag'
  }

  saving.value = false
}

onMounted(loadTag)

useSeoMeta({
  title: 'Edit Tag - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-8">
      <UButton
        to="/admin/tags"
        variant="soft"
        color="neutral"
        icon="i-heroicons-arrow-left"
        size="md"
        class="shadow-sm"
      />
      <div>
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">Edit Tag</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Update tag details</p>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="py-16 text-center">
      <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-purple-50 dark:bg-purple-900/20 mb-4">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-purple-500" />
      </div>
      <p class="text-gray-500 dark:text-gray-400">Loading tag...</p>
    </div>

    <form v-else @submit.prevent="handleSubmit">
      <UAlert v-if="error" color="error" :title="error" class="mb-6" icon="i-heroicons-exclamation-circle" />

      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-purple-50 dark:bg-purple-900/20">
              <UIcon name="i-heroicons-tag" class="size-5 text-purple-500" />
            </div>
            <div>
              <h3 class="font-semibold text-gray-900 dark:text-white">Tag Details</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">Basic tag information</p>
            </div>
          </div>
        </template>

        <div class="space-y-6">
          <UFormField label="Name" name="name" required class="w-full">
            <template #hint>
              <span class="text-xs text-gray-400">Display name for the tag</span>
            </template>
            <UInput
              v-model="form.name"
              placeholder="Enter tag name"
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
                /tag/
              </span>
              <UInput
                v-model="form.slug"
                placeholder="tag-slug"
                class="rounded-l-none flex-1 min-w-0"
              />
            </div>
          </UFormField>
        </div>

        <template #footer>
          <div class="flex flex-col-reverse sm:flex-row sm:justify-end gap-3">
            <UButton to="/admin/tags" variant="soft" color="neutral">
              Cancel
            </UButton>
            <UButton type="submit" :loading="saving" size="lg">
              <UIcon v-if="!saving" name="i-heroicons-check" class="size-4 mr-2" />
              {{ saving ? 'Saving...' : 'Update Tag' }}
            </UButton>
          </div>
        </template>
      </UCard>
    </form>
  </div>
</template>
