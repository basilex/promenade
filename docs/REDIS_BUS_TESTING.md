# Redis Bus Testing Guide

This guide describes how to test the Redis adapter for the event bus.

## Quick Start

### 1. Launch Redis

```bash
# Start Redis via docker-compose
docker-compose -f docker/docker-compose.yml up -d redis

# Verify Redis is running
docker ps | grep redis
```

### 2. Integration Tests

```bash
# Run all Redis bus tests
go test -v ./test/integration/redis_bus_test.go -count=1

# Run specific test
go test -v ./test/integration/redis_bus_test.go -run TestRedisBus_PublishSubscribe
```

**Available tests:**

- `TestRedisBus_PublishSubscribe` - basic pub/sub
- `TestRedisBus_MultipleSubscribers` - multiple subscribers to one topic
- `TestRedisBus_GracefulShutdown` - graceful shutdown
- `TestBusFallback_RedisToMemory` - fallback to memory when Redis unavailable

### 3. Demo Application

```bash
# Test with memory adapter
BUS_ADAPTER=memory go run examples/redis_bus_demo/main.go

# Test with Redis adapter
BUS_ADAPTER=redis REDIS_HOST=localhost REDIS_PORT=6379 \
  go run examples/redis_bus_demo/main.go
```

**Expected result:**

```
[OK] Bus health check passed
[OK] Subscribed to topic
[>>] Publishing test events count=5
[OK] Published event num=1
[<<] Received event type=test.message count=1
...
[RESULTS] Test Results published=5 received=5
[SUCCESS] All messages received!
[OK] Bus closed gracefully
```

## Switching Adapters in Production

### Environment Variables

```bash
# Memory adapter (development)
BUS_ADAPTER=memory
BUS_WORKER_POOL_SIZE=4
BUS_BUFFER_SIZE=100

# Redis adapter (production)
BUS_ADAPTER=redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
```

### Graceful Degradation

In **production** mode, if Redis is unavailable, the system will automatically switch to memory adapter with a warning in logs:

```
level=WARN msg="Failed to initialize Redis event bus, falling back to in-memory adapter"
```

In **test/development** mode, the system will return an error (fail fast).

## Monitoring

### Health Check

```go
ctx := context.Background()
err := eventBus.Health(ctx)
if err != nil {
    // Bus is unhealthy
}
```

### Redis CLI

```bash
# Connect to Redis
docker exec -it promenade_redis redis-cli

# Monitor Pub/Sub
PUBSUB CHANNELS test.*
PUBSUB NUMSUB test.messages
```

## Performance

### Memory vs Redis

**Memory adapter:**

- [+] Faster (no network overhead)
- [+] Isolated in single process
- [-] Does not survive restarts

**Redis adapter:**

- [+] Distributed (multiple instances can listen)
- [+] Persistence possible (if Redis is configured)
- [+] Survives application restarts

### Recommendations

- **Development**: `BUS_ADAPTER=memory`
- **Production (single instance)**: `BUS_ADAPTER=memory`
- **Production (distributed)**: `BUS_ADAPTER=redis`
- **Production (microservices)**: `BUS_ADAPTER=redis`

## Troubleshooting

### Redis connection refused

```bash
# Check if Redis is running
docker ps | grep redis

# Restart Redis
docker-compose -f docker/docker-compose.yml restart redis
```

### Messages not received

1. Verify subscriber is registered **before** publishing event
2. Add `time.Sleep(500 * time.Millisecond)` after Subscribe for Redis Pub/Sub
3. Check that topic name is correct in both Publish and Subscribe

### Graceful shutdown timeout

Increase timeout when closing:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
eventBus.Close(ctx)
```
