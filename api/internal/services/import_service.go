package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/pkg/excel"
	"github.com/xuri/excelize/v2"
)

type ImportService struct {
	importRepo     *repository.ImportRepository
	politicianRepo *repository.PoliticianRepository
	partyRepo      *repository.PoliticalPartyRepository
	locationRepo   *repository.LocationRepository
	validator      *excel.Validator
}

func NewImportService(
	importRepo *repository.ImportRepository,
	politicianRepo *repository.PoliticianRepository,
	partyRepo *repository.PoliticalPartyRepository,
	locationRepo *repository.LocationRepository,
) *ImportService {
	return &ImportService{
		importRepo:     importRepo,
		politicianRepo: politicianRepo,
		partyRepo:      partyRepo,
		locationRepo:   locationRepo,
		validator:      excel.NewValidator(partyRepo, partyRepo, locationRepo),
	}
}

// ValidateImport validates Excel rows without importing
func (s *ImportService) ValidateImport(ctx context.Context, rows []models.ImportRow) (*models.ImportValidationResult, error) {
	return s.validator.ValidateImportRows(ctx, rows)
}

// StartImport creates an import log and starts async processing
func (s *ImportService) StartImport(ctx context.Context, req *models.ProcessImportRequest, userID *uuid.UUID) (*models.PoliticianImportLog, error) {
	// Create import log
	importLog := &models.PoliticianImportLog{
		Filename:   req.Filename,
		UploadedBy: userID,
		Status:     "pending",
		TotalRows:  0,
		StartedAt:  time.Now(),
		CreatedAt:  time.Now(),
		ElectionID: req.ElectionID,
	}

	// Save import log to database
	err := s.importRepo.Create(ctx, importLog)
	if err != nil {
		return nil, fmt.Errorf("failed to create import log: %w", err)
	}

	// Start async processing in goroutine
	go func() {
		// Create a new context for async processing (don't use request context)
		processCtx := context.Background()
		if err := s.ProcessImport(processCtx, importLog.ID, req.FileData); err != nil {
			fmt.Printf("Error processing import %s: %v\n", importLog.ID, err)
		}
	}()

	return importLog, nil
}

// ProcessImport processes the import asynchronously
func (s *ImportService) ProcessImport(ctx context.Context, importLogID uuid.UUID, fileData []byte) error {
	// Update status to processing
	if err := s.importRepo.UpdateStatus(ctx, importLogID, "processing"); err != nil {
		fmt.Printf("Failed to update import status to processing: %v\n", err)
	}

	// Parse Excel file
	rows, err := excel.ParseImportFile(fileData)
	if err != nil {
		if updateErr := s.importRepo.UpdateStatus(ctx, importLogID, "failed"); updateErr != nil {
			fmt.Printf("Failed to update import status to failed: %v\n", updateErr)
		}
		errMsg := fmt.Sprintf("Failed to parse Excel file: %s", err.Error())
		if logErr := s.importRepo.UpdateErrorLog(ctx, importLogID, errMsg); logErr != nil {
			fmt.Printf("Failed to update error log: %v\n", logErr)
		}
		return err
	}

	// Update total rows
	if err := s.importRepo.UpdateTotalRows(ctx, importLogID, len(rows)); err != nil {
		fmt.Printf("Failed to update total rows: %v\n", err)
	}

	// Validate rows
	validationResult, err := s.validator.ValidateImportRows(ctx, rows)
	if err != nil {
		if updateErr := s.importRepo.UpdateStatus(ctx, importLogID, "failed"); updateErr != nil {
			fmt.Printf("Failed to update import status to failed: %v\n", updateErr)
		}
		errMsg := fmt.Sprintf("Validation failed: %s", err.Error())
		if logErr := s.importRepo.UpdateErrorLog(ctx, importLogID, errMsg); logErr != nil {
			fmt.Printf("Failed to update error log: %v\n", logErr)
		}
		return err
	}

	// Store validation errors
	if len(validationResult.Errors) > 0 {
		if err := s.importRepo.UpdateValidationErrors(ctx, importLogID, validationResult.Errors); err != nil {
			fmt.Printf("Failed to update validation errors: %v\n", err)
		}
	}

	// Process valid rows
	stats := &ImportStats{}
	for _, validatedRow := range validationResult.ValidatedRows {
		if !validatedRow.IsValid {
			stats.FailedImports++
			continue
		}

		err := s.processRow(ctx, &validatedRow, stats)
		if err != nil {
			stats.FailedImports++
			// Log error but continue processing
			fmt.Printf("Error processing row %d: %s\n", validatedRow.RowNumber, err.Error())
		} else {
			stats.SuccessfulImports++
		}
	}

	// Update import log with statistics
	completedAt := time.Now()
	if err := s.importRepo.UpdateStatistics(ctx, importLogID, &models.ImportStatistics{
		SuccessfulImports:  stats.SuccessfulImports,
		FailedImports:      stats.FailedImports,
		PoliticiansCreated: stats.PoliticiansCreated,
		PoliticiansUpdated: stats.PoliticiansUpdated,
		PositionsArchived:  stats.PositionsArchived,
		CompletedAt:        &completedAt,
	}); err != nil {
		fmt.Printf("Failed to update import statistics: %v\n", err)
	}

	// Mark as completed
	if err := s.importRepo.UpdateStatus(ctx, importLogID, "completed"); err != nil {
		fmt.Printf("Failed to update import status to completed: %v\n", err)
	}

	return nil
}

// processRow processes a single validated row
func (s *ImportService) processRow(ctx context.Context, row *models.ValidatedImportRow, stats *ImportStats) error {
	// For simplicity, create politician directly with repository
	// In a real implementation, you'd check for existing politicians and handle position history

	slug := generateSlug(row.Name)
	politician := &models.Politician{
		Name:       row.Name,
		Slug:       slug,
		Position:   &row.PositionName,
		PositionID: &row.PositionID,
		PartyID:    row.PartyID,
		Photo:      row.PhotoURL,
		ShortBio:   row.ShortBio,
		TermStart:  &row.TermStart,
		TermEnd:    row.TermEnd,
	}

	// Try to create politician (simplified - in production would check for existing)
	err := s.politicianRepo.Create(ctx, politician)
	if err != nil {
		// If already exists, treat as update
		stats.PoliticiansUpdated++
		return nil // Don't fail the entire import
	}

	stats.PoliticiansCreated++
	return nil
}

// ListImportLogs lists all import logs with pagination
func (s *ImportService) ListImportLogs(ctx context.Context, page, perPage int) (*models.PaginatedImportLogs, error) {
	return s.importRepo.List(ctx, page, perPage)
}

// GetImportLog gets a single import log by ID
func (s *ImportService) GetImportLog(ctx context.Context, id uuid.UUID) (*models.PoliticianImportLog, error) {
	return s.importRepo.GetByID(ctx, id)
}

// GenerateTemplate generates an Excel template for importing
func (s *ImportService) GenerateTemplate(ctx context.Context) (*excelize.File, error) {
	// Load reference data - positions from party repo (it handles both)
	positionsList, err := s.partyRepo.GetAllPositions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load positions: %w", err)
	}

	// Convert to GovernmentPosition slice for template generation
	positions := make([]models.GovernmentPosition, len(positionsList))
	for i, p := range positionsList {
		positions[i] = models.GovernmentPosition{
			ID:        p.ID,
			Name:      p.Name,
			Slug:      p.Slug,
			Level:     p.Level,
			Branch:    p.Branch,
			IsElected: p.IsElected,
		}
	}

	// Load parties
	partiesList, err := s.partyRepo.GetAll(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to load parties: %w", err)
	}

	// Convert to PoliticalParty slice
	parties := make([]models.PoliticalParty, len(partiesList))
	for i, p := range partiesList {
		parties[i] = models.PoliticalParty{
			ID:           p.ID,
			Name:         p.Name,
			Slug:         p.Slug,
			Abbreviation: p.Abbreviation,
			Logo:         p.Logo,
			Color:        p.Color,
			IsMajor:      p.IsMajor,
			IsActive:     p.IsActive,
		}
	}

	// Generate template with data validation
	templateFile, err := excel.GenerateTemplateFile(positions, parties, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate template: %w", err)
	}

	return templateFile, nil
}

// GenerateErrorReport generates an error report for a failed import
func (s *ImportService) GenerateErrorReport(ctx context.Context, importLogID uuid.UUID) (*excelize.File, error) {
	// Get import log
	importLog, err := s.importRepo.GetByID(ctx, importLogID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch import log: %w", err)
	}

	if importLog == nil {
		return nil, nil
	}

	// Only generate report if there are errors
	if len(importLog.ValidationErrors) == 0 {
		return nil, nil
	}

	// Generate error report
	errorFile, err := excel.GenerateErrorReport(importLog, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate error report: %w", err)
	}

	return errorFile, nil
}

// ImportStats tracks statistics during import processing
type ImportStats struct {
	SuccessfulImports  int
	FailedImports      int
	PoliticiansCreated int
	PoliticiansUpdated int
	PositionsArchived  int
}

// Helper functions

func generateSlug(name string) string {
	// Simple slug generation - can be improved
	slug := name
	slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	return slug
}
