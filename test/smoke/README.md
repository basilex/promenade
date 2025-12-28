# Smoke Tests

Fast smoke tests for HTTP handlers with **mirror path** structure (mock-based, no database).

##  Structure (Mirror Path)

```
test/smoke/contexts/          # Mirror: internal/contexts/
 shared/                   # Mirror: internal/contexts/shared/
    country/
       handler_test.go   # Country handler smoke tests
    currency/
       handler_test.go   # Currency handler smoke tests
    language/
       handler_test.go   # Language handler smoke tests
    timezone/
        handler_test.go   # Timezone handler smoke tests
 identity/                 # Mirror: internal/contexts/identity/
     contact/
         handler_test.go   # Contact handler smoke tests
```

##  Purpose

Smoke tests validate **basic HTTP handler functionality** without external dependencies:

-  Fast execution (< 1 second)
-  No database required (mocked use cases)
-  HTTP status code validation
-  Request/response structure testing

##  Usage

```bash
# All smoke tests
make test-smoke

# Specific context
go test ./test/smoke/contexts/shared/... -v
go test ./test/smoke/contexts/identity/... -v

# Single aggregate
go test ./test/smoke/contexts/shared/country -v
```

##  Coverage: 27 tests (5 handlers)

### Shared Context (4 handlers, 20 tests)

-  Country (5 tests) - List, GetByCode, Create, Update, Delete
-  Currency (5 tests) - List, GetByCode, Create, Update, Delete
-  Language (5 tests) - List, GetByCode, Create, Update, Delete
-  Timezone (5 tests) - List, GetByName, Create, Update, Delete

### Identity Context (1 handler, 7 tests)

-  Contact (7 tests) - Create, GetByID, List, Update, Delete, Verify, SetPrimary

**Execution time**: ~0.35s

##  Writing New Smoke Tests

1. Create directory matching production code:

   ```bash
   mkdir -p test/smoke/contexts/{context}/{aggregate}
   ```

2. Create `handler_test.go`:

   ```go
   package {aggregate}_test

   func TestHandler_Smoke(t *testing.T) {
       // Mock use case
       // Setup router
       // Test HTTP endpoints
   }
   ```

3. Follow existing patterns (see `country/handler_test.go`)
