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

  // Regular users (role: "user") cannot access admin panel at all
  // Redirect them to their account page
  if (userRole === 'user') {
    return navigateTo('/account')
  }

  // Define route permissions for admin/author roles
  const adminOnlyRoutes = ['/admin/users', '/admin/roles']

  // Check if route requires admin access
  const isAdminRoute = adminOnlyRoutes.some(route => to.path.startsWith(route))
  if (isAdminRoute && userRole !== 'admin') {
    return navigateTo('/admin')
  }
})
