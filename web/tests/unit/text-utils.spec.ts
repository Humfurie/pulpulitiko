import { describe, it, expect } from 'vitest'
import { useTextUtils } from '~/composables/useTextUtils'

describe('useTextUtils', () => {
  const { countWordsInHtml, htmlToPlainText, sanitizeForSchema } = useTextUtils()

  describe('htmlToPlainText', () => {
    it('extracts plain text from simple HTML', () => {
      const result = htmlToPlainText('<p>Hello world</p>')
      expect(result).toBe('Hello world')
    })

    it('handles nested HTML tags', () => {
      const result = htmlToPlainText('<div><p>Hello <strong>world</strong>!</p></div>')
      expect(result).toBe('Hello world!')
    })

    it('handles HTML entities', () => {
      const result = htmlToPlainText('<p>&lt;tag&gt; &amp; &quot;quotes&quot;</p>')
      expect(result).toBe('<tag> & "quotes"')
    })

    it('handles numeric HTML entities', () => {
      const result = htmlToPlainText('<p>It&#39;s working &#8211; great!</p>')
      expect(result).toBe("It's working – great!")
    })

    it('removes script and style tags', () => {
      const result = htmlToPlainText('<p>Text</p><script>alert("xss")</script><style>.hidden{}</style>')
      expect(result).toBe('Text')
      expect(result).not.toContain('alert')
      expect(result).not.toContain('hidden')
    })

    it('handles empty input', () => {
      expect(htmlToPlainText('')).toBe('')
      expect(htmlToPlainText(undefined)).toBe('')
    })

    it('handles malformed HTML gracefully', () => {
      const result = htmlToPlainText('<p>Unclosed<p>Another')
      expect(result).toContain('Unclosed')
      expect(result).toContain('Another')
    })

    it('preserves spaces between elements', () => {
      const result = htmlToPlainText('<p>First</p><p>Second</p>')
      expect(result).toContain('First')
      expect(result).toContain('Second')
    })

    it('handles deeply nested tags', () => {
      const result = htmlToPlainText('<div><div><div><p>Deep text</p></div></div></div>')
      expect(result).toBe('Deep text')
    })
  })

  describe('countWordsInHtml', () => {
    it('counts words in simple HTML', () => {
      expect(countWordsInHtml('<p>Hello world</p>')).toBe(2)
    })

    it('counts words with nested HTML', () => {
      expect(countWordsInHtml('<p>Hello <strong>beautiful</strong> world!</p>')).toBe(3)
    })

    it('handles HTML entities correctly', () => {
      expect(countWordsInHtml('<p>It&#39;s working</p>')).toBe(2)
    })

    it('handles nbsp entities', () => {
      expect(countWordsInHtml('<p>word1&nbsp;word2</p>')).toBe(2)
    })

    it('ignores script and style content', () => {
      expect(countWordsInHtml('<p>Text</p><script>alert("xss injection")</script>')).toBe(1)
    })

    it('handles empty content', () => {
      expect(countWordsInHtml('')).toBe(0)
      expect(countWordsInHtml(undefined)).toBe(0)
      expect(countWordsInHtml('<p></p>')).toBe(0)
      expect(countWordsInHtml('<p>   </p>')).toBe(0)
    })

    it('collapses multiple spaces', () => {
      expect(countWordsInHtml('<p>Hello    world</p>')).toBe(2)
      expect(countWordsInHtml('<p>Hello\n\n\nworld</p>')).toBe(2)
    })

    it('counts words in lists', () => {
      const html = '<ul><li>First item</li><li>Second item</li></ul>'
      expect(countWordsInHtml(html)).toBe(4)
    })

    it('counts words in complex article structure', () => {
      const html = `
        <h2>Title Here</h2>
        <p>First paragraph with <strong>bold text</strong>.</p>
        <ul>
          <li>List item one</li>
          <li>List item two</li>
        </ul>
        <p>Second paragraph with <em>italic</em> text.</p>
      `
      // Title: 2, First paragraph: 5, List: 6, Second paragraph: 5
      // Total: 18 words
      expect(countWordsInHtml(html)).toBe(18)
    })

    it('handles line breaks and formatting', () => {
      const html = '<p>Line one<br>Line two<br>Line three</p>'
      expect(countWordsInHtml(html)).toBe(6)
    })
  })

  describe('sanitizeForSchema', () => {
    it('extracts plain text from HTML', () => {
      const result = sanitizeForSchema('<p>Hello <strong>world</strong></p>')
      expect(result).toBe('Hello world')
    })

    it('handles HTML entities', () => {
      const result = sanitizeForSchema('<p>&lt;tag&gt; &amp; &quot;quotes&quot;</p>')
      expect(result).toBe('<tag> & "quotes"')
    })

    it('handles numeric entities', () => {
      const result = sanitizeForSchema('<p>It&#39;s great</p>')
      expect(result).toBe("It's great")
    })

    it('collapses multiple spaces', () => {
      const result = sanitizeForSchema('<p>Hello    world   test</p>')
      expect(result).toBe('Hello world test')
    })

    it('collapses newlines and tabs', () => {
      const result = sanitizeForSchema('<p>Hello\n\nworld\t\ttest</p>')
      expect(result).toBe('Hello world test')
    })

    it('truncates at 5000 characters', () => {
      const longText = '<p>' + 'a'.repeat(6000) + '</p>'
      const result = sanitizeForSchema(longText)
      expect(result?.length).toBe(5000)
      expect(result).toMatch(/\.\.\.$/  ) // Ends with ...
    })

    it('does not truncate short content', () => {
      const shortText = '<p>' + 'a'.repeat(1000) + '</p>'
      const result = sanitizeForSchema(shortText)
      expect(result?.length).toBe(1000)
      expect(result).not.toContain('...')
    })

    it('handles undefined input', () => {
      expect(sanitizeForSchema(undefined)).toBeUndefined()
    })

    it('handles empty input', () => {
      expect(sanitizeForSchema('')).toBe('')
    })

    it('removes script tags', () => {
      const result = sanitizeForSchema('<p>Text</p><script>alert("xss")</script>')
      expect(result).toBe('Text')
      expect(result).not.toContain('script')
      expect(result).not.toContain('alert')
    })

    it('removes style tags', () => {
      const result = sanitizeForSchema('<p>Text</p><style>.class{}</style>')
      expect(result).toBe('Text')
      expect(result).not.toContain('style')
      expect(result).not.toContain('class')
    })

    it('handles complex HTML structure', () => {
      const html = `
        <article>
          <h1>Title</h1>
          <p>First paragraph with <strong>emphasis</strong>.</p>
          <ul>
            <li>Item 1</li>
            <li>Item 2</li>
          </ul>
          <p>Second paragraph.</p>
        </article>
      `
      const result = sanitizeForSchema(html)
      expect(result).toContain('Title')
      expect(result).toContain('First paragraph with emphasis')
      expect(result).toContain('Item 1')
      expect(result).toContain('Second paragraph')
      expect(result).not.toContain('<')
      expect(result).not.toContain('>')
    })

    it('trims leading and trailing whitespace', () => {
      const result = sanitizeForSchema('   <p>Text</p>   ')
      expect(result).toBe('Text')
    })

    it('handles blockquotes', () => {
      const result = sanitizeForSchema('<blockquote>Quoted text</blockquote><p>Regular text</p>')
      expect(result).toBe('Quoted text Regular text')
    })

    it('handles mixed HTML entities and numeric entities', () => {
      const result = sanitizeForSchema('<p>&quot;It&#39;s&quot; a test &amp; more</p>')
      expect(result).toBe('"It\'s" a test & more')
    })
  })

  describe('Edge Cases', () => {
    it('handles very long words', () => {
      const longWord = 'a'.repeat(1000)
      expect(countWordsInHtml(`<p>${longWord}</p>`)).toBe(1)
    })

    it('handles special Unicode characters', () => {
      const result = htmlToPlainText('<p>Hello 世界 🌍</p>')
      expect(result).toContain('Hello')
      expect(result).toContain('世界')
      expect(result).toContain('🌍')
    })

    it('handles mixed content types', () => {
      const html = `
        <div>
          <script>var x = 1;</script>
          <p>Text content</p>
          <style>.hidden { display: none; }</style>
          <p>More text</p>
        </div>
      `
      const result = htmlToPlainText(html)
      expect(result).toContain('Text content')
      expect(result).toContain('More text')
      expect(result).not.toContain('var x')
      expect(result).not.toContain('display: none')
    })

    it('handles self-closing tags', () => {
      const result = htmlToPlainText('<p>Before<br/>After</p>')
      expect(result).toContain('Before')
      expect(result).toContain('After')
    })

    it('handles comments in HTML', () => {
      const result = htmlToPlainText('<p>Text<!-- comment -->More</p>')
      expect(result).toContain('Text')
      expect(result).toContain('More')
      expect(result).not.toContain('comment')
    })
  })
})
