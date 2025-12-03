package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// TODO: Restrict in production
		return true
	},
}

// Client represents a connected WebSocket client
type Client struct {
	ID             string
	UserID         uuid.UUID
	IsAdmin        bool
	Conn           *websocket.Conn
	Send           chan []byte
	Hub            *Hub
	ConversationID *uuid.UUID // Currently viewing conversation
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	// Registered clients by user ID
	clients map[uuid.UUID]*Client

	// Admin clients (for broadcasting to all admins)
	admins map[uuid.UUID]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast to specific user
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// BroadcastMessage represents a message to be sent to specific users
type BroadcastMessage struct {
	UserIDs []uuid.UUID
	Message []byte
	ToAdmin bool // If true, send to all admins
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		admins:     make(map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			if client.IsAdmin {
				h.admins[client.UserID] = client
			}
			h.mu.Unlock()

			log.Info().
				Str("user_id", client.UserID.String()).
				Bool("is_admin", client.IsAdmin).
				Msg("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				delete(h.admins, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()

			log.Info().
				Str("user_id", client.UserID.String()).
				Msg("WebSocket client disconnected")

		case msg := <-h.broadcast:
			h.mu.RLock()
			if msg.ToAdmin {
				// Send to all admins
				for _, client := range h.admins {
					select {
					case client.Send <- msg.Message:
					default:
						// Client's buffer is full, skip
					}
				}
			} else {
				// Send to specific users
				for _, userID := range msg.UserIDs {
					if client, ok := h.clients[userID]; ok {
						select {
						case client.Send <- msg.Message:
						default:
							// Client's buffer is full, skip
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID uuid.UUID, msg *models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	h.broadcast <- &BroadcastMessage{
		UserIDs: []uuid.UUID{userID},
		Message: data,
	}
}

// BroadcastToAdmins sends a message to all connected admins
func (h *Hub) BroadcastToAdmins(msg *models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	h.broadcast <- &BroadcastMessage{
		ToAdmin: true,
		Message: data,
	}
}

// BroadcastNewMessage broadcasts a new message to relevant parties
func (h *Hub) BroadcastNewMessage(message *models.Message, conversationUserID uuid.UUID, senderIsAdmin bool) {
	wsMsg := &models.WSMessage{
		Type:           models.WSMessageTypeNewMessage,
		ConversationID: &message.ConversationID,
		Message:        message,
		Timestamp:      time.Now(),
	}

	if senderIsAdmin {
		// Admin sent message, notify the user
		h.BroadcastToUser(conversationUserID, wsMsg)
	} else {
		// User sent message, notify all admins
		h.BroadcastToAdmins(wsMsg)
	}
}

// IsUserOnline checks if a user is currently connected
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub            *Hub
	authService    *services.AuthService
	messageService *services.MessageService
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(hub *Hub, authService *services.AuthService, messageService *services.MessageService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:            hub,
		authService:    authService,
		messageService: messageService,
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter (WebSocket doesn't support custom headers easily)
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	// Validate token
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade to WebSocket")
		return
	}

	isAdmin := claims.Role == "admin"

	client := &Client{
		ID:      uuid.New().String(),
		UserID:  userID,
		IsAdmin: isAdmin,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Hub:     h.hub,
	}

	h.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(h)
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump(h *WebSocketHandler) {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("WebSocket read error")
			}
			break
		}

		// Parse incoming message
		var wsMsg models.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Error().Err(err).Msg("Failed to parse WebSocket message")
			continue
		}

		// Handle different message types
		switch wsMsg.Type {
		case models.WSMessageTypeTyping, models.WSMessageTypeStopTyping:
			// Broadcast typing indicator to the other party
			if wsMsg.ConversationID != nil {
				h.handleTypingIndicator(c, &wsMsg)
			}
		case models.WSMessageTypeMessageRead:
			// Mark messages as read
			if wsMsg.ConversationID != nil {
				h.handleMarkAsRead(c, wsMsg.ConversationID)
			}
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleTypingIndicator broadcasts typing status to the other party
func (h *WebSocketHandler) handleTypingIndicator(c *Client, wsMsg *models.WSMessage) {
	ctx := context.Background()

	// Get conversation to find the other party
	conversation, err := h.messageService.GetConversation(ctx, *wsMsg.ConversationID)
	if err != nil || conversation == nil {
		return
	}

	broadcastMsg := &models.WSMessage{
		Type:           wsMsg.Type,
		ConversationID: wsMsg.ConversationID,
		UserID:         &c.UserID,
		Timestamp:      time.Now(),
	}

	if c.IsAdmin {
		// Admin typing, notify user
		h.hub.BroadcastToUser(conversation.UserID, broadcastMsg)
	} else {
		// User typing, notify admins
		h.hub.BroadcastToAdmins(broadcastMsg)
	}
}

// handleMarkAsRead marks messages as read
func (h *WebSocketHandler) handleMarkAsRead(c *Client, conversationID *uuid.UUID) {
	ctx := context.Background()
	err := h.messageService.MarkAsRead(ctx, *conversationID, c.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to mark messages as read")
	}
}
