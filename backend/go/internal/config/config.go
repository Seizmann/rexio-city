package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	JWTExpiry       time.Duration
	RefreshSecret   string
	RefreshExpiry   time.Duration
	FrontendURL     string
	MediaEndpoint   string
	MediaBucket     string
	MediaAccessKey  string
	MediaSecretKey  string
	MediaURL        string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "10800"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgresql://rexio:rexio@localhost:5432/rexiocity"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiry:     parseDuration("JWT_EXPIRY", "15m"),
		RefreshSecret: getEnv("REFRESH_SECRET", "dev-refresh-secret-change-in-production"),
		RefreshExpiry: parseDuration("REFRESH_EXPIRY", "720h"), // 30 days
		FrontendURL:   getEnv("FRONTEND_URL", "http://localhost:3000"),
		MediaEndpoint: getEnv("MEDIA_ENDPOINT", "http://localhost:9000"),
		MediaBucket:   getEnv("MEDIA_BUCKET", "rexio-city"),
		MediaAccessKey: getEnv("MEDIA_ACCESS_KEY", "minioadmin"),
		MediaSecretKey: getEnv("MEDIA_SECRET_KEY", "minioadmin"),
		MediaURL:      getEnv("MEDIA_URL", "http://localhost:9000"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func parseDuration(key, defaultVal string) time.Duration {
	val := getEnv(key, defaultVal)
	d, err := time.ParseDuration(val)
	if err != nil {
		d, _ = time.ParseDuration(defaultVal)
	}
	return d
}
