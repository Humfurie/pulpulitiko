package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type NotificationService struct {
	repo     *repository.NotificationRepository
	userRepo *repository.UserRepository
}

func NewNotificationService(repo *repository.NotificationRepository, userRepo *repository.UserRepository) *NotificationService {
	return &NotificationService{
		repo:     repo,
		userRepo: userRepo,
	}
}

// CreateMentionNotification creates a notification for a mention
func (s *NotificationService) CreateMentionNotification(ctx context.Context, mentionedUserID, actorID uuid.UUID, commentType string, articleID, politicianID, commentID *uuid.UUID, targetName string) error {
	// Don't notify yourself
	if mentionedUserID == actorID {
		return nil
	}

	// Get actor name for the title
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil || actor == nil {
		return fmt.Errorf("failed to get actor: %w", err)
	}

	var notifType models.NotificationType
	var message string

	if commentType == "article" {
		notifType = models.NotificationTypeMentionArticleComment
		message = fmt.Sprintf("mentioned you in a comment on \"%s\"", targetName)
	} else {
		notifType = models.NotificationTypeMentionPoliticianComment
		message = fmt.Sprintf("mentioned you in a discussion about %s", targetName)
	}

	req := &models.CreateNotificationRequest{
		UserID:       mentionedUserID,
		Type:         notifType,
		Title:        fmt.Sprintf("%s mentioned you", actor.Name),
		Message:      &message,
		ActorID:      &actorID,
		ArticleID:    articleID,
		PoliticianID: politicianID,
		CommentID:    commentID,
	}

	_, err = s.repo.Create(ctx, req)
	return err
}

// CreateReplyNotification creates a notification for a reply
func (s *NotificationService) CreateReplyNotification(ctx context.Context, parentCommentUserID, actorID uuid.UUID, commentType string, articleID, politicianID, commentID *uuid.UUID, targetName string) error {
	// Don't notify yourself
	if parentCommentUserID == actorID {
		return nil
	}

	// Get actor name
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil || actor == nil {
		return fmt.Errorf("failed to get actor: %w", err)
	}

	var notifType models.NotificationType
	var message string

	if commentType == "article" {
		notifType = models.NotificationTypeReplyArticleComment
		message = fmt.Sprintf("replied to your comment on \"%s\"", targetName)
	} else {
		notifType = models.NotificationTypeReplyPoliticianComment
		message = fmt.Sprintf("replied to your comment about %s", targetName)
	}

	req := &models.CreateNotificationRequest{
		UserID:       parentCommentUserID,
		Type:         notifType,
		Title:        fmt.Sprintf("%s replied to you", actor.Name),
		Message:      &message,
		ActorID:      &actorID,
		ArticleID:    articleID,
		PoliticianID: politicianID,
		CommentID:    commentID,
	}

	_, err = s.repo.Create(ctx, req)
	return err
}

// ListNotifications lists paginated notifications for a user
func (s *NotificationService) ListNotifications(ctx context.Context, userID uuid.UUID, page, perPage int, unreadOnly bool) (*models.PaginatedNotifications, error) {
	return s.repo.ListByUser(ctx, userID, page, perPage, unreadOnly)
}

// MarkAsRead marks a single notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// GetUnreadCount returns the count of unread notifications
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}
