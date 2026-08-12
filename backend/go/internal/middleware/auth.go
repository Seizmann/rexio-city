package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

/* ── JWT Auth Middleware ─────────────────────────────────────────── */

// Auth extracts and validates JWT from the Authorization: Bearer header.
// Sets "user_id" (uint) in Fiber locals for downstream handlers.
func Auth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "UNAUTHORIZED", "message": "Missing or invalid authorization header"},
			})
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")
		claims, err := parseJWT(tokenString, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "INVALID_TOKEN", "message": "Invalid or expired token"},
			})
		}

		userID, ok := (*claims)["user_id"].(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "INVALID_TOKEN", "message": "Missing user_id in token"},
			})
		}

		c.Locals("user_id", uint(userID))
		c.Locals("token_expires", (*claims)["exp"])
		return c.Next()
	}
}

// AuthAdmin is the same as Auth but requires an "admin": true claim.
func AuthAdmin(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "UNAUTHORIZED", "message": "Missing or invalid authorization header"},
			})
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")
		claims, err := parseJWT(tokenString, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "INVALID_TOKEN", "message": "Invalid or expired token"},
			})
		}

		isAdmin, _ := (*claims)["admin"].(bool)
		if !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "FORBIDDEN", "message": "Admin access required"},
			})
		}

		userID, _ := (*claims)["user_id"].(float64)
		c.Locals("user_id", uint(userID))
		c.Locals("is_admin", true)
		return c.Next()
	}
}

// ExtractUserIDFromHeader is a helper for optional-auth routes (no middleware).
func ExtractUserIDFromHeader(authHeader, jwtSecret string) uint {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := parseJWT(tokenString, jwtSecret)
	if err != nil {
		return 0
	}
	if userID, ok := (*claims)["user_id"].(float64); ok {
		return uint(userID)
	}
	return 0
}

/* ── CSRF Protection (double-submit cookie pattern) ─────────────── */
//
// How it works:
//   1. On any authenticated response, the backend sets a readable (non-httpOnly)
//      cookie named "rexio_csrf" containing an HMAC-signed token.
//   2. The frontend JS reads this cookie and attaches it as the
//      X-CSRF-Token request header on all state-changing requests.
//   3. This middleware validates the header matches the cookie value.
//
// This stops CSRF because: a cross-origin attacker can trigger the browser
// to send the cookie automatically, but cannot READ the cookie value (same-origin
// policy), so they cannot forge the matching header.

const csrfCookieName = "rexio_csrf"
const csrfHeaderName = "X-CSRF-Token"

// CSRF validates the X-CSRF-Token header against the rexio_csrf cookie.
// Apply this to all state-changing endpoints (POST, PUT, PATCH, DELETE).
func CSRF(csrfSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Safe HTTP methods don't need CSRF protection
		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return c.Next()
		}

		// Auth endpoints (login/signup/refresh) are CSRF-exempt:
		// they don't require a pre-existing session cookie.
		path := c.Path()
		if strings.HasPrefix(path, "/api/auth/login") ||
			strings.HasPrefix(path, "/api/auth/signup") ||
			strings.HasPrefix(path, "/api/auth/refresh") {
			return c.Next()
		}

		// Authorization header (Bearer token) is inherently immune to CSRF because
		// cross-origin browsers cannot attach custom Authorization headers automatically.
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next()
		}

		cookieVal := c.Cookies(csrfCookieName)
		headerVal := c.Get(csrfHeaderName)

		if cookieVal == "" || headerVal == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "CSRF_MISSING", "message": "CSRF token missing"},
			})
		}

		if !validateCSRFToken(cookieVal, headerVal, csrfSecret) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "CSRF_INVALID", "message": "CSRF token invalid"},
			})
		}

		return c.Next()
	}
}

// IssueCSRFCookie generates a new CSRF token and sets the readable cookie.
// Call this on login/signup responses and whenever you want to rotate the CSRF token.
func IssueCSRFCookie(c *fiber.Ctx, csrfSecret string, secure bool, domain string) {
	token := generateCSRFToken(csrfSecret)
	c.Cookie(&fiber.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		Domain:   domain,
		Expires:  time.Now().Add(30 * 24 * time.Hour), // match refresh token expiry
		Secure:   secure,
		HTTPOnly: false, // MUST be readable by JS to implement double-submit
		SameSite: "Strict",
	})
}

// generateCSRFToken creates an HMAC-signed random nonce as the CSRF token.
func generateCSRFToken(secret string) string {
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)
	nonceHex := hex.EncodeToString(nonce)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(nonceHex))
	sig := hex.EncodeToString(mac.Sum(nil))
	return nonceHex + "." + sig
}

// validateCSRFToken checks that the submitted token matches the cookie value.
// Uses constant-time comparison to prevent timing attacks.
func validateCSRFToken(cookieVal, headerVal, secret string) bool {
	// Both must be identical (double-submit pattern: header == cookie)
	return hmac.Equal([]byte(cookieVal), []byte(headerVal))
}

/* ── Token Generation / Parsing ─────────────────────────────────── */

// GenerateJWT creates a signed access token JWT.
func GenerateJWT(userID uint, secret string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a signed refresh token JWT.
// The raw JWT string is returned — the caller hashes it before storing.
func GenerateRefreshToken(userID uint, secret string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     time.Now().Add(expiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseRefreshToken validates a refresh token JWT and returns its claims.
// This function previously had a hardcoded "TODO_CONFIG" secret — now fixed.
func ParseRefreshToken(tokenString, secret string) (*jwt.MapClaims, error) {
	if secret == "" {
		return nil, errors.New("refresh secret not configured")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if claims["type"] != "refresh" {
		return nil, errors.New("not a refresh token")
	}

	return &claims, nil
}

// parseJWT is the internal helper that validates any JWT with the given secret.
func parseJWT(tokenString, secret string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return &claims, nil
}
