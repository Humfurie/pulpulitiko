<script setup lang="ts">
import type { Poll, PollOption, PollCategory } from '~/types'

definePageMeta({
  layout: 'default'
})

const route = useRoute()
const api = useApi()
const { user, getAuthHeaders } = useAuth()

const slug = computed(() => route.params.slug as string)

// Fetch poll
const { data: poll, pending, error, refresh: _refresh } = await useAsyncData<Poll>(
  `poll-${slug.value}`,
  () => api.getPollBySlug(slug.value)
)

// SEO
useSeoMeta({
  title: computed(() => poll.value ? `${poll.value.title} - Polls - Pulpulitiko` : 'Poll - Pulpulitiko'),
  description: computed(() => poll.value?.description || 'Vote on this poll and share your opinion.')
})

// Voting state
const selectedOption = ref<string | null>(null)
const voting = ref(false)
const voteError = ref<string | null>(null)
const hasVoted = computed(() => !!poll.value?.user_vote)

// Initialize selected option if user has already voted
watch(poll, (newPoll) => {
  if (newPoll?.user_vote) {
    selectedOption.value = newPoll.user_vote
  }
}, { immediate: true })

// Check if poll is active
const isPollActive = computed(() => {
  if (!poll.value) return false
  if (poll.value.status !== 'active') return false

  const now = new Date()
  if (poll.value.starts_at && new Date(poll.value.starts_at) > now) return false
  if (poll.value.ends_at && new Date(poll.value.ends_at) < now) return false

  return true
})

// Should show results
const showResults = computed(() => {
  if (!poll.value) return false
  if (hasVoted.value) return true
  if (poll.value.show_results_before_vote) return true
  if (!isPollActive.value) return true
  return false
})

// Cast vote
const castVote = async () => {
  if (!selectedOption.value || voting.value) return

  voting.value = true
  voteError.value = null

  try {
    const authHeaders = getAuthHeaders()
    const response = await api.castVote(poll.value!.id, selectedOption.value, Object.keys(authHeaders).length > 0 ? authHeaders : undefined)
    if (response.success && response.results) {
      // Update poll with new results
      if (poll.value && poll.value.options) {
        poll.value.total_votes = response.results.total_votes
        poll.value.options = response.results.options
        poll.value.user_vote = selectedOption.value
      }
    } else {
      voteError.value = response.message || 'Failed to cast vote'
    }
  } catch (err: unknown) {
    voteError.value = err instanceof Error ? err.message : 'Failed to cast vote'
  } finally {
    voting.value = false
  }
}

// Format helpers
const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

const formatDateTime = (dateStr: string) => {
  return new Date(dateStr).toLocaleString('en-PH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getCategoryLabel = (cat: PollCategory) => {
  const labels: Record<PollCategory, string> = {
    general: 'General',
    election: 'Elections',
    legislation: 'Legislation',
    politician: 'Politicians',
    policy: 'Policy',
    local_issue: 'Local Issues',
    national_issue: 'National Issues'
  }
  return labels[cat] || cat
}

const getCategoryColor = (cat: PollCategory) => {
  const colors: Record<PollCategory, string> = {
    general: 'bg-gray-100 text-gray-800',
    election: 'bg-blue-100 text-blue-800',
    legislation: 'bg-purple-100 text-purple-800',
    politician: 'bg-green-100 text-green-800',
    policy: 'bg-yellow-100 text-yellow-800',
    local_issue: 'bg-orange-100 text-orange-800',
    national_issue: 'bg-red-100 text-red-800'
  }
  return colors[cat] || 'bg-gray-100 text-gray-800'
}

const getOptionPercentage = (option: PollOption) => {
  if (!poll.value || poll.value.total_votes === 0) return 0
  return Math.round((option.vote_count / poll.value.total_votes) * 100)
}

const getWinningOption = () => {
  if (!poll.value?.options || poll.value.options.length === 0) return null
  const firstOption = poll.value.options[0]
  if (!firstOption) return null
  return poll.value.options.reduce((max, opt) => opt.vote_count > max.vote_count ? opt : max, firstOption)
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 py-8">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
      <!-- Back link -->
      <NuxtLink to="/polls" class="inline-flex items-center text-blue-600 hover:text-blue-800 mb-6">
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
        Back to Polls
      </NuxtLink>

      <!-- Loading State -->
      <div v-if="pending" class="bg-white rounded-lg shadow-sm p-8 animate-pulse">
        <div class="h-6 bg-gray-200 rounded w-1/4 mb-4"/>
        <div class="h-8 bg-gray-200 rounded w-3/4 mb-6"/>
        <div class="space-y-4">
          <div v-for="n in 4" :key="n" class="h-12 bg-gray-200 rounded"/>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-8 text-center">
        <svg class="w-12 h-12 text-red-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <h3 class="text-lg font-medium text-red-900 mb-2">Poll Not Found</h3>
        <p class="text-red-600 mb-4">The poll you're looking for doesn't exist or has been removed.</p>
        <NuxtLink to="/polls" class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 inline-block">
          Browse Polls
        </NuxtLink>
      </div>

      <!-- Poll Content -->
      <div v-else-if="poll" class="space-y-6">
        <!-- Main Poll Card -->
        <div class="bg-white rounded-lg shadow-sm p-6 md:p-8">
          <!-- Header -->
          <div class="flex items-start justify-between mb-6">
            <div class="flex items-center space-x-3">
              <span :class="['px-3 py-1 text-sm font-medium rounded-full', getCategoryColor(poll.category)]">
                {{ getCategoryLabel(poll.category) }}
              </span>
              <span v-if="poll.is_featured" class="flex items-center text-yellow-500">
                <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
                <span class="ml-1 text-sm font-medium">Featured</span>
              </span>
            </div>

            <!-- Status Badge -->
            <span
              v-if="!isPollActive"
              class="px-3 py-1 text-sm font-medium rounded-full"
              :class="{
                'bg-red-100 text-red-800': poll.status === 'closed' || poll.status === 'rejected',
                'bg-yellow-100 text-yellow-800': poll.status === 'pending_approval',
                'bg-gray-100 text-gray-800': poll.status === 'draft'
              }"
            >
              {{ poll.status === 'closed' ? 'Closed' : poll.status === 'pending_approval' ? 'Pending' : poll.status }}
            </span>
          </div>

          <!-- Title & Description -->
          <h1 class="text-2xl md:text-3xl font-bold text-gray-900 mb-4">{{ poll.title }}</h1>
          <p v-if="poll.description" class="text-gray-600 mb-6">{{ poll.description }}</p>

          <!-- Poll Meta -->
          <div class="flex flex-wrap items-center gap-4 text-sm text-gray-500 mb-6 pb-6 border-b">
            <div v-if="poll.author" class="flex items-center">
              <img
                v-if="poll.author.avatar"
                :src="poll.author.avatar"
                :alt="poll.author.name"
                class="w-8 h-8 rounded-full mr-2"
              >
              <span v-else class="w-8 h-8 rounded-full bg-gray-300 mr-2 flex items-center justify-center text-sm text-white">
                {{ poll.author.name.charAt(0) }}
              </span>
              <span>{{ poll.author.name }}</span>
            </div>
            <span>{{ formatDate(poll.created_at) }}</span>
            <span class="flex items-center">
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              {{ poll.view_count }} views
            </span>
            <span v-if="poll.ends_at" class="flex items-center text-orange-600 font-medium">
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Ends {{ formatDateTime(poll.ends_at) }}
            </span>
          </div>

          <!-- Vote Error -->
          <div v-if="voteError" class="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-600">
            {{ voteError }}
          </div>

          <!-- Poll Options -->
          <div class="space-y-3">
            <div
              v-for="option in poll.options"
              :key="option.id"
              :class="[
                'relative rounded-lg border-2 transition-all overflow-hidden',
                hasVoted || !isPollActive
                  ? 'cursor-default'
                  : 'cursor-pointer hover:border-blue-400',
                selectedOption === option.id && !hasVoted
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-200',
                poll.user_vote === option.id ? 'border-green-500 bg-green-50' : ''
              ]"
              @click="!hasVoted && isPollActive && (selectedOption = option.id)"
            >
              <!-- Progress bar background -->
              <div
                v-if="showResults"
                class="absolute inset-0 transition-all duration-500"
                :class="[
                  getWinningOption()?.id === option.id ? 'bg-blue-100' : 'bg-gray-100'
                ]"
                :style="{ width: `${getOptionPercentage(option)}%` }"
              />

              <!-- Option content -->
              <div class="relative px-4 py-3 flex items-center justify-between">
                <div class="flex items-center">
                  <!-- Radio button (when voting) -->
                  <div
                    v-if="!hasVoted && isPollActive"
                    :class="[
                      'w-5 h-5 rounded-full border-2 mr-3 flex items-center justify-center',
                      selectedOption === option.id
                        ? 'border-blue-500 bg-blue-500'
                        : 'border-gray-300'
                    ]"
                  >
                    <svg
                      v-if="selectedOption === option.id"
                      class="w-3 h-3 text-white"
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                    </svg>
                  </div>

                  <!-- Checkmark for user's vote -->
                  <svg
                    v-if="poll.user_vote === option.id"
                    class="w-5 h-5 text-green-600 mr-3"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                  </svg>

                  <span class="font-medium text-gray-900">{{ option.text }}</span>
                </div>

                <!-- Results -->
                <div v-if="showResults" class="flex items-center space-x-2">
                  <span class="text-sm text-gray-500">{{ option.vote_count }} votes</span>
                  <span class="text-lg font-semibold" :class="getWinningOption()?.id === option.id ? 'text-blue-600' : 'text-gray-700'">
                    {{ getOptionPercentage(option) }}%
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Vote Button -->
          <div v-if="isPollActive && !hasVoted" class="mt-6">
            <button
              :disabled="!selectedOption || voting"
              class="w-full py-3 px-6 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              @click="castVote"
            >
              <span v-if="voting" class="flex items-center justify-center">
                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                </svg>
                Submitting Vote...
              </span>
              <span v-else>Cast Vote</span>
            </button>
            <p v-if="!user" class="mt-2 text-sm text-gray-500 text-center">
              You're voting anonymously. <NuxtLink to="/login" class="text-blue-600 hover:underline">Sign in</NuxtLink> to track your votes.
            </p>
          </div>

          <!-- Total Votes -->
          <div class="mt-6 text-center text-gray-500">
            <span class="font-semibold text-gray-900">{{ poll.total_votes }}</span> total votes
          </div>
        </div>

        <!-- Related Info -->
        <div v-if="poll.politician || poll.election || poll.bill" class="bg-white rounded-lg shadow-sm p-6">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">Related To</h2>
          <div class="space-y-3">
            <NuxtLink
              v-if="poll.politician"
              :to="`/politicians/${poll.politician.slug}`"
              class="flex items-center p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
            >
              <img
                v-if="poll.politician.photo"
                :src="poll.politician.photo"
                :alt="poll.politician.name"
                class="w-10 h-10 rounded-full object-cover mr-3"
              >
              <div v-else class="w-10 h-10 rounded-full bg-gray-300 mr-3 flex items-center justify-center text-white font-medium">
                {{ poll.politician.name.charAt(0) }}
              </div>
              <div>
                <p class="text-xs text-gray-500 uppercase tracking-wide">Politician</p>
                <p class="font-medium text-gray-900">{{ poll.politician.name }}</p>
              </div>
            </NuxtLink>

            <NuxtLink
              v-if="poll.election"
              :to="`/elections/${poll.election.slug}`"
              class="flex items-center p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
            >
              <div class="w-10 h-10 rounded-lg bg-blue-100 mr-3 flex items-center justify-center">
                <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
                </svg>
              </div>
              <div>
                <p class="text-xs text-gray-500 uppercase tracking-wide">Election</p>
                <p class="font-medium text-gray-900">{{ poll.election.name }}</p>
              </div>
            </NuxtLink>

            <NuxtLink
              v-if="poll.bill"
              :to="`/legislation/${poll.bill.slug}`"
              class="flex items-center p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
            >
              <div class="w-10 h-10 rounded-lg bg-purple-100 mr-3 flex items-center justify-center">
                <svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              </div>
              <div>
                <p class="text-xs text-gray-500 uppercase tracking-wide">Bill</p>
                <p class="font-medium text-gray-900">{{ poll.bill.bill_number }}: {{ poll.bill.title }}</p>
              </div>
            </NuxtLink>
          </div>
        </div>

        <!-- Comments Section -->
        <div class="bg-white rounded-lg shadow-sm p-6">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">
            Discussion ({{ poll.comment_count }})
          </h2>
          <p class="text-gray-500 text-center py-8">
            Comments feature coming soon.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
