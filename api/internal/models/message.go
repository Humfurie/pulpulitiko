package models

import (
	"time"

	"github.com/google/uuid"
)

// ConversationStatus represents the status of a conversation
type ConversationStatus string

const (
	ConversationStatusOpen     ConversationStatus = "open"
	ConversationStatusClosed   ConversationStatus = "closed"
	ConversationStatusArchived ConversationStatus = "archived"
)

// Conversation represents a chat conversation between a user and admin
type Conversation struct {
	ID            uuid.UUID          `json:"id"`
	UserID        uuid.UUID          `json:"user_id"`
	User          *User              `json:"user,omitempty"`
	Subject       *string            `json:"subject,omitempty"`
	Status        ConversationStatus `json:"status"`
	LastMessageAt *time.Time         `json:"last_message_at,omitempty"`
	LastMessage   *Message           `json:"last_message,omitempty"`
	UnreadCount   int                `json:"unread_count,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	Sender         *User      `json:"sender,omitempty"`
	Content        string     `json:"content"`
	IsRead         bool       `json:"is_read"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CreateConversationRequest represents the request to create a new conversation
type CreateConversationRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message" validate:"required,min=1"`
}

// CreateMessageRequest represents the request to send a new message
type CreateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1"`
}

// UpdateConversationRequest represents the request to update a conversation
type UpdateConversationRequest struct {
	Status ConversationStatus `json:"status" validate:"required,oneof=open closed archived"`
}

// ConversationFilter represents filters for listing conversations
type ConversationFilter struct {
	UserID *uuid.UUID
	Status *ConversationStatus
}

// PaginatedConversations represents a paginated list of conversations
type PaginatedConversations struct {
	Conversations []Conversation `json:"conversations"`
	Total         int            `json:"total"`
	Page          int            `json:"page"`
	PerPage       int            `json:"per_page"`
	TotalPages    int            `json:"total_pages"`
}

// PaginatedMessages represents a paginated list of messages
type PaginatedMessages struct {
	Messages   []Message `json:"messages"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
}

// WebSocket message types
type WSMessageType string

const (
	WSMessageTypeNewMessage   WSMessageType = "new_message"
	WSMessageTypeMessageRead  WSMessageType = "message_read"
	WSMessageTypeTyping       WSMessageType = "typing"
	WSMessageTypeStopTyping   WSMessageType = "stop_typing"
	WSMessageTypeUserOnline   WSMessageType = "user_online"
	WSMessageTypeUserOffline  WSMessageType = "user_offline"
	WSMessageTypeConversation WSMessageType = "conversation_update"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type           WSMessageType `json:"type"`
	ConversationID *uuid.UUID    `json:"conversation_id,omitempty"`
	Message        *Message      `json:"message,omitempty"`
	UserID         *uuid.UUID    `json:"user_id,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
}

// UnreadCounts represents unread message counts for a user
type UnreadCounts struct {
	Total         int `json:"total"`
	Conversations int `json:"conversations"`
}
