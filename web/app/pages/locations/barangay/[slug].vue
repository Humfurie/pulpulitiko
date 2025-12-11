<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)

const { data, pending, error } = await useAsyncData(
  `barangay-${slug.value}`,
  () => api.getBarangayBySlug(slug.value)
)

useSeoMeta({
  title: () => data.value ? `${data.value.name} - Pulpulitiko` : 'Barangay - Pulpulitiko',
  description: () => data.value ? `Information about Barangay ${data.value.name}` : 'Barangay information'
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
        <template v-if="data?.city?.province?.region">
          <li class="text-stone-400">/</li>
          <li>
            <NuxtLink
              :to="`/locations/region/${data.city.province.region.slug}`"
              class="text-stone-500 hover:text-orange-600"
            >
              {{ data.city.province.region.name }}
            </NuxtLink>
          </li>
        </template>
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
        <template v-if="data?.city">
          <li class="text-stone-400">/</li>
          <li>
            <NuxtLink
              :to="`/locations/city/${data.city.slug}`"
              class="text-stone-500 hover:text-orange-600"
            >
              {{ data.city.name }}
            </NuxtLink>
          </li>
        </template>
        <li class="text-stone-400">/</li>
        <li class="text-stone-900 dark:text-white font-medium">
          {{ data?.name || 'Barangay' }}
        </li>
      </ol>
    </nav>

    <!-- Loading -->
    <div v-if="pending" class="animate-pulse space-y-4">
      <div class="h-10 bg-stone-200 dark:bg-stone-700 rounded w-1/3" />
      <div class="h-6 bg-stone-200 dark:bg-stone-700 rounded w-1/2" />
      <div class="h-32 bg-stone-200 dark:bg-stone-700 rounded-lg mt-8" />
    </div>

    <!-- Error -->
    <div v-else-if="error" class="text-center py-12">
      <UIcon name="i-heroicons-exclamation-triangle" class="w-12 h-12 text-red-500 mx-auto mb-4" />
      <h2 class="text-xl font-semibold text-stone-900 dark:text-white mb-2">Barangay Not Found</h2>
      <p class="text-stone-600 dark:text-stone-400 mb-4">The barangay you're looking for doesn't exist.</p>
      <UButton to="/locations" color="warning">Back to Locations</UButton>
    </div>

    <!-- Content -->
    <div v-else-if="data">
      <div class="mb-8">
        <div class="flex items-center gap-3 mb-2">
          <UIcon name="i-heroicons-home" class="w-8 h-8 text-purple-500" />
          <h1 class="text-3xl font-bold text-stone-900 dark:text-white">
            Barangay {{ data.name }}
          </h1>
        </div>
        <p class="text-stone-600 dark:text-stone-400">
          <template v-if="data.city">
            {{ data.city.name }},
          </template>
          <template v-if="data.city?.province">
            {{ data.city.province.name }}
          </template>
        </p>
      </div>

      <!-- Barangay Info Card -->
      <UCard class="mb-6">
        <template #header>
          <h2 class="font-semibold text-stone-900 dark:text-white">Barangay Information</h2>
        </template>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <h3 class="text-sm font-medium text-stone-500 dark:text-stone-400 mb-1">PSGC Code</h3>
            <code class="text-sm bg-stone-100 dark:bg-stone-800 px-2 py-1 rounded">
              {{ data.code }}
            </code>
          </div>

          <div v-if="data.population">
            <h3 class="text-sm font-medium text-stone-500 dark:text-stone-400 mb-1">Population</h3>
            <p class="text-stone-900 dark:text-white">{{ data.population.toLocaleString() }}</p>
          </div>

          <div v-if="data.city">
            <h3 class="text-sm font-medium text-stone-500 dark:text-stone-400 mb-1">City/Municipality</h3>
            <NuxtLink
              :to="`/locations/city/${data.city.slug}`"
              class="text-orange-600 hover:text-orange-700 hover:underline"
            >
              {{ data.city.name }}
            </NuxtLink>
          </div>

          <div v-if="data.city?.province">
            <h3 class="text-sm font-medium text-stone-500 dark:text-stone-400 mb-1">Province</h3>
            <NuxtLink
              :to="`/locations/province/${data.city.province.slug}`"
              class="text-orange-600 hover:text-orange-700 hover:underline"
            >
              {{ data.city.province.name }}
            </NuxtLink>
          </div>

          <div v-if="data.city?.province?.region">
            <h3 class="text-sm font-medium text-stone-500 dark:text-stone-400 mb-1">Region</h3>
            <NuxtLink
              :to="`/locations/region/${data.city.province.region.slug}`"
              class="text-orange-600 hover:text-orange-700 hover:underline"
            >
              {{ data.city.province.region.name }}
            </NuxtLink>
          </div>
        </div>
      </UCard>

      <!-- Placeholder for future features -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Local Officials Placeholder -->
        <UCard>
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-user-group" class="w-5 h-5 text-orange-500" />
              <h2 class="font-semibold text-stone-900 dark:text-white">Barangay Officials</h2>
            </div>
          </template>
          <div class="py-6 text-center">
            <UIcon name="i-heroicons-users" class="w-10 h-10 text-stone-300 dark:text-stone-600 mx-auto mb-3" />
            <p class="text-stone-500 dark:text-stone-400 text-sm">
              Barangay officials information coming soon
            </p>
          </div>
        </UCard>

        <!-- Related Articles Placeholder -->
        <UCard>
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-newspaper" class="w-5 h-5 text-blue-500" />
              <h2 class="font-semibold text-stone-900 dark:text-white">Related Articles</h2>
            </div>
          </template>
          <div class="py-6 text-center">
            <UIcon name="i-heroicons-document-text" class="w-10 h-10 text-stone-300 dark:text-stone-600 mx-auto mb-3" />
            <p class="text-stone-500 dark:text-stone-400 text-sm">
              Location-based articles coming soon
            </p>
          </div>
        </UCard>
      </div>

      <!-- Back Navigation -->
      <div class="mt-8 flex gap-4">
        <UButton
          v-if="data.city"
          :to="`/locations/city/${data.city.slug}`"
          variant="outline"
          icon="i-heroicons-arrow-left"
        >
          Back to {{ data.city.name }}
        </UButton>
        <UButton
          to="/locations"
          variant="ghost"
          icon="i-heroicons-map"
        >
          All Locations
        </UButton>
      </div>
    </div>
  </div>
</template>
