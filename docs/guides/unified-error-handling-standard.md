# Unified Error Handling Standard

**Promenade Platform - Master Reference for Error Handling**

**Version**: 2.0 (January 9, 2026)  
**Status**: Production Standard  
**Applies To**: All contexts, aggregates, handlers, and use cases

---

## Executive Summary

This document unifies **two major refactoring efforts** into a single, comprehensive error handling standard:

1. **Phase 1: Handler Security Audit** (December 2025 - January 2026)
   - 36 handlers audited
   - 417 security fixes applied
   - Established validation vs system error patterns

2. **Phase 2: Domain Errors Refactoring** (January 2026)
   - 24 aggregates audited
   - ~225 fmt.Errorf instances identified
   - 18-session plan to eliminate inline errors
   - errors.go introduced for domain constants

**Result**: A **three-layer** error handling architecture ensuring:
-  **Security**: No information leakage to clients
-  **Consistency**: Same patterns across all aggregates
-  **Maintainability**: Domain errors in one place (errors.go)
-  **User Experience**: Clear, actionable error messages

---

## Table of Contents

1. [Historical Context](#historical-context)
2. [The Three-Layer Architecture](#the-three-layer-architecture)
3. [Layer 1: errors.go (Domain Constants)](#layer-1-errorsgo-domain-constants)
4. [Layer 2: usecase.go (Business Logic)](#layer-2-usecasego-business-logic)
5. [Layer 3: handler.go (HTTP Mapping)](#layer-3-handlergo-http-mapping)
6. [Security Patterns](#security-patterns)
7. [Quality Gates](#quality-gates)
8. [Practical Examples](#practical-examples)
9. [Common Mistakes](#common-mistakes)
10. [Migration Checklist](#migration-checklist)

---

## Historical Context

### Phase 1: Handler Security Audit (417 Fixes)

**Problem Identified**: Information leakage through error messages

**Example of Security Issue**:
```go
//  BEFORE - Information leakage
customer, err := h.usecase.GetCustomer(ctx, id)
if err != nil {
    response.InternalError(c, err.Error())  // Exposes: "sql: no rows in result set"
    return
}
```

**Solution Applied**:
```go
//  AFTER - Generic message
customer, err := h.usecase.GetCustomer(ctx, id)
if err != nil {
    response.InternalError(c, "Failed to retrieve customer")  // Safe
    return
}
```

**Key Discovery**: Different error types require different handling:
- **Validation errors** (user input) → Expose details (helps user fix mistakes)
- **System errors** (internal) → Hide details (prevent information leakage)

**Documentation**: `docs/guides/security-patterns.md`

### Phase 2: Domain Errors Refactoring (18 Sessions)

**Problem Identified**: Inline error creation scattered across codebase

**Example of Maintenance Issue**:
```go
//  BEFORE - Inline errors everywhere
if existingCode != nil {
    return nil, fmt.Errorf("location code already exists")
}

// In another method:
if codeExists {
    return nil, fmt.Errorf("code already exists")  // Inconsistent wording
}

// In handler:
if err.Error() == "location code already exists" {  // String comparison anti-pattern
    response.BadRequest(c, "Code exists")
    return
}
```

**Solution Applied**:
```go
//  AFTER - Domain constants in errors.go
var ErrLocationCodeExists = errors.New("location code already exists")

// In usecase:
if existingCode != nil {
    return nil, ErrLocationCodeExists  // Consistent
}

// In handler:
if errors.Is(err, location.ErrLocationCodeExists) {  // Type-safe
    response.BadRequest(c, "Location code already exists")
    return
}
```

**Key Discovery**: Domain errors should be:
- **Centralized** in errors.go (single source of truth)
- **Typed** for type-safe checking (errors.Is)
- **Documented** with clear comments

**Documentation**: docs/guides/domain-errors.md

### Integration: The Unified Standard

Both phases revealed complementary patterns:

| Aspect | Handler Security | Domain Errors |
|--------|-----------------|---------------|
| **Focus** | HTTP layer security | Domain layer consistency |
| **Problem** | Information leakage | Scattered inline errors |
| **Solution** | Generic system messages | Centralized domain constants |
| **Pattern** | Validation vs System | errors.go + usecase + handler |
| **Verification** | String pattern analysis | grep fmt.Errorf count |

**Unified Result**: Three-layer architecture combining security + maintainability

---

## The Three-Layer Architecture

```

  Layer 3: handler.go (HTTP Mapping)                     
    
   • Validation errors: err.Error() (expose)           
   • Domain errors: errors.Is() → user messages        
   • System errors: Generic fallback (hide)            
    
                         ↓                                
  Layer 2: usecase.go (Business Logic)                   
    
   • ZERO fmt.Errorf (strict)                          
   • ZERO inline errors.New                            
   • Only domain constants from errors.go              
    
                         ↓                                
  Layer 1: errors.go (Domain Constants)                  
    
   • Repository errors (data access)                   
   • Business errors (domain rules)                    
   • Technical errors (operation wrappers)             
    

```

**Key Principles**:
1. **Separation of Concerns**: Each layer has specific responsibilities
2. **Security by Design**: Layer 3 controls what users see
3. **Type Safety**: errors.Is() instead of string comparison
4. **Maintainability**: One place to update error messages (errors.go)

---

## Layer 1: errors.go (Domain Constants)

### Purpose

Central registry of all domain errors with three categories:
1. **Repository Errors** - Data access failures
2. **Business Logic Errors** - Domain rule violations
3. **Technical Operation Errors** - Wrapper failures

### Template

```go
package [aggregate]

import "errors"

// Repository Errors - Data access failures
var (
    // Err[Aggregate]NotFound is returned when [aggregate] doesn't exist
    Err[Aggregate]NotFound = errors.New("[aggregate] not found")
)

// Business Logic Errors - Domain rule violations
var (
    // Err[Aggregate][Condition] is returned when [business rule violated]
    Err[Aggregate][Condition] = errors.New("business rule description")
    
    // Err[Aggregate][Constraint] is returned when [constraint violated]
    Err[Aggregate][Constraint] = errors.New("constraint description")
    
    // ... all business validation errors ...
)

// Technical Operation Errors - Wrapper failures
var (
    // Err[Aggregate]CreateFailed is returned when [aggregate] creation fails
    Err[Aggregate]CreateFailed = errors.New("failed to create [aggregate]")
    
    // Err[Aggregate]UpdateFailed is returned when [aggregate] update fails
    Err[Aggregate]UpdateFailed = errors.New("failed to update [aggregate]")
    
    // Err[Aggregate]DeleteFailed is returned when [aggregate] deletion fails
    Err[Aggregate]DeleteFailed = errors.New("failed to delete [aggregate]")
    
    // Err[Aggregate]ListFailed is returned when [aggregate] listing fails
    Err[Aggregate]ListFailed = errors.New("failed to list [aggregates]")
)
```

### Real Example (warehouse/location)

```go
package location

import "errors"

// Repository Errors - Data access failures
var (
    // ErrLocationNotFound is returned when location doesn't exist
    ErrLocationNotFound = errors.New("location not found")
)

// Business Logic Errors - Domain rule violations
var (
    // ErrLocationCodeExists is returned when code already exists
    ErrLocationCodeExists = errors.New("location code already exists")
    
    // ErrParentLocationNotFound is returned when parent location doesn't exist
    ErrParentLocationNotFound = errors.New("parent location not found")
    
    // ErrParentLocationDeleted is returned when parent location is deleted
    ErrParentLocationDeleted = errors.New("parent location is deleted")
    
    // ErrLocationHasChildren is returned when attempting to delete location with children
    ErrLocationHasChildren = errors.New("cannot delete location with children")
    
    // ErrLocationHasItems is returned when attempting to delete location with items
    ErrLocationHasItems = errors.New("cannot delete location with items")
    
    // ErrLocationAlreadyDeleted is returned when location already deleted
    ErrLocationAlreadyDeleted = errors.New("location already deleted")
    
    // ErrLocationDeleted is returned when location is deleted
    ErrLocationDeleted = errors.New("location is deleted")
    
    // ErrCyclicParent is returned when cyclic parent relationship detected
    ErrCyclicParent = errors.New("cyclic parent relationship detected")
    
    // ErrInvalidPagination is returned when pagination parameters invalid
    ErrInvalidPagination = errors.New("invalid pagination parameters")
)

// Technical Operation Errors - Wrapper failures
var (
    // ErrLocationCreateFailed is returned when location creation fails
    ErrLocationCreateFailed = errors.New("failed to create location")
    
    // ErrLocationUpdateFailed is returned when location update fails
    ErrLocationUpdateFailed = errors.New("failed to update location")
    
    // ErrLocationDeleteFailed is returned when location deletion fails
    ErrLocationDeleteFailed = errors.New("failed to delete location")
    
    // ErrLocationListFailed is returned when location listing fails
    ErrLocationListFailed = errors.New("failed to list locations")
)
```

### Rules

 **DO**:
- Use 3 categories (Repository / Business / Technical)
- Document every error with comment
- Use consistent naming: `Err[Aggregate][Action/State]`
- Use standard library `errors.New()`
- Cover all domain scenarios

 **DON'T**:
- Mix categories (keep clear separation)
- Leave errors undocumented
- Use inconsistent naming
- Import fmt or other packages
- Skip any domain error scenarios

---

## Layer 2: usecase.go (Business Logic)

### Purpose

Business logic layer that:
- Validates business rules
- Orchestrates domain operations
- Returns ONLY domain constants (never inline errors)

### Strict Rules

**ABSOLUTE PROHIBITIONS**:
1.  **NEVER** use `fmt.Errorf`
2.  **NEVER** use inline `errors.New`
3.  **NEVER** import "fmt" or "errors" packages
4.  **ALWAYS** return domain constants from errors.go

### Template

```go
package [aggregate]

import (
    "context"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// Note: NO "fmt" or "errors" imports!

type IUseCase interface {
    CreateEntity(ctx context.Context, ...) (*Entity, error)
    GetEntity(ctx context.Context, id uuidv7.UUID) (*Entity, error)
    UpdateEntity(ctx context.Context, ...) error
    DeleteEntity(ctx context.Context, id uuidv7.UUID) error
}

type useCase struct {
    repo IRepository
}

func NewUseCase(repo IRepository) IUseCase {
    return &useCase{repo: repo}
}

func (uc *useCase) CreateEntity(ctx context.Context, ...) (*Entity, error) {
    // 1. Business validation → Domain error
    if violated {
        return nil, ErrBusinessRule  //  Domain constant
    }
    
    // 2. Check existing → Repository error
    existing, err := uc.repo.GetByCode(ctx, code)
    if err == nil && existing != nil {
        return nil, ErrEntityCodeExists  //  Business error
    }
    
    // 3. Create entity → Technical wrapper
    entity, err := NewEntity(...)
    if err != nil {
        return nil, ErrEntityCreateFailed  //  Technical wrapper
    }
    
    // 4. Persist → Technical wrapper
    if err := uc.repo.Create(ctx, entity); err != nil {
        return nil, ErrEntityCreateFailed  //  Technical wrapper
    }
    
    return entity, nil
}

func (uc *useCase) GetEntity(ctx context.Context, id uuidv7.UUID) (*Entity, error) {
    // Repository operation → Repository error (pass through)
    entity, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, ErrEntityNotFound  //  Repository error
    }
    
    // Business validation → Business error
    if entity.DeletedAt != nil {
        return nil, ErrEntityDeleted  //  Business error
    }
    
    return entity, nil
}

func (uc *useCase) DeleteEntity(ctx context.Context, id uuidv7.UUID) error {
    // 1. Get entity
    entity, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return ErrEntityNotFound  //  Repository error
    }
    
    // 2. Check business constraints
    if entity.DeletedAt != nil {
        return ErrEntityAlreadyDeleted  //  Business error
    }
    
    hasChildren, err := uc.repo.HasChildren(ctx, id)
    if err != nil {
        return ErrEntityDeleteFailed  //  Technical wrapper
    }
    if hasChildren {
        return ErrEntityHasChildren  //  Business error
    }
    
    // 3. Delete
    if err := uc.repo.Delete(ctx, id); err != nil {
        return ErrEntityDeleteFailed  //  Technical wrapper
    }
    
    return nil
}
```

### Real Example (warehouse/location - excerpt)

```go
func (uc *useCase) CreateLocation(ctx context.Context, code, name string, locationType LocationType, ...) (*Location, error) {
    // Business validation: Check unique code
    existing, err := uc.repo.GetByCode(ctx, code)
    if err == nil && existing != nil {
        return nil, ErrLocationCodeExists  //  Domain constant
    }
    
    // Create entity
    location, err := NewLocation(code, name, locationType)
    if err != nil {
        return nil, ErrLocationCreateFailed  //  Technical wrapper
    }
    
    // Validate parent if provided
    if parentID != nil {
        parent, err := uc.repo.GetByID(ctx, *parentID)
        if err != nil {
            return nil, ErrParentLocationNotFound  //  Business error
        }
        if parent.DeletedAt != nil {
            return nil, ErrParentLocationDeleted  //  Business error
        }
        location.ParentID = parentID
    }
    
    // Persist
    if err := uc.repo.Create(ctx, location); err != nil {
        return nil, ErrLocationCreateFailed  //  Technical wrapper
    }
    
    return location, nil
}
```

### Rules

 **DO**:
- Return domain constants only
- Wrap repository errors in technical constants
- Validate business rules with business error constants
- Keep imports clean (context + uuidv7 only)

 **DON'T**:
- Use fmt.Errorf or errors.New inline
- Import fmt or errors packages
- Return raw repository errors
- Create error messages dynamically

---

## Layer 3: handler.go (HTTP Mapping)

### Purpose

HTTP layer that:
- Maps domain errors to HTTP status codes
- Controls what error information users see
- Enforces security patterns

### The Three Error Types

#### 1. Validation Errors (User Input) - EXPOSE Details

**Pattern**:
```go
var req CreateDTO
if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(c, err.Error())  //  EXPOSE - Safe to show
    return
}
```

**Why Safe**:
- Error originates from user's own input
- Helps user correct their JSON structure/types
- No system information revealed
- Example: "Key: 'CreateDTO.Code' Error:Field validation for 'Code' failed on the 'required' tag"

**HTTP Status**: 400 Bad Request

#### 2. Business/Domain Errors - User-Friendly Messages

**Pattern**:
```go
entity, err := h.usecase.CreateEntity(c.Request.Context(), ...)
if err != nil {
    // Map specific business rules
    if errors.Is(err, domain.ErrEntityCodeExists) {
        response.BadRequest(c, "Entity code already exists")  //  User-friendly
        return
    }
    
    if errors.Is(err, domain.ErrEntityNotFound) {
        response.NotFound(c, "Entity not found")  //  User-friendly
        return
    }
    
    if errors.Is(err, domain.ErrEntityHasChildren) {
        response.BadRequest(c, "Cannot delete entity with children")  //  Clear reason
        return
    }
    
    // Fallback for unmapped errors (see next section)
    response.InternalError(c, "Failed to create entity")
    return
}
```

**Why Important**:
- Users need to understand business rule violations
- Clear messages improve UX
- No technical details leaked

**HTTP Status**: 
- 400 Bad Request (business rule violations)
- 404 Not Found (entity doesn't exist)

#### 3. System/Technical Errors - HIDE Details

**Pattern**:
```go
// Technical errors fall through to generic message
entity, err := h.usecase.CreateEntity(c.Request.Context(), ...)
if err != nil {
    // ... business error checks above ...
    
    // No explicit check for ErrEntityCreateFailed (technical wrapper)
    // Falls through to generic message
    response.InternalError(c, "Failed to create entity")  //  Hide internals
    return
}
```

**Why Important**:
- Technical errors reveal system internals
- Database errors show schema details
- Generic messages prevent information leakage

**HTTP Status**: 500 Internal Server Error

### Complete Handler Template

```go
package handler

import (
    "errors"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/basilex/promenade/pkg/response"
    "github.com/basilex/promenade/pkg/uuidv7"
    "[aggregate-usecase-import]"
)

type EntityHandler struct {
    usecase [aggregate].IUseCase
}

func NewEntityHandler(uc [aggregate].IUseCase) *EntityHandler {
    return &EntityHandler{usecase: uc}
}

// Create handles POST /entities
func (h *EntityHandler) Create(c *gin.Context) {
    // 1. VALIDATION ERRORS - Expose details (safe)
    var req CreateEntityDTO
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())  //  Validation error
        return
    }
    
    // 2. Format validation - User-friendly message
    var parentID *uuidv7.UUID
    if req.ParentID != "" {
        id, err := uuidv7.Parse(req.ParentID)
        if err != nil {
            response.BadRequest(c, "Invalid parent ID format")  //  Generic format error
            return
        }
        parentID = &id
    }
    
    // 3. BUSINESS LOGIC - Map domain errors
    entity, err := h.usecase.CreateEntity(
        c.Request.Context(),
        req.Code,
        req.Name,
        req.Type,
        parentID,
    )
    if err != nil {
        // Business rule violations → 400
        if errors.Is(err, [aggregate].ErrEntityCodeExists) {
            response.BadRequest(c, "Entity code already exists")
            return
        }
        if errors.Is(err, [aggregate].ErrParentEntityNotFound) {
            response.NotFound(c, "Parent entity not found")
            return
        }
        if errors.Is(err, [aggregate].ErrParentEntityDeleted) {
            response.BadRequest(c, "Parent entity is deleted")
            return
        }
        
        // Not found → 404
        if errors.Is(err, [aggregate].ErrEntityNotFound) {
            response.NotFound(c, "Entity not found")
            return
        }
        
        // Technical errors → 500 (generic message, hide internals)
        response.InternalError(c, "Failed to create entity")
        return
    }
    
    // 4. SUCCESS
    response.Created(c, entity)
}

// GetByID handles GET /entities/:id
func (h *EntityHandler) GetByID(c *gin.Context) {
    // Parse ID
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "Invalid ID format")
        return
    }
    
    // Get entity
    entity, err := h.usecase.GetEntity(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, [aggregate].ErrEntityNotFound) {
            response.NotFound(c, "Entity not found")
            return
        }
        if errors.Is(err, [aggregate].ErrEntityDeleted) {
            response.BadRequest(c, "Entity is deleted")
            return
        }
        response.InternalError(c, "Failed to retrieve entity")
        return
    }
    
    response.Success(c, entity)
}

// Delete handles DELETE /entities/:id
func (h *EntityHandler) Delete(c *gin.Context) {
    // Parse ID
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "Invalid ID format")
        return
    }
    
    // Delete entity
    if err := h.usecase.DeleteEntity(c.Request.Context(), id); err != nil {
        if errors.Is(err, [aggregate].ErrEntityNotFound) {
            response.NotFound(c, "Entity not found")
            return
        }
        if errors.Is(err, [aggregate].ErrEntityAlreadyDeleted) {
            response.BadRequest(c, "Entity already deleted")
            return
        }
        if errors.Is(err, [aggregate].ErrEntityHasChildren) {
            response.BadRequest(c, "Cannot delete entity with children")
            return
        }
        if errors.Is(err, [aggregate].ErrEntityHasItems) {
            response.BadRequest(c, "Cannot delete entity with items")
            return
        }
        response.InternalError(c, "Failed to delete entity")
        return
    }
    
    response.Success(c, gin.H{"message": "Entity deleted successfully"})
}
```

### Rules

 **DO**:
- Expose validation errors (err.Error() in ShouldBindJSON)
- Map business errors with errors.Is() + user-friendly messages
- Hide technical errors with generic fallback
- Use appropriate HTTP status codes (400/404/500)

 **DON'T**:
- Use err.Error() for system/business errors
- Compare error strings (if err.Error() == "...")
- Expose internal error details to users
- Skip domain error mapping (always check errors.Is)

---

## Security Patterns

### The Golden Rule

**Different error types require different handling strategies:**

| Error Type | Source | Handler Pattern | HTTP Status | Expose Details? |
|-----------|--------|----------------|-------------|-----------------|
| **Validation** | User input (ShouldBindJSON) | `err.Error()` | 400 |  YES - Safe |
| **Business** | Domain rules | `errors.Is()` + message | 400/404 |  YES - Controlled |
| **Technical** | System operations | Generic fallback | 500 |  NO - Hide |

### Security Checklist (Per Handler)

Before marking handler as secure, verify:

- [ ] All `ShouldBindJSON()` errors use `response.BadRequest(c, err.Error())`
- [ ] All usecase errors use generic messages (`"Failed to..."`)
- [ ] No `err.Error()` used for system/business errors
- [ ] No string comparisons (`err.Error() == "..."`)
- [ ] No `strings.Contains(err.Error(), ...)`
- [ ] All domain errors mapped with `errors.Is()`
- [ ] Generic fallback message for unmapped errors
- [ ] Tests verify error messages don't leak system details

### Common Security Anti-Patterns

####  Anti-Pattern 1: System Error Leakage
```go
// WRONG - Exposes database details
customer, err := h.usecase.GetCustomer(ctx, id)
if err != nil {
    response.InternalError(c, err.Error())  //  "sql: no rows in result set"
    return
}
```

**Fix**:
```go
// CORRECT - Generic message
customer, err := h.usecase.GetCustomer(ctx, id)
if err != nil {
    if errors.Is(err, customer.ErrCustomerNotFound) {
        response.NotFound(c, "Customer not found")  //  User-friendly
        return
    }
    response.InternalError(c, "Failed to retrieve customer")  //  Generic
    return
}
```

####  Anti-Pattern 2: String Comparison
```go
// WRONG - Brittle, unsafe
if err.Error() == "customer not found" {
    response.NotFound(c, "Not found")
    return
}
```

**Fix**:
```go
// CORRECT - Type-safe
if errors.Is(err, customer.ErrCustomerNotFound) {
    response.NotFound(c, "Customer not found")
    return
}
```

####  Anti-Pattern 3: Missing Validation Error Handling
```go
// WRONG - Generic message for validation error
var req CreateDTO
if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(c, "Invalid request")  //  User can't fix it
    return
}
```

**Fix**:
```go
// CORRECT - Expose validation details
var req CreateDTO
if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(c, err.Error())  //  Shows what's wrong
    return
}
```

---

## Quality Gates

### Automated Verification

Before marking any session complete, run these checks:

#### 1. errors.go Structure Check
```bash
# Verify 3 categories exist
grep -c "// Repository\|// Business\|// Technical" errors.go
# Expected: 3

# Verify all errors documented
grep -c "^    //" errors.go
# Should match number of error constants
```

#### 2. usecase.go Pattern Compliance
```bash
# MUST be 0 (strict enforcement)
grep -c "fmt\.Errorf" usecase.go

# MUST be 0 (all domain constants)
grep -c "errors\.New" usecase.go

# Verify clean imports (should NOT find fmt or errors)
head -20 usecase.go | grep -E "import.*fmt|import.*errors"
```

#### 3. handler.go Security Check
```bash
# Find all err.Error() usage
grep -n "err\.Error()" handler.go

# Verify each is in ShouldBindJSON context (manual review needed)
# All should be within 3 lines after "ShouldBindJSON"

# Verify domain error mapping exists
grep -c "errors\.Is" handler.go
# Should be > 0 (many checks expected)
```

#### 4. Lint & Test Validation
```bash
# Should pass with 0 issues
make ci-lint

# Should pass with 100% success
go test ./internal/contexts/[context]/[aggregate]/... -v

# Smoke tests should pass
make test-smoke
```

### Manual Review Checklist

#### errors.go Review
- [ ] 3 categories clearly separated (Repository/Business/Technical)
- [ ] Every error has explanatory comment
- [ ] Consistent naming: `Err[Aggregate][Action/State]`
- [ ] Uses standard library `errors.New()`
- [ ] Comprehensive coverage (no missing scenarios)

#### usecase.go Review
- [ ] Zero `fmt.Errorf` usages
- [ ] Zero inline `errors.New()` calls
- [ ] Only domain constants from errors.go
- [ ] Clean imports (no fmt, no errors)
- [ ] Consistent pattern across all methods

#### handler.go Review
- [ ] Validation errors expose `err.Error()` (ShouldBindJSON only)
- [ ] Domain errors mapped with `errors.Is()`
- [ ] User-friendly messages for business errors
- [ ] Generic messages for technical errors
- [ ] Appropriate HTTP status codes (400/404/500)

---

## Practical Examples

### Example 1: warehouse/location (Gold Standard)

**Complete implementation from Session 1**

#### errors.go (14 errors, 3 categories)
```go
package location

import "errors"

// Repository Errors
var (
    ErrLocationNotFound = errors.New("location not found")
)

// Business Logic Errors
var (
    ErrLocationCodeExists = errors.New("location code already exists")
    ErrParentLocationNotFound = errors.New("parent location not found")
    ErrParentLocationDeleted = errors.New("parent location is deleted")
    ErrLocationHasChildren = errors.New("cannot delete location with children")
    ErrLocationHasItems = errors.New("cannot delete location with items")
    ErrLocationAlreadyDeleted = errors.New("location already deleted")
    ErrLocationDeleted = errors.New("location is deleted")
    ErrCyclicParent = errors.New("cyclic parent relationship detected")
    ErrInvalidPagination = errors.New("invalid pagination parameters")
)

// Technical Operation Errors
var (
    ErrLocationCreateFailed = errors.New("failed to create location")
    ErrLocationUpdateFailed = errors.New("failed to update location")
    ErrLocationDeleteFailed = errors.New("failed to delete location")
    ErrLocationListFailed = errors.New("failed to list locations")
)
```

#### usecase.go (excerpt - 0 fmt.Errorf)
```go
func (uc *useCase) CreateLocation(ctx context.Context, code, name string, ...) (*Location, error) {
    // Check unique code
    existing, err := uc.repo.GetByCode(ctx, code)
    if err == nil && existing != nil {
        return nil, ErrLocationCodeExists  //  Domain constant
    }
    
    // Create entity
    location, err := NewLocation(code, name, locationType)
    if err != nil {
        return nil, ErrLocationCreateFailed  //  Technical wrapper
    }
    
    // Validate parent
    if parentID != nil {
        parent, err := uc.repo.GetByID(ctx, *parentID)
        if err != nil {
            return nil, ErrParentLocationNotFound  //  Business error
        }
        if parent.DeletedAt != nil {
            return nil, ErrParentLocationDeleted  //  Business error
        }
    }
    
    // Persist
    if err := uc.repo.Create(ctx, location); err != nil {
        return nil, ErrLocationCreateFailed  //  Technical wrapper
    }
    
    return location, nil
}
```

#### handler.go (excerpt - proper security pattern)
```go
func (h *LocationHandler) Create(c *gin.Context) {
    // Validation error - expose details
    var req CreateLocationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())  //  Safe to expose
        return
    }
    
    // Create location
    loc, err := h.usecase.CreateLocation(c.Request.Context(), ...)
    if err != nil {
        // Business errors - user-friendly messages
        if errors.Is(err, location.ErrLocationCodeExists) {
            response.BadRequest(c, "Location code already exists")
            return
        }
        if errors.Is(err, location.ErrParentLocationNotFound) {
            response.NotFound(c, "Parent location not found")
            return
        }
        if errors.Is(err, location.ErrParentLocationDeleted) {
            response.BadRequest(c, "Parent location is deleted")
            return
        }
        
        // Technical errors - generic message
        response.InternalError(c, "Failed to create location")
        return
    }
    
    response.Created(c, loc)
}
```

**Verification**:
-  errors.go: 14 errors, 3 categories, all documented
-  usecase.go: 0 fmt.Errorf, 0 errors.New
-  handler.go: err.Error() only in ShouldBindJSON, 20+ errors.Is() checks
-  Security: No information leakage, proper error exposure

---

## Common Mistakes

### Mistake 1: Using fmt.Errorf in UseCase

```go
//  WRONG
func (uc *useCase) CreateEntity(ctx context.Context, code string) (*Entity, error) {
    if existingCode {
        return nil, fmt.Errorf("entity code %s already exists", code)
    }
}
```

**Why Wrong**: 
- Violates Layer 2 pattern (usecase should use domain constants only)
- Not type-safe (can't use errors.Is() in handler)
- Inconsistent across methods

**Fix**:
```go
//  CORRECT
// In errors.go:
var ErrEntityCodeExists = errors.New("entity code already exists")

// In usecase.go:
func (uc *useCase) CreateEntity(ctx context.Context, code string) (*Entity, error) {
    if existingCode {
        return nil, ErrEntityCodeExists  // Type-safe, consistent
    }
}
```

### Mistake 2: Missing Error Categories in errors.go

```go
//  WRONG - All errors in one block
var (
    ErrEntityNotFound = errors.New("entity not found")
    ErrEntityCodeExists = errors.New("code exists")
    ErrEntityCreateFailed = errors.New("create failed")
)
```

**Why Wrong**:
- No clear separation of concerns
- Hard to understand error purpose
- Maintainability issues

**Fix**:
```go
//  CORRECT - Clear categories
// Repository Errors
var (
    ErrEntityNotFound = errors.New("entity not found")
)

// Business Logic Errors
var (
    ErrEntityCodeExists = errors.New("entity code already exists")
)

// Technical Operation Errors
var (
    ErrEntityCreateFailed = errors.New("failed to create entity")
)
```

### Mistake 3: Exposing System Errors in Handler

```go
//  WRONG
func (h *Handler) GetEntity(c *gin.Context) {
    entity, err := h.usecase.GetEntity(ctx, id)
    if err != nil {
        response.InternalError(c, err.Error())  //  Information leakage
        return
    }
}
```

**Why Wrong**:
- Exposes database errors ("sql: no rows in result set")
- Reveals system internals
- Security risk

**Fix**:
```go
//  CORRECT
func (h *Handler) GetEntity(c *gin.Context) {
    entity, err := h.usecase.GetEntity(ctx, id)
    if err != nil {
        if errors.Is(err, domain.ErrEntityNotFound) {
            response.NotFound(c, "Entity not found")  // User-friendly
            return
        }
        response.InternalError(c, "Failed to retrieve entity")  // Generic
        return
    }
}
```

### Mistake 4: String Comparison Instead of errors.Is

```go
//  WRONG
func (h *Handler) Create(c *gin.Context) {
    entity, err := h.usecase.CreateEntity(ctx, ...)
    if err != nil {
        if err.Error() == "entity code already exists" {  // Brittle!
            response.BadRequest(c, "Code exists")
            return
        }
    }
}
```

**Why Wrong**:
- Brittle (breaks if error message changes)
- Not type-safe
- Hard to refactor

**Fix**:
```go
//  CORRECT
func (h *Handler) Create(c *gin.Context) {
    entity, err := h.usecase.CreateEntity(ctx, ...)
    if err != nil {
        if errors.Is(err, domain.ErrEntityCodeExists) {  // Type-safe!
            response.BadRequest(c, "Entity code already exists")
            return
        }
    }
}
```

### Mistake 5: Missing Validation Error Details

```go
//  WRONG
func (h *Handler) Create(c *gin.Context) {
    var req CreateDTO
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid request")  // Too generic!
        return
    }
}
```

**Why Wrong**:
- User doesn't know what's wrong with their JSON
- Can't fix the problem
- Poor UX

**Fix**:
```go
//  CORRECT
func (h *Handler) Create(c *gin.Context) {
    var req CreateDTO
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())  // Specific validation error
        return
    }
}
```

---

## Migration Checklist

### For Each Aggregate (Sessions 2-18)

#### Step 1: Create errors.go (15-30 min)
- [ ] Create file in aggregate root directory
- [ ] Add 3 category sections (Repository/Business/Technical)
- [ ] Document all existing error scenarios
- [ ] Use consistent naming (`Err[Aggregate][Action/State]`)
- [ ] Add comments above each error
- [ ] Verify structure with grep (should have 3 categories)

#### Step 2: Refactor usecase.go (20-40 min)
- [ ] Find all `fmt.Errorf` calls (grep)
- [ ] Find all inline `errors.New` calls (grep)
- [ ] Replace with domain constants from errors.go
- [ ] Remove fmt and errors imports
- [ ] Verify with grep (should be 0 fmt.Errorf, 0 errors.New)
- [ ] Run tests to ensure no regressions

#### Step 3: Update handler.go (15-30 min)
- [ ] Find all err.Error() usage (grep)
- [ ] Verify each is in ShouldBindJSON context (if not, fix)
- [ ] Add errors.Is() checks for all domain errors
- [ ] Add user-friendly messages for business errors
- [ ] Add generic fallback for technical errors
- [ ] Verify with grep (err.Error() only in validation)

#### Step 4: Verification (10-15 min)
- [ ] Run automated checks (grep commands from Quality Gates)
- [ ] Run lint: `make ci-lint` (should pass)
- [ ] Run tests: `go test ./... -v` (should pass)
- [ ] Run smoke tests: `make test-smoke` (should pass)
- [ ] Manual review against checklist

#### Step 5: Documentation (5-10 min)
- [ ] Update session summary
- [ ] Document any deviations from standard
- [ ] Note fixes count for tracking

**Total Time Per Aggregate**: 65-125 minutes (avg ~90 min)

### Overall Progress Tracking

**Completed**:
-  Session 1: warehouse/location (~70 fixes, gold standard established)

**Remaining** (17 sessions):
1. warehouse/inventory (~30 fixes)
2. warehouse/stockmovement (~25 fixes)
3. warehouse/product (~30 fixes)
4. billing/invoice (~15 fixes)
5. billing/payment (~10 fixes)
6. billing/subscription (~10 fixes)
7. order-mgmt/order (~15 fixes)
8. order-mgmt/contract (~15 fixes)
9. customer-mgmt/customer (~20 fixes)
10. customer-mgmt/company (~15 fixes)
11. customer-mgmt/deal (~10 fixes)
12. customer-mgmt/interaction (~15 fixes)
13. identity/user (~15 fixes)
14. identity/contact (~10 fixes)
15. identity/profile (~10 fixes)
16. identity/role (~5 fixes)
17. identity/permission (~5 fixes)

**Total Remaining**: ~155 fixes, ~25.5 hours

---

## Conclusion

### The Unified Standard Summary

This document combines two major refactoring efforts into a cohesive, production-ready error handling standard:

**Phase 1 (Security)** → No information leakage  
**Phase 2 (Maintainability)** → Centralized domain errors  
**Result** → Three-layer architecture ensuring both security and consistency

### Key Takeaways

1. **errors.go**: Single source of truth for domain errors (3 categories)
2. **usecase.go**: Zero inline errors, domain constants only (strict)
3. **handler.go**: Proper error exposure - validation yes, system no (security)
4. **Quality Gates**: Automated verification prevents regressions (grep checks)
5. **Gold Standard**: warehouse/location is the reference implementation

### Success Metrics

**When standard is fully implemented** (after Session 18):
-  0 fmt.Errorf in all usecases (24 aggregates)
-  100% security pattern compliance (36 handlers)
-  Consistent error handling across all contexts
-  No information leakage to clients
-  Type-safe error checking (errors.Is everywhere)

### Reference Documents

- **This Document**: Master reference for all error handling
- **Phase 1**: `docs/guides/security-patterns.md` (handler security audit)
- **Phase 2**: docs/refactoring/README.md (consolidated status)
- **Gold Standard**: `docs/work-in-progress/LOCATION_AGGREGATE_QUALITY_AUDIT.md` (Session 1)

---

**Version**: 2.0  
**Last Updated**: January 9, 2026  
**Status**: Production Standard  
**Compliance**: Mandatory for all new code and refactoring sessions

