package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/humfurie/pulpulitiko/api/internal/models"
)

type SearchAnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewSearchAnalyticsRepository(db *pgxpool.Pool) *SearchAnalyticsRepository {
	return &SearchAnalyticsRepository{db: db}
}

// TrackSearch records a search query
func (r *SearchAnalyticsRepository) TrackSearch(ctx context.Context, query string, userID *uuid.UUID, sessionID *string, matchedPoliticianID *uuid.UUID, resultsCount int) (*models.SearchQuery, error) {
	sq := &models.SearchQuery{
		ID:                  uuid.New(),
		Query:               query,
		QueryNormalized:     strings.ToLower(strings.TrimSpace(query)),
		UserID:              userID,
		SessionID:           sessionID,
		MatchedPoliticianID: matchedPoliticianID,
		ResultsCount:        resultsCount,
		CreatedAt:           time.Now(),
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO search_queries (id, query, query_normalized, user_id, session_id, matched_politician_id, results_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, sq.ID, sq.Query, sq.QueryNormalized, sq.UserID, sq.SessionID, sq.MatchedPoliticianID, sq.ResultsCount, sq.CreatedAt)

	if err != nil {
		return nil, err
	}

	return sq, nil
}

// TrackClick records a click on a search result
func (r *SearchAnalyticsRepository) TrackClick(ctx context.Context, searchQueryID, articleID uuid.UUID, position int) (*models.SearchClick, error) {
	sc := &models.SearchClick{
		ID:            uuid.New(),
		SearchQueryID: searchQueryID,
		ArticleID:     articleID,
		Position:      position,
		CreatedAt:     time.Now(),
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO search_clicks (id, search_query_id, article_id, position, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, sc.ID, sc.SearchQueryID, sc.ArticleID, sc.Position, sc.CreatedAt)

	if err != nil {
		return nil, err
	}

	return sc, nil
}

// GetTopSearchTerms returns the most popular search terms
func (r *SearchAnalyticsRepository) GetTopSearchTerms(ctx context.Context, timeRange models.TimeRange, limit int) ([]models.TopSearchTerm, error) {
	var startTime *time.Time
	if timeRange != models.TimeRangeLifetime {
		t := time.Now().Add(-timeRange.GetDuration())
		startTime = &t
	}

	query := `
		SELECT
			sq.query_normalized as query,
			COUNT(DISTINCT sq.id) as search_count,
			COUNT(DISTINCT sc.id) as click_count
		FROM search_queries sq
		LEFT JOIN search_clicks sc ON sq.id = sc.search_query_id
		WHERE ($1::timestamp IS NULL OR sq.created_at >= $1)
		GROUP BY sq.query_normalized
		ORDER BY search_count DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, startTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []models.TopSearchTerm
	for rows.Next() {
		var t models.TopSearchTerm
		if err := rows.Scan(&t.Query, &t.Count, &t.ClickCount); err != nil {
			return nil, err
		}
		if t.Count > 0 {
			t.CTR = float64(t.ClickCount) / float64(t.Count) * 100
		}
		terms = append(terms, t)
	}

	return terms, nil
}

// GetSearchTrends returns search volume over time
func (r *SearchAnalyticsRepository) GetSearchTrends(ctx context.Context, timeRange models.TimeRange) ([]models.SearchTrend, error) {
	var startTime *time.Time
	if timeRange != models.TimeRangeLifetime {
		t := time.Now().Add(-timeRange.GetDuration())
		startTime = &t
	}

	// Determine grouping interval based on time range
	var interval string
	switch timeRange {
	case models.TimeRange1Hour:
		interval = "minute"
	case models.TimeRange1Day:
		interval = "hour"
	case models.TimeRange1Week, models.TimeRange1Month:
		interval = "day"
	default:
		interval = "month"
	}

	query := `
		SELECT
			DATE_TRUNC($1, sq.created_at) as period,
			COUNT(DISTINCT sq.id) as search_count,
			COUNT(DISTINCT sq.query_normalized) as unique_terms,
			COUNT(DISTINCT sc.id) as click_count,
			COUNT(DISTINCT sc.article_id) as unique_clicks
		FROM search_queries sq
		LEFT JOIN search_clicks sc ON sq.id = sc.search_query_id
		WHERE ($2::timestamp IS NULL OR sq.created_at >= $2)
		GROUP BY DATE_TRUNC($1, sq.created_at)
		ORDER BY period ASC
	`

	rows, err := r.db.Query(ctx, query, interval, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []models.SearchTrend
	for rows.Next() {
		var t models.SearchTrend
		var period time.Time
		if err := rows.Scan(&period, &t.SearchCount, &t.UniqueTerms, &t.ClickCount, &t.UniqueClicks); err != nil {
			return nil, err
		}
		t.Period = period.Format("2006-01-02T15:04:05Z")
		trends = append(trends, t)
	}

	return trends, nil
}

// GetPoliticianSearchStats returns search stats related to politicians
func (r *SearchAnalyticsRepository) GetPoliticianSearchStats(ctx context.Context, timeRange models.TimeRange, limit int) ([]models.PoliticianSearchStats, error) {
	var startTime *time.Time
	if timeRange != models.TimeRangeLifetime {
		t := time.Now().Add(-timeRange.GetDuration())
		startTime = &t
	}

	query := `
		SELECT
			p.id,
			p.name,
			p.slug,
			COUNT(sq.id) as search_count
		FROM politicians p
		INNER JOIN search_queries sq ON sq.matched_politician_id = p.id
		WHERE p.deleted_at IS NULL
		AND ($1::timestamp IS NULL OR sq.created_at >= $1)
		GROUP BY p.id, p.name, p.slug
		ORDER BY search_count DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, startTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.PoliticianSearchStats
	for rows.Next() {
		var s models.PoliticianSearchStats
		if err := rows.Scan(&s.PoliticianID, &s.PoliticianName, &s.PoliticianSlug, &s.SearchCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetTopClickedArticles returns articles most frequently clicked from search
func (r *SearchAnalyticsRepository) GetTopClickedArticles(ctx context.Context, timeRange models.TimeRange, limit int) ([]models.TopClickedArticle, error) {
	var startTime *time.Time
	if timeRange != models.TimeRangeLifetime {
		t := time.Now().Add(-timeRange.GetDuration())
		startTime = &t
	}

	query := `
		SELECT
			a.id,
			a.title,
			a.slug,
			COUNT(sc.id) as click_count,
			AVG(sc.position) as avg_position
		FROM articles a
		INNER JOIN search_clicks sc ON sc.article_id = a.id
		WHERE a.deleted_at IS NULL
		AND ($1::timestamp IS NULL OR sc.created_at >= $1)
		GROUP BY a.id, a.title, a.slug
		ORDER BY click_count DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, startTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.TopClickedArticle
	for rows.Next() {
		var a models.TopClickedArticle
		if err := rows.Scan(&a.ArticleID, &a.ArticleTitle, &a.ArticleSlug, &a.ClickCount, &a.AvgPosition); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}

	return articles, nil
}

// GetTotalStats returns total search and click counts for a time range
func (r *SearchAnalyticsRepository) GetTotalStats(ctx context.Context, timeRange models.TimeRange) (totalSearches, uniqueTerms, totalClicks int, err error) {
	var startTime *time.Time
	if timeRange != models.TimeRangeLifetime {
		t := time.Now().Add(-timeRange.GetDuration())
		startTime = &t
	}

	query := `
		SELECT
			COUNT(DISTINCT sq.id) as total_searches,
			COUNT(DISTINCT sq.query_normalized) as unique_terms,
			COUNT(DISTINCT sc.id) as total_clicks
		FROM search_queries sq
		LEFT JOIN search_clicks sc ON sq.id = sc.search_query_id
		WHERE ($1::timestamp IS NULL OR sq.created_at >= $1)
	`

	err = r.db.QueryRow(ctx, query, startTime).Scan(&totalSearches, &uniqueTerms, &totalClicks)
	return
}

// FindMatchingPolitician tries to match a search query to a politician
func (r *SearchAnalyticsRepository) FindMatchingPolitician(ctx context.Context, query string) (*uuid.UUID, error) {
	normalized := strings.ToLower(strings.TrimSpace(query))

	var id uuid.UUID
	err := r.db.QueryRow(ctx, `
		SELECT id FROM politicians
		WHERE deleted_at IS NULL
		AND (LOWER(name) = $1 OR LOWER(slug) = $1)
		LIMIT 1
	`, normalized).Scan(&id)

	if err != nil {
		return nil, nil // No match found is not an error
	}

	return &id, nil
}
