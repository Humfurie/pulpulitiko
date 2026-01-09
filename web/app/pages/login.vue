<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const auth = useAuth()
const router = useRouter()
const route = useRoute()

const form = reactive({
  email: '',
  password: ''
})

const loading = ref(false)
const error = ref('')
const showPassword = ref(false)

// Get redirect URL from query params
const redirectTo = computed(() => {
  const redirect = route.query.redirect as string
  return redirect || null
})

async function handleSubmit() {
  loading.value = true
  error.value = ''

  const result = await auth.login(form.email, form.password)

  if (result.success) {
    // Redirect based on role or query param
    if (redirectTo.value) {
      await router.push(redirectTo.value)
    } else if (auth.user.value?.role === 'admin' || auth.user.value?.role === 'author') {
      await router.push('/admin')
    } else {
      await router.push('/')
    }
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
      if (redirectTo.value) {
        await router.push(redirectTo.value)
      } else if (auth.user.value?.role === 'admin' || auth.user.value?.role === 'author') {
        await router.push('/admin')
      } else {
        await router.push('/')
      }
    }
  }
})

useSeoMeta({
  title: 'Sign In - Pulpulitiko'
})
</script>

<template>
  <div class="min-h-[80vh] flex items-center justify-center py-12 px-4">
    <div class="w-full max-w-md">
      <!-- Header -->
      <div class="text-center mb-8">
        <NuxtLink to="/" class="inline-block mb-4">
          <img
            src="/pulpulitiko.png"
            alt="Pulpulitiko"
            class="h-10 w-auto mx-auto dark:hidden"
          >
          <img
            src="/pulpulitiko_dark.png"
            alt="Pulpulitiko"
            class="h-10 w-auto mx-auto hidden dark:block"
          >
        </NuxtLink>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Welcome Back</h1>
        <p class="mt-2 text-gray-600 dark:text-gray-400">
          Sign in to your account
        </p>
      </div>

      <!-- Form Card -->
      <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-xl border border-gray-200 dark:border-gray-800 p-8">
        <form class="space-y-5" @submit.prevent="handleSubmit">
          <!-- Error Alert -->
          <UAlert
            v-if="error"
            color="error"
            icon="i-heroicons-exclamation-circle"
            :title="error"
            class="mb-4"
          />

          <!-- Email Field -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Email Address
            </label>
            <UInput
              v-model="form.email"
              type="email"
              placeholder="you@example.com"
              required
              icon="i-heroicons-envelope"
              size="xl"
              class="w-full"
            />
          </div>

          <!-- Password Field -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Password
              </label>
              <NuxtLink to="/forgot-password" class="text-sm text-primary hover:underline">
                Forgot password?
              </NuxtLink>
            </div>
            <UInput
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="Enter your password"
              required
              icon="i-heroicons-lock-closed"
              size="xl"
              class="w-full"
            >
              <template #trailing>
                <button
                  type="button"
                  class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                  @click="showPassword = !showPassword"
                >
                  <UIcon :name="showPassword ? 'i-heroicons-eye-slash' : 'i-heroicons-eye'" class="w-5 h-5" />
                </button>
              </template>
            </UInput>
          </div>

          <!-- Submit Button -->
          <UButton
            type="submit"
            block
            size="lg"
            :loading="loading"
            :disabled="loading"
          >
            Sign In
          </UButton>
        </form>

        <!-- Divider -->
        <div class="relative my-6">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-gray-200 dark:border-gray-700" />
          </div>
          <div class="relative flex justify-center text-sm">
            <span class="px-4 bg-white dark:bg-gray-900 text-gray-500">or</span>
          </div>
        </div>

        <!-- Register Link -->
        <p class="text-center text-gray-600 dark:text-gray-400">
          Don't have an account?
          <NuxtLink to="/register" class="text-primary font-medium hover:underline">
            Create one
          </NuxtLink>
        </p>
      </div>

      <!-- Back to Home -->
      <p class="text-center text-sm text-gray-500 dark:text-gray-400 mt-6">
        <NuxtLink to="/" class="hover:text-primary hover:underline">
          Back to homepage
        </NuxtLink>
      </p>
    </div>
  </div>
</template>
