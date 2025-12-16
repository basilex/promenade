# Testing Guide for Promenade

## Overview

A comprehensive testing system covering all application layers:

- **Unit Tests** - isolated business logic tests
- **Integration Tests** - tests with real database
- **E2E Tests** - end-to-end tests via HTTP API

## Quick Start

```bash
# Run all tests
make test

# Unit tests only
make test-unit

# Integration tests only (with test DB)
make test-integration

# Coverage report
make test-coverage

# Watch mode (installs gotestsum if needed)
make test-watch
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

- ✅ UserRepository: Create, GetByID, GetByEmail, UpdateStatus, Suspend, Ban, Reactivate, VerifyEmail
- ✅ SessionRepository: Create, GetByID, GetByRefreshToken, GetUserSessions, DeleteByUserID, DeleteExpired

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

### 3. HTTP Integration Tests (Handlers)

_TODO: After unit tests_

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

### ✅ Do

- Use `testify/require` for critical checks (stops test)
- Use `testify/assert` for non-critical checks (continues test)
- Always do cleanup: `defer testDB.CleanupTables(t)`
- Test edge cases: expired sessions, banned users, etc.
- Use fixtures for consistent test data

### ❌ Don't

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

1. ✅ Repository integration tests - **DONE**
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

**Questions?** Check `Makefile.test` for all available commands.
