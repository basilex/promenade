package handler

import (
	"net/http"
	"time"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/pkg/version"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns service health status
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]any}
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"status":  "ok",
		"service": version.ServiceID,
		"time":    time.Now().Unix(),
	})
}
