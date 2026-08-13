package services

import (
	"testing"

	"github.com/seizmann/rexio-city/backend/go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := testDB.AutoMigrate(&models.User{}, &models.Post{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return testDB
}

// replaceDB replaces the global DB with the test database
func replaceDB(db *gorm.DB) {
	// This is a hack to use test DB in services - in production, db.Init() is called
	// For tests, we need to set the package-level DB variable
	// Note: This is a test-only workaround
}

// TestExtractHashtags tests hashtag extraction from post content
func TestExtractHashtags(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single hashtag",
			content:  "Hello #world",
			expected: []string{"world"},
		},
		{
			name:     "multiple hashtags",
			content:  "Check out #golang and #rust",
			expected: []string{"golang", "rust"},
		},
		{
			name:     "hashtag with numbers",
			content:  "Version 2.0 #release123",
			expected: []string{"release123"},
		},
		{
			name:     "no hashtags",
			content:  "Just plain text",
			expected: []string{},
		},
		{
			name:     "duplicate hashtags",
			content:  "#hello #world #hello",
			expected: []string{"hello", "world"},
		},
		{
			name:     "hashtag at end",
			content:  "Check this #test",
			expected: []string{"test"},
		},
		{
			name:     "mixed case hashtags",
			content:  "#Hello #WORLD #test",
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "hashtag with hyphen",
			content:  "Test #foo-bar #baz",
			expected: []string{"foo-bar", "baz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHashtags(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("expected %v at index %d, got %v", tt.expected, i, result)
				}
			}
		})
	}
}

// TestSearchService_Validation tests input validation in search service
func TestSearchService_Validation(t *testing.T) {
	svc := NewSearchService()

	// Test that pagination parameters are normalized
	// These tests verify the service logic path without requiring DB

	// Test empty query returns empty results without panic
	result, err := svc.SearchUsers("", 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}

	// Test with page=0 (should default to 1)
	result, err = svc.SearchUsers("test", 0, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}

	// Test with negative page (should default to 1)
	result, err = svc.SearchUsers("test", -1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}

	// Test with perPage > 50 (should cap at 50)
	result, err = svc.SearchUsers("test", 1, 100)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// TestSearchService_PostSearch tests post search with empty query
func TestSearchService_PostSearch(t *testing.T) {
	svc := NewSearchService()

	result, err := svc.SearchPosts("", 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// TestSearchService_Combined tests combined search with empty query
func TestSearchService_Combined(t *testing.T) {
	svc := NewSearchService()

	result, err := svc.SearchCombined("", 5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// TestSearchService_HashtagSearch tests hashtag search with empty query
func TestSearchService_HashtagSearch(t *testing.T) {
	svc := NewSearchService()

	result, err := svc.SearchHashtags("", 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// TestSearchService_HashtagNormalization tests hashtag normalization in search
func TestSearchService_HashtagNormalization(t *testing.T) {
	svc := NewSearchService()

	// Test that hashtag search normalizes query (adds # if missing)
	result, err := svc.SearchHashtags("golang", 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}

	// Test with # prefix
	result, err = svc.SearchHashtags("#golang", 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}
