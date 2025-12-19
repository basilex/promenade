# Promenade

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Production-ready REST API built with **Clean Architecture**, featuring PostgreSQL with UUID v7, comprehensive testing infrastructure, JWT authentication, and API versioning.

## Key Features

- **Clean Architecture** - Clear separation of concerns (Domain, Use Case, Adapter, Infrastructure)
- **UUID v7 Primary Keys** - Time-ordered UUIDs for optimal performance (2x faster than v4)
- **RBAC System** - Role-Based Access Control with wildcard permissions and 5 system roles
- **Structured Logging** - slog with JSON/text format, context fields (request_id, user_id)
- **Comprehensive Testing** - 388 tests total across all layers - 100% passing
  - Unit: 183 tests (entity validation + domain logic)
  - Integration: 91 tests (repository operations with real PostgreSQL)
  - Smoke: 114 tests (end-to-end critical flows with real database)
- **JWT Authentication** - Secure token-based auth with refresh tokens
- **API Versioning** - v1 and v2 with backward compatibility
- **PostgreSQL + sqlx** - No ORM, pure SQL with transaction support
- **Swagger Documentation** - Auto-generated API docs for both versions
- **Docker Ready** - Full Docker Compose setup for development and testing
- **Database Migrations** - golang-migrate for version control
- **High Performance** - Gin framework with graceful shutdown

## Quick Start

### Prerequisites

- **Go 1.21+**
- **Docker & Docker Compose**
- **Make**
- **golang-migrate** (optional, installed via `make install`)

### Installation

```bash
# Clone the repository
git clone https://github.com/basilex/promenade.git
cd promenade

# Install dependencies and tools
make install

# Start PostgreSQL via Docker
make docker-up

# Run database migrations
make migrate-up

# Start development server
make dev
```

Server will start on http://localhost:8081

### Default Login Credentials

The system includes pre-configured users with different RBAC roles for development and testing:

```bash
# System Administrator (superadmin role - bootstrap user)
Email:    system@promenade.com
Password: passw0rd
Role:     Full system access (*)

# Super Administrator (superadmin role - testing)
Email:    superadmin@promenade.com
Password: passw0rd
Role:     Full system access (*)

# Administrator (admin role)
Email:    admin@promenade.com
Password: passw0rd
Role:     User/content management

# Moderator (moderator role)
Email:    moderator@promenade.com
Password: passw0rd
Role:     Content moderation

# Regular User (user role)
Email:    alexander.vasilenko@gmail.com
Password: 03041965
Role:     Basic user operations
```

**[!] Important:** Change these passwords before deploying to production!

-> **Full credentials reference:** See [docs/CREDENTIALS.md](docs/CREDENTIALS.md) for complete list with API examples.

### Docker Setup (Alternative)

Run everything in Docker containers:

```bash
# Build and run all services (PostgreSQL, Redis, Migrations, API)
make docker-run

# Or step by step:
make docker-build VERSION=0.1.0 ENV=dev
make docker-up

# Check health
curl http://localhost:8080/api/v1/health
```

**Database Auto-Creation:** PostgreSQL container automatically creates three databases on first startup:

- `promenade_prod` - Production database (with migrations applied via migrate service)
- `promenade_dev` - Development database (requires manual `make migrate-up` for local dev)
- `promenade_test` - Test database (used by integration tests)

**Clean Slate:** Use `make docker-clean` to remove all containers and volumes, then `make docker-up` to recreate with fresh databases.

-> **Docker details**: See [docker/README.md](docker/README.md) for comprehensive Docker documentation.

### Quick Test

```bash
# Run all tests (unit + integration)
make test

# Or run integration tests only
make test-integration

# Test authentication with default user (local dev)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}'

# Test authentication (Docker)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}'
```

## -> API Access

### API Endpoints

- **GET /api** - API information with all available versions

  ```bash
  curl http://localhost:8081/api
  # Returns: service info, available versions (v1, v2), links to documentation
  ```

- **GET /api/v1** - API v1 information

  ```bash
  curl http://localhost:8081/api/v1
  # Returns: version info, base path, documentation, health check links
  ```

- **GET /api/v2** - API v2 information
  ```bash
  curl http://localhost:8081/api/v2
  # Returns: version info, base path, documentation, health check links
  ```

### Quick Links

**Local Development** (port 8081):

- **API Root**: http://localhost:8081/api
- **API v1 Base**: http://localhost:8081/api/v1
- **API v2 Base**: http://localhost:8081/api/v2
- **Swagger v1**: http://localhost:8081/api/v1/docs/swagger/index.html
- **Swagger v2**: http://localhost:8081/api/v2/docs/swagger/index.html
- **Health Check**: http://localhost:8081/api/v1/health

**Docker** (port 8080):

- **API Root**: http://localhost:8080/api
- **Swagger v1**: http://localhost:8080/api/v1/docs/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health

### Swagger UI Authentication

**How to authenticate in Swagger UI (browser):**

1. **Login** via `/api/v1/auth/login` endpoint:

   ```json
   {
     "email": "system@promenade.com",
     "password": "passw0rd"
   }
   ```

2. **Copy the `access_token`** from response (e.g., `eyJhbGciOiJIUzI1NiIs...`)

3. **Click the "Authorize" button** 🔓 (green lock icon at the top right)

4. **In the "Value" field, enter**:

   ```
   Bearer eyJhbGciOiJIUzI1NiIs...
   ```

   **⚠️ IMPORTANT:** You must type the word `Bearer`, then a space, then your token

5. **Click "Authorize"** and close the dialog

6. **All protected endpoints** (with 🔒 icon) will now work automatically

**Common mistake:** Entering just the token without `Bearer` prefix results in `401 Unauthorized`.

**cURL equivalent** (for comparison):

```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq -r '.data.access_token')

# Use token (notice "Bearer " prefix)
curl http://localhost:8081/api/v1/roles \
  -H "Authorization: Bearer $TOKEN"
```

### Error Handling

All errors return structured JSON responses via `ErrorHandler` in shared layer:

- **404 Not Found** - Route doesn't exist:

  ```json
  {
    "success": false,
    "message": "route not found",
    "error": "The requested endpoint does not exist",
    "timestamp": 1766125580
  }
  ```

- **405 Method Not Allowed** - HTTP method not supported:
  ```json
  {
    "success": false,
    "message": "method not allowed",
    "error": "The HTTP method is not supported for this endpoint",
    "timestamp": 1766125580
  }
  ```

**Server Timestamp:** All responses include `timestamp` field (Unix seconds) for:

- Client-server time synchronization
- Latency measurement (client_time - server_timestamp)
- Debugging and log correlation across timezones
- Cache freshness detection

**Implementation:** See `internal/adapter/http/shared/handler/error_handler.go` - centralized HTTP error handling with unit test coverage.

## -> Documentation

### Technical Documentation

- **[Authorization Guide](docs/AUTHORIZATION.md)** - RBAC middleware with permissions, roles, and usage examples
- **[Logging Guide](docs/LOGGING.md)** - Structured logging with slog (JSON/text format, context fields)
- [Testing Guide](docs/TESTING_GUIDE.md) - Comprehensive testing setup and best practices
- [Testing Infrastructure](docs/TESTING_INFRASTRUCTURE.md) - Test infrastructure overview
- [Validation](docs/VALIDATION.md) - Multi-layer validation strategy and best practices
- [UUID v7 Migration](docs/UUID_V7_MIGRATION.md) - Migrating from UUID v4 to v7
- [ID Strategies](docs/ID_STRATEGIES.md) - Primary key strategy recommendations
- [Auth Schema](docs/AUTH_SCHEMA.md) - Database schema for authentication system
- [Test Results](docs/TEST_RESULTS.md) - Current test coverage and results
- [User Profiles Test Results](docs/USER_PROFILES_TEST_RESULTS.md) - User profiles module test coverage (72 tests)

**Language Policy:** All documentation and code comments are in English. Russian versions (.ru.md) are kept for reference.

### Database Schema

Complete database schema with relationships and key constraints:

```mermaid
erDiagram
    users ||--o{ user_sessions : "has many"
    users ||--o| user_profiles : "has one"
    users ||--o{ user_contacts : "has many"
    users ||--o{ user_posts : "creates"
    users ||--o{ post_comments : "writes"
    users ||--o{ comment_likes : "likes"
    users ||--o{ password_reset_tokens : "requests"
    users ||--o{ email_verification_tokens : "receives"
    users ||--o{ login_attempts : "attempts"
    users ||--o{ user_roles : "has roles"
    users ||--o{ user_roles : "assigns roles (assigned_by)"

    roles ||--o{ user_roles : "assigned to users"
    roles ||--o{ role_permissions : "has permissions"
    permissions ||--o{ role_permissions : "granted to roles"

    user_profiles }o--|| countries : "located in"

    user_posts ||--o{ post_comments : "has comments"
    post_comments ||--o{ post_comments : "has replies"
    post_comments ||--o{ comment_likes : "receives likes"

    countries ||--o{ country_currencies : "uses"
    currencies ||--o{ country_currencies : "used by"

    users {
        uuid id PK "UUID v7"
        varchar email UK "unique email"
        varchar name
        varchar password "bcrypt hash"
        user_status status "enum"
        timestamptz email_verified_at
        text suspended_reason
        timestamptz suspended_until
        timestamptz last_login_at
        timestamptz created_at
        timestamptz updated_at
    }

    user_sessions {
        uuid id PK "UUID v7"
        uuid user_id FK
        varchar refresh_token UK "hashed JWT"
        text user_agent
        inet ip_address
        timestamptz expires_at
        timestamptz created_at
    }

    user_profiles {
        uuid id PK "UUID v7"
        uuid user_id FK,UK "one-to-one"
        varchar first_name
        varchar last_name
        varchar nickname UK "@username"
        text bio
        date date_of_birth
        varchar gender
        uuid country_id FK
        varchar city
        varchar timezone "IANA format"
        jsonb social_links
        varchar avatar_url
        boolean is_public
        boolean is_verified
        integer profile_views_count
        integer followers_count
        timestamptz last_seen_at
        timestamptz created_at
        timestamptz updated_at
    }

    user_contacts {
        uuid id PK "UUID v7"
        uuid user_id FK
        varchar contact_type "email|phone|telegram..."
        varchar contact_value
        varchar label "Work|Personal|Emergency"
        boolean is_verified
        boolean is_primary
        boolean is_public
        time available_from
        time available_to
        varchar[] available_days
        timestamptz created_at
        timestamptz updated_at
    }

    user_posts {
        uuid id PK "UUID v7"
        uuid user_id FK
        varchar title
        varchar slug UK "per user"
        text excerpt
        text content
        jsonb featured_image
        post_status status "draft|published|archived|scheduled"
        boolean is_public
        boolean is_featured
        timestamptz published_at
        timestamptz scheduled_at
        jsonb tags
        integer view_count
        integer like_count
        integer comment_count
        integer reading_time_minutes
        timestamptz deleted_at "soft delete"
        timestamptz created_at
        timestamptz updated_at
    }

    post_comments {
        uuid id PK "UUID v7"
        uuid post_id FK
        uuid user_id FK
        uuid parent_id FK "self-reference for replies"
        text content "1-5000 chars"
        boolean is_edited
        timestamptz edited_at
        integer like_count
        integer reply_count
        timestamptz deleted_at "soft delete"
        timestamptz created_at
        timestamptz updated_at
    }

    comment_likes {
        uuid comment_id PK,FK
        uuid user_id PK,FK
        timestamptz created_at
    }

    countries {
        uuid id PK "UUID v7"
        varchar name
        varchar code
        char iso2 UK "2-letter code"
        char iso3 UK "3-letter code"
        country_region region "enum"
        timestamptz created_at
        timestamptz updated_at
    }

    currencies {
        uuid id PK "UUID v7"
        varchar name
        varchar code UK "USD|EUR|GBP..."
        varchar symbol "$|€|£..."
        timestamptz created_at
        timestamptz updated_at
    }

    country_currencies {
        uuid country_id PK,FK
        uuid currency_id PK,FK
        boolean is_primary
        timestamptz created_at
    }

    password_reset_tokens {
        uuid id PK "UUID v7"
        uuid user_id FK
        varchar token UK
        boolean used
        timestamptz used_at
        timestamptz expires_at
        timestamptz created_at
    }

    email_verification_tokens {
        uuid id PK "UUID v7"
        uuid user_id FK
        varchar token UK
        boolean used
        timestamptz used_at
        timestamptz expires_at
        timestamptz created_at
    }

    permissions {
        uuid id PK "UUID v7"
        varchar resource "posts|users|comments..."
        varchar action "create|read|update|delete|*"
        text description
        timestamptz created_at
    }

    roles {
        uuid id PK "UUID v7"
        varchar name UK "superadmin|admin|moderator..."
        varchar display_name "Human-readable name"
        text description
        boolean is_system "Cannot be deleted"
        timestamptz created_at
        timestamptz updated_at
    }

    role_permissions {
        uuid role_id PK,FK
        uuid permission_id PK,FK
    }

    user_roles {
        uuid user_id PK,FK
        uuid role_id PK,FK
        timestamptz assigned_at
        uuid assigned_by FK "User who assigned"
        timestamptz expires_at "Optional expiration"
    }

    login_attempts {
        uuid id PK "UUID v7"
        uuid user_id FK
        inet ip_address
        text user_agent
        boolean successful
        text failure_reason
        timestamptz attempted_at
    }
```

**Key Features:**

- **UUID v7** for all primary keys (time-ordered, better performance than UUID v4)
- **RBAC (Role-Based Access Control)** - Flexible permission system with wildcard support (`*:*`, `posts:*`)
  - 5 system roles: superadmin, admin, moderator, user, guest
  - Granular permissions: 33 predefined permissions (users:create, posts:delete, etc.)
  - Optional role expiration for temporary access grants
- **Soft deletes** on user posts and comments (`deleted_at`)
- **Nested comments** via self-referencing `parent_id` in `post_comments`
- **JSONB** for flexible data (social links, preferences, tags, featured images)
- **Enums** for type safety (`user_status`, `post_status`, `country_region`)
- **Composite primary keys** for junction tables (`comment_likes`, `country_currencies`, `role_permissions`, `user_roles`)
- **Cascading deletes** to maintain referential integrity
- **Unique constraints** to prevent duplicates (email, nickname, slug per user, permission resource:action)

## Development Commands

**Modular Makefile System** - Commands organized by context (see [MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md))

### Quick Reference

```bash
make help              # Show all available commands (grouped by module)
make dev               # Start development server (most common workflow)
make test              # Run all tests (unit + integration + smoke)
make docker-run        # Build and run in Docker
```

### Development Workflow (Makefile.dev.mk)

```bash
make install           # Install tools (swag, migrate, golangci-lint)
make dev               # Start server (postgres + migrations + app)
make build             # Build binary to bin/promenade
make run               # Run compiled binary
make lint              # Run golangci-lint
make fmt               # Format code (go fmt + gofmt -s)
make generate          # Generate entity boilerplate
make config-show       # Show current configuration
```

### Testing (Makefile.test.mk)

```bash
make test              # Run all tests (unit + integration)
make test-unit         # Unit tests only (~5s)
make test-integration  # Integration tests (~35s, auto-starts DB)
make test-smoke        # Smoke tests (~4.5s, critical flows)
make test-coverage     # Generate HTML coverage report
make test-db-start     # Start test DB (port 5433)
make test-db-stop      # Stop test DB
```

### Production/DevOps (Makefile.prod.mk)

```bash
# Docker
make docker-build      # Build image (VERSION=0.1.0 ENV=dev)
make docker-run        # Build + start containers
make docker-up         # Start services
make docker-down       # Stop services
make docker-logs       # View logs
make docker-clean      # Remove containers + volumes

# Migrations
make migrate-up        # Apply migrations
make migrate-down      # Rollback last migration
make migrate-create    # Create migration (NAME=xxx)

# Documentation
make swagger-all       # Generate v1 + v2 Swagger docs

# Cleanup
make clean             # Remove artifacts
```

## Architecture

This project strictly follows **Clean Architecture** principles with four distinct layers:

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Handlers (Gin)                      │ ← Adapter Layer
│                    DTOs, Mappers, Routes                     │
├─────────────────────────────────────────────────────────────┤
│                      Use Cases                              │ ← Use Case Layer
│              Business Logic Orchestration                    │
├─────────────────────────────────────────────────────────────┤
│                    Domain Entities                          │ ← Domain Layer
│              Repository Interfaces (Ports)                   │
├─────────────────────────────────────────────────────────────┤
│             Repository Implementations                       │ ← Infrastructure
│              PostgreSQL, Config, JWT                         │
└─────────────────────────────────────────────────────────────┘
```

### Layers Explained

| Layer              | Responsibility                            | Dependencies       |
| ------------------ | ----------------------------------------- | ------------------ |
| **Domain**         | Business entities & repository interfaces | None (independent) |
| **Use Case**       | Application business rules                | Domain only        |
| **Adapter**        | HTTP handlers, DTOs, mappers              | Use Case, Domain   |
| **Infrastructure** | Database, config, external services       | Domain interfaces  |

**Dependency Rule**: Inner layers never depend on outer layers. Dependencies point inward.

### Project Structure

```
promenade/
├── cmd/api/                           # Application entry point
│   └── main.go                        # Bootstrap, DI, server setup
├── internal/
│   ├── domain/
│   │   ├── entity/                    # Business entities (User, Permission, Role, UserProfile, UserContact, UserPost, Country, Currency, Session)
│   │   └── repository/                # Repository interfaces (ports)
│   ├── usecase/                       # Business logic orchestration
│   │   ├── auth_usecase.go           # Login, register, refresh, logout
│   │   ├── permission_usecase.go     # RBAC permissions management
│   │   ├── role_usecase.go           # RBAC roles management
│   │   ├── user_contact_usecase.go   # User contacts management
│   │   ├── user_profile_usecase.go   # User profiles, privacy, moderation
│   │   ├── user_post_usecase.go      # Blog posts, publishing, engagement
│   │   ├── country_usecase.go        # Countries CRUD
│   │   └── currency_usecase.go       # Currencies CRUD
│   ├── adapter/
│   │   ├── http/
│   │   │   ├── shared/middleware/    # Auth, RBAC authorization, CORS, logging, recovery
│   │   │   ├── v1/                   # API v1 (handlers, DTOs, routes)
│   │   │   └── v2/                   # API v2 (handlers, DTOs, routes)
│   │   └── repository/postgres/      # Repository implementations (sqlx)
│   └── infrastructure/
│       ├── config/                    # Configuration loader
│       ├── database/                  # PostgreSQL connection & transactions
│       └── logger/                    # Structured logging
├── pkg/
│   ├── jwt/                          # JWT token manager
│   ├── ptr/                          # Reference helpers for nullable fields
│   ├── uuidv7/                       # UUID v7 generator
│   ├── pagination/                   # Pagination helpers
│   └── validator/                    # Request validation
├── test/
│   ├── helpers/                      # Test database setup & fixtures
│   │   ├── database.go              # TestDB with cleanup
│   │   └── fixtures.go              # User & session fixtures
│   ├── integration/                  # Integration tests (planned)
│   ├── e2e/                         # End-to-end tests (planned)
│   └── mocks/                       # Mock repositories (UserProfile, etc.)
├── migrations/                       # Database migrations (golang-migrate)
├── docker/
│   ├── docker-compose.yml           # Dev database (port 5432)
│   ├── docker-compose.test.yml      # Test database (port 5433)
│   └── Dockerfile                   # Production image
├── docs/                            # Technical documentation
├── scripts/                         # Helper scripts & generators
└── Makefile                         # Development commands
```

### \* Working with Nullable Fields (`pkg/ref`)

When working with database entities that have nullable fields (mapped to SQL `NULL`), Go requires using pointer types (`*string`, `*time.Time`, etc.). The `pkg/ref` package provides convenient helpers to avoid verbose manual pointer creation and prevent common mistakes.

#### Why Reference Helpers?

```go
// [X] Manual approach - verbose and error-prone
reason := "Violation of terms"
user.SuspendedReason = &reason  // Easy to forget "*" after 8 hours at computer

until := time.Now().Add(7 * 24 * time.Hour)
user.SuspendedUntil = &until

// [+] Using ref package - clean and safe
user.SuspendedReason = ref.String("Violation of terms")
user.SuspendedUntil = ref.Time(time.Now().Add(7 * 24 * time.Hour))
```

#### Available Helpers

**Constructor functions** (value → pointer):

```go
ref.String(s string) *string                     // "hello" → *"hello"
ref.Time(t time.Time) *time.Time                // time.Now() → *time.Now()
ref.UUID(u uuidv7.UUID) *uuidv7.UUID            // uuid → *uuid
ref.Int(i int) *int                             // 42 → *42
ref.Bool(b bool) *bool                          // true → *true
```

**Safe getters** (pointer → value with defaults):

```go
ref.StringValue(s *string) string               // nil → "", *"hello" → "hello"
ref.TimeValue(t *time.Time) time.Time          // nil → time.Time{}, *now → now
ref.UUIDValue(u *uuidv7.UUID) uuidv7.UUID      // nil → uuid.UUID{}, *id → id
ref.IntValue(i *int) int                       // nil → 0, *42 → 42
ref.BoolValue(b *bool) bool                    // nil → false, *true → true
```

**Getters with custom defaults**:

```go
ref.StringOr(s *string, default string) string
ref.TimeOr(t *time.Time, default time.Time) time.Time
ref.IntOr(i *int, default int) int
ref.BoolOr(b *bool, default bool) bool
```

**Utility functions**:

```go
ref.IsNil[T any](p *T) bool                    // Check if pointer is nil
ref.IsSet(s *string) bool                      // Check if string pointer is not nil AND not empty
```

#### Real-World Examples

**Creating entities with nullable fields**:

```go
// User suspension
user.Suspend(reason, ref.Time(time.Now().Add(7*24*time.Hour)))

// User profile with optional fields
profile := &entity.UserProfile{
    UserID:      userID,
    Nickname:    "john_doe",
    Bio:         ref.String("Software engineer"),
    DateOfBirth: ref.Time(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)),
    CountryID:   ref.UUID(countryUUID),
    Timezone:    "America/New_York",
}
```

**Test assertions**:

```go
// [X] Old way - manual dereferencing
assert.Equal(t, "Violation of terms", *user.SuspendedReason)
assert.Equal(t, expectedTime, *user.SuspendedUntil)

// [+] New way - safe getters
assert.Equal(t, "Violation of terms", ref.StringValue(user.SuspendedReason))
assert.Equal(t, expectedTime, ref.TimeValue(user.SuspendedUntil))
```

**Conditional logic**:

```go
// Check if bio is set and not empty
if ref.IsSet(profile.Bio) {
    // Display bio
    fmt.Println("Bio:", ref.StringValue(profile.Bio))
} else {
    // Show default message
    fmt.Println("Bio: Not provided")
}

// Get value with fallback
displayName := ref.StringOr(profile.DisplayName, profile.Nickname)
```

#### Benefits

- **Type-safe** - Compiler catches mismatches
- -> **Less verbose** - No temporary variables needed
- - **Intention-clear** - `ref.String("value")` explicitly shows nullable intent
- [!] **Fewer bugs** - Eliminates "forgot to add `*`" mistakes after long coding sessions
- - **Test-friendly** - Safe dereferencing in assertions without panic risk

#### Nullable Fields Philosophy

```go
// Use regular types for required fields (NOT NULL in database)
Email    string       // Always has value, minimum ""
Name     string       // Required
CreatedAt time.Time   // NOT NULL DEFAULT NOW()

// Use pointers for optional fields (NULL in database)
MiddleName     *string     // nil = not set, &"" = empty, &"John" = value
LastLoginAt    *time.Time  // nil = never logged in, &time = last login time
SuspendedUntil *time.Time  // nil = not suspended, &time = suspended until
```

This approach allows distinguishing between three states:

1. **Not set** (nil) - field was never provided
2. **Empty** (&"") - field was explicitly cleared
3. **Value** (&"text") - field has actual data

## 📨 Event Bus System

The project implements a **transport-agnostic event bus** for asynchronous communication between components. This enables scalable, loosely-coupled architecture within the monolith, with a clear path to microservices when needed.

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      Application Layer                          │
│  ┌──────────────┐         ┌──────────────┐                     │
│  │ AuthUseCase  │────────>│  Event Bus   │<────────┐           │
│  └──────────────┘ Publish └──────────────┘ Subscribe           │
│       User                       │                  │           │
│    Registered                    │             ┌────▼────────┐  │
│                                  │             │   Email     │  │
│                                  │             │  Service    │  │
│                                  │             └─────────────┘  │
│                             ┌────▼────────┐                     │
│                             │  Analytics  │                     │
│                             │  Service    │                     │
│                             └─────────────┘                     │
└─────────────────────────────────────────────────────────────────┘

Flow:
1. User registers → AuthUseCase publishes UserRegisteredEvent
2. Event Bus dispatches to all subscribers (non-blocking)
3. Email Service sends welcome email in background
4. Analytics Service tracks registration metrics
5. Main request returns immediately (user doesn't wait)
```

### Key Benefits

| Benefit              | Description                                                        |
| -------------------- | ------------------------------------------------------------------ |
| **Non-blocking**     | Publish() returns immediately - user doesn't wait for side effects |
| **Decoupling**       | Use cases don't know about email/analytics - just publish events   |
| **Scalability**      | Easy to add new subscribers without changing existing code         |
| **Evolution Path**   | Start with in-memory bus → Redis Pub/Sub → NATS → Kafka            |
| **Testing**          | Mock bus for tests, no actual email/analytics calls                |
| **Async Processing** | Worker pool handles events concurrently in goroutines              |

### Available Implementations

#### 1. In-Memory Bus (Current)

**Use case**: Development, testing, and simple monolith deployments

```go
// Production code (cmd/api/main.go)
busConfig := bus.NewBusConfig(
    cfg.Bus.WorkerPoolSize,  // From .env: BUS_WORKER_POOL_SIZE
    cfg.Bus.BufferSize,      // From .env: BUS_BUFFER_SIZE
    cfg.Bus.RetryAttempts,   // From .env: BUS_RETRY_ATTEMPTS
    cfg.Bus.RetryDelay,      // From .env: BUS_RETRY_DELAY
)
eventBus := memory.NewMemoryBus(busConfig)
defer eventBus.Close(ctx)
```

**Features**:

- [+] Zero external dependencies
- [+] Configurable worker pool (10 goroutines default)
- [+] Buffered message queue (1000 messages default)
- [+] Graceful shutdown with proper cleanup
- [+] Built-in health checks and statistics
- [!] No persistence - events lost on restart
- [!] Single-process only - not suitable for horizontal scaling

#### 2. Redis Pub/Sub (Planned)

**Use case**: Multi-instance deployments with shared state

```go
// Future implementation
busConfig := bus.NewBusConfig(...)
eventBus := redis.NewRedisBus(busConfig, redisClient)
```

**Features**:

- [+] Multi-process support (horizontal scaling)
- [+] Pub/Sub pattern for real-time delivery
- [!] No guaranteed delivery - subscribers must be online
- [!] No message persistence after delivery

#### 3. NATS/Kafka (Future)

**Use case**: Microservices with guaranteed delivery

```go
// Future implementation
eventBus := nats.NewNATSBus(busConfig, natsConn)
// or
eventBus := kafka.NewKafkaBus(busConfig, kafkaProducer)
```

**Features**:

- [+] Persistent message storage
- [+] At-least-once delivery guarantees
- [+] Message replay capability
- [+] Dead letter queues for failures
- [!] More complex infrastructure

### Configuration

Event bus is configured via environment variables in `.env` files:

```bash
# .env.development / .env.production
BUS_WORKER_POOL_SIZE=10     # Concurrent workers processing events
BUS_BUFFER_SIZE=1000        # Internal message queue size
BUS_RETRY_ATTEMPTS=3        # Retry failed handlers
BUS_RETRY_DELAY=1s          # Delay between retries
```

**Config struct** (`internal/infrastructure/config/config.go`):

```go
type BusConfig struct {
    WorkerPoolSize int           // Number of concurrent workers
    BufferSize     int           // Message buffer capacity
    RetryAttempts  int           // Max retry attempts on failure
    RetryDelay     time.Duration // Delay between retries
}
```

### Publishing Events

**Step 1: Define domain event** (`internal/domain/event/user_events.go`):

```go
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

**Step 2: Publish from use case** (`internal/usecase/auth_usecase.go`):

```go
func (uc *authUseCase) Register(ctx context.Context, email, name, password string) (*entity.User, error) {
    // 1. Create user in database
    user, err := uc.userRepo.Create(ctx, newUser)
    if err != nil {
        return nil, err
    }

    // 2. Publish event (non-blocking)
    userEvent := event.NewUserRegisteredEvent(user.ID, user.Email, user.Name)
    if err := uc.eventBus.Publish(ctx, userEvent.Type(), userEvent); err != nil {
        logger.Error("Failed to publish user registered event",
            slog.Any("error", err),
            slog.String("user_id", user.ID.String()))
        // Don't fail registration if event publishing fails
    }

    // 3. Return immediately - email is sent in background
    return user, nil
}
```

**Key principle**: Event publishing errors are **logged but not propagated**. The main operation (user registration) should succeed even if events fail.

### Subscribing to Events

**Email notification service** (`internal/infrastructure/notification/email_service.go`):

```go
type EmailService struct {
    bus           bus.Bus
    sender        EmailSender
    templates     *template.Template
}

// Start subscribes to events
func (s *EmailService) Start(ctx context.Context) error {
    // Subscribe to multiple topics
    s.bus.Subscribe(bus.TopicUserRegistered, s.handleUserRegistered)
    s.bus.Subscribe(bus.TopicUserEmailVerified, s.handleUserEmailVerified)
    s.bus.Subscribe(bus.TopicUserPasswordChanged, s.handleUserPasswordChanged)
    s.bus.Subscribe(bus.TopicUserSuspended, s.handleUserSuspended)
    s.bus.Subscribe(bus.TopicUserBanned, s.handleUserBanned)
    return nil
}

// Handle user registration event
func (s *EmailService) handleUserRegistered(ctx context.Context, e bus.Event) error {
    evt, ok := e.(*event.UserRegisteredEvent)
    if !ok {
        return fmt.Errorf("unexpected event type: %T", e)
    }

    // Render HTML template
    data := map[string]interface{}{
        "Name":     evt.Name,
        "Email":    evt.Email,
        "UserID":   evt.UserID.String(),
        "LoginURL": s.appURL + "/api/v1/auth/login",  // From config (APP_URL)
        "AppName":  s.appName,                        // From config (APP_NAME)
        "Year":     time.Now().Year(),
    }

    html, err := s.renderTemplate("welcome.html", data)
    if err != nil {
        return fmt.Errorf("failed to render template: %w", err)
    }

    // Send email (happens in background goroutine)
    email := Email{
        To:      evt.Email,
        Subject: "Welcome to Promenade!",
        HTML:    html,
    }

    return s.sender.Send(ctx, email)
}
```

**Production initialization** (`cmd/api/main.go`):

```go
// Initialize Event Bus
busConfig := bus.NewBusConfig(
    cfg.Bus.WorkerPoolSize,  // From .env: BUS_WORKER_POOL_SIZE=10
    cfg.Bus.BufferSize,      // From .env: BUS_BUFFER_SIZE=1000
    cfg.Bus.RetryAttempts,   // From .env: BUS_RETRY_ATTEMPTS=3
    cfg.Bus.RetryDelay,      // From .env: BUS_RETRY_DELAY=1s
)
eventBus := memory.NewMemoryBus(busConfig)
defer eventBus.Close(context.Background())

// Initialize Email Service with config
emailSender := notification.NewMockEmailSender() // TODO: replace with real SMTP in production
emailService, err := notification.NewEmailService(
    eventBus,
    emailSender,
    "templates/email",          // Template directory
    cfg.Email.FromAddress,      // From .env: EMAIL_FROM_ADDRESS
    cfg.Email.FromName,         // From .env: EMAIL_FROM_NAME
    cfg.Email.AppURL,           // From .env: APP_URL
    cfg.Email.AppName,          // From .env: APP_NAME
)
if err != nil {
    logger.Fatal("Failed to create email service", slog.Any("error", err))
}

// Start service (subscribes to events)
if err := emailService.Start(context.Background()); err != nil {
    logger.Fatal("Failed to start email service", slog.Any("error", err))
}

logger.Info("Email notification service started (async via event bus)",
    slog.String("templates_path", "templates/email"),
    slog.String("from_address", cfg.Email.FromAddress),
    slog.String("app_url", cfg.Email.AppURL))
```

**Environment configuration** (`.env.development`):

```bash
# Email Configuration
EMAIL_FROM_ADDRESS=noreply@promenade.com
EMAIL_FROM_NAME=Promenade Team

# Application
APP_NAME=Promenade
APP_URL=http://localhost:8081
```

### Email Templates

Templates are externalized in `templates/email/` directory:

```
templates/email/
├── welcome.html              # New user registration
├── email_verified.html       # Email verification success
├── password_changed.html     # Security alert
├── account_suspended.html    # Temporary suspension
├── account_banned.html       # Permanent ban
└── README.md                 # Template documentation
```

**Benefits of external templates**:

- [+] Change email design without redeploying service
- [+] Professional HTML emails with CSS styling
- [+] A/B testing different email variants
- [+] Designer-friendly (no Go code required)
- [+] Version control for email content

**Example template** (`templates/email/welcome.html`):

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <style>
      body {
        font-family: Arial, sans-serif;
      }
      .header {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      }
      .button {
        padding: 12px 30px;
        background: #667eea;
        color: white;
      }
    </style>
  </head>
  <body>
    <div class="header">
      <h1>🎉 Welcome to Promenade!</h1>
    </div>
    <div class="content">
      <p>Hi <strong>{{.Name}}</strong>,</p>
      <p>Your account has been successfully created.</p>
      <a href="{{.LoginURL}}" class="button">Start Exploring →</a>
    </div>
    <div class="footer">
      <p>© {{.Year}} Promenade. All rights reserved.</p>
    </div>
  </body>
</html>
```

### Standard Topics

Predefined topic constants in `pkg/bus/topics.go`:

```go
const (
    // User Management
    TopicUserRegistered      = "user.registered"
    TopicUserActivated       = "user.activated"
    TopicUserSuspended       = "user.suspended"
    TopicUserBanned          = "user.banned"
    TopicUserEmailVerified   = "user.email.verified"
    TopicUserPasswordChanged = "user.password.changed"

    // Content Management
    TopicPostPublished       = "post.published"
    TopicPostUnpublished     = "post.unpublished"
    TopicCommentAdded        = "comment.added"
    TopicCommentRemoved      = "comment.removed"
)
```

**Naming convention**: `{domain}.{entity}.{action}` for clarity and consistency.

### Testing

#### Unit Tests with Mock Bus

```go
func TestAuthUseCase_Register(t *testing.T) {
    mockBus := new(MockBus)
    authUC := NewAuthUseCase(userRepo, sessionRepo, jwtManager, mockBus)

    // Test registration logic
    user, err := authUC.Register(ctx, "test@example.com", "John", "password")
    require.NoError(t, err)

    // Verify event was published (but don't actually send email)
    mockBus.AssertCalled(t, "Publish", mock.Anything, "user.registered", mock.Anything)
}
```

#### Integration Tests with Real Bus

```go
func TestEventBusIntegration(t *testing.T) {
    // Use in-memory bus with empty template path (fallback templates)
    eventBus := memory.NewDefaultMemoryBus()
    defer eventBus.Close(context.Background())

    emailSender := notification.NewMockEmailSender()
    // Parameters: eventBus, sender, templatesPath, fromAddress, fromName, appURL, appName
    emailService, err := notification.NewEmailService(
        eventBus,
        emailSender,
        "",                          // empty = use fallback templates
        "noreply@promenade.com",     // from config: EMAIL_FROM_ADDRESS
        "Promenade Team",            // from config: EMAIL_FROM_NAME
        "http://localhost:8081",     // from config: APP_URL
        "Promenade",                 // from config: APP_NAME
    )
    require.NoError(t, err)
    require.NoError(t, emailService.Start(ctx))

    // Publish event
    userEvent := event.NewUserRegisteredEvent(uuid.New(), "test@example.com", "John")
    err = eventBus.Publish(ctx, userEvent.Type(), userEvent)
    require.NoError(t, err)

    // Wait for async processing
    time.Sleep(100 * time.Millisecond)

    // Verify email was "sent" to mock
    require.Equal(t, 1, len(emailSender.GetSentEmails()))
    assert.Equal(t, "test@example.com", emailSender.GetSentEmails()[0].To)
}
```

### Performance & Monitoring

#### Bus Statistics

```go
stats := eventBus.Stats()
fmt.Printf("Topics: %d\n", stats["total_topics"])
fmt.Printf("Subscribers: %d\n", stats["total_subscribers"])
fmt.Printf("Messages Published: %d\n", stats["messages_published"])
fmt.Printf("Messages Processed: %d\n", stats["messages_processed"])
```

#### Health Check

```go
if err := eventBus.Health(ctx); err != nil {
    logger.Error("Event bus is unhealthy", slog.Any("error", err))
}
```

#### Performance Characteristics

**In-Memory Bus**:

- **Latency**: < 1ms to publish (returns immediately)
- **Throughput**: 10,000+ events/sec with default config
- **Worker Pool**: Limits concurrent processing (prevents resource exhaustion)
- **Graceful Shutdown**: Waits for in-flight events to complete

**Tuning for production**:

```bash
# High-volume deployment
BUS_WORKER_POOL_SIZE=50      # More concurrent handlers
BUS_BUFFER_SIZE=10000        # Larger queue for spikes
BUS_RETRY_ATTEMPTS=5         # More aggressive retries
BUS_RETRY_DELAY=2s           # Longer backoff
```

### Migration Path: Monolith → Microservices

#### Phase 1: Monolith with Event Bus (Current)

```
┌─────────────────────────────────┐
│         Monolith Process        │
│  ┌──────────┐   ┌────────────┐ │
│  │ Use Case │──>│ In-Memory  │ │
│  └──────────┘   │    Bus     │ │
│                 └────────────┘ │
│                       │         │
│                 ┌─────▼──────┐  │
│                 │   Email    │  │
│                 │  Service   │  │
│                 └────────────┘  │
└─────────────────────────────────┘
```

#### Phase 2: Multi-Instance with Redis

```
┌──────────────┐         ┌──────────────┐
│ Instance 1   │         │ Instance 2   │
│ ┌──────────┐ │         │ ┌──────────┐ │
│ │ Use Case │─┼───┐ ┐───┼─│ Email    │ │
│ └──────────┘ │   │ │   │ │ Service  │ │
└──────────────┘   │ │   │ └──────────┘ │
                   ▼ ▼   └──────────────┘
              ┌──────────┐
              │  Redis   │
              │ Pub/Sub  │
              └──────────┘
```

#### Phase 3: Microservices with NATS/Kafka

```
┌────────────┐    ┌─────────┐    ┌─────────────┐
│    API     │───>│  NATS/  │───>│   Email     │
│  Service   │    │  Kafka  │    │ Microservice│
└────────────┘    └─────────┘    └─────────────┘
                       │
                       └─────────>┌─────────────┐
                                  │ Analytics   │
                                  │ Microservice│
                                  └─────────────┘
```

**Key insight**: Same event publishing code works across all phases. Only the bus implementation changes:

```go
// Phase 1: In-memory
eventBus := memory.NewMemoryBus(config)

// Phase 2: Redis
eventBus := redis.NewRedisBus(config, redisClient)

// Phase 3: NATS
eventBus := nats.NewNATSBus(config, natsConn)
```

### Best Practices

1. **Events are immutable** - Never modify event after publishing
2. **Events are facts** - Past tense naming (`UserRegistered`, not `RegisterUser`)
3. **Idempotent handlers** - Handle duplicate events gracefully
4. **Don't fail operations on event errors** - Log and continue
5. **Keep events small** - Only essential data (use IDs, not full objects)
6. **Version events** - Add `EventVersion` field for schema evolution
7. **Monitor dead letters** - Track and retry failed events
8. **Use correlation IDs** - Trace events across services

### Demo

Run the event bus demo to see async email notifications in action:

```bash
cd /Users/basilex/Workspace/src/promenade
go run examples/event_bus_demo/main.go
```

**Output**:

```
* Event Bus Demo - Async Email Notifications
================================================

[+] Email service started and listening for events...

-> Simulating user registration: Demo User (demo@example.com)
[+] Event published to bus (returns immediately)
⏳ Email being sent in background goroutine...

-> Emails sent: 1
  1. To: demo@example.com
     Subject: Welcome to Promenade!

-> Event Bus Stats:
  Topics: 5
  Subscribers: 5
  Messages Published: 1
  Messages Processed: 1

* Key Takeaways:
   • Publish() returns immediately - non-blocking
   • Email sent asynchronously in worker pool
   • User doesn't wait for email delivery
```

### Further Reading

- **[Event Bus README](pkg/bus/README.md)** - Detailed technical documentation
- **[Configuration Guide](docs/CONFIGURATION_REFACTORING.md)** - Template and config externalization
- **[Integration Tests](test/integration/event_bus_test.go)** - Full test suite examples

## Configuration

### Environment Files

The project uses a hierarchical environment configuration:

| File               | Purpose              | Committed? | Priority   |
| ------------------ | -------------------- | ---------- | ---------- |
| `.env`             | Base defaults        | [+] Yes    | Lowest     |
| `.env.development` | Development settings | [+] Yes    | Medium     |
| `.env.local`       | Personal overrides   | [X] No     | Highest    |
| `.env.production`  | Production secrets   | [X] No     | Production |

### Key Configuration Variables

```bash
# Application
APP_NAME=Promenade
APP_URL=http://localhost:8081

# Server Configuration
SERVER_PORT=8081
ENVIRONMENT=development           # development, production

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=system
DB_PASSWORD=password
DB_NAME=promenade_dev
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# JWT Configuration
# Generate with: openssl rand -base64 32
JWT_SECRET=your-super-secret-key-change-in-production
JWT_ACCESS_TTL_MINUTES=15        # 15 minutes
JWT_REFRESH_TTL_HOURS=168        # 7 days (168 hours)

# Swagger Configuration
SWAGGER_ENABLED=true
SWAGGER_HOST=localhost:8081

# Rate Limiting
RATE_LIMIT_RPS=100               # Requests per second
RATE_LIMIT_BURST=200             # Burst capacity

# Event Bus Configuration
BUS_WORKER_POOL_SIZE=10          # Concurrent workers
BUS_BUFFER_SIZE=1000             # Message buffer capacity
BUS_RETRY_ATTEMPTS=3             # Max retry attempts on failure
BUS_RETRY_DELAY=1s               # Delay between retries

# Email Configuration
EMAIL_FROM_NAME=Promenade Team
EMAIL_FROM_ADDRESS=noreply@promenade.com
```

### Setup Local Environment

```bash
# Copy example file
cp .env.local.example .env.local

# Edit with your settings
vim .env.local

# Values in .env.local override all other env files
```

### Loading Priority

```
.env.local (highest priority)
    ↓
.env.development
    ↓
.env (base defaults)
```

## Testing

Promenade features a **comprehensive testing infrastructure** with isolated test database and helper utilities.

### Test Statistics

- **388 Total Tests** - 100% passing [+]
  - **183 Unit Tests** (entity validation, domain logic)
    - Country, Currency, Session, UserContact, UserPost, UserProfile, User entities
    - Permission, Role, RBAC system validation
    - Business logic, status transitions, timestamps
  - **91 Integration Tests** (with real PostgreSQL)
    - BaseRepository: 7 tests (transactions, executor pattern)
    - Auth: 10 tests (sessions, token management)
    - Countries & Currencies: 4 tests (CRUD operations)
    - User Management: 33 tests (users, profiles, contacts)
    - Content: 37 tests (posts, comments, likes, replies)
  - **114 Smoke Tests** (end-to-end critical flows)
    - Auth: 8 scenarios (registration → login → sessions → logout)
    - Country/Currency: 12 scenarios (complete CRUD operations)
    - User Management: 35 scenarios (profiles, contacts, posts)
    - Content: 31 scenarios (posts, comments, likes)
    - RBAC: 41 scenarios (permissions, roles, assignments, wildcards)
- **Test Database** - PostgreSQL 16 on port 5433 (isolated from dev DB)
- **Test Execution** - ~45 seconds for full suite (5s unit + 36s integration + 4s smoke)
- | **Coverage**         | Unit Tests | Integration Tests | Smoke Tests | Total   |
  | -------------------- | ---------- | ----------------- | ----------- | ------- |
  | **Country/Currency** | 18         | 4                 | 12          | 34      |
  | **Session/Auth**     | 9          | 10                | 8           | 27      |
  | **UserContact**      | 14         | 13                | 11          | 38      |
  | **UserPost**         | 42         | 30                | 12          | 84      |
  | **PostComment**      | 0          | 22                | 13          | 35      |
  | **UserProfile**      | 31         | 13                | 12          | 56      |
  | **User/RBAC**        | 64         | 9                 | 0           | 73      |
  | **Permission/Role**  | 31         | 16                | 41          | 88      |
  | **CommentLikes**     | 0          | 1                 | 5           | 6       |
  | **BaseRepository**   | 0          | 7                 | 0           | 7       |
  | **Total**            | **183**    | **91**            | **114**     | **388** |

_Note: Smoke tests provide end-to-end verification of critical user flows with real database operations. Tests include table-driven tests with multiple scenarios per function._

### Running Tests

```bash
# Quick test - all tests with auto DB setup
make test                  # Run unit + integration tests (274 tests)

# Individual test suites
make test-unit            # Unit tests only (no database, 183 tests)
make test-integration     # Integration tests (real PostgreSQL, 91 tests)
make test-smoke           # Smoke tests (end-to-end flows, 114 tests)

# Coverage and monitoring
make test-coverage        # Generate HTML coverage report
make test-watch           # Watch mode (re-run on file changes)

# Test database management
make test-db-start        # Start test PostgreSQL (port 5433)
make test-db-stop         # Stop test database
```

### Test Infrastructure

The project includes production-grade test helpers:

```go
// test/helpers/database.go
testDB := helpers.SetupTestDB(t)        // Connect to test database
defer testDB.Close()
defer testDB.CleanupTables(t)           // Clean all tables after test

// test/helpers/fixtures.go
user := helpers.UserFixture()                    // Standard active user
customUser := helpers.UserFixture(func(u *entity.User) {
    u.Email = "custom@test.com"                 // Override fields
})
session := helpers.SessionFixture(user.ID)       // Active session
expiredSession := helpers.ExpiredSessionFixture(user.ID)
```

### Test Examples

**Repository Integration Test:**

```go
func TestUserRepository_Create(t *testing.T) {
    t.Run("creates user successfully", func(t *testing.T) {
        testDB := helpers.SetupTestDB(t)
        defer testDB.Close()
        defer testDB.CleanupTables(t)

        repo := postgres.NewUserRepository(testDB.DB)
        user := helpers.UserFixture()

        err := repo.Create(context.Background(), user)
        require.NoError(t, err)
        assert.NotEqual(t, uuid.Nil, user.ID)
    })
}
```

### \* Smoke Tests

Production-ready **end-to-end smoke tests** verify critical user flows with real database operations. These tests ensure core functionality works correctly in integration.

**Test Suite** (`test/smoke/`):

| Test File                        | Scenarios | Coverage                                                                          |
| -------------------------------- | --------- | --------------------------------------------------------------------------------- |
| `auth_smoke_test.go`             | 8         | Registration, login, GetMe, refresh, logout, sessions, duplicate validation       |
| `country_currency_smoke_test.go` | 12        | Country & Currency CRUD (create, read, update, delete, list, code lookup)         |
| `comment_likes_smoke_test.go`    | 5         | Like/unlike comments, pagination, deleted comments, performance (100 checks)      |
| `rbac_smoke_test.go`             | 28        | Permissions, roles, user assignments, wildcards, expiration, RBAC checks          |
| `rbac_integration_smoke_test.go` | 13        | Real-world RBAC: moderator ban, admin feature, creator restrictions, cross-checks |
| `post_comment_smoke_test.go`     | 13        | Comment CRUD, threading, replies, nested replies, pagination, soft delete, auth   |
| `user_profile_smoke_test.go`     | 12        | Profile CRUD, privacy, verification, ban/unban, views, last seen, list, search    |
| `user_post_smoke_test.go`        | 12        | Post CRUD, draft/publish, featured, schedule, views, soft delete, list, search    |
| `user_contact_smoke_test.go`     | 11        | Contact CRUD (email, phone, telegram), primary, verification, visibility, delete  |
| **Total**                        | **114**   | **All tests passing [+] (9 test files, production-grade coverage)**               |

**Running Smoke Tests:**

```bash
# Run all smoke tests (recommended - includes DB setup)
make test-smoke

# Or run manually with go test
go test -v -count=1 ./test/smoke

# Run specific smoke test
go test -v ./test/smoke -run TestCommentLikes_SmokeTest

# Skip in short mode
go test -short ./test/smoke  # Smoke tests are skipped
```

**Example Smoke Tests:**

```go
// RBAC Smoke Test - comprehensive permission and role management
func TestRBAC_SmokeTest(t *testing.T) {
    // Tests: permissions, roles, user assignments, wildcards, expiration
    t.Run("[+] Create_custom_permissions", func(t *testing.T) {
        perm, err := permUC.CreatePermission(ctx, "invoices", "read", "Can read invoices")
        require.NoError(t, err)
    })

    t.Run("[+] Assign_permissions_to_role", func(t *testing.T) {
        err := roleUC.SyncRolePermissions(ctx, roleID, permIDs)
        require.NoError(t, err)
    })

    t.Run("[+] Check_wildcard_permission", func(t *testing.T) {
        hasPerm, err := roleUC.HasPermission(ctx, userID, "reports:create")
        assert.True(t, hasPerm, "via wildcard reports:*")
    })
}

// Comment Likes Smoke Test - like/unlike flow with performance check
func TestCommentLikes_SmokeTest(t *testing.T) {
    t.Run("[+] Basic_like_flow", func(t *testing.T) {
        // Like comment → verify count → unlike → verify again
    })

    t.Run("* Performance", func(t *testing.T) {
        // Execute 100 HasUserLiked queries and measure performance
    })
}
```

**Key Features:**

- [+] Real database integration (PostgreSQL on port 5433)
- [+] Isolated test data with automatic cleanup
- [+] Critical path verification (create → retrieve → update → delete)
- [+] Performance benchmarks included (100 permission checks in <500ms)
- [+] Fast execution (~4 seconds for all 114 scenarios across 9 test files)
- [+] 100% passing rate with comprehensive coverage
- [+] Idempotent tests with cleanup at start and end (CleanupTables)

See [TESTING_GUIDE.md](docs/TESTING_GUIDE.md) for comprehensive testing documentation.

## API Endpoints & Manual Testing

### Endpoint Coverage

The API provides **79 REST endpoints** across 8 modules with comprehensive functionality.

| Module         | Endpoints | Tested | Status   | Description                                              |
| -------------- | --------- | ------ | -------- | -------------------------------------------------------- |
| **Auth**       | 11        | 6      | [+] 55%  | Registration, login, logout, refresh, session management |
| **Profiles**   | 13        | 4      | [+] 31%  | User profiles with privacy settings and moderation       |
| **Contacts**   | 9         | 3      | [+] 33%  | User contact management (email, phone, social)           |
| **Posts**      | 18        | 5      | [!] 28%  | Blog posts with publishing, scheduling, engagement       |
| **Comments**   | 9         | 7      | [+] 56%  | Threaded comments with likes and moderation              |
| **Countries**  | 9         | 4      | [+] 44%  | Country management with currency relationships           |
| **Currencies** | 9         | 7      | [+] 78%  | Currency management with country relationships           |
| **Health**     | 1         | 1      | [+] 100% | Service health check                                     |
| **TOTAL**      | **79**    | **37** | **47%**  | Core functionality fully operational                     |

### Manual Testing Results (Coffee Tests ☕)

Comprehensive manual testing was performed on all critical endpoints to verify production readiness.

#### [+] Tested & Verified Endpoints

**Authentication Flow (6/11):**

- [+] `POST /api/v1/auth/register` - User registration with validation
- [+] `POST /api/v1/auth/login` - JWT authentication with access + refresh tokens
- [+] `POST /api/v1/auth/refresh` - Token refresh with rotation
- [+] `POST /api/v1/auth/logout` - Refresh token invalidation
- [+] `GET /api/v1/auth/me` - Current user information
- [+] `GET /api/v1/auth/sessions` - Active session listing

**User Profiles (4/13):**

- [+] `POST /api/v1/profiles` - Profile creation with privacy settings
- [+] `GET /api/v1/profiles/me` - Current user profile
- [+] `PUT /api/v1/profiles/:id` - Profile updates
- [+] `GET /api/v1/profiles` - Profile listing with pagination

**User Contacts (3/9):**

- [+] `POST /api/v1/users/contacts` - Contact creation (email, phone, social)
- [+] `GET /api/v1/users/contacts` - User contact listing
- [+] `PUT /api/v1/users/contacts/:id` - Contact updates

**Blog Posts (5/18):**

- [+] `POST /api/v1/posts` - Post creation with draft status
- [+] `POST /api/v1/posts/:id/publish` - Post publishing
- [+] `GET /api/v1/posts/published` - Published posts listing
- [!] `POST /api/v1/posts/:id/like` - Post engagement (partially tested)
- [!] `GET /api/v1/posts/:id` - Single post retrieval (needs data)

**Comments (7/9):**

- [+] `POST /api/v1/comments` - Comment creation
- [+] `PUT /api/v1/comments/:id` - Comment updates
- [+] `GET /api/v1/comments/:id` - Comment retrieval
- [+] `POST /api/v1/comments/:id/like` - Comment likes
- [+] `GET /api/v1/comments?post_id=xxx` - Post comments (query param based)
- `POST /api/v1/comments` (replies) - Thread replies (initiated)
- `GET /api/v1/comments/:id/replies` - Reply listing (initiated)

**RBAC (Role-Based Access Control) (18/18):**

- [+] `POST /api/v1/permissions` - Create permission (requires permissions:create)
- [+] `GET /api/v1/permissions` - List all permissions (requires permissions:read)
- [+] `GET /api/v1/permissions/:id` - Get permission by ID (requires permissions:read)
- [+] `PUT /api/v1/permissions/:id` - Update permission (requires permissions:update)
- [+] `DELETE /api/v1/permissions/:id` - Delete permission (requires permissions:delete)
- [+] `GET /api/v1/permissions?resource=posts` - Find by resource (requires permissions:read)
- [+] `POST /api/v1/roles` - Create role (requires roles:create)
- [+] `GET /api/v1/roles` - List all roles (requires roles:read)
- [+] `GET /api/v1/roles/:id` - Get role by ID (requires roles:read)
- [+] `PUT /api/v1/roles/:id` - Update role (requires roles:update)
- [+] `DELETE /api/v1/roles/:id` - Delete role (requires roles:delete, prevents system roles deletion)
- [+] `POST /api/v1/roles/:id/permissions` - Add permission to role (requires roles:update)
- [+] `DELETE /api/v1/roles/:id/permissions/:permissionId` - Remove permission (requires roles:update)
- [+] `POST /api/v1/roles/:id/permissions/sync` - Sync all permissions (requires roles:update)
- [+] `GET /api/v1/roles/:id/permissions` - List role permissions (requires roles:read)
- [+] `POST /api/v1/roles/:id/users/:userId` - Assign role to user (requires roles:assign)
- [+] `DELETE /api/v1/roles/:id/users/:userId` - Remove role from user (requires roles:assign)
- [+] `GET /api/v1/users/:id/roles` - Get user roles (requires users:read or own user)

### Known Issues Fixed During Testing

1. **Route Conflict in PostCommentRouter** [+] FIXED

   - **Issue**: Path conflict between `/posts/:post_id/comments` and `/posts/:id`
   - **Solution**: Changed to query parameter `GET /api/v1/comments?post_id=xxx`
   - **Impact**: Prevents Gin router panic on startup

2. **UserContactHandler Type Conversion (9 occurrences)** [+] FIXED

   - **Issue**: Incorrect type assertion `userID.(string)` instead of `userID.(uuidv7.UUID)`
   - **Solution**: Fixed all 9 methods in handler
   - **Impact**: Prevents runtime panic on all contact endpoints

3. **ToPostListResponse Nil Pointer** [+] FIXED
   - **Issue**: Accessing `meta.Total` when `meta == nil` caused panic
   - **Solution**: Added nil check before accessing pagination metadata
   - **Impact**: Critical - prevented server crash on `/posts/published`

### API Authentication

All protected endpoints require JWT Bearer token:

```bash
# Get token via login
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Use token in subsequent requests
curl http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

**Token Lifecycle:**

- **Access Token**: Stateless JWT, valid for 1 hour, cannot be revoked
- **Refresh Token**: Stored in database, valid for 7 days, can be revoked via logout
- **Token Rotation**: Each refresh generates new access + refresh token pair

### Production Readiness [+]

- [+] All core user flows tested and operational
- [+] Authentication & authorization working correctly
- [+] Request validation functioning properly
- [+] Error handling comprehensive and informative
- [+] Structured logging capturing all requests
- [+] Panic recovery tested and working
- [+] 328+ automated tests (100% passing)
- [+] 26 endpoints manually verified via HTTP

**Recommended for production use** with continued monitoring and testing of remaining endpoints.

## Structured Logging

Promenade uses **slog** (Go 1.21+ structured logging) for production-ready observability.

### Key Features

- [+] **Structured Format** - JSON (production) or text (development)
- [+] **Context Fields** - Automatic request_id, user_id, trace_id tracking
- [+] **Zero Dependencies** - Built-in Go stdlib
- [+] **Performance** - Zero allocation for common operations
- [+] **Integration Ready** - Works with ELK, Loki, DataDog, New Relic

### Log Output Examples

**Development (text format):**

```
time=2025-12-16T16:55:51+02:00 level=INFO source=promenade/pkg/logger/logger.go:201
msg="Server started" port=8081 environment=development
```

**Production (JSON format):**

```json
{
  "time": "2025-12-16T16:55:51+02:00",
  "level": "INFO",
  "msg": "HTTP request",
  "method": "GET",
  "path": "/api/countries",
  "status": 200,
  "duration": 5830791,
  "request_id": "45400d3f-e98f-4c18-bf57-e9cc719df3e7"
}
```

### Usage

```go
import (
    "log/slog"
    "github.com/basilex/promenade/pkg/logger"
)

// Simple logging
logger.Info("User registered",
    slog.String("email", user.Email),
    slog.String("user_id", userID),
)

// Context-aware (auto-includes request_id)
logger.InfoContext(ctx, "Processing payment",
    slog.String("amount", "100.00"),
    slog.String("currency", "USD"),
)

// Error logging
logger.Error("Database query failed",
    slog.Any("error", err),
    slog.String("query", sqlQuery),
    slog.Duration("duration", queryTime),
)
```

See **[LOGGING.md](docs/LOGGING.md)** for complete guide with examples for all layers.

## Docker Deployment

### Development

```bash
# Start all services (PostgreSQL + migrations)
make docker-up

# View logs
make docker-logs

# Stop all services
make docker-down
```

### Production Build

```bash
# Build optimized Docker image
docker build -t promenade:latest -f docker/Dockerfile .

# Run with production env
docker run -d \
  --name promenade-api \
  -p 8081:8081 \
  --env-file .env.production \
  promenade:latest

# Or use Docker Compose for full stack
docker-compose -f docker/docker-compose.prod.yml up -d
```

### Multi-Stage Dockerfile

The project uses a multi-stage Docker build:

- **Build stage**: Compiles Go binary
- **Runtime stage**: Alpine-based minimal image (~20MB)

## Database

### Migrations

```bash
# Create new migration
make migrate-create NAME=add_users_table

# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status
```

### Current Migrations

| Migration                                   | Description                               | Status      |
| ------------------------------------------- | ----------------------------------------- | ----------- |
| `000001_init_schema_deps.up.sql`            | UUID v7 function, extensions              | [+] Applied |
| `000002_create_auth_schema.up.sql`          | Users, sessions, tokens                   | [+] Applied |
| `000003_create_countries_currencies.up.sql` | Countries & currencies                    | [+] Applied |
| `000004_create_user_contacts.up.sql`        | User contact management                   | [+] Applied |
| `000005_create_user_profiles.up.sql`        | User profiles & social                    | [+] Applied |
| `000006_create_user_posts.up.sql`           | Blog posts system                         | [+] Applied |
| `000007_create_post_comments.up.sql`        | Comments & replies                        | [+] Applied |
| `000008_create_comment_likes_table.up.sql`  | Comment engagement                        | [+] Applied |
| `000009_create_rbac_tables.up.sql`          | **RBAC: permissions, roles, assignments** | [+] Applied |

### Schema Highlights

- **UUID v7 Primary Keys** - Time-ordered for better performance
- **RBAC System** - 4 tables: permissions, roles, role_permissions, user_roles
- **Proper Indexes** - All foreign keys indexed, RBAC lookups optimized
- **Soft Deletes** - Optional soft delete support
- **Timestamps** - created_at, updated_at on all tables
- **Foreign Key Constraints** - Referential integrity enforced
- **System Roles Protection** - is_system flag prevents deletion of core roles

See [AUTH_SCHEMA.md](docs/AUTH_SCHEMA.md) for complete schema documentation and [AUTHORIZATION.md](docs/AUTHORIZATION.md) for RBAC middleware usage.

## Performance

### UUID v7 Benefits

The project uses **UUID v7** (time-ordered) instead of UUID v4 (random):

- [+] **20-50% faster INSERTs** - Better B-tree locality
- [+] **2x faster generation** - 87ns vs 185ns per ID
- [+] **Zero allocations** - Memory efficient
- [+] **Extractable timestamp** - Built-in creation time
- [+] **Index-friendly** - Reduced page splits

See [UUID_V7_MIGRATION.md](docs/UUID_V7_MIGRATION.md) for migration guide and benchmarks.

## Authentication Flow

### Default Users & Credentials

The system comes with **pre-configured users** for development and testing. These are created automatically via database migrations.

| Email                           | Password   | Role           | Permissions                        | Purpose                              |
| ------------------------------- | ---------- | -------------- | ---------------------------------- | ------------------------------------ |
| `system@promenade.com`          | `passw0rd` | **Superadmin** | `*:*` (full access)                | System administration, initial setup |
| `alexander.vasilenko@gmail.com` | `03041965` | **Admin**      | All resources except system config | Project owner, team lead             |
| _(registered users)_            | _(as set)_ | **User**       | Own content management             | Regular users                        |

**Quick Login Examples:**

```bash
# Login as Superadmin (full system access)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "system@promenade.com",
    "password": "passw0rd"
  }'

# Login as Admin (administrative access)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alexander.vasilenko@gmail.com",
    "password": "03041965"
  }'
```

**[!] Security Note:** Change default passwords in production! These credentials are for **development only**.

### Register & Login

```bash
# Register new user
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123!"
  }'

# Login (returns access + refresh tokens)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'

# Response:
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "019b2f90-9c3e-7147-aff6-86deb842e084",
      "email": "user@example.com",
      "name": "John Doe",
      "status": "unverified"
    }
  }
}
```

### Refresh Token

```bash
curl -X POST http://localhost:8081/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### Protected Endpoints

```bash
# Use access token in Authorization header
curl -X GET http://localhost:8081/api/v1/products \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Authorization (RBAC)

The API uses **Role-Based Access Control** with fine-grained permissions. See [AUTHORIZATION.md](docs/AUTHORIZATION.md) for complete guide.

**Quick Examples:**

```go
// Protect endpoint with specific permission
router.POST("/posts",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)

// Require multiple permissions (AND)
router.POST("/posts/:id/publish",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions("posts:create", "posts:publish"),
    handler.PublishPost,
)

// Require any of multiple permissions (OR)
router.GET("/admin",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission("admin:*", "moderator:*"),
    handler.AdminDashboard,
)

// Role-based check
router.GET("/superadmin",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("superadmin"),
    handler.SuperAdminPanel,
)
```

**System Roles:**

| Role         | Permissions        | Description                              | Default Users                   |
| ------------ | ------------------ | ---------------------------------------- | ------------------------------- |
| `superadmin` | `*:*` (all)        | Full system access, cannot be restricted | `system@promenade.com`          |
| `admin`      | Most resources     | System administration, user management   | `alexander.vasilenko@gmail.com` |
| `moderator`  | Content moderation | Can moderate posts, comments, ban users  | _(assign manually)_             |
| `user`       | Own content        | Create/edit own posts, comments, profile | All registered users            |
| `guest`      | Read-only          | View public content only                 | _(unauthenticated)_             |

**Permission Format:** `resource:action` (e.g., `posts:create`, `users:ban`, `*:read`)

**Testing Permissions:**

```bash
# Login as superadmin to test admin endpoints
curl -X POST http://localhost:8081/api/v1/auth/login \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' \
  -H "Content-Type: application/json"

# Use returned token for protected endpoints
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer <access_token>"
```

For detailed usage, wildcard permissions, testing, and best practices, see **[Authorization Guide](docs/AUTHORIZATION.md)**.

## -> API Examples

### User Profiles API

```bash
# Create user profile (requires auth)
curl -X POST http://localhost:8081/api/v1/profiles \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "johndoe",
    "display_name": "John Doe",
    "bio": "Software engineer and tech enthusiast",
    "timezone": "Europe/Kiev",
    "locale": "en",
    "is_public": true
  }'

# Get user profile
curl http://localhost:8081/api/v1/profiles/{id} \
  -H "Authorization: Bearer YOUR_TOKEN"

# Update profile
curl -X PUT http://localhost:8081/api/v1/profiles/{id} \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "Updated bio",
    "website": "https://example.com",
    "location": "Kyiv, Ukraine"
  }'

# Search profiles
curl "http://localhost:8081/api/v1/profiles/search?q=john&limit=10&offset=0"

# Get profile by nickname
curl http://localhost:8081/api/v1/profiles/nickname/johndoe

# Increment profile views (automatically called when viewing)
curl -X POST http://localhost:8081/api/v1/profiles/{id}/views \
  -H "Authorization: Bearer YOUR_TOKEN"

# Admin: Ban profile
curl -X POST http://localhost:8081/api/v1/profiles/{id}/ban \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Violation of terms"}'

# Admin: Unban profile
curl -X POST http://localhost:8081/api/v1/profiles/{id}/unban \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Admin: Verify profile (blue checkmark)
curl -X POST http://localhost:8081/api/v1/profiles/{id}/verify \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### User Contacts API

```bash
# Create contact (requires auth)
curl -X POST http://localhost:8081/api/v1/users/contacts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_type": "email",
    "contact_value": "john@example.com",
    "label": "Work Email",
    "is_primary": true,
    "is_public": false,
    "is_verified": false,
    "available_days": ["monday", "tuesday", "wednesday", "thursday", "friday"],
    "available_from": "09:00",
    "available_to": "18:00",
    "timezone": "Europe/Kiev"
  }'

# List user contacts
curl http://localhost:8081/api/v1/users/contacts \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get contact by ID
curl http://localhost:8081/api/v1/users/contacts/{id} \
  -H "Authorization: Bearer YOUR_TOKEN"

# Update contact
curl -X PUT http://localhost:8081/api/v1/users/contacts/{id} \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"label": "Personal Email", "is_public": true}'

# Set as primary contact
curl -X POST http://localhost:8081/api/v1/users/contacts/{id}/primary \
  -H "Authorization: Bearer YOUR_TOKEN"

# Toggle contact active status
curl -X POST http://localhost:8081/api/v1/users/contacts/{id}/toggle \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get contacts by type
curl http://localhost:8081/api/v1/users/contacts/type/email \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get primary contact by type
curl http://localhost:8081/api/v1/users/contacts/primary/email \
  -H "Authorization: Bearer YOUR_TOKEN"

# Delete contact
curl -X DELETE http://localhost:8081/api/v1/users/contacts/{id} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### User Posts API (Blog/Articles)

```bash
# Create draft post (requires auth)
curl -X POST http://localhost:8081/api/v1/posts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Getting Started with Go",
    "slug": "getting-started-with-go",
    "excerpt": "Learn the basics of Go programming",
    "content": "# Introduction\n\nGo is a statically typed...",
    "tags": ["go", "programming", "tutorial"],
    "categories": ["Development", "Go"],
    "meta_title": "Go Tutorial for Beginners",
    "meta_description": "Complete guide to getting started with Go",
    "is_public": true,
    "is_comments_enabled": true
  }'

# Publish post
curl -X POST http://localhost:8081/api/v1/posts/{id}/publish \
  -H "Authorization: Bearer YOUR_TOKEN"

# Schedule post for future
curl -X POST http://localhost:8081/api/v1/posts/{id}/publish \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"scheduled_at": "2025-12-20T10:00:00Z"}'

# Get published posts (public)
curl "http://localhost:8081/api/v1/posts/published?limit=10&offset=0"

# Search posts
curl "http://localhost:8081/api/v1/posts/search?q=golang&limit=10"

# Get posts by tag
curl "http://localhost:8081/api/v1/posts/tag/golang?limit=10"

# Get featured posts
curl "http://localhost:8081/api/v1/posts/featured?limit=5"

# Get user's posts
curl http://localhost:8081/api/v1/posts/user/{userId} \
  -H "Authorization: Bearer YOUR_TOKEN"

# Like post
curl -X POST http://localhost:8081/api/v1/posts/{id}/like \
  -H "Authorization: Bearer YOUR_TOKEN"

# Unlike post
curl -X DELETE http://localhost:8081/api/v1/posts/{id}/like \
  -H "Authorization: Bearer YOUR_TOKEN"

# View post (increments view count)
curl -X POST http://localhost:8081/api/v1/posts/{id}/view

# Update post
curl -X PUT http://localhost:8081/api/v1/posts/{id} \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Title", "content": "Updated content"}'

# Unpublish post (back to draft)
curl -X POST http://localhost:8081/api/v1/posts/{id}/unpublish \
  -H "Authorization: Bearer YOUR_TOKEN"

# Toggle featured status (admin)
curl -X POST http://localhost:8081/api/v1/posts/{id}/featured \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Toggle comments (admin)
curl -X POST http://localhost:8081/api/v1/posts/{id}/comments \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Delete post (soft delete)
curl -X DELETE http://localhost:8081/api/v1/posts/{id} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Countries & Currencies API

```bash
# List countries
curl http://localhost:8081/api/v1/countries

# Get country by code
curl http://localhost:8081/api/v1/countries/code/UA

# List currencies
curl http://localhost:8081/api/v1/currencies

# Get currency by code
curl http://localhost:8081/api/v1/currencies/code/USD
```

## Best Practices

This project follows industry best practices:

### Code Quality

- [+] **Clean Architecture** - Strict layer separation
- [+] **No ORM** - Direct SQL with sqlx for transparency
- [+] **Dependency Injection** - Manual DI in main.go
- [+] **Interface-based design** - Repository pattern
- [+] **Error handling** - Proper error wrapping
- [+] **Context propagation** - Request context throughout

### Security

- [+] **JWT tokens** - Access + refresh token pattern
- [+] **Password hashing** - bcrypt with cost 10
- [+] **SQL injection prevention** - Parameterized queries
- [+] **RBAC authorization** - 6 middleware functions (RequireAuth, RequirePermission, RequireAnyPermission, RequireAllPermissions, RequireRole, RequireAnyRole)
- [+] **CORS middleware** - Configurable origins
- [+] **Request ID tracking** - X-Request-ID header
- [+] **Graceful shutdown** - Clean connection closure

### Testing

- [+] **Isolated test DB** - Separate port (5433)
- [+] **Fixtures with overrides** - Flexible test data
- [+] **Table-driven tests** - Subtests for clarity
- [+] **Cleanup after tests** - No data leakage
- [+] **Integration tests** - Real database testing

### Development

- [+] **Makefile automation** - Consistent commands
- [+] **Environment hierarchy** - Dev/prod separation
- [+] **Migration versioning** - Reversible schema changes
- [+] **Swagger documentation** - Auto-generated from code
- [+] **Code generation** - Templates for boilerplate

## \* Contributing

We welcome contributions! Please follow these guidelines:

### Development Workflow

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Write tests** for your changes
4. **Ensure** all tests pass: `make test`
5. **Format** code: `make fmt`
6. **Lint**: `make lint`
7. **Commit** with clear messages: `git commit -m 'Add amazing feature'`
8. **Push** to your fork: `git push origin feature/amazing-feature`
9. **Open** a Pull Request

### Code Standards

- **All code and comments in English**
- Follow existing patterns (Clean Architecture layers)
- Add tests for new features
- Update documentation as needed
- Run `make test lint fmt` before committing

### Adding New Features

Use generator scripts for consistency:

```bash
# Interactive generator
./scripts/generate-interactive.sh

# Or manual
./scripts/generate.sh entity MyEntity
./scripts/generate.sh usecase my_usecase
./scripts/generate.sh repository my_repository
```

## API Quick Reference

### Core Endpoints (Most Used)

```bash
# Health Check
GET /api/v1/health

# Authentication
POST /api/v1/auth/register          # Register new user
POST /api/v1/auth/login             # Login (get tokens)
POST /api/v1/auth/refresh           # Refresh access token
POST /api/v1/auth/logout            # Logout (invalidate refresh token)
GET  /api/v1/auth/me                # Get current user (requires auth)
GET  /api/v1/auth/sessions          # List user sessions (requires auth)

# User Profiles
POST /api/v1/profiles               # Create profile (requires auth)
GET  /api/v1/profiles/me            # Get my profile (requires auth)
PUT  /api/v1/profiles/:id           # Update profile (requires auth)
GET  /api/v1/profiles               # List all profiles (public)
GET  /api/v1/profiles/:id           # Get profile by ID (public)

# User Contacts
POST /api/v1/users/contacts         # Create contact (requires auth)
GET  /api/v1/users/contacts         # List my contacts (requires auth)
PUT  /api/v1/users/contacts/:id     # Update contact (requires auth)
DELETE /api/v1/users/contacts/:id   # Delete contact (requires auth)

# Blog Posts
POST /api/v1/posts                  # Create post (requires auth)
POST /api/v1/posts/:id/publish      # Publish post (requires auth)
GET  /api/v1/posts/published        # List published posts (public)
GET  /api/v1/posts/:id              # Get single post (public)
PUT  /api/v1/posts/:id              # Update post (requires auth)
POST /api/v1/posts/:id/like         # Like post (public)
POST /api/v1/posts/:id/view         # Increment view count (public)

# Comments
POST /api/v1/comments               # Create comment (requires auth)
GET  /api/v1/comments?post_id=xxx   # List post comments (public)
GET  /api/v1/comments/:id           # Get comment (public)
PUT  /api/v1/comments/:id           # Update comment (requires auth)
POST /api/v1/comments/:id/like      # Like comment (requires auth)
GET  /api/v1/comments/:id/replies   # Get comment replies (public)

# Countries & Currencies
GET  /api/v1/countries              # List countries
GET  /api/v1/countries/code/:code   # Get country by code
GET  /api/v1/currencies             # List currencies
GET  /api/v1/currencies/code/:code  # Get currency by code

# RBAC (Role-Based Access Control)
POST /api/v1/permissions            # Create permission (requires permissions:create)
GET  /api/v1/permissions            # List permissions (requires permissions:read)
POST /api/v1/roles                  # Create role (requires roles:create)
GET  /api/v1/roles                  # List roles (requires roles:read)
POST /api/v1/roles/:id/permissions  # Add permission to role (requires roles:update)
POST /api/v1/roles/:id/users/:userId # Assign role to user (requires roles:assign)
GET  /api/v1/users/:id/roles        # Get user roles (requires users:read)
```

### Authentication Example

```bash
# 1. Register
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"John Doe","password":"SecurePass123!"}'

# 2. Login (get tokens)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123!"}'
# Response: {"access_token": "...", "refresh_token": "..."}

# 3. Use access token for protected endpoints
curl http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"

# 4. Refresh when access token expires
curl -X POST http://localhost:8081/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'

# 5. Logout (invalidate refresh token)
curl -X POST http://localhost:8081/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

### RBAC Example

```bash
# 1. List all permissions (requires permissions:read)
curl http://localhost:8081/api/v1/permissions \
  -H "Authorization: Bearer <admin_token>"

# 2. Create custom permission (requires permissions:create)
curl -X POST http://localhost:8081/api/v1/permissions \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"resource":"articles","action":"publish","description":"Publish articles"}'

# 3. Create new role (requires roles:create)
curl -X POST http://localhost:8081/api/v1/roles \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"content_manager","display_name":"Content Manager","description":"Can manage all content"}'

# 4. Add permissions to role (requires roles:update)
curl -X POST http://localhost:8081/api/v1/roles/<role_id>/permissions \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"permission_id":"<permission_id>"}'

# 5. Assign role to user (requires roles:assign)
curl -X POST http://localhost:8081/api/v1/roles/<role_id>/users/<user_id> \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"expires_at":"2025-12-31T23:59:59Z"}' # Optional expiration

# 6. Check user's roles
curl http://localhost:8081/api/v1/users/<user_id>/roles \
  -H "Authorization: Bearer <token>"

# 7. Use wildcard permissions for superadmin
# System role "superadmin" has permission "*:*" which grants all access
```

**Permission Format**: `resource:action`

- Examples: `posts:create`, `users:delete`, `comments:*`, `*:*`
- Wildcard `*` matches anything: `posts:*` = all post actions, `*:*` = full access

**System Roles** (cannot be deleted):

- `superadmin` - Full access (_:_)
- `admin` - User/content management
- `moderator` - Content moderation
- `user` - Basic user operations
- `guest` - Read-only access

### Swagger Documentation

Interactive API documentation available at:

- **v1**: http://localhost:8081/api/v1/docs/swagger/index.html
- **v2**: http://localhost:8081/api/v2/docs/swagger/index.html

## \* License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

## \* Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [sqlx](https://github.com/jmoiron/sqlx)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [testify](https://github.com/stretchr/testify)
- Clean Architecture principles by Robert C. Martin

---

**Built with ** using Clean Architecture and best practices\*\*
