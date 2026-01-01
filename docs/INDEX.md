# Promenade Platform - Documentation

**Modern backend platform** for customer management, orders, and business workflows built with **Domain-Driven Design**, **Event-Driven Architecture**, and **Clean Code principles**.

---

## 🎯 Core Concepts

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

**27 tests** (JWT) + **10 tests** (Rate Limiting)

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
- **JSONB Attendees**: Flexible participant tracking with PostgreSQL JSONB arrays
- **Follow-up Management**: Flag interactions requiring follow-up with dates and notes
- **Duration Tracking**: Automatic duration calculation for ended interactions
- **Performance Optimized**: LEFT JOIN queries prevent N+1 problem when listing interactions
- **14 API Endpoints**: Complete CRUD + business logic operations

**Key Features**:
- Flexible interaction types and directions
- Outcome tracking (successful, failed, no_answer, scheduled, cancelled)
- Multi-participant support via JSONB attendees array
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

**Three-tier testing** with clear separation and professional organization.

- **Unit Tests**: In-place, fast feedback (~5s)
- **Smoke Tests**: Mock-based handlers, no DB (~0.4s)
- **Integration Tests**: Real database, full E2E (~14s)
- **Implementation**: [Testing Patterns](guides/testing-patterns.md) | [Quick Reference](guides/testing-quick-reference.md)

**240+ tests**, 90%+ coverage | [Test Report](reference/test-coverage-report.md)

---

### Order Management

**Complete order lifecycle management** with **fully implemented business rules**.

- **Order Creation**: Create orders with currency support
- **Line Items**: Add/remove/update products with automatic totals
- **State Machine**: pending → confirmed → processing → fulfilled - **fully enforced** ✅
- **Business Rules**: All validations implemented in entity methods ✅
- **Money Handling**: Type-safe Money value object (cents-based)
- **14 Endpoints**: Complete CRUD + business logic
- **Implementation**: [Order Management Guide](concepts/order-management.md)

**Status**: All core business logic implemented | Future: Inventory/Payment/Shipping integrations

---

## 🏗️ Architecture

### Bounded Contexts

| Context                 | Status      | Aggregates                            | Documentation                                     |
| ----------------------- | ----------- | ------------------------------------- | ------------------------------------------------- |
| **Shared**              | ✅ Production | Country, Currency, Language, Timezone | [Context Guide](../internal/contexts/shared/README.md)      |
| **Identity**            | ✅ Production | User, Contact, Profile, Role, Permission | [Context Guide](../internal/contexts/identity/README.md)    |
| **Customer Management** | ✅ Production | Customer ✅, Company ✅, Deal ✅, Interaction ✅, Analytics ✅ | [Customer Guide](concepts/customer-management.md) \| [Analytics](../internal/contexts/customer-mgmt/analytics/README.md) |
| **Order Management**    | ✅ Production | Order ✅, OrderLine ✅, Contract 📋, Fulfillment 📋 | [Order Guide](concepts/order-management.md)                   |
| **Billing**             | 📋 Planned   | Invoice, Payment, Subscription        | Coming Q2 2026                                    |
| **Warehouse**           | 📋 Planned   | Inventory, Stock                      | Coming Q3 2026                                    |

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
| **jsonb**       | PostgreSQL JSONB utilities         | 8     | [Guide](../pkg/jsonb/README.md)            |

**Total**: 250+ tests across 12 packages | [Package Overview](../pkg/README.md)

---

## 📚 Documentation Sections

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

### Guides

Step-by-step implementation guides:

- [Getting Started](guides/getting-started.md) - Quick start, installation, first steps
- [Local CI Validation](guides/local-ci.md) - Run GitHub Actions checks locally before push
- [RBAC Implementation](guides/rbac.md) - Roles, permissions, JWT integration
- [Rate Limiting](guides/rate-limiting.md) - IP-based protection for authentication
- [CSRF Protection](guides/csrf-protection.md) - Cross-site request forgery defense
- [Health Checks](guides/health-checks.md) - Dependency monitoring and alerting
- [Caching](guides/caching.md) - Redis caching layer implementation
- [Testing Patterns](guides/testing-patterns.md) - Four-tier testing strategy
- [Development Workflow](guides/development-workflow.md) - Daily development process
- [Production Deployment](guides/production-deployment.md) - Docker, Kubernetes, monitoring

**Design Patterns & Conventions**:
- [Naming Conventions](guides/naming-conventions.md) - Files, directories, Go code naming standards
- [Database Conventions](guides/database-conventions.md) - Tables, columns, indexes, migrations
- [Architecture Patterns](guides/architecture-patterns.md) - Repository, UseCase, Handler, Value Object patterns

### Reference

Technical specifications and detailed documentation:

- [API Reference](reference/api-reference.md) - Complete HTTP API documentation
- [Test Coverage Report](reference/test-coverage-report.md) - 240+ tests breakdown
- [Bus Test Coverage](reference/bus-test-coverage.md) - Event Bus test report
- [Index Audit Report](work-in-progress/INDEX_AUDIT_REPORT.md) - Database index analysis (13 tables, 60+ indexes)
- [Migration History](reference/migration-history.md) - Database schema evolution
- [Configuration Reference](reference/configuration.md) - YAML config options

---

## 🚀 Quick Links

### For Developers

- **New to project?** → [Getting Started](guides/getting-started.md)
- **Writing tests?** → [Testing Quick Reference](guides/testing-quick-reference.md)
- **Adding aggregate?** → [Development Workflow](guides/development-workflow.md)
- **Need API docs?** → [API Reference](reference/api-reference.md)

### For DevOps

- **Deploying?** → [Production Deployment](guides/production-deployment.md)
- **Monitoring?** → [Health Checks](guides/health-checks.md)
- **Configuring?** → [Configuration Reference](reference/configuration.md)

### For Architects

- **Understanding architecture?** → [Clean Architecture](concepts/clean-architecture.md)
- **Planning new context?** → [Bounded Contexts Strategy](concepts/bounded-contexts.md)
- **Event-driven design?** → [Event-Driven Architecture](concepts/event-driven.md)

---

## 📊 Project Statistics

- **Code**: Go 1.24+, PostgreSQL 16, Redis 7
- **Tests**: 240+ tests, 90%+ average coverage
- **Documentation**: 15,000+ lines across 30+ files
- **Contexts**: 4 production-ready (Shared, Identity, Customer-Mgmt, Order-Mgmt)
- **Packages**: 12 reusable libraries (bus, jwt, logger, middleware, cache, uuidv7, valueobject, response, migration, saga, aggregate, jsonb)
- **Performance**: 377K events/sec (Memory Bus)

---

## 🤝 Contributing

Read our [Contributing Guide](guides/contributing.md) to learn about:

- Code style and conventions
- Pull request process
- Testing requirements
- Documentation standards

---

## � Work In Progress

**Active development** tasks and progress tracking:

- [Work In Progress](work-in-progress/README.md) - Current tasks and TODOs
- Living documents that change frequently
- Phase progress reports and optimizations

---

## 📝 Recent Updates

**January 1, 2026**:
- ✅ Local CI Validation implementation (5 Makefile commands)
- ✅ Comprehensive documentation review and updates
- ✅ Go 1.24+ sync across all files
- ✅ All 26 linting issues fixed (100% clean)

**December 31, 2025**:
- ✅ Interaction aggregate completed (14 endpoints, 74 tests)
- ✅ N+1 query optimization with LEFT JOIN
- ✅ Customer Management context fully operational

**December 29, 2025**:
- ✅ Health Checks implementation (21 tests)
- ✅ Documentation restructuring
- ✅ Website sync with docs/

**December 28, 2025**:
- ✅ Integration test optimization (78→24 tests, -69%)
- ✅ Go 1.22+ syntax modernization

---

**Version**: 0.1.0  
**Status**: Production-ready  
**License**: MIT  
**Maintainer**: Promenade Team
