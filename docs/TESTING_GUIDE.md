# Testing Guide for Promenade

## Overview

A comprehensive testing system covering all application layers with **400+ tests (100% passing)**:

- **Core Entity Tests** (39) - domain entities and validation
- **Core UseCase Tests** (236) - business logic with auth, RBAC, purge operations
- **Module Entity Tests** (65) - Posts (33, 83.3% coverage), Profiles (21, 80.4% coverage), Analytics (11)
- **Utilities Tests** (51, 89.5% avg) - response, validator, logger, pagination

## Quick Start

```bash
# Run all tests (~20 seconds)
make test                  # 400+ tests across all layers

# Individual test suites
make test-core             # Core tests (275 tests: 39 entity + 236 usecase)
make test-modules          # All module tests
make test-module-posts     # Posts module (33 tests, 83.3% coverage)
make test-module-profiles  # Profiles module (21 tests, 80.4% coverage)
make test-module-analytics # Analytics module (11 tests)

# Coverage and monitoring
make test-coverage         # HTML coverage report
```

## Test Structure

### 1. Core Entity Tests (39 tests)

Located alongside code: `internal/domain/entity/*_test.go`

**Entities tested:**

- Country (validation, ISO codes, currency relationships)
- Currency (validation, symbols, country relationships)
- Language (ISO codes, native names)
- Timezone (IANA database, UTC offsets)
- Permission (resource-action format, wildcards)
- Role (RBAC roles, system roles)
- User (password hashing/checking, status management)
- Session (token generation, expiration, refresh)
- Purge (retention policies, entity tracking)

Example:

```go
func TestUser_HashPassword(t *testing.T) {
    user := &entity.User{Email: "test@example.com"}

    err := user.HashPassword("password123")
    require.NoError(t, err)
    assert.NotEmpty(t, user.Password)

    ok := user.CheckPassword("password123")
    assert.True(t, ok)
}
```

**Coverage:** Comprehensive coverage of all domain entity business logic.

---

### 2. Core UseCase Tests (236 tests)

Located alongside code: `internal/usecase/*_test.go`

**Auth UseCase** (Register, Login, Logout, RefreshToken, GetMe, Sessions):

- User registration with email/password validation
- Login with credential verification
- Logout and session cleanup
- JWT token refresh logic
- Get current authenticated user
- Session management operations

**RBAC UseCases** (Country, Currency, Language, Permission, Role, Timezone):

- CountryUseCase: CRUD operations, currency associations
- CurrencyUseCase: CRUD operations, country relationships
- LanguageUseCase: CRUD operations, active lists
- PermissionUseCase: Create, get, list, update, delete, batch operations
- RoleUseCase: CRUD operations, permission assignments, user role management
- TimezoneUseCase: CRUD operations, active timezone lists

**Purge UseCase**:

- Purge entities by retention policies
- Policy registry management
- Preview purge operations before execution

Example:

```go
func TestAuthUseCase_Register(t *testing.T) {
    mockUserRepo := mocks.NewMockUserRepository(t)
    uc := usecase.NewAuthUseCase(mockUserRepo, nil, jwtManager, eventBus)

    mockUserRepo.EXPECT().GetByEmail(mock.Anything, "new@example.com").Return(nil, entity.ErrNotFound)
    mockUserRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil)

    user, err := uc.Register(ctx, "new@example.com", "John Doe", "password123")
    require.NoError(t, err)
    assert.Equal(t, "new@example.com", user.Email)
}
```

---

### 3. Module Entity Tests

**Posts Module** (33 tests, 83.3% coverage):

Located: `internal/modules/posts/domain/entity/*_test.go`

Tests cover:

- PostStatus lifecycle and valid transitions
- UserPost creation, validation, publish/unpublish
- Archive and schedule functionality
- Comment counters and interactions
- Soft delete operations
- Featured image management
- Slug generation and uniqueness validation

**Profiles Module** (21 tests, 80.4% coverage):

Located: `internal/modules/profiles/entity/*_test.go`

Tests cover:

- UserContact validation and availability checks
- UserProfile field management
- Gender and privacy settings
- Social link validation
- User preferences
- Ban/unban operations
- Email verification
- View tracking and last seen timestamps

**Analytics Module** (11 tests):

Located: `internal/modules/analytics/domain/entity/*_test.go` and `usecase/*_test.go`

Tests cover:

- Metric validation (types, scopes, modules)
- MetricAggregate calculations
- UseCase operations: collect metric, get metric, list metrics by scope/module, get latest metric

Example:

```go
func TestUserPost_Publish(t *testing.T) {
    post := &entity.UserPost{
        Status: entity.PostStatusDraft,
    }

    post.Publish()

    assert.Equal(t, entity.PostStatusPublished, post.Status)
    assert.NotNil(t, post.PublishedAt)
    assert.Nil(t, post.ScheduledAt)
}
```

---

### 4. Utilities Tests (51 tests, 89.5% avg coverage)

**response** (26 tests, 100% coverage):

- Success/error responses
- Pagination responses
- Query parameter parsing (page, page_size)

**validator** (5 tests, 80% coverage):

- Custom validators (slug, timezone, language code)
- Validation error handling

**logger** (12 tests, 83.8% coverage):

- Context-aware logging with request IDs
- Log level configuration
- Structured logging with slog

**pagination** (8 tests, 94.1% coverage):

- Page/size calculation
- Offset/limit conversion
- Default values and max limits

```go
func TestAuth_SmokeTest(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping smoke test in short mode")
    }

    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    ctx := context.Background()

    t.Run("[+] Complete_auth_flow", func(t *testing.T) {
        // Register → Login → GetMe → Refresh → Logout
        user, err := authUC.Register(ctx, "test@example.com", "John", "password123")
        require.NoError(t, err)

        tokens, err := authUC.Login(ctx, "test@example.com", "password123", "test-device")
        require.NoError(t, err)
        assert.NotEmpty(t, tokens.AccessToken)
        assert.NotEmpty(t, tokens.RefreshToken)

        // Continue testing full flow...
    })

    t.Logf("[SUCCESS] All auth smoke tests passed!")
}
```

**Running Smoke Tests:**

```bash
# All smoke tests
make test-smoke

# Specific smoke test
go test -v ./test/smoke -run TestRBAC_SmokeTest
go test -v ./test/smoke -run TestUserPost_SmokeTest

# Skip in short mode
go test -short ./test/smoke  # Smoke tests are skipped
```

**Coverage by IModule:**

| IModule          | Scenarios | Coverage                                                 |
| ---------------- | --------- | -------------------------------------------------------- |
| Auth             | 8         | Register, login, sessions, refresh, logout               |
| Country/Currency | 12        | Complete CRUD operations                                 |
| UserContact      | 11        | Email, phone, telegram, primary, verification            |
| UserPost         | 12        | Draft, publish, featured, schedule, views, search        |
| UserProfile      | 12        | Privacy, verification, ban/unban, views, search          |
| PostComment      | 13        | Threading, replies, nested replies, soft delete          |
| CommentLikes     | 5         | Like/unlike, pagination, performance (100 checks)        |
| RBAC             | 28        | Permissions, roles, wildcards, expiration                |
| RBAC Integration | 13        | Real-world permission scenarios (moderator, admin, etc.) |
| **Total**        | **114**   | **All tests passing [+]**                                |

**Key Features:**

- [+] Real PostgreSQL integration (port 5433)
- [+] Critical path verification (CRUD flows)
- [+] Performance benchmarks included
- [+] Fast execution (~4 seconds for 114 tests)
- [+] 100% passing rate

### 4. HTTP Integration Tests (Handlers)

_TODO: After smoke tests expansion_

Test all HTTP endpoints via real Gin router:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `GET /api/auth/me`
- etc.

## Test Helpers

### `test/helpers/database.go`

```go
// Connect to test DB
testDB := helpers.SetupTestDB(t)
defer testDB.Close()

// Clean all tables
testDB.CleanupTables(t)

// Transactional test (auto rollback)
testDB.RunInTransaction(t, func(tx *sqlx.Tx) {
    // Your code with tx
})
```

### `test/helpers/fixtures.go`

```go
// Standard active user
user := helpers.UserFixture()

// Unverified user
user := helpers.UnverifiedUserFixture()

// Suspended user
user := helpers.SuspendedUserFixture()

// Banned user
user := helpers.BannedUserFixture()

// Custom user
user := helpers.UserFixture(func(u *entity.User) {
    u.Email = "custom@test.com"
    u.Status = entity.UserStatusInactive
})

// Session
session := helpers.SessionFixture(userID)

// Expired session
session := helpers.ExpiredSessionFixture(userID)
```

## Best Practices

### [+] Do

- Use `testify/require` for critical checks (stops test)
- Use `testify/assert` for non-critical checks (continues test)
- Always do cleanup: `defer testDB.CleanupTables(t)`
- Test edge cases: expired sessions, banned users, etc.
- Use fixtures for consistent test data

### [X] Don't

- Don't use production DB for tests
- Don't create dependencies between tests
- Don't forget `defer testDB.Close()`
- Don't hardcode test data - use fixtures

## CI/CD Integration

Tests are CI-ready:

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    make test-db-start
    make test
    make test-db-stop
```

## Coverage

```bash
make test-coverage
open coverage.html
```

Goal: **>80% coverage** for critical modules (usecase, repository).

## What's Next

1. [+] Repository integration tests - **DONE**
2. **TODO** Use case unit tests with mocks
3. **TODO** HTTP handler integration tests
4. **TODO** E2E tests for complete flows
5. **TODO** Performance/benchmark tests

## Examples

### Running Specific Tests

```bash
# Single test file
go test -v ./internal/adapter/repository/postgres/user_repository_test.go

# Single test
go test -v ./internal/adapter/repository/postgres -run TestUserRepository_Create

# With race detector
go test -race ./...

# With coverage
go test -cover ./internal/adapter/repository/postgres
```

### Debugging Tests

```bash
# Verbose output
go test -v ./...

# With DB logs
make test-db-logs

# Check DB state during test
docker exec -it promenade_test_db psql -U system -d promenade_test
```

## Troubleshooting

### "connection refused"

```bash
make test-db-start
# Wait 3-5 seconds for DB readiness
```

### "table does not exist"

```bash
make migrate-test-up
```

### "too many open connections"

```bash
make test-db-stop
make test-db-start
```

---

**Questions?** Check `Makefile.test.mk` for all available commands.
