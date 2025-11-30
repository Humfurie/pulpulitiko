package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (name, slug, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		category.Name,
		category.Slug,
		category.Description,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	category := &models.Category{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Slug, &category.Description,
		&category.CreatedAt, &category.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM categories
		WHERE slug = $1
	`

	category := &models.Category{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&category.ID, &category.Name, &category.Slug, &category.Description,
		&category.CreatedAt, &category.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	return category, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM categories
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID, &category.Name, &category.Slug, &category.Description,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateCategoryRequest) error {
	query := `
		UPDATE categories
		SET name = COALESCE($1, name),
			slug = COALESCE($2, slug),
			description = COALESCE($3, description)
		WHERE id = $4
	`

	result, err := r.db.Exec(ctx, query, req.Name, req.Slug, req.Description, id)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM categories WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
