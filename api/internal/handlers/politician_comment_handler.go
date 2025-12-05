package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/middleware"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type PoliticianCommentHandler struct {
	commentService *services.PoliticianCommentService
}

func NewPoliticianCommentHandler(commentService *services.PoliticianCommentService) *PoliticianCommentHandler {
	return &PoliticianCommentHandler{
		commentService: commentService,
	}
}

// ListComments GET /api/politicians/{slug}/comments - List paginated comments for a politician
func (h *PoliticianCommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "politician slug is required")
		return
	}

	// Get pagination params
	page := 1
	perPage := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	// Get current user ID if authenticated (for reaction status)
	var currentUserID *uuid.UUID
	includeHidden := false
	claims := middleware.GetUserClaims(r.Context())
	if claims != nil {
		if userID, err := uuid.Parse(claims.UserID); err == nil {
			currentUserID = &userID
		}
		// Admins can see all comments if they request it
		if claims.Role == "admin" && r.URL.Query().Get("include_hidden") == "true" {
			includeHidden = true
		}
	}

	result, err := h.commentService.ListPoliticianComments(r.Context(), slug, currentUserID, includeHidden, page, perPage)
	if err != nil {
		WriteNotFound(w, err.Error())
		return
	}

	WriteSuccess(w, result)
}

// CreateComment POST /api/politicians/{slug}/comments - Create a new comment
func (h *PoliticianCommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "politician slug is required")
		return
	}

	// Get authenticated user
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		WriteUnauthorized(w, "authentication required")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		WriteUnauthorized(w, "invalid user ID")
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if req.Content == "" {
		WriteBadRequest(w, "content is required")
		return
	}

	comment, err := h.commentService.CreateComment(r.Context(), slug, userID, &req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "politician_comments_user_id_fkey") {
			WriteUnauthorized(w, "user session invalid - please log out and log in again")
			return
		}
		WriteInternalError(w, errMsg)
		return
	}

	WriteSuccessWithStatus(w, http.StatusCreated, comment)
}

// GetComment GET /api/politician-comments/{id} - Get a single comment
func (h *PoliticianCommentHandler) GetComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid comment ID")
		return
	}

	comment, err := h.commentService.GetComment(r.Context(), id)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}
	if comment == nil {
		WriteNotFound(w, "comment not found")
		return
	}

	WriteSuccess(w, comment)
}

// GetReplies GET /api/politician-comments/{id}/replies - Get replies to a comment
func (h *PoliticianCommentHandler) GetReplies(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid comment ID")
		return
	}

	// Get current user ID if authenticated
	var currentUserID *uuid.UUID
	includeHidden := false
	claims := middleware.GetUserClaims(r.Context())
	if claims != nil {
		if userID, err := uuid.Parse(claims.UserID); err == nil {
			currentUserID = &userID
		}
		if claims.Role == "admin" && r.URL.Query().Get("include_hidden") == "true" {
			includeHidden = true
		}
	}

	replies, err := h.commentService.ListReplies(r.Context(), id, currentUserID, includeHidden)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, replies)
}

// UpdateComment PUT /api/politician-comments/{id} - Update a comment
func (h *PoliticianCommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid comment ID")
		return
	}

	// Get authenticated user
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		WriteUnauthorized(w, "authentication required")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		WriteUnauthorized(w, "invalid user ID")
		return
	}

	var req models.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if req.Content == "" {
		WriteBadRequest(w, "content is required")
		return
	}

	comment, err := h.commentService.UpdateComment(r.Context(), id, userID, &req)
	if err != nil {
		if err.Error() == "not authorized to edit this comment" {
			WriteForbidden(w, err.Error())
			return
		}
		if err.Error() == "comment not found" {
			WriteNotFound(w, err.Error())
			return
		}
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, comment)
}

// DeleteComment DELETE /api/politician-comments/{id} - Delete a comment
func (h *PoliticianCommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid comment ID")
		return
	}

	// Get authenticated user
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		WriteUnauthorized(w, "authentication required")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		WriteUnauthorized(w, "invalid user ID")
		return
	}

	isAdmin := claims.Role == "admin"

	if err := h.commentService.DeleteComment(r.Context(), id, userID, isAdmin); err != nil {
		if err.Error() == "not authorized to delete this comment" {
			WriteForbidden(w, err.Error())
			return
		}
		if err.Error() == "comment not found" {
			WriteNotFound(w, err.Error())
			return
		}
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "comment deleted"})
}

// AddReaction POST /api/politician-comments/{id}/reactions - Add a reaction
func (h *PoliticianCommentHandler) AddReaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid comment ID")
		return
	}

	// Get authenticated user
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		WriteUnauthorized(w, "authentication required")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		WriteUnauthorized(w, "invalid user ID")
		return
	}

	var req models.AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if err := h.commentService.AddReaction(r.Context(), id, userID, req.Reaction); err != nil {
		if err.Error() == "invalid reaction type" {
			WriteBadRequest(w, err.Error())
			return
		}
		if err.Error() == "comment not found" {
			WriteNotFound(w, err.Error())
			return
		}
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "reaction added"})
}

// RemoveReaction DELETE /api/politician-comments/{id}/reactions/{reaction} - Remove a reaction
func (h *PoliticianCommentHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid comment ID")
		return
	}

	reaction := chi.URLParam(r, "reaction")
	if reaction == "" {
		WriteBadRequest(w, "reaction type is required")
		return
	}

	// Get authenticated user
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		WriteUnauthorized(w, "authentication required")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		WriteUnauthorized(w, "invalid user ID")
		return
	}

	if err := h.commentService.RemoveReaction(r.Context(), id, userID, reaction); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "reaction removed"})
}

// GetCommentCount GET /api/politicians/{slug}/comments/count - Get comment count
func (h *PoliticianCommentHandler) GetCommentCount(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "politician slug is required")
		return
	}

	count, err := h.commentService.GetCommentCount(r.Context(), slug)
	if err != nil {
		WriteNotFound(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]int{"count": count})
}

// ModerateComment PUT /api/admin/politician-comments/{id}/moderate - Moderate a comment
func (h *PoliticianCommentHandler) ModerateComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid comment ID")
		return
	}

	// Get authenticated admin user
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		WriteUnauthorized(w, "authentication required")
		return
	}

	if claims.Role != "admin" {
		WriteForbidden(w, "admin access required")
		return
	}

	moderatorID, err := uuid.Parse(claims.UserID)
	if err != nil {
		WriteUnauthorized(w, "invalid user ID")
		return
	}

	var req models.ModerateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	// Validate status
	validStatuses := map[models.CommentStatus]bool{
		models.CommentStatusActive:      true,
		models.CommentStatusUnderReview: true,
		models.CommentStatusSpam:        true,
		models.CommentStatusHidden:      true,
	}
	if !validStatuses[req.Status] {
		WriteBadRequest(w, "invalid status: must be 'active', 'under_review', 'spam', or 'hidden'")
		return
	}

	comment, err := h.commentService.ModerateComment(r.Context(), id, moderatorID, &req)
	if err != nil {
		if err.Error() == "comment not found" {
			WriteNotFound(w, err.Error())
			return
		}
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, comment)
}
