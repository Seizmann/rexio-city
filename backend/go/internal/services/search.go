package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

// SearchService handles search-related operations
type SearchService struct{}

// NewSearchService creates a new search service
func NewSearchService() *SearchService {
	return &SearchService{}
}

// SearchResult contains search results grouped by type
type SearchResult struct {
	Users       []SearchUserResult    `json:"users,omitempty"`
	Posts       []SearchPostResult    `json:"posts,omitempty"`
	Hashtags    []SearchHashtagResult `json:"hashtags,omitempty"`
	Total       int                   `json:"total"`
	HasUsers    bool                  `json:"has_users"`
	HasPosts    bool                  `json:"has_posts"`
	HasHashtags bool                  `json:"has_hashtags"`
}

// SearchUserResult contains user data for search results
type SearchUserResult struct {
	ID          uint    `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

// SearchPostResult contains post data for search results
type SearchPostResult struct {
	ID        uint             `json:"id"`
	PublicID  string           `json:"public_id"`
	Content   string           `json:"content"`
	UserID    uint             `json:"user_id"`
	CreatedAt string           `json:"created_at"`
	User      SearchUserResult `json:"user"`
}

// SearchHashtagResult contains hashtag search data
type SearchHashtagResult struct {
	Hashtag string             `json:"hashtag"`
	Count   int                `json:"count"`
	Posts   []SearchPostResult `json:"posts,omitempty"`
}

// searchRegex matches hashtags (#word) but only standalone words
// Allows letters, numbers, underscores, and hyphens
var hashtagRegex = regexp.MustCompile(`\B#[\w-]+`)

// extractHashtags extracts unique hashtags from post content
func extractHashtags(content string) []string {
	matches := hashtagRegex.FindAllString(content, -1)
	seen := make(map[string]bool)
	var hashtags []string
	for _, h := range matches {
		// Normalize: remove leading #, lowercase
		tag := strings.ToLower(h[1:])
		if !seen[tag] {
			seen[tag] = true
			hashtags = append(hashtags, tag)
		}
	}
	return hashtags
}

// SearchUsers searches for users by username or display name
func (s *SearchService) SearchUsers(query string, page, perPage int) (*SearchResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	var users []models.User
	var total int64

	// Parameterized query to prevent SQL injection
	db.GetDB().Model(&models.User{}).
		Where("username LIKE ? OR display_name LIKE ?", "%"+query+"%", "%"+query+"%").
		Count(&total)

	offset := (page - 1) * perPage
	db.GetDB().
		Where("username LIKE ? OR display_name LIKE ?", "%"+query+"%", "%"+query+"%").
		Offset(offset).
		Limit(perPage).
		Find(&users)

	results := make([]SearchUserResult, len(users))
	for i, u := range users {
		results[i] = SearchUserResult{
			ID:          u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarURL,
		}
	}

	return &SearchResult{
		Users:    results,
		Total:    int(total),
		HasUsers: len(results) > 0,
	}, nil
}

// SearchPosts searches for posts by content (case-insensitive, excludes soft-deleted)
func (s *SearchService) SearchPosts(query string, page, perPage int) (*SearchResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	var posts []models.Post
	var total int64

	// Use ILIKE for case-insensitive partial match
	db.GetDB().Model(&models.Post{}).
		Where("content ILIKE ? AND deleted_at IS NULL", "%"+query+"%").
		Count(&total)

	offset := (page - 1) * perPage
	db.GetDB().
		Where("content ILIKE ? AND deleted_at IS NULL", "%"+query+"%").
		Offset(offset).
		Limit(perPage).
		Preload("User").
		Find(&posts)

	results := make([]SearchPostResult, len(posts))
	for i, p := range posts {
		result := SearchPostResult{
			ID:        p.ID,
			PublicID:  p.PublicID,
			Content:   p.Content,
			UserID:    p.UserID,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			User: SearchUserResult{
				ID:          p.User.ID,
				Username:    p.User.Username,
				DisplayName: p.User.DisplayName,
				AvatarURL:   p.User.AvatarURL,
			},
		}
		results[i] = result
	}

	return &SearchResult{
		Posts:    results,
		Total:    int(total),
		HasPosts: len(results) > 0,
	}, nil
}

// SearchHashtags searches for hashtags by matching against extracted hashtags from posts
func (s *SearchService) SearchHashtags(query string, page, perPage int) (*SearchResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	query = strings.ToLower(strings.TrimSpace(query))
	if !strings.HasPrefix(query, "#") {
		query = "#" + query
	}
	query = strings.ToLower(query)

	// Find all posts containing this hashtag
	var posts []models.Post
	db.GetDB().
		Where("content ILIKE ? AND deleted_at IS NULL", "%"+query+"%").
		Preload("User").
		Find(&posts)

	// Extract all hashtags from matching posts
	seenPosts := make(map[uint]bool)
	hashtagCounts := make(map[string]int)
	hashtagPosts := make(map[string][]SearchPostResult)

	for _, p := range posts {
		if seenPosts[p.ID] {
			continue
		}
		seenPosts[p.ID] = true

		hashtags := extractHashtags(p.Content)
		for _, tag := range hashtags {
			hashtagCounts["#"+tag]++
			if len(hashtagPosts["#"+tag]) < 3 { // Store up to 3 sample posts per hashtag
				hashtagPosts["#"+tag] = append(hashtagPosts["#"+tag], SearchPostResult{
					ID:        p.ID,
					PublicID:  p.PublicID,
					Content:   p.Content,
					UserID:    p.UserID,
					CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
					User: SearchUserResult{
						ID:          p.User.ID,
						Username:    p.User.Username,
						DisplayName: p.User.DisplayName,
						AvatarURL:   p.User.AvatarURL,
					},
				})
			}
		}
	}

	// Filter hashtags matching the query
	var results []SearchHashtagResult
	total := 0
	for tag, count := range hashtagCounts {
		if strings.HasPrefix(tag, query) {
			total++
			result := SearchHashtagResult{
				Hashtag: tag,
				Count:   count,
				Posts:   hashtagPosts[tag],
			}
			results = append(results, result)
		}
	}

	// Sort by count descending (simple bubble sort for small datasets)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Count > results[i].Count {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Paginate
	if page*perPage < len(results) {
		results = results[(page-1)*perPage : page*perPage]
	} else if len(results) > 0 {
		results = results[(page-1)*perPage:]
	}

	return &SearchResult{
		Hashtags:    results,
		Total:       total,
		HasHashtags: len(results) > 0,
	}, nil
}

// SearchCombined performs a combined search returning top results from each type
func (s *SearchService) SearchCombined(query string, limit int) (*SearchResult, error) {
	if limit < 1 || limit > 20 {
		limit = 5
	}

	// Search users
	usersResult, err := s.SearchUsers(query, 1, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Search posts
	postsResult, err := s.SearchPosts(query, 1, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}

	// Search hashtags (limit to top 3)
	hashtagsResult, err := s.SearchHashtags(query, 1, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to search hashtags: %w", err)
	}

	// Calculate total
	total := usersResult.Total + postsResult.Total + hashtagsResult.Total

	return &SearchResult{
		Users:       usersResult.Users,
		Posts:       postsResult.Posts,
		Hashtags:    hashtagsResult.Hashtags,
		Total:       total,
		HasUsers:    usersResult.HasUsers,
		HasPosts:    postsResult.HasPosts,
		HasHashtags: hashtagsResult.HasHashtags,
	}, nil
}
