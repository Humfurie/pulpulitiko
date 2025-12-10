package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type ElectionHandler struct {
	service *services.ElectionService
}

func NewElectionHandler(service *services.ElectionService) *ElectionHandler {
	return &ElectionHandler{service: service}
}

// Elections

func (h *ElectionHandler) CreateElection(w http.ResponseWriter, r *http.Request) {
	var req models.CreateElectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	election, err := h.service.CreateElection(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, election)
}

func (h *ElectionHandler) GetElectionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid election ID")
		return
	}

	election, err := h.service.GetElectionByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if election == nil {
		WriteNotFound(w, "Election not found")
		return
	}

	WriteSuccess(w, election)
}

func (h *ElectionHandler) GetElectionBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	election, err := h.service.GetElectionBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if election == nil {
		WriteNotFound(w, "Election not found")
		return
	}

	WriteSuccess(w, election)
}

func (h *ElectionHandler) ListElections(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(query.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := &models.ElectionFilter{}

	if electionType := query.Get("election_type"); electionType != "" {
		filter.ElectionType = &electionType
	}
	if status := query.Get("status"); status != "" {
		filter.Status = &status
	}
	if yearStr := query.Get("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filter.Year = &year
		}
	}
	if isFeatured := query.Get("is_featured"); isFeatured == "true" {
		featured := true
		filter.IsFeatured = &featured
	}
	if search := query.Get("search"); search != "" {
		filter.Search = &search
	}

	result, err := h.service.ListElections(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, result)
}

func (h *ElectionHandler) GetUpcomingElections(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	elections, err := h.service.GetUpcomingElections(r.Context(), limit)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, elections)
}

func (h *ElectionHandler) GetFeaturedElections(w http.ResponseWriter, r *http.Request) {
	elections, err := h.service.GetFeaturedElections(r.Context())
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, elections)
}

func (h *ElectionHandler) GetElectionCalendar(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	year := time.Now().Year()
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	items, err := h.service.GetElectionCalendar(r.Context(), year)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, items)
}

func (h *ElectionHandler) UpdateElection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid election ID")
		return
	}

	var req models.UpdateElectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	election, err := h.service.UpdateElection(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if election == nil {
		WriteNotFound(w, "Election not found")
		return
	}

	WriteSuccess(w, election)
}

func (h *ElectionHandler) DeleteElection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid election ID")
		return
	}

	if err := h.service.DeleteElection(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "Election deleted"})
}

// Election Positions

func (h *ElectionHandler) CreateElectionPosition(w http.ResponseWriter, r *http.Request) {
	var req models.CreateElectionPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	position, err := h.service.CreateElectionPosition(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, position)
}

func (h *ElectionHandler) GetElectionPositions(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid election ID")
		return
	}

	positions, err := h.service.GetElectionPositions(r.Context(), id)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, positions)
}

// Candidates

func (h *ElectionHandler) CreateCandidate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	candidate, err := h.service.CreateCandidate(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, candidate)
}

func (h *ElectionHandler) GetCandidateByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid candidate ID")
		return
	}

	candidate, err := h.service.GetCandidateByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if candidate == nil {
		WriteNotFound(w, "Candidate not found")
		return
	}

	WriteSuccess(w, candidate)
}

func (h *ElectionHandler) GetCandidatesForPosition(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "positionId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid position ID")
		return
	}

	candidates, err := h.service.GetCandidatesForPosition(r.Context(), id)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, candidates)
}

func (h *ElectionHandler) ListCandidates(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(query.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filter := &models.CandidateFilter{}

	if electionID := query.Get("election_id"); electionID != "" {
		if id, err := uuid.Parse(electionID); err == nil {
			filter.ElectionID = &id
		}
	}
	if positionID := query.Get("position_id"); positionID != "" {
		if id, err := uuid.Parse(positionID); err == nil {
			filter.PositionID = &id
		}
	}
	if politicianID := query.Get("politician_id"); politicianID != "" {
		if id, err := uuid.Parse(politicianID); err == nil {
			filter.PoliticianID = &id
		}
	}
	if partyID := query.Get("party_id"); partyID != "" {
		if id, err := uuid.Parse(partyID); err == nil {
			filter.PartyID = &id
		}
	}
	if status := query.Get("status"); status != "" {
		filter.Status = &status
	}
	if isWinner := query.Get("is_winner"); isWinner == "true" {
		winner := true
		filter.IsWinner = &winner
	}

	result, err := h.service.ListCandidates(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, result)
}

func (h *ElectionHandler) UpdateCandidate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid candidate ID")
		return
	}

	var req models.UpdateCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	candidate, err := h.service.UpdateCandidate(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if candidate == nil {
		WriteNotFound(w, "Candidate not found")
		return
	}

	WriteSuccess(w, candidate)
}

// Voter Education

func (h *ElectionHandler) CreateVoterEducation(w http.ResponseWriter, r *http.Request) {
	var req models.CreateVoterEducationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	ve, err := h.service.CreateVoterEducation(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, ve)
}

func (h *ElectionHandler) GetVoterEducationBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	ve, err := h.service.GetVoterEducationBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if ve == nil {
		WriteNotFound(w, "Content not found")
		return
	}

	// Increment view count asynchronously
	go func() {
		_ = h.service.IncrementVoterEducationViewCount(r.Context(), ve.ID)
	}()

	WriteSuccess(w, ve)
}

func (h *ElectionHandler) ListVoterEducation(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(query.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var electionID *uuid.UUID
	if eid := query.Get("election_id"); eid != "" {
		if id, err := uuid.Parse(eid); err == nil {
			electionID = &id
		}
	}

	var category *string
	if cat := query.Get("category"); cat != "" {
		category = &cat
	}

	result, err := h.service.ListVoterEducation(r.Context(), electionID, category, page, perPage)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, result)
}
