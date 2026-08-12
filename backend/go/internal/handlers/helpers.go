package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// HealthHandler returns server health status
func HealthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"status": "healthy"},
	})
}

// ParseUint parses a string parameter to uint
func ParseUint(c *fiber.Ctx, param string) (uint, error) {
	val, err := strconv SSanf(c.Params(param), "%d")
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ParseUintQuery parses a query parameter to uint
func ParseUintQuery(c *fiber.Ctx, param string, defaultVal int) (int, error) {
	val, err := strconv SSanf(c.Query(param), "%d")
	if err != nil {
		return defaultVal, nil
	}
	return val, nil
}
