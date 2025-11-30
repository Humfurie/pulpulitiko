<script setup lang="ts">
onMounted(() => {
  // Mark HTML as ready to show content
  document.documentElement.classList.add('ready')

  // Remove the loader with fade animation
  const loader = document.getElementById('__nuxt_loader')
  if (loader) {
    loader.classList.add('fade-out')
    setTimeout(() => {
      loader.remove()
    }, 300)
  }
})
</script>

<template>
  <NuxtLayout>
    <NuxtRouteAnnouncer />
    <NuxtPage />
  </NuxtLayout>
</template>
