package models

import (
	"time"

	"github.com/google/uuid"
)

// Election Type constants
const (
	ElectionTypeNational   = "national"
	ElectionTypeLocal      = "local"
	ElectionTypeBarangay   = "barangay"
	ElectionTypeSpecial    = "special"
	ElectionTypePlebiscite = "plebiscite"
	ElectionTypeRecall     = "recall"
)

// Election Status constants
const (
	ElectionStatusUpcoming  = "upcoming"
	ElectionStatusOngoing   = "ongoing"
	ElectionStatusCompleted = "completed"
	ElectionStatusCancelled = "cancelled"
)

// Candidate Status constants
const (
	CandidateStatusFiled        = "filed"
	CandidateStatusQualified    = "qualified"
	CandidateStatusDisqualified = "disqualified"
	CandidateStatusWithdrawn    = "withdrawn"
	CandidateStatusSubstituted  = "substituted"
)

// Election represents an election event
type Election struct {
	ID                       uuid.UUID  `json:"id"`
	Name                     string     `json:"name"`
	Slug                     string     `json:"slug"`
	ElectionType             string     `json:"election_type"`
	Description              *string    `json:"description,omitempty"`
	ElectionDate             time.Time  `json:"election_date"`
	RegistrationStart        *time.Time `json:"registration_start,omitempty"`
	RegistrationEnd          *time.Time `json:"registration_end,omitempty"`
	CampaignStart            *time.Time `json:"campaign_start,omitempty"`
	CampaignEnd              *time.Time `json:"campaign_end,omitempty"`
	Status                   string     `json:"status"`
	IsFeatured               bool       `json:"is_featured"`
	VoterTurnoutPercentage   *float64   `json:"voter_turnout_percentage,omitempty"`
	TotalRegisteredVoters    *int       `json:"total_registered_voters,omitempty"`
	TotalVotesCast           *int       `json:"total_votes_cast,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	DeletedAt                *time.Time `json:"deleted_at,omitempty"`

	// Joined fields
	Positions  []ElectionPositionListItem `json:"positions,omitempty"`
	Candidates []CandidateListItem        `json:"candidates,omitempty"`
}

type ElectionListItem struct {
	ID                     uuid.UUID `json:"id"`
	Name                   string    `json:"name"`
	Slug                   string    `json:"slug"`
	ElectionType           string    `json:"election_type"`
	ElectionDate           time.Time `json:"election_date"`
	Status                 string    `json:"status"`
	IsFeatured             bool      `json:"is_featured"`
	VoterTurnoutPercentage *float64  `json:"voter_turnout_percentage,omitempty"`
	PositionCount          int       `json:"position_count"`
	CandidateCount         int       `json:"candidate_count"`
}

// ElectionPosition represents a position being contested in an election
type ElectionPosition struct {
	ID                 uuid.UUID  `json:"id"`
	ElectionID         uuid.UUID  `json:"election_id"`
	PositionID         uuid.UUID  `json:"position_id"`
	RegionID           *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID         *uuid.UUID `json:"province_id,omitempty"`
	CityMunicipalityID *uuid.UUID `json:"city_municipality_id,omitempty"`
	BarangayID         *uuid.UUID `json:"barangay_id,omitempty"`
	DistrictID         *uuid.UUID `json:"district_id,omitempty"`
	SeatsAvailable     int        `json:"seats_available"`
	Description        *string    `json:"description,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`

	// Joined fields
	Position   *GovernmentPositionInfo `json:"position,omitempty"`
	Location   *string                 `json:"location,omitempty"` // Human-readable location
	Candidates []CandidateListItem     `json:"candidates,omitempty"`
}

type ElectionPositionListItem struct {
	ID             uuid.UUID               `json:"id"`
	PositionID     uuid.UUID               `json:"position_id"`
	SeatsAvailable int                     `json:"seats_available"`
	Position       *GovernmentPositionInfo `json:"position,omitempty"`
	Location       *string                 `json:"location,omitempty"`
	CandidateCount int                     `json:"candidate_count"`
}

// Candidate represents a candidate for an election position
type Candidate struct {
	ID                 uuid.UUID  `json:"id"`
	ElectionPositionID uuid.UUID  `json:"election_position_id"`
	PoliticianID       uuid.UUID  `json:"politician_id"`
	PartyID            *uuid.UUID `json:"party_id,omitempty"`
	BallotNumber       *int       `json:"ballot_number,omitempty"`
	BallotName         *string    `json:"ballot_name,omitempty"`
	CampaignSlogan     *string    `json:"campaign_slogan,omitempty"`
	Platform           *string    `json:"platform,omitempty"`
	Status             string     `json:"status"`
	FilingDate         *time.Time `json:"filing_date,omitempty"`
	IsIncumbent        bool       `json:"is_incumbent"`
	IsWinner           bool       `json:"is_winner"`
	VotesReceived      *int       `json:"votes_received,omitempty"`
	VotePercentage     *float64   `json:"vote_percentage,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Joined fields
	Politician *PoliticianListItem `json:"politician,omitempty"`
	Party      *PartyBrief         `json:"party,omitempty"`
}

type CandidateListItem struct {
	ID             uuid.UUID           `json:"id"`
	PoliticianID   uuid.UUID           `json:"politician_id"`
	BallotNumber   *int                `json:"ballot_number,omitempty"`
	BallotName     *string             `json:"ballot_name,omitempty"`
	Status         string              `json:"status"`
	IsIncumbent    bool                `json:"is_incumbent"`
	IsWinner       bool                `json:"is_winner"`
	VotesReceived  *int                `json:"votes_received,omitempty"`
	VotePercentage *float64            `json:"vote_percentage,omitempty"`
	Politician     *PoliticianListItem `json:"politician,omitempty"`
	Party          *PartyBrief         `json:"party,omitempty"`
}

// ElectionResult represents aggregate results for a position
type ElectionResult struct {
	ID                 uuid.UUID `json:"id"`
	ElectionPositionID uuid.UUID `json:"election_position_id"`
	TotalVotes         int       `json:"total_votes"`
	ValidVotes         int       `json:"valid_votes"`
	InvalidVotes       int       `json:"invalid_votes"`
	RegisteredVoters   *int      `json:"registered_voters,omitempty"`
	TurnoutPercentage  *float64  `json:"turnout_percentage,omitempty"`
	IsFinal            bool      `json:"is_final"`
	LastUpdated        time.Time `json:"last_updated"`
	CreatedAt          time.Time `json:"created_at"`
}

// PrecinctResult represents results from a specific precinct
type PrecinctResult struct {
	ID           uuid.UUID  `json:"id"`
	CandidateID  uuid.UUID  `json:"candidate_id"`
	PrecinctID   string     `json:"precinct_id"`
	PrecinctName *string    `json:"precinct_name,omitempty"`
	BarangayID   *uuid.UUID `json:"barangay_id,omitempty"`
	Votes        int        `json:"votes"`
	CreatedAt    time.Time  `json:"created_at"`
}

// VoterEducation represents educational content for voters
type VoterEducation struct {
	ID          uuid.UUID  `json:"id"`
	ElectionID  *uuid.UUID `json:"election_id,omitempty"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	ContentType string     `json:"content_type"`
	Category    *string    `json:"category,omitempty"`
	IsFeatured  bool       `json:"is_featured"`
	IsPublished bool       `json:"is_published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ViewCount   int        `json:"view_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`

	// Joined fields
	Election *ElectionListItem `json:"election,omitempty"`
}

type VoterEducationListItem struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	ContentType string     `json:"content_type"`
	Category    *string    `json:"category,omitempty"`
	IsFeatured  bool       `json:"is_featured"`
	ViewCount   int        `json:"view_count"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

// Request types

type CreateElectionRequest struct {
	Name              string  `json:"name" validate:"required,max=300"`
	Slug              string  `json:"slug" validate:"required,max=300"`
	ElectionType      string  `json:"election_type" validate:"required,oneof=national local barangay special plebiscite recall"`
	Description       *string `json:"description,omitempty"`
	ElectionDate      string  `json:"election_date" validate:"required"` // YYYY-MM-DD
	RegistrationStart *string `json:"registration_start,omitempty"`      // YYYY-MM-DD
	RegistrationEnd   *string `json:"registration_end,omitempty"`        // YYYY-MM-DD
	CampaignStart     *string `json:"campaign_start,omitempty"`          // YYYY-MM-DD
	CampaignEnd       *string `json:"campaign_end,omitempty"`            // YYYY-MM-DD
	Status            string  `json:"status" validate:"required,oneof=upcoming ongoing completed cancelled"`
	IsFeatured        bool    `json:"is_featured"`
}

type UpdateElectionRequest struct {
	Name                     *string  `json:"name,omitempty" validate:"omitempty,max=300"`
	Slug                     *string  `json:"slug,omitempty" validate:"omitempty,max=300"`
	Description              *string  `json:"description,omitempty"`
	ElectionDate             *string  `json:"election_date,omitempty"`     // YYYY-MM-DD
	RegistrationStart        *string  `json:"registration_start,omitempty"` // YYYY-MM-DD
	RegistrationEnd          *string  `json:"registration_end,omitempty"`   // YYYY-MM-DD
	CampaignStart            *string  `json:"campaign_start,omitempty"`     // YYYY-MM-DD
	CampaignEnd              *string  `json:"campaign_end,omitempty"`       // YYYY-MM-DD
	Status                   *string  `json:"status,omitempty" validate:"omitempty,oneof=upcoming ongoing completed cancelled"`
	IsFeatured               *bool    `json:"is_featured,omitempty"`
	VoterTurnoutPercentage   *float64 `json:"voter_turnout_percentage,omitempty"`
	TotalRegisteredVoters    *int     `json:"total_registered_voters,omitempty"`
	TotalVotesCast           *int     `json:"total_votes_cast,omitempty"`
}

type CreateElectionPositionRequest struct {
	ElectionID         uuid.UUID  `json:"election_id" validate:"required"`
	PositionID         uuid.UUID  `json:"position_id" validate:"required"`
	RegionID           *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID         *uuid.UUID `json:"province_id,omitempty"`
	CityMunicipalityID *uuid.UUID `json:"city_municipality_id,omitempty"`
	BarangayID         *uuid.UUID `json:"barangay_id,omitempty"`
	DistrictID         *uuid.UUID `json:"district_id,omitempty"`
	SeatsAvailable     int        `json:"seats_available" validate:"min=1"`
	Description        *string    `json:"description,omitempty"`
}

type CreateCandidateRequest struct {
	ElectionPositionID uuid.UUID  `json:"election_position_id" validate:"required"`
	PoliticianID       uuid.UUID  `json:"politician_id" validate:"required"`
	PartyID            *uuid.UUID `json:"party_id,omitempty"`
	BallotNumber       *int       `json:"ballot_number,omitempty"`
	BallotName         *string    `json:"ballot_name,omitempty" validate:"omitempty,max=200"`
	CampaignSlogan     *string    `json:"campaign_slogan,omitempty" validate:"omitempty,max=500"`
	Platform           *string    `json:"platform,omitempty"`
	Status             string     `json:"status" validate:"required,oneof=filed qualified disqualified withdrawn substituted"`
	FilingDate         *string    `json:"filing_date,omitempty"` // YYYY-MM-DD
	IsIncumbent        bool       `json:"is_incumbent"`
}

type UpdateCandidateRequest struct {
	PartyID        *uuid.UUID `json:"party_id,omitempty"`
	BallotNumber   *int       `json:"ballot_number,omitempty"`
	BallotName     *string    `json:"ballot_name,omitempty" validate:"omitempty,max=200"`
	CampaignSlogan *string    `json:"campaign_slogan,omitempty" validate:"omitempty,max=500"`
	Platform       *string    `json:"platform,omitempty"`
	Status         *string    `json:"status,omitempty" validate:"omitempty,oneof=filed qualified disqualified withdrawn substituted"`
	IsIncumbent    *bool      `json:"is_incumbent,omitempty"`
	IsWinner       *bool      `json:"is_winner,omitempty"`
	VotesReceived  *int       `json:"votes_received,omitempty"`
	VotePercentage *float64   `json:"vote_percentage,omitempty"`
}

type CreateVoterEducationRequest struct {
	ElectionID  *uuid.UUID `json:"election_id,omitempty"`
	Title       string     `json:"title" validate:"required,max=300"`
	Slug        string     `json:"slug" validate:"required,max=300"`
	Content     string     `json:"content" validate:"required"`
	ContentType string     `json:"content_type" validate:"required,oneof=article faq guide video"`
	Category    *string    `json:"category,omitempty" validate:"omitempty,max=100"`
	IsFeatured  bool       `json:"is_featured"`
	IsPublished bool       `json:"is_published"`
}

type UpdateVoterEducationRequest struct {
	ElectionID  *uuid.UUID `json:"election_id,omitempty"`
	Title       *string    `json:"title,omitempty" validate:"omitempty,max=300"`
	Slug        *string    `json:"slug,omitempty" validate:"omitempty,max=300"`
	Content     *string    `json:"content,omitempty"`
	ContentType *string    `json:"content_type,omitempty" validate:"omitempty,oneof=article faq guide video"`
	Category    *string    `json:"category,omitempty" validate:"omitempty,max=100"`
	IsFeatured  *bool      `json:"is_featured,omitempty"`
	IsPublished *bool      `json:"is_published,omitempty"`
}

// Filter types

type ElectionFilter struct {
	ElectionType   *string
	Status         *string
	Year           *int
	IsFeatured     *bool
	Search         *string
	IncludeDeleted bool
}

type CandidateFilter struct {
	ElectionID   *uuid.UUID
	PositionID   *uuid.UUID
	PoliticianID *uuid.UUID
	PartyID      *uuid.UUID
	Status       *string
	IsWinner     *bool
}

// Paginated types

type PaginatedElections struct {
	Elections  []ElectionListItem `json:"elections"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
}

type PaginatedCandidates struct {
	Candidates []CandidateListItem `json:"candidates"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
	TotalPages int                 `json:"total_pages"`
}

type PaginatedVoterEducation struct {
	Items      []VoterEducationListItem `json:"items"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	PerPage    int                      `json:"per_page"`
	TotalPages int                      `json:"total_pages"`
}

// Calendar view type
type ElectionCalendarItem struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	ElectionType string    `json:"election_type"`
	ElectionDate time.Time `json:"election_date"`
	Status       string    `json:"status"`
}
