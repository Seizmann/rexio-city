package services

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/text/unicode/norm"

	"github.com/seizmann/rexio-city/backend/go/internal/config"
	"github.com/seizmann/rexio-city/backend/go/internal/db"
	"github.com/seizmann/rexio-city/backend/go/internal/middleware"
	"github.com/seizmann/rexio-city/backend/go/internal/models"
)

const (
	// Argon2 parameters
	saltLength = 16
	keyLength  = 32
	timeCost   = 3
	memoryCost = 64 * 1024 // 64 MB
	parallelism = 1
)

// AuthService handles authentication logic
type AuthService struct{}

// NewAuthService creates a new auth service
func NewAuthService() *AuthService {
	return &AuthService{}
}

// SignupInput contains signup request data
type SignupInput struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// SignupOutput contains signup response data
type SignupOutput struct {
	User         models.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
}

// LoginInput contains login request data
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginOutput contains login response data
type LoginOutput struct {
	User         models.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
}

// RefreshInput contains refresh request data
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshOutput contains refresh response data
type RefreshOutput struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// HashPassword hashes a password using argon2id
func HashPassword(password string) (string, error) {
	salt := generateRandomBytes(saltLength)
	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLength)

	// Encode as hex
	saltHex := fmt.Sprintf("%x", salt)
	hashHex := fmt.Sprintf("%x", hash)

	// Return in format: hex(salt):hex(hash)
	encodedPassword := saltHex + ":" + hashHex
	return encodedPassword, nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(password, hash string) bool {
	parts := []byte(hash)
	colonIndex := -1
	for i, b := range parts {
		if b == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return false
	}

	saltHex := string(parts[:colonIndex])
	hashHex := string(parts[colonIndex+1:])

	// Decode hex strings
	salt, err := decodeHex(saltHex)
	if err != nil {
		return false
	}

	expectedHash, err := decodeHex(hashHex)
	if err != nil {
		return false
	}

	// Verify password
	computedHash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLength)
	return constantTimeCompare(computedHash, expectedHash)
}

// Signup creates a new user and returns JWT tokens
func (s *AuthService) Signup(input SignupInput) (*SignupOutput, error) {
	cfg := config.Load()

	// Validate username
	if !isValidUsername(input.Username) {
		return nil, errors.New("username must be 3-15 characters, lowercase letters, numbers, and underscores only")
	}

	// Validate password length
	if len(input.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Check if username exists
	var existingUser models.User
	result := db.GetDB().Where("username = ?", input.Username).First(&existingUser)
	if result.RowsAffected > 0 {
		return nil, errors.New("username already taken")
	}

	// Check if email exists (if provided)
	if input.Email != "" {
		result = db.GetDB().Where("email = ?", input.Email).First(&existingUser)
		if result.RowsAffected > 0 {
			return nil, errors.New("email already registered")
		}
	}

	// Hash password
	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.User{
		Username:    norm.NFC.String(input.Username),
		DisplayName: &input.DisplayName,
		Email:       &input.Email,
		PasswordHash: hashedPassword,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result = db.GetDB().Create(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	// Generate tokens
	accessToken, err := middleware.GenerateJWT(uint(user.ID), cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(uint(user.ID), cfg.RefreshSecret, cfg.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &SignupOutput{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(cfg.JWTExpiry.Seconds()),
	}, nil
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(input LoginInput) (*LoginOutput, error) {
	cfg := config.Load()

	// Find user by email
	var user models.User
	result := db.GetDB().Where("email = ?", input.Email).First(&user)
	if result.RowsAffected == 0 {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if !VerifyPassword(input.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	// Generate tokens
	accessToken, err := middleware.GenerateJWT(uint(user.ID), cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(uint(user.ID), cfg.RefreshSecret, cfg.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginOutput{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(cfg.JWTExpiry.Seconds()),
	}, nil
}

// RefreshToken validates and refreshes a refresh token
func (s *AuthService) RefreshToken(input RefreshInput) (*RefreshOutput, error) {
	cfg := config.Load()

	if input.RefreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	// Parse and validate refresh token
	token, err := middleware.ParseRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Extract user ID from claims
	userID, ok := (*token)["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Generate new access token
	accessToken, err := middleware.GenerateJWT(uint(userID), cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &RefreshOutput{
		AccessToken: accessToken,
		ExpiresIn:   int(cfg.JWTExpiry.Seconds()),
	}, nil
}

// Helper functions

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
	// In production, use crypto/rand
	for i := range b {
		b[i] = byte(i % 256)
	}
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
			if c >= '0' && c <= '9' {
				d = c - '0'
			} else if c >= 'a' && c <= 'f' {
				d = c - 'a' + 10
			} else if c >= 'A' && c <= 'F' {
				d = c - 'A' + 10
			} else {
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
