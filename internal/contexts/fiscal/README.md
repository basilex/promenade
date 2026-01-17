# Fiscal Context (PRRO Compliance)

**Ukrainian fiscal compliance** for retail businesses via cash register integration.

**Status**: MVP core delivered; sandbox validation in progress  \
**Priority**: Critical blocker - 70% of target market  \
**Timeline**: January 2026

---

## Overview

The Fiscal context integrates with Ukrainian PRRO systems to issue fiscal receipts and comply with tax regulations for cash transactions.

### Key Features

- Cash register management
- Fiscal receipt printing
- Checkbox API integration (MVP)
- Receipt status tracking
- Shift open/close automation (scheduler)
- Z-report references on shift close
- Vchasno.Kasa integration (future)

---

## Architecture

### Aggregates

1. **CashRegister** ([cashregister/README.md](cashregister/README.md))
   - PRRO registration and management
   - API credentials (encrypted)
   - Location-based assignment
   - Active/inactive status

2. **Receipt** ([receipt/README.md](receipt/README.md))
   - Fiscal receipt creation
   - Order integration
   - Payment type tracking
   - Status: pending → printed → cancelled

3. **Shift Automation** (via CashRegister)
   - Shift open/close via Checkbox API
   - Z-report references on close
   - Daily cadence via scheduler jobs

### External Integrations

- **Checkbox API** (`pkg/fiscal/checkbox/`)
  - Production: `https://api.checkbox.ua/api/v1`
  - Sandbox: `https://api.sandbox.checkbox.ua/api/v1`
  - Documentation: https://dev.checkbox.ua/uk/docs/api/

---

## Current Status (January 2026)

### Completed

1. **Core Aggregates + Repositories**
   - CashRegister + Receipt aggregates
   - PostgreSQL repositories + migrations
   - Provider receipt ID persistence

2. **Checkbox Integration**
   - Create/print/cancel receipt
   - Shift open/close + Z-report response

3. **Order Event Flow**
   - order.confirmed → receipt creation
   - Optional auto-print on confirmation

4. **Scheduler Integration**
   - Retry printing job
   - Shift open/close jobs

5. **Testing**
   - Unit tests for receipts, printers, checkbox client
   - Integration tests for order events + shift automation

### In Progress

6. **Sandbox Validation**
   - End-to-end checklist for receipt/shift/Z-report edge cases
   - Smoke test coverage for fiscal handlers

### Planned

7. **Daily Z-Report Cadence + Monitoring**
   - Operational runbook + alerting

8. **Beta Testing**
   - Identify 3-5 retail stores
   - Real-world validation
   - Feedback collection

---

## Usage Examples

### Register Cash Register

```bash
POST /api/v1/fiscal/cash-registers
Content-Type: application/json

{
  "location_id": "01HX...",
  "name": "Register #1",
  "provider": "checkbox",
  "api_key": "your_checkbox_api_key",
  "fiscal_number": "1234567890",
  "registration_number": "RK123456"
}
```

### Create Fiscal Receipt

```bash
POST /api/v1/fiscal/receipts
Content-Type: application/json

{
  "order_id": "01HY...",
  "cash_register_id": "01HZ...",
  "payment_type": "cash",
  "lines": [
    {
      "name": "Product 1",
      "quantity": 2,
      "price": 10000,
      "tax_rate": 20
    }
  ]
}
```

**Response:**
```json
{
  "id": "01HZ...",
  "fiscal_number": "1234567890",
  "fiscal_url": "https://tax.gov.ua/receipt/...",
  "qr_code": "data:image/png;base64,...",
  "status": "printed"
}
```

---

## MVP Scope (2 Weeks)

### Included

- Checkbox API integration only
- Cash register CRUD
- Receipt creation from orders
- Receipt printing + cancellation
- Shift open/close automation with Z-report references
- Basic error handling
- Fiscal number tracking

### Excluded (Future)

- Vchasno.Kasa integration
- Shift management UI
- Receipt corrections
- Multiple tax rates per receipt

---

## Technical Decisions (MVP)

### Why Checkbox Only?

1. **Market Share**: 60%+ of the PRRO market
2. **API Quality**: Well-documented, stable
3. **Fast MVP**: 1 provider = 2 weeks vs 3 providers = 5-6 weeks
4. **Validation**: Get feedback before adding more providers

### Why No Shift UI?

1. **Retail Reality**: Cashiers open/close shifts manually in the terminal
2. **MVP Speed**: UI adds 1-2 weeks
3. **Backend First**: Focus on API correctness

### Why Basic Error Handling?

1. **Network Reality**: Retry logic is complex
2. **MVP Focus**: Happy path first, edge cases later
3. **Real Feedback**: Beta testers will find real errors

---

## Next Steps

- Add fiscal smoke tests (failure + compensation cases)
- Finalize daily Z-report cadence + monitoring
- Complete Checkbox sandbox validation checklist
- Prepare beta onboarding checklist and sample data

---

## Testing Strategy

Fiscal is **high risk** (money + compliance), so allow a larger test budget while keeping the same pattern.

### Baseline Budget (per aggregate)

- Unit tests: 20–30 (invariants, state transitions, validation)
- Smoke tests: 6–9 (CRUD + key error mappings)
- Integration tests: 6–10 (happy path + not found + constraint)

### Fiscal Budget (high risk)

- Unit tests: baseline +30–50%
- Smoke tests: baseline +30–50%
- Integration tests: baseline +30–50%

### Coverage Target

- 75–80% by default
- 85–90% for high‑risk operations (printing, fiscal numbers, cancellation)

---

## Beta Testing Plan (Week 2)

### Target: 3-5 Retail Stores

**Ideal Beta Testers:**
1. Small grocery store (5-10 receipts/day)
2. Cafe/Restaurant (20-50 receipts/day)
3. Service business (1-5 receipts/day)

**What to Validate:**
- Receipt printing works reliably
- Fiscal numbers are correct
- QR codes work (tax.gov.ua)
- Error messages are clear
- Performance is acceptable

**Success Criteria:**
- 90%+ receipts printed successfully
- <3s average response time
- Clear error messages for failures
- Positive feedback from 2/3 testers

---

## Documentation

- API Documentation: (TODO - Week 2)
- Checkbox Integration Guide: (TODO - Week 2)
- Troubleshooting Guide: (TODO - Week 2)

---

## Related Documentation

- [Ukraine Market Strategy 2026](../../docs/roadmap/UKRAINE_MARKET_STRATEGY_2026.md)
- [Ukraine Compliance Roadmap](../../docs/roadmap/UKRAINE_COMPLIANCE_ROADMAP.md)
- Checkbox API Docs: https://dev.checkbox.ua/uk/docs/api/

---

**Status**: MVP core delivered; validation in progress  \
**Next**: Sandbox checklist + smoke test coverage  \
**Timeline**: January 2026  \
**Last Updated**: January 17, 2026

