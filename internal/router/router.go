package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gpslakshan/hireflow/internal/config"
	"github.com/gpslakshan/hireflow/internal/handler"
	"github.com/gpslakshan/hireflow/internal/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/gpslakshan/hireflow/docs" // ← triggers docs.go init()
)

func Setup(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	companyHandler *handler.CompanyHandler,
	jobHandler *handler.JobHandler,
	appHandler *handler.ApplicationHandler,
	uploadHandler *handler.UploadHandler,
) *gin.Engine {
	r := gin.New()

	// ── Global middleware ──────────────────────────────────
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(cfg),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // how long browsers cache preflight responses
	}))

	r.SetTrustedProxies([]string{"127.0.0.1"})

	// ── Routes ────────────────────────────────────────────
	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.Me)
	}

	companies := v1.Group("/companies")
	{
		companies.GET("", companyHandler.GetAll)
		companies.GET("/:id", companyHandler.GetByID)
		companies.POST("",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("admin"),
			companyHandler.Create,
		)
		companies.PUT("/:id",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("recruiter"),
			companyHandler.Update,
		)
		companies.DELETE("/:id",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("admin"),
			companyHandler.Delete,
		)
		companies.POST("/:id/jobs",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("recruiter"),
			jobHandler.Create,
		)
	}

	jobs := v1.Group("/jobs")
	{
		jobs.GET("", jobHandler.GetAll)
		jobs.GET("/:id", jobHandler.GetByID)
		jobs.PUT("/:id",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("recruiter"),
			jobHandler.Update,
		)
		jobs.PATCH("/:id/close",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("recruiter"),
			jobHandler.Close,
		)
		jobs.DELETE("/:id",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("recruiter"),
			jobHandler.Delete,
		)
		jobs.POST("/:id/apply",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("candidate"),
			appHandler.Apply,
		)
		jobs.GET("/:id/applications",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("recruiter"),
			appHandler.GetByJob,
		)
	}

	applications := v1.Group("/applications")
	applications.Use(middleware.AuthMiddleware(cfg))
	{
		applications.GET("/my",
			middleware.RequireRole("candidate"),
			appHandler.GetMyApplications,
		)
		applications.GET("/:id", appHandler.GetByID)
		applications.PATCH("/:id/status",
			middleware.RequireRole("recruiter"),
			appHandler.UpdateStatus,
		)
		applications.DELETE("/:id",
			middleware.RequireRole("candidate"),
			appHandler.Withdraw,
		)
	}

	uploads := v1.Group("/uploads")
	uploads.Use(middleware.AuthMiddleware(cfg))
	{
		uploads.POST("/cv-upload-url",
			middleware.RequireRole("candidate"),
			uploadHandler.GenerateCVUploadURL,
		)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

// getAllowedOrigins returns CORS origins based on environment.
// In development we allow localhost. In production only real domains.
func getAllowedOrigins(cfg *config.Config) []string {
	if cfg.AppEnv == "production" {
		return []string{
			"https://hireflow.com",
			"https://app.hireflow.com",
		}
	}
	return []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:4200",
	}
}
