<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const auth = useAuth()
const router = useRouter()

const form = reactive({
  email: '',
  password: ''
})

const loading = ref(false)
const error = ref('')

async function handleSubmit() {
  loading.value = true
  error.value = ''

  const result = await auth.login(form.email, form.password)

  if (result.success) {
    await router.push('/admin')
  } else {
    error.value = result.error || 'Login failed'
  }

  loading.value = false
}

// Redirect if already logged in
onMounted(async () => {
  if (auth.token.value) {
    await auth.fetchCurrentUser()
    if (auth.isAuthenticated.value) {
      await router.push('/admin')
    }
  }
})

useSeoMeta({
  title: 'Login - Pulpulitiko Admin'
})
</script>

<template>
  <div class="min-h-[80vh] flex items-center justify-center">
    <UCard class="w-full max-w-md">
      <template #header>
        <div class="text-center">
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Admin Login</h1>
          <p class="mt-2 text-gray-600 dark:text-gray-400">Sign in to manage articles</p>
        </div>
      </template>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <UAlert
          v-if="error"
          color="error"
          icon="i-heroicons-exclamation-circle"
          :title="error"
          class="mb-4"
        />

        <UFormField label="Email" name="email">
          <UInput
            v-model="form.email"
            type="email"
            placeholder="admin@pulpulitiko.ph"
            required
            icon="i-heroicons-envelope"
          />
        </UFormField>

        <UFormField label="Password" name="password">
          <UInput
            v-model="form.password"
            type="password"
            placeholder="Enter your password"
            required
            icon="i-heroicons-lock-closed"
          />
        </UFormField>

        <UButton
          type="submit"
          block
          :loading="loading"
          :disabled="loading"
        >
          Sign In
        </UButton>
      </form>

      <template #footer>
        <p class="text-center text-sm text-gray-500 dark:text-gray-400">
          <NuxtLink to="/" class="text-primary hover:underline">
            Back to homepage
          </NuxtLink>
        </p>
      </template>
    </UCard>
  </div>
</template>
