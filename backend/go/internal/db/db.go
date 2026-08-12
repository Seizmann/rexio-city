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
