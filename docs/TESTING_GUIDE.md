# Testing Guide for Promenade

## Overview

A comprehensive testing system covering all application layers with **388 tests (100% passing)**:

- **Unit Tests** (183) - isolated business logic and entity validation
- **Integration Tests** (91) - repository operations with real PostgreSQL
- **Smoke Tests** (114) - end-to-end critical flows with real database
- **E2E Tests** - HTTP API tests (TODO)

## Quick Start

```bash
# Run all tests (unit + integration)
make test                  # 274 tests in ~41s

# Individual test suites
make test-unit            # 183 unit tests (~5s)
make test-integration     # 91 integration tests (~36s)
make test-smoke           # 114 smoke tests (~4s)

# Coverage and monitoring
make test-coverage        # HTML coverage report
make test-watch           # Watch mode (gotestsum)
```

## Test Database

Integration tests use a separate test database on port **5433**:

```bash
# Start test DB
make test-db-start

# Stop and clean
make test-db-stop

# View logs
make test-db-logs
```

**Important:** Test DB is completely isolated from dev/prod databases.

## Test Structure

### 1. Integration Tests (Repositories)

Located alongside code: `internal/adapter/repository/postgres/*_test.go`

Example:

```go
func TestUserRepository_Create(t *testing.T) {
    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    repo := postgres.NewUserRepository(testDB.DB)
    ctx := context.Background()

    t.Run("creates user successfully", func(t *testing.T) {
        user := helpers.UserFixture()
        err := repo.Create(ctx, user)
        require.NoError(t, err)

        retrieved, err := repo.GetByID(ctx, user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.Email, retrieved.Email)
    })
}
```

**Coverage:**

- [+] UserRepository: Create, GetByID, GetByEmail, UpdateStatus, Suspend, Ban, Reactivate, VerifyEmail
- [+] SessionRepository: Create, GetByID, GetByRefreshToken, GetUserSessions, DeleteByUserID, DeleteExpired

### 2. Unit Tests (Use Cases)

_TODO: Next step_

Will test business logic with mocked repositories:

- Register
- Login
- RefreshToken
- Logout
- SuspendUser
- BanUser
- ReactivateUser
- ChangePassword

### 3. Smoke Tests (End-to-End Critical Flows)

Located in: `test/smoke/*_smoke_test.go`

**114 smoke tests** verify critical user flows with real database operations.

Example:

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

    t.Logf("🎉 All auth smoke tests passed!")
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

**Coverage by Module:**

| Module           | Scenarios | Coverage                                                 |
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
2. ⏳ Use case unit tests with mocks
3. ⏳ HTTP handler integration tests
4. ⏳ E2E tests for complete flows
5. ⏳ Performance/benchmark tests

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
