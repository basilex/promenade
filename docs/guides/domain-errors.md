# Domain Errors in Promenade

**Professional error handling** with type-safe domain errors and clear separation between business rules and system failures.

---

## Table of Contents

1. [Introduction](#introduction)
2. [Gold Standard Pattern](#gold-standard-pattern)
3. [UseCase Implementation](#usecase-implementation)
4. [Handler Error Mapping](#handler-error-mapping)
5. [Testing Patterns](#testing-patterns)
6. [Common Mistakes](#common-mistakes)
7. [Migration Guide](#migration-guide)
8. [Code Review Checklist](#code-review-checklist)

---

## Introduction

### What Are Domain Errors?

**Domain errors** are predefined error constants that represent specific business rule violations in your domain model. Unlike generic error messages, domain errors are:

- **Type-safe**: Can be checked with `errors.Is()` instead of string comparison
- **Self-documenting**: Error constant names clearly describe the failure
- **Testable**: Easy to assert in tests without fragile string matching
- **Consistent**: Same error used everywhere in the codebase
- **Secure**: Prevent information leakage by mapping to safe user messages

### Why They Matter

**For Developers**:
- Type safety prevents bugs (compile-time checks)
- Clear intent (error constant name explains why it failed)
- Easy refactoring (rename constant, find all usages)
- Better IDE support (autocomplete, go-to-definition)

**For Users**:
- Consistent error messages across the application
- Proper HTTP status codes (404 for not found, 400 for validation)
- No sensitive information leaked in error messages
- Better user experience with actionable error messages

**For Testing**:
- No fragile string comparisons in tests
- Clear assertions: `assert.True(t, errors.Is(err, ErrNotFound))`
- Easy to test all error paths
- Refactoring-safe tests

### Relationship to Clean Architecture

Domain errors live in the **domain layer** (entity/aggregate packages) and flow outward:

```
Domain Layer (entity.go)
  ↓ Returns domain errors
UseCase Layer (usecase.go)
  ↓ Returns domain errors (or wraps system errors)
Adapter Layer (handler.go)
  ↓ Maps domain errors to HTTP responses
```

**Critical Rule**: Domain errors defined once in `errors.go`, used everywhere.

### When to Use vs fmt.Errorf

**Use Domain Errors** for:
- Business rule violations (`ErrInsufficientStock`, `ErrAccountLocked`)
- Not found scenarios (`ErrCustomerNotFound`, `ErrOrderNotFound`)
- Validation failures (`ErrInvalidEmail`, `ErrPriceNegative`)
- Authorization failures (`ErrUnauthorized`, `ErrAccessDenied`)

**Use fmt.Errorf** for:
- System failures (database connection lost, network timeout)
- External service errors (payment gateway unavailable)
- Unexpected internal errors (parsing failed, file not readable)
- Wrapping errors for context (`fmt.Errorf("failed to create order: %w", err)`)

**Rule of Thumb**: If users need to understand and act on the error → domain error. If it's a system/infrastructure failure → fmt.Errorf.

---

## Gold Standard Pattern

The **Gold Standard** is the warehouse location aggregate - perfect example of domain error implementation.

### errors.go Structure

Every aggregate should have an `errors.go` file with three categories:

**File**: `internal/contexts/warehouse/location/errors.go`

```go
package location

import "errors"

// Repository Errors - Data access failures
var (
	// ErrLocationNotFound is returned when location doesn't exist
	ErrLocationNotFound = errors.New("location not found")
	
	// ErrLocationUnauthorized is returned when user lacks permission
	ErrLocationUnauthorized = errors.New("location access unauthorized")
)

// Business Logic Errors - Domain rule violations
var (
	// ErrLocationCodeExists is returned when code already exists
	ErrLocationCodeExists = errors.New("location code already exists")
	
	// ErrLocationAlreadyDeleted is returned when operating on deleted location
	ErrLocationAlreadyDeleted = errors.New("location already deleted")
	
	// ErrLocationHasChildren is returned when deleting location with children
	ErrLocationHasChildren = errors.New("cannot delete location with children")
	
	// ErrLocationInvalidType is returned when location type is invalid
	ErrLocationInvalidType = errors.New("invalid location type")
	
	// ErrLocationParentNotFound is returned when parent location doesn't exist
	ErrLocationParentNotFound = errors.New("parent location not found")
	
	// ErrLocationParentDeleted is returned when parent location is deleted
	ErrLocationParentDeleted = errors.New("parent location is deleted")
	
	// ErrLocationInvalidName is returned when name is empty or too long
	ErrLocationInvalidName = errors.New("location name is invalid")
	
	// ErrLocationInvalidCode is returned when code is empty or too long
	ErrLocationInvalidCode = errors.New("location code is invalid")
	
	// ErrLocationInvalidCapacity is returned when capacity is negative
	ErrLocationInvalidCapacity = errors.New("location capacity cannot be negative")
)

// Technical Operation Errors - Operation wrappers
var (
	// ErrLocationCreateFailed is returned when creation fails
	ErrLocationCreateFailed = errors.New("failed to create location")
	
	// ErrLocationUpdateFailed is returned when update fails
	ErrLocationUpdateFailed = errors.New("failed to update location")
	
	// ErrLocationDeleteFailed is returned when deletion fails
	ErrLocationDeleteFailed = errors.New("failed to delete location")
)
```

### Naming Conventions

**Error Constant Names**:
```
Err + {Aggregate} + {Condition}
```

**Examples**:
- `ErrLocationNotFound` (not `ErrNotFound` - aggregate name required)
- `ErrInventorySKUExists` (clear what exists)
- `ErrCustomerAlreadyExists` (clear business rule)
- `ErrDealInvalidStage` (clear validation issue)

**Error Message Text**:
- Lowercase (except proper nouns)
- Descriptive and concise
- No punctuation at end
- Present tense

**Examples**:
```go
// ✅ Good
ErrLocationNotFound = errors.New("location not found")
ErrInsufficientStock = errors.New("insufficient stock available")

// ❌ Bad
ErrLocationNotFound = errors.New("Location not found.")  // Capitalized, has period
ErrStockLow = errors.New("stock")  // Not descriptive enough
```

### Code Organization

**Group related errors together**:

```go
// Repository Errors (data access)
var (
	ErrNotFound = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
)

// Validation Errors (input validation)
var (
	ErrInvalidEmail = errors.New("invalid email format")
	ErrInvalidPhone = errors.New("invalid phone format")
	ErrInvalidAmount = errors.New("invalid amount")
)

// Business Logic Errors (domain rules)
var (
	ErrAlreadyExists = errors.New("already exists")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrAccountLocked = errors.New("account locked")
)
```

**One error per line** with clear comment:
```go
// ErrLocationNotFound is returned when location doesn't exist
ErrLocationNotFound = errors.New("location not found")
```

---

## UseCase Implementation

### Domain Errors vs System Errors

**UseCase Layer Rules**:
1. **Return domain errors** for business rule violations
2. **Wrap system errors** with context using `fmt.Errorf(..., %w, err)`
3. **Never use inline errors.New()** - all errors must be constants

### Pattern 1: Return Domain Errors

```go
func (uc *useCase) CreateLocation(ctx context.Context, code, name string, locationType LocationType) (*Location, error) {
	// Check for duplicate code (business rule)
	existing, _ := uc.repo.GetLocationByCode(ctx, code)
	if existing != nil {
		return nil, ErrLocationCodeExists  // ✅ Domain constant
	}
	
	// Validate parent if provided (business rule)
	if parentID != nil {
		parent, err := uc.repo.GetLocation(ctx, *parentID)
		if err != nil {
			return nil, ErrParentLocationNotFound  // ✅ Domain constant
		}
		if parent.DeletedAt != nil {
			return nil, ErrParentLocationDeleted  // ✅ Domain constant
		}
	}
	
	// Create location (technical operation)
	loc := NewLocation(code, name, locationType)
	if err := uc.repo.Create(ctx, loc); err != nil {
		return nil, ErrLocationCreateFailed  // ✅ Technical wrapper
	}
	
	return loc, nil
}
```

### Pattern 2: Check Entity Validation

```go
func (uc *useCase) UpdateLocation(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	// Get existing location
	loc, err := uc.repo.GetLocation(ctx, id)
	if err != nil {
		return ErrLocationNotFound  // ✅ Domain constant
	}
	
	// Check if deleted (business rule)
	if loc.DeletedAt != nil {
		return ErrLocationAlreadyDeleted  // ✅ Domain constant
	}
	
	// Apply updates
	if name, ok := updates["name"].(string); ok {
		if err := loc.SetName(name); err != nil {
			return ErrLocationInvalidName  // ✅ Domain constant
		}
	}
	
	// Persist changes
	if err := uc.repo.Update(ctx, loc); err != nil {
		return ErrLocationUpdateFailed  // ✅ Technical wrapper
	}
	
	return nil
}
```

### Pattern 3: Wrap System Errors

```go
func (uc *useCase) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	// Get location
	loc, err := uc.repo.GetLocation(ctx, id)
	if err != nil {
		return ErrLocationNotFound  // ✅ Domain constant
	}
	
	// Check business rules
	if loc.DeletedAt != nil {
		return ErrLocationAlreadyDeleted  // ✅ Domain constant
	}
	
	// Check for children (business rule requiring DB query)
	hasChildren, err := uc.repo.HasChildren(ctx, id)
	if err != nil {
		// ✅ Wrap system error with context
		return fmt.Errorf("failed to check children: %w", err)
	}
	if hasChildren {
		return ErrLocationHasChildren  // ✅ Domain constant
	}
	
	// Perform deletion
	if err := uc.repo.Delete(ctx, id); err != nil {
		return ErrLocationDeleteFailed  // ✅ Technical wrapper
	}
	
	return nil
}
```

### What NOT to Do

**❌ Inline errors.New()** (eliminated in Phase 2):
```go
// WRONG - no domain constant
if code == "" {
	return nil, errors.New("code is required")
}
```

**❌ fmt.Errorf() for business rules** (eliminated in Phase 2):
```go
// WRONG - should be domain constant
if existing != nil {
	return nil, fmt.Errorf("location code %s already exists", code)
}
```

**❌ String building in errors** (security risk):
```go
// WRONG - information leakage
return nil, fmt.Errorf("user %s not found", email)
```

**✅ Correct approach**:
```go
// Always return domain constants
if code == "" {
	return nil, ErrLocationInvalidCode
}

if existing != nil {
	return nil, ErrLocationCodeExists
}

// Context in logs, not in errors
logger.FromContext(ctx).Warn("duplicate code attempt",
	slog.String("code", code),
	slog.String("user_id", userID),
)
return nil, ErrLocationCodeExists
```

---

## Handler Error Mapping

### HTTP Status Code Mapping

**Handlers must discriminate between error types** to return appropriate HTTP status codes:

```go
func (h *LocationHandler) Create(c *gin.Context) {
	var req CreateLocationRequest
	
	// 1. Validation errors - EXPOSE details (user input issues)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())  // 400 - Safe to expose
		return
	}
	
	// 2. Call use case
	loc, err := h.usecase.CreateLocation(c.Request.Context(), req.Code, req.Name, req.Type)
	
	// 3. Domain errors - MAP to user-friendly messages with errors.Is()
	if err != nil {
		if errors.Is(err, location.ErrLocationCodeExists) {
			response.BadRequest(c, "Location code already exists")  // 400
			return
		}
		if errors.Is(err, location.ErrParentLocationNotFound) {
			response.NotFound(c, "Parent location not found")  // 404
			return
		}
		if errors.Is(err, location.ErrParentLocationDeleted) {
			response.BadRequest(c, "Parent location is deleted")  // 400
			return
		}
		if errors.Is(err, location.ErrLocationInvalidType) {
			response.BadRequest(c, "Invalid location type")  // 400
			return
		}
		
		// 4. System errors - HIDE details (security)
		response.InternalError(c, "Failed to create location")  // 500 - Generic
		return
	}
	
	response.Created(c, toLocationResponse(loc))  // 201
}
```

### Status Code Guidelines

| Error Type | HTTP Code | When to Use |
|------------|-----------|-------------|
| **Not Found** | 404 | ErrXxxNotFound errors |
| **Bad Request** | 400 | Validation errors, ErrXxxAlreadyExists, ErrXxxInvalid* |
| **Conflict** | 409 | State conflicts (ErrXxxAlreadyExists when it's a conflict, not validation) |
| **Forbidden** | 403 | ErrXxxUnauthorized, permission issues |
| **Internal Error** | 500 | System errors, unknown errors (fallback) |
| **Unauthorized** | 401 | Authentication failures |

### Security Considerations

**Three-Layer Error Architecture** (from Handler Security Audit):

**Layer 1 - Validation Errors** (EXPOSE details):
```go
// User needs validation feedback
if err := c.ShouldBindJSON(&req); err != nil {
	response.BadRequest(c, err.Error())  // ✅ Safe - user's input
}
```

**Layer 2 - Domain Errors** (MAP to user-friendly):
```go
// Map specific domain errors to clear messages
if errors.Is(err, location.ErrLocationNotFound) {
	response.NotFound(c, "Location not found")  // ✅ Clear, no sensitive data
}
```

**Layer 3 - System Errors** (HIDE details):
```go
// Generic fallback for unknown errors
response.InternalError(c, "Failed to create location")  // ✅ No implementation details
```

**Never Expose**:
- Database error messages (`sql: no rows in result set`)
- Internal paths or stack traces
- User emails or IDs in error messages
- Implementation details

### Complete Handler Example

```go
func (h *LocationHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")  // 400
		return
	}
	
	err = h.usecase.DeleteLocation(c.Request.Context(), id)
	if err != nil {
		// Map domain errors to HTTP codes
		if errors.Is(err, location.ErrLocationNotFound) {
			response.NotFound(c, "Location not found")  // 404
			return
		}
		if errors.Is(err, location.ErrLocationAlreadyDeleted) {
			response.BadRequest(c, "Location already deleted")  // 400
			return
		}
		if errors.Is(err, location.ErrLocationHasChildren) {
			response.BadRequest(c, "Cannot delete location with children")  // 400
			return
		}
		
		// System error fallback
		response.InternalError(c, "Failed to delete location")  // 500
		return
	}
	
	response.NoContent(c)  // 204
}
```

---

## Testing Patterns

### Using errors.Is()

**Always use `errors.Is()` for type-safe error checking** instead of string comparison:

**✅ Correct** (type-safe):
```go
func TestUseCase_CreateLocation_CodeExists(t *testing.T) {
	// ... setup ...
	
	loc, err := uc.CreateLocation(ctx, "EXISTING-CODE", "Test", location.TypeWarehouse)
	
	assert.Error(t, err)
	assert.True(t, errors.Is(err, location.ErrLocationCodeExists))  // ✅ Type-safe
	assert.Nil(t, loc)
}
```

**❌ Wrong** (fragile, eliminated in Phase 2):
```go
// WRONG - string comparison (anti-pattern)
assert.Equal(t, "location code already exists", err.Error())
```

### Test Coverage

**Test all error paths** for each use case:

```go
func TestUseCase_CreateLocation(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		setupFn func(*mockRepo)
		wantErr error
	}{
		{
			name: "success",
			code: "WH-001",
			setupFn: func(m *mockRepo) {
				m.GetByCodeFunc = func(ctx, code string) (*Location, error) {
					return nil, nil  // Not found
				}
			},
			wantErr: nil,
		},
		{
			name: "duplicate code",
			code: "WH-001",
			setupFn: func(m *mockRepo) {
				m.GetByCodeFunc = func(ctx, code string) (*Location, error) {
					return &Location{Code: code}, nil  // Found
				}
			},
			wantErr: location.ErrLocationCodeExists,
		},
		{
			name: "parent not found",
			code: "WH-002",
			setupFn: func(m *mockRepo) {
				m.GetLocationFunc = func(ctx, id uuid.UUID) (*Location, error) {
					return nil, location.ErrLocationNotFound
				}
			},
			wantErr: location.ErrParentLocationNotFound,
		},
		// ... more test cases
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockRepo()
			if tt.setupFn != nil {
				tt.setupFn(mockRepo)
			}
			
			uc := NewUseCase(mockRepo)
			loc, err := uc.CreateLocation(ctx, tt.code, "Test", location.TypeWarehouse)
			
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))  // ✅ Type-safe
				assert.Nil(t, loc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, loc)
			}
		})
	}
}
```

### Integration Test Example

```go
func TestLocationRepository_GetByCode_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	defer integration.TeardownTestDB(t, db)
	
	repo := postgres.NewLocationRepository(db)
	ctx := context.Background()
	
	loc, err := repo.GetLocationByCode(ctx, "NON-EXISTENT")
	
	assert.Error(t, err)
	assert.True(t, errors.Is(err, location.ErrLocationNotFound))  // ✅ Type-safe
	assert.Nil(t, loc)
}
```

### Smoke Test Example

```go
func TestLocationHandler_Create_CodeExists(t *testing.T) {
	router := smoke.SetupRouter()
	
	mockUC := &MockLocationUseCase{
		CreateLocationFunc: func(ctx, code, name string, locType location.Type) (*location.Location, error) {
			return nil, location.ErrLocationCodeExists  // ✅ Return domain error
		},
	}
	
	handler := NewLocationHandler(mockUC)
	router.POST("/locations", handler.Create)
	
	body := map[string]interface{}{
		"code": "EXISTING",
		"name": "Test",
		"type": "warehouse",
	}
	
	w := smoke.MakeRequest(t, router, "POST", "/locations", body)
	
	smoke.AssertErrorResponse(t, w, 400, "Location code already exists")  // ✅ Check mapped message
}
```

---

## Common Mistakes

### 1. String Comparison Anti-Pattern

**❌ Wrong** (eliminated in Phase 2):
```go
// Fragile - breaks if error message changes
if err != nil {
	if err.Error() == "not found" {
		return c.JSON(404, "Not found")
	}
}
```

**✅ Correct**:
```go
// Type-safe - compiler checks
if err != nil {
	if errors.Is(err, location.ErrLocationNotFound) {
		return c.JSON(404, "Location not found")
	}
}
```

### 2. Returning Wrong Error Constant

**❌ Wrong**:
```go
// Using wrong aggregate's error
if customer == nil {
	return location.ErrLocationNotFound  // WRONG aggregate!
}
```

**✅ Correct**:
```go
// Use correct aggregate's error
if customer == nil {
	return customer.ErrCustomerNotFound  // ✅ Right aggregate
}
```

### 3. Information Leakage

**❌ Wrong** (security risk):
```go
// Exposing sensitive data in error
return fmt.Errorf("user %s with email %s not found", userID, email)

// Exposing system details
response.InternalError(c, err.Error())  // Shows "sql: connection lost"
```

**✅ Correct**:
```go
// Generic user-facing message
return customer.ErrCustomerNotFound

// Log details separately
logger.Error("customer lookup failed",
	slog.String("user_id", userID),
	slog.String("email", email),
)

// Generic handler response
response.InternalError(c, "Failed to retrieve customer")
```

### 4. Missing errors.go File

**❌ Wrong**:
```go
// Inline errors in entity.go or usecase.go
if code == "" {
	return errors.New("code required")
}
```

**✅ Correct**:
```go
// Create errors.go with all error constants
// File: internal/contexts/warehouse/location/errors.go
var ErrLocationInvalidCode = errors.New("location code is invalid")

// Use in entity.go or usecase.go
if code == "" {
	return ErrLocationInvalidCode
}
```

### 5. Not Using errors.Is() in Tests

**❌ Wrong**:
```go
// String comparison in tests (fragile)
assert.Equal(t, "location not found", err.Error())
```

**✅ Correct**:
```go
// Type-safe error checking
assert.True(t, errors.Is(err, location.ErrLocationNotFound))
```

---

## Migration Guide

### Step 1: Create errors.go

**For each aggregate**, create `errors.go` file:

```go
// internal/contexts/{context}/{aggregate}/errors.go
package {aggregate}

import "errors"

// Repository Errors
var (
	Err{Aggregate}NotFound = errors.New("{aggregate} not found")
)

// Business Logic Errors
var (
	Err{Aggregate}AlreadyExists = errors.New("{aggregate} already exists")
	// ... add more as needed
)
```

### Step 2: Replace fmt.Errorf in UseCase

**Find all fmt.Errorf calls**:
```bash
grep -n "fmt.Errorf" internal/contexts/{context}/{aggregate}/usecase.go
```

**Replace with domain errors**:
```go
// Before
if existing != nil {
	return nil, fmt.Errorf("location with code %s already exists", code)
}

// After
if existing != nil {
	return nil, ErrLocationCodeExists
}
```

### Step 3: Update Handlers

**Replace string comparisons**:
```go
// Before
if err != nil {
	if err.Error() == "not found" {
		response.NotFound(c, "Location not found")
	}
}

// After
if err != nil {
	if errors.Is(err, location.ErrLocationNotFound) {
		response.NotFound(c, "Location not found")
	}
}
```

### Step 4: Update Tests

**Replace string assertions**:
```go
// Before
assert.Equal(t, "location not found", err.Error())

// After
assert.True(t, errors.Is(err, location.ErrLocationNotFound))
```

### Step 5: Validate

**Run all tests**:
```bash
go test ./... -v
```

**Run linter**:
```bash
golangci-lint run ./...
```

---

## Code Review Checklist

**Before Approving PR, Verify**:

### errors.go File
- [ ] Each aggregate has `errors.go` file
- [ ] Error constants follow naming convention (`Err{Aggregate}{Condition}`)
- [ ] Error messages are lowercase, descriptive, no punctuation
- [ ] Errors grouped by category (Repository, Validation, Business Logic)
- [ ] Each error has clear comment

### UseCase Layer
- [ ] **Zero fmt.Errorf for business rules** (all domain constants)
- [ ] **Zero inline errors.New()** (all domain constants)
- [ ] Domain errors returned for business rule violations
- [ ] System errors wrapped with `fmt.Errorf(..., %w, err)` for context
- [ ] No sensitive information in error messages

### Handler Layer
- [ ] Handlers use `errors.Is()` for error discrimination
- [ ] Validation errors EXPOSE details (user input)
- [ ] Domain errors MAP to user-friendly messages
- [ ] System errors HIDE details (generic fallback)
- [ ] Proper HTTP status codes (404, 400, 409, 403, 500)
- [ ] No information leakage (no SQL errors, paths, IDs exposed)

### Tests
- [ ] Tests use `errors.Is()` instead of string comparison
- [ ] All error paths tested
- [ ] Mock repositories return domain errors
- [ ] Integration tests verify actual error returns

### Documentation
- [ ] Context README mentions domain errors pattern
- [ ] Complex error handling explained in comments
- [ ] Migration notes for breaking changes

---

## References

- **Gold Standard**: `internal/contexts/warehouse/location/` - 14 domain constants, zero fmt.Errorf
- **Refactoring Plan**: `docs/reference/DOMAIN_ERRORS_REFACTORING_PLAN.md` - Complete 18-session plan
- **Test Examples**: `test/integration/contexts/` - 34+ tests using errors.Is()
- **Security Patterns**: `docs/guides/security-patterns.md` - Handler security audit results

---

## Statistics

**Phase 2 Domain Errors Refactoring** (January 9-13, 2026):
- **Duration**: 2 weeks (18 sessions)
- **Effort**: ~22.5 hours
- **Aggregates with errors.go**: 24/24 (100%)
- **fmt.Errorf refactored**: 225+ calls
- **Domain errors defined**: 150+ constants
- **Integration tests updated**: 34+ with errors.Is()
- **Quality**: Production-ready

---

**Last Updated**: January 13, 2026  
**Status**: Complete (Phase 2 finished)  
**Maintainer**: Promenade Team
