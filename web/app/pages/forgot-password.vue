<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const auth = useAuth()

const form = reactive({
  email: ''
})

const loading = ref(false)
const error = ref('')
const success = ref(false)

async function handleSubmit() {
  loading.value = true
  error.value = ''

  const result = await auth.forgotPassword(form.email)

  if (result.success) {
    success.value = true
  } else {
    error.value = result.error || 'Failed to send reset email'
  }

  loading.value = false
}

useSeoMeta({
  title: 'Forgot Password - Pulpulitiko'
})
</script>

<template>
  <div class="min-h-[80vh] flex items-center justify-center py-12 px-4">
    <div class="w-full max-w-md">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 mb-4">
          <UIcon name="i-heroicons-key" class="w-8 h-8 text-primary" />
        </div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Forgot Password?</h1>
        <p class="mt-2 text-gray-600 dark:text-gray-400">
          Enter your email and we'll send you a reset link
        </p>
      </div>

      <!-- Form Card -->
      <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-xl border border-gray-200 dark:border-gray-800 p-8">
        <!-- Success State -->
        <div v-if="success" class="text-center py-4">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/20 mb-4">
            <UIcon name="i-heroicons-envelope-open" class="w-8 h-8 text-green-600 dark:text-green-400" />
          </div>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">Check your email</h2>
          <p class="text-gray-600 dark:text-gray-400 mb-6">
            If an account exists with {{ form.email }}, you will receive a password reset link shortly.
          </p>
          <NuxtLink to="/login">
            <UButton variant="soft" size="lg">
              Return to Sign In
            </UButton>
          </NuxtLink>
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

          <!-- Submit Button -->
          <UButton
            type="submit"
            block
            size="lg"
            :loading="loading"
            :disabled="loading"
          >
            Send Reset Link
          </UButton>
        </form>

        <!-- Divider -->
        <div v-if="!success" class="relative my-6">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-gray-200 dark:border-gray-700" />
          </div>
          <div class="relative flex justify-center text-sm">
            <span class="px-4 bg-white dark:bg-gray-900 text-gray-500">or</span>
          </div>
        </div>

        <!-- Back to Login Link -->
        <p v-if="!success" class="text-center text-gray-600 dark:text-gray-400">
          Remember your password?
          <NuxtLink to="/login" class="text-primary font-medium hover:underline">
            Sign in
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
