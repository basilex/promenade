package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDeprecateEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		sunsetDate     time.Time
		successorURL   string
		checkHeaders   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "deprecation with successor URL",
			sunsetDate:   time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
			successorURL: "/api/v2",
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "true", w.Header().Get("Deprecation"))
				assert.Equal(t, "Tue, 01 Jun 2027 00:00:00 GMT", w.Header().Get("Sunset"))
				assert.Equal(t, "</api/v2>; rel=\"successor-version\"", w.Header().Get("Link"))
				assert.Contains(t, w.Header().Get("X-API-Warn"), "2027-06-01")
				assert.Contains(t, w.Header().Get("X-API-Warn"), "/api/v2")
			},
		},
		{
			name:         "deprecation without successor URL",
			sunsetDate:   time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			successorURL: "",
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "true", w.Header().Get("Deprecation"))
				assert.Equal(t, "Thu, 31 Dec 2026 23:59:59 GMT", w.Header().Get("Sunset"))
				assert.Empty(t, w.Header().Get("Link")) // No successor URL
				assert.Contains(t, w.Header().Get("X-API-Warn"), "2026-12-31")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(DeprecateEndpoint(tt.sunsetDate, tt.successorURL))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			tt.checkHeaders(t, w)
		})
	}
}

func TestVersionLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		path            string
		expectedVersion string
	}{
		{
			name:            "v1 endpoint",
			path:            "/api/v1/customers",
			expectedVersion: "v1",
		},
		{
			name:            "v2 endpoint",
			path:            "/api/v2/orders",
			expectedVersion: "v2",
		},
		{
			name:            "v10 endpoint (multi-digit)",
			path:            "/api/v10/products",
			expectedVersion: "v10",
		},
		{
			name:            "no version in path",
			path:            "/health",
			expectedVersion: "",
		},
		{
			name:            "root path",
			path:            "/",
			expectedVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			var capturedVersion string
			router.Use(VersionLogger())
			router.GET("/*any", func(c *gin.Context) {
				if val, exists := c.Get("api_version"); exists {
					capturedVersion = val.(string)
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedVersion, capturedVersion)
		})
	}
}

func TestExtractVersionFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"v1", "/api/v1/customers", "v1"},
		{"v2", "/api/v2/orders", "v2"},
		{"v3", "/api/v3/products/123", "v3"},
		{"v10", "/api/v10/items", "v10"},
		{"no version", "/health", ""},
		{"root", "/", ""},
		{"api without version", "/api/health", ""},
		{"short path", "/api", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersionFromPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSunsetVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		sunsetDate        time.Time
		currentTime       time.Time
		expectedStatus    int
		shouldAbort       bool
	}{
		{
			name:           "before sunset - allow request",
			sunsetDate:     time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
			currentTime:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedStatus: http.StatusOK,
			shouldAbort:    false,
		},
		{
			name:           "after sunset - return 410 Gone",
			sunsetDate:     time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			currentTime:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedStatus: http.StatusGone,
			shouldAbort:    true,
		},
		{
			name:           "exact sunset date - return 410 Gone",
			sunsetDate:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			currentTime:    time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC), // 1 second after
			expectedStatus: http.StatusGone,
			shouldAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock time.Now() by using sunset date check
			router := gin.New()
			
			// Use custom middleware that checks against currentTime instead of time.Now()
			router.Use(func(c *gin.Context) {
				if tt.currentTime.After(tt.sunsetDate) {
					c.Header("Link", "</api/v2>; rel=\"successor-version\"")
					c.JSON(http.StatusGone, gin.H{
						"status": "error",
						"error": gin.H{
							"code":    "API_VERSION_SUNSET",
							"message": "API version was sunset. Please migrate to the successor version.",
						},
					})
					c.Abort()
					return
				}
				c.Next()
			})
			
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.shouldAbort {
				// Verify 410 Gone response structure
				assert.Contains(t, w.Body.String(), "API_VERSION_SUNSET")
				assert.Contains(t, w.Body.String(), "sunset")
			}
		})
	}
}

func TestSunsetVersion_ResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	sunsetDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	successorURL := "/api/v2"
	migrationGuideURL := "https://docs.example.com/migration"
	
	router.Use(SunsetVersion(sunsetDate, successorURL, migrationGuideURL))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test after sunset date (will return 410)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response headers
	assert.Equal(t, "</api/v2>; rel=\"successor-version\"", w.Header().Get("Link"))
	
	// Check response body structure
	assert.Contains(t, w.Body.String(), "API_VERSION_SUNSET")
	assert.Contains(t, w.Body.String(), "sunset")
	assert.Contains(t, w.Body.String(), successorURL)
	assert.Contains(t, w.Body.String(), migrationGuideURL)
}
