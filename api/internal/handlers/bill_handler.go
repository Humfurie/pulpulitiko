package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type BillHandler struct {
	service  *services.BillService
	validate *validator.Validate
}

func NewBillHandler(service *services.BillService) *BillHandler {
	return &BillHandler{
		service:  service,
		validate: validator.New(),
	}
}

// Legislative Sessions

func (h *BillHandler) GetCurrentSession(w http.ResponseWriter, r *http.Request) {
	session, err := h.service.GetCurrentSession(r.Context())
	if err != nil {
		WriteInternalError(w, "Failed to get current session")
		return
	}
	if session == nil {
		WriteNotFound(w, "No current session found")
		return
	}
	WriteSuccess(w, session)
}

func (h *BillHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.service.ListSessions(r.Context())
	if err != nil {
		WriteInternalError(w, "Failed to list sessions")
		return
	}
	WriteSuccess(w, sessions)
}

// Committees

func (h *BillHandler) ListCommittees(w http.ResponseWriter, r *http.Request) {
	var chamber *string
	if c := r.URL.Query().Get("chamber"); c != "" {
		chamber = &c
	}

	committees, err := h.service.ListCommittees(r.Context(), chamber)
	if err != nil {
		WriteInternalError(w, "Failed to list committees")
		return
	}
	WriteSuccess(w, committees)
}

func (h *BillHandler) GetCommitteeBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	committee, err := h.service.GetCommitteeBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "Failed to get committee")
		return
	}
	if committee == nil {
		WriteNotFound(w, "Committee not found")
		return
	}
	WriteSuccess(w, committee)
}

// Bills - Public Endpoints

func (h *BillHandler) ListBills(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	filter := &models.BillFilter{}

	if chamber := r.URL.Query().Get("chamber"); chamber != "" {
		filter.Chamber = &chamber
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	if sessionID := r.URL.Query().Get("session_id"); sessionID != "" {
		if id, err := uuid.Parse(sessionID); err == nil {
			filter.SessionID = &id
		}
	}
	if topicID := r.URL.Query().Get("topic_id"); topicID != "" {
		if id, err := uuid.Parse(topicID); err == nil {
			filter.TopicID = &id
		}
	}
	if authorID := r.URL.Query().Get("author_id"); authorID != "" {
		if id, err := uuid.Parse(authorID); err == nil {
			filter.AuthorID = &id
		}
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}

	bills, err := h.service.ListBills(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "Failed to list bills")
		return
	}
	WriteSuccess(w, bills)
}

func (h *BillHandler) GetBillBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	bill, err := h.service.GetBillBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "Failed to get bill")
		return
	}
	if bill == nil {
		WriteNotFound(w, "Bill not found")
		return
	}
	WriteSuccess(w, bill)
}

func (h *BillHandler) GetBillByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid bill ID")
		return
	}

	bill, err := h.service.GetBillByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "Failed to get bill")
		return
	}
	if bill == nil {
		WriteNotFound(w, "Bill not found")
		return
	}
	WriteSuccess(w, bill)
}

// Bill Topics

func (h *BillHandler) ListAllTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.service.ListAllTopics(r.Context())
	if err != nil {
		WriteInternalError(w, "Failed to list topics")
		return
	}
	WriteSuccess(w, topics)
}

// Politician Voting Records

func (h *BillHandler) GetPoliticianVotingHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid politician ID")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	history, err := h.service.GetPoliticianVotingHistory(r.Context(), id, page, perPage)
	if err != nil {
		WriteInternalError(w, "Failed to get voting history")
		return
	}
	WriteSuccess(w, history)
}

func (h *BillHandler) GetPoliticianVotingRecord(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid politician ID")
		return
	}

	record, err := h.service.GetPoliticianVotingRecord(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "Failed to get voting record")
		return
	}
	WriteSuccess(w, record)
}

// Admin Endpoints

func (h *BillHandler) CreateBill(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		WriteBadRequest(w, err.Error())
		return
	}

	bill, err := h.service.CreateBill(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, "Failed to create bill")
		return
	}
	WriteCreated(w, bill)
}

func (h *BillHandler) UpdateBill(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid bill ID")
		return
	}

	var req models.UpdateBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		WriteBadRequest(w, err.Error())
		return
	}

	bill, err := h.service.UpdateBill(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, "Failed to update bill")
		return
	}
	if bill == nil {
		WriteNotFound(w, "Bill not found")
		return
	}
	WriteSuccess(w, bill)
}

func (h *BillHandler) DeleteBill(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid bill ID")
		return
	}

	err = h.service.DeleteBill(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "Failed to delete bill")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BillHandler) AddBillStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid bill ID")
		return
	}

	var req models.AddBillStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		WriteBadRequest(w, err.Error())
		return
	}

	err = h.service.AddBillStatus(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, "Failed to add bill status")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *BillHandler) AddBillVote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid bill ID")
		return
	}

	var req models.AddBillVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		WriteBadRequest(w, err.Error())
		return
	}

	vote, err := h.service.AddBillVote(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, "Failed to add bill vote")
		return
	}
	WriteCreated(w, vote)
}

func (h *BillHandler) GetBillVotes(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid bill ID")
		return
	}

	votes, err := h.service.GetBillVotes(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "Failed to get bill votes")
		return
	}
	WriteSuccess(w, votes)
}

func (h *BillHandler) GetPoliticianVotesForBillVote(w http.ResponseWriter, r *http.Request) {
	voteIDStr := chi.URLParam(r, "voteId")
	voteID, err := uuid.Parse(voteIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid vote ID")
		return
	}

	votes, err := h.service.GetPoliticianVotesForBill(r.Context(), voteID)
	if err != nil {
		WriteInternalError(w, "Failed to get politician votes")
		return
	}
	WriteSuccess(w, votes)
}
