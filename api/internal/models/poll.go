package models

import (
	"time"

	"github.com/google/uuid"
)

// Poll Status constants
const (
	PollStatusDraft           = "draft"
	PollStatusPendingApproval = "pending_approval"
	PollStatusActive          = "active"
	PollStatusClosed          = "closed"
	PollStatusRejected        = "rejected"
)

// Poll Category constants
const (
	PollCategoryGeneral       = "general"
	PollCategoryElection      = "election"
	PollCategoryLegislation   = "legislation"
	PollCategoryPolitician    = "politician"
	PollCategoryPolicy        = "policy"
	PollCategoryLocalIssue    = "local_issue"
	PollCategoryNationalIssue = "national_issue"
)

// Poll represents a user or admin created poll
type Poll struct {
	ID                    uuid.UUID  `json:"id"`
	UserID                uuid.UUID  `json:"user_id"`
	Title                 string     `json:"title"`
	Slug                  string     `json:"slug"`
	Description           *string    `json:"description,omitempty"`
	Category              string     `json:"category"`
	Status                string     `json:"status"`
	PoliticianID          *uuid.UUID `json:"politician_id,omitempty"`
	ElectionID            *uuid.UUID `json:"election_id,omitempty"`
	BillID                *uuid.UUID `json:"bill_id,omitempty"`
	// Location scoping (optional - if all nil, poll is national)
	RegionID           *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID         *uuid.UUID `json:"province_id,omitempty"`
	CityMunicipalityID *uuid.UUID `json:"city_municipality_id,omitempty"`
	BarangayID         *uuid.UUID `json:"barangay_id,omitempty"`
	IsAnonymous           bool       `json:"is_anonymous"`
	AllowMultipleVotes    bool       `json:"allow_multiple_votes"`
	ShowResultsBeforeVote bool       `json:"show_results_before_vote"`
	IsFeatured            bool       `json:"is_featured"`
	StartsAt              *time.Time `json:"starts_at,omitempty"`
	EndsAt                *time.Time `json:"ends_at,omitempty"`
	ApprovedBy            *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt            *time.Time `json:"approved_at,omitempty"`
	RejectionReason       *string    `json:"rejection_reason,omitempty"`
	TotalVotes            int        `json:"total_votes"`
	ViewCount             int        `json:"view_count"`
	CommentCount          int        `json:"comment_count"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	DeletedAt             *time.Time `json:"deleted_at,omitempty"`

	// Joined fields
	Author     *PollAuthor       `json:"author,omitempty"`
	Options    []PollOption      `json:"options,omitempty"`
	Politician *PoliticianBrief  `json:"politician,omitempty"`
	Election   *ElectionBrief    `json:"election,omitempty"`
	Bill       *BillBrief        `json:"bill,omitempty"`
	Location   *LocationBrief    `json:"location,omitempty"` // Human-readable location
	UserVote   *uuid.UUID        `json:"user_vote,omitempty"` // Option ID user voted for
}

type PollListItem struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	Category     string     `json:"category"`
	Status       string     `json:"status"`
	IsFeatured   bool       `json:"is_featured"`
	TotalVotes   int        `json:"total_votes"`
	CommentCount int        `json:"comment_count"`
	EndsAt       *time.Time `json:"ends_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	Author       *PollAuthor `json:"author,omitempty"`
	OptionCount  int        `json:"option_count"`
	Location     *string    `json:"location,omitempty"` // Human-readable location display name
}

type PollAuthor struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Avatar *string   `json:"avatar,omitempty"`
}

type PoliticianBrief struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Slug  string    `json:"slug"`
	Photo *string   `json:"photo,omitempty"`
}

type ElectionBrief struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

type BillBrief struct {
	ID         uuid.UUID `json:"id"`
	BillNumber string    `json:"bill_number"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
}

// LocationBrief represents a human-readable location for polls
type LocationBrief struct {
	RegionID           *uuid.UUID `json:"region_id,omitempty"`
	RegionName         *string    `json:"region_name,omitempty"`
	ProvinceID         *uuid.UUID `json:"province_id,omitempty"`
	ProvinceName       *string    `json:"province_name,omitempty"`
	CityMunicipalityID *uuid.UUID `json:"city_municipality_id,omitempty"`
	CityMunicipalityName *string  `json:"city_municipality_name,omitempty"`
	BarangayID         *uuid.UUID `json:"barangay_id,omitempty"`
	BarangayName       *string    `json:"barangay_name,omitempty"`
	DisplayName        string     `json:"display_name"` // e.g., "Quezon City, Metro Manila, NCR"
}

// PollOption represents a choice in a poll
type PollOption struct {
	ID           uuid.UUID `json:"id"`
	PollID       uuid.UUID `json:"poll_id"`
	Text         string    `json:"text"`
	DisplayOrder int       `json:"display_order"`
	VoteCount    int       `json:"vote_count"`
	Percentage   float64   `json:"percentage,omitempty"` // Calculated field
	CreatedAt    time.Time `json:"created_at"`
}

// PollVote represents a vote on a poll
type PollVote struct {
	ID        uuid.UUID  `json:"id"`
	PollID    uuid.UUID  `json:"poll_id"`
	OptionID  uuid.UUID  `json:"option_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	IPHash    *string    `json:"-"` // Never expose
	CreatedAt time.Time  `json:"created_at"`
}

// PollComment represents a comment on a poll
type PollComment struct {
	ID               uuid.UUID        `json:"id"`
	PollID           uuid.UUID        `json:"poll_id"`
	UserID           uuid.UUID        `json:"user_id"`
	ParentID         *uuid.UUID       `json:"parent_id,omitempty"`
	Content          string           `json:"content"`
	Status           string           `json:"status"`
	ModeratedBy      *uuid.UUID       `json:"moderated_by,omitempty"`
	ModeratedAt      *time.Time       `json:"moderated_at,omitempty"`
	ModerationReason *string          `json:"moderation_reason,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        *time.Time       `json:"deleted_at,omitempty"`

	// Joined fields
	Author     *CommentAuthor     `json:"author,omitempty"`
	Reactions  []ReactionSummary  `json:"reactions,omitempty"`
	ReplyCount int                `json:"reply_count,omitempty"`
}

// Request types

type CreatePollRequest struct {
	Title                 string     `json:"title" validate:"required,max=300"`
	Slug                  string     `json:"slug" validate:"required,max=300"`
	Description           *string    `json:"description,omitempty"`
	Category              string     `json:"category" validate:"required,oneof=general election legislation politician policy local_issue national_issue"`
	PoliticianID          *uuid.UUID `json:"politician_id,omitempty"`
	ElectionID            *uuid.UUID `json:"election_id,omitempty"`
	BillID                *uuid.UUID `json:"bill_id,omitempty"`
	// Location scoping (optional - if all nil, poll is national)
	RegionID           *uuid.UUID `json:"region_id,omitempty"`
	ProvinceID         *uuid.UUID `json:"province_id,omitempty"`
	CityMunicipalityID *uuid.UUID `json:"city_municipality_id,omitempty"`
	BarangayID         *uuid.UUID `json:"barangay_id,omitempty"`
	IsAnonymous           bool       `json:"is_anonymous"`
	AllowMultipleVotes    bool       `json:"allow_multiple_votes"`
	ShowResultsBeforeVote bool       `json:"show_results_before_vote"`
	StartsAt              *string    `json:"starts_at,omitempty"` // ISO 8601
	EndsAt                *string    `json:"ends_at,omitempty"`   // ISO 8601
	Options               []string   `json:"options" validate:"required,min=2,max=10"`
}

type UpdatePollRequest struct {
	Title                 *string    `json:"title,omitempty" validate:"omitempty,max=300"`
	Slug                  *string    `json:"slug,omitempty" validate:"omitempty,max=300"`
	Description           *string    `json:"description,omitempty"`
	Category              *string    `json:"category,omitempty" validate:"omitempty,oneof=general election legislation politician policy local_issue national_issue"`
	IsAnonymous           *bool      `json:"is_anonymous,omitempty"`
	AllowMultipleVotes    *bool      `json:"allow_multiple_votes,omitempty"`
	ShowResultsBeforeVote *bool      `json:"show_results_before_vote,omitempty"`
	StartsAt              *string    `json:"starts_at,omitempty"`
	EndsAt                *string    `json:"ends_at,omitempty"`
}

type AdminUpdatePollRequest struct {
	UpdatePollRequest
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=draft pending_approval active closed rejected"`
	IsFeatured  *bool   `json:"is_featured,omitempty"`
}

type ApprovePollRequest struct {
	Approved bool    `json:"approved"`
	Reason   *string `json:"reason,omitempty"` // Required if not approved
}

type CastVoteRequest struct {
	OptionID uuid.UUID `json:"option_id" validate:"required"`
}

type CreatePollCommentRequest struct {
	Content  string     `json:"content" validate:"required,max=2000"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// Filter types

type PollFilter struct {
	Category           *string
	Status             *string
	UserID             *uuid.UUID
	PoliticianID       *uuid.UUID
	ElectionID         *uuid.UUID
	IsFeatured         *bool
	Search             *string
	ActiveOnly         bool
	// Location filters
	RegionID           *uuid.UUID
	ProvinceID         *uuid.UUID
	CityMunicipalityID *uuid.UUID
	BarangayID         *uuid.UUID
	// If true, include national polls (no location) along with location-filtered results
	IncludeNational    bool
}

// Paginated types

type PaginatedPolls struct {
	Polls      []PollListItem `json:"polls"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

type PaginatedPollComments struct {
	Comments   []PollComment `json:"comments"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}

// Response types

type PollResults struct {
	PollID     uuid.UUID    `json:"poll_id"`
	TotalVotes int          `json:"total_votes"`
	Options    []PollOption `json:"options"`
}

type VoteResponse struct {
	Success  bool         `json:"success"`
	Message  string       `json:"message"`
	Results  *PollResults `json:"results,omitempty"`
}
