# Dependency Injection in Promenade

## Overview

Promenade uses a **custom, explicit dependency injection (DI) container** instead of traditional DI frameworks (Wire, Fx, Dig). This approach provides better control, clearer code, and natural alignment with Domain-Driven Design principles.

## Philosophy

### Why Custom DI?

**Traditional DI frameworks** like Wire, Fx, and Dig have their place, but they introduce complexity:

- **Wire**: Requires code generation, complex provider functions
- **Fx**: Uses reflection, complex lifecycle hooks, difficult debugging
- **Dig**: Runtime reflection, opaque dependency graph

**Custom DI in Promenade**:

- ✅ **Explicit** — Every dependency is clear and traceable
- ✅ **Type-safe** — Compile-time safety without reflection
- ✅ **Debuggable** — Step through initialization easily
- ✅ **Readable** — No magic, no generated code
- ✅ **DDD-aligned** — Groups dependencies by bounded context

## Architecture

### Entry Point

All domain dependencies are initialized through a single function in [cmd/api/dependencies.go](../../cmd/api/dependencies.go):

```go
func InitRepositories(
    db *sqlx.DB,
    cfg config.Config,
    eventBus bus.EventBus,
    cacheClient cache.Client,
) (*Dependencies, error)
```

### Dependencies Structure

Dependencies are **grouped by bounded context**, mirroring the DDD architecture:

```go
type Dependencies struct {
    Shared     *SharedUseCases        // Reference data (Country, Currency, etc.)
    Banking    *BankingUseCases       // Banking operations
    Accounting *AccountingUseCases    // Accounting ledger
    Warehouse  *WarehouseIntegration  // Inventory management
    Analytics  *AnalyticsComponents   // Reporting and analytics
}
```

Each context group contains:

- **Repositories** — Data access
- **Use cases** — Business logic
- **Event handlers** — Cross-context integration
- **HTTP handlers** — Presentation layer

## Implementation Pattern

### Context Initialization

Each bounded context has its own initialization function:

```go
func initSharedContext(
    db *sqlx.DB,
    cfg config.Config,
    eventBus bus.EventBus,
    cacheClient cache.Client,
) (*SharedUseCases, error) {
    // 1. Create repository (using factory for multi-database support)
    countryRepo, err := countryrepository.NewFactory(db, cfg.Database.Driver)
    if err != nil {
        return nil, fmt.Errorf("failed to create country repository: %w", err)
    }

    // 2. Create use case
    countryUseCase := countryusecase.New(countryRepo, eventBus)

    // 3. Create HTTP handler
    countryHTTPHandler := countryhttp.New(countryUseCase)

    // 4. Return grouped dependencies
    return &SharedUseCases{
        Country: struct {
            UseCase    *countryusecase.UseCase
            HTTPHandler *countryhttp.Handler
        }{
            UseCase:    countryUseCase,
            HTTPHandler: countryHTTPHandler,
        },
        // ... other shared aggregates (Currency, Language, Timezone)
    }, nil
}
```

### Repository Factory Pattern

For multi-database support, repositories use the **factory pattern**:

```go
// adapter/repository/factory.go (infrastructure layer)
func NewFactory(db *sqlx.DB, driver string) (country.Repository, error) {
    switch driver {
    case "postgres":
        return postgresrepo.New(db), nil
    case "mssql":
        return mssqlrepo.New(db), nil
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", driver)
    }
}
```

**Key principle**: Factory lives in **infrastructure layer** (`adapter/repository/`), not domain layer, to avoid circular dependencies.

## Usage

### In Bootstrap

```go
// cmd/api/bootstrap.go
func Bootstrap(cfgPath string) (*App, error) {
    // 1. Initialize infrastructure (logger, database, cache, event bus)
    logger := logger.New(cfg)
    db := database.NewPostgresConnection(cfg) // or NewMSSQLConnection
    eventBus := bus.NewMemoryBus(logger)
    cacheClient := cache.NewRedisClient(cfg)

    // 2. Initialize all domain dependencies (single call)
    deps, err := InitRepositories(db, cfg, eventBus, cacheClient)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize dependencies: %w", err)
    }

    // 3. Attach to application
    app := &App{
        Dependencies: deps,
        DB:          db,
        EventBus:    eventBus,
        Logger:      logger,
    }

    return app, nil
}
```

### In HTTP Server

```go
// cmd/api/server.go
func (app *App) setupRoutes() *echo.Echo {
    e := echo.New()

    // Access dependencies by context
    countryGroup := e.Group("/api/v1/countries")
    countryGroup.GET("", app.Dependencies.Shared.Country.HTTPHandler.List)
    countryGroup.GET("/:id", app.Dependencies.Shared.Country.HTTPHandler.GetByID)
    countryGroup.POST("", app.Dependencies.Shared.Country.HTTPHandler.Create)

    bankAccountGroup := e.Group("/api/v1/bank-accounts")
    bankAccountGroup.GET("", app.Dependencies.Banking.BankAccount.HTTPHandler.List)
    // ... etc.

    return e
}
```

## Benefits

### 1. Compile-Time Type Safety

No reflection means all errors are caught at compile time:

```go
// ✅ Compiler checks this
deps.Shared.Country.UseCase.Create(ctx, dto)

// ❌ Typo caught immediately
deps.Shared.Contry.UseCase.Create(ctx, dto)
```

### 2. IDE Autocomplete

Full IDE support with autocomplete, jump-to-definition, and refactoring:

```go
deps.                   // IDE shows: Shared, Banking, Accounting, Warehouse, Analytics
deps.Banking.           // IDE shows: BankAccount, BankTransaction
deps.Banking.BankAccount. // IDE shows: UseCase, HTTPHandler
```

### 3. Clear Dependency Graph

Dependencies are **explicit and traceable**:

```
InitRepositories()
├─ initSharedContext()
│  ├─ countryRepo = NewFactory(db, driver)
│  ├─ countryUseCase = New(countryRepo, eventBus)
│  └─ countryHTTPHandler = New(countryUseCase)
├─ initBankingContext()
│  ├─ bankAccountRepo = NewFactory(db, driver)
│  ├─ bankAccountUseCase = New(bankAccountRepo, eventBus)
│  └─ bankAccountHTTPHandler = New(bankAccountUseCase)
└─ initAccountingContext()
   └─ ... 7 aggregates
```

### 4. Easy Debugging

Step through initialization line by line:

```go
// Set breakpoint here to inspect initialization
deps, err := InitRepositories(db, cfg, eventBus, cacheClient)
if err != nil {
    // Clear error messages with stack trace
    return nil, fmt.Errorf("failed to initialize dependencies: %w", err)
}
```

### 5. Natural DDD Boundaries

Dependencies are grouped by **bounded context**, matching the domain architecture:

```
Shared Context      → deps.Shared.Country
Banking Context     → deps.Banking.BankAccount
Accounting Context  → deps.Accounting.JournalEntry
Warehouse Context   → deps.Warehouse.OrderEventHandler
```

## Comparison with Traditional DI

| Aspect             | Custom DI (Promenade)  | Wire                         | Fx/Dig                         |
| ------------------ | ---------------------- | ---------------------------- | ------------------------------ |
| **Implementation** | Explicit functions     | Code generation              | Runtime reflection             |
| **Type Safety**    | ✅ Compile-time        | ✅ Compile-time              | ✅ Compile-time (with caveats) |
| **Debugging**      | ✅ Step-through easily | ⚠️ Generated code complexity | ❌ Opaque dependency graph     |
| **IDE Support**    | ✅ Full autocomplete   | ⚠️ Limited (generated code)  | ⚠️ Limited (reflection)        |
| **Performance**    | ✅ Zero overhead       | ✅ Zero overhead             | ⚠️ Reflection overhead         |
| **Learning Curve** | ✅ Simple (plain Go)   | ⚠️ Provider syntax           | ⚠️ Lifecycle concepts          |
| **DDD Boundaries** | ✅ Natural grouping    | ❌ Flat structure            | ❌ Flat structure              |
| **Error Messages** | ✅ Clear and traceable | ⚠️ Generated code references | ❌ Runtime reflection errors   |
| **Testability**    | ✅ Mock any dependency | ✅ Mock dependencies         | ⚠️ Complex mocking             |
| **Refactoring**    | ✅ Safe with IDE       | ⚠️ Regenerate code           | ⚠️ Runtime discovery           |

## Adding New Dependencies

### Step 1: Create Context Init Function

```go
// cmd/api/dependencies.go

func initNewContext(
    db *sqlx.DB,
    cfg config.Config,
    eventBus bus.EventBus,
    cacheClient cache.Client,
) (*NewContextUseCases, error) {
    // 1. Repository
    repo, err := newrepository.NewFactory(db, cfg.Database.Driver)
    if err != nil {
        return nil, fmt.Errorf("failed to create new repository: %w", err)
    }

    // 2. Use case
    useCase := newusecase.New(repo, eventBus)

    // 3. HTTP handler
    httpHandler := newhttp.New(useCase)

    return &NewContextUseCases{
        New: struct {
            UseCase     *newusecase.UseCase
            HTTPHandler *newhttp.Handler
        }{
            UseCase:     useCase,
            HTTPHandler: httpHandler,
        },
    }, nil
}
```

### Step 2: Add to Dependencies Structure

```go
type Dependencies struct {
    Shared     *SharedUseCases
    Banking    *BankingUseCases
    Accounting *AccountingUseCases
    Warehouse  *WarehouseIntegration
    Analytics  *AnalyticsComponents
    NewContext *NewContextUseCases // ← Add here
}
```

### Step 3: Call in InitRepositories

```go
func InitRepositories(...) (*Dependencies, error) {
    // ... existing init calls

    newContextDeps, err := initNewContext(db, cfg, eventBus, cacheClient)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize new context: %w", err)
    }

    return &Dependencies{
        Shared:     sharedDeps,
        Banking:    bankingDeps,
        Accounting: accountingDeps,
        Warehouse:  warehouseDeps,
        Analytics:  analyticsDeps,
        NewContext: newContextDeps, // ← Add here
    }, nil
}
```

## Testing

### Mock Dependencies

Custom DI makes mocking trivial:

```go
// test/integration/contexts/customer-mgmt/customer/create_test.go
func TestCustomerCreate(t *testing.T) {
    // 1. Create mock repository
    mockRepo := &mocks.CustomerRepository{
        CreateFunc: func(ctx context.Context, c *aggregate.Customer) error {
            return nil
        },
    }

    // 2. Create use case with mock
    useCase := usecase.New(mockRepo, eventBus)

    // 3. Test
    err := useCase.Create(ctx, dto)
    assert.NoError(t, err)
}
```

### Integration Tests

Initialize only what you need:

```go
func TestBankingIntegration(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    eventBus := bus.NewMemoryBus(logger)

    // Initialize only Banking context
    bankingDeps, err := initBankingContext(db, cfg, eventBus, nil)
    require.NoError(t, err)

    // Test
    err = bankingDeps.BankAccount.UseCase.Create(ctx, dto)
    assert.NoError(t, err)
}
```

## Best Practices

### 1. Keep Init Functions Small

Each context init function should be < 100 lines:

```go
// ✅ Good: focused and readable
func initSharedContext(...) (*SharedUseCases, error) {
    // Initialize 3-5 aggregates
}

// ❌ Bad: too many responsibilities
func initEverything(...) (*AllDependencies, error) {
    // Initialize 20+ aggregates
}
```

### 2. Group by Bounded Context

Match DDD architecture:

```go
// ✅ Good: grouped by domain
type Dependencies struct {
    Shared     *SharedUseCases
    Banking    *BankingUseCases
    Accounting *AccountingUseCases
}

// ❌ Bad: flat technical grouping
type Dependencies struct {
    Repositories map[string]interface{}
    UseCases     map[string]interface{}
    Handlers     map[string]interface{}
}
```

### 3. Use Factory Pattern for Multi-Database

Keep domain layer database-agnostic:

```go
// ✅ Good: factory in infrastructure layer
countryRepo, err := countryrepository.NewFactory(db, cfg.Database.Driver)

// ❌ Bad: direct driver selection in domain
var countryRepo country.Repository
if driver == "postgres" {
    countryRepo = postgresrepo.New(db)
} else {
    countryRepo = mssqlrepo.New(db)
}
```

### 4. Return Errors with Context

Wrap errors to provide clear debugging information:

```go
// ✅ Good: wrapped errors with context
if err != nil {
    return nil, fmt.Errorf("failed to initialize banking context: %w", err)
}

// ❌ Bad: lost error context
if err != nil {
    return nil, err
}
```

## Related Documentation

- [cmd/api/dependencies.go](../../cmd/api/dependencies.go) — DI container implementation
- [cmd/api/bootstrap.go](../../cmd/api/bootstrap.go) — Application bootstrap
- [Repository Factory Pattern](./repository-factory.md) — Multi-database support
- [Bounded Contexts](../concepts/bounded-contexts.md) — DDD architecture
- [Clean Architecture](../concepts/clean-architecture.md) — Dependency direction

## Summary

Promenade's custom DI container provides:

1. **Type safety** without reflection overhead
2. **Explicit dependencies** that are easy to trace and debug
3. **Natural DDD boundaries** with context-based grouping
4. **Full IDE support** with autocomplete and refactoring
5. **Simple implementation** using plain Go functions

This approach sacrifices some "magic" for clarity, maintainability, and alignment with Domain-Driven Design principles.
