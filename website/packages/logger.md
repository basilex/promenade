# Logger Package

**Structured logging with context support and slog wrapper**

---

## Overview

The `logger` package provides a thin wrapper around Go's standard `log/slog` with additional convenience methods for **context-aware logging**. It supports multiple output formats, log levels, and automatic context field extraction.

**Status**: Production-ready  
**Tests**: 15 tests, 95% coverage  
**Location**: `pkg/logger/`

---

## Features

- **Structured Logging** - JSON or text format output
- **Context Integration** - Automatic extraction of request ID, user ID, trace ID
- **Customizable Levels** - Debug, Info, Warn, Error
- **Source Location** - Optional file/line information
- **Time Formatting** - Configurable timestamp format
- **Thread-Safe** - Safe for concurrent use
- **Production-Ready** - Used in all Promenade services

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
        Level:      "info",
        Format:     "text",
        AddSource:  true,
        TimeFormat: time.RFC3339,
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

### Basic Logging

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

**Output (text format)**:
```
time=2025-12-29T10:30:45+02:00 level=INFO msg="User logged in" user_id=123 ip=192.168.1.1
time=2025-12-29T10:30:46+02:00 level=ERROR msg="Failed to connect" error="connection refused"
```

**Output (JSON format)**:
```json
{"time":"2025-12-29T10:30:45+02:00","level":"INFO","msg":"User logged in","user_id":"123","ip":"192.168.1.1"}
{"time":"2025-12-29T10:30:46+02:00","level":"ERROR","msg":"Failed to connect","error":"connection refused"}
```

### Context-Aware Logging

```go
// Extract logger from context
func (h *UserHandler) Create(c *gin.Context) {
    ctx := c.Request.Context()
    log := logger.FromContext(ctx)
    
    log.Info("Creating user", 
        slog.String("email", req.Email),
        slog.String("ip", c.ClientIP()),
    )
    
    user, err := h.usecase.CreateUser(ctx, req.Email, req.Password)
    if err != nil {
        log.Error("Failed to create user", 
            slog.Any("error", err),
            slog.String("email", req.Email),
        )
        return
    }
    
    log.Info("User created", 
        slog.String("user_id", user.ID.String()),
    )
}
```

### With Context Fields

```go
// Add context fields (request ID, user ID, etc.)
ctx = logger.WithField(ctx, "request_id", requestID)
ctx = logger.WithField(ctx, "user_id", userID)

// Logger automatically includes these fields
logger.FromContext(ctx).Info("Processing request")
// Output: time=... level=INFO msg="Processing request" request_id=abc123 user_id=01JGABC...
```

---

## API Reference

### Initialization

#### Init

```go
func Init(config Config)
```

Initializes the global logger with configuration.

**Example**:
```go
logger.Init(logger.Config{
    Level:  "debug",
    Format: "json",
})
```

### Logging Functions

#### Debug

```go
func Debug(msg string, args ...any)
```

Logs at DEBUG level (development only).

#### Info

```go
func Info(msg string, args ...any)
```

Logs at INFO level (general information).

#### Warn

```go
func Warn(msg string, args ...any)
```

Logs at WARN level (potential issues).

#### Error

```go
func Error(msg string, args ...any)
```

Logs at ERROR level (errors that need attention).

#### Fatal

```go
func Fatal(msg string, args ...any)
```

Logs at ERROR level and calls `os.Exit(1)`.

### Context Functions

#### FromContext

```go
func FromContext(ctx context.Context) *slog.Logger
```

Extracts logger from context with automatic field injection.

**Example**:
```go
log := logger.FromContext(ctx)
log.Info("User action", slog.String("action", "login"))
```

#### WithField

```go
func WithField(ctx context.Context, key string, value any) context.Context
```

Adds field to context for automatic inclusion in logs.

**Example**:
```go
ctx = logger.WithField(ctx, "request_id", "abc123")
logger.FromContext(ctx).Info("Request received") // Includes request_id
```

---

## Configuration Examples

### Development (Text Format)

```yaml
# config/app.postgres-dev.yaml or config/app.sqlite-dev.yaml
logging:
  level: "debug"
  format: "text"
  add_source: true
```

```go
logger.Init(logger.Config{
    Level:     cfg.Logging.Level,
    Format:    cfg.Logging.Format,
    AddSource: cfg.Logging.AddSource,
})
```

### Production (JSON Format)

```yaml
# config/app.postgres-prod.yaml
logging:
  level: "info"
  format: "json"
  add_source: false
```

```go
logger.Init(logger.Config{
    Level:  "info",
    Format: "json",
})
```

---

## Best Practices

### DO

✅ **Use structured fields**:
```go
logger.Info("User created", 
    "user_id", userID.String(),
    "email", email,
)
```

✅ **Use context logger in handlers**:
```go
log := logger.FromContext(ctx)
log.Info("Processing request")
```

✅ **Log errors with context**:
```go
if err != nil {
    logger.Error("Database error", 
        "error", err,
        "query", queryName,
    )
}
```

✅ **Use appropriate log levels**:
- DEBUG: Development debugging
- INFO: Normal operations
- WARN: Potential issues
- ERROR: Actual errors

### DON'T

❌ **Don't use string concatenation**:
```go
// Bad
logger.Info("User " + userID + " created")

// Good
logger.Info("User created", "user_id", userID)
```

❌ **Don't log sensitive data**:
```go
// Bad
logger.Info("Login attempt", "password", password)

// Good
logger.Info("Login attempt", "email", email)
```

❌ **Don't use global logger in domain logic**:
```go
// Bad (in usecase)
logger.Info("Creating user")

// Good (pass context)
logger.FromContext(ctx).Info("Creating user")
```

---

## Testing

### Running Tests

```bash
# Run logger tests
go test ./pkg/logger -v

# With coverage
go test ./pkg/logger -cover
```

### Test Example

```go
func TestLogger_Info(t *testing.T) {
    var buf bytes.Buffer
    
    logger.Init(logger.Config{
        Level:  "info",
        Format: "json",
        Output: &buf,
    })
    
    logger.Info("test message", "key", "value")
    
    output := buf.String()
    assert.Contains(t, output, "test message")
    assert.Contains(t, output, "key")
}
```

---

## Integration

### Gin Middleware

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/basilex/promenade/pkg/logger"
)

func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        requestID := uuid.NewString()
        
        // Add request ID to context
        ctx := logger.WithField(c.Request.Context(), "request_id", requestID)
        c.Request = c.Request.WithContext(ctx)
        
        // Process request
        c.Next()
        
        // Log request completion
        logger.FromContext(ctx).Info("Request completed",
            slog.String("method", c.Request.Method),
            slog.String("path", c.Request.URL.Path),
            slog.Int("status", c.Writer.Status()),
            slog.Duration("duration", time.Since(start)),
        )
    }
}
```

---

## Related Documentation

- [Quick Start Guide](/guide/quick-start) - Get started with Promenade
- [Testing Guide](/guide/testing) - Testing patterns
- [Contributing Guide](/guide/contributing) - Development workflow
- [GitHub Repository](https://github.com/basilex/promenade/tree/dev/pkg/logger) - Source code
