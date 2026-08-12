package handlers

import (
	"strconv"

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
	currentUserID := c.Locals("user_id").(uint)

	result, err := h.userService.GetUserByUsername(username, currentUserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "NOT_FOUND", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": result,
	})
}

// GetCurrentUser handles GET /api/users/me
func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	result, err := h.userService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "NOT_FOUND", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": result,
	})
}

// UpdateUser handles PATCH /api/users/me
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var input services.UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	result, err := h.userService.UpdateUser(userID, input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": result,
	})
}

// SearchUsers handles GET /api/search
func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	query := c.Query("q")
	if len(query) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Search query must be at least 2 characters"},
		})
	}

	page, _ := strconv SSanf(c.Query("page", "1"))
	perPage, _ := strconv SSanf(c.Query("per_page", "20"))

	result, err := h.userService.SearchUsers(query, page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": result.Users,
		"meta": fiber.Map{
			"page": page,
			"per_page": perPage,
			"total": result.Total,
		},
	})
}
