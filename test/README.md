# Testing Guide

Unit tests for Promenade project - tests are located alongside the code they test.

---

## Test Structure

Tests are organized **per component** - each module and core component has its own tests in the same directory:

### Core Tests

```
internal/
├── domain/
│   └── entity/              # Core entity tests (39 tests)
│       ├── user.go
│       ├── user_test.go     # ← Tests here
│       ├── session.go
│       ├── session_test.go  # ← Tests here
│       ├── role.go
│       ├── role_test.go
│       ├── permission.go
│       └── permission_test.go
└── usecase/                 # Core usecase tests (236 tests)
    ├── auth_usecase.go
    ├── auth_usecase_test.go
    ├── role_usecase.go
    └── role_usecase_test.go
```

### IModule Tests

```
internal/modules/
├── posts/
│   └── domain/
│       └── entity/
│           ├── post.go
│           ├── post_test.go      # ← 33 tests, 83.3% coverage
│           ├── comment.go
│           └── comment_test.go
├── profiles/
│   └── entity/
│       ├── user_profile.go
│       ├── user_profile_test.go  # ← 21 tests, 80.4% coverage
│       ├── user_contact.go
│       └── user_contact_test.go
└── analytics/
    └── domain/
        └── entity/
            ├── metric.go
            ├── metric_test.go    # ← 11 tests
            └── metric_aggregate_test.go
```

---

## Running Tests

### All Tests

```bash
make test                    # Run all tests (400+ tests, ~20s)
```

### Core Tests

```bash
make test-core               # Run core tests (275 tests: 39 entity + 236 usecase)
```

### IModule Tests

```bash
make test-modules            # Run all module tests
make test-module-posts       # Posts module (33 tests, 83.3% coverage)
make test-module-profiles    # Profiles module (21 tests, 80.4% coverage)
make test-module-analytics   # Analytics module (11 tests)
```

### Coverage

```bash
make test-coverage           # Generate coverage report (HTML)
# Opens coverage.html in browser
```

---

## Test Categories

### 1. **Core Entity Tests** (`internal/domain/entity/*_test.go`)

**39 tests** covering core domain entities:

- **Country**: Validation, currency relationships, ISO codes
- **Currency**: Validation, country relationships, symbols
- **Language**: ISO codes, native names, validation
- **Timezone**: IANA database, UTC offsets, validation
- **Permission**: Resource-action format, wildcard validation
- **Role**: RBAC roles, system roles, validation
- **User**: Password hashing/checking, status management, email validation
- **Session**: Token generation, expiration, refresh logic
- **Purge**: Retention policies, entity tracking

**Example**:

```bash
go test -v ./internal/domain/entity/
```

**Tests**:

- `TestUser_HashPassword` - Password hashing
- `TestUser_CheckPassword` - Password verification
- `TestSession_IsExpired` - Session expiration
- `TestSession_Validate` - Session validation
- `TestRole_Validate` - Role validation
- `TestPermission_Constants` - Permission constants

---

### 2. **Posts IModule Tests** (`internal/modules/posts/domain/entity/*_test.go`)

**33 tests, 83.3% coverage** - Test posts module entities:

- **PostStatus**: Lifecycle management, valid transitions
- **UserPost**: Creation, validation, publish/unpublish, archive, schedule, counters, soft delete, featured image, slug generation
- **Comment**: Threading, validation, soft delete

**Example**:

```bash
go test -v ./internal/modules/posts/domain/entity/
```

**Tests**:

- `TestPostStatus_IsValid` - Post status validation
- `TestUserPost_Creation` - Post creation
- `TestUserPost_Publish` - Publish operations
- `TestUserPost_Archive` - Archive operations
- `TestUserPost_Slug` - Slug generation
- `TestComment_Validate` - Comment validation
- `TestComment_IsTopLevel` - Comment threading

---

### 3. **Profiles IModule Tests** (`internal/modules/profiles/entity/*_test.go`)

**21 tests, 80.4% coverage** - Test profiles module entities:

- **UserProfile**: Profile fields, privacy settings, gender validation, social links, ban/verify operations
- **UserContact**: Contact types, verification flags, availability

**Example**:

```bash
go test -v ./internal/modules/profiles/entity/
```

**Tests**:

- `TestUserProfile_Fields` - Profile field validation
- `TestGender_IsValid` - Gender enum validation
- `TestUserProfile_Privacy` - Privacy flags
- `TestUserProfile_Ban` - Ban/unban operations
- `TestContactType_IsValid` - Contact type validation
- `TestUserContact_Creation` - Contact creation
- `TestUserContact_Availability` - Availability checks

---

### 4. **Analytics IModule Tests** (`internal/modules/analytics/*_test.go`)

**11 tests** - Test analytics module:

- **Metric**: Validation, types, scopes
- **MetricAggregate**: Calculations
- **UseCase**: Collect, get, list operations

**Example**:

```bash
go test -v ./internal/modules/analytics/...
```

**Tests**:

- `TestMetric_Validation` - Metric validation
- `TestMetricAggregate` - Aggregate calculations
- `TestAnalyticsUseCase_CollectMetric` - Collect metrics
- `TestAnalyticsUseCase_GetMetric` - Get metric by ID
- `TestAnalyticsUseCase_ListMetrics` - List metrics with filters

---

## Test Conventions

### Naming

- **Test files**: `{entity}_test.go` (e.g., `user_test.go`)
- **Test functions**: `Test{Entity}_{Method}` (e.g., `TestUser_HashPassword`)
- **Subtests**: Use `t.Run(name, func(t))` for table-driven tests

### Structure

```go
func TestEntity_Method(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        {"valid case", input1, want1, false},
        {"invalid case", input2, want2, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Method(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Dependencies

- **testify/assert**: `github.com/stretchr/testify/assert`
- **UUID v7**: `github.com/basilex/promenade/pkg/uuidv7`

---

## Test Coverage

**Current Status**: All tests passing

### Core Domain

| Entity     | Tests            | Coverage               |
| ---------- | ---------------- | ---------------------- |
| User       | 3 test functions | Password, status       |
| Session    | 2 test functions | Expiration, validation |
| Role       | 1 test function  | Validation             |
| Permission | 2 test functions | Constants, format      |

### Modules

| IModule  | Entity      | Tests       | Coverage                         |
| -------- | ----------- | ----------- | -------------------------------- |
| Posts    | UserPost    | 2 functions | Status, creation                 |
| Posts    | Comment     | 3 functions | Validation, threading, factory   |
| Profiles | UserProfile | 3 functions | Fields, gender, privacy          |
| Profiles | UserContact | 3 functions | Type validation, creation, flags |

---

## Test Helpers

Located in `test/helpers/`:

- `database.go` - Database test utilities (for future integration tests)
- `fixtures.go` - Test data fixtures (for future integration tests)

**Note**: Current tests are unit tests and don't require database helpers.

---

## 📖 Writing New Tests

### 1. Create Test File

Place test file next to the entity:

```bash
# Example: internal/domain/entity/country.go
touch internal/domain/entity/country_test.go
```

### 2. Write Test

```go
package entity

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCountry_Validate(t *testing.T) {
    tests := []struct {
        name    string
        country Country
        wantErr bool
    }{
        {"valid country", Country{Code: "US", Name: "United States"}, false},
        {"empty code", Country{Code: "", Name: "Test"}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.country.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 3. Run Test

```bash
go test -v ./internal/domain/entity/
```

---

## 🔄 Continuous Testing

### Watch Mode

```bash
make test-watch
# Automatically re-runs tests on file changes
```

### Pre-commit

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
make test-quick || exit 1
```

---

## 🚫 What We Don't Test (Yet)

- **Integration tests**: Database operations, external services
- **E2E tests**: Full application workflows
- **Handler tests**: HTTP handlers (adapter layer)
- **Repository tests**: Database queries
- **Use case tests**: Business logic orchestration

**Future**: Add integration and E2E tests as the application grows.

---

## 📚 Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Effective Go - Testing](https://go.dev/doc/effective_go#testing)

---

**Test Status**: All tests passing | Unit tests only | Fast execution
