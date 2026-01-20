# Banking Context Tests

Comprehensive test suite for Banking bounded context covering repository and handler layers.

## Quick Navigation

 **Test Overview**

- [Status Summary](#status-summary) - Current completion status
- [Test Structure](#test-structure) - File organization
- [Running Tests](#running-tests) - Execution commands

 **Critical Patterns**

- [Money Precision](#1-money-precision-kopiyky--critical) - Kopiyky precision validation
- [Matching Workflow](#2-matching-workflow) - Transaction reconciliation
- [Provider Integration](#3-provider-accounts) - External bank connections

 **Detailed Tests**

- [Integration Tests](#completed-integration-tests-repository-layer-) - Repository layer with PostgreSQL
- [Smoke Tests](#completed-smoke-tests-handler-layer-) - Handler layer with mocks

 **Related Documentation**

- [../../../internal/contexts/banking/README.md](../../../internal/contexts/banking/README.md) - Banking context implementation
- [../../integration/contexts/banking/](../../integration/contexts/banking/) - Integration test source code
- [../../TESTING_STRUCTURE.md](../../TESTING_STRUCTURE.md) - Overall testing strategy

---

## Status Summary

 **Integration Tests**: 27 repository tests completed and passing  
 **Smoke Tests**: 23 handler tests completed and passing  
 **Test Coverage**: Full coverage of repository and handler layers

## Test Structure

```
test/
  smoke/contexts/banking/
    bankaccount/handler_test.go      -  10 handler tests with mocked UseCase
    banktransaction/handler_test.go  -  13 handler tests with mocked UseCase
    README.md                         - This file

  integration/contexts/banking/
    bankaccount/repository_test.go   -  10 repository tests with PostgreSQL
    banktransaction/repository_test.go -  17 repository tests with PostgreSQL
```

## Completed: Smoke Tests (Handler Layer) 

### BankAccount Handler (10 tests)

**Account Management**:

- `TestBankAccountHandler_CreateManual_Success` - Create manual account
- `TestBankAccountHandler_ConnectProvider_Success` - Connect provider account
- `TestBankAccountHandler_GetByID_Success` - Retrieve account by ID
- `TestBankAccountHandler_List_Success` - List organization accounts
- `TestBankAccountHandler_Delete_Success` - Delete account

**Balance Operations**:

- `TestBankAccountHandler_UpdateBalance_Success` - Update balance manually
- `TestBankAccountHandler_RecordSync_Success` - Record provider sync

**Status Management**:

- `TestBankAccountHandler_Activate_Success` - Activate account
- `TestBankAccountHandler_Deactivate_Success` - Deactivate account
- `TestBankAccountHandler_Archive_Success` - Archive account

### BankTransaction Handler (13 tests)

**Transaction Recording**:

- `TestBankTransactionHandler_Record_Debit_Success` - Record debit transaction
- `TestBankTransactionHandler_Record_Credit_Success` - Record credit transaction
- `TestBankTransactionHandler_GetByID_Success` - Retrieve transaction
- `TestBankTransactionHandler_ListByAccount_Success` - List account transactions
- `TestBankTransactionHandler_ListUnmatched_Success` - List unmatched transactions
- `TestBankTransactionHandler_Delete_Success` - Delete transaction

**Transaction Lifecycle**:

- `TestBankTransactionHandler_Book_Success` - Book pending transaction
- `TestBankTransactionHandler_Cancel_Success` - Cancel transaction

**Matching & Reconciliation**:

- `TestBankTransactionHandler_Match_ToInvoice_Success` - Match to invoice
- `TestBankTransactionHandler_Match_ToOrder_Success` - Match to order
- `TestBankTransactionHandler_Unmatch_Success` - Remove match

**Transaction Details**:

- `TestBankTransactionHandler_SetCounterparty_Success` - Update counterparty

## Completed: Integration Tests (Repository Layer) 

### BankAccount Repository (10 tests)

**CRUD**: Create, GetByID, Update, Delete, List, GetByProviderAccountID, CountByOrganization  
**Critical Money Test**: `TestBankAccountRepository_BalancePrecision`

- Verifies exact kopiyky precision: 1234.56 UAH = 123456 kopiyky
- `assert.Equal(t, int64(123456), found.BalanceCents, "Balance precision must be exact to the kopiyok!")`

**Provider Integration**: Monobank/Privat24/PUMB account connections  
**Status Transitions**: Active → Inactive → Archived

### BankTransaction Repository (17 tests)

**CRUD**: Create, GetByID, Update, Delete, ListByAccount, GetByExternalID, ListUnmatched, CountByAccount  
**Critical Money Test**: `TestBankTransactionRepository_AmountPrecision`

- Verifies exact kopiyky precision: 1234.56 UAH = 123456 kopiyky
- `assert.Equal(t, int64(123456), found.AmountCents, "Amount precision must be exact to the kopiyok!")`

**Matching & Reconciliation**:

- MatchToInvoice, MatchToOrder, MatchToPayment - Link transactions to entities
- Unmatch - Remove entity links
- ListUnmatched - Query unreconciled transactions

**Status Transitions**: Pending → Booked → Reconciled / Canceled  
**Provider Integration**: ExternalID uniqueness, Counterparty persistence, Cancel operations

## Test Patterns & Best Practices

### 1. Money Precision (Kopiyky)  CRITICAL

```go
// ALWAYS test money precision to the kopiyok!
assert.Equal(t, int64(123456), found.BalanceCents,
    "Balance must be exact to kopiyok! 1234.56 UAH = 123456 kopiyky")
```

### 2. Matching Workflow

```go
// Transaction must be booked before matching
_ = tx.Book()
_ = tx.Match(MatchedEntityInvoice, invoiceID)
// Status becomes Reconciled after match
```

### 3. Provider Accounts

```go
// Provider accounts have ExternalID for sync
_ = acc.ConnectProvider(ProviderMonobank, "mono_123", "UA...", "305299")
found, _ := repo.GetByProviderAccountID(ctx, ProviderMonobank, "mono_123")
```

### 4. Mock UseCase Pattern (Smoke Tests)

```go
type MockBankAccountUseCase struct {
    CreateManualAccountFunc func(ctx context.Context, ...) (*aggregate.BankAccount, error)
    GetAccountFunc         func(ctx context.Context, id uuidv7.UUID) (*aggregate.BankAccount, error)
}

func (m *MockBankAccountUseCase) CreateManualAccount(ctx context.Context, ...) (*aggregate.BankAccount, error) {
    if m.CreateManualAccountFunc != nil {
        return m.CreateManualAccountFunc(ctx, ...)
    }
    return nil, nil
}
```

### 5. Integration Test Pattern

```go
db := integration.SetupTestDB(t)
db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
    repo := postgres.NewBankAccountRepository(db.DB)
    // Test implementation...
})
```

## Coverage Summary

| Component       | Smoke Tests | Integration Tests | Total  |
| --------------- | ----------- | ----------------- | ------ |
| BankAccount     | 10          | 10                | 20     |
| BankTransaction | 13          | 17                | 30     |
| **Total**       | **23**      | **27**            | **50** |

## Related Documentation

- [../../../internal/contexts/banking/README.md](../../../internal/contexts/banking/README.md) - Banking context architecture
- [../../TESTING_STRUCTURE.md](../../TESTING_STRUCTURE.md) - Testing strategy
- [../../integration/README.md](../../integration/README.md) - Integration test setup

## Test Coverage Goals

-  Integration tests: 27 repository tests with CRUD + business logic
- ⏳ Smoke tests: Handler layer (needs UseCase alignment)
- ⏳ Unit tests: Aggregate business logic (future)
- ⏳ E2E tests: Full reconciliation workflow (future)

## Critical Business Rules Tested

1.  **Money Precision**: All amounts in kopiyky (cents), never float - TESTED in AmountPrecision/BalancePrecision
2.  **Matching Rules**: Only booked transactions can be matched - TESTED in MatchToInvoice/Order/Payment
3. ⏳ **Provider Sync**: Only provider accounts can sync - NOT YET TESTED
4.  **External ID**: Unique per account (not globally unique) - TESTED in ExternalIDUnique
5.  **Status Transitions**: Pending → Booked → Reconciled/Canceled - TESTED in StatusTransitions
6.  **Soft Delete**: Accounts/transactions never hard deleted - TESTED in Delete tests

## Dependencies

- `testify/assert`, `testify/require` - assertions
- `test/integration` - DB test helpers (SetupTestDB, WithTransaction)
- `test/smoke` - HTTP test helpers (SetupRouter, MakeRequest, AssertSuccessResponse)
