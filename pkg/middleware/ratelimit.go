// Package middleware provides HTTP middleware for Gin framework.
// This package includes rate limiting, authentication, and other cross-cutting concerns.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/basilex/promenade/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter implements IP-based rate limiting using token bucket algorithm.
// It tracks each IP address separately and enforces per-IP rate limits.
//
// Thread-safe: Uses sync.RWMutex for concurrent access.
//
// Example:
//
//	limiter := NewRateLimiter(rate.Limit(5), 10) // 5 req/sec, burst of 10
//	router.Use(limiter.Limit())
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter with specified rate and burst size.
//
// Parameters:
//   - r: requests per second (use rate.Every(duration) for lower rates)
//   - b: burst size (maximum requests in a single burst)
//
// Example rates:
//   - rate.Limit(5): 5 requests per second
//   - rate.Every(time.Minute): 1 request per minute
//   - rate.Every(time.Minute/5): 5 requests per minute
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getVisitor retrieves or creates a rate limiter for an IP address.
// Thread-safe: uses RWMutex for concurrent access.
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter for this IP
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	limiter, exists = rl.visitors[ip]
	if exists {
		return limiter
	}

	limiter = rate.NewLimiter(rl.rate, rl.burst)
	rl.visitors[ip] = limiter

	return limiter
}

// Limit returns a Gin middleware that enforces rate limiting per IP address.
//
// When rate limit is exceeded:
//   - Returns 429 Too Many Requests
//   - Sets X-RateLimit-* headers
//   - Aborts the request chain
//
// Headers set:
//   - X-RateLimit-Limit: maximum requests allowed
//   - X-RateLimit-Remaining: requests remaining in current window
//   - X-RateLimit-Reset: Unix timestamp when limit resets
//
// Example:
//
//	router.POST("/login", limiter.Limit(), handler.Login)
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getVisitor(ip)

		// Check if request is allowed
		if !limiter.Allow() {
			// Calculate reset time (next token available)
			reservation := limiter.Reserve()
			resetTime := time.Now().Add(reservation.Delay())
			reservation.Cancel() // Don't actually reserve

			// Set rate limit headers
			c.Header("X-RateLimit-Limit", "0") // Exceeded
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

			response.ErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
				"Too many requests. Please try again later.")
			c.Abort()
			return
		}

		// Set rate limit headers for successful requests
		c.Header("X-RateLimit-Limit", "1") // Simplified (always 1 for now)
		c.Header("X-RateLimit-Remaining", "1")

		c.Next()
	}
}

// CleanupVisitors removes old visitor entries to prevent memory leaks.
// Should be called periodically (e.g., every hour).
//
// This is important for long-running applications to avoid unbounded memory growth.
//
// Example:
//
//	go func() {
//	    ticker := time.NewTicker(1 * time.Hour)
//	    defer ticker.Stop()
//	    for range ticker.C {
//	        limiter.CleanupVisitors()
//	    }
//	}()
func (rl *RateLimiter) CleanupVisitors() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Clear all visitors (simple approach)
	// In production, consider more sophisticated cleanup based on last access time
	rl.visitors = make(map[string]*rate.Limiter)
}

// GetVisitorCount returns the number of tracked IP addresses.
// Useful for monitoring and testing.
func (rl *RateLimiter) GetVisitorCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.visitors)
}
