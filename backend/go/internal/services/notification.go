package services

import (
	"fmt"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

// NotificationService handles notifications
type NotificationService struct{}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// Notification represents a notification
type Notification struct {
	ID        uint        `json:"id"`
	UserID    uint        `json:"user_id"`
	Type      string      `json:"type"`
	ActorID   *uint       `json:"actor_id"`
	PostID    *uint       `json:"post_id"`
	Actor     *UserBasic  `json:"actor"`
	Post      *PostBasic  `json:"post"`
	ReadAt    *time.Time  `json:"read_at"`
	CreatedAt time.Time   `json:"created_at"`
}

// UserBasic contains basic user info
type UserBasic struct {
	ID        uint    `json:"id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

// PostBasic contains basic post info
type PostBasic struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
}

// CreateNotificationInput contains notification creation data
type CreateNotificationInput struct {
	UserID  uint
	Type    string // follower, like, comment, repost, mention, dm_reply
	ActorID *uint
	PostID  *uint
}

// GetNotificationsInput contains pagination params
type GetNotificationsInput struct {
	UserID  uint
	Tab     string // all, unread
	Page    int
	PerPage int
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(input CreateNotificationInput) (*Notification, error) {
	if input.UserID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	if input.Type == "" {
		return nil, fmt.Errorf("notification type is required")
	}

	validTypes := map[string]bool{
		"follower": true,
		"like": true,
		"comment": true,
		"repost": true,
		"mention": true,
		"dm_reply": true,
	}
	if !validTypes[input.Type] {
		return nil, fmt.Errorf("invalid notification type")
	}

	notification := models.Notification{
		UserID:    input.UserID,
		Type:      input.Type,
		ActorID:   input.ActorID,
		PostID:    input.PostID,
		CreatedAt: time.Now(),
	}

	if err := db.GetDB().Create(&notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return s.getNotificationDetail(notification.ID)
}

// GetNotifications retrieves notifications for a user
func (s *NotificationService) GetNotifications(input GetNotificationsInput) ([]Notification, int, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 || input.PerPage > 50 {
		input.PerPage = 20
	}

	var notifications []models.Notification
	var total int64

	query := db.GetDB().Model(&models.Notification{}).Where("user_id = ?", input.UserID)
	if input.Tab == "unread" {
		query = query.Where("read_at IS NULL")
	}
	query.Count(&total)

	offset := (input.Page - 1) * input.PerPage
	query.Order("created_at DESC").
		Offset(offset).
		Limit(input.PerPage).
		Find(&notifications)

	var result []Notification
	for _, notif := range notifications {
		detail, err := s.getNotificationDetail(notif.ID)
		if err != nil {
			continue
		}
		result = append(result, *detail)
	}

	return result, int(total), nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID uint, userID uint) error {
	var notification models.Notification
	result := db.GetDB().First(&notification, notificationID)
	if result.Error != nil {
		return fmt.Errorf("notification not found")
	}

	if notification.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	now := time.Now()
	db.GetDB().Model(&notification).Update("read_at", &now)

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	result := db.GetDB().Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", time.Now())

	return result.Error
}

// GetUnreadCount returns the count of unread notifications
func (s *NotificationService) GetUnreadCount(userID uint) (int, error) {
	var count int64
	db.GetDB().Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count)
	return int(count), nil
}

// getNotificationDetail retrieves detailed notification info
func (s *NotificationService) getNotificationDetail(notificationID uint) (*Notification, error) {
	var notification models.Notification
	result := db.GetDB().Where("id = ?", notificationID).First(&notification)
	if result.Error != nil {
		return nil, result.Error
	}

	notif := &Notification{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		ActorID:   notification.ActorID,
		PostID:    notification.PostID,
		ReadAt:    notification.ReadAt,
		CreatedAt: notification.CreatedAt,
	}

	// Get actor info if exists
	if notification.ActorID != nil {
		var actor models.User
		db.GetDB().First(&actor, *notification.ActorID)
		notif.Actor = &UserBasic{
			ID:        actor.ID,
			Username:  actor.Username,
			AvatarURL: actor.AvatarURL,
		}
	}

	// Get post info if exists
	if notification.PostID != nil {
		var post models.Post
		db.GetDB().First(&post, *notification.PostID)
		notif.Post = &PostBasic{
			ID:      post.ID,
			Content: post.Content,
		}
	}

	return notif, nil
}

// TriggerNotification creates notifications for engagement actions
func (s *NotificationService) TriggerNotification(
	notificationType string,
	targetUserID uint,
	actorID uint,
	postID *uint,
) (*Notification, error) {
	// Don't notify the actor themselves
	if targetUserID == actorID {
		return nil, nil
	}

	input := CreateNotificationInput{
		UserID:  targetUserID,
		Type:    notificationType,
		ActorID: &actorID,
		PostID:  postID,
	}

	return s.CreateNotification(input)
}
