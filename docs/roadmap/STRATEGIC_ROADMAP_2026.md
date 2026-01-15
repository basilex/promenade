# Strategic Roadmap 2026

**Created**: January 13, 2026  \
**Updated**: January 14, 2026 (added Ukrainian market strategy)

---

## Scope

This document consolidates the 2026 strategic plan for core platform milestones and market expansion.

## Near-Term Priorities (Q1-Q2 2026)

1. Complete Phase 3 (LUA scripting + UI metadata)
2. Complete Phase 4 (Scheduler)
3. Start Ukrainian compliance (fiscal, tax invoices, bank statements)
4. Prepare beta program and partner documentation

## Reference Roadmaps

- docs/roadmap/PHASE3_LUA_UI_FOUNDATION.md
- docs/roadmap/UKRAINE_MARKET_STRATEGY_2026.md
- docs/roadmap/UKRAINE_COMPLIANCE_ROADMAP.md
# Promenade Strategic Roadmap 2026

**Дата створення**: 13 січня 2026  
**Оновлено**: 14 січня 2026 (додано Ukrainian Market Strategy)  
**Статус**: Phase 2 Complete (100%), Phase 3-4 в progress  
**Горизонт**: Q1-Q2 2026 (6 місяців технічний roadmap)

> **🇺🇦 NEW**: Детальна стратегія для українського CRM/ERP ринку доступна в [UKRAINE_MARKET_STRATEGY_2026.md](UKRAINE_MARKET_STRATEGY_2026.md)  
> Включає: ПРРО інтеграцію, українську бухгалтерію, банківські виписки, HRM, і повний 12-місячний план завоювання ринку після заборони російського ПЗ (1С, Bitrix24).

---

## 🇺🇦 Ukrainian Market Opportunity

**CRITICAL UPDATE** (14 січня 2026): Заборона російського ПЗ в Україні створила величезну ринкову можливість.

**Market Size**:
- 150,000+ компаній МСБ шукають заміну 1С, Bitrix24, AmoCRM
- $500M+ ринок CRM/ERP в Україні
- 70% колишніх користувачів 1С без альтернативи

**Must-Have Features для Ukrainian Market**:
1. 🔥 **ПРРО Integration** (Програмний РРО) - БЛОКЕР для роздрібної торгівлі
2. 🔥 **Податкові Накладні** (XML генерація для ДПС) - БЛОКЕР для бухгалтерських фірм
3. 🔥 **Банківські Виписки** (Monobank, Privat24, PUMB APIs) - HIGH priority
4. 💼 **HRM з українськими податками** (ЄСВ, ПДФО, Військовий збір)
5. 🚚 **Нова Пошта / Укрпошта Integration**

**Competitive Advantage**:
- Сучасна архітектура (DDD + Event-Driven) vs застаріла 1С
- Швидкість Go (10-50x швидше)
- API-first підхід
- Ціна на 40-60% нижче Terrasoft BPM'online
- Open-source Core (прозорість і довіра)

**Target**: 5-10% ринку МСБ (7,500-15,000 компаній) за 18 місяців  
**Projected ARR**: $9.4M-29.9M при 7,500 клієнтів

**Detailed Strategy**: Див. [UKRAINE_MARKET_STRATEGY_2026.md](UKRAINE_MARKET_STRATEGY_2026.md) (87KB, 12-місячний план)

---

## Поточний Стан (Січень 2026)

### ✅ Завершено

**Phase 1: Architecture & Foundation** (Грудень 2025)
- ✅ DDD з Bounded Contexts (6 production contexts)
- ✅ Event Bus (Memory/Redis, 377K events/sec)
- ✅ Multi-Database Support (Postgres/SQLite/MySQL)
- ✅ JWT + RBAC + Rate Limiting
- ✅ Health Checks + Cache Layer
- ✅ Testing Infrastructure (4-tier: unit/smoke/integration/benchmark)

**Phase 2: Domain Errors Refactoring** (Січень 9-13, 2026)
- ✅ 18 sessions (100% complete)
- ✅ 24 aggregates refactored
- ✅ 225+ fmt.Errorf → domain errors
- ✅ 36 handlers з errors.Is()
- ✅ 1,548 lines documentation
- ✅ Gold Standard patterns established

**Phase 3: LUA Scripting (Week 1)** (Січень 8, 2026)
- ✅ LUA Engine з sandbox (memory 50MB, timeout 5s)
- ✅ Standard Library (Customer/Order/Deal/Query/Date)
- ✅ HTTP Layer (10 REST endpoints)
- ✅ 33 tests passing (21 unit + 12 smoke)
- 🔄 **IN PROGRESS**: Script storage + versioning (Week 2)

### 🎯 Стратегічні Цілі

**Горизонтальна Повнота Backend**:
- Bounded Contexts з повним lifecycle management
- Всі CRUD операції + business logic
- Event-driven communication
- Production-ready quality

**Гнучкість Без Перекомпіляції**:
- LUA scripts для business rules
- UI Metadata для динамічних форм
- Configuration-driven behavior

**Real-Time Capabilities**:
- NATS Gateway для Flutter streams
- WebSocket/SSE для browsers
- Event streaming architecture

**Автоматизація**:
- Cron/Scheduler для planned tasks
- Background jobs
- Recurring processes

---

## Пріоритизація: Що Робити Першим?

### Критерії Вибору

1. **Impact** - Скільки бізнес-функціоналу розблокує?
2. **Dependencies** - Що від цього залежить?
3. **Complexity** - Скільки часу на реалізацію?
4. **Urgency** - Наскільки критично зараз?

### Аналіз Компонентів

#### 1. LUA Engine (Phase 3 Continuation)
**Статус**: 70% done (Engine + HTTP ready, Storage pending)  
**Impact**: 🟢 HIGH - Enables no-code business logic  
**Dependencies**: None (standalone)  
**Complexity**: 🟡 MEDIUM (2 weeks remaining)  
**Urgency**: 🟡 MEDIUM  
**Priority**: **#1 - Завершити Phase 3**

**Що залишилось**:
- Week 2: Script storage (DB migrations + versioning)
- Week 3: UI Metadata System (FormDefinition aggregate)

**Reasoning**: Вже 70% зроблено, логічно завершити. Розблокує гнучкість системи.

---

#### 2. Scheduler/Cron System
**Статус**: 0% (відсутній)  
**Impact**: 🟢 HIGH - Background tasks, recurring jobs  
**Dependencies**: Event Bus (done), Contexts (done)  
**Complexity**: 🟢 LOW-MEDIUM (1-2 weeks)  
**Urgency**: 🔴 HIGH (критично для production)  
**Priority**: **#2 - Після Phase 3**

**Use Cases**:
- Expired contract cleanup
- Recurring invoices
- Scheduled notifications
- Data aggregation jobs
- Report generation

**Reasoning**: Критично для production, відносно швидко реалізувати.

---

#### 3. NATS Gateway
**Статус**: 0% (відсутній)  
**Impact**: 🟢 HIGH - Real-time UI, Flutter streams  
**Dependencies**: Event Bus (done), Stable Contexts  
**Complexity**: 🔴 HIGH (3-4 weeks)  
**Urgency**: 🟡 MEDIUM (after backend stable)  
**Priority**: **#3 - Коли backend самодостатній**

**Features**:
- Event Bus → NATS gateway
- Domain events → NATS subjects
- WebSocket/SSE for browsers
- JetStream for durability
- Flutter stream subscriptions

**Reasoning**: Вимагає стабільного backend. Робити коли функціонал повний.

---

#### 4. Horizontal Completion
**Статус**: 80% (6/7 contexts production)  
**Impact**: 🟡 MEDIUM - Feature completeness  
**Dependencies**: None  
**Complexity**: 🟡 MEDIUM (ongoing)  
**Urgency**: 🟢 LOW (continuous improvement)  
**Priority**: **#4 - Паралельно з іншими**

**Remaining**:
- Warehouse: 100% done ✅
- Scripting: Infrastructure context (exclude from main roadmap)
- Contract: Planned for Q2
- Notifications: Planned for Q2

---

## Рекомендований План

### 🎯 **Phase 3 Completion** (2 weeks, Січень 13-27)

**Week 2: Script Storage** (Січень 13-19)
```
✅ Day 1-2: Database migrations
   - scripting_scripts table
   - scripting_script_versions table
   
✅ Day 3-4: Script aggregate
   - Entity with versioning
   - Repository implementation
   - UseCase layer
   
✅ Day 5: Integration tests
   - Script CRUD
   - Version management
```

**Week 3: UI Metadata Foundation** (Січень 20-27)
```
✅ Day 1-3: FormDefinition aggregate
   - JSONB-based metadata storage
   - Multi-language support
   - RBAC integration
   
✅ Day 4-5: API + Documentation
   - REST endpoints
   - Examples
   - Integration with LUA events
```

**Deliverable**: Production-ready LUA Engine + UI Metadata System

---

### 🎯 **Phase 4: Scheduler System** (2 weeks, Січень 27 - Лютий 10)

**Week 1: Core Scheduler** (Січень 27 - Лютий 2)
```
✅ Day 1-2: Scheduler package
   - pkg/scheduler/scheduler.go
   - Cron expression parser
   - Job registry
   - Execution engine
   
✅ Day 3-4: Job persistence
   - Database migrations
   - Job storage
   - Execution history
   
✅ Day 5: Testing
   - Unit tests (30+)
   - Integration tests
```

**Week 2: Integration & Examples** (Лютий 3-10)
```
✅ Day 1-2: Context integration
   - Expired contracts cleanup (Order-Mgmt)
   - Recurring invoices (Billing)
   - Data aggregation (Analytics)
   
✅ Day 3-4: LUA integration
   - LUA scripts as scheduled jobs
   - Dynamic job creation
   
✅ Day 5: Documentation
   - Scheduler guide
   - Examples
   - API reference
```

**Deliverable**: Production Scheduler з LUA integration

---

### 🎯 **Phase 5: NATS Gateway** (3-4 weeks, Лютий 10 - Березень 10)

**Prerequisite**: Backend функціонал самодостатній (Contexts stable)

**Week 1: NATS Infrastructure** (Лютий 10-16)
```
✅ Day 1-2: NATS setup
   - Docker configuration
   - JetStream enable
   - Subject design
   
✅ Day 3-5: Event Bus → NATS adapter
   - pkg/bus/nats/nats_bus.go
   - Event publishing
   - Stream configuration
```

**Week 2: Gateway Layer** (Лютий 17-23)
```
✅ Day 1-3: NATS Gateway service
   - internal/gateway/nats/gateway.go
   - Event routing
   - Subject mapping
   
✅ Day 4-5: WebSocket bridge
   - WebSocket server
   - NATS subscription
   - Message transformation
```

**Week 3: Flutter Integration** (Лютий 24 - Березень 2)
```
✅ Day 1-2: NATS client setup
   - Flutter NATS package
   - Authentication
   - Stream subscriptions
   
✅ Day 3-5: Real-time features
   - Live notifications
   - Order status updates
   - Deal pipeline changes
```

**Week 4: Testing & Monitoring** (Березень 3-10)
```
✅ Day 1-2: Integration tests
   - E2E event flow
   - Latency tests
   
✅ Day 3-4: Monitoring
   - Metrics (events/sec, latency)
   - Dashboards
   
✅ Day 5: Documentation
   - NATS Gateway guide
   - Flutter integration
```

**Deliverable**: Production NATS Gateway з Flutter real-time

---

## Timeline Overview

```
Січень 2026:
════════════════════════════════════════════════════════
Week 1-2  : ✅ Phase 2 Complete (Domain Errors)
Week 3    : 🔄 Phase 3 Week 2 (Script Storage)
Week 4    : 🔄 Phase 3 Week 3 (UI Metadata)

Лютий 2026:
════════════════════════════════════════════════════════
Week 1    : ⏱️ Phase 4 Week 1 (Scheduler Core)
Week 2    : ⏱️ Phase 4 Week 2 (Scheduler Integration)
Week 3-4  : ⏱️ Phase 5 Week 1-2 (NATS Infrastructure)

Березень 2026:
════════════════════════════════════════════════════════
Week 1-2  : ⏱️ Phase 5 Week 3-4 (Flutter Integration)
Week 3-4  : 🎯 Horizontal Completion (Contract, Notifications)
```

---

## Детальні Технічні Вимоги

### Scheduler System Design

**Architecture**:
```
                    Scheduler System                        

┌─────────────────────────────────────────────────────────┐
│                  pkg/scheduler/                         │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Scheduler  │  │  Job Registry│  │   Executor   │ │
│  │   (Cron)     │  │  (Database)  │  │  (Workers)   │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
│         │                 │                  │         │
│         └─────────────────┴──────────────────┘         │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│              Job Implementations                        │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Go Jobs    │  │  LUA Scripts │  │ HTTP Webhooks│ │
│  │  (Built-in)  │  │  (Dynamic)   │  │  (External)  │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
```

**Job Types**:
1. **Go Jobs** - Compiled into binary (fast, type-safe)
2. **LUA Jobs** - Scripts from database (flexible, no recompile)
3. **HTTP Jobs** - Webhook calls (external integrations)

**Cron Expressions**:
```
Standard format: "minute hour day month weekday"
Examples:
- "0 2 * * *"      → Daily at 2 AM
- "*/15 * * * *"   → Every 15 minutes
- "0 9 * * 1-5"    → Weekdays at 9 AM
- "@daily"         → Shorthand for "0 0 * * *"
- "@hourly"        → Shorthand for "0 * * * *"
```

**Database Schema**:
```sql
CREATE TABLE scheduler_jobs (
    id UUID PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    
    -- Scheduling
    cron_expression VARCHAR(100) NOT NULL,
    timezone VARCHAR(50) DEFAULT 'UTC',
    enabled BOOLEAN DEFAULT true,
    
    -- Job definition
    job_type VARCHAR(20) NOT NULL,  -- 'go', 'lua', 'http'
    handler_name VARCHAR(255),      -- For Go jobs
    lua_script_id UUID,             -- For LUA jobs
    http_url TEXT,                  -- For HTTP jobs
    http_method VARCHAR(10),        -- GET, POST, etc.
    http_headers JSONB,
    
    -- Execution control
    max_retries INTEGER DEFAULT 3,
    retry_delay_seconds INTEGER DEFAULT 60,
    timeout_seconds INTEGER DEFAULT 300,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP
);

CREATE TABLE scheduler_job_executions (
    id UUID PRIMARY KEY,
    job_id UUID REFERENCES scheduler_jobs(id),
    
    -- Execution details
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    duration_ms INTEGER,
    status VARCHAR(20),  -- 'running', 'success', 'failed', 'timeout'
    
    -- Results
    output TEXT,
    error_message TEXT,
    retry_attempt INTEGER DEFAULT 0,
    
    -- Context
    triggered_by VARCHAR(50),  -- 'scheduler', 'manual', 'api'
    triggered_by_user_id UUID
);

CREATE INDEX idx_jobs_next_run ON scheduler_jobs(next_run_at) WHERE enabled = true;
CREATE INDEX idx_executions_job ON scheduler_job_executions(job_id, started_at DESC);
```

**Example Usage**:
```go
// Register Go job
scheduler.RegisterJob("cleanup_expired_contracts", &CleanupJob{
    contractUC: contractUseCase,
})

// Schedule from config
scheduler.AddJob(&Job{
    Name: "cleanup_expired_contracts",
    Cron: "0 2 * * *",  // Daily at 2 AM
    Type: JobTypeGo,
    Handler: "cleanup_expired_contracts",
})

// Schedule LUA job
scheduler.AddLUAJob(&Job{
    Name: "custom_report_generation",
    Cron: "0 9 * * 1",  // Mondays at 9 AM
    LUAScriptID: scriptID,
})

// Manual trigger
scheduler.TriggerNow("cleanup_expired_contracts", userID)
```

---

### NATS Gateway Design

**Architecture**:
```
                    NATS Gateway                           

┌─────────────────────────────────────────────────────────┐
│                   Flutter App                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │  NATS Client │  │  WebSocket   │  │   UI Store   │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
          │                    │
          │ Subscribe          │ HTTP/WS
          ▼                    ▼
┌─────────────────────────────────────────────────────────┐
│              Promenade NATS Gateway                     │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │     NATS     │  │   WebSocket  │  │   Subject    │ │
│  │   Publisher  │  │    Server    │  │   Router     │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
│         ▲                                              │
└─────────┼──────────────────────────────────────────────┘
          │
          │ Domain Events
          │
┌─────────┴───────────────────────────────────────────────┐
│                  Event Bus                              │
│  (Memory / Redis / NATS)                                │
└─────────────────────────────────────────────────────────┘
```

**Subject Design**:
```
promenade.{context}.{aggregate}.{action}

Examples:
- promenade.identity.user.registered
- promenade.customer.deal.won
- promenade.order.order.confirmed
- promenade.billing.invoice.paid
- promenade.warehouse.inventory.low_stock

Wildcards for subscriptions:
- promenade.order.*        → All order events
- promenade.*.deal.*       → All deal events
- promenade.*.*.confirmed  → All confirmed events
```

**WebSocket Protocol**:
```json
// Client → Server: Subscribe
{
  "action": "subscribe",
  "subjects": ["promenade.order.*", "promenade.billing.invoice.paid"],
  "auth": "Bearer <jwt_token>"
}

// Server → Client: Event
{
  "subject": "promenade.order.order.confirmed",
  "timestamp": "2026-01-13T12:00:00Z",
  "data": {
    "order_id": "01HXYZ...",
    "customer_id": "01HABC...",
    "total": 125000
  }
}

// Client → Server: Unsubscribe
{
  "action": "unsubscribe",
  "subjects": ["promenade.order.*"]
}
```

**Security**:
- JWT authentication required
- User permissions checked
- Subject-level ACL (User can only see own orders)
- Rate limiting per connection

**Performance Targets**:
- Latency: < 50ms (event → client)
- Throughput: 10K events/sec
- Concurrent connections: 10K+
- Message size: < 64KB

---

## Success Metrics

### Phase 3 (LUA Engine)
- ✅ 10+ LUA scripts created and tested
- ✅ 3+ UI forms defined via metadata
- ✅ 100% test coverage
- ✅ Documentation complete

### Phase 4 (Scheduler)
- ✅ 5+ scheduled jobs running
- ✅ 0 missed executions
- ✅ Retry mechanism working
- ✅ LUA job integration

### Phase 5 (NATS Gateway)
- ✅ < 100ms event latency
- ✅ 10K concurrent connections
- ✅ Flutter real-time working
- ✅ 99.9% uptime

### Overall
- ✅ Backend horizontally complete (all contexts production)
- ✅ No-code flexibility (LUA + UI Metadata)
- ✅ Real-time capabilities (NATS)
- ✅ Automation (Scheduler)
- ✅ 100% test coverage maintained

---

## Risk Assessment

### Phase 3 (LUA)
**Risk**: LOW  
**Reason**: 70% done, clear scope  
**Mitigation**: Follow existing patterns

### Phase 4 (Scheduler)
**Risk**: LOW-MEDIUM  
**Reason**: Well-understood domain  
**Mitigation**: Use proven libraries (robfig/cron)

### Phase 5 (NATS)
**Risk**: MEDIUM-HIGH  
**Reason**: New technology, complex integration  
**Mitigation**: 
- Start after backend stable
- Prototype first
- Gradual rollout

---

## Альтернативні Підходи

### Option A: Finish Phase 3 → Scheduler → NATS (Recommended)
**Pros**: Логічна послідовність, кожна фаза розблокує наступну  
**Cons**: NATS later (but backend needs stability first)  
**Timeline**: 8 weeks total

### Option B: Scheduler First → Finish Phase 3 → NATS
**Pros**: Scheduler critical for production  
**Cons**: Переривання Phase 3 momentum  
**Timeline**: 8 weeks total

### Option C: Parallel Teams
**Pros**: Faster completion  
**Cons**: Resource constraints, coordination overhead  
**Timeline**: 5-6 weeks (if 2 devs)

---

## Рекомендація

**Обираю Option A**: Finish Phase 3 → Scheduler → NATS

**Reasoning**:
1. Phase 3 вже 70% done - логічно завершити
2. LUA Engine потрібен для Scheduler (LUA jobs)
3. NATS потребує стабільного backend - робити останнім
4. Послідовний підхід = менше context switching

**Next Immediate Steps**:
1. ✅ Завершити Phase 3 Week 2 (Script Storage) - 1 тиждень
2. ✅ Завершити Phase 3 Week 3 (UI Metadata) - 1 тиждень
3. ✅ Phase 4 Scheduler - 2 тижні
4. ✅ Phase 5 NATS Gateway - 3-4 тижні

**Total Timeline**: 7-8 тижнів до повного completion

---

## Наступний Крок

Готовий почати **Phase 3 Week 2** (Script Storage)?

Або є питання щодо roadmap? Щось змінити в пріоритетах?

---

**Created**: January 13, 2026  
**Author**: Promenade Team  
**Status**: Draft for Review
