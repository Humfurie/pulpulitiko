export default defineNuxtPlugin(async () => {
  const auth = useAuth()

  // If there's a token but no user, fetch the user data
  if (auth.token.value && !auth.user.value) {
    await auth.fetchCurrentUser()
  }
})
