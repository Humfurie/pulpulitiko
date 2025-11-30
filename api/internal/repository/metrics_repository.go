package repository

import (
	"context"
	"fmt"

	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricsRepository struct {
	db *pgxpool.Pool
}

func NewMetricsRepository(db *pgxpool.Pool) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func (r *MetricsRepository) GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error) {
	metrics := &models.DashboardMetrics{}

	// Get total articles and views
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(view_count), 0)
		FROM articles
		WHERE status = 'published'
	`).Scan(&metrics.TotalArticles, &metrics.TotalViews)
	if err != nil {
		return nil, fmt.Errorf("failed to get article totals: %w", err)
	}

	// Get total categories
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM categories").Scan(&metrics.TotalCategories)
	if err != nil {
		return nil, fmt.Errorf("failed to count categories: %w", err)
	}

	// Get total tags
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM tags").Scan(&metrics.TotalTags)
	if err != nil {
		return nil, fmt.Errorf("failed to count tags: %w", err)
	}

	// Get top articles by views
	topArticles, err := r.GetTopArticles(ctx, 10)
	if err != nil {
		return nil, err
	}
	metrics.TopArticles = topArticles

	// Get category metrics
	categoryMetrics, err := r.GetCategoryMetrics(ctx)
	if err != nil {
		return nil, err
	}
	metrics.CategoryMetrics = categoryMetrics

	// Get tag metrics
	tagMetrics, err := r.GetTagMetrics(ctx)
	if err != nil {
		return nil, err
	}
	metrics.TagMetrics = tagMetrics

	return metrics, nil
}

func (r *MetricsRepository) GetTopArticles(ctx context.Context, limit int) ([]models.TopArticle, error) {
	query := `
		SELECT a.id, a.slug, a.title, a.view_count, c.name
		FROM articles a
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE a.status = 'published'
		ORDER BY a.view_count DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top articles: %w", err)
	}
	defer rows.Close()

	articles := []models.TopArticle{}
	for rows.Next() {
		var article models.TopArticle
		err := rows.Scan(&article.ID, &article.Slug, &article.Title, &article.ViewCount, &article.CategoryName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func (r *MetricsRepository) GetCategoryMetrics(ctx context.Context) ([]models.CategoryMetric, error) {
	query := `
		SELECT c.id, c.name, c.slug,
			   COUNT(a.id) as article_count,
			   COALESCE(SUM(a.view_count), 0) as total_views
		FROM categories c
		LEFT JOIN articles a ON c.id = a.category_id AND a.status = 'published'
		GROUP BY c.id, c.name, c.slug
		ORDER BY total_views DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get category metrics: %w", err)
	}
	defer rows.Close()

	metrics := []models.CategoryMetric{}
	for rows.Next() {
		var metric models.CategoryMetric
		err := rows.Scan(&metric.ID, &metric.Name, &metric.Slug, &metric.ArticleCount, &metric.TotalViews)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (r *MetricsRepository) GetTagMetrics(ctx context.Context) ([]models.TagMetric, error) {
	query := `
		SELECT t.id, t.name, t.slug,
			   COUNT(DISTINCT at.article_id) as article_count,
			   COALESCE(SUM(a.view_count), 0) as total_views
		FROM tags t
		LEFT JOIN article_tags at ON t.id = at.tag_id
		LEFT JOIN articles a ON at.article_id = a.id AND a.status = 'published'
		GROUP BY t.id, t.name, t.slug
		ORDER BY total_views DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag metrics: %w", err)
	}
	defer rows.Close()

	metrics := []models.TagMetric{}
	for rows.Next() {
		var metric models.TagMetric
		err := rows.Scan(&metric.ID, &metric.Name, &metric.Slug, &metric.ArticleCount, &metric.TotalViews)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}
