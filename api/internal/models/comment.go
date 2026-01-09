package models

import (
	"time"

	"github.com/google/uuid"
)

// CommentStatus represents the moderation status of a comment
type CommentStatus string

const (
	CommentStatusActive      CommentStatus = "active"
	CommentStatusUnderReview CommentStatus = "under_review"
	CommentStatusSpam        CommentStatus = "spam"
	CommentStatusHidden      CommentStatus = "hidden"
)

// Comment represents a comment on an article
type Comment struct {
	ID        uuid.UUID     `json:"id"`
	ArticleID uuid.UUID     `json:"article_id"`
	UserID    uuid.UUID     `json:"user_id"`
	ParentID  *uuid.UUID    `json:"parent_id,omitempty"` // NULL for root comments, set for replies
	Content   string        `json:"content"`             // Markdown content
	Status    CommentStatus `json:"status"`              // Moderation status: active, under_review, spam, hidden
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	DeletedAt *time.Time    `json:"deleted_at,omitempty"`

	// Moderation fields
	ModeratedBy      *uuid.UUID `json:"moderated_by,omitempty"`
	ModeratedAt      *time.Time `json:"moderated_at,omitempty"`
	ModerationReason *string    `json:"moderation_reason,omitempty"`

	// Relations (populated when needed)
	Author    *CommentAuthor    `json:"author,omitempty"` // User info displayed as "author" in JSON for frontend compatibility
	Replies   []Comment         `json:"replies,omitempty"`
	Reactions []ReactionSummary `json:"reactions,omitempty"`
	Mentions  []CommentMention  `json:"mentions,omitempty"`

	// Computed fields
	ReplyCount  int     `json:"reply_count,omitempty"`
	ArticleSlug *string `json:"article_slug,omitempty"` // For user profile comments
}

// CommentAuthor is a minimal user representation for comments (called "author" for frontend compatibility)
type CommentAuthor struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Avatar   *string   `json:"avatar,omitempty"`
	IsSystem bool      `json:"is_system,omitempty"` // True for verified/staff users
}

// CommentReaction represents a user's reaction to a comment
type CommentReaction struct {
	ID        uuid.UUID `json:"id"`
	CommentID uuid.UUID `json:"comment_id"`
	UserID    uuid.UUID `json:"user_id"`
	Reaction  string    `json:"reaction"` // 'heart', 'thumbsup', 'thumbsdown', 'laugh', 'fire', 'eyes'
	CreatedAt time.Time `json:"created_at"`
}

// ReactionSummary shows reaction counts and if current user reacted
type ReactionSummary struct {
	Reaction   string `json:"reaction"`
	Count      int    `json:"count"`
	HasReacted bool   `json:"has_reacted"` // Whether current user has this reaction
}

// CommentMention represents a @mention in a comment
type CommentMention struct {
	ID                uuid.UUID `json:"id"`
	CommentID         uuid.UUID `json:"comment_id"`
	MentionedAuthorID uuid.UUID `json:"mentioned_author_id"`
	CreatedAt         time.Time `json:"created_at"`

	// Populated when needed
	MentionedAuthor *CommentAuthor `json:"mentioned_author,omitempty"`
}

// CreateCommentRequest is the request body for creating a comment
type CreateCommentRequest struct {
	Content  string  `json:"content" validate:"required,min=1,max=10000"`
	ParentID *string `json:"parent_id,omitempty" validate:"omitempty,uuid"` // For replies
}

// UpdateCommentRequest is the request body for updating a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=10000"`
}

// AddReactionRequest is the request body for adding a reaction
type AddReactionRequest struct {
	Reaction string `json:"reaction" validate:"required,oneof=heart thumbsup thumbsdown laugh fire eyes"`
}

// ModerateCommentRequest is the request body for moderating a comment
type ModerateCommentRequest struct {
	Status CommentStatus `json:"status" validate:"required,oneof=active under_review spam hidden"`
	Reason *string       `json:"reason,omitempty"`
}

// CommentFilter for filtering comments
type CommentFilter struct {
	ArticleID      *uuid.UUID
	UserID         *uuid.UUID
	ParentID       *uuid.UUID // NULL to get only root comments
	Status         *CommentStatus
	IncludeDeleted bool
	IncludeHidden  bool // Admin-only: include hidden/spam comments
}

// PaginatedComments for paginated comment responses
type PaginatedComments struct {
	Comments   []Comment `json:"comments"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
}

// ReplyPreview shows a preview of replies for collapsed view
type ReplyPreview struct {
	Count   int             `json:"count"`
	Authors []CommentAuthor `json:"authors"` // First few authors who replied
}

// Supported reactions
var SupportedReactions = []string{"heart", "thumbsup", "thumbsdown", "laugh", "fire", "eyes"}

// IsValidReaction checks if a reaction is supported
func IsValidReaction(reaction string) bool {
	for _, r := range SupportedReactions {
		if r == reaction {
			return true
		}
	}
	return false
}
