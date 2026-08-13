package config

import (
	"os"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	JWTExpiry      time.Duration
	RefreshSecret  string
	RefreshExpiry  time.Duration
	FrontendURL    string
	AdminURL       string
	MediaEndpoint  string
	MediaBucket    string
	MediaAccessKey string
	MediaSecretKey string
	MediaURL       string

	// Cookie config for httpOnly refresh token
	CookieDomain string // e.g. "rexio.pro" in prod, "" for localhost
	CookieSecure bool   // true in prod (HTTPS only), false in dev

	// CSRF: secret used to generate/validate the double-submit CSRF token cookie
	CSRFSecret string

	// Brevo transactional email (new-device login alerts, etc.)
	BrevoAPIKey    string
	BrevoFromEmail string
	BrevoFromName  string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "10800"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgresql://rexio:rexio@localhost:5432/rexiocity"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiry:      parseDuration("JWT_EXPIRY", "15m"),
		RefreshSecret:  getEnv("REFRESH_SECRET", "dev-refresh-secret-change-in-production"),
		RefreshExpiry:  parseDuration("REFRESH_EXPIRY", "720h"), // 30 days
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:3000"),
		AdminURL:       getEnv("ADMIN_URL", "http://localhost:5173"),
		MediaEndpoint:  getEnv("MEDIA_ENDPOINT", "http://localhost:9000"),
		MediaBucket:    getEnv("MEDIA_BUCKET", "rexio-city"),
		MediaAccessKey: getEnv("MEDIA_ACCESS_KEY", "minioadmin"),
		MediaSecretKey: getEnv("MEDIA_SECRET_KEY", "minioadmin"),
		MediaURL:       getEnv("MEDIA_URL", "http://localhost:9000"),

		CookieDomain: getEnv("COOKIE_DOMAIN", ""),        // empty = current host (works for localhost)
		CookieSecure: getEnvBool("COOKIE_SECURE", false), // set true in production

		CSRFSecret: getEnv("CSRF_SECRET", "dev-csrf-secret-change-in-production"),

		BrevoAPIKey:    getEnv("BREVO_API_KEY", ""),
		BrevoFromEmail: getEnv("BREVO_FROM_EMAIL", "noreply@rexio.pro"),
		BrevoFromName:  getEnv("BREVO_FROM_NAME", "RexiO City"),
	}
}

func LoadAdmin() *Config {
	return Load()
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return val == "true" || val == "1" || val == "yes"
}

func parseDuration(key, defaultVal string) time.Duration {
	val := getEnv(key, defaultVal)
	d, err := time.ParseDuration(val)
	if err != nil {
		d, _ = time.ParseDuration(defaultVal)
	}
	return d
}
