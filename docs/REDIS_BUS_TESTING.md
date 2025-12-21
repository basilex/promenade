# Redis Bus Testing Guide

Этот guide описывает как протестировать Redis адаптер для event bus.

## Быстрый старт

### 1. Запуск Redis

```bash
# Запустить Redis через docker-compose
docker-compose -f docker/docker-compose.yml up -d redis

# Проверить что Redis запущен
docker ps | grep redis
```

### 2. Integration тесты

```bash
# Запустить все Redis bus тесты
go test -v ./test/integration/redis_bus_test.go -count=1

# Запустить конкретный тест
go test -v ./test/integration/redis_bus_test.go -run TestRedisBus_PublishSubscribe
```

**Доступные тесты:**

- `TestRedisBus_PublishSubscribe` - базовый pub/sub
- `TestRedisBus_MultipleSubscribers` - несколько подписчиков на один топик
- `TestRedisBus_GracefulShutdown` - graceful shutdown
- `TestBusFallback_RedisToMemory` - fallback на memory при недоступности Redis

### 3. Demo приложение

```bash
# Тест с memory адаптером
BUS_ADAPTER=memory go run examples/redis_bus_demo/main.go

# Тест с Redis адаптером
BUS_ADAPTER=redis REDIS_HOST=localhost REDIS_PORT=6379 \
  go run examples/redis_bus_demo/main.go
```

**Ожидаемый результат:**

```
✅ Bus health check passed
✅ Subscribed to topic
📤 Publishing test events count=5
✅ Published event num=1
📨 Received event type=test.message count=1
...
📊 Test Results published=5 received=5
✅ SUCCESS: All messages received!
✅ Bus closed gracefully
```

## Переключение адаптеров в production

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

В **production** режиме, если Redis недоступен, система автоматически переключится на memory adapter с предупреждением в логах:

```
level=WARN msg="Failed to initialize Redis event bus, falling back to in-memory adapter"
```

В **test/development** режиме, система вернет ошибку (fail fast).

## Мониторинг

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
# Подключиться к Redis
docker exec -it promenade_redis redis-cli

# Мониторинг Pub/Sub
PUBSUB CHANNELS test.*
PUBSUB NUMSUB test.messages
```

## Производительность

### Memory vs Redis

**Memory адаптер:**

- ⚡ Быстрее (нет network overhead)
- 🔒 Изолирован в одном процессе
- ❌ Не переживает рестарты

**Redis адаптер:**

- 🌐 Распределенный (несколько инстансов могут слушать)
- 💾 Возможность persistence (если настроить Redis)
- 🔄 Переживает рестарты приложения

### Рекомендации

- **Development**: `BUS_ADAPTER=memory`
- **Production (single instance)**: `BUS_ADAPTER=memory`
- **Production (distributed)**: `BUS_ADAPTER=redis`
- **Production (microservices)**: `BUS_ADAPTER=redis`

## Troubleshooting

### Redis connection refused

```bash
# Проверить что Redis запущен
docker ps | grep redis

# Перезапустить Redis
docker-compose -f docker/docker-compose.yml restart redis
```

### Messages not received

1. Проверить что subscriber зарегистрирован **до** публикации события
2. Добавить `time.Sleep(500 * time.Millisecond)` после Subscribe для Redis Pub/Sub
3. Проверить что топик правильный в Publish и Subscribe

### Graceful shutdown timeout

Увеличить таймаут при Close:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
eventBus.Close(ctx)
```
