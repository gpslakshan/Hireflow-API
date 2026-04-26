package database

import (
	"fmt"
	"log"

	"github.com/gpslakshan/hireflow/internal/config"
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
