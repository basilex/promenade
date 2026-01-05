# Event Bus Package

The **Event Bus** is the central communication hub for asynchronous, decoupled communication between Bounded Contexts in Promenade Platform.

## Features

- **Multiple Adapters** - Memory (dev/test), Redis (production, distributed)
- **Factory Pattern** - `NewBus(config)` creates adapter-specific bus
- **Type-Safe Events** - `Event` interface with `BaseEvent` implementation
- **Retry Policy** - Configurable exponential backoff
- **Panic Recovery** - Handler panics don't crash the system
- **Graceful Shutdown** - Wait for in-flight events before closing
- **Health Checks** - Monitor bus health (`Health(ctx)`)
- **67 Tests** - 100% passing, production-ready

## Quick Start

### Initialize Bus

```go
// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Create bus (factory selects adapter based on cfg.Bus.Adapter)
eventBus, err := bus.NewBus(cfg.Bus, cfg.Database.Redis)
if err != nil {
    log.Fatal(err)
}
defer eventBus.Close(context.Background())
```

### Publish Events

```go
// In use case layer
func (uc *UserUseCase) RegisterUser(ctx context.Context, email, name, password string) (*User, error) {
    user, err := entity.NewUser(email, name, password)
    if err != nil {
        return nil, err
    }
    
    if err := uc.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // Publish domain event
    event := bus.NewBaseEvent("user.registered", user.ID)
    if err := uc.bus.Publish(ctx, bus.TopicUserRegistered, event); err != nil {
        logger.FromContext(ctx).Error("Failed to publish event", slog.Any("error", err))
    }
    
    return user, nil
}
```

### Subscribe to Events

```go
// In handler/subscriber
func (h *NotificationHandler) RegisterSubscribers(bus bus.EventBus) error {
    // Subscribe to user registration events
    if err := bus.Subscribe(bus.TopicUserRegistered, h.HandleUserRegistered); err != nil {
        return err
    }
    return nil
}

func (h *NotificationHandler) HandleUserRegistered(ctx context.Context, e bus.Event) error {
    userID := e.AggregateID()
    
    // Send welcome email
    if err := h.emailService.SendWelcomeEmail(ctx, userID); err != nil {
        return fmt.Errorf("failed to send welcome email: %w", err)
    }
    
    logger.FromContext(ctx).Info("Welcome email sent", slog.String("user_id", userID.String()))
    return nil
}
```

## Configuration

### Memory Adapter (Development)

```yaml
# config/app.postgres-dev.yaml or config/app.sqlite-dev.yaml
bus:
  adapter: "memory"  # Fast in-process adapter
  worker_pool_size: 10
  buffer_size: 1000
  retry_attempts: 3
  retry_delay: 100ms
```

**Performance**: 377,494 events/sec

### Redis Adapter (Production)

```yaml
# config/app.postgres-prod.yaml
database:
  redis:
    addr: "${REDIS_ADDR}"
    password: "${REDIS_PASSWORD}"
    databases:
      bus: 1  # Event bus database

bus:
  adapter: "redis"  # Distributed adapter
  worker_pool_size: 50
  buffer_size: 10000
  retry_attempts: 5
  retry_delay: 1s
```

## Topic Constants

**All topics are predefined** to prevent typos:

### Identity Context

```go
const (
    TopicUserRegistered      = "identity.user.registered"
    TopicUserActivated       = "identity.user.activated"
    TopicUserSuspended       = "identity.user.suspended"
    TopicUserPasswordChanged = "identity.user.password_changed"
    
    TopicContactCreated  = "identity.contact.created"
    TopicContactVerified = "identity.contact.verified"
    TopicContactUpdated  = "identity.contact.updated"
)
```

### Shared Context

```go
const (
    TopicCountryCreated     = "shared.country.created"
    TopicCurrencyCreated    = "shared.currency.created"
    TopicLanguageCreated    = "shared.language.created"
    TopicTimezoneCreated    = "shared.timezone.created"
)
```

### Customer Management

```go
const (
    TopicCustomerCreated = "customer.customer.created"
    TopicCustomerUpdated = "customer.customer.updated"
    TopicDealWon         = "customer.deal.won"
)
```

**Total**: 30+ predefined topics

## Adapters

### Memory Adapter

**Use Cases**: Development, testing, single-instance deployments

**Characteristics**:
- Ultra-fast: ~377K events/sec
- No external dependencies
- Not distributed
- No persistence

```go
// Automatically selected with adapter: "memory"
eventBus, _ := bus.NewBus(cfg.Bus, cfg.Database.Redis)
```

### Redis Adapter

**Use Cases**: Production, distributed systems

**Characteristics**:
- Distributed: Events propagate across instances
- Persistent: Pub/Sub with optional streams
- Scalable: Handles thousands of subscribers
- Slower: ~50-100ms per event (network latency)

```go
// Automatically selected with adapter: "redis"
eventBus, _ := bus.NewBus(cfg.Bus, cfg.Database.Redis)
```

## Retry Policy

**Exponential backoff** with configurable attempts:

```go
config := bus.NewConfig(
    10,                      // workerPoolSize
    1000,                    // bufferSize
    3,                       // maxAttempts
    100*time.Millisecond,    // initialDelay
    10*time.Second,          // maxDelay
    2.0,                     // multiplier
)
```

**Retry Schedule**:
- Attempt 1: Immediate
- Attempt 2: 100ms delay
- Attempt 3: 200ms delay
- Attempt 4: 400ms delay
- Capped at 10s

## Panic Recovery

**All handlers wrapped in panic recovery**:

```go
defer func() {
    if r := recover(); r != nil {
        logger.Error("Handler panicked", slog.Any("panic", r))
    }
}()
```

One handler panic doesn't affect other handlers or crash the application.

## Testing

### Run Tests

```bash
# All bus tests
go test ./pkg/bus/... -v

# Memory adapter only
go test ./pkg/bus/memory -v

# Redis adapter (requires Redis)
go test ./pkg/bus/redis -v

# Integration tests
go test ./test/integration/pkg/bus -v
```

**Test Coverage**: 67 tests, 100% passing

### Test Example

```go
func TestBus_PublishSubscribe(t *testing.T) {
    eventBus := memory.NewMemoryBus(bus.DefaultConfig())
    defer eventBus.Close(context.Background())
    
    received := make(chan string, 1)
    
    eventBus.Subscribe("test.topic", func(ctx context.Context, e bus.Event) error {
        received <- e.Type()
        return nil
    })
    
    event := bus.NewBaseEvent("test.event", uuidv7.New())
    err := eventBus.Publish(context.Background(), "test.topic", event)
    
    assert.NoError(t, err)
    assert.Equal(t, "test.event", <-received)
}
```

## Best Practices

### DO

-  Use topic constants (`bus.TopicUserRegistered`)
-  Publish after successful database commit
-  Log publish errors but don't fail operation
-  Return `nil` for permanent errors (skip retry)
-  Use context for cancellation
-  Keep handlers idempotent

### DON'T

-  Don't hardcode topic strings
-  Don't publish before database commit
-  Don't panic in handlers
-  Don't block in handlers
-  Don't share state between handler invocations
-  Don't use bus for synchronous RPC

## Performance Metrics

### Memory Adapter

**Throughput**: 377,494 events/sec  
**Latency**:
- Publish: ~0.1μs
- Subscribe: ~0.2μs
- Handler: ~10μs (empty handler)

**Load Test** (1000 events):
- Total time: 2.6ms
- Memory usage: ~1MB

### Redis Adapter

**Throughput**: ~1,000 events/sec (network-bound)  
**Latency**:
- Publish: ~50-100ms (Redis round-trip)
- Subscribe: Real-time (Redis Pub/Sub)

## Health Check

```go
// Check bus health
ctx := context.Background()
if err := bus.Health(ctx); err != nil {
    log.Error("Event bus unhealthy:", err)
}
```

## Context Communication

**Rule**: Contexts communicate ONLY via Event Bus:

```go
//  DON'T - Direct import
import "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"

//  DO - Event Bus
event := bus.NewBaseEvent("user.registered", userID)
eventBus.Publish(ctx, bus.TopicUserRegistered, event)
```

## Next Steps

- [Architecture Guide](/guide/architecture) - Event-Driven patterns
- [Identity Context](/contexts/identity) - Domain events examples
- [Testing Guide](/guide/testing) - Test event handlers
- [GitHub Docs](https://github.com/basilex/promenade/blob/dev/pkg/bus/README.md) - Complete documentation
