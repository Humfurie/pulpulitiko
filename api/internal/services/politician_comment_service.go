package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type PoliticianCommentService struct {
	repo                *repository.PoliticianCommentRepository
	politicianRepo      *repository.PoliticianRepository
	notificationService *NotificationService
}

func NewPoliticianCommentService(repo *repository.PoliticianCommentRepository, politicianRepo *repository.PoliticianRepository, notificationService *NotificationService) *PoliticianCommentService {
	return &PoliticianCommentService{
		repo:                repo,
		politicianRepo:      politicianRepo,
		notificationService: notificationService,
	}
}

// CreateComment creates a new comment on a politician page
func (s *PoliticianCommentService) CreateComment(ctx context.Context, politicianSlug string, userID uuid.UUID, req *models.CreateCommentRequest) (*models.PoliticianComment, error) {
	// Get politician by slug
	politician, err := s.politicianRepo.GetBySlug(ctx, politicianSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get politician: %w", err)
	}
	if politician == nil {
		return nil, fmt.Errorf("politician not found")
	}

	// Check if this is a reply and get parent comment
	var parentComment *models.PoliticianComment
	if req.ParentID != nil && *req.ParentID != "" {
		parentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id")
		}

		parentComment, err = s.repo.GetByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get parent comment: %w", err)
		}
		if parentComment == nil {
			return nil, fmt.Errorf("parent comment not found")
		}
		if parentComment.PoliticianID != politician.ID {
			return nil, fmt.Errorf("parent comment belongs to different politician")
		}
	}

	// Determine initial status based on profanity check
	status := models.CommentStatusActive
	if containsPoliticianCommentProfanity(req.Content) {
		status = models.CommentStatusUnderReview
	}

	comment, err := s.repo.Create(ctx, politician.ID, userID, req, status)
	if err != nil {
		return nil, err
	}

	// Process mentions and create notifications
	if s.notificationService != nil {
		// Save mentions and get mentioned user IDs
		mentionedUserIDs, _ := s.repo.SaveMentions(ctx, comment.ID, req.Content)

		// Create notifications for mentions
		for _, mentionedUserID := range mentionedUserIDs {
			_ = s.notificationService.CreateMentionNotification(
				ctx,
				mentionedUserID,
				userID,
				"politician",
				nil,
				&politician.ID,
				&comment.ID,
				politician.Name,
			)
		}

		// Create notification for reply
		if parentComment != nil {
			_ = s.notificationService.CreateReplyNotification(
				ctx,
				parentComment.UserID,
				userID,
				"politician",
				nil,
				&politician.ID,
				&comment.ID,
				politician.Name,
			)
		}
	}

	// Fetch full comment with user info
	return s.repo.GetByID(ctx, comment.ID)
}

// GetComment retrieves a single comment
func (s *PoliticianCommentService) GetComment(ctx context.Context, id uuid.UUID) (*models.PoliticianComment, error) {
	return s.repo.GetByID(ctx, id)
}

// ListPoliticianComments lists paginated comments for a politician
func (s *PoliticianCommentService) ListPoliticianComments(ctx context.Context, politicianSlug string, currentUserID *uuid.UUID, includeHidden bool, page, perPage int) (*models.PaginatedPoliticianComments, error) {
	// Get politician by slug
	politician, err := s.politicianRepo.GetBySlug(ctx, politicianSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get politician: %w", err)
	}
	if politician == nil {
		return nil, fmt.Errorf("politician not found")
	}

	return s.repo.ListByPolitician(ctx, politician.ID, currentUserID, includeHidden, page, perPage)
}

// ListReplies lists all replies to a comment
func (s *PoliticianCommentService) ListReplies(ctx context.Context, commentID uuid.UUID, currentUserID *uuid.UUID, includeHidden bool) ([]models.PoliticianComment, error) {
	return s.repo.ListReplies(ctx, commentID, currentUserID, includeHidden)
}

// UpdateComment updates a comment's content
func (s *PoliticianCommentService) UpdateComment(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *models.UpdateCommentRequest) (*models.PoliticianComment, error) {
	// Get comment to verify ownership
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, fmt.Errorf("comment not found")
	}
	if comment.UserID != userID {
		return nil, fmt.Errorf("not authorized to edit this comment")
	}

	if err := s.repo.Update(ctx, id, req.Content); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

// DeleteComment soft deletes a comment
func (s *PoliticianCommentService) DeleteComment(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error {
	// Get comment to verify ownership
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if comment == nil {
		return fmt.Errorf("comment not found")
	}

	// Only owner or admin can delete
	if comment.UserID != userID && !isAdmin {
		return fmt.Errorf("not authorized to delete this comment")
	}

	return s.repo.Delete(ctx, id)
}

// AddReaction adds a reaction to a comment
func (s *PoliticianCommentService) AddReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
	if !models.IsValidReaction(reaction) {
		return fmt.Errorf("invalid reaction type")
	}

	// Verify comment exists
	comment, err := s.repo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if comment == nil {
		return fmt.Errorf("comment not found")
	}

	return s.repo.AddReaction(ctx, commentID, userID, reaction)
}

// RemoveReaction removes a reaction from a comment
func (s *PoliticianCommentService) RemoveReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
	return s.repo.RemoveReaction(ctx, commentID, userID, reaction)
}

// GetCommentCount returns total comment count for a politician
func (s *PoliticianCommentService) GetCommentCount(ctx context.Context, politicianSlug string) (int, error) {
	politician, err := s.politicianRepo.GetBySlug(ctx, politicianSlug)
	if err != nil {
		return 0, err
	}
	if politician == nil {
		return 0, fmt.Errorf("politician not found")
	}

	return s.repo.GetCommentCount(ctx, politician.ID)
}

// ModerateComment updates a comment's moderation status (admin only)
func (s *PoliticianCommentService) ModerateComment(ctx context.Context, commentID uuid.UUID, moderatorID uuid.UUID, req *models.ModerateCommentRequest) (*models.PoliticianComment, error) {
	// Verify comment exists
	comment, err := s.repo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, fmt.Errorf("comment not found")
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, commentID, req.Status, moderatorID, req.Reason); err != nil {
		return nil, err
	}

	// Return updated comment
	return s.repo.GetByID(ctx, commentID)
}

// containsPoliticianCommentProfanity checks if content contains profane words
func containsPoliticianCommentProfanity(content string) bool {
	lowerContent := strings.ToLower(content)
	for _, word := range profanityWords {
		if strings.Contains(lowerContent, word) {
			return true
		}
	}
	return false
}
