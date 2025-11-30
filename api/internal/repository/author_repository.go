package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthorRepository struct {
	db *pgxpool.Pool
}

func NewAuthorRepository(db *pgxpool.Pool) *AuthorRepository {
	return &AuthorRepository{db: db}
}

func (r *AuthorRepository) Create(ctx context.Context, author *models.Author) error {
	query := `
		INSERT INTO authors (name, slug, bio, avatar, email)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		author.Name,
		author.Slug,
		author.Bio,
		author.Avatar,
		author.Email,
	).Scan(&author.ID, &author.CreatedAt, &author.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create author: %w", err)
	}

	return nil
}

func (r *AuthorRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Author, error) {
	query := `
		SELECT id, name, slug, bio, avatar, email, created_at, updated_at
		FROM authors
		WHERE id = $1
	`

	author := &models.Author{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&author.ID, &author.Name, &author.Slug, &author.Bio, &author.Avatar,
		&author.Email, &author.CreatedAt, &author.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	return author, nil
}

func (r *AuthorRepository) GetBySlug(ctx context.Context, slug string) (*models.Author, error) {
	query := `
		SELECT id, name, slug, bio, avatar, email, created_at, updated_at
		FROM authors
		WHERE slug = $1
	`

	author := &models.Author{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&author.ID, &author.Name, &author.Slug, &author.Bio, &author.Avatar,
		&author.Email, &author.CreatedAt, &author.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get author by slug: %w", err)
	}

	return author, nil
}

func (r *AuthorRepository) List(ctx context.Context) ([]models.Author, error) {
	query := `
		SELECT id, name, slug, bio, avatar, email, created_at, updated_at
		FROM authors
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list authors: %w", err)
	}
	defer rows.Close()

	authors := []models.Author{}
	for rows.Next() {
		var author models.Author
		err := rows.Scan(
			&author.ID, &author.Name, &author.Slug, &author.Bio, &author.Avatar,
			&author.Email, &author.CreatedAt, &author.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan author: %w", err)
		}
		authors = append(authors, author)
	}

	return authors, nil
}

func (r *AuthorRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateAuthorRequest) error {
	query := `
		UPDATE authors
		SET name = COALESCE($1, name),
			slug = COALESCE($2, slug),
			bio = COALESCE($3, bio),
			avatar = COALESCE($4, avatar),
			email = COALESCE($5, email)
		WHERE id = $6
	`

	result, err := r.db.Exec(ctx, query, req.Name, req.Slug, req.Bio, req.Avatar, req.Email, id)
	if err != nil {
		return fmt.Errorf("failed to update author: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("author not found")
	}

	return nil
}

func (r *AuthorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM authors WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete author: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("author not found")
	}

	return nil
}
