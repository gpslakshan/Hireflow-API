package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gpslakshan/hireflow/internal/config"
	"github.com/gpslakshan/hireflow/internal/handler"
	"github.com/gpslakshan/hireflow/internal/middleware"
)

func Setup(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	companyHandler *handler.CompanyHandler,
	jobHandler *handler.JobHandler,
	appHandler *handler.ApplicationHandler,
) *gin.Engine {
	r := gin.New()                             // gin.New() instead of gin.Default() — we attach our own middleware
	r.Use(gin.Recovery())                      // recover from panics — always include this
	r.SetTrustedProxies([]string{"127.0.0.1"}) // Trust only localhost proxies. In production, set this to your actual proxy IPs or use gin's CIDR notation for broader ranges.

	// API v1 group
	v1 := r.Group("/api/v1")

	// ── Auth (public) ──────────────────────────────────────
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.Me)
	}

	// ── Companies ──────────────────────────────────────────
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
		// POST /companies/:id/jobs
		companies.POST("/:id/jobs",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("recruiter"),
			jobHandler.Create,
		)
	}

	// ── Jobs ───────────────────────────────────────────────
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
		// POST /jobs/:id/apply
		jobs.POST("/:id/apply",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("candidate"),
			appHandler.Apply,
		)
		// GET /jobs/:id/applications
		jobs.GET("/:id/applications",
			middleware.AuthMiddleware(cfg),
			middleware.RequireRole("recruiter"),
			appHandler.GetByJob,
		)
	}

	// ── Applications ───────────────────────────────────────
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

	return r
}
