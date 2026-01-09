<script setup lang="ts">
import type { ApiResponse } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const router = useRouter()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const form = ref({
  name: '',
  slug: '',
  level: 'national',
  branch: 'executive',
  display_order: 0,
  description: '',
  max_terms: undefined as number | undefined,
  term_years: 3,
  is_elected: true
})

const loading = ref(false)
const error = ref('')

function generateSlug() {
  form.value.slug = form.value.name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

async function handleSubmit() {
  loading.value = true
  error.value = ''

  try {
    const response = await $fetch<ApiResponse<any>>(`${baseUrl}/admin/positions`, {
      method: 'POST',
      headers: auth.getAuthHeaders(),
      body: {
        name: form.value.name,
        slug: form.value.slug,
        level: form.value.level,
        branch: form.value.branch,
        display_order: form.value.display_order,
        description: form.value.description || null,
        max_terms: form.value.max_terms || null,
        term_years: form.value.term_years,
        is_elected: form.value.is_elected
      }
    })

    if (response.success) {
      router.push('/admin/positions')
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to create position'
  }

  loading.value = false
}

useSeoMeta({
  title: 'New Position - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-2xl mx-auto">
    <div class="flex items-center gap-4 mb-6">
      <UButton
        to="/admin/positions"
        variant="ghost"
        icon="i-heroicons-arrow-left"
        size="sm"
      >
        Back
      </UButton>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">New Position</h1>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <UCard>
      <form class="space-y-6" @submit.prevent="handleSubmit">
        <UInput
          v-model="form.name"
          label="Name"
          placeholder="e.g., President, Mayor, Senator"
          required
          @input="generateSlug"
        />

        <UInput
          v-model="form.slug"
          label="Slug"
          placeholder="e.g., president, mayor, senator"
          required
        />

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <USelect
            v-model="form.level"
            label="Level"
            :options="[
              { label: 'National', value: 'national' },
              { label: 'Regional', value: 'regional' },
              { label: 'Provincial', value: 'provincial' },
              { label: 'City', value: 'city' },
              { label: 'Municipal', value: 'municipal' },
              { label: 'Barangay', value: 'barangay' },
              { label: 'District', value: 'district' }
            ]"
            required
          />

          <USelect
            v-model="form.branch"
            label="Branch"
            :options="[
              { label: 'Executive', value: 'executive' },
              { label: 'Legislative', value: 'legislative' },
              { label: 'Judicial', value: 'judicial' }
            ]"
            required
          />
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
          <UInput
            v-model.number="form.display_order"
            label="Display Order"
            type="number"
            min="0"
            required
          />

          <UInput
            v-model.number="form.term_years"
            label="Term Years"
            type="number"
            min="1"
            required
          />

          <UInput
            v-model.number="form.max_terms"
            label="Max Terms (optional)"
            type="number"
            min="1"
            placeholder="Leave empty for unlimited"
          />
        </div>

        <UTextarea
          v-model="form.description"
          label="Description (optional)"
          placeholder="Brief description of this position..."
          :rows="3"
        />

        <div class="flex items-center gap-2">
          <UCheckbox
            v-model="form.is_elected"
            label="This is an elected position"
          />
        </div>

        <div class="flex justify-end gap-3">
          <UButton
            to="/admin/positions"
            variant="ghost"
            :disabled="loading"
          >
            Cancel
          </UButton>
          <UButton
            type="submit"
            :loading="loading"
            :disabled="loading"
          >
            Create Position
          </UButton>
        </div>
      </form>
    </UCard>
  </div>
</template>
