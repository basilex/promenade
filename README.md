# Promenade

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Production-ready REST API built with **Clean Architecture**, featuring PostgreSQL with UUID v7, comprehensive testing infrastructure, JWT authentication, and API versioning.

## ✨ Key Features

- 🏗️ **Clean Architecture** - Clear separation of concerns (Domain, Use Case, Adapter, Infrastructure)
- 🔑 **UUID v7 Primary Keys** - Time-ordered UUIDs for optimal performance (2x faster than v4)
- 🧪 **Comprehensive Testing** - Integration tests with isolated test database (13 tests, 100% passing)
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

- **API Base URL**: http://localhost:8081
- **Swagger v1**: http://localhost:8081/swagger/v1/index.html
- **Swagger v2**: http://localhost:8081/swagger/v2/index.html
- **Health Check**: http://localhost:8081/health

## 📚 Documentation

### Technical Documentation

- [Testing Guide](docs/TESTING_GUIDE.md) - Comprehensive testing setup and best practices
- [Testing Infrastructure](docs/TESTING_INFRASTRUCTURE.md) - Test infrastructure overview
- [UUID v7 Migration](docs/UUID_V7_MIGRATION.md) - Migrating from UUID v4 to v7
- [ID Strategies](docs/ID_STRATEGIES.md) - Primary key strategy recommendations
- [Auth Schema](docs/AUTH_SCHEMA.md) - Database schema for authentication system
- [Test Results](docs/TEST_RESULTS.md) - Current test coverage and results

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
│   │   ├── entity/                    # Business entities (User, Product, Role, Session)
│   │   └── repository/                # Repository interfaces (ports)
│   ├── usecase/                       # Business logic orchestration
│   │   ├── auth_usecase.go           # Login, register, refresh, logout
│   │   ├── product_usecase.go        # CRUD operations
│   │   └── role_usecase.go           # Role management
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
│   └── mocks/                       # Mock repositories (planned)
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

- **13 Integration Tests** - 100% passing
- **Test Coverage** - Repository layer fully tested
- **Test Database** - PostgreSQL 16 on port 5433 (isolated from dev DB)
- **Test Execution** - ~3 seconds for full suite

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

## 🐳 Docker Deployment

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

### Products API

```bash
# List products (with pagination)
curl "http://localhost:8081/api/v1/products?page=1&limit=10"

# Create product (requires auth)
curl -X POST http://localhost:8081/api/v1/products \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium Widget",
    "description": "High-quality widget",
    "price": 29.99,
    "stock": 100
  }'

# Get product by ID
curl http://localhost:8081/api/v1/products/{id}

# Update product
curl -X PUT http://localhost:8081/api/v1/products/{id} \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"price": 24.99, "stock": 150}'

# Delete product
curl -X DELETE http://localhost:8081/api/v1/products/{id} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Roles API

```bash
# List all roles
curl http://localhost:8081/api/v1/roles \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get role by ID
curl http://localhost:8081/api/v1/roles/{id} \
  -H "Authorization: Bearer YOUR_TOKEN"
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
