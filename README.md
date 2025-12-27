# Promenade CRM Platform

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![DDD](https://img.shields.io/badge/Architecture-DDD-green.svg)](docs/)

**Production-ready CRM platform** built with **Domain-Driven Design (DDD)**, **Bounded Contexts**, and **CQRS/Event Sourcing patterns**.

---

## 🎯 Architecture Overview

Promenade follows **strict Domain-Driven Design** principles with clear **Bounded Context** separation:

### Bounded Contexts

| Context                 | Aggregates                           | Description                          | Status         |
| ----------------------- | ------------------------------------ | ------------------------------------ | -------------- |
| **Identity**            | User, Contact                        | User identity and contact management | ✅ In Progress |
| **Customer Management** | Customer, Company, Deal, Interaction | CRM core functionality               | 📋 Planned     |
| **Order Management**    | Order, OrderItem, Fulfillment        | Order processing and tracking        | 📋 Planned     |
| **Billing**             | Invoice, Payment, Subscription       | Billing and payments                 | 📋 Planned     |
| **Analytics**           | Report, Dashboard, Metric            | Business intelligence                | 📋 Planned     |

### Core Principles

1. **Bounded Contexts** - Each context is autonomous with own models and database schema
2. **Aggregates** - Business entities with invariants and transactional boundaries
3. **Value Objects** - Immutable domain concepts (Email, Phone, Money, Address)
4. **Domain Events** - Asynchronous communication between contexts
5. **Sagas** - Distributed transactions coordination
6. **CQRS** - Separate read/write models for complex queries

---

## 🏗️ Project Structure

```
promenade/
├── cmd/
│   ├── api/                    # HTTP server entry point
│   └── migrate/                # Migration CLI tool
├── internal/
│   ├── contexts/               # Bounded Contexts (DDD)
│   │   ├── identity/          # Identity Context
│   │   │   ├── user/          # User Aggregate
│   │   │   └── contact/       # Contact Aggregate
│   │   ├── customer-mgmt/     # Customer Management Context (planned)
│   │   └── order-mgmt/        # Order Management Context (planned)
│   └── infrastructure/         # Cross-cutting concerns
│       ├── config/            # Configuration management
│       └── database/          # Database connection & transactions
├── pkg/                        # Shared Domain Primitives
│   ├── aggregate/             # Base Aggregate pattern
│   ├── valueobject/           # Value Objects (Email, Phone, Money, etc.)
│   ├── saga/                  # Saga orchestration
│   ├── uuidv7/               # Time-ordered UUIDs
│   ├── logger/               # Structured logging
│   ├── migration/            # Database migrations
│   └── jsonb/                # JSONB utilities
├── migrations/                 # Database migrations (namespace-based)
│   ├── core/                 # Core infrastructure
│   └── identity/             # Identity context migrations
├── test/                       # Testing infrastructure
│   └── integration/          # Integration test helpers
├── config/                     # Configuration files
│   ├── app.dev.yaml          # Development config
│   ├── app.test.yaml         # Test config
│   └── app.prod.yaml         # Production config
└── docs/                       # Architecture documentation
```

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

# Start PostgreSQL
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
# Development mode
make dev

# Or build and run
make build
./bin/promenade
```

Server starts on **http://localhost:8081**

---

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific context tests
go test ./internal/contexts/identity/... -v

# Integration tests (requires test DB)
make test-integration
```

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

Configuration uses YAML files per environment:

**config/app.dev.yaml**

```yaml
app:
  name: "Promenade CRM"
  environment: "development"
  version: "2.0.0"

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
```

---

## 📚 Key Concepts

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

**Built with Domain-Driven Design and Go** 🚀
