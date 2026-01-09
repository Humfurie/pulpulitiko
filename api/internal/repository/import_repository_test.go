package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	// Connect to test database
	// Update this connection string to match your test database
	connString := "postgres://politics:localdev@localhost:5432/politics_db_test?sslmode=disable"

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Skip("Skipping database tests: cannot connect to test database")
		return nil
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skip("Skipping database tests: cannot ping test database")
		return nil
	}

	// Clean up any existing test data
	_, _ = pool.Exec(ctx, "TRUNCATE TABLE politician_import_logs CASCADE")

	return pool
}

func teardownTestDB(t *testing.T, pool *pgxpool.Pool) {
	if pool != nil {
		ctx := context.Background()
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE politician_import_logs CASCADE")
		pool.Close()
	}
}

func TestImportRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardownTestDB(t, pool)

	repo := NewImportRepository(pool)
	ctx := context.Background()

	t.Run("successfully creates import log", func(t *testing.T) {
		userID := uuid.New()
		electionID := uuid.New()
		startedAt := time.Now()

		log := &models.PoliticianImportLog{
			Filename:   "test_politicians.xlsx",
			UploadedBy: &userID,
			Status:     "pending",
			TotalRows:  0,
			StartedAt:  startedAt,
			CreatedAt:  time.Now(),
			ElectionID: &electionID,
		}

		err := repo.Create(ctx, log)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, log.ID)
	})

	t.Run("creates import log without optional fields", func(t *testing.T) {
		log := &models.PoliticianImportLog{
			Filename:  "test_minimal.xlsx",
			Status:    "pending",
			TotalRows: 0,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}

		err := repo.Create(ctx, log)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, log.ID)
	})

	t.Run("fails with invalid context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		log := &models.PoliticianImportLog{
			Filename:  "test_canceled.xlsx",
			Status:    "pending",
			TotalRows: 0,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}

		err := repo.Create(canceledCtx, log)
		assert.Error(t, err)
	})
}

func TestImportRepository_GetByID(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardownTestDB(t, pool)

	repo := NewImportRepository(pool)
	ctx := context.Background()

	t.Run("successfully retrieves import log by ID", func(t *testing.T) {
		// Create a test log
		userID := uuid.New()
		createdLog := &models.PoliticianImportLog{
			Filename:   "test_get.xlsx",
			UploadedBy: &userID,
			Status:     "completed",
			TotalRows:  100,
			StartedAt:  time.Now(),
			CreatedAt:  time.Now(),
		}
		err := repo.Create(ctx, createdLog)
		require.NoError(t, err)

		// Retrieve the log
		retrievedLog, err := repo.GetByID(ctx, createdLog.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrievedLog)
		assert.Equal(t, createdLog.ID, retrievedLog.ID)
		assert.Equal(t, createdLog.Filename, retrievedLog.Filename)
		assert.Equal(t, createdLog.Status, retrievedLog.Status)
		assert.Equal(t, createdLog.TotalRows, retrievedLog.TotalRows)
	})

	t.Run("returns nil for non-existent ID", func(t *testing.T) {
		nonExistentID := uuid.New()
		log, err := repo.GetByID(ctx, nonExistentID)
		require.NoError(t, err)
		assert.Nil(t, log)
	})

	t.Run("retrieves log with validation errors", func(t *testing.T) {
		// Create a log with validation errors
		validationErrors := []models.ValidationError{
			{Row: 1, Field: "name", Error: "Name is required"},
			{Row: 2, Field: "party", Error: "Party not found"},
		}
		errorsJSON, _ := json.Marshal(validationErrors)

		createdLog := &models.PoliticianImportLog{
			Filename:  "test_with_errors.xlsx",
			Status:    "completed",
			TotalRows: 100,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, createdLog)
		require.NoError(t, err)

		// Update with validation errors
		_, err = pool.Exec(ctx, "UPDATE politician_import_logs SET validation_errors = $1 WHERE id = $2",
			errorsJSON, createdLog.ID)
		require.NoError(t, err)

		// Retrieve and verify
		retrievedLog, err := repo.GetByID(ctx, createdLog.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrievedLog)
		assert.Len(t, retrievedLog.ValidationErrors, 2)
		assert.Equal(t, "name", retrievedLog.ValidationErrors[0].Field)
	})
}

func TestImportRepository_List(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardownTestDB(t, pool)

	repo := NewImportRepository(pool)
	ctx := context.Background()

	t.Run("successfully lists import logs with pagination", func(t *testing.T) {
		// Create multiple test logs
		for i := 0; i < 5; i++ {
			log := &models.PoliticianImportLog{
				Filename:  "test_list_" + string(rune(i+'0')) + ".xlsx",
				Status:    "completed",
				TotalRows: i * 10,
				StartedAt: time.Now(),
				CreatedAt: time.Now(),
			}
			err := repo.Create(ctx, log)
			require.NoError(t, err)
			time.Sleep(1 * time.Millisecond) // Ensure different created_at times
		}

		// Get first page
		result, err := repo.List(ctx, 1, 3)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 5, result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 3, result.PerPage)
		assert.Equal(t, 2, result.TotalPages)
		assert.Len(t, result.ImportLogs, 3)
	})

	t.Run("handles second page correctly", func(t *testing.T) {
		result, err := repo.List(ctx, 2, 3)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.ImportLogs, 2)
	})

	t.Run("returns empty list when no logs exist", func(t *testing.T) {
		// Clean up all logs
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE politician_import_logs CASCADE")

		result, err := repo.List(ctx, 1, 10)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.Total)
		assert.Len(t, result.ImportLogs, 0)
	})

	t.Run("handles out of bounds page gracefully", func(t *testing.T) {
		// Create one log
		log := &models.PoliticianImportLog{
			Filename:  "test_outofbounds.xlsx",
			Status:    "completed",
			TotalRows: 10,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)

		// Request page 100
		result, err := repo.List(ctx, 100, 10)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.ImportLogs, 0)
	})
}

func TestImportRepository_UpdateStatus(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardownTestDB(t, pool)

	repo := NewImportRepository(pool)
	ctx := context.Background()

	t.Run("successfully updates status", func(t *testing.T) {
		// Create a log
		log := &models.PoliticianImportLog{
			Filename:  "test_status.xlsx",
			Status:    "pending",
			TotalRows: 10,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)

		// Update status
		err = repo.UpdateStatus(ctx, log.ID, "processing")
		require.NoError(t, err)

		// Verify update
		updated, err := repo.GetByID(ctx, log.ID)
		require.NoError(t, err)
		assert.Equal(t, "processing", updated.Status)
	})

	t.Run("updates status multiple times", func(t *testing.T) {
		log := &models.PoliticianImportLog{
			Filename:  "test_status_multi.xlsx",
			Status:    "pending",
			TotalRows: 10,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)

		statuses := []string{"processing", "completed", "failed"}
		for _, status := range statuses {
			err = repo.UpdateStatus(ctx, log.ID, status)
			require.NoError(t, err)

			updated, err := repo.GetByID(ctx, log.ID)
			require.NoError(t, err)
			assert.Equal(t, status, updated.Status)
		}
	})

	t.Run("no error for non-existent ID", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := repo.UpdateStatus(ctx, nonExistentID, "processing")
		assert.NoError(t, err) // SQL update with no rows affected is not an error
	})
}

func TestImportRepository_UpdateTotalRows(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardownTestDB(t, pool)

	repo := NewImportRepository(pool)
	ctx := context.Background()

	t.Run("successfully updates total rows", func(t *testing.T) {
		log := &models.PoliticianImportLog{
			Filename:  "test_rows.xlsx",
			Status:    "processing",
			TotalRows: 0,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)

		err = repo.UpdateTotalRows(ctx, log.ID, 150)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, log.ID)
		require.NoError(t, err)
		assert.Equal(t, 150, updated.TotalRows)
	})
}

func TestImportRepository_UpdateErrorLog(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardownTestDB(t, pool)

	repo := NewImportRepository(pool)
	ctx := context.Background()

	t.Run("successfully updates error log", func(t *testing.T) {
		log := &models.PoliticianImportLog{
			Filename:  "test_error.xlsx",
			Status:    "failed",
			TotalRows: 10,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)

		errorMsg := "Failed to parse Excel file: invalid format"
		err = repo.UpdateErrorLog(ctx, log.ID, errorMsg)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, log.ID)
		require.NoError(t, err)
		require.NotNil(t, updated.ErrorLog)
		assert.Equal(t, errorMsg, *updated.ErrorLog)
	})
}

func TestImportRepository_UpdateValidationErrors(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardownTestDB(t, pool)

	repo := NewImportRepository(pool)
	ctx := context.Background()

	t.Run("successfully updates validation errors", func(t *testing.T) {
		log := &models.PoliticianImportLog{
			Filename:  "test_validation.xlsx",
			Status:    "completed",
			TotalRows: 10,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)

		errors := []models.ValidationError{
			{Row: 1, Field: "name", Error: "Name is required", Value: nil},
			{Row: 3, Field: "party", Error: "Party not found", Value: stringPtr("ABC Party")},
		}

		err = repo.UpdateValidationErrors(ctx, log.ID, errors)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, log.ID)
		require.NoError(t, err)
		assert.Len(t, updated.ValidationErrors, 2)
		assert.Equal(t, 1, updated.ValidationErrors[0].Row)
		assert.Equal(t, "name", updated.ValidationErrors[0].Field)
	})

	t.Run("handles empty validation errors", func(t *testing.T) {
		log := &models.PoliticianImportLog{
			Filename:  "test_empty_errors.xlsx",
			Status:    "completed",
			TotalRows: 10,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)

		err = repo.UpdateValidationErrors(ctx, log.ID, []models.ValidationError{})
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, log.ID)
		require.NoError(t, err)
		assert.Len(t, updated.ValidationErrors, 0)
	})
}

func TestImportRepository_UpdateStatistics(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer teardownTestDB(t, pool)

	repo := NewImportRepository(pool)
	ctx := context.Background()

	t.Run("successfully updates all statistics", func(t *testing.T) {
		log := &models.PoliticianImportLog{
			Filename:  "test_stats.xlsx",
			Status:    "processing",
			TotalRows: 100,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)

		completedAt := time.Now()
		stats := &models.ImportStatistics{
			SuccessfulImports:  95,
			FailedImports:      5,
			PoliticiansCreated: 80,
			PoliticiansUpdated: 15,
			PositionsArchived:  10,
			CompletedAt:        &completedAt,
		}

		err = repo.UpdateStatistics(ctx, log.ID, stats)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, log.ID)
		require.NoError(t, err)
		assert.Equal(t, 95, updated.SuccessfulImports)
		assert.Equal(t, 5, updated.FailedImports)
		assert.Equal(t, 80, updated.PoliticiansCreated)
		assert.Equal(t, 15, updated.PoliticiansUpdated)
		assert.Equal(t, 10, updated.PositionsArchived)
		assert.NotNil(t, updated.CompletedAt)
	})

	t.Run("updates statistics with zero values", func(t *testing.T) {
		log := &models.PoliticianImportLog{
			Filename:  "test_zero_stats.xlsx",
			Status:    "failed",
			TotalRows: 100,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)

		completedAt := time.Now()
		stats := &models.ImportStatistics{
			SuccessfulImports:  0,
			FailedImports:      100,
			PoliticiansCreated: 0,
			PoliticiansUpdated: 0,
			PositionsArchived:  0,
			CompletedAt:        &completedAt,
		}

		err = repo.UpdateStatistics(ctx, log.ID, stats)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, log.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, updated.SuccessfulImports)
		assert.Equal(t, 100, updated.FailedImports)
	})
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}
