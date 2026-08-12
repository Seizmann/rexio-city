package services

import (
	"fmt"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

// PostService handles post-related operations
type PostService struct{}

// NewPostService creates a new post service
func NewPostService() *PostService {
	return &PostService{}
}

// CreatePostInput contains post creation data
type CreatePostInput struct {
	UserID uint
	Content string
	MediaURLs []string
	MediaTypes []string
}

// CreatePostOutput contains the created post with user info
type CreatePostOutput struct {
	Post models.Post `json:"post"`
}

// GetPostInput contains parameters for getting a single post
type GetPostInput struct {
	PostID uint
	UserID uint // For checking ownership
}

// GetPostOutput contains post with engagement counts
type GetPostOutput struct {
	Post models.Post `json:"post"`
	Likes int `json:"likes"`
	Comments int `json:"comments"`
	Reposts int `json:"reposts"`
	IsLiked bool `json:"is_liked"`
	IsReposted bool `json:"is_reposted"`
	IsBookmarked bool `json:"is_bookmarked"`
}

// ListPostsInput contains pagination params
type ListPostsInput struct {
	UserID uint
	Page int
	PerPage int
}

// ListPostsOutput contains posts with pagination meta
type ListPostsOutput struct {
	Posts []models.Post `json:"posts"`
	Page int `json:"page"`
	PerPage int `json:"per_page"`
	Total int `json:"total"`
}

// CreatePost creates a new post
func (s *PostService) CreatePost(input CreatePostInput) (*CreatePostOutput, error) {
	if len(input.Content) == 0 {
		return nil, fmt.Errorf("post content cannot be empty")
	}
	if len(input.Content) > 500 {
		return nil, fmt.Errorf("post content cannot exceed 500 characters")
	}

	// Create post
	post := models.Post{
		UserID: input.UserID,
		Content: input.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := db.GetDB().Create(&post)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create post: %w", result.Error)
	}

	// Create media entries if provided
	for i, url := range input.MediaURLs {
		media := models.PostMedia{
			PostID: post.ID,
			MediaURL: url,
			MediaType: input.MediaTypes[i],
			Order: i,
			CreatedAt: time.Now(),
		}
		if err := db.GetDB().Create(&media).Error; err != nil {
			// Rollback post if media creation fails
			db.GetDB().Delete(&post)
			return nil, fmt.Errorf("failed to create media: %w", err)
		}
	}

	// Reload post with user
	db.GetDB().Preload("User").First(&post, post.ID)

	return &CreatePostOutput{Post: post}, nil
}

// GetPost retrieves a single post by ID
func (s *PostService) GetPost(input GetPostInput) (*GetPostOutput, error) {
	var post models.Post
	result := db.GetDB().Preload("User").First(&post, input.PostID)
	if result.Error != nil {
		return nil, fmt.Errorf("post not found")
	}

	// Get engagement counts
	var likeCount int64
	db.GetDB().Model(&models.Like{}).Where("post_id = ?", input.PostID).Count(&likeCount)

	var commentCount int64
	db.GetDB().Model(&models.Comment{}).Where("post_id = ?", input.PostID).Count(&commentCount)

	var repostCount int64
	db.GetDB().Model(&models.Repost{}).Where("post_id = ?", input.PostID).Count(&repostCount)

	// Check engagement status
	isLiked := false
	isReposted := false
	isBookmarked := false

	if input.UserID > 0 {
		db.GetDB().Where("user_id = ? AND post_id = ?", input.UserID, input.PostID).First(&models.Like{}).Scan(&isLiked)
		db.GetDB().Where("user_id = ? AND post_id = ?", input.UserID, input.PostID).First(&models.Repost{}).Scan(&isReposted)
		db.GetDB().Where("user_id = ? AND post_id = ?", input.UserID, input.PostID).First(&models.Bookmark{}).Scan(&isBookmarked)
	}

	return &GetPostOutput{
		Post: post,
		Likes: int(likeCount),
		Comments: int(commentCount),
		Reposts: int(repostCount),
		IsLiked: isLiked,
		IsReposted: isReposted,
		IsBookmarked: isBookmarked,
	}, nil
}

// ListPosts retrieves a list of posts with pagination
func (s *PostService) ListPosts(input ListPostsInput) (*ListPostsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 || input.PerPage > 50 {
		input.PerPage = 20
	}

	var posts []models.Post
	var total int64

	query := db.GetDB().Model(&models.Post{}).Where("deleted_at IS NULL")
	if input.UserID > 0 {
		query = query.Where("user_id = ?", input.UserID)
	}
	query.Count(&total)

	offset := (input.Page - 1) * input.PerPage
	db.GetDB().Preload("User").
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Offset(offset).
		Limit(input.PerPage).
		Find(&posts)

	return &ListPostsOutput{
		Posts: posts,
		Page: input.Page,
		PerPage: input.PerPage,
		Total: int(total),
	}, nil
}

// DeletePost soft deletes a post
func (s *PostService) DeletePost(postID uint, userID uint) error {
	var post models.Post
	result := db.GetDB().First(&post, postID)
	if result.Error != nil {
		return fmt.Errorf("post not found")
	}

	// Check ownership
	if post.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	// Soft delete
	now := time.Now()
	db.GetDB().Model(&post).Update("deleted_at", &now)

	return nil
}

// LikePost adds a like to a post
func (s *PostService) LikePost(postID uint, userID uint) error {
	// Check if already liked
	var existing models.Like
	db.GetDB().Where("user_id = ? AND post_id = ?", userID, postID).First(&existing)
	if existing.ID > 0 {
		return fmt.Errorf("already liked")
	}

	like := models.Like{
		UserID: userID,
		PostID: postID,
		CreatedAt: time.Now(),
	}

	return db.GetDB().Create(&like).Error
}

// UnlikePost removes a like from a post
func (s *PostService) UnlikePost(postID uint, userID uint) error {
	result := db.GetDB().Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.Like{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("like not found")
	}
	return nil
}

// CommentOnPost adds a comment to a post
func (s *PostService) CommentOnPost(postID uint, userID uint, content string, parentID *uint) (*models.Comment, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("comment content cannot be empty")
	}
	if len(content) > 500 {
		return nil, fmt.Errorf("comment content cannot exceed 500 characters")
	}

	comment := models.Comment{
		UserID: userID,
		PostID: postID,
		Content: content,
		ParentID: parentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.GetDB().Create(&comment).Error; err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return &comment, nil
}

// GetPostComments retrieves comments for a post
func (s *PostService) GetPostComments(postID uint) ([]models.Comment, error) {
	var comments []models.Comment
	db.GetDB().Where("post_id = ?", postID).
		Order("created_at ASC").
		Find(&comments)
	return comments, nil
}

// RepostPost creates a repost
func (s *PostService) RepostPost(postID uint, userID uint, comment *string) (*models.Repost, error) {
	// Check if already reposted
	var existing models.Repost
	db.GetDB().Where("user_id = ? AND post_id = ?", userID, postID).First(&existing)
	if existing.ID > 0 {
		return nil, fmt.Errorf("already reposted")
	}

	repost := models.Repost{
		UserID: userID,
		PostID: postID,
		Comment: comment,
		CreatedAt: time.Now(),
	}

	if err := db.GetDB().Create(&repost).Error; err != nil {
		return nil, fmt.Errorf("failed to create repost: %w", err)
	}

	return &repost, nil
}

// UnrepostPost removes a repost
func (s *PostService) UnrepostPost(postID uint, userID uint) error {
	result := db.GetDB().Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.Repost{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("repost not found")
	}
	return nil
}

// BookmarkPost saves a post
func (s *PostService) BookmarkPost(postID uint, userID uint) error {
	// Check if already bookmarked
	var existing models.Bookmark
	db.GetDB().Where("user_id = ? AND post_id = ?", userID, postID).First(&existing)
	if existing.ID > 0 {
		return fmt.Errorf("already bookmarked")
	}

	bookmark := models.Bookmark{
		UserID: userID,
		PostID: postID,
		CreatedAt: time.Now(),
	}

	return db.GetDB().Create(&bookmark).Error
}

// UnbookmarkPost removes a bookmark
func (s *PostService) UnbookmarkPost(postID uint, userID uint) error {
	result := db.GetDB().Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.Bookmark{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("bookmark not found")
	}
	return nil
}
