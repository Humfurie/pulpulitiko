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

type ImportRepository struct {
	db *pgxpool.Pool
}

func NewImportRepository(db *pgxpool.Pool) *ImportRepository {
	return &ImportRepository{db: db}
}

// Create creates a new import log
func (r *ImportRepository) Create(ctx context.Context, log *models.PoliticianImportLog) error {
	query := `
		INSERT INTO politician_import_logs (filename, uploaded_by, status, total_rows, started_at, created_at, election_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		log.Filename,
		log.UploadedBy,
		log.Status,
		log.TotalRows,
		log.StartedAt,
		log.CreatedAt,
		log.ElectionID,
	).Scan(&log.ID)

	if err != nil {
		return fmt.Errorf("failed to create import log: %w", err)
	}

	return nil
}

// GetByID gets an import log by ID
func (r *ImportRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PoliticianImportLog, error) {
	query := `
		SELECT id, filename, uploaded_by, total_rows, successful_imports, failed_imports,
			   politicians_created, politicians_updated, positions_archived,
			   status, error_log, validation_errors, election_id, started_at, completed_at, created_at
		FROM politician_import_logs
		WHERE id = $1
	`

	log := &models.PoliticianImportLog{}
	var validationErrorsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&log.ID,
		&log.Filename,
		&log.UploadedBy,
		&log.TotalRows,
		&log.SuccessfulImports,
		&log.FailedImports,
		&log.PoliticiansCreated,
		&log.PoliticiansUpdated,
		&log.PositionsArchived,
		&log.Status,
		&log.ErrorLog,
		&validationErrorsJSON,
		&log.ElectionID,
		&log.StartedAt,
		&log.CompletedAt,
		&log.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get import log: %w", err)
	}

	// Parse validation errors JSON
	if validationErrorsJSON != nil {
		err = json.Unmarshal(validationErrorsJSON, &log.ValidationErrors)
		if err != nil {
			return nil, fmt.Errorf("failed to parse validation errors: %w", err)
		}
	}

	return log, nil
}

// List lists import logs with pagination
func (r *ImportRepository) List(ctx context.Context, page, perPage int) (*models.PaginatedImportLogs, error) {
	// Count total
	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM politician_import_logs").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count import logs: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * perPage
	query := `
		SELECT id, filename, uploaded_by, total_rows, successful_imports, failed_imports,
			   politicians_created, politicians_updated, positions_archived,
			   status, election_id, started_at, completed_at, created_at
		FROM politician_import_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list import logs: %w", err)
	}
	defer rows.Close()

	var logs []models.PoliticianImportLog
	for rows.Next() {
		var log models.PoliticianImportLog
		err := rows.Scan(
			&log.ID,
			&log.Filename,
			&log.UploadedBy,
			&log.TotalRows,
			&log.SuccessfulImports,
			&log.FailedImports,
			&log.PoliticiansCreated,
			&log.PoliticiansUpdated,
			&log.PositionsArchived,
			&log.Status,
			&log.ElectionID,
			&log.StartedAt,
			&log.CompletedAt,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan import log: %w", err)
		}
		logs = append(logs, log)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedImportLogs{
		ImportLogs: logs,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// UpdateStatus updates the status of an import log
func (r *ImportRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := "UPDATE politician_import_logs SET status = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

// UpdateTotalRows updates the total rows count
func (r *ImportRepository) UpdateTotalRows(ctx context.Context, id uuid.UUID, totalRows int) error {
	query := "UPDATE politician_import_logs SET total_rows = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, query, totalRows, id)
	if err != nil {
		return fmt.Errorf("failed to update total rows: %w", err)
	}
	return nil
}

// UpdateErrorLog updates the error log message
func (r *ImportRepository) UpdateErrorLog(ctx context.Context, id uuid.UUID, errorLog string) error {
	query := "UPDATE politician_import_logs SET error_log = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, query, errorLog, id)
	if err != nil {
		return fmt.Errorf("failed to update error log: %w", err)
	}
	return nil
}

// UpdateValidationErrors updates the validation errors
func (r *ImportRepository) UpdateValidationErrors(ctx context.Context, id uuid.UUID, errors []models.ValidationError) error {
	errorsJSON, err := json.Marshal(errors)
	if err != nil {
		return fmt.Errorf("failed to marshal errors: %w", err)
	}

	query := "UPDATE politician_import_logs SET validation_errors = $1 WHERE id = $2"
	_, err = r.db.Exec(ctx, query, errorsJSON, id)
	if err != nil {
		return fmt.Errorf("failed to update validation errors: %w", err)
	}
	return nil
}

// UpdateStatistics updates import statistics
func (r *ImportRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats *models.ImportStatistics) error {
	query := `
		UPDATE politician_import_logs
		SET successful_imports = $1, failed_imports = $2,
		    politicians_created = $3, politicians_updated = $4,
		    positions_archived = $5, completed_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(ctx, query,
		stats.SuccessfulImports,
		stats.FailedImports,
		stats.PoliticiansCreated,
		stats.PoliticiansUpdated,
		stats.PositionsArchived,
		stats.CompletedAt,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update statistics: %w", err)
	}
	return nil
}
