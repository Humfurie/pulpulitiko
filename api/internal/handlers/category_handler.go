package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
	articleService  *services.ArticleService
}

func NewCategoryHandler(categoryService *services.CategoryService, articleService *services.ArticleService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		articleService:  articleService,
	}
}

// GET /api/categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.List(r.Context())
	if err != nil {
		WriteInternalError(w, "failed to fetch categories")
		return
	}

	WriteSuccess(w, categories)
}

// GET /api/categories/:slug
func (h *CategoryHandler) GetArticlesBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	category, err := h.categoryService.GetBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch category")
		return
	}

	if category == nil {
		WriteNotFound(w, "category not found")
		return
	}

	page, perPage := GetPaginationParams(r)

	status := models.ArticleStatusPublished
	filter := &models.ArticleFilter{
		CategoryID: &category.ID,
		Status:     &status,
	}

	articles, err := h.articleService.List(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch articles")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"category": category,
		"articles": articles,
	})
}

// POST /api/admin/categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCategoryRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	category, err := h.categoryService.Create(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, category)
}

// PUT /api/admin/categories/:id
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid category ID")
		return
	}

	var req models.UpdateCategoryRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	category, err := h.categoryService.Update(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if category == nil {
		WriteNotFound(w, "category not found")
		return
	}

	WriteSuccess(w, category)
}

// DELETE /api/admin/categories/:id
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid category ID")
		return
	}

	if err := h.categoryService.Delete(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "category deleted"})
}
