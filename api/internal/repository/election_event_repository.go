package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ElectionEventRepository struct {
	db *pgxpool.Pool
}

func NewElectionEventRepository(db *pgxpool.Pool) *ElectionEventRepository {
	return &ElectionEventRepository{db: db}
}

// Create creates a new election event
func (r *ElectionEventRepository) Create(ctx context.Context, req *models.CreateElectionEventRequest, createdBy uuid.UUID) (*models.ElectionEvent, error) {
	event := &models.ElectionEvent{}

	err := r.db.QueryRow(ctx, `
		INSERT INTO election_events (name, description, election_date, level, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, election_date, level, status, created_at, updated_at, created_by
	`, req.Name, req.Description, req.ElectionDate, req.Level, createdBy).Scan(
		&event.ID, &event.Name, &event.Description, &event.ElectionDate,
		&event.Level, &event.Status, &event.CreatedAt, &event.UpdatedAt, &event.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create election event: %w", err)
	}

	return event, nil
}

// GetByID retrieves an election event by ID
func (r *ElectionEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ElectionEvent, error) {
	event := &models.ElectionEvent{}

	query := `
		SELECT
			ee.id, ee.name, ee.description, ee.election_date, ee.level, ee.status,
			ee.created_at, ee.updated_at, ee.created_by,
			COALESCE((SELECT COUNT(DISTINCT position_id)
			          FROM politician_position_history
			          WHERE election_id = ee.id), 0) as total_positions,
			COALESCE((SELECT COUNT(*)
			          FROM politician_import_logs
			          WHERE election_id = ee.id AND status = 'completed'), 0) as imported_count
		FROM election_events ee
		WHERE ee.id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.Name, &event.Description, &event.ElectionDate, &event.Level, &event.Status,
		&event.CreatedAt, &event.UpdatedAt, &event.CreatedBy,
		&event.TotalPositions, &event.ImportedCount,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get election event: %w", err)
	}

	return event, nil
}

// List retrieves all election events with pagination
func (r *ElectionEventRepository) List(ctx context.Context, page, perPage int, status *string) (*models.PaginatedElectionEvents, error) {
	offset := (page - 1) * perPage

	// Build query with filters
	query := `
		SELECT
			ee.id, ee.name, ee.election_date, ee.level, ee.status,
			COALESCE((SELECT COUNT(DISTINCT position_id)
			          FROM politician_position_history
			          WHERE election_id = ee.id), 0) as total_positions,
			COALESCE((SELECT COUNT(*)
			          FROM politician_import_logs
			          WHERE election_id = ee.id AND status = 'completed'), 0) as imported_count
		FROM election_events ee
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if status != nil {
		query += fmt.Sprintf(" AND ee.status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	query += " ORDER BY ee.election_date DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, perPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list election events: %w", err)
	}
	defer rows.Close()

	var events []models.ElectionEventListItem
	for rows.Next() {
		var event models.ElectionEventListItem
		err := rows.Scan(
			&event.ID, &event.Name, &event.ElectionDate, &event.Level, &event.Status,
			&event.TotalPositions, &event.ImportedCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan election event: %w", err)
		}
		events = append(events, event)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM election_events ee WHERE 1=1"
	countArgs := []interface{}{}

	if status != nil {
		countQuery += " AND ee.status = $1"
		countArgs = append(countArgs, *status)
	}

	var totalCount int
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count election events: %w", err)
	}

	totalPages := (totalCount + perPage - 1) / perPage

	return &models.PaginatedElectionEvents{
		ElectionEvents: events,
		Total:          totalCount,
		Page:           page,
		PerPage:        perPage,
		TotalPages:     totalPages,
	}, nil
}

// Update updates an election event
func (r *ElectionEventRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateElectionEventRequest) error {
	query := `
		UPDATE election_events
		SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			election_date = COALESCE($4, election_date),
			level = COALESCE($5, level),
			status = COALESCE($6, status),
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, req.Name, req.Description, req.ElectionDate, req.Level, req.Status)
	if err != nil {
		return fmt.Errorf("failed to update election event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("election event not found")
	}

	return nil
}

// UpdateStatus updates only the status of an election event
func (r *ElectionEventRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE election_events
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update election status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("election event not found")
	}

	return nil
}

// Delete deletes an election event
func (r *ElectionEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM election_events WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete election event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("election event not found")
	}

	return nil
}

// GetStatistics retrieves detailed statistics for an election
func (r *ElectionEventRepository) GetStatistics(ctx context.Context, electionID uuid.UUID) (*models.ElectionStatistics, error) {
	stats := &models.ElectionStatistics{
		ElectionID: electionID,
	}

	// Get election name
	err := r.db.QueryRow(ctx, "SELECT name FROM election_events WHERE id = $1", electionID).Scan(&stats.ElectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get election name: %w", err)
	}

	// Get positions affected
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT position_id)
		FROM politician_position_history
		WHERE election_id = $1
	`, electionID).Scan(&stats.TotalPositions)
	if err != nil {
		stats.TotalPositions = 0
	}

	// Get archived positions count
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM politician_position_history
		WHERE election_id = $1 AND ended_reason = 'election'
	`, electionID).Scan(&stats.PositionsArchived)
	if err != nil {
		stats.PositionsArchived = 0
	}

	// Get imported politicians count
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(successful_imports), 0)
		FROM politician_import_logs
		WHERE election_id = $1 AND status = 'completed'
	`, electionID).Scan(&stats.PoliticiansImported)
	if err != nil {
		stats.PoliticiansImported = 0
	}

	return stats, nil
}

// PaginatedElectionEvents for pagination
type PaginatedElectionEvents struct {
	Elections  []models.ElectionEventListItem `json:"elections"`
	Total      int                            `json:"total"`
	Page       int                            `json:"page"`
	PerPage    int                            `json:"per_page"`
	TotalPages int                            `json:"total_pages"`
}
