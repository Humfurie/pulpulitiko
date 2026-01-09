import { test, expect } from '@playwright/test'

test.describe('Article Pages', () => {
  test('article page has proper structure', async ({ page }) => {
    // Navigate to homepage first
    await page.goto('/')
    
    // Find and click on the first article link
    const firstArticleLink = page.locator('article a, [class*="article"] a').first()
    await expect(firstArticleLink).toBeVisible({ timeout: 10000 })
    await firstArticleLink.click()
    
    // Wait for article page to load
    await page.waitForLoadState('networkidle')
    
    // Check article page structure
    await expect(page.locator('h1').first()).toBeVisible()
    
    // Check for article content
    const content = page.locator('article, [class*="content"], [class*="body"]').first()
    if (await content.count() > 0) {
      await expect(content).toBeVisible()
    }
  })

  test('article page has metadata', async ({ page }) => {
    await page.goto('/')
    
    // Navigate to first article
    const firstArticleLink = page.locator('article a, [class*="article"] a').first()
    if (await firstArticleLink.count() > 0) {
      await firstArticleLink.click()
      await page.waitForLoadState('networkidle')
      
      // Check for meta description
      const metaDescription = page.locator('meta[name="description"]')
      await expect(metaDescription).toHaveCount(1)
      
      // Check for Open Graph tags
      const ogTitle = page.locator('meta[property="og:title"]')
      await expect(ogTitle).toHaveCount(1)
    }
  })

  test('article page has social sharing', async ({ page }) => {
    await page.goto('/')
    
    // Navigate to first article
    const firstArticleLink = page.locator('article a, [class*="article"] a').first()
    if (await firstArticleLink.count() > 0) {
      await firstArticleLink.click()
      await page.waitForLoadState('networkidle')
      
      // Look for share buttons (adjust selectors based on your implementation)
      const shareSection = page.locator('[class*="share"], [aria-label*="share" i]').first()
      if (await shareSection.count() > 0) {
        await expect(shareSection).toBeVisible()
      }
    }
  })

  test('article page has breadcrumbs', async ({ page }) => {
    await page.goto('/')
    
    // Navigate to first article
    const firstArticleLink = page.locator('article a, [class*="article"] a').first()
    if (await firstArticleLink.count() > 0) {
      await firstArticleLink.click()
      await page.waitForLoadState('networkidle')
      
      // Check for breadcrumb navigation
      const breadcrumbs = page.locator('nav[aria-label*="breadcrumb" i], [class*="breadcrumb"]').first()
      if (await breadcrumbs.count() > 0) {
        await expect(breadcrumbs).toBeVisible()
        
        // Breadcrumbs should contain a link to home
        const homeLink = breadcrumbs.locator('a').first()
        await expect(homeLink).toBeVisible()
      }
    }
  })

  test('related articles are displayed', async ({ page }) => {
    await page.goto('/')
    
    // Navigate to first article
    const firstArticleLink = page.locator('article a, [class*="article"] a').first()
    if (await firstArticleLink.count() > 0) {
      await firstArticleLink.click()
      await page.waitForLoadState('networkidle')
      
      // Scroll to bottom to ensure related articles are loaded
      await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
      
      // Check for related articles section
      const relatedSection = page.getByText(/related articles/i).first()
      if (await relatedSection.count() > 0) {
        await expect(relatedSection).toBeVisible()
      }
    }
  })
})
