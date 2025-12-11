<script setup lang="ts">
import type { ElectionFilter, ElectionType, ElectionStatus } from '~/types'

const api = useApi()

const page = ref(1)
const perPage = ref(12)
const filter = ref<ElectionFilter>({})

const { data, status } = await useAsyncData(
  () => api.getElections(filter.value, page.value, perPage.value),
  { watch: [page, filter] }
)

const { data: upcomingElections } = await useAsyncData(
  'upcoming-elections',
  () => api.getUpcomingElections(3)
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

const currentYear = new Date().getFullYear()
const years = Array.from({ length: 10 }, (_, i) => currentYear + 2 - i)

function applyFilter(key: keyof ElectionFilter, value: string | number | boolean | undefined) {
  if (value === '' || value === undefined) {
    const { [key]: _, ...rest } = filter.value
    filter.value = rest
  } else {
    filter.value = { ...filter.value, [key]: value }
  }
  page.value = 1
}

function getStatusColor(status: ElectionStatus): string {
  switch (status) {
    case 'upcoming':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
    case 'ongoing':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
    case 'completed':
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
    case 'cancelled':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

function getTypeLabel(type: ElectionType): string {
  return type.charAt(0).toUpperCase() + type.slice(1)
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

function getDaysUntil(dateStr: string): number {
  const electionDate = new Date(dateStr)
  const today = new Date()
  const diffTime = electionDate.getTime() - today.getTime()
  return Math.ceil(diffTime / (1000 * 60 * 60 * 24))
}

useSeoMeta({
  title: 'Elections - Pulpulitiko',
  description: 'Browse Philippine elections - national, local, barangay, and special elections'
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <!-- Hero Section -->
    <div class="bg-gradient-to-r from-blue-600 to-indigo-700 text-white py-12">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h1 class="text-4xl font-bold mb-4">Philippine Elections</h1>
        <p class="text-xl text-blue-100">
          Track upcoming and past elections across the Philippines
        </p>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Upcoming Elections Highlight -->
      <div v-if="upcomingElections?.length" class="mb-8">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">Upcoming Elections</h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <NuxtLink
            v-for="election in upcomingElections"
            :key="election.id"
            :to="`/elections/${election.slug}`"
            class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border-l-4 border-blue-500 hover:shadow-lg transition-shadow"
          >
            <div class="flex items-start justify-between">
              <div>
                <span class="text-sm font-medium text-blue-600 dark:text-blue-400">
                  {{ getTypeLabel(election.election_type) }}
                </span>
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mt-1">
                  {{ election.name }}
                </h3>
                <p class="text-gray-600 dark:text-gray-400 mt-2">
                  {{ formatDate(election.election_date) }}
                </p>
              </div>
              <div class="text-right">
                <div class="text-3xl font-bold text-blue-600">{{ getDaysUntil(election.election_date) }}</div>
                <div class="text-sm text-gray-500">days left</div>
              </div>
            </div>
          </NuxtLink>
        </div>
      </div>

      <!-- Filters -->
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <div class="flex flex-wrap gap-4">
          <!-- Type Filter -->
          <div class="flex-1 min-w-[150px]">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Type
            </label>
            <select
              :value="filter.election_type || ''"
              class="w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white shadow-sm focus:border-blue-500 focus:ring-blue-500"
              @change="applyFilter('election_type', ($event.target as HTMLSelectElement).value as ElectionType)"
            >
              <option v-for="type in electionTypes" :key="type.value" :value="type.value">
                {{ type.label }}
              </option>
            </select>
          </div>

          <!-- Status Filter -->
          <div class="flex-1 min-w-[150px]">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Status
            </label>
            <select
              :value="filter.status || ''"
              class="w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white shadow-sm focus:border-blue-500 focus:ring-blue-500"
              @change="applyFilter('status', ($event.target as HTMLSelectElement).value as ElectionStatus)"
            >
              <option v-for="s in electionStatuses" :key="s.value" :value="s.value">
                {{ s.label }}
              </option>
            </select>
          </div>

          <!-- Year Filter -->
          <div class="flex-1 min-w-[120px]">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Year
            </label>
            <select
              :value="filter.year || ''"
              class="w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white shadow-sm focus:border-blue-500 focus:ring-blue-500"
              @change="applyFilter('year', ($event.target as HTMLSelectElement).value ? parseInt(($event.target as HTMLSelectElement).value) : undefined)"
            >
              <option value="">All Years</option>
              <option v-for="year in years" :key="year" :value="year">{{ year }}</option>
            </select>
          </div>

          <!-- Search -->
          <div class="flex-[2] min-w-[200px]">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Search
            </label>
            <input
              type="text"
              :value="filter.search || ''"
              placeholder="Search elections..."
              class="w-full rounded-md border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white shadow-sm focus:border-blue-500 focus:ring-blue-500"
              @input="applyFilter('search', ($event.target as HTMLInputElement).value)"
            >
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="status === 'pending'" class="flex justify-center py-12">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
      </div>

      <!-- Elections Grid -->
      <div v-else-if="data?.elections.length" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <NuxtLink
          v-for="election in data.elections"
          :key="election.id"
          :to="`/elections/${election.slug}`"
          class="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow"
        >
          <div class="p-6">
            <div class="flex items-center justify-between mb-3">
              <span
                :class="[
                  'px-2 py-1 text-xs font-medium rounded-full',
                  getStatusColor(election.status)
                ]"
              >
                {{ election.status.charAt(0).toUpperCase() + election.status.slice(1) }}
              </span>
              <span
                v-if="election.is_featured"
                class="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 rounded-full"
              >
                Featured
              </span>
            </div>

            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              {{ election.name }}
            </h3>

            <div class="space-y-2 text-sm text-gray-600 dark:text-gray-400">
              <div class="flex items-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <span>{{ formatDate(election.election_date) }}</span>
              </div>

              <div class="flex items-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                </svg>
                <span>{{ getTypeLabel(election.election_type) }} Election</span>
              </div>

              <div class="flex items-center gap-4 mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
                <div>
                  <span class="font-medium text-gray-900 dark:text-white">{{ election.position_count }}</span>
                  <span class="text-gray-500"> positions</span>
                </div>
                <div>
                  <span class="font-medium text-gray-900 dark:text-white">{{ election.candidate_count }}</span>
                  <span class="text-gray-500"> candidates</span>
                </div>
              </div>

              <div v-if="election.voter_turnout_percentage" class="mt-2">
                <div class="flex items-center justify-between text-xs mb-1">
                  <span>Voter Turnout</span>
                  <span class="font-medium">{{ election.voter_turnout_percentage.toFixed(1) }}%</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    class="bg-green-500 h-2 rounded-full"
                    :style="{ width: `${election.voter_turnout_percentage}%` }"
                  />
                </div>
              </div>
            </div>
          </div>
        </NuxtLink>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-12">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No elections found</h3>
        <p class="mt-1 text-sm text-gray-500">Try adjusting your filters.</p>
      </div>

      <!-- Pagination -->
      <div v-if="data && data.total_pages > 1" class="mt-8 flex justify-center">
        <nav class="flex items-center gap-2">
          <button
            :disabled="page === 1"
            class="px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
            @click="page--"
          >
            Previous
          </button>
          <span class="text-sm text-gray-700 dark:text-gray-300">
            Page {{ page }} of {{ data.total_pages }}
          </span>
          <button
            :disabled="page === data.total_pages"
            class="px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
            @click="page++"
          >
            Next
          </button>
        </nav>
      </div>
    </div>
  </div>
</template>
