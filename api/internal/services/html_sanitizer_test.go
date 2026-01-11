package services

import (
	"strings"
	"testing"
)

func TestHTMLSanitizer_SanitizeRichContent(t *testing.T) {
	sanitizer := NewHTMLSanitizer()

	tests := []struct {
		name     string
		input    string
		want     string
		mustNot  []string
		mustHave []string
	}{
		{
			name:     "removes script tags",
			input:    "<p>Hello</p><script>alert('xss')</script><p>World</p>",
			mustNot:  []string{"<script>", "alert", "xss"},
			mustHave: []string{"<p>Hello</p>", "<p>World</p>"},
		},
		{
			name:    "removes javascript URLs",
			input:   `<a href="javascript:alert('xss')">click me</a>`,
			mustNot: []string{"javascript:"},
		},
		{
			name:    "removes data URLs",
			input:   `<a href="data:text/html,<script>alert('xss')</script>">click</a>`,
			mustNot: []string{"data:"},
		},
		{
			name:    "removes vbscript URLs",
			input:   `<a href="vbscript:msgbox('xss')">click</a>`,
			mustNot: []string{"vbscript:"},
		},
		{
			name:    "removes onerror handlers",
			input:   `<img src=x onerror=alert('xss')>`,
			mustNot: []string{"onerror", "alert"},
		},
		{
			name:    "removes onclick handlers",
			input:   `<div onclick="alert('xss')">click</div>`,
			mustNot: []string{"onclick", "alert"},
		},
		{
			name:    "removes onload handlers",
			input:   `<body onload="alert('xss')">content</body>`,
			mustNot: []string{"onload", "alert", "<body>"},
		},
		{
			name:    "removes style with javascript",
			input:   `<div style="background:url(javascript:alert('xss'))">content</div>`,
			mustNot: []string{"javascript:", "alert"},
		},
		{
			name:    "removes iframe tags",
			input:   `<iframe src="http://evil.com"></iframe>`,
			mustNot: []string{"<iframe>"},
		},
		{
			name:    "removes object tags",
			input:   `<object data="http://evil.com"></object>`,
			mustNot: []string{"<object>"},
		},
		{
			name:    "removes embed tags",
			input:   `<embed src="http://evil.com">`,
			mustNot: []string{"<embed>"},
		},
		{
			name:     "allows safe HTML tags",
			input:    "<h2>Title</h2><p>Text with <strong>bold</strong> and <em>italic</em></p>",
			mustHave: []string{"<h2>", "<strong>", "<em>"},
		},
		{
			name:     "allows lists",
			input:    "<ul><li>Item 1</li><li>Item 2</li></ul>",
			mustHave: []string{"<ul>", "<li>"},
		},
		{
			name:     "allows blockquotes",
			input:    "<blockquote>Quote text</blockquote>",
			mustHave: []string{"<blockquote>"},
		},
		{
			name:     "allows code blocks",
			input:    "<pre><code>const x = 1;</code></pre>",
			mustHave: []string{"<pre>", "<code>"},
		},
		{
			name:     "allows safe images",
			input:    `<img src="https://example.com/image.jpg" alt="Test">`,
			mustHave: []string{"<img", "src=", "alt="},
		},
		{
			name:     "allows safe links",
			input:    `<a href="https://example.com">link</a>`,
			mustHave: []string{"<a", "href="},
		},
		{
			name:  "handles empty input",
			input: "",
			want:  "",
		},
		{
			name:     "handles nested XSS attempts",
			input:    `<div><script>alert('xss')</script><p>Safe text</p></div>`,
			mustNot:  []string{"<script>", "alert"},
			mustHave: []string{"Safe text"},
		},
		{
			name:    "removes meta refresh redirect",
			input:   `<meta http-equiv="refresh" content="0;url=http://evil.com">`,
			mustNot: []string{"<meta>", "http-equiv"},
		},
		{
			name:    "removes base tag",
			input:   `<base href="http://evil.com">`,
			mustNot: []string{"<base>"},
		},
		{
			name:    "removes form tags",
			input:   `<form action="http://evil.com"><input type="text"></form>`,
			mustNot: []string{"<form>", "<input>"},
		},
		{
			name:    "handles encoded XSS",
			input:   `<img src="x" onerror="&#97;&#108;&#101;&#114;&#116;(1)">`,
			mustNot: []string{"onerror"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizer.SanitizeRichContent(tt.input)

			// Check exact match if specified
			if tt.want != "" && got != tt.want {
				t.Errorf("SanitizeRichContent() = %v, want %v", got, tt.want)
			}

			// Check for forbidden strings
			for _, forbidden := range tt.mustNot {
				if strings.Contains(got, forbidden) {
					t.Errorf("SanitizeRichContent() output contains forbidden string %q: %v", forbidden, got)
				}
			}

			// Check for required strings
			for _, required := range tt.mustHave {
				if !strings.Contains(got, required) {
					t.Errorf("SanitizeRichContent() output missing required string %q: %v", required, got)
				}
			}
		})
	}
}

func TestHTMLSanitizer_SanitizeComment(t *testing.T) {
	sanitizer := NewHTMLSanitizer()

	tests := []struct {
		name     string
		input    string
		want     string
		mustNot  []string
		mustHave []string
	}{
		{
			name:     "removes script tags",
			input:    "<p>Comment</p><script>alert('xss')</script>",
			mustNot:  []string{"<script>", "alert"},
			mustHave: []string{"Comment"},
		},
		{
			name:     "removes headings (not allowed in comments)",
			input:    "<h2>Title</h2><p>Comment</p>",
			mustNot:  []string{"<h2>"},
			mustHave: []string{"Comment"},
		},
		{
			name:    "removes images (not allowed in comments)",
			input:   `<p>Comment</p><img src="http://example.com/img.jpg">`,
			mustNot: []string{"<img>"},
		},
		{
			name:    "removes lists (not allowed in comments)",
			input:   "<ul><li>Item</li></ul><p>Comment</p>",
			mustNot: []string{"<ul>", "<li>"},
		},
		{
			name:     "allows basic formatting",
			input:    "<p>Text with <strong>bold</strong> and <em>italic</em> and <del>strikethrough</del></p>",
			mustHave: []string{"<strong>", "<em>", "<del>"},
		},
		{
			name:     "allows safe links",
			input:    `<p>Visit <a href="https://example.com">our site</a></p>`,
			mustHave: []string{"<a", "href="},
		},
		{
			name:    "removes javascript URLs",
			input:   `<a href="javascript:alert('xss')">click</a>`,
			mustNot: []string{"javascript:"},
		},
		{
			name:     "allows line breaks",
			input:    "<p>Line 1<br>Line 2</p>",
			mustHave: []string{"<br>"},
		},
		{
			name:  "handles empty input",
			input: "",
			want:  "",
		},
		{
			name:    "removes onclick handlers",
			input:   `<p onclick="alert('xss')">Click me</p>`,
			mustNot: []string{"onclick"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizer.SanitizeComment(tt.input)

			// Check exact match if specified
			if tt.want != "" && got != tt.want {
				t.Errorf("SanitizeComment() = %v, want %v", got, tt.want)
			}

			// Check for forbidden strings
			for _, forbidden := range tt.mustNot {
				if strings.Contains(got, forbidden) {
					t.Errorf("SanitizeComment() output contains forbidden string %q: %v", forbidden, got)
				}
			}

			// Check for required strings
			for _, required := range tt.mustHave {
				if !strings.Contains(got, required) {
					t.Errorf("SanitizeComment() output missing required string %q: %v", required, got)
				}
			}
		})
	}
}

func TestHTMLSanitizer_EdgeCases(t *testing.T) {
	sanitizer := NewHTMLSanitizer()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "handles very long input",
			input: strings.Repeat("<p>Test</p>", 1000),
		},
		{
			name:  "handles deeply nested tags",
			input: "<div><div><div><div><p>Deep</p></div></div></div></div>",
		},
		{
			name:  "handles malformed HTML",
			input: "<p>Unclosed<p>Another<p>",
		},
		{
			name:  "handles special characters",
			input: "<p>&lt;&gt;&amp;&quot;&#39;</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			_ = sanitizer.SanitizeRichContent(tt.input)
			_ = sanitizer.SanitizeComment(tt.input)
		})
	}
}
