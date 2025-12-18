# Configuration Changes

## Overview

Refactored email templates and event bus configuration from hardcoded values to externalized resources.

## Changes Made

### 1. Email Templates Externalized

**Before**: Templates hardcoded in `email_service.go`
**After**: HTML templates in `templates/email/` directory

```go
// Old
email := Email{
    Body: fmt.Sprintf(`Hi %s, Welcome...`, evt.Name),
    HTML: fmt.Sprintf(`<html>...</html>`, evt.Name),
}

// New
html, err := s.renderTemplate("welcome.html", data)
email := Email{
    Subject: "Welcome to Promenade!",
    HTML:    html,
}
```

**Benefits**:

- [+] Change templates without redeploying service
- [+] Professional HTML emails with CSS styling
- [+] Easy A/B testing of email variants
- [+] Designer-friendly (no Go code needed)

### 2. Event Bus Configuration from Environment

**Before**: Hardcoded in `pkg/bus/bus.go`

```go
func DefaultBusConfig() BusConfig {
    return BusConfig{
        WorkerPoolSize: 10,
        BufferSize:     1000,
        ...
    }
}
```

**After**: Loaded from `.env` files

```bash
# .env.development
BUS_WORKER_POOL_SIZE=10
BUS_BUFFER_SIZE=1000
BUS_RETRY_ATTEMPTS=3
BUS_RETRY_DELAY=1s
```

```go
// config/config.go
type BusConfig struct {
    WorkerPoolSize int
    BufferSize     int
    RetryAttempts  int
    RetryDelay     time.Duration
}

// main.go
busConfig := bus.NewBusConfig(
    cfg.Bus.WorkerPoolSize,
    cfg.Bus.BufferSize,
    cfg.Bus.RetryAttempts,
    cfg.Bus.RetryDelay,
)
eventBus := memory.NewMemoryBus(busConfig)
```

**Benefits**:

- [+] Different configs per environment (dev/staging/prod)
- [+] Tune performance without code changes
- [+] Scale worker pool for high-load scenarios
- [+] Adjust retry policies dynamically

## Files Modified

### Email Templates System

- Created: `templates/email/*.html` (5 templates)
- Modified: `internal/infrastructure/notification/email_service.go`
  - Added `renderTemplate()` method
  - Added `renderFallbackTemplate()` for tests
  - All handlers now use templates
- Modified: `cmd/api/main.go` - pass templates path
- Modified: `examples/event_bus_demo/main.go` - use templates path
- Modified: `test/integration/event_bus_test.go` - empty path for tests

### Event Bus Configuration

- Modified: `internal/infrastructure/config/config.go`
  - Added `BusConfig` struct
  - Load from environment variables
- Modified: `.env.development`, `.env.example`
  - Added 4 bus configuration variables
- Modified: `pkg/bus/bus.go`
  - Changed `DefaultBusConfig()` → `NewBusConfig(params...)`
- Modified: `pkg/bus/memory/memory_bus.go`
  - Deprecated `NewDefaultMemoryBus()` (fallback for tests)
- Modified: `cmd/api/main.go`
  - Read config from environment
  - Create bus with configuration

## Testing

All tests pass:

```bash
make test  # [+] PASS (unit + integration)
```

Test strategy:

- **Production**: Uses real template files from `templates/email/`
- **Tests**: Uses empty path `""` → fallback templates (no file I/O)
- **Demo**: Uses real templates to demonstrate functionality

## Production Deployment

### Docker

```dockerfile
# Include templates in image
COPY templates/ /app/templates/

# Or mount as volume
volumes:
  - ./templates:/app/templates:ro
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: email-templates
data:
  welcome.html: |
    <!DOCTYPE html>
    ...
```

### Environment Variables

```bash
# Tune for production load
BUS_WORKER_POOL_SIZE=50
BUS_BUFFER_SIZE=10000
BUS_RETRY_ATTEMPTS=5
BUS_RETRY_DELAY=2s
```

## Migration Notes

### For Existing Deployments

1. **Add environment variables** to `.env.production`:

   ```bash
   BUS_WORKER_POOL_SIZE=20
   BUS_BUFFER_SIZE=5000
   BUS_RETRY_ATTEMPTS=3
   BUS_RETRY_DELAY=1s
   ```

2. **Include templates** in deployment pipeline
3. **Update tests** if they instantiate email service directly:

   ```go
   // Old
   emailService := notification.NewEmailService(bus, sender)

   // New
   emailService, err := notification.NewEmailService(bus, sender, "templates/email")
   // Or for tests:
   emailService, err := notification.NewEmailService(bus, sender, "")
   ```

### Backward Compatibility

- [+] Tests work without template files (fallback templates)
- [+] `NewDefaultMemoryBus()` still exists (deprecated, for tests)
- [!] `NewEmailService` signature changed (3rd parameter added)

## Future Enhancements

1. **Template hot-reloading** without service restart
2. **Template versioning** (A/B testing)
3. **Internationalization** (i18n) support
4. **Template preview** endpoint for testing
5. **Email template editor** in admin panel
