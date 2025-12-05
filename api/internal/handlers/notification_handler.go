package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/middleware"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// ListNotifications GET /api/notifications - List notifications for authenticated user
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
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

	// Get pagination params
	page := 1
	perPage := 20

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

	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	result, err := h.notificationService.ListNotifications(r.Context(), userID, page, perPage, unreadOnly)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, result)
}

// GetUnreadCount GET /api/notifications/unread-count - Get unread notification count
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
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

	count, err := h.notificationService.GetUnreadCount(r.Context(), userID)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]int{"count": count})
}

// MarkAsRead POST /api/notifications/{id}/read - Mark a notification as read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
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

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid notification ID")
		return
	}

	if err := h.notificationService.MarkAsRead(r.Context(), id, userID); err != nil {
		if err.Error() == "notification not found" {
			WriteNotFound(w, err.Error())
			return
		}
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "notification marked as read"})
}

// MarkAllAsRead POST /api/notifications/read-all - Mark all notifications as read
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
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

	if err := h.notificationService.MarkAllAsRead(r.Context(), userID); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "all notifications marked as read"})
}

// DeleteNotification DELETE /api/notifications/{id} - Delete a notification
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
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

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid notification ID")
		return
	}

	if err := h.notificationService.DeleteNotification(r.Context(), id, userID); err != nil {
		if err.Error() == "notification not found" {
			WriteNotFound(w, err.Error())
			return
		}
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "notification deleted"})
}
