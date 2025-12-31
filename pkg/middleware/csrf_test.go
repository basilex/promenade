package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDefaultCSRFConfig(t *testing.T) {
	config := DefaultCSRFConfig()

	assert.Equal(t, 32, config.TokenLength)
	assert.Equal(t, "csrf_token", config.CookieName)
	assert.Equal(t, "X-CSRF-Token", config.HeaderName)
	assert.Equal(t, 43200, config.CookieMaxAge)
	assert.Equal(t, "/", config.CookiePath)
	assert.True(t, config.CookieSecure)
	assert.True(t, config.CookieHTTPOnly)
	assert.Equal(t, http.SameSiteStrictMode, config.CookieSameSite)
	assert.Contains(t, config.SkipMethods, "GET")
	assert.Contains(t, config.SkipMethods, "HEAD")
	assert.Contains(t, config.SkipMethods, "OPTIONS")
}

func TestCSRFMiddleware_SkipSafeMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware(DefaultCSRFConfig()))
	router.GET("/test", func(c *gin.Context) {
		token := GetCSRFToken(c)
		c.JSON(http.StatusOK, gin.H{"csrf_token": token})
	})

	// GET request should pass without CSRF token
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check CSRF cookie was set
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies)
	assert.Equal(t, "csrf_token", cookies[0].Name)
	assert.NotEmpty(t, cookies[0].Value)
}

func TestCSRFMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := DefaultCSRFConfig()
	router := gin.New()
	router.Use(CSRFMiddleware(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Generate token
	token := generateCSRFToken(32)

	// POST request with valid CSRF token
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "csrf_token",
		Value: token,
	})
	req.Header.Set("X-CSRF-Token", token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_MissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware(DefaultCSRFConfig()))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// POST request without CSRF cookie
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-CSRF-Token", "some-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF_TOKEN_INVALID")
	assert.Contains(t, w.Body.String(), "CSRF token not found in cookie")
}

func TestCSRFMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware(DefaultCSRFConfig()))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// POST request without CSRF header
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "csrf_token",
		Value: "some-token",
	})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF_TOKEN_INVALID")
	assert.Contains(t, w.Body.String(), "CSRF token not found in header")
}

func TestCSRFMiddleware_TokenMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware(DefaultCSRFConfig()))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// POST request with mismatched tokens
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "csrf_token",
		Value: "token-in-cookie",
	})
	req.Header.Set("X-CSRF-Token", "token-in-header")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF_TOKEN_INVALID")
	assert.Contains(t, w.Body.String(), "CSRF token mismatch")
}

func TestCSRFMiddleware_CustomErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := DefaultCSRFConfig()
	customCalled := false
	config.ErrorHandler = func(c *gin.Context) {
		customCalled = true
		c.JSON(http.StatusTeapot, gin.H{"custom": "error"})
		c.Abort()
	}

	router := gin.New()
	router.Use(CSRFMiddleware(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// POST request without CSRF cookie (should trigger error handler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.True(t, customCalled)
	assert.Contains(t, w.Body.String(), "custom")
}

func TestCSRFMiddleware_HEADMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware(DefaultCSRFConfig()))
	router.HEAD("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// HEAD request should pass without CSRF token
	w := httptest.NewRecorder()
	req := httptest.NewRequest("HEAD", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_OPTIONSMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CSRFMiddleware(DefaultCSRFConfig()))
	router.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// OPTIONS request should pass without CSRF token
	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGenerateCSRFToken(t *testing.T) {
	token1 := generateCSRFToken(32)
	token2 := generateCSRFToken(32)

	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2, "Tokens should be unique")
}

func TestGetCSRFToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test when token doesn't exist
	token := GetCSRFToken(c)
	assert.Empty(t, token)

	// Test when token exists
	c.Set("csrf_token", "test-token-123")
	token = GetCSRFToken(c)
	assert.Equal(t, "test-token-123", token)

	// Test when token is wrong type
	c.Set("csrf_token", 123)
	token = GetCSRFToken(c)
	assert.Empty(t, token)
}

func TestContains(t *testing.T) {
	slice := []string{"GET", "POST", "PUT"}

	assert.True(t, contains(slice, "GET"))
	assert.True(t, contains(slice, "POST"))
	assert.False(t, contains(slice, "DELETE"))
	assert.False(t, contains(slice, ""))
}
