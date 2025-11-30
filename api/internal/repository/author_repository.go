package repository

import (
	"context"
	"encoding/json"
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
	socialLinksJSON, err := json.Marshal(author.SocialLinks)
	if err != nil {
		return fmt.Errorf("failed to marshal social links: %w", err)
	}

	query := `
		INSERT INTO authors (name, slug, bio, avatar, email, phone, address, social_links, role_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		author.Name,
		author.Slug,
		author.Bio,
		author.Avatar,
		author.Email,
		author.Phone,
		author.Address,
		socialLinksJSON,
		author.RoleID,
	).Scan(&author.ID, &author.CreatedAt, &author.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create author: %w", err)
	}

	return nil
}

func (r *AuthorRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Author, error) {
	query := `
		SELECT a.id, a.name, a.slug, a.bio, a.avatar, a.email, a.phone, a.address, a.social_links,
		       a.role_id, COALESCE(r.slug, '') as role_slug, a.created_at, a.updated_at, a.deleted_at
		FROM authors a
		LEFT JOIN roles r ON a.role_id = r.id
		WHERE a.id = $1 AND a.deleted_at IS NULL
	`

	author := &models.Author{}
	var socialLinksJSON []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&author.ID, &author.Name, &author.Slug, &author.Bio, &author.Avatar,
		&author.Email, &author.Phone, &author.Address, &socialLinksJSON,
		&author.RoleID, &author.Role, &author.CreatedAt, &author.UpdatedAt, &author.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	if len(socialLinksJSON) > 0 {
		var socialLinks models.SocialLinks
		if err := json.Unmarshal(socialLinksJSON, &socialLinks); err == nil {
			author.SocialLinks = &socialLinks
		}
	}

	return author, nil
}

func (r *AuthorRepository) GetBySlug(ctx context.Context, slug string) (*models.Author, error) {
	query := `
		SELECT a.id, a.name, a.slug, a.bio, a.avatar, a.email, a.phone, a.address, a.social_links,
		       a.role_id, COALESCE(r.slug, '') as role_slug, a.created_at, a.updated_at, a.deleted_at
		FROM authors a
		LEFT JOIN roles r ON a.role_id = r.id
		WHERE a.slug = $1 AND a.deleted_at IS NULL
	`

	author := &models.Author{}
	var socialLinksJSON []byte
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&author.ID, &author.Name, &author.Slug, &author.Bio, &author.Avatar,
		&author.Email, &author.Phone, &author.Address, &socialLinksJSON,
		&author.RoleID, &author.Role, &author.CreatedAt, &author.UpdatedAt, &author.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get author by slug: %w", err)
	}

	if len(socialLinksJSON) > 0 {
		var socialLinks models.SocialLinks
		if err := json.Unmarshal(socialLinksJSON, &socialLinks); err == nil {
			author.SocialLinks = &socialLinks
		}
	}

	return author, nil
}

func (r *AuthorRepository) List(ctx context.Context) ([]models.Author, error) {
	query := `
		SELECT a.id, a.name, a.slug, a.bio, a.avatar, a.email, a.phone, a.address, a.social_links,
		       a.role_id, COALESCE(r.slug, '') as role_slug, a.created_at, a.updated_at, a.deleted_at
		FROM authors a
		LEFT JOIN roles r ON a.role_id = r.id
		WHERE a.deleted_at IS NULL
		ORDER BY a.name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list authors: %w", err)
	}
	defer rows.Close()

	authors := []models.Author{}
	for rows.Next() {
		var author models.Author
		var socialLinksJSON []byte
		err := rows.Scan(
			&author.ID, &author.Name, &author.Slug, &author.Bio, &author.Avatar,
			&author.Email, &author.Phone, &author.Address, &socialLinksJSON,
			&author.RoleID, &author.Role, &author.CreatedAt, &author.UpdatedAt, &author.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan author: %w", err)
		}

		if len(socialLinksJSON) > 0 {
			var socialLinks models.SocialLinks
			if err := json.Unmarshal(socialLinksJSON, &socialLinks); err == nil {
				author.SocialLinks = &socialLinks
			}
		}

		authors = append(authors, author)
	}

	return authors, nil
}

func (r *AuthorRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateAuthorRequest) error {
	var socialLinksJSON []byte
	if req.SocialLinks != nil {
		var err error
		socialLinksJSON, err = json.Marshal(req.SocialLinks)
		if err != nil {
			return fmt.Errorf("failed to marshal social links: %w", err)
		}
	}

	// Parse role_id if provided
	var roleID *uuid.UUID
	if req.RoleID != nil && *req.RoleID != "" {
		parsed, err := uuid.Parse(*req.RoleID)
		if err != nil {
			return fmt.Errorf("invalid role_id: %w", err)
		}
		roleID = &parsed
	}

	query := `
		UPDATE authors
		SET name = COALESCE($1, name),
			slug = COALESCE($2, slug),
			bio = COALESCE($3, bio),
			avatar = COALESCE($4, avatar),
			email = COALESCE($5, email),
			phone = COALESCE($6, phone),
			address = COALESCE($7, address),
			social_links = COALESCE($8, social_links),
			role_id = COALESCE($9, role_id)
		WHERE id = $10 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		req.Name, req.Slug, req.Bio, req.Avatar, req.Email,
		req.Phone, req.Address, socialLinksJSON, roleID, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update author: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("author not found")
	}

	return nil
}

func (r *AuthorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE authors SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete author: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("author not found")
	}

	return nil
}

func (r *AuthorRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE authors SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore author: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("author not found or not deleted")
	}

	return nil
}

func (r *AuthorRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM authors WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete author: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("author not found")
	}

	return nil
}

// GetRoleIDBySlug returns the role ID for a given role slug
func (r *AuthorRepository) GetRoleIDBySlug(ctx context.Context, slug string) (*uuid.UUID, error) {
	query := "SELECT id FROM roles WHERE slug = $1 AND deleted_at IS NULL"

	var roleID uuid.UUID
	err := r.db.QueryRow(ctx, query, slug).Scan(&roleID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &roleID, nil
}
