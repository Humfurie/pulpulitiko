package repository

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepository struct {
	db *pgxpool.Pool
}

func NewCommentRepository(db *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new comment
func (r *CommentRepository) Create(ctx context.Context, articleID, userID uuid.UUID, req *models.CreateCommentRequest, status models.CommentStatus) (*models.Comment, error) {
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		parsed, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
		parentID = &parsed
	}

	comment := &models.Comment{}
	query := `
		INSERT INTO comments (article_id, user_id, parent_id, content, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, article_id, user_id, parent_id, content, status, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, articleID, userID, parentID, req.Content, status).Scan(
		&comment.ID, &comment.ArticleID, &comment.UserID, &comment.ParentID,
		&comment.Content, &comment.Status, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Extract and save mentions
	mentions := extractMentions(req.Content)
	if len(mentions) > 0 {
		if err := r.saveMentions(ctx, comment.ID, mentions); err != nil {
			// Log but don't fail - mentions are secondary
			fmt.Printf("Warning: failed to save mentions: %v\n", err)
		}
	}

	return comment, nil
}

// GetByID retrieves a comment by ID with user info
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	query := `
		SELECT c.id, c.article_id, c.user_id, c.parent_id, c.content, c.status,
		       c.moderated_by, c.moderated_at, c.moderation_reason,
		       c.created_at, c.updated_at, c.deleted_at,
		       u.id, u.name, u.avatar, COALESCE(u.is_system, false)
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`

	comment := &models.Comment{}
	author := &models.CommentAuthor{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&comment.ID, &comment.ArticleID, &comment.UserID, &comment.ParentID,
		&comment.Content, &comment.Status,
		&comment.ModeratedBy, &comment.ModeratedAt, &comment.ModerationReason,
		&comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt,
		&author.ID, &author.Name, &author.Avatar, &author.IsSystem,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	comment.Author = author
	return comment, nil
}

// ListByArticle retrieves all root comments for an article with replies
// Only shows 'active' comments to regular users. Admin can see all via includeHidden parameter.
func (r *CommentRepository) ListByArticle(ctx context.Context, articleID uuid.UUID, currentUserID *uuid.UUID, includeHidden bool) ([]models.Comment, error) {
	// Get root comments (parent_id IS NULL)
	// Only show active comments unless admin requests hidden ones
	statusFilter := "AND c.status = 'active'"
	if includeHidden {
		statusFilter = "" // Admin can see all
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.article_id, c.user_id, c.parent_id, c.content, c.status,
		       c.created_at, c.updated_at,
		       u.id, u.name, u.avatar, COALESCE(u.is_system, false),
		       (SELECT COUNT(*) FROM comments r WHERE r.parent_id = c.id AND r.deleted_at IS NULL AND r.status = 'active') as reply_count
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.article_id = $1 AND c.parent_id IS NULL AND c.deleted_at IS NULL %s
		ORDER BY c.created_at DESC
	`, statusFilter)

	rows, err := r.db.Query(ctx, query, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		var author models.CommentAuthor

		err := rows.Scan(
			&comment.ID, &comment.ArticleID, &comment.UserID, &comment.ParentID,
			&comment.Content, &comment.Status, &comment.CreatedAt, &comment.UpdatedAt,
			&author.ID, &author.Name, &author.Avatar, &author.IsSystem,
			&comment.ReplyCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		comment.Author = &author

		// Get reactions for this comment
		reactions, err := r.GetReactionSummary(ctx, comment.ID, currentUserID)
		if err == nil {
			comment.Reactions = reactions
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

// ListReplies retrieves all replies for a parent comment
// Only shows 'active' replies unless includeHidden is true (admin only)
func (r *CommentRepository) ListReplies(ctx context.Context, parentID uuid.UUID, currentUserID *uuid.UUID, includeHidden bool) ([]models.Comment, error) {
	statusFilter := "AND c.status = 'active'"
	if includeHidden {
		statusFilter = ""
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.article_id, c.user_id, c.parent_id, c.content, c.status,
		       c.created_at, c.updated_at,
		       u.id, u.name, u.avatar, COALESCE(u.is_system, false)
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.parent_id = $1 AND c.deleted_at IS NULL %s
		ORDER BY c.created_at ASC
	`, statusFilter)

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list replies: %w", err)
	}
	defer rows.Close()

	var replies []models.Comment
	for rows.Next() {
		var comment models.Comment
		var author models.CommentAuthor

		err := rows.Scan(
			&comment.ID, &comment.ArticleID, &comment.UserID, &comment.ParentID,
			&comment.Content, &comment.Status, &comment.CreatedAt, &comment.UpdatedAt,
			&author.ID, &author.Name, &author.Avatar, &author.IsSystem,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reply: %w", err)
		}

		comment.Author = &author

		// Get reactions for this reply
		reactions, err := r.GetReactionSummary(ctx, comment.ID, currentUserID)
		if err == nil {
			comment.Reactions = reactions
		}

		replies = append(replies, comment)
	}

	return replies, nil
}

// Update updates a comment's content
func (r *CommentRepository) Update(ctx context.Context, id uuid.UUID, content string) error {
	query := `UPDATE comments SET content = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, content, id)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	// Update mentions
	mentions := extractMentions(content)
	// Clear old mentions and save new ones
	r.db.Exec(ctx, `DELETE FROM comment_mentions WHERE comment_id = $1`, id)
	if len(mentions) > 0 {
		r.saveMentions(ctx, id, mentions)
	}

	return nil
}

// Delete soft deletes a comment
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE comments SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// UpdateStatus updates the moderation status of a comment (admin only)
func (r *CommentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.CommentStatus, moderatorID uuid.UUID, reason *string) error {
	query := `
		UPDATE comments
		SET status = $1, moderated_by = $2, moderated_at = NOW(), moderation_reason = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, status, moderatorID, reason, id)
	if err != nil {
		return fmt.Errorf("failed to update comment status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// ListAllComments lists all comments for admin moderation (all statuses)
func (r *CommentRepository) ListAllComments(ctx context.Context, filter *models.CommentFilter, currentUserID *uuid.UUID) ([]models.Comment, error) {
	query := `
		SELECT c.id, c.article_id, c.user_id, c.parent_id, c.content, c.status,
		       c.moderated_by, c.moderated_at, c.moderation_reason,
		       c.created_at, c.updated_at,
		       u.id, u.name, u.avatar, COALESCE(u.is_system, false),
		       a.slug as article_slug, a.title as article_title
		FROM comments c
		JOIN users u ON c.user_id = u.id
		JOIN articles a ON c.article_id = a.id
		WHERE c.deleted_at IS NULL
	`

	args := []interface{}{}
	argNum := 1

	if filter != nil && filter.Status != nil {
		query += fmt.Sprintf(" AND c.status = $%d", argNum)
		args = append(args, *filter.Status)
		argNum++
	}

	query += " ORDER BY c.created_at DESC LIMIT 100"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list all comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		var author models.CommentAuthor
		var articleSlug, articleTitle string

		err := rows.Scan(
			&comment.ID, &comment.ArticleID, &comment.UserID, &comment.ParentID,
			&comment.Content, &comment.Status,
			&comment.ModeratedBy, &comment.ModeratedAt, &comment.ModerationReason,
			&comment.CreatedAt, &comment.UpdatedAt,
			&author.ID, &author.Name, &author.Avatar, &author.IsSystem,
			&articleSlug, &articleTitle,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		comment.Author = &author
		// Store article info in a way the frontend can use (could add ArticleSlug field to model later)
		_ = articleSlug
		_ = articleTitle

		// Get reactions
		reactions, err := r.GetReactionSummary(ctx, comment.ID, currentUserID)
		if err == nil {
			comment.Reactions = reactions
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

// AddReaction adds a reaction to a comment (replaces any existing reaction)
func (r *CommentRepository) AddReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
	query := `
		INSERT INTO comment_reactions (comment_id, user_id, reaction)
		VALUES ($1, $2, $3)
		ON CONFLICT (comment_id, user_id) DO UPDATE SET reaction = EXCLUDED.reaction, created_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, commentID, userID, reaction)
	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	return nil
}

// RemoveReaction removes a reaction from a comment
func (r *CommentRepository) RemoveReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
	query := `DELETE FROM comment_reactions WHERE comment_id = $1 AND user_id = $2 AND reaction = $3`

	_, err := r.db.Exec(ctx, query, commentID, userID, reaction)
	if err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	return nil
}

// GetReactionSummary gets reaction counts for a comment
func (r *CommentRepository) GetReactionSummary(ctx context.Context, commentID uuid.UUID, currentUserID *uuid.UUID) ([]models.ReactionSummary, error) {
	query := `
		SELECT reaction, COUNT(*) as count
		FROM comment_reactions
		WHERE comment_id = $1
		GROUP BY reaction
		ORDER BY count DESC
	`

	rows, err := r.db.Query(ctx, query, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reaction summary: %w", err)
	}
	defer rows.Close()

	var summaries []models.ReactionSummary
	for rows.Next() {
		var summary models.ReactionSummary
		if err := rows.Scan(&summary.Reaction, &summary.Count); err != nil {
			return nil, err
		}

		// Check if current user has this reaction
		if currentUserID != nil {
			var hasReacted bool
			r.db.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM comment_reactions WHERE comment_id = $1 AND user_id = $2 AND reaction = $3)
			`, commentID, *currentUserID, summary.Reaction).Scan(&hasReacted)
			summary.HasReacted = hasReacted
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetReplyPreview gets a preview of replies for collapsed view
func (r *CommentRepository) GetReplyPreview(ctx context.Context, parentID uuid.UUID) (*models.ReplyPreview, error) {
	// Get count
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM comments WHERE parent_id = $1 AND deleted_at IS NULL
	`, parentID).Scan(&count)
	if err != nil {
		return nil, err
	}

	// Get first few users who commented
	query := `
		SELECT DISTINCT u.id, u.name, u.avatar
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.parent_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at ASC
		LIMIT 5
	`

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []models.CommentAuthor
	for rows.Next() {
		var author models.CommentAuthor
		if err := rows.Scan(&author.ID, &author.Name, &author.Avatar); err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	return &models.ReplyPreview{
		Count:   count,
		Authors: authors,
	}, nil
}

// GetCommentCount returns total comment count for an article
func (r *CommentRepository) GetCommentCount(ctx context.Context, articleID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM comments WHERE article_id = $1 AND deleted_at IS NULL
	`, articleID).Scan(&count)
	return count, err
}

// saveMentions saves @mentions for a comment
func (r *CommentRepository) saveMentions(ctx context.Context, commentID uuid.UUID, mentions []string) error {
	for _, slug := range mentions {
		// Get author by slug (mentions still reference authors for article writers)
		var authorID uuid.UUID
		err := r.db.QueryRow(ctx, `SELECT id FROM authors WHERE slug = $1 AND deleted_at IS NULL`, slug).Scan(&authorID)
		if err != nil {
			continue // Skip invalid mentions
		}

		_, err = r.db.Exec(ctx, `
			INSERT INTO comment_mentions (comment_id, mentioned_author_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, commentID, authorID)
		if err != nil {
			continue
		}
	}
	return nil
}

// extractMentions extracts @username mentions from content
func extractMentions(content string) []string {
	re := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	var mentions []string
	seen := make(map[string]bool)
	for _, match := range matches {
		slug := strings.ToLower(match[1])
		if !seen[slug] {
			mentions = append(mentions, slug)
			seen[slug] = true
		}
	}
	return mentions
}

// GetMentions gets all mentions for a comment
func (r *CommentRepository) GetMentions(ctx context.Context, commentID uuid.UUID) ([]models.CommentMention, error) {
	query := `
		SELECT m.id, m.comment_id, m.mentioned_author_id, m.created_at,
		       a.id, a.name, a.avatar
		FROM comment_mentions m
		JOIN authors a ON m.mentioned_author_id = a.id
		WHERE m.comment_id = $1
	`

	rows, err := r.db.Query(ctx, query, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mentions []models.CommentMention
	for rows.Next() {
		var mention models.CommentMention
		var author models.CommentAuthor

		if err := rows.Scan(
			&mention.ID, &mention.CommentID, &mention.MentionedAuthorID, &mention.CreatedAt,
			&author.ID, &author.Name, &author.Avatar,
		); err != nil {
			return nil, err
		}

		mention.MentionedAuthor = &author
		mentions = append(mentions, mention)
	}

	return mentions, nil
}
