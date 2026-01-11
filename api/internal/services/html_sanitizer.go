package services

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// HTMLSanitizer provides HTML sanitization for user-generated content
// It implements defense-in-depth by sanitizing HTML before storage
type HTMLSanitizer struct {
	richContentPolicy *bluemonday.Policy
	commentPolicy     *bluemonday.Policy
}

// NewHTMLSanitizer creates a new HTML sanitizer with configured policies
func NewHTMLSanitizer() *HTMLSanitizer {
	// Rich content policy for articles and voter education
	// Allows rich formatting while preventing XSS attacks
	richPolicy := bluemonday.UGCPolicy()

	// Allow TipTap editor classes on div and p elements
	richPolicy.AllowAttrs("class").Matching(regexp.MustCompile("^prose-.*$")).OnElements("div", "p")

	// Additional TipTap elements
	richPolicy.AllowElements("h2", "h3", "h4", "h5", "h6")
	richPolicy.AllowElements("ul", "ol", "li")
	richPolicy.AllowElements("blockquote", "pre", "code")
	richPolicy.AllowElements("strong", "em", "u", "del", "s")
	richPolicy.AllowElements("br", "hr")

	// Allow images with safe attributes
	richPolicy.AllowAttrs("src", "alt", "title").OnElements("img")

	// Allow links with safe attributes
	richPolicy.AllowAttrs("href", "title", "target", "rel").OnElements("a")

	// Require parseable URLs (blocks javascript:, data:, vbscript:, etc.)
	richPolicy.RequireParseableURLs(true)

	// Simple comment policy - only basic formatting
	commentPolicy := bluemonday.NewPolicy()
	commentPolicy.AllowElements("p", "br")
	commentPolicy.AllowElements("strong", "em", "del")

	// Allow links with href attribute - must allow element AND schemes
	commentPolicy.AllowElements("a")
	commentPolicy.AllowURLSchemes("http", "https", "mailto")
	commentPolicy.AllowAttrs("href").OnElements("a")

	return &HTMLSanitizer{
		richContentPolicy: richPolicy,
		commentPolicy:     commentPolicy,
	}
}

// SanitizeRichContent sanitizes HTML for rich content (articles, voter education)
// Allows headings, lists, blockquotes, images, and links with safe attributes
func (s *HTMLSanitizer) SanitizeRichContent(html string) string {
	if html == "" {
		return ""
	}
	return s.richContentPolicy.Sanitize(html)
}

// SanitizeComment sanitizes HTML for comments
// Only allows basic formatting: paragraphs, line breaks, bold, italic, strikethrough, and links
func (s *HTMLSanitizer) SanitizeComment(html string) string {
	if html == "" {
		return ""
	}
	return s.commentPolicy.Sanitize(html)
}
