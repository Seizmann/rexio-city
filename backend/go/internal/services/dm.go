package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

// DMService handles direct messaging
type DMService struct{}

// NewDMService creates a new DM service
func NewDMService() *DMService {
	return &DMService{}
}

// DMConnection represents a WebSocket connection
type DMConnection struct {
	ID     string
	UserID uint
	Conn   *websocket.Conn
	Send   chan []byte
}

// Conversation represents a DM conversation
type Conversation struct {
	ID             uint             `json:"id"`
	Participants   []Participant    `json:"participants"`
	LatestMessage  *DMMessage       `json:"latest_message"`
	CreatedAt      time.Time        `json:"created_at"`
}

// Participant represents a user in a conversation
type Participant struct {
	UserID    uint    `json:"user_id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

// DMMessage represents a decrypted DM message
type DMMessage struct {
	ID             uint      `json:"id"`
	ConversationID uint      `json:"conversation_id"`
	SenderID       uint      `json:"sender_id"`
	SenderName     string    `json:"sender_name"`
	Content        string    `json:"content"`
	EncryptedData  string    `json:"encrypted_data"`
	IV             string    `json:"iv"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateConversationInput contains conversation creation data
type CreateConversationInput struct {
	UserID           uint
	ParticipantIDs []uint
}

// SendMessageInput contains message sending data
type SendMessageInput struct {
	ConversationID uint
	SenderID       uint
	EncryptedData  string
	IV             string
}

// GetConversationsInput contains pagination params
type GetConversationsInput struct {
	UserID  uint
	Page    int
	PerPage int
}

// GetMessagesInput contains message retrieval params
type GetMessagesInput struct {
	ConversationID uint
	UserID         uint
	Page           int
	PerPage        int
}

// NewConversation creates a new DM conversation
func (s *DMService) NewConversation(input CreateConversationInput) (*Conversation, error) {
	if len(input.ParticipantIDs) < 1 {
		return nil, fmt.Errorf("at least one participant required")
	}

	// Check if conversation already exists
	var existing models.DMConversation
	result := db.GetDB().
		Joins("INNER JOIN dm_participants ON dm_conversations.id = dm_participants.conversation_id").
		Where("dm_participants.user_id = ?", input.UserID).
		First(&existing)
	if result.Error == nil {
		var participantCount int64
		db.GetDB().Model(&models.DMParticipant{}).
			Where("conversation_id = ?", existing.ID).
			Count(&participantCount)

		if int(participantCount) == len(input.ParticipantIDs)+1 {
			return s.getConversationDetail(existing.ID)
		}
	}

	// Create new conversation
	conversation := models.DMConversation{
		CreatedAt: time.Now(),
	}
	if err := db.GetDB().Create(&conversation).Error; err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Add participants
	allParticipantIDs := append([]uint{input.UserID}, input.ParticipantIDs...)
	for _, userID := range allParticipantIDs {
		participant := models.DMParticipant{
			ConversationID: conversation.ID,
			UserID:         userID,
		}
		if err := db.GetDB().Create(&participant).Error; err != nil {
			return nil, fmt.Errorf("failed to add participant: %w", err)
		}
	}

	return s.getConversationDetail(conversation.ID)
}

// SendMessage sends a message in a conversation
func (s *DMService) SendMessage(input SendMessageInput) (*DMMessage, error) {
	// Validate conversation exists
	var conversation models.DMConversation
	result := db.GetDB().First(&conversation, input.ConversationID)
	if result.Error != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Verify user is participant
	var participantCount int64
	db.GetDB().Model(&models.DMParticipant{}).
		Where("conversation_id = ? AND user_id = ?", input.ConversationID, input.SenderID).
		Count(&participantCount)
	if participantCount == 0 {
		return nil, fmt.Errorf("not a participant of this conversation")
	}

	// Create message
	message := models.DMMessage{
		ConversationID: input.ConversationID,
		SenderID:       input.SenderID,
		EncryptedData:  []byte(input.EncryptedData),
		IV:             []byte(input.IV),
		CreatedAt:      time.Now(),
	}

	if err := db.GetDB().Create(&message).Error; err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Get sender info
	var sender models.User
	db.GetDB().First(&sender, input.SenderID)

	return &DMMessage{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		SenderName:     sender.Username,
		EncryptedData:  input.EncryptedData,
		IV:             input.IV,
		CreatedAt:      message.CreatedAt,
	}, nil
}

// GetConversations retrieves conversations for a user
func (s *DMService) GetConversations(input GetConversationsInput) ([]Conversation, int, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 || input.PerPage > 50 {
		input.PerPage = 20
	}

	var conversations []models.DMConversation
	var total int64

	db.GetDB().
		Joins("INNER JOIN dm_participants ON dm_conversations.id = dm_participants.conversation_id").
		Where("dm_participants.user_id = ?", input.UserID).
		Order("dm_conversations.created_at DESC").
		Count(&total).
		Offset((input.Page - 1) * input.PerPage).
		Limit(input.PerPage).
		Find(&conversations)

	var result []Conversation
	for _, conv := range conversations {
		detail, err := s.getConversationDetail(conv.ID)
		if err != nil {
			continue
		}
		result = append(result, *detail)
	}

	return result, int(total), nil
}

// GetMessages retrieves messages in a conversation
func (s *DMService) GetMessages(input GetMessagesInput) ([]DMMessage, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 || input.PerPage > 100 {
		input.PerPage = 50
	}

	var messages []models.DMMessage
	db.GetDB().
		Where("conversation_id = ?", input.ConversationID).
		Order("created_at ASC").
		Offset((input.Page - 1) * input.PerPage).
		Limit(input.PerPage).
		Find(&messages)

	var result []DMMessage
	for _, msg := range messages {
		var sender models.User
		db.GetDB().First(&sender, msg.SenderID)

		result = append(result, DMMessage{
			ID:             msg.ID,
			ConversationID: msg.ConversationID,
			SenderID:       msg.SenderID,
			SenderName:     sender.Username,
			EncryptedData:  base64.StdEncoding.EncodeToString(msg.EncryptedData),
			IV:             base64.StdEncoding.EncodeToString(msg.IV),
			CreatedAt:      msg.CreatedAt,
		})
	}

	return result, nil
}

// getConversationDetail retrieves detailed conversation info
func (s *DMService) getConversationDetail(conversationID uint) (*Conversation, error) {
	var conversation models.DMConversation
	result := db.GetDB().First(&conversation, conversationID)
	if result.Error != nil {
		return nil, result.Error
	}

	// Get participants
	var participants []models.DMParticipant
	db.GetDB().Where("conversation_id = ?", conversationID).Find(&participants)

	var participantList []Participant
	for _, p := range participants {
		var user models.User
		db.GetDB().First(&user, p.UserID)
		participantList = append(participantList, Participant{
			UserID:    user.ID,
			Username:  user.Username,
			AvatarURL: user.AvatarURL,
		})
	}

	// Get latest message
	var latestMessage *DMMessage
	var lastMsg models.DMMessage
	db.GetDB().Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		First(&lastMsg)
	if lastMsg.ID > 0 {
		var sender models.User
		db.GetDB().First(&sender, lastMsg.SenderID)
		latestMessage = &DMMessage{
			ID:             lastMsg.ID,
			SenderID:       lastMsg.SenderID,
			SenderName:     sender.Username,
			EncryptedData:  base64.StdEncoding.EncodeToString(lastMsg.EncryptedData),
			IV:             base64.StdEncoding.EncodeToString(lastMsg.IV),
			CreatedAt:      lastMsg.CreatedAt,
		}
	}

	return &Conversation{
		ID:            conversation.ID,
		Participants:  participantList,
		LatestMessage: latestMessage,
		CreatedAt:     conversation.CreatedAt,
	}, nil
}

// EncryptMessage encrypts a message using AES-256-GCM
func EncryptMessage(plaintext string, key []byte) (string, string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), base64.StdEncoding.EncodeToString(nonce), nil
}

// DecryptMessage decrypts a message using AES-256-GCM
func DecryptMessage(encryptedData string, iv string, key []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// GenerateDMKey generates a unique key for a conversation
func GenerateDMKey(userID1, userID2 uint) string {
	if userID1 < userID2 {
		return fmt.Sprintf("dm_key_%d_%d", userID1, userID2)
	}
	return fmt.Sprintf("dm_key_%d_%d", userID2, userID1)
}

// GetUserDMKey generates a DM key for user pairing
func GetUserDMKey(userID uint, partnerID uint) string {
	return GenerateDMKey(userID, partnerID)
}
