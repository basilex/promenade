# Naming Conventions Guide

**Comprehensive naming standards** for files, directories, Go code, and database objects in Promenade Platform.

---

## Table of Contents

- [File & Directory Naming](#file--directory-naming)
- [Go Code Naming](#go-code-naming)
- [Database Naming](#database-naming)
- [Examples](#examples)

---

## File & Directory Naming

### Directory Structure Pattern

**Per Aggregate Structure** (within each context):

Promenade uses a **simple, flat structure** for aggregates to minimize nesting and improve discoverability:

```
internal/contexts/{context}/{aggregate}/
  entity.go                     # Domain entity (aggregate root)
  entity_test.go                # Entity unit tests
  repository.go                 # IRepository interface
  usecase.go                    # IUseCase interface + implementation
  usecase_test.go               # Use case unit tests
  errors.go                     # Domain errors (optional)
  adapter/
      http/
         handler.go             # HTTP handlers
         dto.go                 # Data transfer objects
         dto_test.go            # DTO tests (optional)
      repository/postgres/
          base_repository.go    # BaseRepository (per context)
          {aggregate}_repository.go # PostgreSQL implementation
```

**Real Example - Company Aggregate**:

```bash
$ tree internal/contexts/customer-mgmt/company
internal/contexts/customer-mgmt/company/
├── entity.go
├── entity_test.go
├── repository.go
├── usecase.go
├── usecase_test.go
├── errors.go
└── adapter/
    ├── http/
    │   ├── handler.go
    │   ├── dto.go
    │   └── dto_test.go
    └── repository/postgres/
        ├── base_repository.go
        └── company_repository.go
```

**Why This Structure?**
- **Simple**: Fewer directories, easier navigation
- **Flat**: Files directly in `adapter/http/`, not nested in `handler/` and `dto/`
- **Consistent**: Same pattern across all 21 tables/aggregates
- **Scalable**: Works well for aggregates with 1-10 files

**When to Add Subdirectories?**

If an aggregate grows to have many files (10+ handlers, 20+ DTOs), consider:
```
adapter/http/
  handlers/
     customer_handler.go
     order_handler.go
  dto/
     customer_dto.go
     order_dto.go
```

**Current Status**: All aggregates use simple structure ✅

### File Naming Rules

| File Type         | Pattern                         | Example                      |
|-------------------|---------------------------------|------------------------------|
| **Entity**        | `entity.go`                     | `entity.go`                  |
| **Entity Tests**  | `entity_test.go`                | `entity_test.go`             |
| **Repository**    | `{aggregate}_repository.go`     | `customer_repository.go`     |
| **Repository Tests** | `{aggregate}_repository_test.go` | `customer_repository_test.go` |
| **Use Case**      | `usecase.go`                    | `usecase.go`                 |
| **Use Case Tests**| `usecase_test.go`               | `usecase_test.go`            |
| **Handler**       | `handler.go` (in `adapter/http/`) | `handler.go`               |
| **Handler Tests** | `handler_test.go` (optional)    | `handler_test.go`            |
| **DTOs**          | `dto.go` (in `adapter/http/`)   | `dto.go`                     |
| **DTO Tests**     | `dto_test.go` (optional)        | `dto_test.go`                |
| **Errors**        | `errors.go` (optional)          | `errors.go`                  |
| **Migrations**    | `{number}_{name}.{up|down}.sql` | `000001_customers.up.sql`    |

**Key Rules**:
- Use `snake_case` for file names
- Entity and UseCase files are generic (`entity.go`, `usecase.go`)
- Repository files are aggregate-specific (`{aggregate}_repository.go`)
- Handler and DTO files are generic (`handler.go`, `dto.go`) in `adapter/http/`
- Test files mirror production files with `_test.go` suffix
- Migration files are numbered sequentially with descriptive name
- Errors file is optional (`errors.go`) for domain-specific errors

---

## Go Code Naming

### Interfaces

**Pattern**: `I{Entity}{Type}` with capital `I` prefix

```go
// Repository interface
type ICustomerRepository interface {
    Create(ctx context.Context, customer *Customer) error
    GetByID(ctx context.Context, id uuidv7.UUID) (*Customer, error)
    Update(ctx context.Context, customer *Customer) error
    Delete(ctx context.Context, id uuidv7.UUID) error
}

// Use case interface
type ICustomerUseCase interface {
    CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*Customer, error)
    GetCustomer(ctx context.Context, id uuidv7.UUID) (*Customer, error)
}
```

**Rules**:
- Always prefix with `I` for interfaces
- Use full words, no abbreviations (`IRepository`, NOT `IRepo`)
- Repository: `I{Entity}Repository`
- Use Case: `I{Entity}UseCase` (specific) OR `IUseCase` (generic per aggregate)

### Implementations (Structs)

**Pattern**: lowercase private structs

```go
// Repository implementation
type customerRepository struct {
    *BaseRepository
}

func NewCustomerRepository(db *sqlx.DB) ICustomerRepository {
    return &customerRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

// Use case implementation
type useCase struct {
    repo ICustomerRepository
}

func NewUseCase(repo ICustomerRepository) ICustomerUseCase {
    return &useCase{repo: repo}
}
```

**Rules**:
- Private structs: lowercase first letter
- Repository: `{aggregate}Repository` (e.g., `customerRepository`)
- Use Case: `useCase` (generic, NOT `customerUseCase`)
- Always return interface, not concrete type

### Constructors

**Pattern**: `New{Type}()` for public constructors

```go
// Repository constructor (aggregate-specific)
func NewCustomerRepository(db *sqlx.DB) ICustomerRepository {
    return &customerRepository{BaseRepository: NewBaseRepository(db)}
}

// Use case constructor (SIMPLE, generic)
func NewUseCase(repo ICustomerRepository) ICustomerUseCase {
    return &useCase{repo: repo}
}

// Handler constructor (aggregate-specific)
func NewCustomerHandler(uc ICustomerUseCase) *CustomerHandler {
    return &CustomerHandler{usecase: uc}
}

// Entity factory method
func NewCustomer(email, name string, tier CustomerTier) (*Customer, error) {
    emailVO, err := valueobject.NewEmail(email)
    if err != nil {
        return nil, err
    }
    return &Customer{
        ID:     uuidv7.New(),
        Email:  emailVO,
        Name:   name,
        Tier:   tier,
        Status: CustomerStatusLead,
    }, nil
}
```

**Rules**:
- Repository: `New{Entity}Repository()`
- Use Case: `NewUseCase()` (SIMPLE, no entity prefix)
- Handler: `New{Entity}Handler()`
- Entity: `New{Entity}()` (factory method)
- Value Object: `New{Type}()` (e.g., `NewEmail()`)

### Methods

**Repository Methods**:

```go
// Standard CRUD
func (r *customerRepository) Create(ctx context.Context, customer *Customer) error
func (r *customerRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Customer, error)
func (r *customerRepository) Update(ctx context.Context, customer *Customer) error
func (r *customerRepository) Delete(ctx context.Context, id uuidv7.UUID) error

// Query methods
func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*Customer, error)
func (r *customerRepository) ListCustomers(ctx context.Context, filters Filters) ([]*Customer, error)
func (r *customerRepository) CountByStatus(ctx context.Context, status CustomerStatus) (int, error)
```

**Use Case Methods**:

```go
// Business operations (imperative verbs)
func (uc *useCase) CreateCustomer(ctx context.Context, email, name string) (*Customer, error)
func (uc *useCase) GetCustomer(ctx context.Context, id uuidv7.UUID) (*Customer, error)
func (uc *useCase) UpdateCustomerTier(ctx context.Context, id uuidv7.UUID, tier CustomerTier) error
func (uc *useCase) DeleteCustomer(ctx context.Context, id uuidv7.UUID) error
func (uc *useCase) TransitionToProspect(ctx context.Context, id uuidv7.UUID) error
```

**Handler Methods**:

```go
// Match HTTP verbs
func (h *CustomerHandler) Create(c *gin.Context)
func (h *CustomerHandler) GetByID(c *gin.Context)
func (h *CustomerHandler) Update(c *gin.Context)
func (h *CustomerHandler) Delete(c *gin.Context)
func (h *CustomerHandler) List(c *gin.Context)
```

**Rules**:
- Repository: `GetByXxx`, `ListXxx`, `CountXxx`, `ExistsByXxx`
- Use Case: `{Verb}{Entity}`, `{BusinessOperation}`
- Handler: Match HTTP methods (`Create`, `GetByID`, `Update`, `Delete`, `List`)

### Variables & Parameters

**Naming Rules**:

```go
// Context always first parameter, always 'ctx'
func Method(ctx context.Context, id uuidv7.UUID) error

// Short names for common types
var (
    r    *customerRepository  // Repository receiver
    uc   *useCase             // Use case receiver
    h    *CustomerHandler     // Handler receiver
    req  CreateCustomerDTO    // Request
    resp CustomerResponse     // Response
    err  error                // Error
)

// Descriptive names for business entities
var (
    customer    *Customer
    customers   []*Customer
    email       string
    tier        CustomerTier
    status      CustomerStatus
)

// IDs with suffix
var (
    customerID uuidv7.UUID
    userID     uuidv7.UUID
    companyID  uuidv7.UUID
)
```

### Error Variables

**Pattern**: `Err{Entity}{Condition}`

```go
var (
    ErrCustomerNotFound      = errors.New("customer not found")
    ErrEmailAlreadyExists    = errors.New("email already exists")
    ErrInvalidStatus         = errors.New("invalid customer status")
    ErrUnauthorizedAccess    = errors.New("unauthorized access")
    ErrInvalidTierTransition = errors.New("invalid tier transition")
)

// Usage with wrapping
if errors.Is(err, ErrCustomerNotFound) {
    return nil, fmt.Errorf("failed to get customer: %w", err)
}
```

**Rules**:
- Prefix with `Err`
- Entity name + Condition
- Use `errors.Is()` for checking
- Wrap with `fmt.Errorf()` and `%w` for context

### Constants & Enums

**Pattern**: `{Entity}{Type}{Value}`

```go
// Status constants
const (
    CustomerStatusLead     CustomerStatus = "lead"
    CustomerStatusProspect CustomerStatus = "prospect"
    CustomerStatusCustomer CustomerStatus = "customer"
    CustomerStatusChurned  CustomerStatus = "churned"
)

// Tier constants
const (
    CustomerTierFree       CustomerTier = "free"
    CustomerTierBasic      CustomerTier = "basic"
    CustomerTierPro        CustomerTier = "pro"
    CustomerTierEnterprise CustomerTier = "enterprise"
)

// Database constants
const (
    DefaultPageSize = 20
    MaxPageSize     = 100
)
```

**Rules**:
- Entity name prefix for domain constants
- PascalCase for exported constants
- Group related constants with `const ()`

---

## Examples by Component

### Entity Example

```go
package customer

// Entity file: entity.go
type Customer struct {
    aggregate.BaseAggregate
    
    ID       uuidv7.UUID
    Email    valueobject.Email
    Name     string
    Status   CustomerStatus
    Tier     CustomerTier
    Tags     []string
    
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}

// Factory method
func NewCustomer(email, name string, tier CustomerTier) (*Customer, error) {
    emailVO, err := valueobject.NewEmail(email)
    if err != nil {
        return nil, err
    }
    
    return &Customer{
        BaseAggregate: aggregate.NewBase(),
        ID:            uuidv7.New(),
        Email:         emailVO,
        Name:          name,
        Status:        CustomerStatusLead,
        Tier:          tier,
        Tags:          []string{},
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }, nil
}
```

### Repository Example

```go
package postgres

// Repository file: customer_repository.go
type customerRepository struct {
    *BaseRepository
}

func NewCustomerRepository(db *sqlx.DB) customer.ICustomerRepository {
    return &customerRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

func (r *customerRepository) Create(ctx context.Context, c *customer.Customer) error {
    query := `
        INSERT INTO customer_customers (
            id, email, name, status, tier, tags, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8
        )
    `
    _, err := r.Exec(ctx, query, 
        c.ID, c.Email.Value(), c.Name, c.Status, c.Tier, 
        pq.Array(c.Tags), c.CreatedAt, c.UpdatedAt,
    )
    return err
}

func (r *customerRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*customer.Customer, error) {
    query := `
        SELECT id, email, name, status, tier, tags, created_at, updated_at
        FROM customer_customers
        WHERE id = $1 AND deleted_at IS NULL
    `
    var row customerRow
    if err := r.Get(ctx, &row, query, id); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, customer.ErrCustomerNotFound
        }
        return nil, err
    }
    return row.toEntity()
}
```

### Use Case Example

```go
package customer

// Use case file: usecase.go
type IUseCase interface {
    CreateCustomer(ctx context.Context, email, name string, tier CustomerTier) (*Customer, error)
    GetCustomer(ctx context.Context, id uuidv7.UUID) (*Customer, error)
    UpdateCustomerTier(ctx context.Context, id uuidv7.UUID, tier CustomerTier) error
    DeleteCustomer(ctx context.Context, id uuidv7.UUID) error
}

type useCase struct {
    repo IRepository
}

func NewUseCase(repo IRepository) IUseCase {
    return &useCase{repo: repo}
}

func (uc *useCase) CreateCustomer(ctx context.Context, email, name string, tier CustomerTier) (*Customer, error) {
    // Check if email exists
    exists, err := uc.repo.ExistsByEmail(ctx, email)
    if err != nil {
        return nil, fmt.Errorf("failed to check email existence: %w", err)
    }
    if exists {
        return nil, ErrEmailAlreadyExists
    }
    
    // Create customer entity
    customer, err := NewCustomer(email, name, tier)
    if err != nil {
        return nil, fmt.Errorf("failed to create customer entity: %w", err)
    }
    
    // Save to database
    if err := uc.repo.Create(ctx, customer); err != nil {
        return nil, fmt.Errorf("failed to save customer: %w", err)
    }
    
    return customer, nil
}
```

### Handler Example

```go
package handler

// Handler file: customer_handler.go
type CustomerHandler struct {
    usecase customer.IUseCase
}

func NewCustomerHandler(uc customer.IUseCase) *CustomerHandler {
    return &CustomerHandler{usecase: uc}
}

func (h *CustomerHandler) Create(c *gin.Context) {
    var req dto.CreateCustomerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }
    
    customer, err := h.usecase.CreateCustomer(c.Request.Context(), req.Email, req.Name, req.Tier)
    if errors.Is(err, customer.ErrEmailAlreadyExists) {
        response.Error(c, http.StatusConflict, "EMAIL_EXISTS", err.Error())
        return
    }
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
        return
    }
    
    response.Success(c, dto.ToCustomerResponse(customer))
}

func (h *CustomerHandler) GetByID(c *gin.Context) {
    id, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid customer ID")
        return
    }
    
    customer, err := h.usecase.GetCustomer(c.Request.Context(), id)
    if errors.Is(err, customer.ErrCustomerNotFound) {
        response.Error(c, http.StatusNotFound, "CUSTOMER_NOT_FOUND", err.Error())
        return
    }
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "GET_FAILED", err.Error())
        return
    }
    
    response.Success(c, dto.ToCustomerResponse(customer))
}
```

---

## Anti-Patterns to Avoid

### ❌ DON'T

```go
// DON'T: Abbreviations in interfaces
type ICustomerRepo interface {}      // Use ICustomerRepository

// DON'T: Public struct implementations
type CustomerRepository struct {}    // Use private customerRepository

// DON'T: Entity-specific UseCase constructor
func NewCustomerUseCase() {}         // Use simple NewUseCase()

// DON'T: Wrong method prefixes
func FindByID()                      // Use GetByID()
func FetchCustomers()                // Use ListCustomers()

// DON'T: Generic error names
var ErrNotFound                      // Use ErrCustomerNotFound

// DON'T: Mixed naming styles
type user_repository struct {}       // Use userRepository (camelCase)
```

### ✅ DO

```go
// DO: Full interface names
type ICustomerRepository interface {}

// DO: Private implementations
type customerRepository struct {}

// DO: Simple UseCase constructor
func NewUseCase() IUseCase {}

// DO: Standard method prefixes
func (r *customerRepository) GetByID()
func (r *customerRepository) ListCustomers()

// DO: Specific error names
var ErrCustomerNotFound

// DO: Consistent naming
type customerRepository struct {}
```

---

## Summary

| Component       | Pattern                          | Example                    |
|-----------------|----------------------------------|----------------------------|
| **Interface**   | `I{Entity}{Type}`                | `ICustomerRepository`      |
| **Struct**      | lowercase `{entity}{type}`       | `customerRepository`       |
| **Constructor** | `New{Entity}{Type}()` / `NewUseCase()` | `NewCustomerRepository()`, `NewUseCase()` |
| **Method**      | `{Verb}{Entity}` / `GetByXxx`    | `CreateCustomer`, `GetByEmail` |
| **Error**       | `Err{Entity}{Condition}`         | `ErrCustomerNotFound`      |
| **File**        | `{aggregate}_{type}.go`          | `customer_repository.go`   |
| **Directory**   | `{context}/{aggregate}/`         | `customer-mgmt/customer/`  |

---

**See Also**:
- [Database Conventions Guide](database-conventions.md) - Tables, columns, indexes
- [Architecture Patterns Guide](architecture-patterns.md) - Repository, UseCase, Handler patterns
- [Testing Patterns](testing-patterns.md) - Test organization and naming

---

**Last Updated**: January 1, 2026  
**Status**: Production Standard  
**Maintainer**: Promenade Team
