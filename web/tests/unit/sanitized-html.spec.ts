import { describe, it, expect } from 'vitest'
import { useSanitizedHtml } from '~/composables/useSanitizedHtml'

describe('useSanitizedHtml', () => {
  const { sanitizeRichContent, sanitizeComment } = useSanitizedHtml()

  describe('sanitizeRichContent', () => {
    it('removes script tags', () => {
      const result = sanitizeRichContent('<p>Hello</p><script>alert("xss")</script><p>World</p>')
      expect(result).toBe('<p>Hello</p><p>World</p>')
      expect(result).not.toContain('script')
      expect(result).not.toContain('alert')
    })

    it('removes javascript: URLs in links', () => {
      const result = sanitizeRichContent('<a href="javascript:alert(\'xss\')">click</a>')
      expect(result).not.toContain('javascript:')
      expect(result).not.toContain('alert')
    })

    it('removes data: URLs', () => {
      const result = sanitizeRichContent('<a href="data:text/html,<script>alert(\'xss\')</script>">click</a>')
      expect(result).not.toContain('data:')
    })

    it('removes vbscript: URLs', () => {
      const result = sanitizeRichContent('<a href="vbscript:msgbox(\'xss\')">click</a>')
      expect(result).not.toContain('vbscript:')
    })

    it('removes onerror handlers', () => {
      const result = sanitizeRichContent('<img src=x onerror=alert("xss")>')
      expect(result).not.toContain('onerror')
      expect(result).not.toContain('alert')
    })

    it('removes onclick handlers', () => {
      const result = sanitizeRichContent('<div onclick="alert(\'xss\')">click</div>')
      expect(result).not.toContain('onclick')
      expect(result).not.toContain('alert')
    })

    it('removes onload handlers', () => {
      const result = sanitizeRichContent('<body onload="alert(\'xss\')">content</body>')
      expect(result).not.toContain('onload')
      expect(result).not.toContain('alert')
    })

    it('removes style with javascript', () => {
      const result = sanitizeRichContent('<div style="background:url(javascript:alert(\'xss\'))">content</div>')
      expect(result).not.toContain('javascript:')
    })

    it('removes iframe tags', () => {
      const result = sanitizeRichContent('<iframe src="http://evil.com"></iframe>')
      expect(result).not.toContain('iframe')
    })

    it('removes object tags', () => {
      const result = sanitizeRichContent('<object data="http://evil.com"></object>')
      expect(result).not.toContain('object')
    })

    it('removes embed tags', () => {
      const result = sanitizeRichContent('<embed src="http://evil.com">')
      expect(result).not.toContain('embed')
    })

    it('allows safe HTML tags for rich content', () => {
      const html = '<h2>Title</h2><p>Text with <strong>bold</strong> and <em>italic</em></p>'
      const result = sanitizeRichContent(html)
      expect(result).toContain('<h2>')
      expect(result).toContain('<strong>')
      expect(result).toContain('<em>')
    })

    it('allows lists', () => {
      const html = '<ul><li>Item 1</li><li>Item 2</li></ul>'
      const result = sanitizeRichContent(html)
      expect(result).toContain('<ul>')
      expect(result).toContain('<li>')
    })

    it('allows blockquotes', () => {
      const html = '<blockquote>Quote text</blockquote>'
      const result = sanitizeRichContent(html)
      expect(result).toContain('<blockquote>')
    })

    it('allows code blocks', () => {
      const html = '<pre><code>const x = 1;</code></pre>'
      const result = sanitizeRichContent(html)
      expect(result).toContain('<pre>')
      expect(result).toContain('<code>')
    })

    it('allows safe images', () => {
      const html = '<img src="https://example.com/image.jpg" alt="Test">'
      const result = sanitizeRichContent(html)
      expect(result).toContain('<img')
      expect(result).toContain('src')
      expect(result).toContain('alt')
    })

    it('allows safe links', () => {
      const html = '<a href="https://example.com">link</a>'
      const result = sanitizeRichContent(html)
      expect(result).toContain('<a')
      expect(result).toContain('href')
    })

    it('handles empty input', () => {
      const result = sanitizeRichContent('')
      expect(result).toBe('')
    })

    it('handles undefined input', () => {
      const result = sanitizeRichContent(undefined)
      expect(result).toBe('')
    })

    it('handles nested XSS attempts', () => {
      const html = '<div><script>alert("xss")</script><p>Safe text</p></div>'
      const result = sanitizeRichContent(html)
      expect(result).not.toContain('script')
      expect(result).not.toContain('alert')
      expect(result).toContain('Safe text')
    })

    it('removes meta refresh redirect', () => {
      const html = '<meta http-equiv="refresh" content="0;url=http://evil.com">'
      const result = sanitizeRichContent(html)
      expect(result).not.toContain('meta')
      expect(result).not.toContain('http-equiv')
    })

    it('removes form tags', () => {
      const html = '<form action="http://evil.com"><input type="text"></form>'
      const result = sanitizeRichContent(html)
      expect(result).not.toContain('form')
      expect(result).not.toContain('input')
    })

    it('handles encoded XSS attempts', () => {
      const html = '<img src="x" onerror="&#97;&#108;&#101;&#114;&#116;(1)">'
      const result = sanitizeRichContent(html)
      expect(result).not.toContain('onerror')
    })
  })

  describe('sanitizeComment', () => {
    it('removes script tags from comments', () => {
      const html = '<p>Comment</p><script>alert("xss")</script>'
      const result = sanitizeComment(html)
      expect(result).not.toContain('script')
      expect(result).not.toContain('alert')
      expect(result).toContain('Comment')
    })

    it('removes headings (not allowed in comments)', () => {
      const html = '<h2>Title</h2><p>Comment</p>'
      const result = sanitizeComment(html)
      expect(result).not.toContain('<h2>')
      expect(result).toContain('Comment')
    })

    it('removes images (not allowed in comments)', () => {
      const html = '<p>Comment</p><img src="http://example.com/img.jpg">'
      const result = sanitizeComment(html)
      expect(result).not.toContain('<img')
    })

    it('removes lists (not allowed in comments)', () => {
      const html = '<ul><li>Item</li></ul><p>Comment</p>'
      const result = sanitizeComment(html)
      expect(result).not.toContain('<ul>')
      expect(result).not.toContain('<li>')
    })

    it('allows basic formatting in comments', () => {
      const html = '<p>Text with <strong>bold</strong> and <em>italic</em> and <del>strikethrough</del></p>'
      const result = sanitizeComment(html)
      expect(result).toContain('<strong>')
      expect(result).toContain('<em>')
      expect(result).toContain('<del>')
    })

    it('allows safe links in comments', () => {
      const html = '<p>Visit <a href="https://example.com">our site</a></p>'
      const result = sanitizeComment(html)
      expect(result).toContain('<a')
      expect(result).toContain('href')
    })

    it('removes javascript: URLs from comment links', () => {
      const html = '<a href="javascript:alert(\'xss\')">click</a>'
      const result = sanitizeComment(html)
      expect(result).not.toContain('javascript:')
    })

    it('allows line breaks in comments', () => {
      const html = '<p>Line 1<br>Line 2</p>'
      const result = sanitizeComment(html)
      expect(result).toContain('<br>')
    })

    it('handles empty comment input', () => {
      const result = sanitizeComment('')
      expect(result).toBe('')
    })

    it('handles undefined comment input', () => {
      const result = sanitizeComment(undefined)
      expect(result).toBe('')
    })

    it('removes onclick handlers from comments', () => {
      const html = '<p onclick="alert(\'xss\')">Click me</p>'
      const result = sanitizeComment(html)
      expect(result).not.toContain('onclick')
    })

    it('removes blockquotes (not allowed in comments)', () => {
      const html = '<p>Text</p><blockquote>Quote</blockquote>'
      const result = sanitizeComment(html)
      expect(result).not.toContain('<blockquote>')
    })
  })

  describe('Edge Cases', () => {
    it('handles very long content', { timeout: 15000 }, () => {
      const longHtml = '<p>' + 'Test '.repeat(1000) + '</p>'
      const result = sanitizeRichContent(longHtml)
      expect(result).toContain('<p>')
      expect(result).toContain('Test')
    })

    it('handles deeply nested tags', () => {
      const html = '<div><div><div><p>Deep</p></div></div></div>'
      const result = sanitizeRichContent(html)
      expect(result).toContain('Deep')
    })

    it('handles malformed HTML', () => {
      const html = '<p>Unclosed<p>Another<p>'
      // Should not throw
      expect(() => sanitizeRichContent(html)).not.toThrow()
    })

    it('handles special characters', () => {
      const html = '<p>&lt;&gt;&amp;&quot;&#39;</p>'
      const result = sanitizeRichContent(html)
      expect(result).toContain('<p>')
    })

    it('handles mixed case XSS attempts', () => {
      const html = '<ScRiPt>alert("xss")</sCrIpT>'
      const result = sanitizeRichContent(html)
      expect(result).not.toContain('cript')
      expect(result).not.toContain('alert')
    })
  })
})
