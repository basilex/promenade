# Event-Driven Architecture

**Event-Driven Architecture (EDA)** enables asynchronous, decoupled communication between Bounded Contexts using the **Publish-Subscribe pattern** via Event Bus.

---

## Core Concept

**Definition**: System components communicate by publishing and subscribing to domain events rather than direct method calls.

**Key Principle**: **Temporal Decoupling** - Publisher doesn't know when or if subscriber processes the event. Subscriber doesn't know who published it.

---

## Why Event-Driven?

### Problems It Solves

**Without Events** (direct calls):

```go
// Tight coupling - UserService knows about CustomerService, BillingService
func (s *UserService) Register(email, password string) error {
    user := s.userRepo.Create(email, password)
    
    // Direct dependency - blocks if service is down
    s.customerService.CreateCustomer(user.ID)  // BLOCKS
    s.billingService.CreateAccount(user.ID)    // BLOCKS
    
    return nil
}
```

**Problems**:
- ❌ Tight coupling (3 services in one method)
- ❌ Synchronous - registration blocks waiting for customer/billing
- ❌ Cascade failures - if billing is down, registration fails
- ❌ Hard to add new features (need to modify UserService)

**With Events**:

```go
// Decoupled - UserService only publishes event
func (s *UserService) Register(email, password string) error {
    user := s.userRepo.Create(email, password)
    
    // Publish event (fire-and-forget)
    event := bus.NewBaseEvent("user.registered", user.ID)
    s.eventBus.Publish(ctx, bus.TopicUserRegistered, event)
    
    return nil  // Done! No waiting
}

// Customer Management listens
func (h *CustomerHandler) HandleUserRegistered(ctx context.Context, e bus.Event) error {
    customer := customer.NewCustomer(e.AggregateID())
    return h.customerRepo.Create(ctx, customer)
}

// Billing also listens
func (h *BillingHandler) HandleUserRegistered(ctx context.Context, e bus.Event) error {
    account := account.NewAccount(e.AggregateID())
    return h.accountRepo.Create(ctx, account)
}
```

**Benefits**:
- ✅ Loose coupling - services don't know about each other
- ✅ Asynchronous - registration returns immediately
- ✅ Resilient - if billing is down, retry policy handles it
- ✅ Extensible - add new listener without changing publisher

---

## Promenade Event Bus

### Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Event Bus (Abstraction)                │
│                                                           │
│  Interface: Publish, Subscribe, Unsubscribe, Close       │
│  Events: BaseEvent (type, aggregateID, timestamp)        │
│  Handlers: EventHandler func(ctx, event) error           │
└───────────────────┬──────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
┌───────────────┐       ┌───────────────┐
│Memory Adapter │       │Redis Adapter  │
│               │       │               │
│In-process     │       │Distributed    │
│377K ev/sec    │       │Persistent     │
│Dev/Test       │       │Production     │
└───────────────┘       └───────────────┘
```

**Factory Pattern**: `bus.NewBus(config)` creates adapter based on config

---

### Memory Adapter (Development)

**Characteristics**:

- **Ultra-fast**: 377,494 events/sec (no network overhead)
- **Simple**: No external dependencies
- **In-process**: Events don't cross application boundaries
- **Non-persistent**: Events lost on restart

**Configuration**:

```yaml
# config/app.postgres-dev.yaml or config/app.sqlite-dev.yaml
bus:
  adapter: "memory"
  worker_pool_size: 10
  buffer_size: 1000
  retry_attempts: 3
  retry_delay: 100ms
```

**When to use**:

- Local development
- Unit/integration tests
- Single-server deployments
- Non-critical event processing

---

### Redis Adapter (Production)

**Characteristics**:

- **Distributed**: Events propagate across all application instances
- **Persistent**: Can use Redis Streams for event storage
- **Scalable**: Handles thousands of subscribers
- **Reliable**: Automatic reconnection on network failures

**Configuration**:

```yaml
# config/app.postgres-prod.yaml
database:
  redis:
    addr: "redis.prod.example.com:6379"
    password: "${REDIS_PASSWORD}"
    databases:
      bus: 1  # Event Bus database

bus:
  adapter: "redis"
  worker_pool_size: 50
  buffer_size: 10000
  retry_attempts: 5
  retry_delay: 1s
```

**When to use**:

- Production environments
- Multiple application instances
- Microservices architecture
- Critical event processing requiring persistence

---

## Domain Events

### Event Structure

```go
// Event interface
type Event interface {
    Type() string                    // "user.registered"
    AggregateID() uuid.UUID          // User ID
    OccurredAt() time.Time           // When event happened
    Metadata() map[string]interface{} // Additional data
}

// BaseEvent implementation
type BaseEvent struct {
    EventType      string
    EventAggregateID uuid.UUID
    EventOccurredAt time.Time
    EventMetadata   map[string]interface{}
}
```

### Event Naming Convention

**Pattern**: `{context}.{aggregate}.{action}` (lowercase, dot-separated)

**Examples**:

- `identity.user.registered` - User registered in Identity context
- `identity.contact.verified` - Contact verified in Identity context
- `customer.deal.won` - Deal won in Customer Management context
- `order.order.shipped` - Order shipped in Order Management context
- `billing.payment.received` - Payment received in Billing context

**Rules**:

- Context name first (identity, customer, order, billing)
- Aggregate name second (user, contact, customer, order)
- Past tense action (registered, verified, shipped, not register, verify, ship)

---

## Topic Constants

**All topics are predefined constants** to prevent typos:

```go
// pkg/bus/topics.go

// Identity Context
const (
    TopicUserRegistered      = "identity.user.registered"
    TopicUserActivated       = "identity.user.activated"
    TopicUserSuspended       = "identity.user.suspended"
    TopicContactVerified     = "identity.contact.verified"
)

// Customer Management Context
const (
    TopicCustomerCreated = "customer.customer.created"
    TopicDealWon         = "customer.deal.won"
)

// Order Management Context
const (
    TopicOrderCreated   = "order.order.created"
    TopicOrderShipped   = "order.order.shipped"
)
```

**Benefits**:

- IDE autocomplete
- Compile-time safety
- Easy refactoring
- Documentation in code

---

## Publishing Events

### In Use Case Layer

```go
// internal/contexts/identity/user/usecase.go

func (uc *useCase) Register(ctx context.Context, email, name, password string) (*User, error) {
    // Business logic
    user, err := NewUser(email, name, password)
    if err != nil {
        return nil, err
    }
    
    // Persist
    if err := uc.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // Publish event (after successful commit)
    event := bus.NewBaseEvent("identity.user.registered", user.ID)
    event.EventMetadata = map[string]interface{}{
        "email": user.Email.Value(),
        "name":  user.Name,
    }
    
    if err := uc.eventBus.Publish(ctx, bus.TopicUserRegistered, event); err != nil {
        // Log error but don't fail operation
        logger.FromContext(ctx).Error("Failed to publish event",
            slog.Any("error", err),
            slog.String("event_type", "user.registered"),
        )
    }
    
    return user, nil
}
```

**Key Points**:

- Publish AFTER successful database commit
- Log publish errors but don't fail operation
- Include relevant metadata (email, name)
- Use topic constants (`bus.TopicUserRegistered`)

---

## Subscribing to Events

### Handler Pattern

```go
// internal/contexts/customer-mgmt/customer/subscriber.go

type CustomerSubscriber struct {
    customerUC customer.IUseCase
}

func NewCustomerSubscriber(uc customer.IUseCase) *CustomerSubscriber {
    return &CustomerSubscriber{customerUC: uc}
}

// Register subscriptions
func (s *CustomerSubscriber) RegisterSubscribers(eventBus bus.EventBus) error {
    // Listen to user registration events
    if err := eventBus.Subscribe(bus.TopicUserRegistered, s.HandleUserRegistered); err != nil {
        return fmt.Errorf("failed to subscribe to user.registered: %w", err)
    }
    
    return nil
}

// Handle user.registered event
func (s *CustomerSubscriber) HandleUserRegistered(ctx context.Context, e bus.Event) error {
    userID := e.AggregateID()
    
    // Extract metadata
    metadata := e.Metadata()
    email := metadata["email"].(string)
    
    // Create customer
    customer, err := customer.NewCustomer(userID, email)
    if err != nil {
        return fmt.Errorf("failed to create customer: %w", err)
    }
    
    if err := s.customerUC.CreateCustomer(ctx, customer); err != nil {
        return fmt.Errorf("failed to save customer: %w", err)
    }
    
    logger.FromContext(ctx).Info("Customer created from user registration",
        slog.String("user_id", userID.String()),
        slog.String("customer_id", customer.ID.String()),
    )
    
    return nil
}
```

**Key Points**:

- Subscribe in dedicated subscriber struct
- Return error for retry (temporary failures)
- Return nil to skip retry (permanent errors)
- Log success/failure for monitoring

---

## Retry Policy

### Exponential Backoff

**Configuration**:

```go
config := bus.NewConfig(
    10,                     // workerPoolSize
    1000,                   // bufferSize
    3,                      // maxAttempts
    100*time.Millisecond,   // initialDelay
    10*time.Second,         // maxDelay
    2.0,                    // multiplier
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
    // Temporary error - will retry
    if err := h.service.Call(); err != nil {
        if isTemporaryError(err) {
            return err  // Retry
        }
        
        // Permanent error - log and skip
        logger.FromContext(ctx).Error("Permanent error", slog.Any("error", err))
        return nil  // Don't retry
    }
    
    return nil
}
```

**Best Practice**: Return error for retryable failures, nil for permanent errors.

---

## Panic Recovery

**All handlers are wrapped in panic recovery**:

```go
defer func() {
    if r := recover(); r != nil {
        logger.FromContext(ctx).Error("Event handler panicked",
            slog.Any("panic", r),
            slog.String("event_type", event.Type()),
        )
    }
}()
```

**Result**: One handler panic doesn't crash the application or affect other handlers.

---

## Event Sourcing (Future)

**Event Sourcing** stores all events for later replay:

```go
// Save event to event store
eventStore.Append(event)

// Rebuild aggregate from events
events := eventStore.GetEvents(aggregateID)
aggregate := replayEvents(events)
```

**Benefits**:

- Complete audit trail
- Time travel (state at any point)
- Event replay for new projections
- Debugging (reproduce exact state)

**Status**: Planned for Q2 2026

---

## Sagas (Distributed Transactions)

**Saga** coordinates long-running transactions across contexts:

```go
type OrderFulfillmentSaga struct {
    saga.BaseSaga
    OrderID     uuid.UUID
    CustomerID  uuid.UUID
    PaymentID   uuid.UUID
}

func (s *OrderFulfillmentSaga) Start(ctx context.Context) error {
    // Step 1: Charge payment
    if err := s.chargePayment(ctx); err != nil {
        return err
    }
    
    // Step 2: Reserve inventory
    if err := s.reserveInventory(ctx); err != nil {
        s.refundPayment(ctx)  // Compensate
        return err
    }
    
    // Step 3: Ship order
    if err := s.shipOrder(ctx); err != nil {
        s.releaseInventory(ctx)  // Compensate
        s.refundPayment(ctx)     // Compensate
        return err
    }
    
    return nil
}
```

**Saga Documentation**: [pkg/saga/README.md](../../pkg/saga/README.md)

---

## Best Practices

### DO ✅

- **Publish after commit** to ensure consistency
- **Use topic constants** to prevent typos
- **Keep events small** - only essential data
- **Version events** if schema changes
- **Make handlers idempotent** - safe to retry
- **Log publish errors** but don't fail operation
- **Return nil for permanent errors** to skip retry

### DON'T ❌

- **DON'T publish before commit** (risk inconsistency)
- **DON'T hardcode topic strings**
- **DON'T include large payloads** (use ID + metadata)
- **DON'T panic in handlers** (use error returns)
- **DON'T block in handlers** (use goroutines for slow operations)
- **DON'T ignore context cancellation**
- **DON'T use events for synchronous RPC** (use HTTP)

---

## Performance Metrics

### Memory Adapter

**Throughput**: 377,494 events/sec  
**Latency**:

- Publish: ~0.1μs
- Subscribe: ~0.2μs
- Handler execution: ~10μs (empty handler)

**Load Test** (1000 events):

- Total time: 2.6ms
- Memory usage: ~1MB

### Redis Adapter

**Throughput**: ~1,000 events/sec (network-bound)  
**Latency**:

- Publish: ~50-100ms (Redis round-trip)
- Subscribe: Real-time (Redis Pub/Sub)

**Load Test** (50 concurrent events):

- Total time: ~2 seconds
- Memory usage: ~5MB

---

## Real-World Example: Order Fulfillment

### Multi-Context Orchestration

```
1. Order Context: order.created
   ↓
2. Billing Context: Charge payment → payment.charged
   ↓
3. Warehouse Context: Reserve inventory → inventory.reserved
   ↓
4. Order Context: Mark confirmed → order.confirmed
   ↓
5. Warehouse Context: Ship → order.shipped
   ↓
6. Order Context: Mark shipped → order.completed
```

**Events Published**:

1. `order.order.created` (Order Context)
2. `billing.payment.charged` (Billing Context)
3. `warehouse.inventory.reserved` (Warehouse Context)
4. `order.order.confirmed` (Order Context)
5. `order.order.shipped` (Order Context)

**Subscribers**:

- Billing listens to `order.created`
- Warehouse listens to `payment.charged`
- Order listens to `inventory.reserved` and `order.shipped`

---

## Debugging

### Enable Debug Logging

```yaml
# config/app.postgres-dev.yaml or config/app.sqlite-dev.yaml
logging:
  level: "debug"
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
if err := eventBus.Health(ctx); err != nil {
    log.Error("Event bus unhealthy:", err)
}
```

---

## Related Documentation

- [Event Bus Implementation](../../pkg/bus/README.md) - Complete technical reference
- [Bounded Contexts Strategy](bounded-contexts.md) - Context communication patterns
- [Saga Pattern](../../pkg/saga/README.md) - Distributed transactions
- [Clean Architecture](clean-architecture.md) - Overall architecture

---

**Last Updated**: December 29, 2025  
**Status**: Production-ready (67 tests, 100% passing)  
**Performance**: 377K events/sec (Memory), ~1K events/sec (Redis)  
**Maintainer**: Promenade Team
