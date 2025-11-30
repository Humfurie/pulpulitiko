package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(db *pgxpool.Pool) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(ctx context.Context, tag *models.Tag) error {
	query := `
		INSERT INTO tags (name, slug)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, tag.Name, tag.Slug).Scan(&tag.ID, &tag.CreatedAt, &tag.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM tags
		WHERE id = $1 AND deleted_at IS NULL
	`

	tag := &models.Tag{}
	err := r.db.QueryRow(ctx, query, id).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt, &tag.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

func (r *TagRepository) GetBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM tags
		WHERE slug = $1 AND deleted_at IS NULL
	`

	tag := &models.Tag{}
	err := r.db.QueryRow(ctx, query, slug).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt, &tag.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag by slug: %w", err)
	}

	return tag, nil
}

func (r *TagRepository) List(ctx context.Context) ([]models.Tag, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM tags
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt, &tag.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *TagRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateTagRequest) error {
	query := `
		UPDATE tags
		SET name = COALESCE($1, name),
			slug = COALESCE($2, slug)
		WHERE id = $3
	`

	result, err := r.db.Exec(ctx, query, req.Name, req.Slug, id)
	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE tags SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

func (r *TagRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE tags SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found or not deleted")
	}

	return nil
}

func (r *TagRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM tags WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}
