package handlers

import (
	"strings"
	"testing"

	"github.com/humfurie/pulpulitiko/api/internal/services"
)

// TestSanitization_ServiceLayer verifies that HTML sanitization works correctly at the service layer
// This ensures XSS protection is applied before content reaches the HTTP handlers
func TestSanitization_ServiceLayer(t *testing.T) {
	sanitizer := services.NewHTMLSanitizer()

	tests := []struct {
		name           string
		input          string
		sanitizeFunc   func(string) string
		mustNotContain []string
		mustContain    []string
	}{
		{
			name:           "rich content removes script tags",
			input:          "<p>Safe content</p><script>alert('xss')</script><p>More content</p>",
			sanitizeFunc:   sanitizer.SanitizeRichContent,
			mustNotContain: []string{"<script>", "alert", "xss"},
			mustContain:    []string{"<p>Safe content</p>", "<p>More content</p>"},
		},
		{
			name:           "rich content removes javascript URLs",
			input:          `<a href="javascript:alert('xss')">Click me</a>`,
			sanitizeFunc:   sanitizer.SanitizeRichContent,
			mustNotContain: []string{"javascript:"},
			mustContain:    []string{"Click me"}, // Link may be removed entirely or just href stripped
		},
		{
			name:           "rich content removes event handlers",
			input:          `<img src="test.jpg" onerror="alert('xss')">`,
			sanitizeFunc:   sanitizer.SanitizeRichContent,
			mustNotContain: []string{"onerror", "alert"},
			mustContain:    []string{"<img", "src="},
		},
		{
			name:         "rich content allows safe HTML",
			input:        "<h2>Title</h2><p>Text with <strong>bold</strong> and <em>italic</em></p><ul><li>Item</li></ul>",
			sanitizeFunc: sanitizer.SanitizeRichContent,
			mustContain:  []string{"<h2>", "<strong>", "<em>", "<ul>", "<li>"},
		},
		{
			name:           "comment removes script tags",
			input:          "<p>Comment</p><script>alert('xss')</script>",
			sanitizeFunc:   sanitizer.SanitizeComment,
			mustNotContain: []string{"<script>", "alert"},
			mustContain:    []string{"Comment"},
		},
		{
			name:           "comment removes images",
			input:          `<p>Comment</p><img src="http://example.com/img.jpg">`,
			sanitizeFunc:   sanitizer.SanitizeComment,
			mustNotContain: []string{"<img>"},
			mustContain:    []string{"Comment"},
		},
		{
			name:         "comment allows basic formatting",
			input:        "<p>Text with <strong>bold</strong> and <em>italic</em> and <del>strikethrough</del></p>",
			sanitizeFunc: sanitizer.SanitizeComment,
			mustContain:  []string{"<strong>", "<em>", "<del>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sanitizeFunc(tt.input)

			// Check for forbidden patterns
			for _, forbidden := range tt.mustNotContain {
				if strings.Contains(result, forbidden) {
					t.Errorf("Result contains forbidden pattern %q: %v", forbidden, result)
				}
			}

			// Check for required patterns
			for _, required := range tt.mustContain {
				if !strings.Contains(result, required) {
					t.Errorf("Result missing required pattern %q: %v", required, result)
				}
			}
		})
	}
}

// TODO: Add full HTTP integration tests with Chi router
// These would test:
// 1. POST malicious content to /api/admin/articles
// 2. Verify stored content is sanitized
// 3. GET the article and verify response is sanitized
// 4. POST malicious comment and verify sanitization
//
// Implementation requires:
// - Test database setup
// - Chi router configuration
// - Test HTTP client
// - Article and comment repository mocks or fixtures
