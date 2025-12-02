<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const auth = useAuth()
const route = useRoute()
const router = useRouter()

const token = computed(() => route.query.token as string)

const form = reactive({
  password: '',
  confirmPassword: ''
})

const loading = ref(false)
const error = ref('')
const success = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)

// Check if token is present
const hasToken = computed(() => !!token.value)

async function handleSubmit() {
  // Validate passwords match
  if (form.password !== form.confirmPassword) {
    error.value = 'Passwords do not match'
    return
  }

  // Validate password length
  if (form.password.length < 8) {
    error.value = 'Password must be at least 8 characters'
    return
  }

  loading.value = true
  error.value = ''

  const result = await auth.resetPassword(token.value, form.password)

  if (result.success) {
    success.value = true
  } else {
    error.value = result.error || 'Failed to reset password'
  }

  loading.value = false
}

function goToLogin() {
  router.push('/login')
}

useSeoMeta({
  title: 'Reset Password - Pulpulitiko'
})
</script>

<template>
  <div class="min-h-[80vh] flex items-center justify-center py-12 px-4">
    <div class="w-full max-w-md">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 mb-4">
          <UIcon name="i-heroicons-lock-closed" class="w-8 h-8 text-primary" />
        </div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Reset Password</h1>
        <p class="mt-2 text-gray-600 dark:text-gray-400">
          Enter your new password
        </p>
      </div>

      <!-- Form Card -->
      <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-xl border border-gray-200 dark:border-gray-800 p-8">
        <!-- No Token State -->
        <div v-if="!hasToken" class="text-center py-4">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-red-100 dark:bg-red-900/20 mb-4">
            <UIcon name="i-heroicons-exclamation-triangle" class="w-8 h-8 text-red-600 dark:text-red-400" />
          </div>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">Invalid Link</h2>
          <p class="text-gray-600 dark:text-gray-400 mb-6">
            This password reset link is invalid or has expired. Please request a new one.
          </p>
          <NuxtLink to="/forgot-password">
            <UButton variant="soft" size="lg">
              Request New Link
            </UButton>
          </NuxtLink>
        </div>

        <!-- Success State -->
        <div v-else-if="success" class="text-center py-4">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/20 mb-4">
            <UIcon name="i-heroicons-check-circle" class="w-8 h-8 text-green-600 dark:text-green-400" />
          </div>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">Password Reset!</h2>
          <p class="text-gray-600 dark:text-gray-400 mb-6">
            Your password has been successfully reset. You can now sign in with your new password.
          </p>
          <UButton size="lg" @click="goToLogin">
            Sign In
          </UButton>
        </div>

        <!-- Form State -->
        <form v-else class="space-y-5" @submit.prevent="handleSubmit">
          <!-- Error Alert -->
          <UAlert
            v-if="error"
            color="error"
            icon="i-heroicons-exclamation-circle"
            :title="error"
            class="mb-4"
          />

          <!-- New Password Field -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              New Password
            </label>
            <UInput
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="Enter new password"
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
            <p class="mt-1 text-xs text-gray-500">Minimum 8 characters</p>
          </div>

          <!-- Confirm Password Field -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Confirm Password
            </label>
            <UInput
              v-model="form.confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              placeholder="Confirm new password"
              required
              icon="i-heroicons-lock-closed"
              size="xl"
              class="w-full"
            >
              <template #trailing>
                <button
                  type="button"
                  class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                  @click="showConfirmPassword = !showConfirmPassword"
                >
                  <UIcon :name="showConfirmPassword ? 'i-heroicons-eye-slash' : 'i-heroicons-eye'" class="w-5 h-5" />
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
            Reset Password
          </UButton>
        </form>
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
