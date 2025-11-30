package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type AuthorService struct {
	repo *repository.AuthorRepository
}

func NewAuthorService(repo *repository.AuthorRepository) *AuthorService {
	return &AuthorService{repo: repo}
}

// Public methods

func (s *AuthorService) List(ctx context.Context) ([]models.Author, error) {
	return s.repo.List(ctx)
}

func (s *AuthorService) GetBySlug(ctx context.Context, slug string) (*models.Author, error) {
	return s.repo.GetBySlug(ctx, slug)
}

// Admin methods

func (s *AuthorService) Create(ctx context.Context, req *models.CreateAuthorRequest) (*models.Author, error) {
	author := &models.Author{
		Name:        req.Name,
		Slug:        req.Slug,
		Bio:         req.Bio,
		Avatar:      req.Avatar,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		SocialLinks: req.SocialLinks,
	}

	// Handle role_id - either from direct role_id or by looking up role slug
	if req.RoleID != nil && *req.RoleID != "" {
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			return nil, err
		}
		author.RoleID = &roleID
	} else if req.Role != nil && *req.Role != "" {
		// Look up role by slug if role_id not provided
		roleID, err := s.repo.GetRoleIDBySlug(ctx, *req.Role)
		if err != nil {
			return nil, err
		}
		author.RoleID = roleID
	} else {
		// Default to "author" role
		roleID, _ := s.repo.GetRoleIDBySlug(ctx, "author")
		author.RoleID = roleID
	}

	if err := s.repo.Create(ctx, author); err != nil {
		return nil, err
	}

	return author, nil
}

func (s *AuthorService) GetByID(ctx context.Context, id uuid.UUID) (*models.Author, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AuthorService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateAuthorRequest) (*models.Author, error) {
	if err := s.repo.Update(ctx, id, req); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *AuthorService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *AuthorService) Restore(ctx context.Context, id uuid.UUID) error {
	return s.repo.Restore(ctx, id)
}

func (s *AuthorService) GetByEmail(ctx context.Context, email string) (*models.Author, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *AuthorService) UpdateByEmail(ctx context.Context, email string, req *models.UpdateAuthorRequest) (*models.Author, error) {
	if err := s.repo.UpdateByEmail(ctx, email, req); err != nil {
		return nil, err
	}
	return s.repo.GetByEmail(ctx, email)
}
