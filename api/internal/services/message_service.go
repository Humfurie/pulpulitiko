package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type MessageService struct {
	repo *repository.MessageRepository
}

func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

// CreateConversation creates a new conversation with an initial message
func (s *MessageService) CreateConversation(ctx context.Context, userID uuid.UUID, req *models.CreateConversationRequest) (*models.Conversation, *models.Message, error) {
	// Always create a new conversation (allows multiple conversations per user)
	var subject *string
	if req.Subject != "" {
		subject = &req.Subject
	}

	conversation, err := s.repo.CreateConversation(ctx, userID, subject)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Create the initial message
	message, err := s.repo.CreateMessage(ctx, conversation.ID, userID, req.Message)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Get full conversation with user info
	conversation, err = s.repo.GetConversationByID(ctx, conversation.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Get full message with sender info
	message, err = s.repo.GetMessageByID(ctx, message.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get message: %w", err)
	}

	return conversation, message, nil
}

// GetConversation retrieves a conversation by ID
func (s *MessageService) GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	return s.repo.GetConversationByID(ctx, id)
}

// GetUserConversation gets or returns nil if no conversation exists for user
func (s *MessageService) GetUserConversation(ctx context.Context, userID uuid.UUID) (*models.Conversation, error) {
	return s.repo.GetConversationByUserID(ctx, userID)
}

// ListConversations lists all conversations (admin only)
func (s *MessageService) ListConversations(ctx context.Context, filter *models.ConversationFilter, page, perPage int) (*models.PaginatedConversations, error) {
	return s.repo.ListConversations(ctx, filter, page, perPage)
}

// GetUserConversations gets all conversations for a specific user
func (s *MessageService) GetUserConversations(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error) {
	return s.repo.GetUserConversations(ctx, userID)
}

// UpdateConversationStatus updates the status of a conversation
func (s *MessageService) UpdateConversationStatus(ctx context.Context, id uuid.UUID, status models.ConversationStatus) error {
	return s.repo.UpdateConversationStatus(ctx, id, status)
}

// SendMessage sends a message in a conversation
func (s *MessageService) SendMessage(ctx context.Context, conversationID, senderID uuid.UUID, req *models.CreateMessageRequest) (*models.Message, error) {
	// Verify conversation exists
	conversation, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conversation == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Create the message
	message, err := s.repo.CreateMessage(ctx, conversationID, senderID, req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Get full message with sender info
	message, err = s.repo.GetMessageByID(ctx, message.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return message, nil
}

// GetMessages retrieves messages in a conversation with pagination
func (s *MessageService) GetMessages(ctx context.Context, conversationID uuid.UUID, page, perPage int) (*models.PaginatedMessages, error) {
	return s.repo.ListMessages(ctx, conversationID, page, perPage)
}

// MarkAsRead marks all messages in a conversation as read
func (s *MessageService) MarkAsRead(ctx context.Context, conversationID, readerID uuid.UUID) error {
	return s.repo.MarkMessagesAsRead(ctx, conversationID, readerID)
}

// GetUnreadCounts gets unread message counts for a user
func (s *MessageService) GetUnreadCounts(ctx context.Context, userID uuid.UUID, isAdmin bool) (*models.UnreadCounts, error) {
	return s.repo.GetUnreadCounts(ctx, userID, isAdmin)
}

// CanAccessConversation checks if a user can access a conversation
func (s *MessageService) CanAccessConversation(ctx context.Context, conversationID, userID uuid.UUID, isAdmin bool) (bool, error) {
	conversation, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return false, err
	}
	if conversation == nil {
		return false, nil
	}

	// Admin can access all conversations
	if isAdmin {
		return true, nil
	}

	// Users can only access their own conversations
	return conversation.UserID == userID, nil
}
