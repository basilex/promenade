# Context & Aggregate Naming Conventions

**Navigation**: [Home](../../README.md) > [Docs](../INDEX.md) > [Guides](README.md) > Naming Conventions

---

## Why This Standard Exists

**Problem**: Inconsistent structure across contexts leads to:

- Confusion when navigating codebase
- Different patterns in different contexts (company vs analytics vs banking)
- Difficulty onboarding new developers
- Hard to maintain and scale

**Solution**: Unified structure based on **DDD + Clean Architecture** principles with strict naming conventions.

---

## Core Principles

### 1. Bounded Context Structure

Each bounded context is a **business domain** with clear boundaries:

```
/internal/contexts/{context-name}/
```

Examples:

- `banking` (not `bank-management`)
- `customer-mgmt` (not `customer-management` - keep short)
- `order-mgmt` (not `orders`)
- `identity` (not `auth` or `user-management`)

**Rule**: Use kebab-case, short but descriptive names

---

### 2. Aggregate-Centric Organization

Each context contains **aggregates** (DDD aggregate roots). Structure is organized by aggregate, not by layer:

```
/context-name/
  /aggregate-name/           WRONG - we don't do this

/context-name/
  /aggregate/                CORRECT - all aggregates in one folder
    account.go
    transaction.go
```

---

## Standard Directory Structure

### Level 1: Context Root

```
/internal/contexts/{context-name}/
  README.md                 # Context documentation (domain, aggregates, use cases)
  errors.go                 # Domain error constants (ErrNotFound, ErrInvalidStatus, etc.)

  /aggregate/               # Domain layer - aggregate roots
  /repository/              # Domain layer - repository interfaces
  /usecase/                 # Application layer - business use cases
  /dto/                     # Application layer - data transfer objects
  /adapter/                 # Infrastructure layer - implementations
  /integration/             # Integration layer - event handlers
```

### Level 2: Aggregate (Domain Layer)

```
/aggregate/
  {aggregate_name}.go                 # Aggregate root entity
  {aggregate_name}_test.go            # Unit tests
```

**Naming Rules**:

- File names: snake_case (e.g., `bank_account.go`, `sales_report.go`)
- Aggregate struct: PascalCase (e.g., `BankAccount`, `SalesReport`)
- One aggregate per file
- Tests in `*_test.go` files

**Example**:

```
/aggregate/
  bank_account.go           # type BankAccount struct { ... }
  bank_account_test.go      # TestBankAccount_Connect, TestBankAccount_Sync
  bank_transaction.go       # type BankTransaction struct { ... }
  bank_transaction_test.go
```

### Level 3: Repository Interfaces (Domain Layer)

```
/repository/
  {aggregate_name}_repository.go      # Repository interface
```

**Naming Rules**:

- Interface name: `I{AggregateName}Repository` (e.g., `IBankAccountRepository`)
- Methods: CRUD + domain-specific queries
- No implementation details (pure interfaces)

**Example**:

```go
// repository/bank_account_repository.go
package repository

type IBankAccountRepository interface {
    Create(ctx context.Context, account *aggregate.BankAccount) error
    GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.BankAccount, error)
    GetByOrganization(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.BankAccount, error)
    Update(ctx context.Context, account *aggregate.BankAccount) error
    Delete(ctx context.Context, id uuidv7.UUID) error
}
```

### Level 4: Use Cases (Application Layer)

```
/usecase/
  {use_case_name}.go                  # Use case implementation
  {use_case_name}_test.go             # Use case tests
```

**Naming Rules**:

- File names: snake_case, verb-based (e.g., `connect_bank_account.go`, `sync_transactions.go`)
- Struct name: PascalCase (e.g., `ConnectBankAccountUseCase`)
- One use case per file
- Use case represents a **single business operation**

**Example**:

```
/usecase/
  connect_bank_account.go       # ConnectBankAccountUseCase
  connect_bank_account_test.go
  sync_transactions.go          # SyncTransactionsUseCase
  sync_transactions_test.go
  match_transaction.go          # MatchTransactionUseCase
  list_accounts.go              # ListAccountsUseCase
```

**Use Case Structure**:

```go
// usecase/connect_bank_account.go
package usecase

type ConnectBankAccountUseCase struct {
    repo repository.IBankAccountRepository
    bus  bus.EventBus
}

func NewConnectBankAccountUseCase(
    repo repository.IBankAccountRepository,
    bus bus.EventBus,
) *ConnectBankAccountUseCase {
    return &ConnectBankAccountUseCase{repo: repo, bus: bus}
}

func (uc *ConnectBankAccountUseCase) Execute(
    ctx context.Context,
    req *dto.ConnectBankAccountRequest,
) (*dto.BankAccountResponse, error) {
    // Business logic here
}
```

### Level 5: DTOs (Application Layer)

```
/dto/
  {aggregate_name}_dto.go             # Request/Response DTOs
  {aggregate_name}_dto_test.go        # DTO tests (if needed)
```

**Naming Rules**:

- File names: snake_case with `_dto.go` suffix
- Struct names: PascalCase with Request/Response suffix
- Group related DTOs in one file

**Example**:

```
/dto/
  bank_account_dto.go       # ConnectBankAccountRequest, BankAccountResponse
  bank_transaction_dto.go   # CreateTransactionRequest, TransactionResponse, ListTransactionsResponse
```

**DTO Structure**:

```go
// dto/bank_account_dto.go
package dto

type ConnectBankAccountRequest struct {
    OrganizationID  string `json:"organization_id" binding:"required"`
    Provider        string `json:"provider" binding:"required"`
    AccountNumber   string `json:"account_number" binding:"required"`
    BankName        string `json:"bank_name" binding:"required"`
}

type BankAccountResponse struct {
    ID              string    `json:"id"`
    OrganizationID  string    `json:"organization_id"`
    Provider        string    `json:"provider"`
    AccountNumber   string    `json:"account_number"`
    BankName        string    `json:"bank_name"`
    Status          string    `json:"status"`
    BalanceCents    int64     `json:"balance_cents"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

### Level 6: Adapters (Infrastructure Layer)

```
/adapter/
  /repository/
    /postgres/
      base_repository.go                      # Common DB operations
      {aggregate_name}_repository.go          # Implementation
      {aggregate_name}_repository_test.go     # Integration tests
  /http/
    {aggregate_name}_handler.go               # HTTP handlers
    {aggregate_name}_handler_test.go          # Handler tests
  /provider/                                  # External service adapters
    {provider_name}_provider.go
```

**Naming Rules**:

- Postgres repositories: snake_case with `_repository.go` suffix
- HTTP handlers: snake_case with `_handler.go` suffix
- Providers: snake_case with `_provider.go` suffix

**Example**:

```
/adapter/
  /repository/
    /postgres/
      base_repository.go
      bank_account_repository.go        # BankAccountRepository (implements IBankAccountRepository)
      bank_transaction_repository.go
  /http/
    bank_account_handler.go             # BankAccountHandler
    bank_transaction_handler.go
  /provider/
    monobank_provider.go                # MonobankProvider
    privat24_provider.go                # Privat24Provider
```

### Level 7: Integration (Event Handlers)

```
/integration/
  {event_handler_name}.go               # Event handler
  {event_handler_name}_test.go          # Handler tests
```

**Naming Rules**:

- File names: snake_case, event-based (e.g., `order_created_handler.go`)
- Struct name: PascalCase (e.g., `OrderCreatedHandler`)
- One handler per event type

**Example**:

```
/integration/
  order_created_handler.go          # OrderCreatedHandler - subscribes to "order.created"
  payment_completed_handler.go      # PaymentCompletedHandler - subscribes to "payment.completed"
```

---

## Complete Example: Banking Context

```
/internal/contexts/banking/
  README.md                                     # Banking context documentation
  errors.go                                     # Domain errors

  /aggregate/                                   # Domain layer
    bank_account.go                             # BankAccount aggregate
    bank_account_test.go
    bank_transaction.go                         # BankTransaction aggregate
    bank_transaction_test.go

  /repository/                                  # Domain layer
    bank_account_repository.go                  # IBankAccountRepository interface
    bank_transaction_repository.go              # IBankTransactionRepository interface

  /usecase/                                     # Application layer
    connect_bank_account.go                     # ConnectBankAccountUseCase
    connect_bank_account_test.go
    sync_transactions.go                        # SyncTransactionsUseCase
    sync_transactions_test.go
    match_transaction.go                        # MatchTransactionUseCase
    list_accounts.go                            # ListAccountsUseCase
    list_transactions.go                        # ListTransactionsUseCase

  /dto/                                         # Application layer
    bank_account_dto.go                         # Request/Response DTOs for accounts
    bank_transaction_dto.go                     # Request/Response DTOs for transactions

  /adapter/                                     # Infrastructure layer
    /repository/
      /postgres/
        base_repository.go                      # Common DB operations
        bank_account_repository.go              # BankAccountRepository
        bank_account_repository_test.go
        bank_transaction_repository.go          # BankTransactionRepository
        bank_transaction_repository_test.go
    /http/
      bank_account_handler.go                   # BankAccountHandler
      bank_account_handler_test.go
      bank_transaction_handler.go               # BankTransactionHandler
    /provider/
      bank_provider.go                          # IBankProvider interface
      monobank_provider.go                      # MonobankProvider
      privat24_provider.go                      # Privat24Provider
      pumb_provider.go                          # PUMBProvider

  /integration/                                 # Integration layer
    invoice_created_handler.go                  # InvoiceCreatedHandler
    payment_received_handler.go                 # PaymentReceivedHandler
```

---

## File Naming Rules Summary

| Layer          | Directory                       | File Pattern                     | Example                      |
| -------------- | ------------------------------- | -------------------------------- | ---------------------------- |
| Domain         | `/aggregate/`                   | `{aggregate_name}.go`            | `bank_account.go`            |
| Domain         | `/repository/`                  | `{aggregate_name}_repository.go` | `bank_account_repository.go` |
| Application    | `/usecase/`                     | `{use_case_name}.go`             | `connect_bank_account.go`    |
| Application    | `/dto/`                         | `{aggregate_name}_dto.go`        | `bank_account_dto.go`        |
| Infrastructure | `/adapter/repository/postgres/` | `{aggregate_name}_repository.go` | `bank_account_repository.go` |
| Infrastructure | `/adapter/http/`                | `{aggregate_name}_handler.go`    | `bank_account_handler.go`    |
| Infrastructure | `/adapter/provider/`            | `{provider_name}_provider.go`    | `monobank_provider.go`       |
| Integration    | `/integration/`                 | `{event}_handler.go`             | `order_created_handler.go`   |

**Global Rules**:

- All file names: **snake_case**
- All struct names: **PascalCase**
- All package names: **lowercase** (single word)
- Test files: `*_test.go` suffix

---

## Package Naming

Each directory is a package:

```
/aggregate/         → package aggregate
/repository/        → package repository
/usecase/           → package usecase
/dto/               → package dto
/adapter/http/      → package http
/adapter/repository/postgres/ → package postgres
/integration/       → package integration
```

**Rule**: Package name = directory name (single word, lowercase)

---

## Struct Naming Conventions

### Aggregates

```go
type BankAccount struct { ... }           
type bankAccount struct { ... }           
type BankAccountAggregate struct { ... }   (redundant)
```

### Repositories (Interfaces)

```go
type IBankAccountRepository interface { ... }     
type BankAccountRepository interface { ... }       (no 'I' prefix)
type BankAccountRepo interface { ... }             (no abbreviations)
```

### Repositories (Implementations)

```go
type BankAccountRepository struct { ... }          (implements IBankAccountRepository)
type PostgresBankAccountRepository struct { ... }  (redundant - in postgres package)
```

### Use Cases

```go
type ConnectBankAccountUseCase struct { ... }     
type ConnectBankAccount struct { ... }             (missing UseCase suffix)
type ConnectBankAccountService struct { ... }      (not a service)
```

### DTOs

```go
type ConnectBankAccountRequest struct { ... }     
type BankAccountResponse struct { ... }           
type BankAccountDTO struct { ... }                 (redundant - in dto package)
```

### Handlers

```go
type BankAccountHandler struct { ... }            
type BankAccountHTTPHandler struct { ... }         (redundant - in http package)
```

---

## Method Naming Conventions

### Aggregates (Domain Logic)

```go
func (a *BankAccount) Connect(provider Provider) error { ... }        
func (a *BankAccount) UpdateBalance(cents int64) { ... }              
func (a *BankAccount) Activate() error { ... }                        
func (a *BankAccount) GetBalance() int64 { ... }                       (getter)
func (a *BankAccount) IsActive() bool { ... }                          (predicate)
```

### Repositories (CRUD + Domain Queries)

```go
func (r *IBankAccountRepository) Create(ctx, account) error           
func (r *IBankAccountRepository) GetByID(ctx, id) (*Account, error)   
func (r *IBankAccountRepository) Update(ctx, account) error           
func (r *IBankAccountRepository) Delete(ctx, id) error                
func (r *IBankAccountRepository) GetByOrganization(ctx, orgID) ([]*Account, error) 
```

### Use Cases (Business Operations)

```go
func (uc *ConnectBankAccountUseCase) Execute(ctx, req) (*Response, error) 
```

**Rule**: Use cases always have `Execute` method (single responsibility)

### Handlers (HTTP Endpoints)

```go
func (h *BankAccountHandler) RegisterRoutes(router *gin.RouterGroup)  
func (h *BankAccountHandler) Create(c *gin.Context)                   
func (h *BankAccountHandler) GetByID(c *gin.Context)                  
func (h *BankAccountHandler) List(c *gin.Context)                     
func (h *BankAccountHandler) Update(c *gin.Context)                   
func (h *BankAccountHandler) Delete(c *gin.Context)                   
```

---

## Import Conventions

### Within Same Context

```go
// In usecase/connect_bank_account.go
import (
    "context"

    "github.com/basilex/promenade/internal/contexts/banking/aggregate"
    "github.com/basilex/promenade/internal/contexts/banking/dto"
    "github.com/basilex/promenade/internal/contexts/banking/repository"
)
```

### Cross-Context (Via Events Only!)

```go
//  WRONG - direct import from another context
import "github.com/basilex/promenade/internal/contexts/order-mgmt/aggregate"

//  CORRECT - via event bus
import "github.com/basilex/promenade/pkg/bus"

// Subscribe to order.created event
bus.Subscribe("order.created", handler.HandleOrderCreated)
```

---

## Migration from Old Structure

### Before (company - flat structure)

```
/company/
  entity.go
  entity_test.go
  usecase.go
  usecase_test.go
  repository.go
  errors.go
  adapter/http/handler.go
  adapter/http/dto.go
  adapter/repository/postgres/company_repository.go
```

### After (standardized)

```
/company/
  README.md
  errors.go

  /aggregate/
    company.go
    company_test.go

  /repository/
    company_repository.go           # Interface

  /usecase/
    create_company.go
    update_company.go
    list_companies.go
    get_company.go
    delete_company.go

  /dto/
    company_dto.go

  /adapter/
    /repository/postgres/
      base_repository.go
      company_repository.go         # Implementation
    /http/
      company_handler.go
```

---

## Special Cases

### Read Models (CQRS)

For analytics/reporting contexts with read models:

```
/analytics/
  /aggregate/              # If there are write models
  /read_model/             # Read-only projections
    sales_report.go
    customer_stats.go
  /usecase/
    get_sales_report.go    # Query use case
  /dto/
    sales_report_dto.go
```

### Sagas (Long-Running Processes)

```
/order-mgmt/
  /saga/
    fulfillment_saga.go
    fulfillment_saga_test.go
  /adapter/
    /saga/
      fulfillment_saga_impl.go
```

### Shared/Common Code

```
/shared/                   # NOT a bounded context
  /country/
  /currency/
  /language/
```

**Rule**: Shared contexts are **reference data** (no complex business logic)

---

## Benefits of This Standard

1. **Predictability**: Every context looks the same
2. **Navigation**: Easy to find files (know structure = know location)
3. **Onboarding**: New developers learn structure once
4. **Testing**: Clear separation makes testing easier
5. **Scalability**: Add new contexts/aggregates without confusion
6. **Maintenance**: Consistent patterns reduce cognitive load
7. **Refactoring**: Easy to move/rename/split aggregates

---

## Enforcement

### Pre-commit Checks

- [ ] TODO: Add linter to verify structure
- [ ] TODO: Add script to validate naming conventions

### Code Review Checklist

- [ ] Does new context follow standard structure?
- [ ] Are file names snake_case?
- [ ] Are struct names PascalCase?
- [ ] Are use cases in `/usecase/` with `UseCase` suffix?
- [ ] Are DTOs in `/dto/` with Request/Response suffix?
- [ ] Are repository interfaces in `/repository/`?
- [ ] Are repository implementations in `/adapter/repository/{db}/`?

---

## References

- [Bounded Contexts Strategy](../concepts/bounded-contexts.md)
- [Clean Architecture](../concepts/clean-architecture.md)
- [DDD Patterns](../concepts/ddd-patterns.md)
- [Context Refactoring Checklist](../refactoring/context-structure-refactoring-checklist.md)

---

**Last Updated**: January 17, 2026  
**Status**:  Approved Standard
