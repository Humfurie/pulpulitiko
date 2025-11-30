<script setup lang="ts">
import type { Category, Tag, ApiResponse, CreateArticleRequest } from '~/types'

definePageMeta({
  layout: 'admin'
})

const auth = useAuth()
const router = useRouter()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const loading = ref(false)
const error = ref('')

const form = reactive<CreateArticleRequest>({
  slug: '',
  title: '',
  summary: '',
  content: '',
  featured_image: '',
  category_id: '',
  status: 'draft',
  tag_ids: []
})

const categories = ref<Category[]>([])
const tags = ref<Tag[]>([])

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

async function handleSubmit() {
  loading.value = true
  error.value = ''

  try {
    const payload = {
      ...form,
      category_id: form.category_id || undefined,
      tag_ids: form.tag_ids?.length ? form.tag_ids : undefined
    }

    await $fetch(`${baseUrl}/admin/articles`, {
      method: 'POST',
      headers: auth.getAuthHeaders(),
      body: payload
    })

    await router.push('/admin/articles')
  } catch (e: any) {
    error.value = e?.data?.error?.message || 'Failed to create article'
  }

  loading.value = false
}

onMounted(loadData)

useSeoMeta({
  title: 'New Article - Pulpulitiko Admin'
})
</script>

<template>
  <div>
    <div class="flex items-center gap-4 mb-6">
      <UButton to="/admin/articles" variant="ghost" icon="i-heroicons-arrow-left" />
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">New Article</h1>
    </div>

    <form @submit.prevent="handleSubmit">
      <UAlert v-if="error" color="error" :title="error" class="mb-4" />

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Main content -->
        <div class="lg:col-span-2 space-y-6">
          <UCard>
            <div class="space-y-4">
              <UFormField label="Title" name="title" required>
                <UInput
                  v-model="form.title"
                  placeholder="Article title"
                  @blur="!form.slug && generateSlug()"
                />
              </UFormField>

              <UFormField label="Slug" name="slug" required>
                <UInput v-model="form.slug" placeholder="article-slug" />
              </UFormField>

              <UFormField label="Summary" name="summary">
                <UTextarea
                  v-model="form.summary"
                  placeholder="Brief summary of the article"
                  :rows="3"
                />
              </UFormField>

              <UFormField label="Content" name="content" required>
                <UTextarea
                  v-model="form.content"
                  placeholder="Article content (HTML supported)"
                  :rows="15"
                />
              </UFormField>
            </div>
          </UCard>
        </div>

        <!-- Sidebar -->
        <div class="space-y-6">
          <UCard>
            <template #header>
              <h3 class="font-semibold text-gray-900 dark:text-white">Publish</h3>
            </template>

            <div class="space-y-4">
              <UFormField label="Status" name="status">
                <USelect
                  v-model="form.status"
                  :options="[
                    { label: 'Draft', value: 'draft' },
                    { label: 'Published', value: 'published' },
                    { label: 'Archived', value: 'archived' }
                  ]"
                />
              </UFormField>

              <UButton type="submit" block :loading="loading">
                Create Article
              </UButton>
            </div>
          </UCard>

          <UCard>
            <template #header>
              <h3 class="font-semibold text-gray-900 dark:text-white">Category</h3>
            </template>

            <USelect
              v-model="form.category_id"
              :options="[
                { label: 'None', value: '' },
                ...categories.map(c => ({ label: c.name, value: c.id }))
              ]"
            />
          </UCard>

          <UCard>
            <template #header>
              <h3 class="font-semibold text-gray-900 dark:text-white">Featured Image</h3>
            </template>

            <UInput
              v-model="form.featured_image"
              placeholder="Image URL"
            />
          </UCard>
        </div>
      </div>
    </form>
  </div>
</template>
