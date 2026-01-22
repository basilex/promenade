# ADR-0003: Use Redis for Caching Layer

**Status**: Accepted  
**Date**: 2026-01-15  
**Deciders**: Core Architecture Team  
**Tags**: caching, performance, infrastructure

---

## Context

Promenade requires a caching layer to:

- **Reduce database load**: Reference data (countries, currencies) rarely changes
- **Improve response times**: Cache hot paths (customer lists, product catalog)
- **Session storage**: User sessions, temporary tokens
- **Rate limiting**: Track API request counts per IP/user

**Performance Requirements**:

- Read latency: <5ms (p99)
- Write latency: <10ms (p99)
- High availability: Multi-node support

**Access Patterns**:

- **High read**: Reference data (90% reads, 10% writes)
- **TTL-based**: Sessions expire after 24h
- **Distributed**: Multi-instance API servers (horizontal scaling)

---

## Considered Options

### Option 1: In-Memory (sync.Map / map[string]interface{})

- **Pros**:
  - Zero dependencies
  - Fastest (sub-microsecond latency)
  - No network overhead
- **Cons**:
  -  **Not distributed**: Each API instance has separate cache (cache inconsistency)
  -  **No persistence**: Lost on restart
  -  **Memory limits**: Cannot scale beyond single process
  -  **No eviction**: Manual LRU implementation needed
- **Example**:
  ```go
  cache := sync.Map{}
  cache.Store("country:US", country)
  ```

### Option 2: Memcached

- **Pros**:
  -  Distributed (multi-node support)
  -  Simple protocol (no complexity)
  -  Mature (20+ years)
- **Cons**:
  -  **No persistence**: Lost on restart (same as in-memory)
  -  **Limited data structures**: Key-value only (no lists, sets, sorted sets)
  -  **No pub/sub**: Cannot invalidate cache across nodes
  -  **No atomic operations**: Race conditions possible (incrementing counters)
- **Example**:
  ```go
  mc.Set(&memcache.Item{Key: "country:US", Value: data})
  ```

### Option 3: Redis

- **Pros**:
  -  **Distributed**: Multi-node (Redis Cluster, Sentinel)
  -  **Persistence**: RDB snapshots + AOF log (survive restarts)
  -  **Data structures**: Strings, lists, sets, sorted sets, hashes
  -  **Atomic operations**: `INCR`, `DECR`, `SETNX` (rate limiting)
  -  **Pub/Sub**: Event invalidation (`PUBLISH cache:invalidate country:US`)
  -  **TTL**: Automatic expiration (no manual cleanup)
  -  **Lua scripts**: Complex atomic operations (e.g., sliding window rate limit)
- **Cons**:
  -  Single-threaded (bottleneck at high write volume)
    - Mitigation: Redis 6+ has I/O threads (read scaling)
  -  Memory cost (higher than Memcached for same data)
    - Mitigation: LRU eviction (`maxmemory-policy allkeys-lru`)
- **Example**:
  ```go
  redis.Set(ctx, "country:US", data, 1*time.Hour)
  ```

### Option 4: Hazelcast / Apache Ignite (In-Memory Data Grids)

- **Pros**:
  - Distributed cache with partitioning
  - Near-cache (local cache per node)
- **Cons**:
  -  **Complex**: Requires cluster coordination (Zookeeper/etcd)
  -  **Heavyweight**: JVM (Hazelcast) or heavy Go binary (Ignite)
  -  **Overkill**: Promenade doesn't need distributed computing

---

## Decision

**We will use Redis** as the caching layer for Promenade.

**Rationale**:

1. **Distributed**: Multi-node support (horizontal scaling)
2. **Persistence**: RDB + AOF (survive restarts)
3. **Rich data structures**: Lists, sets, sorted sets (leaderboards, rate limiting)
4. **Atomic operations**: `INCR` for counters (rate limiting, metrics)
5. **Pub/Sub**: Cache invalidation across nodes
6. **Mature ecosystem**: 15+ years, battle-tested (Twitter, GitHub, Stack Overflow)

**Use Cases**:

- **Reference data**: Countries (195), currencies (180), languages (50)
- **Session storage**: User sessions (TTL: 24h)
- **Rate limiting**: API throttling (sliding window algorithm)
- **Hot paths**: Customer lists, product catalog
- **Cache invalidation**: Event-driven (e.g., `customer.updated` → invalidate `customer:123`)

**Implementation**:

```go
// pkg/cache/redis.go
import "github.com/redis/go-redis/v9"

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
    return &RedisCache{
        client: redis.NewClient(&redis.Options{
            Addr: addr,
            DB:   0,
        }),
    }
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
    val, err := c.client.Get(ctx, key).Bytes()
    if errors.Is(err, redis.Nil) {
        return nil, ErrCacheMiss
    }
    return val, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}
```

**Cache-Aside Pattern** (Repository Layer):

```go
// internal/contexts/shared/country/adapter/repository/postgres/country_repository.go
func (r *CountryRepository) FindByCode(ctx context.Context, code string) (*aggregate.Country, error) {
    // 1. Check cache
    cacheKey := fmt.Sprintf("country:%s", code)
    if data, err := r.cache.Get(ctx, cacheKey); err == nil {
        var country aggregate.Country
        json.Unmarshal(data, &country)
        return &country, nil  // Cache hit
    }

    // 2. Query database (cache miss)
    country, err := r.findByCodeDB(ctx, code)
    if err != nil {
        return nil, err
    }

    // 3. Write to cache (TTL: 1 hour)
    data, _ := json.Marshal(country)
    r.cache.Set(ctx, cacheKey, data, 1*time.Hour)

    return country, nil
}
```

**Rate Limiting** (Middleware):

```go
// pkg/middleware/ratelimit.go
func RateLimitMiddleware(redis *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := fmt.Sprintf("ratelimit:%s", c.ClientIP())

        // Sliding window: 100 requests per minute
        count, _ := redis.Incr(c.Request.Context(), key).Result()
        if count == 1 {
            redis.Expire(c.Request.Context(), key, 1*time.Minute)
        }

        if count > 100 {
            c.JSON(429, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## Consequences

### Positive

-  **Performance**: <5ms read latency (p99), 10-100x faster than database
-  **Scalability**: Horizontal scaling (Redis Cluster)
-  **Flexibility**: Rich data structures (lists, sets, sorted sets)
-  **Reliability**: Persistence (RDB + AOF) survives restarts
-  **Atomicity**: `INCR`, `DECR`, `SETNX` for race-free counters
-  **Pub/Sub**: Event-driven cache invalidation

### Negative

-  **Single point of failure**: Redis crash → cache miss storm
  - Mitigation 1: Redis Sentinel (automatic failover)
  - Mitigation 2: Circuit breaker (fallback to database on Redis failure)
-  **Memory cost**: ~1GB RAM for 1M cached entities
  - Mitigation: LRU eviction (`maxmemory 2gb`, `maxmemory-policy allkeys-lru`)
-  **Operational complexity**: Monitoring, backups, upgrades
  - Mitigation: Docker Compose (dev), managed Redis (prod: AWS ElastiCache, Azure Cache)

### Neutral

- ℹ Requires Redis server (Docker Compose in dev, managed service in prod)
- ℹ Cache invalidation must be explicit (no magic)
- ℹ Cache consistency requires event-driven architecture (Event Bus integration)

---

## Implementation Notes

**Rollout**:

- **Phase 1**  (2026-01-15): `pkg/cache` package created
- **Phase 2**  (2026-01-16): Reference data cached (countries, currencies)
- **Phase 3**  (2026-01-17): Rate limiting middleware
- **Phase 4** ⏳ (Phase 2): Event-driven cache invalidation (Event Bus integration)

**Configuration**:

```yaml
# config/app.postgres-dev.yaml
redis:
  addr: localhost:6379
  db: 0
  password: ""
  max_retries: 3
  pool_size: 10
```

**Docker Compose**:

```yaml
# docker/docker-compose.postgres.dev.yml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
```

**Rollback Plan**:

- Remove Redis from repositories (fallback to database)
- Estimated effort: 1-2 days (remove cache calls)

---

## References

- [Redis Documentation](https://redis.io/docs/)
- [Cache-Aside Pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/cache-aside)
- [Redis Pub/Sub](https://redis.io/docs/interact/pubsub/)
- [Rate Limiting with Redis](https://redis.io/glossary/rate-limiting/)
- [pkg/cache/README.md](../../pkg/cache/README.md) - Implementation details

---

## Revisit Criteria

**Reconsider this decision if**:

- Redis becomes a bottleneck (>80% CPU usage, >10ms p99 latency)
- Cache hit rate falls below 70% (cache thrashing)
- Memory cost exceeds budget (>5GB for cache)
- KeyDB or Dragonfly offers significantly better performance (>2x faster)
- Team prefers Memcached for simplicity (no rich data structures needed)

---

## Changelog

| Date       | Change                                              | Author              |
| ---------- | --------------------------------------------------- | ------------------- |
| 2026-01-15 | Initial version                                     | Alexander Vasilenko |
| 2026-01-22 | Added rate limiting example and performance metrics | Alexander Vasilenko |
