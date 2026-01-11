import DOMPurify from 'isomorphic-dompurify'

/**
 * Composable for sanitizing HTML content to prevent XSS attacks
 * Provides defense-in-depth protection for user-generated content
 *
 * SECURITY NOTE: This policy MUST match the backend sanitization policy in
 * api/internal/services/html_sanitizer.go to ensure consistent behavior.
 * The backend is the authoritative source of truth for security.
 */
export function useSanitizedHtml() {
  /**
   * Sanitize rich content (articles, voter education)
   * Allows rich formatting including headings, lists, images, and links
   *
   * Policy alignment with backend (bluemonday.UGCPolicy):
   * - Same allowed tags
   * - class attribute only on div/p with prose-* pattern
   * - Explicit URL scheme whitelist (http, https, mailto)
   * - No data-* attributes
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
        'hr',
        // Container (for TipTap editor classes)
        'div'
      ],
      ALLOWED_ATTR: ['href', 'src', 'alt', 'title', 'target', 'rel'],
      // Restrict class attribute to only div/p with prose-* pattern (matches backend)
      // eslint-disable-next-line no-useless-escape
      ALLOWED_URI_REGEXP: /^(?:(?:(?:f|ht)tps?|mailto):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i,
      ALLOW_DATA_ATTR: false, // Prevent data-* XSS vectors
      SAFE_FOR_TEMPLATES: true, // Escape template syntax
      // Hook to enforce class attribute restrictions (matches backend policy)
      HOOK_AFTER_SANITIZE_ATTRIBUTES: (node: Element) => {
        // Only allow class attribute on div and p elements with prose-* pattern
        if (node.hasAttribute('class')) {
          const tagName = node.tagName.toLowerCase()
          const className = node.getAttribute('class') || ''

          // Remove class if not on div/p OR doesn't match prose-* pattern
          if (!['div', 'p'].includes(tagName) || !className.match(/^prose-.*$/)) {
            node.removeAttribute('class')
          }
        }
        return node
      }
    })
  }

  /**
   * Sanitize comment content
   * Only allows basic formatting - no images, headings, or lists
   *
   * Policy alignment with backend commentPolicy:
   * - Same allowed tags
   * - Explicit URL scheme whitelist
   * - No class attributes
   */
  function sanitizeComment(html: string | undefined): string {
    if (!html) return ''

    return DOMPurify.sanitize(html, {
      ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'del', 'a'],
      ALLOWED_ATTR: ['href'],
      // Explicit URL scheme whitelist (matches backend: http, https, mailto)
      // eslint-disable-next-line no-useless-escape
      ALLOWED_URI_REGEXP: /^(?:(?:(?:f|ht)tps?|mailto):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i,
      ALLOW_DATA_ATTR: false,
      SAFE_FOR_TEMPLATES: true
    })
  }

  return {
    sanitizeRichContent,
    sanitizeComment
  }
}
