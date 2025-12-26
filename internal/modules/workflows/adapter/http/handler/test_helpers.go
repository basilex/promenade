package handler

import (
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

// setupTestRouter creates a test router for handler tests
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// authMiddleware adds user_id to the request context for testing
// CRITICAL: Must set uuidv7.UUID type, not string, to match middleware.GetUserID expectations
func authMiddleware(userID uuidv7.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

// addAuthContext adds user_id directly to the context (for legacy tests)
// CRITICAL: Must set uuidv7.UUID type, not string, to match middleware.GetUserID expectations
func addAuthContext(c *gin.Context, userID uuidv7.UUID) {
	c.Set("user_id", userID)
}
