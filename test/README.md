# Test Directory

Testing infrastructure for Promenade project.

## Quick Start

```bash
# Run all tests (auto-manages test DB)
make test              # Unit + Integration (274 tests)

# Individual test suites
make test-unit         # Unit tests only (183 tests)
make test-integration  # Integration tests (91 tests)
make test-smoke        # Smoke tests (114 tests, end-to-end flows)

# Coverage and monitoring
make test-coverage     # Generate HTML coverage report
make test-watch        # Watch mode

# Manual DB management
make test-db-start     # Start test database (port 5433)
make test-db-stop      # Stop test database
```

## Structure

```
test/
├── helpers/
│   ├── database.go   # Test DB management (SetupTestDB, CleanupTables)
│   └── fixtures.go   # Test data generators (UserFixture, SessionFixture, etc.)
├── smoke/           # End-to-end smoke tests [+] (114 tests, 9 files)
│   ├── auth_smoke_test.go                 # Auth flow (8 scenarios)
│   ├── country_currency_smoke_test.go     # Country/Currency (12 scenarios)
│   ├── user_contact_smoke_test.go         # Contacts (11 scenarios)
│   ├── user_post_smoke_test.go            # Posts (12 scenarios)
│   ├── user_profile_smoke_test.go         # Profiles (12 scenarios)
│   ├── post_comment_smoke_test.go         # Comments (13 scenarios)
│   ├── comment_likes_smoke_test.go        # Likes (5 scenarios)
│   ├── rbac_smoke_test.go                 # RBAC (28 scenarios)
│   └── rbac_integration_smoke_test.go     # RBAC Integration (13 scenarios)
├── integration/      # Integration tests (TODO - future expansion)
├── e2e/             # E2E tests (TODO - future expansion)
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

### Via Make (Recommended)

```bash
make test               # Unit + Integration (274 tests)
make test-unit          # Unit tests only (183 tests)
make test-integration   # Integration tests only (91 tests)
make test-smoke         # Smoke tests (114 tests, end-to-end flows)
make test-coverage      # Generate HTML coverage report
make test-watch         # Watch mode (gotestsum)
```

### Via Script

```bash
./scripts/run-tests.sh all          # All tests
./scripts/run-tests.sh unit         # Unit only
./scripts/run-tests.sh integration  # Integration only
./scripts/run-tests.sh smoke        # Smoke tests only
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

## Smoke Tests

**End-to-end critical flow testing** with real database operations.

```bash
# Run all smoke tests
make test-smoke

# Run specific smoke test
go test -v ./test/smoke -run TestAuth_SmokeTest
go test -v ./test/smoke -run TestRBAC_SmokeTest
```

**Smoke Test Template:**

```go
package smoke

import (
    "context"
    "testing"
    "github.com/basilex/promenade/test/helpers"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFeature_SmokeTest(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping smoke test in short mode")
    }

    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    ctx := context.Background()

    t.Run("[+] Critical_flow", func(t *testing.T) {
        // Test critical user flow end-to-end
    })

    t.Logf("🎉 All feature smoke tests passed!")
}
```

**Key Features:**

- [+] Real database integration (PostgreSQL on port 5433)
- [+] Critical path verification (create → read → update → delete)
- [+] Performance benchmarks (100 permission checks in <500ms)
- [+] 100% passing rate, ~4 seconds execution
- [+] Automatic cleanup between tests

## Best Practices

[+] **Do**:

- Setup/teardown per subtest
- Use fixtures for test data
- Use `require` for critical checks
- Use `assert` for non-critical checks
- Clean tables after tests
- Test edge cases
- Write smoke tests for critical flows

[X] **Don't**:

- Reuse data between tests
- Forget cleanup
- Use production database
- Hardcode test data
- Create test dependencies
- Skip smoke tests before deployment

## Documentation

- [TESTING_GUIDE.md](../docs/TESTING_GUIDE.md) - Complete guide
- [TESTING_INFRASTRUCTURE.md](../docs/TESTING_INFRASTRUCTURE.md) - Architecture
- [TEST_RESULTS.md](../docs/TEST_RESULTS.md) - Latest results

## Test Statistics

**Total: 388 tests - 100% passing [+]**

```
Unit Tests:        183 tests [+] (entity validation, domain logic)
Integration Tests:  91 tests [+] (repository operations with PostgreSQL)
Smoke Tests:       114 tests [+] (end-to-end critical flows)
```

**Test Execution Time:**

- Unit: ~5 seconds
- Integration: ~36 seconds
- Smoke: ~4 seconds
- **Total: ~45 seconds**

**Coverage by Module:**

| Module           | Unit    | Integration | Smoke   | Total   |
| ---------------- | ------- | ----------- | ------- | ------- |
| Country/Currency | 18      | 4           | 12      | 34      |
| Auth/Session     | 9       | 10          | 8       | 27      |
| UserContact      | 14      | 13          | 11      | 38      |
| UserPost         | 42      | 30          | 12      | 84      |
| PostComment      | 0       | 22          | 13      | 35      |
| UserProfile      | 31      | 13          | 12      | 56      |
| User/RBAC        | 64      | 9           | 0       | 73      |
| Permission/Role  | 31      | 16          | 41      | 88      |
| CommentLikes     | 0       | 1           | 5       | 6       |
| BaseRepository   | 0       | 7           | 0       | 7       |
| **Total**        | **183** | **91**      | **114** | **388** |

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

**Status**: All test layers complete [+]  
**Test Count**: 388 tests (183 unit + 91 integration + 114 smoke)  
**Success Rate**: 100%  
**Execution Time**: ~45 seconds for full suite
