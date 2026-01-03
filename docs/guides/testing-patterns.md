# Testing Patterns Guide

**Comprehensive guide for writing tests in Promenade Platform (DDD Architecture)**

---

## 📖 Table of Contents

1. [Testing Philosophy](#-testing-philosophy)
2. [Three-Tier Testing Strategy](#-three-tier-testing-strategy)
3. [Integration Test Pattern](#-integration-test-pattern)
4. [Smoke Test Pattern](#-smoke-test-pattern)
5. [Unit Test Pattern](#-unit-test-pattern)
6. [Best Practices](#-best-practices)
7. [Common Pitfalls](#-common-pitfalls)
8. [Test Coverage Requirements](#-test-coverage-requirements)

---

##  Testing Philosophy

Promenade follows **strict DDD principles** with **three-tier testing strategy**:

1. **Unit Tests** - Fast feedback, test business logic in isolation
2. **Smoke Tests** - Mock-based handler validation, NO database
3. **Integration Tests** - Full E2E with real database

**Key Principle**: Each test type has **ONE standardized pattern** - максимальна ідентичність патерну для легкого розуміння та підтримки.

---

## 🏗 Three-Tier Testing Strategy

### Overview

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

## 🔵 Integration Test Pattern

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

//  ЕТАЛОННИЙ ПАТЕРН для інтеграційних тестів:

func TestUserRepository_Create(t *testing.T) {
	// 1. SetupTestDBWithCleanTables - TRUNCATE всіх таблиць (чиста DB для кожного test function)
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	// 2. Subtests - використовувати UUID тільки якщо потрібна uniqueness в межах test function
	t.Run("create user successfully", func(t *testing.T) {
		// Static email OK - SetupTestDBWithCleanTables дає чисту DB
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

// Якщо потрібна uniqueness між subtests:
func TestUserRepository_ListUsers(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewUserRepository(db.DB)
	ctx := context.Background()

	// Створюємо кілька users - UUID в email для uniqueness
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

## 🟡 Smoke Test Pattern

### Purpose

**Fast validation** of HTTP handler contracts **without database**. Mock all UseCase dependencies.

### Directory Structure

**Mirror path structure**:

```
internal/contexts/identity/user/adapter/http/
 handler.go

test/smoke/contexts/identity/user/
 handler_test.go                ← mirrors adapter/http/
```

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

//  ЕТАЛОННИЙ ПАТЕРН для smoke тестів:

// 1. MockUseCase - імплементує ВСІ методи з IUseCase interface
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

func (m *MockUserUseCase) GetUser(ctx context.Context, userID uuidv7.UUID) (*user.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

// ... implement ALL IUseCase methods

// 2. Router setup - gin test mode
func setupUserRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// 3. ONE test function for entire handler
func TestUserHandler_Smoke(t *testing.T) {
	mockUC := new(MockUserUseCase)
	handler := userHTTP.NewUserHandler(mockUC)
	router := setupUserRouter()

	// Register routes
	router.POST("/users/register", handler.Register)
	router.GET("/users/:id", handler.GetByID)
	router.POST("/users/:id/suspend", handler.Suspend)

	// 4. SUBTEST for each endpoint
	t.Run("Register returns 201", func(t *testing.T) {
		userID := uuidv7.New()

		// Create mock entity (no real hashing/validation needed)
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		// Mock usecase call
		mockUC.On("Register", mock.Anything, "test@example.com", "Test User", "password123").Return(u, nil).Once()

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

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code, "Register should return 201")
		mockUC.AssertExpectations(t)
	})

	t.Run("GetByID returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusActive,
		}

		mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GetByID should return 200")
		mockUC.AssertExpectations(t)
	})

	t.Run("Suspend returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		u := &user.User{
			ID:           userID,
			PasswordHash: "hashedPassword",
			Status:       user.UserStatusSuspended,
		}

		// ⚠️ ВАЖЛИВО: Handler може викликати КІЛЬКА методів UseCase
		// Додати mock для ВСІХ викликів
		mockUC.On("SuspendUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
		mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once() // Handler retrieves updated user

		req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/suspend", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Suspend should return 200")
		mockUC.AssertExpectations(t)
	})
}
```

### Key Rules

####  DO

1. **Implement ALL IUseCase methods** - Mock must satisfy interface completely
2. **Use gin.TestMode** - Disable debug logging
3. **ONE test function per handler** - TestXxxHandler_Smoke with subtests
4. **Mock ALL usecase calls** - Handler may call multiple methods (not just primary operation)
5. **Create entities directly** - No need for factory methods (NewEntity)
6. **AssertExpectations in each subtest** - Verify all mocks were called

####  DON'T

1. **DON'T use real database** - Smoke tests are mock-only
2. **DON'T call factory methods** - Create entity structs directly
3. **DON'T forget secondary mock calls** - Many handlers call GetEntity after main operation
4. **DON'T test business logic** - Only HTTP contract (status codes, request/response)
5. **DON'T use testify/suite** - Keep it simple with subtests

### Critical Pattern: Multiple Mock Calls

Many handlers call **UseCase method + GetEntity** to return updated entity:

```go
// Handler.go
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	// Primary operation
	err := h.usecase.VerifyEmail(ctx, userID)
	if err != nil {
		// handle error
	}

	// SECOND CALL - retrieve updated user
	user, err := h.usecase.GetUser(ctx, userID)
	if err != nil {
		// handle error
	}

	response.Success(c, ToUserResponse(user))
}
```

**Smoke test MUST mock BOTH calls**:

```go
t.Run("VerifyEmail returns 200", func(t *testing.T) {
	userID := uuidv7.New()
	u := &user.User{ID: userID, Status: user.UserStatusActive}

	// ⚠️ Mock PRIMARY call
	mockUC.On("VerifyEmail", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()

	// ⚠️ Mock SECONDARY call (handler retrieves updated user)
	mockUC.On("GetUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/users/"+userID.String()+"/verify-email", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
})
```

### Running Smoke Tests

```bash
# All smoke tests
make test-smoke

# Specific context
go test ./test/smoke/contexts/identity/... -v

# Specific aggregate
go test ./test/smoke/contexts/identity/user -v
```

---

## 🟢 Unit Test Pattern

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

## ⚠️ Common Pitfalls

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

### Smoke Tests

####  Forgetting secondary mock calls

```go
//  BAD - missing GetUser mock
t.Run("VerifyEmail returns 200", func(t *testing.T) {
	mockUC.On("VerifyEmail", mock.Anything, userID).Return(nil).Once()
	// Handler ALSO calls GetUser but no mock! → PANIC
})
```

```go
//  GOOD - mock ALL handler calls
t.Run("VerifyEmail returns 200", func(t *testing.T) {
	u := &user.User{ID: userID}
	mockUC.On("VerifyEmail", mock.Anything, userID).Return(nil).Once()
	mockUC.On("GetUser", mock.Anything, userID).Return(u, nil).Once() // Don't forget!
})
```

####  Using factory methods

```go
//  BAD - NewUser hashes password, validates email (unnecessary in smoke test)
u, err := user.NewUser("test@example.com", "password123")
if err != nil {
	// Handle validation error in smoke test!?
}
```

```go
//  GOOD - create struct directly
u := &user.User{
	ID:           uuidv7.New(),
	PasswordHash: "hashedPassword", // No real hashing
	Status:       user.UserStatusActive,
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
go test ./internal/contexts/identity/user -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Coverage report
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | grep total
```

---

## 📚 Additional Resources

- [Main Testing Guide](../test/README.md) - Testing overview
- [Testing Structure](../test/TESTING_STRUCTURE.md) - Directory organization
- [Integration Test Utilities](../test/integration/testutils.go) - Helper functions
- [DDD Architecture](concepts/clean-architecture.md) - Domain-Driven Design principles
- [Event Bus Testing](../pkg/bus/README.md) - Event-driven testing

---

## 🎓 Summary Checklist

Before writing tests, verify:

- [ ] Correct test type (unit/smoke/integration)?
- [ ] Using standardized pattern?
- [ ] Mirror path structure (integration/smoke)?
- [ ] SetupTestDBWithCleanTables for integration tests?
- [ ] Mock ALL usecase calls for smoke tests?
- [ ] Table-driven tests for multiple scenarios?
- [ ] Clear, descriptive test names?
- [ ] AssertExpectations for all mocks?
- [ ] No UUID pollution in test data?
- [ ] Context propagation (`ctx` first parameter)?

---

**Last Updated**: 2025-12-28  
**Status**:  Production-ready  
**Maintainer**: Promenade Team
