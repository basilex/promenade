# Accounting Context - Integration Tests

Integration tests for the **Accounting** bounded context, covering core accounting functionality with real database operations.

## Test Coverage

All core aggregates have comprehensive integration tests:

### **Account** (`account/repository_test.go`)

- CRUD operations (Create, GetByID, GetByCode, Update, Delete)
- Batch operations (CreateMany, UpdateMany)
- Query operations (ListByOrganization, ListByType, ListChildren)
- Active/Inactive filtering
- Parent-child relationships

### **JournalEntry** (`journalentry/repository_test.go`)

- CRUD operations
- Multiple journal entry lines
- Entry posting workflow
- Entry reversal workflow
- Status filtering (Draft, Posted, Reversed)
- Balanced debit/credit validation

### **FiscalPeriod** (`fiscalperiod/repository_test.go`)

- CRUD operations (Create, GetByID, Update, Delete)
- Period lifecycle (Close, Reopen, Lock)
- Query operations (ListByOrganization, ListByYear, ListOpen)
- Date-based lookups (GetByOrganizationAndDate)
- Business rules validation
  - Cannot lock open period
  - Cannot reopen locked period
  - Can post transaction to open period
- **13 tests total**

### **TaxCode** (`taxcode/repository_test.go`)

- CRUD operations (Create, GetByID, GetByCode, Update, Delete)
- Rate management (ChangeRate)
- GL account mapping (SetGLAccounts)
- Status management (Activate/Deactivate)
- Query operations (ListByOrganization, ListByType, ListActive)
- Tax calculations
  - Calculate tax from net amount
  - Calculate tax from gross amount
  - Different tax types (VAT, Sales, Withholding)
- Business rules validation
  - Invalid rate validation
  - Empty code/name validation
- **15 tests total**

### **CostCenter** (`costcenter/repository_test.go`)

- CRUD operations (Create, GetByID, GetByCode, Update, Delete)
- Details management (Update details)
- Status management (Activate/Deactivate)
- Manager assignment (SetManager)
- Hierarchical structure (parent-child relationships)
- Query operations (ListByOrganization, ListByType, ListActive)
- Business rules validation
  - Empty code/name validation
  - Invalid center type validation
  - Cannot be own parent
  - SetParent validation
- Center types testing (Department, Project, Location, Product, Other)
- **16 tests total**

### **Budget** (`budget/repository_test.go`)

- CRUD operations (Create, GetByID, GetByName, Delete)
- Budget lines management
  - Add line
  - Update line
  - Remove line
- Budget lifecycle
  - Approve budget
  - Activate budget
  - Close budget
- Actuals tracking
  - Update actuals
  - Variance calculation
- Query operations (ListByOrganization, ListByStatus, ListActive)
- Business rules validation
  - Empty name validation
  - Invalid fiscal year
  - Cannot modify approved budget
  - Cannot approve without lines
  - Cannot activate without approval
  - Cannot update actuals when not active
- **21+ tests total**

### **Reconciliation** (`reconciliation/repository_test.go`)

- CRUD operations (Create, GetByID, Update, Delete)
- Reconciliation lifecycle
  - Complete reconciliation
  - Approve reconciliation
- Reconciliation items management
  - Add item
  - Mark item matched
  - Remove item
- Query operations (ListByOrganization, ListByBankAccount, ListByStatus)
- Balance calculations
  - Calculate difference
  - Update balances
- Business rules validation
  - Cannot complete with unmatched items
  - Cannot modify completed reconciliation
  - Reopen completed reconciliation
  - Cannot reopen approved reconciliation
- **17 tests total**

## Running Tests

```bash
# Run all accounting integration tests
make test-integration CONTEXT=accounting

# Run specific aggregate tests
go test ./test/integration/contexts/accounting/account/...
go test ./test/integration/contexts/accounting/journalentry/...
go test ./test/integration/contexts/accounting/fiscalperiod/...
go test ./test/integration/contexts/accounting/taxcode/...
go test ./test/integration/contexts/accounting/costcenter/...
go test ./test/integration/contexts/accounting/budget/...
go test ./test/integration/contexts/accounting/reconciliation/...

# Run with verbose output
go test -v ./test/integration/contexts/accounting/...

# Run specific test
go test -v ./test/integration/contexts/accounting/account/ -run TestAccountRepository_Create
go test -v ./test/integration/contexts/accounting/fiscalperiod/ -run TestFiscalPeriod_BusinessRules
```

## Test Statistics

| Aggregate      | Tests    | Coverage Areas                               |
| -------------- | -------- | -------------------------------------------- |
| Account        | ~18      | CRUD, batch ops, queries, hierarchy          |
| JournalEntry   | ~15      | CRUD, posting, reversal, balanced validation |
| FiscalPeriod   | 13       | CRUD, lifecycle, business rules              |
| TaxCode        | 15       | CRUD, calculations, GL mapping, types        |
| CostCenter     | 16       | CRUD, hierarchy, manager, types              |
| Budget         | 21+      | CRUD, lines, lifecycle, actuals, variance    |
| Reconciliation | 17       | CRUD, items, matching, lifecycle, balances   |
| **TOTAL**      | **102+** | **Comprehensive coverage**                   |

## Test Structure

Each aggregate has its own directory with `repository_test.go`:

```
test/integration/contexts/accounting/
 README.md                          # This file
 helpers.go                         # Shared test utilities
 account/
    repository_test.go             # 18 tests - CRUD, batch, queries
 journalentry/
    repository_test.go             # 15 tests - posting, reversal
 fiscalperiod/
    repository_test.go             # 13 tests - periods, lifecycle
 taxcode/
    repository_test.go             # 15 tests - tax calculations
    helpers_test.go                # Tax test helpers
 costcenter/
    repository_test.go             # 16 tests - hierarchy, manager
 budget/
    repository_test.go             # 21 tests - lines, actuals, variance
    helpers_test.go                # Budget test helpers
 reconciliation/
     repository_test.go             # 17 tests - matching, balances
     helpers_test.go                # Reconciliation test helpers
```

## Test Patterns

### 1. **Database Transactions**

All tests use `db.WithTransaction()` to ensure isolation:

```go
func TestAccountRepository_Create(t *testing.T) {
    db := integration.SetupTestDB(t)
    db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
        repo := postgres.NewAccountRepository(db.DB)
        // Test logic here
    })
}
```

### 2. **Test Data Creation**

Use aggregate constructors to create valid test data:

```go
account, err := aggregate.NewAccount(orgID, "1000", "Cash", aggregate.AccountTypeAsset, "USD")
require.NoError(t, err)
```

### 3. **Assertions**

- Use `require.NoError()` for operations that must succeed
- Use `assert.Equal()` for value comparisons
- Use `assert.Error()` for expected failures

### 4. **Journal Entry Testing**

Always ensure balanced entries (total debits = total credits):

```go
entry.AddLine(accountID1, 10000, 0, "USD", "Debit")  // Debit $100.00
entry.AddLine(accountID2, 0, 10000, "USD", "Credit") // Credit $100.00
```

## Key Accounting Concepts Tested

### Double-Entry Bookkeeping

- Every entry has balanced debits and credits
- `Total Debits = Total Credits`

### Account Types

- **Asset** (1000-1999): Cash, Receivables, Inventory
- **Liability** (2000-2999): Payables, Loans
- **Equity** (3000-3999): Capital, Retained Earnings
- **Revenue** (4000-4999): Sales, Service Income
- **Expense** (5000-5999): Salaries, Rent, Utilities
- **COGS** (6000-6999): Cost of Goods Sold

### Entry Lifecycle

1. **Draft** → Entry created, can be edited
2. **Posted** → Entry posted to ledger, immutable
3. **Reversed** → Entry cancelled via reversal entry

### Account Hierarchy

Accounts can have parent-child relationships:

- Parent: `1000 - Current Assets`
  - Child: `1010 - Cash`
  - Child: `1020 - Bank Accounts`

## Database Schema

Tests interact with these tables:

- `accounting.accounts` - Chart of accounts
- `accounting.journal_entries` - Journal entry headers
- `accounting.journal_entry_lines` - Journal entry lines (debits/credits)

## Performance Considerations

- Use transactions for isolation (automatic rollback)
- Tests run in parallel by default
- Each test creates fresh test data
- Database is reset between test runs

## Debugging Failed Tests

```bash
# Run single test with verbose output
go test -v ./test/integration/contexts/accounting/account/ -run TestAccountRepository_Create

# Check database state (tests use transactions, so changes rollback)
# Enable logging in repository implementation
```

## Related Documentation

- [Accounting Context README](../../../internal/contexts/accounting/README.md)
- [Integration Test Guide](../../README.md)
- [Test Structure](../../TESTING_STRUCTURE.md)
- [Accounting Business Rules](../../../docs/concepts/accounting.md)
