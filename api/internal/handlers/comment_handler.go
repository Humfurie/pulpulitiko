package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/middleware"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type CommentHandler struct {
	commentService *services.CommentService
}

func NewCommentHandler(commentService *services.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// ListComments GET /api/articles/{slug}/comments - List comments for an article
func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "article slug is required")
		return
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

	comments, err := h.commentService.ListArticleComments(r.Context(), slug, currentUserID, includeHidden)
	if err != nil {
		WriteNotFound(w, err.Error())
		return
	}

	WriteSuccess(w, comments)
}

// CreateComment POST /api/articles/{slug}/comments - Create a new comment
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "article slug is required")
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
		// Check for foreign key violation (user doesn't exist)
		if strings.Contains(errMsg, "comments_user_id_fkey") {
			WriteUnauthorized(w, "user session invalid - please log out and log in again")
			return
		}
		WriteInternalError(w, errMsg)
		return
	}

	WriteSuccessWithStatus(w, http.StatusCreated, comment)
}

// GetComment GET /api/comments/{id} - Get a single comment
func (h *CommentHandler) GetComment(w http.ResponseWriter, r *http.Request) {
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

// GetReplies GET /api/comments/{id}/replies - Get replies to a comment
func (h *CommentHandler) GetReplies(w http.ResponseWriter, r *http.Request) {
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
		// Admins can see all replies if they request it
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

// UpdateComment PUT /api/comments/{id} - Update a comment
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
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

// DeleteComment DELETE /api/comments/{id} - Delete a comment
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
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

// AddReaction POST /api/comments/{id}/reactions - Add a reaction to a comment
func (h *CommentHandler) AddReaction(w http.ResponseWriter, r *http.Request) {
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

// RemoveReaction DELETE /api/comments/{id}/reactions/{reaction} - Remove a reaction
func (h *CommentHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
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

// GetCommentCount GET /api/articles/{slug}/comments/count - Get comment count for article
func (h *CommentHandler) GetCommentCount(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "article slug is required")
		return
	}

	count, err := h.commentService.GetCommentCount(r.Context(), slug)
	if err != nil {
		WriteNotFound(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]int{"count": count})
}

// ModerateComment PUT /api/admin/comments/{id}/moderate - Moderate a comment (admin only)
func (h *CommentHandler) ModerateComment(w http.ResponseWriter, r *http.Request) {
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

// ListAllComments GET /api/admin/comments - List all comments for moderation (admin only)
func (h *CommentHandler) ListAllComments(w http.ResponseWriter, r *http.Request) {
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

	var currentUserID *uuid.UUID
	if userID, err := uuid.Parse(claims.UserID); err == nil {
		currentUserID = &userID
	}

	// Optional status filter
	var filter *models.CommentFilter
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		status := models.CommentStatus(statusParam)
		filter = &models.CommentFilter{
			Status: &status,
		}
	}

	comments, err := h.commentService.ListAllComments(r.Context(), filter, currentUserID)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, comments)
}
