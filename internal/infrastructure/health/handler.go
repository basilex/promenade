package health

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// IChecker defines the interface for health checking
type IChecker interface {
	CheckAll(ctx context.Context) *Report
	CheckDatabase(ctx context.Context) Check
	CheckRedis(ctx context.Context) Check
	CheckEventBus(ctx context.Context) Check
}

// Handler provides HTTP endpoints for health checks
type Handler struct {
	checker IChecker
}

// NewHandler creates a new health check handler
func NewHandler(checker IChecker) *Handler {
	return &Handler{
		checker: checker,
	}
}

// RegisterRoutes registers health check endpoints
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	health := r.Group("/health")
	{
		health.GET("", h.CheckAll)
		health.GET("/db", h.CheckDatabase)
		health.GET("/redis", h.CheckRedis)
		health.GET("/bus", h.CheckEventBus)
	}
}

// CheckAll godoc
// @Summary Check all dependencies
// @Description Returns health status of all dependencies (database, redis, event bus)
// @Tags Health
// @Produce json
// @Success 200 {object} Report "All systems healthy"
// @Success 503 {object} Report "One or more systems unhealthy"
// @Router /health [get]
func (h *Handler) CheckAll(c *gin.Context) {
	report := h.checker.CheckAll(c.Request.Context())

	statusCode := http.StatusOK
	if report.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if report.Status == StatusDegraded {
		statusCode = http.StatusOK // Still operational, just degraded
	}

	c.JSON(statusCode, report)
}

// CheckDatabase godoc
// @Summary Check database health
// @Description Returns PostgreSQL database health status
// @Tags Health
// @Produce json
// @Success 200 {object} Check "Database healthy"
// @Success 503 {object} Check "Database unhealthy"
// @Router /health/db [get]
func (h *Handler) CheckDatabase(c *gin.Context) {
	check := h.checker.CheckDatabase(c.Request.Context())

	statusCode := http.StatusOK
	if check.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, check)
}

// CheckRedis godoc
// @Summary Check Redis health
// @Description Returns Redis health status (if configured)
// @Tags Health
// @Produce json
// @Success 200 {object} Check "Redis healthy or not configured"
// @Success 503 {object} Check "Redis unhealthy"
// @Router /health/redis [get]
func (h *Handler) CheckRedis(c *gin.Context) {
	check := h.checker.CheckRedis(c.Request.Context())

	statusCode := http.StatusOK
	if check.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, check)
}

// CheckEventBus godoc
// @Summary Check Event Bus health
// @Description Returns Event Bus health status
// @Tags Health
// @Produce json
// @Success 200 {object} Check "Event Bus healthy"
// @Success 503 {object} Check "Event Bus unhealthy"
// @Router /health/bus [get]
func (h *Handler) CheckEventBus(c *gin.Context) {
	check := h.checker.CheckEventBus(c.Request.Context())

	statusCode := http.StatusOK
	if check.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, check)
}
