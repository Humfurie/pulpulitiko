package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/humfurie/pulpulitiko/api/internal/models"
)

type ElectionRepository struct {
	db *pgxpool.Pool
}

func NewElectionRepository(db *pgxpool.Pool) *ElectionRepository {
	return &ElectionRepository{db: db}
}

// Elections

func (r *ElectionRepository) CreateElection(ctx context.Context, req *models.CreateElectionRequest) (*models.Election, error) {
	electionDate, err := time.Parse("2006-01-02", req.ElectionDate)
	if err != nil {
		return nil, fmt.Errorf("invalid election_date format: %w", err)
	}

	var registrationStart, registrationEnd, campaignStart, campaignEnd *time.Time
	if req.RegistrationStart != nil {
		t, _ := time.Parse("2006-01-02", *req.RegistrationStart)
		registrationStart = &t
	}
	if req.RegistrationEnd != nil {
		t, _ := time.Parse("2006-01-02", *req.RegistrationEnd)
		registrationEnd = &t
	}
	if req.CampaignStart != nil {
		t, _ := time.Parse("2006-01-02", *req.CampaignStart)
		campaignStart = &t
	}
	if req.CampaignEnd != nil {
		t, _ := time.Parse("2006-01-02", *req.CampaignEnd)
		campaignEnd = &t
	}

	election := &models.Election{}
	err = r.db.QueryRow(ctx, `
		INSERT INTO elections (name, slug, election_type, description, election_date, registration_start, registration_end, campaign_start, campaign_end, status, is_featured)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, name, slug, election_type, description, election_date, registration_start, registration_end, campaign_start, campaign_end, status, is_featured, created_at, updated_at
	`, req.Name, req.Slug, req.ElectionType, req.Description, electionDate, registrationStart, registrationEnd, campaignStart, campaignEnd, req.Status, req.IsFeatured).Scan(
		&election.ID, &election.Name, &election.Slug, &election.ElectionType, &election.Description,
		&election.ElectionDate, &election.RegistrationStart, &election.RegistrationEnd, &election.CampaignStart, &election.CampaignEnd,
		&election.Status, &election.IsFeatured, &election.CreatedAt, &election.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create election: %w", err)
	}

	return election, nil
}

func (r *ElectionRepository) GetElectionByID(ctx context.Context, id uuid.UUID) (*models.Election, error) {
	election := &models.Election{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, election_type, description, election_date, registration_start, registration_end,
		       campaign_start, campaign_end, status, is_featured, voter_turnout_percentage, total_registered_voters,
		       total_votes_cast, created_at, updated_at
		FROM elections
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&election.ID, &election.Name, &election.Slug, &election.ElectionType, &election.Description,
		&election.ElectionDate, &election.RegistrationStart, &election.RegistrationEnd, &election.CampaignStart, &election.CampaignEnd,
		&election.Status, &election.IsFeatured, &election.VoterTurnoutPercentage, &election.TotalRegisteredVoters,
		&election.TotalVotesCast, &election.CreatedAt, &election.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get election: %w", err)
	}

	return election, nil
}

func (r *ElectionRepository) GetElectionBySlug(ctx context.Context, slug string) (*models.Election, error) {
	election := &models.Election{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, election_type, description, election_date, registration_start, registration_end,
		       campaign_start, campaign_end, status, is_featured, voter_turnout_percentage, total_registered_voters,
		       total_votes_cast, created_at, updated_at
		FROM elections
		WHERE slug = $1 AND deleted_at IS NULL
	`, slug).Scan(
		&election.ID, &election.Name, &election.Slug, &election.ElectionType, &election.Description,
		&election.ElectionDate, &election.RegistrationStart, &election.RegistrationEnd, &election.CampaignStart, &election.CampaignEnd,
		&election.Status, &election.IsFeatured, &election.VoterTurnoutPercentage, &election.TotalRegisteredVoters,
		&election.TotalVotesCast, &election.CreatedAt, &election.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get election: %w", err)
	}

	// Load positions
	election.Positions, _ = r.GetElectionPositions(ctx, election.ID)

	return election, nil
}

func (r *ElectionRepository) ListElections(ctx context.Context, filter *models.ElectionFilter, page, perPage int) (*models.PaginatedElections, error) {
	offset := (page - 1) * perPage

	whereClause := "WHERE e.deleted_at IS NULL"
	args := []interface{}{}
	argNum := 1

	if filter != nil {
		if filter.ElectionType != nil {
			whereClause += fmt.Sprintf(" AND e.election_type = $%d", argNum)
			args = append(args, *filter.ElectionType)
			argNum++
		}
		if filter.Status != nil {
			whereClause += fmt.Sprintf(" AND e.status = $%d", argNum)
			args = append(args, *filter.Status)
			argNum++
		}
		if filter.Year != nil {
			whereClause += fmt.Sprintf(" AND EXTRACT(YEAR FROM e.election_date) = $%d", argNum)
			args = append(args, *filter.Year)
			argNum++
		}
		if filter.IsFeatured != nil {
			whereClause += fmt.Sprintf(" AND e.is_featured = $%d", argNum)
			args = append(args, *filter.IsFeatured)
			argNum++
		}
		if filter.Search != nil && *filter.Search != "" {
			whereClause += fmt.Sprintf(" AND (e.name ILIKE $%d OR e.description ILIKE $%d)", argNum, argNum)
			args = append(args, "%"+*filter.Search+"%")
			argNum++
		}
	}

	// Count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM elections e %s", whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count elections: %w", err)
	}

	// List
	query := fmt.Sprintf(`
		SELECT e.id, e.name, e.slug, e.election_type, e.election_date, e.status, e.is_featured, e.voter_turnout_percentage,
		       COALESCE((SELECT COUNT(*) FROM election_positions WHERE election_id = e.id), 0) as position_count,
		       COALESCE((SELECT COUNT(*) FROM candidates c JOIN election_positions ep ON c.election_position_id = ep.id WHERE ep.election_id = e.id), 0) as candidate_count
		FROM elections e
		%s
		ORDER BY e.election_date DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)
	args = append(args, perPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list elections: %w", err)
	}
	defer rows.Close()

	var elections []models.ElectionListItem
	for rows.Next() {
		var e models.ElectionListItem
		err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.ElectionType, &e.ElectionDate, &e.Status, &e.IsFeatured, &e.VoterTurnoutPercentage,
			&e.PositionCount, &e.CandidateCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan election: %w", err)
		}
		elections = append(elections, e)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedElections{
		Elections:  elections,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *ElectionRepository) GetUpcomingElections(ctx context.Context, limit int) ([]models.ElectionListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT e.id, e.name, e.slug, e.election_type, e.election_date, e.status, e.is_featured, e.voter_turnout_percentage,
		       COALESCE((SELECT COUNT(*) FROM election_positions WHERE election_id = e.id), 0) as position_count,
		       COALESCE((SELECT COUNT(*) FROM candidates c JOIN election_positions ep ON c.election_position_id = ep.id WHERE ep.election_id = e.id), 0) as candidate_count
		FROM elections e
		WHERE e.deleted_at IS NULL AND e.status = 'upcoming' AND e.election_date >= CURRENT_DATE
		ORDER BY e.election_date ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming elections: %w", err)
	}
	defer rows.Close()

	var elections []models.ElectionListItem
	for rows.Next() {
		var e models.ElectionListItem
		err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.ElectionType, &e.ElectionDate, &e.Status, &e.IsFeatured, &e.VoterTurnoutPercentage,
			&e.PositionCount, &e.CandidateCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan election: %w", err)
		}
		elections = append(elections, e)
	}

	return elections, nil
}

func (r *ElectionRepository) GetFeaturedElections(ctx context.Context) ([]models.ElectionListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT e.id, e.name, e.slug, e.election_type, e.election_date, e.status, e.is_featured, e.voter_turnout_percentage,
		       COALESCE((SELECT COUNT(*) FROM election_positions WHERE election_id = e.id), 0) as position_count,
		       COALESCE((SELECT COUNT(*) FROM candidates c JOIN election_positions ep ON c.election_position_id = ep.id WHERE ep.election_id = e.id), 0) as candidate_count
		FROM elections e
		WHERE e.deleted_at IS NULL AND e.is_featured = TRUE
		ORDER BY e.election_date DESC
		LIMIT 5
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured elections: %w", err)
	}
	defer rows.Close()

	var elections []models.ElectionListItem
	for rows.Next() {
		var e models.ElectionListItem
		err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.ElectionType, &e.ElectionDate, &e.Status, &e.IsFeatured, &e.VoterTurnoutPercentage,
			&e.PositionCount, &e.CandidateCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan election: %w", err)
		}
		elections = append(elections, e)
	}

	return elections, nil
}

func (r *ElectionRepository) GetElectionCalendar(ctx context.Context, year int) ([]models.ElectionCalendarItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, election_type, election_date, status
		FROM elections
		WHERE deleted_at IS NULL AND EXTRACT(YEAR FROM election_date) = $1
		ORDER BY election_date ASC
	`, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get election calendar: %w", err)
	}
	defer rows.Close()

	var items []models.ElectionCalendarItem
	for rows.Next() {
		var item models.ElectionCalendarItem
		err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.ElectionType, &item.ElectionDate, &item.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan calendar item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ElectionRepository) UpdateElection(ctx context.Context, id uuid.UUID, req *models.UpdateElectionRequest) (*models.Election, error) {
	setClauses := []string{}
	args := []interface{}{id}

	if req.Name != nil {
		args = append(args, *req.Name)
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", len(args)))
	}
	if req.Slug != nil {
		args = append(args, *req.Slug)
		setClauses = append(setClauses, fmt.Sprintf("slug = $%d", len(args)))
	}
	if req.Description != nil {
		args = append(args, *req.Description)
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", len(args)))
	}
	if req.ElectionDate != nil {
		date, _ := time.Parse("2006-01-02", *req.ElectionDate)
		args = append(args, date)
		setClauses = append(setClauses, fmt.Sprintf("election_date = $%d", len(args)))
	}
	if req.Status != nil {
		args = append(args, *req.Status)
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", len(args)))
	}
	if req.IsFeatured != nil {
		args = append(args, *req.IsFeatured)
		setClauses = append(setClauses, fmt.Sprintf("is_featured = $%d", len(args)))
	}
	if req.VoterTurnoutPercentage != nil {
		args = append(args, *req.VoterTurnoutPercentage)
		setClauses = append(setClauses, fmt.Sprintf("voter_turnout_percentage = $%d", len(args)))
	}
	if req.TotalRegisteredVoters != nil {
		args = append(args, *req.TotalRegisteredVoters)
		setClauses = append(setClauses, fmt.Sprintf("total_registered_voters = $%d", len(args)))
	}
	if req.TotalVotesCast != nil {
		args = append(args, *req.TotalVotesCast)
		setClauses = append(setClauses, fmt.Sprintf("total_votes_cast = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		return r.GetElectionByID(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE elections SET %s
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, slug, election_type, description, election_date, status, is_featured, voter_turnout_percentage, total_registered_voters, total_votes_cast, created_at, updated_at
	`, strings.Join(setClauses, ", "))

	election := &models.Election{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&election.ID, &election.Name, &election.Slug, &election.ElectionType, &election.Description,
		&election.ElectionDate, &election.Status, &election.IsFeatured, &election.VoterTurnoutPercentage,
		&election.TotalRegisteredVoters, &election.TotalVotesCast, &election.CreatedAt, &election.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update election: %w", err)
	}

	return election, nil
}

func (r *ElectionRepository) DeleteElection(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE elections SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("failed to delete election: %w", err)
	}
	return nil
}

// Election Positions

func (r *ElectionRepository) CreateElectionPosition(ctx context.Context, req *models.CreateElectionPositionRequest) (*models.ElectionPosition, error) {
	position := &models.ElectionPosition{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO election_positions (election_id, position_id, region_id, province_id, city_municipality_id, barangay_id, district_id, seats_available, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, election_id, position_id, region_id, province_id, city_municipality_id, barangay_id, district_id, seats_available, description, created_at
	`, req.ElectionID, req.PositionID, req.RegionID, req.ProvinceID, req.CityMunicipalityID, req.BarangayID, req.DistrictID, req.SeatsAvailable, req.Description).Scan(
		&position.ID, &position.ElectionID, &position.PositionID, &position.RegionID, &position.ProvinceID,
		&position.CityMunicipalityID, &position.BarangayID, &position.DistrictID, &position.SeatsAvailable, &position.Description, &position.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create election position: %w", err)
	}
	return position, nil
}

func (r *ElectionRepository) GetElectionPositions(ctx context.Context, electionID uuid.UUID) ([]models.ElectionPositionListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ep.id, ep.position_id, ep.seats_available,
		       gp.id, gp.name, gp.slug, gp.level, gp.branch, gp.is_elected,
		       COALESCE(r.name, pr.name, cm.name, b.name, cd.name, '') as location_name,
		       COALESCE((SELECT COUNT(*) FROM candidates WHERE election_position_id = ep.id), 0) as candidate_count
		FROM election_positions ep
		JOIN government_positions gp ON ep.position_id = gp.id
		LEFT JOIN regions r ON ep.region_id = r.id
		LEFT JOIN provinces pr ON ep.province_id = pr.id
		LEFT JOIN cities_municipalities cm ON ep.city_municipality_id = cm.id
		LEFT JOIN barangays b ON ep.barangay_id = b.id
		LEFT JOIN congressional_districts cd ON ep.district_id = cd.id
		WHERE ep.election_id = $1
		ORDER BY gp.display_order, location_name
	`, electionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get election positions: %w", err)
	}
	defer rows.Close()

	var positions []models.ElectionPositionListItem
	for rows.Next() {
		var p models.ElectionPositionListItem
		var posInfo models.GovernmentPositionInfo
		var locationName string
		err := rows.Scan(
			&p.ID, &p.PositionID, &p.SeatsAvailable,
			&posInfo.ID, &posInfo.Name, &posInfo.Slug, &posInfo.Level, &posInfo.Branch, &posInfo.IsElected,
			&locationName, &p.CandidateCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		p.Position = &posInfo
		if locationName != "" {
			p.Location = &locationName
		}
		positions = append(positions, p)
	}

	return positions, nil
}

// Candidates

func (r *ElectionRepository) CreateCandidate(ctx context.Context, req *models.CreateCandidateRequest) (*models.Candidate, error) {
	var filingDate *time.Time
	if req.FilingDate != nil {
		t, _ := time.Parse("2006-01-02", *req.FilingDate)
		filingDate = &t
	}

	candidate := &models.Candidate{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO candidates (election_position_id, politician_id, party_id, ballot_number, ballot_name, campaign_slogan, platform, status, filing_date, is_incumbent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, election_position_id, politician_id, party_id, ballot_number, ballot_name, campaign_slogan, platform, status, filing_date, is_incumbent, is_winner, votes_received, vote_percentage, created_at, updated_at
	`, req.ElectionPositionID, req.PoliticianID, req.PartyID, req.BallotNumber, req.BallotName, req.CampaignSlogan, req.Platform, req.Status, filingDate, req.IsIncumbent).Scan(
		&candidate.ID, &candidate.ElectionPositionID, &candidate.PoliticianID, &candidate.PartyID,
		&candidate.BallotNumber, &candidate.BallotName, &candidate.CampaignSlogan, &candidate.Platform,
		&candidate.Status, &candidate.FilingDate, &candidate.IsIncumbent, &candidate.IsWinner,
		&candidate.VotesReceived, &candidate.VotePercentage, &candidate.CreatedAt, &candidate.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create candidate: %w", err)
	}
	return candidate, nil
}

func (r *ElectionRepository) GetCandidateByID(ctx context.Context, id uuid.UUID) (*models.Candidate, error) {
	candidate := &models.Candidate{}
	var pol models.PoliticianListItem
	var party models.PartyBrief
	var partyID *uuid.UUID

	err := r.db.QueryRow(ctx, `
		SELECT c.id, c.election_position_id, c.politician_id, c.party_id, c.ballot_number, c.ballot_name,
		       c.campaign_slogan, c.platform, c.status, c.filing_date, c.is_incumbent, c.is_winner,
		       c.votes_received, c.vote_percentage, c.created_at, c.updated_at,
		       p.id, p.name, p.slug, p.photo, p.position, p.party,
		       pp.id, pp.name, pp.slug, pp.abbreviation, pp.logo, pp.color
		FROM candidates c
		JOIN politicians p ON c.politician_id = p.id
		LEFT JOIN political_parties pp ON c.party_id = pp.id
		WHERE c.id = $1
	`, id).Scan(
		&candidate.ID, &candidate.ElectionPositionID, &candidate.PoliticianID, &partyID,
		&candidate.BallotNumber, &candidate.BallotName, &candidate.CampaignSlogan, &candidate.Platform,
		&candidate.Status, &candidate.FilingDate, &candidate.IsIncumbent, &candidate.IsWinner,
		&candidate.VotesReceived, &candidate.VotePercentage, &candidate.CreatedAt, &candidate.UpdatedAt,
		&pol.ID, &pol.Name, &pol.Slug, &pol.Photo, &pol.Position, &pol.Party,
		&party.ID, &party.Name, &party.Slug, &party.Abbreviation, &party.Logo, &party.Color,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate: %w", err)
	}

	candidate.PartyID = partyID
	candidate.Politician = &pol
	if partyID != nil {
		candidate.Party = &party
	}

	return candidate, nil
}

func (r *ElectionRepository) GetCandidatesForPosition(ctx context.Context, positionID uuid.UUID) ([]models.CandidateListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.politician_id, c.ballot_number, c.ballot_name, c.status, c.is_incumbent, c.is_winner, c.votes_received, c.vote_percentage,
		       p.id, p.name, p.slug, p.photo, p.position, p.party,
		       pp.id, pp.name, pp.slug, pp.abbreviation, pp.logo, pp.color
		FROM candidates c
		JOIN politicians p ON c.politician_id = p.id
		LEFT JOIN political_parties pp ON c.party_id = pp.id
		WHERE c.election_position_id = $1
		ORDER BY COALESCE(c.votes_received, 0) DESC, c.ballot_number
	`, positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}
	defer rows.Close()

	var candidates []models.CandidateListItem
	for rows.Next() {
		var c models.CandidateListItem
		var pol models.PoliticianListItem
		var party models.PartyBrief
		var partyID, partyName, partySlug, partyAbbr, partyLogo, partyColor *string

		err := rows.Scan(
			&c.ID, &c.PoliticianID, &c.BallotNumber, &c.BallotName, &c.Status, &c.IsIncumbent, &c.IsWinner, &c.VotesReceived, &c.VotePercentage,
			&pol.ID, &pol.Name, &pol.Slug, &pol.Photo, &pol.Position, &pol.Party,
			&partyID, &partyName, &partySlug, &partyAbbr, &partyLogo, &partyColor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan candidate: %w", err)
		}

		c.Politician = &pol
		if partyID != nil {
			party.Name = *partyName
			party.Slug = *partySlug
			party.Abbreviation = partyAbbr
			party.Logo = partyLogo
			party.Color = partyColor
			c.Party = &party
		}
		candidates = append(candidates, c)
	}

	return candidates, nil
}

func (r *ElectionRepository) ListCandidates(ctx context.Context, filter *models.CandidateFilter, page, perPage int) (*models.PaginatedCandidates, error) {
	offset := (page - 1) * perPage

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argNum := 1

	if filter != nil {
		if filter.ElectionID != nil {
			whereClause += fmt.Sprintf(" AND ep.election_id = $%d", argNum)
			args = append(args, *filter.ElectionID)
			argNum++
		}
		if filter.PositionID != nil {
			whereClause += fmt.Sprintf(" AND c.election_position_id = $%d", argNum)
			args = append(args, *filter.PositionID)
			argNum++
		}
		if filter.PoliticianID != nil {
			whereClause += fmt.Sprintf(" AND c.politician_id = $%d", argNum)
			args = append(args, *filter.PoliticianID)
			argNum++
		}
		if filter.PartyID != nil {
			whereClause += fmt.Sprintf(" AND c.party_id = $%d", argNum)
			args = append(args, *filter.PartyID)
			argNum++
		}
		if filter.Status != nil {
			whereClause += fmt.Sprintf(" AND c.status = $%d", argNum)
			args = append(args, *filter.Status)
			argNum++
		}
		if filter.IsWinner != nil {
			whereClause += fmt.Sprintf(" AND c.is_winner = $%d", argNum)
			args = append(args, *filter.IsWinner)
			argNum++
		}
	}

	// Count
	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM candidates c
		JOIN election_positions ep ON c.election_position_id = ep.id
		%s
	`, whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count candidates: %w", err)
	}

	// List
	query := fmt.Sprintf(`
		SELECT c.id, c.politician_id, c.ballot_number, c.ballot_name, c.status, c.is_incumbent, c.is_winner, c.votes_received, c.vote_percentage,
		       p.id, p.name, p.slug, p.photo, p.position, p.party,
		       pp.id, pp.name, pp.slug, pp.abbreviation, pp.logo, pp.color
		FROM candidates c
		JOIN election_positions ep ON c.election_position_id = ep.id
		JOIN politicians p ON c.politician_id = p.id
		LEFT JOIN political_parties pp ON c.party_id = pp.id
		%s
		ORDER BY COALESCE(c.votes_received, 0) DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)
	args = append(args, perPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list candidates: %w", err)
	}
	defer rows.Close()

	var candidates []models.CandidateListItem
	for rows.Next() {
		var c models.CandidateListItem
		var pol models.PoliticianListItem
		var party models.PartyBrief
		var partyID, partyName, partySlug, partyAbbr, partyLogo, partyColor *string

		err := rows.Scan(
			&c.ID, &c.PoliticianID, &c.BallotNumber, &c.BallotName, &c.Status, &c.IsIncumbent, &c.IsWinner, &c.VotesReceived, &c.VotePercentage,
			&pol.ID, &pol.Name, &pol.Slug, &pol.Photo, &pol.Position, &pol.Party,
			&partyID, &partyName, &partySlug, &partyAbbr, &partyLogo, &partyColor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan candidate: %w", err)
		}

		c.Politician = &pol
		if partyID != nil {
			party.Name = *partyName
			party.Slug = *partySlug
			party.Abbreviation = partyAbbr
			party.Logo = partyLogo
			party.Color = partyColor
			c.Party = &party
		}
		candidates = append(candidates, c)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedCandidates{
		Candidates: candidates,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *ElectionRepository) UpdateCandidate(ctx context.Context, id uuid.UUID, req *models.UpdateCandidateRequest) (*models.Candidate, error) {
	setClauses := []string{}
	args := []interface{}{id}

	if req.PartyID != nil {
		args = append(args, *req.PartyID)
		setClauses = append(setClauses, fmt.Sprintf("party_id = $%d", len(args)))
	}
	if req.BallotNumber != nil {
		args = append(args, *req.BallotNumber)
		setClauses = append(setClauses, fmt.Sprintf("ballot_number = $%d", len(args)))
	}
	if req.BallotName != nil {
		args = append(args, *req.BallotName)
		setClauses = append(setClauses, fmt.Sprintf("ballot_name = $%d", len(args)))
	}
	if req.CampaignSlogan != nil {
		args = append(args, *req.CampaignSlogan)
		setClauses = append(setClauses, fmt.Sprintf("campaign_slogan = $%d", len(args)))
	}
	if req.Platform != nil {
		args = append(args, *req.Platform)
		setClauses = append(setClauses, fmt.Sprintf("platform = $%d", len(args)))
	}
	if req.Status != nil {
		args = append(args, *req.Status)
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", len(args)))
	}
	if req.IsIncumbent != nil {
		args = append(args, *req.IsIncumbent)
		setClauses = append(setClauses, fmt.Sprintf("is_incumbent = $%d", len(args)))
	}
	if req.IsWinner != nil {
		args = append(args, *req.IsWinner)
		setClauses = append(setClauses, fmt.Sprintf("is_winner = $%d", len(args)))
	}
	if req.VotesReceived != nil {
		args = append(args, *req.VotesReceived)
		setClauses = append(setClauses, fmt.Sprintf("votes_received = $%d", len(args)))
	}
	if req.VotePercentage != nil {
		args = append(args, *req.VotePercentage)
		setClauses = append(setClauses, fmt.Sprintf("vote_percentage = $%d", len(args)))
	}

	if len(setClauses) == 0 {
		return r.GetCandidateByID(ctx, id)
	}

	_, err := r.db.Exec(ctx, fmt.Sprintf("UPDATE candidates SET %s WHERE id = $1", strings.Join(setClauses, ", ")), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update candidate: %w", err)
	}

	return r.GetCandidateByID(ctx, id)
}

// Voter Education

func (r *ElectionRepository) CreateVoterEducation(ctx context.Context, req *models.CreateVoterEducationRequest) (*models.VoterEducation, error) {
	var publishedAt *time.Time
	if req.IsPublished {
		now := time.Now()
		publishedAt = &now
	}

	ve := &models.VoterEducation{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO voter_education (election_id, title, slug, content, content_type, category, is_featured, is_published, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, election_id, title, slug, content, content_type, category, is_featured, is_published, published_at, view_count, created_at, updated_at
	`, req.ElectionID, req.Title, req.Slug, req.Content, req.ContentType, req.Category, req.IsFeatured, req.IsPublished, publishedAt).Scan(
		&ve.ID, &ve.ElectionID, &ve.Title, &ve.Slug, &ve.Content, &ve.ContentType, &ve.Category,
		&ve.IsFeatured, &ve.IsPublished, &ve.PublishedAt, &ve.ViewCount, &ve.CreatedAt, &ve.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create voter education: %w", err)
	}
	return ve, nil
}

func (r *ElectionRepository) GetVoterEducationBySlug(ctx context.Context, slug string) (*models.VoterEducation, error) {
	ve := &models.VoterEducation{}
	err := r.db.QueryRow(ctx, `
		SELECT id, election_id, title, slug, content, content_type, category, is_featured, is_published, published_at, view_count, created_at, updated_at
		FROM voter_education
		WHERE slug = $1 AND deleted_at IS NULL
	`, slug).Scan(
		&ve.ID, &ve.ElectionID, &ve.Title, &ve.Slug, &ve.Content, &ve.ContentType, &ve.Category,
		&ve.IsFeatured, &ve.IsPublished, &ve.PublishedAt, &ve.ViewCount, &ve.CreatedAt, &ve.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get voter education: %w", err)
	}
	return ve, nil
}

func (r *ElectionRepository) ListVoterEducation(ctx context.Context, electionID *uuid.UUID, category *string, page, perPage int) (*models.PaginatedVoterEducation, error) {
	offset := (page - 1) * perPage

	whereClause := "WHERE deleted_at IS NULL AND is_published = TRUE"
	args := []interface{}{}
	argNum := 1

	if electionID != nil {
		whereClause += fmt.Sprintf(" AND election_id = $%d", argNum)
		args = append(args, *electionID)
		argNum++
	}
	if category != nil {
		whereClause += fmt.Sprintf(" AND category = $%d", argNum)
		args = append(args, *category)
		argNum++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM voter_education %s", whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count voter education: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, title, slug, content_type, category, is_featured, view_count, published_at
		FROM voter_education
		%s
		ORDER BY is_featured DESC, published_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)
	args = append(args, perPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list voter education: %w", err)
	}
	defer rows.Close()

	var items []models.VoterEducationListItem
	for rows.Next() {
		var item models.VoterEducationListItem
		err := rows.Scan(&item.ID, &item.Title, &item.Slug, &item.ContentType, &item.Category, &item.IsFeatured, &item.ViewCount, &item.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan voter education: %w", err)
		}
		items = append(items, item)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedVoterEducation{
		Items:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *ElectionRepository) IncrementVoterEducationViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE voter_education SET view_count = view_count + 1 WHERE id = $1`, id)
	return err
}
