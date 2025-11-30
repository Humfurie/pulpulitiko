<script setup lang="ts">
definePageMeta({
  layout: 'admin'
})

const auth = useAuth()

useSeoMeta({
  title: 'Dashboard - Pulpulitiko Admin'
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">Dashboard</h1>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <UCard>
        <div class="flex items-center gap-4">
          <div class="p-3 bg-primary/10 rounded-lg">
            <span class="i-heroicons-document-text size-6 text-primary" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Articles</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">-</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center gap-4">
          <div class="p-3 bg-blue-500/10 rounded-lg">
            <span class="i-heroicons-folder size-6 text-blue-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Categories</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">-</p>
          </div>
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center gap-4">
          <div class="p-3 bg-purple-500/10 rounded-lg">
            <span class="i-heroicons-tag size-6 text-purple-500" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Tags</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">-</p>
          </div>
        </div>
      </UCard>
    </div>

    <UCard class="mt-6">
      <template #header>
        <h2 class="font-semibold text-gray-900 dark:text-white">Quick Actions</h2>
      </template>

      <div class="flex gap-4">
        <UButton to="/admin/articles/new" icon="i-heroicons-plus">
          New Article
        </UButton>
        <UButton to="/admin/articles" variant="outline" icon="i-heroicons-document-text">
          Manage Articles
        </UButton>
      </div>
    </UCard>

    <UCard class="mt-6">
      <template #header>
        <h2 class="font-semibold text-gray-900 dark:text-white">Welcome</h2>
      </template>
      <p class="text-gray-600 dark:text-gray-400">
        Logged in as <strong>{{ auth.user.value?.name }}</strong> ({{ auth.user.value?.role }})
      </p>
    </UCard>
  </div>
</template>
