<script setup lang="ts">
import type { CreatePoliticianRequest } from '~/types'

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
const uploadingPhoto = ref(false)
const error = ref('')
const photoInput = ref<HTMLInputElement | null>(null)

const form = reactive<CreatePoliticianRequest>({
  name: '',
  slug: '',
  photo: '',
  position: '',
  party: '',
  short_bio: ''
})

function generateSlug() {
  form.slug = form.name
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .trim()
}

async function uploadPhoto(file: File) {
  if (!file.type.startsWith('image/')) {
    error.value = 'Please select an image file'
    return
  }

  uploadingPhoto.value = true
  error.value = ''

  try {
    const result = await api.uploadFile(file, auth.getAuthHeaders())
    form.photo = result.url
  } catch (e: unknown) {
    const err = e as { message?: string }
    error.value = err.message || 'Failed to upload photo'
  } finally {
    uploadingPhoto.value = false
  }
}

function handlePhotoSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    uploadPhoto(file)
    target.value = ''
  }
}

function removePhoto() {
  form.photo = ''
}

async function handleSubmit() {
  loading.value = true
  error.value = ''

  try {
    const payload: CreatePoliticianRequest = {
      name: form.name,
      slug: form.slug,
      photo: form.photo || undefined,
      position: form.position || undefined,
      party: form.party || undefined,
      short_bio: form.short_bio || undefined
    }

    await $fetch(`${baseUrl}/admin/politicians`, {
      method: 'POST',
      headers: auth.getAuthHeaders(),
      body: payload
    })

    await router.push('/admin/politicians')
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to create politician'
  }

  loading.value = false
}

useSeoMeta({
  title: 'New Politician - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
      <div class="flex items-center gap-4">
        <UButton
          to="/admin/politicians"
          variant="soft"
          color="neutral"
          icon="i-heroicons-arrow-left"
          size="md"
          class="shadow-sm"
        />
        <div>
          <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">New Politician</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Add a new politician profile</p>
        </div>
      </div>
      <div class="sm:hidden">
        <UButton type="submit" form="politician-form" :loading="loading" class="w-full" size="lg">
          Create Politician
        </UButton>
      </div>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-8" icon="i-heroicons-exclamation-circle" />

    <form id="politician-form" @submit.prevent="handleSubmit">
      <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-primary-50 dark:bg-primary-900/20">
              <UIcon name="i-heroicons-user-circle" class="size-5 text-primary-500" />
            </div>
            <div>
              <h3 class="font-semibold text-gray-900 dark:text-white">Politician Details</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">Basic information about the politician</p>
            </div>
          </div>
        </template>

        <div class="space-y-6">
          <!-- Photo -->
          <div class="flex items-start gap-6">
            <div
              class="relative size-24 rounded-full overflow-hidden bg-gray-100 dark:bg-gray-800 ring-2 ring-gray-200 dark:ring-gray-700 cursor-pointer group"
              @click="photoInput?.click()"
            >
              <img
                v-if="form.photo"
                :src="form.photo"
                alt="Photo"
                class="w-full h-full object-cover"
              >
              <div v-else class="w-full h-full flex items-center justify-center">
                <UIcon name="i-heroicons-user" class="size-12 text-gray-400" />
              </div>
              <div class="absolute inset-0 bg-black/50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                <UIcon v-if="uploadingPhoto" name="i-heroicons-arrow-path" class="size-6 text-white animate-spin" />
                <UIcon v-else name="i-heroicons-camera" class="size-6 text-white" />
              </div>
            </div>
            <div class="flex-1">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Profile Photo</p>
              <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">Click to upload a photo</p>
              <div class="flex gap-2">
                <UButton size="sm" variant="soft" :loading="uploadingPhoto" @click="photoInput?.click()">
                  Upload
                </UButton>
                <UButton v-if="form.photo" size="sm" variant="soft" color="error" @click="removePhoto">
                  Remove
                </UButton>
              </div>
            </div>
            <input
              ref="photoInput"
              type="file"
              accept="image/*"
              class="hidden"
              @change="handlePhotoSelect"
            >
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <UFormField label="Name" name="name" required class="w-full">
              <template #hint>
                <span class="text-xs text-gray-400">Full name of the politician</span>
              </template>
              <UInput
                v-model="form.name"
                placeholder="Juan Dela Cruz"
                size="lg"
                @blur="!form.slug && generateSlug()"
              />
            </UFormField>

            <UFormField label="Slug" name="slug" required class="w-full">
              <template #hint>
                <span class="text-xs text-gray-400">URL-friendly identifier</span>
              </template>
              <div class="flex items-center w-full">
                <span class="inline-flex items-center px-3 py-2 text-sm text-gray-500 bg-gray-50 dark:bg-gray-800 border border-r-0 border-gray-300 dark:border-gray-700 rounded-l-lg shrink-0">
                  /politician/
                </span>
                <UInput
                  v-model="form.slug"
                  placeholder="juan-dela-cruz"
                  class="rounded-l-none flex-1 min-w-0"
                />
              </div>
            </UFormField>

            <UFormField label="Position" name="position" class="w-full">
              <template #hint>
                <span class="text-xs text-gray-400">Current political position</span>
              </template>
              <UInput
                v-model="form.position"
                placeholder="Senator, Mayor, etc."
                icon="i-heroicons-briefcase"
              />
            </UFormField>

            <UFormField label="Party" name="party" class="w-full">
              <template #hint>
                <span class="text-xs text-gray-400">Political party affiliation</span>
              </template>
              <UInput
                v-model="form.party"
                placeholder="Political party name"
                icon="i-heroicons-flag"
              />
            </UFormField>
          </div>

          <UFormField label="Short Bio" name="short_bio" class="w-full">
            <template #hint>
              <span class="text-xs text-gray-400">Brief biography or description</span>
            </template>
            <UTextarea
              v-model="form.short_bio"
              placeholder="Write a brief bio about this politician..."
              :rows="4"
              autoresize
            />
          </UFormField>
        </div>

        <template #footer>
          <div class="flex flex-col-reverse sm:flex-row sm:justify-end gap-3">
            <UButton to="/admin/politicians" variant="soft" color="neutral">
              Cancel
            </UButton>
            <UButton type="submit" :loading="loading" size="lg">
              <UIcon v-if="!loading" name="i-heroicons-plus" class="size-4 mr-2" />
              {{ loading ? 'Creating...' : 'Create Politician' }}
            </UButton>
          </div>
        </template>
      </UCard>
    </form>
  </div>
</template>
