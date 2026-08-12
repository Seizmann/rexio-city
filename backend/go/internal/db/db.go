package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init initializes the database connection
func Init(cfg *config.Config) {
	dsn := cfg.DatabaseURL
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Disable prepared statement caching for Supabase / PgBouncer pooler compatibility (prevents 42P05 errors)
	}), &gorm.Config{
		PrepareStmt: false,
		Logger:      logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test connection
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Ensure public_id column and index exist on posts table
	DB.Exec("ALTER TABLE posts ADD COLUMN IF NOT EXISTS public_id VARCHAR(32)")
	DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_posts_public_id ON posts (public_id)")
	DB.Exec("UPDATE posts SET public_id = SUBSTRING(MD5(RANDOM()::text || id::text) FROM 1 FOR 16) WHERE public_id IS NULL OR public_id = ''")

	// Ensure sessions table exists (migration 002). We use raw SQL so this is
	// idempotent and safe to run at startup even if already applied.
	DB.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		id                BIGSERIAL PRIMARY KEY,
		user_id           BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash        VARCHAR(64) UNIQUE NOT NULL,
		parent_session_id BIGINT REFERENCES sessions(id) ON DELETE SET NULL,
		device_info       TEXT,
		ip_address        VARCHAR(45),
		created_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		last_used_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		expires_at        TIMESTAMP WITH TIME ZONE NOT NULL,
		revoked_at        TIMESTAMP WITH TIME ZONE
	)`)
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_parent ON sessions(parent_session_id)")
	// Drop old bare refresh_tokens table (superseded by sessions)
	DB.Exec("DROP TABLE IF EXISTS refresh_tokens")

	log.Println("Database connected successfully")
}


// GetDB returns the global database instance
func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("Database not initialized. Call db.Init() first")
	}
	return DB
}

// Migrate runs automatic migrations
func Migrate() {
	fmt.Println("Running database migrations...")
	// Migrations will be run manually via SQL files
	// GORM AutoMigrate is disabled for production safety
	fmt.Println("Note: Apply migrations manually using the SQL files in migrations/")
}
