# Smoke Tests

**Quick HTTP-level validation** for all handlers - focus on status codes and basic response structure.

**Status**: **COMPLETE** - 27/27 handlers, 252 tests, 100% pass rate

---

## Completion Summary

### Statistics

- **Total Tests**: 252
- **Total Handlers**: 27 (100% complete)
- **Pass Rate**: 100%
- **Test Duration**: ~1.0 seconds (cached)
- **Contexts Covered**: 8 (Identity, Customer Management, Order Management, Billing, Shared, UI, Fiscal, Accounting)

### Handler Breakdown

**Identity Context** (5 handlers, 43 tests):

- User: 8 tests (Register, Login, GetByID, List)
- Contact: 8 tests
- Profile: 9 tests (includes ListPublic)
- Role: 9 tests (includes Update)
- Permission: 9 tests (includes Update)

**Customer Management Context** (4 handlers, 33 tests):

- Company: 8 tests
- Customer: 11 tests (includes CreateB2B, QualifyAsProspect) Most complex
- Deal: 8 tests
- Interaction: 6 tests

**Order Management Context** (1 handler, 10 tests):

- Order: 10 tests (includes Confirm, Cancel lifecycle methods)

**Shared Context** (4 handlers, 36 tests):

- Country: 9 tests (GetByCode pattern)
- Currency: 9 tests (GetByCode pattern)
- Language: 9 tests (GetByCode pattern)
- Timezone: 9 tests (GetByName with wildcard route)

**Billing Context** (2 handlers, 16 tests):

- Invoice: 8 tests
- Payment: 8 tests

**UI Context** (1 handler, 8 tests):

- Form: 8 tests

**Fiscal Context** (2 handlers, 29 tests):

- Cash Register: 15 tests
- Receipt: 14 tests

**Accounting Context** (7 handlers, 56 tests):

- Account: 9 tests (Chart of Accounts CRUD)
- Budget: 8 tests (Budget management with lines and approval)
- Cost Center: 8 tests (Cost center hierarchy and management)
- Fiscal Period: 7 tests (Period lifecycle with close/reopen)
- Journal Entry: 8 tests (Double-entry bookkeeping)
- Reconciliation: 7 tests (Bank reconciliation workflow)
- Tax Code: 9 tests (Tax configuration management)

---

## Purpose

Smoke tests verify **critical HTTP layer functionality** with minimal setup:

1. **HTTP Status Codes** - 200, 201, 404, 400, 500 returned correctly
2. **Response Format** - JSON structure matches expectations
3. **Basic Routing** - URLs correctly mapped to handlers
4. **Error Handling** - Domain errors properly mapped to HTTP codes

**NOT tested here**:

- Business logic (covered by UseCase tests)
- Database operations (covered by Repository integration tests)
- Complex scenarios (covered by manual QA)

---

## Test Strategy

### Standard Handler Coverage (6-9 tests):

Baseline smoke budget follows [docs/guides/testing-patterns.md](../../docs/guides/testing-patterns.md).

**Minimum** (6 tests):

1. **Create Success** - POST returns 201
2. **Create ValidationError** - Invalid body returns 400
3. **GetByID Success** - GET returns 200
4. **GetByID NotFound** - Invalid ID returns 404
5. **List Success** - GET returns 200 with array
6. **List EmptyResult** - GET returns 200 with empty array

**Extended** (9 tests) - adds: 7. **Update Success** - PUT returns 200 8. **Delete Success** - DELETE returns 200 or 204 9. **Delete NotFound** - Invalid ID returns 404 or 500

### Complex Handler Strategy (12+ tests):

For handlers with 20+ UseCase methods (e.g., Customer - 23 methods):

- **Full mock implementation** (all 23 methods) to prevent compilation errors
- **Selective testing** (12 core tests) covering essential operations
- **Focus**: CRUD + key lifecycle transitions
- **Skip**: Advanced features (stats, bulk operations, complex state management)

---

## Key Patterns & Discoveries

### 1. Mock UseCase Pattern

**Structure**: Function fields + nil-check methods implementing interface

```go
// Mock with function fields
type MockCustomerUseCase struct {
    CreateCustomerFunc func(ctx context.Context, ...) (*customer.Customer, error)
    GetCustomerFunc    func(ctx context.Context, id uuidv7.UUID) (*customer.Customer, error)
    // ... 21 more methods for complex handlers
}

// Nil-check methods implementing ICustomerUseCase
func (m *MockCustomerUseCase) CreateCustomer(ctx context.Context, ...) (*customer.Customer, error) {
    if m.CreateCustomerFunc != nil {
        return m.CreateCustomerFunc(ctx, ...)
    }
    return nil, fmt.Errorf("CreateCustomerFunc not implemented")
}
```

**Key Points**:

- Simple handlers: 5-6 function fields
- Complex handlers (Customer): 23 function fields for full UseCase coverage
- Always return objects, not nil (handlers call multiple methods)
- Mock all methods to prevent compilation errors

### 2. Error Code Patterns by Context

**Identity & Order Management Contexts**:

- Use **simple "NOT_FOUND"** for all not found errors
- Use **"BAD_REQUEST"** for validation errors
- Example: `response.NotFound(c, err.Error())` → "NOT_FOUND"

**Customer Management Context**:

- Uses **standard response helpers** (NotFound, BadRequest, InternalError)
- NOT entity-specific codes (no "CUSTOMER_NOT_FOUND")
- Delete/QualifyAsProspect return **500 INTERNAL_ERROR** on errors (not 404)
- Example: `response.InternalError(c, err.Error())` → "INTERNAL_ERROR"

**Shared Context** (unique):

- Uses **entity-specific codes**: COUNTRY_NOT_FOUND, CURRENCY_NOT_FOUND, LANGUAGE_NOT_FOUND, TIMEZONE_NOT_FOUND
- Uses **VALIDATION_ERROR** (not BAD_REQUEST) for validation
- Delete returns **204 No Content** (no JSON body)
- Example: `response.ErrorCode(c, 404, "COUNTRY_NOT_FOUND", err.Error())`

### 3. Route Patterns

**Standard Routes**:

- Create: `POST /{entity}` → 201 Created
- GetByID: `GET /{entity}/:id` → 200 OK
- List: `GET /{entity}` → 200 OK
- Update: `PUT /{entity}/:id` → 200 OK
- Delete: `DELETE /{entity}/:id` → 200 OK (Customer Management) or 204 No Content (Shared)

**Shared Context Unique Routes**:

- GetByCode: `GET /{entity}/:code` (string parameter, not UUID)
- GetByName: `GET /timezones/*name` (wildcard route for slash support in timezone names like "Europe/Kyiv")
- Update/Delete: Use `:id` UUID parameter (not :code)

### 4. Handler Constructor Patterns

**Discovered Patterns**:

- Identity: `NewUserHandler`, `NewContactHandler`, `NewProfileHandler`
- Customer Management: `NewCustomerHandler`, `NewCompanyHandler`, `NewDealHandler`
- Order Management: `NewOrderHandler`
- Shared: `NewCountryHandler`, `NewCurrencyHandler`, `NewLanguageHandler`, `NewTimezoneHandler`

**Rule**: Always `New{Entity}Handler(usecase)`, never `NewHandler(usecase)`

### 5. Complex Handler Insights (Customer)

**Customer handler** is the most complex in the system:

- **23 UseCase methods** (vs 5-6 for simple handlers)
- **20 HTTP handlers** (vs 5-6 for simple handlers)
- **Multiple method calls per handler**:
  - Create/CreateB2B call: CreateCustomer → SetCustomerPhone → AddTagToCustomer → GetCustomer
  - QualifyAsProspect calls: QualifyAsProspect → GetCustomer

**Strategy Applied**:

- Created **full 23-method mock** to prevent compilation errors
- Implemented **12 strategic tests** (not all 20 handlers)
- Focused on: B2C/B2B creation, CRUD basics, one lifecycle transition
- Skipped: Tier management, tagging, statistics, user linking, reassignment

**Result**: Smoke test coverage achieved without excessive complexity

### 6. Shared Context Quirks

**Timezone Special Handling**:

- DTO uses `utc_offset` as **string "+HH:MM"** (not integer seconds)
- Route uses **wildcard `/*name`** to capture slashes (e.g., "Europe/Kyiv")
- Handler trims leading slash: `strings.TrimPrefix(c.Param("name"), "/")`

**Delete Pattern**:

- Returns **204 No Content** (no JSON body)
- Test asserts: `AssertSuccessResponse(t, w, 204)` (no JSON parsing)

**Error Codes**:

- Delete wraps errors as **DELETE_ERROR** with 500 status
- All other errors use entity-specific codes

---

## Structure

```
test/smoke/
 README.md                  # This file
 testutils.go              # Shared utilities (5 helper functions)
 contexts/                 # Mirror structure (20 handlers, 196 tests)
    identity/               # 5 handlers, 43 tests
       user/handler_test.go
       contact/handler_test.go
       profile/handler_test.go
       role/handler_test.go
       permission/handler_test.go
    customer-mgmt/          # 4 handlers, 33 tests
       company/handler_test.go
       customer/handler_test.go  # Most complex: 12 tests, 23-method mock
       deal/handler_test.go
       interaction/handler_test.go
    order-mgmt/             # 1 handler, 10 tests
       order/handler_test.go
    billing/                # 2 handlers, 16 tests
       invoice/handler_test.go
       payment/handler_test.go
    shared/                 # 4 handlers, 36 tests
       country/handler_test.go
       currency/handler_test.go
       language/handler_test.go
       timezone/handler_test.go
     ui/
         metadata/
             form/handler_test.go
    fiscal/                 # 2 handlers, 15 tests
        cashregister/handler_test.go
        receipt/handler_test.go
```

## Writing Smoke Tests

### 1. Use Helper Utilities (testutils.go)

**Available helpers**:

- `SetupRouter()` - Create Gin test router
- `MakeRequest(t, router, method, path, body)` - Execute HTTP request
- `AssertSuccessResponse(t, resp, expectedCode)` - Validate 200/201 response
- `AssertErrorResponse(t, resp, expectedHTTPCode, expectedErrorCode)` - Validate error response
- `FakeUUID()` - Generate test UUID v7

**Example - Order Handler**:

```go
import "github.com/basilex/promenade/test/smoke"

func TestOrderHandler_Create_Success(t *testing.T) {
    router := smoke.SetupRouter()

    // Mock UseCase with function field
    mockUC := &MockOrderUseCase{
        CreateOrderFunc: func(ctx context.Context, ...) (*order.Order, error) {
            return fakeOrder(), nil  // Always return object, not nil
        },
    }

    // Create handler + router
    handler := orderHTTP.NewOrderHandler(mockUC)
    router.POST("/orders", handler.Create)

    // Make request
    body := map[string]any{
        "customer_id": smoke.FakeUUID(),
        "currency": "USD",
    }
    w := smoke.MakeRequest(t, router, "POST", "/orders", body)

    // Assert
    smoke.AssertSuccessResponse(t, w, 201)
}
```

### 2. Mock Pattern for Simple Handlers

**Example - Contact Handler (6 methods)**:

```go
type MockContactUseCase struct {
    CreateFunc func(ctx context.Context, ...) (*contact.Contact, error)
    GetByIDFunc func(ctx context.Context, id uuidv7.UUID) (*contact.Contact, error)
    ListFunc func(ctx context.Context, ...) ([]*contact.Contact, error)
    UpdateFunc func(ctx context.Context, ...) error
    DeleteFunc func(ctx context.Context, id uuidv7.UUID) error
}

// Implement IContactUseCase with nil checks
func (m *MockContactUseCase) CreateContact(ctx context.Context, ...) (*contact.Contact, error) {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, ...)
    }
    return nil, fmt.Errorf("CreateFunc not implemented")
}
// ... 5 more methods
```

### 3. Mock Pattern for Complex Handlers

**Example - Customer Handler (23 methods)**:

```go
type MockCustomerUseCase struct {
    // Core CRUD
    CreateCustomerFunc func(ctx context.Context, ...) (*customer.Customer, error)
    CreateB2BCustomerFunc func(ctx context.Context, ...) (*customer.Customer, error)
    GetCustomerFunc func(ctx context.Context, id uuidv7.UUID) (*customer.Customer, error)
    UpdateCustomerFunc func(ctx context.Context, ...) error
    DeleteCustomerFunc func(ctx context.Context, id uuidv7.UUID) error

    // Lifecycle
    QualifyAsProspectFunc func(ctx context.Context, customerID uuidv7.UUID) error
    ConvertToCustomerFunc func(ctx context.Context, customerID uuidv7.UUID) error
    ChurnCustomerFunc func(ctx context.Context, customerID uuidv7.UUID) error

    // ... 15 more methods (tier management, tagging, assignments, stats, etc.)
}

// Implement all 23 methods with nil checks (prevents compilation errors)
```

**Key Strategy**:

- Implement **ALL 23 methods** (full mock) to prevent compilation errors
- Test only **12 core operations** (selective testing)
- Focus: B2C/B2B creation, CRUD, one lifecycle transition
- Skip: Advanced features (tiers, tags, stats)

### 4. Handler-Specific Patterns

**Multiple Method Calls** (Create handlers):

```go
// Customer Create calls: CreateCustomer → SetCustomerPhone → AddTagToCustomer → GetCustomer
mockUC := &MockCustomerUseCase{
    CreateCustomerFunc: func(...) (*customer.Customer, error) {
        return fakeCustomer(), nil
    },
    SetCustomerPhoneFunc: func(...) error { return nil },
    AddTagToCustomerFunc: func(...) error { return nil },
    GetCustomerFunc: func(...) (*customer.Customer, error) {
        return fakeCustomer(), nil  // Reload after updates
    },
}
```

**Wildcard Routes** (Timezone):

```go
// Route: GET /timezones/*name (not :name)
router.GET("/timezones/*name", handler.GetTimezoneByName)

// Test URL
w := smoke.MakeRequest(t, router, "GET", "/timezones/Europe/Kyiv", nil)
```

### 5. Focus on Status Codes

**What matters**:

- `200 OK` for successful GET
- `201 Created` for successful POST
- `404 Not Found` for missing resource
- `400 Bad Request` for validation errors
- `500 Internal Server Error` for server errors
- Response has `{"status":"success"}` or `{"status":"error"}`

**What doesn't matter** (for smoke tests):

- Exact field values in response
- Database state changes
- Event publishing
- Complex business rules

---

## Running Tests

```bash
# All smoke tests (196 tests)
go test ./test/smoke/... -v

# All smoke tests (using Makefile)
make test-smoke

# Specific context
go test ./test/smoke/contexts/identity/... -v
go test ./test/smoke/contexts/customer-mgmt/... -v
go test ./test/smoke/contexts/order-mgmt/... -v
go test ./test/smoke/contexts/shared/... -v
go test ./test/smoke/contexts/fiscal/... -v

# Specific handler
go test ./test/smoke/contexts/customer-mgmt/customer -v

# With coverage
go test ./test/smoke/... -cover
```

**Expected output**:

```
ok  github.com/basilex/promenade/test/smoke/contexts/identity/user      (cached)
ok  github.com/basilex/promenade/test/smoke/contexts/identity/contact   (cached)
...
ok  github.com/basilex/promenade/test/smoke/contexts/shared/timezone    (cached)

 All 196 tests PASSED
```

---

## Final Statistics

| Context             | Handlers | Tests   | Status |
| ------------------- | -------- | ------- | ------ |
| Identity            | 5        | 43      |        |
| Customer Management | 4        | 33      |        |
| Order Management    | 1        | 10      |        |
| Billing             | 2        | 16      |        |
| Shared              | 4        | 36      |        |
| Warehouse           | 1        | 23      |        |
| UI                  | 1        | 8       |        |
| Fiscal              | 2        | 15      |        |
| **Total**           | **20**   | **196** |        |

**Test Distribution**:

- Simple handlers (6-8 tests): 14 handlers
- Extended handlers (9-10 tests): 3 handlers
- Complex handlers (12 tests): 1 handler (Customer)

**Effort**:

- ~15 minutes per simple handler
- ~4 hours total development time
- 100% pass rate achieved

---

## Examples

**Reference Implementations**:

- `test/smoke/contexts/order-mgmt/order/handler_test.go` - Standard handler pattern
- `test/smoke/contexts/identity/user/handler_test.go` - Authentication handlers with bcrypt
- `test/smoke/contexts/customer-mgmt/customer/handler_test.go` - Complex handler (23 methods, 12 tests)
- `test/smoke/contexts/billing/invoice/handler_test.go` - Invoice handler with 18-method mock
- `test/smoke/contexts/billing/payment/handler_test.go` - Payment handler pattern
- `test/smoke/contexts/shared/timezone/handler_test.go` - Wildcard routes + unique patterns
- `test/smoke/testutils.go` - Helper functions and utilities

**Mock Examples**:

- Simple mock: Contact (6 methods)
- Complex mock: Customer (23 methods with full implementation)

---

## When to Write Smoke Tests

**Always**:

- When creating new handler
- When adding new endpoint to existing handler
- When changing response format
- When modifying error handling

**Never skip**:

- Even for "simple" handlers
- Tests are fast to write with utilities (~15 min/handler)
- Catch silly mistakes (wrong HTTP code, typo in error code)
- Prevent regressions in HTTP layer

---

**Last Updated**: January 16, 2026  
**Status**: COMPLETE - 20/20 handlers, 196 tests, 100% pass rate
