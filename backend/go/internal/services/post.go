package services

import (
	"fmt"
	"strconv"
	"strings"
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
	UserID     uint
	Content    string
	MediaURLs  []string
	MediaTypes []string
}

// CreatePostOutput contains the created post with user info
type CreatePostOutput struct {
	Post models.Post `json:"post"`
}

// GetPostInput contains parameters for getting a single post
type GetPostInput struct {
	Identifier string // Can be 16-char public_id or numeric ID string
	UserID     uint   // For checking ownership
}

// GetPostOutput contains post with engagement counts
type GetPostOutput struct {
	Post         models.Post `json:"post"`
	Likes        int         `json:"likes"`
	Comments     int         `json:"comments"`
	Reposts      int         `json:"reposts"`
	IsLiked      bool        `json:"is_liked"`
	IsReposted   bool        `json:"is_reposted"`
	IsBookmarked bool        `json:"is_bookmarked"`
}

// ListPostsInput contains pagination and user filter/viewer params
type ListPostsInput struct {
	FilterUserID uint // Filter posts written by this author ID (0 = all posts)
	ViewerUserID uint // Check is_liked, is_reposted, is_bookmarked for this viewer ID
	Page         int
	PerPage      int
}

// ListPostsOutput contains posts with pagination meta
type ListPostsOutput struct {
	Posts   []FeedPost `json:"posts"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
	Total   int        `json:"total"`
}

// FindPostByIdentifier retrieves a post by its 16-char public_id or numeric ID
func (s *PostService) FindPostByIdentifier(identifier string) (*models.Post, error) {
	var post models.Post
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, fmt.Errorf("invalid identifier")
	}

	// 1. Exact match by public_id
	if err := db.GetDB().Model(&models.Post{}).Preload("User").Where("public_id = ? AND (deleted_at IS NULL)", identifier).First(&post).Error; err == nil {
		return &post, nil
	}

	// 2. Case-insensitive match by public_id
	if err := db.GetDB().Model(&models.Post{}).Preload("User").Where("LOWER(public_id) = LOWER(?) AND (deleted_at IS NULL)", identifier).First(&post).Error; err == nil {
		return &post, nil
	}

	// 3. Fallback to numeric ID if identifier is digits
	if numID, err := strconv.ParseUint(identifier, 10, 64); err == nil {
		if err := db.GetDB().Model(&models.Post{}).Preload("User").Where("id = ? AND (deleted_at IS NULL)", uint(numID)).First(&post).Error; err == nil {
			return &post, nil
		}
	}

	return nil, fmt.Errorf("post not found")
}

// CreatePost creates a new post with a 16-character random public_id
func (s *PostService) CreatePost(input CreatePostInput) (*CreatePostOutput, error) {
	if len(input.Content) == 0 {
		return nil, fmt.Errorf("post content cannot be empty")
	}
	if len(input.Content) > 500 {
		return nil, fmt.Errorf("post content cannot exceed 500 characters")
	}

	publicID := GenerateRandomPublicID(16)

	// Create post
	post := models.Post{
		PublicID:  publicID,
		UserID:    input.UserID,
		Content:   input.Content,
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
			PostID:    post.ID,
			MediaURL:  url,
			MediaType: input.MediaTypes[i],
			Order:     i,
			CreatedAt: time.Now(),
		}
		if err := db.GetDB().Create(&media).Error; err != nil {
			// Rollback post if media creation fails
			db.GetDB().Delete(&post)
			return nil, fmt.Errorf("failed to create media: %w", err)
		}
	}

	// Reload post with user
	db.GetDB().Model(&models.Post{}).Preload("User").First(&post, post.ID)

	return &CreatePostOutput{Post: post}, nil
}

// GetPost retrieves a single post by public_id or numeric ID
func (s *PostService) GetPost(input GetPostInput) (*GetPostOutput, error) {
	post, err := s.FindPostByIdentifier(input.Identifier)
	if err != nil {
		return nil, err
	}

	if post.PublicID == "" {
		post.PublicID = GenerateRandomPublicID(16)
		db.GetDB().Model(&models.Post{}).Where("id = ?", post.ID).Update("public_id", post.PublicID)
	}

	// Get engagement counts
	var likeCount int64
	db.GetDB().Model(&models.Like{}).Where("post_id = ?", post.ID).Count(&likeCount)

	var commentCount int64
	db.GetDB().Model(&models.Comment{}).Where("post_id = ?", post.ID).Count(&commentCount)

	var repostCount int64
	db.GetDB().Model(&models.Repost{}).Where("post_id = ?", post.ID).Count(&repostCount)

	// Check engagement status safely
	isLiked := false
	isReposted := false
	isBookmarked := false

	if input.UserID > 0 {
		var like models.Like
		db.GetDB().Model(&models.Like{}).Where("user_id = ? AND post_id = ?", input.UserID, post.ID).First(&like)
		isLiked = like.ID > 0

		var repost models.Repost
		db.GetDB().Model(&models.Repost{}).Where("user_id = ? AND post_id = ?", input.UserID, post.ID).First(&repost)
		isReposted = repost.ID > 0

		var bookmark models.Bookmark
		db.GetDB().Model(&models.Bookmark{}).Where("user_id = ? AND post_id = ?", input.UserID, post.ID).First(&bookmark)
		isBookmarked = bookmark.ID > 0
	}

	return &GetPostOutput{
		Post:         *post,
		Likes:        int(likeCount),
		Comments:     int(commentCount),
		Reposts:      int(repostCount),
		IsLiked:      isLiked,
		IsReposted:   isReposted,
		IsBookmarked: isBookmarked,
	}, nil
}

// ListPosts retrieves a list of posts with engagement counts and pagination
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
	if input.FilterUserID > 0 {
		query = query.Where("user_id = ?", input.FilterUserID)
	}
	query.Count(&total)

	offset := (input.Page - 1) * input.PerPage
	query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(input.PerPage).
		Find(&posts)

	feedPosts := make([]FeedPost, 0, len(posts))
	for _, post := range posts {
		if post.PublicID == "" {
			post.PublicID = GenerateRandomPublicID(16)
			db.GetDB().Model(&models.Post{}).Where("id = ?", post.ID).Update("public_id", post.PublicID)
		}

		feedPost := FeedPost{
			ID:        post.ID,
			PublicID:  post.PublicID,
			UserID:    post.UserID,
			Content:   post.Content,
			CreatedAt: post.CreatedAt,
			User:      post.User,
		}

		var likeCount int64
		db.GetDB().Model(&models.Like{}).Where("post_id = ?", post.ID).Count(&likeCount)
		feedPost.Likes = int(likeCount)
		feedPost.LikeCount = int(likeCount)

		var commentCount int64
		db.GetDB().Model(&models.Comment{}).Where("post_id = ?", post.ID).Count(&commentCount)
		feedPost.Comments = int(commentCount)
		feedPost.CommentCount = int(commentCount)

		var repostCount int64
		db.GetDB().Model(&models.Repost{}).Where("post_id = ?", post.ID).Count(&repostCount)
		feedPost.Reposts = int(repostCount)
		feedPost.RepostCount = int(repostCount)

		var bookmarkCount int64
		db.GetDB().Model(&models.Bookmark{}).Where("post_id = ?", post.ID).Count(&bookmarkCount)
		feedPost.Bookmarks = int(bookmarkCount)
		feedPost.BookmarkCount = int(bookmarkCount)

		// Check engagement status for the currently viewing user (ViewerUserID)
		if input.ViewerUserID > 0 {
			var like models.Like
			db.GetDB().Model(&models.Like{}).Where("user_id = ? AND post_id = ?", input.ViewerUserID, post.ID).First(&like)
			feedPost.IsLiked = like.ID > 0

			var repost models.Repost
			db.GetDB().Model(&models.Repost{}).Where("user_id = ? AND post_id = ?", input.ViewerUserID, post.ID).First(&repost)
			feedPost.IsReposted = repost.ID > 0

			var bookmark models.Bookmark
			db.GetDB().Model(&models.Bookmark{}).Where("user_id = ? AND post_id = ?", input.ViewerUserID, post.ID).First(&bookmark)
			feedPost.IsBookmarked = bookmark.ID > 0
		}

		feedPosts = append(feedPosts, feedPost)
	}

	return &ListPostsOutput{
		Posts:   feedPosts,
		Page:    input.Page,
		PerPage: input.PerPage,
		Total:   int(total),
	}, nil
}

// DeletePost soft deletes a post
func (s *PostService) DeletePost(identifier string, userID uint) error {
	post, err := s.FindPostByIdentifier(identifier)
	if err != nil {
		return err
	}

	// Check ownership
	if post.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	// Soft delete
	now := time.Now()
	db.GetDB().Model(post).Update("deleted_at", &now)

	return nil
}

// LikePost adds a like to a post
func (s *PostService) LikePost(identifier string, userID uint) error {
	post, err := s.FindPostByIdentifier(identifier)
	if err != nil {
		return err
	}

	// Check if already liked
	var existing models.Like
	db.GetDB().Model(&models.Like{}).Where("user_id = ? AND post_id = ?", userID, post.ID).First(&existing)
	if existing.ID > 0 {
		return fmt.Errorf("already liked")
	}

	like := models.Like{
		UserID:    userID,
		PostID:    post.ID,
		CreatedAt: time.Now(),
	}

	return db.GetDB().Create(&like).Error
}

// UnlikePost removes a like from a post
func (s *PostService) UnlikePost(identifier string, userID uint) error {
	post, err := s.FindPostByIdentifier(identifier)
	if err != nil {
		return err
	}

	result := db.GetDB().Where("user_id = ? AND post_id = ?", userID, post.ID).Delete(&models.Like{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("like not found")
	}
	return nil
}

// CommentOnPost adds a comment to a post and preloads author User relation
func (s *PostService) CommentOnPost(identifier string, userID uint, content string, parentID *uint) (*models.Comment, error) {
	post, err := s.FindPostByIdentifier(identifier)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("comment content cannot be empty")
	}
	if len(content) > 500 {
		return nil, fmt.Errorf("comment content cannot exceed 500 characters")
	}

	comment := models.Comment{
		UserID:    userID,
		PostID:    post.ID,
		Content:   content,
		ParentID:  parentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.GetDB().Create(&comment).Error; err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Preload author user info before returning
	db.GetDB().Model(&models.Comment{}).Preload("User").First(&comment, comment.ID)

	return &comment, nil
}

// GetPostComments retrieves comments for a post by identifier
func (s *PostService) GetPostComments(identifier string) ([]models.Comment, error) {
	post, err := s.FindPostByIdentifier(identifier)
	if err != nil {
		return nil, err
	}

	var comments []models.Comment
	db.GetDB().Model(&models.Comment{}).Preload("User").
		Where("post_id = ?", post.ID).
		Order("created_at ASC").
		Find(&comments)
	return comments, nil
}

// RepostPost creates a repost
func (s *PostService) RepostPost(identifier string, userID uint, comment *string) (*models.Repost, error) {
	post, err := s.FindPostByIdentifier(identifier)
	if err != nil {
		return nil, err
	}

	// Check if already reposted
	var existing models.Repost
	db.GetDB().Model(&models.Repost{}).Where("user_id = ? AND post_id = ?", userID, post.ID).First(&existing)
	if existing.ID > 0 {
		return nil, fmt.Errorf("already reposted")
	}

	repost := models.Repost{
		UserID:    userID,
		PostID:    post.ID,
		Comment:   comment,
		CreatedAt: time.Now(),
	}

	if err := db.GetDB().Create(&repost).Error; err != nil {
		return nil, fmt.Errorf("failed to create repost: %w", err)
	}

	return &repost, nil
}

// UnrepostPost removes a repost
func (s *PostService) UnrepostPost(identifier string, userID uint) error {
	post, err := s.FindPostByIdentifier(identifier)
	if err != nil {
		return err
	}

	result := db.GetDB().Where("user_id = ? AND post_id = ?", userID, post.ID).Delete(&models.Repost{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("repost not found")
	}
	return nil
}

// BookmarkPost saves a post
func (s *PostService) BookmarkPost(identifier string, userID uint) error {
	post, err := s.FindPostByIdentifier(identifier)
	if err != nil {
		return err
	}

	// Check if already bookmarked
	var existing models.Bookmark
	db.GetDB().Model(&models.Bookmark{}).Where("user_id = ? AND post_id = ?", userID, post.ID).First(&existing)
	if existing.ID > 0 {
		return fmt.Errorf("already bookmarked")
	}

	bookmark := models.Bookmark{
		UserID:    userID,
		PostID:    post.ID,
		CreatedAt: time.Now(),
	}

	return db.GetDB().Create(&bookmark).Error
}

// UnbookmarkPost removes a bookmark
func (s *PostService) UnbookmarkPost(identifier string, userID uint) error {
	post, err := s.FindPostByIdentifier(identifier)
	if err != nil {
		return err
	}

	result := db.GetDB().Where("user_id = ? AND post_id = ?", userID, post.ID).Delete(&models.Bookmark{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("bookmark not found")
	}
	return nil
}
