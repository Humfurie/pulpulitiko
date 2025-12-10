<script setup lang="ts">
const route = useRoute()
const api = useApi()
const slug = route.params.slug as string

const { data: content, status } = await useAsyncData(
  `voter-education-${slug}`,
  () => api.getVoterEducationBySlug(slug)
)

function getContentTypeIcon(type: string): string {
  switch (type) {
    case 'guide':
      return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
    case 'faq':
      return 'M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
    case 'video':
      return 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z'
    default:
      return 'M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z'
  }
}

function getContentTypeLabel(type: string): string {
  switch (type) {
    case 'guide':
      return 'Guide'
    case 'faq':
      return 'FAQ'
    case 'video':
      return 'Video'
    default:
      return 'Article'
  }
}

function getCategoryLabel(cat: string | undefined): string {
  if (!cat) return ''
  const mapping: Record<string, string> = {
    registration: 'Registration',
    voting_day: 'Voting Day',
    requirements: 'Requirements',
    general: 'General Information'
  }
  return mapping[cat] || cat.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

useSeoMeta({
  title: () => content.value ? `${content.value.title} - Voter Education - Pulpulitiko` : 'Voter Education - Pulpulitiko',
  description: () => content.value?.content?.substring(0, 160) || 'Voter education content'
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <!-- Loading State -->
    <div v-if="status === 'pending'" class="flex justify-center items-center min-h-screen">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-green-600" />
    </div>

    <!-- Error State -->
    <div v-else-if="!content" class="max-w-7xl mx-auto px-4 py-16 text-center">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">Content Not Found</h1>
      <p class="text-gray-600 dark:text-gray-400 mb-8">The content you're looking for doesn't exist.</p>
      <NuxtLink
        to="/voter-education"
        class="inline-flex items-center px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
      >
        Back to Voter Education
      </NuxtLink>
    </div>

    <!-- Content -->
    <template v-else>
      <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <!-- Breadcrumb -->
        <div class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-6">
          <NuxtLink to="/voter-education" class="hover:text-green-600">Voter Education</NuxtLink>
          <span>/</span>
          <span class="text-gray-900 dark:text-white">{{ content.title }}</span>
        </div>

        <!-- Header -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8 mb-8">
          <div class="flex items-center gap-3 mb-4">
            <div class="flex items-center gap-2 text-green-600 dark:text-green-400">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="getContentTypeIcon(content.content_type)" />
              </svg>
              <span class="text-sm font-medium">{{ getContentTypeLabel(content.content_type) }}</span>
            </div>
            <span
              v-if="content.category"
              class="text-sm px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded"
            >
              {{ getCategoryLabel(content.category) }}
            </span>
            <span
              v-if="content.is_featured"
              class="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 rounded-full"
            >
              Featured
            </span>
          </div>

          <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-4">
            {{ content.title }}
          </h1>

          <div class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
            <div v-if="content.published_at" class="flex items-center gap-1">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <span>{{ new Date(content.published_at).toLocaleDateString('en-PH', { year: 'numeric', month: 'long', day: 'numeric' }) }}</span>
            </div>
            <div class="flex items-center gap-1">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              <span>{{ content.view_count.toLocaleString() }} views</span>
            </div>
          </div>

          <!-- Related Election -->
          <div v-if="content.election" class="mt-4 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
            <div class="flex items-center gap-2">
              <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
              <span class="text-sm text-blue-600 dark:text-blue-400">Related to:</span>
              <NuxtLink
                :to="`/elections/${content.election.slug}`"
                class="font-medium text-blue-700 dark:text-blue-300 hover:underline"
              >
                {{ content.election.name }}
              </NuxtLink>
            </div>
          </div>
        </div>

        <!-- Content Body -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8">
          <div
            class="prose prose-lg dark:prose-invert max-w-none"
            v-html="content.content"
          />
        </div>

        <!-- Back Link -->
        <div class="mt-8 text-center">
          <NuxtLink
            to="/voter-education"
            class="inline-flex items-center gap-2 text-green-600 hover:text-green-700 dark:text-green-400"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Back to Voter Education
          </NuxtLink>
        </div>
      </div>
    </template>
  </div>
</template>
