# Promenade CRM→ERP Technical Roadmap 2026

**Date**: January 17, 2026 \
**Status**: Draft (based on expert review) \
**Owner**: Platform Team \
**Horizon**: 12 months

---

## Goal

Move from strong CRM + fiscal + warehouse to ERP-grade accounting, analytics, and integration while preserving DDD boundaries and event-driven data flows. The primary business outcome is **accurate Profit & Loss** derived from fiscal + warehouse data.

## Principles (must follow)

- **Bounded contexts only** in [internal/contexts](../internal/contexts); no cross-context imports. Use the Event Bus for integration.
- **CQRS for analytics**: heavy reporting goes to read models, not transactional tables.
- **DB-agnostic migrations** in [migrations](../migrations) using TEXT for JSON/UUID and IDs in Go via `uuidv7.New()`.
- **Domain errors** in `errors.go` per aggregate; handlers map with `errors.Is()` (see [docs/guides/security-patterns.md](../guides/security-patterns.md)).

---

## Phase 1 (1–2 months): Banking + Sales Reporting

**Objective**: Immediate automation and visible business value.

### A. Banking Imports (Monobank, Privat24, PUMB)

**New context**: `internal/contexts/banking/` (BankAccount, BankStatement, BankTransaction) \
**Integration**: scheduled pulls + manual sync endpoints \
**Events**: `banking.statement.imported`, `banking.transaction.created`

**Deliverables**:

- Provider clients + credential storage (encrypted secrets via config)
- Normalized statement model + reconciliation rules
- Matching engine: transactions ↔ invoices/orders
- Scheduler jobs for daily import and retries

### B. Sales Report (Analytics Read Model)

**New read model**: analytics tables keyed by Order + Deal + Manager \
**Event inputs**: order.confirmed, order.paid, fiscal.receipt.printed

**Deliverables**:

- Read model schema + materializer worker
- `/api/v1/analytics/sales-report` with filters (category, manager, date range)

**Acceptance**:

- Report query does not hit transactional tables
- Import jobs resilient with idempotent event handling

---

## Phase 2 (3–6 months): Accounting Core + CRM Loyalty

**Objective**: Double-entry accounting + customer intelligence.

### A. Accounting (Double Entry) + P&L

**New context**: `internal/contexts/accounting/`

- Aggregates: ChartOfAccounts, JournalEntry, Ledger, PostingRule
- Value objects: Money, Currency, ExchangeRate
- Events in: order.paid, fiscal.receipt.printed, warehouse.stock.moved, banking.transaction.created

**Deliverables**:

- Posting rules for common flows (sale, refund, inventory change, fees)
- P&L read model + `/api/v1/accounting/pnl`
- Trial balance and ledger APIs for accountants

### B. Multi-currency + NBU Rates

**New shared service**: CurrencyRates (NBU API) \
**Usage**: store rates in read model; convert at booking time

### C. CRM Timeline + RFM Analytics

**New analytics worker**: build per-customer timeline (orders, interactions, payments) \
**RFM**: Recency/Frequency/Monetary segmentation as materialized view

**Acceptance**:

- P&L equals sum of journal entries (period-based)
- RFM segments updated via events, not manual cron

---

## Phase 3 (6–12 months): Warehouse Advanced + Compliance

**Objective**: ERP-grade inventory costing + auditability + integrations.

### A. Inventory Count + Adjustments

**Warehouse extensions**:

- InventoryCount document (snapshot, variance)
- Adjustment documents for write-offs and gains

### B. Cost Accounting (FIFO/LIFO/Avg)

**Warehouse costing engine**:

- Cost layers per SKU
- COGS posting into Accounting on stock moves

### C. Audit Log (Compliance)

**New context or shared module**:

- Append-only audit log (who/when/before/after)
- Required for price/stock/status changes

### D. Webhooks (Outgoing)

**Integration**:

- Outbox pattern + retry policy
- Events: order.status.changed, payment.received, inventory.low

**Acceptance**:

- Costing matches COGS in P&L
- Audit log covers price + stock changes
- Webhook delivery is durable and retryable

---

## Long-term (12+ months): Manufacturing + HRM

**Manufacturing**: BOM, routing, production orders \
**HRM**: time tracking + payroll (align with [docs/roadmap/UKRAINE_COMPLIANCE_ROADMAP.md](UKRAINE_COMPLIANCE_ROADMAP.md))

---

## Implementation Notes (by area)

- **Events**: add topics in [pkg/bus/topics.go](../pkg/bus/topics.go). Handlers must log publish errors without failing the operation.
- **Repositories**: embed BaseRepository and use `getExecutor(ctx)` for transaction propagation.
- **JSON fields**: use `jsonstore.Field[T]` from [pkg/jsonstore](../pkg/jsonstore/README.md).
- **Tests**: follow patterns in [test/README.md](../test/README.md) (unit + smoke + integration).

---

## Immediate Next Steps (this sprint)

1. Draft schemas for `banking` and `accounting` contexts.
2. Define event contracts for `banking.statement.imported` and `accounting.entry.posted`.
3. Add analytics read model for sales report and wire event consumers.
4. Prepare migration namespaces for `banking` and `accounting`.

**Maintainer**: Promenade Team
