package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gpslakshan/hireflow/internal/config"
	"github.com/gpslakshan/hireflow/internal/database"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect to database
	db := database.Connect(cfg)

	// Retrieve underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	defer sqlDB.Close()

	// 3. Set Gin mode based on environment
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 4. Start server
	log.Printf("server starting on port %s", cfg.AppPort)
	if err := gin.Default().Run(fmt.Sprintf(":%s", cfg.AppPort)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
