package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	JWTExpiry       string
	RefreshSecret   string
	RefreshExpiry   string
	FrontendURL     string
	AdminURL        string
	BrevoAPIKey     string
	BrevoFromEmail  string
	R2AccountID     string
	R2AccessKeyID   string
	R2SecretKey     string
	R2Bucket        string
	GoogleClientID  string
	GoogleClientSecret string
	GitHubClientID  string
	GitHubClientSecret string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "10800"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		RedisURL:           getEnv("REDIS_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpiry:          getEnv("JWT_EXPIRY", "15m"),
		RefreshSecret:      getEnv("REFRESH_TOKEN_SECRET", ""),
		RefreshExpiry:      getEnv("REFRESH_TOKEN_EXPIRY", "720h"),
		FrontendURL:        getEnv("FRONTEND_URL", "https://city.rexio.pro"),
		AdminURL:           getEnv("ADMIN_URL", "https://admin.rexio.pro"),
		BrevoAPIKey:        getEnv("BREVO_API_KEY", ""),
		BrevoFromEmail:     getEnv("BREVO_FROM_EMAIL", "noreply@rexio.pro.bd"),
		R2AccountID:        getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:      getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretKey:        getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:           getEnv("R2_BUCKET", "rexio-city-media"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
	}
}

func LoadAdmin() *Config {
	return &Config{
		Port:            getEnv("PORT", "10900"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		RedisURL:        getEnv("REDIS_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		JWTExpiry:       getEnv("JWT_EXPIRY", "15m"),
		RefreshSecret:   getEnv("REFRESH_TOKEN_SECRET", ""),
		RefreshExpiry:   getEnv("REFRESH_TOKEN_EXPIRY", "720h"),
		FrontendURL:     getEnv("FRONTEND_URL", "https://city.rexio.pro"),
		AdminURL:        getEnv("ADMIN_URL", "https://admin.rexio.pro"),
		BrevoAPIKey:     getEnv("BREVO_API_KEY", ""),
		BrevoFromEmail:  getEnv("BREVO_FROM_EMAIL", "noreply@rexio.pro.bd"),
		R2AccountID:     getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:   getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretKey:     getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:        getEnv("R2_BUCKET", "rexio-city-media"),
		GoogleClientID:  getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:  getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
