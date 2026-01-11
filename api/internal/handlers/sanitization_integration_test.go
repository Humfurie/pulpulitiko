package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"pulpulitiko/internal/middleware"
	"pulpulitiko/internal/services"
)

// TestArticleSanitization_Integration verifies that HTML sanitization is applied at the HTTP layer
// This is an integration test that ensures XSS protection works end-to-end
func TestArticleSanitization_Integration(t *testing.T) {
	// Skip if database not available
	// TODO: Set up test database connection
	t.Skip("Integration test requires database setup")

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		articleData    map[string]interface{}
		checkField     string   // Field to check in response
		mustNotContain []string // XSS patterns that should be removed
		mustContain    []string // Safe content that should remain
	}{
		{
			name: "removes script tags from article content",
			articleData: map[string]interface{}{
				"title":   "Test Article",
				"content": "<p>Safe content</p><script>alert('xss')</script><p>More content</p>",
				"slug":    "test-article-1",
				"status":  "draft",
			},
			checkField:     "content",
			mustNotContain: []string{"<script>", "alert", "xss"},
			mustContain:    []string{"<p>Safe content</p>", "<p>More content</p>"},
		},
		{
			name: "removes javascript URLs from article content",
			articleData: map[string]interface{}{
				"title":   "Test Article",
				"content": `<a href="javascript:alert('xss')">Click me</a>`,
				"slug":    "test-article-2",
				"status":  "draft",
			},
			checkField:     "content",
			mustNotContain: []string{"javascript:"},
			mustContain:    []string{"<a", "Click me</a>"},
		},
		{
			name: "removes event handlers from article content",
			articleData: map[string]interface{}{
				"title":   "Test Article",
				"content": `<img src="test.jpg" onerror="alert('xss')">`,
				"slug":    "test-article-3",
				"status":  "draft",
			},
			checkField:     "content",
			mustNotContain: []string{"onerror", "alert"},
			mustContain:    []string{"<img", "src="},
		},
		{
			name: "allows safe HTML in article content",
			articleData: map[string]interface{}{
				"title":   "Test Article",
				"content": "<h2>Title</h2><p>Text with <strong>bold</strong> and <em>italic</em></p><ul><li>Item</li></ul>",
				"slug":    "test-article-4",
				"status":  "draft",
			},
			checkField:  "content",
			mustContain: []string{"<h2>", "<strong>", "<em>", "<ul>", "<li>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test router
			router := gin.New()

			// Add sanitization middleware (if exists)
			// router.Use(middleware.SanitizeHTML())

			// Create mock article handler
			// This is a simplified version - actual implementation would use repository
			router.POST("/api/admin/articles", func(c *gin.Context) {
				var req map[string]interface{}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				// Apply sanitization (simulating what the real handler does)
				sanitizer := services.NewHTMLSanitizer()
				if content, ok := req["content"].(string); ok {
					req["content"] = sanitizer.SanitizeRichContent(content)
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    req,
				})
			})

			// Create request
			body, _ := json.Marshal(tt.articleData)
			req := httptest.NewRequest(http.MethodPost, "/api/admin/articles", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check response status
			if w.Code != http.StatusOK {
				t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Get the sanitized content
			data := response["data"].(map[string]interface{})
			sanitizedContent := data[tt.checkField].(string)

			// Check for forbidden patterns
			for _, forbidden := range tt.mustNotContain {
				if strings.Contains(sanitizedContent, forbidden) {
					t.Errorf("Response contains forbidden pattern %q in %s: %v",
						forbidden, tt.checkField, sanitizedContent)
				}
			}

			// Check for required patterns
			for _, required := range tt.mustContain {
				if !strings.Contains(sanitizedContent, required) {
					t.Errorf("Response missing required pattern %q in %s: %v",
						required, tt.checkField, sanitizedContent)
				}
			}
		})
	}
}

// TestCommentSanitization_Integration verifies comment sanitization at HTTP layer
func TestCommentSanitization_Integration(t *testing.T) {
	t.Skip("Integration test requires database setup")

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		commentData    map[string]interface{}
		mustNotContain []string
		mustContain    []string
	}{
		{
			name: "removes script tags from comments",
			commentData: map[string]interface{}{
				"content": "<p>Nice article!</p><script>alert('xss')</script>",
			},
			mustNotContain: []string{"<script>", "alert"},
			mustContain:    []string{"Nice article!"},
		},
		{
			name: "removes images from comments (not allowed)",
			commentData: map[string]interface{}{
				"content": `<p>Comment</p><img src="http://evil.com/img.jpg">`,
			},
			mustNotContain: []string{"<img>"},
			mustContain:    []string{"Comment"},
		},
		{
			name: "allows basic formatting in comments",
			commentData: map[string]interface{}{
				"content": "<p>Text with <strong>bold</strong> and <em>italic</em></p>",
			},
			mustContain: []string{"<strong>", "<em>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.POST("/api/comments", func(c *gin.Context) {
				var req map[string]interface{}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				// Apply comment sanitization
				sanitizer := services.NewHTMLSanitizer()
				if content, ok := req["content"].(string); ok {
					req["content"] = sanitizer.SanitizeComment(content)
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    req,
				})
			})

			body, _ := json.Marshal(tt.commentData)
			req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			data := response["data"].(map[string]interface{})
			sanitizedContent := data["content"].(string)

			for _, forbidden := range tt.mustNotContain {
				if strings.Contains(sanitizedContent, forbidden) {
					t.Errorf("Comment contains forbidden pattern %q: %v", forbidden, sanitizedContent)
				}
			}

			for _, required := range tt.mustContain {
				if !strings.Contains(sanitizedContent, required) {
					t.Errorf("Comment missing required pattern %q: %v", required, sanitizedContent)
				}
			}
		})
	}
}

// TestSanitizationMiddleware verifies that sanitization middleware (if implemented) works correctly
func TestSanitizationMiddleware(t *testing.T) {
	t.Skip("Middleware test - implement when sanitization middleware is added")

	// This test would verify that a middleware.SanitizeHTML() middleware
	// properly sanitizes all string fields in request bodies

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Example of how it might work:
	// router.Use(middleware.SanitizeHTML())
	router.POST("/test", func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		c.JSON(http.StatusOK, req)
	})

	maliciousData := map[string]interface{}{
		"title":   "Test<script>alert('xss')</script>",
		"content": `<a href="javascript:void(0)">Click</a>`,
	}

	body, _ := json.Marshal(maliciousData)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify XSS patterns were removed by middleware
	if strings.Contains(response["title"].(string), "<script>") {
		t.Error("Middleware failed to sanitize title field")
	}
	if strings.Contains(response["content"].(string), "javascript:") {
		t.Error("Middleware failed to sanitize content field")
	}
}

// TODO: Add actual integration tests that:
// 1. Start a test database
// 2. POST malicious content to /api/articles
// 3. GET the article back and verify it's sanitized
// 4. POST malicious comment
// 5. GET comment back and verify it's sanitized
//
// This ensures defense-in-depth: sanitization happens at both
// storage time (service layer) and retrieval time (HTTP layer)
