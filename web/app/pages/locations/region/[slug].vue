<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)

const { data, pending, error } = await useAsyncData(
  `region-${slug.value}`,
  () => api.getRegionBySlug(slug.value)
)

useSeoMeta({
  title: () => data.value ? `${data.value.region.name} - Pulpulitiko` : 'Region - Pulpulitiko',
  description: () => data.value ? `Browse provinces in ${data.value.region.name}` : 'Browse provinces'
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <!-- Breadcrumb -->
    <nav class="mb-6">
      <ol class="flex items-center gap-2 text-sm">
        <li>
          <NuxtLink to="/locations" class="text-stone-500 hover:text-orange-600">
            Locations
          </NuxtLink>
        </li>
        <li class="text-stone-400">/</li>
        <li class="text-stone-900 dark:text-white font-medium">
          {{ data?.region?.name || 'Region' }}
        </li>
      </ol>
    </nav>

    <!-- Loading -->
    <div v-if="pending" class="animate-pulse space-y-4">
      <div class="h-10 bg-stone-200 dark:bg-stone-700 rounded w-1/3" />
      <div class="h-6 bg-stone-200 dark:bg-stone-700 rounded w-1/2" />
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-8">
        <div v-for="i in 6" :key="i" class="h-24 bg-stone-200 dark:bg-stone-700 rounded-lg" />
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="text-center py-12">
      <UIcon name="i-heroicons-exclamation-triangle" class="w-12 h-12 text-red-500 mx-auto mb-4" />
      <h2 class="text-xl font-semibold text-stone-900 dark:text-white mb-2">Region Not Found</h2>
      <p class="text-stone-600 dark:text-stone-400 mb-4">The region you're looking for doesn't exist.</p>
      <UButton to="/locations" color="warning">Back to Locations</UButton>
    </div>

    <!-- Content -->
    <div v-else-if="data">
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-stone-900 dark:text-white mb-2">
          {{ data.region.name }}
        </h1>
        <p class="text-stone-600 dark:text-stone-400">
          {{ data.provinces.length }} {{ data.provinces.length === 1 ? 'province' : 'provinces' }}
        </p>
      </div>

      <!-- Provinces Grid -->
      <div v-if="data.provinces.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <NuxtLink
          v-for="province in data.provinces"
          :key="province.id"
          :to="`/locations/province/${province.slug}`"
          class="block p-5 bg-white dark:bg-stone-800 rounded-lg border border-stone-200 dark:border-stone-700 hover:border-orange-300 dark:hover:border-orange-600 hover:shadow-md transition-all"
        >
          <div class="flex items-start justify-between">
            <div>
              <h3 class="font-semibold text-stone-900 dark:text-white">
                {{ province.name }}
              </h3>
              <p class="text-sm text-stone-500 dark:text-stone-400 mt-1">
                {{ province.city_count }} cities/municipalities
              </p>
            </div>
            <UIcon name="i-heroicons-chevron-right" class="w-5 h-5 text-stone-400" />
          </div>
        </NuxtLink>
      </div>

      <!-- No provinces -->
      <div v-else class="text-center py-12 bg-stone-50 dark:bg-stone-800/50 rounded-xl">
        <UIcon name="i-heroicons-map" class="w-12 h-12 text-stone-400 mx-auto mb-4" />
        <p class="text-stone-600 dark:text-stone-400">No provinces found in this region.</p>
      </div>
    </div>
  </div>
</template>
