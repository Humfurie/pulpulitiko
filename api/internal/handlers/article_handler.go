package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type ArticleHandler struct {
	service *services.ArticleService
}

func NewArticleHandler(service *services.ArticleService) *ArticleHandler {
	return &ArticleHandler{service: service}
}

// GET /api/articles
func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := GetPaginationParams(r)

	filter := &models.ArticleFilter{}

	// Only show published articles for public API
	status := models.ArticleStatusPublished
	filter.Status = &status

	// Note: category filtering by slug would need to be resolved to ID via category service
	// For simplicity, we skip this filter in the handler - use /categories/:slug endpoint instead
	_ = r.URL.Query().Get("category")

	articles, err := h.service.List(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch articles")
		return
	}

	WriteSuccess(w, articles)
}

// GET /api/articles/:slug
func (h *ArticleHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	article, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch article")
		return
	}

	if article == nil {
		WriteNotFound(w, "article not found")
		return
	}

	// Only return published articles for public API
	if article.Status != models.ArticleStatusPublished {
		WriteNotFound(w, "article not found")
		return
	}

	WriteSuccess(w, article)
}

// GET /api/articles/trending
func (h *ArticleHandler) GetTrending(w http.ResponseWriter, r *http.Request) {
	articles, err := h.service.GetTrending(r.Context(), 10)
	if err != nil {
		WriteInternalError(w, "failed to fetch trending articles")
		return
	}

	WriteSuccess(w, articles)
}

// GET /api/search
func (h *ArticleHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		WriteBadRequest(w, "search query is required")
		return
	}

	page, perPage := GetPaginationParams(r)

	articles, err := h.service.Search(r.Context(), query, page, perPage)
	if err != nil {
		WriteInternalError(w, "search failed")
		return
	}

	WriteSuccess(w, articles)
}

// POST /api/admin/articles
func (h *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateArticleRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	article, err := h.service.Create(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, article)
}

// PUT /api/admin/articles/:id
func (h *ArticleHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid article ID")
		return
	}

	var req models.UpdateArticleRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	article, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if article == nil {
		WriteNotFound(w, "article not found")
		return
	}

	WriteSuccess(w, article)
}

// DELETE /api/admin/articles/:id
func (h *ArticleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid article ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "article deleted"})
}

// GET /api/admin/articles
func (h *ArticleHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	page, perPage := GetPaginationParams(r)

	filter := &models.ArticleFilter{}

	if status := r.URL.Query().Get("status"); status != "" {
		s := models.ArticleStatus(status)
		filter.Status = &s
	}

	articles, err := h.service.List(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch articles")
		return
	}

	WriteSuccess(w, articles)
}

// GET /api/admin/articles/:id
func (h *ArticleHandler) AdminGetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid article ID")
		return
	}

	article, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch article")
		return
	}

	if article == nil {
		WriteNotFound(w, "article not found")
		return
	}

	WriteSuccess(w, article)
}

// POST /api/articles/:slug/view
func (h *ArticleHandler) IncrementViewCount(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	if err := h.service.IncrementViewCount(r.Context(), slug); err != nil {
		WriteInternalError(w, "failed to increment view count")
		return
	}

	WriteSuccess(w, map[string]string{"message": "view count incremented"})
}

// POST /api/admin/articles/:id/restore
func (h *ArticleHandler) Restore(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid article ID")
		return
	}

	if err := h.service.Restore(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "article restored"})
}

// GET /api/articles/:slug/related
func (h *ArticleHandler) GetRelatedArticles(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	// First get the article to get its category and tags
	article, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch article")
		return
	}

	if article == nil {
		WriteNotFound(w, "article not found")
		return
	}

	// Get tag IDs
	tagIDs := make([]uuid.UUID, len(article.Tags))
	for i, tag := range article.Tags {
		tagIDs[i] = tag.ID
	}

	related, err := h.service.GetRelatedArticles(r.Context(), article.ID, article.CategoryID, tagIDs, 4)
	if err != nil {
		WriteInternalError(w, "failed to fetch related articles")
		return
	}

	WriteSuccess(w, related)
}
