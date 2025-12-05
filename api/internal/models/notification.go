package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeMentionArticleComment    NotificationType = "mention_article_comment"
	NotificationTypeMentionPoliticianComment NotificationType = "mention_politician_comment"
	NotificationTypeReplyArticleComment      NotificationType = "reply_article_comment"
	NotificationTypeReplyPoliticianComment   NotificationType = "reply_politician_comment"
	NotificationTypeCommentReaction          NotificationType = "comment_reaction"
)

// Notification represents a user notification
type Notification struct {
	ID           uuid.UUID        `json:"id"`
	UserID       uuid.UUID        `json:"user_id"`
	Type         NotificationType `json:"type"`
	Title        string           `json:"title"`
	Message      *string          `json:"message,omitempty"`
	ActorID      *uuid.UUID       `json:"actor_id,omitempty"`
	ArticleID    *uuid.UUID       `json:"article_id,omitempty"`
	PoliticianID *uuid.UUID       `json:"politician_id,omitempty"`
	CommentID    *uuid.UUID       `json:"comment_id,omitempty"`
	IsRead       bool             `json:"is_read"`
	ReadAt       *time.Time       `json:"read_at,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`

	// Relations (populated when needed)
	Actor      *NotificationActor `json:"actor,omitempty"`
	ArticleRef *NotificationRef   `json:"article,omitempty"`
	PoliticianRef *NotificationRef `json:"politician,omitempty"`
}

// NotificationActor is minimal user info for notifications
type NotificationActor struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Avatar *string   `json:"avatar,omitempty"`
}

// NotificationRef is a minimal reference for articles/politicians
type NotificationRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// PaginatedNotifications for paginated responses
type PaginatedNotifications struct {
	Notifications []Notification `json:"notifications"`
	Total         int            `json:"total"`
	UnreadCount   int            `json:"unread_count"`
	Page          int            `json:"page"`
	PerPage       int            `json:"per_page"`
	TotalPages    int            `json:"total_pages"`
}

// CreateNotificationRequest for creating notifications
type CreateNotificationRequest struct {
	UserID       uuid.UUID
	Type         NotificationType
	Title        string
	Message      *string
	ActorID      *uuid.UUID
	ArticleID    *uuid.UUID
	PoliticianID *uuid.UUID
	CommentID    *uuid.UUID
}
