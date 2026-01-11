export default defineEventHandler((event) => {
  const config = useRuntimeConfig()
  const siteUrl = config.public.siteUrl

  const robotsTxt = `# Pulpulitiko robots.txt
User-agent: *
Allow: /

# Sitemap location
Sitemap: ${siteUrl}/sitemap.xml

# Disallow admin and authentication paths
Disallow: /admin/
Disallow: /api/
Disallow: /account/
Disallow: /login/
Disallow: /register/
`

  setHeader(event, 'Content-Type', 'text/plain; charset=utf-8')
  setHeader(event, 'Cache-Control', 'public, max-age=86400') // 24 hours

  return robotsTxt
})
