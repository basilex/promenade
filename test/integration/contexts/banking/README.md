# Banking Integration Tests

Repository-layer integration tests with real PostgreSQL database.

## Overview

27 comprehensive tests covering BankAccount and BankTransaction repositories with:

- Real database operations
- Transaction rollback after each test
- Money precision validation (kopiyky)
- Provider integration scenarios
- Status transition workflows
- Matching and reconciliation

## Test Structure

```
test/integration/contexts/banking/
  bankaccount/
    repository_test.go       # 10 tests - BankAccount repository
  banktransaction/
    repository_test.go       # 17 tests - BankTransaction repository
  README.md                  # This file
```

## Running Tests

```bash
# All Banking integration tests
go test ./test/integration/contexts/banking/...

# Verbose output
go test -v ./test/integration/contexts/banking/...

# Specific aggregate
go test ./test/integration/contexts/banking/bankaccount
go test ./test/integration/contexts/banking/banktransaction

# With coverage
go test -cover ./test/integration/contexts/banking/...
```

## BankAccount Tests (10 tests)

### CRUD Operations

- **Create** - Create account and verify persistence
- **GetByID_NotFound** - Query non-existent account returns error
- **Update** - Modify account details
- **Delete** - Soft delete preserves data
- **List** - Pagination and filtering

### Provider Integration

- **GetByProviderAccountID** - Query by provider + external ID
- **ProviderAccounts** - Multiple provider types (Monobank, Privat24, PUMB)

### Organization Queries

- **CountByOrganization** - Count accounts per organization

### Critical Tests

- **BalancePrecision**  - Exact kopiyky precision validation
- **StatusTransitions** - Active → Inactive → Archived workflow

## BankTransaction Tests (17 tests)

### CRUD Operations

- **Create** - Create transaction and verify
- **GetByID_NotFound** - Query non-existent returns error
- **Update** - Modify transaction details
- **Delete** - Soft delete preserves history
- **ListByAccount** - List account transactions with pagination

### Provider Sync

- **GetByExternalID** - Query by provider transaction ID
- **ExternalIDUnique** - Verify external_id uniqueness per account

### Unmatched Queries

- **ListUnmatched** - List transactions needing reconciliation
- **CountByAccount** - Transaction count per account

### Matching & Reconciliation

- **MatchToInvoice** - Match transaction to invoice, verify reconciled status
- **MatchToOrder** - Match to order
- **MatchToPayment** - Match to payment
- **Unmatch** - Remove match, return to booked status

### Lifecycle

- **StatusTransitions** - Pending → Booked → Reconciled workflow
- **Cancel** - Cancel pending transaction

### Metadata

- **Counterparty** - Store counterparty name and IBAN
- **AmountPrecision**  - Exact kopiyky precision validation

## Test Patterns

### Setup Pattern

```go
func TestBankAccountRepository_Create(t *testing.T) {
    // Setup test database
    db := integration.SetupTestDB(t)

    // Run in transaction (auto-rollback after test)
    db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
        // Create repository
        repo := postgres.NewBankAccountRepository(db.DB)

        // Create aggregate
        acc, err := aggregate.NewBankAccount(
            uuidv7.New(),        // organizationID
            "Test Account",      // name
            "Test Bank",         // bankName
            "UAH",              // currencyCode
            uuidv7.New(),       // lastUpdatedBy
        )
        require.NoError(t, err)

        // Execute operation
        err = repo.Create(ctx, acc)
        require.NoError(t, err)

        // Verify result
        found, err := repo.GetByID(ctx, acc.GetID())
        require.NoError(t, err)
        assert.Equal(t, acc.GetID(), found.GetID())
    })
}
```

### Money Precision Pattern 

```go
// CRITICAL: Always verify kopiyky precision
transaction.AmountCents = 123456  // 1234.56 UAH
repo.Create(ctx, transaction)

found, _ := repo.GetByID(ctx, transaction.GetID())
assert.Equal(t, int64(123456), found.AmountCents,
    "Amount precision must be exact to the kopiyok! 1234.56 UAH = 123456 kopiyky")
```

### Status Transition Pattern

```go
// Verify status workflow
acc, _ := aggregate.NewBankAccount(...)
repo.Create(ctx, acc)

// Deactivate
_ = acc.Deactivate()
repo.Update(ctx, acc)
found, _ := repo.GetByID(ctx, acc.GetID())
assert.Equal(t, aggregate.BankAccountStatusInactive, found.Status)

// Activate
_ = acc.Activate()
repo.Update(ctx, acc)
found, _ = repo.GetByID(ctx, acc.GetID())
assert.Equal(t, aggregate.BankAccountStatusActive, found.Status)

// Archive (final)
_ = acc.Archive()
repo.Update(ctx, acc)
found, _ = repo.GetByID(ctx, acc.GetID())
assert.Equal(t, aggregate.BankAccountStatusArchived, found.Status)
```

### Matching Workflow Pattern

```go
// Create and book transaction
tx, _ := txAggregate.NewBankTransaction(accountID, DirectionCredit, 75000, "UAH", time.Now(), "Payment", userID)
repo.Create(ctx, tx)

_ = tx.Book()
repo.Update(ctx, tx)

// Match to invoice
invoiceID := uuidv7.New()
_ = tx.Match(txAggregate.MatchedEntityInvoice, invoiceID)
repo.Update(ctx, tx)

// Verify reconciliation
found, _ := repo.GetByID(ctx, tx.GetID())
assert.Equal(t, txAggregate.MatchedEntityInvoice, found.MatchedEntityType)
assert.Equal(t, invoiceID, found.MatchedEntityID)
assert.Equal(t, txAggregate.TransactionStatusReconciled, found.Status)
```

## Test Database Setup

Uses `integration.SetupTestDB(t)` which:

- Creates test database connection
- Runs migrations
- Provides transaction rollback after each test
- Cleans up resources automatically

### Transaction Isolation

Each test runs in isolated transaction via `db.WithTransaction()`:

- Changes rolled back after test
- No test pollution
- Parallel test execution safe
- Fast cleanup

## Critical Test Cases

### 1. Money Precision (2 tests)

**Why Critical**: Financial data must be exact to the kopiyok (cent)

- BankAccount balance precision
- BankTransaction amount precision
- No floating-point errors
- Example: 1234.56 UAH = 123456 kopiyky exactly

### 2. Provider Integration (4 tests)

**Why Critical**: External bank sync deduplication

- ExternalID uniqueness per account
- GetByProviderAccountID lookup
- Multiple provider support (Monobank, Privat24, PUMB)
- Prevents duplicate transaction imports

### 3. Status Transitions (2 tests)

**Why Critical**: Enforces business rules

- BankAccount: Active → Inactive → Archived
- BankTransaction: Pending → Booked → Reconciled
- Invalid transitions prevented at domain layer

### 4. Matching & Reconciliation (4 tests)

**Why Critical**: Core reconciliation workflow

- Match to Invoice/Order/Payment
- Unmatch functionality
- Status changes on match/unmatch
- ListUnmatched query for reconciliation UI

## Coverage Summary

| Aggregate       | Tests  | Lines    | Coverage                            |
| --------------- | ------ | -------- | ----------------------------------- |
| BankAccount     | 10     | ~250     | Repository CRUD + Domain Queries    |
| BankTransaction | 17     | ~420     | Repository CRUD + Matching + Status |
| **Total**       | **27** | **~670** | **Full Repository Layer**           |

## Common Test Issues

### Issue 1: Constructor Parameter Mismatch

```go
//  Wrong - missing parameters
acc, _ := aggregate.NewBankAccount(orgID, "Name")

//  Correct - all required parameters
acc, _ := aggregate.NewBankAccount(orgID, "Name", "Bank", "UAH", lastUpdatedBy)
```

### Issue 2: Status Constants

```go
//  Wrong - old naming
assert.Equal(t, aggregate.StatusActive, acc.Status)

//  Correct - prefixed naming
assert.Equal(t, aggregate.BankAccountStatusActive, acc.Status)
```

### Issue 3: Method Signatures

```go
//  Wrong - methods return error
transaction.SetCounterparty("Name", "IBAN")

//  Correct - SetCounterparty is void
tx.SetCounterparty("Name", "IBAN")  // No return value
```

## Related Documentation

- [../../../internal/contexts/banking/README.md](../../../internal/contexts/banking/README.md) - Banking context architecture
- [../smoke/contexts/banking/README.md](../smoke/contexts/banking/README.md) - Smoke tests documentation
- [../../TESTING_STRUCTURE.md](../../TESTING_STRUCTURE.md) - Overall testing strategy
- [../../README.md](../../README.md) - Integration test setup guide
