<script setup lang="ts">
import type { PollListItem, PollCategory, PollStatus, PaginatedPolls } from '~/types'

definePageMeta({
  layout: 'admin'
})

useSeoMeta({
  title: 'Manage Polls - Admin - Pulpulitiko'
})

const api = useApi()
const { authHeaders } = useAuth()

// Filters
const page = ref(1)
const perPage = ref(20)
const categoryFilter = ref<PollCategory | ''>('')
const statusFilter = ref<PollStatus | ''>('')
const searchQuery = ref('')

// Fetch polls
const { data: pollsData, pending, refresh } = await useAsyncData<PaginatedPolls>(
  'admin-polls',
  () => api.adminGetPolls(
    authHeaders.value!,
    {
      category: categoryFilter.value || undefined,
      status: statusFilter.value || undefined,
      search: searchQuery.value || undefined
    },
    page.value,
    perPage.value
  ),
  { watch: [page, categoryFilter, statusFilter] }
)

const polls = computed(() => pollsData.value?.polls || [])
const totalPages = computed(() => pollsData.value?.total_pages || 1)
const total = computed(() => pollsData.value?.total || 0)

// Search
const handleSearch = () => {
  page.value = 1
  refresh()
}

// Category options
const categories: { value: PollCategory | ''; label: string }[] = [
  { value: '', label: 'All Categories' },
  { value: 'general', label: 'General' },
  { value: 'election', label: 'Elections' },
  { value: 'legislation', label: 'Legislation' },
  { value: 'politician', label: 'Politicians' },
  { value: 'policy', label: 'Policy' },
  { value: 'local_issue', label: 'Local Issues' },
  { value: 'national_issue', label: 'National Issues' }
]

// Status options
const statuses: { value: PollStatus | ''; label: string }[] = [
  { value: '', label: 'All Statuses' },
  { value: 'draft', label: 'Draft' },
  { value: 'pending_approval', label: 'Pending Approval' },
  { value: 'active', label: 'Active' },
  { value: 'closed', label: 'Closed' },
  { value: 'rejected', label: 'Rejected' }
]

// Actions
const actionPoll = ref<PollListItem | null>(null)
const showApproveModal = ref(false)
const showRejectModal = ref(false)
const rejectReason = ref('')
const processing = ref(false)

const approvePoll = async () => {
  if (!actionPoll.value) return
  processing.value = true
  try {
    await api.approvePoll(actionPoll.value.id, true, undefined, authHeaders.value!)
    await refresh()
    showApproveModal.value = false
    actionPoll.value = null
  } catch (err) {
    console.error('Failed to approve poll:', err)
  } finally {
    processing.value = false
  }
}

const rejectPoll = async () => {
  if (!actionPoll.value) return
  processing.value = true
  try {
    await api.approvePoll(actionPoll.value.id, false, rejectReason.value || undefined, authHeaders.value!)
    await refresh()
    showRejectModal.value = false
    rejectReason.value = ''
    actionPoll.value = null
  } catch (err) {
    console.error('Failed to reject poll:', err)
  } finally {
    processing.value = false
  }
}

const closePoll = async (poll: PollListItem) => {
  if (!confirm('Are you sure you want to close this poll?')) return
  try {
    await api.closePoll(poll.id, authHeaders.value!)
    await refresh()
  } catch (err) {
    console.error('Failed to close poll:', err)
  }
}

const deletePoll = async (poll: PollListItem) => {
  if (!confirm('Are you sure you want to delete this poll? This action cannot be undone.')) return
  try {
    await api.adminDeletePoll(poll.id, authHeaders.value!)
    await refresh()
  } catch (err) {
    console.error('Failed to delete poll:', err)
  }
}

const toggleFeatured = async (poll: PollListItem) => {
  try {
    await api.adminUpdatePoll(poll.id, { is_featured: !poll.is_featured }, authHeaders.value!)
    await refresh()
  } catch (err) {
    console.error('Failed to toggle featured:', err)
  }
}

// Helpers
const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

const getStatusColor = (status: PollStatus) => {
  const colors: Record<PollStatus, string> = {
    draft: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
    pending_approval: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    active: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    closed: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
    rejected: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
  }
  return colors[status] || 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
}

const getStatusLabel = (status: PollStatus) => {
  return statuses.find(s => s.value === status)?.label || status
}

const getCategoryLabel = (cat: PollCategory) => {
  return categories.find(c => c.value === cat)?.label || cat
}
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Polls Management</h1>
        <p class="text-gray-500 dark:text-gray-400">Manage community polls, approve submissions, and moderate content</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 mb-6">
      <div class="flex flex-col md:flex-row gap-4">
        <!-- Search -->
        <div class="flex-1">
          <form @submit.prevent="handleSearch" class="flex">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search polls..."
              class="flex-1 rounded-l-lg border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500"
            />
            <button
              type="submit"
              class="px-4 py-2 bg-blue-600 text-white rounded-r-lg hover:bg-blue-700"
            >
              Search
            </button>
          </form>
        </div>

        <!-- Category Filter -->
        <div class="w-full md:w-48">
          <select
            v-model="categoryFilter"
            class="w-full rounded-lg border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500"
          >
            <option v-for="cat in categories" :key="cat.value" :value="cat.value">
              {{ cat.label }}
            </option>
          </select>
        </div>

        <!-- Status Filter -->
        <div class="w-full md:w-48">
          <select
            v-model="statusFilter"
            class="w-full rounded-lg border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white focus:ring-blue-500 focus:border-blue-500"
          >
            <option v-for="status in statuses" :key="status.value" :value="status.value">
              {{ status.label }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- Stats -->
    <div class="mb-4 text-sm text-gray-500 dark:text-gray-400">
      Showing {{ polls.length }} of {{ total }} polls
    </div>

    <!-- Polls Table -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden">
      <div v-if="pending" class="p-8 text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
      </div>

      <div v-else-if="polls.length === 0" class="p-8 text-center text-gray-500 dark:text-gray-400">
        No polls found.
      </div>

      <table v-else class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead class="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Poll
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Category
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Status
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Votes
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Created
            </th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
          <tr v-for="poll in polls" :key="poll.id" class="hover:bg-gray-50 dark:hover:bg-gray-700/50">
            <td class="px-6 py-4">
              <div class="flex items-center">
                <button
                  @click="toggleFeatured(poll)"
                  :class="[
                    'mr-2 transition-colors',
                    poll.is_featured ? 'text-yellow-500' : 'text-gray-300 dark:text-gray-600 hover:text-yellow-400'
                  ]"
                  :title="poll.is_featured ? 'Remove from featured' : 'Add to featured'"
                >
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                  </svg>
                </button>
                <div>
                  <NuxtLink
                    :to="`/polls/${poll.slug}`"
                    target="_blank"
                    class="font-medium text-gray-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400 line-clamp-1"
                  >
                    {{ poll.title }}
                  </NuxtLink>
                  <p v-if="poll.author" class="text-sm text-gray-500 dark:text-gray-400">
                    by {{ poll.author.name }}
                  </p>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span class="text-sm text-gray-900 dark:text-gray-200">{{ getCategoryLabel(poll.category) }}</span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="['px-2 py-1 text-xs font-medium rounded-full', getStatusColor(poll.status)]">
                {{ getStatusLabel(poll.status) }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-200">
              {{ poll.total_votes }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
              {{ formatDate(poll.created_at) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <div class="flex items-center justify-end space-x-2">
                <!-- Approve/Reject for pending polls -->
                <template v-if="poll.status === 'pending_approval'">
                  <button
                    @click="actionPoll = poll; showApproveModal = true"
                    class="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300"
                    title="Approve"
                  >
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                  </button>
                  <button
                    @click="actionPoll = poll; showRejectModal = true"
                    class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                    title="Reject"
                  >
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </template>

                <!-- Close for active polls -->
                <button
                  v-if="poll.status === 'active'"
                  @click="closePoll(poll)"
                  class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
                  title="Close Poll"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                </button>

                <!-- Delete -->
                <button
                  @click="deletePoll(poll)"
                  class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                  title="Delete"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="mt-6 flex justify-center">
      <nav class="flex items-center space-x-2">
        <button
          @click="page = page - 1"
          :disabled="page <= 1"
          class="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Previous
        </button>

        <span class="text-gray-600 dark:text-gray-300">
          Page {{ page }} of {{ totalPages }}
        </span>

        <button
          @click="page = page + 1"
          :disabled="page >= totalPages"
          class="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Next
        </button>
      </nav>
    </div>

    <!-- Approve Modal -->
    <div
      v-if="showApproveModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="showApproveModal = false"
    >
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full mx-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Approve Poll</h3>
        <p class="text-gray-600 dark:text-gray-300 mb-6">
          Are you sure you want to approve this poll? It will become visible to all users.
        </p>
        <div class="flex justify-end space-x-3">
          <button
            @click="showApproveModal = false"
            class="px-4 py-2 text-gray-600 dark:text-gray-300 hover:text-gray-800 dark:hover:text-white"
          >
            Cancel
          </button>
          <button
            @click="approvePoll"
            :disabled="processing"
            class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
          >
            {{ processing ? 'Approving...' : 'Approve' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Reject Modal -->
    <div
      v-if="showRejectModal"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="showRejectModal = false"
    >
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full mx-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Reject Poll</h3>
        <p class="text-gray-600 dark:text-gray-300 mb-4">
          Please provide a reason for rejection (optional but recommended).
        </p>
        <textarea
          v-model="rejectReason"
          rows="3"
          placeholder="Reason for rejection..."
          class="w-full rounded-lg border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white focus:ring-red-500 focus:border-red-500 mb-4"
        ></textarea>
        <div class="flex justify-end space-x-3">
          <button
            @click="showRejectModal = false; rejectReason = ''"
            class="px-4 py-2 text-gray-600 dark:text-gray-300 hover:text-gray-800 dark:hover:text-white"
          >
            Cancel
          </button>
          <button
            @click="rejectPoll"
            :disabled="processing"
            class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
          >
            {{ processing ? 'Rejecting...' : 'Reject' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
