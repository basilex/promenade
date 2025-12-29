# Configuration Package

**YAML-based configuration management** with environment variable overrides for Promenade Platform.

---

## Overview

The `config` package provides **structured configuration loading** from YAML files with:

- **Environment-specific configs**: dev, test, prod
- **Environment variable overrides**: Sensitive values (DB_PASSWORD, JWT_SECRET)
- **Validation**: Startup validation of critical settings
- **Type-safe**: Strongly-typed configuration structs

---

## Quick Start

### Load Configuration

```go
import "github.com/basilex/promenade/internal/infrastructure/config"

// Load config based on ENVIRONMENT variable
cfg, err := config.Load()
if err != nil {
    log.Fatal("Failed to load config:", err)
}

// Validate configuration
if err := cfg.Validate(); err != nil {
    log.Fatal("Invalid configuration:", err)
}

// Use configuration
fmt.Println("Server port:", cfg.Server.Port)
fmt.Println("Database host:", cfg.Database.Postgres.Host)
```

---

## Configuration Files

### Location

```
config/
 app.dev.yaml      # Development (Memory bus, localhost DB)
 app.test.yaml     # Testing (test DB on port 5433)
 app.prod.yaml     # Production (Redis bus, env vars)
```

### Selection

**Environment variable** `ENVIRONMENT` determines which file to load:

```bash
ENVIRONMENT=dev ./promenade        # Loads app.dev.yaml
ENVIRONMENT=test ./promenade       # Loads app.test.yaml
ENVIRONMENT=production ./promenade # Loads app.prod.yaml
```

**Default**: `dev` if not specified

---

## Configuration Structure

### AppConfig

```go
type AppConfig struct {
    App      AppSection      // Application metadata
    Server   ServerSection   // HTTP server settings
    Database DatabaseSection // PostgreSQL + Redis
    Bus      BusSection      // Event Bus configuration
    JWT      JWTSection      // JWT authentication
    Logging  LoggingSection  // Logging configuration
}
```

### App Section

```yaml
app:
  name: "Promenade Platform"
  environment: "development"  # dev, test, production
  version: "0.1.0"
```

### Server Section

```yaml
server:
  host: "0.0.0.0"
  port: 8081
  read_timeout: 10s
  write_timeout: 10s
```

### Database Section

```yaml
database:
  postgres:
    host: "localhost"      # Override with DB_HOST
    port: 5432
    user: "system"         # Override with DB_USER
    password: "passw0rd"   # Override with DB_PASSWORD
    database: "promenade_dev"
    ssl_mode: "disable"
  redis:
    addr: "localhost:6379" # Override with REDIS_ADDR
    password: ""           # Override with REDIS_PASSWORD
    pool_size: 10
    max_retries: 3
    databases:
      revocation: 0  # JWT token revocation
      bus: 1         # Event Bus (Redis adapter)
      cache: 2       # Application cache
      sessions: 3    # User sessions
```

### Event Bus Section

```yaml
bus:
  adapter: "memory"        # memory (dev) or redis (prod)
  worker_pool_size: 10
  buffer_size: 1000
  retry_attempts: 3
  retry_delay: 1s
  retry_max_delay: 30s
  retry_multiplier: 2.0
```

### JWT Section

```yaml
jwt:
  secret: "dev-secret-key-change-in-production"  # Override with JWT_SECRET
  access_token_duration: 15m   # 15 minutes
  refresh_token_duration: 168h # 7 days
  issuer: "promenade-platform"
```

### Logging Section

```yaml
logging:
  level: "debug"       # debug, info, warn, error
  format: "text"       # text or json
  add_source: true     # Include source file:line in logs
```

---

## Environment Variable Overrides

**Sensitive values** should use environment variables in production:

```yaml
# config/app.prod.yaml
database:
  postgres:
    host: "${DB_HOST}"           # Required
    user: "${DB_USER}"           # Required
    password: "${DB_PASSWORD}"   # Required
  redis:
    addr: "${REDIS_ADDR:-localhost:6379}"  # Optional with default
    password: "${REDIS_PASSWORD}"          # Optional

jwt:
  secret: "${JWT_SECRET}"        # Required in production
```

**Syntax**:
- `${VAR}` - Required, fails if not set
- `${VAR:-default}` - Optional, uses default if not set

---

## Validation

### Startup Validation

**Configuration is validated on startup** before any initialization:

```go
// cmd/api/main.go
cfg, err := config.Load()
if err != nil {
    log.Fatal("Failed to load config:", err)
}

// Validate before using
if err := cfg.Validate(); err != nil {
    log.Fatal("Invalid configuration:", err)
}
```

### Validation Rules

**JWT Secret**:
- **Production**: Minimum 32 characters, blocks default dev secret
- **Development**: No restrictions (allows short secrets for convenience)
- **All environments**: Empty secret blocked

**Database**:
- PostgreSQL host, port, user, database required
- Redis optional (graceful degradation if not available)

**Server**:
- Port must be valid (1-65535)
- Timeouts must be positive

---

## Usage Examples

### Access Configuration

```go
// In main.go
cfg, _ := config.Load()

// Initialize database
db, err := database.NewPostgresConnection(&cfg.Database.Postgres)

// Initialize JWT manager
jwtManager := jwt.NewManager(jwt.Config{
    SecretKey:            cfg.JWT.Secret,
    AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
    RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
    Issuer:               cfg.JWT.Issuer,
})

// Initialize Event Bus
eventBus, err := bus.NewBus(cfg.Bus, cfg.Database.Redis)
```

### Testing Configuration

```go
// In tests
func TestWithCustomConfig(t *testing.T) {
    // Set environment for test config
    os.Setenv("ENVIRONMENT", "test")
    defer os.Unsetenv("ENVIRONMENT")
    
    cfg, err := config.Load()
    require.NoError(t, err)
    
    assert.Equal(t, 5433, cfg.Database.Postgres.Port) // Test DB port
}
```

---

## Files

```
internal/infrastructure/config/
 yaml_config.go       # Configuration structs and loading
 yaml_config_test.go  # Configuration validation tests
```

---

## Related Documentation

- [Main README](../../../README.md)
- [Documentation Index](../../../docs/INDEX.md)
- [Database Package](../database/README.md)
- [Health Checks](../health/README.md)

---

**Status**: Production-ready  
**Tests**: 12 validation tests  
**Maintainer**: Promenade Team
