package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	notification := &models.Notification{}
	query := `
		INSERT INTO notifications (user_id, type, title, message, actor_id, article_id, politician_id, comment_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, type, title, message, actor_id, article_id, politician_id, comment_id, is_read, read_at, created_at
	`

	err := r.db.QueryRow(ctx, query,
		req.UserID, req.Type, req.Title, req.Message,
		req.ActorID, req.ArticleID, req.PoliticianID, req.CommentID,
	).Scan(
		&notification.ID, &notification.UserID, &notification.Type, &notification.Title, &notification.Message,
		&notification.ActorID, &notification.ArticleID, &notification.PoliticianID, &notification.CommentID,
		&notification.IsRead, &notification.ReadAt, &notification.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT n.id, n.user_id, n.type, n.title, n.message, n.actor_id, n.article_id, n.politician_id, n.comment_id,
		       n.is_read, n.read_at, n.created_at,
		       u.id, u.name, u.avatar
		FROM notifications n
		LEFT JOIN users u ON n.actor_id = u.id
		WHERE n.id = $1
	`

	notification := &models.Notification{}
	var actor models.NotificationActor
	var actorID, actorName, actorAvatar interface{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&notification.ID, &notification.UserID, &notification.Type, &notification.Title, &notification.Message,
		&notification.ActorID, &notification.ArticleID, &notification.PoliticianID, &notification.CommentID,
		&notification.IsRead, &notification.ReadAt, &notification.CreatedAt,
		&actorID, &actorName, &actorAvatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if actorID != nil {
		actor.ID = actorID.(uuid.UUID)
		actor.Name = actorName.(string)
		if actorAvatar != nil {
			av := actorAvatar.(string)
			actor.Avatar = &av
		}
		notification.Actor = &actor
	}

	return notification, nil
}

// ListByUser retrieves paginated notifications for a user
func (r *NotificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, perPage int, unreadOnly bool) (*models.PaginatedNotifications, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	// Build filter
	filter := ""
	if unreadOnly {
		filter = " AND n.is_read = FALSE"
	}

	// Count total
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM notifications n WHERE n.user_id = $1%s`, filter)
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Count unread
	var unreadCount int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`, userID).Scan(&unreadCount); err != nil {
		return nil, fmt.Errorf("failed to count unread: %w", err)
	}

	// Get notifications with related data
	query := fmt.Sprintf(`
		SELECT n.id, n.user_id, n.type, n.title, n.message, n.actor_id, n.article_id, n.politician_id, n.comment_id,
		       n.is_read, n.read_at, n.created_at,
		       u.id, u.name, u.avatar,
		       a.id, a.title, a.slug,
		       p.id, p.name, p.slug
		FROM notifications n
		LEFT JOIN users u ON n.actor_id = u.id
		LEFT JOIN articles a ON n.article_id = a.id
		LEFT JOIN politicians p ON n.politician_id = p.id
		WHERE n.user_id = $1%s
		ORDER BY n.created_at DESC
		LIMIT $2 OFFSET $3
	`, filter)

	rows, err := r.db.Query(ctx, query, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		var actorID *uuid.UUID
		var actorName, actorAvatar *string
		var articleID *uuid.UUID
		var articleTitle, articleSlug *string
		var politicianID *uuid.UUID
		var politicianName, politicianSlug *string

		err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.ActorID, &n.ArticleID, &n.PoliticianID, &n.CommentID,
			&n.IsRead, &n.ReadAt, &n.CreatedAt,
			&actorID, &actorName, &actorAvatar,
			&articleID, &articleTitle, &articleSlug,
			&politicianID, &politicianName, &politicianSlug,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if actorID != nil && actorName != nil {
			n.Actor = &models.NotificationActor{
				ID:     *actorID,
				Name:   *actorName,
				Avatar: actorAvatar,
			}
		}

		if articleID != nil && articleTitle != nil && articleSlug != nil {
			n.ArticleRef = &models.NotificationRef{
				ID:   *articleID,
				Name: *articleTitle,
				Slug: *articleSlug,
			}
		}

		if politicianID != nil && politicianName != nil && politicianSlug != nil {
			n.PoliticianRef = &models.NotificationRef{
				ID:   *politicianID,
				Name: *politicianName,
				Slug: *politicianSlug,
			}
		}

		notifications = append(notifications, n)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedNotifications{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          page,
		PerPage:       perPage,
		TotalPages:    totalPages,
	}, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE, read_at = NOW() WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to mark as read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE, read_at = NOW() WHERE user_id = $1 AND is_read = FALSE`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all as read: %w", err)
	}

	return nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`, userID).Scan(&count)
	return count, err
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}
