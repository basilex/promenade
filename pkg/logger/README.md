# Logger Package

**Purpose:** Structured logging with context support (slog wrapper)  
**Status:** Production-ready  
**Tests:** 15 tests, 95% coverage

---

## Overview

The `logger` package provides a thin wrapper around Go's standard `log/slog` with additional convenience methods for context-aware logging. It supports multiple output formats, log levels, and automatic context field extraction.

## Features

- **Structured Logging:** JSON or text format output
- **Context Integration:** Automatic extraction of request ID, user ID, trace ID
- **Customizable Levels:** Debug, Info, Warn, Error
- **Source Location:** Optional file/line information
- **Time Formatting:** Configurable timestamp format
- **Thread-Safe:** Safe for concurrent use

---

## Installation

```go
import "github.com/basilex/promenade/pkg/logger"
```

---

## Quick Start

### Basic Initialization

```go
package main

import (
    "github.com/basilex/promenade/pkg/logger"
)

func main() {
    // Initialize default logger
    logger.Init(logger.Config{
        Level:  "info",
        Format: "text",
        AddSource: true,
    })
    
    logger.Info("Application started", "version", "1.0.0")
}
```

### Configuration

```go
type Config struct {
    Level      string    // "debug", "info", "warn", "error"
    Format     string    // "json", "text"
    Output     io.Writer // defaults to os.Stdout
    AddSource  bool      // include file:line in logs
    TimeFormat string    // defaults to RFC3339
}
```

---

## Usage Examples

### 1. Basic Logging

```go
// Info level
logger.Info("User logged in", "user_id", "123", "ip", "192.168.1.1")

// Debug level
logger.Debug("Processing request", "endpoint", "/api/users")

// Warning
logger.Warn("Rate limit approaching", "current", 95, "max", 100)

// Error
logger.Error("Failed to connect to database", 
    "error", err, 
    "host", "localhost:5432",
)
```

**Output (text format):**
```
time=2025-12-28T10:30:45+02:00 level=INFO msg="User logged in" user_id=123 ip=192.168.1.1
time=2025-12-28T10:30:46+02:00 level=ERROR msg="Failed to connect" error="connection refused"
```

**Output (JSON format):**
```json
{"time":"2025-12-28T10:30:45+02:00","level":"INFO","msg":"User logged in","user_id":"123","ip":"192.168.1.1"}
{"time":"2025-12-28T10:30:46+02:00","level":"ERROR","msg":"Failed to connect","error":"connection refused"}
```

### 2. Context-Aware Logging

Extract logger with context values (request ID, user ID, trace ID):

```go
import (
    "context"
    "github.com/basilex/promenade/pkg/logger"
)

func HandleRequest(ctx context.Context) {
    // Logger automatically includes context fields
    log := logger.FromContext(ctx)
    
    log.Info("Processing request")
    // Output: time=... level=INFO msg="Processing request" request_id=abc123 user_id=456
}
```

### 3. Adding Context Values

```go
// In HTTP middleware
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        
        // Store in context
        ctx := context.WithValue(c.Request.Context(), 
            logger.RequestIDKey, requestID)
        c.Request = c.Request.WithContext(ctx)
        
        c.Next()
    }
}

// Available context keys:
logger.RequestIDKey // "request_id"
logger.UserIDKey    // "user_id"
logger.TraceIDKey   // "trace_id"
```

### 4. Custom Logger Instance

```go
// Create custom logger for specific module
moduleLogger := logger.New(logger.Config{
    Level:  "debug",
    Format: "json",
    AddSource: false,
})

moduleLogger.Debug("Module initialized")
```

### 5. Development vs Production

**Development** (`config/app.postgres-dev.yaml`):
```yaml
logging:
  level: "debug"
  format: "text"
  add_source: true
```

**Production** (`config/app.postgres-prod.yaml`):
```yaml
logging:
  level: "info"
  format: "json"
  add_source: false
```

### 6. Structured Fields

```go
// Use key-value pairs for structured data
logger.Info("Order created",
    "order_id", order.ID,
    "customer_id", order.CustomerID,
    "amount", order.Total,
    "currency", order.Currency,
)

// Group related fields
logger.Info("Payment processed",
    slog.Group("payment",
        "id", payment.ID,
        "method", payment.Method,
        "status", payment.Status,
    ),
    slog.Group("customer",
        "id", customer.ID,
        "email", customer.Email,
    ),
)
```

---

## API Reference

### Initialization Functions

```go
// Init initializes the default logger
func Init(cfg Config)

// New creates a new logger instance
func New(cfg Config) *Logger

// Default returns the default logger
func Default() *Logger
```

### Logging Methods

```go
// Info logs at INFO level
func Info(msg string, args ...any)

// Debug logs at DEBUG level
func Debug(msg string, args ...any)

// Warn logs at WARN level
func Warn(msg string, args ...any)

// Error logs at ERROR level
func Error(msg string, args ...any)
```

### Context Methods

```go
// FromContext extracts logger with context values
func FromContext(ctx context.Context) *Logger

// WithContext returns logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger
```

---

## Best Practices

### DO

- **Use structured logging** - Pass key-value pairs instead of formatting strings
  ```go
  logger.Info("User registered", "email", email, "name", name) // Good
  ```

- **Extract logger from context** - Use `logger.FromContext(ctx)` in handlers
  ```go
  log := logger.FromContext(ctx)
  log.Info("Processing...")
  ```

- **Log errors with context** - Include relevant IDs and state
  ```go
  logger.Error("Failed to create order", 
      "error", err,
      "user_id", userID,
      "cart_items", len(items),
  )
  ```

- **Use appropriate log levels**:
  - **Debug:** Detailed diagnostic information
  - **Info:** General informational messages
  - **Warn:** Warning conditions (not errors)
  - **Error:** Error conditions requiring attention

### DON'T

- **Don't use string formatting** - Avoid `fmt.Sprintf` in log messages
  ```go
  logger.Info(fmt.Sprintf("User %s logged in", email)) // Bad
  logger.Info("User logged in", "email", email)        // Good
  ```

- **Don't log sensitive data** - Never log passwords, tokens, credit cards
  ```go
  logger.Info("Auth attempt", "password", password) // NEVER DO THIS
  logger.Info("Auth attempt", "email", email)       // Good
  ```

- **Don't log in tight loops** - Use debug level or aggregate
  ```go
  for _, item := range millionItems {
      logger.Info("Processing item") // Bad - spam
  }
  
  logger.Info("Processing items", "count", len(millionItems)) // Good
  ```

---

## Testing

The package includes comprehensive tests for all features:

```bash
# Run logger tests
go test ./pkg/logger -v

# With coverage
go test ./pkg/logger -cover
```

**Test Coverage:**
- Configuration validation
- Log level filtering
- Context value extraction
- Output format verification
- Thread safety
- Nil logger handling

---

## Performance

**Benchmarks:**
- **Text format:** ~1.2 μs per log call
- **JSON format:** ~1.5 μs per log call
- **Context extraction:** ~200 ns overhead

**Memory:**
- Zero allocations for simple logs (no interfaces)
- Minimal allocations with structured fields

---

## Related Packages

- `pkg/bus` - Event Bus (uses logger for event tracing)
- `pkg/migration` - Database migrations (uses logger for progress)
- Standard library `log/slog` - Underlying logging implementation

---

## Examples

### Complete HTTP Handler

```go
package main

import (
    "context"
    "github.com/gin-gonic/gin"
    "github.com/basilex/promenade/pkg/logger"
    "github.com/basilex/promenade/pkg/response"
)

func main() {
    // Initialize logger
    logger.Init(logger.Config{
        Level:  "info",
        Format: "json",
    })
    
    r := gin.Default()
    
    // Add request ID middleware
    r.Use(func(c *gin.Context) {
        requestID := generateRequestID()
        ctx := context.WithValue(c.Request.Context(), 
            logger.RequestIDKey, requestID)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    })
    
    r.GET("/users/:id", func(c *gin.Context) {
        log := logger.FromContext(c.Request.Context())
        
        userID := c.Param("id")
        log.Info("Fetching user", "user_id", userID)
        
        user, err := fetchUser(userID)
        if err != nil {
            log.Error("Failed to fetch user", 
                "error", err,
                "user_id", userID,
            )
            response.InternalError(c, "User not found")
            return
        }
        
        log.Info("User fetched successfully", "user_id", userID)
        response.Success(c, user)
    })
    
    r.Run(":8080")
}
```

---

**Last Updated:** 2025-12-28  
**Status:** Production-ready  
**Maintainer:** Promenade Team
