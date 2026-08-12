package services

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

// SessionService manages DB-persisted refresh token sessions.
// All token rotation and reuse detection logic lives here.
type SessionService struct{}

func NewSessionService() *SessionService { return &SessionService{} }

// CreateSession persists a new session record for a freshly issued refresh token.
// rawToken is the JWT string; we hash it before storage.
func (s *SessionService) CreateSession(userID uint, rawToken, deviceInfo, ipAddress string, expiresAt time.Time, parentID *uint) (*models.Session, error) {
	hash := hashToken(rawToken)

	session := models.Session{
		UserID:          userID,
		TokenHash:       hash,
		ParentSessionID: parentID,
		DeviceInfo:      deviceInfo,
		IPAddress:       ipAddress,
		CreatedAt:       time.Now(),
		LastUsedAt:      time.Now(),
		ExpiresAt:       expiresAt,
	}

	if result := db.GetDB().Create(&session); result.Error != nil {
		return nil, fmt.Errorf("failed to persist session: %w", result.Error)
	}
	return &session, nil
}

// RotateSession looks up the session by rawToken hash, validates it, marks it revoked,
// and returns the session so the caller can issue a new child session.
//
// If the token hash exists but is already revoked, this is a REUSE ATTACK —
// we revoke all sessions for this user and return ErrTokenReuse.
func (s *SessionService) RotateSession(rawToken string, userID uint) (*models.Session, error) {
	hash := hashToken(rawToken)

	var session models.Session
	result := db.GetDB().Where("token_hash = ?", hash).First(&session)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, errors.New("session not found")
	}

	// Reuse detection: already revoked token presented again
	if session.RevokedAt != nil {
		// Revoke ALL active sessions for this user — potential account compromise
		_ = s.RevokeAllSessions(session.UserID)
		return nil, ErrTokenReuse
	}

	// Expired session
	if time.Now().After(session.ExpiresAt) {
		_ = s.revokeSession(&session)
		return nil, errors.New("session expired")
	}

	// Wrong user — shouldn't happen but be explicit
	if session.UserID != userID {
		return nil, errors.New("session user mismatch")
	}

	// Mark old session as rotated-out (revoked)
	if err := s.revokeSession(&session); err != nil {
		return nil, fmt.Errorf("failed to rotate session: %w", err)
	}

	return &session, nil
}

// FindActiveSession retrieves a non-revoked, non-expired session by raw token.
// Used on the refresh endpoint before rotation.
func (s *SessionService) FindActiveSession(rawToken string) (*models.Session, error) {
	hash := hashToken(rawToken)

	var session models.Session
	result := db.GetDB().Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now()).First(&session)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, errors.New("session not found or expired")
	}

	// Touch last_used_at
	now := time.Now()
	db.GetDB().Model(&session).Update("last_used_at", now)
	session.LastUsedAt = now
	return &session, nil
}

// ListActiveSessions returns all non-revoked, non-expired sessions for a user.
func (s *SessionService) ListActiveSessions(userID uint) ([]models.Session, error) {
	var sessions []models.Session
	result := db.GetDB().
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("last_used_at DESC").
		Find(&sessions)
	return sessions, result.Error
}

// RevokeSession revokes a single session by ID, scoped to a specific user (ownership check).
func (s *SessionService) RevokeSessionByID(sessionID, userID uint) error {
	var session models.Session
	result := db.GetDB().Where("id = ? AND user_id = ?", sessionID, userID).First(&session)
	if result.Error != nil || result.RowsAffected == 0 {
		return errors.New("session not found")
	}
	return s.revokeSession(&session)
}

// RevokeAllSessions revokes every session belonging to a user.
// Called on logout-all and on reuse detection.
func (s *SessionService) RevokeAllSessions(userID uint) error {
	now := time.Now()
	return db.GetDB().
		Model(&models.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

// revokeSession sets revoked_at on a session record.
func (s *SessionService) revokeSession(session *models.Session) error {
	now := time.Now()
	session.RevokedAt = &now
	return db.GetDB().Model(session).Update("revoked_at", now).Error
}

// hashToken returns the SHA-256 hex string of a raw token.
// We only store the hash, never the raw token value.
func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return fmt.Sprintf("%x", sum)
}

// ErrTokenReuse signals that a revoked refresh token was presented again,
// indicating a potential session hijack. The entire session family is revoked.
var ErrTokenReuse = errors.New("token reuse detected: all sessions revoked")
