# Event IBus Package

Universal message bus implementation with dual adapters (Memory + Redis) for event-driven architecture in Promenade.

## Overview

The `pkg/bus` package provides a **transport-agnostic event bus** with pluggable adapters for different deployment scenarios:

- **Memory Adapter** - In-memory Pub/Sub for development, testing, and single-instance deployments
- **Redis Adapter** - Distributed Pub/Sub for multi-instance production deployments with horizontal scaling

This enables asynchronous, decoupled communication between components with a clear evolution path from monolith to microservices.

## Architecture Benefits

### Current (Monolith + Async)

**Architecture:**

- Single binary with embedded workers
- Event bus coordinates async communication
- Workers: Email, Analytics, Logs
- Single Database (ACID preserved)

**Flow:**

1. UseCase publishes event to Event IBus
2. Event IBus dispatches to multiple workers concurrently
3. Workers (Email, Analytics, Logs) process independently
4. All within single process boundary

### Future (Microservices)

**Architecture:**

- Gateway routes requests
- Distributed message broker (NATS/Kafka)
- Independent services: Email, Analytics, User
- Each service has own database

**Flow:**

1. Gateway receives request
2. Publishes event to distributed broker (NATS/Kafka)
3. Services subscribe to relevant topics
4. Services process independently and scale horizontally

**Same code, different transport!** Just swap `memory.IBus` → `redis.IBus` or `nats.IBus`

## Features

- [+] **Transport-agnostic** — in-memory, Redis Pub/Sub, NATS, Kafka
- [+] **Non-blocking** — publish returns immediately, processing in background
- [+] **Multiple subscribers** — many workers can listen to same event
- [+] **Worker pool** — configurable concurrent processing
- [+] **Graceful shutdown** — waits for in-flight messages
- [+] **Type-safe events** — domain events with full type information
- [+] **Bounded contexts** — events organized by domain

## Quick Start

### 1. Define Domain Event

```go
// internal/domain/event/user_events.go
type UserRegisteredEvent struct {
    *bus.BaseEvent
    UserID uuidv7.UUID `json:"user_id"`
    Email  string      `json:"email"`
    Name   string      `json:"name"`
}

func NewUserRegisteredEvent(userID uuidv7.UUID, email, name string) *UserRegisteredEvent {
    return &UserRegisteredEvent{
        BaseEvent: bus.NewBaseEvent(bus.TopicUserRegistered, userID),
        UserID:    userID,
        Email:     email,
        Name:      name,
    }
}
```

### 2. Publish Event from Use Case

```go
// internal/usecase/auth_usecase.go
func (uc *authUseCase) Register(ctx context.Context, email, name, password string) (*entity.User, error) {
    // ... create user in database ...

    // Publish event (асинхронно, не блокирует регистрацию)
    event := event.NewUserRegisteredEvent(user.ID, user.Email, user.Name)
    if err := uc.eventBus.Publish(ctx, bus.TopicUserRegistered, event); err != nil {
        // Логируем, но не фейлим регистрацию
        log.Printf("Failed to publish event: %v", err)
    }

    return user, nil
}
```

### 3. Subscribe to Events

```go
// internal/infrastructure/notification/email_service.go
func (s *EmailService) Start(ctx context.Context) error {
    return s.bus.Subscribe(bus.TopicUserRegistered, s.handleUserRegistered)
}

func (s *EmailService) handleUserRegistered(ctx context.Context, e bus.Event) error {
    evt := e.(*event.UserRegisteredEvent)

    // Отправка email в фоне, не блокируя основной поток
    return s.sender.Send(ctx, Email{
        To:      evt.Email,
        Subject: "Welcome!",
        Body:    fmt.Sprintf("Hi %s, welcome to Promenade!", evt.Name),
    })
}
```

### 4. Initialize in main.go

```go
// cmd/api/main.go
func main() {
    // Load config
    cfg, _ := config.Load()

    // Initialize event bus with configuration
    busConfig := bus.NewBusConfig(
        cfg.IBus.WorkerPoolSize,  // BUS_WORKER_POOL_SIZE=10
        cfg.IBus.BufferSize,      // BUS_BUFFER_SIZE=1000
        cfg.IBus.RetryAttempts,   // BUS_RETRY_ATTEMPTS=3
        cfg.IBus.RetryDelay,      // BUS_RETRY_DELAY=1s
        cfg.IBus.RetryMaxDelay,   // BUS_RETRY_MAX_DELAY=5s
        cfg.IBus.RetryMultiplier, // BUS_RETRY_MULTIPLIER=2.0
    )
    eventBus := memory.NewMemoryBus(busConfig)
    defer eventBus.Close(context.Background())

    // Start notification service with config
    emailSender := notification.NewMockEmailSender()
    emailService, _ := notification.NewEmailService(
        eventBus,
        emailSender,
        "templates/email",        // Template path
        cfg.Email.FromAddress,    // EMAIL_FROM_ADDRESS
        cfg.Email.FromName,       // EMAIL_FROM_NAME
        cfg.Email.AppURL,         // APP_URL
        cfg.Email.AppName,        // APP_NAME
    )
    emailService.Start(context.Background())

    // Pass event bus to use cases
    authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, jwtManager, eventBus)
}
```

## Available Implementations

### 1. In-Memory IBus (Default - Development/Testing)

**Use case:** Development, testing, single-instance deployments

```go
import (
    "github.com/basilex/promenade/pkg/bus"
    "github.com/basilex/promenade/internal/infrastructure/config"
)

// Via factory (recommended)
busConfig := config.BusConfig{
    Adapter:         "memory",
    WorkerPoolSize:  4,
    BufferSize:      100,
    RetryAttempts:   3,
    RetryDelay:      1 * time.Second,
    RetryMaxDelay:   30 * time.Second,
    RetryMultiplier: 2.0,
}
eventBus, err := bus.NewBus(busConfig)
defer eventBus.Close(ctx)
```

**Features:**

- [+] Zero external dependencies
- [+] Perfect for development and testing
- [+] Thread-safe with mutex protection
- [+] Configurable worker pool and buffer
- [+] Health checks and graceful shutdown
- [!] Events lost on restart (no persistence)
- [!] Single-process only (no horizontal scaling)

### 2. Redis Pub/Sub (Production-Ready)

**Use case:** Multi-instance deployments, horizontal scaling, distributed systems

```go
import (
    "github.com/basilex/promenade/pkg/bus"
    "github.com/basilex/promenade/internal/infrastructure/config"
)

// Via factory with Redis config
busConfig := config.BusConfig{
    Adapter:         "redis",
    WorkerPoolSize:  4,
    BufferSize:      100,
    RetryAttempts:   3,
    RetryDelay:      1 * time.Second,
    RetryMaxDelay:   30 * time.Second,
    RetryMultiplier: 2.0,
    Redis: config.RedisConfig{
        Host:     "localhost",  // or redis service name in Docker
        Port:     6379,
        Password: "",           // optional
        DB:       0,
        PoolSize: 10,
    },
}
eventBus, err := bus.NewBus(busConfig) // Auto-falls back to memory on failure
defer eventBus.Close(ctx)
```

**Features:**

- [+] Multi-process support (horizontal scaling)
- [+] Distributed Pub/Sub pattern
- [+] JSON serialization for cross-service compatibility
- [+] Production-ready timeouts (5s dial, 3s read/write, 10s ping)
- [+] Automatic fallback to memory adapter on connection failure
- [+] Health checks via Redis PING
- [!] No guaranteed delivery (subscribers must be online)
- [!] At-most-once semantics (no persistence after delivery)

**Redis Architecture:**

| Component             | Description                                                                     |
| --------------------- | ------------------------------------------------------------------------------- |
| **Redis Server**      | Central message broker on port 6379                                             |
| **Channels (Topics)** | user.registered, user.email_verified, user.password_changed, post.created, etc. |
| **Instance #1**       | Publisher: AuthUC, PostUC publish events                                        |
| **Instance #2**       | Subscriber: EmailService processes events                                       |

**Message Flow:**

1. Instance #1 publishes UserRegisteredEvent to Redis channel
2. Redis broadcasts to all subscribers
3. Instance #2's EmailService receives event and sends welcome email
4. Both instances can publish/subscribe simultaneously
5. Horizontal scaling: Add more instances, all participate in pub/sub

### 3. NATS/Kafka (Future - Microservices)

    // ... redis config

})

````

**Pros:**

- Persistent if configured
- Cross-process communication
- Familiar infrastructure

**Cons:**

- At-most-once delivery
- No message ordering guarantees
- Not for high-throughput

### NATS (TODO)

```go
import "github.com/basilex/promenade/pkg/bus/nats"

bus := nats.NewNATSBus(nats.Config{
    URL: "nats://localhost:4222",
    // ... nats config
})
````

**Pros:**

- High performance
- At-least-once delivery
- Message ordering
- Perfect for microservices

**Cons:**

- Additional infrastructure

## Standard Topics

```go
// User Management
bus.TopicUserRegistered      // "user.registered"
bus.TopicUserActivated       // "user.activated"
bus.TopicUserSuspended       // "user.suspended"
bus.TopicUserBanned          // "user.banned"
bus.TopicUserEmailVerified   // "user.email.verified"
bus.TopicUserPasswordChanged // "user.password.changed"

// Content
bus.TopicPostPublished       // "post.published"
bus.TopicCommentAdded        // "comment.added"

// RBAC
bus.TopicRoleAssigned        // "role.assigned"
bus.TopicRoleRevoked         // "role.revoked"
```

## Testing

```go
func TestUserRegistration(t *testing.T) {
    bus := memory.NewDefaultMemoryBus()
    defer bus.Close(context.Background())

    received := make(chan *event.UserRegisteredEvent, 1)

    handler := func(ctx context.Context, e bus.Event) error {
        received <- e.(*event.UserRegisteredEvent)
        return nil
    }

    bus.Subscribe(bus.TopicUserRegistered, handler)

    // Test your code that publishes events
    // ...

    select {
    case evt := <-received:
        assert.Equal(t, "john@example.com", evt.Email)
    case <-time.After(1 * time.Second):
        t.Fatal("event not received")
    }
}
```

## Best Practices

### [+] DO

1. **Keep handlers idempotent** — events may be delivered more than once
2. **Use event versioning** — include version in event type: `user.registered.v1`
3. **Log publishing errors** — but don't fail the business operation
4. **Use domain events** — not integration events (those come later)
5. **Fail gracefully** — handle handler errors without crashing

### [X] DON'T

1. **Don't put critical logic in events** — events complement, don't replace transactions
2. **Don't expect immediate processing** — it's async by design
3. **Don't publish inside database transactions** — event may fire before commit
4. **Don't overuse** — not everything needs to be async
5. **Don't forget to close the bus** — graceful shutdown is important

## Performance Considerations

- **Publishing:** < 1ms (non-blocking, just queues message)
- **Processing:** Depends on handler (e.g., email: 100-500ms)
- **Throughput:** 10,000+ events/sec with memory bus
- **Memory:** ~1KB per queued event
- **Worker pool:** Default 10 workers, tune based on handler speed

## Migration Path

### Phase 1: Current (Monolith + Async Workers)

- [+] In-memory event bus
- [+] Email/notification workers in same binary
- [+] Single database, ACID preserved
- [+] Fast deployment, simple ops

### Phase 2: Add Persistence (Still Monolith)

- [*] Swap to Redis Pub/Sub
- [*] Events survive restarts
- [*] Still single binary
- [*] Prepare for scale-out

### Phase 3: Extract Services (Microservices)

- [*] Move email service to separate binary
- [*] Use NATS/Kafka for cross-service communication
- [*] API Gateway pattern
- [*] Gradual extraction of other services

## FAQ

**Q: When should I use the event bus vs direct calls?**
A: Use events for non-critical operations (emails, analytics, logging). Use direct calls for critical business logic.

**Q: What happens if a handler fails?**
A: With the memory bus, the error is logged and the event is dropped. With Redis/NATS, you can configure retries.

**Q: Can I have multiple handlers for the same event?**
A: Yes! That's the point. Multiple workers can subscribe to the same topic.

**Q: How do I test code that publishes events?**
A: Use the memory bus in tests and subscribe to events to verify they were published with correct data.

**Q: Should I publish events before or after database commit?**
A: After. Otherwise, the event might be processed before the data is visible.

## See Also

- [Notification Service](../../internal/infrastructure/notification/) — Email/SMS via event bus
- [Domain Events](../../internal/domain/event/) — All available events
- [Auth UseCase](../../internal/usecase/auth_usecase.go) — Example of event publishing
