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

	// 4. Repositories
	userRepo := repository.NewUserRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	jobRepo := repository.NewJobRepository(db)
	appRepo := repository.NewApplicationRepository(db)

	// 5. Services
	authService := service.NewAuthService(userRepo, cfg)
	companyService := service.NewCompanyService(companyRepo)
	jobService := service.NewJobService(jobRepo, companyRepo)
	appService := service.NewApplicationService(appRepo, jobRepo)

	// 6. Handlers
	authHandler := handler.NewAuthHandler(authService)
	companyHandler := handler.NewCompanyHandler(companyService)
	jobHandler := handler.NewJobHandler(jobService)
	appHandler := handler.NewApplicationHandler(appService)

	// 7. Router
	r := router.Setup(cfg, authHandler, companyHandler, jobHandler, appHandler)

	log.Printf("server starting on port %s", cfg.AppPort)
	if err := r.Run(fmt.Sprintf(":%s", cfg.AppPort)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
