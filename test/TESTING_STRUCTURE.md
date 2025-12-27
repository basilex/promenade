# Testing Structure Guide

## 🎯 Testing Philosophy

Promenade follows industry best practices with **mirror path structure** for integration tests, inspired by Java testing conventions.

## 📂 Directory Structure

```
promenade/
├── internal/                       # Production code
│   ├── contexts/
│   │   ├── identity/
│   │   │   ├── contact/           # Contact aggregate
│   │   │   │   ├── entity/
│   │   │   │   ├── repository/
│   │   │   │   ├── usecase/
│   │   │   │   └── *_test.go     # ✅ Unit tests (in-place)
│   │   │   └── user/              # User aggregate
│   │   └── shared/
│   └── infrastructure/
│
├── pkg/                            # Shared packages
│   ├── bus/
│   │   ├── bus.go
│   │   ├── bus_test.go            # ✅ Unit tests (in-place)
│   │   ├── memory/
│   │   │   ├── memory_bus.go
│   │   │   └── memory_bus_test.go # ✅ Unit tests (in-place)
│   │   └── redis/
│   │       ├── redis_bus.go
│   │       └── redis_bus_test.go  # ✅ Unit tests (in-place)
│   └── logger/
│
└── test/                           # Test infrastructure
    ├── integration/                # Integration tests (MIRROR PATH)
    │   ├── testutils.go           # Shared test utilities
    │   ├── contexts/               # ⭐ Mirror: internal/contexts/
    │   │   └── identity/
    │   │       └── contact/        # ⭐ Mirror: internal/contexts/identity/contact/
    │   │           └── contact_api_test.go  # HTTP API integration tests
    │   └── pkg/                    # ⭐ Mirror: pkg/
    │       └── bus/                # ⭐ Mirror: pkg/bus/
    │           └── bus_integration_test.go  # Event bus integration tests
    │
    ├── smoke/                      # Smoke tests (quick production checks)
    │   └── health_check_test.go
    │
    └── stress/                     # Load/stress tests
        └── bus_load_test.go
```

---

## 🧪 Test Types

### 1. Unit Tests (In-Place)

**Location**: Same directory as production code  
**File pattern**: `*_test.go`  
**Package**: Same as production code (e.g., `package contact`)  
**Purpose**: Test single units in isolation

**Example:**

```
internal/contexts/identity/contact/usecase/
├── contact_usecase.go
└── contact_usecase_test.go     # ✅ Unit test
```

**Characteristics:**

- ✅ Fast (< 1 second)
- ✅ No external dependencies
- ✅ Pure logic testing
- ✅ Mock repositories/dependencies

**Run:**

```bash
# All unit tests
go test ./... -short

# Specific package
go test ./internal/contexts/identity/contact/usecase -v

# With coverage
go test ./... -cover -coverprofile=coverage.out
```

---

### 2. Integration Tests (Mirror Path)

**Location**: `test/integration/` with **mirror path** structure  
**File pattern**: `*_test.go`  
**Package**: `{feature}_test` (e.g., `package contact_test`, `package bus_test`)  
**Purpose**: Test component interactions (HTTP API, Database, Event Bus, External Services)

**Mirror Path Examples:**

| Production Code                       | Integration Test                              |
| ------------------------------------- | --------------------------------------------- |
| `internal/contexts/identity/contact/` | `test/integration/contexts/identity/contact/` |
| `pkg/bus/`                            | `test/integration/pkg/bus/`                   |
| `internal/contexts/customer/`         | `test/integration/contexts/customer/`         |

**Characteristics:**

- ⏱ Slower (1-30 seconds)
- 🐘 Requires PostgreSQL
- 🔴 Requires Redis (optional, can skip)
- 🌐 Tests real HTTP endpoints
- 💾 Tests real database queries

**Run:**

```bash
# All integration tests (requires test DB)
go test ./test/integration/... -v

# Specific integration test
go test ./test/integration/contexts/identity/contact -v

# Skip Redis tests
go test ./test/integration/... -short -v
```

---

### 3. Smoke Tests (Quick Production Checks)

**Location**: `test/smoke/`  
**Purpose**: Fast sanity checks for production deployments

**Example:**

```go
// test/smoke/health_check_test.go
func TestHealthEndpoint(t *testing.T) {
    resp, err := http.Get("http://localhost:8081/health")
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

**Run:**

```bash
go test ./test/smoke -v
```

---

### 4. Stress Tests (Load Testing)

**Location**: `test/stress/`  
**Purpose**: Performance, load, and stress testing

**Example:**

```go
// test/stress/bus_load_test.go
func TestBus_HighLoad_10kEvents(t *testing.T) {
    // Publish 10,000 events and measure throughput
}
```

**Run:**

```bash
go test ./test/stress -v -timeout 30m
```

---

## 🔧 Test Utilities

### Shared Test Utilities

**File**: `test/integration/testutils.go`  
**Purpose**: Shared setup functions for integration tests

**Functions:**

```go
// SetupTestDB initializes test database with migrations
func SetupTestDB(t *testing.T) *TestDB

// CleanupTestDB drops all tables (destructive!)
func CleanupTestDB(t *testing.T, db *sqlx.DB)
```

**Usage in integration tests:**

```go
package contact_test

import (
    "testing"
    "github.com/basilex/promenade/test/integration"
)

func TestContactAPI_Create(t *testing.T) {
    testDB := integration.SetupTestDB(t)
    defer testDB.Cleanup()

    // Test code here...
}
```

---

## 📝 Naming Conventions

### Test Functions

| Type        | Pattern                                | Example                       |
| ----------- | -------------------------------------- | ----------------------------- |
| Unit        | `Test{Component}_{Scenario}`           | `TestContactUseCase_Create`   |
| Integration | `Test{Feature}_{Component}_{Scenario}` | `TestContactAPI_FullWorkflow` |
| Benchmark   | `Benchmark{Component}_{Operation}`     | `BenchmarkMemoryBus_Publish`  |

### Test Files

| Type        | Pattern                    | Example                   |
| ----------- | -------------------------- | ------------------------- |
| Unit        | `{file}_test.go`           | `contact_usecase_test.go` |
| Integration | `{feature}_{type}_test.go` | `contact_api_test.go`     |

### Test Packages

| Type        | Package            | Import Alias                  |
| ----------- | ------------------ | ----------------------------- |
| Unit        | Same as production | -                             |
| Integration | `{feature}_test`   | `integration` (for testutils) |

---

## 🚀 Running Tests

### Development Workflow

```bash
# 1. Quick unit tests (< 1 sec)
make test-unit

# 2. Integration tests (requires test DB)
make test-integration

# 3. All tests
make test

# 4. With coverage
make test-coverage
```

### CI/CD Pipeline

```bash
# GitHub Actions / GitLab CI
make test-ci
```

**`.github/workflows/tests.yml`:**

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5433:5432
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.23"

      - name: Run unit tests
        run: go test ./... -short -v

      - name: Run integration tests
        run: go test ./test/integration/... -v
```

---

## 📊 Test Coverage

### Generate Coverage Report

```bash
# HTML report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# Terminal summary
go test ./... -cover
```

### Coverage Targets

| Package     | Target | Current |
| ----------- | ------ | ------- |
| Core domain | 90%    | -       |
| Use cases   | 80%    | -       |
| Adapters    | 70%    | -       |
| Integration | 60%    | -       |

---

## 🎯 Best Practices

### ✅ DO

- ✅ **Unit tests in-place** next to production code
- ✅ **Integration tests mirror path** in `test/integration/`
- ✅ **Skip expensive tests** with `testing.Short()`
- ✅ **Use testify** for assertions (`assert`, `require`)
- ✅ **Clean up resources** with `defer testDB.Cleanup()`
- ✅ **Test error paths** not just happy path
- ✅ **Use table-driven tests** for multiple scenarios

### ❌ DON'T

- ❌ **DON'T mix unit and integration tests** in same file
- ❌ **DON'T share state** between tests (use `t.Parallel()` safely)
- ❌ **DON'T use production database** for tests
- ❌ **DON'T ignore test cleanup** (resource leaks)
- ❌ **DON'T skip flaky tests** (fix them!)

---

## 📚 Examples

### Unit Test Example

```go
// internal/contexts/identity/contact/usecase/contact_usecase_test.go
package usecase_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestContactUseCase_Create(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockContactRepository(t)
    uc := NewContactUseCase(mockRepo)

    // Act
    contact, err := uc.CreateContact(ctx, userID, "email", "test@example.com")

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "email", contact.Type)
}
```

### Integration Test Example

```go
// test/integration/contexts/identity/contact/contact_api_test.go
package contact_test

import (
    "testing"
    "github.com/basilex/promenade/test/integration"
)

func TestContactAPI_Create(t *testing.T) {
    testDB := integration.SetupTestDB(t)
    defer testDB.Cleanup()

    // Create test user
    userID := createTestUser(t, testDB.DB)

    // Setup router
    router := setupTestRouter(testDB.DB)

    // Make HTTP request
    resp := makeRequest(t, router, "POST", "/api/v1/contacts", payload)

    // Assert response
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
```

---

## 🔍 Troubleshooting

### Common Issues

**1. Test DB connection failed**

```bash
# Start test database
make test-db-start

# Check status
docker ps | grep promenade_test_db
```

**2. Redis tests failing**

```bash
# Skip Redis tests
go test ./... -short -v

# Or start Redis
docker run -d -p 6379:6379 redis:7-alpine
```

**3. Port already in use**

```bash
# Kill process on port 5433
lsof -ti:5433 | xargs kill -9
```

**4. Test timeout**

```bash
# Increase timeout
go test ./test/integration/... -timeout 30m
```

---

## 📖 Additional Resources

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Testify Package](https://github.com/stretchr/testify)
- [Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

## 🎓 Summary

| Test Type       | Location       | Speed     | Dependencies | When         |
| --------------- | -------------- | --------- | ------------ | ------------ |
| **Unit**        | In-place       | ⚡ Fast   | None         | Always       |
| **Integration** | Mirror path    | 🐢 Slower | DB, Redis    | Before merge |
| **Smoke**       | `test/smoke/`  | ⚡ Fast   | Production   | After deploy |
| **Stress**      | `test/stress/` | 🐌 Slow   | Load testing | Pre-release  |

**Testing Pyramid:**

```
        /\
       /  \  Unit Tests (80%)
      /____\
     /      \
    /        \  Integration Tests (15%)
   /__________\
  /            \
 /              \ Smoke/Stress (5%)
/________________\
```

---

**Last Updated**: 2025-12-27  
**Author**: Promenade Team  
**Go Version**: 1.23+
