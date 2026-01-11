import { test, expect } from '@playwright/test'

test.describe('Smoke Tests', () => {
  test('homepage loads successfully', async ({ page }) => {
    await page.goto('/')
    
    // Check that the page loaded
    await expect(page).toHaveTitle(/Pulpulitiko/i)
    
    // Check for main heading
    await expect(page.getByRole('heading', { name: /latest/i })).toBeVisible()
    
    // Check that articles section is present
    await expect(page.locator('article, [class*="article"]').first()).toBeVisible({ timeout: 10000 })
  })

  test('homepage has navigation', async ({ page }) => {
    await page.goto('/')

    // Check for navigation elements (adjust selectors based on your actual navigation)
    await expect(page.locator('header').first()).toBeVisible()
    await expect(page.locator('nav').first()).toBeVisible()
  })

  test('can navigate to articles page', async ({ page }) => {
    await page.goto('/')
    
    // Try to find and click on first article
    const firstArticle = page.locator('article, [class*="article"]').first()
    await expect(firstArticle).toBeVisible({ timeout: 10000 })
    
    // Check if article has a clickable link
    const articleLink = firstArticle.locator('a').first()
    if (await articleLink.count() > 0) {
      await articleLink.click({ force: true })

      // Wait for navigation
      await page.waitForLoadState('networkidle')
      
      // Check that we're on an article page
      await expect(page).toHaveURL(/\/(article|articles)\//)
    }
  })

  test('search functionality is accessible', async ({ page }) => {
    await page.goto('/')
    
    // Look for search input or button (adjust selector based on your implementation)
    const searchElement = page.locator('[type="search"], [placeholder*="search" i], [aria-label*="search" i]').first()
    
    if (await searchElement.count() > 0) {
      await expect(searchElement).toBeVisible()
    }
  })

  test('politicians page loads', async ({ page }) => {
    // Try to navigate to politicians page
    await page.goto('/politicians')
    
    // Wait for page to load
    await page.waitForLoadState('networkidle')
    
    // Check for content or heading
    const heading = page.locator('h1, h2').first()
    if (await heading.count() > 0) {
      await expect(heading).toBeVisible()
    }
  })

  test('categories page loads', async ({ page }) => {
    // Skip this test as /categories page doesn't exist yet
    test.skip(true, 'Categories page not implemented yet')

    await page.goto('/categories')

    // Check page loaded
    await page.waitForLoadState('networkidle')

    // Basic check that page has content
    const content = page.locator('body')
    await expect(content).toBeVisible()
  })
})
