package models

import (
	"time"

	"github.com/google/uuid"
)

type Politician struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Slug         string       `json:"slug"`
	Photo        *string      `json:"photo,omitempty"`
	Position     *string      `json:"position,omitempty"`
	Party        *string      `json:"party,omitempty"`
	ShortBio     *string      `json:"short_bio,omitempty"`
	FullBio      *string      `json:"full_bio,omitempty"`
	Mission      *string      `json:"mission,omitempty"`
	Vision       *string      `json:"vision,omitempty"`
	Achievements *string      `json:"achievements,omitempty"`
	TermStart    *time.Time   `json:"term_start,omitempty"`
	TermEnd      *time.Time   `json:"term_end,omitempty"`
	BirthDate    *time.Time   `json:"birth_date,omitempty"`
	BirthPlace   *string      `json:"birth_place,omitempty"`
	Education    *string      `json:"education,omitempty"`
	Website      *string      `json:"website,omitempty"`
	SocialLinks  *SocialLinks `json:"social_links,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	DeletedAt    *time.Time   `json:"deleted_at,omitempty"`

	// Government structure fields
	Level      *string    `json:"level,omitempty"`       // national, regional, provincial, city, municipal, barangay
	Branch     *string    `json:"branch,omitempty"`      // executive, legislative, judicial
	PositionID *uuid.UUID `json:"position_id,omitempty"` // FK to government_positions
	PartyID    *uuid.UUID `json:"party_id,omitempty"`    // FK to political_parties
	DistrictID *uuid.UUID `json:"district_id,omitempty"` // FK to congressional_districts

	// Computed fields (populated when needed)
	ArticleCount int                     `json:"article_count,omitempty"`
	PartyInfo    *PartyBrief             `json:"party_info,omitempty"`
	PositionInfo *GovernmentPositionInfo `json:"position_info,omitempty"`
}

// GovernmentPositionInfo is a lightweight version for embedding in Politician
type GovernmentPositionInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Level     string    `json:"level"`
	Branch    string    `json:"branch"`
	IsElected bool      `json:"is_elected"`
}

// PartyBrief is a lightweight version of PoliticalParty for embedding
type PartyBrief struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Abbreviation *string   `json:"abbreviation,omitempty"`
	Logo         *string   `json:"logo,omitempty"`
	Color        *string   `json:"color,omitempty"`
}

type PoliticianListItem struct {
	ID           uuid.UUID   `json:"id"`
	Name         string      `json:"name"`
	Slug         string      `json:"slug"`
	Photo        *string     `json:"photo,omitempty"`
	Position     *string     `json:"position,omitempty"`
	Party        *string     `json:"party,omitempty"`
	Level        *string     `json:"level,omitempty"`
	Branch       *string     `json:"branch,omitempty"`
	TermStart    *time.Time  `json:"term_start,omitempty"`
	TermEnd      *time.Time  `json:"term_end,omitempty"`
	ArticleCount int         `json:"article_count"`
	PartyInfo    *PartyBrief `json:"party_info,omitempty"`
}

type CreatePoliticianRequest struct {
	Name       string     `json:"name" validate:"required,min=2,max=200"`
	Slug       string     `json:"slug" validate:"required,min=2,max=200"`
	Photo      *string    `json:"photo,omitempty" validate:"omitempty,max=500"`
	Position   *string    `json:"position,omitempty" validate:"omitempty,max=200"`
	Party      *string    `json:"party,omitempty" validate:"omitempty,max=200"`
	ShortBio   *string    `json:"short_bio,omitempty"`
	TermStart  *string    `json:"term_start,omitempty"` // Format: YYYY-MM-DD
	TermEnd    *string    `json:"term_end,omitempty"`   // Format: YYYY-MM-DD
	Level      *string    `json:"level,omitempty"`      // national, regional, provincial, city, municipal, barangay
	Branch     *string    `json:"branch,omitempty"`     // executive, legislative, judicial
	PositionID *uuid.UUID `json:"position_id,omitempty"`
	PartyID    *uuid.UUID `json:"party_id,omitempty"`
	DistrictID *uuid.UUID `json:"district_id,omitempty"`
}

type UpdatePoliticianRequest struct {
	Name       *string    `json:"name,omitempty" validate:"omitempty,min=2,max=200"`
	Slug       *string    `json:"slug,omitempty" validate:"omitempty,min=2,max=200"`
	Photo      *string    `json:"photo,omitempty" validate:"omitempty,max=500"`
	Position   *string    `json:"position,omitempty" validate:"omitempty,max=200"`
	Party      *string    `json:"party,omitempty" validate:"omitempty,max=200"`
	ShortBio   *string    `json:"short_bio,omitempty"`
	TermStart  *string    `json:"term_start,omitempty"` // Format: YYYY-MM-DD
	TermEnd    *string    `json:"term_end,omitempty"`   // Format: YYYY-MM-DD
	Level      *string    `json:"level,omitempty"`
	Branch     *string    `json:"branch,omitempty"`
	PositionID *uuid.UUID `json:"position_id,omitempty"`
	PartyID    *uuid.UUID `json:"party_id,omitempty"`
	DistrictID *uuid.UUID `json:"district_id,omitempty"`
}

type PoliticianFilter struct {
	Search         *string
	Party          *string
	Level          *string
	Branch         *string
	PartyID        *uuid.UUID
	PositionID     *uuid.UUID
	LocationID     *uuid.UUID // Can be region, province, city, or barangay
	LocationType   *string    // "region", "province", "city", "barangay"
	IncludeDeleted bool
}

type PaginatedPoliticians struct {
	Politicians []PoliticianListItem `json:"politicians"`
	Total       int                  `json:"total"`
	Page        int                  `json:"page"`
	PerPage     int                  `json:"per_page"`
	TotalPages  int                  `json:"total_pages"`
}
