<script setup lang="ts">
import type { PoliticalParty, ApiResponse } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const route = useRoute()
const auth = useAuth()
const config = useRuntimeConfig()
const router = useRouter()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const partyId = route.params.id as string

const form = ref({
  name: '',
  slug: '',
  abbreviation: '',
  logo: '',
  color: '#3B82F6',
  description: '',
  founded_year: undefined as number | undefined,
  website: '',
  is_major: false,
  is_active: true
})

const loading = ref(false)
const saving = ref(false)
const error = ref('')

async function fetchParty() {
  loading.value = true
  error.value = ''

  try {
    const response = await $fetch<ApiResponse<PoliticalParty>>(`${baseUrl}/parties/${partyId}`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      const party = response.data
      form.value = {
        name: party.name,
        slug: party.slug,
        abbreviation: party.abbreviation || '',
        logo: party.logo || '',
        color: party.color || '#3B82F6',
        description: party.description || '',
        founded_year: party.founded_year || undefined,
        website: party.website || '',
        is_major: party.is_major,
        is_active: party.is_active
      }
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load party'
  }

  loading.value = false
}

async function handleSubmit() {
  saving.value = true
  error.value = ''

  try {
    const response = await $fetch<ApiResponse<PoliticalParty>>(`${baseUrl}/admin/parties/${partyId}`, {
      method: 'PUT',
      headers: auth.getAuthHeaders(),
      body: {
        name: form.value.name,
        slug: form.value.slug,
        abbreviation: form.value.abbreviation || null,
        logo: form.value.logo || null,
        color: form.value.color || null,
        description: form.value.description || null,
        founded_year: form.value.founded_year || null,
        website: form.value.website || null,
        is_major: form.value.is_major,
        is_active: form.value.is_active
      }
    })

    if (response.success) {
      router.push('/admin/parties')
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to update party'
  }

  saving.value = false
}

onMounted(fetchParty)

useSeoMeta({
  title: 'Edit Party - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-2xl mx-auto">
    <div class="flex items-center gap-4 mb-6">
      <UButton
        to="/admin/parties"
        variant="ghost"
        icon="i-heroicons-arrow-left"
        size="sm"
      >
        Back
      </UButton>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Edit Party</h1>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <UCard v-if="loading" class="py-12 text-center">
      <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      <p class="mt-2 text-gray-500">Loading party...</p>
    </UCard>

    <UCard v-else>
      <form class="space-y-6" @submit.prevent="handleSubmit">
        <UInput
          v-model="form.name"
          label="Name"
          placeholder="e.g., Liberal Party, Conservative Party"
          required
        />

        <UInput
          v-model="form.slug"
          label="Slug"
          placeholder="e.g., liberal-party, conservative-party"
          required
        />

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <UInput
            v-model="form.abbreviation"
            label="Abbreviation (optional)"
            placeholder="e.g., LP, CP"
          />

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Party Color
            </label>
            <div class="flex items-center gap-2">
              <input
                v-model="form.color"
                type="color"
                class="w-12 h-10 rounded border border-gray-300 dark:border-gray-600"
              >
              <UInput
                v-model="form.color"
                placeholder="#3B82F6"
                class="flex-1"
              />
            </div>
          </div>
        </div>

        <UInput
          v-model="form.logo"
          label="Logo URL (optional)"
          placeholder="https://example.com/logo.png"
          type="url"
        />

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <UInput
            v-model.number="form.founded_year"
            label="Founded Year (optional)"
            type="number"
            min="1800"
            :max="new Date().getFullYear()"
            placeholder="e.g., 1946"
          />

          <UInput
            v-model="form.website"
            label="Website (optional)"
            placeholder="https://example.com"
            type="url"
          />
        </div>

        <UTextarea
          v-model="form.description"
          label="Description (optional)"
          placeholder="Brief description of this party..."
          :rows="3"
        />

        <div class="flex flex-col gap-3">
          <UCheckbox
            v-model="form.is_major"
            label="This is a major party"
          />
          <UCheckbox
            v-model="form.is_active"
            label="Party is currently active"
          />
        </div>

        <div class="flex justify-end gap-3">
          <UButton
            to="/admin/parties"
            variant="ghost"
            :disabled="saving"
          >
            Cancel
          </UButton>
          <UButton
            type="submit"
            :loading="saving"
            :disabled="saving || loading"
          >
            Save Changes
          </UButton>
        </div>
      </form>
    </UCard>
  </div>
</template>
