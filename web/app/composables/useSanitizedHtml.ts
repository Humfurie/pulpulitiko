import DOMPurify from 'isomorphic-dompurify'

/**
 * Composable for sanitizing HTML content to prevent XSS attacks
 * Provides defense-in-depth protection for user-generated content
 */
export function useSanitizedHtml() {
  /**
   * Sanitize rich content (articles, voter education)
   * Allows rich formatting including headings, lists, images, and links
   */
  function sanitizeRichContent(html: string | undefined): string {
    if (!html) return ''

    return DOMPurify.sanitize(html, {
      ALLOWED_TAGS: [
        // Headings
        'h2',
        'h3',
        'h4',
        'h5',
        'h6',
        // Text formatting
        'p',
        'br',
        'strong',
        'em',
        'u',
        'del',
        's',
        // Lists
        'ul',
        'ol',
        'li',
        // Quotes and code
        'blockquote',
        'pre',
        'code',
        // Links and images
        'a',
        'img',
        // Horizontal rule
        'hr'
      ],
      ALLOWED_ATTR: ['href', 'src', 'alt', 'title', 'class'],
      ALLOW_DATA_ATTR: false, // Prevent data-* XSS vectors
      SAFE_FOR_TEMPLATES: true // Escape template syntax
    })
  }

  /**
   * Sanitize comment content
   * Only allows basic formatting - no images, headings, or lists
   */
  function sanitizeComment(html: string | undefined): string {
    if (!html) return ''

    return DOMPurify.sanitize(html, {
      ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'del', 'a'],
      ALLOWED_ATTR: ['href'],
      ALLOW_DATA_ATTR: false,
      SAFE_FOR_TEMPLATES: true
    })
  }

  return {
    sanitizeRichContent,
    sanitizeComment
  }
}
