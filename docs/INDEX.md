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

### Caching Layer

**Redis-based caching system** improves performance and reduces database load.

- **Adapters**: Redis (production), NoOp (testing)
- **TTL Strategy**: Reference (1h-24h), User (10-30m), Session (30m-1h)
- **Pattern**: Cache-aside with write-through invalidation
- **Features**: Graceful degradation, pattern-based deletion, JSON marshaling
- **Implementation**: [Caching Guide](CACHING.md)

**Cached**: Countries, Currencies, Languages, Timezones, Profiles, Customers

---

### Testing Strategy

**Three-tier testing** with clear separation and professional organization.

- **Unit Tests**: In-place, fast feedback (~5s)
- **Smoke Tests**: Mock-based handlers, no DB (~0.4s)
- **Integration Tests**: Real database, full E2E (~14s)
- **Implementation**: [Testing Patterns](guides/testing-patterns.md) | [Quick Reference](guides/testing-quick-reference.md)

**240+ tests**, 90%+ coverage | [Test Report](reference/test-coverage-report.md)

---

## 🏗️ Architecture

### Bounded Contexts

| Context                 | Status      | Aggregates                            | Documentation                                     |
| ----------------------- | ----------- | ------------------------------------- | ------------------------------------------------- |
| **Shared**              | ✅ Production | Country, Currency, Language, Timezone | [Context Guide](../internal/contexts/shared/README.md)      |
| **Identity**            | ✅ Production | User, Contact, Profile, Role, Permission | [Context Guide](../internal/contexts/identity/README.md)    |
| **Customer Management** | ✅ Production | Customer                              | [Context Guide](../internal/contexts/customer-mgmt/README.md) |
| **Order Management**    | 📋 Planned   | Order, OrderItem, Fulfillment         | Coming Q2 2026                                    |
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
| **uuidv7**      | Time-ordered UUIDs                 | 10    | [Guide](../pkg/uuidv7/README.md)           |
| **valueobject** | Domain value objects               | 45    | [Guide](../pkg/valueobject/README.md)      |
| **response**    | HTTP response helpers              | 13    | [Guide](../pkg/response/README.md)         |
| **migration**   | Database migrations                | 8     | [Guide](../pkg/migration/README.md)        |
| **saga**        | Distributed transactions           | 28    | [Guide](../pkg/saga/README.md)             |
| **aggregate**   | Base aggregate pattern             | 5     | [Guide](../pkg/aggregate/README.md)        |
| **jsonb**       | PostgreSQL JSONB utilities         | 8     | [Guide](../pkg/jsonb/README.md)            |

**Total**: 196+ tests across 10 packages | [Package Overview](../pkg/README.md)

---

## 📚 Documentation Sections

### Concepts

Fundamental architectural principles and design patterns:

- [Clean Architecture with DDD](concepts/clean-architecture.md) - Bounded Contexts, Aggregates, Value Objects
- [Event-Driven Architecture](concepts/event-driven.md) - Event Bus, Domain Events, Sagas
- [Bounded Contexts Strategy](concepts/bounded-contexts.md) - Context isolation and communication

### Guides

Step-by-step implementation guides:

- [Getting Started](guides/getting-started.md) - Quick start, installation, first steps
- [RBAC Implementation](guides/rbac.md) - Roles, permissions, JWT integration
- [Rate Limiting](guides/rate-limiting.md) - IP-based protection for authentication
- [Health Checks](guides/health-checks.md) - Dependency monitoring and alerting
- [Testing Patterns](guides/testing-patterns.md) - Three-tier testing strategy
- [Development Workflow](guides/development-workflow.md) - Daily development process
- [Production Deployment](guides/production-deployment.md) - Docker, Kubernetes, monitoring

### Reference

Technical specifications and detailed documentation:

- [API Reference](reference/api-reference.md) - Complete HTTP API documentation
- [Test Coverage Report](reference/test-coverage-report.md) - 240+ tests breakdown
- [Bus Test Coverage](reference/bus-test-coverage.md) - Event Bus test report
- [Index Audit Report](INDEX_AUDIT_REPORT.md) - Database index analysis (13 tables, 60+ indexes)
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

- **Code**: Go 1.23+, PostgreSQL 16, Redis 7
- **Tests**: 240+ tests, 90%+ average coverage
- **Documentation**: 15,000+ lines across 30+ files
- **Contexts**: 3 production-ready, 3 planned
- **Packages**: 10 reusable libraries
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

## �📝 Recent Updates

**December 29, 2025**:
- ✅ Health Checks implementation (21 tests)
- ✅ Documentation restructuring
- ✅ Website sync with docs/

**December 28, 2025**:
- ✅ Integration test optimization (78→24 tests, -69%)
- ✅ Go 1.22+ syntax modernization

**December 27, 2025**:
- ✅ Event Bus comprehensive testing (67 tests)

---

**Version**: 0.1.0  
**Status**: Production-ready  
**License**: MIT  
**Maintainer**: Promenade Team
