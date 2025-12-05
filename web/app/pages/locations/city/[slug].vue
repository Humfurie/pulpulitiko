<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)

const { data, pending, error } = await useAsyncData(
  `city-${slug.value}`,
  () => api.getCityBySlug(slug.value)
)

useSeoMeta({
  title: () => data.value ? `${data.value.city.name} - Pulpulitiko` : 'City - Pulpulitiko',
  description: () => data.value ? `Browse barangays in ${data.value.city.name}` : 'Browse barangays'
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <!-- Breadcrumb -->
    <nav class="mb-6">
      <ol class="flex items-center gap-2 text-sm flex-wrap">
        <li>
          <NuxtLink to="/locations" class="text-stone-500 hover:text-orange-600">
            Locations
          </NuxtLink>
        </li>
        <template v-if="data?.city?.province">
          <li class="text-stone-400">/</li>
          <li>
            <NuxtLink
              :to="`/locations/province/${data.city.province.slug}`"
              class="text-stone-500 hover:text-orange-600"
            >
              {{ data.city.province.name }}
            </NuxtLink>
          </li>
        </template>
        <li class="text-stone-400">/</li>
        <li class="text-stone-900 dark:text-white font-medium">
          {{ data?.city?.name || 'City' }}
        </li>
      </ol>
    </nav>

    <!-- Loading -->
    <div v-if="pending" class="animate-pulse space-y-4">
      <div class="h-10 bg-stone-200 dark:bg-stone-700 rounded w-1/3" />
      <div class="h-6 bg-stone-200 dark:bg-stone-700 rounded w-1/2" />
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3 mt-8">
        <div v-for="i in 12" :key="i" class="h-16 bg-stone-200 dark:bg-stone-700 rounded-lg" />
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="text-center py-12">
      <UIcon name="i-heroicons-exclamation-triangle" class="w-12 h-12 text-red-500 mx-auto mb-4" />
      <h2 class="text-xl font-semibold text-stone-900 dark:text-white mb-2">City Not Found</h2>
      <p class="text-stone-600 dark:text-stone-400 mb-4">The city or municipality you're looking for doesn't exist.</p>
      <UButton to="/locations" color="orange">Back to Locations</UButton>
    </div>

    <!-- Content -->
    <div v-else-if="data">
      <div class="mb-8">
        <div class="flex items-center gap-3 mb-2">
          <h1 class="text-3xl font-bold text-stone-900 dark:text-white">
            {{ data.city.name }}
          </h1>
          <UBadge v-if="data.city.is_capital" color="orange" size="sm">Provincial Capital</UBadge>
          <UBadge v-else-if="data.city.is_huc" color="blue" size="sm">Highly Urbanized City</UBadge>
          <UBadge v-else-if="data.city.is_city" color="blue" size="sm" variant="subtle">City</UBadge>
          <UBadge v-else size="sm" variant="subtle">Municipality</UBadge>
        </div>
        <p class="text-stone-600 dark:text-stone-400">
          {{ data.barangays.total }} barangays
          <template v-if="data.city.population">
            &middot; Population: {{ data.city.population.toLocaleString() }}
          </template>
        </p>
      </div>

      <!-- Barangays Grid -->
      <div v-if="data.barangays.barangays.length > 0">
        <h2 class="text-xl font-semibold text-stone-800 dark:text-stone-200 mb-4">Barangays</h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          <div
            v-for="barangay in data.barangays.barangays"
            :key="barangay.id"
            class="p-4 bg-white dark:bg-stone-800 rounded-lg border border-stone-200 dark:border-stone-700"
          >
            <h3 class="font-medium text-stone-900 dark:text-white text-sm">
              {{ barangay.name }}
            </h3>
          </div>
        </div>

        <!-- Pagination info -->
        <div v-if="data.barangays.total_pages > 1" class="mt-6 text-center text-stone-500 dark:text-stone-400">
          Showing {{ data.barangays.barangays.length }} of {{ data.barangays.total }} barangays
        </div>
      </div>

      <!-- No barangays -->
      <div v-else class="text-center py-12 bg-stone-50 dark:bg-stone-800/50 rounded-xl">
        <UIcon name="i-heroicons-home" class="w-12 h-12 text-stone-400 mx-auto mb-4" />
        <p class="text-stone-600 dark:text-stone-400">No barangays found.</p>
      </div>
    </div>
  </div>
</template>
