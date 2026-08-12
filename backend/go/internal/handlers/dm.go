package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

// DMHandler handles direct messaging
type DMHandler struct {
	dmService          *services.DMService
	readReceiptService *services.ReadReceiptService
	typingIndicator    *services.TypingIndicatorService
	wsClients          map[string]*services.DMConnection
	mu                 sync.RWMutex
}

// NewDMHandler creates a new DM handler
func NewDMHandler() *DMHandler {
	handler := &DMHandler{
		dmService:          services.NewDMService(),
		readReceiptService: services.NewReadReceiptService(),
		typingIndicator:    services.NewTypingIndicatorService(),
		wsClients:          make(map[string]*services.DMConnection),
	}

	// Set up broadcast function for typing indicators
	handler.typingIndicator.SetBroadcastFunc(handler.broadcastTypingIndicator)

	return handler
}

// GetConversations handles GET /api/dm/conversations
func (h *DMHandler) GetConversations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	conversations, total, err := h.dmService.GetConversations(services.GetConversationsInput{
		UserID:  userID,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": conversations,
		"meta": fiber.Map{
			"page": page,
			"per_page": perPage,
			"total": total,
		},
	})
}

// GetMessages handles GET /api/dm/conversations/:id/messages
func (h *DMHandler) GetMessages(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	conversationID := c.Params("id")

	var convID uint
	if _, err := fmt.Sscanf(conversationID, "%d", &convID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid conversation ID"},
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "50"))

	messages, err := h.dmService.GetMessages(services.GetMessagesInput{
		ConversationID: convID,
		UserID:         userID,
		Page:           page,
		PerPage:        perPage,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	// Get unread count
	unreadCount, _ := h.readReceiptService.GetUnreadCount(convID, userID)

	return c.JSON(fiber.Map{
		"success": true,
		"data": messages,
		"meta": fiber.Map{
			"unread_count": unreadCount,
		},
	})
}

// SendMessage handles POST /api/dm/conversations/:id/messages
func (h *DMHandler) SendMessage(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	conversationID := c.Params("id")

	var convID uint
	if _, err := fmt.Sscanf(conversationID, "%d", &convID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid conversation ID"},
		})
	}

	type request struct {
		EncryptedData string `json:"encrypted_data"`
		IV            string `json:"iv"`
	}
	var input request
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	message, err := h.dmService.SendMessage(services.SendMessageInput{
		ConversationID: convID,
		SenderID:       userID,
		EncryptedData:  input.EncryptedData,
		IV:             input.IV,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	// Broadcast to other participants via WebSocket
	h.broadcastToConversation(convID, message)

	// Trigger notification for DM
	_ = services.NewEventService().OnDMReceived(userID, convID, convID)

	return c.JSON(fiber.Map{
		"success": true,
		"data": message,
	})
}

// ConnectWS handles WebSocket connection for real-time DMs
func (h *DMHandler) ConnectWS(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	
	// Use websocket.New for Fiber integration
	return websocket.New(func(c *websocket.Conn) {
		// Create connection
		dmConn := &services.DMConnection{
			ID:     fmt.Sprintf("ws_%d", userID),
			UserID: userID,
			Conn:   c,
			Send:   make(chan []byte, 256),
		}

		h.mu.Lock()
		h.wsClients[dmConn.ID] = dmConn
		h.mu.Unlock()

		// Send welcome message
		c.WriteMessage(websocket.TextMessage, []byte(`{"type":"connected","user_id":`+fmt.Sprintf("%d", userID)+`}`))

		// Handle incoming messages
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				h.mu.Lock()
				delete(h.wsClients, dmConn.ID)
				h.mu.Unlock()
				c.Close()
				break
			}
			// Process message
			h.handleWSMessage(dmConn, message)
		}
	}, websocket.Config{})(c)
}

// handleWSMessage handles incoming WebSocket messages
func (h *DMHandler) handleWSMessage(conn *services.DMConnection, message []byte) {
	// Parse message
	type wsMessage struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}
	var msg wsMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	// Handle different message types
	switch msg.Type {
	case "ping":
		conn.Send <- []byte(`{"type":"pong"}`)
	case "typing_start":
		// User started typing
		if convID, ok := msg.Data["conversation_id"].(float64); ok {
			h.typingIndicator.StartTyping(uint(convID), conn.UserID)
		}
	case "typing_stop":
		// User stopped typing
		if convID, ok := msg.Data["conversation_id"].(float64); ok {
			h.typingIndicator.StopTyping(uint(convID), conn.UserID)
		}
	case "read_receipt":
		// Message read
		if convID, ok := msg.Data["conversation_id"].(float64); ok {
			if msgID, ok := msg.Data["message_id"].(float64); ok {
				h.readReceiptService.MarkMessageRead(uint(convID), uint(msgID), conn.UserID)
			}
		}
	case "chat_message":
		// Handle incoming chat message
		if convID, ok := msg.Data["conversation_id"].(float64); ok {
			// Broadcast to other participants
			h.broadcastToConversation(uint(convID), nil)
		}
	}
}

// broadcastToConversation broadcasts a message to all participants
func (h *DMHandler) broadcastToConversation(conversationID uint, message *services.DMMessage) {
	// In real implementation, track which users are in which conversations
	// For now, broadcast to all connected clients
	for _, client := range h.wsClients {
		msg := fmt.Sprintf(`{"type":"new_message","conversation_id":%d}`, conversationID)
		if message != nil {
			msg = fmt.Sprintf(`{"type":"new_message","conversation_id":%d,"message":%s}`, 
				conversationID, messageJSON(message))
		}
		select {
		case client.Send <- []byte(msg):
		default:
			// Client buffer full, skip
		}
	}
}

// broadcastTypingIndicator broadcasts typing indicators
func (h *DMHandler) broadcastTypingIndicator(conversationID uint, userID uint, isTyping bool) {
	msgType := "typing_stop"
	if isTyping {
		msgType = "typing_start"
	}
	
	for _, client := range h.wsClients {
		if client.UserID != userID {
			payload := fmt.Sprintf(`{"type":"%s","user_id":%d,"conversation_id":%d}`,
				msgType, userID, conversationID)
			select {
			case client.Send <- []byte(payload):
			default:
				// Client buffer full, skip
			}
		}
	}
}

// CreateConversation handles POST /api/dm/conversations
func (h *DMHandler) CreateConversation(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	type request struct {
		ParticipantIDs []uint `json:"participant_ids"`
	}
	var input request
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	conversation, err := h.dmService.NewConversation(services.CreateConversationInput{
		UserID:           userID,
		ParticipantIDs:   input.ParticipantIDs,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": conversation,
	})
}

// GetUnreadCount handles GET /api/dm/conversations/:id/unread-count
func (h *DMHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	conversationID := c.Params("id")

	var convID uint
	if _, err := fmt.Sscanf(conversationID, "%d", &convID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid conversation ID"},
		})
	}

	count, err := h.readReceiptService.GetUnreadCount(convID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{"unread_count": count},
	})
}

// MarkConversationRead handles PUT /api/dm/conversations/:id/read
func (h *DMHandler) MarkConversationRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	conversationID := c.Params("id")

	var convID uint
	if _, err := fmt.Sscanf(conversationID, "%d", &convID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid conversation ID"},
		})
	}

	err := h.readReceiptService.MarkConversationRead(convID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{"message": "Conversation marked as read"},
	})
}

// GetTypingUsers handles GET /api/dm/conversations/:id/typing
func (h *DMHandler) GetTypingUsers(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	conversationID := c.Params("id")

	var convID uint
	if _, err := fmt.Sscanf(conversationID, "%d", &convID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid conversation ID"},
		})
	}

	typingUsers := h.typingIndicator.GetTypingStatus(convID, userID)

	return c.JSON(fiber.Map{
		"success": true,
		"data": typingUsers,
	})
}

// messageJSON converts a DMMessage to JSON
func messageJSON(msg *services.DMMessage) string {
	data, _ := json.Marshal(msg)
	return string(data)
}
