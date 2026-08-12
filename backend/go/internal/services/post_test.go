package services_test

import (
	"testing"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupPostTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := testDB.AutoMigrate(&models.User{}, &models.Post{}, &models.PostMedia{}); err != nil {
		t.Fatalf("failed to migrate post test db: %v", err)
	}

	user := models.User{
		Username:     "mediauser",
		PasswordHash: "dummyhash",
	}
	testDB.Create(&user)

	return testDB
}

func TestPostMediaAttachment(t *testing.T) {
	testDB := setupPostTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()

	post := models.Post{
		PublicID:  "testpublicid12345",
		UserID:    1,
		Content:   "Post with image and video",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := testDB.Create(&post).Error; err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	mediaItems := []models.PostMedia{
		{
			PostID:    post.ID,
			MediaURL:  "/uploads/media/photo1.jpg",
			MediaType: "photo",
			Order:     0,
			CreatedAt: time.Now(),
		},
		{
			PostID:    post.ID,
			MediaURL:  "/uploads/media/video1.mp4",
			MediaType: "video",
			Order:     1,
			CreatedAt: time.Now(),
		},
	}

	for _, m := range mediaItems {
		if err := testDB.Create(&m).Error; err != nil {
			t.Fatalf("failed to create media item: %v", err)
		}
	}

	var fetchedPost models.Post
	if err := testDB.Preload("User").Preload("Media").First(&fetchedPost, post.ID).Error; err != nil {
		t.Fatalf("failed to fetch post with preloaded media: %v", err)
	}

	if len(fetchedPost.Media) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(fetchedPost.Media))
	}
	if fetchedPost.Media[0].MediaType != "photo" || fetchedPost.Media[1].MediaType != "video" {
		t.Fatalf("media types mismatch: got %s and %s", fetchedPost.Media[0].MediaType, fetchedPost.Media[1].MediaType)
	}
}
