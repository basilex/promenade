---
title: "Testing Infrastructure"
description: "400+ tests with manual mocks pattern"
weight: 5
---

## Comprehensive Testing

Promenade has **400+ tests** covering all layers with a **manual mocks** approach.

### Test Coverage

**Core Tests (275 tests):**

- Domain entities: 39 tests
- Use cases: 236 tests (auth, RBAC, reference data)

**Module Tests:**

- Posts: 33 tests (83.3% coverage)
- Profiles: 21 tests (80.4% coverage)
- Analytics: 11 tests
- Billing: 375 tests (100% coverage)

**Utilities:** 51 tests, 89.5% average coverage

### Test Types

```bash
# All tests (~20 seconds)
make test

# Core tests only
make test-core

# Module tests
make test-module-posts
make test-module-billing

# Coverage report
make test-coverage
```

### Manual Mocks Pattern

**No mockgen** - simple, explicit mocks:

```go
// Mock repository inline
type mockUserRepo struct {
    users map[string]*User
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    return nil, ErrUserNotFound
}

// Use in tests
func TestRegisterUser(t *testing.T) {
    repo := &mockUserRepo{users: make(map[string]*User)}
    usecase := NewUserUseCase(repo)

    user, err := usecase.Register(ctx, "test@example.com", "password")
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### Integration Tests

Tests with **real PostgreSQL**:

```bash
# Start test database
make test-db-start

# Run integration tests
make test-integration

# Stop test database
make test-db-stop
```

### Test Structure

Tests live **alongside code**:

```
internal/usecase/
├── auth_usecase.go
├── auth_usecase_test.go      # Unit tests here
├── role_usecase.go
└── role_usecase_test.go
```

### Benefits

✅ **Fast** - Unit tests run in ~5 seconds  
✅ **Simple** - No mock generation tools needed  
✅ **Comprehensive** - 100% business logic coverage  
✅ **Real DB** - Integration tests use PostgreSQL

[Testing Guide →](/promenade/docs/TESTING_GUIDE)
