# Architecture Overview

Promenade follows **strict Domain-Driven Design** principles with **Bounded Context** separation and **Event-Driven Architecture** at its core.

## Core Principles

### 1. Domain-Driven Design (DDD)

**Bounded Contexts** - Each context is autonomous with own models and database schema:
- **Identity** - User, Contact, Profile aggregates
- **Customer Management** - Customer, Company, Deal aggregates
- **Order Management** - Order, OrderItem, Fulfillment aggregates
- **Billing** - Invoice, Payment, Subscription aggregates

**Aggregates** - Business entities with invariants and transactional boundaries:
```go
type Contact struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    Type       ContactType
    Email      *valueobject.Email  // Value object
    Phone      *valueobject.Phone  // Value object
    IsPrimary  bool
    IsVerified bool
}

func NewEmailContact(userID uuid.UUID, email, label string) (*Contact, error) {
    emailVO, err := valueobject.NewEmail(email)
    if err != nil {
        return nil, fmt.Errorf("invalid email: %w", err)
    }
    return &Contact{
        ID:     uuidv7.New(),
        UserID: userID,
        Email:  &emailVO,
    }, nil
}
```

**Value Objects** - Immutable domain concepts:
```go
type Email struct {
    value string
}

func NewEmail(email string) (Email, error) {
    trimmed := strings.TrimSpace(strings.ToLower(email))
    if !emailRegex.MatchString(trimmed) {
        return Email{}, fmt.Errorf("invalid email format")
    }
    return Email{value: trimmed}, nil
}
```

### 2. Event-Driven Architecture

**Event Bus** - Central communication hub:
- **Memory Adapter** - 377K events/sec, zero dependencies (dev/test)
- **Redis Adapter** - Distributed, persistent (production)
- **Retry Policy** - Exponential backoff with panic recovery
- **Graceful Shutdown** - Wait for in-flight events

Example:
```go
// Publish domain event
event := bus.NewBaseEvent("contact.verified", contactID)
eventBus.Publish(ctx, bus.TopicContactVerified, event)

// Subscribe to events
eventBus.Subscribe(bus.TopicContactVerified, func(ctx context.Context, e bus.Event) error {
    // Handle contact verification
    return notificationService.SendVerificationEmail(ctx, e.AggregateID())
})
```

### 3. Clean Architecture Layers

```

           HTTP Handlers (Gin)            ← Adapter Layer

          Use Cases (Business Logic)      ← Application Layer

    Aggregates + Value Objects (Domain)   ← Domain Layer

   Repository (PostgreSQL + Redis)        ← Infrastructure Layer

```

**Dependency Rule**: Inner layers never depend on outer layers.

## Bounded Contexts Structure

Each context follows the same pattern:

```
internal/contexts/{context}/
 {aggregate}/
    entity.go              # Domain entity (aggregate root)
    entity_test.go         # Entity tests
    repository.go          # Repository interface
    usecase.go            # Business logic (use cases)
    usecase_test.go       # Use case tests
    adapter/
        http/
           handler/
               {aggregate}_handler.go  # HTTP handlers
               dto/
                   {aggregate}_dto.go  # Data transfer objects
        repository/postgres/
            base_repository.go         # BaseRepository (per context)
            {aggregate}_repository.go  # PostgreSQL implementation
 router.go                  # Context router (registers routes)
```

### Example: Identity Context

```go
// internal/contexts/identity/contact/entity.go
type Contact struct {
    aggregate.BaseAggregate
    ID         uuid.UUID
    UserID     uuid.UUID
    Type       ContactType
    Email      *valueobject.Email
    IsPrimary  bool
    IsVerified bool
}

// internal/contexts/identity/contact/repository.go
type IRepository interface {
    Create(ctx context.Context, contact *Contact) error
    GetByID(ctx context.Context, id uuid.UUID) (*Contact, error)
    Update(ctx context.Context, contact *Contact) error
}

// internal/contexts/identity/contact/usecase.go
type IUseCase interface {
    CreateEmailContact(ctx context.Context, userID uuid.UUID, email, label string) (*Contact, error)
    VerifyContact(ctx context.Context, contactID uuid.UUID) error
}

// internal/contexts/identity/router.go
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
    identity := api.Group("/identity")
    contacts := identity.Group("/contacts")
    contacts.Use(jwt.AuthMiddleware(r.jwtManager))
    {
        contacts.POST("", r.contactHandler.Create)
        contacts.GET("/:id", r.contactHandler.GetByID)
    }
}
```

## Context Communication

**Rule**: Contexts communicate ONLY via Event Bus (no direct dependencies).

```go
// User Context publishes event
event := bus.NewBaseEvent("user.registered", userID)
eventBus.Publish(ctx, bus.TopicUserRegistered, event)

// Customer Context subscribes
eventBus.Subscribe(bus.TopicUserRegistered, func(ctx context.Context, e bus.Event) error {
    // Create customer profile automatically
    return customerUC.CreateFromUser(ctx, e.AggregateID())
})
```

## Database Strategy

### UUID v7 (Time-Ordered)

**2x faster inserts** than UUID v4 due to natural ordering:

```go
id := uuidv7.New()  // Time-ordered UUID
```

### No ORM

Raw SQL with **sqlx** + **BaseRepository** pattern:

```go
type BaseRepository struct {
    db *sqlx.DB
    tm *database.TransactionManager
}

func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
    executor := r.getExecutor(ctx)  // Auto-detects transaction
    return executor.GetContext(ctx, dest, query, args...)
}
```

### Soft Delete

Always include in SELECT queries:

```sql
SELECT * FROM identity_contacts 
WHERE id = $1 AND deleted_at IS NULL
```

## Authentication & Authorization

### JWT Tokens

- **Access Token**: 15 minutes (API requests)
- **Refresh Token**: 7 days (token renewal)
- **Token Revocation**: Redis-backed blacklist

```go
// Generate tokens
tokenPair, err := jwtManager.GenerateTokenPair(userID, email, roles)

// Validate token
claims, err := jwtManager.ValidateAccessToken(token)

// Revoke token
tokenRevoker.Revoke(ctx, token, claims.ExpiresAt)
```

### RBAC (Role-Based Access Control)

5 system roles with 29+ permissions:

```go
// Protect routes
admin := api.Group("/admin")
admin.Use(jwt.AuthMiddleware(jwtManager))
admin.Use(jwt.RequireRole("admin"))
{
    admin.POST("/users", handler.CreateUser)
}
```

## Testing Strategy

### Three-Tier Testing

1. **Unit Tests** (in-place) - Fast feedback, test individual components
2. **Smoke Tests** (`test/smoke/`) - Mock-based handler validation, no DB
3. **Integration Tests** (`test/integration/`) - Full E2E with real DB

```bash
make test              # All tests (240+ tests, ~40s)
make test-unit         # Unit tests (~5s)
make test-smoke        # Smoke tests (~0.3s)
make test-integration  # Integration tests (~5s)
```

## Configuration

Environment-specific YAML configs with env var overrides:

```yaml
# config/app.postgres-dev.yaml or config/app.sqlite-dev.yaml
app:
  name: "Promenade Platform"
  environment: "development"

database:
  postgres:
    host: "localhost"
    database: "promenade_dev"

bus:
  adapter: "memory"  # Fast in-process adapter
```

```yaml
# config/app.postgres-prod.yaml
database:
  postgres:
    host: "${DB_HOST}"           # Environment variable
    password: "${DB_PASSWORD}"

bus:
  adapter: "redis"  # Distributed adapter
```

## Next Steps

- [Testing Guide](/guide/testing) - Write professional tests
- [Identity Context](/contexts/identity) - Explore User & Contact aggregates
- [Event Bus](/packages/bus) - Deep dive into event-driven patterns
- [Contributing](/guide/contributing) - Join the project
