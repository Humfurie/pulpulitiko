package models

import (
	"time"

	"github.com/google/uuid"
)

// ElectionEvent tracks elections that trigger bulk position changes
type ElectionEvent struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	ElectionDate time.Time  `json:"election_date"`
	Level        string     `json:"level"`  // 'national', 'local', 'barangay', 'regional'
	Status       string     `json:"status"` // 'scheduled', 'in_progress', 'completed', 'cancelled'
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty"`

	// Computed fields
	TotalPositions int `json:"total_positions,omitempty"` // Number of positions affected
	ImportedCount  int `json:"imported_count,omitempty"`  // Number of politicians imported for this election
}

// ElectionEventListItem for listing elections
type ElectionEventListItem struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	ElectionDate   time.Time `json:"election_date"`
	Level          string    `json:"level"`
	Status         string    `json:"status"`
	TotalPositions int       `json:"total_positions"`
	ImportedCount  int       `json:"imported_count"`
}

// CreateElectionEventRequest for creating an election
type CreateElectionEventRequest struct {
	Name         string    `json:"name" validate:"required,min=2,max=200"`
	Description  *string   `json:"description,omitempty"`
	ElectionDate time.Time `json:"election_date" validate:"required"`
	Level        string    `json:"level" validate:"required,oneof=national local barangay regional provincial city municipal"`
}

// UpdateElectionEventRequest for updating an election
type UpdateElectionEventRequest struct {
	Name         *string    `json:"name,omitempty" validate:"omitempty,min=2,max=200"`
	Description  *string    `json:"description,omitempty"`
	ElectionDate *time.Time `json:"election_date,omitempty"`
	Level        *string    `json:"level,omitempty" validate:"omitempty,oneof=national local barangay regional provincial city municipal"`
	Status       *string    `json:"status,omitempty" validate:"omitempty,oneof=scheduled in_progress completed cancelled"`
}

// ProcessElectionResultsRequest for importing election results
type ProcessElectionResultsRequest struct {
	ElectionID uuid.UUID `json:"election_id" validate:"required"`
	FileData   []byte    `json:"file_data" validate:"required"`
	Filename   string    `json:"filename" validate:"required"`
}

// ElectionStatistics provides statistics about an election
type ElectionStatistics struct {
	ElectionID          uuid.UUID             `json:"election_id"`
	ElectionName        string                `json:"election_name"`
	TotalPositions      int                   `json:"total_positions"`
	PositionsArchived   int                   `json:"positions_archived"`
	PoliticiansImported int                   `json:"politicians_imported"`
	ImportLogs          []PoliticianImportLog `json:"import_logs,omitempty"`
}

// PaginatedElectionEvents for paginated election event lists
type PaginatedElectionEvents struct {
	ElectionEvents []ElectionEventListItem `json:"data"`
	Total          int                     `json:"total"`
	Page           int                     `json:"page"`
	PerPage        int                     `json:"per_page"`
	TotalPages     int                     `json:"total_pages"`
}
