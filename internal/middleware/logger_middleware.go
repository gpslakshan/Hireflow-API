package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// LoggerMiddleware logs every HTTP request as a structured JSON event.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process the request
		c.Next()

		// Log after the handler has run — we now have the status code
		duration := time.Since(start)
		status := c.Writer.Status()

		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Int64("duration_ms", duration.Milliseconds()).
			Str("client_ip", c.ClientIP()).
			Msg("request")
	}
}
