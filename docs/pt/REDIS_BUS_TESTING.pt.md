# Guia de Testes do Redis Bus

[🇬🇧 English](../REDIS_BUS_TESTING.md) | [🇺🇦 Українська](../uk/REDIS_BUS_TESTING.uk.md) | [🇩🇪 Deutsch](../de/REDIS_BUS_TESTING.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/REDIS_BUS_TESTING.es.md)

---

Este guia descreve como testar o adapter Redis para o event bus.

## Início Rápido

### 1. Iniciar o Redis

```bash
# Start Redis via docker-compose
docker-compose -f docker/docker-compose.yml up -d redis

# Verify Redis is running
docker ps | grep redis
```

### 2. Testes de Integração

```bash
# Run all Redis bus tests
go test -v ./test/integration/redis_bus_test.go -count=1

# Run specific test
go test -v ./test/integration/redis_bus_test.go -run TestRedisBus_PublishSubscribe
```

**Testes disponíveis:**

- `TestRedisBus_PublishSubscribe` - pub/sub básico
- `TestRedisBus_MultipleSubscribers` - múltiplos assinantes em um tópico
- `TestRedisBus_GracefulShutdown` - encerramento gracioso
- `TestBusFallback_RedisToMemory` - fallback para memory quando Redis indisponível

### 3. Aplicação de Demonstração

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

## Alternância de Adapters em Produção

### Variáveis de Ambiente

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

### Degradação Graciosa

No modo **production**, se o Redis estiver indisponível, o sistema mudará automaticamente para o memory adapter com um aviso nos logs:

```
level=WARN msg="Failed to initialize Redis event bus, falling back to in-memory adapter"
```

Nos modos **test/development**, o sistema retornará um erro (fail fast).

## Monitoramento

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

## Desempenho

### Memory vs Redis

**Memory adapter:**

- [+] Mais rápido (sem overhead de rede)
- [+] Isolado em um único processo
- [-] Não sobrevive a reinicializações

**Redis adapter:**

- [+] Distribuído (múltiplas instâncias podem ouvir)
- [+] Persistência possível (se Redis estiver configurado)
- [+] Sobrevive a reinicializações da aplicação

### Recomendações

- **Development**: `BUS_ADAPTER=memory`
- **Production (instância única)**: `BUS_ADAPTER=memory`
- **Production (distribuído)**: `BUS_ADAPTER=redis`
- **Production (microsserviços)**: `BUS_ADAPTER=redis`

## Resolução de Problemas

### Redis connection refused

```bash
# Check if Redis is running
docker ps | grep redis

# Restart Redis
docker-compose -f docker/docker-compose.yml restart redis
```

### Mensagens não são recebidas

1. Verifique se o assinante está registrado **antes** de publicar o evento
2. Adicione `time.Sleep(500 * time.Millisecond)` após Subscribe para Redis Pub/Sub
3. Verifique se o nome do tópico está correto tanto em Publish quanto em Subscribe

### Timeout de encerramento gracioso

Aumente o timeout ao fechar:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
eventBus.Close(ctx)
```
