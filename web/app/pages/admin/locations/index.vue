<script setup lang="ts">
import type { RegionListItem, ProvinceListItem, CityMunicipalityListItem, BarangayListItem, ApiResponse } from '~/types'
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

// Tab state
const activeTab = ref<'regions' | 'provinces' | 'cities' | 'barangays'>('regions')

// Selected parent for hierarchical filtering
const selectedRegionId = ref<string>('')
const selectedProvinceId = ref<string>('')
const selectedCityId = ref<string>('')

// Data and loading states
const regions = ref<RegionListItem[]>([])
const provinces = ref<ProvinceListItem[]>([])
const cities = ref<CityMunicipalityListItem[]>([])
const barangays = ref<BarangayListItem[]>([])
const loading = ref(false)
const error = ref('')

// Pagination
const page = ref(1)
const total = ref(0)
const totalPages = ref(1)
const perPage = 20

// Column definitions
const regionColumns: TableColumn<RegionListItem>[] = [
  { accessorKey: 'code', header: 'PSGC Code' },
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'province_count', header: 'Provinces' },
  { id: 'actions', header: '' }
]

const provinceColumns: TableColumn<ProvinceListItem>[] = [
  { accessorKey: 'code', header: 'PSGC Code' },
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'city_count', header: 'Cities' },
  { id: 'actions', header: '' }
]

const cityColumns: TableColumn<CityMunicipalityListItem>[] = [
  { accessorKey: 'code', header: 'PSGC Code' },
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'type', header: 'Type' },
  { accessorKey: 'barangay_count', header: 'Barangays' },
  { id: 'actions', header: '' }
]

const barangayColumns: TableColumn<BarangayListItem>[] = [
  { accessorKey: 'code', header: 'PSGC Code' },
  { accessorKey: 'name', header: 'Name' },
  { id: 'actions', header: '' }
]

// Fetch functions
async function fetchRegions() {
  loading.value = true
  error.value = ''
  try {
    const response = await $fetch<ApiResponse<RegionListItem[]>>(`${baseUrl}/locations/regions`, {
      headers: auth.getAuthHeaders()
    })
    if (response.success) {
      regions.value = response.data
      total.value = response.data.length
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load regions'
  }
  loading.value = false
}

async function fetchProvinces() {
  if (!selectedRegionId.value) {
    provinces.value = []
    return
  }
  loading.value = true
  error.value = ''
  try {
    const response = await $fetch<ApiResponse<ProvinceListItem[]>>(`${baseUrl}/locations/provinces/by-region/${selectedRegionId.value}`, {
      headers: auth.getAuthHeaders()
    })
    if (response.success) {
      provinces.value = response.data
      total.value = response.data.length
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load provinces'
  }
  loading.value = false
}

async function fetchCities() {
  if (!selectedProvinceId.value) {
    cities.value = []
    return
  }
  loading.value = true
  error.value = ''
  try {
    const response = await $fetch<ApiResponse<CityMunicipalityListItem[]>>(`${baseUrl}/locations/cities/by-province/${selectedProvinceId.value}`, {
      headers: auth.getAuthHeaders()
    })
    if (response.success) {
      cities.value = response.data
      total.value = response.data.length
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load cities'
  }
  loading.value = false
}

async function fetchBarangays() {
  if (!selectedCityId.value) {
    barangays.value = []
    return
  }
  loading.value = true
  error.value = ''
  try {
    const response = await $fetch<ApiResponse<{ barangays: BarangayListItem[], total: number }>>(`${baseUrl}/locations/barangays/by-city/${selectedCityId.value}?page=${page.value}&per_page=${perPage}`, {
      headers: auth.getAuthHeaders()
    })
    if (response.success) {
      barangays.value = response.data.barangays
      total.value = response.data.total
      totalPages.value = Math.ceil(response.data.total / perPage)
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to load barangays'
  }
  loading.value = false
}

// Get city type label
function getCityType(city: CityMunicipalityListItem): string {
  if (city.is_huc) return 'HUC'
  if (city.is_city && city.is_capital) return 'City (Capital)'
  if (city.is_city) return 'City'
  if (city.is_capital) return 'Municipality (Capital)'
  return 'Municipality'
}

function getCityTypeColor(city: CityMunicipalityListItem): 'error' | 'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'neutral' {
  if (city.is_huc) return 'info'
  if (city.is_capital) return 'warning'
  if (city.is_city) return 'primary'
  return 'neutral'
}

// Tab change handler
function onTabChange(tab: 'regions' | 'provinces' | 'cities' | 'barangays') {
  activeTab.value = tab
  page.value = 1
  error.value = ''

  if (tab === 'regions') {
    fetchRegions()
  } else if (tab === 'provinces') {
    if (selectedRegionId.value) fetchProvinces()
  } else if (tab === 'cities') {
    if (selectedProvinceId.value) fetchCities()
  } else if (tab === 'barangays') {
    if (selectedCityId.value) fetchBarangays()
  }
}

// Parent selection handlers
watch(selectedRegionId, (newVal) => {
  selectedProvinceId.value = ''
  selectedCityId.value = ''
  provinces.value = []
  cities.value = []
  barangays.value = []
  if (newVal && activeTab.value === 'provinces') {
    fetchProvinces()
  }
})

watch(selectedProvinceId, (newVal) => {
  selectedCityId.value = ''
  cities.value = []
  barangays.value = []
  if (newVal && activeTab.value === 'cities') {
    fetchCities()
  }
})

watch(selectedCityId, (newVal) => {
  barangays.value = []
  page.value = 1
  if (newVal && activeTab.value === 'barangays') {
    fetchBarangays()
  }
})

watch(page, () => {
  if (activeTab.value === 'barangays') {
    fetchBarangays()
  }
})

// Initial load
onMounted(() => {
  fetchRegions()
})

useSeoMeta({
  title: 'Locations - Pulpulitiko Admin'
})

const tabs = [
  { label: 'Regions', value: 'regions' as const, icon: 'i-heroicons-globe-asia-australia' },
  { label: 'Provinces', value: 'provinces' as const, icon: 'i-heroicons-map' },
  { label: 'Cities', value: 'cities' as const, icon: 'i-heroicons-building-office-2' },
  { label: 'Barangays', value: 'barangays' as const, icon: 'i-heroicons-home' }
]
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Locations</h1>
      <div class="flex gap-2">
        <UButton
          to="/locations"
          variant="outline"
          icon="i-heroicons-eye"
        >
          View Public Page
        </UButton>
      </div>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />

    <!-- Tabs -->
    <div class="flex gap-2 mb-6 border-b border-gray-200 dark:border-gray-700">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        :class="[
          'flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
          activeTab === tab.value
            ? 'border-primary text-primary'
            : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
        ]"
        @click="onTabChange(tab.value)"
      >
        <UIcon :name="tab.icon" class="w-4 h-4" />
        {{ tab.label }}
      </button>
    </div>

    <!-- Regions Tab -->
    <UCard v-if="activeTab === 'regions'">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-sm text-gray-500">{{ regions.length }} regions</span>
        </div>
      </template>

      <div v-if="loading" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      </div>

      <UTable
        v-else-if="regions.length"
        :data="regions"
        :columns="regionColumns"
      >
        <template #code-cell="{ row }">
          <code class="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
            {{ row.original.code }}
          </code>
        </template>

        <template #name-cell="{ row }">
          <div>
            <p class="font-medium text-gray-900 dark:text-white">{{ row.original.name }}</p>
            <p class="text-sm text-gray-500">{{ row.original.slug }}</p>
          </div>
        </template>

        <template #province_count-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400">
            {{ row.original.province_count }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-2 justify-end">
            <UButton
              variant="ghost"
              size="sm"
              icon="i-heroicons-arrow-right"
              @click="selectedRegionId = row.original.id; onTabChange('provinces')"
            >
              View Provinces
            </UButton>
          </div>
        </template>
      </UTable>

      <div v-else class="py-8 text-center text-gray-500">
        No regions found.
      </div>
    </UCard>

    <!-- Provinces Tab -->
    <UCard v-if="activeTab === 'provinces'">
      <template #header>
        <div class="flex items-center gap-4">
          <USelectMenu
            v-model="selectedRegionId"
            :items="regions.map(r => ({ label: r.name, value: r.id }))"
            placeholder="Select a region"
            class="w-64"
            value-key="value"
          />
          <span v-if="selectedRegionId" class="text-sm text-gray-500">
            {{ provinces.length }} provinces
          </span>
        </div>
      </template>

      <div v-if="!selectedRegionId" class="py-8 text-center text-gray-500">
        <UIcon name="i-heroicons-arrow-up" class="w-8 h-8 mx-auto mb-2" />
        <p>Select a region to view its provinces</p>
      </div>

      <div v-else-if="loading" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      </div>

      <UTable
        v-else-if="provinces.length"
        :data="provinces"
        :columns="provinceColumns"
      >
        <template #code-cell="{ row }">
          <code class="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
            {{ row.original.code }}
          </code>
        </template>

        <template #name-cell="{ row }">
          <div>
            <p class="font-medium text-gray-900 dark:text-white">{{ row.original.name }}</p>
            <p class="text-sm text-gray-500">{{ row.original.slug }}</p>
          </div>
        </template>

        <template #city_count-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400">
            {{ row.original.city_count }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-2 justify-end">
            <UButton
              variant="ghost"
              size="sm"
              icon="i-heroicons-arrow-right"
              @click="selectedProvinceId = row.original.id; onTabChange('cities')"
            >
              View Cities
            </UButton>
          </div>
        </template>
      </UTable>

      <div v-else class="py-8 text-center text-gray-500">
        No provinces found in this region.
      </div>
    </UCard>

    <!-- Cities Tab -->
    <UCard v-if="activeTab === 'cities'">
      <template #header>
        <div class="flex items-center gap-4 flex-wrap">
          <USelectMenu
            v-model="selectedRegionId"
            :items="regions.map(r => ({ label: r.name, value: r.id }))"
            placeholder="Select a region"
            class="w-48"
            value-key="value"
          />
          <USelectMenu
            v-model="selectedProvinceId"
            :items="provinces.map(p => ({ label: p.name, value: p.id }))"
            placeholder="Select a province"
            class="w-48"
            :disabled="!selectedRegionId"
            value-key="value"
          />
          <span v-if="selectedProvinceId" class="text-sm text-gray-500">
            {{ cities.length }} cities/municipalities
          </span>
        </div>
      </template>

      <div v-if="!selectedProvinceId" class="py-8 text-center text-gray-500">
        <UIcon name="i-heroicons-arrow-up" class="w-8 h-8 mx-auto mb-2" />
        <p>Select a region and province to view cities/municipalities</p>
      </div>

      <div v-else-if="loading" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      </div>

      <UTable
        v-else-if="cities.length"
        :data="cities"
        :columns="cityColumns"
      >
        <template #code-cell="{ row }">
          <code class="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
            {{ row.original.code }}
          </code>
        </template>

        <template #name-cell="{ row }">
          <div>
            <p class="font-medium text-gray-900 dark:text-white">{{ row.original.name }}</p>
            <p class="text-sm text-gray-500">{{ row.original.slug }}</p>
          </div>
        </template>

        <template #type-cell="{ row }">
          <UBadge :color="getCityTypeColor(row.original)" variant="subtle" size="sm">
            {{ getCityType(row.original) }}
          </UBadge>
        </template>

        <template #barangay_count-cell="{ row }">
          <span class="text-gray-600 dark:text-gray-400">
            {{ row.original.barangay_count }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-2 justify-end">
            <UButton
              variant="ghost"
              size="sm"
              icon="i-heroicons-arrow-right"
              @click="selectedCityId = row.original.id; onTabChange('barangays')"
            >
              View Barangays
            </UButton>
          </div>
        </template>
      </UTable>

      <div v-else class="py-8 text-center text-gray-500">
        No cities/municipalities found in this province.
      </div>
    </UCard>

    <!-- Barangays Tab -->
    <UCard v-if="activeTab === 'barangays'">
      <template #header>
        <div class="flex items-center gap-4 flex-wrap">
          <USelectMenu
            v-model="selectedRegionId"
            :items="regions.map(r => ({ label: r.name, value: r.id }))"
            placeholder="Select a region"
            class="w-48"
            value-key="value"
          />
          <USelectMenu
            v-model="selectedProvinceId"
            :items="provinces.map(p => ({ label: p.name, value: p.id }))"
            placeholder="Select a province"
            class="w-48"
            :disabled="!selectedRegionId"
            value-key="value"
          />
          <USelectMenu
            v-model="selectedCityId"
            :items="cities.map(c => ({ label: c.name, value: c.id }))"
            placeholder="Select a city"
            class="w-48"
            :disabled="!selectedProvinceId"
            value-key="value"
          />
          <span v-if="selectedCityId" class="text-sm text-gray-500">
            {{ total }} barangays
          </span>
        </div>
      </template>

      <div v-if="!selectedCityId" class="py-8 text-center text-gray-500">
        <UIcon name="i-heroicons-arrow-up" class="w-8 h-8 mx-auto mb-2" />
        <p>Select a region, province, and city to view barangays</p>
      </div>

      <div v-else-if="loading" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      </div>

      <UTable
        v-else-if="barangays.length"
        :data="barangays"
        :columns="barangayColumns"
      >
        <template #code-cell="{ row }">
          <code class="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
            {{ row.original.code }}
          </code>
        </template>

        <template #name-cell="{ row }">
          <div>
            <p class="font-medium text-gray-900 dark:text-white">{{ row.original.name }}</p>
            <p class="text-sm text-gray-500">{{ row.original.slug }}</p>
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center gap-2 justify-end">
            <UButton
              :to="`/locations/barangay/${row.original.slug}`"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
            >
              View
            </UButton>
          </div>
        </template>
      </UTable>

      <div v-else class="py-8 text-center text-gray-500">
        No barangays found in this city/municipality.
      </div>

      <template v-if="totalPages > 1" #footer>
        <div class="flex justify-center">
          <UPagination v-model:page="page" :total="total" :items-per-page="perPage" />
        </div>
      </template>
    </UCard>

    <!-- Info Card -->
    <UCard class="mt-6">
      <template #header>
        <div class="flex items-center gap-2">
          <UIcon name="i-heroicons-information-circle" class="w-5 h-5 text-blue-500" />
          <span class="font-medium">About Philippine Locations</span>
        </div>
      </template>
      <div class="text-sm text-gray-600 dark:text-gray-400 space-y-2">
        <p>Location data follows the Philippine Standard Geographic Code (PSGC) structure:</p>
        <ul class="list-disc list-inside space-y-1 ml-2">
          <li><strong>17 Regions</strong> - Administrative regions of the Philippines</li>
          <li><strong>82 Provinces</strong> - Second-level administrative divisions</li>
          <li><strong>Cities &amp; Municipalities</strong> - Local government units (LGUs)</li>
          <li><strong>42,000+ Barangays</strong> - Smallest administrative division</li>
        </ul>
        <p class="mt-4">
          <strong>City Types:</strong>
          <span class="inline-flex gap-2 ml-2">
            <UBadge color="info" variant="subtle" size="xs">HUC</UBadge>
            <span>= Highly Urbanized City</span>
          </span>
          <span class="inline-flex gap-2 ml-2">
            <UBadge color="warning" variant="subtle" size="xs">Capital</UBadge>
            <span>= Provincial Capital</span>
          </span>
        </p>
      </div>
    </UCard>
  </div>
</template>
