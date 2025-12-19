package models

import (
	"time"

	"github.com/google/uuid"
)

type PoliticalParty struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Abbreviation *string    `json:"abbreviation,omitempty"`
	Logo         *string    `json:"logo,omitempty"`
	Color        *string    `json:"color,omitempty"`
	Description  *string    `json:"description,omitempty"`
	FoundedYear  *int       `json:"founded_year,omitempty"`
	Website      *string    `json:"website,omitempty"`
	IsMajor      bool       `json:"is_major"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`

	// Computed fields
	MemberCount int `json:"member_count,omitempty"`
}

type PoliticalPartyListItem struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Abbreviation *string   `json:"abbreviation,omitempty"`
	Logo         *string   `json:"logo,omitempty"`
	Color        *string   `json:"color,omitempty"`
	IsMajor      bool      `json:"is_major"`
	IsActive     bool      `json:"is_active"`
	MemberCount  int       `json:"member_count"`
}

type CreatePoliticalPartyRequest struct {
	Name         string  `json:"name" validate:"required,min=2,max=200"`
	Slug         string  `json:"slug" validate:"required,min=2,max=200"`
	Abbreviation *string `json:"abbreviation,omitempty" validate:"omitempty,max=50"`
	Logo         *string `json:"logo,omitempty" validate:"omitempty,max=500"`
	Color        *string `json:"color,omitempty" validate:"omitempty,max=20"`
	Description  *string `json:"description,omitempty"`
	FoundedYear  *int    `json:"founded_year,omitempty"`
	Website      *string `json:"website,omitempty" validate:"omitempty,max=500"`
	IsMajor      bool    `json:"is_major"`
	IsActive     bool    `json:"is_active"`
}

type UpdatePoliticalPartyRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=2,max=200"`
	Slug         *string `json:"slug,omitempty" validate:"omitempty,min=2,max=200"`
	Abbreviation *string `json:"abbreviation,omitempty" validate:"omitempty,max=50"`
	Logo         *string `json:"logo,omitempty" validate:"omitempty,max=500"`
	Color        *string `json:"color,omitempty" validate:"omitempty,max=20"`
	Description  *string `json:"description,omitempty"`
	FoundedYear  *int    `json:"founded_year,omitempty"`
	Website      *string `json:"website,omitempty" validate:"omitempty,max=500"`
	IsMajor      *bool   `json:"is_major,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

type PaginatedPoliticalParties struct {
	Parties    []PoliticalPartyListItem `json:"parties"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	PerPage    int                      `json:"per_page"`
	TotalPages int                      `json:"total_pages"`
}

// Government Position represents a normalized position type
type GovernmentPosition struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Level        string    `json:"level"`  // national, regional, provincial, city, municipal, barangay
	Branch       string    `json:"branch"` // executive, legislative, judicial
	DisplayOrder int       `json:"display_order"`
	Description  *string   `json:"description,omitempty"`
	MaxTerms     *int      `json:"max_terms,omitempty"`
	TermYears    int       `json:"term_years"`
	IsElected    bool      `json:"is_elected"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GovernmentPositionListItem struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Level        string    `json:"level"`
	Branch       string    `json:"branch"`
	DisplayOrder int       `json:"display_order"`
	IsElected    bool      `json:"is_elected"`
}

type CreateGovernmentPositionRequest struct {
	Name         string  `json:"name" validate:"required"`
	Slug         string  `json:"slug" validate:"required"`
	Level        string  `json:"level" validate:"required,oneof=national regional provincial city municipal barangay district"`
	Branch       string  `json:"branch" validate:"required,oneof=executive legislative judicial"`
	DisplayOrder int     `json:"display_order" validate:"gte=0"`
	Description  *string `json:"description,omitempty"`
	MaxTerms     *int    `json:"max_terms,omitempty" validate:"omitempty,gte=0"`
	TermYears    int     `json:"term_years" validate:"required,gte=1"`
	IsElected    bool    `json:"is_elected"`
}

type UpdateGovernmentPositionRequest struct {
	Name         *string `json:"name,omitempty"`
	Slug         *string `json:"slug,omitempty"`
	Level        *string `json:"level,omitempty" validate:"omitempty,oneof=national regional provincial city municipal barangay district"`
	Branch       *string `json:"branch,omitempty" validate:"omitempty,oneof=executive legislative judicial"`
	DisplayOrder *int    `json:"display_order,omitempty" validate:"omitempty,gte=0"`
	Description  *string `json:"description,omitempty"`
	MaxTerms     *int    `json:"max_terms,omitempty" validate:"omitempty,gte=0"`
	TermYears    *int    `json:"term_years,omitempty" validate:"omitempty,gte=1"`
	IsElected    *bool   `json:"is_elected,omitempty"`
}

// Politician Jurisdiction maps a politician to their jurisdiction
type PoliticianJurisdiction struct {
	ID           uuid.UUID  `json:"id"`
	PoliticianID uuid.UUID  `json:"politician_id"`
	RegionID     *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID   *uuid.UUID `json:"province_id,omitempty"`
	CityID       *uuid.UUID `json:"city_id,omitempty"`
	BarangayID   *uuid.UUID `json:"barangay_id,omitempty"`
	IsNational   bool       `json:"is_national"`
	CreatedAt    time.Time  `json:"created_at"`

	// Joined fields
	Region   *RegionListItem           `json:"region,omitempty"`
	Province *ProvinceListItem         `json:"province,omitempty"`
	City     *CityMunicipalityListItem `json:"city,omitempty"`
	Barangay *BarangayListItem         `json:"barangay,omitempty"`
}

type CreatePoliticianJurisdictionRequest struct {
	PoliticianID uuid.UUID  `json:"politician_id" validate:"required"`
	RegionID     *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID   *uuid.UUID `json:"province_id,omitempty"`
	CityID       *uuid.UUID `json:"city_id,omitempty"`
	BarangayID   *uuid.UUID `json:"barangay_id,omitempty"`
	IsNational   bool       `json:"is_national"`
}
