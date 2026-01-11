export function useTextUtils() {
  /**
   * Converts HTML to plain text, correctly handling HTML entities and nested tags
   * Uses DOMParser for accurate parsing in browser, falls back to regex for SSR
   */
  function htmlToPlainText(html: string | undefined): string {
    if (!html) return ''

    // Try DOMParser first (browser environment) - handles ALL HTML entities correctly
    if (typeof DOMParser !== 'undefined') {
      try {
        // Remove script and style tags first
        const cleanHtml = html
          .replace(/<script[^>]*>.*?<\/script>/gi, '')
          .replace(/<style[^>]*>.*?<\/style>/gi, '')

        // Use DOMParser for complete HTML entity decoding
        // This handles all entities including &apos;, &mdash;, &ldquo;, etc.
        const doc = new DOMParser().parseFromString(cleanHtml, 'text/html')
        return (doc.body.textContent || '')
          .replace(/\s+/g, ' ') // Collapse multiple spaces
          .trim()
      } catch (e) {
        // Fall through to regex approach if DOMParser fails
        console.warn('DOMParser failed, falling back to regex approach', e)
      }
    }

    // Fallback regex-based approach for SSR or if DOMParser fails
    // This covers common entities but may miss uncommon ones
    return html
      .replace(/<script[^>]*>.*?<\/script>/gi, '')
      .replace(/<style[^>]*>.*?<\/style>/gi, '')
      // Replace closing block tags with space to preserve word boundaries
      .replace(/<\/(p|div|li|h[1-6]|blockquote|pre|article|section|header|footer|main|aside)>/gi, ' ')
      // Replace br tags with space
      .replace(/<br[^>]*>/gi, ' ')
      // Remove all remaining HTML tags
      .replace(/<[^>]+>/g, '')
      // Decode common HTML entities (incomplete - DOMParser is preferred)
      .replace(/&nbsp;/g, ' ')
      .replace(/&amp;/g, '&')
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&quot;/g, '"')
      .replace(/&#39;/g, "'")
      .replace(/&apos;/g, "'")
      .replace(/&mdash;/g, '—')
      .replace(/&ndash;/g, '–')
      .replace(/&ldquo;/g, '"')
      .replace(/&rdquo;/g, '"')
      .replace(/&lsquo;/g, ''')
      .replace(/&rsquo;/g, ''')
      .replace(/&#(\d+);/g, (_, num) => String.fromCharCode(parseInt(num, 10)))
      .replace(/&#x([0-9a-f]+);/gi, (_, hex) => String.fromCharCode(parseInt(hex, 16)))
      // Collapse multiple spaces
      .replace(/\s+/g, ' ')
      .trim()
  }

  /**
   * Counts words in HTML content using accurate DOMParser-based extraction
   * Handles nested tags, HTML entities, and multiple spaces correctly
   */
  function countWordsInHtml(html: string | undefined): number {
    const text = htmlToPlainText(html)

    // Split on whitespace, filter empty strings
    const words = text
      .trim()
      .split(/\s+/)
      .filter(word => word.length > 0)

    return words.length
  }

  /**
   * Sanitizes HTML for Schema.org structured data
   * Extracts plain text and handles all HTML entities (including numeric)
   * Truncates to 5000 characters as recommended by Google
   */
  function sanitizeForSchema(html: string | undefined): string | undefined {
    if (html === undefined) return undefined
    if (html === '') return ''

    const text = htmlToPlainText(html)
      .replace(/\s+/g, ' ') // Collapse multiple spaces
      .trim()

    // Truncate to 5000 chars (Google recommendation)
    return text.length > 5000
      ? text.substring(0, 4997) + '...'
      : text
  }

  return {
    htmlToPlainText,
    countWordsInHtml,
    sanitizeForSchema
  }
}
