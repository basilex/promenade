# Promenade Platform

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Tests](https://img.shields.io/badge/Tests-240+-success?style=flat)](test/)
[![Coverage](https://img.shields.io/badge/Coverage-90%25+-success?style=flat)](test/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![DDD](https://img.shields.io/badge/Architecture-DDD-green.svg)](docs/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**Modern backend platform for customer management, orders, and business workflows** — built with **Domain-Driven Design (DDD)**, **Bounded Contexts**, and **Event-Driven Architecture**.

> **Not a traditional CRM** — Promenade is a **modular platform** that grows with your needs.  
> Start with customer management, add orders when needed, integrate billing when ready.

📚 **[View Full Documentation](https://basilex.github.io/promenade/)** | 🚀 **[Quick Start Guide](https://basilex.github.io/promenade/guide/getting-started)** | 📖 **[API Reference](https://basilex.github.io/promenade/guide/api-reference)**

---

## Architecture Overview

Promenade follows **strict Domain-Driven Design** principles with clear **Bounded Context** separation and **Event-Driven Architecture** at its core.

### Core Architecture Principles

1. **Bounded Contexts** - Each context is autonomous with own models and database schema
2. **Aggregates** - Business entities with invariants and transactional boundaries
3. **Value Objects** - Immutable domain concepts (Email, Phone, Money, Address)
4. **Domain Events** - Asynchronous communication between contexts via Event Bus
5. **Sagas** - Distributed transactions coordination (planned)
6. **CQRS** - Separate read/write models for complex queries (planned)

### Event-Driven Architecture

**Event Bus** is the central nervous system of Promenade:

- **Multiple Adapters**: Memory (dev: 377K events/sec) and Redis (prod: distributed)
- **Retry Policy**: Exponential backoff with configurable attempts
- **Panic Recovery**: Handler failures don't crash the system
- **Graceful Shutdown**: Ensures all events processed before shutdown
- **67 Tests**: Comprehensive test coverage (100% passing)

**See**: [pkg/bus/README.md](pkg/bus/README.md) for complete Event Bus documentation

### Role-Based Access Control (RBAC)

Promenade implements **enterprise-grade RBAC** for fine-grained authorization:

- **Roles**: Named collections of permissions (superadmin, admin, manager, user, guest)
- **Permissions**: Specific actions on resources (e.g., "users:create", "customers:delete")
- **JWT Integration**: Roles embedded in tokens for stateless authorization
- **Flexible**: Easy to add custom roles and permissions without code changes
- **5 System Roles**: Pre-configured with 29+ permissions across all resources
- **14 API Endpoints**: Complete management of roles and permissions

**See**: [docs/RBAC.md](docs/RBAC.md) for complete RBAC implementation guide

### Rate Limiting

Promenade implements **IP-based rate limiting** to protect against brute-force attacks:

- **Token Bucket Algorithm**: Each IP gets separate rate limiter with configurable rate and burst
- **Authentication Protection**: Login (5/min), Register (3/min) endpoints rate-limited
- **X-RateLimit Headers**: Standard headers (Limit, Remaining, Reset) in all responses
- **Thread-Safe**: Concurrent access with RWMutex, production-ready
- **Memory Management**: Periodic cleanup prevents memory leaks
- **10 Tests**: Comprehensive test coverage (100% passing)

**See**: [docs/RATE_LIMITING.md](docs/RATE_LIMITING.md) for complete Rate Limiting documentation

### Multi-Database Support

Promenade implements **database-agnostic architecture** supporting multiple SQL databases:

- **Dialect Pattern**: Abstracts SQL syntax differences (placeholders, JSON types)
- **Supported Databases**: PostgreSQL (production), SQLite (dev/test), MySQL (planned), SQL Server (planned)
- **JSONB Abstraction**: `jsonstore.Field[T]` works across all databases (JSONB → TEXT → JSON)
- **UUID Generation**: All IDs generated in Go code (`uuidv7.New()`), not database defaults
- **Timestamp Management**: Business logic calls `.Touch()` to update UpdatedAt (no database triggers)
- **101 Tests**: pkg/database + pkg/jsonstore with 100% coverage

**Key Features**:
- Write queries with `?` placeholders, auto-convert to database-specific syntax ($1 for Postgres, ? for SQLite)
- Store JSON as TEXT (cross-database compatible) with type-safe Go wrappers
- Migrations work identically across databases (no vendor-specific DDL)

**Quick Switch Example**:
```bash
# PostgreSQL (production)
DATABASE_DRIVER=postgres ENVIRONMENT=production ./bin/promenade

# SQLite (development/demo)
DATABASE_DRIVER=sqlite ENVIRONMENT=development ./bin/promenade

# Configuration files (driver-environment format)
config/app.postgres-dev.yaml   # PostgreSQL + development
config/app.postgres-test.yaml  # PostgreSQL + testing
config/app.postgres-prod.yaml  # PostgreSQL + production
config/app.sqlite-dev.yaml     # SQLite + development
config/app.sqlite-test.yaml    # SQLite + testing
config/app.sqlite-prod.yaml    # SQLite + production
```

**See**: [docs/guides/database-adapters.md](docs/guides/database-adapters.md) | [docs/guides/jsonb-strategy.md](docs/guides/jsonb-strategy.md) | [config/SQLITE.md](config/SQLITE.md)

### Workspace Management

Promenade implements **explicit workspace state management** for seamless multi-database development:

**Design Philosophy**:
- **Single Source of Truth**: `.promenade.workspace` file defines `DATABASE_DRIVER` and `ENVIRONMENT`
- **Fail-Fast Validation**: Commands validate workspace before execution
- **Smart Switchers**: 9 commands cover all database × environment combinations
- **Environment Awareness**: Runners (`dev`, `test-all`, `prod`) enforce correct environment usage

**Core Commands**:
```bash
# Configure workspace (one-time)
make switch-postgres-dev   # PostgreSQL + development
make switch-sqlite-test    # SQLite + testing
make switch-postgres-prod  # PostgreSQL + production

# Check current state
make workspace             # Show DATABASE_DRIVER + ENVIRONMENT

# Use environment-aware runners
make dev                   # Requires ENVIRONMENT=development
make test-all              # Warns if not ENVIRONMENT=test
make prod                  # Requires ENVIRONMENT=production

# Workspace commands (no validation needed)
make build                 # Go compilation
make fmt                   # Code formatting
make lint                  # Linting
```

**Architecture Benefits**:
- ✅ No command explosion (50+ commands work with all databases)
- ✅ Prevents wrong environment execution (fail-fast validation)
- ✅ Natural developer workflow (configure once, work anywhere)
- ✅ Clear state visibility (`make workspace` always shows current config)
- ✅ Modular Makefile system (main + dev/test/prod modules)

**Example Workflow**:
```bash
# Day 1: PostgreSQL development
make switch-postgres-dev && make dev

# Day 2: SQLite testing  
make switch-sqlite-test && make test-all

# Day 3: Production deployment
make switch-postgres-prod && make prod
```

**See**: [Workspace Management Guide](docs/guides/workspace-management.md) for complete architecture documentation

### Health Checks

Promenade provides **comprehensive health monitoring** for all dependencies:

- **4 HTTP Endpoints**: `/health`, `/health/db`, `/health/redis`, `/health/bus`
- **3 Status Levels**: healthy, degraded, unhealthy
- **5-Second Timeout**: Prevents hanging checks
- **Graceful Degradation**: Optional dependencies (Redis) handled gracefully
- **Proper HTTP Status Codes**: 200 (healthy/degraded), 503 (unhealthy)
- **21 Tests**: Comprehensive test coverage (100% passing)

**Endpoints**:
- `GET /health` - Overall system health (PostgreSQL + Redis + Event Bus)
- `GET /health/db` - PostgreSQL database health
- `GET /health/redis` - Redis health (if configured)
- `GET /health/bus` - Event Bus health

**See**: [docs/HEALTH_CHECKS.md](docs/HEALTH_CHECKS.md) for complete Health Check documentation

### Local CI Validation

Promenade provides **local CI simulation** to catch issues before pushing to GitHub:

- **Pre-Push Validation**: Run all CI checks locally with `make pre-push`
- **Selective Checks**: Run lint (`make ci-lint`), tests (`make ci-test`), or build (`make ci-build`) independently
- **Time Savings**: 4+ minutes saved per failed push by catching issues early
- **GitHub Actions Parity**: Exact same checks as CI pipeline
- **5 Commands**: ci-check, ci-lint, ci-test, ci-build, pre-push

**Usage**:
```bash
# Before every push (REQUIRED)
make pre-push

# Or run checks separately
make ci-lint        # golangci-lint (0 issues)
make ci-test        # All tests + race detector
make ci-build       # Test compilation
```

**See**: [docs/guides/local-ci.md](docs/guides/local-ci.md) for complete Local CI Validation guide

### Caching Layer

Promenade implements **Redis-based caching** for improved performance and reduced database load:

- **Multiple Adapters**: Redis (production/dev), NoOp (testing/fallback)
- **Resource-Specific TTL**: Reference data (1h-24h), User data (10-30m), Sessions (30m-1h)
- **Cache-Aside Pattern**: Read-through cache with write-through invalidation
- **Graceful Degradation**: Application continues if Redis unavailable
- **Pattern-Based Invalidation**: Efficient bulk deletion with SCAN
- **JSON Marshaling**: Automatic serialization of complex types

**Cached Resources**:
- Reference data: Countries, Currencies, Languages, Timezones
- User data: Profiles, Customer records
- Session data: Temporary state

**See**: [docs/guides/caching.md](docs/guides/caching.md) for complete Caching implementation guide

### Customer Management

Promenade implements **complete CRM functionality** for B2C and B2B customer lifecycle management:

- **Customer Lifecycle**: Track customers through states (Lead → Prospect → Customer → Churned)
- **B2C & B2B Support**: Handle individual consumers and business contacts
- **Customer Segmentation**: Organize by tier (free, basic, pro, enterprise) and tags
- **Sales Pipeline**: Assign customers to sales reps, track source and status
- **14 API Endpoints**: Complete CRUD + business logic operations
- **Auto-validated Data**: Email and Phone value objects with validation

**Key Features**:
- State machine enforces lifecycle transitions
- JSONB tags for flexible metadata
- Soft delete support
- Sales rep assignment
- Integration with Identity and Order Management contexts

**See**: [Customer Management Guide](docs/concepts/customer-management.md) | [Company Management Guide](docs/concepts/company-management.md) | [Deal Management Guide](docs/concepts/deal-management.md)

### Deal Management

Promenade provides **complete sales pipeline management** with deal lifecycle tracking and **fully implemented business rules**:

- **Deal Lifecycle**: Track deals through stages (lead → qualified → proposal → negotiation → closed)
- **Probability Tracking**: Automatic probability calculation per stage (10% → 100%)
- **Money Handling**: Type-safe Money value object for deal values
- **Pipeline Statistics**: Real-time stats by stage and won deals analytics
- **12 API Endpoints**: Complete CRUD + business logic operations
- **Stage Transitions**: Enforced state machine (lead → ... → closed_won/closed_lost)

**Key Features**:
- State machine enforces valid stage transitions
- Automatic probability updates per stage
- Win/Loss tracking with actual close dates
- Filter deals by stage, customer, or sales rep
- Pipeline and revenue statistics
- Integration with Customer aggregate

**Implementation Status**: All core business rules implemented in entity (`MoveToStage()`, `MarkWon()`, `MarkLost()`). Tested via HTTP API with 100% success rate.

**See**: [docs/concepts/deal-management.md](docs/concepts/deal-management.md) for complete Deal Management guide

### Interaction Management

Promenade provides **comprehensive customer interaction tracking** for calls, emails, meetings, and notes:

- **Interaction Types**: Track calls (inbound/outbound), emails, meetings, notes
- **Interaction Lifecycle**: Create → In Progress → Completed with outcome tracking
- **JSONB Attendees**: Flexible participant tracking with PostgreSQL JSONB arrays
- **Follow-up Management**: Flag interactions requiring follow-up with dates and notes
- **Duration Tracking**: Automatic duration calculation for ended interactions
- **Performance Optimized**: LEFT JOIN queries prevent N+1 problem when listing interactions
- **14 API Endpoints**: Complete CRUD + business logic operations

**Key Features**:
- Flexible interaction types and directions
- Outcome tracking (successful, failed, no_answer, scheduled, cancelled)
- Multi-participant support via JSONB attendees array
- Follow-up scheduling and tracking
- Company association for B2B interactions
- Time tracking with started_at, ended_at, duration_sec

**Implementation Status**: All core business rules implemented. N+1 optimization with LEFT JOIN. Fully tested with 74 tests + 2 benchmarks.

**See**: [docs/concepts/interaction-management.md](docs/concepts/interaction-management.md) for complete Interaction Management guide

### Customer Analytics

Promenade implements **CQRS read models** for business intelligence and reporting:

- **9 Query Methods**: Optimized analytical queries (direct SQL, no repository pattern)
- **8 GET Endpoints**: Customer, deal, sales rep, revenue, and interaction analytics
- **Real-time Metrics**: Customer overview, deal pipeline, sales rep performance
- **Time Series**: Revenue trends with configurable granularity (day/week/month)
- **Funnel Analysis**: Customer lifecycle transitions and deal stage conversions
- **Performance Optimized**: Denormalized queries with LEFT JOIN across aggregates

**Key Features**:
- CQRS pattern separates read (analytics) from write (CRUD) models
- Direct database access bypasses repository abstraction for maximum query performance
- Aggregations span multiple entities (Customer + Deal + Interaction)
- No business logic in analytics (read-only reporting)

**Available Analytics**:
1. **Customer Overview** - Total customers, status/tier distribution, top sales reps
2. **Customer Lifecycle** - Conversion funnel (Lead → Prospect → Customer → Churned)
3. **Customer Segmentation** - Distribution by status, tier, lifetime value
4. **Deal Pipeline** - Pipeline by stage, win/loss rates, avg days to close
5. **Deal Conversions** - Stage-to-stage conversion rates, bottleneck identification
6. **Sales Rep Performance** - Active customers, deals won/lost, revenue, win rates
7. **Revenue Time Series** - Revenue over time with new customer tracking
8. **Interaction Insights** - Interaction counts by type/outcome, follow-up metrics

**Implementation Status**: All 8 endpoints live in production. 15 tests (7 unit + 8 integration). CQRS pattern fully implemented.

**See**: [Analytics README](internal/contexts/customer-mgmt/analytics/README.md) for complete API reference with request/response examples

### Order Management

Promenade provides **complete order lifecycle management** with state transitions and **fully implemented business rules**:

- **Order Creation**: Create orders with currency support for customers
- **Line Items**: Add/remove/update products with automatic total calculation
- **State Machine**: pending → confirmed → processing → fulfilled (or cancelled) - **fully enforced** ✅
- **Business Rules**: All validations implemented (line count, status checks, terminal state protection) ✅
- **Money Handling**: Type-safe Money value object (cents-based precision)
- **14 API Endpoints**: Complete CRUD + business logic operations
- **Auto-generated Numbers**: ORD-YYYY-NNNNNN format for easy tracking

**Implementation Status**: All core business rules implemented in entity (`Confirm()`, `StartProcessing()`, `MarkFulfilled()`, `Cancel()`). Future: Inventory/Payment/Shipping integrations.

**See**: [docs/concepts/order-management.md](docs/concepts/order-management.md) for complete Order Management guide

### Bounded Contexts

| Context                 | Aggregates                            | Description                   | Status           | Documentation                                  |
| ----------------------- | ------------------------------------- | ----------------------------- | ---------------- | ---------------------------------------------- |
| **Shared**              | Country, Currency, Language, Timezone | Reference data (read-only)    | ✅ Production     | [README](internal/contexts/shared/README.md)   |
| **Identity**            | User, Contact, Profile, Role, Permission | User management & RBAC     | ✅ Production     | [README](internal/contexts/identity/README.md) |
| **Customer Management** | Customer ✅, Company ✅, Deal ✅, Interaction ✅, Analytics ✅ | CRM, sales pipeline & BI    | ✅ Production     | [Guide](docs/concepts/customer-management.md) \| [Analytics](internal/contexts/customer-mgmt/analytics/README.md) |
| **Order Management**    | Order ✅, OrderLine ✅, Contract 📋, Fulfillment 📋 | Order processing       | ✅ Production     | [Guide](docs/concepts/order-management.md)      |
| **Billing**             | Invoice ✅, Payment ✅, Subscription 📋 | Billing and payments          | ✅ Production    | [Invoice Guide](docs/concepts/invoice-management.md) \| [Payment Guide](docs/concepts/payment-management.md)   |
| **Warehouse**           | Inventory, Stock                      | Inventory management          | 📋 Planned Q3 2026 | Coming soon                                 |

**Context Isolation**: Contexts communicate ONLY via Event Bus (no direct dependencies)

---

## Design Patterns & Conventions

Promenade follows **strict design patterns** and **naming conventions** to ensure consistency, maintainability, and code quality across the entire codebase.

### Core Patterns

**Repository Pattern** - Data access abstraction with `BaseRepository` for common operations:
- Interface: `IRepository` (e.g., `ICustomerRepository`)
- Implementation: Private struct with `BaseRepository` embedded
- Methods: `Create`, `GetByID`, `Update`, `Delete`, `List*`

**Use Case Pattern** - Business logic encapsulation:
- Interface: `IUseCase` (single interface per aggregate)
- Implementation: Private `useCase` struct (generic name)
- Constructor: Simple `NewUseCase()` (NOT `New{Entity}UseCase`)

**Handler Pattern** - HTTP endpoint handling:
- Struct: `{Entity}Handler` (e.g., `CustomerHandler`)
- Methods: Match HTTP verbs (`Create`, `GetByID`, `Update`, `Delete`, `List`)
- DTOs: Request/Response objects in `dto/` subdirectory

**Value Objects** - Immutable domain primitives:
- Email, Phone, Money, Address with validation
- Factory methods: `NewEmail()`, `NewPhone()`, `NewMoney()`
- Immutability: No setters, only getters

### Naming Conventions

**File Naming**:
- Entities: `entity.go`, `entity_test.go`
- Use Cases: `usecase.go`, `usecase_test.go`
- Repositories: `{aggregate}_repository.go`
- Handlers: `{aggregate}_handler.go`
- DTOs: `{aggregate}_dto.go`

**Database Naming**:
- Tables: `<context>_<aggregate>` (e.g., `customer_customers`, `order_orders`)
- Columns: `snake_case` (e.g., `user_id`, `created_at`, `is_active`)
- Indexes: `idx_<table>_<column>` (e.g., `idx_customers_email`)
- Foreign Keys: `fk_<table>_<ref_table>` (e.g., `fk_orders_customers`)

**Go Naming**:
- Interfaces: `I{Entity}Repository`, `I{Entity}UseCase`
- Implementations: lowercase `repository`, `useCase` (private structs)
- Constructors: `New{Entity}Repository()`, `NewUseCase()`
- Errors: `Err{Entity}{Condition}` (e.g., `ErrCustomerNotFound`)

**Detailed Documentation**:
- [Naming Conventions Guide](docs/guides/naming-conventions.md) - Files, directories, Go code
- [Database Conventions Guide](docs/guides/database-conventions.md) - Tables, columns, indexes, migrations
- [Architecture Patterns Guide](docs/guides/architecture-patterns.md) - Repository, UseCase, Handler implementations

---

## Project Structure

```
promenade/
 cmd/
    api/                    # HTTP server entry point (211 lines)
       main.go            # Bootstrap: Config → Logger → DB → Migrations → Event Bus → Server
    migrate/                # Migration CLI tool
 internal/
    contexts/               # Bounded Contexts (DDD)
       shared/            #  Shared Context (Reference Data: Country, Currency, Language, Timezone)
          README.md      # Complete documentation (~450 lines)
          domain/entity/ # 4 aggregates (Country, Currency, Language, Timezone)
          domain/repository/ # Repository interfaces
          usecase/       # Business logic (GetCountries, GetCurrencies, etc.)
          adapter/http/  # HTTP handlers (Gin)
          adapter/repository/postgres/ # PostgreSQL implementation
       identity/          #  Identity Context (User, Contact, Profile, Role, Permission)
          README.md      # Complete documentation (~550 lines)
          user/          #  User Aggregate (authentication, password management)
          contact/       #  Contact Aggregate (email, phone, address)
          profile/       #  Profile Aggregate (display name, bio, avatar)
          role/          #  Role Aggregate (RBAC roles)
          permission/    #  Permission Aggregate (RBAC permissions)
       customer-mgmt/     #  Customer Management Context (Customer)
       order-mgmt/        #  Order Management Context (Order, OrderLine)
       billing/           #  Billing Context (planned)
       README.md          # Contexts overview
    infrastructure/         # Cross-cutting concerns
        config/            # Configuration management (YAML + env vars)
        database/          # Database connection & transactions
 pkg/                        # Shared Domain Primitives & Utilities
    README.md              #  Complete package overview (~400 lines)
    bus/                   #  Event Bus (Memory/Redis, 67 tests, 377K/sec)
       README.md          #  Complete documentation (~600 lines)
    aggregate/             # Base Aggregate pattern (DDD)
    valueobject/           # Value Objects (Email, Phone, Money, Address)
    saga/                  # Saga orchestration ( planned Q1 2026)
    uuidv7/               # Time-ordered UUIDs (RFC 9562, 2x faster inserts)
    logger/               # Structured logging (slog wrapper, context-aware)
    migration/            # Namespace-based database migrations
    response/             # Standard HTTP responses & pagination
    jsonb/                # PostgreSQL JSONB utilities
    reference/            # Reference data validation helpers
 migrations/                 # Database migrations (namespace-based)
    core/                 # Core infrastructure (UUID v7, auth, RBAC)
    shared/               # Shared context migrations (reference data)
    identity/             # Identity context migrations (users, contacts)
    customer-mgmt/        # Customer Management context migrations
    order-mgmt/           # Order Management context migrations
 test/                       # Testing infrastructure
    README.md             # Testing structure documentation
    integration/          # Integration tests (real DB)
        testutils.go      # Shared test utilities
        contexts/         # Mirror path: repository integration tests
    benchmark/            # Benchmark tests (performance measurement)
 config/                     # Configuration files
    # PostgreSQL configurations
    app.postgres-dev.yaml   # PostgreSQL + development
    app.postgres-test.yaml  # PostgreSQL + testing
    app.postgres-prod.yaml  # PostgreSQL + production
    
    # SQLite configurations
    app.sqlite-dev.yaml     # SQLite + development
    app.sqlite-test.yaml    # SQLite + testing
    app.sqlite-prod.yaml    # SQLite + production
 docs/                       # Architecture documentation (8 guides)
     INDEX.md              # Documentation index
     CLEAN_ARCHITECTURE_SUMMARY.md # DDD with Bounded Contexts
     PHASE1_ARCHITECTURE_PREPARATION.md # Migration roadmap
     TESTING_PATTERNS.md  # Comprehensive testing guide
     TESTING_QUICK_REFERENCE.md # Testing cheat sheet
     BUS_TEST_COVERAGE.md # Event Bus test report
```

### Key Directories Explained

- **cmd/api/main.go**: Application entry point with proper initialization order
- **internal/contexts/**: Bounded contexts (autonomous, isolated)
- **pkg/**: Shared packages (context-agnostic, reusable across all contexts)
- **migrations/**: Namespace-based migrations (run in order: core → shared → identity)
- **test/**: Three-tier testing (unit in-place, integration/benchmark in mirror path)
- **config/**: Environment-specific YAML configs (dev/test/prod)

---

## Quick Start

### Prerequisites

- **Go 1.24+**
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
make migrate-postgres  # PostgreSQL migrations
make migrate-sqlite    # SQLite migrations
```

### 3. Start Application

```bash
# PostgreSQL development (Docker + migrations + API)
make dev-postgres

# Or build and run separately
make build
DATABASE_DRIVER=postgres ENVIRONMENT=development ./bin/promenade
```

Server starts on **http://localhost:8081**

### 3a. Alternative: SQLite Embedded Mode

**No Docker required** - perfect for demos, laptops, quick testing:

```bash
# Build binary
make build

# Start with embedded SQLite database
make dev-sqlite

# Or manually
DATABASE_DRIVER=sqlite ENVIRONMENT=development ./bin/promenade
```

Database creates at `./data/promenade.db` (auto-created directory).

**When to use SQLite**:
- Local development without Docker
- Sales demos on laptops
- Quick prototyping
- CI/CD testing
- Single-user deployments

**Limitations**:
- Single writer (not for production with multiple instances)
- No distributed transactions
- Limited to 1 concurrent connection

**See**: [config/SQLITE.md](config/SQLITE.md) for complete documentation

### 4. Health Check

```bash
curl http://localhost:8081/health
```

### 5. API Authentication

Promenade uses **JWT (JSON Web Tokens)** for authentication with access and refresh tokens.

**Register a new user:**

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Login to get JWT tokens:**

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-12-28T19:04:06Z",
    "token_type": "Bearer",
    "user": { ... }
  }
}
```

**Use access token for protected endpoints:**

```bash
curl -X GET http://localhost:8081/api/v1/identity/users \
  -H "Authorization: Bearer <access_token>"
```

**Refresh expired access token:**

```bash
curl -X POST http://localhost:8081/api/v1/identity/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh_token>"
  }'
```

**Token Configuration:**
- **Access Token**: 15 minutes (for API requests)
- **Refresh Token**: 7 days (to generate new access tokens)
- **Algorithm**: HS256 (HMAC with SHA-256)

**See**: [pkg/jwt/README.md](pkg/jwt/README.md) for detailed JWT documentation

---

### Available Make Commands

```bash
make help              # Show all available commands

# Workspace Configuration (configure once, work anywhere)
make switch-postgres-dev   # PostgreSQL + development
make switch-postgres-test  # PostgreSQL + test
make switch-postgres-prod  # PostgreSQL + production
make switch-sqlite-dev     # SQLite + development (no Docker needed)
make switch-sqlite-test    # SQLite + test
make workspace             # Show current workspace configuration

# Development Runners (environment-aware)
make dev                   # Start development server (requires ENVIRONMENT=development)
make dev-fresh             # Fresh start with clean database
make build                 # Build binary
make run                   # Build and run
make fmt                   # Format code
make lint                  # Run linters

# Testing (four-tier strategy)
make test-all              # All tests runner (warns if not test environment)
make test                  # All tests with race detector (~40s)
make test-unit             # Unit tests only (~5s)
make test-smoke            # Smoke tests (handler validation, no DB, ~2s)
make test-integration      # Integration tests with real DB (~14s)
make test-benchmark        # Benchmark tests (~5s per benchmark)
make test-coverage         # HTML coverage report

# Database Management
make docker-up             # Start database containers (database-aware)
make docker-down           # Stop database containers
make db-reset              # Drop and recreate database
make migrate               # Run all migrations (auto-detects driver)
make migrate-core          # Core migrations only

# Local CI Validation (run before push)
make pre-push              # Run all CI checks locally (lint + test + build)
make ci-lint               # Run golangci-lint (same as CI)
make ci-test               # Run all tests (same as CI)
make ci-build              # Test build (same as CI)

# Production
make prod                  # Production runner (requires ENVIRONMENT=production)
make swagger-all           # Generate API documentation
```

---

## Testing

Promenade uses a **four-tier testing strategy** with clear separation of concerns:

### Test Organization

**Four-tier architecture**:

1. **Unit Tests** (in-place) - Fast feedback, test individual components
2. **Smoke Tests** (`test/smoke/contexts/`) - HTTP handler validation (80/20 rule)
3. **Integration Tests** (`test/integration/contexts/`) - Full E2E with real database
4. **Benchmark Tests** (`test/benchmark/contexts/`) - Performance measurement with real DB

```
# Unit tests - alongside production code
internal/contexts/identity/contact/
 entity.go
 entity_test.go              #  Entity unit tests
 usecase.go
 usecase_test.go             #  UseCase unit tests

# Smoke tests - mirror path structure (NO DB, HTTP validation)
test/smoke/contexts/
 identity/
    user/handler_test.go      # 8 tests
    contact/handler_test.go   # 8 tests
    ...
 customer-mgmt/
    customer/handler_test.go  # 12 tests (most complex - 23-method mock)
    ...
 # Total: 123 tests across 15 handlers (100% pass rate) ✅

# Integration tests - mirror path structure (real DB)
test/integration/contexts/
 shared/
    country/repository_test.go     # 6 repository integration tests
    currency/repository_test.go    # 6 repository integration tests
    ...
 identity/
     contact/repository_test.go     # 9 repository integration tests
     profile/repository_test.go     # 17 repository integration subtests

# Benchmark tests - mirror path structure (real DB)
test/benchmark/contexts/
 identity/
    user/
       repository_bench_test.go    # Repository performance benchmarks
```

### Running Tests

```bash
# All tests (360+ tests, ~60 seconds with race detector)
make test-all               # Runner with environment check
make test                   # All tests with race detector

# By type (four-tier strategy)
make test-unit              # Unit tests only (~5s)
make test-smoke             # Smoke tests (handler validation, no DB, ~2s)
make test-integration       # Integration tests with real DB (~14s)
make test-benchmark         # Benchmark tests (5s per benchmark)
make test-benchmark-all     # Extended benchmarks (10s per benchmark)

# By context
go test ./test/integration/contexts/identity/... -v
go test ./test/smoke/contexts/... -v
go test -bench=. ./test/benchmark/contexts/identity/user -v

# By package
go test ./pkg/jwt/... -v
go test ./pkg/bus/... -v
go test ./pkg/logger/... -v

# With coverage
make test-coverage          # Generate HTML coverage report

# Local CI validation (before push)
make pre-push               # Run all CI checks (lint + test + build)
```

### Test Statistics

| Component                  | Tests | Coverage | Duration | Type        |
| -------------------------- | ----- | -------- | -------- | ----------- |
| **pkg/jwt**                | 18    | 87%      | cached   | Unit        |
| **pkg/bus**                | 67    | 100%     | ~2s      | Unit        |
| **pkg/logger**             | 15    | 95%      | cached   | Unit        |
| **pkg/uuidv7**             | 10    | 100%     | cached   | Unit        |
| **pkg/response**           | 13    | 100%     | cached   | Unit        |
| **pkg/saga**               | 28    | 100%     | cached   | Unit        |
| **pkg/migration**          | 8     | 90%      | cached   | Unit        |
| **pkg/valueobject**        | 45    | 95%      | cached   | Unit        |
| **pkg/middleware**         | 25    | 93%      | cached   | Unit        |
| **pkg/cache**              | 8     | 85%      | cached   | Unit        |
| **Smoke (all contexts)**   | 123   | -        | ~0.5s    | Smoke       |
| **Identity (integration)** | 35    | -        | ~2.8s    | Integration |
| **Shared (integration)**   | 24    | -        | ~7.8s    | Integration |
| **Customer (integration)** | 14    | -        | ~3.6s    | Integration |
| **Billing (integration)**  | 5     | -        | ~1.2s    | Integration |
| **Order-mgmt (integration)** | 6  | -        | ~1.5s    | Integration |
| **User ListUsers (bench)** | 4     | -        | ~5s      | Benchmark   |

**Total**: 360+ tests across 45+ packages, 90%+ average coverage

### Test Database

Integration and benchmark tests use a separate test database (automatically started):

```bash
# Start test database (PostgreSQL on 5433, Redis on 6380)
make test-db-start          # Automatically skipped in CI/CD environments

# Run integration tests (starts DB automatically)
make test-integration

# Run benchmark tests (requires test DB)
make test-benchmark

# Stop test database
make test-db-stop           # Automatically skipped in CI/CD environments
```

**CI/CD Awareness**: Test database commands detect CI environments (CI or GITHUB_ACTIONS) and skip Docker operations when PostgreSQL service is already running.

**Mirror Path Navigation**: Tests mirror production code structure for easy discovery

- `internal/contexts/shared/country/` → `test/integration/contexts/shared/country/`
- `internal/contexts/identity/user/` → `test/benchmark/contexts/identity/user/`

**See**: [test/README.md](test/README.md) for complete testing documentation

---

## Domain Model Examples

### Identity Context

#### Contact Aggregate

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

#### Profile Aggregate

```go
// Profile is an aggregate root for user profile information
type Profile struct {
    aggregate.BaseAggregate

    ID          uuidv7.UUID
    UserID      uuidv7.UUID // 1:1 with User
    DisplayName string      // public name (required)

    // Personal Info
    FirstName   string
    LastName    string
    Gender      Gender // male, female, other, not_specify
    DateOfBirth *time.Time

    // Business Info
    Bio       string // max 500 chars
    AvatarURL string

    // Localization
    Timezone string // IANA (e.g., "Europe/Kyiv")
    Language string // ISO 639-1 (e.g., "uk")
    Country  string // ISO 3166-1 (e.g., "UA")

    // Social Links (all require https://)
    Website, LinkedIn, Twitter, GitHub string

    // Status
    IsPublic   bool // profile visibility
    IsActive   bool // activation status
    IsBanned   bool // moderation flag
    IsVerified bool // verified badge
}

// Factory method
func NewProfile(userID uuidv7.UUID, displayName string) (*Profile, error) {
    if displayName == "" {
        return nil, fmt.Errorf("display name is required")
    }

    return &Profile{
        BaseAggregate: aggregate.NewBase(),
        ID:            uuidv7.New(),
        UserID:        userID,
        DisplayName:   displayName,
        IsPublic:      false,
        IsActive:      true,
    }, nil
}
```

### Value Objects

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

## Configuration

Promenade uses **YAML configuration files** per environment with **environment variable overrides**:

### Configuration Files

```
config/
 # PostgreSQL configurations
 app.postgres-dev.yaml   # PostgreSQL + development
 app.postgres-test.yaml  # PostgreSQL + testing
 app.postgres-prod.yaml  # PostgreSQL + production
 
 # SQLite configurations
 app.sqlite-dev.yaml     # SQLite + development
 app.sqlite-test.yaml    # SQLite + testing
 app.sqlite-prod.yaml    # SQLite + production
```

### Development Configuration

**config/app.postgres-dev.yaml**

```yaml
app:
  name: "Promenade Platform"
  environment: "development"
  version: "0.1.0"

server:
  host: "0.0.0.0"
  port: 8081
  read_timeout: 10s
  write_timeout: 10s

database:
  postgres:
    host: "localhost"
    port: 5432
    user: "system"
    password: "passw0rd"
    database: "promenade_dev"
    ssl_mode: "disable"
  redis:
    addr: "localhost:6379"
    password: ""
    pool_size: 10
    max_retries: 3
    databases:
      revocation: 0  # JWT token revocation
      bus: 1         # Event bus (if adapter=redis)
      cache: 2       # Application cache
      sessions: 3    # User sessions

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

**config/app.postgres-prod.yaml**

```yaml
# ... (same app/server sections)

database:
  postgres:
    host: "${DB_HOST}" # Environment variable override
    port: 5432
    user: "${DB_USER}"
    password: "${DB_PASSWORD}"
    database: "promenade_prod"
    ssl_mode: "require"
  redis:
    addr: "${REDIS_ADDR:-localhost:6379}"
    password: "${REDIS_PASSWORD}"
    pool_size: 20
    max_retries: 3
    databases:
      revocation: 0
      bus: 1
      cache: 2
      sessions: 3

# Redis Event Bus for production (distributed)
bus:
  adapter: "redis"
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

Environment is detected via `DATABASE_DRIVER` and `ENVIRONMENT` env vars:

```bash
DATABASE_DRIVER=postgres ENVIRONMENT=production ./bin/promenade   # Loads app.postgres-prod.yaml
DATABASE_DRIVER=sqlite ENVIRONMENT=test ./bin/promenade           # Loads app.sqlite-test.yaml
DATABASE_DRIVER=postgres ENVIRONMENT=development ./bin/promenade  # Loads app.postgres-dev.yaml (default)
```

**See**: [internal/infrastructure/config/](internal/infrastructure/config/) for config implementation

---

## Key Concepts

### Package Library

Promenade includes a comprehensive **package library** (`pkg/`) with reusable, context-agnostic components:

| Package         | Purpose                               | Tests | Status     | Documentation                           |
| --------------- | ------------------------------------- | ----- | ---------- | --------------------------------------- |
| **bus**         | Event Bus (Memory/Redis adapters)     | 67    | Production | [README](pkg/bus/README.md)             |
| **jwt**         | JWT authentication & RBAC middleware  | 18    | Production | [README](pkg/jwt/README.md)             |
| **logger**      | Structured logging (slog wrapper)     | 15    | Production | [README](pkg/logger/README.md)          |
| **middleware**  | HTTP middleware (rate limit, CSRF)    | 25    | Production | [README](pkg/middleware/README.md)      |
| **cache**       | Redis-based caching layer             | 8     | Production | [README](pkg/cache/README.md)           |
| **migration**   | Namespace-based DB migrations         | 8     | Production | [README](pkg/migration/README.md)       |
| **response**    | Standard HTTP responses               | 12    | Production | [README](pkg/response/README.md)        |
| **uuidv7**      | Time-ordered UUIDs (RFC 9562)         | 10    | Production | [README](pkg/uuidv7/README.md)          |
| **valueobject** | DDD Value Objects                     | 25    | Production | [README](pkg/valueobject/README.md)     |
| **aggregate**   | Base Aggregate pattern                | 5     | Production | [README](pkg/aggregate/README.md)       |
| **jsonb**       | PostgreSQL JSONB utilities            | 8     | Production | [README](pkg/jsonb/README.md)           |
| **saga**        | Distributed transaction orchestration | 28    | Production | [README](pkg/saga/README.md)            |

**Total**: 229+ tests across 12 packages

**Key Highlights**:

- **bus**: Central event-driven communication hub (377K events/sec with Memory adapter)
- **jwt**: JWT token generation/validation with RBAC middleware (87% test coverage)
- **middleware**: Rate limiting and CSRF protection for API security (93% test coverage)
- **uuidv7**: Time-ordered UUIDs provide 2x faster inserts than UUID v4
- **valueobject**: Email, Phone, Money, Address with immutability and validation
- **aggregate**: Base pattern for all domain aggregates (event sourcing support)

**See**: [pkg/README.md](pkg/README.md) for complete package library documentation

---

## Domain Model Example: Identity Context

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

## Database

### Migrations

Migrations are namespace-based for context isolation:

```
migrations/
 core/                        # Core infrastructure
    000001_core_init_uuid_v7.up.sql
    000002_core_auth_full.up.sql
    000003_core_rbac_full.up.sql
 identity/                    # Identity context
     000001_identity_contacts.up.sql
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

## Roadmap

### Phase 1: Foundation (Completed)

- [x] DDD primitives (Aggregate, Value Objects, Saga)
- [x] Project structure (Bounded Contexts)
- [x] Event Bus (Memory + Redis adapters)
- [x] Database migrations system
- [x] Testing infrastructure (four-tier strategy: unit, smoke, integration, benchmark)

### Phase 2: Identity Context (Completed ✅)

- [x] Contact aggregate (email, phone, address)
- [x] Profile aggregate (personal info, social links, localization)
- [x] User aggregate (registration, authentication, password management)
- [x] Repository implementations (PostgreSQL)
- [x] HTTP API (REST with Gin) - all 3 aggregates
- [x] Unit tests (85+ tests per aggregate)
- [x] Integration tests (32 User + 9 Contact + 17 Profile subtests)
- [x] Password policies (8+ chars, digit, letter, bcrypt hashing)
- [x] Account management (status: active/suspended/banned, locking after failed logins)

### Phase 3: Authentication & Authorization (Completed)

- [x] JWT authentication (token generation, validation)
- [x] JWT middleware for protected endpoints
- [x] Session management (Redis storage, TTL)
- [x] Token refresh mechanism
- [x] Role-Based Access Control (RBAC)
- [x] Role & Permission aggregates
- [x] RBAC middleware
- [x] Role management API (7 endpoints)
- [x] Permission management API (7 endpoints)

### Phase 4: Customer Management Context (Completed ✅)

- [x] Customer aggregate (lifecycle, segmentation)
- [x] Company aggregate (B2B support, 14 endpoints)
- [x] Deal aggregate (pipeline, stages, 12 endpoints)
- [x] Interaction aggregate (calls, emails, meetings, 14 endpoints) - **COMPLETED Dec 31, 2025**

### Phase 5: Order Management Context (Completed ✅)

- [x] Order aggregate (creation, fulfillment, 14 endpoints)
- [x] OrderLine entity (integrated with Order)
- [ ] Contract aggregate - Planned Q1 2026
- [ ] Fulfillment saga (payment → inventory → shipping) - Planned Q2 2026

### Phase 6: Analytics & Reporting (Planned Q2 2026)

- [ ] CQRS read models
- [ ] Dashboards
- [ ] Business metrics

---

## Contributing

This is a learning project focused on DDD architecture. Contributions welcome!

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Follow DDD principles and existing patterns
4. Write tests (maintain 80%+ coverage)
5. Commit changes (`git commit -m 'Add amazing feature'`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Open Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Email**: alexander.vasilenko@gmail.com

---

## **Built with Domain-Driven Design and Go**

## Documentation

### Complete Documentation Library

Promenade includes **comprehensive documentation** covering all aspects of the architecture:

**Core Documentation**:

- [Documentation Index](docs/INDEX.md) - Quick navigation hub with concepts and links
- [Clean Architecture with DDD](docs/concepts/clean-architecture.md) - Bounded Contexts, Aggregates, Value Objects
- [Event-Driven Architecture](docs/concepts/event-driven.md) - Event Bus, Domain Events, Sagas
- [Bounded Contexts Strategy](docs/concepts/bounded-contexts.md) - Context isolation and communication

**Implementation Guides**:

- [RBAC Implementation](docs/guides/rbac.md) - Roles, permissions, JWT integration
- [Rate Limiting](docs/guides/rate-limiting.md) - IP-based protection for authentication
- [Health Checks](docs/guides/health-checks.md) - Dependency monitoring and alerting
- [Local CI Validation](docs/guides/local-ci.md) - Run GitHub Actions checks locally before push
- [Testing Patterns](docs/guides/testing-patterns.md) - Four-tier testing strategy (~965 lines)
- [Testing Quick Reference](docs/guides/testing-quick-reference.md) - One-page cheat sheet
- [Smoke Tests Guide](test/smoke/README.md) - Complete HTTP handler validation guide (~450 lines)

**Technical Reference**:

- [Test Coverage Report](docs/reference/test-coverage-report.md) - 360+ tests breakdown
- [Bus Test Coverage](docs/reference/bus-test-coverage.md) - Event Bus test report (67 tests, 100% passing)
- [Refactoring Roadmap](docs/reference/refactoring-roadmap.md) - Technical debt and improvements

**Bounded Contexts**:

- [Contexts Overview](internal/contexts/README.md) - All bounded contexts catalog
- [Identity Context](internal/contexts/identity/README.md) - User, Contact, Profile, RBAC (~550 lines)
- [Shared Context](internal/contexts/shared/README.md) - Reference data (Country, Currency, Language, Timezone) (~450 lines)
- [Customer Management](internal/contexts/customer-mgmt/README.md) - Customer aggregate

**Package Library**:

- [Package Overview](pkg/README.md) - All shared packages documentation (~400 lines)
- [Event Bus](pkg/bus/README.md) - Central communication hub (Memory/Redis adapters) (~600 lines)
- [JWT Authentication](pkg/jwt/README.md) - Token generation, validation, RBAC middleware
- [UUID v7](pkg/uuidv7/README.md) - Time-ordered UUIDs for better performance
- [Logger](pkg/logger/README.md) - Structured logging with context
- [Value Objects](pkg/valueobject/README.md) - Email, Phone, Money, Address

**Infrastructure**:

- [Configuration](internal/infrastructure/config/README.md) - YAML config management
- [Database](internal/infrastructure/database/README.md) - PostgreSQL connection and transactions
- [Health Checks](internal/infrastructure/health/README.md) - Dependency monitoring
- [Testing Guide](test/README.md) - Testing structure and best practices
- [Migrations](migrations/README.md) - Namespace-based migration system

**Total Documentation**: 8 core guides + 6 context READMEs with examples, best practices, and architecture decisions

---
