# Banking Context

**Banking** - bounded context for managing bank accounts and transactions with external bank provider integration.

## Quick Navigation

📋 **Design & Concepts**

- [Architecture Overview](#architecture) - Clean architecture layers and structure
- [Domain Model](#domain-model) - BankAccount and BankTransaction aggregates
- [Business Rules](#business-rules) - Domain constraints and workflows
- [Provider Integration](#provider-integration) - External bank connections

📚 **Implementation Details**

- [Database Schema](#database-schema) - Table structures and relationships
- [API Endpoints](#api-endpoints) - REST API documentation
- [Testing Strategy](#testing) - Test structure and patterns
- [Implementation Status](#implementation-status) - Current progress

🔗 **Related Documentation**

- [../../docs/concepts/banking-integration.md](../../docs/concepts/banking-integration.md) - Banking integration patterns
- [../../docs/guides/money-precision.md](../../docs/guides/money-precision.md) - Money handling best practices
- [./bankaccount/README.md](./bankaccount/README.md) - BankAccount aggregate details
- [./banktransaction/README.md](./banktransaction/README.md) - BankTransaction aggregate details

---

## Purpose

Provides functionality for:

- Managing bank accounts (manual and provider-connected)
- Recording bank transactions (debits/credits)
- Matching transactions to invoices/orders/payments
- Synchronizing with external providers (Monobank, Privat24, PUMB)
- Bank statement import and processing

## Aggregates

### BankAccount

- **Location**: `bankaccount/aggregate/bank_account.go`
- **Purpose**: Represents a bank account connected to external provider or managed manually
- **Key Features**:
  - Support for multiple providers (Manual, Monobank, Privat24, PUMB)
  - Balance tracking
  - Synchronization status
  - Account lifecycle (Active, Inactive, Archived)

### BankTransaction

- **Location**: `banktransaction/aggregate/bank_transaction.go`
- **Purpose**: Represents a single bank transaction (debit or credit)
- **Key Features**:
  - Direction tracking (debit/credit)
  - Transaction status (Pending, Booked, Canceled, Reconciled)
  - Matching to business entities (Invoice, Order, Payment)
  - Counterparty information

## Architecture

```
banking/
  bankaccount/              # Bank Account Aggregate
    aggregate/
      bank_account.go       # Aggregate root
    repository/
      bank_account_repository.go  # Repository interface
    adapter/
      repository/postgres/
        bank_account_repository.go  # PostgreSQL implementation
    usecase/                # Use cases
    dto/                    # DTOs
    errors.go               # Domain errors

  banktransaction/          # Bank Transaction Aggregate
    aggregate/
      bank_transaction.go   # Aggregate root
    repository/
      bank_transaction_repository.go  # Repository interface
    adapter/
      repository/postgres/
        bank_transaction_repository.go  # PostgreSQL implementation
    usecase/                # Use cases
    dto/                    # DTOs
    errors.go               # Domain errors

  router.go                 # HTTP routes
  README.md                 # This file
```

## Domain Model

### BankAccount

**Properties**:

- `OrganizationID` - Organization owner
- `Name` - Account display name
- `BankName` - Bank institution name
- `IBAN` - International Bank Account Number
- `AccountNumber` - Local account number
- `CurrencyCode` - Currency (UAH, USD, EUR)
- `Provider` - Bank provider (manual, monobank, privat24, pumb)
- `ProviderAccountID` - External provider account ID
- `Status` - Account status (active, inactive, archived)
- `BalanceCents` - Current balance in cents
- `LastSyncAt` - Last synchronization timestamp

## Business Rules

### BankAccount Business Rules

1. **Provider Constraints**
   - Manual accounts cannot be synced with external providers
   - Provider-connected accounts require `provider_account_id`
   - Provider type cannot be changed after creation

2. **Status Transitions**
   - Active → Inactive: Deactivate()
   - Inactive → Active: Activate()
   - Active/Inactive → Archived: Archive()
   - Archived accounts cannot be activated

3. **Balance Management**
   - Balance is always stored in cents for precision (kopiyky)
   - Example: 1234.56 UAH = 123456 kopiyky
   - Synchronization updates balance from provider

### BankTransaction Business Rules

1. **Amount Precision**
   - Amount must be positive
   - Stored in cents/kopiyky for exact precision
   - Example: 1234.56 UAH = 123456 kopiyky

2. **Direction Rules**
   - Direction must be `debit` or `credit`
   - Cannot be changed after creation

3. **Status Workflow**
   - Pending → Booked: `Book()`
   - Pending → Canceled: `Cancel()`
   - Booked → Reconciled: `Match()` to entity
   - Canceled transactions cannot be booked or matched

4. **Matching Rules**
   - Transaction must be booked before matching
   - Can match to: Invoice, Order, or Payment
   - Matched transactions cannot be matched again (must unmatch first)
   - Unmatching returns status to Booked

5. **External ID Uniqueness**
   - `external_id` must be unique per account
   - Used for provider synchronization deduplication

## Domain Model

### BankAccount

**Properties**:

- `BankAccountID` - Related bank account
- `StatementID` - Related bank statement
- `ExternalID` - Provider's transaction ID
- `Direction` - Debit or Credit
- `AmountCents` - Amount in cents
- `CurrencyCode` - Currency
- `CounterpartyName` - Counterparty name
- `CounterpartyIBAN` - Counterparty IBAN
- `Description` - Transaction description
- `TransactionAt` - Transaction timestamp
- `BookedAt` - Booking timestamp
- `Status` - Transaction status
- `MatchedEntityType` - Type of matched entity (invoice, order, payment)
- `MatchedEntityID` - ID of matched entity
- `MatchedAt` - Matching timestamp

## Use Cases

### BankAccount Use Cases

**Account Management**:

- `CreateManualAccount` - Create manually managed account
- `ConnectProviderAccount` - Connect to external bank provider
- `GetAccount` - Retrieve account by ID
- `ListAccounts` - List accounts for organization
- `UpdateAccountDetails` - Update account information
- `DeleteAccount` - Soft delete account

**Balance & Sync**:

- `UpdateBalance` - Update account balance manually
- `RecordProviderSync` - Record provider synchronization

**Status Management**:

- `ActivateAccount` - Activate inactive account
- `DeactivateAccount` - Deactivate active account
- `ArchiveAccount` - Archive account permanently

### BankTransaction Use Cases

**Transaction Management**:

- `RecordTransaction` - Record new transaction
- `RecordTransactionWithExternalID` - Record from provider sync
- `GetTransaction` - Retrieve transaction by ID
- `ListTransactionsByAccount` - List account transactions
- `ListUnmatchedTransactions` - List unreconciled transactions
- `DeleteTransaction` - Soft delete transaction

**Transaction Lifecycle**:

- `BookTransaction` - Book pending transaction
- `CancelTransaction` - Cancel pending transaction

**Matching & Reconciliation**:

- `MatchToInvoice` - Match transaction to invoice
- `MatchToOrder` - Match transaction to order
- `MatchToPayment` - Match transaction to payment
- `UnmatchTransaction` - Remove entity match

**Transaction Details**:

- `SetCounterparty` - Update counterparty information
- `UpdateTransactionDescription` - Update description

## Business Rules

### banking_bank_accounts

- Primary key: `id` (UUID v7)
- Optimistic locking: `version`
- Soft delete: `deleted_at`
- Indexes: organization_id, provider+provider_account_id, status

### banking_bank_transactions

- Primary key: `id` (UUID v7)
- Foreign keys: `bank_account_id`, `statement_id`
- Optimistic locking: `version`
- Soft delete: `deleted_at`
- Indexes: bank_account_id, external_id, matched_entity_type+matched_entity_id

## Provider Integration

### Supported Providers

1. **Monobank** - Ukrainian monobank API
2. **Privat24** - PrivatBank API
3. **PUMB** - PUMB Bank API

### Integration Flow

1. Connect account via provider API credentials
2. Store provider_account_id
3. Sync transactions periodically
4. Match transactions to business entities
5. Reconcile accounts

## API Endpoints (Planned)

### Bank Accounts

- `POST /api/v1/banking/accounts` - Create account
- `GET /api/v1/banking/accounts/:id` - Get account
- `PUT /api/v1/banking/accounts/:id` - Update account
- `DELETE /api/v1/banking/accounts/:id` - Delete account
- `GET /api/v1/banking/accounts` - List accounts
- `POST /api/v1/banking/accounts/:id/connect` - Connect provider
- `POST /api/v1/banking/accounts/:id/sync` - Sync transactions

### Bank Transactions

- `POST /api/v1/banking/transactions` - Create transaction
- `GET /api/v1/banking/transactions/:id` - Get transaction
- `PUT /api/v1/banking/transactions/:id` - Update transaction
- `DELETE /api/v1/banking/transactions/:id` - Delete transaction
- `GET /api/v1/banking/transactions` - List transactions
- `POST /api/v1/banking/transactions/:id/match` - Match to entity
- `POST /api/v1/banking/transactions/:id/unmatch` - Unmatch
- `POST /api/v1/banking/transactions/:id/book` - Book transaction
- `POST /api/v1/banking/transactions/:id/reconcile` - Reconcile

## Testing

- **Unit Tests**: `bankaccount/aggregate/*_test.go`, `banktransaction/aggregate/*_test.go`
- **Integration Tests**: `test/integration/contexts/banking/*_test.go`
- **Smoke Tests**: `test/smoke/contexts/banking/*_test.go`

## Dependencies

**Internal**:

- `pkg/aggregate` - Base aggregate pattern
- `pkg/uuidv7` - UUID v7 generation
- `pkg/jsonstore` - JSON field handling
- `internal/infrastructure/database` - Database utilities

**External**:

- `jmoiron/sqlx` - SQL extensions
- Provider SDKs (Monobank, Privat24, PUMB)

## Implementation Status

- ✅ **BankAccount aggregate** - Entity + business rules complete
- ✅ **BankTransaction aggregate** - Entity + business rules complete
- ✅ **Repository interfaces** - IBankAccountRepository, IBankTransactionRepository
- ✅ **PostgreSQL implementations** - Full CRUD + domain queries
- ✅ **Domain errors** - Typed error constants
- ✅ **Use cases** - All use cases implemented
- ✅ **DTOs** - Request/response data transfer objects
- ✅ **HTTP handlers** - REST API endpoints complete
- ✅ **Integration tests** - 27 tests (10 BankAccount + 17 BankTransaction)
- ✅ **Smoke tests** - 23 tests (10 BankAccount + 13 BankTransaction)
- ⏳ **Provider adapters** - Monobank/Privat24/PUMB (planned)
- ⏳ **Statement import** - CSV/Excel import (planned)

### Test Coverage

**Integration Tests** (Repository layer with real PostgreSQL):

- ✅ `test/integration/contexts/banking/bankaccount/repository_test.go` - 10 tests
- ✅ `test/integration/contexts/banking/banktransaction/repository_test.go` - 17 tests
- 💰 **Money precision tests** - Exact kopiyky validation

**Smoke Tests** (Handler layer with mocked UseCases):

- ✅ `test/smoke/contexts/banking/bankaccount/handler_test.go` - 10 tests
- ✅ `test/smoke/contexts/banking/banktransaction/handler_test.go` - 13 tests

See [test/smoke/contexts/banking/README.md](../../../test/smoke/contexts/banking/README.md) for detailed test documentation.

## Related Contexts

- **Billing**: Transactions matched to invoices/payments
- **Order Management**: Transactions matched to orders
- **Accounting**: Transactions posted to journal entries

## Future Enhancements

- Real-time transaction webhooks from providers
- Automatic transaction categorization via ML
- Bank reconciliation reports
- Multi-currency support
- Commission and fee tracking
- Cash flow forecasting
