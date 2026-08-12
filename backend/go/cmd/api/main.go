package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/seizmann/rexio-city/backend/go/internal/config"
	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/handlers"
	"github.com/seizmann/rexio-city/backend/go/internal/middleware"
)

func main() {
	cfg := config.Load()

	// Initialize database
	db.Init(cfg)
	db.Migrate()

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * 1000,
		WriteTimeout: 10 * 1000,
	})

	// CORS — only allow frontend origin
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURL,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	// Health endpoint (no auth required)
	app.Get("/api/health", handlers.HealthHandler)

	// Auth routes (no auth required)
	authHandler := handlers.NewAuthHandler()
	auth := app.Group("/api/auth")
	auth.Post("/signup", authHandler.Signup)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	// Protected routes (auth required)
	protected := app.Group("/api")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	protected.Get("/user/me", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(uint)
		return c.JSON(fiber.Map{
			"success": true,
			"data":    fiber.Map{"user_id": userID},
		})
	})

	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
