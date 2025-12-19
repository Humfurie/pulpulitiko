package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoliticianCommentRepository struct {
	db *pgxpool.Pool
}

func NewPoliticianCommentRepository(db *pgxpool.Pool) *PoliticianCommentRepository {
	return &PoliticianCommentRepository{db: db}
}

// Create creates a new politician comment
func (r *PoliticianCommentRepository) Create(ctx context.Context, politicianID, userID uuid.UUID, req *models.CreateCommentRequest, status models.CommentStatus) (*models.PoliticianComment, error) {
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		parsed, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
		parentID = &parsed
	}

	comment := &models.PoliticianComment{}
	query := `
		INSERT INTO politician_comments (politician_id, user_id, parent_id, content, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, politician_id, user_id, parent_id, content, status, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, politicianID, userID, parentID, req.Content, status).Scan(
		&comment.ID, &comment.PoliticianID, &comment.UserID, &comment.ParentID,
		&comment.Content, &comment.Status, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

// GetByID retrieves a politician comment by ID with user info
func (r *PoliticianCommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PoliticianComment, error) {
	query := `
		SELECT c.id, c.politician_id, c.user_id, c.parent_id, c.content, c.status,
		       c.moderated_by, c.moderated_at, c.moderation_reason,
		       c.created_at, c.updated_at, c.deleted_at,
		       u.id, u.name, u.avatar, COALESCE(u.is_system, false)
		FROM politician_comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`

	comment := &models.PoliticianComment{}
	author := &models.CommentAuthor{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&comment.ID, &comment.PoliticianID, &comment.UserID, &comment.ParentID,
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

// ListByPolitician retrieves paginated root comments for a politician
func (r *PoliticianCommentRepository) ListByPolitician(ctx context.Context, politicianID uuid.UUID, currentUserID *uuid.UUID, includeHidden bool, page, perPage int) (*models.PaginatedPoliticianComments, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	// Count total root comments
	statusFilter := "AND c.status = 'active'"
	if includeHidden {
		statusFilter = ""
	}

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM politician_comments c
		WHERE c.politician_id = $1 AND c.parent_id IS NULL AND c.deleted_at IS NULL %s
	`, statusFilter)
	if err := r.db.QueryRow(ctx, countQuery, politicianID).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count comments: %w", err)
	}

	// Get paginated root comments
	query := fmt.Sprintf(`
		SELECT c.id, c.politician_id, c.user_id, c.parent_id, c.content, c.status,
		       c.created_at, c.updated_at,
		       u.id, u.name, u.avatar, COALESCE(u.is_system, false),
		       (SELECT COUNT(*) FROM politician_comments r WHERE r.parent_id = c.id AND r.deleted_at IS NULL AND r.status = 'active') as reply_count
		FROM politician_comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.politician_id = $1 AND c.parent_id IS NULL AND c.deleted_at IS NULL %s
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`, statusFilter)

	rows, err := r.db.Query(ctx, query, politicianID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []models.PoliticianComment
	for rows.Next() {
		var comment models.PoliticianComment
		var author models.CommentAuthor

		err := rows.Scan(
			&comment.ID, &comment.PoliticianID, &comment.UserID, &comment.ParentID,
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

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedPoliticianComments{
		Comments:   comments,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// ListReplies retrieves all replies for a parent comment
func (r *PoliticianCommentRepository) ListReplies(ctx context.Context, parentID uuid.UUID, currentUserID *uuid.UUID, includeHidden bool) ([]models.PoliticianComment, error) {
	statusFilter := "AND c.status = 'active'"
	if includeHidden {
		statusFilter = ""
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.politician_id, c.user_id, c.parent_id, c.content, c.status,
		       c.created_at, c.updated_at,
		       u.id, u.name, u.avatar, COALESCE(u.is_system, false)
		FROM politician_comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.parent_id = $1 AND c.deleted_at IS NULL %s
		ORDER BY c.created_at ASC
	`, statusFilter)

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list replies: %w", err)
	}
	defer rows.Close()

	var replies []models.PoliticianComment
	for rows.Next() {
		var comment models.PoliticianComment
		var author models.CommentAuthor

		err := rows.Scan(
			&comment.ID, &comment.PoliticianID, &comment.UserID, &comment.ParentID,
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
func (r *PoliticianCommentRepository) Update(ctx context.Context, id uuid.UUID, content string) error {
	query := `UPDATE politician_comments SET content = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, content, id)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// Delete soft deletes a comment
func (r *PoliticianCommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE politician_comments SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// UpdateStatus updates the moderation status of a comment
func (r *PoliticianCommentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.CommentStatus, moderatorID uuid.UUID, reason *string) error {
	query := `
		UPDATE politician_comments
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

// GetCommentCount returns total comment count for a politician
func (r *PoliticianCommentRepository) GetCommentCount(ctx context.Context, politicianID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM politician_comments WHERE politician_id = $1 AND deleted_at IS NULL AND status = 'active'
	`, politicianID).Scan(&count)
	return count, err
}

// AddReaction adds a reaction to a comment
func (r *PoliticianCommentRepository) AddReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
	query := `
		INSERT INTO politician_comment_reactions (comment_id, user_id, reaction)
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
func (r *PoliticianCommentRepository) RemoveReaction(ctx context.Context, commentID, userID uuid.UUID, reaction string) error {
	query := `DELETE FROM politician_comment_reactions WHERE comment_id = $1 AND user_id = $2 AND reaction = $3`

	_, err := r.db.Exec(ctx, query, commentID, userID, reaction)
	if err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	return nil
}

// GetReactionSummary gets reaction counts for a comment
func (r *PoliticianCommentRepository) GetReactionSummary(ctx context.Context, commentID uuid.UUID, currentUserID *uuid.UUID) ([]models.ReactionSummary, error) {
	query := `
		SELECT reaction, COUNT(*) as count
		FROM politician_comment_reactions
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
			_ = r.db.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM politician_comment_reactions WHERE comment_id = $1 AND user_id = $2 AND reaction = $3)
			`, commentID, *currentUserID, summary.Reaction).Scan(&hasReacted)
			summary.HasReacted = hasReacted
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// SaveMentions saves @mentions for a comment and returns the mentioned user IDs
func (r *PoliticianCommentRepository) SaveMentions(ctx context.Context, commentID uuid.UUID, content string) ([]uuid.UUID, error) {
	// Get all users to match against
	rows, err := r.db.Query(ctx, `SELECT id, name FROM users WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type userInfo struct {
		ID   uuid.UUID
		Name string
	}
	var users []userInfo
	for rows.Next() {
		var u userInfo
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			continue
		}
		users = append(users, u)
	}

	// Sort users by name length (longest first) to match longest names first
	sort.Slice(users, func(i, j int) bool {
		return len(users[i].Name) > len(users[j].Name)
	})

	var mentionedUserIDs []uuid.UUID
	contentLower := strings.ToLower(content)
	seen := make(map[uuid.UUID]bool)

	// Find all @ positions and try to match user names
	for i := 0; i < len(contentLower); i++ {
		if contentLower[i] == '@' && i+1 < len(contentLower) {
			remaining := contentLower[i+1:]
			for _, user := range users {
				nameLower := strings.ToLower(user.Name)
				if strings.HasPrefix(remaining, nameLower) {
					// Check that it's a complete word (followed by space, punctuation, or end)
					afterName := i + 1 + len(nameLower)
					if afterName >= len(contentLower) || !isAlphanumeric(contentLower[afterName]) {
						if !seen[user.ID] {
							seen[user.ID] = true
							// Save mention
							_, err = r.db.Exec(ctx, `
								INSERT INTO politician_comment_mentions (comment_id, mentioned_user_id)
								VALUES ($1, $2)
								ON CONFLICT DO NOTHING
							`, commentID, user.ID)
							if err == nil {
								mentionedUserIDs = append(mentionedUserIDs, user.ID)
							}
						}
						break // Found longest match, move on
					}
				}
			}
		}
	}

	return mentionedUserIDs, nil
}

func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// GetMentionedUsers gets all mentioned users for a comment
func (r *PoliticianCommentRepository) GetMentionedUsers(ctx context.Context, commentID uuid.UUID) ([]models.CommentAuthor, error) {
	query := `
		SELECT u.id, u.name, u.avatar, COALESCE(u.is_system, false)
		FROM politician_comment_mentions m
		JOIN users u ON m.mentioned_user_id = u.id
		WHERE m.comment_id = $1
	`

	rows, err := r.db.Query(ctx, query, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.CommentAuthor
	for rows.Next() {
		var user models.CommentAuthor
		if err := rows.Scan(&user.ID, &user.Name, &user.Avatar, &user.IsSystem); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
