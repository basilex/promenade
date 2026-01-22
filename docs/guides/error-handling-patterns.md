# Error Handling Patterns

## Overview

This guide defines best practices for domain error handling in Promenade. Proper error design improves debugging, security, and user experience.

## Core Principles

### 1. Error Types Classification

** GOOD: Domain-Specific Errors** — Represent business rules and domain concepts:

```go
// Validation errors - input validation
ErrCustomerNameEmpty = errors.New("customer name cannot be empty")
ErrCustomerEmailInvalid = errors.New("invalid customer email")

// Business logic errors - domain rule violations
ErrInvalidStatusTransition = errors.New("invalid status transition")
ErrPaymentAlreadyCompleted = errors.New("payment already completed")
ErrRefundAmountExceedsPayment = errors.New("refund amount exceeds payment")

// Not found errors - entity lookup failures
ErrCustomerNotFound = errors.New("customer not found")
ErrInvoiceNotFound = errors.New("invoice not found")

// Already exists errors - uniqueness violations
ErrCustomerAlreadyExists = errors.New("customer already exists")
```

** BAD: Generic System Errors** — No semantic value, just wrap underlying errors:

```go
// These should be REMOVED - they add no domain context
ErrCustomerCreateFailed = errors.New("failed to create customer")
ErrCustomerUpdateFailed = errors.New("failed to update customer")
ErrCustomerGetFailed = errors.New("failed to retrieve customer")
ErrCustomerDeleteFailed = errors.New("failed to delete customer")
```

**Why avoid generic errors?**

- They provide no business context
- They duplicate information from underlying errors
- They create false expectations (handlers can't discriminate them from system errors)
- They clutter the codebase

### 2. Error Wrapping Best Practices

When repository/database operations fail, wrap the error with business context using `fmt.Errorf`:

```go
//  BAD: Return generic constant
customer, err := r.repo.Create(ctx, customer)
if err != nil {
    return nil, ErrCustomerCreateFailed  // No context!
}

//  GOOD: Wrap with business context
customer, err := r.repo.Create(ctx, customer)
if err != nil {
    return nil, fmt.Errorf("create customer: %w", err)
}

//  EVEN BETTER: Add business identifiers
customer, err := r.repo.Create(ctx, customer)
if err != nil {
    return nil, fmt.Errorf("create customer %s: %w", customer.Email, err)
}
```

### 3. Three-Layer Error Handling Pattern

**Layer 1: Repository (Infrastructure)**

- Returns raw database errors or wrapped system errors
- Example: `sql.ErrNoRows`, `fmt.Errorf("query customer: %w", err)`

**Layer 2: Use Case (Application)**

- Translates infrastructure errors to domain errors
- Discriminates error types and returns domain constants
- Wraps unexpected errors with business context

**Layer 3: HTTP Handler (Presentation)**

- Discriminates domain errors with `errors.Is()`
- Maps to HTTP status codes
- Sanitizes error messages for security

```go
// LAYER 2: Use Case - Error translation
func (uc *CustomerUseCase) GetByID(ctx context.Context, id string) (*aggregate.Customer, error) {
    customer, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        // Discriminate specific cases
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrCustomerNotFound  // Domain error
        }
        // Wrap unexpected system errors
        return nil, fmt.Errorf("get customer %s: %w", id, err)
    }
    return customer, nil
}

// LAYER 3: HTTP Handler - HTTP mapping
func (h *CustomerHandler) Get(c echo.Context) error {
    customer, err := h.uc.GetByID(c.Request().Context(), id)
    if err != nil {
        // Discriminate domain errors
        if errors.Is(err, customer.ErrCustomerNotFound) {
            return response.NotFound(c, "customer not found")
        }
        // Hide system error details (security)
        slog.Error("failed to get customer", slog.String("id", id), slog.String("error", err.Error()))
        return response.InternalServerError(c, "operation failed")
    }
    return response.OK(c, dto.FromAggregate(customer))
}
```

## Migration Guide

### Step 1: Audit Existing Errors

Review all `errors.go` files and categorize:

-  **Keep**: Domain-specific errors (validation, business rules, not-found, already-exists)
-  **Remove**: Generic system errors (`Err*CreateFailed`, `Err*UpdateFailed`, `Err*GetFailed`)

### Step 2: Update Use Cases

Replace generic error constants with proper error wrapping:

```go
// BEFORE
func (uc *InvoiceUseCase) Create(ctx context.Context, req dto.CreateInvoiceRequest) (*aggregate.Invoice, error) {
    invoice, err := uc.repo.Create(ctx, invoice)
    if err != nil {
        return nil, ErrInvoiceCreateFailed  // Generic!
    }
    return invoice, nil
}

// AFTER
func (uc *InvoiceUseCase) Create(ctx context.Context, req dto.CreateInvoiceRequest) (*aggregate.Invoice, error) {
    invoice, err := uc.repo.Create(ctx, invoice)
    if err != nil {
        return nil, fmt.Errorf("create invoice for customer %s: %w", invoice.CustomerID, err)
    }
    return invoice, nil
}
```

### Step 3: Update Error Discrimination

Use `errors.Is()` for specific checks, fall back to generic for system errors:

```go
// BEFORE
if errors.Is(err, ErrCustomerGetFailed) {  // This is useless!
    return response.InternalServerError(c, "failed to get customer")
}

// AFTER
if errors.Is(err, customer.ErrCustomerNotFound) {
    return response.NotFound(c, "customer not found")
}
// For unexpected errors, log and hide details
slog.Error("customer operation failed", slog.String("error", err.Error()))
return response.InternalServerError(c, "operation failed")
```

### Step 4: Verify Test Cases

Update tests that check for generic errors:

```go
// BEFORE
assert.ErrorIs(t, err, ErrCustomerCreateFailed)

// AFTER - Check specific domain error or error message pattern
if err != nil && !errors.Is(err, customer.ErrCustomerAlreadyExists) {
    // Or check error contains expected context
    assert.Contains(t, err.Error(), "create customer")
}
```

## Examples by Context

### Customer Management Context

**errors.go** (cleaned):

```go
package customer

import "errors"

// Validation Errors
var (
    ErrCustomerNameEmpty      = errors.New("customer name cannot be empty")
    ErrCustomerEmailInvalid   = errors.New("invalid customer email")
    ErrCustomerPhoneInvalid   = errors.New("invalid customer phone")
)

// Not Found Errors
var (
    ErrCustomerNotFound = errors.New("customer not found")
)

// Business Logic Errors
var (
    ErrCustomerAlreadyExists     = errors.New("customer already exists")
    ErrInvalidStatusTransition   = errors.New("invalid status transition")
    ErrCannotActivateChurnedCustomer = errors.New("cannot activate churned customer")
)

// REMOVED: All generic errors (ErrCustomerCreateFailed, ErrCustomerUpdateFailed, etc.)
```

### Billing Payment Context

**errors.go** (cleaned):

```go
package payment

import "errors"

// Validation Errors
var (
    ErrInvalidPaymentAmount      = errors.New("invalid payment amount")
    ErrPaymentMethodRequired     = errors.New("payment method is required")
    ErrTransactionIDRequired     = errors.New("transaction ID is required")
    ErrRefundAmountExceedsPayment = errors.New("refund amount exceeds payment amount")
)

// Not Found Errors
var (
    ErrPaymentNotFound = errors.New("payment not found")
)

// Business Logic Errors
var (
    ErrPaymentAlreadyProcessed  = errors.New("payment already processed")
    ErrPaymentAlreadyCompleted  = errors.New("payment already completed")
    ErrPaymentAlreadyRefunded   = errors.New("payment already refunded")
    ErrPaymentCancelled         = errors.New("payment is cancelled")
    ErrPaymentFailed            = errors.New("payment processing failed")
)

// REMOVED: All generic errors (ErrPaymentCreateFailed, ErrPaymentProcessingFailed, etc.)
```

## Security Considerations

### Never Expose Internal Errors in HTTP Responses

```go
//  DANGEROUS: Exposes stack traces, SQL errors, file paths
return c.JSON(http.StatusInternalServerError, map[string]string{
    "error": err.Error(),  // Could expose: "pq: duplicate key value violates unique constraint..."
})

//  SAFE: Generic message + detailed logging
if err != nil {
    slog.Error("customer creation failed",
        slog.String("email", req.Email),
        slog.String("error", err.Error()))
    return response.InternalServerError(c, "operation failed")
}
```

### Log Context, Not Secrets

```go
//  BAD: Logs sensitive data
slog.Error("payment failed", slog.String("card_number", req.CardNumber))

//  GOOD: Logs identifiers only
slog.Error("payment failed",
    slog.String("payment_id", payment.ID),
    slog.String("customer_id", payment.CustomerID),
    slog.String("error", err.Error()))
```

## Testing Guidelines

### Test Domain Errors

```go
func TestCreateCustomer_AlreadyExists(t *testing.T) {
    // ... setup ...

    _, err := uc.Create(ctx, dto.CreateCustomerRequest{Email: existingEmail})

    // Check specific domain error
    assert.ErrorIs(t, err, customer.ErrCustomerAlreadyExists)
}
```

### Test Error Wrapping Context

```go
func TestCreateCustomer_RepositoryFailure(t *testing.T) {
    mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
        Return(nil, fmt.Errorf("database connection failed"))

    _, err := uc.Create(ctx, dto.CreateCustomerRequest{Email: "test@example.com"})

    // Verify error contains business context
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "create customer")
    assert.Contains(t, err.Error(), "test@example.com")
}
```

## Action Items

### High Priority

- [ ] Remove all generic `Err*CreateFailed`, `Err*UpdateFailed`, `Err*GetFailed` constants
- [ ] Update use cases to use `fmt.Errorf` wrapping with business context
- [ ] Audit HTTP handlers for error detail exposure

### Medium Priority

- [ ] Document domain-specific error scenarios for each context
- [ ] Add integration tests for error HTTP status code mapping
- [ ] Review logging statements for sensitive data leaks

### Low Priority

- [ ] Create error catalog documentation for API consumers
- [ ] Add error code constants for client-side error discrimination
- [ ] Implement error metrics/monitoring for production troubleshooting

## References

- [Security Patterns](security-patterns.md) - Error message sanitization
- [Testing Guidelines](../reference/testing-guidelines.md) - Error test patterns
- Go Blog: [Error Handling and Go](https://blog.golang.org/error-handling-and-go)
- Go Blog: [Working with Errors in Go 1.13](https://blog.golang.org/go1.13-errors)
