# Accounting Context

**Accounting** - bounded context for complete accounting system with double-entry bookkeeping, financial management, and reporting.

## Quick Navigation

**Design & Implementation**

- [Architecture Overview](#architecture) - Clean architecture layers and structure
- [Domain Model](#domain-model) - 7 aggregates with business logic
- [HTTP API](#http-api) - 78 REST endpoints with Swagger documentation
- [Performance Optimizations](#performance-optimizations) - Caching, indexing, and materialized views

  **Modules**

- [Chart of Accounts](#1-chart-of-accounts) - Account hierarchy management
- [Journal Entries](#2-journal-entries) - Double-entry bookkeeping
- [Fiscal Periods](#3-fiscal-periods) - Period management and closing
- [Tax Codes](#4-tax-codes) - Tax calculation support
- [Budgets](#5-budgets) - Budget planning and tracking
- [Cost Centers](#6-cost-centers) - Cost and profit center management
- [Bank Reconciliation](#7-bank-reconciliation) - Statement reconciliation

  **Related Documentation**

- [../../docs/concepts/accounting-double-entry.md](../../docs/concepts/accounting-double-entry.md) - Double-entry patterns
- [../../docs/guides/money-precision.md](../../docs/guides/money-precision.md) - Money handling
- [./audit/README.md](./audit/README.md) - Audit logging system
- [./integration/README.md](./integration/README.md) - Event-driven integration

---

## Purpose

Complete accounting system providing:

- **Chart of Accounts** - Hierarchical account management (Plan de Cuentas)
- **Journal Entries** - Double-entry bookkeeping with validation
- **Fiscal Periods** - Period management with lock/close controls
- **Tax Codes** - Tax calculation and reporting
- **Budgets** - Budget planning with variance tracking
- **Cost Centers** - Cost allocation and profitability analysis
- **Bank Reconciliation** - Bank statement reconciliation
- **Audit Trail** - Complete before/after state tracking
- **Event Sourcing** - Journal entry event store
- **Caching** - Performance-optimized with 5-15 min TTL

## Why Accounting Context?

**Complete ERP replacement** requires proper accounting foundation:

```
BankTransaction → Dr 311 (Bank Accounts) / Cr 702 (Sales Revenue)
Invoice         → Dr 361 (Accounts Receivable) / Cr 702 (Sales Revenue)
Payment         → Dr 311 (Bank Accounts) / Cr 361 (Accounts Receivable)
Receipt         → Dr 301 (Cash Register) / Cr 702 (Sales Revenue)
```

**Ukrainian Accounting Standards (P(S)BU) compliant** with:

- Double-entry bookkeeping
- Chart of Accounts
- General Ledger
- Trial Balance

---

## Architecture

### Directory Structure

```
accounting/
 router.go                    # HTTP routes registration
 README.md                    # This file

 account/                     # Chart of Accounts
    aggregate/account.go
    repository/
    adapter/
       repository/postgres/
       http/account_handler.go
    usecase/account_usecase.go
    dto/account_dto.go
    cache/account_cache.go
    errors.go

 journalentry/                # Journal Entries
    aggregate/
    repository/
    adapter/http/journal_entry_handler.go
    usecase/journal_entry_usecase.go
    dto/journal_entry_dto.go
    eventstore/event_store.go
    errors.go

 fiscalperiod/                # Fiscal Periods
    aggregate/fiscal_period.go
    adapter/http/fiscal_period_handler.go
    usecase/fiscal_period_usecase.go
    dto/fiscal_period_dto.go
    cache/fiscal_period_cache.go
    errors.go

 taxcode/                     # Tax Codes
    aggregate/tax_code.go
    adapter/http/tax_code_handler.go
    usecase/tax_code_usecase.go
    dto/tax_code_dto.go
    cache/tax_code_cache.go
    errors.go

 budget/                      # Budgets
    aggregate/budget.go
    adapter/http/budget_handler.go
    usecase/budget_usecase.go
    dto/budget_dto.go
    errors.go

 costcenter/                  # Cost Centers
    aggregate/cost_center.go
    adapter/http/cost_center_handler.go
    usecase/cost_center_usecase.go
    dto/cost_center_dto.go
    errors.go

 reconciliation/                  # Bank Reconciliation
    types.go                       # Status, TransactionType
    aggregate/reconciliation.go
    repository/reconciliation_repository.go
    adapter/
       repository/postgres/reconciliation_repository.go
       http/reconciliation_handler.go
    usecase/reconciliation_usecase.go
    dto/reconciliation_dto.go
    errors.go

 audit/                       # Audit System
    audit_logger.go

 integration/                 # Event-driven integration
     bank_event_handler.go
     billing_event_handler.go
     fiscal_event_handler.go
```

### Clean Architecture Layers

```

  Presentation Layer (HTTP Handlers)
  - DTOs, Request/Response validation
  - Swagger documentation
  - Error mapping (domain → HTTP)

                       ↓

  Application Layer (Use Cases)
  - Business logic orchestration
  - Transaction management
  - Cache integration
  - Audit logging

                       ↓

  Domain Layer (Aggregates)
  - Business rules validation
  - Domain events
  - Aggregate invariants

                       ↓

  Infrastructure Layer (Repositories, Cache, Events)
  - PostgreSQL/SQLite repositories
  - In-memory caching (5-15 min TTL)
  - Event store (journal entries)
  - Audit logger

```

---

## Domain Model

### Summary Table

| Aggregate          | Endpoints | Status | Cache TTL | Features                       |
| ------------------ | --------- | ------ | --------- | ------------------------------ |
| **Account**        | 10        |        | 5 min     | Hierarchical, Type validation  |
| **JournalEntry**   | 10        |        | -         | Event sourcing, Audit log      |
| **FiscalPeriod**   | 10        |        | 10 min    | Close/Lock workflow            |
| **TaxCode**        | 12        |        | 15 min    | Tax calculation                |
| **Budget**         | 13        |        | -         | Approval workflow, Variance    |
| **CostCenter**     | 12        |        | -         | Hierarchical, Manager tracking |
| **Reconciliation** | 11        |        | -         | Item matching, Approval        |

---

### 1. Chart of Accounts

**Aggregate**: `Account` ([account/aggregate/account.go](account/aggregate/account.go))

**Purpose**: Hierarchical chart of accounts following Ukrainian accounting standards.

**Key Properties**:

- `Code` - Account code (e.g., "311", "702")
- `Name` - Account name
- `AccountType` - asset, liability, equity, revenue, expense
- `ParentID` - Hierarchical structure
- `IsActive` - Active status
- `Level` - Hierarchy depth

**Business Rules**:

- Code must be unique per organization
- Cannot delete account with child accounts
- Cannot be own parent

**Cache**: 5 min TTL (byID, byCode, byOrg)

**Endpoints**: 10 (CRUD, parent management, activate/deactivate, type filtering)

---

### 2. Journal Entries

**Aggregate**: `JournalEntry` ([journalentry/aggregate/journal_entry.go](journalentry/aggregate/journal_entry.go))

**Purpose**: Double-entry bookkeeping transactions with complete audit trail.

**Key Properties**:

- `EntryDate` - Transaction date
- `Status` - draft, posted, reversed
- `Lines` - []JournalEntryLine (debit/credit pairs)
- `SourceType` - Source event type
- `SourceID` - Source entity ID

**Business Rules**:

- **Golden Rule**: Sum(Debits) = Sum(Credits)
- Must have at least 2 lines
- Cannot modify posted entries
- Fiscal period must be open

**Event Sourcing**: 6 event types (Created, LineAdded, Posted, Reversed, etc.)

**Endpoints**: 10 (CRUD, line management, post, reverse, period filtering)

---

### 3. Fiscal Periods

**Aggregate**: `FiscalPeriod` ([fiscalperiod/aggregate/fiscal_period.go](fiscalperiod/aggregate/fiscal_period.go))

**Purpose**: Manage accounting periods with close/lock controls.

**Key Properties**:

- `Name` - e.g., "January 2026"
- `StartDate`, `EndDate` - Period boundaries
- `Status` - open, closed, locked
- `Year`, `Quarter` - Grouping

**Business Rules**:

- Cannot create overlapping periods
- Cannot post to closed/locked periods
- Cannot reopen locked period

**Cache**: 10 min TTL (byID, byDate, byOrg)

**Endpoints**: 10 (CRUD, close/reopen/lock, date queries)

---

### 4. Tax Codes

**Aggregate**: `TaxCode` ([taxcode/aggregate/tax_code.go](taxcode/aggregate/tax_code.go))

**Purpose**: Tax calculation support for VAT, income tax, etc.

**Key Properties**:

- `Code` - e.g., "VAT20", "PDV20"
- `TaxType` - vat, income, withholding, sales, custom
- `Rate` - e.g., 20.0 for 20%
- `TaxAccountID` - Account to post tax

**Business Rules**:

- Rate must be between 0 and 100
- Tax account must be liability or expense

**Cache**: 15 min TTL (longest - tax data is static)

**Endpoints**: 12 (CRUD, account associations, type filtering, calculate)

---

### 5. Budgets

**Aggregate**: `Budget` ([budget/aggregate/budget.go](budget/aggregate/budget.go))

**Purpose**: Budget planning with variance tracking and approval workflow.

**Key Properties**:

- `Name` - e.g., "2026 Operating Budget"
- `FiscalYear` - Year reference
- `Status` - draft, approved, active, closed
- `Lines` - []BudgetLine with actual vs budget

**Business Rules**:

- Cannot modify approved/active budgets
- Cannot activate without approval
- Must have at least one line

**Workflow**: Draft → Approve → Activate → Close

**Endpoints**: 13 (CRUD, line management, approve/activate/close, status filtering)

---

### 6. Cost Centers

**Aggregate**: `CostCenter` ([costcenter/aggregate/cost_center.go](costcenter/aggregate/cost_center.go))

**Purpose**: Cost allocation and profitability analysis with hierarchical structure.

**Key Properties**:

- `Code` - e.g., "CC-SALES", "PC-RETAIL"
- `CenterType` - cost_center, profit_center, investment_center
- `ParentID` - Hierarchical structure
- `ManagerID` - Responsible manager

**Business Rules**:

- Cannot delete center with children
- Cannot be own parent
- Parent must be same or higher level type

**Endpoints**: 12 (CRUD, parent/manager management, hierarchical queries, type filtering)

---

### 7. Bank Reconciliation

**Aggregate**: `Reconciliation` ([reconciliation/aggregate/reconciliation.go](reconciliation/aggregate/reconciliation.go))

**Purpose**: Reconcile bank statements with accounting records.

**Key Properties**:

- `BankStatementBalanceCents` - Statement balance
- `BookBalanceCents` - Book balance
- `Status` - in_progress, completed, approved (from reconciliation.Status)
- `Items` - []Item with matching

**Business Rules**:

- Must match all items before completion
- Cannot reopen approved reconciliation
- Adjusted balance must match statement

**Formula**: Book + Deposits - Checks - Fees + Interest = Statement

**Endpoints**: 11 (CRUD, item management, match/complete/reopen, bank account queries)

---

## HTTP API

### Complete Endpoint Summary

**78 REST Endpoints** across 7 handlers with full Swagger documentation.

| Module             | Base Path                            | Endpoints | Key Operations                                                |
| ------------------ | ------------------------------------ | --------- | ------------------------------------------------------------- |
| **Account**        | `/api/v1/accounting/accounts`        | 10        | Create, Update, SetParent, Activate, Deactivate, List by type |
| **JournalEntry**   | `/api/v1/accounting/journal-entries` | 10        | Create, AddLine, RemoveLine, Post, Reverse, List by period    |
| **FiscalPeriod**   | `/api/v1/accounting/fiscal-periods`  | 10        | Create, Update, Close, Reopen, Lock, Get by date              |
| **TaxCode**        | `/api/v1/accounting/tax-codes`       | 12        | Create, Update, SetTaxAccount, Calculate, List by type        |
| **Budget**         | `/api/v1/accounting/budgets`         | 13        | Create, AddLine, Approve, Activate, Close, List by status     |
| **CostCenter**     | `/api/v1/accounting/cost-centers`    | 12        | Create, SetParent, SetManager, Activate, List children        |
| **Reconciliation** | `/api/v1/accounting/reconciliations` | 11        | Create, AddItem, MarkMatched, Complete, Reopen, List by bank  |

### Authentication

All endpoints require JWT authentication:

```bash
Authorization: Bearer <token>
```

Organization context extracted from token (`organization_id`).

### Example Requests

#### Create Account

```bash
POST /api/v1/accounting/accounts
Content-Type: application/json

{
  "code": "311",
  "name": "Bank Accounts",
  "account_type": "asset",
  "currency_code": "UAH"
}
```

#### Create Journal Entry with Lines

```bash
POST /api/v1/accounting/journal-entries
{
  "fiscal_period_id": "{period-id}",
  "entry_date": "2026-01-20",
  "description": "Initial capital",
  "reference_number": "JE-001"
}

POST /api/v1/accounting/journal-entries/{id}/lines
{
  "account_id": "{311-id}",
  "debit_amount_cents": 100000000,
  "credit_amount_cents": 0,
  "currency_code": "UAH",
  "description": "Dr 311"
}

POST /api/v1/accounting/journal-entries/{id}/post
```

#### Calculate Tax

```bash
POST /api/v1/accounting/tax-codes/{id}/calculate
{
  "base_amount_cents": 100000
}

# Response:
{
  "base_amount_cents": 100000,
  "tax_amount_cents": 20000,
  "total_amount_cents": 120000,
  "rate": 20.0
}
```

---

## Performance Optimizations

### Database Indexing

**12 Composite Indices**:

```sql
-- Account lookups (3)
CREATE INDEX idx_accounts_org_code ON accounting_accounts(organization_id, code);
CREATE INDEX idx_accounts_org_type_active ON accounting_accounts(organization_id, account_type, is_active);
CREATE INDEX idx_accounts_parent_level ON accounting_accounts(parent_id, level);

-- Journal entry queries (3)
CREATE INDEX idx_je_org_period_date ON accounting_journal_entries(organization_id, fiscal_period_id, entry_date);
CREATE INDEX idx_je_status_date ON accounting_journal_entries(status, entry_date);
CREATE INDEX idx_je_lines_account_entry ON accounting_journal_entry_lines(account_id, journal_entry_id);

-- Fiscal period date range (2)
CREATE INDEX idx_fp_org_dates ON accounting_fiscal_periods(organization_id, start_date, end_date);
CREATE INDEX idx_fp_year_quarter ON accounting_fiscal_periods(year, quarter);

-- Tax code lookups (2)
CREATE INDEX idx_tc_org_code ON accounting_tax_codes(organization_id, code);
CREATE INDEX idx_tc_type_active ON accounting_tax_codes(tax_type, is_active);

-- Budget queries (1)
CREATE INDEX idx_budget_org_year_status ON accounting_budgets(organization_id, fiscal_year, status);

-- Cost center hierarchy (1)
CREATE INDEX idx_cc_org_type_active ON accounting_cost_centers(organization_id, center_type, is_active);
```

### Materialized Views

**3 Materialized Views** for reporting:

1. `accounting_account_balances` - Account balances by period
2. `accounting_budget_variance` - Budget vs actual variance
3. `accounting_trial_balance` - Trial balance report

Refresh: `REFRESH MATERIALIZED VIEW CONCURRENTLY`

### Caching

**Three-level Strategy**:

| Cache             | TTL    | Keys                 | Use Case                 |
| ----------------- | ------ | -------------------- | ------------------------ |
| AccountCache      | 5 min  | byID, byCode, byOrg  | Frequent account lookups |
| FiscalPeriodCache | 10 min | byID, byDate, byOrg  | Period validation        |
| TaxCodeCache      | 15 min | byID, byCode, byType | Tax calculations         |

---

## Audit & Event Sourcing

### Audit Logger

**Complete Before/After State Tracking**:

```go
type AuditRecord struct {
    EntityType      string  // "account", "journal_entry", etc.
    EntityID        uuidv7.UUID
    Action          string  // "create", "update", "delete"
    UserID          uuidv7.UUID
    BeforeState     map[string]interface{}
    AfterState      map[string]interface{}
    ChangedFields   []string
    Timestamp       time.Time
}
```

All use cases integrated with audit logging.

### Event Sourcing

**Journal Entry Event Store** with 6 event types:

- `journal_entry_created`
- `journal_entry_line_added`
- `journal_entry_line_removed`
- `journal_entry_posted`
- `journal_entry_reversed`
- `journal_entry_description_updated`

Events enable:

- Complete audit trail
- Event replay (rebuild aggregate from events)
- Debugging and troubleshooting

---

## Implementation Status

### Phase 1: Core Implementation (COMPLETE)

- [x] 7 Aggregates with business logic
- [x] 7 Repositories (PostgreSQL implementation)
- [x] 7 Use Cases with cache & audit
- [x] 7 HTTP Handlers with Swagger docs
- [x] 78 REST Endpoints
- [x] 8 Migrations (6 base + 2 optimizations)
- [x] 12 Composite Indices
- [x] 3 Materialized Views
- [x] 3 Caching Layers
- [x] Audit Logger
- [x] Event Store (journal entries)
- [x] Router Registration

### Phase 2: Integration & Testing (IN PROGRESS)

- [x] Bootstrap Wiring (`cmd/api/bootstrap.go`)
- [x] Event Integration (bank, billing, fiscal handlers)
- [x] Swagger Generation (78 endpoints documented)
- [ ] Unit Tests (20-30 per aggregate)
- [ ] Smoke Tests (6-9 per handler)
- [ ] Integration Tests (6-10 per repository)

### Phase 3: Advanced Features (PLANNED)

- [ ] Reporting (Trial Balance, Financial Statements)
- [ ] Period Closing Automation
- [ ] Multi-Currency Support
- [ ] Budget Alerts
- [ ] Cost Center Reporting
- [ ] Advanced Reconciliation

---

## Usage Examples

### Complete Workflow Example

```bash
# 1. Setup Chart of Accounts
POST /api/v1/accounting/accounts
{"code": "311", "name": "Bank Accounts", "account_type": "asset"}

POST /api/v1/accounting/accounts
{"code": "702", "name": "Sales Revenue", "account_type": "revenue"}

# 2. Create Fiscal Period
POST /api/v1/accounting/fiscal-periods
{"name": "January 2026", "start_date": "2026-01-01", "end_date": "2026-01-31", "year": 2026, "quarter": 1}

# 3. Record Sales Transaction
POST /api/v1/accounting/journal-entries
{"fiscal_period_id": "{period-id}", "entry_date": "2026-01-20", "description": "Sales"}

POST /api/v1/accounting/journal-entries/{id}/lines
{"account_id": "{311-id}", "debit_amount_cents": 120000, "credit_amount_cents": 0}

POST /api/v1/accounting/journal-entries/{id}/lines
{"account_id": "{702-id}", "debit_amount_cents": 0, "credit_amount_cents": 120000}

POST /api/v1/accounting/journal-entries/{id}/post

# 4. Create Budget
POST /api/v1/accounting/budgets
{"name": "2026 Sales Budget", "fiscal_year": 2026}

POST /api/v1/accounting/budgets/{id}/lines
{"account_id": "{702-id}", "budget_amount_cents": 10000000}

POST /api/v1/accounting/budgets/{id}/approve
POST /api/v1/accounting/budgets/{id}/activate

# 5. Bank Reconciliation
POST /api/v1/accounting/bank-reconciliations
{
  "bank_account_id": "{bank-id}",
  "account_id": "{311-id}",
  "reconciliation_date": "2026-01-31",
  "statement_date": "2026-01-31",
  "bank_statement_balance_cents": 100000000,
  "book_balance_cents": 98000000
}

POST /api/v1/accounting/bank-reconciliations/{id}/items
{"transaction_type": "outstanding", "description": "Check #1001", "amount_cents": 2000000}

POST /api/v1/accounting/bank-reconciliations/{id}/complete
```

---

## Error Handling

### Security Pattern (Gold Standard)

**Validation errors** → Expose details (user can fix)  
 **Not-found errors** → Generic message (prevent enumeration)  
 **System errors** → Hide details (prevent leakage)

Example:

```go
func handleAccountError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, account.ErrAccountNotFound):
        response.NotFound(c, "account not found")  // 404 generic

    case errors.Is(err, account.ErrCodeRequired):
        response.BadRequest(c, err.Error())  // 400 with details

    default:
        response.InternalError(c, "operation failed")  // 500 generic
    }
}
```

---

## Related Contexts

- **Banking**: Bank transactions → Journal entries
- **Billing**: Invoices/Payments → A/R entries
- **Fiscal**: Cash register → Cash entries
- **Shared**: Currency, reference data
- **Identity**: User tracking, audit

---

## Next Steps

1.  **Complete handlers** (DONE)
2.  **Create router.go** (DONE)
3.  **Bootstrap wiring** (DONE - registered in `cmd/api/bootstrap.go`)
4.  **Event integration** (DONE - bank, billing, fiscal handlers)
5.  **Swagger generation** (DONE - 78 endpoints documented)
6.  ⏳ **Testing** (unit, smoke, integration)(bank, billing, fiscal handlers)
7.  ⏳ **Swagger generation**

---

**Status**: **Phase 1 Complete + Integration Done**  
**API Surface**: 78 REST endpoints fully operational  
**Swagger**: Complete API documentation generated  
**Bootstrap**: Registered in bootstrap.go with event handlers  
**Test Coverage**: ⏳ Next Priority  
**Production Ready**: Testing Phase

---

_Last Updated: 2026-01-20_  
_Total Endpoints: 78 across 7 modules_  
_Lines of Code: ~15,000+_  
_Integration: Bank, Billing, Fiscal event handlers active_
