# BankTransaction Aggregate

Bank transaction management with matching, reconciliation, and provider synchronization support.

## Quick Navigation

- [Overview](#overview) - Purpose and key features
- [Domain Model](#domain-model) - Properties and types
- [Business Rules](#business-rules) - Domain constraints
- [Use Cases](#use-cases) - Available operations
- [API Examples](#api-examples) - Usage patterns

🔗 **Related**: [../README.md](../README.md) - Banking Context Overview

---

## Overview

**BankTransaction** represents a single bank transaction (debit or credit) that can be:

- Recorded manually
- Imported from bank statements
- Synchronized from external providers
- Matched to business entities (invoices, orders, payments)

### Key Features

- Debit/credit direction tracking
- Status workflow (Pending → Booked → Reconciled)
- Entity matching (Invoice, Order, Payment)
- Counterparty information storage
- External ID for provider sync deduplication
- Amount precision in kopiyky

## Domain Model

### Properties

```go
type BankTransaction struct {
    aggregate.BaseAggregate

    // Account Reference
    BankAccountID uuidv7.UUID
    StatementID   *uuidv7.UUID  // Related bank statement

    // Provider Integration
    ExternalID string  // Provider's transaction ID (unique per account)

    // Transaction Details
    Direction    TransactionDirection  // debit, credit
    AmountCents  int64                // Amount in kopiyky
    CurrencyCode string               // UAH, USD, EUR
    Description  string               // Transaction description

    // Counterparty
    CounterpartyName string
    CounterpartyIBAN string

    // Timestamps
    TransactionAt time.Time   // When transaction occurred
    BookedAt      *time.Time  // When booked in bank system

    // Status
    Status TransactionStatus  // pending, booked, canceled, reconciled

    // Matching
    MatchedEntityType *MatchedEntityType  // invoice, order, payment
    MatchedEntityID   *uuidv7.UUID       // Entity ID
    MatchedAt         *time.Time         // When matched

    // Metadata
    RawPayload    jsonstore.Field[map[string]interface{}]  // Provider raw data
    LastUpdatedBy uuidv7.UUID
}
```

### Types

**TransactionDirection**:

- `DirectionDebit` - Money out (payment, withdrawal)
- `DirectionCredit` - Money in (receipt, deposit)

**TransactionStatus**:

- `TransactionStatusPending` - Created, not yet booked
- `TransactionStatusBooked` - Confirmed by bank
- `TransactionStatusCanceled` - Canceled/reversed
- `TransactionStatusReconciled` - Matched and reconciled

**MatchedEntityType**:

- `MatchedEntityInvoice` - Matched to customer invoice
- `MatchedEntityOrder` - Matched to order payment
- `MatchedEntityPayment` - Matched to supplier payment

## Business Rules

### 1. Amount Precision

- Amount must be positive (> 0)
- Stored in cents/kopiyky for exact precision
- Example: 1234.56 UAH = 123456 kopiyky
- No floating-point calculations

### 2. Direction Rules

- Direction must be `debit` or `credit`
- Cannot be changed after creation
- Debit = money out, Credit = money in

### 3. Status Workflow

```
Pending → Booked → Reconciled
   ↓
Canceled
```

**State Transitions**:

- Pending → Booked: `Book()` - Confirm transaction
- Pending → Canceled: `Cancel()` - Cancel before booking
- Booked → Reconciled: `Match()` - Match to entity
- Reconciled → Booked: `Unmatch()` - Remove matching

**Constraints**:

- Canceled transactions cannot be booked or matched
- Must be Booked before Matching
- Cannot match already matched transaction (must unmatch first)

### 4. Matching Rules

- Transaction must be Booked before matching
- Can match to one entity type at a time
- Matching changes status to Reconciled
- Unmatching returns status to Booked
- Cannot match Pending or Canceled transactions

### 5. External ID Uniqueness

- `external_id` must be unique per account
- Used for provider sync deduplication
- Prevents duplicate imports
- Optional for manual transactions

### 6. Validation Rules

- `amount_cents` - Required, > 0
- `currency_code` - Required, ISO 4217
- `direction` - Required (debit/credit)
- `transaction_at` - Required, not future
- `external_id` - Unique per account if provided

## Use Cases

### Transaction Management

**RecordTransaction**(accountID, direction, amountCents, currencyCode, transactionAt, description, lastUpdatedBy)

- Records new manual transaction
- Initial status = Pending
- No external_id

**RecordTransactionWithExternalID**(accountID, externalID, direction, amountCents, ...)

- Records transaction from provider sync
- Checks external_id uniqueness
- Returns existing if duplicate
- Initial status = Booked

**GetTransaction**(id) → BankTransaction

- Retrieves transaction by ID
- Returns ErrTransactionNotFound if not exists

**ListTransactionsByAccount**(accountID, limit, offset) → []BankTransaction

- Lists account transactions
- Ordered by transaction_at DESC
- Supports pagination

**ListUnmatchedTransactions**(accountID, limit, offset) → []BankTransaction

- Lists booked but unmatched transactions
- Used for reconciliation workflow
- Excludes matched and canceled

**DeleteTransaction**(id)

- Soft deletes transaction
- Preserves historical data

### Transaction Lifecycle

**BookTransaction**(id, lastUpdatedBy)

- Books pending transaction
- Sets BookedAt timestamp
- Status: Pending → Booked
- Returns ErrTransactionAlreadyBooked if already booked
- Returns ErrTransactionCanceled if canceled

**CancelTransaction**(id, lastUpdatedBy)

- Cancels pending transaction
- Status: Pending → Canceled
- Cannot cancel booked transactions
- Returns ErrTransactionAlreadyBooked

### Matching & Reconciliation

**MatchToInvoice**(transactionID, invoiceID, lastUpdatedBy)

- Matches transaction to invoice
- Status: Booked → Reconciled
- Sets MatchedEntityType = Invoice
- Sets MatchedAt timestamp
- Returns ErrTransactionNotBooked if not booked
- Returns ErrTransactionAlreadyMatched if matched

**MatchToOrder**(transactionID, orderID, lastUpdatedBy)

- Matches transaction to order
- Same rules as MatchToInvoice

**MatchToPayment**(transactionID, paymentID, lastUpdatedBy)

- Matches transaction to payment
- Same rules as MatchToInvoice

**UnmatchTransaction**(transactionID, lastUpdatedBy)

- Removes entity match
- Status: Reconciled → Booked
- Clears MatchedEntityType/ID/At
- Returns ErrTransactionNotMatched

### Transaction Details

**SetCounterparty**(id, counterpartyName, counterpartyIBAN)

- Updates counterparty information
- Validates IBAN format if provided
- Can be called multiple times

**UpdateTransactionDescription**(id, description, lastUpdatedBy)

- Updates transaction description
- Useful for manual corrections

## API Examples

### Recording Manual Transaction

```bash
POST /api/v1/banking/transactions
{
  "bank_account_id": "01JGXYZ...",
  "direction": "credit",
  "amount_cents": 150000,  // 1500.00 UAH
  "currency_code": "UAH",
  "transaction_at": "2026-01-20T10:30:00Z",
  "description": "Payment from client",
  "last_updated_by": "01JGXYZ..."
}
```

### Booking Transaction

```bash
POST /api/v1/banking/transactions/:id/book
{
  "last_updated_by": "01JGXYZ..."
}
```

### Matching to Invoice

```bash
POST /api/v1/banking/transactions/:id/match
{
  "entity_type": "invoice",
  "entity_id": "01JGXYZ...",
  "last_updated_by": "01JGXYZ..."
}
```

### Unmatching

```bash
POST /api/v1/banking/transactions/:id/unmatch
{
  "last_updated_by": "01JGXYZ..."
}
```

### Setting Counterparty

```bash
PUT /api/v1/banking/transactions/:id/counterparty
{
  "counterparty_name": "ТОВ Постачальник",
  "counterparty_iban": "UA123456789012345678901234567"
}
```

## Testing

**Integration Tests**: `test/integration/contexts/banking/banktransaction/repository_test.go`

- 17 tests covering CRUD, matching, status transitions, precision

**Smoke Tests**: `test/smoke/contexts/banking/banktransaction/handler_test.go`

- 13 tests with mocked UseCases

### Critical Test: Amount Precision

```go
tx.AmountCents = 123456 // 1234.56 UAH
repo.Update(ctx, tx)
found, _ := repo.GetByID(ctx, tx.GetID())
assert.Equal(t, int64(123456), found.AmountCents,
    "Amount precision must be exact to the kopiyok!")
```

### Critical Test: Matching Workflow

```go
// Create transaction
tx, _ := NewBankTransaction(accountID, DirectionCredit, 100000, "UAH", time.Now(), "Payment", userID)
repo.Create(ctx, tx)

// Book it
_ = tx.Book()
repo.Update(ctx, tx)
assert.Equal(t, TransactionStatusBooked, tx.Status)

// Match to invoice
invoiceID := uuidv7.New()
_ = tx.Match(MatchedEntityInvoice, invoiceID)
repo.Update(ctx, tx)
assert.Equal(t, TransactionStatusReconciled, tx.Status)
assert.Equal(t, MatchedEntityInvoice, tx.MatchedEntityType)
assert.Equal(t, invoiceID, tx.MatchedEntityID)

// Unmatch
_ = tx.Unmatch()
repo.Update(ctx, tx)
assert.Equal(t, TransactionStatusBooked, tx.Status)
assert.Nil(t, tx.MatchedEntityType)
```

## Files

```
banktransaction/
  aggregate/
    bank_transaction.go          # Aggregate root + business logic
  repository/
    bank_transaction_repository.go # Repository interface
  adapter/
    repository/postgres/
      bank_transaction_repository.go # PostgreSQL implementation
    http/
      bank_transaction_handler.go    # HTTP handlers
  usecase/
    bank_transaction_usecase.go      # Use case implementations
  dto/
    bank_transaction_dto.go          # Request/response DTOs
  errors.go                          # Domain errors
  README.md                          # This file
```

## Reconciliation Workflow Example

```go
// 1. Import transactions from provider
transactions := providerClient.FetchTransactions(accountID)
for _, tx := range transactions {
    usecase.RecordTransactionWithExternalID(ctx, accountID, tx.ExternalID, ...)
}

// 2. List unmatched transactions
unmatched, _ := usecase.ListUnmatchedTransactions(ctx, accountID, 100, 0)

// 3. Match to invoices
for _, tx := range unmatched {
    invoice := findInvoiceByAmount(tx.AmountCents)
    if invoice != nil {
        usecase.MatchToInvoice(ctx, tx.GetID(), invoice.ID, userID)
    }
}

// 4. Verify reconciliation
remaining, _ := usecase.ListUnmatchedTransactions(ctx, accountID, 100, 0)
log.Printf("Unmatched: %d", len(remaining))
```

## Related Contexts

- **BankAccount** - Parent account
- **Billing** - Invoice payments
- **Order Management** - Order payments
- **Accounting** - Journal entries from transactions
