# Redis IBus Testing Leitfaden

[ English](../REDIS_BUS_TESTING.md) | [ Українська](../uk/REDIS_BUS_TESTING.uk.md) |  **Deutsch** | [ Português](../pt/REDIS_BUS_TESTING.pt.md) | [ Español](../es/REDIS_BUS_TESTING.es.md)

---

Dieser Leitfaden beschreibt, wie der Redis Adapter für den Event IBus getestet wird.

## Schnellstart

### 1. Redis starten

```bash
# Start Redis via docker-compose
docker-compose -f docker/docker-compose.yml up -d redis

# Verify Redis is running
docker ps | grep redis
```

### 2. Integrationstests

```bash
# Run all Redis bus tests
go test -v ./test/integration/redis_bus_test.go -count=1

# Run specific test
go test -v ./test/integration/redis_bus_test.go -run TestRedisBus_PublishSubscribe
```

**Verfügbare Tests:**

- `TestRedisBus_PublishSubscribe` - grundlegende Pub/Sub-Funktionalität
- `TestRedisBus_MultipleSubscribers` - mehrere Subscriber auf ein Topic
- `TestRedisBus_GracefulShutdown` - ordnungsgemäßes Herunterfahren
- `TestBusFallback_RedisToMemory` - Fallback auf Memory bei Redis-Ausfall

### 3. Demo-Anwendung

```bash
# Test with memory adapter
BUS_ADAPTER=memory go run examples/redis_bus_demo/main.go

# Test with Redis adapter
BUS_ADAPTER=redis REDIS_HOST=localhost REDIS_PORT=6379 \
  go run examples/redis_bus_demo/main.go
```

**Erwartetes Ergebnis:**

```
[OK] IBus health check passed
[OK] Subscribed to topic
[>>] Publishing test events count=5
[OK] Published event num=1
[<<] Received event type=test.message count=1
...
[RESULTS] Test Results published=5 received=5
[SUCCESS] All messages received!
[OK] IBus closed gracefully
```

## Adapter-Wechsel in der Produktion

### Umgebungsvariablen

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

Im **Production**-Modus schaltet das System automatisch auf den Memory Adapter um, wenn Redis nicht verfügbar ist, mit einer Warnung in den Logs:

```
level=WARN msg="Failed to initialize Redis event bus, falling back to in-memory adapter"
```

Im **Test/Development**-Modus gibt das System einen Fehler zurück (fail fast).

## Überwachung

### Health Check

```go
ctx := context.Background()
err := eventBus.Health(ctx)
if err != nil {
    // IBus is unhealthy
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

## Leistung

### Memory vs Redis

**Memory Adapter:**

- [+] Schneller (kein Netzwerk-Overhead)
- [+] Isoliert in einem Prozess
- [-] Überlebt Neustarts nicht

**Redis Adapter:**

- [+] Verteilt (mehrere Instanzen können zuhören)
- [+] Persistenz möglich (wenn Redis konfiguriert ist)
- [+] Überlebt Anwendungsneustarts

### Empfehlungen

- **Development**: `BUS_ADAPTER=memory`
- **Production (einzelne Instanz)**: `BUS_ADAPTER=memory`
- **Production (verteilt)**: `BUS_ADAPTER=redis`
- **Production (Microservices)**: `BUS_ADAPTER=redis`

## Fehlerbehebung

### Redis connection refused

```bash
# Check if Redis is running
docker ps | grep redis

# Restart Redis
docker-compose -f docker/docker-compose.yml restart redis
```

### Nachrichten werden nicht empfangen

1. Stellen Sie sicher, dass der Subscriber **vor** der Event-Publikation registriert ist
2. Fügen Sie `time.Sleep(500 * time.Millisecond)` nach Subscribe für Redis Pub/Sub hinzu
3. Prüfen Sie, dass der Topic-Name sowohl in Publish als auch Subscribe korrekt ist

### Timeout beim ordnungsgemäßen Herunterfahren

Erhöhen Sie das Timeout beim Schließen:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
eventBus.Close(ctx)
```
