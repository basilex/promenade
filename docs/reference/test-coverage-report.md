# Test Coverage Report

**Promenade Platform** - Comprehensive test coverage across all components with **three-tier testing strategy**.

---

## Overall Statistics

| Metric                  | Value        |
| ----------------------- | ------------ |
| **Total Tests**         | 240+         |
| **Average Coverage**    | 90%+         |
| **Test Execution Time** | ~40s (all)   |
| **Test Organization**   | 3-tier       |
| **Packages Tested**     | 40+          |

---

## Test Organization

### Three-Tier Strategy

1. **Unit Tests** (in-place) - Fast feedback, isolated components (~5s)
2. **Smoke Tests** (mirror path) - Mock-based handlers, no DB (~0.4s)
3. **Integration Tests** (mirror path) - Real database, full E2E (~14s)

**See**: [Testing Patterns Guide](../guides/testing-patterns.md) for complete documentation

---

## Package Test Coverage

### Core Packages

| Package         | Tests | Coverage | Duration | Type |
| --------------- | ----- | -------- | -------- | ---- |
| **pkg/jwt**     | 18    | 87%      | cached   | Unit |
| **pkg/bus**     | 67    | 100%     | ~2s      | Unit |
| **pkg/logger**  | 15    | 95%      | cached   | Unit |
| **pkg/uuidv7**  | 10    | 100%     | cached   | Unit |
| **pkg/response**| 13    | 100%     | cached   | Unit |
| **pkg/saga**    | 28    | 100%     | cached   | Unit |
| **pkg/migration** | 8   | 90%      | cached   | Unit |
| **pkg/valueobject** | 45 | 95%     | cached   | Unit |
| **pkg/aggregate** | 5   | 100%     | cached   | Unit |
| **pkg/jsonb**   | 8     | 100%     | cached   | Unit |

**Total**: 217 tests across 10 packages

---

### Context Test Coverage

#### Identity Context

**Unit Tests** (in-place):
- User Entity: 15 tests, 85% coverage
- Contact Entity: 35 tests, 96% coverage
- Profile Entity: 35 tests, 96% coverage
- User UseCase: 52 tests, 70% coverage
- Contact UseCase: 52 tests, 70% coverage
- Profile UseCase: 52 tests, 70% coverage

**Smoke Tests** (mock-based):
- User Handler: 15 tests
- Contact Handler: 7 tests
- Profile Handler: 8 tests

**Integration Tests** (real DB):
- User Repository: 32 tests
- Contact Repository: 9 tests
- Profile Repository: 17 subtests

**Total**: 394 tests

---

#### Shared Context

**Unit Tests**:
- Country Entity: Tests present
- Currency Entity: Tests present
- Language Entity: Tests present
- Timezone Entity: Tests present

**Smoke Tests**:
- Country Handler: 5 tests
- Currency Handler: 5 tests
- Language Handler: 5 tests
- Timezone Handler: 5 tests

**Integration Tests**:
- Country Repository: 6 tests
- Currency Repository: 6 tests
- Language Repository: 6 tests
- Timezone Repository: 6 tests

**Total**: 48 tests

---

#### Customer Management Context

**Smoke Tests**:
- Customer Handler: 18 tests

**Integration Tests**:
- Customer Repository: 14 tests

**Total**: 32 tests

---

## Running Tests

### Quick Commands

```bash
# All tests (360+ tests, ~60s with race detector)
make test

# By type
make test-unit              # Unit tests only (~5s)
make test-smoke             # Smoke tests (~0.4s)
make test-integration       # Integration tests (~14s)

# With coverage
make test-coverage          # HTML coverage report
```

### Context-Specific

```bash
# Identity Context
go test ./internal/contexts/identity/... -v
go test ./test/smoke/contexts/identity/... -v
go test ./test/integration/contexts/identity/... -v

# Shared Context
go test ./internal/contexts/shared/... -v
go test ./test/smoke/contexts/shared/... -v
go test ./test/integration/contexts/shared/... -v

# Package Tests
go test ./pkg/bus/... -v
go test ./pkg/jwt/... -v
```

---

## Test Breakdown by Type

### Unit Tests (~5s)

**Location**: In-place (`*_test.go` next to production code)

**Components Tested**:
- Entities (validation, factory methods)
- Use cases (business logic)
- Value objects (immutability, validation)
- Package utilities

**Total**: ~150 tests

---

### Smoke Tests (~0.4s)

**Location**: `test/smoke/contexts/` (mirror path)

**Components Tested**:
- HTTP handlers with mock use cases
- Request/response validation
- Error handling

**Total**: ~50 tests

---

### Integration Tests (~14s)

**Location**: `test/integration/contexts/` (mirror path)

**Components Tested**:
- Repositories with real database
- End-to-end workflows
- Database constraints

**Total**: ~80 tests

---

## Coverage Goals

### Current Targets

| Component         | Target | Current | Status |
| ----------------- | ------ | ------- | ------ |
| Core Packages     | 90%    | 95%     | ✅      |
| Domain Entities   | 85%    | 90%     | ✅      |
| Use Cases         | 80%    | 70%     | ⚠️      |
| HTTP Handlers     | 75%    | 80%     | ✅      |
| Repositories      | 85%    | 90%     | ✅      |

### Improvement Areas

- **Use Cases**: Need more edge case tests (70% → 80%)
- **Error Paths**: Some error scenarios not covered
- **Concurrent Access**: Limited concurrency tests

---

## Test Quality Metrics

### Test Characteristics

**✅ Good Coverage**:
- All critical paths tested
- Business rules validated
- Error handling verified
- Database constraints tested

**✅ Fast Execution**:
- Unit tests: ~5s (fast feedback)
- Smoke tests: ~0.4s (mock-based)
- Integration tests: ~14s (acceptable for E2E)

**✅ Well Organized**:
- Three-tier structure
- Mirror path for smoke/integration
- Clear naming conventions

---

## Related Documentation

- [Testing Patterns Guide](../guides/testing-patterns.md) - Complete testing documentation
- [Testing Quick Reference](../guides/testing-quick-reference.md) - Cheat sheet
- [Test README](../../test/README.md) - Testing infrastructure

---

**Last Updated**: December 29, 2025  
**Status**: 240+ tests, 90%+ average coverage  
**Maintainer**: Promenade Team
