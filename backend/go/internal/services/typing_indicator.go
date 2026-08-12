package services

import (
	"fmt"
	"sync"
	"time"
)

// TypingIndicatorService handles typing indicators
type TypingIndicatorService struct {
	mu             sync.RWMutex
	typingUsers    map[string]map[uint]time.Time // conversationID -> userID -> timestamp
	broadcastFunc  func(conversationID uint, userID uint, isTyping bool)
}

// NewTypingIndicatorService creates a new typing indicator service
func NewTypingIndicatorService() *TypingIndicatorService {
	return &TypingIndicatorService{
		typingUsers: make(map[string]map[uint]time.Time),
	}
}

// SetBroadcastFunc sets the function to broadcast typing indicators
func (s *TypingIndicatorService) SetBroadcastFunc(fn func(conversationID uint, userID uint, isTyping bool)) {
	s.broadcastFunc = fn
}

// StartTyping marks a user as typing in a conversation
func (s *TypingIndicatorService) StartTyping(conversationID uint, userID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	convKey := fmt.Sprintf("%d", conversationID)
	if s.typingUsers[convKey] == nil {
		s.typingUsers[convKey] = make(map[uint]time.Time)
	}
	s.typingUsers[convKey][userID] = time.Now()

	// Broadcast to other participants
	if s.broadcastFunc != nil {
		s.broadcastFunc(conversationID, userID, true)
	}

	// Auto-clear after 5 seconds
	go s.clearTypingAfterTimeout(conversationID, userID)
}

// StopTyping marks a user as stopped typing
func (s *TypingIndicatorService) StopTyping(conversationID uint, userID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	convKey := fmt.Sprintf("%d", conversationID)
	delete(s.typingUsers[convKey], userID)

	// Broadcast to other participants
	if s.broadcastFunc != nil {
		s.broadcastFunc(conversationID, userID, false)
	}
}

// IsTyping returns whether a user is currently typing
func (s *TypingIndicatorService) IsTyping(conversationID uint, userID uint) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	convKey := fmt.Sprintf("%d", conversationID)
	users, ok := s.typingUsers[convKey]
	if !ok {
		return false
	}

	timestamp, ok := users[userID]
	if !ok {
		return false
	}

	// Check if within timeout (5 seconds)
	return time.Since(timestamp) < 5*time.Second
}

// GetTypingUsers returns list of users currently typing in a conversation
func (s *TypingIndicatorService) GetTypingUsers(conversationID uint) []uint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	convKey := fmt.Sprintf("%d", conversationID)
	users, ok := s.typingUsers[convKey]
	if !ok {
		return []uint{}
	}

	var typing []uint
	for userID, timestamp := range users {
		if time.Since(timestamp) < 5*time.Second {
			typing = append(typing, userID)
		}
	}
	return typing
}

// clearTypingAfterTimeout automatically clears typing status after timeout
func (s *TypingIndicatorService) clearTypingAfterTimeout(conversationID uint, userID uint) {
	time.Sleep(5 * time.Second)
	s.StopTyping(conversationID, userID)
}

// CleanupExpiredTyping removes expired typing indicators
func (s *TypingIndicatorService) CleanupExpiredTyping() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for convKey, users := range s.typingUsers {
		for userID, timestamp := range users {
			if now.Sub(timestamp) > 5*time.Second {
				delete(users, userID)
			}
		}
		if len(users) == 0 {
			delete(s.typingUsers, convKey)
		}
	}
}

// GetTypingStatus returns the typing status for a conversation
func (s *TypingIndicatorService) GetTypingStatus(conversationID uint, excludeUserID uint) []TypingUser {
	s.mu.RLock()
	defer s.mu.RUnlock()

	convKey := fmt.Sprintf("%d", conversationID)
	users, ok := s.typingUsers[convKey]
	if !ok {
		return []TypingUser{}
	}

	var result []TypingUser
	for userID, timestamp := range users {
		if userID != excludeUserID && time.Since(timestamp) < 5*time.Second {
			result = append(result, TypingUser{
				UserID:    userID,
				StartedAt: timestamp,
			})
		}
	}
	return result
}

// TypingUser represents a user who is typing
type TypingUser struct {
	UserID    uint      `json:"user_id"`
	StartedAt time.Time `json:"started_at"`
}

// Error types
var (
	ErrInvalidConversationID = fmt.Errorf("invalid conversation ID")
	ErrInvalidUserID         = fmt.Errorf("invalid user ID")
)
