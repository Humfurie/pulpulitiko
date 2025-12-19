<script setup lang="ts">
import type { BillStatus } from '~/types'

const api = useApi()
const route = useRoute()
const slug = computed(() => route.params.slug as string)

const { data: bill, pending, error } = useAsyncData(
  `bill-${slug.value}`,
  () => api.getBillBySlug(slug.value)
)

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
  filed: 'bg-gray-100 text-gray-800 border-gray-300',
  pending_committee: 'bg-yellow-100 text-yellow-800 border-yellow-300',
  in_committee: 'bg-blue-100 text-blue-800 border-blue-300',
  reported_out: 'bg-indigo-100 text-indigo-800 border-indigo-300',
  pending_second_reading: 'bg-purple-100 text-purple-800 border-purple-300',
  approved_second_reading: 'bg-purple-200 text-purple-900 border-purple-400',
  pending_third_reading: 'bg-pink-100 text-pink-800 border-pink-300',
  approved_third_reading: 'bg-pink-200 text-pink-900 border-pink-400',
  transmitted: 'bg-cyan-100 text-cyan-800 border-cyan-300',
  consolidated: 'bg-teal-100 text-teal-800 border-teal-300',
  ratified: 'bg-emerald-100 text-emerald-800 border-emerald-300',
  signed_into_law: 'bg-green-100 text-green-800 border-green-300',
  vetoed: 'bg-red-100 text-red-800 border-red-300',
  lapsed: 'bg-orange-100 text-orange-800 border-orange-300',
  withdrawn: 'bg-stone-100 text-stone-800 border-stone-300',
  archived: 'bg-slate-100 text-slate-800 border-slate-300'
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

function formatShortDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

// Active tab
const activeTab = ref<'overview' | 'timeline' | 'votes' | 'committees'>('overview')

useSeoMeta({
  title: () => bill.value ? `${bill.value.bill_number} - ${bill.value.short_title || bill.value.title}` : 'Bill Details',
  description: () => bill.value?.summary || 'View details about this legislative bill'
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Loading -->
    <div v-if="pending" class="max-w-5xl mx-auto px-4 py-12">
      <div class="animate-pulse space-y-4">
        <div class="h-8 bg-gray-200 rounded w-1/3" />
        <div class="h-12 bg-gray-200 rounded w-2/3" />
        <div class="h-4 bg-gray-200 rounded w-1/4" />
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error || !bill" class="max-w-5xl mx-auto px-4 py-12 text-center">
      <svg class="w-16 h-16 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <h2 class="mt-4 text-xl font-semibold text-gray-900">Bill not found</h2>
      <p class="mt-2 text-gray-600">The bill you're looking for doesn't exist or has been removed.</p>
      <NuxtLink to="/legislation" class="mt-4 inline-block text-red-600 hover:text-red-700">
        &larr; Back to Legislation
      </NuxtLink>
    </div>

    <!-- Bill Content -->
    <template v-else>
      <!-- Header -->
      <div class="bg-white border-b">
        <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <!-- Breadcrumb -->
          <nav class="mb-4">
            <NuxtLink to="/legislation" class="text-sm text-gray-500 hover:text-gray-700">
              &larr; Back to Legislation
            </NuxtLink>
          </nav>

          <!-- Bill Number & Chamber -->
          <div class="flex items-center gap-3 mb-2">
            <span class="font-mono text-lg font-bold text-gray-700">{{ bill.bill_number }}</span>
            <span
:class="[
              'px-2 py-0.5 text-xs font-medium rounded capitalize',
              bill.chamber === 'senate' ? 'bg-blue-100 text-blue-700' : 'bg-amber-100 text-amber-700'
            ]">
              {{ bill.chamber }}
            </span>
          </div>

          <!-- Title -->
          <h1 class="text-2xl md:text-3xl font-bold text-gray-900">
            {{ bill.title }}
          </h1>

          <!-- Short Title if different -->
          <p v-if="bill.short_title && bill.short_title !== bill.title" class="mt-2 text-lg text-gray-600">
            {{ bill.short_title }}
          </p>

          <!-- Status & Dates -->
          <div class="mt-4 flex flex-wrap items-center gap-4">
            <span
:class="[
              'px-4 py-1.5 text-sm font-semibold rounded-full border',
              statusColors[bill.status]
            ]">
              {{ statusLabels[bill.status] }}
            </span>

            <span class="text-gray-500">
              Filed {{ formatDate(bill.filed_date) }}
            </span>

            <span v-if="bill.last_action_date" class="text-gray-500">
              Last action {{ formatDate(bill.last_action_date) }}
            </span>
          </div>

          <!-- Republic Act Number -->
          <div v-if="bill.republic_act_number" class="mt-4 inline-block bg-green-50 border border-green-200 rounded-lg px-4 py-2">
            <span class="text-sm text-green-700 font-medium">
              Republic Act No. {{ bill.republic_act_number }}
            </span>
            <span v-if="bill.date_signed" class="text-sm text-green-600 ml-2">
              (Signed {{ formatDate(bill.date_signed) }})
            </span>
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="bg-white border-b sticky top-0 z-10">
        <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
          <nav class="flex gap-8">
            <button
              v-for="tab in ['overview', 'timeline', 'votes', 'committees'] as const"
              :key="tab"
              :class="[
                'py-4 px-1 text-sm font-medium border-b-2 -mb-px capitalize',
                activeTab === tab
                  ? 'border-red-600 text-red-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              ]"
              @click="activeTab = tab"
            >
              {{ tab }}
            </button>
          </nav>
        </div>
      </div>

      <!-- Tab Content -->
      <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <!-- Overview Tab -->
        <div v-if="activeTab === 'overview'" class="space-y-8">
          <!-- Summary -->
          <section v-if="bill.summary" class="bg-white rounded-lg shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Summary</h2>
            <p class="text-gray-700 whitespace-pre-line">{{ bill.summary }}</p>
          </section>

          <!-- Authors -->
          <section v-if="bill.authors && bill.authors.length > 0" class="bg-white rounded-lg shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Authors</h2>
            <div class="space-y-3">
              <!-- Principal Authors -->
              <div v-if="bill.principal_authors && bill.principal_authors.length > 0">
                <h3 class="text-sm font-medium text-gray-500 mb-2">Principal Authors</h3>
                <div class="flex flex-wrap gap-2">
                  <NuxtLink
                    v-for="author in bill.principal_authors"
                    :key="author.id"
                    :to="`/politician/${author.slug}`"
                    class="inline-flex items-center gap-2 px-3 py-1.5 bg-red-50 text-red-700 rounded-full text-sm hover:bg-red-100"
                  >
                    <img
                      v-if="author.photo"
                      :src="author.photo"
                      :alt="author.name"
                      class="w-6 h-6 rounded-full object-cover"
                    >
                    <span>{{ author.name }}</span>
                  </NuxtLink>
                </div>
              </div>

              <!-- Co-Authors -->
              <div v-if="bill.authors.filter(a => !a.is_principal_author).length > 0">
                <h3 class="text-sm font-medium text-gray-500 mb-2">Co-Authors</h3>
                <div class="flex flex-wrap gap-2">
                  <NuxtLink
                    v-for="author in bill.authors.filter(a => !a.is_principal_author)"
                    :key="author.id"
                    :to="`/politician/${author.politician?.slug}`"
                    class="inline-flex items-center gap-2 px-3 py-1.5 bg-gray-100 text-gray-700 rounded-full text-sm hover:bg-gray-200"
                  >
                    <span>{{ author.politician?.name }}</span>
                  </NuxtLink>
                </div>
              </div>
            </div>
          </section>

          <!-- Topics -->
          <section v-if="bill.topics && bill.topics.length > 0" class="bg-white rounded-lg shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Topics</h2>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="topic in bill.topics"
                :key="topic.id"
                class="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm"
              >
                {{ topic.name }}
              </span>
            </div>
          </section>

          <!-- Full Text -->
          <section v-if="bill.full_text" class="bg-white rounded-lg shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Full Text</h2>
            <div class="prose prose-gray max-w-none">
              <pre class="whitespace-pre-wrap text-sm text-gray-700 font-sans">{{ bill.full_text }}</pre>
            </div>
          </section>
        </div>

        <!-- Timeline Tab -->
        <div v-else-if="activeTab === 'timeline'" class="bg-white rounded-lg shadow p-6">
          <h2 class="text-lg font-semibold text-gray-900 mb-6">Status History</h2>

          <div v-if="bill.status_history && bill.status_history.length > 0" class="relative">
            <!-- Timeline line -->
            <div class="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200" />

            <div class="space-y-6">
              <div
                v-for="(item, index) in bill.status_history"
                :key="item.id"
                class="relative pl-10"
              >
                <!-- Dot -->
                <div
:class="[
                  'absolute left-2 w-5 h-5 rounded-full border-2 bg-white',
                  index === 0 ? 'border-red-600' : 'border-gray-300'
                ]" />

                <!-- Content -->
                <div>
                  <div class="flex items-center gap-3">
                    <span
:class="[
                      'px-2 py-0.5 text-xs font-medium rounded',
                      statusColors[item.status]
                    ]">
                      {{ statusLabels[item.status] }}
                    </span>
                    <span class="text-sm text-gray-500">{{ formatShortDate(item.action_date) }}</span>
                  </div>
                  <p v-if="item.action_description" class="mt-1 text-gray-600">
                    {{ item.action_description }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <p v-else class="text-gray-500 text-center py-8">No status history available</p>
        </div>

        <!-- Votes Tab -->
        <div v-else-if="activeTab === 'votes'" class="space-y-6">
          <div v-if="bill.votes && bill.votes.length > 0">
            <div
              v-for="vote in bill.votes"
              :key="vote.id"
              class="bg-white rounded-lg shadow p-6"
            >
              <div class="flex items-center justify-between mb-4">
                <div>
                  <h3 class="font-semibold text-gray-900">
                    {{ vote.reading === 'second' ? '2nd' : '3rd' }} Reading Vote
                  </h3>
                  <p class="text-sm text-gray-500">
                    {{ vote.chamber === 'senate' ? 'Senate' : 'House' }} • {{ formatDate(vote.vote_date) }}
                  </p>
                </div>
                <span
:class="[
                  'px-3 py-1 text-sm font-semibold rounded-full',
                  vote.is_passed ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                ]">
                  {{ vote.is_passed ? 'Passed' : 'Failed' }}
                </span>
              </div>

              <!-- Vote Counts -->
              <div class="grid grid-cols-4 gap-4 text-center">
                <div class="p-3 bg-green-50 rounded-lg">
                  <div class="text-2xl font-bold text-green-600">{{ vote.yeas }}</div>
                  <div class="text-sm text-green-700">Yes</div>
                </div>
                <div class="p-3 bg-red-50 rounded-lg">
                  <div class="text-2xl font-bold text-red-600">{{ vote.nays }}</div>
                  <div class="text-sm text-red-700">No</div>
                </div>
                <div class="p-3 bg-yellow-50 rounded-lg">
                  <div class="text-2xl font-bold text-yellow-600">{{ vote.abstentions }}</div>
                  <div class="text-sm text-yellow-700">Abstain</div>
                </div>
                <div class="p-3 bg-gray-50 rounded-lg">
                  <div class="text-2xl font-bold text-gray-600">{{ vote.absent }}</div>
                  <div class="text-sm text-gray-700">Absent</div>
                </div>
              </div>

              <p v-if="vote.notes" class="mt-4 text-sm text-gray-600">
                {{ vote.notes }}
              </p>
            </div>
          </div>

          <div v-else class="bg-white rounded-lg shadow p-12 text-center">
            <svg class="w-12 h-12 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
            <p class="mt-4 text-gray-500">No voting records available yet</p>
          </div>
        </div>

        <!-- Committees Tab -->
        <div v-else-if="activeTab === 'committees'" class="bg-white rounded-lg shadow p-6">
          <h2 class="text-lg font-semibold text-gray-900 mb-6">Committee Referrals</h2>

          <div v-if="bill.committees && bill.committees.length > 0" class="space-y-4">
            <div
              v-for="ref in bill.committees"
              :key="ref.id"
              class="flex items-center justify-between p-4 border rounded-lg"
            >
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="font-medium text-gray-900">{{ ref.committee?.name }}</h3>
                  <span v-if="ref.is_primary" class="px-2 py-0.5 text-xs bg-red-100 text-red-700 rounded">
                    Primary
                  </span>
                </div>
                <p class="text-sm text-gray-500">
                  {{ ref.committee?.chamber === 'senate' ? 'Senate' : 'House' }} •
                  Referred {{ formatShortDate(ref.referred_date) }}
                </p>
              </div>
              <span
:class="[
                'px-3 py-1 text-sm font-medium rounded capitalize',
                ref.status === 'approved' ? 'bg-green-100 text-green-700' :
                ref.status === 'disapproved' ? 'bg-red-100 text-red-700' :
                'bg-yellow-100 text-yellow-700'
              ]">
                {{ ref.status }}
              </span>
            </div>
          </div>

          <p v-else class="text-gray-500 text-center py-8">No committee referrals</p>
        </div>
      </div>
    </template>
  </div>
</template>
