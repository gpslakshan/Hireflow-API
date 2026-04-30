package database

import (
	"fmt"
	"log"

	"github.com/gpslakshan/hireflow/internal/config"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to PostgreSQL using GORM.
// Returns a *gorm.DB instance to be shared across the application.
func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	// GORM logger level — verbose in dev, silent in prod
	logLevel := logger.Info
	if cfg.AppEnv == "production" {
		logLevel = logger.Silent
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		// Fatal — app cannot run without a database
		log.Fatalf("failed to connect to database: %v", err)
	}

	log.Println("database connection established")
	return db
}

// Migrate runs GORM AutoMigrate for all entities.
// Creates tables if they don't exist, adds missing columns.
// Does NOT delete columns or change types — safe to run on startup.
func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&entity.Company{}, // Company first — User has FK to Company
		&entity.User{},
		&entity.Job{},
		&entity.Application{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("database migration completed")
}
