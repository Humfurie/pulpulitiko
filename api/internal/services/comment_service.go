package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

// Profanity word list (common profanity to flag for review)
// This is a basic list - consider using a more comprehensive library in production
var profanityWords = []string{
	"fuck", "shit", "damn", "bitch", "ass", "asshole",
	"bastard", "crap", "dick", "piss", "slut", "whore",
	"putang", "puta", "gago", "bobo", "tanga", "ulol", "tarantado",
	"leche", "bwisit", "tangina", "pakyu", "punyeta",
}

// containsProfanity checks if content contains any profane words
func containsProfanity(content string) bool {
	lowerContent := strings.ToLower(content)
	for _, word := range profanityWords {
		if strings.Contains(lowerContent, word) {
			return true
		}
	}
	return false
}

type CommentService struct {
	repo                *repository.CommentRepository
	articleRepo         *repository.ArticleRepository
	notificationService *NotificationService
	sanitizer           *HTMLSanitizer
}

func NewCommentService(repo *repository.CommentRepository, articleRepo *repository.ArticleRepository, notificationService *NotificationService, sanitizer *HTMLSanitizer) *CommentService {
	return &CommentService{
		repo:                repo,
		articleRepo:         articleRepo,
		notificationService: notificationService,
		sanitizer:           sanitizer,
	}
}

// CreateComment creates a new comment on an article
func (s *CommentService) CreateComment(ctx context.Context, articleSlug string, userID uuid.UUID, req *models.CreateCommentRequest) (*models.Comment, error) {
	// Get article by slug
	article, err := s.articleRepo.GetBySlug(ctx, articleSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}
	if article == nil {
		return nil, fmt.Errorf("article not found")
	}

	// Check if this is a reply and get parent comment
	var parentComment *models.Comment
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
		if parentComment.ArticleID != article.ID {
			return nil, fmt.Errorf("parent comment belongs to different article")
		}
		// Single-level threading is enforced at DB level
	}

	// Sanitize comment content to prevent XSS attacks
	sanitizedContent := s.sanitizer.SanitizeComment(req.Content)

	// Create sanitized request
	sanitizedReq := &models.CreateCommentRequest{
		Content:  sanitizedContent,
		ParentID: req.ParentID,
	}

	// Determine initial status based on profanity check
	status := models.CommentStatusActive
	if containsProfanity(sanitizedContent) {
		status = models.CommentStatusUnderReview
	}

	comment, err := s.repo.Create(ctx, article.ID, userID, sanitizedReq, status)
	if err != nil {
		return nil, err
	}

	// Process mentions and create notifications
	if s.notificationService != nil {
		// Save mentions and get mentioned user IDs (use sanitized content)
		mentionedUserIDs, _ := s.repo.SaveMentions(ctx, comment.ID, sanitizedContent)

		// Create notifications for mentions
		for _, mentionedUserID := range mentionedUserIDs {
			_ = s.notificationService.CreateMentionNotification(
				ctx,
				mentionedUserID,
				userID,
				"article",
				&article.ID,
				nil,
				&comment.ID,
				article.Title,
			)
		}

		// Create notification for reply
		if parentComment != nil {
			_ = s.notificationService.CreateReplyNotification(
				ctx,
				parentComment.UserID,
				userID,
				"article",
				&article.ID,
				nil,
				&comment.ID,
				article.Title,
			)
		}
	}

	// Fetch full comment with user info
	return s.repo.GetByID(ctx, comment.ID)
}

// GetComment retrieves a single comment
func (s *CommentService) GetComment(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	return s.repo.GetByID(ctx, id)
}

// ListArticleComments lists all comments for an article
// includeHidden is for admins only to see moderated comments
func (s *CommentService) ListArticleComments(ctx context.Context, articleSlug string, currentUserID *uuid.UUID, includeHidden bool) ([]models.Comment, error) {
	// Get article by slug
	article, err := s.articleRepo.GetBySlug(ctx, articleSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}
	if article == nil {
		return nil, fmt.Errorf("article not found")
	}

	return s.repo.ListByArticle(ctx, article.ID, currentUserID, includeHidden)
}

// ListReplies lists all replies to a comment
// includeHidden is for admins only to see moderated comments
func (s *CommentService) ListReplies(ctx context.Context, commentID uuid.UUID, currentUserID *uuid.UUID, includeHidden bool) ([]models.Comment, error) {
	return s.repo.ListReplies(ctx, commentID, currentUserID, includeHidden)
}

// UpdateComment updates a comment's content
func (s *CommentService) UpdateComment(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *models.UpdateCommentRequest) (*models.Comment, error) {
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

	// Sanitize comment content to prevent XSS attacks
	sanitizedContent := s.sanitizer.SanitizeComment(req.Content)

	if err := s.repo.Update(ctx, id, sanitizedContent); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

// DeleteComment soft deletes a comment
func (s *CommentService) DeleteComment(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error {
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
func (s *CommentService) AddReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
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
func (s *CommentService) RemoveReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
	return s.repo.RemoveReaction(ctx, commentID, userID, reaction)
}

// GetReplyPreview gets a preview of replies for collapsed view
func (s *CommentService) GetReplyPreview(ctx context.Context, commentID uuid.UUID) (*models.ReplyPreview, error) {
	return s.repo.GetReplyPreview(ctx, commentID)
}

// GetCommentCount returns total comment count for an article
func (s *CommentService) GetCommentCount(ctx context.Context, articleSlug string) (int, error) {
	article, err := s.articleRepo.GetBySlug(ctx, articleSlug)
	if err != nil {
		return 0, err
	}
	if article == nil {
		return 0, fmt.Errorf("article not found")
	}

	return s.repo.GetCommentCount(ctx, article.ID)
}

// ModerateComment updates a comment's moderation status (admin only)
func (s *CommentService) ModerateComment(ctx context.Context, commentID uuid.UUID, moderatorID uuid.UUID, req *models.ModerateCommentRequest) (*models.Comment, error) {
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

// ListAllComments lists all comments for admin moderation panel
func (s *CommentService) ListAllComments(ctx context.Context, filter *models.CommentFilter, currentUserID *uuid.UUID) ([]models.Comment, error) {
	return s.repo.ListAllComments(ctx, filter, currentUserID)
}
