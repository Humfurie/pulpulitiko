<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

useSeoMeta({
  title: 'Political Parties - Pulpulitiko',
  description: 'Browse Philippine political parties and their members'
})

const api = useApi()

const page = ref(1)
const perPage = 20
const showMajorOnly = ref(false)

const { data, pending } = await useAsyncData(
  'parties',
  () => api.getParties(page.value, perPage, showMajorOnly.value, true),
  { watch: [page, showMajorOnly] }
)

watch(showMajorOnly, () => {
  page.value = 1
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <!-- Header -->
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-stone-900 dark:text-white mb-2">
        Political Parties
      </h1>
      <p class="text-stone-600 dark:text-stone-400">
        Browse Philippine political parties and learn about their history, leadership, and platform
      </p>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-4 mb-6">
      <UCheckbox v-model="showMajorOnly" label="Show major parties only" />
      <span v-if="data" class="text-sm text-stone-500">
        {{ data.total }} parties found
      </span>
    </div>

    <!-- Loading -->
    <div v-if="pending" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="i in 6" :key="i" class="animate-pulse">
        <div class="bg-stone-200 dark:bg-stone-700 rounded-lg h-40" />
      </div>
    </div>

    <!-- Parties Grid -->
    <div v-else-if="data && data.parties.length > 0">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <NuxtLink
          v-for="party in data.parties"
          :key="party.id"
          :to="`/party/${party.slug}`"
          class="block p-5 bg-white dark:bg-stone-800 rounded-lg border border-stone-200 dark:border-stone-700 hover:border-orange-300 dark:hover:border-orange-600 hover:shadow-md transition-all"
        >
          <div class="flex items-start gap-4">
            <!-- Logo/Abbreviation -->
            <div
              v-if="party.logo"
              class="w-14 h-14 rounded-lg overflow-hidden flex-shrink-0 bg-white border border-stone-200 dark:border-stone-700"
            >
              <img :src="party.logo" :alt="party.name" class="w-full h-full object-contain" >
            </div>
            <div
              v-else
              class="w-14 h-14 rounded-lg flex items-center justify-center flex-shrink-0"
              :style="{ backgroundColor: party.color ? `${party.color}20` : '#f5f5f5' }"
            >
              <span
                class="text-xl font-bold"
                :style="{ color: party.color || '#78716c' }"
              >
                {{ party.abbreviation?.substring(0, 2) || party.name.charAt(0) }}
              </span>
            </div>

            <!-- Info -->
            <div class="flex-1 min-w-0">
              <div class="flex items-start gap-2">
                <h3 class="font-semibold text-stone-900 dark:text-white line-clamp-2">
                  {{ party.name }}
                </h3>
              </div>
              <div class="flex flex-wrap items-center gap-2 mt-2">
                <UBadge v-if="party.abbreviation" variant="subtle" size="xs">
                  {{ party.abbreviation }}
                </UBadge>
                <UBadge v-if="party.is_major" color="warning" size="xs">Major</UBadge>
                <span v-if="party.member_count > 0" class="text-xs text-stone-500">
                  {{ party.member_count }} member(s)
                </span>
              </div>
            </div>

            <!-- Color indicator -->
            <div
              v-if="party.color"
              class="w-3 h-12 rounded-full flex-shrink-0"
              :style="{ backgroundColor: party.color }"
            />
          </div>
        </NuxtLink>
      </div>

      <!-- Pagination -->
      <div v-if="data.total_pages > 1" class="mt-8 flex justify-center">
        <UPagination v-model:page="page" :total="data.total" :items-per-page="perPage" />
      </div>
    </div>

    <!-- No Results -->
    <div v-else class="text-center py-12">
      <UIcon name="i-heroicons-flag" class="w-16 h-16 text-stone-300 dark:text-stone-600 mx-auto mb-4" />
      <h2 class="text-xl font-semibold text-stone-700 dark:text-stone-300 mb-2">
        No Parties Found
      </h2>
      <p class="text-stone-500 dark:text-stone-400">
        No political parties match your criteria
      </p>
    </div>
  </div>
</template>
