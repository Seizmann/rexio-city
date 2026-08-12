package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

// FeedHandler handles feed-related requests
type FeedHandler struct {
	feedService *services.FeedService
}

// NewFeedHandler creates a new feed service
func NewFeedHandler() *FeedHandler {
	return &FeedHandler{
		feedService: services.NewFeedService(),
	}
}

// ListFeed handles GET /api/feed
func (h *FeedHandler) ListFeed(c *fiber.Ctx) error {
	var userID uint
	if u, ok := c.Locals("user_id").(uint); ok {
		userID = u
	}

	tab := c.Query("tab", "foryou")
	if tab != "following" && tab != "foryou" {
		tab = "foryou"
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	input := services.ListFeedInput{
		UserID:  userID,
		Tab:     tab,
		Page:    page,
		PerPage: perPage,
	}

	result, err := h.feedService.ListFeed(input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result.Posts,
		"meta": fiber.Map{
			"page":     page,
			"per_page": perPage,
			"total":    result.Total,
			"tab":      tab,
		},
	})
}
