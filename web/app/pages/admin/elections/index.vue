<script setup lang="ts">
import type { ElectionFilter, ElectionStatus, ElectionType } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'auth'
})

const api = useApi()

const page = ref(1)
const perPage = ref(20)
const filter = ref<ElectionFilter>({})

const { data, status, refresh: _refresh } = await useAsyncData(
  () => api.getElections(filter.value, page.value, perPage.value),
  { watch: [page, filter] }
)

const electionTypes: { value: ElectionType | ''; label: string }[] = [
  { value: '', label: 'All Types' },
  { value: 'national', label: 'National' },
  { value: 'local', label: 'Local' },
  { value: 'barangay', label: 'Barangay' },
  { value: 'special', label: 'Special' },
  { value: 'plebiscite', label: 'Plebiscite' },
  { value: 'recall', label: 'Recall' }
]

const electionStatuses: { value: ElectionStatus | ''; label: string }[] = [
  { value: '', label: 'All Statuses' },
  { value: 'upcoming', label: 'Upcoming' },
  { value: 'ongoing', label: 'Ongoing' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' }
]

function applyFilter(key: keyof ElectionFilter, value: string | undefined) {
  if (value === '' || value === undefined) {
    const { [key]: _, ...rest } = filter.value
    filter.value = rest
  } else {
    filter.value = { ...filter.value, [key]: value }
  }
  page.value = 1
}

function getStatusColor(status: ElectionStatus): 'primary' | 'success' | 'neutral' | 'error' {
  switch (status) {
    case 'upcoming':
      return 'primary'
    case 'ongoing':
      return 'success'
    case 'completed':
      return 'neutral'
    case 'cancelled':
      return 'error'
    default:
      return 'neutral'
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

useSeoMeta({
  title: 'Elections - Admin - Pulpulitiko'
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Elections</h1>
        <p class="text-gray-500 dark:text-gray-400">Manage election data and candidates</p>
      </div>
      <UButton
        to="/admin/elections/new"
        icon="i-heroicons-plus"
        color="primary"
      >
        Add Election
      </UButton>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
      <div class="flex flex-wrap gap-4">
        <div class="flex-1 min-w-[150px]">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type</label>
          <select
            :value="filter.election_type || ''"
            class="w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
            @change="applyFilter('election_type', ($event.target as HTMLSelectElement).value as ElectionType)"
          >
            <option v-for="type in electionTypes" :key="type.value" :value="type.value">
              {{ type.label }}
            </option>
          </select>
        </div>

        <div class="flex-1 min-w-[150px]">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Status</label>
          <select
            :value="filter.status || ''"
            class="w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
            @change="applyFilter('status', ($event.target as HTMLSelectElement).value as ElectionStatus)"
          >
            <option v-for="s in electionStatuses" :key="s.value" :value="s.value">
              {{ s.label }}
            </option>
          </select>
        </div>

        <div class="flex-[2] min-w-[200px]">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Search</label>
          <UInput
            :model-value="filter.search || ''"
            placeholder="Search elections..."
            icon="i-heroicons-magnifying-glass"
            @update:model-value="applyFilter('search', $event as string)"
          />
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="status === 'pending'" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600" />
    </div>

    <!-- Elections Table -->
    <div v-else-if="data?.elections.length" class="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead class="bg-gray-50 dark:bg-gray-900">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Election
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Type
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Date
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Status
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Stats
            </th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
          <tr
            v-for="election in data.elections"
            :key="election.id"
            class="hover:bg-gray-50 dark:hover:bg-gray-700/50"
          >
            <td class="px-6 py-4">
              <div>
                <NuxtLink
                  :to="`/admin/elections/${election.id}`"
                  class="font-medium text-gray-900 dark:text-white hover:text-primary-600"
                >
                  {{ election.name }}
                </NuxtLink>
                <div v-if="election.is_featured" class="mt-1">
                  <UBadge color="warning" variant="subtle" size="xs">Featured</UBadge>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400 capitalize">
              {{ election.election_type }}
            </td>
            <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
              {{ formatDate(election.election_date) }}
            </td>
            <td class="px-6 py-4">
              <UBadge :color="getStatusColor(election.status)" variant="subtle" class="capitalize">
                {{ election.status }}
              </UBadge>
            </td>
            <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
              <div>{{ election.position_count }} positions</div>
              <div>{{ election.candidate_count }} candidates</div>
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-2">
                <UButton
                  :to="`/elections/${election.slug}`"
                  size="xs"
                  variant="ghost"
                  icon="i-heroicons-eye"
                  target="_blank"
                />
                <UButton
                  :to="`/admin/elections/${election.id}`"
                  size="xs"
                  variant="ghost"
                  icon="i-heroicons-pencil-square"
                />
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-12 bg-white dark:bg-gray-800 rounded-lg">
      <UIcon name="i-heroicons-clipboard-document-list" class="w-12 h-12 mx-auto text-gray-400 mb-4" />
      <h3 class="text-sm font-medium text-gray-900 dark:text-white">No elections found</h3>
      <p class="mt-1 text-sm text-gray-500">Get started by creating a new election.</p>
      <div class="mt-6">
        <UButton to="/admin/elections/new" icon="i-heroicons-plus">
          Add Election
        </UButton>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="data && data.total_pages > 1" class="mt-6 flex justify-center">
      <nav class="flex items-center gap-2">
        <UButton
          :disabled="page === 1"
          variant="outline"
          size="sm"
          @click="page--"
        >
          Previous
        </UButton>
        <span class="text-sm text-gray-700 dark:text-gray-300">
          Page {{ page }} of {{ data.total_pages }}
        </span>
        <UButton
          :disabled="page === data.total_pages"
          variant="outline"
          size="sm"
          @click="page++"
        >
          Next
        </UButton>
      </nav>
    </div>
  </div>
</template>
