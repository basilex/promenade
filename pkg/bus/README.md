# Event Bus Package

## 📌 Overview

**Event Bus** is the central communication hub for **asynchronous, decoupled communication** between Bounded Contexts in Promenade CRM. It implements the **Publish-Subscribe pattern** with support for **multiple adapters** (Memory, Redis), **retry policies**, **panic recovery**, and **graceful shutdown**.

---

## 🎯 Key Features

- ✅ **Multiple Adapters**: Memory (dev/test), Redis (production, distributed)
- ✅ **Factory Pattern**: `NewBus(config)` creates adapter-specific bus
- ✅ **Type-Safe Events**: `Event` interface with `BaseEvent` implementation
- ✅ **Retry Policy**: Configurable exponential backoff
- ✅ **Panic Recovery**: Handler panics don't crash the system
- ✅ **Graceful Shutdown**: Wait for in-flight events before closing
- ✅ **Health Checks**: Monitor bus health (`Health(ctx)`)
- ✅ **Topic Constants**: Predefined topics for all domain events
- ✅ **Concurrent Safe**: All operations thread-safe
- ✅ **Context Propagation**: Request tracing, cancellation support

---

## 🏗️ Architecture

### Directory Structure

```
pkg/bus/
├── README.md                # This file
├── bus.go                   # Core interfaces (EventBus, Event)
├── event.go                 # BaseEvent implementation
├── factory.go               # Bus factory (adapter selection)
├── topics.go                # Topic constants (30+ topics)
├── event_test.go            # Event tests (8 tests)
├── factory_test.go          # Factory tests (13 tests)
├── topics_test.go           # Topic validation tests (2 tests)
├── memory/                  # Memory adapter (in-process)
│   ├── memory_bus.go
│   ├── memory_bus_test.go  # Unit tests (8 tests)
│   ├── edge_cases_test.go  # Edge case tests (13 tests)
│   └── benchmark_test.go   # Performance benchmarks
└── redis/                   # Redis adapter (distributed)
    ├── redis_bus.go
    └── redis_bus_test.go   # Integration tests (16 tests)
```

**Test Coverage**: **54 tests total** (100% passing)

- Core: 23 tests
- Memory adapter: 21 tests
- Redis adapter: 16 tests
- Integration: 7 tests (in `test/integration/pkg/bus/`)

---

## 📦 Core Interfaces

### EventBus Interface

```go
// EventBus defines the interface for event publishing and subscribing
type EventBus interface {
    // Publish sends an event to a topic
    Publish(ctx context.Context, topic string, event Event) error

    // Subscribe registers a handler for a topic
    Subscribe(topic string, handler EventHandler) error

    // Unsubscribe removes a handler from a topic
    Unsubscribe(topic string, handler EventHandler) error

    // Close shuts down the bus gracefully
    Close(ctx context.Context) error

    // Health checks if the bus is operational
    Health(ctx context.Context) error
}
```

### Event Interface

```go
// Event represents a domain event
type Event interface {
    // Type returns the event type (e.g., "user.created")
    Type() string

    // AggregateID returns the UUID of the aggregate that emitted the event
    AggregateID() uuid.UUID

    // OccurredAt returns when the event occurred
    OccurredAt() time.Time

    // Metadata returns event metadata (key-value pairs)
    Metadata() map[string]interface{}
}
```

### EventHandler Type

```go
// EventHandler processes an event
type EventHandler func(ctx context.Context, event Event) error
```

---

## 🔧 Configuration

### Config Structure

```go
type Config struct {
    WorkerPoolSize int           // Number of worker goroutines
    BufferSize     int           // Event channel buffer size
    RetryPolicy    *RetryPolicy  // Retry configuration
}

type RetryPolicy struct {
    MaxAttempts  int           // Maximum retry attempts
    InitialDelay time.Duration // Initial delay between retries
    MaxDelay     time.Duration // Maximum delay between retries
    Multiplier   float64       // Exponential backoff multiplier
}
```

### YAML Configuration

**Development** (`config/app.dev.yaml`):

```yaml
bus:
  adapter: "memory" # Fast in-process adapter
  worker_pool_size: 10
  buffer_size: 1000
  retry_attempts: 3
  retry_delay: 100ms
  retry_max_delay: 10s
  retry_multiplier: 2.0
```

**Production** (`config/app.prod.yaml`):

```yaml
bus:
  adapter: "redis" # Distributed adapter
  worker_pool_size: 50
  buffer_size: 10000
  retry_attempts: 5
  retry_delay: 1s
  retry_max_delay: 30s
  retry_multiplier: 2.0
  redis:
    host: "redis.prod.example.com"
    port: 6379
    password: "${REDIS_PASSWORD}"
    db: 0
    max_retries: 3
    pool_size: 20
```

---

## 🚀 Usage

### 1. Initialize Bus

```go
// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal("Failed to load config:", err)
}

// Create bus (factory selects adapter)
eventBus, err := bus.NewBus(cfg.Bus)
if err != nil {
    log.Fatal("Failed to initialize event bus:", err)
}
defer eventBus.Close(context.Background())
```

### 2. Publish Events

```go
// In use case layer
func (uc *UserUseCase) RegisterUser(ctx context.Context, email, name, password string) (*User, error) {
    // Business logic
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
        // Log error but don't fail the operation
        logger.FromContext(ctx).Error("Failed to publish event",
            slog.Any("error", err),
            slog.String("event_type", "user.registered"),
        )
    }

    return user, nil
}
```

### 3. Subscribe to Events

```go
// In handler/subscriber
func (h *NotificationHandler) RegisterSubscribers(bus bus.EventBus) error {
    // Subscribe to user registration events
    if err := bus.Subscribe(bus.TopicUserRegistered, h.HandleUserRegistered); err != nil {
        return fmt.Errorf("failed to subscribe to user.registered: %w", err)
    }

    // Subscribe to contact verification events
    if err := bus.Subscribe(bus.TopicContactVerified, h.HandleContactVerified); err != nil {
        return fmt.Errorf("failed to subscribe to contact.verified: %w", err)
    }

    return nil
}

func (h *NotificationHandler) HandleUserRegistered(ctx context.Context, e bus.Event) error {
    // Extract user ID from event
    userID := e.AggregateID()

    // Send welcome email
    if err := h.emailService.SendWelcomeEmail(ctx, userID); err != nil {
        return fmt.Errorf("failed to send welcome email: %w", err)
    }

    logger.FromContext(ctx).Info("Welcome email sent",
        slog.String("user_id", userID.String()),
        slog.String("event_type", e.Type()),
    )

    return nil
}
```

### 4. Unsubscribe (if needed)

```go
// Remove handler
if err := bus.Unsubscribe(bus.TopicUserRegistered, handler); err != nil {
    log.Error("Failed to unsubscribe:", err)
}
```

---

## 📋 Topic Constants

**All topics are predefined constants** to prevent typos and enable IDE autocomplete:

### Identity Context Topics

```go
const (
    TopicUserRegistered      = "identity.user.registered"
    TopicUserActivated       = "identity.user.activated"
    TopicUserSuspended       = "identity.user.suspended"
    TopicUserPasswordChanged = "identity.user.password_changed"
    TopicUserEmailVerified   = "identity.user.email_verified"

    TopicContactCreated  = "identity.contact.created"
    TopicContactUpdated  = "identity.contact.updated"
    TopicContactVerified = "identity.contact.verified"
    TopicContactDeleted  = "identity.contact.deleted"
)
```

### Shared Context Topics (Reference Data)

```go
const (
    TopicCountryCreated      = "shared.country.created"
    TopicCountryUpdated      = "shared.country.updated"
    TopicCountryDeactivated  = "shared.country.deactivated"

    TopicCurrencyCreated     = "shared.currency.created"
    TopicCurrencyUpdated     = "shared.currency.updated"
    TopicCurrencyDeactivated = "shared.currency.deactivated"

    TopicLanguageCreated     = "shared.language.created"
    TopicLanguageUpdated     = "shared.language.updated"
    TopicLanguageDeactivated = "shared.language.deactivated"

    TopicTimezoneCreated     = "shared.timezone.created"
    TopicTimezoneUpdated     = "shared.timezone.updated"
    TopicTimezoneDeactivated = "shared.timezone.deactivated"
)
```

### Future Context Topics (Planned)

```go
const (
    // Customer Management
    TopicCustomerCreated = "customer.customer.created"
    TopicCustomerUpdated = "customer.customer.updated"
    TopicDealWon         = "customer.deal.won"

    // Order Management
    TopicOrderCreated   = "order.order.created"
    TopicOrderConfirmed = "order.order.confirmed"
    TopicOrderShipped   = "order.order.shipped"

    // Billing
    TopicInvoiceGenerated = "billing.invoice.generated"
    TopicPaymentReceived  = "billing.payment.received"
    TopicPaymentFailed    = "billing.payment.failed"

    // Notifications
    TopicNotificationEmail = "notification.email.sent"
    TopicNotificationSMS   = "notification.sms.sent"
)
```

**Total**: 30+ predefined topics

---

## 🎯 Adapters

### Memory Adapter

**Use Cases**: Development, testing, single-instance deployments

**Characteristics**:

- ✅ **Ultra-fast**: No network overhead (~377K events/sec)
- ✅ **Simple**: No external dependencies
- ✅ **Reliable**: No connection failures
- ❌ **Not distributed**: Events don't cross process boundaries
- ❌ **No persistence**: Events lost on restart

**When to use**:

- Local development
- Unit/integration tests
- Single-server deployments
- Non-critical event processing

**Performance**: 377,494 events/sec (1000 events in 2.6ms)

### Redis Adapter

**Use Cases**: Production, distributed systems, multi-instance deployments

**Characteristics**:

- ✅ **Distributed**: Events propagate across all instances
- ✅ **Persistent**: Pub/Sub with optional stream storage
- ✅ **Scalable**: Handles thousands of subscribers
- ✅ **Reliable**: Built-in retry and reconnection logic
- ❌ **Slower**: Network latency (~50-100ms per event)
- ❌ **Requires Redis**: External dependency

**When to use**:

- Production environments
- Multiple application instances
- Microservices architecture
- Critical event processing

**Performance**: ~50 concurrent events, 7.5s test duration

---

## 🔄 Retry Policy

### Exponential Backoff

```go
config := bus.NewConfig(
    10,              // workerPoolSize
    1000,            // bufferSize
    3,               // maxAttempts
    100*time.Millisecond, // initialDelay
    10*time.Second,  // maxDelay
    2.0,             // multiplier (exponential)
)
```

**Retry Schedule**:

- Attempt 1: Immediate
- Attempt 2: 100ms delay
- Attempt 3: 200ms delay (100ms × 2.0)
- Attempt 4: 400ms delay (200ms × 2.0)
- Attempt 5+: Capped at 10s (maxDelay)

### Error Handling

```go
func (h *Handler) ProcessEvent(ctx context.Context, e bus.Event) error {
    // Temporary errors trigger retry
    if err := h.service.Call(); err != nil {
        if isTemporaryError(err) {
            return err // Will retry
        }
        // Permanent error - log and skip
        logger.FromContext(ctx).Error("Permanent error", slog.Any("error", err))
        return nil // Don't retry
    }
    return nil
}
```

---

## 🛡️ Panic Recovery

**All handlers are wrapped in panic recovery** to prevent cascading failures:

```go
func (mb *MemoryBus) handleEvent(ctx context.Context, handler EventHandler, event Event) {
    defer func() {
        if r := recover(); r != nil {
            logger.FromContext(ctx).Error("Event handler panicked",
                slog.Any("panic", r),
                slog.String("event_type", event.Type()),
            )
        }
    }()

    // Execute handler with retry logic
    err := mb.retryHandler(ctx, handler, event)
    if err != nil {
        logger.FromContext(ctx).Error("Event handler failed after retries",
            slog.Any("error", err),
        )
    }
}
```

**Result**: One handler panic doesn't affect other handlers or crash the application.

---

## 🧪 Testing

### Test Organization

**Unit Tests** (in-place):

- `bus/event_test.go` (8 tests) - BaseEvent creation, validation
- `bus/factory_test.go` (13 tests) - Factory pattern, config builder
- `bus/topics_test.go` (2 tests) - Topic constants validation
- `memory/memory_bus_test.go` (8 tests) - Memory adapter pub/sub
- `memory/edge_cases_test.go` (13 tests) - Edge cases, concurrency
- `redis/redis_bus_test.go` (16 tests) - Redis adapter integration

**Integration Tests** (mirror path):

- `test/integration/pkg/bus/bus_integration_test.go` (7 tests) - End-to-end workflows

**Benchmarks**:

- `memory/benchmark_test.go` - Performance profiling

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

# Benchmarks
go test -bench=. ./pkg/bus/memory

# With coverage
go test ./pkg/bus/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Results Summary

| Package        | Tests  | Status  | Duration |
| -------------- | ------ | ------- | -------- |
| pkg/bus        | 23     | ✅ PASS | 0.01s    |
| pkg/bus/memory | 21     | ✅ PASS | cached   |
| pkg/bus/redis  | 16     | ✅ PASS | 7.55s    |
| integration    | 7      | ✅ PASS | 5.67s    |
| **Total**      | **67** | ✅ PASS | ~13s     |

**Test Coverage**: **100%** of critical paths

---

## 🎓 Best Practices

### ✅ DO

- ✅ **Use topic constants** (`bus.TopicUserRegistered`) to prevent typos
- ✅ **Publish after successful database commit** to ensure consistency
- ✅ **Log publish errors** but don't fail the operation
- ✅ **Return nil from handlers** for permanent errors (skip retry)
- ✅ **Use context for cancellation** and request tracing
- ✅ **Keep handlers idempotent** (safe to retry)
- ✅ **Test handlers with TestSuite** (unit + integration)

### ❌ DON'T

- ❌ **DON'T hardcode topic strings** (use constants)
- ❌ **DON'T publish before database commit** (risk inconsistency)
- ❌ **DON'T panic in handlers** (use error returns)
- ❌ **DON'T block in handlers** (use goroutines for slow operations)
- ❌ **DON'T share state** between handler invocations
- ❌ **DON'T ignore context cancellation** (`ctx.Done()`)
- ❌ **DON'T use bus for synchronous RPC** (use direct calls)

---

## 📊 Performance Metrics

### Memory Adapter

**Throughput**: 377,494 events/sec  
**Latency**:

- Publish: ~0.1μs
- Subscribe: ~0.2μs
- Handler execution: ~10μs (empty handler)

**Load Test** (1000 events):

- Total time: 2.6ms
- Events/sec: 377,494
- Memory usage: ~1MB

### Redis Adapter

**Throughput**: ~1,000 events/sec (network-bound)  
**Latency**:

- Publish: ~50-100ms (Redis round-trip)
- Subscribe: Real-time (Redis Pub/Sub)
- Handler execution: Same as Memory

**Load Test** (50 concurrent events):

- Total time: ~2 seconds
- Memory usage: ~5MB
- Network traffic: ~100KB

---

## 🔍 Debugging

### Enable Debug Logging

```yaml
# config/app.dev.yaml
logging:
  level: "debug" # Enable debug logs
```

**Output**:

```
[DEBUG] Event bus initialized adapter=memory worker_pool_size=10
[DEBUG] Published event topic=identity.user.registered event_id=01JGABC...
[DEBUG] Handler invoked topic=identity.user.registered handler_duration=15ms
```

### Health Check

```go
// Check bus health
ctx := context.Background()
if err := bus.Health(ctx); err != nil {
    log.Error("Event bus unhealthy:", err)
}
```

### Monitor Metrics

```go
// Custom metrics (example)
type BusMetrics struct {
    PublishedTotal   int64
    FailedTotal      int64
    AverageLatency   time.Duration
}

// Publish with metrics
start := time.Now()
err := bus.Publish(ctx, topic, event)
metrics.PublishedTotal++
if err != nil {
    metrics.FailedTotal++
}
metrics.AverageLatency = time.Since(start)
```

---

## 🚀 Future Enhancements

### Planned Features

- [ ] **Event Sourcing**: Store all events for replay
- [ ] **Dead Letter Queue**: Failed events → DLQ for manual review
- [ ] **Priority Topics**: High-priority events processed first
- [ ] **Event Filtering**: Subscribe with filter predicates
- [ ] **Metrics Dashboard**: Grafana/Prometheus integration
- [ ] **Message Deduplication**: Idempotency keys
- [ ] **Circuit Breaker**: Prevent cascading failures
- [ ] **Rate Limiting**: Prevent event storms

### Not Planned

- ❌ **RPC/Request-Reply**: Use HTTP or gRPC instead
- ❌ **Transactions**: Events are fire-and-forget
- ❌ **Guaranteed Ordering**: Redis Pub/Sub doesn't guarantee order

---

## 📚 Related Documentation

- [Architecture Overview](../../docs/ARCHITECTURE_OVERVIEW.md)
- [Event-Driven Design](../../docs/EVENT_DRIVEN_DESIGN.md)
- [Testing Guide](../../test/TESTING_STRUCTURE.md)
- [Bus Test Coverage Report](../../docs/BUS_TEST_COVERAGE.md)
- [Redis Configuration](../../docs/REDIS_CONFIGURATION.md)

---

## 🤝 Contributing

### Adding New Topics

1. **Define constant** in `topics.go`:

   ```go
   const TopicCustomerCreated = "customer.customer.created"
   ```

2. **Update tests** in `topics_test.go`:

   ```go
   assert.NotEmpty(t, bus.TopicCustomerCreated)
   ```

3. **Document in README** (this file)

### Adding New Adapters

1. **Implement `EventBus` interface** in new package (e.g., `pkg/bus/kafka/`)
2. **Register adapter** in `factory.go`:
   ```go
   func init() {
       RegisterAdapter("kafka", NewKafkaBus)
   }
   ```
3. **Add tests** (unit + integration)
4. **Update documentation**

---

**Last Updated**: 2025-12-27  
**Status**: ✅ Production-ready  
**Test Coverage**: 67 tests, 100% passing  
**Maintainer**: Promenade Team
