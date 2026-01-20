# Accounting Migrations

Database migrations for the **Accounting Context** - professional double-entry bookkeeping system implementing Ukrainian P(S)BO standards.

---

## Overview

Professional accounting system with 21+ tables covering:

- **Double-entry bookkeeping** - Journal entries with debit/credit validation
- **Fiscal period management** - Month/quarter/year close and lock
- **Bank reconciliation** - Statement matching with outstanding items
- **Tax compliance** - Ukrainian VAT, corporate income tax, personal income tax, social contributions
- **Budgeting** - Budget vs actual with variance analysis
- **Financial reporting** - Balance Sheet, P&L, Cash Flow, Trial Balance
- **Audit trail** - Complete change tracking for compliance
- **Management accounting** - Cost/profit centers with allocations

All tables follow the `accounting_*` naming prefix convention.

---

## Migration Files

### 000001_accounting_core.up.sql (CORE)

Creates 4 foundational tables:

- **`accounting_chart_of_accounts`** - Hierarchical chart of accounts (Plan of Accounts)
  - Ukrainian P(S)BO classes 1-9 (Asset, Liability, Equity, Income, Expense)
  - Parent-child relationships for subaccounts
  - Multi-currency support with default currency
  - Active/inactive status, soft delete

- **`accounting_journal_entries`** - Journal entry headers
  - Entry date, description, status workflow (`draft` → `posted` → `reversed`)
  - Source tracking (manual, bank_transaction, invoice, payment, receipt, adjustment)
  - Audit trail (posted_by, posted_at, reversed_by, reversed_at)
  - Optimistic locking (version field)

- **`accounting_journal_entry_lines`** - Debit/credit lines
  - Direction: `debit` or `credit`
  - Amount in cents (integer for precision)
  - Multi-currency with exchange rate
  - Links to journal entry and account

- **`accounting_posting_rules`** - Automated posting rules
  - Maps domain events to accounting entries
  - Configurable debit/credit accounts, amount formulas
  - Rule activation control

**Indexes**: Organization-based partitioning, fast lookups by code/status/date, source entity tracking.

**Down Migration**: `000001_accounting_core.down.sql`

---

### 000002_seed_chart_of_accounts.up.sql (SEED DATA)

Seeds 36 Ukrainian P(S)BO accounts across all classes:

- **Class 1-3: Assets** (10 accounts) - Fixed assets, Inventory, Cash/Bank, Receivables
- **Class 4: Equity** (4 accounts) - Registered capital, Retained earnings
- **Class 5-6: Liabilities** (8 accounts) - Loans, Accounts payable, Tax payable, Payroll
- **Class 7: Income** (6 accounts) - Sales revenue, Other income
- **Class 8-9: Expenses** (8 accounts) - COGS, Administrative, Sales, Financial expenses

Key accounts: 311 (Bank), 301 (Cash), 361 (Receivables), 631 (Payables), 702 (Sales), 902 (Expenses)

**Down Migration**: `000002_seed_chart_of_accounts.down.sql`

---

### 000003_fiscal_periods_and_ledger.up.sql (PERIOD MANAGEMENT)

Creates 4 tables for fiscal period management and ledger balances:

- **`accounting_fiscal_periods`** - Fiscal periods (month/quarter/year)
  - Period status: `open`, `closed`, `locked`
  - Lock date enforcement - prevents retroactive modifications
  - Period close/open tracking (closed_by, closed_at, locked_by, locked_at)

- **`accounting_ledger`** - Running balances per account per period
  - Opening balance, debits, credits, closing balance
  - Denormalized for fast financial statement generation
  - Updated automatically on journal entry posting

- **`accounting_account_groups`** - Logical grouping for reports
  - Report types: `balance_sheet`, `income_statement`, `cash_flow`
  - Hierarchical structure with parent_id
  - Section classification (Assets, Liabilities, Revenue, Expenses)

- **`accounting_account_group_members`** - Many-to-many account-to-group mapping

**Down Migration**: `000003_fiscal_periods_and_ledger.down.sql`

---

### 000004_adjustments_reconciliation_tax.up.sql (ADVANCED FEATURES)

Creates 7 tables for adjustments, reconciliation, tax, and budgeting:

- **`accounting_journal_entry_links`** - Track entry relationships
  - Link types: `reversal`, `adjustment`, `correction`, `reclassification`
  - Bidirectional linking between original and new entries

- **`accounting_bank_reconciliations`** + `items` - Bank statement matching
  - Reconciliation workflow: `in_progress`, `completed`, `reviewed`
  - Outstanding items tracking (uncleared checks, deposits in transit)
  - Statement balance vs book balance reconciliation

- **`accounting_tax_codes`** - Ukrainian tax definitions
  - Tax types: `vat`, `income_tax`, `payroll_tax`, `other`
  - Rate storage with precision (20.0000%)
  - GL account mapping for tax payable/receivable

- **`accounting_journal_entry_line_taxes`** - Line-level tax tracking
  - Links tax codes to journal lines
  - Taxable amount and tax amount separation

- **`accounting_budgets`** + `lines` - Budget management
  - Budget master with approval workflow (`draft` → `approved` → `active`)
  - Monthly/period allocations with actual vs budget variance
  - Budget categories for department/project tracking

**Down Migration**: `000004_adjustments_reconciliation_tax.down.sql`

---

### 000005_audit_reports_advanced.up.sql (AUDIT & REPORTING)

Creates 6 tables for audit trail, reporting, and management accounting:

- **`accounting_audit_log`** - Complete change tracking
  - Entity type + ID for any accounting record
  - Before/after JSON snapshots (old_values, new_values)
  - User, IP address, action, reason tracking
  - Immutable log for compliance

- **`accounting_report_templates`** - Configurable financial reports
  - Report types: `balance_sheet`, `income_statement`, `cash_flow`, `trial_balance`
  - JSON structure for report layout and calculations
  - Default templates with customization support

- **`accounting_report_snapshots`** - Saved reports with approval
  - Report data in JSON format
  - Approval workflow: `draft` → `final` → `approved` → `published`
  - Summary metrics (total assets, liabilities, equity, revenue, expenses, net income)

- **`accounting_account_restrictions`** - Internal controls
  - Restriction types: `no_manual_entry`, `approval_required`, `date_range_limit`
  - Approval threshold amounts
  - Enforce segregation of duties

- **`accounting_cost_centers`** - Management accounting
  - Center types: `cost_center`, `profit_center`, `investment_center`
  - Hierarchical structure with manager assignment
  - Cost/profit center hierarchy

- **`accounting_journal_entry_line_allocations`** - Cost center allocations
  - Allocate journal lines to multiple cost centers
  - Percentage or amount-based allocation

**Down Migration**: `000005_audit_reports_advanced.down.sql`

---

### 000006_seed_periods_tax_groups.up.sql (REFERENCE DATA)

Seeds reference data for fiscal periods, tax codes, account groups, and cost centers:

- **17 Fiscal Periods for 2026**: 12 months + 4 quarters + 1 year (all status `open`)
- **9 Ukrainian Tax Codes**:
  - ПДВ: 20%, 7%, 0% (export), exempt
  - Податок на прибуток: 18%
  - Payroll taxes: ПДФО 18%, Військовий збір 1.5%, ЄСВ 22%
- **17 Account Groups**: Balance Sheet (Assets/Liabilities/Equity) + Income Statement (Revenue/Expenses)
- **4 Cost Centers**: Administration, Sales, Operations, IT

**Down Migration**: `000006_seed_periods_tax_groups.down.sql`

---

## Schema Summary

**Total: 21 tables**

### Core (4 tables)

- `accounting_chart_of_accounts` - Hierarchical chart
- `accounting_journal_entries` - Entry headers
- `accounting_journal_entry_lines` - Debit/credit lines
- `accounting_posting_rules` - Event-to-accounting mappings

### Period Management (4 tables)

- `accounting_fiscal_periods` - Month/quarter/year periods
- `accounting_ledger` - Running balances
- `accounting_account_groups` - Report grouping
- `accounting_account_group_members` - Account-to-group mapping

### Advanced Features (7 tables)

- `accounting_journal_entry_links` - Entry relationships
- `accounting_bank_reconciliations` + `items` - Bank statement matching
- `accounting_tax_codes` - Tax definitions
- `accounting_journal_entry_line_taxes` - Line-level taxes
- `accounting_budgets` + `lines` - Budget vs actual

### Audit & Reporting (6 tables)

- `accounting_audit_log` - Change tracking
- `accounting_report_templates` - Report definitions
- `accounting_report_snapshots` - Saved reports
- `accounting_account_restrictions` - Control rules
- `accounting_cost_centers` - Management accounting
- `accounting_journal_entry_line_allocations` - Cost allocations

---

## Key Features

 **Double-Entry Bookkeeping** - Balanced debit/credit validation  
 **Fiscal Period Close/Lock** - Prevents retroactive modifications  
 **Bank Reconciliation** - Statement matching with outstanding items  
 **Tax Compliance** - Ukrainian ПДВ, податок на прибуток, ПДФО, ЄСВ tracking  
 **Budgeting** - Budget vs actual with variance analysis  
 **Financial Reporting** - Balance Sheet, P&L, Cash Flow, Trial Balance  
 **Audit Trail** - Complete change log for compliance  
 **Management Accounting** - Cost/profit center tracking  
 **Internal Controls** - Account restrictions, approval workflows

---

## Running Migrations

### All Accounting Migrations

```bash
make migrate-module MODULE=accounting
```

### Rollback Specific Steps

```bash
make migrate-rollback MODULE=accounting STEPS=1  # Rollback 000006
make migrate-rollback MODULE=accounting STEPS=3  # Rollback 000006, 000005, 000004
```

### Migration Status

```bash
./cmd/migrate/migrate status accounting
```

---

## Integration with Bounded Contexts

**Event Sources (Incoming)**:

- `banking.transaction.created` → Auto-post bank entries
- `billing.invoice.created` → Record AR and revenue
- `billing.invoice.paid` → Record payment application
- `order-mgmt.order.fulfilled` → Record COGS and inventory reduction
- `warehouse.inventory.adjusted` → Record inventory write-offs

**Event Publications (Outgoing)**:

- `accounting.journal_entry.posted` → Notify other contexts
- `accounting.period.closed` → Trigger end-of-period processes
- `accounting.reconciliation.completed` → Update banking context

See [internal/contexts/accounting/integration](../../internal/contexts/accounting/integration/) for event handlers.

---

## Database Compatibility

All migrations are **multi-DB compatible**:

- **PostgreSQL** (production) - uses `TIMESTAMP`, `BOOLEAN`, `TEXT`
- **SQLite** (dev/test) - uses `TEXT` for dates, `INTEGER` for boolean

**IDs**: Always `TEXT` (UUID v7 strings), never auto-increment.

---

## Ukrainian Accounting Standards (П(С)БО)

Chart of accounts follows **Plan of Accounts for Ukrainian enterprises**:

- **Class 0**: Off-balance accounts (not implemented)
- **Class 1**: Non-current assets (Необоротні активи)
- **Class 2**: Inventories (Запаси)
- **Class 3**: Cash, receivables (Грошові кошти, розрахунки)
- **Class 4**: Equity (Власний капітал)
- **Class 5**: Long-term liabilities (Довгострокові зобов'язання)
- **Class 6**: Current liabilities (Поточні зобов'язання)
- **Class 7**: Income (Доходи)
- **Class 8**: Expenses (Витрати)
- **Class 9**: Cost of sales (Витрати діяльності)

---

## Compliance Features

-  **Period Close** - Monthly/quarterly/yearly close with validation
-  **Period Lock** - Prevent retroactive changes to closed periods
-  **Audit Trail** - Complete before/after change log
-  **Bank Reconciliation** - Match bank statements with book entries
-  **Tax Tracking** - ПДВ, податок на прибуток, ПДФО, ЄСВ
-  **Approval Workflows** - Budget approval, report approval, entry approval
-  **Internal Controls** - Account restrictions, segregation of duties
-  **Financial Statements** - Balance Sheet (form 1), P&L (form 2)

---

## Future Enhancements

- [ ] Analytical dimensions (projects, departments, products)
- [ ] Multi-level approval workflows
- [ ] Automated accruals and deferrals
- [ ] Fixed assets depreciation schedules
- [ ] Intercompany eliminations
- [ ] Consolidation for group reporting
- [ ] IFRS parallel accounting

---

## References

- [П(С)БО (Ukrainian Accounting Standards)](https://zakon.rada.gov.ua/laws/show/z0391-99)
- [Chart of Accounts Guide](https://buhgalter911.com/news/news-1034361.html)
- [Double-Entry Bookkeeping](https://en.wikipedia.org/wiki/Double-entry_bookkeeping)
