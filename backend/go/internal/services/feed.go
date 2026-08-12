package services

import (
	"crypto/rand"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

const publicIDChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomPublicID generates a cryptographically random 16+ character public identifier
func GenerateRandomPublicID(length int) string {
	if length < 16 {
		length = 16
	}
	b := make([]byte, length)
	_, _ = rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = publicIDChars[int(b[i])%len(publicIDChars)]
	}
	return string(b)
}

// FeedService handles feed-related operations
type FeedService struct{}

// NewFeedService creates a new feed service
func NewFeedService() *FeedService {
	return &FeedService{}
}

// FeedPost contains post data with engagement info
type FeedPost struct {
	ID            uint        `json:"id"`
	PublicID      string      `json:"public_id"`
	UserID        uint        `json:"user_id"`
	Content       string      `json:"content"`
	CreatedAt     time.Time   `json:"created_at"`
	User          models.User `json:"user"`
	Likes         int         `json:"likes"`
	LikeCount     int         `json:"like_count"`
	Comments      int         `json:"comments"`
	CommentCount  int         `json:"comment_count"`
	Reposts       int         `json:"reposts"`
	RepostCount   int         `json:"repost_count"`
	Bookmarks     int         `json:"bookmarks"`
	BookmarkCount int         `json:"bookmark_count"`
	IsLiked       bool        `json:"is_liked"`
	IsReposted    bool        `json:"is_reposted"`
	IsBookmarked  bool        `json:"is_bookmarked"`
}

// ListFeedInput contains feed request parameters
type ListFeedInput struct {
	UserID  uint
	Tab     string // "following" or "foryou"
	Page    int
	PerPage int
}

// ListFeedOutput contains feed posts with pagination
type ListFeedOutput struct {
	Posts   []FeedPost `json:"posts"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
	Total   int        `json:"total"`
}

// ListFeed retrieves posts for the feed
func (s *FeedService) ListFeed(input ListFeedInput) (*ListFeedOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 || input.PerPage > 50 {
		input.PerPage = 20
	}
	if input.Tab != "following" && input.Tab != "foryou" {
		input.Tab = "foryou"
	}

	var posts []models.Post
	var total int64
	query := db.GetDB().Model(&models.Post{}).Where("deleted_at IS NULL")

	// Following tab: show posts from followed users + own posts
	if input.Tab == "following" && input.UserID > 0 {
		var followeeIDs []uint
		db.GetDB().Model(&models.Follow{}).Where("follower_id = ?", input.UserID).Pluck("followee_id", &followeeIDs)
		followeeIDs = append(followeeIDs, input.UserID)
		query = query.Where("user_id IN ?", followeeIDs)
	}

	query.Count(&total)

	offset := (input.Page - 1) * input.PerPage
	query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(input.PerPage).
		Find(&posts)

	// Build feed posts with engagement info
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

		// Get engagement counts
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

		// Check engagement status
		if input.UserID > 0 {
			var like models.Like
			db.GetDB().Where("user_id = ? AND post_id = ?", input.UserID, post.ID).First(&like)
			feedPost.IsLiked = like.ID > 0

			var repost models.Repost
			db.GetDB().Where("user_id = ? AND post_id = ?", input.UserID, post.ID).First(&repost)
			feedPost.IsReposted = repost.ID > 0

			var bookmark models.Bookmark
			db.GetDB().Where("user_id = ? AND post_id = ?", input.UserID, post.ID).First(&bookmark)
			feedPost.IsBookmarked = bookmark.ID > 0
		}

		feedPosts = append(feedPosts, feedPost)
	}

	return &ListFeedOutput{
		Posts:   feedPosts,
		Page:    input.Page,
		PerPage: input.PerPage,
		Total:   int(total),
	}, nil
}
