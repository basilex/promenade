# Fiscal Context (ПРРО - Програмний РРО)

**Ukrainian Fiscal Compliance** - Cash register integration for retail businesses in Ukraine.

**Status**: 🚧 **MVP IN PROGRESS** (Week 1 of 2)  
**Priority**: 🔥 **CRITICAL BLOCKER** - 70% of target market  
**Timeline**: January 14 - January 28, 2026

---

## Overview

Fiscal Context provides integration with Ukrainian ПРРО (Програмний РРО) systems for retail businesses. This enables legal compliance with Ukrainian tax regulations for cash transactions.

### Key Features

- ✅ Cash Register Management (ПРРО)
- ✅ Fiscal Receipt Printing
- ✅ Checkbox API Integration (MVP)
- 🚧 Receipt Status Tracking (Week 2)
- 🚧 Shift Management (Week 2)
- 🚧 Z-Reports (Week 2)
- ⏳ Вчасно.Каса Integration (Future)

---

## Architecture

### Aggregates

1. **CashRegister** (`cashregister/`)
   - ПРРО registration and management
   - API credentials (encrypted)
   - Location-based assignment
   - Active/Inactive status

2. **Receipt** (`receipt/`) - 🚧 IN PROGRESS
   - Fiscal receipt creation
   - Order integration
   - Payment type tracking
   - Status: pending → printed → cancelled

3. **Shift** (`shift/`) - ⏳ PLANNED
   - Shift management (open/close)
   - Z-Reports generation
   - Daily summaries

### External Integrations

- **Checkbox API** (`pkg/fiscal/checkbox/`)
  - Production: `https://api.checkbox.ua/api/v1`
  - Sandbox: `https://api.sandbox.checkbox.ua/api/v1`
  - Documentation: https://dev.checkbox.ua/uk/docs/api/

---

## Current Status (Week 1)

### ✅ Completed

1. **Project Structure**
   - ✅ Directory structure created
   - ✅ CashRegister aggregate (entity, errors, repository, usecase)
   - ✅ Database migration (fiscal_cash_registers, fiscal_receipts)
   - ✅ Checkbox API client (basic implementation)

### 🚧 In Progress (Week 1, Day 2-7)

2. **Receipt Aggregate**
   - ⏳ Entity implementation
   - ⏳ Repository (PostgreSQL)
   - ⏳ UseCase (business logic)

3. **HTTP API**
   - ⏳ Cash Register handlers
   - ⏳ Receipt handlers
   - ⏳ Router setup

4. **Integration with Order Context**
   - ⏳ Order → Receipt event flow
   - ⏳ Auto-print on order confirmation

### ⏳ Planned (Week 2)

5. **Testing**
   - ⏳ Unit tests (30+ tests target)
   - ⏳ Integration tests (15+ tests)
   - ⏳ Smoke tests (10+ tests)

6. **Beta Testing**
   - ⏳ Find 3-5 retail stores
   - ⏳ Real-world validation
   - ⏳ Feedback collection

---

## Usage Examples

### Register Cash Register (ПРРО)

```bash
POST /api/v1/fiscal/cash-registers
Content-Type: application/json

{
  "location_id": "01HX...",
  "name": "Каса №1",
  "provider": "checkbox",
  "api_key": "your_checkbox_api_key",
  "fiscal_number": "1234567890",
  "registration_number": "РК123456"
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

### What's INCLUDED (MVP)

- ✅ Checkbox API integration ONLY
- ✅ Cash register CRUD
- ✅ Receipt creation from orders
- ✅ Basic error handling
- ✅ Fiscal number tracking

### What's EXCLUDED (Future)

- ❌ Вчасно.Каса integration (add later)
- ❌ Shift management UI (manual for now)
- ❌ Z-Reports automation (manual for now)
- ❌ Receipt corrections (rare use case)
- ❌ Multiple tax rates per receipt (edge case)

---

## Technical Decisions (MVP)

### Why Checkbox ONLY?

1. **Market Share**: 60%+ of ПРРО market
2. **API Quality**: Well-documented, stable
3. **Fast MVP**: 1 provider = 2 weeks vs 3 providers = 5-6 weeks
4. **Validation**: Get feedback before adding more providers

### Why No Shift UI?

1. **Retail Reality**: Cashiers open/close shifts manually in terminal
2. **MVP Speed**: UI adds 1-2 weeks
3. **Backend First**: Focus on API correctness

### Why Basic Error Handling?

1. **Network Reality**: Retry logic is complex
2. **MVP Focus**: Happy path first, edge cases later
3. **Real Feedback**: Beta testers will find real errors

---

## Next Steps (Week 1, Days 2-7)

### Day 2 (Today) ✅

- [x] Create Fiscal Context structure
- [x] Implement CashRegister aggregate
- [x] Create database migration
- [x] Implement Checkbox API client
- [x] Write README

### Days 3-4 (January 15-16)

- [ ] Implement Receipt aggregate
- [ ] Create PostgreSQL repository
- [ ] Implement business logic (UseCase)

### Days 5-6 (January 17-18)

- [ ] HTTP handlers (DTOs, validation)
- [ ] Router setup
- [ ] Bootstrap integration

### Day 7 (January 19)

- [ ] Unit tests (30+ tests)
- [ ] Manual testing (Postman)
- [ ] Code review

---

## Testing Strategy

### Unit Tests (30+ target)

- CashRegister entity (10 tests)
- Receipt entity (10 tests)
- UseCase business logic (10+ tests)

### Integration Tests (15+ target)

- PostgreSQL repository (10 tests)
- Checkbox API mock (5 tests)

### Smoke Tests (10+ target)

- HTTP handlers (10 tests)
- Status codes validation
- Response format checks

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

**Status**: Day 1 Complete ✅  
**Next**: Receipt aggregate implementation (Day 2)  
**Timeline**: On track for 2-week MVP  
**Last Updated**: January 14, 2026

