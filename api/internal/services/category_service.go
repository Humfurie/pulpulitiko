package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/pkg/cache"
)

const CategoryCacheTTL = 30 * time.Minute

type CategoryService struct {
	repo  *repository.CategoryRepository
	cache *cache.RedisCache
}

func NewCategoryService(repo *repository.CategoryRepository, cache *cache.RedisCache) *CategoryService {
	return &CategoryService{
		repo:  repo,
		cache: cache,
	}
}

func (s *CategoryService) Create(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error) {
	category := &models.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, err
	}

	s.cache.Delete(ctx, cache.CategoriesKey())

	return category, nil
}

func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	cacheKey := cache.CategoryKey(id.String())

	var category models.Category
	if err := s.cache.Get(ctx, cacheKey, &category); err == nil {
		return &category, nil
	}

	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	s.cache.Set(ctx, cacheKey, result, CategoryCacheTTL)

	return result, nil
}

func (s *CategoryService) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *CategoryService) List(ctx context.Context) ([]models.Category, error) {
	cacheKey := cache.CategoriesKey()

	var categories []models.Category
	if err := s.cache.Get(ctx, cacheKey, &categories); err == nil {
		return categories, nil
	}

	result, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, result, CategoryCacheTTL)

	return result, nil
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateCategoryRequest) (*models.Category, error) {
	if err := s.repo.Update(ctx, id, req); err != nil {
		return nil, err
	}

	s.cache.Delete(ctx, cache.CategoryKey(id.String()))
	s.cache.Delete(ctx, cache.CategoriesKey())

	return s.repo.GetByID(ctx, id)
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.cache.Delete(ctx, cache.CategoryKey(id.String()))
	s.cache.Delete(ctx, cache.CategoriesKey())

	return nil
}
