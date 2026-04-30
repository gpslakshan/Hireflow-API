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
) *gin.Engine {
	r := gin.New() // gin.New() instead of gin.Default() — we attach our own middleware

	r.Use(gin.Recovery()) // recover from panics — always include this

	// Trust only localhost proxies
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// API v1 group
	v1 := r.Group("/api/v1")

	// Public auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.Me)
	}

	return r
}
