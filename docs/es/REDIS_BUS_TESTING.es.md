# Guía de Pruebas de Redis Bus

[🇬🇧 English](../REDIS_BUS_TESTING.md) | [🇺🇦 Українська](../uk/REDIS_BUS_TESTING.uk.md) | [🇩🇪 Deutsch](../de/REDIS_BUS_TESTING.de.md) | [🇵🇹 Português](../pt/REDIS_BUS_TESTING.pt.md) | 🇪🇸 **Español**

---

Esta guía describe cómo probar el adapter Redis para el event bus.

## Inicio Rápido

### 1. Iniciar Redis

```bash
# Start Redis via docker-compose
docker-compose -f docker/docker-compose.yml up -d redis

# Verify Redis is running
docker ps | grep redis
```

### 2. Pruebas de Integración

```bash
# Run all Redis bus tests
go test -v ./test/integration/redis_bus_test.go -count=1

# Run specific test
go test -v ./test/integration/redis_bus_test.go -run TestRedisBus_PublishSubscribe
```

**Pruebas disponibles:**

- `TestRedisBus_PublishSubscribe` - pub/sub básico
- `TestRedisBus_MultipleSubscribers` - múltiples suscriptores a un tema
- `TestRedisBus_GracefulShutdown` - cierre gracioso
- `TestBusFallback_RedisToMemory` - fallback a memory cuando Redis no está disponible

### 3. Aplicación de Demostración

```bash
# Test with memory adapter
BUS_ADAPTER=memory go run examples/redis_bus_demo/main.go

# Test with Redis adapter
BUS_ADAPTER=redis REDIS_HOST=localhost REDIS_PORT=6379 \
  go run examples/redis_bus_demo/main.go
```

**Resultado esperado:**

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

## Cambio de Adapters en Producción

### Variables de Entorno

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

### Degradación Graciosa

En modo **production**, si Redis no está disponible, el sistema cambiará automáticamente al memory adapter con una advertencia en los logs:

```
level=WARN msg="Failed to initialize Redis event bus, falling back to in-memory adapter"
```

En modos **test/development**, el sistema devolverá un error (fail fast).

## Monitoreo

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

## Rendimiento

### Memory vs Redis

**Memory adapter:**

- [+] Más rápido (sin overhead de red)
- [+] Aislado en un solo proceso
- [-] No sobrevive a reinicios

**Redis adapter:**

- [+] Distribuido (múltiples instancias pueden escuchar)
- [+] Persistencia posible (si Redis está configurado)
- [+] Sobrevive a reinicios de la aplicación

### Recomendaciones

- **Development**: `BUS_ADAPTER=memory`
- **Production (instancia única)**: `BUS_ADAPTER=memory`
- **Production (distribuido)**: `BUS_ADAPTER=redis`
- **Production (microservicios)**: `BUS_ADAPTER=redis`

## Solución de Problemas

### Redis connection refused

```bash
# Check if Redis is running
docker ps | grep redis

# Restart Redis
docker-compose -f docker/docker-compose.yml restart redis
```

### Los mensajes no se reciben

1. Verifique que el suscriptor esté registrado **antes** de publicar el evento
2. Agregue `time.Sleep(500 * time.Millisecond)` después de Subscribe para Redis Pub/Sub
3. Verifique que el nombre del tema sea correcto tanto en Publish como en Subscribe

### Timeout de cierre gracioso

Aumente el timeout al cerrar:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
eventBus.Close(ctx)
```
