package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoliticianRepository struct {
	db *pgxpool.Pool
}

func NewPoliticianRepository(db *pgxpool.Pool) *PoliticianRepository {
	return &PoliticianRepository{db: db}
}

func (r *PoliticianRepository) Create(ctx context.Context, politician *models.Politician) error {
	query := `
		INSERT INTO politicians (name, slug, photo, position, party, short_bio, term_start, term_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		politician.Name,
		politician.Slug,
		politician.Photo,
		politician.Position,
		politician.Party,
		politician.ShortBio,
		politician.TermStart,
		politician.TermEnd,
	).Scan(&politician.ID, &politician.CreatedAt, &politician.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create politician: %w", err)
	}

	return nil
}

func (r *PoliticianRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Politician, error) {
	query := `
		SELECT id, name, slug, photo, position, party, short_bio, term_start, term_end, created_at, updated_at, deleted_at
		FROM politicians
		WHERE id = $1 AND deleted_at IS NULL
	`

	politician := &models.Politician{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&politician.ID,
		&politician.Name,
		&politician.Slug,
		&politician.Photo,
		&politician.Position,
		&politician.Party,
		&politician.ShortBio,
		&politician.TermStart,
		&politician.TermEnd,
		&politician.CreatedAt,
		&politician.UpdatedAt,
		&politician.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get politician: %w", err)
	}

	return politician, nil
}

func (r *PoliticianRepository) GetBySlug(ctx context.Context, slug string) (*models.Politician, error) {
	query := `
		SELECT id, name, slug, photo, position, party, short_bio, term_start, term_end, created_at, updated_at, deleted_at
		FROM politicians
		WHERE slug = $1 AND deleted_at IS NULL
	`

	politician := &models.Politician{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&politician.ID,
		&politician.Name,
		&politician.Slug,
		&politician.Photo,
		&politician.Position,
		&politician.Party,
		&politician.ShortBio,
		&politician.TermStart,
		&politician.TermEnd,
		&politician.CreatedAt,
		&politician.UpdatedAt,
		&politician.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get politician by slug: %w", err)
	}

	return politician, nil
}

func (r *PoliticianRepository) List(ctx context.Context, filter *models.PoliticianFilter, page, perPage int) (*models.PaginatedPoliticians, error) {
	// Build base query with article count and party info
	baseQuery := `
		SELECT p.id, p.name, p.slug, p.photo, p.position, p.party, p.term_start, p.term_end,
			(SELECT COUNT(*) FROM articles a WHERE a.primary_politician_id = p.id AND a.deleted_at IS NULL) +
			(SELECT COUNT(*) FROM article_politicians ap JOIN articles a ON ap.article_id = a.id WHERE ap.politician_id = p.id AND a.deleted_at IS NULL) as article_count,
			pp.id, pp.name, pp.slug, pp.abbreviation, pp.logo, pp.color
		FROM politicians p
		LEFT JOIN political_parties pp ON p.party_id = pp.id
		WHERE p.deleted_at IS NULL
	`
	countQuery := "SELECT COUNT(*) FROM politicians p WHERE p.deleted_at IS NULL"

	args := []interface{}{}
	argNum := 1
	conditions := []string{}

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.position ILIKE $%d OR p.party ILIKE $%d)", argNum, argNum, argNum))
			args = append(args, "%"+*filter.Search+"%")
			argNum++
		}

		if filter.Party != nil && *filter.Party != "" {
			conditions = append(conditions, fmt.Sprintf("p.party = $%d", argNum))
			args = append(args, *filter.Party)
			argNum++
		}

		if filter.PartyID != nil {
			conditions = append(conditions, fmt.Sprintf("p.party_id = $%d", argNum))
			args = append(args, *filter.PartyID)
			argNum++
		}

		if filter.IncludeDeleted {
			// Remove the deleted_at IS NULL condition
			baseQuery = strings.Replace(baseQuery, "WHERE p.deleted_at IS NULL", "WHERE 1=1", 1)
			countQuery = strings.Replace(countQuery, "WHERE p.deleted_at IS NULL", "WHERE 1=1", 1)
		}
	}

	if len(conditions) > 0 {
		condStr := " AND " + strings.Join(conditions, " AND ")
		baseQuery += condStr
		countQuery += condStr
	}

	// Get total count
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count politicians: %w", err)
	}

	// Add pagination and ordering
	offset := (page - 1) * perPage
	baseQuery += fmt.Sprintf(" ORDER BY p.name ASC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, perPage, offset)

	// Execute main query
	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list politicians: %w", err)
	}
	defer rows.Close()

	politicians := []models.PoliticianListItem{}
	for rows.Next() {
		var p models.PoliticianListItem
		var partyID *uuid.UUID
		var partyName, partySlug *string
		var partyAbbr, partyLogo, partyColor *string

		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Photo, &p.Position, &p.Party,
			&p.TermStart, &p.TermEnd, &p.ArticleCount,
			&partyID, &partyName, &partySlug, &partyAbbr, &partyLogo, &partyColor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan politician: %w", err)
		}

		// Populate PartyInfo if party exists
		if partyID != nil && partyName != nil && partySlug != nil {
			p.PartyInfo = &models.PartyBrief{
				ID:           *partyID,
				Name:         *partyName,
				Slug:         *partySlug,
				Abbreviation: partyAbbr,
				Logo:         partyLogo,
				Color:        partyColor,
			}
		}

		politicians = append(politicians, p)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedPoliticians{
		Politicians: politicians,
		Total:       total,
		Page:        page,
		PerPage:     perPage,
		TotalPages:  totalPages,
	}, nil
}

// ListAll returns all politicians without pagination (for dropdowns/selects)
func (r *PoliticianRepository) ListAll(ctx context.Context) ([]models.Politician, error) {
	query := `
		SELECT id, name, slug, photo, position, party, short_bio, term_start, term_end, created_at, updated_at
		FROM politicians
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list politicians: %w", err)
	}
	defer rows.Close()

	politicians := []models.Politician{}
	for rows.Next() {
		var p models.Politician
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Photo, &p.Position, &p.Party, &p.ShortBio, &p.TermStart, &p.TermEnd, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan politician: %w", err)
		}
		politicians = append(politicians, p)
	}

	return politicians, nil
}

// Search returns politicians matching the query (for autocomplete)
func (r *PoliticianRepository) Search(ctx context.Context, query string, limit int) ([]models.Politician, error) {
	sqlQuery := `
		SELECT id, name, slug, photo, position, party, short_bio, term_start, term_end, created_at, updated_at
		FROM politicians
		WHERE deleted_at IS NULL AND (name ILIKE $1 OR position ILIKE $1 OR party ILIKE $1)
		ORDER BY name ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search politicians: %w", err)
	}
	defer rows.Close()

	politicians := []models.Politician{}
	for rows.Next() {
		var p models.Politician
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Photo, &p.Position, &p.Party, &p.ShortBio, &p.TermStart, &p.TermEnd, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan politician: %w", err)
		}
		politicians = append(politicians, p)
	}

	return politicians, nil
}

func (r *PoliticianRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdatePoliticianRequest) error {
	// Parse term dates if provided
	var termStart, termEnd interface{}
	if req.TermStart != nil {
		termStart = *req.TermStart
	}
	if req.TermEnd != nil {
		termEnd = *req.TermEnd
	}

	query := `
		UPDATE politicians
		SET name = COALESCE($1, name),
			slug = COALESCE($2, slug),
			photo = COALESCE($3, photo),
			position = COALESCE($4, position),
			party = COALESCE($5, party),
			short_bio = COALESCE($6, short_bio),
			term_start = COALESCE($7::date, term_start),
			term_end = COALESCE($8::date, term_end),
			updated_at = NOW()
		WHERE id = $9 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		req.Name,
		req.Slug,
		req.Photo,
		req.Position,
		req.Party,
		req.ShortBio,
		termStart,
		termEnd,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update politician: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("politician not found")
	}

	return nil
}

func (r *PoliticianRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE politicians SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete politician: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("politician not found")
	}

	return nil
}

func (r *PoliticianRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE politicians SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore politician: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("politician not found or not deleted")
	}

	return nil
}

func (r *PoliticianRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM politicians WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete politician: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("politician not found")
	}

	return nil
}

// GetArticleMentionedPoliticians returns the mentioned politicians for an article
func (r *PoliticianRepository) GetArticleMentionedPoliticians(ctx context.Context, articleID uuid.UUID) ([]models.Politician, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.photo, p.position, p.party, p.short_bio, p.created_at, p.updated_at
		FROM politicians p
		JOIN article_politicians ap ON p.id = ap.politician_id
		WHERE ap.article_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.name ASC
	`

	rows, err := r.db.Query(ctx, query, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get article mentioned politicians: %w", err)
	}
	defer rows.Close()

	politicians := []models.Politician{}
	for rows.Next() {
		var p models.Politician
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Photo, &p.Position, &p.Party, &p.ShortBio, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan politician: %w", err)
		}
		politicians = append(politicians, p)
	}

	return politicians, nil
}

// SetArticleMentionedPoliticians sets the mentioned politicians for an article
func (r *PoliticianRepository) SetArticleMentionedPoliticians(ctx context.Context, articleID uuid.UUID, politicianIDs []uuid.UUID) error {
	// Delete existing associations
	_, err := r.db.Exec(ctx, "DELETE FROM article_politicians WHERE article_id = $1", articleID)
	if err != nil {
		return fmt.Errorf("failed to clear article politicians: %w", err)
	}

	// Insert new associations
	if len(politicianIDs) > 0 {
		for _, politicianID := range politicianIDs {
			_, err := r.db.Exec(ctx,
				"INSERT INTO article_politicians (article_id, politician_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				articleID, politicianID,
			)
			if err != nil {
				return fmt.Errorf("failed to add article politician: %w", err)
			}
		}
	}

	return nil
}
