import { test, expect } from '@playwright/test'

test.describe('Dynamic robots.txt', () => {
  test('returns correct content type and headers', async ({ page }) => {
    const response = await page.goto('/robots.txt')

    expect(response?.status()).toBe(200)
    expect(response?.headers()['content-type']).toContain('text/plain')

    // Cache-Control header check (case-insensitive)
    const headers = response?.headers() || {}
    const cacheControl = headers['cache-control'] || headers['Cache-Control'] || ''
    expect(cacheControl).toContain('public')
    expect(cacheControl).toContain('max-age')
  })

  test('uses correct site URL for sitemap', async ({ page }) => {
    const response = await page.goto('/robots.txt')
    expect(response?.status()).toBe(200)

    const text = await response?.text()

    // Should contain Sitemap directive
    expect(text).toContain('Sitemap:')

    // Should use the configured site URL (not hardcoded)
    // The URL should match the environment's NUXT_PUBLIC_SITE_URL
    expect(text).toMatch(/Sitemap: https?:\/\/.*\/sitemap\.xml/)

    // Verify it contains the domain (will vary by environment)
    expect(text).toContain('pulpulitiko')
  })

  test('includes all disallow rules', async ({ page }) => {
    const response = await page.goto('/robots.txt')
    const text = await response?.text()

    // Verify all protected paths are disallowed
    expect(text).toContain('Disallow: /admin/')
    expect(text).toContain('Disallow: /api/')
    expect(text).toContain('Disallow: /account/')
    expect(text).toContain('Disallow: /login/')
    expect(text).toContain('Disallow: /register/')
  })

  test('allows all user agents', async ({ page }) => {
    const response = await page.goto('/robots.txt')
    const text = await response?.text()

    expect(text).toContain('User-agent: *')
    expect(text).toContain('Allow: /')
  })

  test('has valid robots.txt format', async ({ page }) => {
    const response = await page.goto('/robots.txt')
    const text = await response?.text()

    // Should contain User-agent at or near the start (may have comment before it)
    expect(text?.trim()).toMatch(/^(#.*\n)?User-agent:/s)

    // Should have newlines between directives
    expect(text).toMatch(/User-agent:.*\n.*Allow:/)
    expect(text).toMatch(/Sitemap:/)
  })
})
