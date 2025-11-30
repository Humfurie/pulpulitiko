<script setup lang="ts">
const props = defineProps<{
  currentPage: number
  totalPages: number
  baseUrl?: string
}>()

const emit = defineEmits<{
  (e: 'change', page: number): void
}>()

const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const current = props.currentPage
  const total = props.totalPages

  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }

  pages.push(1)

  if (current > 3) {
    pages.push('...')
  }

  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }

  if (current < total - 2) {
    pages.push('...')
  }

  pages.push(total)

  return pages
})

function goToPage(page: number) {
  if (page >= 1 && page <= props.totalPages && page !== props.currentPage) {
    emit('change', page)
  }
}
</script>

<template>
  <nav v-if="totalPages > 1" class="flex items-center justify-center gap-1" aria-label="Pagination">
    <UButton
      variant="ghost"
      icon="i-heroicons-chevron-left"
      :disabled="currentPage <= 1"
      @click="goToPage(currentPage - 1)"
    />

    <template v-for="(page, index) in visiblePages" :key="index">
      <span
        v-if="page === '...'"
        class="px-3 py-2 text-gray-500"
      >
        ...
      </span>
      <UButton
        v-else
        :variant="page === currentPage ? 'solid' : 'ghost'"
        :color="page === currentPage ? 'primary' : 'neutral'"
        class="min-w-[40px]"
        @click="goToPage(page as number)"
      >
        {{ page }}
      </UButton>
    </template>

    <UButton
      variant="ghost"
      icon="i-heroicons-chevron-right"
      :disabled="currentPage >= totalPages"
      @click="goToPage(currentPage + 1)"
    />
  </nav>
</template>
