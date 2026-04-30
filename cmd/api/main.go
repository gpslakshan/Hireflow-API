package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gpslakshan/hireflow/internal/config"
	"github.com/gpslakshan/hireflow/internal/database"
	"github.com/gpslakshan/hireflow/internal/handler"
	"github.com/gpslakshan/hireflow/internal/repository"
	"github.com/gpslakshan/hireflow/internal/router"
	"github.com/gpslakshan/hireflow/internal/service"
)

func main() {
	// 1. Config
	cfg := config.Load()

	// 2. Database
	db := database.Connect(cfg)
	database.Migrate(db)

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	defer sqlDB.Close()

	// 3. Gin mode
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 4. Wire dependencies — bottom up
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	// 5. Setup router and start server
	r := router.Setup(cfg, authHandler)

	log.Printf("server starting on port %s", cfg.AppPort)
	if err := r.Run(fmt.Sprintf(":%s", cfg.AppPort)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
