package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/pkg/cache"
)

const (
	ArticleCacheTTL     = 15 * time.Minute
	ArticleListCacheTTL = 5 * time.Minute
	TrendingCacheTTL    = 10 * time.Minute
)

type ArticleService struct {
	repo           *repository.ArticleRepository
	politicianRepo *repository.PoliticianRepository
	cache          *cache.RedisCache
}

func NewArticleService(repo *repository.ArticleRepository, politicianRepo *repository.PoliticianRepository, cache *cache.RedisCache) *ArticleService {
	return &ArticleService{
		repo:           repo,
		politicianRepo: politicianRepo,
		cache:          cache,
	}
}

func (s *ArticleService) Create(ctx context.Context, req *models.CreateArticleRequest) (*models.Article, error) {
	article := &models.Article{
		Slug:          req.Slug,
		Title:         req.Title,
		Summary:       req.Summary,
		Content:       req.Content,
		FeaturedImage: req.FeaturedImage,
		Status:        models.ArticleStatusDraft,
	}

	if req.Status != "" {
		article.Status = models.ArticleStatus(req.Status)
	}

	if req.AuthorID != nil {
		id, err := uuid.Parse(*req.AuthorID)
		if err != nil {
			return nil, fmt.Errorf("invalid author ID: %w", err)
		}
		article.AuthorID = &id
	}

	if req.CategoryID != nil {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category ID: %w", err)
		}
		article.CategoryID = &id
	}

	if req.PrimaryPoliticianID != nil {
		id, err := uuid.Parse(*req.PrimaryPoliticianID)
		if err != nil {
			return nil, fmt.Errorf("invalid primary politician ID: %w", err)
		}
		article.PrimaryPoliticianID = &id
	}

	if err := s.repo.Create(ctx, article); err != nil {
		return nil, err
	}

	// Set tags if provided
	if len(req.TagIDs) > 0 {
		tagUUIDs := make([]uuid.UUID, len(req.TagIDs))
		for i, tagID := range req.TagIDs {
			id, err := uuid.Parse(tagID)
			if err != nil {
				return nil, fmt.Errorf("invalid tag ID: %w", err)
			}
			tagUUIDs[i] = id
		}
		if err := s.repo.SetArticleTags(ctx, article.ID, tagUUIDs); err != nil {
			return nil, err
		}
	}

	// Set mentioned politicians if provided
	if len(req.PoliticianIDs) > 0 {
		politicianUUIDs := make([]uuid.UUID, len(req.PoliticianIDs))
		for i, politicianID := range req.PoliticianIDs {
			id, err := uuid.Parse(politicianID)
			if err != nil {
				return nil, fmt.Errorf("invalid politician ID: %w", err)
			}
			politicianUUIDs[i] = id
		}
		if err := s.politicianRepo.SetArticleMentionedPoliticians(ctx, article.ID, politicianUUIDs); err != nil {
			return nil, err
		}
	}

	// Invalidate list cache
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixArticleList+"*")

	return s.repo.GetByID(ctx, article.ID)
}

func (s *ArticleService) GetByID(ctx context.Context, id uuid.UUID) (*models.Article, error) {
	cacheKey := cache.ArticleKey(id.String())

	var article models.Article
	if err := s.cache.Get(ctx, cacheKey, &article); err == nil {
		return &article, nil
	}

	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, ArticleCacheTTL)

	return result, nil
}

func (s *ArticleService) GetBySlug(ctx context.Context, slug string) (*models.Article, error) {
	cacheKey := cache.ArticleSlugKey(slug)

	var article models.Article
	if err := s.cache.Get(ctx, cacheKey, &article); err == nil {
		return &article, nil
	}

	result, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, ArticleCacheTTL)

	return result, nil
}

func (s *ArticleService) List(ctx context.Context, filter *models.ArticleFilter, page, perPage int) (*models.PaginatedArticles, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filterHash := hashFilter(filter)
	cacheKey := cache.ArticleListKey(page, perPage, filterHash)

	var result models.PaginatedArticles
	if err := s.cache.Get(ctx, cacheKey, &result); err == nil {
		return &result, nil
	}

	articles, err := s.repo.List(ctx, filter, page, perPage)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, articles, ArticleListCacheTTL)

	return articles, nil
}

func (s *ArticleService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateArticleRequest) (*models.Article, error) {
	updates := make(map[string]interface{})

	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.FeaturedImage != nil {
		updates["featured_image"] = *req.FeaturedImage
	}
	if req.AuthorID != nil {
		authorID, err := uuid.Parse(*req.AuthorID)
		if err != nil {
			return nil, fmt.Errorf("invalid author ID: %w", err)
		}
		updates["author_id"] = authorID
	}
	if req.CategoryID != nil {
		categoryID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category ID: %w", err)
		}
		updates["category_id"] = categoryID
	}
	if req.PrimaryPoliticianID != nil {
		politicianID, err := uuid.Parse(*req.PrimaryPoliticianID)
		if err != nil {
			return nil, fmt.Errorf("invalid primary politician ID: %w", err)
		}
		updates["primary_politician_id"] = politicianID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == string(models.ArticleStatusPublished) {
			updates["published_at"] = time.Now()
		}
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, err
	}

	// Update tags if provided
	if req.TagIDs != nil {
		tagUUIDs := make([]uuid.UUID, len(req.TagIDs))
		for i, tagID := range req.TagIDs {
			tid, err := uuid.Parse(tagID)
			if err != nil {
				return nil, fmt.Errorf("invalid tag ID: %w", err)
			}
			tagUUIDs[i] = tid
		}
		if err := s.repo.SetArticleTags(ctx, id, tagUUIDs); err != nil {
			return nil, err
		}
	}

	// Update mentioned politicians if provided
	if req.PoliticianIDs != nil {
		politicianUUIDs := make([]uuid.UUID, len(req.PoliticianIDs))
		for i, politicianID := range req.PoliticianIDs {
			pid, err := uuid.Parse(politicianID)
			if err != nil {
				return nil, fmt.Errorf("invalid politician ID: %w", err)
			}
			politicianUUIDs[i] = pid
		}
		if err := s.politicianRepo.SetArticleMentionedPoliticians(ctx, id, politicianUUIDs); err != nil {
			return nil, err
		}
	}

	// Invalidate caches
	s.invalidateArticleCache(ctx, id)

	return s.repo.GetByID(ctx, id)
}

func (s *ArticleService) Delete(ctx context.Context, id uuid.UUID) error {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return fmt.Errorf("article not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateArticleCache(ctx, id)
	_ = s.cache.Delete(ctx, cache.ArticleSlugKey(article.Slug))

	return nil
}

func (s *ArticleService) Restore(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Restore(ctx, id); err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateArticleCache(ctx, id)

	return nil
}

func (s *ArticleService) GetTrending(ctx context.Context, limit int) ([]models.ArticleListItem, error) {
	if limit < 1 || limit > 20 {
		limit = 10
	}

	cacheKey := cache.TrendingKey()

	var articles []models.ArticleListItem
	if err := s.cache.Get(ctx, cacheKey, &articles); err == nil {
		if len(articles) > limit {
			return articles[:limit], nil
		}
		return articles, nil
	}

	// For now, trending is based on recent published articles
	// In a real app, this would be based on view counts, shares, etc.
	ids, err := s.repo.GetTrendingIDs(ctx, 20)
	if err != nil {
		return nil, err
	}

	articles, err = s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, articles, TrendingCacheTTL)

	if len(articles) > limit {
		return articles[:limit], nil
	}
	return articles, nil
}

func (s *ArticleService) Search(ctx context.Context, query string, page, perPage int) (*models.PaginatedArticles, error) {
	filter := &models.ArticleFilter{
		Search: &query,
		Status: func() *models.ArticleStatus {
			status := models.ArticleStatusPublished
			return &status
		}(),
	}
	return s.List(ctx, filter, page, perPage)
}

func (s *ArticleService) IncrementViewCount(ctx context.Context, slug string) error {
	return s.repo.IncrementViewCountBySlug(ctx, slug)
}

func (s *ArticleService) GetRelatedArticles(ctx context.Context, articleID uuid.UUID, categoryID *uuid.UUID, tagIDs []uuid.UUID, limit int) ([]models.ArticleListItem, error) {
	if limit < 1 || limit > 10 {
		limit = 4
	}

	cacheKey := fmt.Sprintf("related:%s", articleID.String())

	var articles []models.ArticleListItem
	if err := s.cache.Get(ctx, cacheKey, &articles); err == nil {
		if len(articles) > limit {
			return articles[:limit], nil
		}
		return articles, nil
	}

	articles, err := s.repo.GetRelatedArticles(ctx, articleID, categoryID, tagIDs, 10)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, articles, ArticleCacheTTL)

	if len(articles) > limit {
		return articles[:limit], nil
	}
	return articles, nil
}

func (s *ArticleService) invalidateArticleCache(ctx context.Context, id uuid.UUID) {
	_ = s.cache.Delete(ctx, cache.ArticleKey(id.String()))
	_ = s.cache.Delete(ctx, cache.TrendingKey())
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixArticleList+"*")
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixArticleSlug+"*")
}

func hashFilter(filter *models.ArticleFilter) string {
	if filter == nil {
		return "nil"
	}

	data := fmt.Sprintf("%v:%v:%v:%v:%v:%v",
		filter.Status,
		filter.CategoryID,
		filter.TagID,
		filter.AuthorID,
		filter.PoliticianID,
		filter.Search,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
