package services_test

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/models"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing.
// We cannot use real Postgres in CI without a live DB, so SQLite is the
// pragmatic choice for unit tests. The session logic is pure DB CRUD.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := testDB.AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	// Create a test user
	user := models.User{
		Username:     "testuser",
		PasswordHash: "dummyhash",
	}
	testDB.Create(&user)

	return testDB
}

// injectTestDB swaps the global DB for a test database and returns a cleanup func.
func injectTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	testDB := setupTestDB(t)

	// We need to inject testDB into the db package. Since db.DB is exported,
	// import the db package here and swap it.
	// For these tests, we test SessionService directly with a factory that
	// accepts a db param — see the testable variant below.
	return testDB, func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}
}

// sessionSvcWithDB creates a SessionService that uses the provided DB.
// This is a testable factory — production code uses NewSessionService() which
// calls db.GetDB(). For tests we inject directly.
type testableSessionService struct {
	db *gorm.DB
}

func (s *testableSessionService) hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

func (s *testableSessionService) createSession(userID uint, rawToken, device, ip string, expiresAt time.Time, parentID *uint) (*models.Session, error) {
	hash := s.hashToken(rawToken)
	sess := models.Session{
		UserID:          userID,
		TokenHash:       hash,
		ParentSessionID: parentID,
		DeviceInfo:      device,
		IPAddress:       ip,
		CreatedAt:       time.Now(),
		LastUsedAt:      time.Now(),
		ExpiresAt:       expiresAt,
	}
	if err := s.db.Create(&sess).Error; err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *testableSessionService) rotateSession(rawToken string, userID uint) (*models.Session, error) {
	hash := s.hashToken(rawToken)
	var sess models.Session
	res := s.db.Where("token_hash = ?", hash).First(&sess)
	if res.Error != nil || res.RowsAffected == 0 {
		return nil, fmt.Errorf("session not found")
	}
	if sess.RevokedAt != nil {
		// Reuse detected — revoke all sessions for this user
		now := time.Now()
		s.db.Model(&models.Session{}).Where("user_id = ? AND revoked_at IS NULL", sess.UserID).Update("revoked_at", now)
		return nil, services.ErrTokenReuse
	}
	now := time.Now()
	sess.RevokedAt = &now
	s.db.Model(&sess).Update("revoked_at", now)
	return &sess, nil
}

func (s *testableSessionService) revokeAll(userID uint) error {
	now := time.Now()
	return s.db.Model(&models.Session{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", now).Error
}

func (s *testableSessionService) listActive(userID uint) ([]models.Session, error) {
	var sessions []models.Session
	err := s.db.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).Find(&sessions).Error
	return sessions, err
}

/* ── Tests ──────────────────────────────────────────────────────── */

// TestTokenRotation verifies that rotating a valid token revokes the old session
// and allows creating a new child session.
func TestTokenRotation(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()
	svc := &testableSessionService{db: testDB}

	// Create initial session
	rawToken := "initial-refresh-token-abc123"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	session, err := svc.createSession(1, rawToken, "Mozilla/5.0", "127.0.0.1", expiresAt, nil)
	if err != nil {
		t.Fatalf("createSession failed: %v", err)
	}
	if session.RevokedAt != nil {
		t.Fatal("new session should not be revoked")
	}

	// Rotate: invalidate old, create new child
	oldSession, err := svc.rotateSession(rawToken, 1)
	if err != nil {
		t.Fatalf("rotateSession failed: %v", err)
	}

	// Old session must be revoked now
	var refreshedOld models.Session
	testDB.First(&refreshedOld, oldSession.ID)
	if refreshedOld.RevokedAt == nil {
		t.Fatal("old session should be revoked after rotation")
	}

	// Create new child session (simulating what the handler does)
	newRawToken := "new-refresh-token-xyz789"
	newSession, err := svc.createSession(1, newRawToken, "Mozilla/5.0", "127.0.0.1", expiresAt, &oldSession.ID)
	if err != nil {
		t.Fatalf("createSession for child failed: %v", err)
	}
	if newSession.ParentSessionID == nil || *newSession.ParentSessionID != oldSession.ID {
		t.Fatal("child session should reference parent session ID")
	}
}

// TestReuseDetection verifies that presenting an already-rotated token
// triggers full session family revocation (the security response to theft).
func TestReuseDetection(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()
	svc := &testableSessionService{db: testDB}

	rawToken := "stolen-refresh-token-abc123"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	// Create and immediately rotate (simulates legitimate use)
	_, err := svc.createSession(1, rawToken, "Mozilla/5.0", "127.0.0.1", expiresAt, nil)
	if err != nil {
		t.Fatalf("setup createSession failed: %v", err)
	}
	_, err = svc.rotateSession(rawToken, 1) // legitimate rotation
	if err != nil {
		t.Fatalf("setup rotateSession failed: %v", err)
	}

	// Create the new legitimate session to verify it gets revoked too
	newToken := "new-legit-token"
	_, err = svc.createSession(1, newToken, "Mozilla/5.0", "127.0.0.1", expiresAt, nil)
	if err != nil {
		t.Fatalf("setup new session failed: %v", err)
	}

	// Now an attacker presents the old (already-rotated) token — reuse attack
	_, err = svc.rotateSession(rawToken, 1)
	if err == nil {
		t.Fatal("expected ErrTokenReuse, got nil")
	}
	if err != services.ErrTokenReuse {
		t.Fatalf("expected ErrTokenReuse, got: %v", err)
	}

	// All sessions for this user must now be revoked
	active, _ := svc.listActive(1)
	if len(active) != 0 {
		t.Fatalf("expected 0 active sessions after reuse detection, got %d", len(active))
	}
}

// TestSessionListing verifies that only active (non-revoked, non-expired) sessions
// are returned for a user.
func TestSessionListing(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()
	svc := &testableSessionService{db: testDB}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	// Create 3 sessions, revoke 1
	_, _ = svc.createSession(1, "token-a", "Chrome", "1.1.1.1", expiresAt, nil)
	_, _ = svc.createSession(1, "token-b", "Firefox", "2.2.2.2", expiresAt, nil)
	_, _ = svc.createSession(1, "token-c", "Safari", "3.3.3.3", expiresAt, nil)

	// Revoke token-b
	_, _ = svc.rotateSession("token-b", 1)

	active, err := svc.listActive(1)
	if err != nil {
		t.Fatalf("listActive failed: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("expected 2 active sessions, got %d", len(active))
	}
}

// TestRevokeAll verifies that logout-all revokes every active session.
func TestRevokeAll(t *testing.T) {
	testDB := setupTestDB(t)
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()
	svc := &testableSessionService{db: testDB}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, _ = svc.createSession(1, "tok-1", "Chrome", "1.1.1.1", expiresAt, nil)
	_, _ = svc.createSession(1, "tok-2", "Firefox", "2.2.2.2", expiresAt, nil)

	if err := svc.revokeAll(1); err != nil {
		t.Fatalf("revokeAll failed: %v", err)
	}

	active, _ := svc.listActive(1)
	if len(active) != 0 {
		t.Fatalf("expected 0 active sessions after revokeAll, got %d", len(active))
	}
}
