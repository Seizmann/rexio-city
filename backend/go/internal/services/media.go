package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// MediaService handles media uploads
type MediaService struct {
	s3Client *s3.S3
	bucket   string
	endpoint string
	url      string
}

// NewMediaService creates a new media service
func NewMediaService(endpoint, bucket, accessKey, secretKey, url string) *MediaService {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("auto"),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		logError("failed to create S3 session: %v", err)
	}

	var s3Client *s3.S3
	if sess != nil {
		s3Client = s3.New(sess)
	}

	return &MediaService{
		s3Client: s3Client,
		bucket:   bucket,
		endpoint: endpoint,
		url:      url,
	}
}

// UploadResult contains upload result info
type UploadResult struct {
	URL  string `json:"url"`
	Type string `json:"type"` // photo, video, voice
	Size int64  `json:"size"`
}

// UploadMedia uploads a media file with local disk fallback if R2/S3 fails
func (s *MediaService) UploadMedia(file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// Validate file size (max 50MB)
	maxSize := int64(50 * 1024 * 1024)
	if header.Size > maxSize {
		return nil, fmt.Errorf("file size exceeds 50MB limit")
	}

	// Determine media type
	contentType := header.Header.Get("Content-Type")
	mediaType := s.getMediaType(contentType)
	if mediaType == "" {
		// Fallback for image extensions
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".gif" {
			mediaType = "photo"
			contentType = "image/" + strings.TrimPrefix(ext, ".")
		} else {
			return nil, fmt.Errorf("unsupported file type: %s", contentType)
		}
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), header.Size, ext)
	key := fmt.Sprintf("media/%s", filename)

	// Try S3/R2 upload if client is available
	if s.s3Client != nil && s.bucket != "" {
		_, s3Err := s.s3Client.PutObject(&s3.PutObjectInput{
			Bucket:        aws.String(s.bucket),
			Key:           aws.String(key),
			Body:          bytes.NewReader(content),
			ContentType:   aws.String(contentType),
			ContentLength: aws.Int64(int64(len(content))),
		})
		if s3Err == nil {
			mediaURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(s.url, "/"), key)
			return &UploadResult{
				URL:  mediaURL,
				Type: mediaType,
				Size: int64(len(content)),
			}, nil
		}
	}

	// Fallback to local disk storage in ./uploads/media/
	uploadsDir := filepath.Join(".", "uploads", "media")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}

	filePath := filepath.Join(uploadsDir, filename)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to save local file: %w", err)
	}

	mediaURL := fmt.Sprintf("/uploads/media/%s", filename)
	return &UploadResult{
		URL:  mediaURL,
		Type: mediaType,
		Size: int64(len(content)),
	}, nil
}

// DeleteMedia deletes a media file
func (s *MediaService) DeleteMedia(key string) error {
	if s.s3Client != nil {
		_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		})
		return err
	}
	return nil
}

// getMediaType determines the media type from content type
func (s *MediaService) getMediaType(contentType string) string {
	if strings.HasPrefix(contentType, "image/") {
		return "photo"
	}
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}
	if strings.HasPrefix(contentType, "audio/") {
		return "voice"
	}
	return ""
}

func logError(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
