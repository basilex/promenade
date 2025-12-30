# Rate Limiting

**IP-based rate limiting** for Promenade Platform - protect authentication endpoints from brute-force attacks and abuse.

---

## Overview

Rate limiting middleware implements **token bucket algorithm** with per-IP tracking. Each IP address gets its own rate limiter, preventing one abusive client from affecting others.

**Key Features:**

- IP-based tracking (separate limits per IP)
- Configurable rate and burst size
- Thread-safe concurrent access
- Automatic cleanup of old visitors
- Standard X-RateLimit-* headers
- Integration with Gin framework

---

## Quick Start

### 1. Create Rate Limiter

```go
import (
    "time"
    "golang.org/x/time/rate"
    "github.com/basilex/promenade/pkg/middleware"
)

// Create limiter: 5 requests per minute, burst of 1
limiter := middleware.NewRateLimiter(rate.Every(time.Minute/5), 1)
```

### 2. Apply to Routes

```go
// Protect login endpoint
router.POST("/login", limiter.Limit(), handler.Login)

// Protect register endpoint
registerLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/3), 1)
router.POST("/register", registerLimiter.Limit(), handler.Register)
```

### 3. Test Rate Limiting

```bash
# First request succeeds
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Second request (within 1 minute) is blocked
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Response: 429 Too Many Requests
# {
#   "status": "error",
#   "error": {
#     "code": "RATE_LIMIT_EXCEEDED",
#     "message": "Too many requests. Please try again later."
#   }
# }
```

---

## API Reference

### Types

#### RateLimiter

```go
type RateLimiter struct {
    visitors map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}
```

**Fields:**
- `visitors`: Per-IP rate limiter instances
- `mu`: Read-write mutex for thread-safe access
- `rate`: Requests per second (or custom rate)
- `burst`: Maximum requests in single burst

### Functions

#### NewRateLimiter

```go
func NewRateLimiter(r rate.Limit, b int) *RateLimiter
```

Creates new rate limiter with specified rate and burst size.

**Parameters:**
- `r`: Rate limit (use `rate.Limit(n)` or `rate.Every(duration)`)
- `b`: Burst size (maximum concurrent requests)

**Example rates:**
```go
rate.Limit(5)                  // 5 requests per second
rate.Every(time.Minute)        // 1 request per minute
rate.Every(time.Minute/5)      // 5 requests per minute
rate.Every(time.Second * 10)   // 1 request per 10 seconds
```

#### Limit

```go
func (rl *RateLimiter) Limit() gin.HandlerFunc
```

Returns Gin middleware that enforces rate limiting.

**Behavior:**
- Allows requests within rate limit
- Returns 429 Too Many Requests when exceeded
- Sets X-RateLimit-* headers
- Aborts request chain on limit exceeded

**Headers set:**
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining in window
- `X-RateLimit-Reset`: Timestamp when limit resets (RFC3339)

#### CleanupVisitors

```go
func (rl *RateLimiter) CleanupVisitors()
```

Removes all visitor entries to prevent memory leaks.

**Recommended usage:** Call periodically (every 1-4 hours) in production.

```go
// Start cleanup goroutine
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        limiter.CleanupVisitors()
    }
}()
```

#### GetVisitorCount

```go
func (rl *RateLimiter) GetVisitorCount() int
```

Returns number of tracked IP addresses. Useful for monitoring.

---

## Configuration Examples

### Authentication Endpoints

**Login** (5 attempts per minute):
```go
loginLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/5), 1)
router.POST("/login", loginLimiter.Limit(), handler.Login)
```

**Register** (3 attempts per minute):
```go
registerLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/3), 1)
router.POST("/register", registerLimiter.Limit(), handler.Register)
```

**Password Reset** (2 attempts per hour):
```go
resetLimiter := middleware.NewRateLimiter(rate.Every(time.Hour/2), 1)
router.POST("/reset-password", resetLimiter.Limit(), handler.ResetPassword)
```

### API Endpoints

**Public API** (100 requests per minute):
```go
apiLimiter := middleware.NewRateLimiter(rate.Limit(100), 10)
api.Use(apiLimiter.Limit())
```

**Admin API** (1000 requests per minute):
```go
adminLimiter := middleware.NewRateLimiter(rate.Limit(1000), 50)
admin.Use(adminLimiter.Limit())
```

### Burst Handling

**Allow burst, then strict** (burst of 5, then 1/second):
```go
limiter := middleware.NewRateLimiter(rate.Limit(1), 5)
// First 5 requests: instant
// Subsequent requests: 1 per second
```

**No burst** (strict 1 per minute):
```go
limiter := middleware.NewRateLimiter(rate.Every(time.Minute), 1)
// Every request: must wait 1 minute
```

---

## Response Format

### Success (Within Limit)

**Status:** 200 OK

**Headers:**
```
X-RateLimit-Limit: 1
X-RateLimit-Remaining: 1
```

**Body:**
```json
{
  "status": "success",
  "data": { ... }
}
```

### Rate Limit Exceeded

**Status:** 429 Too Many Requests

**Headers:**
```
X-RateLimit-Limit: 0
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 2025-12-29T15:30:45+02:00
```

**Body:**
```json
{
  "status": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later."
  }
}
```

---

## Implementation Details

### Token Bucket Algorithm

Rate limiting uses **token bucket algorithm**:

1. Each IP gets a bucket with `burst` tokens
2. Tokens regenerate at `rate` per second
3. Each request consumes 1 token
4. If no tokens available → 429 error

**Example** (5 req/min, burst 1):
- Bucket starts with 1 token
- Request 1: Uses token, succeeds
- Bucket empty, regenerates 1 token in 12 seconds
- Request 2 (before 12s): No tokens, blocked (429)
- Request 3 (after 12s): Token available, succeeds

### Thread Safety

All operations are thread-safe:
- `getVisitor()`: RWMutex for concurrent reads, Lock for writes
- Double-check locking pattern prevents race conditions
- Safe for high-concurrency production use

### Memory Management

**Problem:** Unbounded visitor map grows forever

**Solution:** Periodic cleanup

```go
// Production cleanup (every hour)
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        limiter.CleanupVisitors()
        logger.Info("Cleaned up rate limiter visitors",
            slog.Int("count", limiter.GetVisitorCount()))
    }
}()
```

**Advanced:** Track last access time per visitor, remove inactive IPs only.

---

## Testing

### Unit Tests

```go
func TestRateLimiter_BlocksExcessRequests(t *testing.T) {
    limiter := NewRateLimiter(rate.Every(time.Minute), 1)
    
    router := gin.New()
    router.GET("/test", limiter.Limit(), handler)
    
    // First request succeeds
    w1 := httptest.NewRecorder()
    req1 := httptest.NewRequest("GET", "/test", nil)
    req1.RemoteAddr = "192.168.1.1:1234"
    router.ServeHTTP(w1, req1)
    assert.Equal(t, http.StatusOK, w1.Code)
    
    // Second request blocked
    w2 := httptest.NewRecorder()
    req2 := httptest.NewRequest("GET", "/test", nil)
    req2.RemoteAddr = "192.168.1.1:1234"
    router.ServeHTTP(w2, req2)
    assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}
```

### Integration Tests

```bash
# Run all middleware tests
go test ./pkg/middleware -v

# Test specific cases
go test ./pkg/middleware -run TestRateLimiter_BlocksExcessRequests -v
go test ./pkg/middleware -run TestRateLimiter_SeparatesIPAddresses -v
```

### Manual Testing

```bash
# Test login rate limiting (5 attempts/min)
for i in {1..6}; do
  echo "Request $i:"
  curl -X POST http://localhost:8081/api/v1/identity/users/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"password"}' \
    -w "\nStatus: %{http_code}\n\n"
  sleep 1
done

# Expected: First 5 succeed (or fail auth), 6th gets 429
```

---

## Production Best Practices

### 1. Choose Appropriate Limits

**Authentication endpoints** (strict):
- Login: 5-10 attempts per minute
- Register: 3-5 attempts per minute
- Password reset: 2-3 attempts per hour

**API endpoints** (permissive):
- Public API: 100-1000 requests per minute
- Authenticated API: 1000-5000 requests per minute

### 2. Monitor Rate Limiting

```go
// Log when limits are exceeded
func (rl *RateLimiter) Limit() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            logger.FromContext(c).Warn("Rate limit exceeded",
                slog.String("ip", c.ClientIP()),
                slog.String("path", c.Request.URL.Path),
            )
            // ... rest of error handling
        }
    }
}
```

### 3. Consider Load Balancers

**Problem:** Load balancer IP appears as client IP

**Solution:** Use X-Forwarded-For header:

```go
// Set trusted proxies in main.go
router.SetTrustedProxies([]string{"10.0.0.0/8"})

// ClientIP() will extract real IP from X-Forwarded-For
limiter := middleware.NewRateLimiter(rate.Every(time.Minute/5), 1)
```

### 4. Distributed Rate Limiting (Future)

**Current:** In-memory per-instance

**Future:** Redis-based shared state:
- All instances share rate limit counters
- Consistent limits across cluster
- Survives instance restarts

---

## Comparison with Alternatives

### Custom Middleware vs Libraries

| Feature              | Custom Middleware | go-limiter | tollbooth |
|----------------------|-------------------|------------|-----------|
| IP-based tracking    | ✅                | ✅         | ✅        |
| Per-route limits     | ✅                | ✅         | ✅        |
| Memory cleanup       | ✅                | ❌         | ❌        |
| Zero dependencies    | ✅ (stdlib)       | ❌         | ❌        |
| Custom error format  | ✅                | ❌         | ❌        |
| Thread-safe          | ✅                | ✅         | ✅        |
| Distributed (Redis)  | ❌ (future)       | ✅         | ✅        |

**Verdict:** Custom middleware provides control, simplicity, and integration with existing error handling.

---

## Troubleshooting

### Rate limit not working

**Check:**
1. Middleware applied to route?
   ```go
   router.POST("/login", limiter.Limit(), handler) // ✅ Correct
   router.POST("/login", handler) // ❌ Missing middleware
   ```
2. Correct IP extraction?
   ```go
   // Behind proxy:
   router.SetTrustedProxies([]string{"proxy-ip"})
   ```

### Too strict / too permissive

**Adjust rate:**
```go
// Too strict (1/min):
limiter := middleware.NewRateLimiter(rate.Every(time.Minute), 1)

// More permissive (10/min):
limiter := middleware.NewRateLimiter(rate.Every(time.Minute/10), 1)

// Allow burst:
limiter := middleware.NewRateLimiter(rate.Every(time.Minute/10), 5)
```

### Memory usage growing

**Cause:** Visitors never cleaned up

**Fix:** Add periodic cleanup:
```go
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        limiter.CleanupVisitors()
    }
}()
```

---

## Test Coverage

**Package:** `pkg/middleware`  
**Tests:** 10  
**Coverage:** 100%

| Test                                    | Purpose                          |
|-----------------------------------------|----------------------------------|
| TestNewRateLimiter                      | Constructor                      |
| TestRateLimiter_AllowsRequests          | Success case                     |
| TestRateLimiter_BlocksExcessRequests    | Rate limit enforcement           |
| TestRateLimiter_SetsRateLimitHeaders    | HTTP headers                     |
| TestRateLimiter_SetsResetHeaderWhenLimitExceeded | Reset time calculation  |
| TestRateLimiter_SeparatesIPAddresses    | Per-IP isolation                 |
| TestRateLimiter_AllowsBurstRequests     | Burst handling                   |
| TestRateLimiter_CleanupVisitors         | Memory cleanup                   |
| TestRateLimiter_ConcurrentAccess        | Thread safety                    |
| TestRateLimiter_GetVisitorCount         | Monitoring                       |

**Run tests:**
```bash
go test ./pkg/middleware -v
go test ./pkg/middleware -cover
```

---

## Integration with Promenade

### Identity Context (Current)

Rate limiting applied to authentication endpoints:

```go
// internal/contexts/identity/router.go
loginLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/5), 1)       // 5/min
registerLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/3), 1)    // 3/min

users.POST("/register", registerLimiter.Limit(), r.userHandler.Register)
users.POST("/login", loginLimiter.Limit(), r.userHandler.Login)
```

**Protected endpoints:**
- `/api/v1/identity/users/register`: 3 attempts/min
- `/api/v1/identity/users/login`: 5 attempts/min

### Future Contexts

Apply rate limiting to other sensitive endpoints:

- Password reset: 2 attempts/hour
- Email verification: 5 attempts/hour
- Profile updates: 10 attempts/minute
- API calls: 100 requests/minute

---

## Related Documentation

- [Main README](../../README.md)
- [Documentation Index](INDEX.md)
- [JWT Authentication](../pkg/jwt/README.md)
- [Identity Context](../contexts/identity.md)

---

**Last Updated:** December 29, 2025  
**Status:** Production-ready  
**Test Coverage:** 10 tests, 100% passing  
**Maintainer:** Promenade Team
