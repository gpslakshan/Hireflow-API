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
	"github.com/rs/zerolog/log"
)

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

	// 5. Wire dependencies
	userRepo := repository.NewUserRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	jobRepo := repository.NewJobRepository(db)
	appRepo := repository.NewApplicationRepository(db)

	authService := service.NewAuthService(userRepo, cfg)
	companyService := service.NewCompanyService(companyRepo)
	jobService := service.NewJobService(jobRepo, companyRepo)
	appService := service.NewApplicationService(appRepo, jobRepo)

	authHandler := handler.NewAuthHandler(authService)
	companyHandler := handler.NewCompanyHandler(companyService)
	jobHandler := handler.NewJobHandler(jobService)
	appHandler := handler.NewApplicationHandler(appService)

	// 6. Router
	r := router.Setup(cfg, authHandler, companyHandler, jobHandler, appHandler)

	// 7. Build http.Server manually so we can shut it down gracefully
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.AppPort),
		Handler:      r,
		ReadTimeout:  10 * time.Second, // max time to read request
		WriteTimeout: 10 * time.Second, // max time to write response
		IdleTimeout:  60 * time.Second, // max time for keep-alive connections
	}

	// 8. Start server in a goroutine so it doesn't block the shutdown logic
	go func() {
		log.Info().Msgf("server starting on port %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed to start")
		}
	}()

	// 9. Block until we receive a termination signal (Ctrl+C or kill)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// 10. Give in-flight requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped cleanly")
}
