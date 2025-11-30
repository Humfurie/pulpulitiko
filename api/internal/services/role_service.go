package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type RoleService struct {
	roleRepo       *repository.RoleRepository
	permissionRepo *repository.PermissionRepository
}

func NewRoleService(roleRepo *repository.RoleRepository, permissionRepo *repository.PermissionRepository) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// ListRoles returns all roles with permission counts
func (s *RoleService) ListRoles(ctx context.Context, includeDeleted bool) ([]models.RoleWithPermissionCount, error) {
	return s.roleRepo.List(ctx, includeDeleted)
}

// GetRoleByID returns a role by ID with its permissions
func (s *RoleService) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

// GetRoleBySlug returns a role by slug
func (s *RoleService) GetRoleBySlug(ctx context.Context, slug string) (*models.Role, error) {
	role, err := s.roleRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, req *models.CreateRoleRequest) (*models.Role, error) {
	// Check for duplicate slug
	existing, _ := s.roleRepo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, fmt.Errorf("role with slug '%s' already exists", req.Slug)
	}

	return s.roleRepo.Create(ctx, req)
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(ctx context.Context, id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error) {
	// Check if role exists
	existing, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("role not found")
	}

	// Check for duplicate slug if being changed
	if req.Slug != nil && *req.Slug != existing.Slug {
		other, _ := s.roleRepo.GetBySlug(ctx, *req.Slug)
		if other != nil {
			return nil, fmt.Errorf("role with slug '%s' already exists", *req.Slug)
		}
	}

	return s.roleRepo.Update(ctx, id, req)
}

// DeleteRole soft deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.roleRepo.Delete(ctx, id)
}

// RestoreRole restores a soft-deleted role
func (s *RoleService) RestoreRole(ctx context.Context, id uuid.UUID) error {
	return s.roleRepo.Restore(ctx, id)
}

// GetPermissionsByRoleID returns permissions for a role
func (s *RoleService) GetPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error) {
	return s.roleRepo.GetPermissionsByRoleID(ctx, roleID)
}

// GetPermissionSlugsByRoleID returns permission slugs for middleware checks
func (s *RoleService) GetPermissionSlugsByRoleID(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	return s.roleRepo.GetPermissionSlugsByRoleID(ctx, roleID)
}

// HasPermission checks if a role has a specific permission
func (s *RoleService) HasPermission(ctx context.Context, roleID uuid.UUID, permissionSlug string) (bool, error) {
	slugs, err := s.roleRepo.GetPermissionSlugsByRoleID(ctx, roleID)
	if err != nil {
		return false, err
	}

	for _, slug := range slugs {
		if slug == permissionSlug {
			return true, nil
		}
	}
	return false, nil
}

// ListPermissions returns all permissions
func (s *RoleService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.permissionRepo.List(ctx)
}

// ListPermissionsGrouped returns permissions grouped by category
func (s *RoleService) ListPermissionsGrouped(ctx context.Context) ([]models.PermissionCategory, error) {
	return s.permissionRepo.ListGroupedByCategory(ctx)
}
