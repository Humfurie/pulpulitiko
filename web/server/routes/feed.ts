export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  // Use internal URL for server-side requests, removing /api suffix
  const apiBaseUrl = config.apiInternalUrl.replace('/api', '')

  const response = await $fetch.raw(`${apiBaseUrl}/feed`)

  // Set correct content type for RSS
  setHeader(event, 'Content-Type', 'application/rss+xml; charset=utf-8')
  setHeader(event, 'Cache-Control', 'public, max-age=900') // 15 minutes cache

  return response._data
})
