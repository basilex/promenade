# Event Bus Package

Universal message bus implementation for event-driven architecture in Promenade.

## Overview

The `pkg/bus` package provides a transport-agnostic event bus that enables asynchronous, decoupled communication between different parts of the application. This is the foundation for evolving from a monolith to microservices without changing business logic.

## Architecture Benefits

### Current (Monolith + Async)

```
┌─────────────────────────────────────────┐
│         Single Binary                    │
│                                          │
│  ┌──────────┐                            │
│  │ UseCase  │──publish──┐                │
│  └──────────┘           │                │
│                         ▼                │
│                    ┌─────────┐           │
│                    │Event Bus│           │
│                    └────┬────┘           │
│                         │                │
│       ┌─────────────────┼────────────┐   │
│       │                 │            │   │
│  ┌────▼─────┐   ┌───────▼──┐  ┌─────▼┐  │
│  │  Email   │   │Analytics │  │ Logs │  │
│  │ Worker   │   │  Worker  │  │Worker│  │
│  └──────────┘   └──────────┘  └──────┘  │
│                                          │
│  Single Database (ACID preserved)       │
└─────────────────────────────────────────┘
```

### Future (Microservices)

```
┌─────────────┐    ┌──────────────┐
│   Gateway   │───▶│  NATS/Kafka  │
└─────────────┘    └──────┬───────┘
                          │
        ┌─────────────────┼──────────────┐
        │                 │              │
   ┌────▼────┐      ┌─────▼──────┐  ┌───▼───┐
   │ Email   │      │ Analytics  │  │ User  │
   │ Service │      │  Service   │  │Service│
   └─────────┘      └────────────┘  └───────┘
```

**Same code, different transport!** Just swap `memory.Bus` → `redis.Bus` or `nats.Bus`

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
    bus.BaseEvent
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
    Name   string    `json:"name"`
}

func NewUserRegisteredEvent(userID uuid.UUID, email, name string) *UserRegisteredEvent {
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
    // Initialize event bus (in-memory для начала)
    eventBus := memory.NewDefaultMemoryBus()
    defer eventBus.Close(context.Background())

    // Start notification service
    emailSender := notification.NewMockEmailSender()
    emailService := notification.NewEmailService(eventBus, emailSender)
    emailService.Start(context.Background())

    // Pass event bus to use cases
    authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, jwtManager, eventBus)
}
```

## Available Implementations

### In-Memory Bus (Development/Simple Deployments)

```go
import "github.com/basilex/promenade/pkg/bus/memory"

bus := memory.NewDefaultMemoryBus()
// или с кастомной конфигурацией
bus := memory.NewMemoryBus(bus.BusConfig{
    WorkerPoolSize: 20,
    BufferSize:     5000,
})
```

**Pros:**

- Zero dependencies
- Perfect for development
- Simple single-instance deployments
- Fast tests

**Cons:**

- Events lost on restart
- No cross-process communication

### Redis Pub/Sub (TODO)

```go
import "github.com/basilex/promenade/pkg/bus/redis"

bus := redis.NewRedisBus(redis.Config{
    Address: "localhost:6379",
    // ... redis config
})
```

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
```

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
