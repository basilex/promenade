# Package Library Overview

**Reusable, context-agnostic components** for Promenade Platform.

## Available Packages

### Event Bus 🚀

**Location**: `pkg/bus/`  
**Documentation**: [Event Bus](/packages/bus)

Central communication hub with Memory and Redis adapters.

**Key Features**:
- 377K events/sec (Memory adapter)
- Retry policy with exponential backoff
- Panic recovery
- 67 tests (100% coverage)

**Quick Example**:
```go
event := bus.NewBaseEvent("user.registered", userID)
eventBus.Publish(ctx, bus.TopicUserRegistered, event)
```

### JWT Authentication 🔐

**Location**: `pkg/jwt/`  
**Status**: Production-ready  
**Tests**: 18 tests (87% coverage)

Token-based authentication with RBAC support.

**Key Features**:
- Access + Refresh token pairs
- Role-based authorization
- Token revocation (Redis)
- Gin middleware

**Quick Example**:
```go
tokenPair, _ := jwtManager.GenerateTokenPair(userID, email, roles)
// Returns: access token (15min) + refresh token (7 days)
```

### UUID v7 ⚡

**Location**: `pkg/uuidv7/`  
**Status**: Production-ready  
**Tests**: 10 tests (100% coverage)

Time-ordered UUIDs for 2x faster database inserts.

**Quick Example**:
```go
id := uuidv7.New()  // Time-ordered UUID (RFC 9562)
ts := uuidv7.ExtractTime(id)  // Get timestamp
```

### Logger 📝

**Location**: `pkg/logger/`  
**Status**: Production-ready  
**Tests**: 15 tests (95% coverage)

Structured logging with context propagation.

**Quick Example**:
```go
logger.Info("User registered", 
    slog.String("user_id", userID.String()),
    slog.String("email", email),
)
```

### Response Helper 📦

**Location**: `pkg/response/`  
**Status**: Production-ready  
**Tests**: 13 tests (100% coverage)

Standard HTTP responses with pagination.

**Quick Example**:
```go
response.Success(c, user)                           // 200 OK
response.Error(c, 404, "USER_NOT_FOUND", "...")    // Error
response.Paginated(c, users, total, page, pageSize) // Paginated
```

### Value Objects 💎

**Location**: `pkg/valueobject/`  
**Status**: Production-ready  
**Tests**: 45 tests (95% coverage)

Immutable domain concepts (Email, Phone, Money, Address).

**Quick Example**:
```go
email, err := valueobject.NewEmail("user@example.com")
phone, err := valueobject.NewPhone("+380", "501234567")
money, err := valueobject.NewMoney(100.50, "USD")
```

### Migration Manager 🗄️

**Location**: `pkg/migration/`  
**Status**: Production-ready  
**Tests**: 8 tests (90% coverage)

Namespace-based database migrations.

**Quick Example**:
```bash
make migrate-new CONTEXT=identity NAME=add_users
make migrate-identity
```

### Aggregate Pattern 🏗️

**Location**: `pkg/aggregate/`  
**Status**: Production-ready  
**Tests**: 5 tests

Base aggregate for domain entities.

**Quick Example**:
```go
type User struct {
    aggregate.BaseAggregate
    ID    uuid.UUID
    Email string
}
```

### JSONB Utilities 📊

**Location**: `pkg/jsonb/`  
**Status**: Production-ready  
**Tests**: 8 tests

PostgreSQL JSONB helpers.

**Quick Example**:
```go
data := map[string]interface{}{"key": "value"}
jsonbData := jsonb.Marshal(data)  // For PostgreSQL JSONB column
```

### Saga Pattern 🔄

**Location**: `pkg/saga/`  
**Status**: Production-ready  
**Tests**: 28 tests (100% coverage)

Distributed transaction orchestration.

**Quick Example**:
```go
saga := saga.NewSaga("order-fulfillment")
saga.AddStep("charge-payment", chargeStep, compensateCharge)
saga.AddStep("reserve-inventory", reserveStep, compensateReserve)
saga.Execute(ctx)
```

## Package Statistics

| Package         | Tests | Coverage | Status     |
| --------------- | ----- | -------- | ---------- |
| **bus**         | 67    | 100%     | Production |
| **jwt**         | 18    | 87%      | Production |
| **logger**      | 15    | 95%      | Production |
| **uuidv7**      | 10    | 100%     | Production |
| **response**    | 13    | 100%     | Production |
| **valueobject** | 45    | 95%      | Production |
| **migration**   | 8     | 90%      | Production |
| **aggregate**   | 5     | 90%      | Production |
| **jsonb**       | 8     | 95%      | Production |
| **saga**        | 28    | 100%     | Production |
| **Total**       | **217** | **95%** | ✅        |

## Usage in Contexts

All packages are context-agnostic and can be used across all contexts:

```go
// Identity Context uses
import (
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/basilex/promenade/pkg/valueobject"
    "github.com/basilex/promenade/pkg/jwt"
)

// Customer Management uses
import (
    "github.com/basilex/promenade/pkg/bus"
    "github.com/basilex/promenade/pkg/logger"
)
```

## Next Steps

- [Event Bus](/packages/bus) - Deep dive into event-driven patterns
- [Architecture Guide](/guide/architecture) - How packages fit together
- [Identity Context](/contexts/identity) - See packages in action
- [GitHub Docs](https://github.com/basilex/promenade/tree/dev/pkg) - Complete package documentation
