# Double-Entry Bookkeeping in Promenade

**Complete replacement for 1C/Bitrix24 accounting** - Automated journal entries from business events.

---

## Overview

Promenade implements **double-entry bookkeeping** (подвійна бухгалтерія) as a core financial capability. Every business transaction automatically generates accounting journal entries following strict double-entry rules.

**Why This Matters:**

- **Complete financial visibility**: Every hryvnia is tracked
- **Automated accounting**: No manual entry duplication
- **Audit trail**: Full history of all financial operations
- **1C replacement**: Professional accounting without external software

---

## Double-Entry Fundamentals

### The Golden Rule

**Every transaction has two sides**: For every debit, there must be an equal credit.

```
Σ Debits = Σ Credits
```

### Account Types & Balance Effects

| Account Type          | Normal Balance | Increases With | Decreases With |
| --------------------- | -------------- | -------------- | -------------- |
| **Asset** (Актив)     | Debit          | Debit          | Credit         |
| **Liability** (Пасив) | Credit         | Credit         | Debit          |
| **Equity** (Капітал)  | Credit         | Credit         | Debit          |
| **Revenue** (Дохід)   | Credit         | Credit         | Debit          |
| **Expense** (Витрата) | Debit          | Debit          | Credit         |

### Basic Accounting Equation

```
Assets = Liabilities + Equity
Активи = Пасиви + Капітал
```

**Revenue increases Equity**, **Expenses decrease Equity**:

```
Assets = Liabilities + Equity + (Revenue - Expenses)
```

---

## Ukrainian Chart of Accounts

Promenade uses simplified Ukrainian chart of accounts based on П(С)БО (National Accounting Standards).

### Key Accounts

#### 1. Assets (Активи)

```
30 - Грошові кошти (Cash & Bank)
  301 - Каса в національній валюті (Cash Register)
  311 - Розрахункові рахунки в банках (Bank Accounts)

36 - Розрахунки з покупцями (Accounts Receivable)
  361 - Розрахунки з вітчизняними покупцями (Domestic Customers)
```

#### 2. Equity (Капітал)

```
40 - Зареєстрований капітал (Registered Capital)
  401 - Статутний капітал (Share Capital)
```

#### 3. Liabilities (Зобов'язання)

```
63 - Розрахунки з постачальниками (Accounts Payable)
  631 - Розрахунки з вітчизняними постачальниками (Domestic Suppliers)
```

#### 4. Revenue (Доходи)

```
70 - Доходи від реалізації (Sales Revenue)
  702 - Дохід від реалізації товарів (Product Sales)
  703 - Дохід від реалізації робіт і послуг (Service Revenue)
```

#### 5. Expenses (Витрати)

```
90 - Собівартість реалізації (Cost of Sales)
  902 - Адміністративні витрати (Administrative Expenses)
93 - Витрати на збут (Selling Expenses)
```

---

## Automatic Journal Entries

### Business Event → Accounting Entry Flow

```
Business Operation (Banking, Billing, Fiscal)
         ↓
   Event Published (Event Bus)
         ↓
   Integration Handler Receives Event
         ↓
   Journal Entry Created (Draft)
         ↓
   Validation (Debits = Credits)
         ↓
   Entry Posted to Ledger
         ↓
   Account Balances Updated
```

---

## Integration Patterns

### 1. Bank Transaction (Incoming Payment)

**Business Event**: Customer payment received in bank account

**Event**: `bank.transaction.recorded`

```json
{
  "transaction_id": "uuid",
  "direction": "credit",
  "amount_cents": 10000,
  "currency_code": "UAH",
  "description": "Payment from customer ABC"
}
```

**Journal Entry**:

```
Дт 311 (Bank Account)       100.00 UAH
Кт 702 (Sales Revenue)      100.00 UAH

Description: Bank transaction: Payment from customer ABC
```

**Accounting Meaning**:

- **Asset increased** (more money in bank)
- **Revenue recognized** (earned income)

---

### 2. Bank Transaction (Outgoing Payment)

**Business Event**: Payment to supplier from bank account

**Event**: `bank.transaction.recorded`

```json
{
  "transaction_id": "uuid",
  "direction": "debit",
  "amount_cents": 5000,
  "currency_code": "UAH",
  "description": "Office supplies"
}
```

**Journal Entry**:

```
Дт 902 (Operating Expenses)  50.00 UAH
Кт 311 (Bank Account)        50.00 UAH

Description: Bank transaction: Office supplies
```

**Accounting Meaning**:

- **Expense recorded** (money spent on operations)
- **Asset decreased** (less money in bank)

---

### 3. Invoice Generated

**Business Event**: Invoice issued to customer

**Event**: `invoice.generated`

```json
{
  "invoice_id": "uuid",
  "customer_id": "uuid",
  "amount_cents": 20000,
  "currency_code": "UAH",
  "description": "Software subscription"
}
```

**Journal Entry**:

```
Дт 361 (Accounts Receivable)  200.00 UAH
Кт 702 (Sales Revenue)        200.00 UAH

Description: Invoice generated: Software subscription
```

**Accounting Meaning**:

- **Asset increased** (customer owes us money)
- **Revenue recognized** (earned income, accrual basis)

---

### 4. Payment Received for Invoice

**Business Event**: Customer paid invoice

**Event**: `payment.received`

```json
{
  "payment_id": "uuid",
  "invoice_id": "uuid",
  "amount_cents": 20000,
  "currency_code": "UAH",
  "payment_date": "2026-01-20T10:00:00Z"
}
```

**Journal Entry**:

```
Дт 311 (Bank Account)          200.00 UAH
Кт 361 (Accounts Receivable)   200.00 UAH

Description: Payment received for invoice #123
```

**Accounting Meaning**:

- **Asset transformation** (receivable → cash)
- **No new revenue** (already recognized when invoice created)

---

### 5. Cash Register Receipt (Fiscal)

**Business Event**: Cash sale recorded in cash register

**Event**: `fiscal.receipt.created`

```json
{
  "receipt_id": "uuid",
  "amount_cents": 15000,
  "currency_code": "UAH",
  "description": "Cash sale"
}
```

**Journal Entry**:

```
Дт 301 (Cash Register)    150.00 UAH
Кт 702 (Sales Revenue)    150.00 UAH

Description: Cash sale recorded
```

**Accounting Meaning**:

- **Asset increased** (cash in register)
- **Revenue recognized** (cash sale)

---

## Complete Business Scenario

### Example: Software Subscription Sale

**Timeline**:

1. **January 1**: Invoice issued (200 UAH)
2. **January 15**: Payment received (200 UAH)

#### Step 1: Invoice Issued (2026-01-01)

```
Дт 361 (Accounts Receivable)  200.00 UAH
Кт 702 (Sales Revenue)        200.00 UAH
```

**Balance Sheet Effect**:

- Assets: +200 UAH (Receivable)
- Equity: +200 UAH (Revenue → Retained Earnings)

**Income Statement Effect**:

- Revenue: +200 UAH

#### Step 2: Payment Received (2026-01-15)

```
Дт 311 (Bank Account)          200.00 UAH
Кт 361 (Accounts Receivable)   200.00 UAH
```

**Balance Sheet Effect**:

- Assets: +200 UAH (Bank), -200 UAH (Receivable) = **0 net change**
- No equity change (just asset transformation)

**Income Statement Effect**:

- No change (revenue already recognized)

#### Final Account Balances

| Account       | Debit | Credit | Balance     |
| ------------- | ----- | ------ | ----------- |
| 311 (Bank)    | 200   | -      | **+200 Дт** |
| 361 (A/R)     | 200   | 200    | **0**       |
| 702 (Revenue) | -     | 200    | **+200 Кт** |

---

## Financial Reports

### Trial Balance (Оборотна відомість)

Ensures accounting system is balanced:

```
Account                         Debit    Credit
-----------------------------------------------
301 (Cash Register)            1,500        -
311 (Bank Account)            10,200        -
361 (Accounts Receivable)      5,000        -
401 (Share Capital)                -   10,000
702 (Sales Revenue)                -   15,500
902 (Operating Expenses)       8,800        -
-----------------------------------------------
TOTAL                         25,500   25,500   Balanced
```

### Balance Sheet (Баланс)

```
ASSETS
  Cash & Bank:              11,700
  Accounts Receivable:       5,000
  Total Assets:             16,700

LIABILITIES
  (none)                         0

EQUITY
  Share Capital:            10,000
  Retained Earnings:         6,700
  Total Equity:             16,700

Assets = Liabilities + Equity  
16,700 = 0 + 16,700
```

### Income Statement (Звіт про прибутки)

```
REVENUE
  Sales Revenue:            15,500

EXPENSES
  Operating Expenses:        8,800

NET INCOME:                  6,700
```

---

## Implementation Architecture

### Components

1. **Account Aggregate** (`internal/contexts/accounting/account/`)
   - Chart of accounts
   - Account hierarchy
   - Account types

2. **JournalEntry Aggregate** (`internal/contexts/accounting/journalentry/`)
   - Double-entry transactions
   - Debit/Credit lines
   - Posting logic

3. **Ledger Aggregate** (`internal/contexts/accounting/ledger/`)
   - Account balance tracking
   - Period aggregation
   - Financial reports

4. **Integration Handlers** (`internal/contexts/accounting/integration/`)
   - Bank event handler
   - Billing event handler
   - Fiscal event handler

### Event-Driven Integration

```go
// Bank transaction arrives
bankTransaction := BankTransaction{
    Direction: "credit",
    Amount: 100 UAH,
}

// Event published
eventBus.Publish(ctx, "bank.transaction.recorded", bankTransaction)

// Integration handler receives event
func (h *BankEventHandler) HandleTransactionRecorded(event) {
    // Create journal entry
    entry := NewJournalEntry(...)
    entry.AddLine(account311, 10000, 0, "UAH", "Bank receipt")
    entry.AddLine(account702, 0, 10000, "UAH", "Sales revenue")

    // Post to ledger
    entry.Post(systemUserID)
    repository.Save(entry)
}
```

---

## Benefits vs 1C/Bitrix24

| Feature               | Promenade       | 1C         | Bitrix24         |
| --------------------- | --------------- | ---------- | ---------------- |
| **Automated entries** |  Event-driven |  Manual  |  No accounting |
| **Double-entry**      |  Native       |  Yes     |  No            |
| **Chart of accounts** |  Customizable |  Yes     |  No            |
| **Integration**       |  Built-in     |  Complex |  External      |
| **Audit trail**       |  Complete     |  Yes     |  Limited       |
| **Cost**              |  Free         |  License |  Subscription  |
| **Self-hosted**       |  Yes          |  Yes     |  Limited       |

---

## Testing Strategy

### Unit Tests

- Account creation and validation
- JournalEntry double-entry validation
- Balance calculations
- Entry posting logic

### Integration Tests

- End-to-end event flow
- Bank transaction → Journal entry
- Invoice → Payment → Journal entries
- Account balance updates

### Smoke Tests

- API endpoints validation
- DTO validation
- Mock use case layer

---

## Next Steps

### Phase 1: Core Implementation (Current)

-  Account aggregate
-  JournalEntry aggregate
-  Ledger aggregate
-  Database migrations
-  Integration handlers
-  Repositories and use cases
-  HTTP handlers

### Phase 2: Automation

- Event subscriber registration
- Automatic posting rules
- Period closing automation

### Phase 3: Reporting

- Trial balance
- Balance sheet
- Income statement
- Cash flow statement

---

## Related Documentation

- [Accounting Context README](../internal/contexts/accounting/README.md) - Full context documentation
- [Banking Context README](../internal/contexts/banking/README.md) - Bank integration
- [Billing Context README](../internal/contexts/billing/README.md) - Invoice & payment integration
- [Event Bus README](../pkg/bus/README.md) - Event-driven architecture

---

**Status**:  Design Complete, Ready for Implementation  
**Target**: Complete 1C/Bitrix24 replacement with professional accounting
