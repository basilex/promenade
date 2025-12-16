package middleware

import (
	"log/slog"
	"time"

	"github.com/basilex/promenade/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Get request ID from context
		requestID := c.GetString("request_id")

		c.Next()

		duration := time.Since(startTime)

		// Structured logging with context
		logger.Default().Info("HTTP request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", duration),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.String("request_id", requestID),
			slog.Int("response_size", c.Writer.Size()),
		)
	}
}
