package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/middleware"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type AuthorHandler struct {
	authorService  *services.AuthorService
	articleService *services.ArticleService
}

func NewAuthorHandler(authorService *services.AuthorService, articleService *services.ArticleService) *AuthorHandler {
	return &AuthorHandler{
		authorService:  authorService,
		articleService: articleService,
	}
}

// GET /api/authors
func (h *AuthorHandler) List(w http.ResponseWriter, r *http.Request) {
	authors, err := h.authorService.List(r.Context())
	if err != nil {
		WriteInternalError(w, "failed to fetch authors")
		return
	}

	WriteSuccess(w, authors)
}

// GET /api/authors/:slug
func (h *AuthorHandler) GetArticlesBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	author, err := h.authorService.GetBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch author")
		return
	}

	if author == nil {
		WriteNotFound(w, "author not found")
		return
	}

	page, perPage := GetPaginationParams(r)

	status := models.ArticleStatusPublished
	filter := &models.ArticleFilter{
		AuthorID: &author.ID,
		Status:   &status,
	}

	articles, err := h.articleService.List(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch articles")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"author":   author,
		"articles": articles,
	})
}

// Admin endpoints - requires admin role

// GET /api/admin/users
func (h *AuthorHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	authors, err := h.authorService.List(r.Context())
	if err != nil {
		WriteInternalError(w, "failed to fetch users")
		return
	}

	WriteSuccess(w, authors)
}

// GET /api/admin/users/:id
func (h *AuthorHandler) AdminGetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid user ID")
		return
	}

	author, err := h.authorService.GetByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch user")
		return
	}

	if author == nil {
		WriteNotFound(w, "user not found")
		return
	}

	WriteSuccess(w, author)
}

// POST /api/admin/users
func (h *AuthorHandler) AdminCreate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAuthorRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	author, err := h.authorService.Create(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, author)
}

// PUT /api/admin/users/:id
func (h *AuthorHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid user ID")
		return
	}

	var req models.UpdateAuthorRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	author, err := h.authorService.Update(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if author == nil {
		WriteNotFound(w, "user not found")
		return
	}

	WriteSuccess(w, author)
}

// DELETE /api/admin/users/:id
func (h *AuthorHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid user ID")
		return
	}

	if err := h.authorService.Delete(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "user deleted"})
}

// POST /api/admin/users/:id/restore
func (h *AuthorHandler) AdminRestore(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid user ID")
		return
	}

	if err := h.authorService.Restore(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "user restored"})
}

// Account endpoints - for current authenticated user

// GET /api/auth/account
func (h *AuthorHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	author, err := h.authorService.GetByEmail(r.Context(), claims.Email)
	if err != nil {
		WriteInternalError(w, "failed to fetch account")
		return
	}

	if author == nil {
		WriteNotFound(w, "account not found")
		return
	}

	WriteSuccess(w, author)
}

// PUT /api/auth/account
func (h *AuthorHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	var req models.UpdateAuthorRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	// Users cannot change their own role or email via this endpoint
	req.RoleID = nil
	req.Email = nil

	author, err := h.authorService.UpdateByEmail(r.Context(), claims.Email, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if author == nil {
		WriteNotFound(w, "account not found")
		return
	}

	WriteSuccess(w, author)
}
