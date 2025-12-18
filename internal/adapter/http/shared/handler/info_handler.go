package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIInfo represents API metadata
type APIInfo struct {
	Service     string    `json:"service"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
}

// APIVersionInfo represents API version details
type APIVersionInfo struct {
	Version       string    `json:"version"`
	BasePath      string    `json:"base_path"`
	Documentation string    `json:"documentation"`
	HealthCheck   string    `json:"health_check"`
	Timestamp     time.Time `json:"timestamp"`
}

// APIVersionsInfo represents available API versions
type APIVersionsInfo struct {
	Service   string          `json:"service"`
	Versions  []string        `json:"versions"`
	V1        *APIVersionLink `json:"v1"`
	V2        *APIVersionLink `json:"v2,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// APIVersionLink represents a link to a specific API version
type APIVersionLink struct {
	BasePath      string `json:"base_path"`
	Documentation string `json:"documentation"`
	HealthCheck   string `json:"health_check"`
}

// InfoHandler handles API information endpoints
type InfoHandler struct {
	serviceName string
	version     string
	environment string
	host        string
	port        string
}

// NewInfoHandler creates a new info handler
func NewInfoHandler(serviceName, version, environment, host, port string) *InfoHandler {
	return &InfoHandler{
		serviceName: serviceName,
		version:     version,
		environment: environment,
		host:        host,
		port:        port,
	}
}

// GetAPIInfo handles GET /api - shows available API versions
func (h *InfoHandler) GetAPIInfo(c *gin.Context) {
	baseURL := h.getBaseURL(c)

	info := APIVersionsInfo{
		Service:   h.serviceName,
		Versions:  []string{"v1", "v2"},
		Timestamp: time.Now().UTC(),
		V1: &APIVersionLink{
			BasePath:      baseURL + "/api/v1",
			Documentation: baseURL + "/api/v1/docs/swagger/index.html",
			HealthCheck:   baseURL + "/api/v1/health",
		},
		V2: &APIVersionLink{
			BasePath:      baseURL + "/api/v2",
			Documentation: baseURL + "/api/v2/docs/swagger/index.html",
			HealthCheck:   baseURL + "/api/v2/health",
		},
	}

	c.JSON(http.StatusOK, info)
}

// GetV1Info handles GET /api/v1 - shows v1 API information
func (h *InfoHandler) GetV1Info(c *gin.Context) {
	baseURL := h.getBaseURL(c)

	info := APIVersionInfo{
		Version:       "v1",
		BasePath:      "/api/v1",
		Documentation: baseURL + "/api/v1/docs/swagger/index.html",
		HealthCheck:   baseURL + "/api/v1/health",
		Timestamp:     time.Now().UTC(),
	}

	c.JSON(http.StatusOK, info)
}

// GetV2Info handles GET /api/v2 - shows v2 API information
func (h *InfoHandler) GetV2Info(c *gin.Context) {
	baseURL := h.getBaseURL(c)

	info := APIVersionInfo{
		Version:       "v2",
		BasePath:      "/api/v2",
		Documentation: baseURL + "/api/v2/docs/swagger/index.html",
		HealthCheck:   baseURL + "/api/v2/health",
		Timestamp:     time.Now().UTC(),
	}

	c.JSON(http.StatusOK, info)
}

// getBaseURL constructs the base URL from request or config
func (h *InfoHandler) getBaseURL(c *gin.Context) string {
	// Try to get from request
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	// Check X-Forwarded-Proto header (for reverse proxies)
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	host := c.Request.Host
	if host == "" {
		// Fallback to config
		host = h.host
		if host == "0.0.0.0" || host == "" {
			host = "localhost"
		}
		host = host + ":" + h.port
	}

	return scheme + "://" + host
}
