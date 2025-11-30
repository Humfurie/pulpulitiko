package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: db}
}

// List returns all roles with permission counts
func (r *RoleRepository) List(ctx context.Context, includeDeleted bool) ([]models.RoleWithPermissionCount, error) {
	query := `
		SELECT r.id, r.name, r.slug, r.description, r.is_system, r.created_at, r.updated_at, r.deleted_at,
			   COUNT(rp.permission_id) as permission_count
		FROM roles r
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
	`
	if !includeDeleted {
		query += " WHERE r.deleted_at IS NULL"
	}
	query += " GROUP BY r.id ORDER BY r.created_at DESC"

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []models.RoleWithPermissionCount
	for rows.Next() {
		var role models.RoleWithPermissionCount
		err := rows.Scan(
			&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsSystem,
			&role.CreatedAt, &role.UpdatedAt, &role.DeletedAt, &role.PermissionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetByID returns a role by ID with all its permissions
func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	query := `
		SELECT id, name, slug, description, is_system, created_at, updated_at, deleted_at
		FROM roles
		WHERE id = $1 AND deleted_at IS NULL
	`

	var role models.Role
	err := r.db.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsSystem,
		&role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Get permissions for this role
	permissions, err := r.GetPermissionsByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}
	role.Permissions = permissions

	return &role, nil
}

// GetBySlug returns a role by slug
func (r *RoleRepository) GetBySlug(ctx context.Context, slug string) (*models.Role, error) {
	query := `
		SELECT id, name, slug, description, is_system, created_at, updated_at, deleted_at
		FROM roles
		WHERE slug = $1 AND deleted_at IS NULL
	`

	var role models.Role
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsSystem,
		&role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Get permissions for this role
	permissions, err := r.GetPermissionsByRoleID(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	role.Permissions = permissions

	return &role, nil
}

// GetPermissionsByRoleID returns all permissions for a role
func (r *RoleRepository) GetPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.description, p.category, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.name
	`

	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var p models.Permission
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.Category, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, p)
	}

	return permissions, nil
}

// GetPermissionSlugsByRoleID returns only permission slugs for a role (for middleware checks)
func (r *RoleRepository) GetPermissionSlugsByRoleID(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	query := `
		SELECT p.slug
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`

	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission slugs: %w", err)
	}
	defer rows.Close()

	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, fmt.Errorf("failed to scan slug: %w", err)
		}
		slugs = append(slugs, slug)
	}

	return slugs, nil
}

// Create creates a new role with permissions
func (r *RoleRepository) Create(ctx context.Context, req *models.CreateRoleRequest) (*models.Role, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Insert role
	query := `
		INSERT INTO roles (name, slug, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, description, is_system, created_at, updated_at, deleted_at
	`

	var role models.Role
	err = tx.QueryRow(ctx, query, req.Name, req.Slug, req.Description).Scan(
		&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsSystem,
		&role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Add permissions
	if len(req.PermissionIDs) > 0 {
		for _, permIDStr := range req.PermissionIDs {
			permID, err := uuid.Parse(permIDStr)
			if err != nil {
				continue // Skip invalid UUIDs
			}
			_, err = tx.Exec(ctx,
				"INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				role.ID, permID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to add permission: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Fetch permissions
	role.Permissions, _ = r.GetPermissionsByRoleID(ctx, role.ID)

	return &role, nil
}

// Update updates a role and its permissions
func (r *RoleRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateRoleRequest) (*models.Role, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Get existing role
	var role models.Role
	err = tx.QueryRow(ctx,
		"SELECT id, name, slug, description, is_system, created_at, updated_at FROM roles WHERE id = $1 AND deleted_at IS NULL",
		id,
	).Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Update fields
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Slug != nil {
		role.Slug = *req.Slug
	}
	if req.Description != nil {
		role.Description = req.Description
	}

	// Update role
	_, err = tx.Exec(ctx,
		"UPDATE roles SET name = $1, slug = $2, description = $3, updated_at = NOW() WHERE id = $4",
		role.Name, role.Slug, role.Description, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Update permissions if provided
	if req.PermissionIDs != nil {
		// Remove all existing permissions
		_, err = tx.Exec(ctx, "DELETE FROM role_permissions WHERE role_id = $1", id)
		if err != nil {
			return nil, fmt.Errorf("failed to clear permissions: %w", err)
		}

		// Add new permissions
		for _, permIDStr := range req.PermissionIDs {
			permID, err := uuid.Parse(permIDStr)
			if err != nil {
				continue
			}
			_, err = tx.Exec(ctx,
				"INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				id, permID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to add permission: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Fetch updated role
	return r.GetByID(ctx, id)
}

// Delete soft deletes a role (only non-system roles)
func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if system role
	var isSystem bool
	err := r.db.QueryRow(ctx, "SELECT is_system FROM roles WHERE id = $1", id).Scan(&isSystem)
	if err != nil {
		return fmt.Errorf("failed to check role: %w", err)
	}
	if isSystem {
		return fmt.Errorf("cannot delete system role")
	}

	result, err := r.db.Exec(ctx,
		"UPDATE roles SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// Restore restores a soft-deleted role
func (r *RoleRepository) Restore(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		"UPDATE roles SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to restore role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role not found or not deleted")
	}

	return nil
}

// HardDelete permanently deletes a role
func (r *RoleRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	// Check if system role
	var isSystem bool
	err := r.db.QueryRow(ctx, "SELECT is_system FROM roles WHERE id = $1", id).Scan(&isSystem)
	if err != nil {
		return fmt.Errorf("failed to check role: %w", err)
	}
	if isSystem {
		return fmt.Errorf("cannot delete system role")
	}

	_, err = r.db.Exec(ctx, "DELETE FROM roles WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to hard delete role: %w", err)
	}

	return nil
}
