# Strategic Roadmap 2026

**Created**: January 13, 2026  \
**Updated**: January 17, 2026 (Phase 3 completed; fiscal PRRO E2E + scheduler shift/retry wired)

---

## Scope

This document consolidates the 2026 strategic plan for core platform milestones and market expansion.

## Near-Term Priorities (Q1-Q2 2026)

1. Complete Phase 3 (script storage + UI metadata API)
2. Fiscal PRRO core end-to-end (receipt flow, order events, Checkbox sandbox)
3. Scheduler integration for fiscal retries, shifts, and daily Z-reports
4. Tax invoices + bank statements after PRRO is stable
5. Prepare beta program and partner documentation

## Reference Roadmaps

- docs/roadmap/PHASE3_LUA_UI_FOUNDATION.md
- docs/roadmap/UKRAINE_MARKET_STRATEGY_2026.md
- docs/roadmap/UKRAINE_COMPLIANCE_ROADMAP.md

---

## Current Status (January 2026)

- Phase 1–2 complete (architecture + domain errors refactor).
- Phase 3 complete: LUA engine + HTTP layer + script storage/versioning + UI metadata FormDefinition API.
- Scheduler core package production-ready (`pkg/scheduler`); fiscal retries + shift open/close jobs wired.
- Fiscal context in progress: cash register aggregate done; receipt aggregate and routes added; fiscal migrations extended; checkbox cancel + shift flows added.
- Order Management Contract aggregate complete (see docs/roadmap/PHASE3_CONTRACT_SUMMARY.md).

## Priority Sequence (Q1 2026)

1. Phase 3 completion (script storage + UI metadata API).
2. Fiscal PRRO core end-to-end (receipt creation → print/cancel → order event flow → Checkbox sandbox).
3. Scheduler integration for fiscal retries, shifts, and Z-report automation.
4. Tax invoices and bank statements after PRRO stabilizes.
5. Beta program + partner documentation.

## Q1 Milestones

- M1 (Done): Script storage + versioning (migrations, repository, use case).
- M2 (Done): UI metadata FormDefinition aggregate + API.
- M3 (Done): Receipt E2E flow with order events and handler error mapping.
- M4 (In Progress): Checkbox provider integration complete (sandbox validation).
- M5 (Planned): 10 retail beta testers onboarded.

## Next Steps

1. Add receipt + shift smoke tests (failure + compensation cases).
2. Finalize daily Z-report automation cadence + monitoring.
3. Complete Checkbox sandbox validation checklist (incl. shift/Z-report edge cases).
4. Prepare beta onboarding checklist and sample data.

<!--
*** End Patch

**Architecture**:
```
                    Scheduler System                        


                  pkg/scheduler/                         
                                                         
       
     Scheduler      Job Registry     Executor    
     (Cron)         (Database)      (Workers)    
       
                                                     
                  

                           
                           

              Job Implementations                        
                                                         
       
     Go Jobs        LUA Scripts    HTTP Webhooks 
    (Built-in)      (Dynamic)       (External)   
       

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
  id TEXT PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    
    -- Scheduling
    cron_expression VARCHAR(100) NOT NULL,
    timezone VARCHAR(50) DEFAULT 'UTC',
    enabled BOOLEAN DEFAULT true,
    
    -- Job definition
    job_type VARCHAR(20) NOT NULL,  -- 'go', 'lua', 'http'
    handler_name VARCHAR(255),      -- For Go jobs
    lua_script_id TEXT,             -- For LUA jobs
    http_url TEXT,                  -- For HTTP jobs
    http_method VARCHAR(10),        -- GET, POST, etc.
    http_headers TEXT,
    
    -- Execution control
    max_retries INTEGER DEFAULT 3,
    retry_delay_seconds INTEGER DEFAULT 60,
    timeout_seconds INTEGER DEFAULT 300,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP
);

CREATE TABLE scheduler_job_executions (
  id TEXT PRIMARY KEY,
  job_id TEXT REFERENCES scheduler_jobs(id),
    
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
    triggered_by_user_id TEXT
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


                   Flutter App                           
       
    NATS Client     WebSocket        UI Store    
       

                              
           Subscribe           HTTP/WS
                              

              Promenade NATS Gateway                     
                                                         
       
       NATS          WebSocket       Subject     
     Publisher        Server         Router      
       
                                                       

          
           Domain Events
          

                  Event Bus                              
  (Memory / Redis / NATS)                                

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
-  10+ LUA scripts created and tested
-  3+ UI forms defined via metadata
-  100% test coverage
-  Documentation complete

### Phase 4 (Scheduler)
-  5+ scheduled jobs running
-  0 missed executions
-  Retry mechanism working
-  LUA job integration

### Phase 5 (NATS Gateway)
-  < 100ms event latency
-  10K concurrent connections
-  Flutter real-time working
-  99.9% uptime

### Overall
-  Backend horizontally complete (all contexts production)
-  No-code flexibility (LUA + UI Metadata)
-  Real-time capabilities (NATS)
-  Automation (Scheduler)
-  100% test coverage maintained

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
1.  Завершити Phase 3 Week 2 (Script Storage) - 1 тиждень
2.  Завершити Phase 3 Week 3 (UI Metadata) - 1 тиждень
3.  Phase 4 Scheduler - 2 тижні
4.  Phase 5 NATS Gateway - 3-4 тижні

**Total Timeline**: 7-8 тижнів до повного completion

---

## Наступний Крок

Готовий почати **Phase 3 Week 2** (Script Storage)?

Або є питання щодо roadmap? Щось змінити в пріоритетах?

---

**Created**: January 13, 2026  
**Author**: Promenade Team  
**Status**: Draft for Review

-->
