# Test Integration Utilities

**Shared test utilities** for Promenade - helpers for setting up test databases, fixtures, and common test scenarios.

---

## Overview

This directory contains **shared test utilities**, not tests themselves. All tests are located **in-place** alongside the code they test.

### What's Here

- **`testutils.go`** - Database setup helpers, test fixtures, common assertions
- **`README.md`** - This documentation

### What's NOT Here

❌ Actual test files (they live in-place with production code)  
❌ Integration tests (moved to `*_test.go` files alongside repositories)  
❌ Unit tests (always in-place with code)

---

## In-Place Testing

All tests now follow **in-place testing** pattern:

```
pkg/bus/
├── bus.go
├── bus_test.go                 # ✅ Unit tests
├── bus_integration_test.go     # ✅ Integration tests
└── memory/
    ├── memory_bus.go
    └── memory_bus_test.go      # ✅ Adapter tests

internal/contexts/shared/country/
├── entity.go
├── entity_test.go              # ✅ Entity tests
├── usecase.go
├── usecase_test.go             # ✅ UseCase tests
└── adapter/repository/postgres/
    ├── country_repository.go
    └── country_repository_test.go  # ✅ DB integration tests
```

---

## Using Test Utilities

make test-integration-notifications

````

**Posts Module Tests:**

- `PostRepository` - Post CRUD, ListByUserID, ListByStatus, soft delete
- `ICommentRepository` - Comment CRUD, ListByPostID, nested comments
- `LikeRepository` - Like/Unlike, CountByPostID

**Profiles Module Tests:**

- `IUserProfileRepository` - Profile CRUD, GetByUserID, IncrementProfileViews
- `IUserContactRepository` - Contact CRUD, GetByUserID

**Analytics Module Tests:**

- `IMetricRepository` - Store metrics, Query, Aggregate, DeleteOlderThan

**Notifications Module Tests:**

- `INotificationRepository` - Notification CRUD, ListByUserID, GetUnreadCount, UpdateStatus
- `IUserPreferenceRepository` - Preferences CRUD, GetByUserID, quiet hours logic

### Check Database Status

```bash
make test-integration-check
````

---

## Using Test Utilities

### Import in Your Tests

```go
import "github.com/basilex/promenade/test/integration"
```

### Setup Test Database

```go
// Setup test database with migrations
testDB := integration.SetupTestDB(t)

// Clean all tables
testDB.CleanAllTables()

// Run code in transaction (auto-rollback)
testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
    // Test code here
})

// Get context with logger
ctx := testDB.GetContext()
```

### Example: Repository Integration Test

```go
package postgres

import (
    "testing"
    "github.com/basilex/promenade/test/integration"
)

func TestCountryRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    testDB := integration.SetupTestDB(t)
    defer testDB.CleanAllTables()

    repo := NewCountryRepository(testDB.DB)
    ctx := testDB.GetContext()

    t.Run("Create and GetByID", func(t *testing.T) {
        country, err := repo.Create(ctx, &Country{...})
        assert.NoError(t, err)

        retrieved, err := repo.GetByID(ctx, country.ID)
        assert.NoError(t, err)
        assert.Equal(t, country.Name, retrieved.Name)
    })
}
```

---

## Running Integration Tests

### All Tests

```bash
go test ./... -v
```

### Context-Specific

```bash
go test ./internal/contexts/shared/... -v
go test ./internal/contexts/identity/... -v
```

### Skip Integration Tests

```bash
go test ./... -short  # Skips tests with testing.Short() check
```

---

## Database Requirements

Integration tests require PostgreSQL:

```bash
# Start test database
make docker-up

# Run migrations
make migrate
```

        require.NoError(t, err)
        assert.Equal(t, user.Email, retrieved.Email)
    })

}

```

---

## Test Organization

```

internal/
adapter/
repository/
postgres/
user_repository.go
integration_test.go # ← Core repository tests
reference_integration_test.go # ← Reference data tests

modules/
posts/
adapter/
repository/
postgres/
post_repository.go
integration_test.go # ← Posts module tests

     profiles/
         adapter/
             repository/
                 postgres/
                     integration_test.go       # ← Profiles module tests

````

---

## Test Coverage

### Core Repositories

- [x] **IUserRepository** - 7 tests
  - Create, GetByID, GetByEmail, Update, Delete, ExistsByEmail, List
- [x] **IRoleRepository** - 5 tests
  - GetByID, GetByName, List, ExistsByName, GetUserRoles
- [x] **IPermissionRepository** - 5 tests
  - GetByID, GetByResourceAction, List, FindByResource, GetRolePermissions
- [x] **ISessionRepository** - 5 tests
  - GetByID, GetByRefreshToken, GetUserSessions, CountUserSessions, Delete

### Reference Data Repositories

- [x] **ICountryRepository** - 5 tests
  - GetByID, GetByCode, List, ListByRegion, ExistsByCode
- [x] **ICurrencyRepository** - 4 tests
  - GetByID, GetByCode, List, ExistsByCode
- [x] **ILanguageRepository** - 5 tests
  - GetByID, GetByCode, List, ListActive, ExistsByCode
- [x] **ITimezoneRepository** - 5 tests
  - GetByID, GetByName, List, ListActive, ExistsByName

### Module Repositories

- [x] **PostRepository** (posts module) - 7 tests
  - Create, GetByID, Update, Delete (soft), ListByUserID, ListByStatus, CountByUserID
- [x] **ICommentRepository** (posts module) - 7 tests
  - Create, GetByID, Create reply, ListByPostID, ListByUserID, CountByPostID, Delete (soft)
- [x] **LikeRepository** (posts module) - 4 tests
  - Create, GetByID, UserHasLikedPost, CountByPostID, Delete

**Total**: 54 integration tests across all repositories

---

## Best Practices

### 1. Clean State

Always start with clean tables:

```go
testDB := integration.SetupTestDBWithCleanTables(t)
````

### 2. Use Fixtures

Don't manually insert test data - use fixtures:

```go
//  Don't do this
testDB.MustExec(t, "INSERT INTO users ...", ...)

//  Do this instead
user := fixtures.CreateUser(t, "test@example.com", "password123")
```

### 3. Test Real Scenarios

Test realistic data flows:

```go
t.Run("User with roles and permissions", func(t *testing.T) {
    user := fixtures.CreateUser(t, "admin@example.com", "password123")
    role := fixtures.CreateRole(t, "admin", "Admin", false)
    perm := fixtures.CreatePermission(t, "posts", "delete", "Delete posts")

    fixtures.AssignRoleToUser(t, user.ID, role.ID)
    fixtures.AssignPermissionToRole(t, role.ID, perm.ID)

    // Now test permission checking
    perms, err := repo.GetUserPermissions(ctx, user.ID)
    require.NoError(t, err)
    assert.Len(t, perms, 1)
})
```

### 4. Soft Delete Testing

Always verify soft delete behavior:

```go
t.Run("Soft deleted records not returned", func(t *testing.T) {
    post := createTestPost(t)
    repo.Delete(ctx, post.ID)

    // Should not be found
    _, err := repo.GetByID(ctx, post.ID)
    assert.Error(t, err)

    // Should not appear in lists
    posts, _, _ := repo.List(ctx, 1, 10)
    for _, p := range posts {
        assert.NotEqual(t, post.ID, p.ID)
    }
})
```

### 5. Pagination Testing

Test edge cases:

```go
t.Run("Pagination edge cases", func(t *testing.T) {
    // Page 0 should return page 1
    posts, _, err := repo.List(ctx, 0, 10)
    require.NoError(t, err)

    // Large page size should be limited
    posts, _, err = repo.List(ctx, 1, 10000)
    require.NoError(t, err)
    assert.LessOrEqual(t, len(posts), 100) // max page size
})
```

---

## Troubleshooting

### Test Database Not Available

```bash
 Test database not available. Run 'make docker-up' first
```

**Solution**:

```bash
make docker-up
make test-integration-setup
```

### Migration Errors

```bash
Failed to run migrations: ...
```

**Solution**:

```bash
# Check migration files
ls -la migrations/core/
ls -la migrations/posts/

# Verify database connection
psql -h localhost -U promenade -d promenade_test -c "\dt"
```

### Dirty Database State

Tests fail due to leftover data.

**Solution**:

```go
// Use clean tables setup
testDB := integration.SetupTestDBWithCleanTables(t)

// Or manually clean specific tables
testDB.CleanAllTables()
```

### Foreign Key Violations

```bash
ERROR: insert or update on table violates foreign key constraint
```

**Solution**:
Create parent records first using fixtures:

```go
//  Wrong order
post := createPost(t, "nonexistent-user-id")

//  Correct order
user := fixtures.CreateUser(t, "test@example.com", "password123")
post := createPost(t, user.ID)
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Run integration tests
  run: |
    docker-compose up -d postgres
    sleep 5
    make test-integration-setup
    make test-integration
  env:
    TEST_DB_HOST: localhost
    TEST_DB_PORT: 5432
    TEST_DB_USER: promenade
    TEST_DB_PASSWORD: promenade
    TEST_DB_NAME: promenade_test
```

---

## Next Steps

- [ ] Add profiles module integration tests
- [ ] Add analytics module integration tests
- [ ] Add warehouse module integration tests
- [ ] Transaction rollback tests
- [ ] Concurrent access tests
- [ ] Performance benchmarks
