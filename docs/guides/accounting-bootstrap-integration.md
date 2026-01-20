# Accounting Context - Bootstrap Integration Guide

**How to wire accounting context into the application**

---

## Overview

This guide shows how to register accounting components in `cmd/api/bootstrap.go` to enable automatic journal entry creation from business events.

---

## Step-by-Step Integration

### 1. Import Required Packages

```go
// cmd/api/bootstrap.go

import (
    // ... existing imports ...

    // Accounting Context
    accountRepo "github.com/basilex/promenade/internal/contexts/accounting/account/adapter/repository/postgres"
    accountUseCase "github.com/basilex/promenade/internal/contexts/accounting/account/usecase"

    journalEntryRepo "github.com/basilex/promenade/internal/contexts/accounting/journalentry/adapter/repository/postgres"
    journalEntryUseCase "github.com/basilex/promenade/internal/contexts/accounting/journalentry/usecase"

    ledgerRepo "github.com/basilex/promenade/internal/contexts/accounting/ledger/adapter/repository/postgres"
    ledgerUseCase "github.com/basilex/promenade/internal/contexts/accounting/ledger/usecase"

    accountingIntegration "github.com/basilex/promenade/internal/contexts/accounting/integration"

    accountHandler "github.com/basilex/promenade/internal/contexts/accounting/account/adapter/http"
    journalEntryHandler "github.com/basilex/promenade/internal/contexts/accounting/journalentry/adapter/http"
)
```

### 2. Add Fields to App Struct

```go
// App holds all application dependencies
type App struct {
    // ... existing fields ...

    // Accounting Context
    AccountUseCase           accountUseCase.IAccountUseCase
    JournalEntryUseCase      journalEntryUseCase.IJournalEntryUseCase
    LedgerUseCase            ledgerUseCase.ILedgerUseCase
    AccountingBankHandler    *accountingIntegration.BankEventHandler
    AccountingBillingHandler *accountingIntegration.BillingEventHandler
    AccountHandler           *accountHandler.AccountHandler
    JournalEntryHandler      *journalEntryHandler.JournalEntryHandler
}
```

### 3. Initialize Repositories

```go
// Bootstrap initializes all application dependencies
func Bootstrap(cfg *config.AppConfig) (*App, error) {
    // ... existing initialization ...

    // Initialize Accounting repositories
    accountRepository := accountRepo.NewAccountRepository(db)
    journalEntryRepository := journalEntryRepo.NewJournalEntryRepository(db)
    ledgerRepository := ledgerRepo.NewLedgerRepository(db)

    logger.Info("Accounting repositories initialized")

    // ... continue with use cases ...
}
```

### 4. Initialize Use Cases

```go
    // Initialize Accounting use cases
    accountUC := accountUseCase.NewAccountUseCase(accountRepository)
    journalEntryUC := journalEntryUseCase.NewJournalEntryUseCase(
        journalEntryRepository,
        accountRepository,
        ledgerRepository,
    )
    ledgerUC := ledgerUseCase.NewLedgerUseCase(ledgerRepository, accountRepository)

    app.AccountUseCase = accountUC
    app.JournalEntryUseCase = journalEntryUC
    app.LedgerUseCase = ledgerUC

    logger.Info("Accounting use cases initialized")
```

### 5. Load Standard Accounts

```go
    // Load standard accounts for automatic journal entries
    standardAccounts, err := loadStandardAccounts(ctx, accountRepository)
    if err != nil {
        logger.Warn("Failed to load standard accounts, automatic entries may fail", "error", err)
        // Use default IDs if loading fails
        standardAccounts = getDefaultStandardAccounts()
    }

    logger.Info("Standard accounts loaded",
        slog.String("bank_account", standardAccounts.BankAccount.String()),
        slog.String("sales_revenue", standardAccounts.SalesRevenue.String()),
    )
```

### 6. Initialize Integration Handlers

```go
    // Initialize Accounting integration handlers
    bankEventHandler := accountingIntegration.NewBankEventHandler(
        journalEntryUC,
        standardAccounts,
    )

    billingEventHandler := accountingIntegration.NewBillingEventHandler(
        journalEntryUC,
        standardAccounts,
    )

    app.AccountingBankHandler = bankEventHandler
    app.AccountingBillingHandler = billingEventHandler

    logger.Info("Accounting integration handlers initialized")
```

### 7. Initialize HTTP Handlers

```go
    // Initialize Accounting HTTP handlers
    accountHTTPHandler := accountHandler.NewAccountHandler(accountUC)
    journalEntryHTTPHandler := journalEntryHandler.NewJournalEntryHandler(journalEntryUC)

    app.AccountHandler = accountHTTPHandler
    app.JournalEntryHandler = journalEntryHTTPHandler

    logger.Info("Accounting HTTP handlers initialized")
```

### 8. Register Event Subscribers

```go
    // Register Accounting event subscribers
    registerAccountingEventHandlers(eventBus, app)

    logger.Info("Accounting event handlers registered")
```

### 9. Helper Functions

```go
// loadStandardAccounts loads standard accounts from database
func loadStandardAccounts(ctx context.Context, repo accountRepo.IAccountRepository) (accountingIntegration.StandardAccounts, error) {
    accounts := accountingIntegration.StandardAccounts{}

    // Load by account code
    bankAccount, err := repo.GetByCode(ctx, "311")
    if err != nil {
        return accounts, fmt.Errorf("failed to load bank account (311): %w", err)
    }
    accounts.BankAccount = bankAccount.ID

    cashRegister, err := repo.GetByCode(ctx, "301")
    if err != nil {
        return accounts, fmt.Errorf("failed to load cash register (301): %w", err)
    }
    accounts.CashRegister = cashRegister.ID

    accountsReceivable, err := repo.GetByCode(ctx, "361")
    if err != nil {
        return accounts, fmt.Errorf("failed to load accounts receivable (361): %w", err)
    }
    accounts.AccountsReceivable = accountsReceivable.ID

    salesRevenue, err := repo.GetByCode(ctx, "702")
    if err != nil {
        return accounts, fmt.Errorf("failed to load sales revenue (702): %w", err)
    }
    accounts.SalesRevenue = salesRevenue.ID

    operatingExpenses, err := repo.GetByCode(ctx, "902")
    if err != nil {
        return accounts, fmt.Errorf("failed to load operating expenses (902): %w", err)
    }
    accounts.OperatingExpenses = operatingExpenses.ID

    accountsPayable, err := repo.GetByCode(ctx, "631")
    if err != nil {
        return accounts, fmt.Errorf("failed to load accounts payable (631): %w", err)
    }
    accounts.AccountsPayable = accountsPayable.ID

    return accounts, nil
}

// getDefaultStandardAccounts returns hardcoded account IDs as fallback
func getDefaultStandardAccounts() accountingIntegration.StandardAccounts {
    return accountingIntegration.StandardAccounts{
        BankAccount:        uuidv7.MustParse("01000000-0000-0000-0000-000000000311"),
        CashRegister:       uuidv7.MustParse("01000000-0000-0000-0000-000000000301"),
        AccountsReceivable: uuidv7.MustParse("01000000-0000-0000-0000-000000000361"),
        SalesRevenue:       uuidv7.MustParse("07000000-0000-0000-0000-000000000702"),
        OperatingExpenses:  uuidv7.MustParse("09000000-0000-0000-0000-000000000902"),
        AccountsPayable:    uuidv7.MustParse("06000000-0000-0000-0000-000000000631"),
    }
}

// registerAccountingEventHandlers registers all accounting event subscribers
func registerAccountingEventHandlers(eventBus bus.IBus, app *App) {
    // Bank transaction events
    eventBus.Subscribe(bus.BankTransactionRecorded, app.AccountingBankHandler.HandleTransactionRecorded)
    eventBus.Subscribe(bus.BankTransactionReconciled, app.AccountingBankHandler.HandleTransactionReconciled)

    // Billing events
    eventBus.Subscribe(bus.TopicInvoiceGenerated, app.AccountingBillingHandler.HandleInvoiceGenerated)
    eventBus.Subscribe(bus.TopicPaymentReceived, app.AccountingBillingHandler.HandlePaymentReceived)

    logger.Info("Registered accounting event subscribers",
        slog.Int("subscribers", 4),
    )
}
```

---

## Complete Bootstrap Function

```go
func Bootstrap(cfg *config.AppConfig) (*App, error) {
    app := &App{Config: cfg}
    ctx := context.Background()

    // 1. Initialize logger
    if err := initLogger(cfg); err != nil {
        return nil, err
    }

    // 2. Initialize database
    db, err := initDatabase(cfg)
    if err != nil {
        return nil, err
    }
    app.DB = db

    // 3. Run migrations
    if err := runMigrations(db); err != nil {
        return nil, err
    }

    // 4. Initialize Redis and cache
    redisClient := initRedis(cfg)
    app.RedisClient = redisClient

    cacheClient, err := initCache(cfg, redisClient)
    if err != nil {
        return nil, err
    }
    app.CacheClient = cacheClient

    // 5. Initialize Event Bus
    eventBus, err := initEventBus(cfg, redisClient)
    if err != nil {
        return nil, err
    }
    app.EventBus = eventBus

    // 6. Initialize Auth (JWT, Token Revoker)
    jwtManager, tokenRevoker := initAuth(cfg, redisClient)
    app.JWTManager = jwtManager
    app.TokenRevoker = tokenRevoker

    // 7. Initialize repositories
    // ... existing repositories (banking, billing, etc.) ...

    // Accounting repositories
    accountRepository := accountRepo.NewAccountRepository(db)
    journalEntryRepository := journalEntryRepo.NewJournalEntryRepository(db)
    ledgerRepository := ledgerRepo.NewLedgerRepository(db)

    // 8. Initialize use cases
    // ... existing use cases ...

    // Accounting use cases
    accountUC := accountUseCase.NewAccountUseCase(accountRepository)
    journalEntryUC := journalEntryUseCase.NewJournalEntryUseCase(
        journalEntryRepository,
        accountRepository,
        ledgerRepository,
    )
    ledgerUC := ledgerUseCase.NewLedgerUseCase(ledgerRepository, accountRepository)

    app.AccountUseCase = accountUC
    app.JournalEntryUseCase = journalEntryUC
    app.LedgerUseCase = ledgerUC

    // 9. Load standard accounts
    standardAccounts, err := loadStandardAccounts(ctx, accountRepository)
    if err != nil {
        logger.Warn("Failed to load standard accounts", "error", err)
        standardAccounts = getDefaultStandardAccounts()
    }

    // 10. Initialize integration handlers
    bankEventHandler := accountingIntegration.NewBankEventHandler(
        journalEntryUC,
        standardAccounts,
    )
    billingEventHandler := accountingIntegration.NewBillingEventHandler(
        journalEntryUC,
        standardAccounts,
    )

    app.AccountingBankHandler = bankEventHandler
    app.AccountingBillingHandler = billingEventHandler

    // 11. Initialize HTTP handlers
    // ... existing handlers ...

    accountHTTPHandler := accountHandler.NewAccountHandler(accountUC)
    journalEntryHTTPHandler := journalEntryHandler.NewJournalEntryHandler(journalEntryUC)

    app.AccountHandler = accountHTTPHandler
    app.JournalEntryHandler = journalEntryHTTPHandler

    // 12. Register event subscribers
    // ... existing subscribers ...
    registerAccountingEventHandlers(eventBus, app)

    // 13. Initialize health checker
    healthChecker := initHealthChecker(db, redisClient, eventBus)
    app.HealthChecker = healthChecker

    logger.Info("Application bootstrap completed successfully")
    return app, nil
}
```

---

## Router Integration

Add accounting routes to `internal/contexts/accounting/router.go`:

```go
package accounting

import (
    "github.com/gin-gonic/gin"

    accountHandler "github.com/basilex/promenade/internal/contexts/accounting/account/adapter/http"
    journalEntryHandler "github.com/basilex/promenade/internal/contexts/accounting/journalentry/adapter/http"
    "github.com/basilex/promenade/pkg/middleware"
)

// RegisterRoutes registers all accounting context routes
func RegisterRoutes(r *gin.RouterGroup,
    accountH *accountHandler.AccountHandler,
    journalEntryH *journalEntryHandler.JournalEntryHandler,
    authMiddleware gin.HandlerFunc,
    rbacMiddleware func(permission string) gin.HandlerFunc,
) {
    accounting := r.Group("/accounting")
    accounting.Use(authMiddleware)

    // Accounts
    accounts := accounting.Group("/accounts")
    {
        accounts.GET("", rbacMiddleware("accounting:accounts:read"), accountH.List)
        accounts.GET("/:id", rbacMiddleware("accounting:accounts:read"), accountH.Get)
        accounts.POST("", rbacMiddleware("accounting:accounts:write"), accountH.Create)
        accounts.PUT("/:id", rbacMiddleware("accounting:accounts:write"), accountH.Update)
        accounts.DELETE("/:id", rbacMiddleware("accounting:accounts:write"), accountH.Deactivate)
    }

    // Journal Entries
    entries := accounting.Group("/journal-entries")
    {
        entries.GET("", rbacMiddleware("accounting:entries:read"), journalEntryH.List)
        entries.GET("/:id", rbacMiddleware("accounting:entries:read"), journalEntryH.Get)
        entries.POST("", rbacMiddleware("accounting:entries:write"), journalEntryH.Create)
        entries.PUT("/:id", rbacMiddleware("accounting:entries:write"), journalEntryH.Update)
        entries.POST("/:id/post", rbacMiddleware("accounting:entries:post"), journalEntryH.Post)
        entries.POST("/:id/reverse", rbacMiddleware("accounting:entries:post"), journalEntryH.Reverse)
    }

    // Reports
    reports := accounting.Group("/reports")
    {
        reports.GET("/trial-balance", rbacMiddleware("accounting:reports:read"), journalEntryH.TrialBalance)
        reports.GET("/balance-sheet", rbacMiddleware("accounting:reports:read"), journalEntryH.BalanceSheet)
        reports.GET("/income-statement", rbacMiddleware("accounting:reports:read"), journalEntryH.IncomeStatement)
    }
}
```

Add to main router in `cmd/api/server.go`:

```go
// Register accounting routes
accounting.RegisterRoutes(
    v1,
    app.AccountHandler,
    app.JournalEntryHandler,
    authMiddleware,
    rbacMiddleware,
)
```

---

## Testing the Integration

### 1. Run Migrations

```bash
make migrate-module MODULE=accounting
```

### 2. Start Server

```bash
make dev
```

### 3. Verify Standard Accounts Loaded

```bash
curl http://localhost:8081/api/v1/accounting/accounts | jq
```

Should see accounts: 301, 311, 361, 702, 902, 631

### 4. Create Bank Transaction

```bash
curl -X POST http://localhost:8081/api/v1/banking/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "bank_account_id": "uuid",
    "direction": "credit",
    "amount_cents": 10000,
    "currency_code": "UAH",
    "description": "Test payment"
  }'
```

### 5. Verify Journal Entry Created

```bash
curl http://localhost:8081/api/v1/accounting/journal-entries | jq
```

Should see automatic entry:

```json
{
  "id": "uuid",
  "description": "Bank transaction: Test payment",
  "status": "posted",
  "lines": [
    {
      "account_id": "311",
      "debit_cents": 10000,
      "credit_cents": 0
    },
    {
      "account_id": "702",
      "debit_cents": 0,
      "credit_cents": 10000
    }
  ]
}
```

---

## Troubleshooting

### Issue: "Failed to load standard accounts"

**Cause**: Migration not run or account codes changed

**Solution**:

```bash
make migrate-module MODULE=accounting
# Check accounts exist
psql -d promenade -c "SELECT code, name FROM accounts WHERE code IN ('311', '702');"
```

### Issue: "Journal entry not created for bank transaction"

**Cause**: Event subscriber not registered or event bus not working

**Solution**:

1. Check event bus health: `curl http://localhost:8081/health`
2. Check logs for "Registered accounting event subscribers"
3. Verify event published: check logs for "bank.transaction.recorded"

### Issue: "Journal entry created but not posted"

**Cause**: Validation failed (debits ≠ credits) or posting error

**Solution**:

1. Check logs for validation errors
2. Verify entry balance: `SELECT * FROM journal_entry_lines WHERE journal_entry_id = ?`
3. Manual post: `POST /api/v1/accounting/journal-entries/:id/post`

---

## Performance Considerations

### Event Processing

- Journal entry creation is **asynchronous** (via Event Bus)
- Bank transaction API returns immediately
- Journal entry created in background (< 100ms)
- Failures logged, don't block business operations

### Database Indexes

Already created in migrations:

- `idx_accounts_code` - Fast account lookup by code
- `idx_journal_entries_source` - Fast lookup by source event
- `idx_journal_entry_lines_account` - Fast balance calculation

### Caching

Consider caching standard accounts:

```go
// Load once at startup, cache in memory
var standardAccountsCache accountingIntegration.StandardAccounts

func getStandardAccounts() accountingIntegration.StandardAccounts {
    return standardAccountsCache // No DB query
}
```

---

## Next Steps

1. **Implement repositories** - Complete PostgreSQL implementations
2. **Implement use cases** - Business logic layer
3. **Implement HTTP handlers** - REST API endpoints
4. **Add unit tests** - Test aggregates and use cases
5. **Add integration tests** - Test event flows end-to-end
6. **Add RBAC permissions** - Define accounting permissions
7. **Add Swagger docs** - Document API endpoints

---

**Status**:  Guide Complete  
**Target**: Production-ready accounting integration
