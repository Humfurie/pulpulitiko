package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ArticleRepository struct {
	db *pgxpool.Pool
}

func NewArticleRepository(db *pgxpool.Pool) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(ctx context.Context, article *models.Article) error {
	query := `
		INSERT INTO articles (slug, title, summary, content, featured_image, author_id, category_id, status, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	var publishedAt *time.Time
	if article.Status == models.ArticleStatusPublished && article.PublishedAt == nil {
		now := time.Now()
		publishedAt = &now
	} else {
		publishedAt = article.PublishedAt
	}

	err := r.db.QueryRow(ctx, query,
		article.Slug,
		article.Title,
		article.Summary,
		article.Content,
		article.FeaturedImage,
		article.AuthorID,
		article.CategoryID,
		article.Status,
		publishedAt,
	).Scan(&article.ID, &article.CreatedAt, &article.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create article: %w", err)
	}

	article.PublishedAt = publishedAt
	return nil
}

func (r *ArticleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Article, error) {
	query := `
		SELECT a.id, a.slug, a.title, a.summary, a.content, a.featured_image,
			   a.author_id, a.category_id, a.status, a.view_count, a.published_at, a.created_at, a.updated_at,
			   au.id, au.name, au.slug, au.bio, au.avatar, au.email,
			   c.id, c.name, c.slug, c.description
		FROM articles a
		LEFT JOIN authors au ON a.author_id = au.id AND au.deleted_at IS NULL
		LEFT JOIN categories c ON a.category_id = c.id AND c.deleted_at IS NULL
		WHERE a.id = $1 AND a.deleted_at IS NULL
	`

	article := &models.Article{}
	var authorID, categoryID *uuid.UUID
	var authorName, authorSlug, authorBio, authorAvatar, authorEmail *string
	var categoryName, categorySlug, categoryDescription *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&article.ID, &article.Slug, &article.Title, &article.Summary, &article.Content, &article.FeaturedImage,
		&article.AuthorID, &article.CategoryID, &article.Status, &article.ViewCount, &article.PublishedAt, &article.CreatedAt, &article.UpdatedAt,
		&authorID, &authorName, &authorSlug, &authorBio, &authorAvatar, &authorEmail,
		&categoryID, &categoryName, &categorySlug, &categoryDescription,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	if authorID != nil {
		article.Author = &models.Author{
			ID:     *authorID,
			Name:   *authorName,
			Slug:   *authorSlug,
			Bio:    authorBio,
			Avatar: authorAvatar,
			Email:  authorEmail,
		}
	}
	if categoryID != nil {
		article.Category = &models.Category{
			ID:          *categoryID,
			Name:        *categoryName,
			Slug:        *categorySlug,
			Description: categoryDescription,
		}
	}

	tags, err := r.GetArticleTags(ctx, article.ID)
	if err != nil {
		return nil, err
	}
	article.Tags = tags

	return article, nil
}

func (r *ArticleRepository) GetBySlug(ctx context.Context, slug string) (*models.Article, error) {
	query := `
		SELECT a.id, a.slug, a.title, a.summary, a.content, a.featured_image,
			   a.author_id, a.category_id, a.status, a.view_count, a.published_at, a.created_at, a.updated_at,
			   au.id, au.name, au.slug, au.bio, au.avatar, au.email,
			   c.id, c.name, c.slug, c.description
		FROM articles a
		LEFT JOIN authors au ON a.author_id = au.id AND au.deleted_at IS NULL
		LEFT JOIN categories c ON a.category_id = c.id AND c.deleted_at IS NULL
		WHERE a.slug = $1 AND a.deleted_at IS NULL
	`

	article := &models.Article{}
	var authorID, categoryID *uuid.UUID
	var authorName, authorSlug, authorBio, authorAvatar, authorEmail *string
	var categoryName, categorySlug, categoryDescription *string

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&article.ID, &article.Slug, &article.Title, &article.Summary, &article.Content, &article.FeaturedImage,
		&article.AuthorID, &article.CategoryID, &article.Status, &article.ViewCount, &article.PublishedAt, &article.CreatedAt, &article.UpdatedAt,
		&authorID, &authorName, &authorSlug, &authorBio, &authorAvatar, &authorEmail,
		&categoryID, &categoryName, &categorySlug, &categoryDescription,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article by slug: %w", err)
	}

	if authorID != nil {
		article.Author = &models.Author{
			ID:     *authorID,
			Name:   *authorName,
			Slug:   *authorSlug,
			Bio:    authorBio,
			Avatar: authorAvatar,
			Email:  authorEmail,
		}
	}
	if categoryID != nil {
		article.Category = &models.Category{
			ID:          *categoryID,
			Name:        *categoryName,
			Slug:        *categorySlug,
			Description: categoryDescription,
		}
	}

	tags, err := r.GetArticleTags(ctx, article.ID)
	if err != nil {
		return nil, err
	}
	article.Tags = tags

	return article, nil
}

func (r *ArticleRepository) List(ctx context.Context, filter *models.ArticleFilter, page, perPage int) (*models.PaginatedArticles, error) {
	whereClause := []string{"a.deleted_at IS NULL"}
	args := []interface{}{}
	argNum := 1

	if filter != nil {
		if filter.Status != nil {
			whereClause = append(whereClause, fmt.Sprintf("a.status = $%d", argNum))
			args = append(args, *filter.Status)
			argNum++
		}
		if filter.CategoryID != nil {
			whereClause = append(whereClause, fmt.Sprintf("a.category_id = $%d", argNum))
			args = append(args, *filter.CategoryID)
			argNum++
		}
		if filter.AuthorID != nil {
			whereClause = append(whereClause, fmt.Sprintf("a.author_id = $%d", argNum))
			args = append(args, *filter.AuthorID)
			argNum++
		}
		if filter.TagID != nil {
			whereClause = append(whereClause, fmt.Sprintf("EXISTS (SELECT 1 FROM article_tags at WHERE at.article_id = a.id AND at.tag_id = $%d)", argNum))
			args = append(args, *filter.TagID)
			argNum++
		}
		if filter.Search != nil && *filter.Search != "" {
			whereClause = append(whereClause, fmt.Sprintf("to_tsvector('english', a.title || ' ' || COALESCE(a.summary, '') || ' ' || a.content) @@ plainto_tsquery('english', $%d)", argNum))
			args = append(args, *filter.Search)
			argNum++
		}
		if filter.IncludeDeleted {
			whereClause[0] = "1=1"
		}
	}

	where := strings.Join(whereClause, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM articles a WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count articles: %w", err)
	}

	// Get articles
	offset := (page - 1) * perPage
	args = append(args, perPage, offset)

	query := fmt.Sprintf(`
		SELECT a.id, a.slug, a.title, a.summary, a.featured_image, a.status, a.view_count, a.published_at, a.created_at,
			   au.name, c.name, c.slug
		FROM articles a
		LEFT JOIN authors au ON a.author_id = au.id
		LEFT JOIN categories c ON a.category_id = c.id
		WHERE %s
		ORDER BY a.published_at DESC NULLS LAST, a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argNum, argNum+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list articles: %w", err)
	}
	defer rows.Close()

	articles := []models.ArticleListItem{}
	for rows.Next() {
		var article models.ArticleListItem
		err := rows.Scan(
			&article.ID, &article.Slug, &article.Title, &article.Summary, &article.FeaturedImage,
			&article.Status, &article.ViewCount, &article.PublishedAt, &article.CreatedAt,
			&article.AuthorName, &article.CategoryName, &article.CategorySlug,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedArticles{
		Articles:   articles,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *ArticleRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClauses := []string{}
	args := []interface{}{}
	argNum := 1

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, argNum))
		args = append(args, value)
		argNum++
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE articles SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argNum)

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update article: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("article not found")
	}

	return nil
}

func (r *ArticleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE articles SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("article not found")
	}

	return nil
}

func (r *ArticleRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE articles SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore article: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("article not found or not deleted")
	}

	return nil
}

func (r *ArticleRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM articles WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete article: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("article not found")
	}

	return nil
}

func (r *ArticleRepository) GetArticleTags(ctx context.Context, articleID uuid.UUID) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.slug
		FROM tags t
		JOIN article_tags at ON t.id = at.tag_id
		WHERE at.article_id = $1
	`

	rows, err := r.db.Query(ctx, query, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get article tags: %w", err)
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *ArticleRepository) SetArticleTags(ctx context.Context, articleID uuid.UUID, tagIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Delete existing tags
	_, err = tx.Exec(ctx, "DELETE FROM article_tags WHERE article_id = $1", articleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Insert new tags
	for _, tagID := range tagIDs {
		_, err = tx.Exec(ctx, "INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2)", articleID, tagID)
		if err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *ArticleRepository) GetTrendingIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	query := `
		SELECT id FROM articles
		WHERE status = 'published' AND deleted_at IS NULL
		ORDER BY view_count DESC, published_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending articles: %w", err)
	}
	defer rows.Close()

	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *ArticleRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.ArticleListItem, error) {
	if len(ids) == 0 {
		return []models.ArticleListItem{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT a.id, a.slug, a.title, a.summary, a.featured_image, a.status, a.view_count, a.published_at, a.created_at,
			   au.name, c.name, c.slug
		FROM articles a
		LEFT JOIN authors au ON a.author_id = au.id AND au.deleted_at IS NULL
		LEFT JOIN categories c ON a.category_id = c.id AND c.deleted_at IS NULL
		WHERE a.id IN (%s) AND a.deleted_at IS NULL
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles by ids: %w", err)
	}
	defer rows.Close()

	articlesMap := make(map[uuid.UUID]models.ArticleListItem)
	for rows.Next() {
		var article models.ArticleListItem
		err := rows.Scan(
			&article.ID, &article.Slug, &article.Title, &article.Summary, &article.FeaturedImage,
			&article.Status, &article.ViewCount, &article.PublishedAt, &article.CreatedAt,
			&article.AuthorName, &article.CategoryName, &article.CategorySlug,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articlesMap[article.ID] = article
	}

	// Preserve order from input IDs
	articles := make([]models.ArticleListItem, 0, len(ids))
	for _, id := range ids {
		if article, ok := articlesMap[id]; ok {
			articles = append(articles, article)
		}
	}

	return articles, nil
}

func (r *ArticleRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE articles SET view_count = view_count + 1 WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}
	return nil
}

func (r *ArticleRepository) IncrementViewCountBySlug(ctx context.Context, slug string) error {
	query := "UPDATE articles SET view_count = view_count + 1 WHERE slug = $1 AND status = 'published'"
	_, err := r.db.Exec(ctx, query, slug)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}
	return nil
}
