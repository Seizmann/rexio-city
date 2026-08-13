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

// PresignedUploadRequest holds metadata for presigned URL generation
type PresignedUploadRequest struct {
	Filename   string `json:"filename"`
	ContentType string `json:"content_type"`
	Size       int64  `json:"size"`
}

// PresignedUploadResult contains the presigned URL for direct R2 upload
type PresignedUploadResult struct {
	URL           string `json:"url"`      // Presigned PUT URL
	MediaURL      string `json:"media_url"` // Final CDN URL after upload
	Key           string `json:"key"`      // R2 object key
	MaxSize       int64  `json:"max_size"` // Max allowed size in bytes (30MB)
}

// UploadMedia uploads a media file with local disk fallback if R2/S3 fails
func (s *MediaService) UploadMedia(file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// Validate file size (max 500MB for video attachments)
	maxSize := int64(500 * 1024 * 1024)
	if header.Size > maxSize {
		return nil, fmt.Errorf("file size exceeds 500MB limit")
	}

	// Determine media type
	contentType := header.Header.Get("Content-Type")
	mediaType := s.getMediaType(contentType)
	if mediaType == "" {
		// Fallback for file extensions
		ext := strings.ToLower(filepath.Ext(header.Filename))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".svg":
			mediaType = "photo"
			contentType = "image/" + strings.TrimPrefix(ext, ".")
		case ".mp4", ".mov", ".webm", ".mkv", ".avi":
			mediaType = "video"
			contentType = "video/" + strings.TrimPrefix(ext, ".")
		case ".mp3", ".wav", ".ogg", ".m4a":
			mediaType = "voice"
			contentType = "audio/" + strings.TrimPrefix(ext, ".")
		default:
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

// GeneratePresignedURL generates a presigned PUT URL for direct R2 upload
// The browser can then PUT the file directly to R2 without going through Vercel
func (s *MediaService) GeneratePresignedURL(req PresignedUploadRequest) (*PresignedUploadResult, error) {
	if s.s3Client == nil || s.bucket == "" {
		return nil, fmt.Errorf("R2 client not configured")
	}

	// Validate file type
	mediaType := s.getMediaType(req.ContentType)
	if mediaType == "" {
		// Fallback for file extensions
		ext := strings.ToLower(filepath.Ext(req.Filename))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".svg":
			mediaType = "photo"
		case ".mp4", ".mov", ".webm", ".mkv", ".avi":
			mediaType = "video"
		case ".mp3", ".wav", ".ogg", ".m4a":
			mediaType = "voice"
		default:
			return nil, fmt.Errorf("unsupported file type: %s", req.ContentType)
		}
	}

	// Validate size (30MB PRD limit)
	maxSize := int64(30 * 1024 * 1024)
	if req.Size > maxSize {
		return nil, fmt.Errorf("file size exceeds 30MB limit")
	}

	// Generate unique key following PRD Section 6 pattern
	ext := filepath.Ext(req.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("media/%d_%d%s", time.Now().UnixNano(), req.Size, ext)

	// Generate presigned PUT URL (expires in 10 minutes)
	reqObj, _ := s.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(req.ContentType),
	})
	reqURL, err := reqObj.Presign(10 * time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Build final CDN URL
	mediaURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(s.url, "/"), key)

	return &PresignedUploadResult{
		URL:       reqURL,
		MediaURL:  mediaURL,
		Key:       key,
		MaxSize:   maxSize,
	}, nil
}

// VerifyUpload checks the actual object size in R2 after direct upload
func (s *MediaService) VerifyUpload(key string, maxSize int64) error {
	if s.s3Client == nil || s.bucket == "" {
		return fmt.Errorf("R2 client not configured")
	}

	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	output, err := s.s3Client.HeadObject(headInput)
	if err != nil {
		return fmt.Errorf("failed to verify upload: %w", err)
	}

	actualSize := aws.Int64Value(output.ContentLength)
	if actualSize > maxSize {
		// Clean up the oversized file
		s.DeleteMedia(key)
		return fmt.Errorf("upload exceeds 30MB limit (actual: %d bytes)", actualSize)
	}

	return nil
}

// BuildMediaURL constructs the CDN URL for a given key
func (s *MediaService) BuildMediaURL(key string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.url, "/"), key)
}

func logError(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
