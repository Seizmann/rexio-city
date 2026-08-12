package services

import (
	"fmt"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

// FollowService handles follow-related operations
type FollowService struct{}

// NewFollowService creates a new follow service
func NewFollowService() *FollowService {
	return &FollowService{}
}

// FollowUser makes userA follow userB
func (s *FollowService) FollowUser(followerID, followeeID uint) error {
	// Cannot follow yourself
	if followerID == followeeID {
		return fmt.Errorf("cannot follow yourself")
	}

	// Check if already following
	var count int64
	db.GetDB().Model(&models.Follow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Count(&count)
	if count > 0 {
		return fmt.Errorf("already following")
	}

	follow := models.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		CreatedAt: time.Now(),
	}

	if err := db.GetDB().Create(&follow).Error; err != nil {
		return fmt.Errorf("failed to follow user: %w", err)
	}

	return nil
}

// UnfollowUser makes userA unfollow userB
func (s *FollowService) UnfollowUser(followerID, followeeID uint) error {
	result := db.GetDB().Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Delete(&models.Follow{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("not following")
	}
	return nil
}

// IsFollowing checks if userA follows userB
func (s *FollowService) IsFollowing(followerID, followeeID uint) (bool, error) {
	var count int64
	db.GetDB().Model(&models.Follow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Count(&count)
	return count > 0, nil
}

// GetFollowers retrieves followers of a user
func (s *FollowService) GetFollowers(userID uint, page, perPage int) ([]models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	var users []models.User
	var total int64

	db.GetDB().Model(&models.Follow{}).
		Where("followee_id = ?", userID).
		Count(&total)

	offset := (page - 1) * perPage
	db.GetDB().Table("users").
		Joins("INNER JOIN follows ON users.id = follows.follower_id").
		Where("follows.followee_id = ?", userID).
		Offset(offset).
		Limit(perPage).
		Find(&users)

	return users, int(total), nil
}

// GetFollowing retrieves users that a user follows
func (s *FollowService) GetFollowing(userID uint, page, perPage int) ([]models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	var users []models.User
	var total int64

	db.GetDB().Model(&models.Follow{}).
		Where("follower_id = ?", userID).
		Count(&total)

	offset := (page - 1) * perPage
	db.GetDB().Table("users").
		Joins("INNER JOIN follows ON users.id = follows.followee_id").
		Where("follows.follower_id = ?", userID).
		Offset(offset).
		Limit(perPage).
		Find(&users)

	return users, int(total), nil
}

// GetUserFollowCounts returns follower and following counts for a user
func (s *FollowService) GetUserFollowCounts(userID uint) (followers, following int, err error) {
	var followCount int64
	var followingCount int64
	db.GetDB().Model(&models.Follow{}).Where("followee_id = ?", userID).Count(&followCount)
	db.GetDB().Model(&models.Follow{}).Where("follower_id = ?", userID).Count(&followingCount)
	return int(followCount), int(followingCount), nil
}
