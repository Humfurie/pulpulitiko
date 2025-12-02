<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const auth = useAuth()
const router = useRouter()

const form = reactive({
  name: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const loading = ref(false)
const error = ref('')
const showPassword = ref(false)
const showConfirmPassword = ref(false)

// Password validation
const passwordValidation = computed(() => {
  const password = form.password
  return {
    minLength: password.length >= 8,
    hasUppercase: /[A-Z]/.test(password),
    hasLowercase: /[a-z]/.test(password),
    hasNumber: /[0-9]/.test(password)
  }
})

const isPasswordValid = computed(() => {
  return passwordValidation.value.minLength &&
    passwordValidation.value.hasUppercase &&
    passwordValidation.value.hasLowercase &&
    passwordValidation.value.hasNumber
})

const passwordsMatch = computed(() => {
  return form.password === form.confirmPassword && form.confirmPassword.length > 0
})

const canSubmit = computed(() => {
  return form.name.trim().length >= 2 &&
    form.email.includes('@') &&
    isPasswordValid.value &&
    passwordsMatch.value
})

async function handleSubmit() {
  if (!canSubmit.value) return

  loading.value = true
  error.value = ''

  const result = await auth.register(form.name, form.email, form.password)

  if (result.success) {
    await router.push('/')
  } else {
    error.value = result.error || 'Registration failed'
  }

  loading.value = false
}

// Redirect if already logged in
onMounted(async () => {
  if (auth.token.value) {
    await auth.fetchCurrentUser()
    if (auth.isAuthenticated.value) {
      await router.push('/')
    }
  }
})

useSeoMeta({
  title: 'Create Account - Pulpulitiko'
})
</script>

<template>
  <div class="min-h-[80vh] flex items-center justify-center py-12 px-4">
    <div class="w-full max-w-md">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 mb-4">
          <UIcon name="i-heroicons-user-plus" class="w-8 h-8 text-primary" />
        </div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Create Account</h1>
        <p class="mt-2 text-gray-600 dark:text-gray-400">
          Join the community and share your thoughts
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

          <!-- Name Field -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Full Name
            </label>
            <UInput
              v-model="form.name"
              type="text"
              placeholder="Juan Dela Cruz"
              required
              icon="i-heroicons-user"
              size="xl"
              class="w-full"
            />
          </div>

          <!-- Email Field -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Email Address
            </label>
            <UInput
              v-model="form.email"
              type="email"
              placeholder="juan@example.com"
              required
              icon="i-heroicons-envelope"
              size="xl"
              class="w-full"
            />
          </div>

          <!-- Password Field -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Password
            </label>
            <UInput
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="Create a strong password"
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

            <!-- Password Requirements -->
            <div v-if="form.password" class="mt-3 space-y-1.5">
              <div class="flex items-center gap-2 text-sm">
                <UIcon
                  :name="passwordValidation.minLength ? 'i-heroicons-check-circle' : 'i-heroicons-x-circle'"
                  :class="passwordValidation.minLength ? 'text-green-500' : 'text-gray-400'"
                  class="w-4 h-4"
                />
                <span :class="passwordValidation.minLength ? 'text-green-600 dark:text-green-400' : 'text-gray-500'">
                  At least 8 characters
                </span>
              </div>
              <div class="flex items-center gap-2 text-sm">
                <UIcon
                  :name="passwordValidation.hasUppercase ? 'i-heroicons-check-circle' : 'i-heroicons-x-circle'"
                  :class="passwordValidation.hasUppercase ? 'text-green-500' : 'text-gray-400'"
                  class="w-4 h-4"
                />
                <span :class="passwordValidation.hasUppercase ? 'text-green-600 dark:text-green-400' : 'text-gray-500'">
                  One uppercase letter
                </span>
              </div>
              <div class="flex items-center gap-2 text-sm">
                <UIcon
                  :name="passwordValidation.hasLowercase ? 'i-heroicons-check-circle' : 'i-heroicons-x-circle'"
                  :class="passwordValidation.hasLowercase ? 'text-green-500' : 'text-gray-400'"
                  class="w-4 h-4"
                />
                <span :class="passwordValidation.hasLowercase ? 'text-green-600 dark:text-green-400' : 'text-gray-500'">
                  One lowercase letter
                </span>
              </div>
              <div class="flex items-center gap-2 text-sm">
                <UIcon
                  :name="passwordValidation.hasNumber ? 'i-heroicons-check-circle' : 'i-heroicons-x-circle'"
                  :class="passwordValidation.hasNumber ? 'text-green-500' : 'text-gray-400'"
                  class="w-4 h-4"
                />
                <span :class="passwordValidation.hasNumber ? 'text-green-600 dark:text-green-400' : 'text-gray-500'">
                  One number
                </span>
              </div>
            </div>
          </div>

          <!-- Confirm Password Field -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Confirm Password
            </label>
            <UInput
              v-model="form.confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              placeholder="Confirm your password"
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

            <!-- Password Match Indicator -->
            <div v-if="form.confirmPassword" class="mt-2 flex items-center gap-2 text-sm">
              <UIcon
                :name="passwordsMatch ? 'i-heroicons-check-circle' : 'i-heroicons-x-circle'"
                :class="passwordsMatch ? 'text-green-500' : 'text-red-500'"
                class="w-4 h-4"
              />
              <span :class="passwordsMatch ? 'text-green-600 dark:text-green-400' : 'text-red-500'">
                {{ passwordsMatch ? 'Passwords match' : 'Passwords do not match' }}
              </span>
            </div>
          </div>

          <!-- Terms Notice -->
          <p class="text-xs text-gray-500 dark:text-gray-400">
            By creating an account, you agree to our community guidelines and terms of service.
          </p>

          <!-- Submit Button -->
          <UButton
            type="submit"
            block
            size="lg"
            :loading="loading"
            :disabled="!canSubmit || loading"
          >
            Create Account
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

        <!-- Login Link -->
        <p class="text-center text-gray-600 dark:text-gray-400">
          Already have an account?
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
