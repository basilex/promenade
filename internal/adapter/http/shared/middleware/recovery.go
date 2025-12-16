package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := c.GetString("request_id")

				logger.Default().Error("PANIC recovered",
					slog.Any("panic", err),
					slog.String("stack", string(debug.Stack())),
					slog.String("request_id", requestID),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("ip", c.ClientIP()),
				)

				response.Error(c, http.StatusInternalServerError, "internal server error", nil)
				c.Abort()
			}
		}()

		c.Next()
	}
}
