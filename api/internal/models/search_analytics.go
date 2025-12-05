package models

import (
	"time"

	"github.com/google/uuid"
)

// TimeRange represents a time filter for analytics
type TimeRange string

const (
	TimeRange1Hour    TimeRange = "1h"
	TimeRange1Day     TimeRange = "1d"
	TimeRange1Week    TimeRange = "1w"
	TimeRange1Month   TimeRange = "1m"
	TimeRange1Year    TimeRange = "1y"
	TimeRange5Years   TimeRange = "5y"
	TimeRangeLifetime TimeRange = "lifetime"
)

// GetTimeRangeDuration returns the duration for a given time range
func (t TimeRange) GetDuration() time.Duration {
	switch t {
	case TimeRange1Hour:
		return time.Hour
	case TimeRange1Day:
		return 24 * time.Hour
	case TimeRange1Week:
		return 7 * 24 * time.Hour
	case TimeRange1Month:
		return 30 * 24 * time.Hour
	case TimeRange1Year:
		return 365 * 24 * time.Hour
	case TimeRange5Years:
		return 5 * 365 * 24 * time.Hour
	case TimeRangeLifetime:
		return 0 // Special case: no time limit
	default:
		return 24 * time.Hour
	}
}

// IsValid checks if the time range is valid
func (t TimeRange) IsValid() bool {
	switch t {
	case TimeRange1Hour, TimeRange1Day, TimeRange1Week, TimeRange1Month,
		TimeRange1Year, TimeRange5Years, TimeRangeLifetime:
		return true
	default:
		return false
	}
}

// SearchQuery represents a recorded search query
type SearchQuery struct {
	ID                  uuid.UUID  `json:"id"`
	Query               string     `json:"query"`
	QueryNormalized     string     `json:"query_normalized"`
	UserID              *uuid.UUID `json:"user_id,omitempty"`
	SessionID           *string    `json:"session_id,omitempty"`
	MatchedPoliticianID *uuid.UUID `json:"matched_politician_id,omitempty"`
	ResultsCount        int        `json:"results_count"`
	CreatedAt           time.Time  `json:"created_at"`
}

// SearchClick represents a click on a search result
type SearchClick struct {
	ID            uuid.UUID `json:"id"`
	SearchQueryID uuid.UUID `json:"search_query_id"`
	ArticleID     uuid.UUID `json:"article_id"`
	Position      int       `json:"position"`
	CreatedAt     time.Time `json:"created_at"`
}

// TrackSearchRequest is the request body for tracking a search
type TrackSearchRequest struct {
	Query        string  `json:"query" validate:"required,min=1,max=500"`
	SessionID    *string `json:"session_id,omitempty"`
	ResultsCount int     `json:"results_count"`
}

// TrackSearchResponse is returned after tracking a search
type TrackSearchResponse struct {
	SearchQueryID uuid.UUID `json:"search_query_id"`
}

// TrackClickRequest is the request body for tracking a search result click
type TrackClickRequest struct {
	SearchQueryID uuid.UUID `json:"search_query_id" validate:"required"`
	ArticleID     uuid.UUID `json:"article_id" validate:"required"`
	Position      int       `json:"position" validate:"min=1"`
}

// TopSearchTerm represents a top search term with count
type TopSearchTerm struct {
	Query      string  `json:"query"`
	Count      int     `json:"count"`
	ClickCount int     `json:"click_count"`
	CTR        float64 `json:"ctr"` // Click-through rate
}

// SearchTrend represents search volume over time
type SearchTrend struct {
	Period       string `json:"period"` // Date/hour string depending on granularity
	SearchCount  int    `json:"search_count"`
	UniqueTerms  int    `json:"unique_terms"`
	ClickCount   int    `json:"click_count"`
	UniqueClicks int    `json:"unique_clicks"`
}

// PoliticianSearchStats represents search stats related to a politician
type PoliticianSearchStats struct {
	PoliticianID   uuid.UUID `json:"politician_id"`
	PoliticianName string    `json:"politician_name"`
	PoliticianSlug string    `json:"politician_slug"`
	SearchCount    int       `json:"search_count"`
}

// TopClickedArticle represents an article frequently clicked from search
type TopClickedArticle struct {
	ArticleID    uuid.UUID `json:"article_id"`
	ArticleTitle string    `json:"article_title"`
	ArticleSlug  string    `json:"article_slug"`
	ClickCount   int       `json:"click_count"`
	AvgPosition  float64   `json:"avg_position"`
}

// SearchAnalytics represents the complete analytics data
type SearchAnalytics struct {
	TimeRange          TimeRange               `json:"time_range"`
	TotalSearches      int                     `json:"total_searches"`
	UniqueSearchTerms  int                     `json:"unique_search_terms"`
	TotalClicks        int                     `json:"total_clicks"`
	OverallCTR         float64                 `json:"overall_ctr"`
	TopSearchTerms     []TopSearchTerm         `json:"top_search_terms"`
	SearchTrends       []SearchTrend           `json:"search_trends"`
	PoliticianSearches []PoliticianSearchStats `json:"politician_searches"`
	TopClickedArticles []TopClickedArticle     `json:"top_clicked_articles"`
}
