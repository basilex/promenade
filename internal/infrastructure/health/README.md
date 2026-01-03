**Navigation**: [Home](../../../README.md) > [Internal](../../README.md) > [Infrastructure](../README.md) > Health Checks

---

# Health Checks Package

**Comprehensive health monitoring** for all Promenade Platform dependencies with graceful degradation.

---

## Overview

The `health` package provides:

- **4 HTTP Endpoints**: Overall, Database, Redis, Event Bus health checks
- **3 Status Levels**: healthy, degraded, unhealthy
- **5-Second Timeout**: Prevents hanging checks
- **Graceful Degradation**: Optional dependencies handled gracefully
- **Kubernetes-Ready**: Proper liveness/readiness probes

---

## Quick Start

### Initialize Health Checker

```go
import (
    "github.com/basilex/promenade/internal/infrastructure/health"
)

// Initialize health checker
healthChecker := health.NewChecker(db, redisClient, eventBus, cfg.App.Version)

// Create HTTP handler
healthHandler := health.NewHandler(healthChecker)

// Register routes
healthHandler.RegisterRoutes(router)
```

---

## API Endpoints

### Overall Health

**GET /health**

Returns combined health status of all dependencies.

**Response** (200 OK - Healthy):

```json
{
  "status": "healthy",
  "version": "0.1.0",
  "checks": {
    "database": {
      "status": "healthy",
      "message": "PostgreSQL connection healthy"
    },
    "redis": {
      "status": "healthy",
      "message": "Redis connection healthy"
    },
    "event_bus": {
      "status": "healthy",
      "message": "Event bus operational"
    }
  }
}
```

**Response** (200 OK - Degraded):

```json
{
  "status": "degraded",
  "version": "0.1.0",
  "checks": {
    "database": {
      "status": "healthy",
      "message": "PostgreSQL connection healthy"
    },
    "redis": {
      "status": "unhealthy",
      "message": "Redis ping failed: connection refused"
    },
    "event_bus": {
      "status": "healthy",
      "message": "Event bus operational"
    }
  }
}
```

**Response** (503 Service Unavailable - Unhealthy):

```json
{
  "status": "unhealthy",
  "version": "0.1.0",
  "checks": {
    "database": {
      "status": "unhealthy",
      "message": "PostgreSQL ping failed: connection refused"
    }
  }
}
```

---

### Database Health

**GET /health/db**

Returns PostgreSQL database health status.

**Checks**:
1. Connection ping
2. Simple query execution (`SELECT 1`)

**Response** (200 OK):

```json
{
  "status": "healthy",
  "message": "PostgreSQL connection healthy"
}
```

**Response** (503 Service Unavailable):

```json
{
  "status": "unhealthy",
  "message": "Database ping failed: connection refused"
}
```

---

### Redis Health

**GET /health/redis**

Returns Redis health status (graceful if Redis not configured).

**Response** (200 OK):

```json
{
  "status": "healthy",
  "message": "Redis connection healthy"
}
```

**Response** (200 OK - Not Configured):

```json
{
  "status": "healthy",
  "message": "Redis not configured"
}
```

**Response** (503 Service Unavailable):

```json
{
  "status": "unhealthy",
  "message": "Redis ping failed: connection refused"
}
```

---

### Event Bus Health

**GET /health/bus**

Returns Event Bus health status.

**Response** (200 OK):

```json
{
  "status": "healthy",
  "message": "Event bus operational"
}
```

**Response** (503 Service Unavailable):

```json
{
  "status": "unhealthy",
  "message": "Event bus health check failed: adapter unavailable"
}
```

---

## Status Levels

### Healthy

All dependencies are operational.

- **HTTP Status**: 200 OK
- **Response**: `"status": "healthy"`
- **Action**: Normal operation

### Degraded

Some non-critical dependencies are unavailable, but core functionality works.

- **HTTP Status**: 200 OK (still serving traffic)
- **Response**: `"status": "degraded"`
- **Example**: Redis down (token revocation disabled, but auth still works)
- **Action**: Monitor, investigate, but don't panic

### Unhealthy

Critical dependencies (Database, Event Bus) are unavailable.

- **HTTP Status**: 503 Service Unavailable
- **Response**: `"status": "unhealthy"`
- **Action**: Stop routing traffic, alert on-call

---

## Kubernetes Integration

### Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8081
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

**Purpose**: Restart pod if unhealthy

### Readiness Probe

```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8081
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 5
  failureThreshold: 2
```

**Purpose**: Stop routing traffic if unhealthy

---

## Prometheus Integration

### Metrics Endpoint (Planned)

```yaml
# Future: Expose metrics
GET /metrics

# health_check_status{dependency="database"} 1  # 1=healthy, 0=unhealthy
# health_check_status{dependency="redis"} 1
# health_check_status{dependency="event_bus"} 1
# health_check_duration_seconds{dependency="database"} 0.002
```

---

## Usage Examples

### Manual Health Check

```bash
# Overall health
curl http://localhost:8081/health

# Database only
curl http://localhost:8081/health/db

# Redis only
curl http://localhost:8081/health/redis

# Event Bus only
curl http://localhost:8081/health/bus
```

### Programmatic Health Check

```go
// In handler or middleware
checker := health.NewChecker(db, redis, eventBus, version)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

report := checker.CheckAll(ctx)

if report.Status == "unhealthy" {
    log.Error("System unhealthy", slog.Any("checks", report.Checks))
    // Take action: stop processing, alert, etc.
}
```

---

## Timeout Protection

**All checks have 5-second timeout** to prevent hanging:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

report := checker.CheckAll(ctx)
```

**Result**: If check takes longer than 5 seconds, it's marked as unhealthy.

---

## Files

```
internal/infrastructure/health/
 health.go           # Checker logic (CheckAll, CheckDatabase, CheckRedis, CheckEventBus)
 handler.go          # HTTP handlers (4 endpoints)
 health_test.go      # Checker tests (11 tests)
 handler_test.go     # Handler tests (10 tests)
```

---

## Testing

### Run Tests

```bash
# All health check tests
go test ./internal/infrastructure/health -v

# With coverage
go test ./internal/infrastructure/health -cover
```

**Test Coverage**: 21 tests, 100% passing

---

## Troubleshooting

### Database Unhealthy

**Symptom**: `/health/db` returns 503

**Possible Causes**:
- PostgreSQL not running
- Wrong connection string
- Network issues
- Max connections reached

**Debug**:
```bash
# Check PostgreSQL
docker ps | grep postgres

# Test connection
psql -h localhost -U system -d promenade_dev
```

### Redis Unhealthy

**Symptom**: `/health/redis` returns 503

**Impact**: Token revocation disabled (degraded), but auth still works

**Possible Causes**:
- Redis not running
- Wrong connection string
- Network issues

**Debug**:
```bash
# Check Redis
docker ps | grep redis

# Test connection
redis-cli -h localhost -p 6379 ping
```

### Event Bus Unhealthy

**Symptom**: `/health/bus` returns 503

**Possible Causes**:
- Redis adapter configured but Redis unavailable
- Event Bus initialization failed

**Debug**:
```bash
# Check Event Bus config
grep "adapter" config/app.postgres-dev.yaml

# Check Redis if using redis adapter
redis-cli ping
```

---

## Related Documentation

- [Main README](../../../README.md)
- [Documentation Index](../../../docs/INDEX.md)
- [Health Checks Guide](../../../docs/guides/health-checks.md) - Complete implementation guide (680+ lines)
- [Configuration](../config/README.md)
- [Database](../database/README.md)

---

**Status**: Production-ready  
**Tests**: 21 tests, 100% passing  
**Maintainer**: Promenade Team
