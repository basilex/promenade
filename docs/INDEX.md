# Promenade Platform - Documentation

**Modern backend platform** for customer management, orders, and business workflows built with **Domain-Driven Design**, **Event-Driven Architecture**, and **Clean Code principles**.

---

##  For Business Decision-Makers

**Non-technical overview of platform capabilities, value proposition, and implementation status**

Comprehensive business documentation available in multiple languages:

| Language | Document | Target Audience |
|----------|----------|-----------------|
|  English | [Business Overview](business/BUSINESS_OVERVIEW.md) | Executives, managers, investors |
|  Ukrainian | [Business Overview](business/BUSINESS_OVERVIEW_UK.md) | Executives, managers, investors |
|  Deutsch | [Geschäftsübersicht](business/BUSINESS_OVERVIEW_DE.md) | Führungskräfte, Manager, Investoren |

**What's included**: Executive summary, value propositions, core capabilities (CRM, Orders, Warehouse, Billing), use cases with ROI, deployment options, roadmap Q1-Q4 2026, success metrics.

**Status**: 80% complete, 182+ API endpoints, 2465+ automated tests, 90%+ code coverage.

---

## Core Concepts

### Domain-Driven Design (DDD)

**Bounded Contexts** isolate business domains with clear boundaries and autonomous evolution.

- **Concept**: Each context owns its domain model, database schema, and business logic
- **Communication**: Contexts interact only via Event Bus (no direct dependencies)
- **Implementation**: [Clean Architecture Summary](concepts/clean-architecture.md)

**Key patterns**: Aggregates, Entities, Value Objects, Domain Events, Repositories, Use Cases

---

### Event-Driven Architecture

**Event Bus** is the central nervous system for asynchronous, decoupled communication.

- **Adapters**: Memory (377K events/sec), Redis (distributed)
- **Features**: Retry policy, panic recovery, graceful shutdown
- **Implementation**: [Event Bus Guide](../pkg/bus/README.md)

**67 tests**, 100% passing | [Test Coverage Report](reference/bus-test-coverage.md)

---

### Authentication & Authorization

**JWT + RBAC** provide secure, stateless authentication with fine-grained access control.

- **JWT**: 15-min access tokens, 7-day refresh tokens, Redis revocation
- **RBAC**: 5 system roles, 29+ permissions, flexible role assignment
- **Rate Limiting**: IP-based protection (Login: 5/min, Register: 3/min)
- **Implementation**: [RBAC Guide](guides/rbac.md) | [Rate Limiting](guides/rate-limiting.md)

**27 tests** (JWT) + **25 tests** (Middleware)

---

### API Documentation (Swagger/OpenAPI)

**Interactive API documentation** with Swagger UI for exploring and testing all endpoints.

- **OpenAPI 3.0**: Auto-generated specification from code annotations
- **Swagger UI**: Interactive browser interface at `/api/docs/index.html`
- **182+ Endpoints**: All contexts fully documented
- **Try-It-Out**: Test APIs directly from browser
- **JWT Integration**: Bearer token authentication in UI
- **Export Formats**: JSON/YAML for Postman, Insomnia
- **Implementation**: [API Documentation Guide](guides/api-documentation.md)

**Quick Access**: `make dev` → http://localhost:8081/api/docs/index.html

---

### Postman Collection

**Pre-built Postman collection** with 182+ endpoints, authentication flows, and automated testing.

- **Auto-Generated**: OpenAPI spec → Postman collection (33K lines)
- **Multi-Environment**: Dev/Staging/Prod configs with pre-configured variables
- **Authentication Flow**: Auto-save JWT tokens after login, auto-refresh expired tokens
- **Test Automation**: Pre-request and test scripts for validation and token management
- **Newman Integration**: CI/CD ready with `newman run` command
- **Common Workflows**: 5 documented use cases (auth, customer, pipeline, order, billing)
- **Implementation**: [Postman Collection Guide](../postman/README.md)

**Quick Start**: Import `postman/Promenade_API.postman_collection.json` + `Development.postman_environment.json`

---

### API Versioning

**URL-based versioning strategy** for safe API evolution without breaking clients.

- **Versioning Approach**: URL-based (`/api/v1`, `/api/v2`) - clear, cacheable, simple
- **Deprecation Policy**: 12-month support window, RFC 8594 compliant headers
- **Breaking Changes**: 7 categories with clear definitions (field removal, type changes, etc.)
- **Migration Guides**: Templates with before/after code examples
- **Version Detection**: API endpoints for checking version status and sunset dates
- **Implementation**: [Versioning Strategy](guides/api-versioning.md) | [Practical Examples](guides/api-versioning-examples.md)

**37 middleware tests**, 100% passing | Production-ready deprecation/sunset middleware

---

### Health Monitoring

**Comprehensive health checks** monitor all dependencies with graceful degradation.

- **Endpoints**: `/health`, `/health/db`, `/health/redis`, `/health/bus`
- **Status Levels**: healthy, degraded, unhealthy
- **Features**: 5-second timeout, proper HTTP codes, Kubernetes-ready
- **Implementation**: [Health Checks Guide](guides/health-checks.md)

**21 tests**, 100% passing

---

### Local CI Validation

**Local CI simulation** catches issues before pushing to GitHub.

- **Commands**: `make pre-push`, `make ci-lint`, `make ci-test`, `make ci-build`
- **Time Savings**: 4+ minutes saved per failed push
- **GitHub Actions Parity**: Exact same checks as CI pipeline
- **Features**: golangci-lint, race detector, build validation
- **Implementation**: [Local CI Guide](guides/local-ci.md)

**Usage**: `make pre-push` before every git push (REQUIRED)

---

### Caching Layer

**Redis-based caching system** improves performance and reduces database load.

- **Adapters**: Redis (production), NoOp (testing)
- **TTL Strategy**: Reference (1h-24h), User (10-30m), Session (30m-1h)
- **Pattern**: Cache-aside with write-through invalidation
- **Features**: Graceful degradation, pattern-based deletion, JSON marshaling
- **Implementation**: [Caching Guide](guides/caching.md)

**Cached**: Countries, Currencies, Languages, Timezones, Profiles, Customers

---

### Multi-Database Support

**Database-agnostic architecture** supporting multiple SQL databases.

- **Dialect Pattern**: Abstracts SQL syntax differences (placeholders, JSON types)
- **Supported**: PostgreSQL (production), SQLite (dev/test), MySQL (planned), SQL Server (planned)
- **JSON Abstraction**: `jsonstore.Field[T]` stores JSON as TEXT across all databases
- **UUID Generation**: All IDs in Go code (`uuidv7.New()`), not DB defaults
- **Timestamp Management**: `.Touch()` updates timestamps (no DB triggers)
- **Implementation**: [Database Adapters](guides/database-adapters.md) | [JSONB Strategy](guides/jsonb-strategy.md)

**101 Tests**: pkg/database + pkg/jsonstore with 100% coverage

---

### Workspace Management

**Explicit workspace state management** for seamless multi-database development.

- **Single Source of Truth**: `.promenade.workspace` file defines `DATABASE_DRIVER` and `ENVIRONMENT`
- **Fail-Fast Validation**: Commands validate workspace before execution
- **Smart Switchers**: 9 commands cover all database × environment combinations
- **Environment Awareness**: Runners (`dev`, `test-all`, `prod`) enforce correct environment usage
- **Modular Architecture**: Main Makefile (workspace + help) + specialized modules (dev/test/prod)
- **Workspace Commands**: Go tools (`build`, `fmt`, `lint`) grouped logically with state management
- **Implementation**: [Workspace Management Guide](guides/workspace-management.md)

**Key Benefits**:
- No command explosion (50+ commands work with all databases)
- Prevents wrong environment execution (fail-fast validation)
- Natural developer workflow (configure once, work anywhere)
- Clear state visibility (`make workspace` shows current config)

**Example**: `make switch-postgres-dev && make dev` → PostgreSQL development ready

---

### Customer Management

**Complete CRM functionality** for B2C and B2B customer lifecycle management.

- **Customer Lifecycle**: Lead → Prospect → Customer → Churned (state machine)
- **B2C & B2B Support**: Individual consumers and business contacts
- **Customer Segmentation**: Tiers (free, basic, pro, enterprise) + flexible tags
- **Sales Pipeline**: Assign to reps, track source and status
- **14 Endpoints**: Complete CRUD + business logic operations
- **Implementation**: [Customer Management Guide](concepts/customer-management.md)

**Live tested**: All 14 endpoints working | State transitions validated

---

### Deal Management

**Complete sales pipeline management** with deal lifecycle tracking and **fully implemented business rules**.

- **Deal Lifecycle**: Track deals through stages (lead → qualified → proposal → negotiation → closed)
- **Probability Tracking**: Automatic probability calculation per stage (10% → 100%)
- **Money Handling**: Type-safe Money value object for deal values
- **Pipeline Statistics**: Real-time stats by stage and won deals analytics
- **12 API Endpoints**: Complete CRUD + business logic operations
- **Stage Transitions**: Enforced state machine (lead → ... → closed_won/closed_lost)

**Key Features**:
- State machine enforces valid stage transitions
- Automatic probability updates per stage
- Win/Loss tracking with actual close dates
- Filter deals by stage, customer, or sales rep
- Pipeline and revenue statistics
- Integration with Customer aggregate

**Implementation Status**: All core business rules implemented in entity (`MoveToStage()`, `MarkWon()`, `MarkLost()`). Tested via HTTP API with 100% success rate.

**See**: [Deal Management Guide](concepts/deal-management.md) for complete documentation

---

### Interaction Management

**Comprehensive customer interaction tracking** for calls, emails, meetings, and notes with outcome tracking and follow-up management.

- **Interaction Types**: Track calls (inbound/outbound), emails, meetings, notes
- **Interaction Lifecycle**: Create → In Progress → Completed with outcome tracking
- **JSON Attendees**: Flexible participant tracking with TEXT-stored JSON arrays
- **Follow-up Management**: Flag interactions requiring follow-up with dates and notes
- **Duration Tracking**: Automatic duration calculation for ended interactions
- **Performance Optimized**: LEFT JOIN queries prevent N+1 problem when listing interactions
- **14 API Endpoints**: Complete CRUD + business logic operations

**Key Features**:
- Flexible interaction types and directions
- Outcome tracking (successful, failed, no_answer, scheduled, cancelled)
- Multi-participant support via JSON attendees array (stored as TEXT)
- Follow-up scheduling and tracking
- Company association for B2B interactions
- Time tracking with started_at, ended_at, duration_sec

**Implementation Status**: All core business rules implemented. N+1 optimization with LEFT JOIN. Fully tested with 74 tests + 2 benchmarks.

**See**: [Interaction Management Guide](concepts/interaction-management.md) for complete documentation

---

### Customer Analytics

**CQRS read models** for business intelligence and reporting.

- **9 Query Methods**: Optimized analytical queries (direct SQL, no repository pattern)
- **8 GET Endpoints**: Customer, deal, sales rep, revenue, and interaction analytics
- **Real-time Metrics**: Customer overview, deal pipeline, sales rep performance
- **Time Series**: Revenue trends with configurable granularity (day/week/month)
- **Funnel Analysis**: Customer lifecycle transitions and deal stage conversions
- **Performance Optimized**: Denormalized queries with LEFT JOIN across aggregates
- **Implementation**: [Analytics README](../internal/contexts/customer-mgmt/analytics/README.md)

**Key Characteristics**:
- CQRS pattern separates read (analytics) from write (CRUD)
- Direct DB access bypasses repository for maximum performance
- Aggregations span multiple entities (Customer + Deal + Interaction)
- No business logic (read-only reporting)

**All 8 endpoints live**: Customer overview, lifecycle, segmentation, deal pipeline, conversions, sales rep performance, revenue time series, interaction insights

---

### Testing Strategy

**Four-tier testing** with clear separation and professional organization.

- **Unit Tests**: In-place, fast feedback (~5s) - 2232+ tests
- **Smoke Tests**: HTTP handler validation, no DB (~2s) - 196 tests, 100% pass rate
- **Integration Tests**: Real database, full E2E (~14s) - 19 packages
- **Benchmark Tests**: Performance measurement with real DB
- **Implementation**: [Testing Patterns](guides/testing-patterns.md) | [Quick Reference](guides/testing-quick-reference.md) | [Smoke Tests Guide](../test/smoke/README.md)

**2465+ tests total**, 90%+ coverage | [Test Report](reference/test-coverage-report.md)

---

### Order Management

**Complete order lifecycle management** with **fully implemented business rules**.

- **Order Creation**: Create orders with currency support
- **Line Items**: Add/remove/update products with automatic totals
- **State Machine**: pending → confirmed → processing → fulfilled - **fully enforced**
- **Business Rules**: All validations implemented in entity methods
- **Money Handling**: Type-safe Money value object (cents-based)
- **14 Endpoints**: Complete CRUD + business logic
- **Implementation**: [Order Management Guide](concepts/order-management.md)

**Status**: All core business logic implemented

---

### Fulfillment Saga

**Distributed transaction orchestration** for order fulfillment process - coordinates Payment, Inventory, and Shipping operations with automatic compensation on failures.

**Concept**: The Fulfillment Saga implements the **Saga pattern** to manage complex, multi-step order fulfillment without requiring distributed ACID transactions. It ensures data consistency across multiple bounded contexts (Payment, Inventory, Shipping) through choreographed state transitions and compensating actions.

**Design**:
- **State Machine**: 7 states (pending → payment_processing → inventory_processing → shipping_processing → completed)
- **Compensation Logic**: Automatic rollback on failures (compensating → compensated → cancelled)
- **Optimistic Locking**: Version-based concurrency control prevents lost updates
- **JSON Storage**: Flexible arrays for completed steps and reserved items (stored as TEXT)
- **UTC Timestamps**: Consistent timezone handling across all operations
- **Idempotent Steps**: Safe to retry operations without side effects

**Key Features**:
- Orchestrator pattern coordinates all fulfillment steps
- PostgreSQL persistence with full ACID guarantees
- Concurrent saga execution with conflict detection
- Monitoring endpoint for in-progress sagas
- 100% test coverage (57 tests: 42 unit + 15 integration)

**State Flow**:
```
pending → payment_processing → inventory_processing → shipping_processing → completed
              ↓ (on failure)
         compensating → compensated → cancelled
```

**Implementation Status**: Production-ready. All 57 tests passing (100%). Repository with optimistic locking. Orchestrator with compensation logic.

**See**: [Fulfillment Saga README](../internal/contexts/order-mgmt/fulfillment/README.md) for complete implementation guide

---

### Warehouse Management

**Complete inventory management system** with stock tracking, product catalog, warehouse locations, and **automated order-inventory integration**.

- **Product Catalog**: Manage products with SKUs, categories, brands, and physical properties
- **Inventory Tracking**: Real-time stock levels (on hand, reserved, available, committed)
- **Stock Movements**: Immutable audit trail for all inventory changes
- **Warehouse Locations**: Hierarchical location management (warehouses, zones, bins)
- **Warehouse Integration**: Automated synchronization with Order Management via Event Bus
- **4 Aggregates**: Product, Inventory, StockMovement, Location (all production ready)
- **58 Endpoints**: Complete CRUD + business logic + integration operations
- **Implementation**: [Warehouse Management Guide](concepts/warehouse-management.md) | [Context README](../internal/contexts/warehouse/README.md)

**Key Architecture**:
```
Order.Confirm() → order.confirmed event → ReservationService.ReserveForOrder()
Order.Cancel()  → order.cancelled event → ReservationService.ReleaseForOrder()
Order.Fulfill() → order.fulfilled event → ReservationService.CommitForOrder()
```

**Status**: Production ready (433 tests passing) | Automated order-inventory integration operational

---

### LUA Scripting Engine

**Embedded LUA scripting engine** for Promenade Platform - enables dynamic business logic without Go recompilation.

**Concept**: Business analysts and power users can create custom validation rules, workflows, pricing logic, and automations using LUA scripts stored in the database. Scripts execute in a secure sandbox with controlled access to Promenade APIs.

**Design**:
- **Sandboxed Execution**: Memory limits (50MB), CPU timeout (5s), restricted filesystem/network access
- **Standard Library**: Safe access to Promenade APIs (Customer, Order, Deal, Notify, Query, Date modules)
- **Context Support**: Cancellation and timeout via Go context
- **Script Storage**: Scripts stored in database with versioning
- **Type Conversion**: Automatic conversion between LUA and Go types

**Key Features**:
- Execute LUA scripts with parameters and return values
- Validate scripts before saving (syntax checking)
- Timeout protection prevents infinite loops
- Panic recovery for handler safety
- 33 tests (21 unit + 12 smoke) - 100% passing
- **10 REST Endpoints**: Complete HTTP API at `/api/v1/scripts/*`

**Use Cases**:
```lua
-- Dynamic pricing based on customer tier
function calculatePrice(basePrice, customerTier)
    if customerTier == "premium" then
        return basePrice * 0.8  -- 20% discount
    end
    return basePrice
end

-- Auto-approve deals under threshold
function shouldAutoApprove(deal)
    local tier = Customer.GetTier(deal.customer_id)
    return tier == "premium" and deal.amount < 10000
end
```

**Implementation Status**:  HTTP Layer Complete (Week 1 Day 4 Complete) | Standard Library integrated with real UseCases | 10 REST endpoints operational

**Backend Flow Examples**:
- `Customer.GetTier(id)` → CustomerUseCase.GetCustomer() → returns cust.Tier
- `Order.GetTotal(id)` → OrderUseCase.GetOrder() → returns ord.Total.Amount (Money value object)
- `Deal.Approve(id)` → DealUseCase.MarkDealAsWon() → executes real business logic
- `Query.Execute(sql)` → sqlx.DB.QueryContext() → secure SELECT-only execution

**See**: [LUA Scripting Guide](../pkg/scripting/README.md) for complete documentation with examples

---

### Job Scheduler

**Production-ready cron-based job scheduler** for Promenade Platform - enables scheduled workflows, maintenance tasks, and event-driven automation without manual intervention.

**Concept**: The Job Scheduler implements a robust cron-based scheduling system for executing periodic and time-based tasks. It provides a worker pool architecture for concurrent job execution with retry logic, health monitoring, and graceful shutdown. Business users can schedule LUA scripts, event publications, HTTP webhooks, and custom operations using familiar cron expressions.

**Design**:
- **Cron-Based Scheduling**: Standard cron expressions for flexible timing (`0 2 * * *` = daily at 2 AM)
- **Worker Pool**: Configurable concurrent job execution with goroutine management
- **Job Types**: LUA scripts, Event Bus publishing, HTTP requests, Custom executors
- **Retry Logic**: Exponential backoff for failed jobs (configurable attempts and delays)
- **Graceful Lifecycle**: Start/Stop with context cancellation and worker coordination
- **Health Monitoring**: Built-in health checks with configurable intervals

**Key Features**:
- Execute scheduled jobs with 4 job types (LUA, Event, HTTP, Custom)
- Worker pool with configurable concurrency (default: 10 workers)
- Retry failed jobs with exponential backoff (default: 3 attempts, 60s delay)
- Per-job timeout configuration (default: 5 minutes, overrideable)
- Job lifecycle management (add/update/remove/disable operations)
- Panic recovery prevents single job failure from crashing engine
- Health monitoring logs worker status and queue length
- **43 tests** (91.5% coverage) including 5 integration scenarios

**Architecture**:
```
Cron Trigger → Job Queue → Worker Pool → Executor → Result → Status Update
                  ↓            ↓             ↓
            (buffered)   (goroutines)   (pluggable)
```

**Use Cases**:
- **Periodic Cleanup**: Delete old logs, temp files, expired sessions
- **Event Publishing**: Schedule daily reports, notifications, data syncs
- **HTTP Webhooks**: Call external APIs for inventory sync, status updates
- **LUA Automation**: Execute custom business logic on schedule
- **Backup Jobs**: Database backups, file exports, data archival
- **Monitoring Tasks**: Health checks, metric collection, alerting

**Job Types**:
- **LUA** (`JobTypeLUA`): Execute LUA scripts with sandboxed environment
- **Event** (`JobTypeEvent`): Publish events to Event Bus for async processing
- **HTTP** (`JobTypeHTTP`): Make HTTP requests to external APIs/webhooks
- **Custom** (`JobTypeCustom`): Implement custom executors for specific logic

**Implementation Status**:  **Production Ready** - 43 tests passing (91.5% coverage), 5 integration scenarios validated, graceful lifecycle management operational

**See**: [Job Scheduler Guide](../pkg/scheduler/README.md) for complete documentation with architecture details, advanced usage patterns, 4 practical examples, API reference, troubleshooting guide, and feature roadmap

---

## Architecture

### Bounded Contexts

| Context                 | Status     | Aggregates                                                  | Documentation                                     |
| ----------------------- | ---------- | ----------------------------------------------------------- | ------------------------------------------------- |
| **Shared**              | Production | Country, Currency, Language, Timezone                       | [Context Guide](../internal/contexts/shared/README.md)      |
| **Identity**            | Production | User, Contact, Profile, Role, Permission                    | [Context Guide](../internal/contexts/identity/README.md)    |
| **Customer Management** | Production | Customer, Company, Deal, Interaction, Analytics (all live)  | [Customer Guide](concepts/customer-management.md) \| [Analytics](../internal/contexts/customer-mgmt/analytics/README.md) |
| **Order Management**    | Production | Order, OrderLine (live) \| Contract, Fulfillment (planned) | [Order Guide](concepts/order-management.md)                   |
| **Billing**             | Production | Invoice, Payment, Subscription (all live)                   | [Invoice Guide](concepts/invoice-management.md) \| [Payment Guide](concepts/payment-management.md) |
| **Warehouse**           | Production (100%) | Inventory, StockMovement, Product, Location (all live) | [Context Guide](../internal/contexts/warehouse/README.md) |
| **Fiscal**              | In progress | CashRegister, Receipt (in progress)                         | [Context Guide](../internal/contexts/fiscal/README.md) |

**Read**: [Bounded Contexts Overview](concepts/bounded-contexts.md)

---

### Package Library

**Reusable, context-agnostic components** for domain primitives and infrastructure.

| Package         | Purpose                            | Tests | Documentation                              |
| --------------- | ---------------------------------- | ----- | ------------------------------------------ |
| **bus**         | Event Bus (Memory/Redis)           | 67    | [Guide](../pkg/bus/README.md)              |
| **jwt**         | Authentication & RBAC              | 18    | [Guide](../pkg/jwt/README.md)              |
| **logger**      | Structured logging                 | 15    | [Guide](../pkg/logger/README.md)           |
| **middleware**  | HTTP middleware (rate limit, CSRF) | 25    | [Guide](../pkg/middleware/README.md)       |
| **cache**       | Redis-based caching layer          | 8     | [Guide](../pkg/cache/README.md)            |
| **uuidv7**      | Time-ordered UUIDs                 | 10    | [Guide](../pkg/uuidv7/README.md)           |
| **valueobject** | Domain value objects               | 45    | [Guide](../pkg/valueobject/README.md)      |
| **response**    | HTTP response helpers              | 13    | [Guide](../pkg/response/README.md)         |
| **migration**   | Database migrations                | 8     | [Guide](../pkg/migration/README.md)        |
| **saga**        | Distributed transactions           | 28    | [Guide](../pkg/saga/README.md)             |
| **aggregate**   | Base aggregate pattern             | 5     | [Guide](../pkg/aggregate/README.md)        |
| **jsonb**       | PostgreSQL JSONB utilities (optional) | 8  | [Guide](../pkg/jsonb/README.md)            |
| **scripting**   | LUA scripting engine               | 21    | [Guide](../pkg/scripting/README.md)        |
| **scheduler**   | Cron-based job scheduler           | 43    | [Guide](../pkg/scheduler/README.md)        |

**Total**: 291+ tests across 13 packages | [Package Overview](../pkg/README.md)

---

## Documentation Sections

### Concepts

Fundamental architectural principles and design patterns:

- [Clean Architecture with DDD](concepts/clean-architecture.md) - Bounded Contexts, Aggregates, Value Objects
- [Event-Driven Architecture](concepts/event-driven.md) - Event Bus, Domain Events, Sagas
- [Bounded Contexts Strategy](concepts/bounded-contexts.md) - Context isolation and communication
- [Customer Management](concepts/customer-management.md) - CRM functionality, customer lifecycle, B2C & B2B
- [Company Management](concepts/company-management.md) - B2B organizations, hierarchical structure, legal entities
- [Deal Management](concepts/deal-management.md) - Sales pipeline, deal stages, probability tracking
- [Interaction Management](concepts/interaction-management.md) - Customer interaction tracking, calls, emails, meetings, notes
- [Order Management](concepts/order-management.md) - Complete order lifecycle, state machine, business rules
- [Warehouse Management](concepts/warehouse-management.md) - Inventory tracking, stock movements, automated order-inventory integration

### Guides

Step-by-step implementation guides:

**Developer Portal**:
- [Quick Start Guide](guides/quick-start.md) - 5-minute practical tutorial with curl examples
- [Authentication Flow](guides/authentication-flow.md) - Complete JWT authentication documentation
- [Common Use Cases](guides/common-use-cases.md) - 7 real-world business scenarios
- [Troubleshooting Guide](guides/troubleshooting.md) - Common issues and solutions

**API & Integration**:
- [API Versioning Strategy](guides/api-versioning.md) - URL-based versioning, deprecation policy, RFC 8594 headers
- [API Versioning Examples](guides/api-versioning-examples.md) - Practical code examples for versioning implementation
- [Local CI Validation](guides/local-ci.md) - Run GitHub Actions checks locally before push
- [RBAC Implementation](guides/rbac.md) - Roles, permissions, JWT integration
- [Rate Limiting](guides/rate-limiting.md) - IP-based protection for authentication
- [CSRF Protection](guides/csrf-protection.md) - Cross-site request forgery defense
- [Health Checks](guides/health-checks.md) - Dependency monitoring and alerting
- [Caching](guides/caching.md) - Redis caching layer implementation
- [Testing Patterns](guides/testing-patterns.md) - Four-tier testing strategy

**Design Patterns & Conventions**:
- [Naming Conventions](guides/naming-conventions.md) - Files, directories, Go code naming standards
- [Database Conventions](guides/database-conventions.md) - Tables, columns, indexes, migrations
- [Architecture Patterns](guides/architecture-patterns.md) - Repository, UseCase, Handler, Value Object patterns
- [UI Metadata Guide](guides/ui-metadata.md) - FormDefinition schema, JSON metadata, API endpoints
- [Domain Errors Guide](guides/domain-errors.md) - Gold Standard error handling (644 lines, 8 sections, comprehensive patterns)
- [Documentation Style Guide](guides/documentation-style-guide.md) - Official no-emoji policy with automated enforcement (`scripts/clean-emojies.py`)

### Reference

Technical specifications and detailed documentation:

- [API Reference](reference/api-reference.md) - Complete HTTP API documentation
- [Test Coverage Report](reference/test-coverage-report.md) - 450+ tests breakdown
- [Bus Test Coverage](reference/bus-test-coverage.md) - Event Bus test report
- [Index Audit Report](reference/index-audit-report.md) - Database index analysis (13 tables, 60+ indexes)
- [Soft Delete Audit](reference/soft-delete-audit.md) - Soft delete implementation audit (78 queries, 100% compliance)
- [N+1 Optimization](reference/n-plus-one-optimization.md) - Query optimization report (95% reduction)
- [Table Naming Strategy](reference/table-naming-strategy.md) - Database table naming conventions
- [Migration History](reference/migration-history.md) - Database schema evolution
- [Configuration Reference](reference/configuration.md) - YAML config options

### Roadmaps

Strategic planning and implementation timelines:

- [Strategic Roadmap 2026](roadmap/STRATEGIC_ROADMAP_2026.md) - Q1-Q2 2026 complete implementation plan (LUA, Scheduler, NATS)
- [Ukraine Market Strategy 2026](roadmap/UKRAINE_MARKET_STRATEGY_2026.md) - NEW: Strategy to capture the Ukrainian CRM/ERP market (fiscal compliance, accounting, banking, 12-month plan)
- [Ukraine Compliance Roadmap](roadmap/UKRAINE_COMPLIANCE_ROADMAP.md) - NEW: Technical implementation of Ukrainian compliance (Q1-Q2 2026, fiscal receipts, tax invoices, bank statements, HRM)
- [Phase 3: LUA + UI Metadata](roadmap/PHASE3_LUA_UI_FOUNDATION.md) - 3-week implementation (January 8-28, 2026)

---

## Quick Links

### For Developers

- **New to project?** → [Quick Start Guide](guides/quick-start.md) (5-minute tutorial)
- **Authentication?** → [Authentication Flow](guides/authentication-flow.md) (JWT + RBAC)
- **Real workflows?** → [Common Use Cases](guides/common-use-cases.md) (7 scenarios)
- **Issues?** → [Troubleshooting Guide](guides/troubleshooting.md)
- **Writing tests?** → [Testing Quick Reference](guides/testing-quick-reference.md)
- **Adding aggregate?** → [Architecture Patterns](guides/architecture-patterns.md)

### For DevOps

- **Monitoring?** → [Health Checks](guides/health-checks.md)
- **Caching?** → [Caching Guide](guides/caching.md)

### For Architects

- **Understanding architecture?** → [Clean Architecture](concepts/clean-architecture.md)
- **Planning new context?** → [Bounded Contexts Strategy](concepts/bounded-contexts.md)
- **Event-driven design?** → [Event-Driven Architecture](concepts/event-driven.md)

---

##  Project Statistics

- **Code**: Go 1.24+, PostgreSQL 16, Redis 7
- **Tests**: 2477+ tests (2232+ unit, 196 smoke, 76+ integration), 90%+ coverage
- **Documentation**: 15,000+ lines across 55+ files
- **Contexts**: 6 production-ready (Shared, Identity, Customer-Mgmt, Order-Mgmt, Billing, Warehouse), 1 in progress (Fiscal)
- **Packages**: 12 reusable libraries (bus, jwt, logger, middleware, cache, uuidv7, valueobject, response, migration, saga, aggregate, jsonb)
- **Performance**: 377K events/sec (Memory Bus)
- **API Documentation**: 182+ endpoints (Swagger UI), 33K lines Postman collection

---

##  Contributing

Contributions are welcome! Please read our style guidelines:

- **Code Style**: Follow [Naming Conventions](guides/naming-conventions.md) and [Architecture Patterns](guides/architecture-patterns.md)
- **Documentation**: Follow [Documentation Style Guide](guides/documentation-style-guide.md) (no-emoji policy)
- **Testing**: Write tests per [Testing Patterns](guides/testing-patterns.md) (four-tier strategy)
- **Pull Requests**: Run `make pre-push` before submitting (see [Local CI Guide](guides/local-ci.md))

---

##  Work In Progress

**Active development** tasks and progress tracking:

- [Work In Progress](work-in-progress/README.md) - Current tasks and TODOs
- Living documents that change frequently
- Phase progress reports and optimizations

---

##  Recent Updates

**January 5, 2026**:
- Phase 1 COMPLETE: API Documentation & Developer Portal (5 days, 2x faster than planned)
- Swagger UI operational (182+ endpoints at `/api/docs/index.html`)
- Postman collection ready (33K lines, auto-generated with test automation)
- Developer Portal complete (4 guides: Quick Start, Auth Flow, Use Cases, Troubleshooting)
- Warehouse Context progress (55% complete - Inventory + StockMovement + Product aggregates production-ready)
- Product aggregate completed (139 tests, 16 endpoints)
- Subscription aggregate completed (8 endpoints, 120 tests)
- Billing context fully operational (Invoice, Payment, Subscription)
- Test infrastructure validated (2477+ tests: 325 warehouse, 196 smoke, 100% pass rate)
- All lint issues resolved (0 issues)

**January 1, 2026**:
- Local CI Validation implementation (5 Makefile commands)
- Comprehensive documentation review and updates
- Go 1.24+ sync across all files
- All 26 linting issues fixed (100% clean)

**December 31, 2025**:
- Interaction aggregate completed (14 endpoints, 74 tests)
- N+1 query optimization with LEFT JOIN
- Customer Management context fully operational

**December 29, 2025**:
- Health Checks implementation (21 tests)
- Documentation restructuring
- Website sync with docs/

**December 28, 2025**:
- Integration test optimization (78→24 tests, -69%)
- Go 1.22+ syntax modernization

---

**Version**: 0.1.0  
**Status**: Production-ready  
**License**: MIT  
**Maintainer**: Promenade Team
