# Domain Error Patterns - Consolidated Guide

**Purpose**: Unified patterns discovered across all Phase 2 sessions  
**Sessions Covered**: 2-10 (11 sessions, 341 eliminations)  
**Status**: Production-ready patterns

---

## Core Principles

### 1. Three-Layer Error Architecture

**Layer 1: errors.go** - Domain constants
```go
// Repository Errors
var ErrCustomerNotFound = errors.New("customer not found")

// Business Logic Errors
var ErrCustomerAlreadyExists = errors.New("customer already exists")
var ErrInvalidCustomerTier = errors.New("invalid customer tier")

// Technical Operation Errors
var ErrCustomerCreateFailed = errors.New("failed to create customer")
```

**Layer 2: usecase.go** - Zero inline errors
```go
//  NEVER DO THIS
func (uc *useCase) Create(...) error {
    return fmt.Errorf("failed to create: %w", err)  // WRONG
}

//  ALWAYS DO THIS
func (uc *useCase) Create(...) error {
    return ErrCustomerCreateFailed  // Domain constant
}
```

**Layer 3: handler.go** - Security-aware mapping
```go
// Validation errors: EXPOSE details
if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(c, err.Error())  // User needs feedback
    return
}

// Domain errors: MAP to user-friendly messages
if errors.Is(err, customer.ErrCustomerNotFound) {
    response.NotFound(c, "Customer not found")  // Clear message
    return
}

// System errors: HIDE details (security)
response.InternalError(c, "Failed to create customer")  // Generic
```

### 2. errors.Is() Pattern (Type-Safe)

**Replace string comparison** with type-safe checks:

```go
//  OLD - String comparison (fragile)
if err != nil && err.Error() == "not found" {
    return ErrCustomerNotFound
}

//  NEW - Type-safe with errors.Is()
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return ErrCustomerNotFound
    }
}
```

**Benefits**:
- Compile-time safety (typos caught immediately)
- Works with wrapped errors (`fmt.Errorf("%w", err)`)
- Refactor-friendly (rename constant, all usages update)

### 3. Handler Error Mapping

**Standard mapping patterns by error type**:

```go
// Not Found (404)
if errors.Is(err, customer.ErrCustomerNotFound) {
    response.NotFound(c, "Customer not found")
    return
}

// Bad Request (400)
if errors.Is(err, customer.ErrInvalidCustomerTier) {
    response.BadRequest(c, "Invalid tier value")
    return
}

// Conflict (409)
if errors.Is(err, customer.ErrCustomerAlreadyExists) {
    response.Conflict(c, "Customer with this email already exists")
    return
}

// Internal Error (500) - Fallback
response.InternalError(c, "Failed to process request")
```

---

## Pattern Library

### Pattern 1: Repository Error Wrapping

**Purpose**: Convert database errors to domain errors

**Implementation**:
```go
// In repository
func (r *repository) GetCustomer(ctx context.Context, id uuid.UUID) (*Customer, error) {
    var row customerRow
    query := `SELECT * FROM customers WHERE id = $1 AND deleted_at IS NULL`
    
    err := r.Get(ctx, &row, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, customer.ErrCustomerNotFound  // Domain error
        }
        return nil, customer.ErrCustomerGetFailed  // Technical wrapper
    }
    
    return row.toEntity()
}
```

**Sessions**: 2, 3, 4, 5, 6, 7, 8, 9, 10 (universal pattern)

### Pattern 2: Business Logic Validation

**Purpose**: Validate business rules before database operations

**Implementation**:
```go
// In use case
func (uc *useCase) UpgradeTier(ctx context.Context, id uuid.UUID, tier string) error {
    // Validate tier value
    validTiers := []string{"free", "basic", "pro", "enterprise"}
    if !contains(validTiers, tier) {
        return customer.ErrInvalidCustomerTier  // Domain error
    }
    
    // Get customer
    customer, err := uc.repo.GetCustomer(ctx, id)
    if err != nil {
        return err  // Propagate domain error
    }
    
    // Business rule: Can't downgrade from enterprise
    if customer.Tier == "enterprise" && tier != "enterprise" {
        return customer.ErrCannotDowngradeFromEnterprise
    }
    
    // Apply change
    customer.Tier = tier
    return uc.repo.UpdateCustomer(ctx, customer)
}
```

**Sessions**: 4, 5, 6, 7, 8 (business-heavy contexts)

### Pattern 3: Security-First Handler Design

**Purpose**: Prevent information leakage through error messages

**Implementation**:
```go
func (h *Handler) Create(c *gin.Context) {
    var req CreateRequest
    
    // Validation: EXPOSE details (safe)
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())  // User input feedback
        return
    }
    
    // Business logic
    customer, err := h.usecase.CreateCustomer(c.Request.Context(), req.Email, req.Name)
    if err != nil {
        // Map known errors
        if errors.Is(err, customer.ErrCustomerAlreadyExists) {
            response.Conflict(c, "Customer with this email already exists")
            return
        }
        
        // Hide system errors (CRITICAL)
        response.InternalError(c, "Failed to create customer")
        return
    }
    
    response.Created(c, toDTO(customer))
}
```

**Critical**: Never expose:
- `sql: no rows in result set` → "Not found"
- `pq: duplicate key value` → "Already exists"
- Database field names or table structures
- Internal paths or stack traces

**Sessions**: All (Phase 1 Security Audit - 417 fixes)

### Pattern 4: Aggregate Root Error Propagation

**Purpose**: Entity methods return domain errors directly

**Implementation**:
```go
// In entity
func (c *Customer) UpgradeTier(newTier string) error {
    validTiers := []string{"free", "basic", "pro", "enterprise"}
    if !contains(validTiers, newTier) {
        return ErrInvalidCustomerTier  // Domain error
    }
    
    if c.Tier == "enterprise" && newTier != "enterprise" {
        return ErrCannotDowngradeFromEnterprise
    }
    
    c.Tier = newTier
    c.Touch()  // Update timestamp
    return nil
}

// In use case
func (uc *useCase) UpgradeTier(ctx context.Context, id uuid.UUID, tier string) error {
    customer, err := uc.repo.GetCustomer(ctx, id)
    if err != nil {
        return err
    }
    
    // Entity enforces business rules
    if err := customer.UpgradeTier(tier); err != nil {
        return err  // Propagate entity error
    }
    
    return uc.repo.UpdateCustomer(ctx, customer)
}
```

**Sessions**: 5 (Order critical bug), 6, 7, 8, 9 (DDD-heavy contexts)

### Pattern 5: Testing with errors.Is()

**Purpose**: Type-safe error assertions in tests

**Implementation**:
```go
func TestUseCase_CreateCustomer_AlreadyExists(t *testing.T) {
    // Setup
    mockRepo := &MockRepository{
        GetCustomerFunc: func(ctx context.Context, email string) (*Customer, error) {
            return fakeCustomer(), nil  // Customer exists
        },
    }
    uc := NewUseCase(mockRepo)
    
    // Execute
    _, err := uc.CreateCustomer(ctx, "test@example.com", "Test User")
    
    // Assert with errors.Is() - Type-safe
    assert.Error(t, err)
    assert.True(t, errors.Is(err, customer.ErrCustomerAlreadyExists))
    
    //  DON'T DO THIS - String comparison
    // assert.Equal(t, "customer already exists", err.Error())
}
```

**Sessions**: 10 (Entity Tests - 64 anti-patterns fixed)

### Pattern 6: JSONB Field Validation

**Purpose**: Validate JSONB arrays/objects before database operations

**Implementation**:
```go
// In entity
func (c *Customer) AddTag(tag string) error {
    if tag == "" {
        return ErrEmptyTag
    }
    
    // Get current tags
    tags := c.Tags.Get()
    
    // Check duplicates
    if contains(tags, tag) {
        return ErrTagAlreadyExists
    }
    
    // Add tag
    tags = append(tags, tag)
    c.Tags.Set(tags)
    c.Touch()
    return nil
}
```

**Sessions**: 4, 7, 8 (contexts with JSONB fields)

### Pattern 7: Soft Delete Error Handling

**Purpose**: Distinguish between "not found" and "deleted"

**Implementation**:
```go
// In repository
func (r *repository) GetCustomer(ctx context.Context, id uuid.UUID) (*Customer, error) {
    var row customerRow
    
    // Check if exists (including deleted)
    queryAll := `SELECT * FROM customers WHERE id = $1`
    err := r.Get(ctx, &row, queryAll, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, customer.ErrCustomerNotFound
        }
        return nil, customer.ErrCustomerGetFailed
    }
    
    // Check if deleted
    if row.DeletedAt != nil {
        return nil, customer.ErrCustomerDeleted  // Specific error
    }
    
    return row.toEntity()
}
```

**Sessions**: 2, 3, 4, 6, 7 (contexts with soft delete)

### Pattern 8: Foreign Key Error Handling

**Purpose**: Map FK constraint violations to domain errors

**Implementation**:
```go
// In repository
func (r *repository) CreateOrder(ctx context.Context, order *Order) error {
    query := `INSERT INTO orders (id, customer_id, ...) VALUES ($1, $2, ...)`
    
    err := r.Exec(ctx, query, order.ID, order.CustomerID, ...)
    if err != nil {
        // Check for FK violation (PostgreSQL error code 23503)
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
            if strings.Contains(pqErr.Constraint, "customer_id") {
                return order.ErrCustomerNotFound  // Domain error
            }
        }
        return order.ErrOrderCreateFailed
    }
    
    return nil
}
```

**Sessions**: 5, 8, 9 (contexts with FK relationships)

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: String Comparison 

**Bad**:
```go
if err != nil && err.Error() == "not found" {
    return ErrCustomerNotFound
}
```

**Why**: Fragile, typos undetected, breaks with wrapped errors

**Good**:
```go
if err != nil && errors.Is(err, sql.ErrNoRows) {
    return ErrCustomerNotFound
}
```

### Anti-Pattern 2: Inline fmt.Errorf() 

**Bad**:
```go
func (uc *useCase) Create(...) error {
    if name == "" {
        return fmt.Errorf("name is required")  // Inline
    }
}
```

**Why**: Not a domain constant, can't use errors.Is(), hard to reuse

**Good**:
```go
// In errors.go
var ErrNameRequired = errors.New("name is required")

// In usecase.go
func (uc *useCase) Create(...) error {
    if name == "" {
        return ErrNameRequired  // Domain constant
    }
}
```

### Anti-Pattern 3: Information Leakage 

**Bad**:
```go
// Handler exposes database errors
if err != nil {
    response.InternalError(c, err.Error())  // "sql: no rows in result set"
}
```

**Why**: Security risk, leaks implementation details

**Good**:
```go
// Handler maps to user-friendly messages
if err != nil {
    if errors.Is(err, customer.ErrCustomerNotFound) {
        response.NotFound(c, "Customer not found")
        return
    }
    response.InternalError(c, "Failed to process request")  // Generic
}
```

### Anti-Pattern 4: Error Wrapping Business Logic 

**Bad**:
```go
// Wrapping a domain error (Session 5 critical bug)
if err := order.Confirm(); err != nil {
    return fmt.Errorf("failed to confirm order: %w", err)  // WRONG!
}
```

**Why**: Changes error type, breaks errors.Is() checks in handlers

**Good**:
```go
// Return domain error directly
if err := order.Confirm(); err != nil {
    return err  // ErrOrderAlreadyConfirmed, ErrOrderCancelled, etc.
}
```

### Anti-Pattern 5: Missing Domain Constants 

**Bad**:
```go
// Repository returns generic errors
if err != nil {
    return err  // Propagates sql.ErrNoRows
}
```

**Why**: Leaks database abstraction, handler can't map properly

**Good**:
```go
// Repository returns domain errors
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return customer.ErrCustomerNotFound
    }
    return customer.ErrCustomerGetFailed
}
```

---

## Metrics Summary

| Session | Context | Aggregate | Eliminations | Constants | Pattern Highlights |
|---------|---------|-----------|--------------|-----------|-------------------|
| 2 | Warehouse | Inventory | 17 | 22 | Repository wrapping, JSONB validation |
| 3 | Warehouse | StockMovement | 28 | 13 | Audit trail errors, movement types |
| 4 | Warehouse | Product | 60 | 27 | SKU uniqueness, FK validation |
| 5 | Customer-Mgmt | Deal | 34 | 19 | Business rules, stage transitions |
| 5b | Order-Mgmt | Order | 29 | 31 | **Critical bug fix** (error wrapping) |
| 6 | Identity | User | 32 | 24 | Security patterns, handler hardening |
| 7 | Customer-Mgmt | Company | 26 | 17 | Hierarchical validation, parent checks |
| 8 | Billing | Subscription | 30 | 16 | Lifecycle validation, state machine |
| 9 | Order-Mgmt | Contract | 22 | 14 | Version control, approval workflow |
| 10 | Warehouse | Inventory (Tests) | 64 | 13 | **Test anti-patterns**, errors.Is() |
| **TOTAL** | **6 contexts** | **10 aggregates** | **341** | **210** | **11 sessions** |

---

## Quick Reference

### Checklist for New Aggregate

- [ ] Create `errors.go` with domain constants (3 categories)
- [ ] Zero `fmt.Errorf()` in `usecase.go`
- [ ] Map domain errors in handlers (validation vs system)
- [ ] Use `errors.Is()` everywhere (never string comparison)
- [ ] Test error paths with `errors.Is()` assertions
- [ ] Document error mapping in handler comments
- [ ] Security review: No information leakage

### Error Categories Template

```go
// errors.go
package aggregate

import "errors"

// Repository Errors (data access)
var (
    ErrEntityNotFound = errors.New("entity not found")
    ErrEntityDeleted = errors.New("entity already deleted")
)

// Business Logic Errors (domain rules)
var (
    ErrInvalidValue = errors.New("invalid value")
    ErrCannotPerformAction = errors.New("cannot perform action in current state")
)

// Technical Operation Errors (wrappers)
var (
    ErrEntityCreateFailed = errors.New("failed to create entity")
    ErrEntityUpdateFailed = errors.New("failed to update entity")
    ErrEntityDeleteFailed = errors.New("failed to delete entity")
)
```

---

**Version**: 1.0  
**Last Updated**: January 11, 2026  
**Sessions**: 2-10 (11 total)  
**Status**: Production patterns from 341 eliminations across 10 aggregates
