package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type SearchAnalyticsService struct {
	repo *repository.SearchAnalyticsRepository
}

func NewSearchAnalyticsService(repo *repository.SearchAnalyticsRepository) *SearchAnalyticsService {
	return &SearchAnalyticsService{repo: repo}
}

// TrackSearch records a search query and tries to match it to a politician
func (s *SearchAnalyticsService) TrackSearch(ctx context.Context, query string, userID *uuid.UUID, sessionID *string, resultsCount int) (*models.SearchQuery, error) {
	// Try to match the query to a politician
	matchedPoliticianID, _ := s.repo.FindMatchingPolitician(ctx, query)

	return s.repo.TrackSearch(ctx, query, userID, sessionID, matchedPoliticianID, resultsCount)
}

// TrackClick records a click on a search result
func (s *SearchAnalyticsService) TrackClick(ctx context.Context, searchQueryID, articleID uuid.UUID, position int) (*models.SearchClick, error) {
	return s.repo.TrackClick(ctx, searchQueryID, articleID, position)
}

// GetAnalytics returns comprehensive search analytics for a time range
func (s *SearchAnalyticsService) GetAnalytics(ctx context.Context, timeRange models.TimeRange) (*models.SearchAnalytics, error) {
	if !timeRange.IsValid() {
		timeRange = models.TimeRange1Day
	}

	// Get all analytics data concurrently
	type result struct {
		topTerms    []models.TopSearchTerm
		trends      []models.SearchTrend
		polStats    []models.PoliticianSearchStats
		topArticles []models.TopClickedArticle
		totalSearch int
		uniqueTerms int
		totalClicks int
		err         error
	}

	ch := make(chan result, 5)

	// Top search terms
	go func() {
		terms, err := s.repo.GetTopSearchTerms(ctx, timeRange, 20)
		ch <- result{topTerms: terms, err: err}
	}()

	// Search trends
	go func() {
		trends, err := s.repo.GetSearchTrends(ctx, timeRange)
		ch <- result{trends: trends, err: err}
	}()

	// Politician search stats
	go func() {
		stats, err := s.repo.GetPoliticianSearchStats(ctx, timeRange, 10)
		ch <- result{polStats: stats, err: err}
	}()

	// Top clicked articles
	go func() {
		articles, err := s.repo.GetTopClickedArticles(ctx, timeRange, 10)
		ch <- result{topArticles: articles, err: err}
	}()

	// Total stats
	go func() {
		total, unique, clicks, err := s.repo.GetTotalStats(ctx, timeRange)
		ch <- result{totalSearch: total, uniqueTerms: unique, totalClicks: clicks, err: err}
	}()

	analytics := &models.SearchAnalytics{
		TimeRange: timeRange,
	}

	// Collect results
	for i := 0; i < 5; i++ {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		if r.topTerms != nil {
			analytics.TopSearchTerms = r.topTerms
		}
		if r.trends != nil {
			analytics.SearchTrends = r.trends
		}
		if r.polStats != nil {
			analytics.PoliticianSearches = r.polStats
		}
		if r.topArticles != nil {
			analytics.TopClickedArticles = r.topArticles
		}
		if r.totalSearch > 0 || r.uniqueTerms > 0 || r.totalClicks > 0 {
			analytics.TotalSearches = r.totalSearch
			analytics.UniqueSearchTerms = r.uniqueTerms
			analytics.TotalClicks = r.totalClicks
		}
	}

	// Calculate overall CTR
	if analytics.TotalSearches > 0 {
		analytics.OverallCTR = float64(analytics.TotalClicks) / float64(analytics.TotalSearches) * 100
	}

	// Ensure slices are not nil
	if analytics.TopSearchTerms == nil {
		analytics.TopSearchTerms = []models.TopSearchTerm{}
	}
	if analytics.SearchTrends == nil {
		analytics.SearchTrends = []models.SearchTrend{}
	}
	if analytics.PoliticianSearches == nil {
		analytics.PoliticianSearches = []models.PoliticianSearchStats{}
	}
	if analytics.TopClickedArticles == nil {
		analytics.TopClickedArticles = []models.TopClickedArticle{}
	}

	return analytics, nil
}
