<script setup lang="ts">
import type { BillFilter, BillStatus, LegislativeChamber } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const api = useApi()

// Filter state
const chamber = ref<LegislativeChamber | ''>('')
const status = ref<BillStatus | ''>('')
const search = ref('')
const currentPage = ref(1)
const perPage = 20

// Grouping configuration
type LegislationGroupBy = '' | 'chamber' | 'status'
const groupBy = ref<LegislationGroupBy>('')

// Data
const { data: sessionsData } = useAsyncData('admin-sessions', () => api.getLegislativeSessions())
const { data: topicsData } = useAsyncData('admin-topics', () => api.getBillTopics())

const sessions = computed(() => sessionsData.value || [])
const topics = computed(() => topicsData.value || [])

// Build filter
const filter = computed<BillFilter>(() => ({
  chamber: chamber.value || undefined,
  status: status.value || undefined,
  search: search.value || undefined
}))

// Fetch bills
const { data: billsData, pending } = useAsyncData(
  'admin-bills',
  () => api.getBills(filter.value, currentPage.value, perPage),
  { watch: [filter, currentPage] }
)

const bills = computed(() => billsData.value?.bills || [])
const total = computed(() => billsData.value?.total || 0)
const totalPages = computed(() => billsData.value?.total_pages || 1)

// Status display mapping (needed before grouping)
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

type BadgeColor = 'error' | 'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'neutral'
const statusColors: Record<BillStatus, BadgeColor> = {
  filed: 'neutral',
  pending_committee: 'warning',
  in_committee: 'info',
  reported_out: 'info',
  pending_second_reading: 'info',
  approved_second_reading: 'info',
  pending_third_reading: 'info',
  approved_third_reading: 'info',
  transmitted: 'info',
  consolidated: 'info',
  ratified: 'success',
  signed_into_law: 'success',
  vetoed: 'error',
  lapsed: 'warning',
  withdrawn: 'neutral',
  archived: 'neutral'
}

// Group by options
const groupByOptions = [
  { label: 'No Grouping', value: '' },
  { label: 'By Chamber', value: 'chamber' },
  { label: 'By Status', value: 'status' }
]

// Use grouping composable
const {
  expandedGroups,
  groupedItems: groupedBills,
  toggleGroup,
  expandAll,
  collapseAll,
  hasExpandedGroups,
  allGroupsExpanded
} = useGrouping(
  bills,
  groupBy,
  (bill, groupByValue) => {
    if (groupByValue === 'chamber') {
      return bill.chamber ? bill.chamber.charAt(0).toUpperCase() + bill.chamber.slice(1) : 'Unknown Chamber'
    } else if (groupByValue === 'status') {
      return statusLabels[bill.status] || 'Unknown Status'
    }
    return ''
  },
  {
    storageKey: 'admin-legislation-expanded-groups',
    defaultExpanded: false
  }
)

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-PH', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

useSeoMeta({
  title: 'Manage Legislation - Admin'
})
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Legislation Management</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Manage bills, sessions, and committees</p>
      </div>
      <UButton to="/admin/legislation/new" color="primary" icon="i-heroicons-plus">
        Add Bill
      </UButton>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-900 rounded-lg shadow p-4 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-5 gap-4">
        <UInput
          v-model="search"
          placeholder="Search bills..."
          icon="i-heroicons-magnifying-glass"
        />
        <USelect
          v-model="chamber"
          :items="[
            { label: 'All Chambers', value: '' },
            { label: 'Senate', value: 'senate' },
            { label: 'House', value: 'house' }
          ]"
          value-key="value"
          placeholder="Chamber"
        />
        <USelect
          v-model="status"
          :items="[
            { label: 'All Statuses', value: '' },
            ...Object.entries(statusLabels).map(([key, label]) => ({ label, value: key }))
          ]"
          value-key="value"
          placeholder="Status"
        />
        <USelect
          v-model="groupBy"
          :options="groupByOptions"
          placeholder="Group by"
        />
        <div class="flex items-center gap-3 self-center">
          <span class="text-sm text-gray-500">
            {{ total }} bill{{ total !== 1 ? 's' : '' }}
          </span>
          <div v-if="groupBy" class="flex gap-1">
            <UButton
              variant="ghost"
              size="xs"
              icon="i-heroicons-chevron-double-down"
              :disabled="allGroupsExpanded"
              title="Expand all groups"
              @click="expandAll"
            >
              Expand All
            </UButton>
            <UButton
              variant="ghost"
              size="xs"
              icon="i-heroicons-chevron-double-up"
              :disabled="!hasExpandedGroups"
              title="Collapse all groups"
              @click="collapseAll"
            >
              Collapse All
            </UButton>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Stats -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-white dark:bg-gray-900 rounded-lg shadow p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">Sessions</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ sessions.length }}</div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-lg shadow p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">Topics</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ topics.length }}</div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-lg shadow p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">Total Bills</div>
        <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ total }}</div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-lg shadow p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">Current Session</div>
        <div class="text-lg font-bold text-gray-900 dark:text-white">
          {{ sessions.find(s => s.is_current)?.congress_number || '-' }}th Congress
        </div>
      </div>
    </div>

    <!-- Bills Table -->
    <div class="bg-white dark:bg-gray-900 rounded-lg shadow overflow-hidden">
      <div v-if="pending" class="p-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 animate-spin text-gray-400" />
      </div>

      <div v-else-if="bills.length === 0" class="p-8 text-center">
        <UIcon name="i-heroicons-document-text" class="w-12 h-12 mx-auto text-gray-400 mb-4" />
        <p class="text-gray-500">No bills found</p>
      </div>

      <!-- Grouped View -->
      <div v-else-if="bills.length && groupBy" class="p-6 space-y-6">
        <div v-for="(groupBills, groupName) in groupedBills" :key="groupName" class="space-y-2">
          <!-- Group Header -->
          <button
            :aria-expanded="expandedGroups.has(groupName)"
            :aria-controls="`legislation-group-${groupName}`"
            class="flex items-center gap-2 w-full px-4 py-3 bg-gray-50 dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-750 rounded-lg transition-colors"
            @click="toggleGroup(groupName)"
          >
            <UIcon
              :name="expandedGroups.has(groupName) ? 'i-heroicons-chevron-down' : 'i-heroicons-chevron-right'"
              class="size-5 text-gray-600 dark:text-gray-400"
            />
            <span class="font-semibold text-gray-900 dark:text-white">{{ groupName }}</span>
            <UBadge color="primary" variant="subtle" size="sm">{{ groupBills.length }}</UBadge>
          </button>

          <!-- Group Content -->
          <div
            v-if="expandedGroups.has(groupName)"
            :id="`legislation-group-${groupName}`"
            role="region"
            :aria-label="`${groupName} bills`"
            class="overflow-x-auto"
          >
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
              <thead class="bg-gray-50 dark:bg-gray-800">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Bill
                  </th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Chamber
                  </th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Filed
                  </th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800">
                <tr v-for="bill in groupBills" :key="bill.id" class="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td class="px-6 py-4">
                    <div class="text-sm font-mono font-medium text-gray-900 dark:text-white">
                      {{ bill.bill_number }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400 truncate max-w-md">
                      {{ bill.short_title || bill.title }}
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span class="capitalize text-sm">{{ bill.chamber }}</span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <UBadge :color="statusColors[bill.status]" variant="subtle" size="xs">
                      {{ statusLabels[bill.status] }}
                    </UBadge>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                    {{ formatDate(bill.filed_date) }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right">
                    <div class="flex items-center justify-end gap-2">
                      <UButton
                        :to="`/legislation/${bill.slug}`"
                        variant="ghost"
                        size="xs"
                        icon="i-heroicons-eye"
                        target="_blank"
                      />
                      <UButton
                        :to="`/admin/legislation/${bill.id}`"
                        variant="ghost"
                        size="xs"
                        icon="i-heroicons-pencil"
                      />
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Ungrouped View -->
      <table v-else class="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
        <thead class="bg-gray-50 dark:bg-gray-800">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Bill
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Chamber
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Status
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Filed
            </th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800">
          <tr v-for="bill in bills" :key="bill.id" class="hover:bg-gray-50 dark:hover:bg-gray-800">
            <td class="px-6 py-4">
              <div class="text-sm font-mono font-medium text-gray-900 dark:text-white">
                {{ bill.bill_number }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400 truncate max-w-md">
                {{ bill.short_title || bill.title }}
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span class="capitalize text-sm">{{ bill.chamber }}</span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <UBadge :color="statusColors[bill.status]" variant="subtle" size="xs">
                {{ statusLabels[bill.status] }}
              </UBadge>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
              {{ formatDate(bill.filed_date) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right">
              <div class="flex items-center justify-end gap-2">
                <UButton
                  :to="`/legislation/${bill.slug}`"
                  variant="ghost"
                  size="xs"
                  icon="i-heroicons-eye"
                  target="_blank"
                />
                <UButton
                  :to="`/admin/legislation/${bill.id}`"
                  variant="ghost"
                  size="xs"
                  icon="i-heroicons-pencil"
                />
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="px-6 py-4 border-t border-gray-200 dark:border-gray-800 flex items-center justify-between">
        <div class="text-sm text-gray-500">
          Showing {{ (currentPage - 1) * perPage + 1 }} to {{ Math.min(currentPage * perPage, total) }} of {{ total }}
        </div>
        <div class="flex gap-2">
          <UButton
            :disabled="currentPage === 1"
            variant="outline"
            size="sm"
            @click="currentPage--"
          >
            Previous
          </UButton>
          <UButton
            :disabled="currentPage === totalPages"
            variant="outline"
            size="sm"
            @click="currentPage++"
          >
            Next
          </UButton>
        </div>
      </div>
    </div>
  </div>
</template>
