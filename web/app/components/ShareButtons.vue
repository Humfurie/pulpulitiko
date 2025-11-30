<script setup lang="ts">
const props = defineProps<{
  title: string
  url?: string
}>()

const shareUrl = computed(() => {
  if (import.meta.client) {
    return props.url || window.location.href
  }
  return props.url || ''
})

const encodedUrl = computed(() => encodeURIComponent(shareUrl.value))
const encodedTitle = computed(() => encodeURIComponent(props.title))

const shareLinks = computed(() => [
  {
    name: 'Facebook',
    icon: 'i-simple-icons-facebook',
    url: `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl.value}`,
    color: 'hover:text-blue-600'
  },
  {
    name: 'X (Twitter)',
    icon: 'i-simple-icons-x',
    url: `https://twitter.com/intent/tweet?url=${encodedUrl.value}&text=${encodedTitle.value}`,
    color: 'hover:text-gray-900 dark:hover:text-white'
  },
  {
    name: 'LinkedIn',
    icon: 'i-simple-icons-linkedin',
    url: `https://www.linkedin.com/shareArticle?mini=true&url=${encodedUrl.value}&title=${encodedTitle.value}`,
    color: 'hover:text-blue-700'
  },
  {
    name: 'WhatsApp',
    icon: 'i-simple-icons-whatsapp',
    url: `https://wa.me/?text=${encodedTitle.value}%20${encodedUrl.value}`,
    color: 'hover:text-green-500'
  }
])

async function copyLink() {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    // Could add toast notification here
  } catch {
    // Fallback for older browsers
  }
}
</script>

<template>
  <div class="flex items-center gap-2">
    <span class="text-sm text-gray-500 dark:text-gray-400 mr-2">Share:</span>
    <a
      v-for="link in shareLinks"
      :key="link.name"
      :href="link.url"
      target="_blank"
      rel="noopener noreferrer"
      :title="`Share on ${link.name}`"
      class="p-2 rounded-full text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
      :class="link.color"
    >
      <UIcon :name="link.icon" class="w-5 h-5" />
    </a>
    <button
      title="Copy link"
      class="p-2 rounded-full text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-primary transition-colors"
      @click="copyLink"
    >
      <UIcon name="i-heroicons-link" class="w-5 h-5" />
    </button>
  </div>
</template>
