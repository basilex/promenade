package middleware

import (
    "log"
    "net/http"
    "runtime/debug"
    
    "github.com/gin-gonic/gin"
    "github.com/basilex/promenade/internal/adapter/http/shared/response"
)

func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("PANIC: %v\n%s", err, debug. Stack())
                response.Error(c, http.StatusInternalServerError, "internal server error", nil)
                c.Abort()
            }
        }()

        c.Next()
    }
}
