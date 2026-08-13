package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

// SearchHandler handles GET /api/search
type SearchHandler struct {
	searchService *services.SearchService
	rateLimiter   *simpleRateLimiter
}

// NewSearchHandler creates a new search handler
func NewSearchHandler() *SearchHandler {
	return &SearchHandler{
		searchService: services.NewSearchService(),
		rateLimiter:   newSimpleRateLimiter(30, time.Minute), // 30 requests per minute per IP
	}
}

// searchQuery represents the parsed query parameters
type searchQuery struct {
	Q       string `json:"q"`
	Type    string `json:"type"` // "user", "post", "hashtag", or "" for all
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
}

// validateSearchQuery validates and normalizes search parameters
// Returns error via fiber.Map on invalid input, nil on success
func validateSearchQuery(c *fiber.Ctx) (*searchQuery, error) {
	q := strings.TrimSpace(c.Query("q", ""))
	if len(q) == 0 {
		return nil, c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Search query cannot be empty"},
		})
	}
	if len(q) > 100 {
		return nil, c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Search query too long (max 100 characters)"},
		})
	}

	searchType := strings.ToLower(strings.TrimSpace(c.Query("type", "")))
	// Validate type
	if searchType != "" && searchType != "user" && searchType != "post" && searchType != "hashtag" {
		return nil, c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid type parameter. Must be 'user', 'post', or 'hashtag'"},
		})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.Query("per_page", "10"))
	if err != nil || perPage < 1 || perPage > 20 {
		perPage = 10
	}

	return &searchQuery{
		Q:       q,
		Type:    searchType,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// Search handles GET /api/search
func (h *SearchHandler) Search(c *fiber.Ctx) error {
	// Rate limit check
	if !h.rateLimiter.Allow(c.IP()) {
		return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "RATE_LIMITED", "message": "Search requests limited. Try again in a minute."},
		})
	}

	// Validate input
	query, err := validateSearchQuery(c)
	if err != nil {
		return err // Already sent response
	}

	// Build search result
	var result *services.SearchResult
	if query.Type == "" {
		// Combined search — return top N of each type
		result, err = h.searchService.SearchCombined(query.Q, query.PerPage)
	} else if query.Type == "user" {
		result, err = h.searchService.SearchUsers(query.Q, query.Page, query.PerPage)
	} else if query.Type == "post" {
		result, err = h.searchService.SearchPosts(query.Q, query.Page, query.PerPage)
	} else if query.Type == "hashtag" {
		result, err = h.searchService.SearchHashtags(query.Q, query.Page, query.PerPage)
	}

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
		"meta": fiber.Map{
			"page":     query.Page,
			"per_page": query.PerPage,
			"total":    result.Total,
			"query":    query.Q,
			"type":     query.Type,
		},
	})
}

// simpleRateLimiter is a basic in-memory rate limiter (IP-based, sliding window)
type simpleRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newSimpleRateLimiter(limit int, window time.Duration) *simpleRateLimiter {
	return &simpleRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (r *simpleRateLimiter) Allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Get existing requests for this IP
	reqs := r.requests[ip]
	// Filter out old requests
	valid := reqs[:0]
	for _, t := range reqs {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= r.limit {
		r.requests[ip] = valid
		return false
	}

	r.requests[ip] = append(valid, now)
	return true
}
