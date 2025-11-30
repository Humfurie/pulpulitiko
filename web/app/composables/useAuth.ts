import type { User, LoginResponse, ApiResponse } from '~/types'

export function useAuth() {
  const config = useRuntimeConfig()
  const baseUrl = import.meta.server
    ? config.apiInternalUrl
    : config.public.apiUrl

  const user = useState<User | null>('auth-user', () => null)
  const token = useCookie('auth-token', {
    maxAge: 60 * 60 * 24 * 7, // 7 days
    sameSite: 'lax'
  })

  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(email: string, password: string): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await $fetch<ApiResponse<LoginResponse>>(`${baseUrl}/auth/login`, {
        method: 'POST',
        body: { email, password }
      })

      if (response.success) {
        token.value = response.data.token
        user.value = response.data.user
        return { success: true }
      }

      return { success: false, error: 'Login failed' }
    } catch (error: any) {
      return { success: false, error: error?.data?.error?.message || 'Invalid credentials' }
    }
  }

  async function logout() {
    token.value = null
    user.value = null
    await navigateTo('/login')
  }

  async function fetchCurrentUser() {
    if (!token.value) return

    try {
      const response = await $fetch<ApiResponse<User>>(`${baseUrl}/auth/me`, {
        headers: {
          Authorization: `Bearer ${token.value}`
        }
      })

      if (response.success) {
        user.value = response.data
      }
    } catch {
      // Token invalid, clear it
      token.value = null
      user.value = null
    }
  }

  function getAuthHeaders(): Record<string, string> {
    if (!token.value) return {}
    return { Authorization: `Bearer ${token.value}` }
  }

  return {
    user: readonly(user),
    token: readonly(token),
    isAuthenticated,
    isAdmin,
    login,
    logout,
    fetchCurrentUser,
    getAuthHeaders
  }
}
