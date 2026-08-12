package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/middleware"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(),
	}
}

// SignupRequest represents a signup request body
type SignupRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// LoginRequest represents a login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents a refresh request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Signup handles user registration
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var input SignupRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	result, err := h.authService.Signup(services.SignupInput{
		Username:    input.Username,
		Email:       input.Email,
		Password:    input.Password,
		DisplayName: input.DisplayName,
	})
	if err != nil {
		status := fiber.StatusBadRequest
		if err.Error() == "username already taken" || err.Error() == "email already registered" {
			status = fiber.StatusConflict
		}
		return c.JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user":         result.User,
			"access_token": result.AccessToken,
			"refresh_token": result.RefreshToken,
			"expires_in":   result.ExpiresIn,
		},
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	result, err := h.authService.Login(services.LoginInput{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "AUTH_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user":         result.User,
			"access_token": result.AccessToken,
			"refresh_token": result.RefreshToken,
			"expires_in":   result.ExpiresIn,
		},
	})
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var input RefreshRequest
	if err := c.BodyParser(&input); err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	result, err := h.authService.RefreshToken(services.RefreshInput{
		RefreshToken: input.RefreshToken,
	})
	if err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "AUTH_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"access_token": result.AccessToken,
			"expires_in":   result.ExpiresIn,
		},
	})
}

// Health handles health checks
func (h *AuthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"status": "healthy"},
	})
}
