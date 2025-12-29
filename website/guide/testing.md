# Three-Tier Testing Strategy

Promenade uses a **professional three-tier testing strategy** with clear separation of concerns and comprehensive coverage.

## Test Organization

### 1. Unit Tests (In-Place)

**Location**: Same directory as production code  
**Purpose**: Test individual components in isolation  
**Speed**: Fast (~5 seconds)

```
internal/contexts/identity/contact/
 entity.go
 entity_test.go         ← Unit tests here
 usecase.go
 usecase_test.go        ← Unit tests here
```

**Example**:
```go
// internal/contexts/identity/contact/entity_test.go
func TestContact_NewEmailContact(t *testing.T) {
    userID := uuidv7.New()
    contact, err := NewEmailContact(userID, "test@example.com", "Work")
    
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, contact.ID)
    assert.Equal(t, userID, contact.UserID)
    assert.Equal(t, ContactTypeEmail, contact.Type)
}
```

### 2. Smoke Tests (Mock-Based)

**Location**: `test/smoke/contexts/` (mirror path)  
**Purpose**: Fast HTTP handler validation with mocks  
**Speed**: Very fast (~0.3 seconds)

```
test/smoke/contexts/
 identity/
    contact/handler_test.go   ← Mock-based handler tests
```

**Example**:
```go
// test/smoke/contexts/identity/contact/handler_test.go
type MockContactUseCase struct {
    mock.Mock
}

func TestHandler_CreateEmailContact(t *testing.T) {
    gin.SetMode(gin.TestMode)
    mockUC := new(MockContactUseCase)
    handler := NewContactHandler(mockUC)
    
    // Setup expectations
    mockUC.On("CreateEmailContact", mock.Anything, mock.Anything, "test@example.com", "Work", false).
        Return(&contact.Contact{ID: uuidv7.New()}, nil)
    
    // Make HTTP request
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(`{...}`))
    
    handler.Create(c)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    mockUC.AssertExpectations(t)
}
```

### 3. Integration Tests (Real DB)

**Location**: `test/integration/contexts/` (mirror path)  
**Purpose**: Full E2E testing with real database  
**Speed**: Slower (~14 seconds)

```
test/integration/contexts/
 identity/
    contact/repository_test.go   ← Real DB tests
```

**Example**:
```go
// test/integration/contexts/identity/contact/repository_test.go
func TestContactRepository_Create(t *testing.T) {
    db := integration.SetupTestDBWithCleanTables(t)
    repo := postgres.NewContactRepository(db.DB)
    ctx := context.Background()
    
    userID := uuidv7.New()
    contact, _ := contact.NewEmailContact(userID, "test@example.com", "Work")
    
    err := repo.Create(ctx, contact)
    require.NoError(t, err)
    
    // Verify persistence
    retrieved, err := repo.GetByID(ctx, contact.ID)
    require.NoError(t, err)
    assert.Equal(t, contact.ID, retrieved.ID)
}
```

## Running Tests

### Quick Commands

```bash
# All tests (240+ tests, ~40s with race detector)
make test

# By type
make test-unit              # Unit tests only (~5s)
make test-smoke             # Smoke tests (~0.3s)
make test-integration       # Integration tests (~14s)

# Context-specific
go test ./internal/contexts/identity/... -v
go test ./test/smoke/contexts/identity/... -v
go test ./test/integration/contexts/identity/... -v

# With coverage
make test-coverage          # HTML coverage report
```

### Test Comparison

| Type            | Location                      | Database   | Speed     | Run When     |
| --------------- | ----------------------------- | ---------- | --------- | ------------ |
| **Unit**        | In-place (`*_test.go`)        | No (mocks) | Fast (~5s)| Every save   |
| **Smoke**       | `/test/smoke/contexts/`       | No (mocks) | Very fast (~0.3s) | Every commit |
| **Integration** | `/test/integration/contexts/` | Real DB    | Slow (~14s) | Before merge |

## Test Coverage

### Identity Context

- **User**: 32 unit + 11 smoke + 14 integration = **57 tests**
- **Contact**: 35 unit + 7 smoke + 9 integration = **51 tests**
- **Profile**: 35 unit + 8 smoke + 17 integration = **60 tests**

### Shared Context

- **Country**: 6 unit + 5 smoke + 6 integration = **17 tests**
- **Currency**: 6 unit + 5 smoke + 6 integration = **17 tests**
- **Language**: 6 unit + 5 smoke + 6 integration = **17 tests**
- **Timezone**: 6 unit + 5 smoke + 6 integration = **17 tests**

### Package Tests

- **Event Bus**: 67 tests (100% coverage)
- **JWT**: 18 tests (87% coverage)
- **Logger**: 15 tests (95% coverage)
- **UUID v7**: 10 tests (100% coverage)

**Total**: 240+ tests, 90%+ average coverage

## Writing Tests

### 1. Write Unit Test

```go
// internal/contexts/identity/contact/entity_test.go
package contact

func TestContact_SetAsPrimary(t *testing.T) {
    contact := &Contact{IsPrimary: false}
    contact.SetAsPrimary()
    assert.True(t, contact.IsPrimary)
}
```

### 2. Write Smoke Test

```go
// test/smoke/contexts/identity/contact/handler_test.go
package contact_test

func TestHandler_GetByID_Success(t *testing.T) {
    mockUC := new(MockContactUseCase)
    handler := NewContactHandler(mockUC)
    
    contact := &contact.Contact{ID: uuidv7.New()}
    mockUC.On("GetContact", mock.Anything, contact.ID).Return(contact, nil)
    
    // Test HTTP handler
}
```

### 3. Write Integration Test

```go
// test/integration/contexts/identity/contact/repository_test.go
package contact_test

func TestContactRepository_GetByUserID(t *testing.T) {
    db := integration.SetupTestDBWithCleanTables(t)
    repo := postgres.NewContactRepository(db.DB)
    
    // Create test data
    // Execute query
    // Verify results
}
```

## Test Helpers

### Database Setup

```go
// test/integration/testutils.go
func SetupTestDBWithCleanTables(t *testing.T) *TestDB {
    // 1. Start test database (PostgreSQL on 5433)
    // 2. Run migrations
    // 3. Clean all tables
    // 4. Return connection with cleanup callback
}
```

### Mock Creation

```go
// test/smoke/contexts/identity/contact/mocks_test.go
type MockContactUseCase struct {
    mock.Mock
}

func (m *MockContactUseCase) CreateEmailContact(...) (*Contact, error) {
    args := m.Called(...)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Contact), args.Error(1)
}
```

## Best Practices

### DO

- ✅ Write unit tests for business logic
- ✅ Write smoke tests for HTTP handlers
- ✅ Write integration tests for repositories
- ✅ Use table-driven tests for multiple cases
- ✅ Use `require.NoError()` for setup code
- ✅ Use `assert.Equal()` for assertions
- ✅ Run tests before committing

### DON'T

- ❌ Don't ignore test errors with `_`
- ❌ Don't share state between tests
- ❌ Don't use real database in unit tests
- ❌ Don't test implementation details
- ❌ Don't skip cleanup in integration tests

## Continuous Testing

### Watch Mode

```bash
# Watch for changes and run tests
go test ./internal/contexts/identity/contact -v -watch
```

### Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
make test-unit || exit 1
```

### CI/CD

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make test
```

## Debugging Tests

### Verbose Output

```bash
go test ./... -v                    # Verbose mode
go test ./... -v -run TestSpecific  # Run specific test
```

### Race Detector

```bash
go test ./... -race   # Detect race conditions
```

### Coverage Analysis

```bash
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Next Steps

- [Getting Started](/guide/getting-started) - Setup test environment
- [Architecture](/guide/architecture) - Understand DDD testing
- [Identity Context](/contexts/identity) - Example tests
- [GitHub Docs](https://github.com/basilex/promenade/blob/dev/test/README.md) - Complete testing guide
