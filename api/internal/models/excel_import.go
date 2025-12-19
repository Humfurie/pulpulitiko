package models

import (
	"time"

	"github.com/google/uuid"
)

// PoliticianImportLog tracks Excel import operations
type PoliticianImportLog struct {
	ID         uuid.UUID  `json:"id"`
	Filename   string     `json:"filename"`
	UploadedBy *uuid.UUID `json:"uploaded_by,omitempty"`

	// Statistics
	TotalRows          int `json:"total_rows"`
	SuccessfulImports  int `json:"successful_imports"`
	FailedImports      int `json:"failed_imports"`
	PoliticiansCreated int `json:"politicians_created"`
	PoliticiansUpdated int `json:"politicians_updated"`
	PositionsArchived  int `json:"positions_archived"`

	// Status
	Status           string            `json:"status"` // 'pending', 'processing', 'completed', 'failed'
	ErrorLog         *string           `json:"error_log,omitempty"`
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`

	// Election linkage
	ElectionID *uuid.UUID `json:"election_id,omitempty"`

	// Timestamps
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	// Joined objects
	Election       *ElectionEventBrief `json:"election,omitempty"`
	UploadedByUser *UserBrief          `json:"uploaded_by_user,omitempty"`
}

// UserBrief is a lightweight user representation
type UserBrief struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

// ValidationError represents an error in import validation
type ValidationError struct {
	Row         int      `json:"row"`
	Field       string   `json:"field"`
	Error       string   `json:"error"`
	Value       *string  `json:"value,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"` // Suggested fixes
}

// ImportRow represents a single row from Excel import
type ImportRow struct {
	RowNumber        int     `json:"row_number"`
	Name             string  `json:"name"`
	Position         string  `json:"position"`
	JurisdictionType string  `json:"jurisdiction_type"` // 'national', 'region', 'province', 'city', 'barangay', 'district'
	JurisdictionName string  `json:"jurisdiction_name"`
	Party            string  `json:"party"`
	TermStart        string  `json:"term_start"`         // Date string
	TermEnd          *string `json:"term_end,omitempty"` // Optional date string
	PhotoURL         *string `json:"photo_url,omitempty"`
	ShortBio         *string `json:"short_bio,omitempty"`
	BirthDate        *string `json:"birth_date,omitempty"`
}

// ValidatedImportRow is an ImportRow with resolved IDs
type ValidatedImportRow struct {
	RowNumber    int       `json:"row_number"`
	Name         string    `json:"name"`
	PositionID   uuid.UUID `json:"position_id"`
	PositionName string    `json:"position_name"`

	// Jurisdiction
	JurisdictionType string     `json:"jurisdiction_type"`
	RegionID         *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID       *uuid.UUID `json:"province_id,omitempty"`
	CityID           *uuid.UUID `json:"city_id,omitempty"`
	BarangayID       *uuid.UUID `json:"barangay_id,omitempty"`
	DistrictID       *uuid.UUID `json:"district_id,omitempty"`
	IsNational       bool       `json:"is_national"`

	PartyID   *uuid.UUID `json:"party_id,omitempty"`
	PartyName *string    `json:"party_name,omitempty"`
	TermStart time.Time  `json:"term_start"`
	TermEnd   *time.Time `json:"term_end,omitempty"`
	PhotoURL  *string    `json:"photo_url,omitempty"`
	ShortBio  *string    `json:"short_bio,omitempty"`
	BirthDate *time.Time `json:"birth_date,omitempty"`

	// Validation result
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// ImportValidationResult is the result of validating an Excel file
type ImportValidationResult struct {
	TotalRows     int                  `json:"total_rows"`
	ValidRows     int                  `json:"valid_rows"`
	InvalidRows   int                  `json:"invalid_rows"`
	Errors        []ValidationError    `json:"errors"`
	ValidatedRows []ValidatedImportRow `json:"validated_rows,omitempty"`
}

// ProcessImportRequest for importing politicians from Excel
type ProcessImportRequest struct {
	FileData     []byte     `json:"file_data" validate:"required"`
	Filename     string     `json:"filename" validate:"required"`
	ElectionID   *uuid.UUID `json:"election_id,omitempty"`
	ValidateOnly bool       `json:"validate_only"` // If true, only validate without importing
}

// ImportProgressUpdate for real-time progress updates (WebSocket/SSE)
type ImportProgressUpdate struct {
	ImportID          uuid.UUID `json:"import_id"`
	Status            string    `json:"status"`
	ProcessedRows     int       `json:"processed_rows"`
	TotalRows         int       `json:"total_rows"`
	SuccessfulImports int       `json:"successful_imports"`
	FailedImports     int       `json:"failed_imports"`
	CurrentRow        int       `json:"current_row"`
	Message           string    `json:"message,omitempty"`
}

// ExportPoliticiansRequest for exporting politicians to Excel
type ExportPoliticiansRequest struct {
	IncludeHistory bool       `json:"include_history"`       // Include historical positions
	PositionID     *uuid.UUID `json:"position_id,omitempty"` // Filter by position
	PartyID        *uuid.UUID `json:"party_id,omitempty"`    // Filter by party
	Level          *string    `json:"level,omitempty"`       // Filter by government level
	CurrentOnly    bool       `json:"current_only"`          // Only export current position holders
}

// ExcelTemplate defines the structure of the import template
type ExcelTemplate struct {
	SheetName      string
	Headers        []string
	DataValidation map[string][]string // Column name -> list of valid values
}

// PaginatedImportLogs for paginated list of import logs
type PaginatedImportLogs struct {
	ImportLogs []PoliticianImportLog `json:"import_logs"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PerPage    int                   `json:"per_page"`
	TotalPages int                   `json:"total_pages"`
}

// ImportStatistics for updating import log statistics
type ImportStatistics struct {
	SuccessfulImports  int
	FailedImports      int
	PoliticiansCreated int
	PoliticiansUpdated int
	PositionsArchived  int
	CompletedAt        *time.Time
}
