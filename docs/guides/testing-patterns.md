# Testing Patterns Guide

**Comprehensive guide for writing tests in Promenade Platform (DDD Architecture)**

---

##  Table of Contents

1. [Testing Philosophy](#-testing-philosophy)
2. [Three-Tier Testing Strategy](#-three-tier-testing-strategy)
3. [Unit Test Pattern](#-unit-test-pattern)
4. [Integration Test Pattern](#-integration-test-pattern)
5. [Benchmark Test Pattern](#-benchmark-test-pattern)
6. [Best Practices](#-best-practices)
7. [Common Pitfalls](#-common-pitfalls)
8. [Test Coverage Requirements](#-test-coverage-requirements)

---

##  Testing Philosophy

Promenade follows **strict DDD principles** with **three-tier testing strategy**:

1. **Unit Tests** - Fast feedback, test business logic in isolation
2. **Integration Tests** - Full E2E with real database
3. **Benchmark Tests** - Performance measurement and optimization validation

**Key Principle**: Each test type has **ONE standardized pattern** - maximum pattern consistency for easy understanding and maintenance.

---

##  Three-Tier Testing Strategy

### Overview

```
Unit Tests (in-place)          → Fast (5s)    → Business logic
Integration Tests (real DB)    → Medium (14s) → Repository E2E + Full validation
Benchmark Tests (real DB)      → Variable     → Performance measurement
```

### When to Use Each Type

| Test Type       | When to Use                             | What to Test                   | Dependencies    |
| --------------- | --------------------------------------- | ------------------------------ | --------------- |
| **Unit**        | Business logic, entities, value objects | Domain rules, validation       | None (pure Go)  |
| **Integration** | Repository, database queries            | SQL operations, transactions   | Real PostgreSQL |
| **Benchmark**   | After optimizations, N+1 fixes          | Query performance, memory      | Real PostgreSQL |

---

##  Integration Test Pattern

### Directory Structure

**Mirror path structure** - integration tests mirror production code:

```
internal/contexts/identity/user/
 entity.go
 usecase.go
 adapter/repository/postgres/
     user_repository.go

test/integration/contexts/identity/user/
 repository_test.go             ← mirrors adapter/repository/postgres/
```

### Standardized Pattern

```go
package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

//  REFERENCE PATTERN for integration tests:

func TestUserRepository_Create(t *testing.T) {
	// 1. SetupTestDBWithCleanTables - TRUNCATE all tables (clean DB for each test function)
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	// 2. Subtests - use UUID only when uniqueness is needed within a test function
	t.Run("create user successfully", func(t *testing.T) {
		// Static email OK - SetupTestDBWithCleanTables provides a clean DB
		u, err := user.NewUser("test@example.com", "password123")
		require.NoError(t, err)

		err = repo.Create(ctx, u)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, u.ID)
	})

	t.Run("duplicate email fails", func(t *testing.T) {
		// Reuse same email - DB was truncated, no conflict
		u1, _ := user.NewUser("test@example.com", "password123")
		u2, _ := user.NewUser("test@example.com", "password456")

		require.NoError(t, repo.Create(ctx, u1))
		err := repo.Create(ctx, u2)
		assert.Error(t, err) // Should fail on duplicate email
	})
}

// If uniqueness is required between subtests:
func TestUserRepository_ListUsers(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	// Create multiple users - include UUID in email for uniqueness
	for i := 0; i < 5; i++ {
		email := fmt.Sprintf("user_%d@example.com", i)
		u, _ := user.NewUser(email, "password123")
		require.NoError(t, repo.Create(ctx, u))
	}

	t.Run("list all users", func(t *testing.T) {
		users, total, err := repo.ListUsers(ctx, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Len(t, users, 5)
	})
}
```

### Key Rules

####  DO

1. **Use `SetupTestDBWithCleanTables(t)`** - TRUNCATE all tables at start of each test function
2. **Static emails for single-entity tests** - No UUID pollution
3. **UUID only for intra-function uniqueness** - Multiple entities in ONE test function
4. **Context first parameter** - `func Method(ctx context.Context, ...)`
5. **require.NoError for critical checks** - Stop test on setup failure
6. **assert for expectations** - Continue test to see all failures

####  DON'T

1. **DON'T use `SetupTestDB(t)`** - It doesn't clean between test functions (causes duplicate key errors)
2. **DON'T use WithTransaction** - Requires repository signature changes (*sqlx.Tx vs *sqlx.DB)
3. **DON'T use CleanAllTables manually** - SetupTestDBWithCleanTables already does it
4. **DON'T share state between test functions** - Each function is independent
5. **DON'T use random data generators** - Use predictable test data

### Helper Functions

#### createTestUser (for Contact/Profile tests)

```go
// Helper creates test user with UUID in email (for uniqueness between test functions)
func createTestUser(t *testing.T, db *integration.TestDB, userID uuidv7.UUID) {
	t.Helper()

	uniqueEmail := fmt.Sprintf("testuser_%s@example.com", userID.String())

	_, err := db.DB.Exec(`
		INSERT INTO identity_users (id, email, password_hash, status)
		VALUES ($1, $2, $3, $4)
	`, userID, uniqueEmail, "password_hash", "active")
	require.NoError(t, err)
}

func TestContactRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewContactRepository(db.DB)
	ctx := context.Background()

	// Create test user first (for FK constraint)
	userID := uuidv7.New()
	createTestUser(t, db, userID)

	t.Run("create email contact", func(t *testing.T) {
		c, _ := contact.NewEmailContact(userID, "work@example.com", "Work")
		err := repo.Create(ctx, c)
		require.NoError(t, err)
	})
}
```

### Running Integration Tests

```bash
# All integration tests (auto-starts test DB)
make test-integration

# Specific context
go test ./test/integration/contexts/identity/... -v

# Specific aggregate
go test ./test/integration/contexts/identity/user -v

# With coverage
go test ./test/integration/... -cover
```

---

##  Unit Test Pattern

### Purpose

Test **business logic in isolation** - entities, use cases, value objects.

### Location

**In-place** - same directory as production code:

```
internal/contexts/identity/user/
 entity.go
 entity_test.go          ← unit test
 usecase.go
 usecase_test.go         ← unit test
```

### Entity Tests

```go
package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestUser_NewUser(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		wantErr   bool
		errContains string
	}{
		{
			name:     "valid user",
			email:    "test@example.com",
			password: "ValidPass123!",
			wantErr:  false,
		},
		{
			name:        "invalid email",
			email:       "not-an-email",
			password:    "ValidPass123!",
			wantErr:     true,
			errContains: "invalid email",
		},
		{
			name:        "weak password",
			email:       "test@example.com",
			password:    "123",
			wantErr:     true,
			errContains: "password must be at least",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.email, tt.password)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEqual(t, uuidv7.Nil, user.ID)
				assert.Equal(t, UserStatusActive, user.Status)
			}
		})
	}
}

func TestUser_VerifyEmail(t *testing.T) {
	u, _ := NewUser("test@example.com", "ValidPass123!")
	assert.False(t, u.IsEmailVerified)

	u.VerifyEmail()

	assert.True(t, u.IsEmailVerified)
	assert.NotNil(t, u.EmailVerifiedAt)
}

func TestUser_Suspend(t *testing.T) {
	u, _ := NewUser("test@example.com", "ValidPass123!")
	assert.Equal(t, UserStatusActive, u.Status)

	err := u.Suspend()

	require.NoError(t, err)
	assert.Equal(t, UserStatusSuspended, u.Status)
}

func TestUser_Suspend_AlreadySuspended(t *testing.T) {
	u, _ := NewUser("test@example.com", "ValidPass123!")
	u.Suspend()

	err := u.Suspend()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already suspended")
}
```

### UseCase Tests (with Mock Repository)

```go
package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockRepository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

// ... implement other methods

func TestUseCase_Register(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := user.NewUseCase(mockRepo)
	ctx := context.Background()

	t.Run("register new user successfully", func(t *testing.T) {
		email := "newuser@example.com"
		name := "New User"
		password := "ValidPass123!"

		// Mock: email doesn't exist
		mockRepo.On("ExistsByEmail", ctx, email).Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()

		u, err := uc.Register(ctx, email, name, password)

		require.NoError(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, email, u.Email.Value())
		mockRepo.AssertExpectations(t)
	})

	t.Run("register with existing email fails", func(t *testing.T) {
		email := "existing@example.com"

		// Mock: email already exists
		mockRepo.On("ExistsByEmail", ctx, email).Return(true, nil).Once()

		u, err := uc.Register(ctx, email, "Name", "ValidPass123!")

		assert.Error(t, err)
		assert.ErrorIs(t, err, user.ErrEmailAlreadyExists)
		assert.Nil(t, u)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_SuspendUser(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := user.NewUseCase(mockRepo)
	ctx := context.Background()

	t.Run("suspend active user", func(t *testing.T) {
		userID := uuidv7.New()
		u, _ := user.NewUser("user@example.com", "ValidPass123!")
		u.ID = userID

		mockRepo.On("GetByID", ctx, userID).Return(u, nil).Once()
		mockRepo.On("Update", ctx, u).Return(nil).Once()

		err := uc.SuspendUser(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, user.UserStatusSuspended, u.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("suspend non-existent user fails", func(t *testing.T) {
		userID := uuidv7.New()

		mockRepo.On("GetByID", ctx, userID).Return(nil, user.ErrUserNotFound).Once()

		err := uc.SuspendUser(ctx, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, user.ErrUserNotFound)
		mockRepo.AssertExpectations(t)
	})
}
```

### Running Unit Tests

```bash
# All unit tests (fast)
make test-unit

# Specific package
go test ./internal/contexts/identity/user -v

# With coverage
go test ./internal/contexts/identity/user -cover
```

---

##  Benchmark Test Pattern

### Purpose

**Performance measurement** and **optimization validation** with real database. Used to validate N+1 query fixes, measure query performance, and track performance regressions.

### Directory Structure

**Mirror path structure** in `test/benchmark/contexts/`:

```
test/benchmark/contexts/identity/user/
 repository_bench_test.go    ← Benchmark tests with real DB

internal/contexts/identity/user/
 adapter/repository/postgres/
    user_repository.go        ← Production repository
```

### Standardized Pattern

```go
package user_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	"github.com/basilex/promenade/test/integration"
)

// BenchmarkUserRepository_ListUsers measures query performance
func BenchmarkUserRepository_ListUsers(b *testing.B) {
	// Setup real database with test data
	db := integration.SetupTestDBWithCleanTables(&testing.T{})
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	// Create test data
	for i := 0; i < 100; i++ {
		u, _ := user.NewUser("user"+strconv.Itoa(i)+"@example.com", "password123")
		_ = repo.Create(ctx, u)
	}

	// Reset timer before benchmark loop
	b.ResetTimer()

	// Benchmark loop
	for i := 0; i < b.N; i++ {
		_, err := repo.ListUsers(ctx, 1, 20)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUserRepository_ListUsers_WithRoles measures N+1 fix
func BenchmarkUserRepository_ListUsers_WithRoles(b *testing.B) {
	db := integration.SetupTestDBWithCleanTables(&testing.T{})
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	// Create test data with roles
	for i := 0; i < 100; i++ {
		u, _ := user.NewUser("user"+strconv.Itoa(i)+"@example.com", "password123")
		_ = repo.Create(ctx, u)
		// Assign roles
		_ = repo.AssignRole(ctx, u.ID, "user")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.ListUsers(ctx, 1, 20)
		if err != nil {
			b.Fatal(err)
		}
	}
}
```

### Key Rules

####  DO

1. **Use real database** - Benchmarks measure real query performance
2. **Create test data** - Populate database with realistic dataset
3. **Reset timer** - Call `b.ResetTimer()` after setup
4. **Measure specific operations** - Benchmark individual methods
5. **Compare before/after** - Run benchmarks before and after optimizations

####  DON'T

1. **DON'T use mocks** - Benchmarks need real database
2. **DON'T forget ResetTimer** - Setup time skews results
3. **DON'T test trivial operations** - Focus on expensive operations (queries, aggregations)
4. **DON'T ignore memory** - Use `-benchmem` flag to track allocations

### Running Benchmark Tests

```bash
# All benchmarks (5s per benchmark)
make test-benchmark

# Extended benchmarks (10s per benchmark)
make test-benchmark-all

# Specific benchmark
go test -bench=. ./test/benchmark/contexts/identity/user -v

# With memory stats
go test -bench=. -benchmem ./test/benchmark/contexts/identity/user

# Compare before/after optimization
go test -bench=ListUsers -benchmem ./test/benchmark/contexts/identity/user > before.txt
# Apply optimization
go test -bench=ListUsers -benchmem ./test/benchmark/contexts/identity/user > after.txt
benchcmp before.txt after.txt
```

---

##  Best Practices

### General

1. **Table-Driven Tests** - Use for multiple scenarios
2. **Clear Test Names** - Describe expected behavior
3. **AAA Pattern** - Arrange, Act, Assert
4. **One Assertion Per Test** - Or use subtests
5. **No Test Interdependence** - Tests can run in any order
6. **Context Propagation** - Always pass `ctx` as first parameter

### Naming

```go
//  Good
func TestUser_NewUser_InvalidEmail_ReturnsError(t *testing.T) {}
func TestUserRepository_Create_DuplicateEmail_ReturnsError(t *testing.T) {}
func TestUserHandler_Register_MissingPassword_Returns400(t *testing.T) {}

//  Bad
func TestUser(t *testing.T) {}
func Test1(t *testing.T) {}
func TestCreate(t *testing.T) {}
```

### Assertions

```go
//  Use require for critical setup
require.NoError(t, err, "failed to create test user")

//  Use assert for expectations
assert.Equal(t, expected, actual)
assert.Contains(t, err.Error(), "expected substring")

//  Don't swallow errors
if err != nil {
	// Silent failure - BAD
}
```

### Test Data

```go
//  Predictable test data
email := "test@example.com"
name := "Test User"
password := "TestPass123!"

//  Random data (hard to debug failures)
email := fmt.Sprintf("user_%d@example.com", rand.Int())
```

### Mock Setup

```go
//  Clear mock expectations
mockRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()
mockRepo.AssertExpectations(t) // Verify all mocks called

//  Vague mocks
mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
```

---

##  Common Pitfalls

### Integration Tests

####  Using SetupTestDB instead of SetupTestDBWithCleanTables

```go
//  BAD - shared DB, duplicate key errors
func TestUserRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t) // Doesn't clean between functions
	// ...
}

func TestUserRepository_Update(t *testing.T) {
	db := integration.SetupTestDB(t) // Same DB, leftover data
	// ...
}
```

```go
//  GOOD - clean DB for each function
func TestUserRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t) // TRUNCATE all tables
	// ...
}

func TestUserRepository_Update(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t) // Fresh start
	// ...
}
```

####  Over-using UUID in emails

```go
//  BAD - UUID pollution
func TestUserRepository_GetByID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)

	email := fmt.Sprintf("test_%s@example.com", uuidv7.New().String()) // Unnecessary!
	u, _ := user.NewUser(email, "password")
	// ...
}
```

```go
//  GOOD - static email (DB is clean)
func TestUserRepository_GetByID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)

	u, _ := user.NewUser("test@example.com", "password") // Simple and clear
	// ...
}
```

####  Using WithTransaction

```go
//  BAD - requires repository changes
func TestUserRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	tx, _ := db.BeginTx(ctx, nil)
	defer tx.Rollback()

	repo := postgres.NewUserRepository(tx) // Expects *sqlx.DB not *sqlx.Tx
	// ...
}
```

```go
//  GOOD - use SetupTestDBWithCleanTables
func TestUserRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB) // Works with *sqlx.DB
	// ...
}
```

### Unit Tests

####  Testing implementation details

```go
//  BAD - testing internal methods
func TestUser_hashPassword(t *testing.T) {
	hash, err := hashPassword("password")
	// Testing private method
}
```

```go
//  GOOD - testing public behavior
func TestUser_NewUser_PasswordIsHashed(t *testing.T) {
	u, _ := NewUser("test@example.com", "password123")
	assert.NotEqual(t, "password123", u.PasswordHash) // Verify password was hashed
}
```

####  Not using mock.AssertExpectations

```go
//  BAD - mock setup but no verification
mockRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil)
uc.Register(ctx, "test@example.com", "Name", "password")
// Did Create get called? We don't know!
```

```go
//  GOOD - verify all mocks called
mockRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()
uc.Register(ctx, "test@example.com", "Name", "password")
mockRepo.AssertExpectations(t) // Fails if Create not called exactly once
```

---

##  Test Coverage Requirements

### Per Component

| Component         | Coverage | Why                          |
| ----------------- | -------- | ---------------------------- |
| **Entities**      | 95%+     | Core business logic          |
| **Use Cases**     | 85%+     | Business operations          |
| **Repositories**  | 80%+     | Data access                  |
| **Handlers**      | 70%+     | HTTP contracts (unit tests)  |
| **Value Objects** | 95%+     | Domain primitives            |

### Overall Project

- **Minimum**: 80% total coverage
- **Target**: 90% total coverage
- **Critical paths**: 100% (authentication, authorization, payment)

### Checking Coverage

```bash
# All tests with coverage
make test-coverage

# Specific package
go test ./internal/contexts/identity/user -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Coverage report
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | grep total
```

---

##  Additional Resources

- [Main Testing Guide](../test/README.md) - Testing overview
- [Testing Structure](../test/TESTING_STRUCTURE.md) - Directory organization
- [Integration Test Utilities](../test/integration/testutils.go) - Helper functions
- [DDD Architecture](concepts/clean-architecture.md) - Domain-Driven Design principles
- [Event Bus Testing](../pkg/bus/README.md) - Event-driven testing

---

##  Summary Checklist

Before writing tests, verify:

- [ ] Correct test type (unit/integration/benchmark)?
- [ ] Using standardized pattern?
- [ ] Mirror path structure for integration tests?
- [ ] SetupTestDBWithCleanTables for integration tests?
- [ ] Real database for integration tests?
- [ ] Table-driven tests for multiple scenarios?
- [ ] Clear, descriptive test names?
- [ ] AssertExpectations for all mocks?
- [ ] No UUID pollution in test data?
- [ ] Context propagation (`ctx` first parameter)?

---

**Last Updated**: 2025-12-28  
**Status**:  Production-ready  
**Maintainer**: Promenade Team
