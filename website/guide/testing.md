# Testing Strategy

**Comprehensive three-tier testing for Domain-Driven Design architecture**

---

## Testing Philosophy

Promenade follows **strict DDD principles** with **three-tier testing strategy**:

1. **Unit Tests** - Fast feedback, test business logic in isolation
2. **Smoke Tests** - Mock-based handler validation, NO database
3. **Integration Tests** - Full E2E with real database

**Key Principle**: Each test type has **ONE standardized pattern** for maximum consistency and maintainability.

---

## Three-Tier Strategy Overview

```
Unit Tests (in-place)          → Fast (5s)   → Business logic
Smoke Tests (mock-based)       → Fast (0.4s) → Handler HTTP contracts
Integration Tests (real DB)    → Slow (6s)   → Repository E2E
```

### When to Use Each Type

| Test Type       | When to Use                             | What to Test                   | Dependencies    |
| --------------- | --------------------------------------- | ------------------------------ | --------------- |
| **Unit**        | Business logic, entities, value objects | Domain rules, validation       | None (mocks)    |
| **Smoke**       | HTTP handlers, API contracts            | Request/response, status codes | Mock UseCase    |
| **Integration** | Repository, database queries            | SQL operations, transactions   | Real PostgreSQL |

---

## Integration Test Pattern

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

func TestUserRepository_Create(t *testing.T) {
	// 1. SetupTestDBWithCleanTables - TRUNCATE all tables
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	// 2. Subtests with static test data
	t.Run("create user successfully", func(t *testing.T) {
		u, err := user.NewUser("test@example.com", "password123")
		require.NoError(t, err)

		err = repo.Create(ctx, u)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, u.ID)
	})

	t.Run("duplicate email fails", func(t *testing.T) {
		u1, _ := user.NewUser("test@example.com", "password123")
		u2, _ := user.NewUser("test@example.com", "password456")

		require.NoError(t, repo.Create(ctx, u1))
		err := repo.Create(ctx, u2)
		assert.Error(t, err) // Should fail on duplicate email
	})
}
```

### Key Rules

**DO:**
- ✅ Use `SetupTestDBWithCleanTables(t)` - TRUNCATE all tables at start
- ✅ Static emails for single-entity tests - No UUID pollution
- ✅ `require.NoError` for critical checks - Stop test on failure
- ✅ `assert` for expectations - Continue to see all failures

**DON'T:**
- ❌ Use `SetupTestDB(t)` - Doesn't clean between functions (causes duplicate key errors)
- ❌ Use `WithTransaction` - Requires repository signature changes
- ❌ Share state between test functions - Each function is independent
- ❌ Use random data generators - Use predictable test data

---

## Smoke Test Pattern

### Purpose

**Fast validation** of HTTP handler contracts **without database**. Mock all UseCase dependencies.

### Standardized Pattern

```go
package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	userHTTP "github.com/basilex/promenade/internal/contexts/identity/user/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// 1. MockUseCase - implements ALL IUseCase methods
type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) Register(ctx context.Context, email, name, password string) (*user.User, error) {
	args := m.Called(ctx, email, name, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

// 2. ONE test function per handler
func TestUserHandler_Smoke(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := new(MockUserUseCase)
	handler := userHTTP.NewUserHandler(mockUC)
	router := gin.New()

	router.POST("/users/register", handler.Register)

	// 3. SUBTEST for each endpoint
	t.Run("Register returns 201", func(t *testing.T) {
		u := &user.User{
			ID:           uuidv7.New(),
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		// Mock usecase call
		mockUC.On("Register", mock.Anything, "test@example.com", "Test User", "password123").
			Return(u, nil).Once()

		// HTTP request
		reqBody := userHTTP.RegisterRequest{
			Email:    "test@example.com",
			Name:     "Test User",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUC.AssertExpectations(t)
	})
}
```

### Key Rules

**DO:**
- ✅ Implement ALL IUseCase methods - Mock must satisfy interface completely
- ✅ Use `gin.TestMode` - Disable debug logging
- ✅ ONE test function per handler with subtests
- ✅ Mock ALL usecase calls - Handler may call multiple methods
- ✅ Create entities directly - No need for factory methods
- ✅ `AssertExpectations` in each subtest

**DON'T:**
- ❌ Use real database - Smoke tests are mock-only
- ❌ Call factory methods - Create entity structs directly
- ❌ Forget secondary mock calls - Many handlers call GetEntity after operation
- ❌ Test business logic - Only HTTP contract (status codes, request/response)

---

## Unit Test Pattern

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
)

func TestUser_NewUser(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		wantErr     bool
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
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
			}
		})
	}
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
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func TestUseCase_Register(t *testing.T) {
	mockRepo := new(MockRepository)
	uc := user.NewUseCase(mockRepo)
	ctx := context.Background()

	t.Run("register new user successfully", func(t *testing.T) {
		mockRepo.On("ExistsByEmail", ctx, "new@example.com").Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()

		u, err := uc.Register(ctx, "new@example.com", "Name", "Pass123!")

		require.NoError(t, err)
		assert.NotNil(t, u)
		mockRepo.AssertExpectations(t)
	})

	t.Run("register with existing email fails", func(t *testing.T) {
		mockRepo.On("ExistsByEmail", ctx, "exists@example.com").Return(true, nil).Once()

		u, err := uc.Register(ctx, "exists@example.com", "Name", "Pass123!")

		assert.Error(t, err)
		assert.Nil(t, u)
		mockRepo.AssertExpectations(t)
	})
}
```

---

## Running Tests

### Quick Commands

```bash
# All tests with race detector (~40s)
make test

# By type
make test-unit              # Unit tests only (~5s)
make test-smoke             # Smoke tests (~0.4s)
make test-integration       # Integration tests (~6s)

# With coverage
make test-coverage          # HTML coverage report
```

### Context-Specific Tests

```bash
# Identity Context
go test ./internal/contexts/identity/... -v
go test ./test/smoke/contexts/identity/... -v
go test ./test/integration/contexts/identity/... -v

# Shared Context
go test ./internal/contexts/shared/... -v

# Specific aggregate
go test ./internal/contexts/identity/user -v
```

### Package Tests

```bash
# Event Bus
go test ./pkg/bus/... -v

# JWT
go test ./pkg/jwt/... -v

# All packages
go test ./pkg/... -v
```

---

## Best Practices

### General

1. **Table-Driven Tests** - Use for multiple scenarios
2. **Clear Test Names** - Describe expected behavior
3. **AAA Pattern** - Arrange, Act, Assert
4. **One Assertion Per Test** - Or use subtests
5. **No Test Interdependence** - Tests can run in any order
6. **Context Propagation** - Always pass `ctx` as first parameter

### Naming Convention

```go
// ✅ Good
func TestUser_NewUser_InvalidEmail_ReturnsError(t *testing.T) {}
func TestUserRepository_Create_DuplicateEmail_ReturnsError(t *testing.T) {}
func TestUserHandler_Register_MissingPassword_Returns400(t *testing.T) {}

// ❌ Bad
func TestUser(t *testing.T) {}
func Test1(t *testing.T) {}
```

### Assertions

```go
// Use require for critical setup (stops test on failure)
require.NoError(t, err, "failed to create test user")

// Use assert for expectations (continues test)
assert.Equal(t, expected, actual)
assert.Contains(t, err.Error(), "expected substring")
```

---

## Common Pitfalls

### Integration Tests

**❌ Using SetupTestDB instead of SetupTestDBWithCleanTables**

```go
// BAD - shared DB, duplicate key errors
db := integration.SetupTestDB(t)

// GOOD - clean DB for each function
db := integration.SetupTestDBWithCleanTables(t)
```

**❌ Over-using UUID in emails**

```go
// BAD - unnecessary UUID pollution
email := fmt.Sprintf("test_%s@example.com", uuidv7.New().String())

// GOOD - static email (DB is clean)
email := "test@example.com"
```

### Smoke Tests

**❌ Forgetting secondary mock calls**

```go
// BAD - missing GetUser mock
mockUC.On("VerifyEmail", mock.Anything, userID).Return(nil).Once()
// Handler ALSO calls GetUser but no mock! → PANIC

// GOOD - mock ALL handler calls
mockUC.On("VerifyEmail", mock.Anything, userID).Return(nil).Once()
mockUC.On("GetUser", mock.Anything, userID).Return(u, nil).Once()
```

**❌ Using factory methods**

```go
// BAD - NewUser hashes password, validates email
u, err := user.NewUser("test@example.com", "password")

// GOOD - create struct directly
u := &user.User{
	ID:           uuidv7.New(),
	PasswordHash: "hashedPassword",
	Status:       user.UserStatusActive,
}
```

### Unit Tests

**❌ Not using mock.AssertExpectations**

```go
// BAD - no verification
mockRepo.On("Create", ctx, mock.Anything).Return(nil)
uc.Register(ctx, "test@example.com", "Name", "password")

// GOOD - verify all mocks called
mockRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
uc.Register(ctx, "test@example.com", "Name", "password")
mockRepo.AssertExpectations(t) // Fails if not called exactly once
```

---

## Test Coverage Requirements

### Per Component

| Component         | Coverage | Why                          |
| ----------------- | -------- | ---------------------------- |
| **Entities**      | 95%+     | Core business logic          |
| **Use Cases**     | 85%+     | Business operations          |
| **Repositories**  | 80%+     | Data access                  |
| **Handlers**      | 70%+     | HTTP contracts (smoke tests) |
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
go test ./internal/contexts/identity/user -cover

# HTML report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Test Statistics

| Component                  | Tests | Coverage | Duration | Type        |
| -------------------------- | ----- | -------- | -------- | ----------- |
| **pkg/jwt**                | 18    | 87%      | cached   | Unit        |
| **pkg/bus**                | 67    | 100%     | ~2s      | Unit        |
| **pkg/logger**             | 12    | 95%      | cached   | Unit        |
| **pkg/uuidv7**             | 7     | 100%     | cached   | Unit        |
| **Identity User**          | 15    | 85%      | ~2.1s    | Smoke       |
| **Identity Contact**       | 7     | -        | cached   | Smoke       |
| **Identity Profile**       | 8     | -        | cached   | Smoke       |
| **Customer Management**    | 18    | -        | ~1.6s    | Smoke       |
| **Shared (all contexts)**  | 20    | -        | cached   | Smoke       |
| **Identity (integration)** | 35    | -        | ~2.8s    | Integration |
| **Shared (integration)**   | 24    | -        | ~7.8s    | Integration |
| **Customer (integration)** | 14    | -        | ~3.6s    | Integration |

**Total**: 360+ tests across 45+ packages, 90%+ average coverage

---

## Summary Checklist

Before writing tests, verify:

- [ ] Correct test type (unit/smoke/integration)?
- [ ] Using standardized pattern?
- [ ] Mirror path structure (integration/smoke)?
- [ ] `SetupTestDBWithCleanTables` for integration tests?
- [ ] Mock ALL usecase calls for smoke tests?
- [ ] Table-driven tests for multiple scenarios?
- [ ] Clear, descriptive test names?
- [ ] `AssertExpectations` for all mocks?
- [ ] No UUID pollution in test data?
- [ ] Context propagation (`ctx` first parameter)?

---

## Next Steps

- [Testing Structure](https://github.com/basilex/promenade/tree/main/test/README.md) - Directory organization
- [Integration Test Utilities](https://github.com/basilex/promenade/tree/main/test/integration/testutils.go) - Helper functions
- [DDD Architecture](/guide/architecture) - Domain-Driven Design principles
- [Contributing](/guide/contributing) - Development workflow
