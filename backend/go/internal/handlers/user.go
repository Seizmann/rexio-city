package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(),
	}
}

// GetUser handles GET /api/users/:username
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	username := c.Params("username")
	var currentUserID uint
	if val := c.Locals("user_id"); val != nil {
		if id, ok := val.(uint); ok {
			currentUserID = id
		}
	}

	result, err := h.userService.GetUserByUsername(username, currentUserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "NOT_FOUND", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GetCurrentUser handles GET /api/users/me
func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	var userID uint
	if val := c.Locals("user_id"); val != nil {
		if id, ok := val.(uint); ok {
			userID = id
		}
	}
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "UNAUTHORIZED", "message": "Unauthorized"},
		})
	}

	result, err := h.userService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "NOT_FOUND", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// UpdateUser handles PATCH /api/users/me
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	var userID uint
	if val := c.Locals("user_id"); val != nil {
		if id, ok := val.(uint); ok {
			userID = id
		}
	}
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "UNAUTHORIZED", "message": "Unauthorized"},
		})
	}

	var input services.UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	result, err := h.userService.UpdateUser(userID, input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// SearchUsers is deprecated — use SearchHandler.Search instead.
// Kept for backwards compatibility.
func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	return c.Redirect("/api/search?q="+c.Query("q")+"&type=user", fiber.StatusMovedPermanently)
}
