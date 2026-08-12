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

	// Media upload (no auth required for testing, add auth later)
	mediaHandler := handlers.NewMediaHandler(
		cfg.MediaEndpoint,
		cfg.MediaBucket,
		cfg.MediaAccessKey,
		cfg.MediaSecretKey,
		cfg.MediaURL,
	)
	media := app.Group("/api/media")
	media.Post("/upload", mediaHandler.UploadMedia)

	// Protected routes (auth required)
	protected := app.Group("/api")
	protected.Use(middleware.Auth(cfg.JWTSecret))

	// User routes
	userHandler := handlers.NewUserHandler()
	protected.Get("/users/me", userHandler.GetCurrentUser)
	protected.Patch("/users/me", userHandler.UpdateUser)
	protected.Get("/users/:username", userHandler.GetUser)
	protected.Get("/search", userHandler.SearchUsers)

	// Post routes
	postHandler := handlers.NewPostHandler()
	protected.Post("/posts", postHandler.CreatePost)
	protected.Get("/posts", postHandler.ListPosts)
	protected.Get("/posts/:id", postHandler.GetPost)
	protected.Delete("/posts/:id", postHandler.DeletePost)
	protected.Post("/posts/:id/like", postHandler.LikePost)
	protected.Delete("/posts/:id/like", postHandler.UnlikePost)
	protected.Post("/posts/:id/comments", postHandler.CommentOnPost)
	protected.Get("/posts/:id/comments", postHandler.GetPostComments)
	protected.Post("/posts/:id/repost", postHandler.RepostPost)
	protected.Delete("/posts/:id/repost", postHandler.UnrepostPost)
	protected.Post("/posts/:id/bookmark", postHandler.BookmarkPost)
	protected.Delete("/posts/:id/bookmark", postHandler.UnbookmarkPost)

	// Feed routes
	feedHandler := handlers.NewFeedHandler()
	protected.Get("/feed", feedHandler.ListFeed)

	// Follow routes
	followHandler := handlers.NewFollowHandler()
	protected.Post("/users/:id/follow", followHandler.FollowUser)
	protected.Delete("/users/:id/follow", followHandler.UnfollowUser)
	protected.Get("/users/:id/followers", followHandler.GetFollowers)
	protected.Get("/users/:id/following", followHandler.GetFollowing)
	protected.Get("/users/:id/follow-counts", followHandler.GetFollowCounts)
	protected.Get("/users/:id/is-following", followHandler.IsFollowing)

	// DM routes
	dmHandler := handlers.NewDMHandler()
	protected.Get("/dm/conversations", dmHandler.GetConversations)
	protected.Post("/dm/conversations", dmHandler.CreateConversation)
	protected.Get("/dm/conversations/:id/messages", dmHandler.GetMessages)
	protected.Post("/dm/conversations/:id/messages", dmHandler.SendMessage)

	// WebSocket for DMs
	app.Get("/ws/dm", dmHandler.ConnectWS)

	// Notification routes
	notificationHandler := handlers.NewNotificationHandler()
	protected.Get("/notifications", notificationHandler.GetNotifications)
	protected.Put("/notifications/:id/read", notificationHandler.MarkAsRead)
	protected.Put("/notifications/read-all", notificationHandler.MarkAllAsRead)
	protected.Get("/notifications/unread-count", notificationHandler.GetUnreadCount)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
