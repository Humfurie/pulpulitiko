package models

import "github.com/google/uuid"

// CategoryMetric represents article count and views per category
type CategoryMetric struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	ArticleCount int       `json:"article_count"`
	TotalViews   int       `json:"total_views"`
}

// TagMetric represents article count and views per tag
type TagMetric struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	ArticleCount int       `json:"article_count"`
	TotalViews   int       `json:"total_views"`
}

// TopArticle represents a top viewed article
type TopArticle struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	ViewCount    int       `json:"view_count"`
	CategoryName *string   `json:"category_name,omitempty"`
}

// DashboardMetrics represents all metrics for the admin dashboard
type DashboardMetrics struct {
	TotalArticles   int              `json:"total_articles"`
	TotalViews      int              `json:"total_views"`
	TotalCategories int              `json:"total_categories"`
	TotalTags       int              `json:"total_tags"`
	TopArticles     []TopArticle     `json:"top_articles"`
	CategoryMetrics []CategoryMetric `json:"category_metrics"`
	TagMetrics      []TagMetric      `json:"tag_metrics"`
}
