package services

import (
	"fmt"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

// ReadReceiptService handles message read receipts
type ReadReceiptService struct{}

// NewReadReceiptService creates a new read receipt service
func NewReadReceiptService() *ReadReceiptService {
	return &ReadReceiptService{}
}

// ReadReceipt represents a message read status
type ReadReceipt struct {
	ID             uint      `json:"id"`
	ConversationID uint      `json:"conversation_id"`
	MessageID      uint      `json:"message_id"`
	ViewerID       uint      `json:"viewer_id"`
	ReadAt         time.Time `json:"read_at"`
}

// GetUnreadCount returns unread message count for a conversation
func (s *ReadReceiptService) GetUnreadCount(conversationID uint, userID uint) (int, error) {
	var lastReadAt *time.Time
	result := db.GetDB().
		Table("dm_read_receipts").
		Where("conversation_id = ? AND viewer_id = ?", conversationID, userID).
		Order("read_at DESC").
		First(&lastReadAt)
	
	if result.Error != nil {
		// No read receipts yet, all messages are unread
		var count int64
		db.GetDB().Model(&models.DMMessage{}).
			Where("conversation_id = ? AND sender_id != ?", conversationID, userID).
			Count(&count)
		return int(count), nil
	}

	// Count messages sent after last read
	var count int64
	db.GetDB().Model(&models.DMMessage{}).
		Where("conversation_id = ? AND sender_id != ? AND created_at > ?", 
			conversationID, userID, lastReadAt).
			Count(&count)
	return int(count), nil
}

// MarkMessageRead marks a message as read
func (s *ReadReceiptService) MarkMessageRead(conversationID uint, messageID uint, viewerID uint) error {
	// Verify user is participant
	var participantCount int64
	db.GetDB().Model(&models.DMParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, viewerID).
		Count(&participantCount)
	if participantCount == 0 {
		return fmt.Errorf("not a participant of this conversation")
	}

	// Check if already read
	var existing ReadReceipt
	result := db.GetDB().
		Where("conversation_id = ? AND message_id = ? AND viewer_id = ?", 
			conversationID, messageID, viewerID).
		First(&existing)
	if result.RowsAffected > 0 {
		return nil // Already read
	}

	receipt := ReadReceipt{
		ConversationID: conversationID,
		MessageID:      messageID,
		ViewerID:       viewerID,
		ReadAt:         time.Now(),
	}

	return db.GetDB().Create(&receipt).Error
}

// MarkConversationRead marks all messages in a conversation as read
func (s *ReadReceiptService) MarkConversationRead(conversationID uint, userID uint) error {
	// Get all unread messages
	var messages []models.DMMessage
	db.GetDB().Where("conversation_id = ?", conversationID).Find(&messages)

	// Create read receipts for each message
	for _, msg := range messages {
		_ = s.MarkMessageRead(conversationID, msg.ID, userID)
	}

	return nil
}

// GetReadReceipts returns read receipts for a message
func (s *ReadReceiptService) GetReadReceipts(messageID uint) ([]ReadReceipt, error) {
	var receipts []ReadReceipt
	db.GetDB().
		Where("message_id = ?", messageID).
		Order("read_at ASC").
		Find(&receipts)
	return receipts, nil
}

// GetLastReadAt returns the last time a user read messages in a conversation
func (s *ReadReceiptService) GetLastReadAt(conversationID uint, userID uint) (*time.Time, error) {
	var lastReadAt time.Time
	result := db.GetDB().
		Table("dm_read_receipts").
		Where("conversation_id = ? AND viewer_id = ?", conversationID, userID).
		Order("read_at DESC").
		First(&lastReadAt)
	if result.Error != nil {
		return nil, result.Error
	}
	return &lastReadAt, nil
}
