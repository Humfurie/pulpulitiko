package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoliticalPartyRepository struct {
	db *pgxpool.Pool
}

func NewPoliticalPartyRepository(db *pgxpool.Pool) *PoliticalPartyRepository {
	return &PoliticalPartyRepository{db: db}
}

// Political Party CRUD

func (r *PoliticalPartyRepository) Create(ctx context.Context, req *models.CreatePoliticalPartyRequest) (*models.PoliticalParty, error) {
	party := &models.PoliticalParty{}

	err := r.db.QueryRow(ctx, `
		INSERT INTO political_parties (name, slug, abbreviation, logo, color, description, founded_year, website, is_major, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, name, slug, abbreviation, logo, color, description, founded_year, website, is_major, is_active, created_at, updated_at
	`, req.Name, req.Slug, req.Abbreviation, req.Logo, req.Color, req.Description, req.FoundedYear, req.Website, req.IsMajor, req.IsActive).Scan(
		&party.ID, &party.Name, &party.Slug, &party.Abbreviation, &party.Logo, &party.Color,
		&party.Description, &party.FoundedYear, &party.Website, &party.IsMajor, &party.IsActive,
		&party.CreatedAt, &party.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create political party: %w", err)
	}

	return party, nil
}

func (r *PoliticalPartyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PoliticalParty, error) {
	party := &models.PoliticalParty{}

	err := r.db.QueryRow(ctx, `
		SELECT pp.id, pp.name, pp.slug, pp.abbreviation, pp.logo, pp.color, pp.description,
		       pp.founded_year, pp.website, pp.is_major, pp.is_active, pp.created_at, pp.updated_at, pp.deleted_at,
		       COALESCE((SELECT COUNT(*) FROM politicians WHERE party_id = pp.id AND deleted_at IS NULL), 0) as member_count
		FROM political_parties pp
		WHERE pp.id = $1 AND pp.deleted_at IS NULL
	`, id).Scan(
		&party.ID, &party.Name, &party.Slug, &party.Abbreviation, &party.Logo, &party.Color,
		&party.Description, &party.FoundedYear, &party.Website, &party.IsMajor, &party.IsActive,
		&party.CreatedAt, &party.UpdatedAt, &party.DeletedAt, &party.MemberCount,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get political party: %w", err)
	}

	return party, nil
}

func (r *PoliticalPartyRepository) GetBySlug(ctx context.Context, slug string) (*models.PoliticalParty, error) {
	party := &models.PoliticalParty{}

	err := r.db.QueryRow(ctx, `
		SELECT pp.id, pp.name, pp.slug, pp.abbreviation, pp.logo, pp.color, pp.description,
		       pp.founded_year, pp.website, pp.is_major, pp.is_active, pp.created_at, pp.updated_at, pp.deleted_at,
		       COALESCE((SELECT COUNT(*) FROM politicians WHERE party_id = pp.id AND deleted_at IS NULL), 0) as member_count
		FROM political_parties pp
		WHERE pp.slug = $1 AND pp.deleted_at IS NULL
	`, slug).Scan(
		&party.ID, &party.Name, &party.Slug, &party.Abbreviation, &party.Logo, &party.Color,
		&party.Description, &party.FoundedYear, &party.Website, &party.IsMajor, &party.IsActive,
		&party.CreatedAt, &party.UpdatedAt, &party.DeletedAt, &party.MemberCount,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get political party by slug: %w", err)
	}

	return party, nil
}

func (r *PoliticalPartyRepository) List(ctx context.Context, page, perPage int, majorOnly, activeOnly bool) (*models.PaginatedPoliticalParties, error) {
	offset := (page - 1) * perPage

	// Build query with filters
	query := `
		SELECT pp.id, pp.name, pp.slug, pp.abbreviation, pp.logo, pp.color, pp.is_major, pp.is_active,
		       COALESCE((SELECT COUNT(*) FROM politicians WHERE party_id = pp.id AND deleted_at IS NULL), 0) as member_count
		FROM political_parties pp
		WHERE pp.deleted_at IS NULL
	`
	countQuery := `SELECT COUNT(*) FROM political_parties pp WHERE pp.deleted_at IS NULL`

	if majorOnly {
		query += " AND pp.is_major = TRUE"
		countQuery += " AND pp.is_major = TRUE"
	}
	if activeOnly {
		query += " AND pp.is_active = TRUE"
		countQuery += " AND pp.is_active = TRUE"
	}

	query += " ORDER BY pp.is_major DESC, pp.name ASC LIMIT $1 OFFSET $2"

	// Get total count
	var total int
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count political parties: %w", err)
	}

	// Get parties
	rows, err := r.db.Query(ctx, query, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list political parties: %w", err)
	}
	defer rows.Close()

	parties := []models.PoliticalPartyListItem{}
	for rows.Next() {
		var party models.PoliticalPartyListItem
		err := rows.Scan(
			&party.ID, &party.Name, &party.Slug, &party.Abbreviation, &party.Logo, &party.Color,
			&party.IsMajor, &party.IsActive, &party.MemberCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan political party: %w", err)
		}
		parties = append(parties, party)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedPoliticalParties{
		Parties:    parties,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *PoliticalPartyRepository) GetAll(ctx context.Context, activeOnly bool) ([]models.PoliticalPartyListItem, error) {
	query := `
		SELECT pp.id, pp.name, pp.slug, pp.abbreviation, pp.logo, pp.color, pp.is_major, pp.is_active,
		       COALESCE((SELECT COUNT(*) FROM politicians WHERE party_id = pp.id AND deleted_at IS NULL), 0) as member_count
		FROM political_parties pp
		WHERE pp.deleted_at IS NULL
	`
	if activeOnly {
		query += " AND pp.is_active = TRUE"
	}
	query += " ORDER BY pp.is_major DESC, pp.name ASC"

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all political parties: %w", err)
	}
	defer rows.Close()

	parties := []models.PoliticalPartyListItem{}
	for rows.Next() {
		var party models.PoliticalPartyListItem
		err := rows.Scan(
			&party.ID, &party.Name, &party.Slug, &party.Abbreviation, &party.Logo, &party.Color,
			&party.IsMajor, &party.IsActive, &party.MemberCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan political party: %w", err)
		}
		parties = append(parties, party)
	}

	return parties, nil
}

func (r *PoliticalPartyRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdatePoliticalPartyRequest) (*models.PoliticalParty, error) {
	party := &models.PoliticalParty{}

	err := r.db.QueryRow(ctx, `
		UPDATE political_parties SET
			name = COALESCE($2, name),
			slug = COALESCE($3, slug),
			abbreviation = COALESCE($4, abbreviation),
			logo = COALESCE($5, logo),
			color = COALESCE($6, color),
			description = COALESCE($7, description),
			founded_year = COALESCE($8, founded_year),
			website = COALESCE($9, website),
			is_major = COALESCE($10, is_major),
			is_active = COALESCE($11, is_active)
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, slug, abbreviation, logo, color, description, founded_year, website, is_major, is_active, created_at, updated_at
	`, id, req.Name, req.Slug, req.Abbreviation, req.Logo, req.Color, req.Description, req.FoundedYear, req.Website, req.IsMajor, req.IsActive).Scan(
		&party.ID, &party.Name, &party.Slug, &party.Abbreviation, &party.Logo, &party.Color,
		&party.Description, &party.FoundedYear, &party.Website, &party.IsMajor, &party.IsActive,
		&party.CreatedAt, &party.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update political party: %w", err)
	}

	return party, nil
}

func (r *PoliticalPartyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE political_parties SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete political party: %w", err)
	}
	return nil
}

// Government Position methods

func (r *PoliticalPartyRepository) GetAllPositions(ctx context.Context) ([]models.GovernmentPositionListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, level, branch, display_order, is_elected
		FROM government_positions
		ORDER BY display_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get government positions: %w", err)
	}
	defer rows.Close()

	positions := []models.GovernmentPositionListItem{}
	for rows.Next() {
		var pos models.GovernmentPositionListItem
		err := rows.Scan(&pos.ID, &pos.Name, &pos.Slug, &pos.Level, &pos.Branch, &pos.DisplayOrder, &pos.IsElected)
		if err != nil {
			return nil, fmt.Errorf("failed to scan government position: %w", err)
		}
		positions = append(positions, pos)
	}

	return positions, nil
}

func (r *PoliticalPartyRepository) GetPositionsByLevel(ctx context.Context, level string) ([]models.GovernmentPositionListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, level, branch, display_order, is_elected
		FROM government_positions
		WHERE level = $1
		ORDER BY display_order ASC
	`, level)
	if err != nil {
		return nil, fmt.Errorf("failed to get government positions by level: %w", err)
	}
	defer rows.Close()

	positions := []models.GovernmentPositionListItem{}
	for rows.Next() {
		var pos models.GovernmentPositionListItem
		err := rows.Scan(&pos.ID, &pos.Name, &pos.Slug, &pos.Level, &pos.Branch, &pos.DisplayOrder, &pos.IsElected)
		if err != nil {
			return nil, fmt.Errorf("failed to scan government position: %w", err)
		}
		positions = append(positions, pos)
	}

	return positions, nil
}

func (r *PoliticalPartyRepository) GetPositionByID(ctx context.Context, id uuid.UUID) (*models.GovernmentPosition, error) {
	pos := &models.GovernmentPosition{}

	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, level, branch, display_order, description, max_terms, term_years, is_elected, created_at, updated_at
		FROM government_positions
		WHERE id = $1
	`, id).Scan(
		&pos.ID, &pos.Name, &pos.Slug, &pos.Level, &pos.Branch, &pos.DisplayOrder,
		&pos.Description, &pos.MaxTerms, &pos.TermYears, &pos.IsElected, &pos.CreatedAt, &pos.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get government position: %w", err)
	}

	return pos, nil
}

func (r *PoliticalPartyRepository) GetPositionBySlug(ctx context.Context, slug string) (*models.GovernmentPosition, error) {
	pos := &models.GovernmentPosition{}

	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, level, branch, display_order, description, max_terms, term_years, is_elected, created_at, updated_at
		FROM government_positions
		WHERE slug = $1
	`, slug).Scan(
		&pos.ID, &pos.Name, &pos.Slug, &pos.Level, &pos.Branch, &pos.DisplayOrder,
		&pos.Description, &pos.MaxTerms, &pos.TermYears, &pos.IsElected, &pos.CreatedAt, &pos.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get government position by slug: %w", err)
	}

	return pos, nil
}

func (r *PoliticalPartyRepository) CreatePosition(ctx context.Context, req *models.CreateGovernmentPositionRequest) (*models.GovernmentPosition, error) {
	pos := &models.GovernmentPosition{}

	err := r.db.QueryRow(ctx, `
		INSERT INTO government_positions (name, slug, level, branch, display_order, description, max_terms, term_years, is_elected)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, slug, level, branch, display_order, description, max_terms, term_years, is_elected, created_at, updated_at
	`, req.Name, req.Slug, req.Level, req.Branch, req.DisplayOrder, req.Description, req.MaxTerms, req.TermYears, req.IsElected).Scan(
		&pos.ID, &pos.Name, &pos.Slug, &pos.Level, &pos.Branch, &pos.DisplayOrder,
		&pos.Description, &pos.MaxTerms, &pos.TermYears, &pos.IsElected, &pos.CreatedAt, &pos.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create government position: %w", err)
	}

	return pos, nil
}

func (r *PoliticalPartyRepository) UpdatePosition(ctx context.Context, id uuid.UUID, req *models.UpdateGovernmentPositionRequest) (*models.GovernmentPosition, error) {
	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argCount))
		args = append(args, *req.Name)
		argCount++
	}
	if req.Slug != nil {
		updates = append(updates, fmt.Sprintf("slug = $%d", argCount))
		args = append(args, *req.Slug)
		argCount++
	}
	if req.Level != nil {
		updates = append(updates, fmt.Sprintf("level = $%d", argCount))
		args = append(args, *req.Level)
		argCount++
	}
	if req.Branch != nil {
		updates = append(updates, fmt.Sprintf("branch = $%d", argCount))
		args = append(args, *req.Branch)
		argCount++
	}
	if req.DisplayOrder != nil {
		updates = append(updates, fmt.Sprintf("display_order = $%d", argCount))
		args = append(args, *req.DisplayOrder)
		argCount++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argCount))
		args = append(args, *req.Description)
		argCount++
	}
	if req.MaxTerms != nil {
		updates = append(updates, fmt.Sprintf("max_terms = $%d", argCount))
		args = append(args, *req.MaxTerms)
		argCount++
	}
	if req.TermYears != nil {
		updates = append(updates, fmt.Sprintf("term_years = $%d", argCount))
		args = append(args, *req.TermYears)
		argCount++
	}
	if req.IsElected != nil {
		updates = append(updates, fmt.Sprintf("is_elected = $%d", argCount))
		args = append(args, *req.IsElected)
		argCount++
	}

	if len(updates) == 0 {
		return r.GetPositionByID(ctx, id)
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())
	argCount++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE government_positions
		SET %s
		WHERE id = $%d
		RETURNING id, name, slug, level, branch, display_order, description, max_terms, term_years, is_elected, created_at, updated_at
	`, strings.Join(updates, ", "), argCount)

	pos := &models.GovernmentPosition{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&pos.ID, &pos.Name, &pos.Slug, &pos.Level, &pos.Branch, &pos.DisplayOrder,
		&pos.Description, &pos.MaxTerms, &pos.TermYears, &pos.IsElected, &pos.CreatedAt, &pos.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("government position not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update government position: %w", err)
	}

	return pos, nil
}

func (r *PoliticalPartyRepository) DeletePosition(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM government_positions
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete government position: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("government position not found")
	}

	return nil
}

// Politician Jurisdiction methods

func (r *PoliticalPartyRepository) CreateJurisdiction(ctx context.Context, req *models.CreatePoliticianJurisdictionRequest) (*models.PoliticianJurisdiction, error) {
	jurisdiction := &models.PoliticianJurisdiction{}

	err := r.db.QueryRow(ctx, `
		INSERT INTO politician_jurisdictions (politician_id, region_id, province_id, city_id, barangay_id, is_national)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, politician_id, region_id, province_id, city_id, barangay_id, is_national, created_at
	`, req.PoliticianID, req.RegionID, req.ProvinceID, req.CityID, req.BarangayID, req.IsNational).Scan(
		&jurisdiction.ID, &jurisdiction.PoliticianID, &jurisdiction.RegionID, &jurisdiction.ProvinceID,
		&jurisdiction.CityID, &jurisdiction.BarangayID, &jurisdiction.IsNational, &jurisdiction.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create politician jurisdiction: %w", err)
	}

	return jurisdiction, nil
}

func (r *PoliticalPartyRepository) GetJurisdictionsByPolitician(ctx context.Context, politicianID uuid.UUID) ([]models.PoliticianJurisdiction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT pj.id, pj.politician_id, pj.region_id, pj.province_id, pj.city_id, pj.barangay_id, pj.is_national, pj.created_at,
		       r.id, r.name, r.slug,
		       p.id, p.name, p.slug,
		       c.id, c.name, c.slug, c.is_city, c.is_huc, c.is_capital,
		       b.id, b.name, b.slug
		FROM politician_jurisdictions pj
		LEFT JOIN regions r ON pj.region_id = r.id
		LEFT JOIN provinces p ON pj.province_id = p.id
		LEFT JOIN cities_municipalities c ON pj.city_id = c.id
		LEFT JOIN barangays b ON pj.barangay_id = b.id
		WHERE pj.politician_id = $1
	`, politicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to get politician jurisdictions: %w", err)
	}
	defer rows.Close()

	jurisdictions := []models.PoliticianJurisdiction{}
	for rows.Next() {
		var j models.PoliticianJurisdiction
		var regionID, regionName, regionSlug *string
		var provinceID, provinceName, provinceSlug *string
		var cityID, cityName, citySlug *string
		var cityIsCity, cityIsHUC, cityIsCapital *bool
		var barangayID, barangayName, barangaySlug *string

		err := rows.Scan(
			&j.ID, &j.PoliticianID, &j.RegionID, &j.ProvinceID, &j.CityID, &j.BarangayID, &j.IsNational, &j.CreatedAt,
			&regionID, &regionName, &regionSlug,
			&provinceID, &provinceName, &provinceSlug,
			&cityID, &cityName, &citySlug, &cityIsCity, &cityIsHUC, &cityIsCapital,
			&barangayID, &barangayName, &barangaySlug,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan politician jurisdiction: %w", err)
		}

		if regionID != nil {
			rid, _ := uuid.Parse(*regionID)
			j.Region = &models.RegionListItem{ID: rid, Name: *regionName, Slug: *regionSlug}
		}
		if provinceID != nil {
			pid, _ := uuid.Parse(*provinceID)
			j.Province = &models.ProvinceListItem{ID: pid, Name: *provinceName, Slug: *provinceSlug}
		}
		if cityID != nil {
			cid, _ := uuid.Parse(*cityID)
			j.City = &models.CityMunicipalityListItem{
				ID: cid, Name: *cityName, Slug: *citySlug,
				IsCity: *cityIsCity, IsHUC: *cityIsHUC, IsCapital: *cityIsCapital,
			}
		}
		if barangayID != nil {
			bid, _ := uuid.Parse(*barangayID)
			j.Barangay = &models.BarangayListItem{ID: bid, Name: *barangayName, Slug: *barangaySlug}
		}

		jurisdictions = append(jurisdictions, j)
	}

	return jurisdictions, nil
}

func (r *PoliticalPartyRepository) DeleteJurisdiction(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM politician_jurisdictions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete politician jurisdiction: %w", err)
	}
	return nil
}

func (r *PoliticalPartyRepository) DeleteAllJurisdictionsForPolitician(ctx context.Context, politicianID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM politician_jurisdictions WHERE politician_id = $1`, politicianID)
	if err != nil {
		return fmt.Errorf("failed to delete all politician jurisdictions: %w", err)
	}
	return nil
}

// Find representatives by location
func (r *PoliticalPartyRepository) FindRepresentativesByBarangay(ctx context.Context, barangayID uuid.UUID) ([]models.PoliticianListItem, error) {
	// Get the hierarchy for this barangay
	var cityID, provinceID, regionID uuid.UUID

	err := r.db.QueryRow(ctx, `
		SELECT b.city_id, c.province_id, p.region_id
		FROM barangays b
		JOIN cities_municipalities c ON b.city_id = c.id
		JOIN provinces p ON c.province_id = p.id
		WHERE b.id = $1
	`, barangayID).Scan(&cityID, &provinceID, &regionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get barangay hierarchy: %w", err)
	}

	// Find all politicians who represent any level of this location
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (p.id) p.id, p.name, p.slug, p.photo, p.position, p.party, p.term_start, p.term_end,
		       COALESCE((SELECT COUNT(*) FROM article_politicians WHERE politician_id = p.id), 0) as article_count
		FROM politicians p
		JOIN politician_jurisdictions pj ON p.id = pj.politician_id
		WHERE p.deleted_at IS NULL
		  AND (
		    pj.is_national = TRUE
		    OR pj.region_id = $1
		    OR pj.province_id = $2
		    OR pj.city_id = $3
		    OR pj.barangay_id = $4
		  )
		ORDER BY p.id, p.level, p.name
	`, regionID, provinceID, cityID, barangayID)
	if err != nil {
		return nil, fmt.Errorf("failed to find representatives: %w", err)
	}
	defer rows.Close()

	politicians := []models.PoliticianListItem{}
	for rows.Next() {
		var pol models.PoliticianListItem
		err := rows.Scan(
			&pol.ID, &pol.Name, &pol.Slug, &pol.Photo, &pol.Position, &pol.Party,
			&pol.TermStart, &pol.TermEnd, &pol.ArticleCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan politician: %w", err)
		}
		politicians = append(politicians, pol)
	}

	return politicians, nil
}
