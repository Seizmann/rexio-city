package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

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
		Region:         aws.String("auto"),
		Endpoint:       aws.String(endpoint),
		Credentials:    credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		panic(fmt.Errorf("failed to create S3 session: %w", err))
	}

	return &MediaService{
		s3Client: s3.New(sess),
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

// UploadMedia uploads a media file
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
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	key := fmt.Sprintf("media/%d%s", header.Size, ext)

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Upload to S3/MinIO
	_, err = s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(content),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(content))),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate URL
	url := fmt.Sprintf("%s/%s/%s", s.url, s.bucket, key)

	return &UploadResult{
		URL:  url,
		Type: mediaType,
		Size: int64(len(content)),
	}, nil
}

// DeleteMedia deletes a media file
func (s *MediaService) DeleteMedia(key string) error {
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
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
