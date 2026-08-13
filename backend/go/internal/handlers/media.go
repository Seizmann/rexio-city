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

// UploadMedia handles POST /api/media/upload
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
