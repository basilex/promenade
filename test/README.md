# Test Directory

Testing infrastructure for Promenade project.

## Quick Start

```bash
# Start test database
make test-db-start

# Run all integration tests
make test-integration

# Run with coverage
make test-coverage

# Stop test database
make test-db-stop
```

## Structure

```
test/
├── helpers/
│   ├── database.go   # Test DB management
│   └── fixtures.go   # Test data generators
├── integration/      # Integration tests (TODO)
├── e2e/             # End-to-end tests (TODO)
└── mocks/           # Mock implementations (TODO)
```

## Helpers

### `database.go`

```go
// Setup test database connection
testDB := helpers.SetupTestDB(t)
defer testDB.Close()

// Clean all tables
testDB.CleanupTables(t)

// Run in transaction (auto-rollback)
testDB.RunInTransaction(t, func(tx *sqlx.Tx) {
    // Your test code
})
```

### `fixtures.go`

```go
// Create test users
user := helpers.UserFixture()
unverified := helpers.UnverifiedUserFixture()
suspended := helpers.SuspendedUserFixture()
banned := helpers.BannedUserFixture()

// Custom user
user := helpers.UserFixture(func(u *entity.User) {
    u.Email = "custom@test.com"
    u.Status = entity.UserStatusInactive
})

// Create test sessions
session := helpers.SessionFixture(userID)
expired := helpers.ExpiredSessionFixture(userID)
```

## Test Database

- **Host**: localhost
- **Port**: 5433 (not 5432!)
- **Database**: promenade_test
- **User**: system
- **Password**: passw0rd

Isolated from development database.

## Running Tests

### Via Make

```bash
make test               # All tests (unit + integration)
make test-unit          # Only unit tests
make test-integration   # Only integration tests
make test-coverage      # Generate coverage report
make test-watch         # Watch mode
```

### Via Script

```bash
./scripts/run-tests.sh all          # All tests
./scripts/run-tests.sh unit         # Unit only
./scripts/run-tests.sh integration  # Integration only
./scripts/run-tests.sh coverage     # With coverage
```

### Direct Go Test

```bash
# Single package
go test -v ./internal/adapter/repository/postgres

# With race detector
go test -race ./...

# Specific test
go test -v ./internal/adapter/repository/postgres -run TestUserRepository_Create
```

## Writing Tests

### Integration Test Template

```go
package postgres_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/basilex/promenade/internal/adapter/repository/postgres"
    "github.com/basilex/promenade/test/helpers"
)

func TestYourFeature(t *testing.T) {
    ctx := context.Background()

    t.Run("test case 1", func(t *testing.T) {
        testDB := helpers.SetupTestDB(t)
        defer testDB.Close()
        defer testDB.CleanupTables(t)

        repo := postgres.NewYourRepository(testDB.DB)

        // Your test code
        entity := helpers.YourFixture()
        err := repo.Create(ctx, entity)
        require.NoError(t, err)

        // Assertions
        retrieved, err := repo.GetByID(ctx, entity.ID)
        require.NoError(t, err)
        assert.Equal(t, entity.Field, retrieved.Field)
    })
}
```

## Best Practices

[+] **Do**:

- Setup/teardown per subtest
- Use fixtures for test data
- Use `require` for critical checks
- Use `assert` for non-critical checks
- Clean tables after tests
- Test edge cases

[X] **Don't**:

- Reuse data between tests
- Forget cleanup
- Use production database
- Hardcode test data
- Create test dependencies

## Documentation

- [TESTING_GUIDE.md](../docs/TESTING_GUIDE.md) - Complete guide
- [TESTING_INFRASTRUCTURE.md](../docs/TESTING_INFRASTRUCTURE.md) - Architecture
- [TEST_RESULTS.md](../docs/TEST_RESULTS.md) - Latest results

## Current Coverage

```
Repository Layer:  45.3% [+]
UUID Package:      88.9% [+]
Use Cases:         0.0%  ⏳ (TODO)
Handlers:          0.0%  ⏳ (TODO)
```

## CI/CD Integration

```yaml
# .github/workflows/test.yml
steps:
  - name: Start test database
    run: make test-db-start

  - name: Run tests
    run: make test-coverage

  - name: Upload coverage
    uses: codecov/codecov-action@v3
    with:
      files: ./coverage.out

  - name: Cleanup
    if: always()
    run: make test-db-stop
```

## Troubleshooting

### Connection refused

```bash
# Ensure test DB is running
docker ps | grep promenade_test_db

# Or start it
make test-db-start
```

### Table doesn't exist

```bash
# Run migrations
make migrate-test-up
```

### Too many connections

```bash
# Restart test DB
make test-db-stop
make test-db-start
```

## Next Steps

1. ⏳ Add use case unit tests with mocks
2. ⏳ Add handler integration tests
3. ⏳ Add E2E tests
4. ⏳ Add benchmark tests
5. ⏳ Integrate with CI/CD

---

**Status**: Repository tests complete [+]  
**Test Count**: 13 integration tests  
**Success Rate**: 100%
