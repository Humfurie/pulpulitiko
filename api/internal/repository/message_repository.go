package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

// CreateConversation creates a new conversation
func (r *MessageRepository) CreateConversation(ctx context.Context, userID uuid.UUID, subject *string) (*models.Conversation, error) {
	conversation := &models.Conversation{}
	query := `
		INSERT INTO conversations (user_id, subject, status)
		VALUES ($1, $2, 'open')
		RETURNING id, user_id, subject, status, last_message_at, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, userID, subject).Scan(
		&conversation.ID, &conversation.UserID, &conversation.Subject,
		&conversation.Status, &conversation.LastMessageAt,
		&conversation.CreatedAt, &conversation.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conversation, nil
}

// GetConversationByID retrieves a conversation by ID with user info
func (r *MessageRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	query := `
		SELECT c.id, c.user_id, c.subject, c.status, c.last_message_at, c.created_at, c.updated_at,
		       u.id, u.name, u.email, u.avatar
		FROM conversations c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`

	conversation := &models.Conversation{}
	user := &models.User{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&conversation.ID, &conversation.UserID, &conversation.Subject,
		&conversation.Status, &conversation.LastMessageAt,
		&conversation.CreatedAt, &conversation.UpdatedAt,
		&user.ID, &user.Name, &user.Email, &user.Avatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	conversation.User = user
	return conversation, nil
}

// GetConversationByUserID gets the open conversation for a user (creates one if not exists)
func (r *MessageRepository) GetConversationByUserID(ctx context.Context, userID uuid.UUID) (*models.Conversation, error) {
	query := `
		SELECT c.id, c.user_id, c.subject, c.status, c.last_message_at, c.created_at, c.updated_at,
		       u.id, u.name, u.email, u.avatar
		FROM conversations c
		JOIN users u ON c.user_id = u.id
		WHERE c.user_id = $1 AND c.status = 'open'
		ORDER BY c.created_at DESC
		LIMIT 1
	`

	conversation := &models.Conversation{}
	user := &models.User{}

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&conversation.ID, &conversation.UserID, &conversation.Subject,
		&conversation.Status, &conversation.LastMessageAt,
		&conversation.CreatedAt, &conversation.UpdatedAt,
		&user.ID, &user.Name, &user.Email, &user.Avatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	conversation.User = user
	return conversation, nil
}

// ListConversations lists all conversations with pagination
func (r *MessageRepository) ListConversations(ctx context.Context, filter *models.ConversationFilter, page, perPage int) (*models.PaginatedConversations, error) {
	offset := (page - 1) * perPage

	// Build query conditions
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argNum := 1

	if filter != nil {
		if filter.UserID != nil {
			whereClause += fmt.Sprintf(" AND c.user_id = $%d", argNum)
			args = append(args, *filter.UserID)
			argNum++
		}
		if filter.Status != nil {
			whereClause += fmt.Sprintf(" AND c.status = $%d", argNum)
			args = append(args, *filter.Status)
			argNum++
		}
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM conversations c %s`, whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count conversations: %w", err)
	}

	// Get conversations with last message and unread count
	query := fmt.Sprintf(`
		SELECT c.id, c.user_id, c.subject, c.status, c.last_message_at, c.created_at, c.updated_at,
		       u.id, u.name, u.email, u.avatar,
		       (SELECT COUNT(*) FROM messages m WHERE m.conversation_id = c.id AND m.is_read = false AND m.sender_id = c.user_id) as unread_count
		FROM conversations c
		JOIN users u ON c.user_id = u.id
		%s
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, perPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var conv models.Conversation
		var user models.User

		err := rows.Scan(
			&conv.ID, &conv.UserID, &conv.Subject, &conv.Status,
			&conv.LastMessageAt, &conv.CreatedAt, &conv.UpdatedAt,
			&user.ID, &user.Name, &user.Email, &user.Avatar,
			&conv.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		conv.User = &user

		// Get last message preview
		lastMsg, _ := r.GetLastMessage(ctx, conv.ID)
		conv.LastMessage = lastMsg

		conversations = append(conversations, conv)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedConversations{
		Conversations: conversations,
		Total:         total,
		Page:          page,
		PerPage:       perPage,
		TotalPages:    totalPages,
	}, nil
}

// UpdateConversationStatus updates the status of a conversation
func (r *MessageRepository) UpdateConversationStatus(ctx context.Context, id uuid.UUID, status models.ConversationStatus) error {
	query := `UPDATE conversations SET status = $1 WHERE id = $2`

	result, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update conversation status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("conversation not found")
	}

	return nil
}

// CreateMessage creates a new message in a conversation
func (r *MessageRepository) CreateMessage(ctx context.Context, conversationID, senderID uuid.UUID, content string) (*models.Message, error) {
	message := &models.Message{}
	query := `
		INSERT INTO messages (conversation_id, sender_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, conversation_id, sender_id, content, is_read, read_at, created_at
	`

	err := r.db.QueryRow(ctx, query, conversationID, senderID, content).Scan(
		&message.ID, &message.ConversationID, &message.SenderID,
		&message.Content, &message.IsRead, &message.ReadAt, &message.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation's last_message_at
	_, err = r.db.Exec(ctx, `UPDATE conversations SET last_message_at = $1 WHERE id = $2`, message.CreatedAt, conversationID)
	if err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to update last_message_at: %v\n", err)
	}

	return message, nil
}

// GetMessageByID retrieves a message by ID with sender info
func (r *MessageRepository) GetMessageByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.content, m.is_read, m.read_at, m.created_at,
		       u.id, u.name, u.email, u.avatar
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.id = $1
	`

	message := &models.Message{}
	sender := &models.User{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&message.ID, &message.ConversationID, &message.SenderID,
		&message.Content, &message.IsRead, &message.ReadAt, &message.CreatedAt,
		&sender.ID, &sender.Name, &sender.Email, &sender.Avatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	message.Sender = sender
	return message, nil
}

// ListMessages lists messages in a conversation with pagination
func (r *MessageRepository) ListMessages(ctx context.Context, conversationID uuid.UUID, page, perPage int) (*models.PaginatedMessages, error) {
	offset := (page - 1) * perPage

	// Count total
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`, conversationID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.content, m.is_read, m.read_at, m.created_at,
		       u.id, u.name, u.email, u.avatar
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, conversationID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var sender models.User

		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID,
			&msg.Content, &msg.IsRead, &msg.ReadAt, &msg.CreatedAt,
			&sender.ID, &sender.Name, &sender.Email, &sender.Avatar,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Sender = &sender
		messages = append(messages, msg)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedMessages{
		Messages:   messages,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GetLastMessage gets the last message in a conversation
func (r *MessageRepository) GetLastMessage(ctx context.Context, conversationID uuid.UUID) (*models.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.content, m.is_read, m.read_at, m.created_at,
		       u.id, u.name, u.email, u.avatar
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at DESC
		LIMIT 1
	`

	message := &models.Message{}
	sender := &models.User{}

	err := r.db.QueryRow(ctx, query, conversationID).Scan(
		&message.ID, &message.ConversationID, &message.SenderID,
		&message.Content, &message.IsRead, &message.ReadAt, &message.CreatedAt,
		&sender.ID, &sender.Name, &sender.Email, &sender.Avatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last message: %w", err)
	}

	message.Sender = sender
	return message, nil
}

// MarkMessagesAsRead marks all messages in a conversation as read for a recipient
func (r *MessageRepository) MarkMessagesAsRead(ctx context.Context, conversationID, readerID uuid.UUID) error {
	// Mark messages as read where the reader is NOT the sender
	query := `
		UPDATE messages
		SET is_read = true, read_at = NOW()
		WHERE conversation_id = $1 AND sender_id != $2 AND is_read = false
	`

	_, err := r.db.Exec(ctx, query, conversationID, readerID)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	return nil
}

// GetUnreadCounts gets unread message counts for a user
func (r *MessageRepository) GetUnreadCounts(ctx context.Context, userID uuid.UUID, isAdmin bool) (*models.UnreadCounts, error) {
	counts := &models.UnreadCounts{}

	if isAdmin {
		// Admin sees unread messages from all users (messages where sender is not admin)
		query := `
			SELECT
				COUNT(DISTINCT m.id) as total_messages,
				COUNT(DISTINCT c.id) as total_conversations
			FROM messages m
			JOIN conversations c ON m.conversation_id = c.id
			JOIN users u ON m.sender_id = u.id
			JOIN roles r ON u.role_id = r.id
			WHERE m.is_read = false AND r.slug != 'admin'
		`
		err := r.db.QueryRow(ctx, query).Scan(&counts.Total, &counts.Conversations)
		if err != nil {
			return nil, fmt.Errorf("failed to get admin unread counts: %w", err)
		}
	} else {
		// User sees unread messages in their conversations (from admins)
		query := `
			SELECT
				COUNT(DISTINCT m.id) as total_messages,
				COUNT(DISTINCT c.id) as total_conversations
			FROM messages m
			JOIN conversations c ON m.conversation_id = c.id
			WHERE c.user_id = $1 AND m.is_read = false AND m.sender_id != $1
		`
		err := r.db.QueryRow(ctx, query, userID).Scan(&counts.Total, &counts.Conversations)
		if err != nil {
			return nil, fmt.Errorf("failed to get user unread counts: %w", err)
		}
	}

	return counts, nil
}

// GetUserConversations gets all conversations for a specific user
func (r *MessageRepository) GetUserConversations(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error) {
	query := `
		SELECT c.id, c.user_id, c.subject, c.status, c.last_message_at, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM messages m WHERE m.conversation_id = c.id AND m.is_read = false AND m.sender_id != $1) as unread_count
		FROM conversations c
		WHERE c.user_id = $1
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var conv models.Conversation

		err := rows.Scan(
			&conv.ID, &conv.UserID, &conv.Subject, &conv.Status,
			&conv.LastMessageAt, &conv.CreatedAt, &conv.UpdatedAt,
			&conv.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		// Get last message preview
		lastMsg, _ := r.GetLastMessage(ctx, conv.ID)
		conv.LastMessage = lastMsg

		conversations = append(conversations, conv)
	}

	return conversations, nil
}
