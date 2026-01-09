# Security Patterns

**Best practices for secure error handling in HTTP handlers** - Preventing information leakage while maintaining user experience.

---

## Overview

This document describes the security patterns established during the comprehensive security audit (January 2026) of all 36 HTTP handlers across 7 bounded contexts in the Promenade platform.

**Audit Results**:
- **Handlers Audited**: 36/36 (100%)
- **Security Fixes**: 417 applied
- **Sessions**: 18 completed
- **Code Quality**: 0 lint issues
- **Tests**: 2465+ passing (100%)

---

## Core Principle

**Different types of errors require different handling strategies**:

1. **Validation Errors** (user input) → **Detailed feedback** (safe to expose)
2. **System Errors** (internal operations) → **Generic messages** (hide implementation)

This separation prevents information leakage while maintaining good user experience.

---

## The Gold Standard Pattern

### Validation Errors - Expose Details

**When**: User provides invalid input (binding errors, format errors, constraint violations)

**Pattern**:
```go
func (h *Handler) Create(c *gin.Context) {
    var req CreateDTO
    
    // ✅ CORRECT - Expose validation details
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())  // User needs to fix their input
        return
    }
    
    // Continue with business logic...
}
```

**Why Safe**: 
- Error originates from user's own input
- Helps user correct their mistakes
- No system information revealed
- Standard HTTP 400 Bad Request

**Example Error Messages** (Safe):
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Key: 'CreateOrderDTO.CustomerID' Error:Field validation for 'CustomerID' failed on the 'required' tag"
  }
}
```

### System Errors - Generic Messages

**When**: Internal operations fail (database errors, service errors, business logic errors)

**Pattern**:
```go
func (h *Handler) Create(c *gin.Context) {
    var req CreateDTO
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    
    // ✅ CORRECT - Generic message for system errors
    order, err := h.usecase.CreateOrder(c.Request.Context(), req.CustomerID, req.Currency)
    if err != nil {
        response.InternalError(c, "Failed to create order")  // Hide implementation details
        return
    }
    
    response.Success(c, order)
}
```

**Why Secure**:
- Hides internal implementation details
- Prevents entity enumeration attacks
- No information about system structure
- Standard HTTP 500 Internal Server Error

**Example Error Messages** (Secure):
```json
{
  "status": "error",
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to create order"
  }
}
```

---

## Anti-Patterns (What NOT to Do)

### ❌ Anti-Pattern 1: String Comparison on System Errors

**Problem**: Checking error text to determine response type

```go
// ❌ WRONG - Security risk!
order, err := h.usecase.CreateOrder(c.Request.Context(), req.CustomerID, req.Currency)
if err != nil {
    if err.Error() == "not found" {  // ⚠️ INFORMATION LEAKAGE
        response.NotFound(c, "Order not found")
        return
    }
    if err.Error() == "already exists" {  // ⚠️ ENTITY ENUMERATION
        response.BadRequest(c, "Order already exists")
        return
    }
    response.InternalError(c, "Failed to create order")
    return
}
```

**Security Risks**:
1. **Entity Enumeration**: Attacker can determine if entities exist
2. **Brittle Code**: Breaks if error text changes
3. **Information Leakage**: Reveals internal error handling logic
4. **Reconnaissance**: Attacker learns about system structure

**Solution**: Remove string comparisons, use generic messages

```go
// ✅ CORRECT - Security hardened
order, err := h.usecase.CreateOrder(c.Request.Context(), req.CustomerID, req.Currency)
if err != nil {
    response.InternalError(c, "Failed to create order")  // Generic message always
    return
}
```

### ❌ Anti-Pattern 2: Exposing Specific Error Messages

**Problem**: Returning specific system error messages to client

```go
// ❌ WRONG - Exposes internal details
customer, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
if err != nil {
    response.InternalError(c, err.Error())  // ⚠️ Exposes implementation
    return
}
```

**What Gets Leaked**:
```json
{
  "error": {
    "message": "pq: relation 'customers' does not exist"  // ⚠️ Database structure
  }
}
```

or

```json
{
  "error": {
    "message": "customer not found in cache, fallback failed: redis: connection refused"  // ⚠️ Architecture details
  }
}
```

**Solution**: Always use generic messages

```go
// ✅ CORRECT
customer, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
if err != nil {
    response.InternalError(c, "Failed to retrieve customer")  // Safe
    return
}
```

### ❌ Anti-Pattern 3: Different Messages for Different Error Types

**Problem**: Different error messages reveal information about system state

```go
// ❌ WRONG - Entity enumeration attack vector
order, err := h.usecase.UpdateOrder(c.Request.Context(), orderID, req)
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        response.NotFound(c, "Order not found")  // ⚠️ Confirms non-existence
        return
    }
    if strings.Contains(err.Error(), "permission denied") {
        response.Forbidden(c, "Permission denied")  // ⚠️ Confirms existence
        return
    }
    response.InternalError(c, "Failed to update order")
    return
}
```

**Attack Scenario**:
1. Attacker tries updating order with random IDs
2. "Order not found" → Order doesn't exist
3. "Permission denied" → Order exists but no access
4. Attacker maps valid order IDs in the system

**Solution**: Consistent generic message for all cases

```go
// ✅ CORRECT - No information leakage
order, err := h.usecase.UpdateOrder(c.Request.Context(), orderID, req)
if err != nil {
    response.InternalError(c, "Failed to update order")  // Same message always
    return
}
```

---

## Context-Specific Patterns

### Identity Context

**Example**: User Handler

```go
func (h *UserHandler) GetByID(c *gin.Context) {
    userID, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "Invalid user ID format")  // Validation error - safe
        return
    }
    
    user, err := h.usecase.GetUser(c.Request.Context(), userID)
    if err != nil {
        response.InternalError(c, "Failed to retrieve user")  // System error - generic
        return
    }
    
    response.Success(c, toUserDTO(user))
}
```

### Customer Management Context

**Example**: Customer Handler

```go
func (h *CustomerHandler) QualifyAsProspect(c *gin.Context) {
    customerID, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "Invalid customer ID format")  // Validation - safe
        return
    }
    
    err = h.usecase.QualifyAsProspect(c.Request.Context(), customerID)
    if err != nil {
        response.InternalError(c, "Failed to qualify customer")  // System - generic
        return
    }
    
    response.Success(c, gin.H{"message": "Customer qualified as prospect"})
}
```

### Order Management Context

**Example**: Order Handler

```go
func (h *OrderHandler) Confirm(c *gin.Context) {
    orderID, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "Invalid order ID format")  // Validation - safe
        return
    }
    
    err = h.usecase.ConfirmOrder(c.Request.Context(), orderID)
    if err != nil {
        response.InternalError(c, "Failed to confirm order")  // System - generic
        return
    }
    
    response.Success(c, gin.H{"message": "Order confirmed successfully"})
}
```

---

## Response Helpers

**Standard response helpers** in `pkg/response/response.go`:

```go
// For validation errors - expose details
func BadRequest(c *gin.Context, message string) {
    Error(c, http.StatusBadRequest, "VALIDATION_ERROR", message)
}

// For system errors - generic messages only
func InternalError(c *gin.Context, message string) {
    Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func NotFound(c *gin.Context, message string) {
    Error(c, http.StatusNotFound, "NOT_FOUND", message)
}
```

**Usage Guidelines**:
- `BadRequest()`: Only for `ShouldBindJSON()` errors
- `InternalError()`: For all usecase/repository errors
- `NotFound()`: Only with generic messages (never specific entity details)

---

## Testing Security Patterns

### Unit Tests

**Verify generic error handling**:

```go
func TestHandler_Create_SystemError(t *testing.T) {
    mockUC := &MockUseCase{
        CreateFunc: func(ctx context.Context, ...) (*Entity, error) {
            return nil, errors.New("database connection failed")  // Specific error
        },
    }
    
    handler := NewHandler(mockUC)
    router := gin.New()
    router.POST("/entities", handler.Create)
    
    resp := makeRequest(t, router, "POST", "/entities", validBody)
    
    // Assert generic message returned
    assert.Equal(t, 500, resp.Code)
    assert.Contains(t, resp.Body.String(), "Failed to create entity")
    assert.NotContains(t, resp.Body.String(), "database")  // ✅ Implementation hidden
}
```

### Smoke Tests

**Mock all usecase methods** to test handler responses:

```go
func TestHandler_GetByID_NotFound(t *testing.T) {
    mockUC := &MockUseCase{
        GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*Entity, error) {
            return nil, errors.New("entity not found in database")  // Specific
        },
    }
    
    handler := NewHandler(mockUC)
    router := gin.New()
    router.GET("/entities/:id", handler.GetByID)
    
    resp := makeRequest(t, router, "GET", "/entities/"+testUUID, nil)
    
    // Assert generic error
    assert.Equal(t, 500, resp.Code)
    assert.Contains(t, resp.Body.String(), "Failed to retrieve entity")
    assert.NotContains(t, resp.Body.String(), "database")  // ✅ Safe
}
```

---

## Code Review Checklist

**Before merging any HTTP handler changes**:

- [ ] All `ShouldBindJSON()` errors use `response.BadRequest(c, err.Error())`
- [ ] All usecase errors use generic messages (`"Failed to..."`)
- [ ] No `err.Error()` string comparisons for system errors
- [ ] No `strings.Contains(err.Error(), ...)` checks
- [ ] No specific error messages in `response.InternalError()`
- [ ] Validation errors (user input) are detailed
- [ ] System errors are generic
- [ ] Tests verify generic error messages

---

## Migration Guide

**For existing handlers that need security hardening**:

### Step 1: Identify Anti-Patterns

Search for problematic patterns:

```bash
# Find string comparisons on errors
grep -r 'err.Error() ==' internal/contexts/*/adapter/http/handler/

# Find specific error messages
grep -r 'response.InternalError.*err.Error()' internal/contexts/
```

### Step 2: Apply Fixes

**Before** (insecure):
```go
customer, err := h.usecase.GetCustomer(ctx, id)
if err != nil {
    if err.Error() == "not found" {
        response.NotFound(c, "Customer not found")
        return
    }
    response.InternalError(c, err.Error())
    return
}
```

**After** (secure):
```go
customer, err := h.usecase.GetCustomer(ctx, id)
if err != nil {
    response.InternalError(c, "Failed to retrieve customer")
    return
}
```

### Step 3: Validate Changes

```bash
# Run linter
make ci-lint

# Run tests
make test-all

# Verify no string comparisons remain
grep -r 'err.Error() ==' internal/contexts/*/adapter/http/handler/ | wc -l  # Should be 0
```

---

## Audit Statistics

**Comprehensive Security Audit** (January 2026):

| Metric | Value |
|--------|-------|
| Handlers Audited | 36/36 (100%) |
| Security Fixes Applied | 417 |
| Sessions Completed | 18 |
| Contexts Covered | 7 (Identity, Customer-Mgmt, Order-Mgmt, Billing, Warehouse, Scripting, Shared) |
| Code Quality | 0 lint issues |
| Test Coverage | 2465+ tests, 100% pass rate |
| Validation Rate | ~89% (validation errors preserved) |
| Final Commit | [4eac8c9](https://github.com/basilex/promenade/commit/4eac8c9) |

**Context Breakdown**:

| Context | Handlers | Fixes | Status |
|---------|----------|-------|--------|
| Identity | 5 | 40 | Complete |
| Customer Management | 5 | 9 | Complete |
| Order Management | 2 | 0 | Complete |
| Billing | 3 | 1 | Complete |
| Warehouse | 4 | 0 | Complete |
| Scripting | 1 | 5 | Complete |
| Shared | 4 | 0 | Complete |
| Various (Sessions 1-13) | 12 | 362 | Complete |

---

## Future Considerations

### Advanced Error Handling

**For specific use cases**, consider domain-specific error types:

```go
// Define domain errors in usecase layer
var (
    ErrEntityNotFound     = errors.New("entity not found")
    ErrPermissionDenied   = errors.New("permission denied")
    ErrInvalidState       = errors.New("invalid state transition")
)

// Check with errors.Is() in handler
entity, err := h.usecase.GetEntity(ctx, id)
if errors.Is(err, usecase.ErrEntityNotFound) {
    response.NotFound(c, "Entity not found")  // OK - using domain error
    return
}
if err != nil {
    response.InternalError(c, "Failed to retrieve entity")  // Generic fallback
    return
}
```

**Note**: This pattern is acceptable because:
- Errors are defined in domain layer (not strings)
- Type-safe comparison with `errors.Is()`
- Still avoids exposing implementation details

### Logging

**Always log specific errors** internally while returning generic messages:

```go
customer, err := h.usecase.UpdateCustomer(ctx, id, req)
if err != nil {
    // Log specific error for debugging
    logger.FromContext(ctx).Error("Failed to update customer",
        slog.String("customer_id", id.String()),
        slog.Any("error", err),  // Full error details in logs
    )
    
    // Return generic message to client
    response.InternalError(c, "Failed to update customer")  // Generic
    return
}
```

---

## Related Documentation

- [Main README](../../README.md) - Project overview
- [Testing Patterns](testing-patterns.md) - Test strategy
- [API Documentation](api-documentation.md) - API reference
- [RBAC Guide](rbac.md) - Authorization patterns

---

**Last Updated**: January 9, 2026  
**Status**: Production-ready security patterns  
**Audit**: 36 handlers, 417 fixes, 100% complete  
**Maintainer**: Promenade Security Team
