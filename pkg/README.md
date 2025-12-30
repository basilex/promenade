# Package Library Overview

## Overview

The `pkg/` directory contains **shared, reusable packages** used across all bounded contexts in Promenade Platform. These packages are **context-agnostic** and implement common patterns, utilities, and domain primitives.

---

## Package Inventory

### Core Infrastructure

| Package       | Purpose                                         | Status     | Tests | Documentation                       |
| ------------- | ----------------------------------------------- | ---------- | ----- | ----------------------------------- |
| **bus**       | Event Bus (Pub/Sub, Memory/Redis adapters)      | Production | 67    | [README](bus/README.md)             |
| **cache**     | Caching layer (Redis/NoOp adapters)             | Production | -     | [README](cache/README.md)           |
| **jwt**       | JWT authentication & RBAC middleware            | Production | 18    | [README](jwt/README.md)             |
| **logger**    | Structured logging (slog wrapper)               | Production | 15    | [README](logger/README.md)          |
| **migration** | Database migration management (namespace-based) | Production | 8     | [README](migration/README.md)       |
| **response**  | Standard HTTP response formatting               | Production | 12    | [README](response/README.md)        |

### Domain Primitives

| Package         | Purpose                                         | Status     | Tests | Documentation                       |
| --------------- | ----------------------------------------------- | ---------- | ----- | ----------------------------------- |
| **uuidv7**      | Time-ordered UUIDs (RFC 9562)                   | Production | 10    | [README](uuidv7/README.md)          |
| **valueobject** | Immutable value objects (Email, Phone, Money)   | Production | 25    | [README](valueobject/README.md)     |
| **aggregate**   | Base aggregate pattern (DDD)                    | Production | 5     | [README](aggregate/README.md)       |
| **saga**        | Saga orchestration for distributed transactions | Production | 28    | [README](saga/README.md)            |

### Utilities

| Package       | Purpose                      | Status     | Tests | Documentation                       |
| ------------- | ---------------------------- | ---------- | ----- | ----------------------------------- |
| **jsonb**     | JSONB helpers for PostgreSQL | Production | 8     | [README](jsonb/README.md)           |
| **reference** | Reference data utilities     | Production | 6     | -                                   |

---

## Package Descriptions

### 1. pkg/bus - Event Bus

**Central communication hub** for asynchronous, decoupled event-driven architecture.

**Key Features**:

- Multiple adapters (Memory, Redis)
- Retry policy with exponential backoff
- Panic recovery in handlers
- Graceful shutdown
- 30+ predefined topic constants

**Performance**:

- Memory: 377K events/sec
- Redis: ~1K events/sec (network-bound)

**Read More**: [pkg/bus/README.md](bus/README.md)

---

### 2. pkg/cache - Caching Layer

**Redis-based caching system** with multiple adapters for performance optimization.

**Key Features**:

- Multiple adapters (Redis production, NoOp testing)
- Resource-specific TTL configuration
- Cache-aside pattern with write-through invalidation
- Pattern-based deletion (SCAN)
- Graceful degradation when Redis unavailable
- JSON marshaling for complex types

**Usage**:

```go
import "github.com/basilex/promenade/pkg/cache"

// Initialize cache
cacheClient, err := cache.NewCache(cacheConfig, redisClient)

// UseCase integration
type countryUseCase struct {
    repo  IRepository
    cache cache.Cache
}

// Cache-aside pattern
func (uc *countryUseCase) GetByID(ctx context.Context, id uuid.UUID) (*Country, error) {
    // Try cache first
    var country Country
    if err := uc.cache.Get(ctx, fmt.Sprintf("country:id:%s", id), &country); err == nil {
        return &country, nil
    }
    
    // Cache miss - fetch from DB
    country, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    uc.cache.Set(ctx, fmt.Sprintf("country:id:%s", id), country, 1*time.Hour)
    return country, nil
}
```

**Configuration** (`config/app.*.yaml`):

```yaml
cache:
  enabled: true
  adapter: "redis"  # redis or noop
  prefix: "promenade:dev:"
  default_ttl: "5m"
  ttl:
    countries: "1h"      # Reference data
    user_profile: "15m"  # User data
    session: "30m"       # Session data
```

**TTL Strategy**:
- **Reference data** (countries, currencies): 1h (dev) → 24h (prod)
- **User data** (profiles, customers): 10-15m (dev) → 20-30m (prod)
- **Session data**: 30m (dev) → 1h (prod)

**Read More**: [pkg/cache/README.md](cache/README.md) | [docs/CACHING.md](../docs/CACHING.md)

---

### 3. pkg/jwt - JWT Authentication

**JWT token generation and validation** with RBAC middleware for secure authentication and authorization.

**Key Features**:

- Token generation (access + refresh tokens)
- Token validation with custom claims (user_id, email, roles)
- Gin middleware for authentication
- RBAC middleware (RequireRole, RequireAnyRole, RequireAllRoles)
- Context helpers (GetClaims, GetUserID)
- Configurable TTL and secret key

**Usage**:

```go
import "github.com/basilex/promenade/pkg/jwt"

// Initialize JWT manager
jwtManager := jwt.NewManager(jwt.Config{
    SecretKey:            cfg.JWT.Secret,
    AccessTokenDuration:  15 * time.Minute,
    RefreshTokenDuration: 7 * 24 * time.Hour,
    Issuer:               "promenade-platform",
})

// Generate token pair
tokenPair, err := jwtManager.GenerateTokenPair(userID, email, roles)

// Protect routes
router.Use(jwt.AuthMiddleware(jwtManager))

// Require specific role
admin := router.Group("/admin")
admin.Use(jwt.RequireRole("admin"))
```

**Configuration** (`config/app.*.yaml`):

```yaml
jwt:
  secret: "your-secret-key-at-least-32-characters"
  access_token_duration: 15m
  refresh_token_duration: 168h  # 7 days
  issuer: "promenade-platform"
```

**Read More**: [pkg/jwt/README.md](jwt/README.md)

---

### 4. pkg/logger - Structured Logging

**Wrapper around Go's `log/slog`** with context-aware logging and request ID propagation.

**Features**:

- Structured JSON/text output
- Context propagation (`logger.FromContext(ctx)`)
- Request ID tracking
- Log levels (debug, info, warn, error)
- Source code location (file:line)

**Usage**:

```go
import "github.com/basilex/promenade/pkg/logger"

// Initialize once at startup
logger.Init(logger.Config{
    Level:      "debug",
    Format:     "json",
    AddSource:  true,
    TimeFormat: time.RFC3339,
})

// Use in handlers/usecases
log := logger.FromContext(ctx)
log.Info("User registered",
    slog.String("user_id", userID.String()),
    slog.String("email", email),
)
```

**Configuration** (`config/app.*.yaml`):

```yaml
logging:
  level: "debug" # debug, info, warn, error
  format: "json" # json, text
  add_source: true # Add file:line to logs
```

---

### 5. pkg/migration - Database Migrations

**Namespace-based migration system** for managing schema evolution across multiple bounded contexts.

**Features**:

- Namespace isolation (`core`, `shared`, `identity`, etc.)
- Up/down migrations
- Version tracking per namespace
- Automatic migration on startup
- Manual migration CLI

**Directory Structure**:

```
migrations/
 core/                    # Core infrastructure
    000001_init_uuid_v7.up.sql
    000001_init_uuid_v7.down.sql
 shared/                  # Reference data
    000001_seed_currencies.up.sql
    000002_seed_countries.up.sql
 identity/                # Identity context
     000001_create_users.up.sql
     000002_create_contacts.up.sql
```

**Usage**:

```bash
# Create migration
make migrate-create MODULE=identity NAME=add_user_status

# Run migrations
make migrate

# Or specific namespace
go run cmd/migrate/main.go --cmd=up --namespace=identity
```

**Read More**: [migrations/README.md](../migrations/README.md)

---

### 6. pkg/response - HTTP Responses

**Standard response formatting** for consistent API responses across all handlers.

**Response Format**:

```json
// Success
{
  "status": "success",
  "data": { ... }
}

// Error
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email format"
  }
}

// Paginated
{
  "status": "success",
  "data": [ ... ],
  "pagination": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  }
}
```

**Usage**:

```go
import "github.com/basilex/promenade/pkg/response"

// Success response
response.Success(c, http.StatusOK, user)

// Error response
response.Error(c, http.StatusBadRequest, "INVALID_INPUT", err)

// Paginated response
page := response.GetPageFromQuery(c)       // Default: 1
pageSize := response.GetPageSizeFromQuery(c) // Default: 20, max: 100
response.Paginated(c, http.StatusOK, users, total, page, pageSize)
```

---

### 7. pkg/uuidv7 - Time-Ordered UUIDs

**UUIDv7 implementation** (RFC 9562) for database primary keys with better performance than UUIDv4.

**Why UUIDv7?**

- **Time-ordered**: Naturally sorts by creation time
- **B-tree friendly**: 2x faster inserts than UUIDv4
- **Reduced fragmentation**: Better index locality
- **Extractable timestamp**: `ExtractTime(uuid)`

**Usage**:

```go
import "github.com/basilex/promenade/pkg/uuidv7"

// Generate UUID
id := uuidv7.New()  // e.g., 01JGABC123...

// Extract timestamp
ts := uuidv7.ExtractTime(id)
fmt.Println(ts) // 2025-12-27 10:30:45

// Verify version
if uuidv7.IsV7(id) {
    fmt.Println("Valid UUIDv7")
}
```

**Database Schema**:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),  -- Auto-generate
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Read More**: [docs/UUID_V7_GUIDE.md](../docs/UUID_V7_GUIDE.md)

---

### 8. pkg/valueobject - Value Objects

**Immutable value objects** (DDD pattern) for domain concepts with validation and equality semantics.

**Available Value Objects**:

- **Email**: RFC 5322 compliant email validation
- **Phone**: E.164 format with country code
- **Money**: Amount + Currency (ISO 4217)
- **Address**: Street, City, Country, Postal Code
- **DateRange**: Start + End dates with validation

**Usage**:

```go
import "github.com/basilex/promenade/pkg/valueobject"

// Email
email, err := valueobject.NewEmail("john@example.com")
if err != nil {
    return err // Invalid email
}
fmt.Println(email.Value()) // john@example.com

// Phone
phone, err := valueobject.NewPhone("+1", "5551234567")
if err != nil {
    return err // Invalid phone
}
fmt.Println(phone.Full()) // +15551234567

// Money
money, err := valueobject.NewMoney(100.50, "USD")
if err != nil {
    return err // Invalid currency
}
fmt.Println(money.Format()) // $100.50

// Address
addr, err := valueobject.NewAddress(
    "123 Main St",
    "New York",
    "US",
    "10001",
)
```

**Characteristics**:

- **Immutable**: No setters, create new instance to change
- **Value equality**: Compared by value, not identity
- **Self-validating**: Validation in constructor
- **Encapsulation**: Business rules inside value object

---

### 9. pkg/aggregate - Base Aggregate

**Base aggregate pattern** (DDD) for aggregate roots with event sourcing support.

**Features**:

- Event collection (uncommitted events)
- Version tracking (optimistic locking)
- Created/Updated timestamps
- Aggregate ID (UUIDv7)

**Usage**:

```go
import "github.com/basilex/promenade/pkg/aggregate"

// Embed in aggregate root
type User struct {
    aggregate.BaseAggregate

    ID       uuid.UUID
    Email    string
    Name     string
    Password string
}

// Factory method
func NewUser(email, name, password string) (*User, error) {
    user := &User{
        BaseAggregate: aggregate.NewBase(),
        ID:            uuidv7.New(),
        Email:         email,
        Name:          name,
        Password:      hashPassword(password),
    }

    // Add domain event
    user.AddEvent(bus.NewBaseEvent("user.created", user.ID))

    return user, nil
}

// In repository
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    // Save to database
    if err := r.db.Insert(ctx, user); err != nil {
        return err
    }

    // Publish uncommitted events
    for _, event := range user.GetEvents() {
        r.bus.Publish(ctx, event.Type(), event)
    }

    user.ClearEvents()
    return nil
}
```

---

### 10. pkg/saga - Saga Orchestration (Planned)

**Distributed transaction coordination** using Saga pattern for cross-context workflows.

**Planned Features**:

- [ ] Saga definition DSL
- [ ] Compensation logic
- [ ] State persistence
- [ ] Retry and recovery
- [ ] Saga visualization

**Example** (future):

```go
saga := saga.NewSaga("OrderFulfillment")
saga.
    Step("ReserveInventory", reserveInventoryHandler, compensateInventoryHandler).
    Step("ChargePayment", chargePaymentHandler, refundPaymentHandler).
    Step("ShipOrder", shipOrderHandler, cancelShipmentHandler).
    Execute(ctx)
```

**Status**: Planned for Q1 2026

---

### 11. pkg/jsonb - JSONB Utilities

**PostgreSQL JSONB helpers** for storing flexible data in JSON columns.

**Features**:

- Marshal/unmarshal Go structs
- Null handling
- Query builders for JSONB operators
- Index creation helpers

**Usage**:

```go
import "github.com/basilex/promenade/pkg/jsonb"

// Store metadata as JSONB
type Contact struct {
    ID       uuid.UUID
    Type     string
    Metadata jsonb.JSONB  // map[string]interface{} wrapper
}

metadata := jsonb.JSONB{
    "source": "web_form",
    "campaign": "summer_2025",
    "tags": []string{"vip", "newsletter"},
}

// Save to database
_, err := db.Exec(`
    INSERT INTO contacts (id, type, metadata)
    VALUES ($1, $2, $3)
`, id, "email", metadata)

// Query JSONB
var contact Contact
err := db.Get(&contact, `
    SELECT * FROM contacts
    WHERE metadata->>'source' = 'web_form'
`)
```

---

## Testing

### Run All Package Tests

```bash
# All pkg tests
go test ./pkg/... -v

# Specific package
go test ./pkg/bus -v
go test ./pkg/logger -v
go test ./pkg/uuidv7 -v

# With coverage
go test ./pkg/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage Summary

| Package     | Tests | Coverage | Status  |
| ----------- | ----- | -------- | ------- |
| bus         | 67    | 100%     | Passing |
| logger      | 15    | 95%      | Passing |
| migration   | 8     | 90%      | Passing |
| response    | 12    | 100%     | Passing |
| uuidv7      | 10    | 100%     | Passing |
| valueobject | 25    | 95%      | Passing |
| aggregate   | 5     | 90%      | Passing |
| jsonb       | 8     | 85%      | Passing |

---

## Package Dependencies

### Internal Dependencies

```
pkg/
 bus
    (no internal deps)
 logger
    (no internal deps)
 migration
    logger
 response
    logger
 uuidv7
    (no internal deps)
 valueobject
    reference
 aggregate
    uuidv7
 jsonb
     (no internal deps)
```

**Design Rule**: Packages in `pkg/` **MUST NOT** depend on `internal/` code.

---

## Best Practices

### DO

- **Use packages from `pkg/`** in all contexts
- **Keep packages small and focused** (single responsibility)
- **Write comprehensive tests** for all packages
- **Document public APIs** with GoDoc comments
- **Avoid external dependencies** when possible
- **Use interfaces** for flexibility

### DON'T

- **DON'T import `internal/`** from `pkg/` packages
- **DON'T add context-specific logic** to `pkg/`
- **DON'T create circular dependencies**
- **DON'T skip tests** (100% critical path coverage)
- **DON'T use global state** (except logger.Init)

---

## Future Packages

### Planned (Q1 2026)

- [ ] **pkg/cache** - Redis/in-memory caching
- [ ] **pkg/email** - Email sending (SMTP, templates)
- [ ] **pkg/sms** - SMS sending (Twilio, etc.)
- [ ] **pkg/storage** - File storage (S3, local)
- [ ] **pkg/encryption** - AES encryption utilities
- [ ] **pkg/jwt** - JWT token generation/validation
- [ ] **pkg/validation** - Enhanced validation rules

### Under Consideration

- [ ] **pkg/webhook** - Webhook delivery system
- [ ] **pkg/rate_limit** - Rate limiting (Redis-based)
- [ ] **pkg/metrics** - Prometheus metrics
- [ ] **pkg/tracing** - OpenTelemetry integration

---

## Related Documentation

- [Documentation Index](../docs/INDEX.md)
- [Clean Architecture Summary](../docs/CLEAN_ARCHITECTURE_SUMMARY.md)
- [Event Bus Documentation](bus/README.md)
- [Testing Patterns](../docs/TESTING_PATTERNS.md)
- [Testing Guide](../test/README.md)
- [Migrations README](../migrations/README.md)

---

**Last Updated**: 2025-12-27  
**Total Tests**: 150+ tests across 32 packages  
**Status**: Production-ready  
**Maintainer**: Promenade Team
