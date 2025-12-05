package models

import (
	"time"

	"github.com/google/uuid"
)

// PoliticianComment represents a comment on a politician page
type PoliticianComment struct {
	ID           uuid.UUID     `json:"id"`
	PoliticianID uuid.UUID     `json:"politician_id"`
	UserID       uuid.UUID     `json:"user_id"`
	ParentID     *uuid.UUID    `json:"parent_id,omitempty"`
	Content      string        `json:"content"`
	Status       CommentStatus `json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	DeletedAt    *time.Time    `json:"deleted_at,omitempty"`

	// Moderation fields
	ModeratedBy      *uuid.UUID `json:"moderated_by,omitempty"`
	ModeratedAt      *time.Time `json:"moderated_at,omitempty"`
	ModerationReason *string    `json:"moderation_reason,omitempty"`

	// Relations
	Author    *CommentAuthor    `json:"author,omitempty"`
	Replies   []PoliticianComment `json:"replies,omitempty"`
	Reactions []ReactionSummary `json:"reactions,omitempty"`

	// Computed fields
	ReplyCount     int     `json:"reply_count,omitempty"`
	PoliticianSlug *string `json:"politician_slug,omitempty"`
	PoliticianName *string `json:"politician_name,omitempty"`
}

// PaginatedPoliticianComments for paginated responses
type PaginatedPoliticianComments struct {
	Comments   []PoliticianComment `json:"comments"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
	TotalPages int                 `json:"total_pages"`
}

// PoliticianCommentFilter for filtering politician comments
type PoliticianCommentFilter struct {
	PoliticianID   *uuid.UUID
	UserID         *uuid.UUID
	ParentID       *uuid.UUID
	Status         *CommentStatus
	IncludeDeleted bool
	IncludeHidden  bool
}
