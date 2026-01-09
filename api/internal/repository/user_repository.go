package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, name, role_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.RoleID,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.name, COALESCE(a.avatar, u.avatar) as avatar,
		       u.role_id, COALESCE(r.slug, '') as role_slug, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		LEFT JOIN authors a ON a.email = u.email AND a.deleted_at IS NULL
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Avatar,
		&user.RoleID, &user.RoleSlug, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.name, COALESCE(a.avatar, u.avatar) as avatar,
		       u.role_id, COALESCE(r.slug, '') as role_slug, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		LEFT JOIN authors a ON LOWER(a.email) = LOWER(u.email) AND a.deleted_at IS NULL
		WHERE LOWER(u.email) = LOWER($1) AND u.deleted_at IS NULL
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Avatar,
		&user.RoleID, &user.RoleSlug, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.name, COALESCE(a.avatar, u.avatar) as avatar,
		       u.role_id, COALESCE(r.slug, '') as role_slug, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		LEFT JOIN authors a ON a.email = u.email AND a.deleted_at IS NULL
		WHERE u.deleted_at IS NULL
		ORDER BY u.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Avatar,
			&user.RoleID, &user.RoleSlug, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) AdminList(ctx context.Context, filter *models.UserFilter, page, perPage int) (*models.PaginatedUsers, error) {
	baseQuery := `
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		LEFT JOIN authors a ON a.email = u.email AND a.deleted_at IS NULL
		WHERE u.deleted_at IS NULL`

	args := []interface{}{}
	argCount := 0

	if filter.Search != nil && *filter.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (u.name ILIKE $%d OR u.email ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+*filter.Search+"%")
	}

	if filter.RoleSlug != nil && *filter.RoleSlug != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND r.slug = $%d", argCount)
		args = append(args, *filter.RoleSlug)
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	orderClause := "ORDER BY u.created_at DESC"
	if filter.SortBy != nil && *filter.SortBy != "" {
		sortBy := *filter.SortBy
		sortOrder := "ASC"
		if filter.SortOrder != nil && *filter.SortOrder != "" {
			sortOrder = *filter.SortOrder
		}
		if sortBy == "name" {
			orderClause = fmt.Sprintf("ORDER BY u.name %s", sortOrder)
		} else if sortBy == "email" {
			orderClause = fmt.Sprintf("ORDER BY u.email %s", sortOrder)
		} else if sortBy == "created_at" {
			orderClause = fmt.Sprintf("ORDER BY u.created_at %s", sortOrder)
		}
	}

	offset := (page - 1) * perPage
	totalPages := (total + perPage - 1) / perPage

	argCount++
	query := fmt.Sprintf(`
		SELECT u.id, u.email, u.password_hash, u.name, COALESCE(a.avatar, u.avatar) as avatar,
		       u.role_id, COALESCE(r.slug, '') as role_slug, u.created_at, u.updated_at, u.deleted_at
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, baseQuery, orderClause, argCount, argCount+1)
	args = append(args, perPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Avatar, &user.RoleID, &user.RoleSlug, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return &models.PaginatedUsers{
		Users:      users,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE users SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or not deleted")
	}

	return nil
}

func (r *UserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetMentionableUsers returns users that can be mentioned in comments
func (r *UserRepository) GetMentionableUsers(ctx context.Context) ([]models.CommentAuthor, error) {
	query := `
		SELECT id, name, avatar
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY name ASC
		LIMIT 100
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentionable users: %w", err)
	}
	defer rows.Close()

	var users []models.CommentAuthor
	for rows.Next() {
		var user models.CommentAuthor
		if err := rows.Scan(&user.ID, &user.Name, &user.Avatar); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserBySlug retrieves a user by their name slug
func (r *UserRepository) GetUserBySlug(ctx context.Context, slug string) (*models.User, error) {
	// Convert slug back to name pattern (replace hyphens with spaces for ILIKE)
	namePattern := "%" + slug + "%"

	query := `
		SELECT u.id, u.email, u.password_hash, u.name, u.role_id, COALESCE(r.slug, '') as role_slug, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.deleted_at IS NULL AND LOWER(REPLACE(u.name, ' ', '-')) = LOWER($1)
		LIMIT 1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.RoleID, &user.RoleSlug,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		// Try alternative lookup with name pattern
		query = `
			SELECT u.id, u.email, u.password_hash, u.name, u.role_id, COALESCE(r.slug, '') as role_slug, u.created_at, u.updated_at, u.deleted_at
			FROM users u
			LEFT JOIN roles r ON u.role_id = r.id
			WHERE u.deleted_at IS NULL AND u.name ILIKE $1
			LIMIT 1
		`
		err = r.db.QueryRow(ctx, query, namePattern).Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.RoleID, &user.RoleSlug,
			&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		)
		if err == pgx.ErrNoRows {
			return nil, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by slug: %w", err)
	}

	return user, nil
}

// GetUserProfile returns a user's public profile with comment counts
func (r *UserRepository) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	query := `
		SELECT
			u.id, u.name, COALESCE(a.avatar, u.avatar) as avatar, u.created_at,
			(SELECT COUNT(*) FROM comments WHERE user_id = u.id AND parent_id IS NULL AND deleted_at IS NULL) as comment_count,
			(SELECT COUNT(*) FROM comments WHERE user_id = u.id AND parent_id IS NOT NULL AND deleted_at IS NULL) as reply_count
		FROM users u
		LEFT JOIN authors a ON a.email = u.email AND a.deleted_at IS NULL
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`

	profile := &models.UserProfile{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&profile.ID, &profile.Name, &profile.Avatar, &profile.CreatedAt,
		&profile.CommentCount, &profile.ReplyCount,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Generate slug from name
	profile.Slug = strings.ToLower(strings.ReplaceAll(profile.Name, " ", "-"))

	return profile, nil
}

// GetUserComments returns comments made by a user (not replies)
func (r *UserRepository) GetUserComments(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Comment, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT c.id, c.article_id, c.user_id, c.parent_id, c.content, c.status,
		       c.created_at, c.updated_at,
		       a.slug as article_slug,
		       (SELECT COUNT(*) FROM comments r WHERE r.parent_id = c.id AND r.deleted_at IS NULL) as reply_count
		FROM comments c
		JOIN articles a ON c.article_id = a.id
		WHERE c.user_id = $1 AND c.parent_id IS NULL AND c.deleted_at IS NULL AND c.status = 'active'
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(
			&comment.ID, &comment.ArticleID, &comment.UserID, &comment.ParentID,
			&comment.Content, &comment.Status, &comment.CreatedAt, &comment.UpdatedAt,
			&comment.ArticleSlug, &comment.ReplyCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// GetUserReplies returns replies made by a user
func (r *UserRepository) GetUserReplies(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Comment, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT c.id, c.article_id, c.user_id, c.parent_id, c.content, c.status,
		       c.created_at, c.updated_at,
		       a.slug as article_slug
		FROM comments c
		JOIN articles a ON c.article_id = a.id
		WHERE c.user_id = $1 AND c.parent_id IS NOT NULL AND c.deleted_at IS NULL AND c.status = 'active'
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user replies: %w", err)
	}
	defer rows.Close()

	var replies []models.Comment
	for rows.Next() {
		var reply models.Comment
		if err := rows.Scan(
			&reply.ID, &reply.ArticleID, &reply.UserID, &reply.ParentID,
			&reply.Content, &reply.Status, &reply.CreatedAt, &reply.UpdatedAt,
			&reply.ArticleSlug,
		); err != nil {
			return nil, fmt.Errorf("failed to scan reply: %w", err)
		}
		replies = append(replies, reply)
	}

	return replies, nil
}

// CreatePasswordResetToken creates a new password reset token for a user
func (r *UserRepository) CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	return nil
}

// GetPasswordResetToken retrieves a valid (unexpired, unused) password reset token
func (r *UserRepository) GetPasswordResetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1 AND expires_at > NOW() AND used_at IS NULL
	`

	resetToken := &models.PasswordResetToken{}
	err := r.db.QueryRow(ctx, query, token).Scan(
		&resetToken.ID, &resetToken.UserID, &resetToken.Token,
		&resetToken.ExpiresAt, &resetToken.UsedAt, &resetToken.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}

	return resetToken, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used
func (r *UserRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`

	_, err := r.db.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// UpdatePassword updates a user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// InvalidateUserPasswordResetTokens invalidates all existing password reset tokens for a user
func (r *UserRepository) InvalidateUserPasswordResetTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate tokens: %w", err)
	}

	return nil
}
