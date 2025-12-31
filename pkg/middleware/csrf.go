package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRFConfig holds CSRF middleware configuration
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token in bytes (default: 32)
	TokenLength int

	// CookieName is the name of the CSRF cookie (default: "csrf_token")
	CookieName string

	// HeaderName is the name of the CSRF header (default: "X-CSRF-Token")
	HeaderName string

	// CookieMaxAge is the cookie max age in seconds (default: 12 hours)
	CookieMaxAge int

	// CookiePath is the cookie path (default: "/")
	CookiePath string

	// CookieDomain is the cookie domain (default: "")
	CookieDomain string

	// CookieSecure indicates if cookie should only be sent over HTTPS (default: true in production)
	CookieSecure bool

	// CookieHTTPOnly indicates if cookie should be HTTP only (default: true)
	CookieHTTPOnly bool

	// CookieSameSite is the SameSite attribute (default: http.SameSiteStrictMode)
	CookieSameSite http.SameSite

	// SkipMethods is a list of HTTP methods to skip CSRF check (default: GET, HEAD, OPTIONS)
	SkipMethods []string

	// ErrorHandler is a custom error handler (optional)
	ErrorHandler func(c *gin.Context)
}

// DefaultCSRFConfig returns default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:    32,
		CookieName:     "csrf_token",
		HeaderName:     "X-CSRF-Token",
		CookieMaxAge:   43200, // 12 hours
		CookiePath:     "/",
		CookieDomain:   "",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		SkipMethods:    []string{"GET", "HEAD", "OPTIONS"},
		ErrorHandler:   nil,
	}
}

// CSRFMiddleware creates a CSRF protection middleware
// NOTE: Only use this middleware if you use cookie-based authentication
// Bearer JWT tokens in Authorization header are safe from CSRF by design
func CSRFMiddleware(config CSRFConfig) gin.HandlerFunc {
	// Apply defaults
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = 43200 // 12 hours
	}
	if config.CookiePath == "" {
		config.CookiePath = "/"
	}
	if config.SkipMethods == nil {
		config.SkipMethods = []string{"GET", "HEAD", "OPTIONS"}
	}

	return func(c *gin.Context) {
		// Skip CSRF check for safe methods
		if contains(config.SkipMethods, c.Request.Method) {
			// Generate and set CSRF token for GET requests
			if c.Request.Method == "GET" {
				token := generateCSRFToken(config.TokenLength)
				setCSRFCookie(c, config, token)
				c.Set("csrf_token", token)
			}
			c.Next()
			return
		}

		// Get token from cookie
		cookieToken, err := c.Cookie(config.CookieName)
		if err != nil || cookieToken == "" {
			handleCSRFError(c, config, "CSRF token not found in cookie")
			return
		}

		// Get token from header
		headerToken := c.GetHeader(config.HeaderName)
		if headerToken == "" {
			handleCSRFError(c, config, "CSRF token not found in header")
			return
		}

		// Validate tokens match
		if cookieToken != headerToken {
			handleCSRFError(c, config, "CSRF token mismatch")
			return
		}

		// Token is valid, proceed
		c.Set("csrf_token", cookieToken)
		c.Next()
	}
}

// generateCSRFToken generates a random CSRF token
func generateCSRFToken(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based token (not cryptographically secure)
		return base64.URLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.URLEncoding.EncodeToString(b)
}

// setCSRFCookie sets the CSRF token cookie
func setCSRFCookie(c *gin.Context, config CSRFConfig, token string) {
	c.SetCookie(
		config.CookieName,
		token,
		config.CookieMaxAge,
		config.CookiePath,
		config.CookieDomain,
		config.CookieSecure,
		config.CookieHTTPOnly,
	)
	c.SetSameSite(config.CookieSameSite)
}

// handleCSRFError handles CSRF validation errors
func handleCSRFError(c *gin.Context, config CSRFConfig, message string) {
	if config.ErrorHandler != nil {
		config.ErrorHandler(c)
		return
	}

	c.JSON(http.StatusForbidden, gin.H{
		"status": "error",
		"error": gin.H{
			"code":    "CSRF_TOKEN_INVALID",
			"message": message,
		},
	})
	c.Abort()
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// GetCSRFToken retrieves the CSRF token from the context
func GetCSRFToken(c *gin.Context) string {
	token, exists := c.Get("csrf_token")
	if !exists {
		return ""
	}
	tokenStr, ok := token.(string)
	if !ok {
		return ""
	}
	return tokenStr
}
