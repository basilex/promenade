# Integration Tests

Integration tests for Promenade - testing repositories with real PostgreSQL database.

---

## Overview

Integration tests verify that repositories correctly interact with the database, including:

- **CRUD operations** - Create, Read, Update, Delete
- **Complex queries** - Joins, filters, pagination
- **Transactions** - Multi-step operations
- **Constraints** - Foreign keys, unique constraints
- **Soft deletes** - Proper filtering of deleted records

---

## Setup

### 1. Start Test Database

```bash
# Start PostgreSQL via Docker
make docker-up

# Create test database (one-time setup)
make test-integration-setup
```

### 2. Environment Variables (Optional)

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=promenade
export TEST_DB_PASSWORD=promenade
export TEST_DB_NAME=promenade_test
```

---

## Running Tests

### All Integration Tests

```bash
make test-integration
```

### Core Repository Tests

```bash
make test-integration-core
```

Tests:

- `UserRepository` - User CRUD, GetByEmail, ExistsByEmail
- `RoleRepository` - Role management, GetUserRoles
- `PermissionRepository` - Permissions, GetRolePermissions
- `SessionRepository` - Sessions, GetByRefreshToken, CountUserSessions
- `CountryRepository` - Countries, ListByRegion
- `CurrencyRepository` - Currencies
- `LanguageRepository` - Languages, ListActive
- `TimezoneRepository` - Timezones

### Module Repository Tests

```bash
# Posts module
make test-integration-posts

# Profiles module
make test-integration-profiles

# Analytics module
make test-integration-analytics
```

**Posts Module Tests:**

- `PostRepository` - Post CRUD, ListByUserID, ListByStatus, soft delete
- `CommentRepository` - Comment CRUD, ListByPostID, nested comments
- `LikeRepository` - Like/Unlike, CountByPostID

**Profiles Module Tests:**

- `UserProfileRepository` - Profile CRUD, GetByUserID, IncrementProfileViews
- `UserContactRepository` - Contact CRUD, GetByUserID

**Analytics Module Tests:**

- `MetricRepository` - Store metrics, Query, Aggregate, DeleteOlderThan

### Check Database Status

```bash
make test-integration-check
```

---

## Test Structure

### Setup Helper

`test/integration/setup.go` provides:

```go
// Setup test database with migrations
testDB := integration.SetupTestDB(t)

// Setup with clean tables
testDB := integration.SetupTestDBWithCleanTables(t)

// Clean all tables manually
testDB.CleanAllTables()

// Run code in transaction (auto-rollback)
testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
    // Test code here
})

// Execute query that must succeed
testDB.MustExec(t, "INSERT INTO ...", args...)

// Get context with logger
ctx := testDB.GetContext()
```

### Fixtures Helper

`test/integration/fixtures.go` provides data creation:

```go
fixtures := integration.NewFixtures(testDB.DB)

// Create test user
user := fixtures.CreateUser(t, "test@example.com", "password123")

// Create role and assign to user
role := fixtures.CreateRole(t, "admin", "Admin Role", false)
fixtures.AssignRoleToUser(t, user.ID, role.ID)

// Create permission and assign to role
perm := fixtures.CreatePermission(t, "posts", "create", "Create posts")
fixtures.AssignPermissionToRole(t, role.ID, perm.ID)

// Create session
session := fixtures.CreateSession(t, user.ID)

// Create reference data
country := fixtures.CreateCountry(t, "USA", "US", "US", "USA", "north_america")
currency := fixtures.CreateCurrency(t, "US Dollar", "USD", "$")
language := fixtures.CreateLanguage(t, "English", "English", "en", "eng")
timezone := fixtures.CreateTimezone(t, "America/New_York", "EST", "-05:00")
```

### Example Test

```go
func TestUserRepository_Integration(t *testing.T) {
    testDB := integration.SetupTestDBWithCleanTables(t)
    fixtures := integration.NewFixtures(testDB.DB)

    repo := postgres.NewUserRepository(testDB.DB)
    ctx := testDB.GetContext()

    t.Run("Create and GetByID", func(t *testing.T) {
        user := fixtures.CreateUser(t, "test@example.com", "password123")

        retrieved, err := repo.GetByID(ctx, user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.Email, retrieved.Email)
    })
}
```

---

## Test Organization

```
internal/
├── adapter/
│   └── repository/
│       └── postgres/
│           ├── user_repository.go
│           ├── integration_test.go              # ← Core repository tests
│           └── reference_integration_test.go    # ← Reference data tests
│
└── modules/
    ├── posts/
    │   └── adapter/
    │       └── repository/
    │           └── postgres/
    │               ├── post_repository.go
    │               └── integration_test.go       # ← Posts module tests
    │
    └── profiles/
        └── adapter/
            └── repository/
                └── postgres/
                    └── integration_test.go       # ← Profiles module tests
```

---

## Test Coverage

### Core Repositories

- [x] **UserRepository** - 7 tests
  - Create, GetByID, GetByEmail, Update, Delete, ExistsByEmail, List
- [x] **RoleRepository** - 5 tests
  - GetByID, GetByName, List, ExistsByName, GetUserRoles
- [x] **PermissionRepository** - 5 tests
  - GetByID, GetByResourceAction, List, FindByResource, GetRolePermissions
- [x] **SessionRepository** - 5 tests
  - GetByID, GetByRefreshToken, GetUserSessions, CountUserSessions, Delete

### Reference Data Repositories

- [x] **CountryRepository** - 5 tests
  - GetByID, GetByCode, List, ListByRegion, ExistsByCode
- [x] **CurrencyRepository** - 4 tests
  - GetByID, GetByCode, List, ExistsByCode
- [x] **LanguageRepository** - 5 tests
  - GetByID, GetByCode, List, ListActive, ExistsByCode
- [x] **TimezoneRepository** - 5 tests
  - GetByID, GetByName, List, ListActive, ExistsByName

### Module Repositories

- [x] **PostRepository** (posts module) - 7 tests
  - Create, GetByID, Update, Delete (soft), ListByUserID, ListByStatus, CountByUserID
- [x] **CommentRepository** (posts module) - 7 tests
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
```

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
