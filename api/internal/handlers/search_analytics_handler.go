package handlers

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/humfurie/pulpulitiko/api/internal/middleware"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type SearchAnalyticsHandler struct {
	service *services.SearchAnalyticsService
}

func NewSearchAnalyticsHandler(service *services.SearchAnalyticsService) *SearchAnalyticsHandler {
	return &SearchAnalyticsHandler{service: service}
}

// POST /api/search/track - Track a search query
func (h *SearchAnalyticsHandler) TrackSearch(w http.ResponseWriter, r *http.Request) {
	var req models.TrackSearchRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	// Get user ID from context if authenticated
	var userID *uuid.UUID
	if user, ok := r.Context().Value(middleware.UserContextKey).(*models.User); ok && user != nil {
		userID = &user.ID
	}

	searchQuery, err := h.service.TrackSearch(r.Context(), req.Query, userID, req.SessionID, req.ResultsCount)
	if err != nil {
		WriteInternalError(w, "failed to track search")
		return
	}

	WriteSuccess(w, models.TrackSearchResponse{
		SearchQueryID: searchQuery.ID,
	})
}

// POST /api/search/click - Track a click on a search result
func (h *SearchAnalyticsHandler) TrackClick(w http.ResponseWriter, r *http.Request) {
	var req models.TrackClickRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	_, err := h.service.TrackClick(r.Context(), req.SearchQueryID, req.ArticleID, req.Position)
	if err != nil {
		WriteInternalError(w, "failed to track click")
		return
	}

	WriteSuccess(w, map[string]string{"message": "click tracked"})
}

// GET /api/admin/analytics/search - Get search analytics (admin only)
func (h *SearchAnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	timeRangeStr := r.URL.Query().Get("time_range")
	if timeRangeStr == "" {
		timeRangeStr = "1d"
	}

	timeRange := models.TimeRange(timeRangeStr)
	if !timeRange.IsValid() {
		WriteBadRequest(w, "invalid time_range, must be one of: 1h, 1d, 1w, 1m, 1y, 5y, lifetime")
		return
	}

	analytics, err := h.service.GetAnalytics(r.Context(), timeRange)
	if err != nil {
		WriteInternalError(w, "failed to get search analytics")
		return
	}

	WriteSuccess(w, analytics)
}
