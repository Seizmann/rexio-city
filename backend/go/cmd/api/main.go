package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/seizmann/rexio-city/backend/go/internal/config"
	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/handlers"
	"github.com/seizmann/rexio-city/backend/go/internal/middleware"
)

func main() {
	cfg := config.Load()

	// Initialize database (also runs idempotent schema setup)
	db.Init(cfg)
	db.Migrate()

	app := fiber.New(fiber.Config{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		BodyLimit:    500 * 1024 * 1024, // 500MB max request body for video uploads
	})

	// CORS — only allow the Next.js frontend origin (never "*").
	// AllowCredentials=true is required for cookies to be sent cross-origin.
	// X-CSRF-Token header is explicitly allowed for CSRF double-submit pattern.
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURL,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-CSRF-Token",
		ExposeHeaders:    "Set-Cookie",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	// Health endpoint (no auth required)
	app.Get("/api/health", handlers.HealthHandler)

	// Static uploaded files
	app.Static("/uploads", "./uploads")

	// ── Auth routes (no auth required) ───────────────────────────
	authHandler := handlers.NewAuthHandler()
	auth := app.Group("/api/auth")
	auth.Post("/signup", authHandler.Signup)
	auth.Post("/login", authHandler.Login)
	// Refresh reads the httpOnly cookie — no body params needed
	auth.Post("/refresh", authHandler.Refresh)

	// ── Media upload ──────────────────────────────────────────────
	mediaHandler := handlers.NewMediaHandler(
		cfg.MediaEndpoint,
		cfg.MediaBucket,
		cfg.MediaAccessKey,
		cfg.MediaSecretKey,
		cfg.MediaURL,
	)
	media := app.Group("/api/media")
	media.Post("/upload", mediaHandler.UploadMedia)
	media.Post("/upload-request", mediaHandler.GeneratePresignedURL)
	media.Post("/upload-complete", mediaHandler.CompleteUpload)

	// ── Public post endpoints (no auth required) ─────────────────
	postHandler := handlers.NewPostHandler()
	app.Get("/api/posts", postHandler.ListPosts)
	app.Get("/api/posts/:id", postHandler.GetPost)
	app.Get("/api/posts/:id/comments", postHandler.GetPostComments)

	userHandler := handlers.NewUserHandler()
	followHandler := handlers.NewFollowHandler()

	// ── Public user/profile endpoints (no auth required) ─────────
	// These must be registered BEFORE protected routes so they take precedence.
	// If registered after, the protected middleware catches them first.
	app.Get("/api/users/:username", userHandler.GetUser)
	app.Get("/api/users/:id/follow-counts", followHandler.GetFollowCounts)
	app.Get("/api/users/:id/is-following", followHandler.IsFollowing)
	app.Get("/api/users/:id/followers", followHandler.GetFollowers)
	app.Get("/api/users/:id/following", followHandler.GetFollowing)

	// ── Protected routes (JWT auth + CSRF protection) ─────────────
	protected := app.Group("/api")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	protected.Use(middleware.CSRF(cfg.CSRFSecret))

	// User routes (me must be registered BEFORE wildcard :username)
	protected.Get("/users/me", userHandler.GetCurrentUser)
	protected.Patch("/users/me", userHandler.UpdateUser)
	protected.Get("/search", userHandler.SearchUsers)

	// Session management (auth required)
	protected.Get("/auth/sessions", authHandler.ListSessions)
	protected.Post("/auth/sessions/:id/revoke", authHandler.RevokeSession)
	protected.Post("/auth/logout", authHandler.Logout)
	protected.Post("/auth/logout-all", authHandler.LogoutAll)

	// ── Public user/profile endpoints (no auth required) ─────────
	// These must be registered BEFORE protected routes so they take precedence.
	// If registered after, the protected middleware catches them first.
	app.Get("/api/users/:username", userHandler.GetUser)
	app.Get("/api/users/:id/follow-counts", followHandler.GetFollowCounts)
	app.Get("/api/users/:id/is-following", followHandler.IsFollowing)
	app.Get("/api/users/:id/followers", followHandler.GetFollowers)
	app.Get("/api/users/:id/following", followHandler.GetFollowing)

	// Post routes
	protected.Post("/posts", postHandler.CreatePost)
	protected.Delete("/posts/:id", postHandler.DeletePost)
	protected.Post("/posts/:id/like", postHandler.LikePost)
	protected.Delete("/posts/:id/like", postHandler.UnlikePost)
	protected.Post("/posts/:id/comments", postHandler.CommentOnPost)
	protected.Post("/posts/:id/repost", postHandler.RepostPost)
	protected.Delete("/posts/:id/repost", postHandler.UnrepostPost)
	protected.Post("/posts/:id/bookmark", postHandler.BookmarkPost)
	protected.Delete("/posts/:id/bookmark", postHandler.UnbookmarkPost)

	// Feed routes
	feedHandler := handlers.NewFeedHandler()
	protected.Get("/feed", feedHandler.ListFeed)

	// Follow routes (auth required for mutations, but GET follow-lists are public)
	// Note: GetFollowers and GetFollowing are already registered above as public routes.
	protected.Post("/users/:id/follow", followHandler.FollowUser)
	protected.Delete("/users/:id/follow", followHandler.UnfollowUser)

	// DM routes
	dmHandler := handlers.NewDMHandler()
	protected.Get("/dm/conversations", dmHandler.GetConversations)
	protected.Post("/dm/conversations", dmHandler.CreateConversation)
	protected.Get("/dm/conversations/:id/messages", dmHandler.GetMessages)
	protected.Post("/dm/conversations/:id/messages", dmHandler.SendMessage)

	// WebSocket for DMs (auth via token query param or header)
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
