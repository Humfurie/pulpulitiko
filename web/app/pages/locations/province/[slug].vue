<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)

const { data, pending, error } = await useAsyncData(
  `province-${slug.value}`,
  () => api.getProvinceBySlug(slug.value)
)

useSeoMeta({
  title: () => data.value ? `${data.value.province.name} - Pulpulitiko` : 'Province - Pulpulitiko',
  description: () => data.value ? `Browse cities and municipalities in ${data.value.province.name}` : 'Browse cities'
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
        <li class="text-stone-400">/</li>
        <li v-if="data?.province?.region">
          <NuxtLink
            :to="`/locations/region/${data.province.region.slug}`"
            class="text-stone-500 hover:text-orange-600"
          >
            {{ data.province.region.name }}
          </NuxtLink>
        </li>
        <li v-if="data?.province?.region" class="text-stone-400">/</li>
        <li class="text-stone-900 dark:text-white font-medium">
          {{ data?.province?.name || 'Province' }}
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
      <h2 class="text-xl font-semibold text-stone-900 dark:text-white mb-2">Province Not Found</h2>
      <p class="text-stone-600 dark:text-stone-400 mb-4">The province you're looking for doesn't exist.</p>
      <UButton to="/locations" color="orange">Back to Locations</UButton>
    </div>

    <!-- Content -->
    <div v-else-if="data">
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-stone-900 dark:text-white mb-2">
          {{ data.province.name }}
        </h1>
        <p class="text-stone-600 dark:text-stone-400">
          {{ data.cities.length }} cities/municipalities
        </p>
      </div>

      <!-- Cities Grid -->
      <div v-if="data.cities.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <NuxtLink
          v-for="city in data.cities"
          :key="city.id"
          :to="`/locations/city/${city.slug}`"
          class="block p-5 bg-white dark:bg-stone-800 rounded-lg border border-stone-200 dark:border-stone-700 hover:border-orange-300 dark:hover:border-orange-600 hover:shadow-md transition-all"
        >
          <div class="flex items-start justify-between">
            <div>
              <div class="flex items-center gap-2">
                <h3 class="font-semibold text-stone-900 dark:text-white">
                  {{ city.name }}
                </h3>
                <UBadge v-if="city.is_capital" color="orange" size="xs">Capital</UBadge>
                <UBadge v-else-if="city.is_city" color="blue" size="xs" variant="subtle">City</UBadge>
                <UBadge v-else size="xs" variant="subtle">Municipality</UBadge>
              </div>
              <p class="text-sm text-stone-500 dark:text-stone-400 mt-1">
                {{ city.barangay_count }} barangays
              </p>
            </div>
            <UIcon name="i-heroicons-chevron-right" class="w-5 h-5 text-stone-400" />
          </div>
        </NuxtLink>
      </div>

      <!-- No cities -->
      <div v-else class="text-center py-12 bg-stone-50 dark:bg-stone-800/50 rounded-xl">
        <UIcon name="i-heroicons-building-office-2" class="w-12 h-12 text-stone-400 mx-auto mb-4" />
        <p class="text-stone-600 dark:text-stone-400">No cities or municipalities found.</p>
      </div>
    </div>
  </div>
</template>
