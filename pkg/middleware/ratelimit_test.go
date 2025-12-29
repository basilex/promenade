package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(rate.Limit(5), 10)

	assert.NotNil(t, limiter)
	assert.NotNil(t, limiter.visitors)
	assert.Equal(t, rate.Limit(5), limiter.rate)
	assert.Equal(t, 10, limiter.burst)
	assert.Equal(t, 0, limiter.GetVisitorCount())
}

func TestRateLimiter_AllowsRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create limiter: 5 requests per minute
	limiter := NewRateLimiter(rate.Every(time.Minute/5), 1)

	router := gin.New()
	router.GET("/test", limiter.Limit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request should succeed
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestRateLimiter_BlocksExcessRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create very restrictive limiter: 1 request per minute, burst of 1
	limiter := NewRateLimiter(rate.Every(time.Minute), 1)

	router := gin.New()
	router.GET("/test", limiter.Limit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request should succeed
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be blocked
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Contains(t, w2.Body.String(), "RATE_LIMIT_EXCEEDED")
	assert.Contains(t, w2.Body.String(), "Too many requests")
}

func TestRateLimiter_SetsRateLimitHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewRateLimiter(rate.Limit(5), 1)

	router := gin.New()
	router.GET("/test", limiter.Limit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w, req)

	// Check headers are set
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimiter_SetsResetHeaderWhenLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Very restrictive limiter
	limiter := NewRateLimiter(rate.Every(time.Minute), 1)

	router := gin.New()
	router.GET("/test", limiter.Limit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request (allowed)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w1, req1)

	// Second request (blocked)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w2, req2)

	// Check reset header is set
	resetHeader := w2.Header().Get("X-RateLimit-Reset")
	assert.NotEmpty(t, resetHeader)

	// Parse and verify it's a future time
	resetTime, err := time.Parse(time.RFC3339, resetHeader)
	assert.NoError(t, err)
	assert.True(t, resetTime.After(time.Now()))
}

func TestRateLimiter_SeparatesIPAddresses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Very restrictive limiter: 1 request per minute
	limiter := NewRateLimiter(rate.Every(time.Minute), 1)

	router := gin.New()
	router.GET("/test", limiter.Limit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Request from IP 1
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Request from IP 2 (should succeed - different IP)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Second request from IP 1 (should fail)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}

func TestRateLimiter_AllowsBurstRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Limiter with burst of 3
	limiter := NewRateLimiter(rate.Every(time.Minute), 3)

	router := gin.New()
	router.GET("/test", limiter.Limit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First 3 requests should succeed (burst)
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 4th request should fail
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest("GET", "/test", nil)
	req4.RemoteAddr = "192.168.1.1:1234"
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusTooManyRequests, w4.Code)
}

func TestRateLimiter_CleanupVisitors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewRateLimiter(rate.Limit(5), 10)

	router := gin.New()
	router.GET("/test", limiter.Limit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make requests from 3 different IPs
	ips := []string{"192.168.1.1:1234", "192.168.1.2:1234", "192.168.1.3:1234"}
	for _, ip := range ips {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		router.ServeHTTP(w, req)
	}

	// Verify visitors are tracked
	assert.Equal(t, 3, limiter.GetVisitorCount())

	// Cleanup
	limiter.CleanupVisitors()

	// Verify visitors are cleared
	assert.Equal(t, 0, limiter.GetVisitorCount())
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewRateLimiter(rate.Limit(100), 10)

	router := gin.New()
	router.GET("/test", limiter.Limit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make concurrent requests
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 5; j++ {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.1:1234"
				router.ServeHTTP(w, req)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test should complete without panics
	assert.True(t, true)
}

func TestRateLimiter_GetVisitorCount(t *testing.T) {
	limiter := NewRateLimiter(rate.Limit(5), 10)

	// Initially zero
	assert.Equal(t, 0, limiter.GetVisitorCount())

	// Add visitors manually
	limiter.getVisitor("192.168.1.1")
	assert.Equal(t, 1, limiter.GetVisitorCount())

	limiter.getVisitor("192.168.1.2")
	assert.Equal(t, 2, limiter.GetVisitorCount())

	// Same IP doesn't increase count
	limiter.getVisitor("192.168.1.1")
	assert.Equal(t, 2, limiter.GetVisitorCount())
}
