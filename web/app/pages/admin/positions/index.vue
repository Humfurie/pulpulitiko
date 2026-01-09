<script setup lang="ts">
import type { GovernmentPositionListItem, ApiResponse } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const loading = ref(false)
const positions = ref<GovernmentPositionListItem[]>([])
const error = ref('')

async function fetchPositions() {
  loading.value = true
  error.value = ''

  try {
    const response = await $fetch<ApiResponse<GovernmentPositionListItem[]>>(`${baseUrl}/positions`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      positions.value = response.data
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load positions'
  }

  loading.value = false
}

async function deletePosition(id: string) {
  if (!confirm('Are you sure you want to delete this position?')) return

  try {
    await $fetch(`${baseUrl}/admin/positions/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchPositions()
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    alert(err?.data?.error?.message || 'Failed to delete position')
  }
}

function getLevelColor(level: string) {
  switch (level) {
    case 'national': return 'error'
    case 'regional': return 'warning'
    case 'provincial': return 'primary'
    case 'city': case 'municipal': return 'info'
    case 'barangay': return 'success'
    default: return 'neutral'
  }
}

function getBranchColor(branch: string) {
  switch (branch) {
    case 'executive': return 'primary'
    case 'legislative': return 'warning'
    case 'judicial': return 'error'
    default: return 'neutral'
  }
}

onMounted(fetchPositions)

useSeoMeta({
  title: 'Positions - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-6xl mx-auto">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Government Positions</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Manage official government positions</p>
      </div>
      <UButton to="/admin/positions/new" icon="i-heroicons-plus">
        New Position
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <UCard>
      <div v-if="loading" class="py-12 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
        <p class="mt-2 text-gray-500">Loading positions...</p>
      </div>

      <div v-else-if="positions.length" class="divide-y divide-gray-200 dark:divide-gray-800">
        <div
          v-for="position in positions"
          :key="position.id"
          class="flex items-center justify-between py-4 first:pt-0 last:pb-0"
        >
          <div class="flex-1">
            <NuxtLink :to="`/admin/positions/${position.id}`" class="block">
              <div class="flex items-center gap-3 mb-2">
                <p class="font-medium text-gray-900 dark:text-white hover:text-primary">
                  {{ position.name }}
                </p>
                <UBadge :color="getLevelColor(position.level)" variant="subtle" class="capitalize" size="xs">
                  {{ position.level }}
                </UBadge>
                <UBadge :color="getBranchColor(position.branch)" variant="subtle" class="capitalize" size="xs">
                  {{ position.branch }}
                </UBadge>
                <UBadge v-if="position.is_elected" color="success" variant="outline" size="xs">
                  Elected
                </UBadge>
                <UBadge v-else color="neutral" variant="outline" size="xs">
                  Appointed
                </UBadge>
              </div>
              <p class="text-sm text-gray-500">{{ position.slug }}</p>
            </NuxtLink>
          </div>
          <div class="flex items-center gap-2">
            <UButton
              :to="`/admin/positions/${position.id}`"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
            />
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-trash"
              @click="deletePosition(position.id)"
            />
          </div>
        </div>
      </div>

      <div v-else class="py-12 text-center">
        <UIcon name="i-heroicons-briefcase" class="size-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
        <p class="text-gray-500 dark:text-gray-400 mb-4">No positions yet.</p>
        <UButton to="/admin/positions/new" variant="outline">
          Create your first position
        </UButton>
      </div>
    </UCard>
  </div>
</template>
