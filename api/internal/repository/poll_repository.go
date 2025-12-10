package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/humfurie/pulpulitiko/api/internal/models"
)

type PollRepository struct {
	db *pgxpool.Pool
}

func NewPollRepository(db *pgxpool.Pool) *PollRepository {
	return &PollRepository{db: db}
}

// Polls

func (r *PollRepository) CreatePoll(ctx context.Context, userID uuid.UUID, req *models.CreatePollRequest) (*models.Poll, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var startsAt, endsAt *time.Time
	if req.StartsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartsAt)
		if err == nil {
			startsAt = &t
		}
	}
	if req.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err == nil {
			endsAt = &t
		}
	}

	// Determine initial status
	status := models.PollStatusDraft

	var poll models.Poll
	err = tx.QueryRow(ctx, `
		INSERT INTO polls (
			user_id, title, slug, description, category, status,
			politician_id, election_id, bill_id,
			region_id, province_id, city_municipality_id, barangay_id,
			is_anonymous, allow_multiple_votes, show_results_before_vote,
			starts_at, ends_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, user_id, title, slug, description, category, status,
			politician_id, election_id, bill_id,
			region_id, province_id, city_municipality_id, barangay_id,
			is_anonymous, allow_multiple_votes, show_results_before_vote,
			is_featured, starts_at, ends_at,
			total_votes, view_count, comment_count,
			created_at, updated_at
	`, userID, req.Title, req.Slug, req.Description, req.Category, status,
		req.PoliticianID, req.ElectionID, req.BillID,
		req.RegionID, req.ProvinceID, req.CityMunicipalityID, req.BarangayID,
		req.IsAnonymous, req.AllowMultipleVotes, req.ShowResultsBeforeVote,
		startsAt, endsAt,
	).Scan(
		&poll.ID, &poll.UserID, &poll.Title, &poll.Slug, &poll.Description,
		&poll.Category, &poll.Status, &poll.PoliticianID, &poll.ElectionID, &poll.BillID,
		&poll.RegionID, &poll.ProvinceID, &poll.CityMunicipalityID, &poll.BarangayID,
		&poll.IsAnonymous, &poll.AllowMultipleVotes, &poll.ShowResultsBeforeVote,
		&poll.IsFeatured, &poll.StartsAt, &poll.EndsAt,
		&poll.TotalVotes, &poll.ViewCount, &poll.CommentCount,
		&poll.CreatedAt, &poll.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Insert options
	for i, optionText := range req.Options {
		var option models.PollOption
		err = tx.QueryRow(ctx, `
			INSERT INTO poll_options (poll_id, text, display_order)
			VALUES ($1, $2, $3)
			RETURNING id, poll_id, text, display_order, vote_count, created_at
		`, poll.ID, optionText, i+1).Scan(
			&option.ID, &option.PollID, &option.Text, &option.DisplayOrder,
			&option.VoteCount, &option.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		poll.Options = append(poll.Options, option)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &poll, nil
}

func (r *PollRepository) GetPollByID(ctx context.Context, id uuid.UUID) (*models.Poll, error) {
	var poll models.Poll
	var authorID uuid.UUID
	var authorName string
	var authorAvatar *string

	err := r.db.QueryRow(ctx, `
		SELECT p.id, p.user_id, p.title, p.slug, p.description, p.category, p.status,
			p.politician_id, p.election_id, p.bill_id,
			p.region_id, p.province_id, p.city_municipality_id, p.barangay_id,
			p.is_anonymous, p.allow_multiple_votes, p.show_results_before_vote,
			p.is_featured, p.starts_at, p.ends_at,
			p.approved_by, p.approved_at, p.rejection_reason,
			p.total_votes, p.view_count, p.comment_count,
			p.created_at, p.updated_at,
			u.id, u.name, u.avatar
		FROM polls p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`, id).Scan(
		&poll.ID, &poll.UserID, &poll.Title, &poll.Slug, &poll.Description,
		&poll.Category, &poll.Status, &poll.PoliticianID, &poll.ElectionID, &poll.BillID,
		&poll.RegionID, &poll.ProvinceID, &poll.CityMunicipalityID, &poll.BarangayID,
		&poll.IsAnonymous, &poll.AllowMultipleVotes, &poll.ShowResultsBeforeVote,
		&poll.IsFeatured, &poll.StartsAt, &poll.EndsAt,
		&poll.ApprovedBy, &poll.ApprovedAt, &poll.RejectionReason,
		&poll.TotalVotes, &poll.ViewCount, &poll.CommentCount,
		&poll.CreatedAt, &poll.UpdatedAt,
		&authorID, &authorName, &authorAvatar,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	poll.Author = &models.PollAuthor{
		ID:     authorID,
		Name:   authorName,
		Avatar: authorAvatar,
	}

	// Get options
	options, err := r.GetPollOptions(ctx, poll.ID)
	if err != nil {
		return nil, err
	}
	poll.Options = options

	// Get associated entities if present
	if poll.PoliticianID != nil {
		poll.Politician, _ = r.getPoliticianBrief(ctx, *poll.PoliticianID)
	}
	if poll.ElectionID != nil {
		poll.Election, _ = r.getElectionBrief(ctx, *poll.ElectionID)
	}
	if poll.BillID != nil {
		poll.Bill, _ = r.getBillBrief(ctx, *poll.BillID)
	}

	// Get location info if present
	if poll.RegionID != nil || poll.ProvinceID != nil || poll.CityMunicipalityID != nil || poll.BarangayID != nil {
		poll.Location, _ = r.getLocationBrief(ctx, poll.RegionID, poll.ProvinceID, poll.CityMunicipalityID, poll.BarangayID)
	}

	return &poll, nil
}

func (r *PollRepository) GetPollBySlug(ctx context.Context, slug string) (*models.Poll, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `
		SELECT id FROM polls WHERE slug = $1 AND deleted_at IS NULL
	`, slug).Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.GetPollByID(ctx, id)
}

func (r *PollRepository) ListPolls(ctx context.Context, filter *models.PollFilter, page, perPage int) (*models.PaginatedPolls, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, "p.deleted_at IS NULL")

	if filter != nil {
		if filter.Category != nil {
			conditions = append(conditions, fmt.Sprintf("p.category = $%d", argNum))
			args = append(args, *filter.Category)
			argNum++
		}
		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("p.status = $%d", argNum))
			args = append(args, *filter.Status)
			argNum++
		}
		if filter.UserID != nil {
			conditions = append(conditions, fmt.Sprintf("p.user_id = $%d", argNum))
			args = append(args, *filter.UserID)
			argNum++
		}
		if filter.PoliticianID != nil {
			conditions = append(conditions, fmt.Sprintf("p.politician_id = $%d", argNum))
			args = append(args, *filter.PoliticianID)
			argNum++
		}
		if filter.ElectionID != nil {
			conditions = append(conditions, fmt.Sprintf("p.election_id = $%d", argNum))
			args = append(args, *filter.ElectionID)
			argNum++
		}
		if filter.IsFeatured != nil {
			conditions = append(conditions, fmt.Sprintf("p.is_featured = $%d", argNum))
			args = append(args, *filter.IsFeatured)
			argNum++
		}
		if filter.Search != nil && *filter.Search != "" {
			conditions = append(conditions, fmt.Sprintf("(p.title ILIKE $%d OR p.description ILIKE $%d)", argNum, argNum))
			args = append(args, "%"+*filter.Search+"%")
			argNum++
		}
		if filter.ActiveOnly {
			conditions = append(conditions, "p.status = 'active'")
			conditions = append(conditions, "(p.starts_at IS NULL OR p.starts_at <= NOW())")
			conditions = append(conditions, "(p.ends_at IS NULL OR p.ends_at > NOW())")
		}

		// Location filtering
		if filter.RegionID != nil || filter.ProvinceID != nil || filter.CityMunicipalityID != nil || filter.BarangayID != nil {
			var locationConditions []string

			if filter.BarangayID != nil {
				locationConditions = append(locationConditions, fmt.Sprintf("p.barangay_id = $%d", argNum))
				args = append(args, *filter.BarangayID)
				argNum++
			} else if filter.CityMunicipalityID != nil {
				locationConditions = append(locationConditions, fmt.Sprintf("p.city_municipality_id = $%d", argNum))
				args = append(args, *filter.CityMunicipalityID)
				argNum++
			} else if filter.ProvinceID != nil {
				locationConditions = append(locationConditions, fmt.Sprintf("p.province_id = $%d", argNum))
				args = append(args, *filter.ProvinceID)
				argNum++
			} else if filter.RegionID != nil {
				locationConditions = append(locationConditions, fmt.Sprintf("p.region_id = $%d", argNum))
				args = append(args, *filter.RegionID)
				argNum++
			}

			// Optionally include national polls (polls with no location set)
			if filter.IncludeNational {
				locationConditions = append(locationConditions, "(p.region_id IS NULL AND p.province_id IS NULL AND p.city_municipality_id IS NULL AND p.barangay_id IS NULL)")
				conditions = append(conditions, "("+strings.Join(locationConditions, " OR ")+")")
			} else {
				conditions = append(conditions, strings.Join(locationConditions, " AND "))
			}
		}
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM polls p WHERE %s`, whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalPages := (total + perPage - 1) / perPage
	offset := (page - 1) * perPage

	// Fetch polls with location display name
	args = append(args, perPage, offset)
	query := fmt.Sprintf(`
		SELECT p.id, p.title, p.slug, p.category, p.status, p.is_featured,
			p.total_votes, p.comment_count, p.ends_at, p.created_at,
			u.id, u.name, u.avatar,
			(SELECT COUNT(*) FROM poll_options WHERE poll_id = p.id) as option_count,
			COALESCE(
				b.name || ', ' || cm.name || ', ' || prov.name,
				cm.name || ', ' || prov.name,
				prov.name || ', ' || r.name,
				r.name,
				NULL
			) as location_display
		FROM polls p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN regions r ON p.region_id = r.id
		LEFT JOIN provinces prov ON p.province_id = prov.id
		LEFT JOIN cities_municipalities cm ON p.city_municipality_id = cm.id
		LEFT JOIN barangays b ON p.barangay_id = b.id
		WHERE %s
		ORDER BY p.is_featured DESC, p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var polls []models.PollListItem
	for rows.Next() {
		var poll models.PollListItem
		var authorID uuid.UUID
		var authorName string
		var authorAvatar *string

		err := rows.Scan(
			&poll.ID, &poll.Title, &poll.Slug, &poll.Category, &poll.Status,
			&poll.IsFeatured, &poll.TotalVotes, &poll.CommentCount, &poll.EndsAt,
			&poll.CreatedAt, &authorID, &authorName, &authorAvatar, &poll.OptionCount,
			&poll.Location,
		)
		if err != nil {
			return nil, err
		}

		poll.Author = &models.PollAuthor{
			ID:     authorID,
			Name:   authorName,
			Avatar: authorAvatar,
		}

		polls = append(polls, poll)
	}

	return &models.PaginatedPolls{
		Polls:      polls,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *PollRepository) GetFeaturedPolls(ctx context.Context, limit int) ([]models.PollListItem, error) {
	filter := &models.PollFilter{
		IsFeatured: boolPtr(true),
		ActiveOnly: true,
	}
	result, err := r.ListPolls(ctx, filter, 1, limit)
	if err != nil {
		return nil, err
	}
	return result.Polls, nil
}

func (r *PollRepository) UpdatePoll(ctx context.Context, id uuid.UUID, req *models.UpdatePollRequest) (*models.Poll, error) {
	var sets []string
	var args []interface{}
	argNum := 1

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argNum))
		args = append(args, *req.Title)
		argNum++
	}
	if req.Slug != nil {
		sets = append(sets, fmt.Sprintf("slug = $%d", argNum))
		args = append(args, *req.Slug)
		argNum++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argNum))
		args = append(args, *req.Description)
		argNum++
	}
	if req.Category != nil {
		sets = append(sets, fmt.Sprintf("category = $%d", argNum))
		args = append(args, *req.Category)
		argNum++
	}
	if req.IsAnonymous != nil {
		sets = append(sets, fmt.Sprintf("is_anonymous = $%d", argNum))
		args = append(args, *req.IsAnonymous)
		argNum++
	}
	if req.AllowMultipleVotes != nil {
		sets = append(sets, fmt.Sprintf("allow_multiple_votes = $%d", argNum))
		args = append(args, *req.AllowMultipleVotes)
		argNum++
	}
	if req.ShowResultsBeforeVote != nil {
		sets = append(sets, fmt.Sprintf("show_results_before_vote = $%d", argNum))
		args = append(args, *req.ShowResultsBeforeVote)
		argNum++
	}
	if req.StartsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartsAt)
		if err == nil {
			sets = append(sets, fmt.Sprintf("starts_at = $%d", argNum))
			args = append(args, t)
			argNum++
		}
	}
	if req.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err == nil {
			sets = append(sets, fmt.Sprintf("ends_at = $%d", argNum))
			args = append(args, t)
			argNum++
		}
	}

	if len(sets) == 0 {
		return r.GetPollByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE polls SET %s, updated_at = NOW()
		WHERE id = $%d AND deleted_at IS NULL
	`, strings.Join(sets, ", "), argNum)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return r.GetPollByID(ctx, id)
}

func (r *PollRepository) AdminUpdatePoll(ctx context.Context, id uuid.UUID, req *models.AdminUpdatePollRequest) (*models.Poll, error) {
	// First apply the basic update
	_, err := r.UpdatePoll(ctx, id, &req.UpdatePollRequest)
	if err != nil {
		return nil, err
	}

	// Then apply admin-specific updates
	var sets []string
	var args []interface{}
	argNum := 1

	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *req.Status)
		argNum++
	}
	if req.IsFeatured != nil {
		sets = append(sets, fmt.Sprintf("is_featured = $%d", argNum))
		args = append(args, *req.IsFeatured)
		argNum++
	}

	if len(sets) > 0 {
		args = append(args, id)
		query := fmt.Sprintf(`
			UPDATE polls SET %s, updated_at = NOW()
			WHERE id = $%d AND deleted_at IS NULL
		`, strings.Join(sets, ", "), argNum)

		_, err = r.db.Exec(ctx, query, args...)
		if err != nil {
			return nil, err
		}
	}

	return r.GetPollByID(ctx, id)
}

func (r *PollRepository) ApprovePoll(ctx context.Context, id uuid.UUID, approverID uuid.UUID, approved bool, reason *string) error {
	var status string
	if approved {
		status = models.PollStatusActive
	} else {
		status = models.PollStatusRejected
	}

	_, err := r.db.Exec(ctx, `
		UPDATE polls SET
			status = $1,
			approved_by = $2,
			approved_at = NOW(),
			rejection_reason = $3,
			updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`, status, approverID, reason, id)
	return err
}

func (r *PollRepository) ClosePoll(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE polls SET status = 'closed', updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	return err
}

func (r *PollRepository) DeletePoll(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE polls SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	return err
}

func (r *PollRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE polls SET view_count = view_count + 1
		WHERE id = $1
	`, id)
	return err
}

// Poll Options

func (r *PollRepository) GetPollOptions(ctx context.Context, pollID uuid.UUID) ([]models.PollOption, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, poll_id, text, display_order, vote_count, created_at
		FROM poll_options
		WHERE poll_id = $1
		ORDER BY display_order
	`, pollID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []models.PollOption
	for rows.Next() {
		var opt models.PollOption
		err := rows.Scan(&opt.ID, &opt.PollID, &opt.Text, &opt.DisplayOrder, &opt.VoteCount, &opt.CreatedAt)
		if err != nil {
			return nil, err
		}
		options = append(options, opt)
	}

	return options, nil
}

// Voting

func (r *PollRepository) CastVote(ctx context.Context, pollID, optionID uuid.UUID, userID *uuid.UUID, ipHash *string) error {
	var existingVote uuid.UUID

	// Check for existing vote
	if userID != nil {
		err := r.db.QueryRow(ctx, `
			SELECT id FROM poll_votes WHERE poll_id = $1 AND user_id = $2
		`, pollID, userID).Scan(&existingVote)
		if err == nil {
			return fmt.Errorf("you have already voted on this poll")
		}
		if err != pgx.ErrNoRows {
			return err
		}
	} else if ipHash != nil {
		err := r.db.QueryRow(ctx, `
			SELECT id FROM poll_votes WHERE poll_id = $1 AND ip_hash = $2
		`, pollID, ipHash).Scan(&existingVote)
		if err == nil {
			return fmt.Errorf("you have already voted on this poll")
		}
		if err != pgx.ErrNoRows {
			return err
		}
	}

	// Cast vote
	_, err := r.db.Exec(ctx, `
		INSERT INTO poll_votes (poll_id, option_id, user_id, ip_hash)
		VALUES ($1, $2, $3, $4)
	`, pollID, optionID, userID, ipHash)
	return err
}

func (r *PollRepository) HasUserVoted(ctx context.Context, pollID uuid.UUID, userID *uuid.UUID, ipHash *string) (bool, *uuid.UUID) {
	var optionID uuid.UUID

	if userID != nil {
		err := r.db.QueryRow(ctx, `
			SELECT option_id FROM poll_votes WHERE poll_id = $1 AND user_id = $2
		`, pollID, userID).Scan(&optionID)
		if err == nil {
			return true, &optionID
		}
	}

	if ipHash != nil {
		err := r.db.QueryRow(ctx, `
			SELECT option_id FROM poll_votes WHERE poll_id = $1 AND ip_hash = $2
		`, pollID, ipHash).Scan(&optionID)
		if err == nil {
			return true, &optionID
		}
	}

	return false, nil
}

func (r *PollRepository) GetPollResults(ctx context.Context, pollID uuid.UUID) (*models.PollResults, error) {
	var totalVotes int
	err := r.db.QueryRow(ctx, `
		SELECT total_votes FROM polls WHERE id = $1
	`, pollID).Scan(&totalVotes)
	if err != nil {
		return nil, err
	}

	options, err := r.GetPollOptions(ctx, pollID)
	if err != nil {
		return nil, err
	}

	// Calculate percentages
	for i := range options {
		if totalVotes > 0 {
			options[i].Percentage = float64(options[i].VoteCount) / float64(totalVotes) * 100
		}
	}

	return &models.PollResults{
		PollID:     pollID,
		TotalVotes: totalVotes,
		Options:    options,
	}, nil
}

// Poll Comments

func (r *PollRepository) CreatePollComment(ctx context.Context, pollID, userID uuid.UUID, req *models.CreatePollCommentRequest) (*models.PollComment, error) {
	var comment models.PollComment
	err := r.db.QueryRow(ctx, `
		INSERT INTO poll_comments (poll_id, user_id, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, poll_id, user_id, parent_id, content, status, created_at, updated_at
	`, pollID, userID, req.ParentID, req.Content).Scan(
		&comment.ID, &comment.PollID, &comment.UserID, &comment.ParentID,
		&comment.Content, &comment.Status, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Get author info
	var author models.CommentAuthor
	err = r.db.QueryRow(ctx, `
		SELECT id, name, avatar FROM users WHERE id = $1
	`, userID).Scan(&author.ID, &author.Name, &author.Avatar)
	if err == nil {
		comment.Author = &author
	}

	return &comment, nil
}

func (r *PollRepository) GetPollComments(ctx context.Context, pollID uuid.UUID, page, perPage int) (*models.PaginatedPollComments, error) {
	// Count total
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM poll_comments
		WHERE poll_id = $1 AND parent_id IS NULL AND deleted_at IS NULL
	`, pollID).Scan(&total)
	if err != nil {
		return nil, err
	}

	totalPages := (total + perPage - 1) / perPage
	offset := (page - 1) * perPage

	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.poll_id, c.user_id, c.parent_id, c.content, c.status,
			c.created_at, c.updated_at,
			u.id, u.name, u.avatar,
			(SELECT COUNT(*) FROM poll_comments WHERE parent_id = c.id AND deleted_at IS NULL) as reply_count
		FROM poll_comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.poll_id = $1 AND c.parent_id IS NULL AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`, pollID, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.PollComment
	for rows.Next() {
		var comment models.PollComment
		var author models.CommentAuthor

		err := rows.Scan(
			&comment.ID, &comment.PollID, &comment.UserID, &comment.ParentID,
			&comment.Content, &comment.Status, &comment.CreatedAt, &comment.UpdatedAt,
			&author.ID, &author.Name, &author.Avatar, &comment.ReplyCount,
		)
		if err != nil {
			return nil, err
		}

		comment.Author = &author
		comments = append(comments, comment)
	}

	return &models.PaginatedPollComments{
		Comments:   comments,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *PollRepository) DeletePollComment(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE poll_comments SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	return err
}

// Helper methods

func (r *PollRepository) getPoliticianBrief(ctx context.Context, id uuid.UUID) (*models.PoliticianBrief, error) {
	var p models.PoliticianBrief
	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, photo FROM politicians WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&p.ID, &p.Name, &p.Slug, &p.Photo)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PollRepository) getElectionBrief(ctx context.Context, id uuid.UUID) (*models.ElectionBrief, error) {
	var e models.ElectionBrief
	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug FROM elections WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&e.ID, &e.Name, &e.Slug)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *PollRepository) getBillBrief(ctx context.Context, id uuid.UUID) (*models.BillBrief, error) {
	var b models.BillBrief
	err := r.db.QueryRow(ctx, `
		SELECT id, bill_number, title, slug FROM bills WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&b.ID, &b.BillNumber, &b.Title, &b.Slug)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func boolPtr(b bool) *bool {
	return &b
}

func (r *PollRepository) getLocationBrief(ctx context.Context, regionID, provinceID, cityMunicipalityID, barangayID *uuid.UUID) (*models.LocationBrief, error) {
	loc := &models.LocationBrief{}
	var displayParts []string

	// Get barangay info
	if barangayID != nil {
		var name string
		err := r.db.QueryRow(ctx, `SELECT id, name FROM barangays WHERE id = $1`, barangayID).Scan(&loc.BarangayID, &name)
		if err == nil {
			loc.BarangayName = &name
			displayParts = append(displayParts, name)
		}
	}

	// Get city/municipality info
	if cityMunicipalityID != nil {
		var name string
		err := r.db.QueryRow(ctx, `SELECT id, name FROM cities_municipalities WHERE id = $1`, cityMunicipalityID).Scan(&loc.CityMunicipalityID, &name)
		if err == nil {
			loc.CityMunicipalityName = &name
			displayParts = append(displayParts, name)
		}
	}

	// Get province info
	if provinceID != nil {
		var name string
		err := r.db.QueryRow(ctx, `SELECT id, name FROM provinces WHERE id = $1`, provinceID).Scan(&loc.ProvinceID, &name)
		if err == nil {
			loc.ProvinceName = &name
			displayParts = append(displayParts, name)
		}
	}

	// Get region info
	if regionID != nil {
		var name string
		err := r.db.QueryRow(ctx, `SELECT id, name FROM regions WHERE id = $1`, regionID).Scan(&loc.RegionID, &name)
		if err == nil {
			loc.RegionName = &name
			displayParts = append(displayParts, name)
		}
	}

	loc.DisplayName = strings.Join(displayParts, ", ")
	return loc, nil
}
