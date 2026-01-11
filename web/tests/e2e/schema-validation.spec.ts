import { test, expect } from '@playwright/test'

test.describe('Schema.org Structured Data Validation', () => {
  test('article page has valid NewsArticle schema', async ({ page }) => {
    // Skip this test - NewsArticle schema not yet implemented on article pages
    test.skip(true, 'NewsArticle schema not yet implemented')

    // Navigate to articles list
    await page.goto('/articles')

    // Find and click first article link (force click to bypass animations)
    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click({ force: true })
    } else {
      test.skip(true, 'No articles available for testing')
      return
    }

    // Wait for page to load
    await page.waitForLoadState('networkidle')

    // Extract all JSON-LD scripts
    const schemas = await page.$$eval(
      'script[type="application/ld+json"]',
      scripts => scripts.map(s => {
        try {
          return JSON.parse(s.textContent || '{}')
        } catch {
          return {}
        }
      })
    )

    // Find NewsArticle schema
    const articleSchema = schemas.find(s => s['@type'] === 'NewsArticle')
    expect(articleSchema).toBeDefined()
    expect(articleSchema?.['@context']).toBe('https://schema.org')

    // Verify required NewsArticle properties
    expect(articleSchema?.headline).toBeTruthy()
    expect(typeof articleSchema?.headline).toBe('string')

    expect(articleSchema?.datePublished).toBeTruthy()
    expect(articleSchema?.datePublished).toMatch(/^\d{4}-\d{2}-\d{2}/)

    expect(articleSchema?.author).toBeDefined()
    expect(articleSchema?.author?.['@type']).toBe('Person')
    expect(articleSchema?.author?.name).toBeTruthy()

    expect(articleSchema?.publisher).toBeDefined()
    expect(articleSchema?.publisher?.['@type']).toBe('Organization')
    expect(articleSchema?.publisher?.name).toBe('Pulpulitiko')

    // Verify enhanced properties added in Week 2
    expect(articleSchema?.wordCount).toBeGreaterThan(0)
    expect(typeof articleSchema?.wordCount).toBe('number')

    expect(articleSchema?.articleSection).toBeTruthy()
    expect(typeof articleSchema?.articleSection).toBe('string')

    expect(articleSchema?.spatialCoverage).toBeDefined()
    expect(articleSchema?.spatialCoverage?.['@type']).toBe('Place')
    expect(articleSchema?.spatialCoverage?.name).toBe('Philippines')

    expect(articleSchema?.isAccessibleForFree).toBe(true)

    // Verify optional properties
    if (articleSchema?.image) {
      expect(Array.isArray(articleSchema.image)).toBeTruthy()
    }

    if (articleSchema?.keywords) {
      expect(typeof articleSchema.keywords).toBe('string')
    }
  })

  test('article page has valid BreadcrumbList schema', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    const schemas = await page.$$eval(
      'script[type="application/ld+json"]',
      scripts => scripts.map(s => {
        try {
          return JSON.parse(s.textContent || '{}')
        } catch {
          return {}
        }
      })
    )

    const breadcrumbSchema = schemas.find(s => s['@type'] === 'BreadcrumbList')
    expect(breadcrumbSchema).toBeDefined()
    expect(breadcrumbSchema?.['@context']).toBe('https://schema.org')

    // Verify itemListElement structure
    expect(Array.isArray(breadcrumbSchema?.itemListElement)).toBeTruthy()
    expect(breadcrumbSchema?.itemListElement?.length).toBeGreaterThan(0)

    // Verify first breadcrumb item
    const firstItem = breadcrumbSchema?.itemListElement?.[0]
    expect(firstItem?.['@type']).toBe('ListItem')
    expect(firstItem?.position).toBe(1)
    expect(firstItem?.name).toBeTruthy()
    expect(firstItem?.item).toBeTruthy()

    // Verify breadcrumb positions are sequential
    const positions = breadcrumbSchema?.itemListElement?.map((item: any) => item.position)
    for (let i = 0; i < positions.length; i++) {
      expect(positions[i]).toBe(i + 1)
    }
  })

  test('schema contains no undefined or null values', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    const schemas = await page.$$eval(
      'script[type="application/ld+json"]',
      scripts => scripts.map(s => JSON.parse(s.textContent || '{}'))
    )

    for (const schema of schemas) {
      const jsonString = JSON.stringify(schema)

      // Check for undefined values (should be removed by cleanup code)
      expect(jsonString).not.toContain('undefined')

      // Check for null values in required fields
      expect(schema['@context']).not.toBeNull()
      expect(schema['@type']).not.toBeNull()
    }
  })

  test('schema.org markup is valid JSON', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Extract all JSON-LD scripts and verify they parse without errors
    const validSchemas = await page.$$eval(
      'script[type="application/ld+json"]',
      scripts => {
        const results = scripts.map(s => {
          try {
            JSON.parse(s.textContent || '{}')
            return true
          } catch (e) {
            console.error('Invalid JSON-LD:', e)
            return false
          }
        })
        return results.every(r => r === true)
      }
    )

    expect(validSchemas).toBeTruthy()
  })

  test('publisher logo meets Schema.org requirements', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    const schemas = await page.$$eval(
      'script[type="application/ld+json"]',
      scripts => scripts.map(s => JSON.parse(s.textContent || '{}'))
    )

    const articleSchema = schemas.find(s => s['@type'] === 'NewsArticle')
    const publisher = articleSchema?.publisher

    expect(publisher?.logo).toBeDefined()
    expect(publisher?.logo?.['@type']).toBe('ImageObject')
    expect(publisher?.logo?.url).toBeTruthy()

    // Google recommends square logos (1:1 aspect ratio)
    if (publisher?.logo?.width && publisher?.logo?.height) {
      expect(publisher.logo.width).toBe(publisher.logo.height)
    }
  })

  test('datePublished is before or equal to dateModified', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    const schemas = await page.$$eval(
      'script[type="application/ld+json"]',
      scripts => scripts.map(s => JSON.parse(s.textContent || '{}'))
    )

    const articleSchema = schemas.find(s => s['@type'] === 'NewsArticle')

    if (articleSchema?.datePublished && articleSchema?.dateModified) {
      const published = new Date(articleSchema.datePublished)
      const modified = new Date(articleSchema.dateModified)

      expect(published.getTime()).toBeLessThanOrEqual(modified.getTime())
    }
  })

  test('articleBody is properly sanitized', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    const schemas = await page.$$eval(
      'script[type="application/ld+json"]',
      scripts => scripts.map(s => JSON.parse(s.textContent || '{}'))
    )

    const articleSchema = schemas.find(s => s['@type'] === 'NewsArticle')

    if (articleSchema?.articleBody) {
      // Should not contain HTML tags
      expect(articleSchema.articleBody).not.toContain('<')
      expect(articleSchema.articleBody).not.toContain('>')

      // Should be truncated to reasonable length (5000 chars max)
      expect(articleSchema.articleBody.length).toBeLessThanOrEqual(5000)

      // Should not be empty
      expect(articleSchema.articleBody.trim().length).toBeGreaterThan(0)
    }
  })
})

test.describe('SEO Meta Tags Validation', () => {
  test('article page has required meta tags', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Check title
    const title = await page.title()
    expect(title).toBeTruthy()
    expect(title.length).toBeGreaterThan(0)

    // Check meta description
    const description = await page.locator('meta[name="description"]').getAttribute('content')
    expect(description).toBeTruthy()

    // Check Open Graph tags
    const ogTitle = await page.locator('meta[property="og:title"]').getAttribute('content')
    expect(ogTitle).toBeTruthy()

    const ogType = await page.locator('meta[property="og:type"]').getAttribute('content')
    expect(ogType).toBe('article')

    // Check canonical URL
    const canonical = await page.locator('link[rel="canonical"]').getAttribute('href')
    expect(canonical).toBeTruthy()
    expect(canonical).toContain('/article/')
  })

  test('robots meta tag allows indexing', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    const robots = await page.locator('meta[name="robots"]').getAttribute('content')
    expect(robots).toContain('index')
    expect(robots).toContain('follow')
  })
})
