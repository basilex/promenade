package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandler handles HTTP error responses (404, 405, etc.)
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleNotFound handles 404 Not Found errors
// Returns a structured JSON response when a route doesn't exist
func (h *ErrorHandler) HandleNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"success":   false,
		"message":   "route not found",
		"error":     "The requested endpoint does not exist",
		"timestamp": time.Now().Unix(),
	})
}

// HandleMethodNotAllowed handles 405 Method Not Allowed errors
// Returns a structured JSON response when HTTP method is not supported
func (h *ErrorHandler) HandleMethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"success":   false,
		"message":   "method not allowed",
		"error":     "The HTTP method is not supported for this endpoint",
		"timestamp": time.Now().Unix(),
	})
}
