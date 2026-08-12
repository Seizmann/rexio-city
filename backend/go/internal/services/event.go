package services

import (
	"fmt"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

// EventService handles event triggering for notifications
type EventService struct{}

// NewEventService creates a new event service
func NewEventService() *EventService {
	return &EventService{}
}

// OnPostCreated triggers notifications for new post
func (s *EventService) OnPostCreated(post *models.Post) error {
	// Notify followers
	var followers []uint
	db.GetDB().Table("follows").
		Where("followee_id = ?", post.UserID).
		Pluck("follower_id", &followers)

	notificationService := NewNotificationService()
	for _, followerID := range followers {
		_, err := notificationService.TriggerNotification(
			"post",
			followerID,
			post.UserID,
			&post.ID,
		)
		if err != nil {
			fmt.Printf("Failed to create notification: %v\n", err)
		}
	}
	return nil
}

// OnLikeCreated triggers notification for new like
func (s *EventService) OnLikeCreated(like *models.Like) error {
	notificationService := NewNotificationService()
	_, err := notificationService.TriggerNotification(
		"like",
		like.UserID, // post owner
		like.UserID, // liker (same for now, will be different in real impl)
		&like.PostID,
	)
	return err
}

// OnCommentCreated triggers notification for new comment
func (s *EventService) OnCommentCreated(comment *models.Comment) error {
	notificationService := NewNotificationService()
	_, err := notificationService.TriggerNotification(
		"comment",
		comment.UserID, // post owner
		comment.UserID, // commenter
		&comment.PostID,
	)
	return err
}

// OnFollowCreated triggers notification for new follow
func (s *EventService) OnFollowCreated(follow *models.Follow) error {
	notificationService := NewNotificationService()
	_, err := notificationService.TriggerNotification(
		"follower",
		follow.FolloweeID,
		follow.FollowerID,
		nil,
	)
	return err
}

// OnDMReceived triggers notification for new DM
func (s *EventService) OnDMReceived(senderID, recipientID uint, conversationID uint) error {
	notificationService := NewNotificationService()
	postID := uint(0)
	_, err := notificationService.TriggerNotification(
		"dm_reply",
		recipientID,
		senderID,
		&postID,
	)
	return err
}
