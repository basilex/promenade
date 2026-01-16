# Promenade Roadmaps

Strategic planning documents for platform development and market strategy.

---

## Available Roadmaps

### Technical Roadmaps

#### [Strategic Roadmap 2026](STRATEGIC_ROADMAP_2026.md)
**Timeline**: Q1-Q2 2026 (6 months)  
**Focus**: Core platform development  
**Status**: Phase 3 in progress (Week 2 underway); Phase 4 planned

**Key Phases**:
- Phase 1-2: Foundation + Domain Errors (Complete)
- Phase 3: LUA Scripting Engine + UI Metadata (in progress)
- Phase 4: Job Scheduler (Complete)
- Phase 5: NATS Gateway (Planned)

---

#### [Phase 3: LUA + UI Metadata Foundation](PHASE3_LUA_UI_FOUNDATION.md)
**Timeline**: 3 weeks (January 8-28, 2026)  
**Focus**: No-code platform capabilities  
**Status**: Week 1 complete; Week 2 in progress (UI metadata foundation complete, scripting storage pending)

**Deliverables**:
- LUA Scripting Engine (business rules without recompilation)
- UI Metadata System (dynamic forms like Oracle Forms)
- Standard Library (Customer, Order, Deal, Query, Date APIs)

---

### Ukrainian Market Strategy

#### [Ukraine Market Strategy 2026](UKRAINE_MARKET_STRATEGY_2026.md)
**Timeline**: 12 months (Q1 2026 - Q1 2027)  
**Focus**: Business strategy for Ukrainian CRM/ERP market  
**Status**: Planning

**Key Insights**:
- **Market Opportunity**: 150,000+ SMB companies seeking alternatives to 1C/Bitrix24
- **Market Size**: $500M+ CRM/ERP market in Ukraine
- **Target**: 5-10% market share (7,500-15,000 companies) in 18 months
- **Projected ARR**: $9.4M-29.9M with 7,500 customers

**Must-Have Features**:
1. Fiscal integration (Checkbox, Vchasno.Kasa)
2. Tax invoices (XML for the State Tax Service)
3. Bank statements (Monobank, Privat24, PUMB)
4. HRM (payroll + Ukrainian taxes)
5. Nova Poshta / Ukrposhta APIs

**Competitive Advantages**:
- Modern architecture (DDD + Event-Driven) vs legacy 1C
- Go performance (10-50x faster)
- API-first approach
- Pricing 40-60% lower than Terrasoft
- Open-source core

**Content**:
- Executive Summary (revenue projections, funding)
- Market Analysis (competitors, target segments)
- Technical Requirements (fiscal, accounting, banking)
- Go-to-Market Strategy (beta → launch → growth)
- Pricing Strategy ($29-149/month tiers)
- Partnership Strategy (accounting firms, IT integrators)
- 12-Month Timeline (Q1 2026 - Q1 2027)

---

#### [Ukraine Compliance Roadmap](UKRAINE_COMPLIANCE_ROADMAP.md)
**Timeline**: Q1-Q2 2026 (February - June)  
**Focus**: Technical implementation of Ukrainian compliance  
**Status**: Planning, ready to start

**Phases**:

**Phase 5: Fiscal Integration** (3-4 weeks, February 2026)
- Week 7-8: Core fiscal infrastructure
- Week 9-10: Checkbox + Vchasno.Kasa APIs
- Deliverables: Fiscal receipts, Z-reports, X-reports

**Phase 6: Tax Invoices** (2-3 weeks, March 2026)
- Week 11-12: XML generation for the State Tax Service
- Week 13: M.E.Doc integration
- Deliverables: Registration in the Unified Register, VAT declaration

**Phase 7: Bank Statements** (1-2 weeks, March 2026)
- Week 13-14: Monobank, Privat24, PUMB APIs
- Deliverables: Auto-matching with invoices

**Phase 8: HRM** (3-4 weeks, April 2026)
- Week 14-17: Payroll + taxes (USC, PIT, military levy)
- Deliverables: Form 1DF, USC report

**Phase 9: Delivery** (1 week, April 2026)
- Week 17: Nova Poshta API
- Deliverables: Waybill creation, tracking

**Technical Details**:
- Database schemas (PRRO, accounting, banking)
- API implementations (Checkbox, M.E.Doc, Monobank)
- Event flows (Order → FiscalReceipt)
- Testing strategies (30+ unit, 15+ integration per module)

---

## Priority Matrix

| Priority | Feature | Timeline | Impact | Market |
|----------|---------|----------|--------|--------|
| Blocker | Fiscal Integration | Week 7-10 | 70% | Retail |
| Blocker | Tax Invoices | Week 11-12 | 50% | Accounting firms |
| High | Bank Statements | Week 13-14 | 60% | Automation |
| Medium | HRM | Week 14-17 | 40% | Competitive advantage |
| Medium | Nova Poshta | Week 17 | 50% | E-commerce |

---

## Combined Timeline

```
January 2026:

Week 1-2  : Phase 2 Complete (Domain Errors)
Week 3    : Phase 3 Week 2 (Script Storage)
Week 4    : Phase 3 Week 3 (UI Metadata)

February 2026:

Week 5-6  : Phase 4 (Scheduler)
Week 7-10 : PRRO Integration (Checkbox + Vchasno.Kasa)

March 2026:

Week 11-12: Tax Invoices + XML
Week 13-14: Bank Statements (Monobank, Privat24)

April 2026:

Week 14-17: HRM (Payroll + Taxes)
Week 17   : Nova Poshta API

May - June 2026:

Week 18-24: NATS Gateway + Mobile App
Week 25-26: Polish + Documentation

Q3 2026 (Launch):

October: PUBLIC BETA LAUNCH
November: Bug fixes + User Feedback
December: Premium Features

Q4 2026 Goals:

- 500 active users
- 150 paying customers
- $15K MRR
```

---

## Next Steps

### Immediate (Week 3-4, January 2026)
1. Complete Phase 3 scripting storage and UI metadata integration
2. Prepare technical specs for PRRO
3. Find 10 beta testers for PRRO

### Short-term (February 2026)
1. Complete Phase 4 (Scheduler)
2. Start PRRO Integration
3. Register a legal entity in Ukraine

### Medium-term (March - April 2026)
1. Complete all Ukrainian compliance modules
2. 100 beta testers
3. First 10 paying customers

### Long-term (Q3-Q4 2026)
1. PUBLIC BETA LAUNCH
2. Mobile App (Flutter)
3. 500 active users, $15K MRR

---

## Success Metrics

### Technical KPIs
- **Availability**: 99.5% uptime
- **Performance**: < 200ms API response
- **Test Coverage**: 90%+
- **Ukrainian Compliance**: 100% regulatory compliance

### Business KPIs (Ukrainian Market)

**Q1 2026**:
- 20 beta users (PRRO testing)
- 0 paying customers (beta period)
- $0 MRR

**Q2 2026**:
- 100 Beta users
- 10 paying customers
- $1K MRR

**Q3 2026**:
- 200 active users
- 50 paying customers
- $5K MRR

**Q4 2026**:
- 500 active users
- 150 paying customers
- $15K MRR

**Q1 2027** (Target):
- 1,000 active users
- 300 paying customers
- $30K MRR

---

## Related Documentation

**Core Platform**:
- [Main README](../README.md) - Project overview
- [Documentation Index](INDEX.md) - Complete docs
- [Architecture Guide](../concepts/clean-architecture.md) - DDD architecture

**Development Guides**:
- [Quick Start Guide](../guides/quick-start.md) - 5-minute tutorial
- [Testing Patterns](../guides/testing-patterns.md) - Testing strategy
- [API Documentation](../guides/api-documentation.md) - Swagger guide

**Business**:
- [Business Overview](../business/BUSINESS_OVERVIEW.md) - For decision-makers
- [Business Overview (Ukrainian)](../business/BUSINESS_OVERVIEW_UK.md) - For decision-makers

---

**Last Updated**: January 14, 2026  
**Status**: Active Planning & Development  
**Maintainer**: Promenade Team
