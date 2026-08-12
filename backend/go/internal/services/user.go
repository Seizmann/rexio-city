package services

import (
	"fmt"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

// UserService handles user-related operations
type UserService struct{}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{}
}

// UserProfile contains user data with follow status
type UserProfile struct {
	ID         uint    `json:"id"`
	Username   string  `json:"username"`
	DisplayName *string `json:"display_name"`
	Bio        *string `json:"bio"`
	AvatarURL  *string `json:"avatar_url"`
	CoverURL   *string `json:"cover_url"`
	Followers  int     `json:"followers"`
	Following  int     `json:"following"`
	IsFollowing bool   `json:"is_following"`
	CreatedAt  time.Time `json:"created_at"`
}

// GetUserInput contains parameters for getting a user
type GetUserInput struct {
	UserID uint
	Username string
	CurrentUserID uint // For checking follow status
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID uint) (*UserProfile, error) {
	var user models.User
	result := db.GetDB().First(&user, userID)
	if result.Error != nil {
		return nil, fmt.Errorf("user not found")
	}

	return s.buildUserProfile(user, 0)
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string, currentUserID uint) (*UserProfile, error) {
	var user models.User
	result := db.GetDB().Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("user not found")
	}

	return s.buildUserProfile(user, currentUserID)
}

// UpdateUserInput contains fields for updating a user
type UpdateUserInput struct {
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatar_url"`
	CoverURL    *string `json:"cover_url"`
}

// UpdateUser updates user profile information
func (s *UserService) UpdateUser(userID uint, input UpdateUserInput) (*UserProfile, error) {
	var user models.User
	result := db.GetDB().First(&user, userID)
	if result.Error != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	if input.DisplayName != nil {
		if len(*input.DisplayName) > 50 {
			return nil, fmt.Errorf("display name cannot exceed 50 characters")
		}
		user.DisplayName = input.DisplayName
	}

	if input.Bio != nil {
		if len(*input.Bio) > 160 {
			return nil, fmt.Errorf("bio cannot exceed 160 characters")
		}
		user.Bio = input.Bio
	}

	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	if input.CoverURL != nil {
		user.CoverURL = input.CoverURL
	}

	user.UpdatedAt = time.Now()
	db.GetDB().Save(&user)

	return s.buildUserProfile(user, 0)
}

// SearchUsersResult contains search results
type SearchUsersResult struct {
	Users []models.User `json:"users"`
	Total int `json:"total"`
}

// SearchUsers searches for users by username or display name
func (s *UserService) SearchUsers(query string, page, perPage int) (*SearchUsersResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	var users []models.User
	var total int64

	db.GetDB().Model(&models.User{}).
		Where("username LIKE ? OR display_name LIKE ?", "%"+query+"%", "%"+query+"%").
		Count(&total)

	offset := (page - 1) * perPage
	db.GetDB().
		Where("username LIKE ? OR display_name LIKE ?", "%"+query+"%", "%"+query+"%").
		Offset(offset).
		Limit(perPage).
		Find(&users)

	return &SearchUsersResult{
		Users: users,
		Total: int(total),
	}, nil
}

// buildUserProfile constructs a UserProfile from a User model
func (s *UserService) buildUserProfile(user models.User, currentUserID uint) (*UserProfile, error) {
	followers, following, err := newFollowService().GetUserFollowCounts(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get follow counts: %w", err)
	}

	isFollowing := false
	if currentUserID > 0 {
		isFollowing, _ = newFollowService().IsFollowing(currentUserID, user.ID)
	}

	return &UserProfile{
		ID: user.ID,
		Username: user.Username,
		DisplayName: user.DisplayName,
		Bio: user.Bio,
		AvatarURL: user.AvatarURL,
		CoverURL: user.CoverURL,
		Followers: followers,
		Following: following,
		IsFollowing: isFollowing,
		CreatedAt: user.CreatedAt,
	}, nil
}

// Helper to create follow service instance
func newFollowService() *FollowService {
	return NewFollowService()
}
