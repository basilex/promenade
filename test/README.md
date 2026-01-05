# Testing Guide

**Promenade Testing Infrastructure** - Professional test organization with mirror path structure.

---

## Quick Navigation

- **[Testing Patterns Guide](../docs/TESTING_PATTERNS.md)** - **COMPREHENSIVE** testing patterns (RECOMMENDED) (~800 lines)
- **[Testing Structure Guide](TESTING_STRUCTURE.md)** - Directory organization
- **[Migration Summary](MIGRATION_SUMMARY.md)** - Changes from old structure
- **[Bus Test Coverage](../docs/BUS_TEST_COVERAGE.md)** - Event bus test report

---

## Test Organization

Promenade uses **four-tier testing strategy** with clear separation:

### 1. Unit Tests (In-Place)

**Location**: Same directory as production code  
**Purpose**: Test individual components in isolation

```
internal/contexts/identity/contact/
 entity.go
 entity_test.go              #  Entity tests
 usecase.go
 usecase_test.go             #  UseCase tests
 adapter/
     http/handler/
        contact_handler.go
        contact_handler_test.go  #  Handler unit tests
     repository/postgres/
         contact_repository.go
         contact_repository_test.go  #  Repository tests
```

### 2. Smoke Tests (Mirror Path, No DB)

**Location**: `test/smoke/contexts/` (mirror path)  
**Purpose**: HTTP handler validation (80/20 rule)  
**Status**:  **COMPLETE** - 15/15 handlers, 123 tests, 100% pass rate

```
test/smoke/
 README.md                   # Smoke testing guide
 testutils.go                # Helper utilities (5 functions)
 contexts/                   # Mirror path structure
    identity/               # 5 handlers, 43 tests
       user/handler_test.go
       contact/handler_test.go
       profile/handler_test.go
       role/handler_test.go
       permission/handler_test.go
    customer-mgmt/          # 4 handlers, 34 tests
       company/handler_test.go
       customer/handler_test.go  # Most complex: 12 tests, 23-method mock
       deal/handler_test.go
       interaction/handler_test.go
    order-mgmt/             # 1 handler, 10 tests
       order/handler_test.go
    shared/                 # 4 handlers, 36 tests
       country/handler_test.go
       currency/handler_test.go
       language/handler_test.go
       timezone/handler_test.go
```

**Key Characteristics**:
- Mock UseCase with function fields + nil-check methods
- Test HTTP status codes (200/201/404/400/500)
- Validate response format (`{"status":"success"}`)
- No database dependencies (pure mocks)
- Fast execution (~0.5s for all 123 tests, cached)
- 6-12 tests per handler (simple to complex)
- Error code patterns vary by context:
  - Identity/Order Management: Simple "NOT_FOUND"
  - Customer Management: Standard helpers (NotFound, BadRequest, InternalError)
  - Shared: Entity-specific codes (COUNTRY_NOT_FOUND, VALIDATION_ERROR, 204 Delete)

### 3. Integration Tests (Mirror Path, With DB)

**Location**: `test/integration/contexts/` (mirror path)  
**Purpose**: Full E2E testing with real database

```
test/integration/
 testutils.go                # Shared test utilities
 contexts/                   # Mirror path structure
    shared/
       country/repository_test.go       #  Repository tests with real DB
       currency/repository_test.go      #  Repository tests with real DB
       language/repository_test.go      #  Repository tests with real DB
       timezone/repository_test.go      #  Repository tests with real DB
    identity/
        contact/repository_test.go       #  Repository tests with real DB
 pkg/                        # Package integration tests
     bus/
         bus_integration_test.go
```

### 4. Benchmark Tests (Mirror Path, With DB)

**Location**: `test/benchmark/contexts/` (mirror path)  
**Purpose**: Performance measurement and optimization validation

```
test/benchmark/
 contexts/                   # Mirror path structure
    identity/
       user/
          repository_bench_test.go  #  Repository performance benchmarks
```

**Key characteristics**:
- Measure query performance and memory usage
- Validate optimizations (e.g., N+1 query fixes)
- Require real database (like integration tests)
- Use test DB on port 5433
- Run via Makefile targets with automatic DB setup

---

## Running Tests

### Quick Commands

```bash
# Unit tests only (in-place, fast)
go test ./... -short -v

# Smoke tests (handler validation, no DB)
make test-smoke                     # Run all 123 smoke tests

# Integration tests (with real DB)
make test-integration

# Benchmark tests (performance measurement, with real DB)
make test-benchmark

# All tests
make test
```

** Detailed Smoke Test Documentation**: See [test/smoke/README.md](smoke/README.md) for comprehensive guide with patterns, examples, and all discovered highlights.

### Test Comparison

| Type            | Location                      | Database   | Speed         | Run When              |
| --------------- | ----------------------------- | ---------- | ------------- | --------------------- |
| **Unit**        | In-place (`*_test.go`)        | No (mocks) | Fast (~5s)    | Every save            |
| **Smoke**       | `/test/smoke/contexts/`       | No (mocks) | Fast (~2s)    | Before commit         |
| **Integration** | `/test/integration/contexts/` | Real DB    | Medium (~14s) | Before merge          |
| **Benchmark**   | `/test/benchmark/contexts/`   | Real DB    | Variable      | After optimizations   |

# Context-specific tests

go test ./internal/contexts/identity/... -v
go test ./internal/contexts/shared/... -v

# Benchmark tests (manual, for specific optimization validation)
go test -bench=. -benchmem ./test/benchmark/contexts/identity/user

# Package tests

go test ./pkg/bus/... -v
go test ./pkg/uuidv7/... -v

# With coverage

go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

```

### Makefile Targets

```bash
make test                 # All tests
make test-unit            # Unit tests only (fast, < 5s)
make test-integration     # Integration tests (with real DB)
make test-benchmark       # Benchmark tests (auto DB setup, 5s per benchmark)
make test-benchmark-all   # Extended benchmarks (10s per benchmark)
make test-coverage        # HTML coverage report
```

---

## Test Coverage Overview

### Identity Context

```
internal/contexts/identity/
 contact/
    entity_test.go         # 35 tests, 96% coverage
    usecase_test.go        # 52 tests, 70% coverage
    adapter/
        http/handler/
            contact_handler_test.go  # Handler unit tests
 profile/
     entity_test.go         # 35 tests, 96% coverage
     usecase_test.go        # 52 tests, 70% coverage
     adapter/
         http/handler/
             profile_handler_test.go  # Handler unit tests

test/integration/contexts/identity/
 contact/repository_test.go  # 9 integration tests
 profile/repository_test.go  # 17 integration subtests
```

### Smoke Tests (Handler Validation)

```
test/smoke/contexts/
 identity/
    user/handler_test.go      # 8 tests
    contact/handler_test.go   # 8 tests
    profile/handler_test.go   # 9 tests
    role/handler_test.go      # 9 tests
    permission/handler_test.go # 9 tests
 customer-mgmt/
    company/handler_test.go   # 8 tests
    customer/handler_test.go  # 12 tests (23-method mock - most complex)
    deal/handler_test.go      # 8 tests
    interaction/handler_test.go # 6 tests
 order-mgmt/
    order/handler_test.go     # 10 tests
 shared/
    country/handler_test.go   # 9 tests
    currency/handler_test.go  # 9 tests
    language/handler_test.go  # 9 tests
    timezone/handler_test.go  # 9 tests

**Total: 123 smoke tests across 15 handlers (100% pass rate)**
```

### Shared Context (Reference Data)

```
internal/contexts/shared/
 country/
    entity_test.go         # Entity validation tests
    usecase_test.go        # Business logic tests
    adapter/
        http/handler/
            country_handler_test.go
 currency/
 language/
 timezone/

test/integration/contexts/shared/
 country/repository_test.go  # 6 integration tests
 currency/repository_test.go # 6 integration tests
 language/repository_test.go # 6 integration tests
 timezone/repository_test.go # 6 integration tests
```

### Package Tests

```
pkg/
 bus/                       # 67 tests, 100% coverage
    event_test.go         # 8 tests
    factory_test.go       # 13 tests
    topics_test.go        # 2 tests
    memory/
       memory_bus_test.go      # 8 tests
       edge_cases_test.go      # 13 tests
       benchmark_test.go       # Performance tests
    redis/
        redis_bus_test.go       # 16 tests
 logger/logger_test.go      # 15 tests, 95% coverage
 uuidv7/uuidv7_test.go      # 10 tests, 100% coverage
 response/response_test.go  # 12 tests, 100% coverage
 migration/manager_test.go  # 8 tests, 90% coverage
 valueobject/              # 25 tests, 95% coverage
    email_test.go
    phone_test.go
    address_test.go
    money_test.go
 aggregate/aggregate_test.go # 5 tests
 jsonb/jsonb_test.go        # 8 tests
```

---

## Running Tests

### All Tests

```bash
make test                 # All tests (420+ tests, ~60s with race detector)
```

### By Type

```bash
make test-unit              # Unit tests only (~5s)
make test-smoke             # Smoke tests for handlers (~2s, no DB)
make test-integration       # Integration tests (real DB, ~14s)
make test-coverage          # HTML coverage report
```

### By Context

```bash
# Identity Context
go test ./internal/contexts/identity/... -v
go test ./test/integration/contexts/identity/... -v

# Shared Context
go test ./internal/contexts/shared/... -v
go test ./test/integration/contexts/shared/... -v

# Package Tests
go test ./pkg/bus/... -v
go test ./pkg/logger/... -v
```

---

## Test Conventions

### Naming

- **Test files**: `{entity}_test.go` (e.g., `user_test.go`)
- **Test functions**: `Test{Entity}_{Method}` (e.g., `TestUser_HashPassword`)
- **Subtests**: Use `t.Run(name, func(t))` for table-driven tests

### Structure

```go
func TestEntity_Method(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        {"valid case", input1, want1, false},
        {"invalid case", input2, want2, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Method(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Dependencies

- **testify/assert**: `github.com/stretchr/testify/assert`
- **UUID v7**: `github.com/basilex/promenade/pkg/uuidv7`

---

## Test Conventions

### Naming

- **Test files**: `{entity}_test.go` (e.g., `contact_test.go`, `profile_test.go`)
- **Test functions**: `Test{Entity}_{Method}` (e.g., `TestProfile_UpdateDisplayName`)
- **Subtests**: Use `t.Run(name, func(t))` for table-driven tests

### Structure

```go
func TestEntity_Method(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        {"valid case", input1, want1, false},
        {"invalid case", input2, want2, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Method(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Dependencies

- **testify/assert**: `github.com/stretchr/testify/assert`
- **testify/require**: `github.com/stretchr/testify/require`
- **UUID v7**: `github.com/basilex/promenade/pkg/uuidv7`

---

## Writing New Tests

### 1. Unit Tests (In-Place)

Place test file next to the entity:

```bash
# Example: internal/contexts/identity/contact/entity.go
touch internal/contexts/identity/contact/entity_test.go
```

Write test:

```go
package contact

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/basilex/promenade/pkg/uuidv7"
)

func TestContact_NewEmailContact(t *testing.T) {
    userID := uuidv7.New()
    contact, err := NewEmailContact(userID, "test@example.com", "Work")

    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, contact.ID)
    assert.Equal(t, userID, contact.UserID)
    assert.Equal(t, ContactTypeEmail, contact.Type)
}
```

### 2. Integration Tests (With DB)

Create in mirror path:

```bash
# Example: test/integration/contexts/identity/contact/repository_test.go
mkdir -p test/integration/contexts/identity/contact
touch test/integration/contexts/identity/contact/repository_test.go
```

Write test:

```go
package contact_test

import (
    "testing"
    "github.com/basilex/promenade/test/integration"
)

func TestContactRepository_Create(t *testing.T) {
    db, cleanup := integration.SetupTestDB(t)
    defer cleanup()

    repo := postgres.NewContactRepository(db)
    ctx := context.Background()

    userID := uuidv7.New()
    contact, _ := contact.NewEmailContact(userID, "test@example.com", "Work")

    err := repo.Create(ctx, contact)
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, contact.ID)
}
```

---

## Test Helpers

Located in `test/integration/`:

- `testutils.go` - Database setup, cleanup, test utilities
- `SetupTestDB(t)` - Initialize test database with migrations
- Helper functions for creating test data

**Usage**:

```go
db, cleanup := integration.SetupTestDB(t)
defer cleanup()
```

---

## Continuous Testing

### Watch Mode (Manual)

```bash
# Terminal 1: Watch for changes
fswatch -o internal/contexts/identity/contact/*.go | xargs -n1 -I{} go test -v ./internal/contexts/identity/contact

# Or use VS Code extension: "Go Test Explorer"
```

### Pre-commit Hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
make test-unit || exit 1
```

---

## What We Test

**Unit Tests**: Entities, use cases, value objects, handlers  
**Smoke Tests**: HTTP handlers with mocks (123 tests across 15 handlers)   
**Integration Tests**: Repositories with real database  
**Benchmark Tests**: Performance measurement and optimization validation  
**Package Tests**: Shared utilities (bus, logger, uuidv7)

**Not Yet**: End-to-end tests, UI tests, load tests

---

---

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Effective Go - Testing](https://go.dev/doc/effective_go#testing)

---

**Test Status**: All tests passing | 138 smoke tests (100% pass rate) | 420+ total tests
