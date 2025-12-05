package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type PoliticianHandler struct {
	politicianService *services.PoliticianService
	articleService    *services.ArticleService
}

func NewPoliticianHandler(politicianService *services.PoliticianService, articleService *services.ArticleService) *PoliticianHandler {
	return &PoliticianHandler{
		politicianService: politicianService,
		articleService:    articleService,
	}
}

// GET /api/politicians - List all politicians (public)
func (h *PoliticianHandler) List(w http.ResponseWriter, r *http.Request) {
	politicians, err := h.politicianService.ListAll(r.Context())
	if err != nil {
		WriteInternalError(w, "failed to fetch politicians")
		return
	}

	WriteSuccess(w, politicians)
}

// GET /api/politicians/search?q= - Search politicians for autocomplete
func (h *PoliticianHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		WriteSuccess(w, []models.Politician{})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	politicians, err := h.politicianService.Search(r.Context(), query, limit)
	if err != nil {
		WriteInternalError(w, "failed to search politicians")
		return
	}

	WriteSuccess(w, politicians)
}

// GET /api/politicians/:slug - Get politician profile with articles
func (h *PoliticianHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	politician, err := h.politicianService.GetBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch politician")
		return
	}

	if politician == nil {
		WriteNotFound(w, "politician not found")
		return
	}

	page, perPage := GetPaginationParams(r)

	status := models.ArticleStatusPublished
	filter := &models.ArticleFilter{
		PoliticianID: &politician.ID,
		Status:       &status,
	}

	articles, err := h.articleService.List(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch articles")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"politician": politician,
		"articles":   articles,
	})
}

// GET /api/admin/politicians - List all politicians (admin, paginated)
func (h *PoliticianHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	page, perPage := GetPaginationParams(r)

	search := r.URL.Query().Get("search")
	party := r.URL.Query().Get("party")

	filter := &models.PoliticianFilter{}
	if search != "" {
		filter.Search = &search
	}
	if party != "" {
		filter.Party = &party
	}

	politicians, err := h.politicianService.List(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch politicians")
		return
	}

	WriteSuccess(w, politicians)
}

// GET /api/admin/politicians/:id - Get politician by ID (admin)
func (h *PoliticianHandler) AdminGetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid politician ID")
		return
	}

	politician, err := h.politicianService.GetByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch politician")
		return
	}

	if politician == nil {
		WriteNotFound(w, "politician not found")
		return
	}

	WriteSuccess(w, politician)
}

// POST /api/admin/politicians - Create politician
func (h *PoliticianHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePoliticianRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	politician, err := h.politicianService.Create(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, politician)
}

// PUT /api/admin/politicians/:id - Update politician
func (h *PoliticianHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid politician ID")
		return
	}

	var req models.UpdatePoliticianRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	politician, err := h.politicianService.Update(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if politician == nil {
		WriteNotFound(w, "politician not found")
		return
	}

	WriteSuccess(w, politician)
}

// DELETE /api/admin/politicians/:id - Delete politician (soft)
func (h *PoliticianHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid politician ID")
		return
	}

	if err := h.politicianService.Delete(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "politician deleted"})
}

// POST /api/admin/politicians/:id/restore - Restore politician
func (h *PoliticianHandler) Restore(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid politician ID")
		return
	}

	if err := h.politicianService.Restore(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "politician restored"})
}
