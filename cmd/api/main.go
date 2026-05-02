package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gpslakshan/hireflow/internal/config"
	"github.com/gpslakshan/hireflow/internal/database"
	"github.com/gpslakshan/hireflow/internal/handler"
	"github.com/gpslakshan/hireflow/internal/repository"
	"github.com/gpslakshan/hireflow/internal/router"
	"github.com/gpslakshan/hireflow/internal/service"
	"github.com/gpslakshan/hireflow/internal/storage"
	"github.com/rs/zerolog/log"
)

// @title           HireFlow API
// @version         1.0
// @description     A production-grade Job Recruitment REST API built with Go and Gin.

// @contact.name    HireFlow Support
// @contact.email   support@hireflow.com

// @license.name    MIT

// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter your JWT token as: Bearer <token>
func main() {
	// 1. Config
	cfg := config.Load()

	// 2. Logger — must initialise before anything else logs
	config.InitLogger(cfg.AppEnv)

	// 3. Database
	db := database.Connect(cfg)
	database.Migrate(db)
	database.Seed(db, cfg)

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get sql.DB")
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	defer sqlDB.Close()

	// 4. Gin mode
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// ── 5. Storage ───────────────────────────────────────────
	s3Storage, err := storage.NewS3Storage(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialise S3 storage")
	}

	// 6. Repositories
	userRepo := repository.NewUserRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	jobRepo := repository.NewJobRepository(db)
	appRepo := repository.NewApplicationRepository(db)

	// 7. Services
	authService := service.NewAuthService(userRepo, cfg)
	companyService := service.NewCompanyService(companyRepo)
	jobService := service.NewJobService(jobRepo, companyRepo)
	appService := service.NewApplicationService(appRepo, jobRepo, s3Storage)
	uploadService := service.NewUploadService(s3Storage)

	// 8. Handlers
	authHandler := handler.NewAuthHandler(authService)
	companyHandler := handler.NewCompanyHandler(companyService)
	jobHandler := handler.NewJobHandler(jobService)
	appHandler := handler.NewApplicationHandler(appService)
	uploadHandler := handler.NewUploadHandler(uploadService)

	// 9. Router
	r := router.Setup(cfg, authHandler, companyHandler, jobHandler, appHandler, uploadHandler)

	// 10. HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.AppPort),
		Handler:      r,
		ReadTimeout:  10 * time.Second, // max time to read request
		WriteTimeout: 10 * time.Second, // max time to write response
		IdleTimeout:  60 * time.Second, // max time for keep-alive connections
	}

	// ── 11. Start in goroutine ───────────────────────────────
	// Run in background so the main goroutine can listen for
	// shutdown signals without blocking.
	go func() {
		log.Info().Msgf("server starting on port %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed to start")
		}
	}()

	// ── 12. Graceful shutdown ────────────────────────────────
	// Block until SIGINT (Ctrl+C) or SIGTERM (Docker / Kubernetes)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// Give in-flight requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped cleanly")
}
