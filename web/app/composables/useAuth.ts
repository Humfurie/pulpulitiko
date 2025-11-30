import type { User, LoginResponse, ApiResponse } from '~/types'

export function useAuth() {
  const config = useRuntimeConfig()
  const baseUrl = import.meta.server
    ? config.apiInternalUrl
    : config.public.apiUrl

  const user = useState<User | null>('auth-user', () => null)
  const permissions = useState<string[]>('auth-permissions', () => [])
  const token = useCookie('auth-token', {
    maxAge: 60 * 60 * 24 * 7, // 7 days
    sameSite: 'lax'
  })

  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isAuthor = computed(() => user.value?.role === 'author')

  function hasPermission(permission: string): boolean {
    return permissions.value.includes(permission)
  }

  function hasAnyPermission(...perms: string[]): boolean {
    return perms.some(p => permissions.value.includes(p))
  }

  async function login(email: string, password: string): Promise<{ success: boolean; error?: string }> {
    try {
      const response = await $fetch<ApiResponse<LoginResponse>>(`${baseUrl}/auth/login`, {
        method: 'POST',
        body: { email, password }
      })

      if (response.success) {
        token.value = response.data.token
        user.value = response.data.user
        permissions.value = response.data.permissions || []
        return { success: true }
      }

      return { success: false, error: 'Login failed' }
    } catch (e: unknown) {
      const err = e as { data?: { error?: { message?: string } } }
      return { success: false, error: err?.data?.error?.message || 'Invalid credentials' }
    }
  }

  async function logout() {
    token.value = null
    user.value = null
    permissions.value = []
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
    permissions: readonly(permissions),
    isAuthenticated,
    isAdmin,
    isAuthor,
    login,
    logout,
    fetchCurrentUser,
    getAuthHeaders,
    hasPermission,
    hasAnyPermission
  }
}
