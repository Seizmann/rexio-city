package handlers

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/seizmann/rexio-city/backend/go/internal/config"
	"github.com/seizmann/rexio-city/backend/go/internal/middleware"
	"github.com/seizmann/rexio-city/backend/go/internal/services"
)

const refreshCookieName = "rexio_refresh"

// AuthHandler handles authentication requests.
type AuthHandler struct {
	authService    *services.AuthService
	sessionService *services.SessionService
	emailService   *services.EmailService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService:    services.NewAuthService(),
		sessionService: services.NewSessionService(),
		emailService:   services.NewEmailService(),
	}
}

/* ── Request Bodies ─────────────────────────────────────────────── */

type SignupRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

/* ── Signup ─────────────────────────────────────────────────────── */

// Signup handles user registration.
// Returns access_token in JSON body; sets refresh token as httpOnly cookie.
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	cfg := config.Load()

	var input SignupRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	result, rawRefreshToken, err := h.authService.Signup(services.SignupInput{
		Username:    input.Username,
		Email:       input.Email,
		Password:    input.Password,
		DisplayName: input.DisplayName,
		DeviceInfo:  c.Get("User-Agent"),
		IPAddress:   c.IP(),
	})
	if err != nil {
		statusCode := fiber.StatusBadRequest
		if err.Error() == "username already taken" || err.Error() == "email already registered" {
			statusCode = fiber.StatusConflict
		}
		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
	}

	// Set refresh token as httpOnly cookie — JS cannot read this
	setRefreshCookie(c, rawRefreshToken, cfg)

	// Issue the CSRF token cookie (readable by JS for double-submit)
	middleware.IssueCSRFCookie(c, cfg.CSRFSecret, cfg.CookieSecure, cfg.CookieDomain)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user":         result.User,
			"access_token": result.AccessToken,
			"expires_in":   result.ExpiresIn,
		},
	})
}

/* ── Login ──────────────────────────────────────────────────────── */

// Login handles user authentication.
// Returns access_token in JSON body; sets refresh token as httpOnly cookie.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	cfg := config.Load()

	var input LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid request body"},
		})
	}

	result, rawRefreshToken, err := h.authService.Login(services.LoginInput{
		Email:      input.Email,
		Password:   input.Password,
		DeviceInfo: c.Get("User-Agent"),
		IPAddress:  c.IP(),
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "AUTH_ERROR", "message": err.Error()},
		})
	}

	// Set refresh token as httpOnly cookie
	setRefreshCookie(c, rawRefreshToken, cfg)

	// Issue CSRF cookie
	middleware.IssueCSRFCookie(c, cfg.CSRFSecret, cfg.CookieSecure, cfg.CookieDomain)

	// Send new-device email alert asynchronously — don't block the response
	if result.IsNewDevice && result.User.Email != nil {
		// Capture values before goroutine — Fiber reuses context across connections
		userAgent := c.Get("User-Agent")
		ipAddr := c.IP()
		go func() {
			if err := h.emailService.SendNewDeviceAlert(
				*result.User.Email,
				result.User.Username,
				userAgent,
				ipAddr,
				time.Now(),
			); err != nil {
				log.Printf("[auth] new-device email failed for user %d: %v", result.User.ID, err)
			}
		}()
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user":         result.User,
			"access_token": result.AccessToken,
			"expires_in":   result.ExpiresIn,
		},
	})
}

/* ── Token Refresh (cookie-based) ───────────────────────────────── */

// Refresh reads the httpOnly cookie, rotates the session, and returns a new access token.
// The new refresh token is also set as a cookie automatically.
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	cfg := config.Load()

	rawRefreshToken := c.Cookies(refreshCookieName)

	result, newRawToken, err := h.authService.RefreshToken(services.RefreshInput{
		RefreshToken: rawRefreshToken,
		DeviceInfo:   c.Get("User-Agent"),
		IPAddress:    c.IP(),
	})
	if err != nil {
		// Whether it's reuse detection or an expired token, clear the cookie
		clearRefreshCookie(c, cfg)

		code := "AUTH_ERROR"
		if errors.Is(err, services.ErrTokenReuse) {
			code = "SESSION_COMPROMISED"
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": code, "message": err.Error()},
		})
	}

	// Set the new (rotated) refresh token cookie
	setRefreshCookie(c, newRawToken, cfg)

	// Issue new CSRF cookie on session rotation
	middleware.IssueCSRFCookie(c, cfg.CSRFSecret, cfg.CookieSecure, cfg.CookieDomain)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"access_token": result.AccessToken,
			"expires_in":   result.ExpiresIn,
		},
	})
}

/* ── Logout ─────────────────────────────────────────────────────── */

// Logout revokes the current session and clears the cookie.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	cfg := config.Load()
	rawRefreshToken := c.Cookies(refreshCookieName)

	if rawRefreshToken != "" {
		// Best-effort revoke — don't fail the logout if the session is already gone
		_ = h.sessionService.RevokeAllSessions(c.Locals("user_id").(uint))
	}

	clearRefreshCookie(c, cfg)
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "logged out"}})
}

/* ── Session Management Endpoints ───────────────────────────────── */

// ListSessions returns all active sessions (devices) for the logged-in user.
func (h *AuthHandler) ListSessions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	sessions, err := h.sessionService.ListActiveSessions(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": "Failed to list sessions"},
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": sessions})
}

// RevokeSession revokes a specific session by ID (must belong to the calling user).
func (h *AuthHandler) RevokeSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	sessionID, err := c.ParamsInt("id")
	if err != nil || sessionID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INVALID_INPUT", "message": "Invalid session ID"},
		})
	}

	if err := h.sessionService.RevokeSessionByID(uint(sessionID), userID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "NOT_FOUND", "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "session revoked"}})
}

// LogoutAll revokes every session for the user and clears the cookie.
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	cfg := config.Load()
	userID := c.Locals("user_id").(uint)

	if err := h.sessionService.RevokeAllSessions(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "INTERNAL_ERROR", "message": "Failed to revoke sessions"},
		})
	}

	clearRefreshCookie(c, cfg)
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"message": "all sessions revoked"}})
}

/* ── Health ─────────────────────────────────────────────────────── */

func (h *AuthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"status": "healthy"}})
}

/* ── Cookie Helpers ─────────────────────────────────────────────── */

// setRefreshCookie sets the httpOnly refresh token cookie.
func setRefreshCookie(c *fiber.Ctx, rawToken string, cfg *config.Config) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    rawToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		Expires:  time.Now().Add(cfg.RefreshExpiry),
		Secure:   cfg.CookieSecure,
		HTTPOnly: true,
		SameSite: "Lax", // Changed from Strict — Cloudflare tunnel breaks SameSite=Strict+Secure cookies
	})
	clearLegacyRefreshCookie(c, cfg)
}

// clearRefreshCookie deletes the refresh token cookie by expiring it.
func clearRefreshCookie(c *fiber.Ctx, cfg *config.Config) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		Domain:   cfg.CookieDomain,
		Expires:  time.Unix(0, 0),
		Secure:   cfg.CookieSecure,
		HTTPOnly: true,
		SameSite: "Lax", // Changed from Strict — Cloudflare tunnel breaks SameSite=Strict+Secure cookies
	})
	clearLegacyRefreshCookie(c, cfg)
}

// clearLegacyRefreshCookie expires the old rexio_refresh cookie set at path=/api/auth.
func clearLegacyRefreshCookie(c *fiber.Ctx, cfg *config.Config) {
	domainAttr := ""
	if cfg.CookieDomain != "" {
		domainAttr = fmt.Sprintf("; Domain=%s", cfg.CookieDomain)
	}
	secureAttr := ""
	if cfg.CookieSecure {
		secureAttr = "; Secure"
	}
	// Manually append the Set-Cookie header so Fiber doesn't overwrite the primary cookie
	cookieHeader := fmt.Sprintf("%s=; Path=/api/auth%s; Expires=Thu, 01 Jan 1970 00:00:00 GMT; HttpOnly%s; SameSite=Lax",
		refreshCookieName, domainAttr, secureAttr)
	c.Append("Set-Cookie", cookieHeader)
}
