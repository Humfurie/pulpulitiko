package repository

import (
	"context"
	"fmt"

	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionRepository struct {
	db *pgxpool.Pool
}

func NewPermissionRepository(db *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// List returns all permissions
func (r *PermissionRepository) List(ctx context.Context) ([]models.Permission, error) {
	query := `
		SELECT id, name, slug, description, category, created_at
		FROM permissions
		ORDER BY category, name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
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

// ListGroupedByCategory returns permissions grouped by category
func (r *PermissionRepository) ListGroupedByCategory(ctx context.Context) ([]models.PermissionCategory, error) {
	permissions, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	// Group by category
	categoryMap := make(map[string][]models.Permission)
	categoryOrder := []string{}

	for _, p := range permissions {
		if _, exists := categoryMap[p.Category]; !exists {
			categoryOrder = append(categoryOrder, p.Category)
		}
		categoryMap[p.Category] = append(categoryMap[p.Category], p)
	}

	// Build result maintaining order
	result := make([]models.PermissionCategory, 0, len(categoryOrder))
	for _, cat := range categoryOrder {
		result = append(result, models.PermissionCategory{
			Category:    cat,
			Permissions: categoryMap[cat],
		})
	}

	return result, nil
}

// GetByCategory returns permissions for a specific category
func (r *PermissionRepository) GetByCategory(ctx context.Context, category string) ([]models.Permission, error) {
	query := `
		SELECT id, name, slug, description, category, created_at
		FROM permissions
		WHERE category = $1
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by category: %w", err)
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
