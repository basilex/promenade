# Cache Package

**Redis-based caching layer** with graceful degradation and flexible TTL configuration for Promenade Platform.

---

## Features

- **Multiple Adapters**: Redis (production), NoOp (fallback/testing)
- **Factory Pattern**: Auto-selects adapter based on configuration
- **Flexible TTL**: Per-resource type configuration (countries, currencies, users, etc.)
- **Graceful Degradation**: Continues working when Redis unavailable
- **Cache-Aside Pattern**: Read-through cache with write-through invalidation
- **Type-Safe**: JSON marshaling for complex types
- **Context-Aware**: Proper context propagation for cancellation

---

## Quick Start

### 1. Configuration

Add cache section to `config/app.{env}.yaml`:

```yaml
# Development (shorter TTLs)
cache:
  enabled: true
  adapter: "redis"  # or "noop" for testing
  prefix: "promenade:dev:"
  default_ttl: "5m"
  ttl:
    countries: "1h"
    currencies: "1h"
    languages: "1h"
    timezones: "1h"
    user_profile: "15m"
    customer: "10m"
    session: "30m"

# Production (longer TTLs)
cache:
  enabled: true
  adapter: "redis"
  prefix: "promenade:prod:"
  default_ttl: "10m"
  ttl:
    countries: "24h"
    currencies: "24h"
    languages: "24h"
    timezones: "24h"
    user_profile: "30m"
    customer: "20m"
    session: "1h"
```

### 2. Initialize Cache

```go
import (
    "github.com/basilex/promenade/pkg/cache"
    "github.com/redis/go-redis/v9"
)

// Parse config
cacheConfig, err := cfg.Cache.ToCacheConfig()
if err != nil {
    log.Fatal("Failed to parse cache config:", err)
}

// Create Redis client (separate DB for cache)
redisClient := redis.NewClient(&redis.Options{
    Addr:     cfg.Database.Redis.Addr,
    Password: cfg.Database.Redis.Password,
    DB:       cfg.Database.Redis.Databases.Cache, // DB 2 for cache
    PoolSize: 10,
})

// Initialize cache (factory selects Redis or NoOp)
cacheClient, err := cache.NewCache(cacheConfig, redisClient)
if err != nil {
    log.Fatal("Failed to initialize cache:", err)
}
defer cacheClient.Close(context.Background())
```

### 3. Use in UseCase

```go
type useCase struct {
    repo  IRepository
    cache cache.ICache
}

func NewUseCase(repo IRepository, cacheClient cache.ICache) IUseCase {
    return &useCase{
        repo:  repo,
        cache: cacheClient,
    }
}

// Cache TTL constant (1 hour for reference data)
const cacheTTL = 1 * time.Hour

func (uc *useCase) GetByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("entity:id:%s", id.String())
    var entity Entity
    err := uc.cache.Get(ctx, cacheKey, &entity)
    
    if err == nil {
        logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
        return &entity, nil
    }
    
    if !cache.IsCacheMiss(err) {
        logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
    }
    
    // Cache miss - fetch from database
    logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
    entity_, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    if err := uc.cache.Set(ctx, cacheKey, entity_, cacheTTL); err != nil {
        logger.FromContext(ctx).Error("Failed to cache entity", slog.Any("error", err))
    }
    
    return entity_, nil
}
```

---

## Cache Interface

```go
type Cache interface {
    // Get retrieves a value from cache and unmarshals it into dest
    Get(ctx context.Context, key string, dest interface{}) error

    // Set stores a value in cache with the given TTL
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

    // Delete removes a single key from cache
    Delete(ctx context.Context, key string) error

    // DeletePattern removes all keys matching the pattern (e.g., "user:*")
    DeletePattern(ctx context.Context, pattern string) error

    // Clear removes all keys from cache (use with caution)
    Clear(ctx context.Context) error

    // Health checks if the cache is operational
    Health(ctx context.Context) error

    // Close gracefully shuts down the cache
    Close(ctx context.Context) error
}
```

---

## Adapters

### Redis Adapter

**Use Cases**: Production, distributed systems

**Features**:
- JSON marshaling for complex types
- SCAN-based pattern deletion (non-blocking)
- Connection pooling
- Automatic reconnection

**Configuration**:
```yaml
cache:
  enabled: true
  adapter: "redis"
  prefix: "promenade:prod:"
```

### NoOp Adapter

**Use Cases**: Testing, cache disabled, Redis unavailable

**Behavior**:
- All Get operations return cache miss
- All Set/Delete operations succeed silently
- No side effects, no storage

**Configuration**:
```yaml
cache:
  enabled: false
  adapter: "noop"
```

---

## Cache Patterns

### Cache-Aside (Read-Through)

```go
func (uc *useCase) GetByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
    // 1. Try cache
    var entity Entity
    err := uc.cache.Get(ctx, cacheKey, &entity)
    if err == nil {
        return &entity, nil // Cache hit
    }
    
    // 2. Fetch from database (cache miss)
    entity_, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 3. Store in cache
    uc.cache.Set(ctx, cacheKey, entity_, cacheTTL)
    
    return entity_, nil
}
```

### Write-Through Invalidation

```go
func (uc *useCase) Update(ctx context.Context, entity *Entity) error {
    // 1. Update database
    if err := uc.repo.Update(ctx, entity); err != nil {
        return err
    }
    
    // 2. Invalidate cache
    uc.cache.Delete(ctx, fmt.Sprintf("entity:id:%s", entity.ID))
    uc.cache.Delete(ctx, fmt.Sprintf("entity:code:%s", entity.Code))
    uc.cache.Delete(ctx, "entity:list:all")
    
    return nil
}
```

### Pattern-Based Invalidation

```go
func (uc *useCase) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
    // Invalidate all keys matching pattern
    pattern := fmt.Sprintf("user:%s:*", userID.String())
    return uc.cache.DeletePattern(ctx, pattern)
}
```

---

## Key Naming Conventions

**Follow consistent naming patterns** for easy invalidation:

```
{entity}:id:{uuid}          - by ID lookup
{entity}:code:{code}        - by code lookup
{entity}:name:{name}        - by name lookup
{entity}:list:all          - full list
{entity}:list:{filter}     - filtered list
```

**Examples**:
```
country:id:01JGABC-1234-5678-9012-ABCDEF123456
country:code:US
country:list:all

currency:id:01JGABC-5678-1234-9012-FEDCBA654321
currency:code:USD
currency:list:all

user:id:01JGABC-9012-3456-7890-123456789ABC:profile
user:email:user@example.com
```

---

## TTL Strategy

**Reference Data** (rarely changes):
- Countries, Currencies, Languages, Timezones
- Dev: 1 hour
- Prod: 24 hours

**User Data** (changes frequently):
- User profiles, Customers
- Dev: 10-15 minutes
- Prod: 20-30 minutes

**Session Data** (short-lived):
- User sessions, temporary tokens
- Dev: 30 minutes
- Prod: 1 hour

---

## Error Handling

### Cache Miss Error

```go
err := cache.Get(ctx, key, &dest)
if cache.IsCacheMiss(err) {
    // Normal cache miss - fetch from DB
    return fetchFromDB(ctx, key)
}
if err != nil {
    // Redis error - log but continue
    logger.Error("Cache error", slog.Any("error", err))
    return fetchFromDB(ctx, key)
}
```

### Graceful Degradation

```go
// Cache errors are logged but don't break requests
if err := cache.Set(ctx, key, value, ttl); err != nil {
    logger.Error("Failed to cache", slog.Any("error", err))
    // Continue - data still fetched from DB
}
```

---

## Testing

Repository-wide testing strategy and baseline budgets are documented in [docs/guides/testing-patterns.md](../../docs/guides/testing-patterns.md).

### With NoOp Cache

```go
func TestUseCase_GetByID(t *testing.T) {
    mockRepo := new(MockRepository)
    mockRepo.On("GetByID", mock.Anything, id).Return(entity, nil)
    
    // NoOp cache always returns cache miss
    uc := NewUseCase(mockRepo, noop.NewNoOpCache())
    
    result, err := uc.GetByID(ctx, id)
    
    assert.NoError(t, err)
    assert.Equal(t, entity, result)
    mockRepo.AssertExpectations(t) // Should call DB
}
```

### With Real Redis (Integration Tests)

```go
func TestCache_Integration(t *testing.T) {
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   15, // Use separate test DB
    })
    defer redisClient.Close()
    
    cache := redis.NewRedisCache(redisClient, "test:")
    defer cache.Clear(context.Background())
    
    // Test Set/Get
    err := cache.Set(ctx, "key", "value", 1*time.Minute)
    require.NoError(t, err)
    
    var result string
    err = cache.Get(ctx, "key", &result)
    require.NoError(t, err)
    assert.Equal(t, "value", result)
}
```

---

## Redis Database Layout

**Promenade uses separate Redis logical databases** for isolation:

| DB  | Purpose           | Keys Example                |
| --- | ----------------- | --------------------------- |
| 0   | Token Revocation  | `revoked:{token_id}`        |
| 1   | Event Bus         | `bus:{topic}:{event_id}`    |
| 2   | Application Cache | `promenade:prod:country:*`  |
| 3   | User Sessions     | `session:{session_id}`      |

**Key Prefix** (`promenade:dev:` or `promenade:prod:`) prevents conflicts between environments.

---

## Performance

### Redis Adapter

**Throughput**: ~10,000 ops/sec (network-bound)  
**Latency**:
- Get: ~1-5ms (local Redis)
- Set: ~1-5ms (local Redis)
- SCAN: ~10-50ms (depends on keyspace size)

### NoOp Adapter

**Throughput**: Unlimited (in-memory)  
**Latency**: ~0.1μs (no-op)

---

## Best Practices

### DO

- **Use constants for TTL** in usecases (e.g., `const cacheTTL = 1 * time.Hour`)
- **Log cache errors** but don't fail requests
- **Invalidate related keys** on updates (id, code, list)
- **Use context** for cancellation and tracing
- **Test with NoOp cache** for unit tests

### DON'T

- **DON'T cache sensitive data** without encryption
- **DON'T use short keys** (hard to debug)
- **DON'T ignore cache errors** silently
- **DON'T cache everything** (only frequently accessed data)
- **DON'T forget to invalidate** on updates/deletes

---

## Monitoring

### Health Check

```go
// Check cache health
err := cache.Health(ctx)
if err != nil {
    log.Error("Cache unhealthy", err)
}
```

### Metrics (Future)

- Cache hit rate
- Cache miss rate
- Average latency
- Memory usage
- Eviction count

---

## Troubleshooting

### Cache Not Working

1. Check Redis connection: `redis-cli -h localhost -p 6379 ping`
2. Verify config: `cache.enabled = true`
3. Check logs for "Cache initialized"
4. Test health endpoint: `GET /health`

### High Cache Miss Rate

1. Check TTL configuration (not too short)
2. Verify key naming consistency
3. Check if keys are being invalidated too often
4. Monitor Redis memory usage (eviction policy)

### Redis Connection Errors

1. Verify Redis is running: `docker ps | grep redis`
2. Check connection string in config
3. Test connection: `redis-cli -h <host> -p <port> ping`
4. Review firewall rules

---

## Related Documentation

- [Main README](../../README.md)
- [Configuration Guide](../../internal/infrastructure/config/README.md)
- [Shared Context](../../internal/contexts/shared/README.md)
- [Testing Guide](../../test/README.md)

---

**Last Updated**: December 30, 2025  
**Status**: Production-ready  
**Maintainer**: Promenade Team
