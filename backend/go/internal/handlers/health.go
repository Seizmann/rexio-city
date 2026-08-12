package handlers

import "github.com/gofiber/fiber/v2"

// HealthHandler returns server health status
func HealthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"status": "healthy"},
	})
}
