package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

// DMHandler handles direct messaging
type DMHandler struct {
	dmService *services.DMService
	wsClients map[string]*services.DMConnection
}

// NewDMHandler creates a new DM handler
func NewDMHandler() *DMHandler {
	return &DMHandler{
		dmService: services.NewDMService(),
		wsClients: make(map[string]*services.DMConnection),
	}
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

	return c.JSON(fiber.Map{
		"success": true,
		"data": messages,
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

		h.wsClients[dmConn.ID] = dmConn

		// Handle incoming messages
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				delete(h.wsClients, dmConn.ID)
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
		Type string `json:"type"`
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
		select {
		case client.Send <- []byte(msg):
		default:
			// Client buffer full, skip
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
