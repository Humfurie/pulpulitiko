<script setup lang="ts">
import type { LocationSearchResult } from '~/types'

const api = useApi()
const router = useRouter()

const searchQuery = ref('')
const searchResults = ref<LocationSearchResult[]>([])
const isSearching = ref(false)
const showResults = ref(false)

// Debounced search
let searchTimeout: ReturnType<typeof setTimeout> | null = null

watch(searchQuery, (query) => {
  if (searchTimeout) clearTimeout(searchTimeout)

  if (!query || query.length < 2) {
    searchResults.value = []
    showResults.value = false
    return
  }

  isSearching.value = true
  searchTimeout = setTimeout(async () => {
    try {
      searchResults.value = await api.searchLocations(query, 10)
      showResults.value = true
    } catch (error) {
      console.error('Search failed:', error)
    } finally {
      isSearching.value = false
    }
  }, 300)
})

function getLocationRoute(result: LocationSearchResult): string {
  switch (result.type) {
    case 'region':
      return `/locations/region/${result.slug}`
    case 'province':
      return `/locations/province/${result.slug}`
    case 'city':
      return `/locations/city/${result.slug}`
    case 'barangay':
      return `/locations/barangay/${result.slug}`
    default:
      return '/locations'
  }
}

function getTypeIcon(type: string): string {
  switch (type) {
    case 'region':
      return 'i-heroicons-globe-asia-australia'
    case 'province':
      return 'i-heroicons-map'
    case 'city':
      return 'i-heroicons-building-office-2'
    case 'barangay':
      return 'i-heroicons-home'
    default:
      return 'i-heroicons-map-pin'
  }
}

function getTypeBadgeColor(type: string): 'error' | 'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'neutral' {
  switch (type) {
    case 'region':
      return 'info'
    case 'province':
      return 'success'
    case 'city':
      return 'warning'
    case 'barangay':
      return 'secondary'
    default:
      return 'neutral'
  }
}

function selectResult(result: LocationSearchResult) {
  router.push(getLocationRoute(result))
  showResults.value = false
  searchQuery.value = ''
}

function onBlur() {
  // Delay hiding results to allow click
  setTimeout(() => {
    showResults.value = false
  }, 200)
}
</script>

<template>
  <div class="relative">
    <UInput
      v-model="searchQuery"
      icon="i-heroicons-magnifying-glass"
      placeholder="Search for a region, province, city, or barangay..."
      size="lg"
      :loading="isSearching"
      @focus="showResults = searchResults.length > 0"
      @blur="onBlur"
    />

    <!-- Search Results Dropdown -->
    <div
      v-if="showResults && searchResults.length > 0"
      class="absolute top-full left-0 right-0 mt-2 bg-white dark:bg-stone-800 rounded-lg shadow-lg border border-stone-200 dark:border-stone-700 z-50 max-h-96 overflow-y-auto"
    >
      <div
        v-for="result in searchResults"
        :key="result.id"
        role="button"
        tabindex="0"
        class="w-full px-4 py-3 text-left hover:bg-stone-50 dark:hover:bg-stone-700 flex items-center gap-3 border-b border-stone-100 dark:border-stone-700 last:border-0 cursor-pointer"
        @click="selectResult(result)"
        @keydown.enter="selectResult(result)"
        @keydown.space.prevent="selectResult(result)"
      >
        <UIcon :name="getTypeIcon(result.type)" class="w-5 h-5 text-stone-400 flex-shrink-0" />
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <span class="font-medium text-stone-900 dark:text-white truncate">
              {{ result.name }}
            </span>
            <UBadge :color="getTypeBadgeColor(result.type)" size="xs" variant="subtle">
              {{ result.type }}
            </UBadge>
          </div>
          <p class="text-sm text-stone-500 dark:text-stone-400 truncate">
            {{ result.full_path }}
          </p>
        </div>
        <UIcon name="i-heroicons-chevron-right" class="w-4 h-4 text-stone-400 flex-shrink-0" />
      </div>
    </div>

    <!-- No Results -->
    <div
      v-else-if="showResults && searchQuery.length >= 2 && !isSearching"
      class="absolute top-full left-0 right-0 mt-2 bg-white dark:bg-stone-800 rounded-lg shadow-lg border border-stone-200 dark:border-stone-700 z-50 p-4 text-center"
    >
      <UIcon name="i-heroicons-magnifying-glass" class="w-8 h-8 text-stone-400 mx-auto mb-2" />
      <p class="text-stone-600 dark:text-stone-400">No locations found for "{{ searchQuery }}"</p>
    </div>
  </div>
</template>
