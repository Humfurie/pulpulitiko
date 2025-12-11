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

type BillRepository struct {
	db *pgxpool.Pool
}

func NewBillRepository(db *pgxpool.Pool) *BillRepository {
	return &BillRepository{db: db}
}

// Legislative Sessions

func (r *BillRepository) GetCurrentSession(ctx context.Context) (*models.LegislativeSession, error) {
	session := &models.LegislativeSession{}
	err := r.db.QueryRow(ctx, `
		SELECT id, congress_number, session_number, session_type, start_date, end_date, is_current, created_at, updated_at
		FROM legislative_sessions
		WHERE is_current = TRUE
		LIMIT 1
	`).Scan(
		&session.ID, &session.CongressNumber, &session.SessionNumber, &session.SessionType,
		&session.StartDate, &session.EndDate, &session.IsCurrent, &session.CreatedAt, &session.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current session: %w", err)
	}
	return session, nil
}

func (r *BillRepository) ListSessions(ctx context.Context) ([]models.LegislativeSessionListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ls.id, ls.congress_number, ls.session_number, ls.session_type, ls.is_current,
		       COALESCE((SELECT COUNT(*) FROM bills WHERE session_id = ls.id AND deleted_at IS NULL), 0) as bill_count
		FROM legislative_sessions ls
		ORDER BY ls.congress_number DESC, ls.session_number DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.LegislativeSessionListItem
	for rows.Next() {
		var s models.LegislativeSessionListItem
		err := rows.Scan(&s.ID, &s.CongressNumber, &s.SessionNumber, &s.SessionType, &s.IsCurrent, &s.BillCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// Committees

func (r *BillRepository) ListCommittees(ctx context.Context, chamber *string) ([]models.CommitteeListItem, error) {
	query := `
		SELECT c.id, c.chamber, c.name, c.slug, c.is_active,
		       COALESCE((SELECT COUNT(*) FROM bill_committees WHERE committee_id = c.id), 0) as bill_count
		FROM committees c
		WHERE c.deleted_at IS NULL
	`
	args := []interface{}{}
	if chamber != nil {
		query += " AND c.chamber = $1"
		args = append(args, *chamber)
	}
	query += " ORDER BY c.chamber, c.name"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list committees: %w", err)
	}
	defer rows.Close()

	var committees []models.CommitteeListItem
	for rows.Next() {
		var c models.CommitteeListItem
		err := rows.Scan(&c.ID, &c.Chamber, &c.Name, &c.Slug, &c.IsActive, &c.BillCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan committee: %w", err)
		}
		committees = append(committees, c)
	}
	return committees, nil
}

func (r *BillRepository) GetCommitteeBySlug(ctx context.Context, slug string) (*models.Committee, error) {
	committee := &models.Committee{}
	err := r.db.QueryRow(ctx, `
		SELECT id, chamber, name, slug, description, chairperson_id, vice_chairperson_id, is_active, created_at, updated_at
		FROM committees
		WHERE slug = $1 AND deleted_at IS NULL
	`, slug).Scan(
		&committee.ID, &committee.Chamber, &committee.Name, &committee.Slug, &committee.Description,
		&committee.ChairpersonID, &committee.ViceChairpersonID, &committee.IsActive, &committee.CreatedAt, &committee.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get committee: %w", err)
	}
	return committee, nil
}

// Bills

func (r *BillRepository) Create(ctx context.Context, req *models.CreateBillRequest) (*models.Bill, error) {
	filedDate, err := time.Parse("2006-01-02", req.FiledDate)
	if err != nil {
		return nil, fmt.Errorf("invalid filed_date format: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	bill := &models.Bill{}
	err = tx.QueryRow(ctx, `
		INSERT INTO bills (session_id, chamber, bill_number, title, slug, short_title, summary, full_text, significance, status, filed_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, session_id, chamber, bill_number, title, slug, short_title, summary, full_text, significance, status, filed_date, created_at, updated_at
	`, req.SessionID, req.Chamber, req.BillNumber, req.Title, req.Slug, req.ShortTitle, req.Summary, req.FullText, req.Significance, req.Status, filedDate).Scan(
		&bill.ID, &bill.SessionID, &bill.Chamber, &bill.BillNumber, &bill.Title, &bill.Slug,
		&bill.ShortTitle, &bill.Summary, &bill.FullText, &bill.Significance, &bill.Status, &bill.FiledDate,
		&bill.CreatedAt, &bill.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}

	// Add principal authors
	for _, authorID := range req.PrincipalAuthors {
		_, err = tx.Exec(ctx, `
			INSERT INTO bill_authors (bill_id, politician_id, is_principal_author) VALUES ($1, $2, TRUE)
		`, bill.ID, authorID)
		if err != nil {
			return nil, fmt.Errorf("failed to add principal author: %w", err)
		}
	}

	// Add co-authors
	for _, authorID := range req.CoAuthors {
		_, err = tx.Exec(ctx, `
			INSERT INTO bill_authors (bill_id, politician_id, is_principal_author) VALUES ($1, $2, FALSE)
		`, bill.ID, authorID)
		if err != nil {
			return nil, fmt.Errorf("failed to add co-author: %w", err)
		}
	}

	// Add topics
	for _, topicID := range req.TopicIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO bill_topic_assignments (bill_id, topic_id) VALUES ($1, $2)
		`, bill.ID, topicID)
		if err != nil {
			return nil, fmt.Errorf("failed to add topic: %w", err)
		}
	}

	// Add initial status history
	_, err = tx.Exec(ctx, `
		INSERT INTO bill_status_history (bill_id, status, action_date, action_description)
		VALUES ($1, $2, $3, 'Bill filed')
	`, bill.ID, req.Status, filedDate)
	if err != nil {
		return nil, fmt.Errorf("failed to add status history: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return bill, nil
}

func (r *BillRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Bill, error) {
	bill := &models.Bill{}
	err := r.db.QueryRow(ctx, `
		SELECT id, session_id, chamber, bill_number, title, slug, short_title, summary, full_text, significance,
		       status, filed_date, last_action_date, date_signed, republic_act_number, created_at, updated_at
		FROM bills
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&bill.ID, &bill.SessionID, &bill.Chamber, &bill.BillNumber, &bill.Title, &bill.Slug, &bill.ShortTitle,
		&bill.Summary, &bill.FullText, &bill.Significance, &bill.Status, &bill.FiledDate, &bill.LastActionDate,
		&bill.DateSigned, &bill.RepublicActNumber, &bill.CreatedAt, &bill.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bill: %w", err)
	}
	return bill, nil
}

func (r *BillRepository) GetBySlug(ctx context.Context, slug string) (*models.Bill, error) {
	bill := &models.Bill{}
	err := r.db.QueryRow(ctx, `
		SELECT id, session_id, chamber, bill_number, title, slug, short_title, summary, full_text, significance,
		       status, filed_date, last_action_date, date_signed, republic_act_number, created_at, updated_at
		FROM bills
		WHERE slug = $1 AND deleted_at IS NULL
	`, slug).Scan(
		&bill.ID, &bill.SessionID, &bill.Chamber, &bill.BillNumber, &bill.Title, &bill.Slug, &bill.ShortTitle,
		&bill.Summary, &bill.FullText, &bill.Significance, &bill.Status, &bill.FiledDate, &bill.LastActionDate,
		&bill.DateSigned, &bill.RepublicActNumber, &bill.CreatedAt, &bill.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bill: %w", err)
	}

	// Load related data
	bill.Authors, _ = r.GetBillAuthors(ctx, bill.ID)
	bill.StatusHistory, _ = r.GetBillStatusHistory(ctx, bill.ID)
	bill.Topics, _ = r.GetBillTopics(ctx, bill.ID)
	bill.Committees, _ = r.GetBillCommittees(ctx, bill.ID)
	bill.Votes, _ = r.GetBillVotes(ctx, bill.ID)

	// Load principal authors separately for easy access
	for _, author := range bill.Authors {
		if author.IsPrincipalAuthor && author.Politician != nil {
			bill.PrincipalAuthors = append(bill.PrincipalAuthors, *author.Politician)
		}
	}

	return bill, nil
}

func (r *BillRepository) List(ctx context.Context, filter *models.BillFilter, page, perPage int) (*models.PaginatedBills, error) {
	offset := (page - 1) * perPage

	whereClause := "WHERE b.deleted_at IS NULL"
	args := []interface{}{}
	argNum := 1

	if filter != nil {
		if filter.Chamber != nil {
			whereClause += fmt.Sprintf(" AND b.chamber = $%d", argNum)
			args = append(args, *filter.Chamber)
			argNum++
		}
		if filter.Status != nil {
			whereClause += fmt.Sprintf(" AND b.status = $%d", argNum)
			args = append(args, *filter.Status)
			argNum++
		}
		if filter.SessionID != nil {
			whereClause += fmt.Sprintf(" AND b.session_id = $%d", argNum)
			args = append(args, *filter.SessionID)
			argNum++
		}
		if filter.TopicID != nil {
			whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM bill_topic_assignments bta WHERE bta.bill_id = b.id AND bta.topic_id = $%d)", argNum)
			args = append(args, *filter.TopicID)
			argNum++
		}
		if filter.AuthorID != nil {
			whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM bill_authors ba WHERE ba.bill_id = b.id AND ba.politician_id = $%d)", argNum)
			args = append(args, *filter.AuthorID)
			argNum++
		}
		if filter.Search != nil && *filter.Search != "" {
			whereClause += fmt.Sprintf(" AND (b.title ILIKE $%d OR b.bill_number ILIKE $%d OR b.short_title ILIKE $%d)", argNum, argNum, argNum)
			args = append(args, "%"+*filter.Search+"%")
			argNum++
		}
		if filter.FiledAfter != nil {
			whereClause += fmt.Sprintf(" AND b.filed_date >= $%d", argNum)
			args = append(args, *filter.FiledAfter)
			argNum++
		}
		if filter.FiledBefore != nil {
			whereClause += fmt.Sprintf(" AND b.filed_date <= $%d", argNum)
			args = append(args, *filter.FiledBefore)
			argNum++
		}
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM bills b %s", whereClause)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count bills: %w", err)
	}

	// Get bills
	query := fmt.Sprintf(`
		SELECT b.id, b.chamber, b.bill_number, b.title, b.slug, b.short_title, b.status, b.filed_date, b.last_action_date,
		       COALESCE((SELECT COUNT(*) FROM bill_authors WHERE bill_id = b.id), 0) as author_count,
		       COALESCE((SELECT array_agg(bt.name) FROM bill_topics bt JOIN bill_topic_assignments bta ON bt.id = bta.topic_id WHERE bta.bill_id = b.id), '{}') as topic_names
		FROM bills b
		%s
		ORDER BY b.filed_date DESC, b.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)
	args = append(args, perPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list bills: %w", err)
	}
	defer rows.Close()

	var bills []models.BillListItem
	for rows.Next() {
		var b models.BillListItem
		err := rows.Scan(
			&b.ID, &b.Chamber, &b.BillNumber, &b.Title, &b.Slug, &b.ShortTitle, &b.Status, &b.FiledDate, &b.LastActionDate,
			&b.AuthorCount, &b.TopicNames,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bill: %w", err)
		}
		bills = append(bills, b)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedBills{
		Bills:      bills,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *BillRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateBillRequest) (*models.Bill, error) {
	setClauses := []string{}
	args := []interface{}{id}
	argNum := 2

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argNum))
		args = append(args, *req.Title)
		argNum++
	}
	if req.Slug != nil {
		setClauses = append(setClauses, fmt.Sprintf("slug = $%d", argNum))
		args = append(args, *req.Slug)
		argNum++
	}
	if req.ShortTitle != nil {
		setClauses = append(setClauses, fmt.Sprintf("short_title = $%d", argNum))
		args = append(args, *req.ShortTitle)
		argNum++
	}
	if req.Summary != nil {
		setClauses = append(setClauses, fmt.Sprintf("summary = $%d", argNum))
		args = append(args, *req.Summary)
		argNum++
	}
	if req.FullText != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_text = $%d", argNum))
		args = append(args, *req.FullText)
		argNum++
	}
	if req.Significance != nil {
		setClauses = append(setClauses, fmt.Sprintf("significance = $%d", argNum))
		args = append(args, *req.Significance)
		argNum++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *req.Status)
		argNum++
	}
	if req.LastActionDate != nil {
		date, err := time.Parse("2006-01-02", *req.LastActionDate)
		if err != nil {
			return nil, fmt.Errorf("invalid last_action_date format: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("last_action_date = $%d", argNum))
		args = append(args, date)
		argNum++
	}
	if req.DateSigned != nil {
		date, err := time.Parse("2006-01-02", *req.DateSigned)
		if err != nil {
			return nil, fmt.Errorf("invalid date_signed format: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("date_signed = $%d", argNum))
		args = append(args, date)
		argNum++
	}
	if req.RepublicActNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("republic_act_number = $%d", argNum))
		args = append(args, *req.RepublicActNumber)
		argNum++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE bills SET %s
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, session_id, chamber, bill_number, title, slug, short_title, summary, full_text, significance,
		          status, filed_date, last_action_date, date_signed, republic_act_number, created_at, updated_at
	`, strings.Join(setClauses, ", "))

	bill := &models.Bill{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&bill.ID, &bill.SessionID, &bill.Chamber, &bill.BillNumber, &bill.Title, &bill.Slug, &bill.ShortTitle,
		&bill.Summary, &bill.FullText, &bill.Significance, &bill.Status, &bill.FiledDate, &bill.LastActionDate,
		&bill.DateSigned, &bill.RepublicActNumber, &bill.CreatedAt, &bill.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update bill: %w", err)
	}

	return bill, nil
}

func (r *BillRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE bills SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("failed to delete bill: %w", err)
	}
	return nil
}

// Bill Authors

func (r *BillRepository) GetBillAuthors(ctx context.Context, billID uuid.UUID) ([]models.BillAuthor, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ba.id, ba.bill_id, ba.politician_id, ba.is_principal_author, ba.created_at,
		       p.id, p.name, p.slug, p.photo, p.position, p.party
		FROM bill_authors ba
		JOIN politicians p ON ba.politician_id = p.id
		WHERE ba.bill_id = $1
		ORDER BY ba.is_principal_author DESC, p.name
	`, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bill authors: %w", err)
	}
	defer rows.Close()

	var authors []models.BillAuthor
	for rows.Next() {
		var a models.BillAuthor
		var pol models.PoliticianListItem
		err := rows.Scan(
			&a.ID, &a.BillID, &a.PoliticianID, &a.IsPrincipalAuthor, &a.CreatedAt,
			&pol.ID, &pol.Name, &pol.Slug, &pol.Photo, &pol.Position, &pol.Party,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bill author: %w", err)
		}
		a.Politician = &pol
		authors = append(authors, a)
	}
	return authors, nil
}

// Bill Status History

func (r *BillRepository) GetBillStatusHistory(ctx context.Context, billID uuid.UUID) ([]models.BillStatusHistoryItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, bill_id, status, action_description, action_date, created_at
		FROM bill_status_history
		WHERE bill_id = $1
		ORDER BY action_date DESC, created_at DESC
	`, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bill status history: %w", err)
	}
	defer rows.Close()

	var history []models.BillStatusHistoryItem
	for rows.Next() {
		var h models.BillStatusHistoryItem
		err := rows.Scan(&h.ID, &h.BillID, &h.Status, &h.ActionDescription, &h.ActionDate, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status history: %w", err)
		}
		history = append(history, h)
	}
	return history, nil
}

func (r *BillRepository) AddBillStatus(ctx context.Context, billID uuid.UUID, req *models.AddBillStatusRequest) error {
	actionDate, err := time.Parse("2006-01-02", req.ActionDate)
	if err != nil {
		return fmt.Errorf("invalid action_date format: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Add status history entry
	_, err = tx.Exec(ctx, `
		INSERT INTO bill_status_history (bill_id, status, action_description, action_date)
		VALUES ($1, $2, $3, $4)
	`, billID, req.Status, req.ActionDescription, actionDate)
	if err != nil {
		return fmt.Errorf("failed to add status history: %w", err)
	}

	// Update bill status and last_action_date
	_, err = tx.Exec(ctx, `
		UPDATE bills SET status = $1, last_action_date = $2 WHERE id = $3
	`, req.Status, actionDate, billID)
	if err != nil {
		return fmt.Errorf("failed to update bill status: %w", err)
	}

	return tx.Commit(ctx)
}

// Bill Topics

func (r *BillRepository) GetBillTopics(ctx context.Context, billID uuid.UUID) ([]models.BillTopic, error) {
	rows, err := r.db.Query(ctx, `
		SELECT bt.id, bt.name, bt.slug, bt.description, bt.created_at
		FROM bill_topics bt
		JOIN bill_topic_assignments bta ON bt.id = bta.topic_id
		WHERE bta.bill_id = $1
		ORDER BY bt.name
	`, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bill topics: %w", err)
	}
	defer rows.Close()

	var topics []models.BillTopic
	for rows.Next() {
		var t models.BillTopic
		err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Description, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan topic: %w", err)
		}
		topics = append(topics, t)
	}
	return topics, nil
}

func (r *BillRepository) ListAllTopics(ctx context.Context) ([]models.BillTopic, error) {
	rows, err := r.db.Query(ctx, `
		SELECT bt.id, bt.name, bt.slug, bt.description, bt.created_at,
		       COALESCE((SELECT COUNT(*) FROM bill_topic_assignments WHERE topic_id = bt.id), 0) as bill_count
		FROM bill_topics bt
		ORDER BY bt.name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}
	defer rows.Close()

	var topics []models.BillTopic
	for rows.Next() {
		var t models.BillTopic
		err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Description, &t.CreatedAt, &t.BillCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan topic: %w", err)
		}
		topics = append(topics, t)
	}
	return topics, nil
}

// Bill Committees

func (r *BillRepository) GetBillCommittees(ctx context.Context, billID uuid.UUID) ([]models.BillCommittee, error) {
	rows, err := r.db.Query(ctx, `
		SELECT bc.id, bc.bill_id, bc.committee_id, bc.referred_date, bc.is_primary, bc.status, bc.created_at,
		       c.id, c.chamber, c.name, c.slug, c.is_active
		FROM bill_committees bc
		JOIN committees c ON bc.committee_id = c.id
		WHERE bc.bill_id = $1
		ORDER BY bc.is_primary DESC, bc.referred_date
	`, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bill committees: %w", err)
	}
	defer rows.Close()

	var committees []models.BillCommittee
	for rows.Next() {
		var bc models.BillCommittee
		var comm models.CommitteeListItem
		err := rows.Scan(
			&bc.ID, &bc.BillID, &bc.CommitteeID, &bc.ReferredDate, &bc.IsPrimary, &bc.Status, &bc.CreatedAt,
			&comm.ID, &comm.Chamber, &comm.Name, &comm.Slug, &comm.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bill committee: %w", err)
		}
		bc.Committee = &comm
		committees = append(committees, bc)
	}
	return committees, nil
}

// Bill Votes

func (r *BillRepository) GetBillVotes(ctx context.Context, billID uuid.UUID) ([]models.BillVote, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, bill_id, chamber, reading, vote_date, yeas, nays, abstentions, absent, is_passed, notes, created_at
		FROM bill_votes
		WHERE bill_id = $1
		ORDER BY vote_date DESC
	`, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bill votes: %w", err)
	}
	defer rows.Close()

	var votes []models.BillVote
	for rows.Next() {
		var v models.BillVote
		err := rows.Scan(
			&v.ID, &v.BillID, &v.Chamber, &v.Reading, &v.VoteDate, &v.Yeas, &v.Nays, &v.Abstentions, &v.Absent, &v.IsPassed, &v.Notes, &v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vote: %w", err)
		}
		votes = append(votes, v)
	}
	return votes, nil
}

func (r *BillRepository) AddBillVote(ctx context.Context, billID uuid.UUID, req *models.AddBillVoteRequest) (*models.BillVote, error) {
	voteDate, err := time.Parse("2006-01-02", req.VoteDate)
	if err != nil {
		return nil, fmt.Errorf("invalid vote_date format: %w", err)
	}

	vote := &models.BillVote{}
	err = r.db.QueryRow(ctx, `
		INSERT INTO bill_votes (bill_id, chamber, reading, vote_date, yeas, nays, abstentions, absent, is_passed, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, bill_id, chamber, reading, vote_date, yeas, nays, abstentions, absent, is_passed, notes, created_at
	`, billID, req.Chamber, req.Reading, voteDate, req.Yeas, req.Nays, req.Abstentions, req.Absent, req.IsPassed, req.Notes).Scan(
		&vote.ID, &vote.BillID, &vote.Chamber, &vote.Reading, &vote.VoteDate, &vote.Yeas, &vote.Nays,
		&vote.Abstentions, &vote.Absent, &vote.IsPassed, &vote.Notes, &vote.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add bill vote: %w", err)
	}
	return vote, nil
}

// Politician Votes

func (r *BillRepository) GetPoliticianVotesForBill(ctx context.Context, billVoteID uuid.UUID) ([]models.PoliticianVote, error) {
	rows, err := r.db.Query(ctx, `
		SELECT pv.id, pv.bill_vote_id, pv.politician_id, pv.vote, pv.created_at,
		       p.id, p.name, p.slug, p.photo, p.position, p.party
		FROM politician_votes pv
		JOIN politicians p ON pv.politician_id = p.id
		WHERE pv.bill_vote_id = $1
		ORDER BY pv.vote, p.name
	`, billVoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get politician votes: %w", err)
	}
	defer rows.Close()

	var votes []models.PoliticianVote
	for rows.Next() {
		var v models.PoliticianVote
		var pol models.PoliticianListItem
		err := rows.Scan(
			&v.ID, &v.BillVoteID, &v.PoliticianID, &v.Vote, &v.CreatedAt,
			&pol.ID, &pol.Name, &pol.Slug, &pol.Photo, &pol.Position, &pol.Party,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan politician vote: %w", err)
		}
		v.Politician = &pol
		votes = append(votes, v)
	}
	return votes, nil
}

func (r *BillRepository) GetPoliticianVotingHistory(ctx context.Context, politicianID uuid.UUID, page, perPage int) (*models.PaginatedPoliticianVotes, error) {
	offset := (page - 1) * perPage

	var total int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM politician_votes pv
		WHERE pv.politician_id = $1
	`, politicianID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count votes: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT b.id, b.chamber, b.bill_number, b.title, b.slug, b.short_title, b.status, b.filed_date, b.last_action_date,
		       pv.vote, bv.vote_date, bv.reading, bv.is_passed
		FROM politician_votes pv
		JOIN bill_votes bv ON pv.bill_vote_id = bv.id
		JOIN bills b ON bv.bill_id = b.id
		WHERE pv.politician_id = $1 AND b.deleted_at IS NULL
		ORDER BY bv.vote_date DESC
		LIMIT $2 OFFSET $3
	`, politicianID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get voting history: %w", err)
	}
	defer rows.Close()

	var votes []models.PoliticianBillVote
	for rows.Next() {
		var v models.PoliticianBillVote
		err := rows.Scan(
			&v.Bill.ID, &v.Bill.Chamber, &v.Bill.BillNumber, &v.Bill.Title, &v.Bill.Slug, &v.Bill.ShortTitle,
			&v.Bill.Status, &v.Bill.FiledDate, &v.Bill.LastActionDate,
			&v.Vote, &v.VoteDate, &v.Reading, &v.BillPassed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vote: %w", err)
		}
		votes = append(votes, v)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedPoliticianVotes{
		Votes:      votes,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *BillRepository) GetPoliticianVotingRecord(ctx context.Context, politicianID uuid.UUID) (*models.PoliticianVotingRecord, error) {
	record := &models.PoliticianVotingRecord{PoliticianID: politicianID}

	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) as total,
		       COALESCE(SUM(CASE WHEN vote = 'yea' THEN 1 ELSE 0 END), 0) as yeas,
		       COALESCE(SUM(CASE WHEN vote = 'nay' THEN 1 ELSE 0 END), 0) as nays,
		       COALESCE(SUM(CASE WHEN vote = 'abstain' THEN 1 ELSE 0 END), 0) as abstains,
		       COALESCE(SUM(CASE WHEN vote = 'absent' THEN 1 ELSE 0 END), 0) as absents
		FROM politician_votes
		WHERE politician_id = $1
	`, politicianID).Scan(
		&record.TotalVotes, &record.YeaVotes, &record.NayVotes, &record.AbstainVotes, &record.AbsentVotes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get voting record: %w", err)
	}

	if record.TotalVotes > 0 {
		presentVotes := record.TotalVotes - record.AbsentVotes
		record.AttendanceRate = float64(presentVotes) / float64(record.TotalVotes) * 100
	}

	return record, nil
}
