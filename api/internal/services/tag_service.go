package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type TagService struct {
	repo *repository.TagRepository
}

func NewTagService(repo *repository.TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) Create(ctx context.Context, req *models.CreateTagRequest) (*models.Tag, error) {
	tag := &models.Tag{
		Name: req.Name,
		Slug: req.Slug,
	}

	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *TagService) GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TagService) GetBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *TagService) List(ctx context.Context) ([]models.Tag, error) {
	return s.repo.List(ctx)
}

func (s *TagService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateTagRequest) (*models.Tag, error) {
	if err := s.repo.Update(ctx, id, req); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *TagService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
