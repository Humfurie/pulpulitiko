package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// GetMentionableUsers GET /api/users/mentionable - Get users that can be mentioned
func (h *UserHandler) GetMentionableUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetMentionableUsers(r.Context())
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if users == nil {
		users = []models.CommentAuthor{}
	}

	WriteSuccess(w, users)
}

// GetUserProfile GET /api/users/{slug}/profile - Get a user's public profile
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "user slug is required")
		return
	}

	// First, find the user by slug
	user, err := h.userRepo.GetUserBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if user == nil {
		WriteNotFound(w, "user not found")
		return
	}

	// Get the profile with counts
	profile, err := h.userRepo.GetUserProfile(r.Context(), user.ID)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if profile == nil {
		WriteNotFound(w, "user profile not found")
		return
	}

	WriteSuccess(w, profile)
}

// GetUserComments GET /api/users/{slug}/comments - Get a user's comments
func (h *UserHandler) GetUserComments(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "user slug is required")
		return
	}

	// Parse pagination
	page := 1
	pageSize := 10
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 50 {
			pageSize = parsed
		}
	}

	// First, find the user by slug
	user, err := h.userRepo.GetUserBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if user == nil {
		WriteNotFound(w, "user not found")
		return
	}

	// Get comments
	comments, err := h.userRepo.GetUserComments(r.Context(), user.ID, page, pageSize)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if comments == nil {
		comments = []models.Comment{}
	}

	WriteSuccess(w, comments)
}

// GetUserReplies GET /api/users/{slug}/replies - Get a user's replies
func (h *UserHandler) GetUserReplies(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "user slug is required")
		return
	}

	// Parse pagination
	page := 1
	pageSize := 10
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 50 {
			pageSize = parsed
		}
	}

	// First, find the user by slug
	user, err := h.userRepo.GetUserBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if user == nil {
		WriteNotFound(w, "user not found")
		return
	}

	// Get replies
	replies, err := h.userRepo.GetUserReplies(r.Context(), user.ID, page, pageSize)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	if replies == nil {
		replies = []models.Comment{}
	}

	WriteSuccess(w, replies)
}

// AdminList GET /api/admin/users - List all users with pagination, search, and sorting (admin)
func (h *UserHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	page, perPage := GetPaginationParams(r)

	search := r.URL.Query().Get("search")
	roleSlug := r.URL.Query().Get("role")
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")

	filter := &models.UserFilter{}
	if search != "" {
		filter.Search = &search
	}
	if roleSlug != "" {
		filter.RoleSlug = &roleSlug
	}
	if sortBy != "" {
		filter.SortBy = &sortBy
	}
	if sortOrder != "" {
		filter.SortOrder = &sortOrder
	}

	paginatedUsers, err := h.userRepo.AdminList(r.Context(), filter, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch users")
		return
	}

	WriteSuccess(w, paginatedUsers)
}
