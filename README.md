# Promenade CRM Platform

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![DDD](https://img.shields.io/badge/Architecture-DDD-green.svg)](docs/)

**Production-ready CRM platform** built with **Domain-Driven Design (DDD)**, **Bounded Contexts**, and **CQRS/Event Sourcing patterns**.

---

## 🎯 Architecture Overview

Promenade follows **strict Domain-Driven Design** principles with clear **Bounded Context** separation and **Event-Driven Architecture** at its core.

### Core Architecture Principles

1. **Bounded Contexts** - Each context is autonomous with own models and database schema
2. **Aggregates** - Business entities with invariants and transactional boundaries
3. **Value Objects** - Immutable domain concepts (Email, Phone, Money, Address)
4. **Domain Events** - Asynchronous communication between contexts via Event Bus
5. **Sagas** - Distributed transactions coordination (planned)
6. **CQRS** - Separate read/write models for complex queries (planned)

### Event-Driven Architecture 🚀

**Event Bus** is the central nervous system of Promenade:

- **Multiple Adapters**: Memory (dev: 377K events/sec) and Redis (prod: distributed)
- **Retry Policy**: Exponential backoff with configurable attempts
- **Panic Recovery**: Handler failures don't crash the system
- **Graceful Shutdown**: Ensures all events processed before shutdown
- **67 Tests**: Comprehensive test coverage (100% passing)

**See**: [pkg/bus/README.md](pkg/bus/README.md) for complete Event Bus documentation

### Bounded Contexts

| Context                 | Aggregates                            | Description                    | Status        | Documentation                                  |
| ----------------------- | ------------------------------------- | ------------------------------ | ------------- | ---------------------------------------------- |
| **Shared**              | Country, Currency, Language, Timezone | Reference data (read-only)     | ✅ Production | [README](internal/contexts/shared/README.md)   |
| **Identity**            | User, Contact, Profile                | Authentication & authorization | ✅ Production | [README](internal/contexts/identity/README.md) |
| **Customer Management** | Customer, Company, Deal, Interaction  | CRM core functionality         | 📋 Planned    | Coming Q1 2026                                 |
| **Order Management**    | Order, OrderItem, Fulfillment         | Order processing and tracking  | 📋 Planned    | Coming Q2 2026                                 |
| **Billing**             | Invoice, Payment, Subscription        | Billing and payments           | 📋 Planned    | Coming Q2 2026                                 |
| **Analytics**           | Report, Dashboard, Metric             | Business intelligence          | 📋 Planned    | Coming Q3 2026                                 |

**Context Isolation**: Contexts communicate ONLY via Event Bus (no direct dependencies)

---

## 🏗️ Project Structure

```
promenade/
├── cmd/
│   ├── api/                    # HTTP server entry point (211 lines)
│   │   └── main.go            # Bootstrap: Config → Logger → DB → Migrations → Event Bus → Server
│   └── migrate/                # Migration CLI tool
├── internal/
│   ├── contexts/               # Bounded Contexts (DDD)
│   │   ├── shared/            # ✅ Shared Context (Reference Data: Country, Currency, Language, Timezone)
│   │   │   ├── README.md      # Complete documentation (~450 lines)
│   │   │   ├── domain/entity/ # 4 aggregates (Country, Currency, Language, Timezone)
│   │   │   ├── domain/repository/ # Repository interfaces
│   │   │   ├── usecase/       # Business logic (GetCountries, GetCurrencies, etc.)
│   │   │   ├── adapter/http/  # HTTP handlers (Gin)
│   │   │   └── adapter/repository/postgres/ # PostgreSQL implementation
│   │   ├── identity/          # ✅ Identity Context (User, Contact, Profile)
│   │   │   ├── README.md      # Complete documentation (~550 lines)
│   │   │   ├── user/          # User Aggregate (authentication, RBAC)
│   │   │   ├── contact/       # Contact Aggregate (email, phone, address)
│   │   │   └── profile/       # Profile Aggregate (🚧 in progress)
│   │   ├── customer-mgmt/     # 📋 Customer Management Context (planned)
│   │   ├── order-mgmt/        # 📋 Order Management Context (planned)
│   │   ├── billing/           # 📋 Billing Context (planned)
│   │   └── README.md          # Contexts overview
│   └── infrastructure/         # Cross-cutting concerns
│       ├── config/            # Configuration management (YAML + env vars)
│       └── database/          # Database connection & transactions
├── pkg/                        # Shared Domain Primitives & Utilities
│   ├── README.md              # ✅ Complete package overview (~400 lines)
│   ├── bus/                   # ✅ Event Bus (Memory/Redis, 67 tests, 377K/sec)
│   │   └── README.md          # ✅ Complete documentation (~600 lines)
│   ├── aggregate/             # Base Aggregate pattern (DDD)
│   ├── valueobject/           # Value Objects (Email, Phone, Money, Address)
│   ├── saga/                  # Saga orchestration (🚧 planned Q1 2026)
│   ├── uuidv7/               # Time-ordered UUIDs (RFC 9562, 2x faster inserts)
│   ├── logger/               # Structured logging (slog wrapper, context-aware)
│   ├── migration/            # Namespace-based database migrations
│   ├── response/             # Standard HTTP responses & pagination
│   ├── jsonb/                # PostgreSQL JSONB utilities
│   └── reference/            # Reference data validation helpers
├── migrations/                 # Database migrations (namespace-based)
│   ├── core/                 # Core infrastructure (UUID v7, auth, RBAC)
│   ├── shared/               # Shared context migrations (reference data)
│   └── identity/             # Identity context migrations (users, contacts)
├── test/                       # Testing infrastructure
│   ├── README.md             # Testing structure documentation
│   ├── integration/          # Integration test helpers
│   └── smoke/                # Smoke tests
├── config/                     # Configuration files
│   ├── app.dev.yaml          # Development (Memory bus, localhost DB)
│   ├── app.test.yaml         # Testing
│   └── app.prod.yaml         # Production (Redis bus)
└── docs/                       # Architecture documentation (20+ guides)
    ├── INDEX.md              # Documentation index
    ├── ARCHITECTURE_OVERVIEW.md # High-level architecture
    ├── ARCHITECTURE_QUICKREF.md # Quick reference guide
    └── ...                   # 20+ detailed guides
```

### Key Directories Explained

- **cmd/api/main.go**: Application entry point with proper initialization order
- **internal/contexts/**: Bounded contexts (autonomous, isolated)
- **pkg/**: Shared packages (context-agnostic, reusable across all contexts)
- **migrations/**: Namespace-based migrations (run in order: core → shared → identity)
- **test/**: Mirror path testing structure (unit + integration tests)
- **config/**: Environment-specific YAML configs (dev/test/prod)

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.23+**
- **Docker & Docker Compose** (for PostgreSQL)
- **Make**

### 1. Clone & Setup

```bash
git clone https://github.com/basilex/promenade.git
cd promenade

# Start PostgreSQL (localhost:5432)
make docker-up
```

### 2. Run Migrations

```bash
# Migrations run automatically on app startup
# Or manually:
make migrate
```

### 3. Start Application

```bash
# Development mode (Docker + migrations + API)
make dev

# Or build and run separately
make build
./bin/promenade
```

Server starts on **http://localhost:8081**

### 4. Health Check

```bash
curl http://localhost:8081/health
```

### Available Make Commands

```bash
make help              # Show all available commands

# Development
make dev               # Full dev environment (Docker + migrations + API)
make dev-fresh         # Fresh start with clean database
make build             # Build binary
make run               # Build and run
make fmt               # Format code
make lint              # Run linters

# Testing
make test              # Run all tests (156+ tests, ~25 seconds)
make test-unit         # Unit tests only (~5 seconds)
make test-integration  # Integration tests with real DB
make test-coverage     # HTML coverage report

# Database
make docker-up         # Start PostgreSQL
make docker-down       # Stop PostgreSQL
make db-reset          # Drop and recreate database
make migrate           # Run all migrations

# Documentation
make swagger-all       # Generate API documentation
```

---

## 🧪 Testing

Promenade uses a **mirror path testing structure** where tests live alongside code:

### Test Organization

```
pkg/
├── bus/
│   ├── bus.go
│   ├── bus_test.go         # Unit tests (67 tests)
│   ├── memory/
│   │   ├── memory.go
│   │   └── memory_test.go
│   └── redis/
│       ├── redis.go
│       └── redis_test.go

internal/contexts/identity/
├── user/
│   ├── entity/
│   │   ├── user.go
│   │   └── user_test.go
│   └── usecase/
│       ├── user_usecase.go
│       └── user_usecase_test.go
```

### Running Tests

```bash
# All tests (156+ tests, ~25 seconds)
make test

# By category
make test-unit              # Unit tests only (~5 seconds)
make test-integration       # Integration tests with real DB (~36 seconds)

# By context
go test ./internal/contexts/identity/... -v
go test ./internal/contexts/shared/... -v

# By package
go test ./pkg/bus/... -v
go test ./pkg/logger/... -v

# With coverage
make test-coverage          # Generate HTML coverage report
```

### Test Statistics

| Component            | Tests | Coverage | Duration |
| -------------------- | ----- | -------- | -------- |
| **pkg/bus**          | 67    | 100%     | ~2s      |
| **pkg/logger**       | 15    | 95%      | <1s      |
| **pkg/uuidv7**       | 10    | 100%     | <1s      |
| **pkg/response**     | 12    | 100%     | <1s      |
| **pkg/migration**    | 8     | 90%      | ~1s      |
| **pkg/valueobject**  | 25    | 95%      | <1s      |
| **pkg/aggregate**    | 5     | 90%      | <1s      |
| **pkg/jsonb**        | 8     | 85%      | <1s      |
| **pkg/reference**    | 6     | 80%      | <1s      |
| **Identity Context** | TBD   | -        | -        |
| **Shared Context**   | TBD   | -        | -        |

**Total**: 156+ tests, 90%+ average coverage

### Test Database

Integration tests use a separate database on port 5433:

```bash
# Start test database
make test-db-start

# Stop test database
make test-db-stop
```

**See**: [test/README.md](test/README.md) for complete testing documentation

---

## 📊 Domain Model Example: Identity Context

### Contact Aggregate

```go
// Contact is an aggregate root for user contact information
type Contact struct {
    aggregate.BaseAggregate

    ID         uuid.UUID
    UserID     uuid.UUID
    Type       ContactType    // email, phone, address
    Label      string        // "Work", "Home", "Personal"

    // Value Objects (only one populated based on Type)
    Email      *valueobject.Email
    Phone      *valueobject.Phone
    Address    *valueobject.Address

    // Business flags
    IsPrimary  bool  // Only one primary per user per type
    IsVerified bool  // Email confirmation, phone OTP, etc.
    IsPublic   bool  // Visible in public profile
}

// Factory method with business validation
func NewEmailContact(userID uuid.UUID, email, label string) (*Contact, error) {
    emailVO, err := valueobject.NewEmail(email)
    if err != nil {
        return nil, fmt.Errorf("invalid email: %w", err)
    }

    return &Contact{
        BaseAggregate: aggregate.NewBase(),
        ID:            uuidv7.New(),
        UserID:        userID,
        Type:          ContactTypeEmail,
        Label:         label,
        Email:         &emailVO,
        IsPrimary:     false,
        IsVerified:    false,
        IsPublic:      false,
    }, nil
}
```

### Value Object Example

```go
// Email is a value object for email addresses
type Email struct {
    value string
}

func NewEmail(email string) (Email, error) {
    trimmed := strings.TrimSpace(strings.ToLower(email))

    // Validation
    if trimmed == "" {
        return Email{}, fmt.Errorf("email cannot be empty")
    }

    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(trimmed) {
        return Email{}, fmt.Errorf("invalid email format")
    }

    return Email{value: trimmed}, nil
}

func (e Email) Value() string {
    return e.value
}
```

---

## 🔧 Configuration

Promenade uses **YAML configuration files** per environment with **environment variable overrides**:

### Configuration Files

```
config/
├── app.dev.yaml        # Development (Memory bus, localhost DB)
├── app.test.yaml       # Testing (isolated test DB)
└── app.prod.yaml       # Production (Redis bus, production DB)
```

### Development Configuration

**config/app.dev.yaml**

```yaml
app:
  name: "Promenade CRM"
  environment: "development"
  version: "0.1.0"

server:
  host: "0.0.0.0"
  port: 8081
  read_timeout: 10s
  write_timeout: 10s

database:
  host: "localhost"
  port: 5432
  user: "system"
  password: "passw0rd"
  database: "promenade_dev"
  ssl_mode: "disable"

# Event Bus configuration (Memory adapter for dev)
bus:
  adapter: "memory" # memory (dev) or redis (prod)
  worker_pool_size: 10 # Number of concurrent workers
  buffer_size: 1000 # Event buffer capacity
  retry_attempts: 3 # Failed event retry attempts
  retry_delay: 1s # Delay between retries

logging:
  level: "debug" # debug, info, warn, error
  format: "text" # text or json
  add_source: true # Include source file:line
```

### Production Configuration

**config/app.prod.yaml**

```yaml
# ... (same app/server sections)

database:
  host: "${DB_HOST}" # Environment variable override
  port: 5432
  user: "${DB_USER}"
  password: "${DB_PASSWORD}"
  database: "promenade_prod"
  ssl_mode: "require"

# Redis Event Bus for production (distributed)
bus:
  adapter: "redis"
  redis_addr: "${REDIS_ADDR}"
  redis_password: "${REDIS_PASSWORD}"
  worker_pool_size: 20
  buffer_size: 5000
  retry_attempts: 5
  retry_delay: 2s

logging:
  level: "info"
  format: "json" # Structured logging for prod
  add_source: false
```

### Environment Variables

Override sensitive configuration values:

```bash
export DB_HOST="production-db.example.com"
export DB_USER="promenade_app"
export DB_PASSWORD="secure_password"
export REDIS_ADDR="redis.example.com:6379"
export REDIS_PASSWORD="redis_password"
```

### Configuration Loading

The application loads config on startup in [cmd/api/main.go](cmd/api/main.go):

```go
// Load configuration (environment-specific)
cfg, err := config.Load()
if err != nil {
    slog.Error("Failed to load config", slog.Any("error", err))
    os.Exit(1)
}
```

Environment is detected via `ENVIRONMENT` env var (defaults to `dev`):

```bash
ENVIRONMENT=production ./bin/promenade   # Loads app.prod.yaml
ENVIRONMENT=test ./bin/promenade        # Loads app.test.yaml
./bin/promenade                         # Loads app.dev.yaml (default)
```

**See**: [internal/infrastructure/config/](internal/infrastructure/config/) for config implementation

---

## 📚 Key Concepts

### Package Library

Promenade includes a comprehensive **package library** (`pkg/`) with reusable, context-agnostic components:

| Package         | Purpose                               | Tests | Status        | Documentation               |
| --------------- | ------------------------------------- | ----- | ------------- | --------------------------- |
| **bus**         | Event Bus (Memory/Redis adapters)     | 67    | ✅ Production | [README](pkg/bus/README.md) |
| **logger**      | Structured logging (slog wrapper)     | 15    | ✅ Production | -                           |
| **migration**   | Namespace-based DB migrations         | 8     | ✅ Production | -                           |
| **response**    | Standard HTTP responses               | 12    | ✅ Production | -                           |
| **uuidv7**      | Time-ordered UUIDs (RFC 9562)         | 10    | ✅ Production | -                           |
| **valueobject** | DDD Value Objects                     | 25    | ✅ Production | -                           |
| **aggregate**   | Base Aggregate pattern                | 5     | ✅ Production | -                           |
| **jsonb**       | PostgreSQL JSONB utilities            | 8     | ✅ Production | -                           |
| **reference**   | Reference data validation             | 6     | ✅ Production | -                           |
| **saga**        | Distributed transaction orchestration | -     | 📋 Planned    | Coming Q1 2026              |

**Total**: 156+ tests across all packages

**Key Highlights**:

- **bus**: Central event-driven communication hub (377K events/sec with Memory adapter)
- **uuidv7**: Time-ordered UUIDs provide 2x faster inserts than UUID v4
- **valueobject**: Email, Phone, Money, Address with immutability and validation
- **aggregate**: Base pattern for all domain aggregates (event sourcing support)

**See**: [pkg/README.md](pkg/README.md) for complete package library documentation

---

## 📊 Domain Model Example: Identity Context

### Aggregates

**Aggregate** is a cluster of domain objects treated as a single unit with:

- **Aggregate Root** - entry point with global identity
- **Invariants** - business rules enforced within boundary
- **Transactional Consistency** - changes saved atomically

### Value Objects

**Value Object** is an immutable object defined by its attributes:

- No identity (equality by value)
- Validation in constructor
- Examples: Email, Phone, Money, Address, DateRange

### Domain Events

**Domain Event** represents something that happened in the domain:

```go
type ContactVerified struct {
    ContactID uuid.UUID
    UserID    uuid.UUID
    Type      ContactType
    VerifiedAt time.Time
}
```

### Sagas

**Saga** coordinates long-running distributed transactions:

```go
type OrderFulfillmentSaga struct {
    saga.BaseSaga
    OrderID   uuid.UUID
    CustomerID uuid.UUID
    PaymentID  uuid.UUID
}
```

---

## 🗄️ Database

### Migrations

Migrations are namespace-based for context isolation:

```
migrations/
├── core/                        # Core infrastructure
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000002_core_auth_full.up.sql
│   └── 000003_core_rbac_full.up.sql
└── identity/                    # Identity context
    └── 000001_identity_contacts.up.sql
```

### UUID v7

Time-ordered UUIDs for better database performance:

- 2x faster inserts than UUID v4
- Natural ordering by creation time
- Better B-tree index locality

```go
id := uuidv7.New()  // Time-ordered UUID
```

---

## 🛣️ Roadmap

### Phase 1: Foundation ✅ (Completed)

- [x] DDD primitives (Aggregate, Value Objects, Saga)
- [x] Project structure (Bounded Contexts)
- [x] Identity Context: Contact aggregate
- [x] Database migrations system
- [x] Testing infrastructure

### Phase 2: Identity Context (Current - Week 1)

- [ ] User aggregate (registration, authentication)
- [ ] Contact aggregate (email, phone, address)
- [ ] Integration tests
- [ ] HTTP API (REST)

### Phase 3: Customer Management Context (Week 2-3)

- [ ] Customer aggregate (lifecycle, segmentation)
- [ ] Company aggregate (B2B)
- [ ] Deal aggregate (pipeline, stages)
- [ ] Interaction aggregate (calls, emails, meetings)

### Phase 4: Order Management Context (Week 4-5)

- [ ] Order aggregate (creation, fulfillment)
- [ ] OrderItem value object
- [ ] Fulfillment saga (payment → inventory → shipping)

### Phase 5: Analytics & Reporting (Week 6)

- [ ] CQRS read models
- [ ] Dashboards
- [ ] Business metrics

---

## 🤝 Contributing

This is a learning project focused on DDD architecture. Contributions welcome!

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Follow DDD principles and existing patterns
4. Write tests (maintain 80%+ coverage)
5. Commit changes (`git commit -m 'Add amazing feature'`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Open Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 📧 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Email**: alexander.vasilenko@gmail.com

---

## **Built with Domain-Driven Design and Go** 🚀

## 📖 Documentation

### Complete Documentation Library

Promenade includes **comprehensive documentation** covering all aspects of the architecture:

**Core Documentation**:

- [Architecture Overview](docs/ARCHITECTURE_OVERVIEW.md) - High-level architecture and design principles
- [Architecture Quick Reference](docs/ARCHITECTURE_QUICKREF.md) - Quick reference for common patterns
- [Documentation Index](docs/INDEX.md) - Complete documentation catalog

**Bounded Contexts**:

- [Identity Context](internal/contexts/identity/README.md) - User, Contact, Profile aggregates (~550 lines)
- [Shared Context](internal/contexts/shared/README.md) - Reference data (Country, Currency, Language, Timezone) (~450 lines)

**Package Library**:

- [Package Overview](pkg/README.md) - All shared packages documentation (~400 lines)
- [Event Bus](pkg/bus/README.md) - Central communication hub (Memory/Redis adapters) (~600 lines)

**Infrastructure**:

- [Testing Guide](test/README.md) - Testing structure and best practices
- [Migration Architecture](docs/MIGRATION_ARCHITECTURE.md) - Database migration system
- [UUID v7 Guide](docs/UUID_V7_GUIDE.md) - Time-ordered UUID implementation

**Testing & Quality**:

- [Testing Infrastructure](docs/TESTING_INFRASTRUCTURE.md) - Test organization and execution
- Mirror path testing structure (tests alongside code)
- 156+ tests with 90%+ average coverage

**Total Documentation**: 20+ comprehensive guides with examples, best practices, and architecture decisions

---
