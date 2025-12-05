package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/pkg/cache"
)

type PoliticianService struct {
	repo  *repository.PoliticianRepository
	cache *cache.RedisCache
}

func NewPoliticianService(repo *repository.PoliticianRepository, cache *cache.RedisCache) *PoliticianService {
	return &PoliticianService{
		repo:  repo,
		cache: cache,
	}
}

func (s *PoliticianService) Create(ctx context.Context, req *models.CreatePoliticianRequest) (*models.Politician, error) {
	politician := &models.Politician{
		Name:     req.Name,
		Slug:     req.Slug,
		Photo:    req.Photo,
		Position: req.Position,
		Party:    req.Party,
		ShortBio: req.ShortBio,
	}

	if err := s.repo.Create(ctx, politician); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.invalidateCache(ctx)

	return politician, nil
}

func (s *PoliticianService) GetByID(ctx context.Context, id uuid.UUID) (*models.Politician, error) {
	// Try cache first
	cacheKey := cache.PoliticianKey(id.String())
	var politician models.Politician
	if err := s.cache.Get(ctx, cacheKey, &politician); err == nil {
		return &politician, nil
	}

	// Get from DB
	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	// Cache for 1 hour
	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)

	return result, nil
}

func (s *PoliticianService) GetBySlug(ctx context.Context, slug string) (*models.Politician, error) {
	// Try cache first
	cacheKey := cache.PoliticianSlugKey(slug)
	var politician models.Politician
	if err := s.cache.Get(ctx, cacheKey, &politician); err == nil {
		return &politician, nil
	}

	// Get from DB
	result, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	// Cache for 1 hour
	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)

	return result, nil
}

func (s *PoliticianService) List(ctx context.Context, filter *models.PoliticianFilter, page, perPage int) (*models.PaginatedPoliticians, error) {
	return s.repo.List(ctx, filter, page, perPage)
}

func (s *PoliticianService) ListAll(ctx context.Context) ([]models.Politician, error) {
	// Try cache first
	cacheKey := cache.PoliticiansKey()
	var politicians []models.Politician
	if err := s.cache.Get(ctx, cacheKey, &politicians); err == nil {
		return politicians, nil
	}

	// Get from DB
	result, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	// Cache for 15 minutes
	_ = s.cache.Set(ctx, cacheKey, result, 15*time.Minute)

	return result, nil
}

func (s *PoliticianService) Search(ctx context.Context, query string, limit int) ([]models.Politician, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.Search(ctx, query, limit)
}

func (s *PoliticianService) Update(ctx context.Context, id uuid.UUID, req *models.UpdatePoliticianRequest) (*models.Politician, error) {
	if err := s.repo.Update(ctx, id, req); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.invalidatePoliticianCache(ctx, id)

	return s.repo.GetByID(ctx, id)
}

func (s *PoliticianService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get politician first to invalidate slug cache
	politician, _ := s.repo.GetByID(ctx, id)

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	s.invalidatePoliticianCache(ctx, id)
	if politician != nil {
		_ = s.cache.Delete(ctx, cache.PoliticianSlugKey(politician.Slug))
	}

	return nil
}

func (s *PoliticianService) Restore(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Restore(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateCache(ctx)

	return nil
}

// GetArticleMentionedPoliticians returns the mentioned politicians for an article
func (s *PoliticianService) GetArticleMentionedPoliticians(ctx context.Context, articleID uuid.UUID) ([]models.Politician, error) {
	return s.repo.GetArticleMentionedPoliticians(ctx, articleID)
}

// SetArticleMentionedPoliticians sets the mentioned politicians for an article
func (s *PoliticianService) SetArticleMentionedPoliticians(ctx context.Context, articleID uuid.UUID, politicianIDs []uuid.UUID) error {
	return s.repo.SetArticleMentionedPoliticians(ctx, articleID, politicianIDs)
}

func (s *PoliticianService) invalidatePoliticianCache(ctx context.Context, id uuid.UUID) {
	_ = s.cache.Delete(ctx, cache.PoliticianKey(id.String()))
	_ = s.cache.Delete(ctx, cache.PoliticiansKey())
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixPoliticianList+"*")
}

func (s *PoliticianService) invalidateCache(ctx context.Context) {
	_ = s.cache.Delete(ctx, cache.PoliticiansKey())
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixPolitician+"*")
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixPoliticianList+"*")
}
