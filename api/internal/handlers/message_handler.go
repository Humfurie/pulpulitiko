package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/middleware"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type MessageHandler struct {
	service *services.MessageService
	hub     *Hub
}

func NewMessageHandler(service *services.MessageService, hub *Hub) *MessageHandler {
	return &MessageHandler{
		service: service,
		hub:     hub,
	}
}

// CreateConversation creates a new conversation with an initial message
// POST /api/messages/conversations
func (h *MessageHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.CreateConversationRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	conversation, message, err := h.service.CreateConversation(r.Context(), userID, &req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Broadcast to admins that a new message arrived
	h.hub.BroadcastNewMessage(message, userID, false)

	WriteCreated(w, map[string]interface{}{
		"conversation": conversation,
		"message":      message,
	})
}

// GetMyConversations gets the current user's conversations
// GET /api/messages/conversations
func (h *MessageHandler) GetMyConversations(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	conversations, err := h.service.GetUserConversations(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, conversations)
}

// GetConversation retrieves a specific conversation
// GET /api/messages/conversations/{id}
func (h *MessageHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conversationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	userID, _ := uuid.Parse(claims.UserID)
	isAdmin := claims.Role == "admin"

	// Check access
	canAccess, err := h.service.CanAccessConversation(r.Context(), conversationID, userID, isAdmin)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !canAccess {
		WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	conversation, err := h.service.GetConversation(r.Context(), conversationID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if conversation == nil {
		WriteError(w, http.StatusNotFound, "Conversation not found")
		return
	}

	WriteSuccess(w, conversation)
}

// GetMessages retrieves messages in a conversation
// GET /api/messages/conversations/{id}/messages
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conversationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	userID, _ := uuid.Parse(claims.UserID)
	isAdmin := claims.Role == "admin"

	// Check access
	canAccess, err := h.service.CanAccessConversation(r.Context(), conversationID, userID, isAdmin)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !canAccess {
		WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	page, perPage := GetPaginationParams(r)
	messages, err := h.service.GetMessages(r.Context(), conversationID, page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, messages)
}

// SendMessage sends a message in a conversation
// POST /api/messages/conversations/{id}/messages
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conversationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	userID, _ := uuid.Parse(claims.UserID)
	isAdmin := claims.Role == "admin"

	// Check access
	canAccess, err := h.service.CanAccessConversation(r.Context(), conversationID, userID, isAdmin)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !canAccess {
		WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req models.CreateMessageRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	message, err := h.service.SendMessage(r.Context(), conversationID, userID, &req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get conversation to know who to notify
	conversation, _ := h.service.GetConversation(r.Context(), conversationID)
	if conversation != nil {
		h.hub.BroadcastNewMessage(message, conversation.UserID, isAdmin)
	}

	WriteCreated(w, message)
}

// MarkAsRead marks all messages in a conversation as read
// POST /api/messages/conversations/{id}/read
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conversationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	userID, _ := uuid.Parse(claims.UserID)
	isAdmin := claims.Role == "admin"

	// Check access
	canAccess, err := h.service.CanAccessConversation(r.Context(), conversationID, userID, isAdmin)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !canAccess {
		WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	err = h.service.MarkAsRead(r.Context(), conversationID, userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, map[string]bool{"success": true})
}

// GetUnreadCounts gets unread message counts for the current user
// GET /api/messages/unread
func (h *MessageHandler) GetUnreadCounts(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, _ := uuid.Parse(claims.UserID)
	isAdmin := claims.Role == "admin"

	counts, err := h.service.GetUnreadCounts(r.Context(), userID, isAdmin)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, counts)
}

// ===== Admin Endpoints =====

// AdminListConversations lists all conversations (admin only)
// GET /api/admin/messages/conversations
func (h *MessageHandler) AdminListConversations(w http.ResponseWriter, r *http.Request) {
	page, perPage := GetPaginationParams(r)

	// Parse filter from query params
	var filter *models.ConversationFilter
	statusParam := r.URL.Query().Get("status")
	if statusParam != "" {
		status := models.ConversationStatus(statusParam)
		filter = &models.ConversationFilter{Status: &status}
	}

	conversations, err := h.service.ListConversations(r.Context(), filter, page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, conversations)
}

// AdminUpdateConversationStatus updates a conversation's status
// PATCH /api/admin/messages/conversations/{id}/status
func (h *MessageHandler) AdminUpdateConversationStatus(w http.ResponseWriter, r *http.Request) {
	conversationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	var req models.UpdateConversationRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.UpdateConversationStatus(r.Context(), conversationID, req.Status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, map[string]bool{"success": true})
}
