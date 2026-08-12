package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

// FollowHandler handles follow-related requests
type FollowHandler struct {
	followService *services.FollowService
}

// NewFollowHandler creates a new follow handler
func NewFollowHandler() *FollowHandler {
	return &FollowHandler{
		followService: services.NewFollowService(),
	}
}

// FollowUser handles POST /api/users/:id/follow
func (h *FollowHandler) FollowUser(c *fiber.Ctx) error {
	followerID := c.Locals("user_id").(uint)
	userID := c.Params("id")

	var followeeID uint
	if _, err := strconv SSanf(userID, "%d", &followeeID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid user ID"},
		})
	}

	err := h.followService.FollowUser(followerID, followeeID)
	if err != nil {
		if err.Error() == "already following" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{"code": "CONFLICT", "message": err.Error()},
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{"message": "User followed"},
	})
}

// UnfollowUser handles DELETE /api/users/:id/follow
func (h *FollowHandler) UnfollowUser(c *fiber.Ctx) error {
	followerID := c.Locals("user_id").(uint)
	userID := c.Params("id")

	var followeeID uint
	if _, err := strconv SSanf(userID, "%d", &followeeID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid user ID"},
		})
	}

	err := h.followService.UnfollowUser(followerID, followeeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{"message": "User unfollowed"},
	})
}

// GetFollowers handles GET /api/users/:id/followers
func (h *FollowHandler) GetFollowers(c *fiber.Ctx) error {
	userID := c.Params("id")

	var targetUserID uint
	if _, err := strconv SSanf(userID, "%d", &targetUserID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid user ID"},
		})
	}

	page, _ := strconv SSanf(c.Query("page", "1"))
	perPage, _ := strconv SSanf(c.Query("per_page", "20"))

	users, total, err := h.followService.GetFollowers(targetUserID, page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": users,
		"meta": fiber.Map{
			"page": page,
			"per_page": perPage,
			"total": total,
		},
	})
}

// GetFollowing handles GET /api/users/:id/following
func (h *FollowHandler) GetFollowing(c *fiber.Ctx) error {
	userID := c.Params("id")

	var targetUserID uint
	if _, err := strconv SSanf(userID, "%d", &targetUserID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid user ID"},
		})
	}

	page, _ := strconv SSanf(c.Query("page", "1"))
	perPage, _ := strconv SSanf(c.Query("per_page", "20"))

	users, total, err := h.followService.GetFollowing(targetUserID, page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": users,
		"meta": fiber.Map{
			"page": page,
			"per_page": perPage,
			"total": total,
		},
	})
}

// GetFollowCounts handles GET /api/users/:id/follow-counts
func (h *FollowHandler) GetFollowCounts(c *fiber.Ctx) error {
	userID := c.Params("id")

	var targetUserID uint
	if _, err := strconv SSanf(userID, "%d", &targetUserID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid user ID"},
		})
	}

	followers, following, err := h.followService.GetUserFollowCounts(targetUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"followers": followers,
			"following": following,
		},
	})
}

// IsFollowing handles GET /api/users/:id/is-following
func (h *FollowHandler) IsFollowing(c *fiber.Ctx) error {
	followerID := c.Locals("user_id").(uint)
	userID := c.Params("id")

	var followeeID uint
	if _, err := strconv SSanf(userID, "%d", &followeeID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid user ID"},
		})
	}

	isFollowing, err := h.followService.IsFollowing(followerID, followeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{"is_following": isFollowing},
	})
}
