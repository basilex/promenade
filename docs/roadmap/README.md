# Promenade Roadmaps

Strategic planning documents for platform development and market strategy.

---

## 📋 Available Roadmaps

### Technical Roadmaps

#### [Strategic Roadmap 2026](STRATEGIC_ROADMAP_2026.md)
**Timeline**: Q1-Q2 2026 (6 місяців)  
**Focus**: Core platform development  
**Status**: Phase 3-4 in progress

**Key Phases**:
- ✅ Phase 1-2: Foundation + Domain Errors (Complete)
- 🔄 Phase 3: LUA Scripting Engine (70% done)
- ⏱️ Phase 4: Job Scheduler (Planned)
- ⏱️ Phase 5: NATS Gateway (Planned)

---

#### [Phase 3: LUA + UI Metadata Foundation](PHASE3_LUA_UI_FOUNDATION.md)
**Timeline**: 3 weeks (January 8-28, 2026)  
**Focus**: No-code platform capabilities  
**Status**: Week 1 complete, Week 2-3 in progress

**Deliverables**:
- LUA Scripting Engine (business rules without recompilation)
- UI Metadata System (dynamic forms like Oracle Forms)
- Standard Library (Customer, Order, Deal, Query, Date APIs)

---

### Ukrainian Market Strategy 🇺🇦

#### [Ukraine Market Strategy 2026](UKRAINE_MARKET_STRATEGY_2026.md) ⭐
**Timeline**: 12 місяців (Q1 2026 - Q1 2027)  
**Focus**: Business strategy for Ukrainian CRM/ERP market  
**Status**: Planning

**Key Insights**:
- **Market Opportunity**: 150,000+ компаній МСБ шукають заміну 1С/Bitrix24
- **Market Size**: $500M+ ринок CRM/ERP в Україні
- **Target**: 5-10% ринку (7,500-15,000 компаній) за 18 місяців
- **Projected ARR**: $9.4M-29.9M при 7,500 клієнтів

**Must-Have Features**:
1. 🔥 ПРРО Integration (Checkbox, Вчасно.Каса)
2. 🔥 Податкові Накладні (XML для ДПС)
3. 🔥 Банківські Виписки (Monobank, Privat24, PUMB)
4. 💼 HRM (зарплата + українські податки)
5. 🚚 Нова Пошта / Укрпошта APIs

**Competitive Advantages**:
- Сучасна архітектура (DDD + Event-Driven) vs застаріла 1С
- Швидкість Go (10-50x швидше)
- API-first підхід
- Ціна на 40-60% нижче Terrasoft
- Open-source Core

**Content**:
- Executive Summary (revenue projections, funding)
- Market Analysis (competitors, target segments)
- Technical Requirements (ПРРО, accounting, banking)
- Go-to-Market Strategy (beta → launch → growth)
- Pricing Strategy ($29-149/month tiers)
- Partnership Strategy (бухгалтерські фірми, IT-інтегратори)
- 12-Month Timeline (Q1 2026 - Q1 2027)

---

#### [Ukraine Compliance Roadmap](UKRAINE_COMPLIANCE_ROADMAP.md) ⭐
**Timeline**: Q1-Q2 2026 (Лютий - Червень)  
**Focus**: Technical implementation of Ukrainian compliance  
**Status**: Planning, ready to start

**Phases**:

**Phase 5: ПРРО Integration** (3-4 weeks, Лютий 2026)
- Week 7-8: Core ПРРО infrastructure
- Week 9-10: Checkbox + Вчасно.Каса APIs
- Deliverables: Фіскальні чеки, Z-звіти, X-звіти

**Phase 6: Податкові Накладні** (2-3 weeks, Березень 2026)
- Week 11-12: XML generation для ДПС
- Week 13: M.E.Doc integration
- Deliverables: Реєстрація в ЄРПН, декларація ПДВ

**Phase 7: Банківські Виписки** (1-2 weeks, Березень 2026)
- Week 13-14: Monobank, Privat24, PUMB APIs
- Deliverables: Auto-matching з рахунками

**Phase 8: HRM** (3-4 weeks, Квітень 2026)
- Week 14-17: Зарплата + податки (ЄСВ, ПДФО, Військовий збір)
- Deliverables: Форма 1ДФ, звіт ЄСВ

**Phase 9: Delivery** (1 week, Квітень 2026)
- Week 17: Нова Пошта API
- Deliverables: Створення ТТН, трекінг

**Technical Details**:
- Database schemas (ПРРО, accounting, banking)
- API implementations (Checkbox, M.E.Doc, Monobank)
- Event flows (Order → FiscalReceipt)
- Testing strategies (30+ unit, 15+ integration per module)

---

## 🎯 Priority Matrix

| Priority | Feature | Timeline | Impact | Market |
|----------|---------|----------|--------|--------|
| 🔴 БЛОКЕР | ПРРО Integration | Week 7-10 | 70% | Роздрібна торгівля |
| 🔴 БЛОКЕР | Податкові Накладні | Week 11-12 | 50% | Бухгалтерські фірми |
| 🟡 HIGH | Банківські Виписки | Week 13-14 | 60% | Автоматизація |
| 🟡 MEDIUM | HRM | Week 14-17 | 40% | Конкурентна перевага |
| 🟢 MEDIUM | Нова Пошта | Week 17 | 50% | E-commerce |

---

## 📅 Combined Timeline

```
Січень 2026:
════════════════════════════════════════════════════════
Week 1-2  : ✅ Phase 2 Complete (Domain Errors)
Week 3    : 🔄 Phase 3 Week 2 (Script Storage)
Week 4    : 🔄 Phase 3 Week 3 (UI Metadata)

Лютий 2026:
════════════════════════════════════════════════════════
Week 5-6  : ⏱️ Phase 4 (Scheduler)
Week 7-10 : 🇺🇦 ПРРО Integration (Checkbox + Вчасно.Каса)

Березень 2026:
════════════════════════════════════════════════════════
Week 11-12: 🇺🇦 Податкові Накладні + XML
Week 13-14: 🇺🇦 Банківські Виписки (Monobank, Privat24)

Квітень 2026:
════════════════════════════════════════════════════════
Week 14-17: 🇺🇦 HRM (Зарплата + Податки)
Week 17   : 🇺🇦 Нова Пошта API

Травень - Червень 2026:
════════════════════════════════════════════════════════
Week 18-24: ⚡ NATS Gateway + 📱 Mobile App
Week 25-26: 🎨 Polish + 📚 Documentation

Q3 2026 (Launch):
════════════════════════════════════════════════════════
Жовтень: 🚀 PUBLIC BETA LAUNCH
Листопад: 🐛 Bug fixes + User Feedback
Грудень: 💎 Premium Features

Q4 2026 Goals:
════════════════════════════════════════════════════════
- 500 active users
- 150 paying customers
- $15K MRR
```

---

## 🚀 Next Steps

### Immediate (Week 3-4, Січень 2026)
1. ✅ Завершити Phase 3 (LUA + UI Metadata)
2. 📝 Підготувати technical specs для ПРРО
3. 🎯 Знайти 10 beta-тестерів для ПРРО

### Short-term (Лютий 2026)
1. ⏱️ Завершити Phase 4 (Scheduler)
2. 🇺🇦 Почати ПРРО Integration
3. 📝 Зареєструвати ТОВ в Україні

### Medium-term (Березень - Квітень 2026)
1. 🇺🇦 Завершити всі українські комплаєнс модулі
2. 🎯 100 beta-тестерів
3. 💰 Перші 10 paying customers

### Long-term (Q3-Q4 2026)
1. 🚀 PUBLIC BETA LAUNCH
2. 📱 Mobile App (Flutter)
3. 🎯 500 active users, $15K MRR

---

## 📊 Success Metrics

### Technical KPIs
- **Availability**: 99.5% uptime
- **Performance**: < 200ms API response
- **Test Coverage**: 90%+
- **Ukrainian Compliance**: 100% законодавча відповідність

### Business KPIs (Ukrainian Market)

**Q1 2026**:
- 20 Beta users (ПРРО testing)
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

## 📚 Related Documentation

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
- [Business Overview (Ukrainian)](../business/BUSINESS_OVERVIEW_UK.md) - Для керівників

---

**Last Updated**: January 14, 2026  
**Status**: Active Planning & Development  
**Maintainer**: Promenade Team 🇺🇦
