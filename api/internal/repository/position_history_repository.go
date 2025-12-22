package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PositionHistoryRepository struct {
	db *pgxpool.Pool
}

func NewPositionHistoryRepository(db *pgxpool.Pool) *PositionHistoryRepository {
	return &PositionHistoryRepository{db: db}
}

// Create creates a new position history entry
func (r *PositionHistoryRepository) Create(ctx context.Context, history *models.PoliticianPositionHistory) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO politician_position_history (
			politician_id, position_id, party_id,
			region_id, province_id, city_id, barangay_id, district_id, is_national,
			term_start, term_end, is_current, ended_reason, election_id, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`,
		history.PoliticianID, history.PositionID, history.PartyID,
		history.RegionID, history.ProvinceID, history.CityID, history.BarangayID, history.DistrictID, history.IsNational,
		history.TermStart, history.TermEnd, history.IsCurrent, history.EndedReason, history.ElectionID, history.CreatedBy,
	).Scan(&history.ID, &history.CreatedAt, &history.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create position history: %w", err)
	}

	return nil
}

// GetByID retrieves a position history entry by ID with joined data
func (r *PositionHistoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PoliticianPositionHistory, error) {
	history := &models.PoliticianPositionHistory{}

	query := `
		SELECT
			ph.id, ph.politician_id, ph.position_id, ph.party_id,
			ph.region_id, ph.province_id, ph.city_id, ph.barangay_id, ph.district_id, ph.is_national,
			ph.term_start, ph.term_end, ph.is_current, ph.ended_reason, ph.election_id,
			ph.created_at, ph.updated_at, ph.created_by,
			-- Politician
			p.id, p.name, p.slug, p.photo,
			-- Position
			gp.id, gp.name, gp.slug, gp.level, gp.branch, gp.is_elected,
			-- Party
			pp.id, pp.name, pp.slug, pp.abbreviation, pp.logo, pp.color,
			-- Election
			ee.id, ee.name, ee.election_date, ee.level
		FROM politician_position_history ph
		INNER JOIN politicians p ON ph.politician_id = p.id
		INNER JOIN government_positions gp ON ph.position_id = gp.id
		LEFT JOIN political_parties pp ON ph.party_id = pp.id
		LEFT JOIN election_events ee ON ph.election_id = ee.id
		WHERE ph.id = $1
	`

	var politician models.PoliticianBrief
	var position models.GovernmentPosition
	var party models.PartyBrief
	var election models.ElectionEventBrief

	var partyID, partyName, partySlug, partyAbbr, partyLogo, partyColor *string
	var electionID, electionName, electionLevel *string
	var electionDate *pgtype.Date

	err := r.db.QueryRow(ctx, query, id).Scan(
		&history.ID, &history.PoliticianID, &history.PositionID, &history.PartyID,
		&history.RegionID, &history.ProvinceID, &history.CityID, &history.BarangayID, &history.DistrictID, &history.IsNational,
		&history.TermStart, &history.TermEnd, &history.IsCurrent, &history.EndedReason, &history.ElectionID,
		&history.CreatedAt, &history.UpdatedAt, &history.CreatedBy,
		// Politician
		&politician.ID, &politician.Name, &politician.Slug, &politician.Photo,
		// Position
		&position.ID, &position.Name, &position.Slug, &position.Level, &position.Branch, &position.IsElected,
		// Party
		&partyID, &partyName, &partySlug, &partyAbbr, &partyLogo, &partyColor,
		// Election
		&electionID, &electionName, &electionDate, &electionLevel,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get position history: %w", err)
	}

	history.Politician = &politician
	history.Position = &position

	if partyID != nil {
		party.ID = uuid.MustParse(*partyID)
		party.Name = *partyName
		party.Slug = *partySlug
		party.Abbreviation = partyAbbr
		party.Logo = partyLogo
		party.Color = partyColor
		history.Party = &party
	}

	if electionID != nil {
		election.ID = uuid.MustParse(*electionID)
		election.Name = *electionName
		election.ElectionDate = electionDate.Time
		election.Level = *electionLevel
		history.Election = &election
	}

	return history, nil
}

// GetPoliticianHistory retrieves all position history for a politician
func (r *PositionHistoryRepository) GetPoliticianHistory(ctx context.Context, politicianID uuid.UUID) ([]models.PositionHistoryListItem, error) {
	query := `
		SELECT
			ph.id, ph.politician_id, ph.position_id, ph.term_start, ph.term_end,
			ph.is_current, ph.ended_reason,
			p.name as politician_name, p.slug as politician_slug,
			gp.name as position_name,
			pp.name as party_name, pp.color as party_color
		FROM politician_position_history ph
		INNER JOIN politicians p ON ph.politician_id = p.id
		INNER JOIN government_positions gp ON ph.position_id = gp.id
		LEFT JOIN political_parties pp ON ph.party_id = pp.id
		WHERE ph.politician_id = $1
		ORDER BY ph.term_start DESC, ph.is_current DESC
	`

	rows, err := r.db.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to get politician history: %w", err)
	}
	defer rows.Close()

	var history []models.PositionHistoryListItem
	for rows.Next() {
		var item models.PositionHistoryListItem
		err := rows.Scan(
			&item.ID, &item.PoliticianID, &item.PositionID, &item.TermStart, &item.TermEnd,
			&item.IsCurrent, &item.EndedReason,
			&item.PoliticianName, &item.PoliticianSlug,
			&item.PositionName,
			&item.PartyName, &item.PartyColor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position history: %w", err)
		}
		history = append(history, item)
	}

	return history, nil
}

// GetCurrentHolder finds the current holder of a position in a specific jurisdiction
func (r *PositionHistoryRepository) GetCurrentHolder(ctx context.Context, req *models.GetCurrentHolderRequest) (*models.PoliticianPositionHistory, error) {
	query := `
		SELECT
			ph.id, ph.politician_id, ph.position_id, ph.party_id,
			ph.region_id, ph.province_id, ph.city_id, ph.barangay_id, ph.district_id, ph.is_national,
			ph.term_start, ph.term_end, ph.is_current, ph.ended_reason, ph.election_id,
			ph.created_at, ph.updated_at, ph.created_by,
			p.id, p.name, p.slug, p.photo,
			gp.id, gp.name, gp.slug, gp.level, gp.branch, gp.is_elected
		FROM politician_position_history ph
		INNER JOIN politicians p ON ph.politician_id = p.id
		INNER JOIN government_positions gp ON ph.position_id = gp.id
		WHERE ph.position_id = $1
		  AND ph.is_current = TRUE
	`

	args := []interface{}{req.PositionID}
	argIndex := 2

	// Add jurisdiction filter based on what's provided
	if req.IsNational {
		query += " AND ph.is_national = TRUE"
	} else if req.RegionID != nil {
		query += fmt.Sprintf(" AND ph.region_id = $%d", argIndex)
		args = append(args, *req.RegionID)
	} else if req.ProvinceID != nil {
		query += fmt.Sprintf(" AND ph.province_id = $%d", argIndex)
		args = append(args, *req.ProvinceID)
	} else if req.CityID != nil {
		query += fmt.Sprintf(" AND ph.city_id = $%d", argIndex)
		args = append(args, *req.CityID)
	} else if req.BarangayID != nil {
		query += fmt.Sprintf(" AND ph.barangay_id = $%d", argIndex)
		args = append(args, *req.BarangayID)
	} else if req.DistrictID != nil {
		query += fmt.Sprintf(" AND ph.district_id = $%d", argIndex)
		args = append(args, *req.DistrictID)
	}

	history := &models.PoliticianPositionHistory{}
	var politician models.PoliticianBrief
	var position models.GovernmentPosition

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&history.ID, &history.PoliticianID, &history.PositionID, &history.PartyID,
		&history.RegionID, &history.ProvinceID, &history.CityID, &history.BarangayID, &history.DistrictID, &history.IsNational,
		&history.TermStart, &history.TermEnd, &history.IsCurrent, &history.EndedReason, &history.ElectionID,
		&history.CreatedAt, &history.UpdatedAt, &history.CreatedBy,
		&politician.ID, &politician.Name, &politician.Slug, &politician.Photo,
		&position.ID, &position.Name, &position.Slug, &position.Level, &position.Branch, &position.IsElected,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current holder: %w", err)
	}

	history.Politician = &politician
	history.Position = &position

	return history, nil
}

// Update updates a position history entry
func (r *PositionHistoryRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdatePositionHistoryRequest) error {
	query := `
		UPDATE politician_position_history
		SET
			party_id = COALESCE($2, party_id),
			term_start = COALESCE($3, term_start),
			term_end = COALESCE($4, term_end),
			is_current = COALESCE($5, is_current),
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, req.PartyID, req.TermStart, req.TermEnd, req.IsCurrent)
	if err != nil {
		return fmt.Errorf("failed to update position history: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("position history not found")
	}

	return nil
}

// EndTerm ends a politician's current term
func (r *PositionHistoryRepository) EndTerm(ctx context.Context, politicianID uuid.UUID, endDate, endedReason string) error {
	query := `
		UPDATE politician_position_history
		SET
			is_current = FALSE,
			term_end = $2,
			ended_reason = $3,
			updated_at = NOW()
		WHERE politician_id = $1 AND is_current = TRUE
	`

	result, err := r.db.Exec(ctx, query, politicianID, endDate, endedReason)
	if err != nil {
		return fmt.Errorf("failed to end term: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no current term found for politician")
	}

	return nil
}

// EndTermByID ends a specific position history entry
func (r *PositionHistoryRepository) EndTermByID(ctx context.Context, id uuid.UUID, endDate, endedReason string) error {
	query := `
		UPDATE politician_position_history
		SET
			is_current = FALSE,
			term_end = $2,
			ended_reason = $3,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, endDate, endedReason)
	if err != nil {
		return fmt.Errorf("failed to end term: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("position history not found")
	}

	return nil
}

// BulkArchiveForElection archives all current holders of specified positions for an election
func (r *PositionHistoryRepository) BulkArchiveForElection(ctx context.Context, electionID uuid.UUID, positionIDs []uuid.UUID, electionDate string) error {
	if len(positionIDs) == 0 {
		return nil
	}

	query := `
		UPDATE politician_position_history
		SET
			is_current = FALSE,
			term_end = $1,
			ended_reason = 'election',
			election_id = $2,
			updated_at = NOW()
		WHERE position_id = ANY($3) AND is_current = TRUE
	`

	result, err := r.db.Exec(ctx, query, electionDate, electionID, positionIDs)
	if err != nil {
		return fmt.Errorf("failed to bulk archive for election: %w", err)
	}

	_ = result // Successfully archived (may be 0 if no current holders)

	return nil
}

// GetPositionHolders retrieves all politicians who held a specific position
func (r *PositionHistoryRepository) GetPositionHolders(ctx context.Context, positionID uuid.UUID) ([]models.PositionHistoryListItem, error) {
	query := `
		SELECT
			ph.id, ph.politician_id, ph.position_id, ph.term_start, ph.term_end,
			ph.is_current, ph.ended_reason,
			p.name as politician_name, p.slug as politician_slug,
			gp.name as position_name,
			pp.name as party_name, pp.color as party_color
		FROM politician_position_history ph
		INNER JOIN politicians p ON ph.politician_id = p.id
		INNER JOIN government_positions gp ON ph.position_id = gp.id
		LEFT JOIN political_parties pp ON ph.party_id = pp.id
		WHERE ph.position_id = $1
		ORDER BY ph.is_current DESC, ph.term_start DESC
	`

	rows, err := r.db.Query(ctx, query, positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position holders: %w", err)
	}
	defer rows.Close()

	var holders []models.PositionHistoryListItem
	for rows.Next() {
		var item models.PositionHistoryListItem
		err := rows.Scan(
			&item.ID, &item.PoliticianID, &item.PositionID, &item.TermStart, &item.TermEnd,
			&item.IsCurrent, &item.EndedReason,
			&item.PoliticianName, &item.PoliticianSlug,
			&item.PositionName,
			&item.PartyName, &item.PartyColor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position holder: %w", err)
		}
		holders = append(holders, item)
	}

	return holders, nil
}

// Delete removes a position history entry
func (r *PositionHistoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM politician_position_history WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete position history: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("position history not found")
	}

	return nil
}

// GetCurrentHistoryForPolitician finds a politician's current position history entry
func (r *PositionHistoryRepository) GetCurrentHistoryForPolitician(ctx context.Context, politicianID uuid.UUID) (*models.PoliticianPositionHistory, error) {
	query := `
		SELECT
			ph.id, ph.politician_id, ph.position_id, ph.party_id,
			ph.region_id, ph.province_id, ph.city_id, ph.barangay_id, ph.district_id, ph.is_national,
			ph.term_start, ph.term_end, ph.is_current, ph.ended_reason, ph.election_id,
			ph.created_at, ph.updated_at, ph.created_by
		FROM politician_position_history ph
		WHERE ph.politician_id = $1 AND ph.is_current = TRUE
	`

	history := &models.PoliticianPositionHistory{}

	err := r.db.QueryRow(ctx, query, politicianID).Scan(
		&history.ID, &history.PoliticianID, &history.PositionID, &history.PartyID,
		&history.RegionID, &history.ProvinceID, &history.CityID, &history.BarangayID, &history.DistrictID, &history.IsNational,
		&history.TermStart, &history.TermEnd, &history.IsCurrent, &history.EndedReason, &history.ElectionID,
		&history.CreatedAt, &history.UpdatedAt, &history.CreatedBy,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current history: %w", err)
	}

	return history, nil
}
