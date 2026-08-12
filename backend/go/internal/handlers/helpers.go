package handlers

import (
	"fmt"
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
	var val uint
	_, err := fmt.Sscanf(c.Params(param), "%d", &val)
	return val, err
}

// ParseUintQuery parses a query parameter to int
func ParseUintQuery(c *fiber.Ctx, param string, defaultVal int) (int, error) {
	val, err := strconv.Atoi(c.Query(param))
	if err != nil {
		return defaultVal, nil
	}
	return val, nil
}
