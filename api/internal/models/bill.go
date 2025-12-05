package models

import (
	"time"

	"github.com/google/uuid"
)

// Bill Status constants
const (
	BillStatusFiled                 = "filed"
	BillStatusPendingCommittee      = "pending_committee"
	BillStatusInCommittee           = "in_committee"
	BillStatusReportedOut           = "reported_out"
	BillStatusPendingSecondReading  = "pending_second_reading"
	BillStatusApprovedSecondReading = "approved_second_reading"
	BillStatusPendingThirdReading   = "pending_third_reading"
	BillStatusApprovedThirdReading  = "approved_third_reading"
	BillStatusTransmitted           = "transmitted"
	BillStatusConsolidated          = "consolidated"
	BillStatusRatified              = "ratified"
	BillStatusSignedIntoLaw         = "signed_into_law"
	BillStatusVetoed                = "vetoed"
	BillStatusLapsed                = "lapsed"
	BillStatusWithdrawn             = "withdrawn"
	BillStatusArchived              = "archived"
)

// Chamber constants
const (
	ChamberSenate = "senate"
	ChamberHouse  = "house"
)

// Vote type constants
const (
	VoteYea     = "yea"
	VoteNay     = "nay"
	VoteAbstain = "abstain"
	VoteAbsent  = "absent"
)

// LegislativeSession represents a Congress session
type LegislativeSession struct {
	ID             uuid.UUID  `json:"id"`
	CongressNumber int        `json:"congress_number"`
	SessionNumber  int        `json:"session_number"`
	SessionType    string     `json:"session_type"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	IsCurrent      bool       `json:"is_current"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type LegislativeSessionListItem struct {
	ID             uuid.UUID `json:"id"`
	CongressNumber int       `json:"congress_number"`
	SessionNumber  int       `json:"session_number"`
	SessionType    string    `json:"session_type"`
	IsCurrent      bool      `json:"is_current"`
	BillCount      int       `json:"bill_count"`
}

// Committee represents a legislative committee
type Committee struct {
	ID                uuid.UUID          `json:"id"`
	Chamber           string             `json:"chamber"`
	Name              string             `json:"name"`
	Slug              string             `json:"slug"`
	Description       *string            `json:"description,omitempty"`
	ChairpersonID     *uuid.UUID         `json:"chairperson_id,omitempty"`
	ViceChairpersonID *uuid.UUID         `json:"vice_chairperson_id,omitempty"`
	IsActive          bool               `json:"is_active"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	DeletedAt         *time.Time         `json:"deleted_at,omitempty"`
	Chairperson       *PoliticianListItem `json:"chairperson,omitempty"`
	ViceChairperson   *PoliticianListItem `json:"vice_chairperson,omitempty"`
}

type CommitteeListItem struct {
	ID          uuid.UUID `json:"id"`
	Chamber     string    `json:"chamber"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	IsActive    bool      `json:"is_active"`
	BillCount   int       `json:"bill_count"`
}

// Bill represents a legislative bill
type Bill struct {
	ID               uuid.UUID    `json:"id"`
	SessionID        uuid.UUID    `json:"session_id"`
	Chamber          string       `json:"chamber"`
	BillNumber       string       `json:"bill_number"`
	Title            string       `json:"title"`
	Slug             string       `json:"slug"`
	ShortTitle       *string      `json:"short_title,omitempty"`
	Summary          *string      `json:"summary,omitempty"`
	FullText         *string      `json:"full_text,omitempty"`
	Significance     *string      `json:"significance,omitempty"`
	Status           string       `json:"status"`
	FiledDate        time.Time    `json:"filed_date"`
	LastActionDate   *time.Time   `json:"last_action_date,omitempty"`
	DateSigned       *time.Time   `json:"date_signed,omitempty"`
	RepublicActNumber *string     `json:"republic_act_number,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	DeletedAt        *time.Time   `json:"deleted_at,omitempty"`

	// Joined fields
	Session           *LegislativeSessionListItem `json:"session,omitempty"`
	Authors           []BillAuthor                `json:"authors,omitempty"`
	PrincipalAuthors  []PoliticianListItem        `json:"principal_authors,omitempty"`
	Committees        []BillCommittee             `json:"committees,omitempty"`
	StatusHistory     []BillStatusHistoryItem     `json:"status_history,omitempty"`
	Topics            []BillTopic                 `json:"topics,omitempty"`
	Votes             []BillVote                  `json:"votes,omitempty"`
}

type BillListItem struct {
	ID             uuid.UUID    `json:"id"`
	Chamber        string       `json:"chamber"`
	BillNumber     string       `json:"bill_number"`
	Title          string       `json:"title"`
	Slug           string       `json:"slug"`
	ShortTitle     *string      `json:"short_title,omitempty"`
	Status         string       `json:"status"`
	FiledDate      time.Time    `json:"filed_date"`
	LastActionDate *time.Time   `json:"last_action_date,omitempty"`
	AuthorCount    int          `json:"author_count"`
	TopicNames     []string     `json:"topic_names,omitempty"`
}

// BillAuthor represents an author of a bill
type BillAuthor struct {
	ID                uuid.UUID          `json:"id"`
	BillID            uuid.UUID          `json:"bill_id"`
	PoliticianID      uuid.UUID          `json:"politician_id"`
	IsPrincipalAuthor bool               `json:"is_principal_author"`
	CreatedAt         time.Time          `json:"created_at"`
	Politician        *PoliticianListItem `json:"politician,omitempty"`
}

// BillStatusHistoryItem represents a status change in the bill timeline
type BillStatusHistoryItem struct {
	ID                uuid.UUID `json:"id"`
	BillID            uuid.UUID `json:"bill_id"`
	Status            string    `json:"status"`
	ActionDescription *string   `json:"action_description,omitempty"`
	ActionDate        time.Time `json:"action_date"`
	CreatedAt         time.Time `json:"created_at"`
}

// BillCommittee represents a committee referral
type BillCommittee struct {
	ID           uuid.UUID          `json:"id"`
	BillID       uuid.UUID          `json:"bill_id"`
	CommitteeID  uuid.UUID          `json:"committee_id"`
	ReferredDate time.Time          `json:"referred_date"`
	IsPrimary    bool               `json:"is_primary"`
	Status       string             `json:"status"`
	CreatedAt    time.Time          `json:"created_at"`
	Committee    *CommitteeListItem `json:"committee,omitempty"`
}

// BillVote represents a voting session for a bill
type BillVote struct {
	ID          uuid.UUID `json:"id"`
	BillID      uuid.UUID `json:"bill_id"`
	Chamber     string    `json:"chamber"`
	Reading     string    `json:"reading"`
	VoteDate    time.Time `json:"vote_date"`
	Yeas        int       `json:"yeas"`
	Nays        int       `json:"nays"`
	Abstentions int       `json:"abstentions"`
	Absent      int       `json:"absent"`
	IsPassed    bool      `json:"is_passed"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// PoliticianVote represents an individual politician's vote
type PoliticianVote struct {
	ID           uuid.UUID          `json:"id"`
	BillVoteID   uuid.UUID          `json:"bill_vote_id"`
	PoliticianID uuid.UUID          `json:"politician_id"`
	Vote         string             `json:"vote"`
	CreatedAt    time.Time          `json:"created_at"`
	Politician   *PoliticianListItem `json:"politician,omitempty"`
}

// BillTopic represents a topic/category for bills
type BillTopic struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	BillCount   int       `json:"bill_count,omitempty"`
}

// Request types

type CreateBillRequest struct {
	SessionID        uuid.UUID   `json:"session_id" validate:"required"`
	Chamber          string      `json:"chamber" validate:"required,oneof=senate house"`
	BillNumber       string      `json:"bill_number" validate:"required,max=50"`
	Title            string      `json:"title" validate:"required,max=500"`
	Slug             string      `json:"slug" validate:"required,max=500"`
	ShortTitle       *string     `json:"short_title,omitempty" validate:"omitempty,max=200"`
	Summary          *string     `json:"summary,omitempty"`
	FullText         *string     `json:"full_text,omitempty"`
	Significance     *string     `json:"significance,omitempty" validate:"omitempty,max=100"`
	Status           string      `json:"status" validate:"required"`
	FiledDate        string      `json:"filed_date" validate:"required"` // YYYY-MM-DD
	PrincipalAuthors []uuid.UUID `json:"principal_authors,omitempty"`
	CoAuthors        []uuid.UUID `json:"co_authors,omitempty"`
	TopicIDs         []uuid.UUID `json:"topic_ids,omitempty"`
}

type UpdateBillRequest struct {
	Title             *string     `json:"title,omitempty" validate:"omitempty,max=500"`
	Slug              *string     `json:"slug,omitempty" validate:"omitempty,max=500"`
	ShortTitle        *string     `json:"short_title,omitempty" validate:"omitempty,max=200"`
	Summary           *string     `json:"summary,omitempty"`
	FullText          *string     `json:"full_text,omitempty"`
	Significance      *string     `json:"significance,omitempty" validate:"omitempty,max=100"`
	Status            *string     `json:"status,omitempty"`
	LastActionDate    *string     `json:"last_action_date,omitempty"` // YYYY-MM-DD
	DateSigned        *string     `json:"date_signed,omitempty"`      // YYYY-MM-DD
	RepublicActNumber *string     `json:"republic_act_number,omitempty" validate:"omitempty,max=50"`
	TopicIDs          []uuid.UUID `json:"topic_ids,omitempty"`
}

type AddBillStatusRequest struct {
	Status            string `json:"status" validate:"required"`
	ActionDescription string `json:"action_description,omitempty"`
	ActionDate        string `json:"action_date" validate:"required"` // YYYY-MM-DD
}

type AddBillVoteRequest struct {
	Chamber     string `json:"chamber" validate:"required,oneof=senate house"`
	Reading     string `json:"reading" validate:"required,oneof=second third"`
	VoteDate    string `json:"vote_date" validate:"required"` // YYYY-MM-DD
	Yeas        int    `json:"yeas" validate:"min=0"`
	Nays        int    `json:"nays" validate:"min=0"`
	Abstentions int    `json:"abstentions" validate:"min=0"`
	Absent      int    `json:"absent" validate:"min=0"`
	IsPassed    bool   `json:"is_passed"`
	Notes       string `json:"notes,omitempty"`
}

type AddPoliticianVoteRequest struct {
	PoliticianID uuid.UUID `json:"politician_id" validate:"required"`
	Vote         string    `json:"vote" validate:"required,oneof=yea nay abstain absent"`
}

type BillFilter struct {
	Chamber        *string
	Status         *string
	SessionID      *uuid.UUID
	TopicID        *uuid.UUID
	AuthorID       *uuid.UUID
	Search         *string
	FiledAfter     *time.Time
	FiledBefore    *time.Time
	IncludeDeleted bool
}

type PaginatedBills struct {
	Bills      []BillListItem `json:"bills"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// Politician voting record summary
type PoliticianVotingRecord struct {
	PoliticianID  uuid.UUID `json:"politician_id"`
	TotalVotes    int       `json:"total_votes"`
	YeaVotes      int       `json:"yea_votes"`
	NayVotes      int       `json:"nay_votes"`
	AbstainVotes  int       `json:"abstain_votes"`
	AbsentVotes   int       `json:"absent_votes"`
	AttendanceRate float64  `json:"attendance_rate"`
}

type PoliticianBillVote struct {
	Bill       BillListItem `json:"bill"`
	Vote       string       `json:"vote"`
	VoteDate   time.Time    `json:"vote_date"`
	Reading    string       `json:"reading"`
	BillPassed bool         `json:"bill_passed"`
}

type PaginatedPoliticianVotes struct {
	Votes      []PoliticianBillVote `json:"votes"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	TotalPages int                  `json:"total_pages"`
}
