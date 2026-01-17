# CRM→ERP Roadmap Checklist

Source: [CRM_ERP_ROADMAP_2026.md](CRM_ERP_ROADMAP_2026.md)

## Phase 1 (1–2 months): Banking + Sales Reporting

- [ ] Create `banking` context skeleton (directories, aggregates, repo/usecase/handler)
- [ ] Draft DB schemas for BankAccount, BankStatement, BankTransaction
- [ ] Define event contracts: `banking.statement.imported`, `banking.transaction.created`
- [ ] Implement provider clients (Monobank, Privat24, PUMB)
- [ ] Implement credential storage + config wiring
- [ ] Build normalization + reconciliation rules
- [ ] Implement matching engine (transactions ↔ invoices/orders)
- [ ] Add scheduler jobs for imports + retries
- [ ] Add analytics read model schema for sales report
- [ ] Implement materializer worker (event consumers)
- [ ] Add `/api/v1/analytics/sales-report` endpoint + filters
- [ ] Ensure report queries use read model only

## Phase 2 (3–6 months): Accounting Core + CRM Loyalty

- [ ] Create `accounting` context skeleton
- [ ] Draft DB schemas for ChartOfAccounts, JournalEntry, Ledger, PostingRule
- [ ] Define event contracts: `accounting.entry.posted`
- [ ] Implement posting rules for sale, refund, inventory change, fees
- [ ] Build P&L read model + `/api/v1/accounting/pnl`
- [ ] Implement trial balance + ledger APIs
- [ ] Add CurrencyRates service (NBU API) + storage
- [ ] Add multi-currency conversions at booking time
- [ ] Implement CRM timeline materializer
- [ ] Implement RFM segmentation materialized view

## Phase 3 (6–12 months): Warehouse Advanced + Compliance

- [ ] Implement InventoryCount document (snapshot, variance)
- [ ] Implement adjustment documents (write-off, gains)
- [ ] Build costing engine (FIFO/LIFO/Avg)
- [ ] Post COGS to Accounting on stock moves
- [ ] Add audit log (append-only, before/after)
- [ ] Ensure audit coverage for price/stock/status changes
- [ ] Implement outgoing webhooks (outbox + retry)
- [ ] Add webhook events: order.status.changed, payment.received, inventory.low

## Long-term (12+ months): Manufacturing + HRM

- [ ] Add BOM + routing + production orders (manufacturing)
- [ ] Add HRM time tracking + payroll (align with Ukraine compliance roadmap)

## Immediate Next Steps (this sprint)

- [ ] Draft schemas for `banking` and `accounting` contexts
- [ ] Define event contracts for `banking.statement.imported` and `accounting.entry.posted`
- [ ] Add analytics read model for sales report and wire event consumers
- [ ] Prepare migration namespaces for `banking` and `accounting`
