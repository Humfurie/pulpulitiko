export default defineNuxtRouteMiddleware(async (to) => {
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

  const userRole = auth.user.value?.role

  // Define route permissions
  const adminOnlyRoutes = ['/admin/users', '/admin/roles']
  const authorRoutes = ['/admin/articles', '/admin/categories', '/admin/tags']

  // Check if route requires admin access
  const isAdminRoute = adminOnlyRoutes.some(route => to.path.startsWith(route))
  if (isAdminRoute && userRole !== 'admin') {
    return navigateTo('/admin')
  }

  // Check if route requires at least author access
  const isAuthorRoute = authorRoutes.some(route => to.path.startsWith(route))
  if (isAuthorRoute && !['admin', 'author'].includes(userRole || '')) {
    return navigateTo('/admin')
  }
})
