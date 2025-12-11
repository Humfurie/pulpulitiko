<script setup lang="ts">
import type { PoliticianListItem, LocationHierarchy } from '~/types'

definePageMeta({
  layout: 'default'
})

useSeoMeta({
  title: 'Find My Representatives - Pulpulitiko',
  description: 'Find your elected representatives from national to barangay level based on your location'
})

const api = useApi()

const selectedLocation = ref<LocationHierarchy | null>(null)
const representatives = ref<PoliticianListItem[]>([])
const loading = ref(false)
const searched = ref(false)
const error = ref('')

async function onLocationChange(location: LocationHierarchy) {
  selectedLocation.value = location

  if (!location.barangay?.id) {
    representatives.value = []
    searched.value = false
    return
  }

  loading.value = true
  error.value = ''
  searched.value = true

  try {
    representatives.value = await api.findMyRepresentatives(location.barangay.id)
  } catch (e) {
    console.error('Failed to find representatives:', e)
    error.value = 'Failed to find representatives. Please try again.'
    representatives.value = []
  } finally {
    loading.value = false
  }
}

// Group representatives by level
const groupedRepresentatives = computed(() => {
  const groups: Record<string, PoliticianListItem[]> = {
    national: [],
    regional: [],
    provincial: [],
    city: [],
    municipal: [],
    barangay: []
  }

  for (const rep of representatives.value) {
    const level = rep.level || 'national'
    const targetGroup = groups[level]
    if (targetGroup) {
      targetGroup.push(rep)
    } else {
      // Fallback to national (always exists)
      groups.national!.push(rep)
    }
  }

  return groups
})

const levelLabels: Record<string, string> = {
  national: 'National',
  regional: 'Regional',
  provincial: 'Provincial',
  city: 'City',
  municipal: 'Municipal',
  barangay: 'Barangay'
}

const levelIcons: Record<string, string> = {
  national: 'i-heroicons-globe-asia-australia',
  regional: 'i-heroicons-map',
  provincial: 'i-heroicons-building-library',
  city: 'i-heroicons-building-office-2',
  municipal: 'i-heroicons-building-office',
  barangay: 'i-heroicons-home'
}

function getLevelColor(level: string): 'error' | 'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'neutral' {
  const colors: Record<string, 'error' | 'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'neutral'> = {
    national: 'error',
    regional: 'secondary',
    provincial: 'info',
    city: 'warning',
    municipal: 'success',
    barangay: 'primary'
  }
  return colors[level] || 'neutral'
}
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <!-- Header -->
    <div class="mb-8 text-center">
      <h1 class="text-3xl font-bold text-stone-900 dark:text-white mb-2">
        Find My Representatives
      </h1>
      <p class="text-stone-600 dark:text-stone-400 max-w-2xl mx-auto">
        Select your barangay to see all your elected representatives from national government down to your local barangay officials.
      </p>
    </div>

    <!-- Location Picker -->
    <div class="max-w-3xl mx-auto mb-8">
      <UCard>
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-map-pin" class="w-5 h-5 text-orange-500" />
            <span class="font-medium">Select Your Location</span>
          </div>
        </template>
        <LocationPicker @change="onLocationChange" />
      </UCard>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-12">
      <UIcon name="i-heroicons-arrow-path" class="w-10 h-10 text-stone-400 animate-spin mx-auto mb-4" />
      <p class="text-stone-600 dark:text-stone-400">Finding your representatives...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="max-w-xl mx-auto">
      <UAlert color="error" :title="error" icon="i-heroicons-exclamation-triangle" />
    </div>

    <!-- Initial State -->
    <div v-else-if="!searched" class="text-center py-12">
      <UIcon name="i-heroicons-user-group" class="w-16 h-16 text-stone-300 dark:text-stone-600 mx-auto mb-4" />
      <h2 class="text-xl font-semibold text-stone-700 dark:text-stone-300 mb-2">
        Select Your Barangay
      </h2>
      <p class="text-stone-500 dark:text-stone-400">
        Choose your location above to find your representatives
      </p>
    </div>

    <!-- Results -->
    <div v-else-if="representatives.length > 0" class="space-y-8">
      <!-- Location Summary -->
      <div v-if="selectedLocation" class="text-center mb-6">
        <p class="text-stone-600 dark:text-stone-400">
          Representatives for
          <span class="font-medium text-stone-900 dark:text-white">
            {{ selectedLocation.barangay?.name }},
            {{ selectedLocation.city_municipality?.name }},
            {{ selectedLocation.province?.name }}
          </span>
        </p>
        <p class="text-sm text-stone-500 mt-1">
          Found {{ representatives.length }} representative(s)
        </p>
      </div>

      <!-- Representatives by Level -->
      <div
        v-for="(level, levelKey) in groupedRepresentatives"
        :key="levelKey"
        class="space-y-4"
      >
        <template v-if="level.length > 0">
          <div class="flex items-center gap-2">
            <UIcon :name="levelIcons[levelKey]" class="w-5 h-5" :class="`text-${getLevelColor(levelKey)}-500`" />
            <h2 class="text-lg font-semibold text-stone-800 dark:text-stone-200">
              {{ levelLabels[levelKey] }} Level
            </h2>
            <UBadge :color="getLevelColor(levelKey)" variant="subtle" size="sm">
              {{ level.length }}
            </UBadge>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <NuxtLink
              v-for="rep in level"
              :key="rep.id"
              :to="`/politician/${rep.slug}`"
              class="block p-4 bg-white dark:bg-stone-800 rounded-lg border border-stone-200 dark:border-stone-700 hover:border-orange-300 dark:hover:border-orange-600 hover:shadow-md transition-all"
            >
              <div class="flex items-start gap-4">
                <UAvatar
                  :src="rep.photo"
                  :alt="rep.name"
                  size="lg"
                  class="rounded-lg"
                />
                <div class="flex-1 min-w-0">
                  <h3 class="font-semibold text-stone-900 dark:text-white truncate">
                    {{ rep.name }}
                  </h3>
                  <p v-if="rep.position" class="text-sm text-stone-600 dark:text-stone-400">
                    {{ rep.position }}
                  </p>
                  <div class="flex items-center gap-2 mt-2">
                    <UBadge
                      v-if="rep.party_info"
                      :style="rep.party_info.color ? { backgroundColor: `${rep.party_info.color}20`, color: rep.party_info.color } : {}"
                      variant="subtle"
                      size="xs"
                    >
                      {{ rep.party_info.abbreviation || rep.party_info.name }}
                    </UBadge>
                    <UBadge v-else-if="rep.party" variant="subtle" size="xs">
                      {{ rep.party }}
                    </UBadge>
                    <span v-if="rep.branch" class="text-xs text-stone-500 capitalize">
                      {{ rep.branch }}
                    </span>
                  </div>
                </div>
                <UIcon name="i-heroicons-chevron-right" class="w-5 h-5 text-stone-400 flex-shrink-0" />
              </div>
            </NuxtLink>
          </div>
        </template>
      </div>
    </div>

    <!-- No Results -->
    <div v-else class="text-center py-12">
      <UIcon name="i-heroicons-user-group" class="w-16 h-16 text-stone-300 dark:text-stone-600 mx-auto mb-4" />
      <h2 class="text-xl font-semibold text-stone-700 dark:text-stone-300 mb-2">
        No Representatives Found
      </h2>
      <p class="text-stone-500 dark:text-stone-400 max-w-md mx-auto">
        We don't have representative data for this location yet. This information is being updated regularly.
      </p>
    </div>

    <!-- Info Section -->
    <div class="mt-12 max-w-3xl mx-auto">
      <UCard>
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-information-circle" class="w-5 h-5 text-blue-500" />
            <span class="font-medium">About This Feature</span>
          </div>
        </template>
        <div class="text-sm text-stone-600 dark:text-stone-400 space-y-3">
          <p>
            The "Find My Representatives" feature shows you all elected officials who represent your area, from the President down to your Barangay Kagawad.
          </p>
          <p>
            <strong>Government Levels:</strong>
          </p>
          <ul class="list-disc list-inside space-y-1 ml-2">
            <li><strong>National</strong> - President, Vice President, Senators, House Representatives</li>
            <li><strong>Provincial</strong> - Governor, Vice Governor, Provincial Board Members</li>
            <li><strong>City/Municipal</strong> - Mayor, Vice Mayor, Councilors</li>
            <li><strong>Barangay</strong> - Barangay Captain, Kagawad, SK Officials</li>
          </ul>
        </div>
      </UCard>
    </div>
  </div>
</template>
