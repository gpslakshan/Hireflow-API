package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// extractUserID pulls the user_id claim injected by AuthMiddleware
// from the Gin request context.
// Returns an error if the claim is missing or not a valid UUID.
func extractUserID(c *gin.Context) (uuid.UUID, error) {
	raw, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}
	return uuid.Parse(raw.(string))
}
