<script setup lang="ts">
import type { ApiResponse, CreateAuthorRequest, SocialLinks, RoleWithPermissionCount } from '~/types'

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
const loadingRoles = ref(false)
const uploadingAvatar = ref(false)
const error = ref('')
const avatarInput = ref<HTMLInputElement | null>(null)
const roles = ref<RoleWithPermissionCount[]>([])

const form = reactive({
  name: '',
  slug: '',
  email: '',
  phone: '',
  address: '',
  bio: '',
  avatar: '',
  role_id: '',
  social_links: {
    twitter: '',
    facebook: '',
    linkedin: '',
    instagram: '',
    youtube: '',
    tiktok: '',
    website: ''
  }
})

async function loadRoles() {
  loadingRoles.value = true
  try {
    const response = await $fetch<ApiResponse<RoleWithPermissionCount[]>>(`${baseUrl}/admin/roles`, {
      headers: auth.getAuthHeaders()
    })
    if (response.success) {
      roles.value = response.data || []
      // Set default role to 'author' if available
      const authorRole = roles.value.find(r => r.slug === 'author')
      if (authorRole) {
        form.role_id = authorRole.id
      } else if (roles.value.length > 0) {
        form.role_id = roles.value[0].id
      }
    }
  } catch (e: unknown) {
    console.error('Failed to load roles', e)
  }
  loadingRoles.value = false
}

onMounted(loadRoles)

function generateSlug() {
  form.slug = form.name
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .trim()
}

async function uploadAvatar(file: File) {
  if (!file.type.startsWith('image/')) {
    error.value = 'Please select an image file'
    return
  }

  uploadingAvatar.value = true
  error.value = ''

  try {
    const result = await api.uploadFile(file, auth.getAuthHeaders())
    form.avatar = result.url
  } catch (e: unknown) {
    const err = e as { message?: string }
    error.value = err.message || 'Failed to upload avatar'
  } finally {
    uploadingAvatar.value = false
  }
}

function handleAvatarSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    uploadAvatar(file)
    target.value = ''
  }
}

function removeAvatar() {
  form.avatar = ''
}

async function handleSubmit() {
  loading.value = true
  error.value = ''

  try {
    // Clean up empty social links
    const socialLinks: SocialLinks = {}
    if (form.social_links) {
      Object.entries(form.social_links).forEach(([key, value]) => {
        if (value) {
          socialLinks[key as keyof SocialLinks] = value
        }
      })
    }

    const payload: CreateAuthorRequest = {
      name: form.name,
      slug: form.slug,
      email: form.email || undefined,
      phone: form.phone || undefined,
      address: form.address || undefined,
      bio: form.bio || undefined,
      avatar: form.avatar || undefined,
      role_id: form.role_id || undefined,
      social_links: Object.keys(socialLinks).length > 0 ? socialLinks : undefined
    }

    await $fetch(`${baseUrl}/admin/users`, {
      method: 'POST',
      headers: auth.getAuthHeaders(),
      body: payload
    })

    await router.push('/admin/users')
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to create user'
  }

  loading.value = false
}

useSeoMeta({
  title: 'New User - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
      <div class="flex items-center gap-4">
        <UButton
          to="/admin/users"
          variant="soft"
          color="neutral"
          icon="i-heroicons-arrow-left"
          size="md"
          class="shadow-sm"
        />
        <div>
          <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">New User</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Create a new user account</p>
        </div>
      </div>
      <div class="sm:hidden">
        <UButton type="submit" form="user-form" :loading="loading" class="w-full" size="lg">
          Create User
        </UButton>
      </div>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-8" icon="i-heroicons-exclamation-circle" />

    <form id="user-form" @submit.prevent="handleSubmit">
      <div class="space-y-8">
        <!-- Basic Information -->
        <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
          <template #header>
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-lg bg-primary-50 dark:bg-primary-900/20">
                <UIcon name="i-heroicons-user" class="size-5 text-primary-500" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">Basic Information</h3>
                <p class="text-sm text-gray-500 dark:text-gray-400">User's personal details</p>
              </div>
            </div>
          </template>

          <div class="space-y-6">
            <!-- Avatar -->
            <div class="flex items-start gap-6">
              <div
                class="relative size-24 rounded-full overflow-hidden bg-gray-100 dark:bg-gray-800 ring-2 ring-gray-200 dark:ring-gray-700 cursor-pointer group"
                @click="avatarInput?.click()"
              >
                <img
                  v-if="form.avatar"
                  :src="form.avatar"
                  alt="Avatar"
                  class="w-full h-full object-cover"
                >
                <div v-else class="w-full h-full flex items-center justify-center">
                  <UIcon name="i-heroicons-user" class="size-12 text-gray-400" />
                </div>
                <div class="absolute inset-0 bg-black/50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                  <UIcon v-if="uploadingAvatar" name="i-heroicons-arrow-path" class="size-6 text-white animate-spin" />
                  <UIcon v-else name="i-heroicons-camera" class="size-6 text-white" />
                </div>
              </div>
              <div class="flex-1">
                <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Profile Photo</p>
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">Click to upload or drag and drop</p>
                <div class="flex gap-2">
                  <UButton size="sm" variant="soft" :loading="uploadingAvatar" @click="avatarInput?.click()">
                    Upload
                  </UButton>
                  <UButton v-if="form.avatar" size="sm" variant="soft" color="error" @click="removeAvatar">
                    Remove
                  </UButton>
                </div>
              </div>
              <input
                ref="avatarInput"
                type="file"
                accept="image/*"
                class="hidden"
                @change="handleAvatarSelect"
              >
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <UFormField label="Name" name="name" required class="w-full">
                <UInput
                  v-model="form.name"
                  placeholder="John Doe"
                  size="lg"
                  @blur="!form.slug && generateSlug()"
                />
              </UFormField>

              <UFormField label="Slug" name="slug" required class="w-full">
                <template #hint>
                  <span class="text-xs text-gray-400">URL-friendly identifier</span>
                </template>
                <UInput v-model="form.slug" placeholder="john-doe" />
              </UFormField>

              <UFormField label="Email" name="email" class="w-full">
                <UInput
                  v-model="form.email"
                  type="email"
                  placeholder="john@example.com"
                  icon="i-heroicons-envelope"
                />
              </UFormField>

              <UFormField label="Phone" name="phone" class="w-full">
                <UInput
                  v-model="form.phone"
                  type="tel"
                  placeholder="+1 234 567 8900"
                  icon="i-heroicons-phone"
                />
              </UFormField>
            </div>

            <UFormField label="Address" name="address" class="w-full">
              <UTextarea
                v-model="form.address"
                placeholder="Enter full address..."
                :rows="2"
                autoresize
              />
            </UFormField>

            <UFormField label="Bio" name="bio" class="w-full">
              <template #hint>
                <span class="text-xs text-gray-400">Short biography or description</span>
              </template>
              <UTextarea
                v-model="form.bio"
                placeholder="Write a brief bio..."
                :rows="3"
                autoresize
              />
            </UFormField>

            <UFormField label="Role" name="role_id" required class="w-full">
              <USelect
                v-model="form.role_id"
                :items="roles.map(r => ({ label: `${r.name} - ${r.description || 'No description'}`, value: r.id }))"
                value-key="value"
                option-attribute="label"
                :loading="loadingRoles"
                class="w-full"
              />
            </UFormField>
          </div>
        </UCard>

        <!-- Social Links -->
        <UCard class="shadow-sm ring-1 ring-gray-200 dark:ring-gray-800">
          <template #header>
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-lg bg-blue-50 dark:bg-blue-900/20">
                <UIcon name="i-heroicons-globe-alt" class="size-5 text-blue-500" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">Social Links</h3>
                <p class="text-sm text-gray-500 dark:text-gray-400">Connect social media profiles</p>
              </div>
            </div>
          </template>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <UFormField label="Website" name="website">
              <UInput
                v-model="form.social_links!.website"
                placeholder="https://example.com"
                icon="i-heroicons-globe-alt"
              />
            </UFormField>

            <UFormField label="Twitter / X" name="twitter">
              <UInput
                v-model="form.social_links!.twitter"
                placeholder="https://twitter.com/username"
                icon="i-simple-icons-x"
              />
            </UFormField>

            <UFormField label="Facebook" name="facebook">
              <UInput
                v-model="form.social_links!.facebook"
                placeholder="https://facebook.com/username"
                icon="i-simple-icons-facebook"
              />
            </UFormField>

            <UFormField label="LinkedIn" name="linkedin">
              <UInput
                v-model="form.social_links!.linkedin"
                placeholder="https://linkedin.com/in/username"
                icon="i-simple-icons-linkedin"
              />
            </UFormField>

            <UFormField label="Instagram" name="instagram">
              <UInput
                v-model="form.social_links!.instagram"
                placeholder="https://instagram.com/username"
                icon="i-simple-icons-instagram"
              />
            </UFormField>

            <UFormField label="YouTube" name="youtube">
              <UInput
                v-model="form.social_links!.youtube"
                placeholder="https://youtube.com/@username"
                icon="i-simple-icons-youtube"
              />
            </UFormField>

            <UFormField label="TikTok" name="tiktok">
              <UInput
                v-model="form.social_links!.tiktok"
                placeholder="https://tiktok.com/@username"
                icon="i-simple-icons-tiktok"
              />
            </UFormField>
          </div>
        </UCard>

        <!-- Submit Button -->
        <div class="hidden sm:flex justify-end">
          <UButton type="submit" :loading="loading" size="lg">
            Create User
          </UButton>
        </div>
      </div>
    </form>
  </div>
</template>
