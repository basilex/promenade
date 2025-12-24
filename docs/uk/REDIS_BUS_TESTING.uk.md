# Посібник з тестування Redis IBus

[🇬🇧 English](../REDIS_BUS_TESTING.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](REDIS_BUS_TESTING.de.md) | [🇵🇹 Português](REDIS_BUS_TESTING.pt.md) | [🇪🇸 Español](REDIS_BUS_TESTING.es.md)

---

Цей посібник описує, як тестувати Redis adapter для event bus.

## Швидкий старт

### 1. Запуск Redis

```bash
# Start Redis via docker-compose
docker-compose -f docker/docker-compose.yml up -d redis

# Verify Redis is running
docker ps | grep redis
```

### 2. Інтеграційні тести

```bash
# Run all Redis bus tests
go test -v ./test/integration/redis_bus_test.go -count=1

# Run specific test
go test -v ./test/integration/redis_bus_test.go -run TestRedisBus_PublishSubscribe
```

**Доступні тести:**

- `TestRedisBus_PublishSubscribe` - базовий pub/sub
- `TestRedisBus_MultipleSubscribers` - декілька підписників на одну тему
- `TestRedisBus_GracefulShutdown` - коректне завершення
- `TestBusFallback_RedisToMemory` - відмовостійкість з переходом на memory при недоступності Redis

### 3. Демо-додаток

```bash
# Test with memory adapter
BUS_ADAPTER=memory go run examples/redis_bus_demo/main.go

# Test with Redis adapter
BUS_ADAPTER=redis REDIS_HOST=localhost REDIS_PORT=6379 \
  go run examples/redis_bus_demo/main.go
```

**Очікуваний результат:**

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

## Перемикання адаптерів у продакшені

### Змінні оточення

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

### Плавне відновлення

У режимі **production**, якщо Redis недоступний, система автоматично переключиться на memory adapter з попередженням у логах:

```
level=WARN msg="Failed to initialize Redis event bus, falling back to in-memory adapter"
```

У режимах **test/development**, система поверне помилку (fail fast).

## Моніторинг

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

## Продуктивність

### Memory vs Redis

**Memory adapter:**

- [+] Швидший (немає мережевих витрат)
- [+] Ізольований в одному процесі
- [-] Не зберігається після перезапуску

**Redis adapter:**

- [+] Розподілений (декілька інстансів можуть слухати)
- [+] Можлива персистентність (якщо Redis налаштований)
- [+] Переживає перезапуски додатку

### Рекомендації

- **Development**: `BUS_ADAPTER=memory`
- **Production (один інстанс)**: `BUS_ADAPTER=memory`
- **Production (розподілений)**: `BUS_ADAPTER=redis`
- **Production (мікросервіси)**: `BUS_ADAPTER=redis`

## Усунення проблем

### Redis connection refused

```bash
# Check if Redis is running
docker ps | grep redis

# Restart Redis
docker-compose -f docker/docker-compose.yml restart redis
```

### Повідомлення не отримуються

1. Переконайтеся, що підписник зареєстрований **до** публікації події
2. Додайте `time.Sleep(500 * time.Millisecond)` після Subscribe для Redis Pub/Sub
3. Перевірте, що назва теми коректна і в Publish, і в Subscribe

### Таймаут коректного завершення

Збільште таймаут при закритті:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
eventBus.Close(ctx)
```
