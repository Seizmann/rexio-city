package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/config"
	"github.com/seizmann/rexio-city/backend/go/internal/middleware"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

// PostHandler handles post-related requests
type PostHandler struct {
	postService *services.PostService
}

// NewPostHandler creates a new post handler
func NewPostHandler() *PostHandler {
	return &PostHandler{
		postService: services.NewPostService(),
	}
}

// CreatePostRequest represents a create post request
type CreatePostRequest struct {
	Content    string   `json:"content"`
	MediaURLs  []string `json:"media_urls"`
	MediaTypes []string `json:"media_types"`
}

// CommentRequest represents a comment request
type CommentRequest struct {
	Content  string `json:"content"`
	ParentID *uint  `json:"parent_id"`
}

// RepostRequest represents a repost request
type RepostRequest struct {
	Comment *string `json:"comment"`
}

// CreatePost handles POST /api/posts
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var input CreatePostRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	result, err := h.postService.CreatePost(services.CreatePostInput{
		UserID:     userID,
		Content:    input.Content,
		MediaURLs:  input.MediaURLs,
		MediaTypes: input.MediaTypes,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GetPost handles GET /api/posts/:id (publicly accessible, accepts public_id or numeric ID)
func (h *PostHandler) GetPost(c *fiber.Ctx) error {
	postID := c.Params("id")
	var userID uint
	if u, ok := c.Locals("user_id").(uint); ok {
		userID = u
	} else {
		cfg := config.Load()
		userID = middleware.ExtractUserIDFromHeader(c.Get("Authorization"), cfg.JWTSecret)
	}

	result, err := h.postService.GetPost(services.GetPostInput{
		Identifier: postID,
		UserID:     userID,
	})
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

// ListPosts handles GET /api/posts (supports optional ?user_id= query filter)
func (h *PostHandler) ListPosts(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	filterUserID, _ := strconv.Atoi(c.Query("user_id", "0"))

	var viewerID uint
	if u, ok := c.Locals("user_id").(uint); ok {
		viewerID = u
	} else {
		cfg := config.Load()
		viewerID = middleware.ExtractUserIDFromHeader(c.Get("Authorization"), cfg.JWTSecret)
	}

	input := services.ListPostsInput{
		FilterUserID: uint(filterUserID),
		ViewerUserID: viewerID,
		Page:         page,
		PerPage:      perPage,
	}

	result, err := h.postService.ListPosts(input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result.Posts,
		"meta": fiber.Map{
			"page":     page,
			"per_page": perPage,
			"total":    result.Total,
		},
	})
}

// DeletePost handles DELETE /api/posts/:id
func (h *PostHandler) DeletePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID := c.Params("id")

	err := h.postService.DeletePost(postID, userID)
	if err != nil {
		statusCode := fiber.StatusBadRequest
		if err.Error() == "permission denied" {
			statusCode = fiber.StatusForbidden
		}
		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Post deleted"},
	})
}

// LikePost handles POST /api/posts/:id/like
func (h *PostHandler) LikePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID := c.Params("id")

	err := h.postService.LikePost(postID, userID)
	if err != nil {
		if err.Error() == "already liked" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "CONFLICT", "message": err.Error()},
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Post liked"},
	})
}

// UnlikePost handles DELETE /api/posts/:id/like
func (h *PostHandler) UnlikePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID := c.Params("id")

	err := h.postService.UnlikePost(postID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Post unliked"},
	})
}

// CommentOnPost handles POST /api/posts/:id/comments
func (h *PostHandler) CommentOnPost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID := c.Params("id")

	var req CommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	result, err := h.postService.CommentOnPost(postID, userID, req.Content, req.ParentID)
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

// GetPostComments handles GET /api/posts/:id/comments (publicly accessible)
func (h *PostHandler) GetPostComments(c *fiber.Ctx) error {
	postID := c.Params("id")

	result, err := h.postService.GetPostComments(postID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "SERVER_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// RepostPost handles POST /api/posts/:id/repost
func (h *PostHandler) RepostPost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID := c.Params("id")

	var req RepostRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
			})
		}
	}

	result, err := h.postService.RepostPost(postID, userID, req.Comment)
	if err != nil {
		if err.Error() == "already reposted" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "CONFLICT", "message": err.Error()},
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// UnrepostPost removes a repost
func (h *PostHandler) UnrepostPost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID := c.Params("id")

	err := h.postService.UnrepostPost(postID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Post unreposted"},
	})
}

// BookmarkPost handles POST /api/posts/:id/bookmark
func (h *PostHandler) BookmarkPost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID := c.Params("id")

	err := h.postService.BookmarkPost(postID, userID)
	if err != nil {
		if err.Error() == "already bookmarked" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "CONFLICT", "message": err.Error()},
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Post bookmarked"},
	})
}

// UnbookmarkPost removes a bookmark
func (h *PostHandler) UnbookmarkPost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID := c.Params("id")

	err := h.postService.UnbookmarkPost(postID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"message": "Post unbookmarked"},
	})
}
