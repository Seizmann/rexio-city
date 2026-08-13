package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

// MediaHandler handles media uploads
type MediaHandler struct {
	mediaService *services.MediaService
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(endpoint, bucket, accessKey, secretKey, url string) *MediaHandler {
	return &MediaHandler{
		mediaService: services.NewMediaService(endpoint, bucket, accessKey, secretKey, url),
	}
}

// UploadMedia handles POST /api/media/upload (legacy direct upload)
func (h *MediaHandler) UploadMedia(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "File is required"},
		})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "UPLOAD_ERROR", "message": err.Error()},
		})
	}
	defer f.Close()

	result, err := h.mediaService.UploadMedia(f, file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "UPLOAD_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GeneratePresignedURL handles POST /api/media/upload-request
// Returns a presigned URL for direct R2 upload (bypasses Vercel 4.5MB limit)
func (h *MediaHandler) GeneratePresignedURL(c *fiber.Ctx) error {
	var req services.PresignedUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	if req.Filename == "" || req.ContentType == "" || req.Size <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "filename, content_type, and size are required"},
		})
	}

	result, err := h.mediaService.GeneratePresignedURL(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "UPLOAD_ERROR", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// CompleteUpload handles POST /api/media/upload-complete
// Verifies the upload and returns the media URL for post creation
func (h *MediaHandler) CompleteUpload(c *fiber.Ctx) error {
	var req struct {
		Key  string `json:"key"`
		Size int64  `json:"size"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	if req.Key == "" || req.Size <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "key and size are required"},
		})
	}

	// Verify the uploaded file size
	maxSize := int64(30 * 1024 * 1024)
	if err := h.mediaService.VerifyUpload(req.Key, maxSize); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "UPLOAD_ERROR", "message": err.Error()},
		})
	}

	// Return the media URL (the key is used to construct the URL)
	mediaURL := h.mediaService.BuildMediaURL(req.Key)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"url":    mediaURL,
			"key":    req.Key,
			"size":   req.Size,
		},
	})
}
