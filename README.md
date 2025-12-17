# Promenade

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Production-ready REST API built with **Clean Architecture**, featuring PostgreSQL with UUID v7, comprehensive testing infrastructure, JWT authentication, and API versioning.

## ✨ Key Features

- 🏗️ **Clean Architecture** - Clear separation of concerns (Domain, Use Case, Adapter, Infrastructure)
- 🔑 **UUID v7 Primary Keys** - Time-ordered UUIDs for optimal performance (2x faster than v4)
- 📊 **Structured Logging** - slog with JSON/text format, context fields (request_id, user_id)
- 🧪 **Comprehensive Testing** - 149 tests total across all layers - 100% passing
  - Unit: 119 tests (59 entity + 60 use case)
  - Integration: 30 tests (repository layer with real PostgreSQL)
- 🔐 **JWT Authentication** - Secure token-based auth with refresh tokens
- 📚 **API Versioning** - v1 and v2 with backward compatibility
- 🗄️ **PostgreSQL + sqlx** - No ORM, pure SQL with transaction support
- 📖 **Swagger Documentation** - Auto-generated API docs for both versions
- 🐳 **Docker Ready** - Full Docker Compose setup for development and testing
- 🔄 **Database Migrations** - golang-migrate for version control
- ⚡ **High Performance** - Gin framework with graceful shutdown

## 🚀 Quick Start

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

### Quick Test

```bash
# Run all tests (unit + integration)
make test

# Or run integration tests only
make test-integration
```

## 📡 API Access

- **API Base URL**: http://localhost:8081/api/v1
- **Swagger v1**: http://localhost:8081/api/v1/docs/swagger/index.html
- **Swagger v2**: http://localhost:8081/api/v2/docs/swagger/index.html
- **Health Check**: http://localhost:8081/health

## 📚 Documentation

### Technical Documentation

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

## 🛠️ Development Commands

### Core Commands

```bash
make help              # Show all available commands
make install           # Install development tools (migrate, swag, golangci-lint)
make dev               # Start development server (starts DB, runs migrations)
make build             # Build production binary to bin/promenade
make clean             # Clean build artifacts
```

### Database Commands

```bash
make docker-up         # Start PostgreSQL (dev) via Docker Compose
make docker-down       # Stop and remove containers
make docker-logs       # Show PostgreSQL logs
make migrate-up        # Run all pending migrations
make migrate-down      # Rollback last migration
make migrate-create NAME=create_example_table  # Create new migration
```

### Testing Commands

```bash
make test              # Run all tests (unit + integration)
make test-unit         # Run unit tests only
make test-integration  # Run integration tests (auto-starts test DB)
make test-coverage     # Generate HTML coverage report
make test-watch        # Watch mode with gotestsum
make test-db-start     # Start test database (port 5433)
make test-db-stop      # Stop test database
```

### Code Quality

```bash
make fmt               # Format code with gofmt
make lint              # Run golangci-lint
make swagger-all       # Generate Swagger docs (v1 + v2)
```

## 🏗️ Architecture

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
│   │   ├── entity/                    # Business entities (User, UserProfile, UserContact, Country, Currency, Session)
│   │   └── repository/                # Repository interfaces (ports)
│   ├── usecase/                       # Business logic orchestration
│   │   ├── auth_usecase.go           # Login, register, refresh, logout
│   │   ├── user_contact_usecase.go   # User contacts management
│   │   ├── user_profile_usecase.go   # User profiles, privacy, moderation
│   │   ├── country_usecase.go        # Countries CRUD
│   │   └── currency_usecase.go       # Currencies CRUD
│   ├── adapter/
│   │   ├── http/
│   │   │   ├── shared/middleware/    # Auth, CORS, logging, recovery
│   │   │   ├── v1/                   # API v1 (handlers, DTOs, routes)
│   │   │   └── v2/                   # API v2 (handlers, DTOs, routes)
│   │   └── repository/postgres/      # Repository implementations (sqlx)
│   └── infrastructure/
│       ├── config/                    # Configuration loader
│       ├── database/                  # PostgreSQL connection & transactions
│       └── logger/                    # Structured logging
├── pkg/
│   ├── jwt/                          # JWT token manager
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

## ⚙️ Configuration

### Environment Files

The project uses a hierarchical environment configuration:

| File               | Purpose              | Committed? | Priority   |
| ------------------ | -------------------- | ---------- | ---------- |
| `.env`             | Base defaults        | ✅ Yes     | Lowest     |
| `.env.development` | Development settings | ✅ Yes     | Medium     |
| `.env.local`       | Personal overrides   | ❌ No      | Highest    |
| `.env.production`  | Production secrets   | ❌ No      | Production |

### Key Configuration Variables

```bash
# Server
SERVER_PORT=8081
SERVER_HOST=0.0.0.0
GIN_MODE=debug                    # debug, release

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=promenade
DB_PASSWORD=promenade
DB_NAME=promenade
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key-here
JWT_ACCESS_EXPIRY=15m            # 15 minutes
JWT_REFRESH_EXPIRY=168h          # 7 days

# Test Database (separate from dev)
TEST_DB_PORT=5433
TEST_DB_NAME=promenade_test
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

## 🧪 Testing

Promenade features a **comprehensive testing infrastructure** with isolated test database and helper utilities.

### Test Statistics

- **149 Total Tests** - 100% passing
  - **119 Unit Tests**
    - Entity: 59 tests (User, UserProfile, Country validation)
    - Use Case: 60 tests (Auth, UserProfile, UserContact business logic)
  - **30 Integration Tests** - Repository layer with real PostgreSQL
    - User, UserProfile, UserContact, Session, Country, Currency repositories
- **Test Database** - PostgreSQL 16 on port 5433 (isolated from dev DB)
- **Test Execution** - ~8 seconds for full suite (unit + integration)
- **Coverage** - All layers tested (entity validation, business logic, database operations)

#### Module Test Breakdown

| Module          | Entity Tests | UseCase Tests | Integration Tests | Total   |
| --------------- | ------------ | ------------- | ----------------- | ------- |
| **User**        | 0            | 0             | 7                 | 7       |
| **UserProfile** | 29           | 32            | 11                | 72      |
| **UserContact** | 0            | 0             | 3                 | 3       |
| **Auth**        | 0            | 0             | 0                 | 0       |
| **Country**     | 30           | 0             | 2                 | 32      |
| **Currency**    | 0            | 0             | 2                 | 2       |
| **Session**     | 0            | 0             | 5                 | 5       |
| **Other**       | 0            | 28            | 0                 | 28      |
| **Total**       | **59**       | **60**        | **30**            | **149** |

### Running Tests

```bash
# Quick test - all tests with auto DB setup
make test

# Integration tests only (with isolated test DB)
make test-integration

# Unit tests (no database required)
make test-unit

# Coverage report (opens in browser)
make test-coverage

# Watch mode (re-run on file changes)
make test-watch
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

See [TESTING_GUIDE.md](docs/TESTING_GUIDE.md) for comprehensive testing documentation.

## � Structured Logging

Promenade uses **slog** (Go 1.21+ structured logging) for production-ready observability.

### Key Features

- ✅ **Structured Format** - JSON (production) or text (development)
- ✅ **Context Fields** - Automatic request_id, user_id, trace_id tracking
- ✅ **Zero Dependencies** - Built-in Go stdlib
- ✅ **Performance** - Zero allocation for common operations
- ✅ **Integration Ready** - Works with ELK, Loki, DataDog, New Relic

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

## �🐳 Docker Deployment

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

## 🗄️ Database

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

### Schema Highlights

- **UUID v7 Primary Keys** - Time-ordered for better performance
- **Proper Indexes** - All foreign keys indexed
- **Soft Deletes** - Optional soft delete support
- **Timestamps** - created_at, updated_at on all tables
- **Foreign Key Constraints** - Referential integrity enforced

See [AUTH_SCHEMA.md](docs/AUTH_SCHEMA.md) for complete schema documentation.

## 🚀 Performance

### UUID v7 Benefits

The project uses **UUID v7** (time-ordered) instead of UUID v4 (random):

- ✅ **20-50% faster INSERTs** - Better B-tree locality
- ✅ **2x faster generation** - 87ns vs 185ns per ID
- ✅ **Zero allocations** - Memory efficient
- ✅ **Extractable timestamp** - Built-in creation time
- ✅ **Index-friendly** - Reduced page splits

See [UUID_V7_MIGRATION.md](docs/UUID_V7_MIGRATION.md) for migration guide and benchmarks.

## 🔐 Authentication Flow

### Register & Login

```bash
# Register new user
curl -X POST http://localhost:8081/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "username": "johndoe"
  }'

# Login
curl -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'

# Response:
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900
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

## 📚 API Examples

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

## 🛡️ Best Practices

This project follows industry best practices:

### Code Quality

- ✅ **Clean Architecture** - Strict layer separation
- ✅ **No ORM** - Direct SQL with sqlx for transparency
- ✅ **Dependency Injection** - Manual DI in main.go
- ✅ **Interface-based design** - Repository pattern
- ✅ **Error handling** - Proper error wrapping
- ✅ **Context propagation** - Request context throughout

### Security

- ✅ **JWT tokens** - Access + refresh token pattern
- ✅ **Password hashing** - bcrypt with cost 10
- ✅ **SQL injection prevention** - Parameterized queries
- ✅ **CORS middleware** - Configurable origins
- ✅ **Request ID tracking** - X-Request-ID header
- ✅ **Graceful shutdown** - Clean connection closure

### Testing

- ✅ **Isolated test DB** - Separate port (5433)
- ✅ **Fixtures with overrides** - Flexible test data
- ✅ **Table-driven tests** - Subtests for clarity
- ✅ **Cleanup after tests** - No data leakage
- ✅ **Integration tests** - Real database testing

### Development

- ✅ **Makefile automation** - Consistent commands
- ✅ **Environment hierarchy** - Dev/prod separation
- ✅ **Migration versioning** - Reversible schema changes
- ✅ **Swagger documentation** - Auto-generated from code
- ✅ **Code generation** - Templates for boilerplate

## 🤝 Contributing

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

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [sqlx](https://github.com/jmoiron/sqlx)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [testify](https://github.com/stretchr/testify)
- Clean Architecture principles by Robert C. Martin

---

**Built with ❤️ using Clean Architecture and best practices**
