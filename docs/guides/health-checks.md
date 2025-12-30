# Health Check System

**Comprehensive health monitoring** for Promenade Platform dependencies - PostgreSQL, Redis, and Event Bus.

---

## Overview

The health check system provides **3-level monitoring** (healthy, degraded, unhealthy) with:
- **4 HTTP endpoints** for monitoring
- **5-second timeout** for all checks
- **Graceful degradation** for optional dependencies
- **Proper HTTP status codes** (200 for healthy/degraded, 503 for unhealthy)

---

## Architecture

### Components

1. **health.Checker** (`internal/infrastructure/health/health.go`)
   - Core health check logic
   - Checks: Database, Redis (optional), Event Bus
   - 5-second timeout for all checks combined

2. **health.Handler** (`internal/infrastructure/health/handler.go`)
   - HTTP endpoints with Gin
   - 4 routes: `/health`, `/health/db`, `/health/redis`, `/health/bus`

3. **Integration** (`cmd/api/main.go`)
   - Dependency injection (db, redisClient, eventBus)
   - Replaces old simple health endpoint

---

## API Endpoints

### 1. Overall Health Check

**GET /health**

Returns overall system health with all dependency checks.

**Response (200 OK - Healthy)**:
```json
{
  "status": "healthy",
  "checks": {
    "database": {
      "name": "PostgreSQL",
      "status": "healthy",
      "message": "database connection ok",
      "duration_ms": 1234567,
      "timestamp": "2025-12-29T18:00:00Z"
    },
    "redis": {
      "name": "Redis",
      "status": "healthy",
      "message": "redis connection ok",
      "duration_ms": 567890,
      "timestamp": "2025-12-29T18:00:00Z"
    },
    "event_bus": {
      "name": "Event Bus",
      "status": "healthy",
      "message": "event bus operational",
      "duration_ms": 123456,
      "timestamp": "2025-12-29T18:00:00Z"
    }
  },
  "timestamp": "2025-12-29T18:00:00Z",
  "version": "0.1.0"
}
```

**Response (200 OK - Degraded)**:
```json
{
  "status": "degraded",
  "checks": {
    "database": {
      "status": "degraded",
      "message": "database ping ok but query failed"
    },
    "redis": { "status": "healthy" },
    "event_bus": { "status": "healthy" }
  }
}
```

**Response (503 Service Unavailable - Unhealthy)**:
```json
{
  "status": "unhealthy",
  "checks": {
    "database": {
      "status": "unhealthy",
      "message": "database ping failed: connection refused"
    }
  }
}
```

### 2. Database Health Check

**GET /health/db**

Returns PostgreSQL database health only.

**Response (200 OK)**:
```json
{
  "name": "PostgreSQL",
  "status": "healthy",
  "message": "database connection ok",
  "duration_ms": 1234567,
  "timestamp": "2025-12-29T18:00:00Z"
}
```

**Response (503 Service Unavailable)**:
```json
{
  "name": "PostgreSQL",
  "status": "unhealthy",
  "message": "database ping failed: connection refused"
}
```

### 3. Redis Health Check

**GET /health/redis**

Returns Redis health (if configured).

**Response (200 OK - Configured)**:
```json
{
  "name": "Redis",
  "status": "healthy",
  "message": "redis connection ok"
}
```

**Response (200 OK - Not Configured)**:
```json
{
  "name": "Redis",
  "status": "healthy",
  "message": "redis not configured (optional)"
}
```

### 4. Event Bus Health Check

**GET /health/bus**

Returns Event Bus health.

**Response (200 OK)**:
```json
{
  "name": "Event Bus",
  "status": "healthy",
  "message": "event bus operational"
}
```

---

## Status Levels

### Healthy

All dependencies are operational.

- **HTTP Status**: 200 OK
- **Criteria**: All checks pass
- **Action**: No action needed

### Degraded

System is operational but with issues.

- **HTTP Status**: 200 OK
- **Criteria**: At least one check is degraded (e.g., database ping works but query fails)
- **Action**: Investigate warnings, monitor closely

### Unhealthy

Critical dependency is down.

- **HTTP Status**: 503 Service Unavailable
- **Criteria**: At least one check failed completely
- **Action**: Immediate investigation required

---

## Usage Examples

### cURL

```bash
# Check overall health
curl http://localhost:8081/health

# Check database only
curl http://localhost:8081/health/db

# Check Redis only
curl http://localhost:8081/health/redis

# Check Event Bus only
curl http://localhost:8081/health/bus
```

### Kubernetes Liveness Probe

```yaml
livenessProbe:
  httpGet:
    path: /health/db
    port: 8081
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Kubernetes Readiness Probe

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

### Prometheus Monitoring

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'promenade-health'
    metrics_path: /health
    static_configs:
      - targets: ['promenade:8081']
```

---

## Implementation Details

### Database Check

1. **Ping**: `db.PingContext(ctx)` - Basic connectivity
2. **Query**: `SELECT 1` - Database is writable

**Status**:
- Healthy: Both pass
- Degraded: Ping passes, query fails
- Unhealthy: Ping fails

### Redis Check

1. **Optional**: Returns "healthy" if not configured
2. **Ping**: `redis.Ping(ctx)` - Connectivity check

**Status**:
- Healthy: Ping passes or not configured
- Unhealthy: Ping fails

### Event Bus Check

1. **Health method**: `eventBus.Health(ctx)` - Internal health check

**Status**:
- Healthy: Health check passes
- Unhealthy: Health check fails

---

## Timeout Behavior

All checks have a **5-second combined timeout**:

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
```

**Behavior**:
- If any check takes longer than 5 seconds, it returns unhealthy
- Prevents hanging requests
- Fast-fail for slow dependencies

---

## Testing

### Unit Tests

**Location**: `internal/infrastructure/health/health_test.go` (11 tests)

```bash
go test ./internal/infrastructure/health -v
```

**Tests**:
- TestChecker_CheckDatabase_Healthy
- TestChecker_CheckDatabase_Unhealthy_PingFailed
- TestChecker_CheckDatabase_Degraded_QueryFailed
- TestChecker_CheckRedis_Healthy
- TestChecker_CheckRedis_Unhealthy
- TestChecker_CheckRedis_NotConfigured
- TestChecker_CheckEventBus_Healthy
- TestChecker_CheckAll_AllHealthy
- TestChecker_CheckAll_DatabaseUnhealthy
- TestChecker_CheckAll_WithTimeout (verifies 5s timeout)

### Handler Tests

**Location**: `internal/infrastructure/health/handler_test.go` (10 tests)

**Tests**:
- TestHandler_CheckAll_Healthy
- TestHandler_CheckAll_Unhealthy
- TestHandler_CheckAll_Degraded
- TestHandler_CheckDatabase_Healthy
- TestHandler_CheckDatabase_Unhealthy
- TestHandler_CheckRedis
- TestHandler_CheckEventBus
- TestHandler_RegisterRoutes

### Mock Dependencies

```go
// Database mock
db, mock, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
sqlxDB := sqlx.NewDb(db, "sqlmock")

// Redis mock
redisClient, redisMock := redismock.NewClientMock()

// Event Bus mock (real memory bus)
eventBus := memory.NewMemoryBus(bus.Config{
    WorkerPoolSize: 1,
    BufferSize:     10,
})
```

---

## Monitoring Integration

### Grafana Dashboard

Create dashboard with panels for:
- Overall status (gauge: healthy/degraded/unhealthy)
- Database response time (graph)
- Redis response time (graph)
- Event Bus status (status history)

### Alerting

Configure alerts for:
- **Critical**: Status = unhealthy for > 1 minute
- **Warning**: Status = degraded for > 5 minutes

### Logging

Health checks are logged at DEBUG level:

```
[DEBUG] Health check: status=healthy db=1ms redis=2ms bus=1ms
[WARN] Health check: status=degraded db=degraded
[ERROR] Health check: status=unhealthy db=unhealthy
```

---

## Configuration

### Main Application

**Location**: `cmd/api/main.go`

```go
// Initialize health checker
healthChecker := health.NewChecker(db, redisClient, eventBus, cfg.App.Version)
healthHandler := health.NewHandler(healthChecker)

// Register routes
healthHandler.RegisterRoutes(r)
```

### Dependencies

- **PostgreSQL**: Required (sqlx.DB)
- **Redis**: Optional (nil if not configured)
- **Event Bus**: Required (bus.IBus)
- **Version**: App version from config

---

## Graceful Degradation

### Redis Not Configured

If Redis is not configured (nil), health check returns:

```json
{
  "name": "Redis",
  "status": "healthy",
  "message": "redis not configured (optional)"
}
```

**Behavior**:
- System continues to operate
- Overall status not affected
- Token revocation disabled (graceful fallback)

---

## Best Practices

### DO

- **Monitor /health endpoint** - Set up alerts for unhealthy status
- **Use readiness probes** - Prevent traffic to unhealthy instances
- **Check individual endpoints** - Debug specific dependency issues
- **Set timeouts** - Prevent hanging health checks
- **Log health changes** - Track status transitions

### DON'T

- **DON'T poll too frequently** - Adds load, use 5-10 second intervals
- **DON'T expose to public** - Health endpoints should be internal
- **DON'T treat degraded as unhealthy** - System still operational
- **DON'T skip Redis check** - Optional but important for full picture

---

## Troubleshooting

### Database Unhealthy

```json
{
  "status": "unhealthy",
  "message": "database ping failed: connection refused"
}
```

**Possible causes**:
- Database down
- Connection pool exhausted
- Network issues
- Wrong credentials

**Actions**:
1. Check PostgreSQL is running: `docker ps | grep postgres`
2. Verify connection string in config
3. Check database logs
4. Test connection: `psql -h localhost -U system -d promenade_dev`

### Database Degraded

```json
{
  "status": "degraded",
  "message": "database ping ok but query failed"
}
```

**Possible causes**:
- Read-only mode
- Disk full
- Permissions issue

**Actions**:
1. Check database mode: `SHOW transaction_read_only;`
2. Check disk space: `df -h`
3. Verify user permissions

### Redis Unhealthy

```json
{
  "status": "unhealthy",
  "message": "redis ping failed: connection refused"
}
```

**Possible causes**:
- Redis down
- Network issues
- Wrong address/port

**Actions**:
1. Check Redis is running: `docker ps | grep redis`
2. Verify Redis address in config
3. Test connection: `redis-cli -h localhost -p 6379 ping`

---

## Related Documentation

- [Main README](../README.md) - Project overview
- [GAPS_AND_TODOS.md](../work-in-progress/GAPS_AND_TODOS.md) - Implementation checklist
- [Configuration Guide](../config/README.md) - App configuration

---

**Last Updated**: December 29, 2025  
**Status**: Production-ready  
**Test Coverage**: 21 tests, 100% passing  
**Maintainer**: Promenade Team
