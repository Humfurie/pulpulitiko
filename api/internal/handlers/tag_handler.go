package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type TagHandler struct {
	tagService     *services.TagService
	articleService *services.ArticleService
}

func NewTagHandler(tagService *services.TagService, articleService *services.ArticleService) *TagHandler {
	return &TagHandler{
		tagService:     tagService,
		articleService: articleService,
	}
}

// GET /api/tags
func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.tagService.List(r.Context())
	if err != nil {
		WriteInternalError(w, "failed to fetch tags")
		return
	}

	WriteSuccess(w, tags)
}

// GET /api/admin/tags - List all tags with pagination, search, and sorting (admin)
func (h *TagHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	page, perPage := GetPaginationParams(r)

	search := r.URL.Query().Get("search")
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")

	filter := &models.TagFilter{}
	if search != "" {
		filter.Search = &search
	}
	if sortBy != "" {
		filter.SortBy = &sortBy
	}
	if sortOrder != "" {
		filter.SortOrder = &sortOrder
	}

	paginatedTags, err := h.tagService.AdminList(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch tags")
		return
	}

	WriteSuccess(w, paginatedTags)
}

// GET /api/tags/:slug
func (h *TagHandler) GetArticlesBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	tag, err := h.tagService.GetBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch tag")
		return
	}

	if tag == nil {
		WriteNotFound(w, "tag not found")
		return
	}

	page, perPage := GetPaginationParams(r)

	status := models.ArticleStatusPublished
	filter := &models.ArticleFilter{
		TagID:  &tag.ID,
		Status: &status,
	}

	articles, err := h.articleService.List(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch articles")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"tag":      tag,
		"articles": articles,
	})
}

// GET /api/admin/tags/:id
func (h *TagHandler) AdminGetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid tag ID")
		return
	}

	tag, err := h.tagService.GetByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch tag")
		return
	}

	if tag == nil {
		WriteNotFound(w, "tag not found")
		return
	}

	WriteSuccess(w, tag)
}

// POST /api/admin/tags
func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTagRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	tag, err := h.tagService.Create(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, tag)
}

// PUT /api/admin/tags/:id
func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid tag ID")
		return
	}

	var req models.UpdateTagRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	tag, err := h.tagService.Update(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if tag == nil {
		WriteNotFound(w, "tag not found")
		return
	}

	WriteSuccess(w, tag)
}

// DELETE /api/admin/tags/:id
func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid tag ID")
		return
	}

	if err := h.tagService.Delete(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "tag deleted"})
}

// POST /api/admin/tags/:id/restore
func (h *TagHandler) Restore(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid tag ID")
		return
	}

	if err := h.tagService.Restore(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "tag restored"})
}
