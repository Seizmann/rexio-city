package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/text/unicode/norm"

	"github.com/seizmann/rexio-city/backend/go/internal/config"
	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/middleware"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

const (
	// Argon2 parameters (OWASP recommended minimums)
	saltLength  = 16
	keyLength   = 32
	timeCost    = 3
	memoryCost  = 64 * 1024 // 64 MB
	parallelism = 1
)

// AuthService handles authentication logic.
type AuthService struct {
	sessions *SessionService
	email    *EmailService
}

// NewAuthService creates a new auth service.
func NewAuthService() *AuthService {
	return &AuthService{
		sessions: NewSessionService(),
		email:    NewEmailService(),
	}
}

/* ── Input / Output DTOs ────────────────────────────────────────── */

type SignupInput struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
	DeviceInfo  string // User-Agent, injected by handler
	IPAddress   string // client IP, injected by handler
}

// SignupOutput omits RefreshToken — it is set as an httpOnly cookie by the handler.
type SignupOutput struct {
	User        models.User `json:"user"`
	AccessToken string      `json:"access_token"`
	ExpiresIn   int         `json:"expires_in"`
	// SessionID is passed back so the handler can tell if this is "new device"
	SessionID uint `json:"-"`
}

type LoginInput struct {
	Email      string
	Password   string
	DeviceInfo string
	IPAddress  string
}

// LoginOutput omits RefreshToken — it is set as an httpOnly cookie by the handler.
type LoginOutput struct {
	User         models.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	ExpiresIn    int         `json:"expires_in"`
	RefreshToken string      `json:"-"` // populated internally, handler sets cookie
	IsNewDevice  bool        `json:"-"` // handler uses this to send email alert
}

type RefreshInput struct {
	RefreshToken string // read from httpOnly cookie by handler, passed here
	DeviceInfo   string
	IPAddress    string
}

// RefreshOutput omits RefreshToken — handler sets the new cookie.
type RefreshOutput struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"-"` // new rotated token, handler sets cookie
}

/* ── Password Helpers ───────────────────────────────────────────── */

// HashPassword hashes a password using argon2id. Stores as "saltHex:hashHex".
func HashPassword(password string) (string, error) {
	salt := generateRandomBytes(saltLength)
	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLength)
	return fmt.Sprintf("%x:%x", salt, hash), nil
}

// VerifyPassword verifies a password against its stored argon2id hash.
func VerifyPassword(password, stored string) bool {
	parts := strings.SplitN(stored, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := decodeHex(parts[0])
	if err != nil {
		return false
	}
	expected, err := decodeHex(parts[1])
	if err != nil {
		return false
	}
	computed := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLength)
	return constantTimeCompare(computed, expected)
}

/* ── Signup ─────────────────────────────────────────────────────── */

// Signup creates a new user, persists a DB session, and returns tokens.
// The caller (handler) is responsible for setting the refresh token as an httpOnly cookie.
func (s *AuthService) Signup(input SignupInput) (*SignupOutput, string, error) {
	cfg := config.Load()

	cleanUsername := strings.TrimSpace(strings.ToLower(input.Username))
	cleanEmail := strings.TrimSpace(strings.ToLower(input.Email))

	if !isValidUsername(cleanUsername) {
		return nil, "", errors.New("username must be 3-15 characters, lowercase letters, numbers, and underscores only")
	}
	if len(input.Password) < 8 {
		return nil, "", errors.New("password must be at least 8 characters")
	}

	// Uniqueness checks
	var existingUser models.User
	if db.GetDB().Where("LOWER(username) = ?", cleanUsername).First(&existingUser).RowsAffected > 0 {
		return nil, "", errors.New("username already taken")
	}
	if cleanEmail != "" {
		if db.GetDB().Where("LOWER(email) = ?", cleanEmail).First(&existingUser).RowsAffected > 0 {
			return nil, "", errors.New("email already registered")
		}
	}

	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = cleanUsername
	}

	user := models.User{
		Username:     norm.NFC.String(cleanUsername),
		DisplayName:  &displayName,
		Email:        &cleanEmail,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if result := db.GetDB().Create(&user); result.Error != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", result.Error)
	}

	accessToken, err := middleware.GenerateJWT(uint(user.ID), cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(uint(user.ID), cfg.RefreshSecret, cfg.RefreshExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(cfg.RefreshExpiry)
	session, err := s.sessions.CreateSession(uint(user.ID), refreshToken, input.DeviceInfo, input.IPAddress, expiresAt, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return &SignupOutput{
		User:        user,
		AccessToken: accessToken,
		ExpiresIn:   int(cfg.JWTExpiry.Seconds()),
		SessionID:   session.ID,
	}, refreshToken, nil
}

/* ── Login ──────────────────────────────────────────────────────── */

// Login authenticates a user, persists a DB session, and returns tokens.
// The caller (handler) sets the refresh token as an httpOnly cookie.
func (s *AuthService) Login(input LoginInput) (*LoginOutput, string, error) {
	cfg := config.Load()

	cleanIdentifier := strings.TrimSpace(strings.ToLower(input.Email))
	if cleanIdentifier == "" {
		return nil, "", errors.New("email or username is required")
	}

	var user models.User
	result := db.GetDB().Where("LOWER(email) = ? OR LOWER(username) = ?", cleanIdentifier, cleanIdentifier).First(&user)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, "", errors.New("invalid email or password")
	}

	if !VerifyPassword(input.Password, user.PasswordHash) {
		return nil, "", errors.New("invalid email or password")
	}

	// Is this a new device? Compare device_info against existing active sessions for this user.
	activeSessions, _ := s.sessions.ListActiveSessions(uint(user.ID))
	isNewDevice := true
	for _, sess := range activeSessions {
		if sess.DeviceInfo == input.DeviceInfo {
			isNewDevice = false
			break
		}
	}

	accessToken, err := middleware.GenerateJWT(uint(user.ID), cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(uint(user.ID), cfg.RefreshSecret, cfg.RefreshExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(cfg.RefreshExpiry)
	if _, err = s.sessions.CreateSession(uint(user.ID), refreshToken, input.DeviceInfo, input.IPAddress, expiresAt, nil); err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginOutput{
		User:        user,
		AccessToken: accessToken,
		ExpiresIn:   int(cfg.JWTExpiry.Seconds()),
		IsNewDevice: isNewDevice,
	}, refreshToken, nil
}

/* ── Refresh (with rotation) ────────────────────────────────────── */

// RefreshToken validates the incoming refresh token from the cookie,
// rotates it (old invalidated, new issued), persists the new session,
// and returns the new access + refresh tokens.
// The caller (handler) reads the cookie and sets the new cookie.
func (s *AuthService) RefreshToken(input RefreshInput) (*RefreshOutput, string, error) {
	cfg := config.Load()

	if input.RefreshToken == "" {
		return nil, "", errors.New("refresh token cookie missing")
	}

	// Validate the JWT signature and extract user_id before touching the DB
	claims, err := middleware.ParseRefreshToken(input.RefreshToken, cfg.RefreshSecret)
	if err != nil {
		return nil, "", errors.New("invalid refresh token")
	}

	userID, ok := (*claims)["user_id"].(float64)
	if !ok {
		return nil, "", errors.New("invalid token claims")
	}

	// Rotate: validates + revokes old session, returns old session for lineage
	oldSession, err := s.sessions.RotateSession(input.RefreshToken, uint(userID))
	if err != nil {
		if errors.Is(err, ErrTokenReuse) {
			// Reuse detected — all sessions already revoked inside RotateSession.
			// Return a clear error; handler should clear the cookie.
			return nil, "", ErrTokenReuse
		}
		return nil, "", errors.New("invalid or expired session")
	}

	// Issue new tokens
	accessToken, err := middleware.GenerateJWT(uint(userID), cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := middleware.GenerateRefreshToken(uint(userID), cfg.RefreshSecret, cfg.RefreshExpiry)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Persist new session with rotation lineage
	expiresAt := time.Now().Add(cfg.RefreshExpiry)
	if _, err = s.sessions.CreateSession(uint(userID), newRefreshToken, input.DeviceInfo, input.IPAddress, expiresAt, &oldSession.ID); err != nil {
		return nil, "", fmt.Errorf("failed to persist new session: %w", err)
	}

	return &RefreshOutput{
		AccessToken: accessToken,
		ExpiresIn:   int(cfg.JWTExpiry.Seconds()),
	}, newRefreshToken, nil
}

/* ── Helpers ────────────────────────────────────────────────────── */

func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 15 {
		return false
	}
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func decodeHex(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, errors.New("invalid hex string")
	}
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		var value byte
		for j := 0; j < 2; j++ {
			var d byte
			c := s[i+j]
			switch {
			case c >= '0' && c <= '9':
				d = c - '0'
			case c >= 'a' && c <= 'f':
				d = c - 'a' + 10
			case c >= 'A' && c <= 'F':
				d = c - 'A' + 10
			default:
				return nil, errors.New("invalid hex character")
			}
			value = (value << 4) | d
		}
		b[i/2] = value
	}
	return b, nil
}

func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
