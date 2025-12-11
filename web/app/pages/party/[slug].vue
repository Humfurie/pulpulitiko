<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const route = useRoute()
const api = useApi()

const slug = computed(() => route.params.slug as string)

const { data: party, pending, error } = await useAsyncData(
  `party-${slug.value}`,
  () => api.getPartyBySlug(slug.value)
)

useSeoMeta({
  title: () => party.value ? `${party.value.name} - Pulpulitiko` : 'Political Party - Pulpulitiko',
  description: () => party.value?.description || `Learn about ${party.value?.name} political party`
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <!-- Loading -->
    <div v-if="pending" class="animate-pulse space-y-6">
      <div class="flex items-center gap-6">
        <div class="w-24 h-24 bg-stone-200 dark:bg-stone-700 rounded-lg" />
        <div class="space-y-3 flex-1">
          <div class="h-8 bg-stone-200 dark:bg-stone-700 rounded w-1/3" />
          <div class="h-4 bg-stone-200 dark:bg-stone-700 rounded w-1/4" />
        </div>
      </div>
      <div class="h-32 bg-stone-200 dark:bg-stone-700 rounded-lg" />
    </div>

    <!-- Error -->
    <div v-else-if="error" class="text-center py-12">
      <UIcon name="i-heroicons-exclamation-triangle" class="w-12 h-12 text-red-500 mx-auto mb-4" />
      <h2 class="text-xl font-semibold text-stone-900 dark:text-white mb-2">Party Not Found</h2>
      <p class="text-stone-600 dark:text-stone-400 mb-4">The political party you're looking for doesn't exist.</p>
      <UButton to="/parties" color="warning">View All Parties</UButton>
    </div>

    <!-- Content -->
    <div v-else-if="party">
      <!-- Party Header -->
      <div class="flex flex-col md:flex-row items-start gap-6 mb-8">
        <!-- Logo -->
        <div
          v-if="party.logo"
          class="w-24 h-24 rounded-lg overflow-hidden flex-shrink-0 bg-white dark:bg-stone-800 border border-stone-200 dark:border-stone-700"
        >
          <img :src="party.logo" :alt="party.name" class="w-full h-full object-contain" >
        </div>
        <div
          v-else
          class="w-24 h-24 rounded-lg flex items-center justify-center flex-shrink-0"
          :style="{ backgroundColor: party.color ? `${party.color}20` : '#f5f5f5' }"
        >
          <span
            class="text-3xl font-bold"
            :style="{ color: party.color || '#78716c' }"
          >
            {{ party.abbreviation || party.name.charAt(0) }}
          </span>
        </div>

        <!-- Info -->
        <div class="flex-1">
          <div class="flex flex-wrap items-center gap-3 mb-2">
            <h1 class="text-3xl font-bold text-stone-900 dark:text-white">
              {{ party.name }}
            </h1>
            <UBadge v-if="party.abbreviation" variant="subtle" size="lg">
              {{ party.abbreviation }}
            </UBadge>
            <UBadge v-if="party.is_major" color="warning" size="sm">Major Party</UBadge>
            <UBadge v-if="!party.is_active" color="error" size="sm">Inactive</UBadge>
          </div>

          <div class="flex flex-wrap items-center gap-4 text-stone-600 dark:text-stone-400">
            <span v-if="party.founded_year" class="flex items-center gap-1">
              <UIcon name="i-heroicons-calendar" class="w-4 h-4" />
              Founded {{ party.founded_year }}
            </span>
            <span v-if="party.member_count" class="flex items-center gap-1">
              <UIcon name="i-heroicons-users" class="w-4 h-4" />
              {{ party.member_count }} registered member(s)
            </span>
            <a
              v-if="party.website"
              :href="party.website"
              target="_blank"
              rel="noopener noreferrer"
              class="flex items-center gap-1 text-orange-600 hover:text-orange-700 hover:underline"
            >
              <UIcon name="i-heroicons-globe-alt" class="w-4 h-4" />
              Website
            </a>
          </div>

          <!-- Color indicator -->
          <div v-if="party.color" class="mt-3 flex items-center gap-2">
            <span class="text-sm text-stone-500">Party Color:</span>
            <div
              class="w-6 h-6 rounded-full border border-stone-300 dark:border-stone-600"
              :style="{ backgroundColor: party.color }"
            />
            <code class="text-xs bg-stone-100 dark:bg-stone-800 px-2 py-1 rounded">
              {{ party.color }}
            </code>
          </div>
        </div>
      </div>

      <!-- Description -->
      <UCard v-if="party.description" class="mb-8">
        <template #header>
          <span class="font-medium">About the Party</span>
        </template>
        <div class="prose prose-stone dark:prose-invert max-w-none">
          <p>{{ party.description }}</p>
        </div>
      </UCard>

      <!-- Party Details Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <!-- Quick Facts -->
        <UCard>
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-information-circle" class="w-5 h-5 text-blue-500" />
              <span class="font-medium">Quick Facts</span>
            </div>
          </template>
          <dl class="space-y-3">
            <div v-if="party.abbreviation">
              <dt class="text-sm text-stone-500 dark:text-stone-400">Abbreviation</dt>
              <dd class="font-medium text-stone-900 dark:text-white">{{ party.abbreviation }}</dd>
            </div>
            <div v-if="party.founded_year">
              <dt class="text-sm text-stone-500 dark:text-stone-400">Year Founded</dt>
              <dd class="font-medium text-stone-900 dark:text-white">{{ party.founded_year }}</dd>
            </div>
            <div>
              <dt class="text-sm text-stone-500 dark:text-stone-400">Status</dt>
              <dd>
                <UBadge :color="party.is_active ? 'success' : 'error'" variant="subtle">
                  {{ party.is_active ? 'Active' : 'Inactive' }}
                </UBadge>
              </dd>
            </div>
            <div>
              <dt class="text-sm text-stone-500 dark:text-stone-400">Classification</dt>
              <dd>
                <UBadge :color="party.is_major ? 'warning' : 'neutral'" variant="subtle">
                  {{ party.is_major ? 'Major Party' : 'Minor Party' }}
                </UBadge>
              </dd>
            </div>
          </dl>
        </UCard>

        <!-- Members Placeholder -->
        <UCard>
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-users" class="w-5 h-5 text-orange-500" />
              <span class="font-medium">Notable Members</span>
            </div>
          </template>
          <div class="py-6 text-center">
            <UIcon name="i-heroicons-user-group" class="w-10 h-10 text-stone-300 dark:text-stone-600 mx-auto mb-3" />
            <p class="text-stone-500 dark:text-stone-400 text-sm">
              List of party members coming soon
            </p>
          </div>
        </UCard>
      </div>

      <!-- Related Articles Placeholder -->
      <UCard>
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-newspaper" class="w-5 h-5 text-blue-500" />
            <span class="font-medium">Related Articles</span>
          </div>
        </template>
        <div class="py-8 text-center">
          <UIcon name="i-heroicons-document-text" class="w-12 h-12 text-stone-300 dark:text-stone-600 mx-auto mb-4" />
          <p class="text-stone-500 dark:text-stone-400">
            Articles mentioning {{ party.name }} will appear here
          </p>
        </div>
      </UCard>

      <!-- Back Navigation -->
      <div class="mt-8">
        <UButton to="/parties" variant="outline" icon="i-heroicons-arrow-left">
          All Political Parties
        </UButton>
      </div>
    </div>
  </div>
</template>
