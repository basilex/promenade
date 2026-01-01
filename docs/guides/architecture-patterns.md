# Architecture Patterns Guide

**Comprehensive guide** to architectural patterns used in Promenade Platform: Repository, UseCase, Handler, and Value Object patterns.

---

## Table of Contents

- [Repository Pattern](#repository-pattern)
- [Use Case Pattern](#use-case-pattern)
- [Handler Pattern](#handler-pattern)
- [Value Object Pattern](#value-object-pattern)
- [Entity Pattern](#entity-pattern)
- [Complete Example](#complete-example)

---

## Repository Pattern

### Purpose

**Repository Pattern** abstracts data access logic, providing a collection-like interface for domain entities.

**Benefits**:
- Isolates domain layer from database implementation
- Enables easy testing with mocks
- Centralizes data access logic
- Supports multiple storage backends

### Structure

```
internal/contexts/{context}/{aggregate}/
  repository.go                    # Interface definition
  adapter/repository/postgres/
      base_repository.go          # Shared base repository
      {aggregate}_repository.go   # PostgreSQL implementation
```

### Interface Definition

**File**: `internal/contexts/customer-mgmt/customer/repository.go`

```go
package customer

import (
    "context"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the data access interface for Customer aggregate
type IRepository interface {
    // Standard CRUD operations
    Create(ctx context.Context, customer *Customer) error
    GetByID(ctx context.Context, id uuidv7.UUID) (*Customer, error)
    Update(ctx context.Context, customer *Customer) error
    Delete(ctx context.Context, id uuidv7.UUID) error
    
    // Query operations
    GetByEmail(ctx context.Context, email string) (*Customer, error)
    ListCustomers(ctx context.Context, filters ListFilters) ([]*Customer, int, error)
    
    // Business queries
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    CountByStatus(ctx context.Context, status CustomerStatus) (int, error)
}
```

### Implementation with BaseRepository

**File**: `internal/contexts/customer-mgmt/customer/adapter/repository/postgres/customer_repository.go`

```go
package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/jmoiron/sqlx"
    "github.com/lib/pq"
    
    "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// customerRepository implements IRepository using PostgreSQL
type customerRepository struct {
    *BaseRepository  // Embeds shared functionality
}

// NewCustomerRepository creates a new PostgreSQL repository
func NewCustomerRepository(db *sqlx.DB) customer.IRepository {
    return &customerRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

// Create inserts a new customer
func (r *customerRepository) Create(ctx context.Context, c *customer.Customer) error {
    query := `
        INSERT INTO customer_customers (
            id, email, name, status, tier, tags, 
            user_id, company_id, assigned_to,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
        )
    `
    
    _, err := r.Exec(ctx, query,
        c.ID,
        c.Email.Value(),
        c.Name,
        c.Status,
        c.Tier,
        pq.Array(c.Tags),
        c.UserID,
        c.CompanyID,
        c.AssignedTo,
        c.CreatedAt,
        c.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to create customer: %w", err)
    }
    
    return nil
}

// GetByID retrieves a customer by ID
func (r *customerRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*customer.Customer, error) {
    query := `
        SELECT id, email, name, status, tier, tags,
               user_id, company_id, assigned_to,
               created_at, updated_at
        FROM customer_customers
        WHERE id = $1 AND deleted_at IS NULL
    `
    
    var row customerRow
    if err := r.Get(ctx, &row, query, id); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, customer.ErrCustomerNotFound
        }
        return nil, fmt.Errorf("failed to get customer: %w", err)
    }
    
    return row.toEntity()
}

// Update modifies an existing customer
func (r *customerRepository) Update(ctx context.Context, c *customer.Customer) error {
    query := `
        UPDATE customer_customers
        SET email = $2,
            name = $3,
            status = $4,
            tier = $5,
            tags = $6,
            user_id = $7,
            company_id = $8,
            assigned_to = $9,
            updated_at = $10
        WHERE id = $1 AND deleted_at IS NULL
    `
    
    result, err := r.Exec(ctx, query,
        c.ID,
        c.Email.Value(),
        c.Name,
        c.Status,
        c.Tier,
        pq.Array(c.Tags),
        c.UserID,
        c.CompanyID,
        c.AssignedTo,
        c.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to update customer: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return customer.ErrCustomerNotFound
    }
    
    return nil
}

// Delete soft-deletes a customer
func (r *customerRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
    query := `
        UPDATE customer_customers
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND deleted_at IS NULL
    `
    
    result, err := r.Exec(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete customer: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return customer.ErrCustomerNotFound
    }
    
    return nil
}
```

### BaseRepository Pattern

**File**: `internal/contexts/customer-mgmt/customer/adapter/repository/postgres/base_repository.go`

```go
package postgres

import (
    "context"
    "database/sql"

    "github.com/jmoiron/sqlx"
    "github.com/basilex/promenade/internal/infrastructure/database"
)

// BaseRepository provides common database operations
type BaseRepository struct {
    db *sqlx.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
    return &BaseRepository{db: db}
}

// getExecutor returns the appropriate executor (transaction or database)
func (r *BaseRepository) getExecutor(ctx context.Context) database.Executor {
    if tx := database.GetTx(ctx); tx != nil {
        return tx
    }
    return r.db
}

// Get executes a query that returns a single row
func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
    return r.getExecutor(ctx).GetContext(ctx, dest, query, args...)
}

// Select executes a query that returns multiple rows
func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
    return r.getExecutor(ctx).SelectContext(ctx, dest, query, args...)
}

// Exec executes a query without returning rows
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query without returning rows
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
    return r.getExecutor(ctx).NamedExecContext(ctx, query, arg)
}
```

### Key Patterns

**Context Propagation**:
```go
// Always pass context as first parameter
func (r *customerRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Customer, error)

// getExecutor checks for transaction in context
executor := r.getExecutor(ctx)
```

**Error Handling**:
```go
// Convert sql.ErrNoRows to domain error
if errors.Is(err, sql.ErrNoRows) {
    return nil, customer.ErrCustomerNotFound
}

// Wrap errors with context
return nil, fmt.Errorf("failed to get customer: %w", err)
```

**Soft Delete**:
```go
// Always filter deleted records
WHERE id = $1 AND deleted_at IS NULL

// Soft delete updates deleted_at
UPDATE customer_customers
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
```

---

## Use Case Pattern

### Purpose

**Use Case Pattern** encapsulates business logic and orchestrates domain operations.

**Benefits**:
- Clear separation of concerns (business logic vs data access)
- Testable without database
- Reusable across different interfaces (HTTP, CLI, gRPC)
- Domain-centric design

### Structure

```
internal/contexts/{context}/{aggregate}/
  usecase.go        # Interface + implementation
  usecase_test.go   # Business logic tests
```

### Interface Definition

**File**: `internal/contexts/customer-mgmt/customer/usecase.go`

```go
package customer

import (
    "context"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines business operations for Customer aggregate
type IUseCase interface {
    // Creation
    CreateCustomer(ctx context.Context, email, name string, tier CustomerTier) (*Customer, error)
    
    // Retrieval
    GetCustomer(ctx context.Context, id uuidv7.UUID) (*Customer, error)
    ListCustomers(ctx context.Context, filters ListFilters) ([]*Customer, int, error)
    
    // Updates
    UpdateCustomerInfo(ctx context.Context, id uuidv7.UUID, name string) error
    UpdateCustomerTier(ctx context.Context, id uuidv7.UUID, tier CustomerTier) error
    
    // Business operations
    TransitionToProspect(ctx context.Context, id uuidv7.UUID) error
    TransitionToCustomer(ctx context.Context, id uuidv7.UUID) error
    ChurnCustomer(ctx context.Context, id uuidv7.UUID) error
    
    // Deletion
    DeleteCustomer(ctx context.Context, id uuidv7.UUID) error
}
```

### Implementation

```go
// useCase implements IUseCase
type useCase struct {
    repo IRepository
}

// NewUseCase creates a new use case instance
func NewUseCase(repo IRepository) IUseCase {
    return &useCase{repo: repo}
}

// CreateCustomer creates a new customer
func (uc *useCase) CreateCustomer(ctx context.Context, email, name string, tier CustomerTier) (*Customer, error) {
    // Business validation: check if email exists
    exists, err := uc.repo.ExistsByEmail(ctx, email)
    if err != nil {
        return nil, fmt.Errorf("failed to check email existence: %w", err)
    }
    if exists {
        return nil, ErrEmailAlreadyExists
    }
    
    // Create domain entity (factory method with validation)
    customer, err := NewCustomer(email, name, tier)
    if err != nil {
        return nil, fmt.Errorf("failed to create customer entity: %w", err)
    }
    
    // Persist to database
    if err := uc.repo.Create(ctx, customer); err != nil {
        return nil, fmt.Errorf("failed to save customer: %w", err)
    }
    
    return customer, nil
}

// TransitionToProspect transitions customer from Lead to Prospect
func (uc *useCase) TransitionToProspect(ctx context.Context, id uuidv7.UUID) error {
    // Retrieve entity
    customer, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to get customer: %w", err)
    }
    
    // Business logic: validate state transition
    if err := customer.TransitionToProspect(); err != nil {
        return err
    }
    
    // Persist changes
    if err := uc.repo.Update(ctx, customer); err != nil {
        return fmt.Errorf("failed to update customer: %w", err)
    }
    
    return nil
}
```

### Key Patterns

**Business Validation**:
```go
// Check uniqueness before creation
exists, err := uc.repo.ExistsByEmail(ctx, email)
if exists {
    return nil, ErrEmailAlreadyExists
}
```

**Entity Creation**:
```go
// Use factory methods for entity creation
customer, err := NewCustomer(email, name, tier)
if err != nil {
    return nil, err
}
```

**State Transitions**:
```go
// Delegate business logic to entity
if err := customer.TransitionToProspect(); err != nil {
    return err
}
```

**Error Wrapping**:
```go
// Add context to errors
return nil, fmt.Errorf("failed to save customer: %w", err)
```

---

## Handler Pattern

### Purpose

**Handler Pattern** manages HTTP requests and responses, delegating business logic to use cases.

**Benefits**:
- Clean separation between HTTP and business logic
- Standard error handling
- Consistent response format
- Easy to test with mocks

### Structure

```
internal/contexts/{context}/{aggregate}/adapter/http/handler/
  {aggregate}_handler.go        # HTTP handlers
  {aggregate}_handler_test.go   # Handler tests
  dto/
      {aggregate}_dto.go        # Request/Response DTOs
```

### Handler Implementation

**File**: `internal/contexts/customer-mgmt/customer/adapter/http/handler/customer_handler.go`

```go
package handler

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    
    "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
    "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/http/handler/dto"
    "github.com/basilex/promenade/pkg/response"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// CustomerHandler handles HTTP requests for Customer aggregate
type CustomerHandler struct {
    usecase customer.IUseCase
}

// NewCustomerHandler creates a new handler
func NewCustomerHandler(uc customer.IUseCase) *CustomerHandler {
    return &CustomerHandler{usecase: uc}
}

// Create handles POST /customers
func (h *CustomerHandler) Create(c *gin.Context) {
    // Bind request
    var req dto.CreateCustomerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }
    
    // Delegate to use case
    customer, err := h.usecase.CreateCustomer(
        c.Request.Context(),
        req.Email,
        req.Name,
        req.Tier,
    )
    
    // Handle domain errors
    if errors.Is(err, customer.ErrEmailAlreadyExists) {
        response.Error(c, http.StatusConflict, "EMAIL_EXISTS", err.Error())
        return
    }
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
        return
    }
    
    // Return success response
    response.Success(c, dto.ToCustomerResponse(customer))
}

// GetByID handles GET /customers/:id
func (h *CustomerHandler) GetByID(c *gin.Context) {
    // Parse ID parameter
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid customer ID")
        return
    }
    
    // Delegate to use case
    customer, err := h.usecase.GetCustomer(c.Request.Context(), id)
    
    // Handle domain errors
    if errors.Is(err, customer.ErrCustomerNotFound) {
        response.Error(c, http.StatusNotFound, "CUSTOMER_NOT_FOUND", err.Error())
        return
    }
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "GET_FAILED", err.Error())
        return
    }
    
    // Return success response
    response.Success(c, dto.ToCustomerResponse(customer))
}

// Update handles PUT /customers/:id
func (h *CustomerHandler) Update(c *gin.Context) {
    // Parse ID
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid customer ID")
        return
    }
    
    // Bind request
    var req dto.UpdateCustomerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }
    
    // Delegate to use case
    if err := h.usecase.UpdateCustomerInfo(c.Request.Context(), id, req.Name); err != nil {
        if errors.Is(err, customer.ErrCustomerNotFound) {
            response.Error(c, http.StatusNotFound, "CUSTOMER_NOT_FOUND", err.Error())
            return
        }
        response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
        return
    }
    
    response.Success(c, gin.H{"message": "Customer updated successfully"})
}
```

### DTOs (Data Transfer Objects)

**File**: `internal/contexts/customer-mgmt/customer/adapter/http/handler/dto/customer_dto.go`

```go
package dto

import (
    "time"
    
    "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// CreateCustomerRequest is the request body for creating a customer
type CreateCustomerRequest struct {
    Email string             `json:"email" binding:"required,email"`
    Name  string             `json:"name" binding:"required,min=1,max=100"`
    Tier  customer.CustomerTier `json:"tier" binding:"required,oneof=free basic pro enterprise"`
}

// UpdateCustomerRequest is the request body for updating a customer
type UpdateCustomerRequest struct {
    Name string `json:"name" binding:"required,min=1,max=100"`
}

// CustomerResponse is the response body for a customer
type CustomerResponse struct {
    ID        uuidv7.UUID           `json:"id"`
    Email     string                `json:"email"`
    Name      string                `json:"name"`
    Status    customer.CustomerStatus `json:"status"`
    Tier      customer.CustomerTier   `json:"tier"`
    Tags      []string              `json:"tags"`
    CreatedAt time.Time             `json:"created_at"`
    UpdatedAt time.Time             `json:"updated_at"`
}

// ToCustomerResponse converts entity to response DTO
func ToCustomerResponse(c *customer.Customer) *CustomerResponse {
    return &CustomerResponse{
        ID:        c.ID,
        Email:     c.Email.Value(),
        Name:      c.Name,
        Status:    c.Status,
        Tier:      c.Tier,
        Tags:      c.Tags,
        CreatedAt: c.CreatedAt,
        UpdatedAt: c.UpdatedAt,
    }
}
```

### Key Patterns

**Request Binding**:
```go
// Use Gin's ShouldBindJSON with validation tags
var req dto.CreateCustomerRequest
if err := c.ShouldBindJSON(&req); err != nil {
    response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
    return
}
```

**Domain Error Mapping**:
```go
// Map domain errors to HTTP status codes
if errors.Is(err, customer.ErrCustomerNotFound) {
    response.Error(c, http.StatusNotFound, "CUSTOMER_NOT_FOUND", err.Error())
    return
}
if errors.Is(err, customer.ErrEmailAlreadyExists) {
    response.Error(c, http.StatusConflict, "EMAIL_EXISTS", err.Error())
    return
}
```

**Response Helpers**:
```go
// Use standardized response helpers
response.Success(c, data)                    // 200 OK
response.Error(c, code, "ERROR_CODE", msg)   // Error with code
```

---

## Value Object Pattern

### Purpose

**Value Objects** are immutable objects defined by their attributes, not identity.

**Benefits**:
- Encapsulates validation logic
- Prevents invalid state
- Improves type safety
- Self-documenting code

### Examples

**Email Value Object**:

```go
package valueobject

import (
    "fmt"
    "regexp"
    "strings"
)

// Email is a value object for email addresses
type Email struct {
    value string
}

// NewEmail creates a new email with validation
func NewEmail(email string) (Email, error) {
    trimmed := strings.TrimSpace(strings.ToLower(email))
    
    if trimmed == "" {
        return Email{}, fmt.Errorf("email cannot be empty")
    }
    
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(trimmed) {
        return Email{}, fmt.Errorf("invalid email format")
    }
    
    return Email{value: trimmed}, nil
}

// MustNewEmail creates email or panics (for tests)
func MustNewEmail(email string) Email {
    e, err := NewEmail(email)
    if err != nil {
        panic(err)
    }
    return e
}

// Value returns the email string
func (e Email) Value() string {
    return e.value
}

// Equals checks equality with another email
func (e Email) Equals(other Email) bool {
    return e.value == other.value
}

// String implements Stringer interface
func (e Email) String() string {
    return e.value
}
```

**Money Value Object**:

```go
package valueobject

import "fmt"

// Money represents monetary value with currency
type Money struct {
    cents    int64  // Amount in cents (precision)
    currency string // ISO 4217 currency code
}

// NewMoney creates a new money value
func NewMoney(cents int64, currency string) (Money, error) {
    if len(currency) != 3 {
        return Money{}, fmt.Errorf("currency must be 3-letter ISO code")
    }
    
    return Money{
        cents:    cents,
        currency: strings.ToUpper(currency),
    }, nil
}

// Cents returns amount in cents
func (m Money) Cents() int64 {
    return m.cents
}

// Currency returns the currency code
func (m Money) Currency() string {
    return m.currency
}

// Add adds two money values (same currency)
func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, fmt.Errorf("cannot add different currencies")
    }
    
    return Money{
        cents:    m.cents + other.cents,
        currency: m.currency,
    }, nil
}

// Multiply multiplies money by a factor
func (m Money) Multiply(factor int) Money {
    return Money{
        cents:    m.cents * int64(factor),
        currency: m.currency,
    }
}

// IsZero checks if amount is zero
func (m Money) IsZero() bool {
    return m.cents == 0
}

// IsPositive checks if amount is positive
func (m Money) IsPositive() bool {
    return m.cents > 0
}
```

---

## Entity Pattern

### Purpose

**Entities** are domain objects with unique identity and lifecycle.

**Characteristics**:
- Have unique identifier (UUID v7)
- Contain business logic
- Enforce invariants
- Emit domain events

### Example

```go
package customer

import (
    "fmt"
    "time"
    
    "github.com/basilex/promenade/pkg/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/basilex/promenade/pkg/valueobject"
)

// Customer is an aggregate root
type Customer struct {
    aggregate.BaseAggregate
    
    // Identity
    ID uuidv7.UUID
    
    // Value Objects
    Email valueobject.Email
    
    // Attributes
    Name   string
    Status CustomerStatus
    Tier   CustomerTier
    Tags   []string
    
    // Relationships
    UserID     *uuidv7.UUID
    CompanyID  *uuidv7.UUID
    AssignedTo *uuidv7.UUID
    
    // Lifecycle
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}

// NewCustomer creates a new customer (factory method)
func NewCustomer(email, name string, tier CustomerTier) (*Customer, error) {
    // Validate email
    emailVO, err := valueobject.NewEmail(email)
    if err != nil {
        return nil, fmt.Errorf("invalid email: %w", err)
    }
    
    // Validate name
    if len(name) == 0 || len(name) > 100 {
        return nil, fmt.Errorf("name must be 1-100 characters")
    }
    
    // Validate tier
    if !tier.IsValid() {
        return nil, fmt.Errorf("invalid customer tier")
    }
    
    now := time.Now()
    return &Customer{
        BaseAggregate: aggregate.NewBase(),
        ID:            uuidv7.New(),
        Email:         emailVO,
        Name:          name,
        Status:        CustomerStatusLead,
        Tier:          tier,
        Tags:          []string{},
        CreatedAt:     now,
        UpdatedAt:     now,
    }, nil
}

// TransitionToProspect transitions from Lead to Prospect
func (c *Customer) TransitionToProspect() error {
    if c.Status != CustomerStatusLead {
        return fmt.Errorf("can only transition to prospect from lead status")
    }
    
    c.Status = CustomerStatusProspect
    c.UpdatedAt = time.Now()
    
    return nil
}

// TransitionToCustomer transitions to Customer status
func (c *Customer) TransitionToCustomer() error {
    if c.Status != CustomerStatusProspect {
        return fmt.Errorf("can only transition to customer from prospect status")
    }
    
    c.Status = CustomerStatusCustomer
    c.UpdatedAt = time.Now()
    
    return nil
}

// UpdateTier updates the customer tier
func (c *Customer) UpdateTier(tier CustomerTier) error {
    if !tier.IsValid() {
        return fmt.Errorf("invalid customer tier")
    }
    
    c.Tier = tier
    c.UpdatedAt = time.Now()
    
    return nil
}

// AddTag adds a tag to the customer
func (c *Customer) AddTag(tag string) {
    // Check if tag already exists
    for _, t := range c.Tags {
        if t == tag {
            return
        }
    }
    
    c.Tags = append(c.Tags, tag)
    c.UpdatedAt = time.Now()
}
```

---

## Complete Example

### Customer Aggregate (Full Flow)

**1. Entity** (`entity.go`):
```go
// Factory method
func NewCustomer(email, name string, tier CustomerTier) (*Customer, error)

// Business methods
func (c *Customer) TransitionToProspect() error
func (c *Customer) AddTag(tag string)
```

**2. Repository Interface** (`repository.go`):
```go
type IRepository interface {
    Create(ctx context.Context, customer *Customer) error
    GetByID(ctx context.Context, id uuidv7.UUID) (*Customer, error)
}
```

**3. Repository Implementation** (`adapter/repository/postgres/customer_repository.go`):
```go
type customerRepository struct {
    *BaseRepository
}

func (r *customerRepository) Create(ctx context.Context, c *Customer) error {
    // SQL INSERT with BaseRepository.Exec
}
```

**4. Use Case** (`usecase.go`):
```go
type IUseCase interface {
    CreateCustomer(ctx context.Context, email, name string, tier CustomerTier) (*Customer, error)
}

func (uc *useCase) CreateCustomer(...) (*Customer, error) {
    // Business logic + repository calls
}
```

**5. Handler** (`adapter/http/handler/customer_handler.go`):
```go
func (h *CustomerHandler) Create(c *gin.Context) {
    // Bind request → Call use case → Return response
}
```

**6. Router** (`adapter/http/router.go`):
```go
customers := api.Group("/customers")
{
    customers.POST("", handler.Create)
    customers.GET("/:id", handler.GetByID)
}
```

---

## Summary

| Pattern           | Purpose                        | File Location                     |
|-------------------|--------------------------------|-----------------------------------|
| **Repository**    | Data access abstraction        | `{aggregate}/repository.go`       |
| **Use Case**      | Business logic orchestration   | `{aggregate}/usecase.go`          |
| **Handler**       | HTTP request handling          | `adapter/http/handler/{aggregate}_handler.go` |
| **Value Object**  | Immutable validated values     | `pkg/valueobject/{type}.go`       |
| **Entity**        | Domain objects with identity   | `{aggregate}/entity.go`           |

---

**See Also**:
- [Naming Conventions Guide](naming-conventions.md) - Files, directories, Go code
- [Database Conventions Guide](database-conventions.md) - Tables, columns, indexes
- [Testing Patterns](testing-patterns.md) - Test organization and strategy

---

**Last Updated**: January 1, 2026  
**Status**: Production Standard  
**Maintainer**: Promenade Team
