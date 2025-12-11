<script setup lang="ts">
import type { BillFilter, BillStatus, LegislativeChamber, LegislativeSessionListItem } from '~/types'

const api = useApi()

// Filter state
const chamber = ref<LegislativeChamber | ''>('')
const status = ref<BillStatus | ''>('')
const sessionId = ref('')
const topicId = ref('')
const search = ref('')
const currentPage = ref(1)
const perPage = 20
const showMobileFilters = ref(false)

// Data
const { data: sessionsData } = useAsyncData('sessions', () => api.getLegislativeSessions())
const { data: topicsData } = useAsyncData('topics', () => api.getBillTopics())

const sessions = computed(() => sessionsData.value || [])
const topics = computed(() => topicsData.value || [])

// Build filter
const filter = computed<BillFilter>(() => ({
  chamber: chamber.value || undefined,
  status: status.value || undefined,
  session_id: sessionId.value || undefined,
  topic_id: topicId.value || undefined,
  search: search.value || undefined
}))

// Fetch bills
const { data: billsData, pending } = useAsyncData(
  'bills',
  () => api.getBills(filter.value, currentPage.value, perPage),
  { watch: [filter, currentPage] }
)

const bills = computed(() => billsData.value?.bills || [])
const total = computed(() => billsData.value?.total || 0)
const totalPages = computed(() => billsData.value?.total_pages || 1)

// Status display mapping
const statusLabels: Record<BillStatus, string> = {
  filed: 'Filed',
  pending_committee: 'Pending Committee',
  in_committee: 'In Committee',
  reported_out: 'Reported Out',
  pending_second_reading: 'Pending 2nd Reading',
  approved_second_reading: 'Approved 2nd Reading',
  pending_third_reading: 'Pending 3rd Reading',
  approved_third_reading: 'Approved 3rd Reading',
  transmitted: 'Transmitted',
  consolidated: 'Consolidated',
  ratified: 'Ratified',
  signed_into_law: 'Signed Into Law',
  vetoed: 'Vetoed',
  lapsed: 'Lapsed',
  withdrawn: 'Withdrawn',
  archived: 'Archived'
}

const statusColors: Record<BillStatus, string> = {
  filed: 'bg-gray-100 text-gray-800',
  pending_committee: 'bg-yellow-100 text-yellow-800',
  in_committee: 'bg-blue-100 text-blue-800',
  reported_out: 'bg-indigo-100 text-indigo-800',
  pending_second_reading: 'bg-purple-100 text-purple-800',
  approved_second_reading: 'bg-purple-200 text-purple-900',
  pending_third_reading: 'bg-pink-100 text-pink-800',
  approved_third_reading: 'bg-pink-200 text-pink-900',
  transmitted: 'bg-cyan-100 text-cyan-800',
  consolidated: 'bg-teal-100 text-teal-800',
  ratified: 'bg-emerald-100 text-emerald-800',
  signed_into_law: 'bg-green-100 text-green-800',
  vetoed: 'bg-red-100 text-red-800',
  lapsed: 'bg-orange-100 text-orange-800',
  withdrawn: 'bg-stone-100 text-stone-800',
  archived: 'bg-slate-100 text-slate-800'
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

function formatSession(session: LegislativeSessionListItem): string {
  const ordinal = ['', '1st', '2nd', '3rd'][session.session_number] || `${session.session_number}th`
  return `${session.congress_number}th Congress - ${ordinal} ${session.session_type} session`
}

function clearFilters() {
  chamber.value = ''
  status.value = ''
  sessionId.value = ''
  topicId.value = ''
  search.value = ''
  currentPage.value = 1
}

useSeoMeta({
  title: 'Legislation Tracker - Pulpulitiko',
  description: 'Track Philippine legislative bills, voting records, and committee activities'
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <div class="bg-white border-b">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 class="text-3xl font-bold text-gray-900">Legislation Tracker</h1>
        <p class="mt-2 text-gray-600">
          Track legislative bills, monitor their progress, and see how your representatives vote
        </p>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Mobile Filter Toggle -->
      <button
        class="lg:hidden w-full mb-4 flex items-center justify-center gap-2 px-4 py-3 bg-white rounded-lg shadow text-gray-700 font-medium"
        @click="showMobileFilters = !showMobileFilters"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
        </svg>
        {{ showMobileFilters ? 'Hide Filters' : 'Show Filters' }}
        <span v-if="chamber || status || sessionId || topicId || search" class="ml-1 px-2 py-0.5 bg-red-100 text-red-700 text-xs rounded-full">
          Active
        </span>
      </button>

      <div class="lg:grid lg:grid-cols-4 lg:gap-8">
        <!-- Filters Sidebar -->
        <div :class="['lg:col-span-1', showMobileFilters ? 'block' : 'hidden lg:block']">
          <div class="bg-white rounded-lg shadow p-6 sticky top-4 mb-6 lg:mb-0">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-lg font-semibold">Filters</h2>
              <button
                v-if="chamber || status || sessionId || topicId || search"
                class="text-sm text-red-600 hover:text-red-700"
                @click="clearFilters"
              >
                Clear all
              </button>
            </div>

            <div class="space-y-4">
              <!-- Search -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Search</label>
                <input
                  v-model="search"
                  type="text"
                  placeholder="Bill number or title..."
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-red-500 focus:border-red-500"
                />
              </div>

              <!-- Chamber -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Chamber</label>
                <select
                  v-model="chamber"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-red-500 focus:border-red-500"
                >
                  <option value="">All Chambers</option>
                  <option value="senate">Senate</option>
                  <option value="house">House of Representatives</option>
                </select>
              </div>

              <!-- Session -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Session</label>
                <select
                  v-model="sessionId"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-red-500 focus:border-red-500"
                >
                  <option value="">All Sessions</option>
                  <option v-for="s in sessions" :key="s.id" :value="s.id">
                    {{ formatSession(s) }} ({{ s.bill_count }})
                  </option>
                </select>
              </div>

              <!-- Status -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
                <select
                  v-model="status"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-red-500 focus:border-red-500"
                >
                  <option value="">All Statuses</option>
                  <option v-for="(label, key) in statusLabels" :key="key" :value="key">
                    {{ label }}
                  </option>
                </select>
              </div>

              <!-- Topic -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Topic</label>
                <select
                  v-model="topicId"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-red-500 focus:border-red-500"
                >
                  <option value="">All Topics</option>
                  <option v-for="t in topics" :key="t.id" :value="t.id">
                    {{ t.name }} ({{ t.bill_count }})
                  </option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <!-- Bills List -->
        <div class="lg:col-span-3 mt-8 lg:mt-0">
          <!-- Results header -->
          <div class="flex items-center justify-between mb-4">
            <p class="text-gray-600">
              <span v-if="pending">Loading...</span>
              <span v-else>{{ total }} bill{{ total !== 1 ? 's' : '' }} found</span>
            </p>
          </div>

          <!-- Bills Grid -->
          <div v-if="pending" class="space-y-4">
            <div v-for="i in 5" :key="i" class="bg-white rounded-lg shadow p-6 animate-pulse">
              <div class="h-4 bg-gray-200 rounded w-1/4 mb-3" />
              <div class="h-6 bg-gray-200 rounded w-3/4 mb-4" />
              <div class="h-4 bg-gray-200 rounded w-1/2" />
            </div>
          </div>

          <div v-else-if="bills.length === 0" class="bg-white rounded-lg shadow p-12 text-center">
            <svg class="w-16 h-16 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <h3 class="mt-4 text-lg font-medium text-gray-900">No bills found</h3>
            <p class="mt-2 text-gray-500">Try adjusting your filters or search terms</p>
          </div>

          <div v-else class="space-y-4">
            <NuxtLink
              v-for="bill in bills"
              :key="bill.id"
              :to="`/legislation/${bill.slug}`"
              class="block bg-white rounded-lg shadow hover:shadow-md transition-shadow p-6"
            >
              <div class="flex items-start justify-between gap-4">
                <div class="flex-1 min-w-0">
                  <!-- Bill Number & Chamber -->
                  <div class="flex items-center gap-2 text-sm text-gray-500 mb-1">
                    <span class="font-mono font-medium text-gray-700">{{ bill.bill_number }}</span>
                    <span>•</span>
                    <span class="capitalize">{{ bill.chamber }}</span>
                  </div>

                  <!-- Title -->
                  <h3 class="text-lg font-semibold text-gray-900 line-clamp-2">
                    {{ bill.short_title || bill.title }}
                  </h3>

                  <!-- Topics -->
                  <div v-if="bill.topic_names && bill.topic_names.length > 0" class="mt-2 flex flex-wrap gap-1">
                    <span
                      v-for="topic in bill.topic_names.slice(0, 3)"
                      :key="topic"
                      class="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded"
                    >
                      {{ topic }}
                    </span>
                    <span v-if="bill.topic_names.length > 3" class="text-xs text-gray-500">
                      +{{ bill.topic_names.length - 3 }} more
                    </span>
                  </div>

                  <!-- Meta -->
                  <div class="mt-3 flex items-center gap-4 text-sm text-gray-500">
                    <span>Filed {{ formatDate(bill.filed_date) }}</span>
                    <span v-if="bill.author_count > 0">
                      {{ bill.author_count }} author{{ bill.author_count !== 1 ? 's' : '' }}
                    </span>
                  </div>
                </div>

                <!-- Status Badge -->
                <span
                  :class="[
                    'shrink-0 px-3 py-1 text-xs font-medium rounded-full',
                    statusColors[bill.status]
                  ]"
                >
                  {{ statusLabels[bill.status] }}
                </span>
              </div>
            </NuxtLink>
          </div>

          <!-- Pagination -->
          <div v-if="totalPages > 1" class="mt-8 flex justify-center">
            <nav class="flex items-center gap-2">
              <button
                :disabled="currentPage === 1"
                class="px-3 py-2 text-sm font-medium text-gray-700 bg-white border rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                @click="currentPage--"
              >
                Previous
              </button>

              <span class="px-4 py-2 text-sm text-gray-600">
                Page {{ currentPage }} of {{ totalPages }}
              </span>

              <button
                :disabled="currentPage === totalPages"
                class="px-3 py-2 text-sm font-medium text-gray-700 bg-white border rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                @click="currentPage++"
              >
                Next
              </button>
            </nav>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
