package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUploadMediaBodyLimit tests that the upload handler returns an error when body exceeds limit
func TestUploadMediaBodyLimit(t *testing.T) {
	// Create a handler with minimal config (no actual S3/R2 connection)
	h := NewMediaHandler("http://localhost:9999", "test-bucket", "test-key", "test-secret", "http://localhost:9999")

	app := fiber.New(fiber.Config{
		BodyLimit: 30 * 1024 * 1024, // 30MB limit as per PRD
	})
	app.Post("/api/media/upload", h.UploadMedia)

	// Create a multipart request that's larger than the limit
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "large_test.jpg")
	require.NoError(t, err)

	// Write 35MB of data (exceeds 30MB limit)
	largeData := bytes.Repeat([]byte("x"), 35*1024*1024)
	_, err = part.Write(largeData)
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)

	// When body exceeds limit, Fiber returns an error (not a response)
	// This is expected behavior - the request is rejected before reaching the handler
	if err != nil {
		assert.Contains(t, err.Error(), "body size exceeds")
		return
	}

	// If we get a response, it should be 413
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}

// TestUploadMediaSmallFile tests successful upload of a small file
func TestUploadMediaSmallFile(t *testing.T) {
	h := NewMediaHandler("http://localhost:9999", "test-bucket", "test-key", "test-secret", "http://localhost:9999")

	app := fiber.New(fiber.Config{
		BodyLimit: 30 * 1024 * 1024,
	})
	app.Post("/api/media/upload", h.UploadMedia)

	// Create a small multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "small_test.jpg")
	require.NoError(t, err)

	// Write a small amount of data (5 bytes)
	smallData := []byte("hello")
	_, err = part.Write(smallData)
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Should succeed (falls back to local disk since no real S3)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify response is valid JSON
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
}
