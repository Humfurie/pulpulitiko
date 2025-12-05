<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

useSeoMeta({
  title: 'Philippine Locations - Pulpulitiko',
  description: 'Browse Philippine regions, provinces, cities, municipalities, and barangays'
})

const api = useApi()

const { data: regions, pending } = await useAsyncData('regions', () => api.getRegions())

// Group regions by island group
const islandGroups = computed(() => {
  if (!regions.value) return []

  const luzonCodes = ['010000000', '020000000', '030000000', '040000000', '170000000', '050000000', '130000000', '140000000']
  const visayasCodes = ['060000000', '070000000', '080000000']
  const mindanaoCodes = ['090000000', '100000000', '110000000', '120000000', '160000000', '190000000']

  const luzon = regions.value.filter(r => luzonCodes.includes(r.code))
  const visayas = regions.value.filter(r => visayasCodes.includes(r.code))
  const mindanao = regions.value.filter(r => mindanaoCodes.includes(r.code))

  return [
    { name: 'Luzon', regions: luzon },
    { name: 'Visayas', regions: visayas },
    { name: 'Mindanao', regions: mindanao }
  ]
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-stone-900 dark:text-white mb-2">
        Philippine Locations
      </h1>
      <p class="text-stone-600 dark:text-stone-400">
        Browse Philippine regions, provinces, cities, municipalities, and barangays
      </p>
    </div>

    <!-- Loading State -->
    <div v-if="pending" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="i in 6" :key="i" class="animate-pulse">
        <div class="bg-stone-200 dark:bg-stone-700 rounded-lg h-40" />
      </div>
    </div>

    <!-- Island Groups -->
    <div v-else class="space-y-10">
      <div v-for="group in islandGroups" :key="group.name">
        <h2 class="text-2xl font-semibold text-stone-800 dark:text-stone-200 mb-4 flex items-center gap-2">
          <UIcon name="i-heroicons-map" class="w-6 h-6 text-orange-500" />
          {{ group.name }}
        </h2>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <NuxtLink
            v-for="region in group.regions"
            :key="region.id"
            :to="`/locations/region/${region.slug}`"
            class="block p-5 bg-white dark:bg-stone-800 rounded-lg border border-stone-200 dark:border-stone-700 hover:border-orange-300 dark:hover:border-orange-600 hover:shadow-md transition-all"
          >
            <div class="flex items-start justify-between">
              <div>
                <h3 class="font-semibold text-stone-900 dark:text-white">
                  {{ region.name }}
                </h3>
                <p class="text-sm text-stone-500 dark:text-stone-400 mt-1">
                  {{ region.province_count }} {{ region.province_count === 1 ? 'province' : 'provinces' }}
                </p>
              </div>
              <UIcon name="i-heroicons-chevron-right" class="w-5 h-5 text-stone-400" />
            </div>
          </NuxtLink>
        </div>
      </div>
    </div>

    <!-- Location Search -->
    <div class="mt-12 bg-stone-50 dark:bg-stone-800/50 rounded-xl p-6">
      <h2 class="text-xl font-semibold text-stone-800 dark:text-stone-200 mb-4">
        Search Locations
      </h2>
      <LocationSearch />
    </div>
  </div>
</template>
