<script setup lang="ts">
import type { Category, Tag, ApiResponse, CreateArticleRequest } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const api = useApi()
const router = useRouter()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const loading = ref(false)
const uploadingFeaturedImage = ref(false)
const error = ref('')
const featuredImageInput = ref<HTMLInputElement | null>(null)

const form = reactive<CreateArticleRequest>({
  slug: '',
  title: '',
  summary: '',
  content: '',
  featured_image: '',
  category_id: null as string | null,
  status: 'draft',
  tag_ids: []
})

const categories = ref<Category[]>([])
const tags = ref<Tag[]>([])

const categoryOptions = computed(() => [
  { label: 'No category', value: null },
  ...categories.value.map(c => ({ label: c.name, value: c.id }))
])

async function loadData() {
  try {
    const [catRes, tagRes] = await Promise.all([
      $fetch<ApiResponse<Category[]>>(`${baseUrl}/categories`),
      $fetch<ApiResponse<Tag[]>>(`${baseUrl}/tags`)
    ])
    if (catRes.success) categories.value = catRes.data
    if (tagRes.success) tags.value = tagRes.data
  } catch (e) {
    console.error('Failed to load categories/tags', e)
  }
}

function generateSlug() {
  form.slug = form.title
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .trim()
}

function toggleTag(tagId: string) {
  if (!form.tag_ids) form.tag_ids = []
  const index = form.tag_ids.indexOf(tagId)
  if (index === -1) {
    form.tag_ids.push(tagId)
  } else {
    form.tag_ids.splice(index, 1)
  }
}

async function uploadFeaturedImage(file: File) {
  if (!file.type.startsWith('image/')) {
    error.value = 'Please select an image file'
    return
  }

  const maxSize = 10 * 1024 * 1024 // 10MB
  if (file.size > maxSize) {
    error.value = 'Image size must be less than 10MB'
    return
  }

  uploadingFeaturedImage.value = true
  error.value = ''

  try {
    const result = await api.uploadFile(file, auth.getAuthHeaders())
    form.featured_image = result.url
  } catch (e: unknown) {
    const err = e as { message?: string }
    error.value = err.message || 'Failed to upload image'
  } finally {
    uploadingFeaturedImage.value = false
  }
}

function handleFeaturedImageSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    uploadFeaturedImage(file)
    target.value = ''
  }
}

function handleFeaturedImageDrop(event: DragEvent) {
  event.preventDefault()
  const file = event.dataTransfer?.files?.[0]
  if (file) {
    uploadFeaturedImage(file)
  }
}

function removeFeaturedImage() {
  form.featured_image = ''
}

async function handleSubmit() {
  loading.value = true
  error.value = ''

  try {
    const payload: Record<string, unknown> = {
      slug: form.slug,
      title: form.title,
      summary: form.summary || undefined,
      content: form.content,
      featured_image: form.featured_image || undefined,
      status: form.status
    }

    // Only include category_id if it's a valid non-empty value
    if (form.category_id && form.category_id !== 'null') {
      payload.category_id = form.category_id
    }

    // Only include tag_ids if there are any
    if (form.tag_ids?.length) {
      payload.tag_ids = form.tag_ids
    }

    await $fetch(`${baseUrl}/admin/articles`, {
      method: 'POST',
      headers: auth.getAuthHeaders(),
      body: payload
    })

    await router.push('/admin/articles')
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to create article'
  }

  loading.value = false
}

onMounted(loadData)

useSeoMeta({
  title: 'New Article - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <!-- Header with better visual hierarchy -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
      <div class="flex items-center gap-4">
        <UButton
          to="/admin/articles"
          variant="soft"
          color="neutral"
          icon="i-heroicons-arrow-left"
          size="md"
          class="shadow-sm"
        />
        <div>
          <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">New Article</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Create a new article for your blog</p>
        </div>
      </div>
      <!-- Mobile action button -->
      <div class="sm:hidden">
        <UButton type="submit" form="article-form" :loading="loading" class="w-full" size="lg">
          <UIcon name="i-heroicons-plus" class="size-5 mr-2" />
          Create Article
        </UButton>
      </div>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-8" icon="i-heroicons-exclamation-circle" />

    <form id="article-form" @submit.prevent="handleSubmit">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <!-- Main content area -->
        <div class="lg:col-span-2 space-y-8">
          <!-- Content Card with enhanced styling -->
          <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
            <template #header>
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-lg bg-primary-50 dark:bg-primary-900/20">
                  <UIcon name="i-heroicons-document-text" class="size-5 text-primary-500" />
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900 dark:text-white">Article Content</h3>
                  <p class="text-sm text-gray-500 dark:text-gray-400">Main article information</p>
                </div>
              </div>
            </template>

            <div class="space-y-6">
              <!-- Title field - prominent styling -->
              <UFormField label="Title" name="title" required class="w-full">
                <template #hint>
                  <span class="text-xs text-gray-400">The main headline of your article</span>
                </template>
                <UInput
                  v-model="form.title"
                  placeholder="Enter a compelling article title..."
                  size="xl"
                  class="font-medium w-full"
                  @blur="!form.slug && generateSlug()"
                />
              </UFormField>

              <!-- Slug field -->
              <UFormField label="Slug" name="slug" required class="w-full">
                <template #hint>
                  <span class="text-xs text-gray-400">URL-friendly identifier (auto-generated from title)</span>
                </template>
                <div class="flex items-center w-full">
                  <span class="inline-flex items-center px-3 py-2 text-sm text-gray-500 bg-gray-50 dark:bg-gray-800 border border-r-0 border-gray-300 dark:border-gray-700 rounded-l-lg shrink-0">
                    /articles/
                  </span>
                  <UInput
                    v-model="form.slug"
                    placeholder="article-slug"
                    class="rounded-l-none flex-1 min-w-0"
                  />
                </div>
              </UFormField>

              <!-- Summary field with larger textarea -->
              <UFormField label="Summary" name="summary" class="w-full">
                <template #hint>
                  <span class="text-xs text-gray-400">Brief description shown in article previews and SEO</span>
                </template>
                <UTextarea
                  v-model="form.summary"
                  placeholder="Write a compelling summary that captures readers' attention..."
                  :rows="4"
                  autoresize
                  :maxrows="8"
                  class="w-full"
                />
              </UFormField>

              <!-- Content editor with label -->
              <UFormField label="Content" name="content" required class="w-full">
                <template #hint>
                  <span class="text-xs text-gray-400">Full article content with rich text formatting</span>
                </template>
                <RichTextEditor
                  v-model="form.content"
                  placeholder="Start writing your article..."
                />
              </UFormField>
            </div>
          </UCard>
        </div>

        <!-- Sidebar with sticky positioning -->
        <div class="space-y-6 lg:sticky lg:top-6 lg:self-start">
          <!-- Publish Card - Primary action -->
          <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800 overflow-hidden">
            <template #header>
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-lg bg-emerald-50 dark:bg-emerald-900/20">
                  <UIcon name="i-heroicons-rocket-launch" class="size-5 text-emerald-500" />
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900 dark:text-white">Publish</h3>
                  <p class="text-sm text-gray-500 dark:text-gray-400">Control article visibility</p>
                </div>
              </div>
            </template>

            <div class="space-y-5">
              <UFormField label="Status" name="status">
                <USelect
                  v-model="form.status"
                  :items="[
                    { label: 'Draft', value: 'draft' },
                    { label: 'Published', value: 'published' },
                    { label: 'Archived', value: 'archived' }
                  ]"
                  value-key="value"
                  size="xl"
                  class="w-full"
                />
              </UFormField>

              <UButton
                type="submit"
                block
                :loading="loading"
                size="lg"
                class="hidden sm:flex font-medium"
                :disabled="loading"
              >
                <UIcon v-if="!loading" name="i-heroicons-plus" class="size-5 mr-2" />
                {{ loading ? 'Creating...' : 'Create Article' }}
              </UButton>
            </div>
          </UCard>

          <!-- Category Card -->
          <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
            <template #header>
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
                  <UIcon name="i-heroicons-folder" class="size-5 text-blue-500" />
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900 dark:text-white">Category</h3>
                  <p class="text-sm text-gray-500 dark:text-gray-400">Organize your article</p>
                </div>
              </div>
            </template>

            <USelect
              v-model="form.category_id"
              :items="categoryOptions"
              value-key="value"
              placeholder="Select a category"
              size="xl"
              class="w-full"
            />
          </UCard>

          <!-- Tags Card with improved styling -->
          <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
            <template #header>
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="p-2 rounded-lg bg-purple-50 dark:bg-purple-900/20">
                    <UIcon name="i-heroicons-tag" class="size-5 text-purple-500" />
                  </div>
                  <div>
                    <h3 class="font-semibold text-gray-900 dark:text-white">Tags</h3>
                    <p class="text-sm text-gray-500 dark:text-gray-400">Add relevant tags</p>
                  </div>
                </div>
                <UBadge v-if="form.tag_ids?.length" color="primary" variant="soft" size="sm">
                  {{ form.tag_ids.length }} selected
                </UBadge>
              </div>
            </template>

            <div v-if="tags.length" class="flex flex-wrap gap-2">
              <UButton
                v-for="tag in tags"
                :key="tag.id"
                size="sm"
                :variant="form.tag_ids?.includes(tag.id) ? 'solid' : 'soft'"
                :color="form.tag_ids?.includes(tag.id) ? 'primary' : 'neutral'"
                class="transition-all duration-200"
                @click="toggleTag(tag.id)"
              >
                <UIcon
                  :name="form.tag_ids?.includes(tag.id) ? 'i-heroicons-check' : 'i-heroicons-plus'"
                  class="size-3.5 mr-1"
                />
                {{ tag.name }}
              </UButton>
            </div>
            <div v-else class="text-center py-4">
              <UIcon name="i-heroicons-tag" class="size-8 text-gray-300 dark:text-gray-600 mx-auto mb-2" />
              <p class="text-sm text-gray-500 dark:text-gray-400">No tags available</p>
              <UButton to="/admin/tags" variant="link" size="sm" class="mt-2">
                Create tags
              </UButton>
            </div>
          </UCard>

          <!-- Featured Image Card with enhanced preview -->
          <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
            <template #header>
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-lg bg-rose-50 dark:bg-rose-900/20">
                  <UIcon name="i-heroicons-photo" class="size-5 text-rose-500" />
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900 dark:text-white">Featured Image</h3>
                  <p class="text-sm text-gray-500 dark:text-gray-400">Article cover image</p>
                </div>
              </div>
            </template>

            <div class="space-y-4">
              <!-- Image preview or upload area -->
              <div
                v-if="form.featured_image"
                class="relative aspect-video rounded-xl overflow-hidden bg-gradient-to-br from-gray-100 to-gray-50 dark:from-gray-800 dark:to-gray-900 ring-1 ring-gray-200 dark:ring-gray-700 group"
              >
                <img
                  :src="form.featured_image"
                  alt="Featured image preview"
                  class="w-full h-full object-cover"
                  @error="form.featured_image = ''"
                >
                <!-- Remove button overlay -->
                <div class="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                  <UButton
                    color="neutral"
                    variant="solid"
                    icon="i-heroicons-arrow-path"
                    size="sm"
                    @click="featuredImageInput?.click()"
                  >
                    Replace
                  </UButton>
                  <UButton
                    color="error"
                    variant="solid"
                    icon="i-heroicons-trash"
                    size="sm"
                    @click="removeFeaturedImage"
                  >
                    Remove
                  </UButton>
                </div>
              </div>

              <!-- Upload area when no image -->
              <div
                v-else
                class="relative aspect-video rounded-xl overflow-hidden bg-gradient-to-br from-gray-100 to-gray-50 dark:from-gray-800 dark:to-gray-900 ring-1 ring-gray-200 dark:ring-gray-700 ring-dashed cursor-pointer hover:ring-primary-500 hover:bg-primary-50/50 dark:hover:bg-primary-900/10 transition-all"
                @click="featuredImageInput?.click()"
                @dragover.prevent
                @drop="handleFeaturedImageDrop"
              >
                <div class="absolute inset-0 flex flex-col items-center justify-center text-gray-400 dark:text-gray-500">
                  <UIcon v-if="uploadingFeaturedImage" name="i-heroicons-arrow-path" class="size-12 mb-2 animate-spin" />
                  <UIcon v-else name="i-heroicons-cloud-arrow-up" class="size-12 mb-2" />
                  <span class="text-sm font-medium">{{ uploadingFeaturedImage ? 'Uploading...' : 'Click or drag to upload' }}</span>
                  <span class="text-xs mt-1">PNG, JPG, GIF, WebP up to 10MB</span>
                </div>
              </div>

              <!-- Hidden file input -->
              <input
                ref="featuredImageInput"
                type="file"
                accept="image/*"
                class="hidden"
                @change="handleFeaturedImageSelect"
              >
            </div>
          </UCard>
        </div>
      </div>
    </form>
  </div>
</template>
