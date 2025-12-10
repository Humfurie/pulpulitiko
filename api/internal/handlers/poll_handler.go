package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type PollHandler struct {
	service *services.PollService
}

func NewPollHandler(service *services.PollService) *PollHandler {
	return &PollHandler{service: service}
}

// Public endpoints

func (h *PollHandler) ListPolls(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(query.Get("per_page"))
	if perPage < 1 || perPage > 50 {
		perPage = 12
	}

	filter := &models.PollFilter{
		ActiveOnly:      true, // Public endpoint only shows active polls
		IncludeNational: true, // By default include national polls
	}

	if category := query.Get("category"); category != "" {
		filter.Category = &category
	}
	if search := query.Get("search"); search != "" {
		filter.Search = &search
	}
	if politicianID := query.Get("politician_id"); politicianID != "" {
		if id, err := uuid.Parse(politicianID); err == nil {
			filter.PoliticianID = &id
		}
	}
	if electionID := query.Get("election_id"); electionID != "" {
		if id, err := uuid.Parse(electionID); err == nil {
			filter.ElectionID = &id
		}
	}

	// Location filters
	if regionID := query.Get("region_id"); regionID != "" {
		if id, err := uuid.Parse(regionID); err == nil {
			filter.RegionID = &id
		}
	}
	if provinceID := query.Get("province_id"); provinceID != "" {
		if id, err := uuid.Parse(provinceID); err == nil {
			filter.ProvinceID = &id
		}
	}
	if cityID := query.Get("city_municipality_id"); cityID != "" {
		if id, err := uuid.Parse(cityID); err == nil {
			filter.CityMunicipalityID = &id
		}
	}
	if barangayID := query.Get("barangay_id"); barangayID != "" {
		if id, err := uuid.Parse(barangayID); err == nil {
			filter.BarangayID = &id
		}
	}
	// Allow client to control whether to include national polls
	if includeNational := query.Get("include_national"); includeNational == "false" {
		filter.IncludeNational = false
	}

	result, err := h.service.ListPolls(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, result)
}

func (h *PollHandler) GetPollBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Get user ID from context if authenticated
	var userID *uuid.UUID
	if uid, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		userID = &uid
	}

	// Get IP hash for anonymous voting check
	ip := getClientIP(r)
	ipHash := services.HashIP(ip, uuid.Nil) // We'll update this once we have the poll ID

	poll, err := h.service.GetPollBySlug(r.Context(), slug, userID, &ipHash)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if poll == nil {
		WriteNotFound(w, "Poll not found")
		return
	}

	// Update IP hash with actual poll ID
	if userID == nil {
		ipHash = services.HashIP(ip, poll.ID)
		hasVoted, optionID := h.service.HasUserVoted(r.Context(), poll.ID, nil, &ipHash)
		if hasVoted {
			poll.UserVote = optionID
		}
	}

	// Increment view count asynchronously
	go func() {
		_ = h.service.IncrementViewCount(r.Context(), poll.ID)
	}()

	WriteSuccess(w, poll)
}

func (h *PollHandler) GetPollByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	var userID *uuid.UUID
	if uid, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		userID = &uid
	}

	ip := getClientIP(r)
	ipHash := services.HashIP(ip, id)

	poll, err := h.service.GetPollByID(r.Context(), id, userID, &ipHash)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if poll == nil {
		WriteNotFound(w, "Poll not found")
		return
	}

	WriteSuccess(w, poll)
}

func (h *PollHandler) GetFeaturedPolls(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 10 {
			limit = l
		}
	}

	polls, err := h.service.GetFeaturedPolls(r.Context(), limit)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, polls)
}

func (h *PollHandler) GetPollResults(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	results, err := h.service.GetPollResults(r.Context(), id)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, results)
}

func (h *PollHandler) CastVote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	pollID, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	var req models.CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	var userID *uuid.UUID
	if uid, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		userID = &uid
	}

	ip := getClientIP(r)

	result, err := h.service.CastVote(r.Context(), pollID, req.OptionID, userID, ip)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if !result.Success {
		WriteBadRequest(w, result.Message)
		return
	}

	WriteSuccess(w, result)
}

// Comments

func (h *PollHandler) GetPollComments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(query.Get("per_page"))
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	comments, err := h.service.GetPollComments(r.Context(), id, page, perPage)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, comments)
}

func (h *PollHandler) CreatePollComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	pollID, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		WriteUnauthorized(w, "Authentication required")
		return
	}

	var req models.CreatePollCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	comment, err := h.service.CreatePollComment(r.Context(), pollID, userID, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, comment)
}

func (h *PollHandler) DeletePollComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid comment ID")
		return
	}

	if err := h.service.DeletePollComment(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "Comment deleted"})
}

// Authenticated user endpoints

func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		WriteUnauthorized(w, "Authentication required")
		return
	}

	var req models.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	poll, err := h.service.CreatePoll(r.Context(), userID, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, poll)
}

func (h *PollHandler) UpdatePoll(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	var req models.UpdatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	poll, err := h.service.UpdatePoll(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if poll == nil {
		WriteNotFound(w, "Poll not found")
		return
	}

	WriteSuccess(w, poll)
}

func (h *PollHandler) SubmitForApproval(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	if err := h.service.SubmitForApproval(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "Poll submitted for approval"})
}

func (h *PollHandler) GetMyPolls(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		WriteUnauthorized(w, "Authentication required")
		return
	}

	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(query.Get("per_page"))
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}

	result, err := h.service.GetUserPolls(r.Context(), userID, page, perPage)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, result)
}

func (h *PollHandler) DeletePoll(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	if err := h.service.DeletePoll(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "Poll deleted"})
}

// Admin endpoints

func (h *PollHandler) AdminListPolls(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(query.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := &models.PollFilter{}

	if category := query.Get("category"); category != "" {
		filter.Category = &category
	}
	if status := query.Get("status"); status != "" {
		filter.Status = &status
	}
	if search := query.Get("search"); search != "" {
		filter.Search = &search
	}
	if isFeatured := query.Get("is_featured"); isFeatured == "true" {
		featured := true
		filter.IsFeatured = &featured
	}

	// Location filters
	if regionID := query.Get("region_id"); regionID != "" {
		if id, err := uuid.Parse(regionID); err == nil {
			filter.RegionID = &id
		}
	}
	if provinceID := query.Get("province_id"); provinceID != "" {
		if id, err := uuid.Parse(provinceID); err == nil {
			filter.ProvinceID = &id
		}
	}
	if cityID := query.Get("city_municipality_id"); cityID != "" {
		if id, err := uuid.Parse(cityID); err == nil {
			filter.CityMunicipalityID = &id
		}
	}
	if barangayID := query.Get("barangay_id"); barangayID != "" {
		if id, err := uuid.Parse(barangayID); err == nil {
			filter.BarangayID = &id
		}
	}

	result, err := h.service.ListPolls(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, result)
}

func (h *PollHandler) AdminUpdatePoll(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	var req models.AdminUpdatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	poll, err := h.service.AdminUpdatePoll(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if poll == nil {
		WriteNotFound(w, "Poll not found")
		return
	}

	WriteSuccess(w, poll)
}

func (h *PollHandler) ApprovePoll(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	approverID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		WriteUnauthorized(w, "Authentication required")
		return
	}

	var req models.ApprovePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.service.ApprovePoll(r.Context(), id, approverID, req.Approved, req.Reason); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	message := "Poll approved"
	if !req.Approved {
		message = "Poll rejected"
	}

	WriteSuccess(w, map[string]string{"message": message})
}

func (h *PollHandler) ClosePoll(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid poll ID")
		return
	}

	if err := h.service.ClosePoll(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "Poll closed"})
}

// Helper function to get client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
