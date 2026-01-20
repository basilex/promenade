# BankAccount Aggregate

Bank account management with support for manual accounts and external provider integration.

## Quick Navigation

- [Overview](#overview) - Purpose and key features
- [Domain Model](#domain-model) - Properties and types
- [Business Rules](#business-rules) - Domain constraints
- [Use Cases](#use-cases) - Available operations
- [API Examples](#api-examples) - Usage patterns

 **Related**: [../README.md](../README.md) - Banking Context Overview

---

## Overview

**BankAccount** represents a bank account that can be:

- **Manual** - Created and managed within the system
- **Provider-connected** - Synchronized with external bank (Monobank, Privat24, PUMB)

### Key Features

- Multi-provider support (Manual, Monobank, Privat24, PUMB)
- Balance tracking with kopiyky precision
- Account lifecycle management (Active/Inactive/Archived)
- Provider synchronization status
- IBAN and account number storage

## Domain Model

### Properties

```go
type BankAccount struct {
    aggregate.BaseAggregate

    // Ownership
    OrganizationID uuidv7.UUID

    // Account Details
    Name           string  // Display name
    BankName       string  // Bank institution name
    IBAN           string  // International Bank Account Number
    AccountNumber  string  // Local account number
    CurrencyCode   string  // UAH, USD, EUR

    // Provider Integration
    Provider          BankProvider  // manual, monobank, privat24, pumb
    ProviderAccountID string        // External provider account ID

    // Status & Balance
    Status       BankAccountStatus  // active, inactive, archived
    BalanceCents int64             // Balance in kopiyky
    LastSyncAt   *time.Time        // Last provider sync timestamp

    // Metadata
    LastUpdatedBy uuidv7.UUID
}
```

### Types

**BankProvider**:

- `ProviderManual` - Manually managed account
- `ProviderMonobank` - Monobank API integration
- `ProviderPrivat24` - PrivatBank API integration
- `ProviderPUMB` - PUMB Bank API integration

**BankAccountStatus**:

- `BankAccountStatusActive` - Active, can accept transactions
- `BankAccountStatusInactive` - Temporarily disabled
- `BankAccountStatusArchived` - Permanently archived

## Business Rules

### 1. Provider Constraints

- Manual accounts cannot be connected to providers
- Provider accounts require valid `provider_account_id`
- Provider type cannot be changed after creation
- Only provider accounts can be synchronized

### 2. Status Transitions

```
Active ←→ Inactive
  ↓
Archived (final state)
```

- Active → Inactive: `Deactivate()`
- Inactive → Active: `Activate()`
- Active/Inactive → Archived: `Archive()`
- Archived accounts cannot be reactivated

### 3. Balance Precision

- Balance always stored in cents (kopiyky)
- Example: 1234.56 UAH = 123456 kopiyky
- Ensures no floating-point precision loss
- Synchronization updates balance from provider

### 4. Validation Rules

- `name` - Required, max 255 characters
- `bank_name` - Required, max 255 characters
- `currency_code` - Required, ISO 4217 (UAH, USD, EUR)
- `IBAN` - Optional, valid format if provided
- `provider_account_id` - Required for provider accounts

## Use Cases

### Account Management

**CreateManualAccount**(organizationID, name, bankName, currencyCode, lastUpdatedBy)

- Creates manually managed account
- Sets Provider = ProviderManual
- Initial status = Active

**ConnectProviderAccount**(organizationID, name, bankName, provider, providerAccountID, lastUpdatedBy)

- Creates provider-connected account
- Requires valid provider and providerAccountID
- Enables automatic synchronization

**GetAccount**(id) → BankAccount

- Retrieves account by ID
- Returns ErrAccountNotFound if not exists

**ListAccounts**(organizationID, limit, offset) → []BankAccount

- Lists organization accounts
- Supports pagination

**UpdateAccountDetails**(id, name, bankName, iban, accountNumber, lastUpdatedBy)

- Updates account information
- Validates IBAN format if provided

**DeleteAccount**(id)

- Soft deletes account
- Preserves transaction history

### Balance Operations

**UpdateBalance**(id, balanceCents, lastUpdatedBy)

- Manually updates account balance
- Used for manual accounts or corrections

**RecordProviderSync**(id, balanceCents, lastUpdatedBy)

- Records provider synchronization
- Updates balance and LastSyncAt timestamp
- Only for provider-connected accounts

### Status Management

**ActivateAccount**(id, lastUpdatedBy)

- Activates inactive account
- Cannot activate archived accounts
- Returns ErrAccountArchived if archived

**DeactivateAccount**(id, lastUpdatedBy)

- Deactivates active account
- Temporarily disables transactions

**ArchiveAccount**(id, lastUpdatedBy)

- Permanently archives account
- Cannot be reversed
- Historical data preserved

## API Examples

### Creating Manual Account

```bash
POST /api/v1/banking/accounts
{
  "organization_id": "01JGXYZ...",
  "name": "UAH Current Account",
  "bank_name": "PrivatBank",
  "currency_code": "UAH",
  "last_updated_by": "01JGXYZ..."
}
```

### Connecting Provider Account

```bash
POST /api/v1/banking/accounts/provider
{
  "organization_id": "01JGXYZ...",
  "name": "Monobank Business",
  "bank_name": "Monobank",
  "provider": "monobank",
  "provider_account_id": "mono_abc123",
  "last_updated_by": "01JGXYZ..."
}
```

### Updating Balance

```bash
PUT /api/v1/banking/accounts/:id/balance
{
  "balance_cents": 500000,  // 5000.00 UAH
  "last_updated_by": "01JGXYZ..."
}
```

### Recording Sync

```bash
PUT /api/v1/banking/accounts/:id/sync
{
  "balance_cents": 502345,  // 5023.45 UAH from provider
  "last_updated_by": "01JGXYZ..."
}
```

## Testing

**Integration Tests**: `test/integration/contexts/banking/bankaccount/repository_test.go`

- 10 tests covering CRUD, providers, balance precision, status transitions

**Smoke Tests**: `test/smoke/contexts/banking/bankaccount/handler_test.go`

- 10 tests with mocked UseCases

### Critical Test: Balance Precision

```go
acc.BalanceCents = 123456 // 1234.56 UAH
repo.Update(ctx, acc)
found, _ := repo.GetByID(ctx, acc.GetID())
assert.Equal(t, int64(123456), found.BalanceCents,
    "Balance precision must be exact to the kopiyok!")
```

## Files

```
bankaccount/
  aggregate/
    bank_account.go          # Aggregate root + business logic
  repository/
    bank_account_repository.go # Repository interface
  adapter/
    repository/postgres/
      bank_account_repository.go # PostgreSQL implementation
    http/
      bank_account_handler.go    # HTTP handlers
  usecase/
    bank_account_usecase.go      # Use case implementations
  dto/
    bank_account_dto.go          # Request/response DTOs
  errors.go                      # Domain errors
  README.md                      # This file
```

## Related Contexts

- **BankTransaction** - Transactions for this account
- **Billing** - Invoice payments via bank accounts
- **Accounting** - Account balance tracking in ledger
