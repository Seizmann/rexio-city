package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPresignedUploadRequestValidation tests that presigned upload request validation works
func TestPresignedUploadRequestValidation(t *testing.T) {
	req := PresignedUploadRequest{
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
		Size:        5 * 1024 * 1024,
	}

	// Verify request structure is valid
	assert.Equal(t, "test.jpg", req.Filename)
	assert.Equal(t, "image/jpeg", req.ContentType)
	assert.Equal(t, int64(5*1024*1024), req.Size)
}

// TestBuildMediaURL tests the media URL builder
func TestBuildMediaURL(t *testing.T) {
	svc := NewMediaService("https://test.r2.cloudflarestorage.com", "test-bucket", "test-key", "test-secret", "https://cdn-test.rexio.pro")
	
	url := svc.BuildMediaURL("media/123_456.jpg")
	assert.Equal(t, "https://cdn-test.rexio.pro/media/123_456.jpg", url)
}

// TestGetMediaType tests media type detection
func TestGetMediaType(t *testing.T) {
	svc := NewMediaService("", "", "", "", "")
	
	assert.Equal(t, "photo", svc.getMediaType("image/jpeg"))
	assert.Equal(t, "photo", svc.getMediaType("image/png"))
	assert.Equal(t, "video", svc.getMediaType("video/mp4"))
	assert.Equal(t, "voice", svc.getMediaType("audio/mp3"))
	assert.Equal(t, "", svc.getMediaType("application/octet-stream"))
}
