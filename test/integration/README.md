# Integration Tests

Integration tests with **real database** organized by **mirror path** structure.

## Structure

```
test/integration/
 testutils.go              # Shared test utilities
 contexts/                 # Mirror: internal/contexts/
    shared/               # Mirror: internal/contexts/shared/
       country/
          repository_test.go  #  Repository integration tests
       currency/
          repository_test.go  #  Repository integration tests
       language/
          repository_test.go  #  Repository integration tests
       timezone/
           repository_test.go  #  Repository integration tests
    identity/             # Mirror: internal/contexts/identity/
        contact/
            repository_test.go  #  Repository integration tests
 pkg/
     bus/
         bus_integration_test.go
```

## Coverage

### Repository Integration Tests (5 aggregates, ~30 tests)

-  **Country** (6 tests) - Create, GetByID, GetByCode, Update, Delete, List
-  **Currency** (6 tests) - Create, GetByID, GetByCode, Update, Delete, List
-  **Language** (6 tests) - Create, GetByID, GetByCode, Update, Delete, List
-  **Timezone** (6 tests) - Create, GetByID, GetByName, Update, Delete, List
-  **Contact** (TBD tests) - Repository CRUD operations

All tests use real PostgreSQL database via `testutils.SetupTestDB()`.

## Purpose

Integration tests validate **full functionality** with real external dependencies:

-  Real PostgreSQL database
-  Real Redis (Event Bus)
-  Full HTTP request/response cycle
-  Database transactions and rollbacks
-  End-to-end workflows

## Usage

```bash
# All integration tests (starts test DB automatically)
make test-integration

# Specific context
go test ./test/integration/contexts/shared/... -v
go test ./test/integration/contexts/identity/... -v

# Single aggregate
go test ./test/integration/contexts/shared/country -v
```

## Key Characteristics

| Aspect   | Integration Tests             |
| -------- | ----------------------------- |
| Location | `/test/integration/contexts/` |
| Database |  Yes (real)                 |
| Speed    | Medium (~14s for all)         |
| Purpose  | Full E2E validation           |
| When     | Before merge/deploy           |

Integration tests use real database connections to validate SQL queries, foreign key constraints, transactions, and full repository behavior. Unlike unit tests with mocks, these catch real database issues.

## Writing Integration Tests

1. Create test file:

   ```bash
   touch test/integration/contexts/{context}/{aggregate}/handler_integration_test.go
   ```

2. Use real database:

   ```go
   package country_test

   func TestCountryHandler_Integration(t *testing.T) {
       // Setup real database
       db := testutils.SetupTestDB(t)
       defer testutils.CleanupTestDB(t, db)

       // Create real repository (not mock)
       repo := repository.NewCountryRepository(db)
       usecase := country.NewUseCase(repo)
       handler := handler.NewHandler(usecase)

       // Test real HTTP → DB flow
   }
   ```

3. Clean up after each test to ensure isolation

## Future Integration Tests

- Real database schema validation
- Transaction rollback testing
- Foreign key constraints
- Database triggers
- Full CRUD workflows with persistence
