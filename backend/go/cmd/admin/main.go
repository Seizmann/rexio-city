package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/seizmann/rexio-city/backend/go/internal/config"
	"github.com/seizmann/rexio-city/backend/go/internal/middleware"
)

func main() {
	cfg := config.LoadAdmin()

	app := fiber.New()

	// CORS — only allow admin origin
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AdminURL,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Auth middleware for admin
	app.Use(middleware.AuthAdmin(cfg.JWTSecret))

	// Routes
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// TODO: Add admin routes

	log.Printf("Admin server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
