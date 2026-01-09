import { test, expect } from '@playwright/test'

test.describe('XSS Prevention - Frontend Sanitization', () => {
  test('article content does not execute scripts', async ({ page }) => {
    // Navigate to articles list
    await page.goto('/articles')

    // Find first article
    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available for testing')
      return
    }

    await page.waitForLoadState('networkidle')

    // Check that no script tags exist in article content area
    const articleContent = page.locator('.article-content, [class*="prose"], main article')
    const scriptInContent = await articleContent.locator('script').count()
    expect(scriptInContent).toBe(0)

    // Check that no inline event handlers exist
    const elementsWithHandlers = await articleContent.locator('[onclick], [onerror], [onload], [onmouseover]').count()
    expect(elementsWithHandlers).toBe(0)

    // Check for javascript: URLs in links
    const links = await articleContent.locator('a').all()
    for (const link of links) {
      const href = await link.getAttribute('href')
      if (href) {
        expect(href.toLowerCase()).not.toContain('javascript:')
        expect(href.toLowerCase()).not.toContain('data:')
        expect(href.toLowerCase()).not.toContain('vbscript:')
      }
    }
  })

  test('comments do not contain malicious scripts', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Wait for comments section to load (if it exists)
    const commentsSection = page.locator('[class*="comment"], .comments, #comments')
    if (await commentsSection.count() > 0) {
      // Check for script tags in comments
      const scriptInComments = await commentsSection.locator('script').count()
      expect(scriptInComments).toBe(0)

      // Check for event handlers
      const handlersInComments = await commentsSection.locator('[onclick], [onerror], [onload]').count()
      expect(handlersInComments).toBe(0)

      // Check comment links
      const commentLinks = await commentsSection.locator('a').all()
      for (const link of commentLinks) {
        const href = await link.getAttribute('href')
        if (href) {
          expect(href.toLowerCase()).not.toContain('javascript:')
          expect(href.toLowerCase()).not.toContain('data:')
        }
      }
    }
  })

  test('images do not have malicious onerror handlers', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Check all images
    const images = await page.locator('img').all()
    for (const img of images) {
      const onerror = await img.getAttribute('onerror')
      const onload = await img.getAttribute('onload')

      expect(onerror).toBeNull()
      expect(onload).toBeNull()
    }
  })

  test('no inline style with javascript: URLs', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Check for javascript: in style attributes
    const elementsWithStyle = await page.locator('[style]').all()
    for (const element of elementsWithStyle) {
      const style = await element.getAttribute('style')
      if (style) {
        expect(style.toLowerCase()).not.toContain('javascript:')
        expect(style.toLowerCase()).not.toContain('expression(')
      }
    }
  })

  test('form elements do not have malicious actions', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Check forms (e.g., comment forms)
    const forms = await page.locator('form').all()
    for (const form of forms) {
      const action = await form.getAttribute('action')
      if (action) {
        expect(action.toLowerCase()).not.toContain('javascript:')
      }

      const onsubmit = await form.getAttribute('onsubmit')
      // onsubmit is allowed for legitimate form handling, but check for obvious XSS
      if (onsubmit) {
        expect(onsubmit.toLowerCase()).not.toContain('alert(')
        expect(onsubmit.toLowerCase()).not.toContain('eval(')
      }
    }
  })

  test('meta refresh tags do not redirect to malicious sites', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Check for meta refresh tags
    const metaRefresh = await page.locator('meta[http-equiv="refresh"]').count()
    expect(metaRefresh).toBe(0)
  })

  test('no iframe tags with malicious sources', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Check for iframes in article content (should be sanitized out)
    const iframes = await page.locator('.article-content iframe, [class*="prose"] iframe').count()
    expect(iframes).toBe(0)

    // If there are any iframes on the page (e.g., for embeds), check they're safe
    const allIframes = await page.locator('iframe').all()
    for (const iframe of allIframes) {
      const src = await iframe.getAttribute('src')
      if (src) {
        expect(src.toLowerCase()).not.toContain('javascript:')
        expect(src.toLowerCase()).not.toContain('data:')
      }
    }
  })

  test('no object or embed tags', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Check for object and embed tags (should be sanitized)
    const objectTags = await page.locator('object').count()
    const embedTags = await page.locator('embed').count()

    expect(objectTags).toBe(0)
    expect(embedTags).toBe(0)
  })

  test('CSP headers are present', async ({ page }) => {
    const response = await page.goto('/articles')

    if (response) {
      const headers = response.headers()

      // Check for CSP header (either enforcement or report-only)
      const hasCSP =
        headers['content-security-policy'] ||
        headers['content-security-policy-report-only']

      expect(hasCSP).toBeTruthy()

      // Check for other security headers
      expect(headers['x-content-type-options']).toBe('nosniff')
      expect(headers['x-frame-options']).toBeTruthy()
    }
  })

  test('voter education content is sanitized', async ({ page }) => {
    await page.goto('/voter-education')

    // Find first voter education content
    const firstItem = page.locator('a[href*="/voter-education/"]').first()
    if (await firstItem.count() > 0) {
      await firstItem.click()
      await page.waitForLoadState('networkidle')

      // Check content area for scripts
      const contentArea = page.locator('[class*="prose"], main')
      const scriptCount = await contentArea.locator('script').count()
      expect(scriptCount).toBe(0)

      // Check for event handlers
      const handlerCount = await contentArea.locator('[onclick], [onerror], [onload]').count()
      expect(handlerCount).toBe(0)
    }
  })
})

test.describe('XSS Prevention - OWASP Top 10 Payloads', () => {
  test('article list handles without XSS vulnerabilities', async ({ page }) => {
    await page.goto('/articles')

    // Ensure page loaded without executing any malicious scripts
    // If XSS was present, it would typically execute during page load

    // Check console for any errors that might indicate XSS attempts
    const consoleErrors: string[] = []
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text())
      }
    })

    await page.waitForLoadState('networkidle')

    // No alerts should be triggered (common XSS payload)
    const dialogPromise = page.waitForEvent('dialog', { timeout: 1000 }).catch(() => null)
    const dialog = await dialogPromise
    expect(dialog).toBeNull()
  })

  test('no eval() or Function() in page source', async ({ page }) => {
    await page.goto('/articles')

    const firstArticle = page.locator('article a, .article-card a, a[href*="/article/"]').first()
    if (await firstArticle.count() > 0) {
      await firstArticle.click()
    } else {
      test.skip(true, 'No articles available')
      return
    }

    await page.waitForLoadState('networkidle')

    // Check that eval() or Function() are not being used in inline scripts
    // This is a basic check - actual code may use these legitimately
    const pageContent = await page.content()
    const articleContent = await page.locator('.article-content, [class*="prose"]').textContent()

    // Ensure malicious eval patterns aren't in content
    if (articleContent) {
      expect(articleContent).not.toContain('eval(')
      expect(articleContent).not.toContain('Function(')
    }
  })
})
