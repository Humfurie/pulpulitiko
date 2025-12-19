<script setup lang="ts">
import type { PoliticalPartyListItem, ApiResponse } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const loading = ref(false)
const parties = ref<PoliticalPartyListItem[]>([])
const error = ref('')

async function fetchParties() {
  loading.value = true
  error.value = ''

  try {
    const response = await $fetch<ApiResponse<PoliticalPartyListItem[]>>(`${baseUrl}/parties`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      parties.value = response.data
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load parties'
  }

  loading.value = false
}

async function deleteParty(id: string) {
  if (!confirm('Are you sure you want to delete this party?')) return

  try {
    await $fetch(`${baseUrl}/admin/parties/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchParties()
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    alert(err?.data?.error?.message || 'Failed to delete party')
  }
}

onMounted(fetchParties)

useSeoMeta({
  title: 'Political Parties - Pulpulitiko Admin'
})
</script>

<template>
  <div class="max-w-6xl mx-auto">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Political Parties</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Manage political parties</p>
      </div>
      <UButton to="/admin/parties/new" icon="i-heroicons-plus">
        New Party
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <UCard>
      <div v-if="loading" class="py-12 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
        <p class="mt-2 text-gray-500">Loading parties...</p>
      </div>

      <div v-else-if="parties.length" class="divide-y divide-gray-200 dark:divide-gray-800">
        <div
          v-for="party in parties"
          :key="party.id"
          class="flex items-center justify-between py-4 first:pt-0 last:pb-0"
        >
          <div class="flex items-center gap-4 flex-1">
            <div
              v-if="party.logo"
              class="flex-shrink-0 w-12 h-12 rounded-lg overflow-hidden bg-gray-100 dark:bg-gray-800"
            >
              <img :src="party.logo" :alt="party.name" class="w-full h-full object-cover" >
            </div>
            <div
              v-else
              class="flex-shrink-0 w-12 h-12 rounded-lg flex items-center justify-center"
              :style="{ backgroundColor: party.color || '#e5e7eb' }"
            >
              <span class="text-white font-bold text-lg">
                {{ party.abbreviation || party.name.charAt(0) }}
              </span>
            </div>

            <div class="flex-1 min-w-0">
              <NuxtLink :to="`/admin/parties/${party.id}`" class="block">
                <div class="flex items-center gap-2 mb-1">
                  <p class="font-medium text-gray-900 dark:text-white hover:text-primary">
                    {{ party.name }}
                  </p>
                  <span v-if="party.abbreviation" class="text-sm text-gray-500">
                    ({{ party.abbreviation }})
                  </span>
                  <UBadge v-if="party.is_major" color="primary" variant="subtle" size="xs">
                    Major Party
                  </UBadge>
                  <UBadge v-if="!party.is_active" color="neutral" variant="outline" size="xs">
                    Inactive
                  </UBadge>
                </div>
                <div class="flex items-center gap-3 text-sm text-gray-500">
                  <span>{{ party.slug }}</span>
                  <span v-if="party.member_count > 0">
                    • {{ party.member_count }} member{{ party.member_count !== 1 ? 's' : '' }}
                  </span>
                </div>
              </NuxtLink>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <UButton
              :to="`/admin/parties/${party.id}`"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
            />
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-trash"
              @click="deleteParty(party.id)"
            />
          </div>
        </div>
      </div>

      <div v-else class="py-12 text-center">
        <UIcon name="i-heroicons-flag" class="size-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
        <p class="text-gray-500 dark:text-gray-400 mb-4">No parties yet.</p>
        <UButton to="/admin/parties/new" variant="outline">
          Create your first party
        </UButton>
      </div>
    </UCard>
  </div>
</template>
