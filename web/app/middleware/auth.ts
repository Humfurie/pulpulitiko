export default defineNuxtRouteMiddleware(async () => {
  const auth = useAuth()

  // Wait for auth to be ready
  if (!auth.token.value) {
    return navigateTo('/login')
  }

  // Fetch user if not already loaded
  if (!auth.user.value) {
    await auth.fetchCurrentUser()
  }

  if (!auth.isAuthenticated.value) {
    return navigateTo('/login')
  }
})
