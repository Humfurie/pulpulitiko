<script setup lang="ts">
import type { PoliticianListItem, PaginatedPoliticians, PoliticalPartyListItem, ApiResponse } from '~/types'
import type { TableColumn } from '@nuxt/ui'
import { useDebounceFn } from '@vueuse/core'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

const page = ref(1)
const loading = ref(false)
const politicians = ref<PoliticianListItem[]>([])
const total = ref(0)
const totalPages = ref(1)
const error = ref('')
const search = ref('')
const partyFilter = ref('')
const parties = ref<PoliticalPartyListItem[]>([])

// Grouping configuration
type PoliticianGroupBy = '' | 'party' | 'position'
const groupBy = ref<PoliticianGroupBy>('')

const columns: TableColumn<PoliticianListItem>[] = [
  { accessorKey: 'photo', header: '' },
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'position', header: 'Position' },
  { accessorKey: 'party', header: 'Party' },
  { accessorKey: 'article_count', header: 'Articles' },
  { id: 'actions', header: 'Actions' }
]

// Party filter options
const partyOptions = computed(() => [
  { label: 'All Parties', value: '' },
  ...parties.value.map(p => ({
    label: p.abbreviation ? `${p.name} (${p.abbreviation})` : p.name,
    value: p.id
  }))
])

// Group by options
const groupByOptions = [
  { label: 'No Grouping', value: '' },
  { label: 'By Party', value: 'party' },
  { label: 'By Position', value: 'position' }
]

// Use grouping composable
const {
  expandedGroups,
  groupedItems: groupedPoliticians,
  toggleGroup,
  expandAll,
  collapseAll,
  hasExpandedGroups,
  allGroupsExpanded
} = useGrouping(
  politicians,
  groupBy,
  (politician, groupByValue) => {
    if (groupByValue === 'party') {
      return politician.party || 'No Party'
    } else if (groupByValue === 'position') {
      return politician.position || 'No Position'
    }
    return ''
  },
  {
    storageKey: 'admin-politicians-expanded-groups',
    defaultExpanded: false
  }
)

async function fetchParties() {
  try {
    const response = await $fetch<ApiResponse<PoliticalPartyListItem[]>>(`${baseUrl}/parties`, {
      headers: auth.getAuthHeaders()
    })
    if (response.success) {
      parties.value = response.data
    }
  } catch (e: unknown) {
    console.error('Failed to load parties:', e)
  }
}

async function fetchPoliticians() {
  loading.value = true
  error.value = ''

  try {
    const params = new URLSearchParams({
      page: String(page.value),
      per_page: '20'
    })
    if (search.value) {
      params.append('search', search.value)
    }
    if (partyFilter.value) {
      params.append('party_id', partyFilter.value)
    }

    const response = await $fetch<ApiResponse<PaginatedPoliticians>>(`${baseUrl}/admin/politicians?${params}`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      politicians.value = response.data.politicians
      total.value = response.data.total
      totalPages.value = response.data.total_pages
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load politicians'
  }

  loading.value = false
}

async function deletePolitician(id: string) {
  if (!confirm('Are you sure you want to delete this politician?')) return

  try {
    await $fetch(`${baseUrl}/admin/politicians/${id}`, {
      method: 'DELETE',
      headers: auth.getAuthHeaders()
    })
    await fetchPoliticians()
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    alert(err?.data?.error?.message || 'Failed to delete politician')
  }
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchPoliticians()
}, 300)

onMounted(() => {
  fetchParties()
  fetchPoliticians()
})
watch(page, fetchPoliticians)
watch(search, debouncedSearch)
watch(partyFilter, () => {
  page.value = 1
  fetchPoliticians()
})

useSeoMeta({
  title: 'Politicians - Pulpulitiko Admin'
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Politicians</h1>
      <UButton to="/admin/politicians/new" icon="i-heroicons-plus">
        New Politician
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4">
          <div class="flex flex-col sm:flex-row gap-4">
            <UInput
              v-model="search"
              placeholder="Search politicians..."
              icon="i-heroicons-magnifying-glass"
              class="flex-1"
            />
            <USelect
              v-model="partyFilter"
              :options="partyOptions"
              placeholder="Filter by party"
              class="w-48"
            />
            <USelect
              v-model="groupBy"
              :options="groupByOptions"
              placeholder="Group by"
              class="w-48"
            />
          </div>
          <div v-if="total > 0" class="flex items-center gap-3">
            <span class="text-sm text-gray-500">
              {{ total }} politician{{ total !== 1 ? 's' : '' }} total
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
      </template>

      <div v-if="loading" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      </div>

      <!-- Grouped View -->
      <div v-else-if="politicians.length && groupBy" class="space-y-6">
        <div v-for="(groupPoliticians, groupName) in groupedPoliticians" :key="groupName" class="space-y-2">
          <!-- Group Header -->
          <button
            :aria-expanded="expandedGroups.has(groupName)"
            :aria-controls="`politicians-group-${groupName}`"
            class="flex items-center gap-2 w-full px-4 py-3 bg-gray-50 dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-750 rounded-lg transition-colors"
            @click="toggleGroup(groupName)"
          >
            <UIcon
              :name="expandedGroups.has(groupName) ? 'i-heroicons-chevron-down' : 'i-heroicons-chevron-right'"
              class="size-5 text-gray-600 dark:text-gray-400"
            />
            <span class="font-semibold text-gray-900 dark:text-white">{{ groupName }}</span>
            <UBadge color="primary" variant="subtle" size="sm">{{ groupPoliticians.length }}</UBadge>
          </button>

          <!-- Group Content -->
          <div
            v-if="expandedGroups.has(groupName)"
            :id="`politicians-group-${groupName}`"
            role="region"
            :aria-label="`${groupName} politicians`"
          >
            <UTable
              :data="groupPoliticians"
              :columns="columns"
            >
            <template #photo-cell="{ row }">
              <UAvatar
                :src="row.original.photo"
                :alt="row.original.name"
                size="sm"
              />
            </template>

            <template #name-cell="{ row }">
              <NuxtLink :to="`/admin/politicians/${row.original.id}`" class="block">
                <p class="font-medium text-gray-900 dark:text-white hover:text-primary">{{ row.original.name }}</p>
                <p class="text-sm text-gray-500">{{ row.original.slug }}</p>
              </NuxtLink>
            </template>

            <template #position-cell="{ row }">
              <span class="text-gray-600 dark:text-gray-400 text-sm">
                {{ row.original.position || '-' }}
              </span>
            </template>

            <template #party-cell="{ row }">
              <NuxtLink
                v-if="row.original.party_info"
                :to="`/admin/parties/${row.original.party_info.id}`"
                class="inline-block"
              >
                <UBadge variant="subtle" class="hover:bg-primary/20 transition-colors cursor-pointer">
                  {{ row.original.party }}
                </UBadge>
              </NuxtLink>
              <UBadge v-else-if="row.original.party" variant="subtle">
                {{ row.original.party }}
              </UBadge>
              <span v-else class="text-gray-400">-</span>
            </template>

            <template #article_count-cell="{ row }">
              <span class="text-gray-600 dark:text-gray-400">
                {{ row.original.article_count }}
              </span>
            </template>

            <template #actions-cell="{ row }">
              <div class="flex items-center gap-2 justify-end">
                <UButton
                  :to="`/admin/politicians/${row.original.id}`"
                  variant="ghost"
                  size="sm"
                  icon="i-heroicons-pencil"
                />
                <UButton
                  variant="ghost"
                  size="sm"
                  color="error"
                  icon="i-heroicons-trash"
                  @click="deletePolitician(row.original.id)"
                />
              </div>
            </template>
            </UTable>
          </div>
        </div>
      </div>

      <!-- Ungrouped View -->
      <UTable
        v-else-if="politicians.length"
        :data="politicians"
        :columns="columns"
      >
        <template #photo-cell="{ row }">
          <UAvatar
            :src="row.original.photo"
            :alt="row.original.name"
            size="sm"
          />
        </template>

        <template #name-cell="{ row }">
          <NuxtLink :to="`/admin/politicians/${row.original.id}`" class="block">
            <p class="font-medium text-gray-900 dark:text-white hover:text-primary">{{ row.original.name }}</p>
            <p class="text-sm text-gray-500">{{ row.original.slug }}</p>
          </NuxtLink>
        </template>

        <template #position-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400 text-sm">
            {{ row.original.position || '-' }}
          </span>
        </template>

        <template #party-cell="{ row }">
          <NuxtLink
            v-if="row.original.party_info"
            :to="`/admin/parties/${row.original.party_info.id}`"
            class="inline-block"
          >
            <UBadge variant="subtle" class="hover:bg-primary/20 transition-colors cursor-pointer">
              {{ row.original.party }}
            </UBadge>
          </NuxtLink>
          <UBadge v-else-if="row.original.party" variant="subtle">
            {{ row.original.party }}
          </UBadge>
          <span v-else class="text-gray-400">-</span>
        </template>

        <template #article_count-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400">
            {{ row.original.article_count }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-2 justify-end">
            <UButton
              :to="`/admin/politicians/${row.original.id}`"
              variant="ghost"
              size="sm"
              icon="i-heroicons-pencil"
            >
              Edit
            </UButton>
            <UButton
              variant="ghost"
              size="sm"
              color="error"
              icon="i-heroicons-trash"
              @click="deletePolitician(row.original.id)"
            >
              Delete
            </UButton>
          </div>
        </template>
      </UTable>

      <div v-else class="py-8 text-center text-gray-500">
        No politicians yet.
        <NuxtLink to="/admin/politicians/new" class="text-primary hover:underline">Create one</NuxtLink>
      </div>

      <template v-if="totalPages > 1" #footer>
        <div class="flex justify-center">
          <UPagination v-model:page="page" :total="total" :items-per-page="20" />
        </div>
      </template>
    </UCard>
  </div>
</template>
