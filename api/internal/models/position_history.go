package models

import (
	"time"

	"github.com/google/uuid"
)

// PoliticianPositionHistory tracks all position assignments over time
type PoliticianPositionHistory struct {
	ID           uuid.UUID  `json:"id"`
	PoliticianID uuid.UUID  `json:"politician_id"`
	PositionID   uuid.UUID  `json:"position_id"`
	PartyID      *uuid.UUID `json:"party_id,omitempty"`

	// Jurisdiction fields (polymorphic)
	RegionID   *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID *uuid.UUID `json:"province_id,omitempty"`
	CityID     *uuid.UUID `json:"city_id,omitempty"`
	BarangayID *uuid.UUID `json:"barangay_id,omitempty"`
	DistrictID *uuid.UUID `json:"district_id,omitempty"`
	IsNational bool       `json:"is_national"`

	// Term dates
	TermStart time.Time  `json:"term_start"`
	TermEnd   *time.Time `json:"term_end,omitempty"`

	// Status
	IsCurrent   bool    `json:"is_current"`
	EndedReason *string `json:"ended_reason,omitempty"`

	// Election linkage
	ElectionID *uuid.UUID `json:"election_id,omitempty"`

	// Metadata
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`

	// Joined objects (populated in queries)
	Politician *PoliticianBrief       `json:"politician,omitempty"`
	Position   *GovernmentPosition    `json:"position,omitempty"`
	Party      *PartyBrief            `json:"party,omitempty"`
	Election   *ElectionEventBrief    `json:"election,omitempty"`
	Region     *Region                `json:"region,omitempty"`
	Province   *Province              `json:"province,omitempty"`
	City       *CityMunicipality      `json:"city,omitempty"`
	Barangay   *Barangay              `json:"barangay,omitempty"`
	District   *CongressionalDistrict `json:"district,omitempty"`
}

// ElectionEventBrief is a lightweight election representation
type ElectionEventBrief struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ElectionDate time.Time `json:"election_date"`
	Level        string    `json:"level"`
}

// PositionHistoryListItem is used for listing position history
type PositionHistoryListItem struct {
	ID           uuid.UUID  `json:"id"`
	PoliticianID uuid.UUID  `json:"politician_id"`
	PositionID   uuid.UUID  `json:"position_id"`
	TermStart    time.Time  `json:"term_start"`
	TermEnd      *time.Time `json:"term_end,omitempty"`
	IsCurrent    bool       `json:"is_current"`
	EndedReason  *string    `json:"ended_reason,omitempty"`

	// Simplified joined data
	PoliticianName string  `json:"politician_name"`
	PoliticianSlug string  `json:"politician_slug"`
	PositionName   string  `json:"position_name"`
	PartyName      *string `json:"party_name,omitempty"`
	PartyColor     *string `json:"party_color,omitempty"`
}

// CreatePositionHistoryRequest for assigning a position to a politician
type CreatePositionHistoryRequest struct {
	PoliticianID uuid.UUID  `json:"politician_id" validate:"required"`
	PositionID   uuid.UUID  `json:"position_id" validate:"required"`
	PartyID      *uuid.UUID `json:"party_id,omitempty"`

	// Jurisdiction (only one should be set based on position level)
	RegionID   *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID *uuid.UUID `json:"province_id,omitempty"`
	CityID     *uuid.UUID `json:"city_id,omitempty"`
	BarangayID *uuid.UUID `json:"barangay_id,omitempty"`
	DistrictID *uuid.UUID `json:"district_id,omitempty"`
	IsNational bool       `json:"is_national"`

	TermStart  time.Time  `json:"term_start" validate:"required"`
	TermEnd    *time.Time `json:"term_end,omitempty"`
	ElectionID *uuid.UUID `json:"election_id,omitempty"`

	// Indicates whether to create new history entry even if same politician
	WithHistory bool `json:"with_history"` // true = always create new record, false = update if same politician
}

// UpdatePositionHistoryRequest for updating a history entry
type UpdatePositionHistoryRequest struct {
	PartyID   *uuid.UUID `json:"party_id,omitempty"`
	TermStart *time.Time `json:"term_start,omitempty"`
	TermEnd   *time.Time `json:"term_end,omitempty"`
	IsCurrent *bool      `json:"is_current,omitempty"`
}

// EndTermRequest for ending a politician's current term
type EndTermRequest struct {
	EndDate     time.Time `json:"end_date" validate:"required"`
	EndedReason string    `json:"ended_reason" validate:"required,oneof=term_expired resigned replaced election deceased other"`
}

// GetCurrentHolderRequest for finding who holds a position in a jurisdiction
type GetCurrentHolderRequest struct {
	PositionID uuid.UUID  `json:"position_id" validate:"required"`
	RegionID   *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID *uuid.UUID `json:"province_id,omitempty"`
	CityID     *uuid.UUID `json:"city_id,omitempty"`
	BarangayID *uuid.UUID `json:"barangay_id,omitempty"`
	DistrictID *uuid.UUID `json:"district_id,omitempty"`
	IsNational bool       `json:"is_national"`
}

// PoliticianPositionTimeline represents a politician's career timeline
type PoliticianPositionTimeline struct {
	PoliticianID    uuid.UUID                 `json:"politician_id"`
	PoliticianName  string                    `json:"politician_name"`
	PoliticianSlug  string                    `json:"politician_slug"`
	CurrentPosition *PositionHistoryListItem  `json:"current_position,omitempty"`
	PastPositions   []PositionHistoryListItem `json:"past_positions"`
	TotalPositions  int                       `json:"total_positions"`
}
