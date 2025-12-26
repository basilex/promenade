---
title: "Event-Driven Architecture"
description: "Dual event bus adapters for async communication"
weight: 6
---

## Event Bus System

Promenade includes a **dual-adapter event bus** for asynchronous, event-driven communication.

### Two Adapters

**Memory Adapter** (default):

- In-memory Pub/Sub
- Goroutines + channels
- Perfect for dev/test
- Zero dependencies

**Redis Adapter** (production):

- Distributed Pub/Sub
- Multi-instance support
- Persistent message queue
- Fault-tolerant

### Configuration

```yaml
# config/app.yaml
bus:
  adapter: "memory" # or "redis"
  worker_pool_size: 4

  # Redis-specific
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
```

### Publishing Events

```go
// Define event
type UserRegisteredEvent struct {
    bus.BaseEvent
    UserID string
    Email  string
}

// Publish
event := &UserRegisteredEvent{
    BaseEvent: bus.BaseEvent{ID: uuid.New().String()},
    UserID:    user.ID,
    Email:     user.Email,
}
eventBus.Publish(ctx, "user.registered", event)
```

### Subscribing to Events

```go
// Subscribe in module initialization
func (m *EmailModule) Initialize(ctx context.Context, core *module.Core) error {
    return core.EventBus.Subscribe(ctx, "user.registered",
        func(ctx context.Context, e bus.Event) error {
            evt := e.(*UserRegisteredEvent)
            return m.sendWelcomeEmail(ctx, evt.Email)
        },
    )
}
```

### Built-in Events

- `user.registered` - New user signup
- `user.deleted` - User account removed
- `post.created` - New post published
- `comment.created` - New comment added
- `purge.completed` - Purge job finished

### Module Communication

Modules use events for **loose coupling**:

```
Posts Module                Email Module
    |                           |
    | user.registered           |
    |-------------------------->|
    |                      Send welcome email
    |                           |
```

**Benefits:**

- No direct dependencies between modules
- Async processing
- Easy to add new subscribers
- Production-ready with Redis

### Graceful Fallback

```go
// Try Redis, fallback to Memory
eventBus, err := bus.NewBus(cfg.Bus)
if err != nil {
    logger.Warn("Redis unavailable, using memory bus")
    eventBus = bus.NewMemoryBus()
}
```

### Benefits

✅ **Flexible** - Switch adapters via config  
✅ **Async** - Non-blocking event processing  
✅ **Scalable** - Redis for multi-instance  
✅ **Simple** - Unified interface for both adapters

[Event Bus Guide →](/promenade/docs/pkg/bus/README)
