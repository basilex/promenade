package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DeprecateEndpoint adds deprecation headers to responses
// Use this middleware to mark endpoints as deprecated before removing them
//
// Example:
//
//	v1 := router.Group("/api/v1")
//	v1.Use(middleware.DeprecateEndpoint(
//	    time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
//	    "/api/v2",
//	))
func DeprecateEndpoint(sunsetDate time.Time, successorURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// RFC 8594 Deprecation header
		c.Header("Deprecation", "true")

		// RFC 8594 Sunset header (HTTP date format)
		c.Header("Sunset", sunsetDate.Format(http.TimeFormat))

		// Link header pointing to successor version
		if successorURL != "" {
			c.Header("Link", fmt.Sprintf("<%s>; rel=\"successor-version\"", successorURL))
		}

		// Custom warning header with human-readable date
		c.Header("X-API-Warn", fmt.Sprintf(
			"This API version is deprecated and will be removed on %s. Please migrate to %s",
			sunsetDate.Format("2006-01-02"),
			successorURL,
		))

		c.Next()
	}
}

// VersionLogger logs API version usage for analytics
// Useful for tracking which versions are still in use
//
// Example:
//
//	router.Use(middleware.VersionLogger())
func VersionLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract version from URL path (e.g., /api/v1/customers → v1)
		version := extractVersionFromPath(c.Request.URL.Path)

		// Store version in context for later use (metrics, logging, etc.)
		c.Set("api_version", version)

		c.Next()
	}
}

// extractVersionFromPath extracts API version from URL path
// Returns empty string if version not found
func extractVersionFromPath(path string) string {
	// Simple pattern matching: /api/v1/... → v1
	if len(path) >= 8 && path[:5] == "/api/" {
		// Check for v1, v2, v3, etc.
		if path[5] == 'v' && len(path) > 7 {
			// Find end of version (next slash or end of string)
			end := 7
			for end < len(path) && path[end] != '/' {
				end++
			}
			return path[5:end] // Return "v1", "v2", etc.
		}
	}
	return ""
}

// SunsetVersion returns 410 Gone for sunsetted API versions
// Use this to completely disable old API versions after deprecation period
//
// Example:
//
//	v1 := router.Group("/api/v1")
//	v1.Use(middleware.SunsetVersion(
//	    time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
//	    "/api/v2",
//	    "https://docs.promenade.example.com/migration/v1-to-v2",
//	))
func SunsetVersion(sunsetDate time.Time, successorURL, migrationGuideURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if current date is past sunset date
		if time.Now().After(sunsetDate) {
			// Return 410 Gone with migration information
			c.Header("Link", fmt.Sprintf("<%s>; rel=\"successor-version\"", successorURL))
			c.JSON(http.StatusGone, gin.H{
				"status": "error",
				"error": gin.H{
					"code":    "API_VERSION_SUNSET",
					"message": fmt.Sprintf("API version was sunset on %s. Please migrate to the successor version.", sunsetDate.Format("2006-01-02")),
					"details": gin.H{
						"sunset_date":       sunsetDate.Format("2006-01-02"),
						"successor_version": successorURL,
						"migration_guide":   migrationGuideURL,
					},
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
