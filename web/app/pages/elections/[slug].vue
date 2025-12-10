<script setup lang="ts">
import type { ElectionStatus, ElectionType, CandidateStatus } from '~/types'

const route = useRoute()
const api = useApi()
const slug = route.params.slug as string

const { data: election, status } = await useAsyncData(
  `election-${slug}`,
  () => api.getElectionBySlug(slug)
)

const { data: positions } = await useAsyncData(
  `election-positions-${slug}`,
  async () => {
    if (election.value?.id) {
      return api.getElectionPositions(election.value.id)
    }
    return []
  },
  { watch: [election] }
)

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

function getCandidateStatusColor(status: CandidateStatus): string {
  switch (status) {
    case 'qualified':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
    case 'filed':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
    case 'disqualified':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    case 'withdrawn':
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
    case 'substituted':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

function getTypeLabel(type: ElectionType): string {
  return type.charAt(0).toUpperCase() + type.slice(1)
}

function formatDate(dateStr: string | undefined): string {
  if (!dateStr) return 'TBD'
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
  title: () => election.value ? `${election.value.name} - Elections - Pulpulitiko` : 'Election - Pulpulitiko',
  description: () => election.value?.description || `Details about ${election.value?.name}`
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <!-- Loading State -->
    <div v-if="status === 'pending'" class="flex justify-center items-center min-h-screen">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
    </div>

    <!-- Error State -->
    <div v-else-if="!election" class="max-w-7xl mx-auto px-4 py-16 text-center">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">Election Not Found</h1>
      <p class="text-gray-600 dark:text-gray-400 mb-8">The election you're looking for doesn't exist.</p>
      <NuxtLink
        to="/elections"
        class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
      >
        Back to Elections
      </NuxtLink>
    </div>

    <!-- Election Content -->
    <template v-else>
      <!-- Hero Section -->
      <div class="bg-gradient-to-r from-blue-600 to-indigo-700 text-white py-12">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div class="flex items-center gap-2 mb-4">
            <NuxtLink to="/elections" class="text-blue-200 hover:text-white">Elections</NuxtLink>
            <span class="text-blue-300">/</span>
            <span>{{ election.name }}</span>
          </div>

          <div class="flex flex-wrap items-start justify-between gap-4">
            <div>
              <div class="flex items-center gap-3 mb-2">
                <span
                  :class="[
                    'px-3 py-1 text-sm font-medium rounded-full',
                    getStatusColor(election.status)
                  ]"
                >
                  {{ election.status.charAt(0).toUpperCase() + election.status.slice(1) }}
                </span>
                <span class="px-3 py-1 text-sm font-medium bg-white/20 rounded-full">
                  {{ getTypeLabel(election.election_type) }}
                </span>
              </div>
              <h1 class="text-4xl font-bold">{{ election.name }}</h1>
              <p v-if="election.description" class="text-xl text-blue-100 mt-2">
                {{ election.description }}
              </p>
            </div>

            <!-- Countdown for upcoming elections -->
            <div v-if="election.status === 'upcoming'" class="text-center bg-white/10 rounded-lg p-6">
              <div class="text-5xl font-bold">{{ getDaysUntil(election.election_date) }}</div>
              <div class="text-blue-200">days until election</div>
            </div>
          </div>
        </div>
      </div>

      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <!-- Key Dates -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 mb-8">
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Key Dates</h2>
          <div class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
            <div v-if="election.registration_start" class="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <div class="text-sm text-gray-500 dark:text-gray-400">Registration Start</div>
              <div class="font-semibold text-gray-900 dark:text-white">{{ formatDate(election.registration_start) }}</div>
            </div>
            <div v-if="election.registration_end" class="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <div class="text-sm text-gray-500 dark:text-gray-400">Registration End</div>
              <div class="font-semibold text-gray-900 dark:text-white">{{ formatDate(election.registration_end) }}</div>
            </div>
            <div v-if="election.campaign_start" class="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <div class="text-sm text-gray-500 dark:text-gray-400">Campaign Start</div>
              <div class="font-semibold text-gray-900 dark:text-white">{{ formatDate(election.campaign_start) }}</div>
            </div>
            <div v-if="election.campaign_end" class="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <div class="text-sm text-gray-500 dark:text-gray-400">Campaign End</div>
              <div class="font-semibold text-gray-900 dark:text-white">{{ formatDate(election.campaign_end) }}</div>
            </div>
            <div class="text-center p-4 bg-blue-50 dark:bg-blue-900/30 rounded-lg border-2 border-blue-500">
              <div class="text-sm text-blue-600 dark:text-blue-400">Election Day</div>
              <div class="font-bold text-blue-700 dark:text-blue-300">{{ formatDate(election.election_date) }}</div>
            </div>
          </div>
        </div>

        <!-- Stats (for completed elections) -->
        <div v-if="election.status === 'completed' && election.voter_turnout_percentage" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 mb-8">
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Election Results</h2>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div v-if="election.total_registered_voters" class="text-center">
              <div class="text-3xl font-bold text-gray-900 dark:text-white">
                {{ election.total_registered_voters.toLocaleString() }}
              </div>
              <div class="text-gray-500 dark:text-gray-400">Registered Voters</div>
            </div>
            <div v-if="election.total_votes_cast" class="text-center">
              <div class="text-3xl font-bold text-gray-900 dark:text-white">
                {{ election.total_votes_cast.toLocaleString() }}
              </div>
              <div class="text-gray-500 dark:text-gray-400">Votes Cast</div>
            </div>
            <div class="text-center">
              <div class="text-3xl font-bold text-green-600">
                {{ election.voter_turnout_percentage.toFixed(1) }}%
              </div>
              <div class="text-gray-500 dark:text-gray-400">Voter Turnout</div>
            </div>
          </div>
        </div>

        <!-- Positions and Candidates -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-6">Positions & Candidates</h2>

          <div v-if="positions?.length" class="space-y-6">
            <div
              v-for="position in positions"
              :key="position.id"
              class="border border-gray-200 dark:border-gray-700 rounded-lg p-4"
            >
              <div class="flex items-center justify-between mb-4">
                <div>
                  <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                    {{ position.position?.name || 'Position' }}
                  </h3>
                  <p v-if="position.location" class="text-sm text-gray-500 dark:text-gray-400">
                    {{ position.location }}
                  </p>
                </div>
                <div class="text-right">
                  <div class="text-sm text-gray-500 dark:text-gray-400">
                    {{ position.seats_available }} seat{{ position.seats_available > 1 ? 's' : '' }} available
                  </div>
                  <div class="text-sm text-gray-500 dark:text-gray-400">
                    {{ position.candidate_count }} candidate{{ position.candidate_count !== 1 ? 's' : '' }}
                  </div>
                </div>
              </div>

              <!-- Position candidates would be loaded separately if needed -->
              <div class="text-sm text-gray-500 dark:text-gray-400">
                <NuxtLink
                  :to="`/elections/${slug}/position/${position.id}`"
                  class="text-blue-600 hover:text-blue-800 dark:text-blue-400"
                >
                  View candidates
                </NuxtLink>
              </div>
            </div>
          </div>

          <div v-else class="text-center py-8 text-gray-500 dark:text-gray-400">
            No positions have been added yet.
          </div>
        </div>

        <!-- Election Candidates Preview -->
        <div v-if="election.candidates?.length" class="mt-8 bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-6">Featured Candidates</h2>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
              v-for="candidate in election.candidates.slice(0, 6)"
              :key="candidate.id"
              class="flex items-center gap-4 p-4 bg-gray-50 dark:bg-gray-700 rounded-lg"
            >
              <div class="flex-shrink-0">
                <div class="w-12 h-12 bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center">
                  <span v-if="candidate.politician?.photo">
                    <img
                      :src="candidate.politician.photo"
                      :alt="candidate.politician.name"
                      class="w-12 h-12 rounded-full object-cover"
                    >
                  </span>
                  <span v-else class="text-gray-500 dark:text-gray-400 text-lg font-bold">
                    {{ candidate.politician?.name?.charAt(0) || '?' }}
                  </span>
                </div>
              </div>
              <div class="flex-1 min-w-0">
                <NuxtLink
                  v-if="candidate.politician"
                  :to="`/politician/${candidate.politician.slug}`"
                  class="font-medium text-gray-900 dark:text-white hover:text-blue-600 truncate block"
                >
                  {{ candidate.ballot_name || candidate.politician.name }}
                </NuxtLink>
                <div class="flex items-center gap-2 mt-1">
                  <span
                    v-if="candidate.ballot_number"
                    class="text-sm font-bold text-blue-600"
                  >
                    #{{ candidate.ballot_number }}
                  </span>
                  <span
                    v-if="candidate.party"
                    class="text-xs px-2 py-0.5 rounded-full"
                    :style="{ backgroundColor: candidate.party.color + '20', color: candidate.party.color }"
                  >
                    {{ candidate.party.abbreviation || candidate.party.name }}
                  </span>
                  <span
                    :class="[
                      'text-xs px-2 py-0.5 rounded-full',
                      getCandidateStatusColor(candidate.status)
                    ]"
                  >
                    {{ candidate.status }}
                  </span>
                </div>
              </div>
              <div v-if="candidate.is_winner" class="flex-shrink-0">
                <span class="text-green-500 text-xl" title="Winner">
                  <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                  </svg>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
