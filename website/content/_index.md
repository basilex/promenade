---
title: "Production-Ready REST API Framework"
description: "Enterprise-grade backend framework built with Clean Architecture principles. Features a revolutionary modular plugin system where business logic lives in independent, licensable modules. No shared domain entities, no tight coupling - true architectural autonomy."
date: 2025-12-25
features:
  - icon: "🏗️"
    title: "Clean Architecture"
    description: "Strict layered architecture with Domain, Use Case, Adapter, and Infrastructure layers. Dependency rule enforced."

  - icon: "🧩"
    title: "Modular System"
    description: "Independent business modules with own entities, repositories, and use cases. Enable/disable modules dynamically."

  - icon: "🗄️"
    title: "PostgreSQL + UUID v7"
    description: "Time-ordered UUIDs for 2x faster inserts. Raw SQL with sqlx - no ORM overhead. BaseRepository pattern."

  - icon: "🔐"
    title: "Auth & RBAC"
    description: "JWT authentication, role-based access control with wildcard permissions. Session management included."

  - icon: "📡"
    title: "Event-Driven"
    description: "Dual event bus adapters (Memory/Redis). Async communication between modules. Production-ready."

  - icon: "🔄"
    title: "Namespace Migrations"
    description: "Each module has independent migration history. True module autonomy without schema conflicts."

  - icon: "✅"
    title: "400+ Tests"
    description: "Comprehensive test suite with unit, integration, and smoke tests. Manual mocks pattern for testability."

  - icon: "📚"
    title: "20+ Docs"
    description: "Extensive documentation in EN/UK/DE. Architecture guides, tutorials, and API references."

  - icon: "🚀"
    title: "Production-Ready"
    description: "Used in real-world applications. Commercial modules available (billing, audit, warehouse)."
---

## Quick Start

```bash
# Clone repository
git clone https://github.com/basilex/promenade.git
cd promenade

# Start PostgreSQL + Redis
make docker-up

# Run migrations
make migrate

# Start development server
make dev
```

Server starts on **http://localhost:8081**

---

## The Promenade Philosophy

**Promenade isn't just another Go framework** - it's a complete rethinking of how backend applications should be structured.

Traditional monoliths suffer from tight coupling. Microservices introduce operational complexity. **Promenade offers a third way**: a modular monolith where business domains are truly independent, yet run in a single process.

### Core Principles

🎯 **Core as Orchestrator, Modules as Workers**

The Core layer provides infrastructure (database, auth, events, logging) but contains **zero business logic**. All business domains live in modules that can be enabled/disabled without affecting each other.

🔌 **True Module Independence**

Modules never import from `internal/domain` or other modules. Each module is a complete vertical slice with its own entities, use cases, and adapters. This isn't just good practice - it's enforced by the architecture.

📊 **Namespace-Based Migrations**

Each module maintains its own migration history in separate namespaces (`migrations/core/`, `migrations/posts/`, `migrations/billing/`). Enable or disable modules without breaking database schema history. No more migration conflicts in team development.

⚡ **Performance by Design**

UUID v7 primary keys provide time-ordered inserts (2x faster than UUID v4). Raw SQL with sqlx eliminates ORM overhead. Connection pooling and partial indexes optimize query performance. Benchmarked in production.

🔐 **Security First**

Built-in JWT authentication, RBAC with wildcard permissions (`posts:*`, `users:delete`), session management, and audit logging. Commercial audit module provides immutable logs with HMAC signatures.

## Core Features

### Clean Architecture

Promenade implements **textbook Clean Architecture** with strict dependency rules:

- **Domain Layer**: Business entities and interfaces - pure business logic, zero framework dependencies
- **Use Case Layer**: Business logic implementation - orchestrates entities and repositories
- **Adapter Layer**: HTTP handlers, repositories - framework integration points
- **Infrastructure**: Database, event bus, configuration - external services

**The Dependency Rule**: Inner layers never depend on outer layers. Use cases depend on domain interfaces, not concrete implementations. This makes your business logic **portable, testable, and maintainable**.

### Module System

The module system is Promenade's killer feature:

#### Core (Always Enabled)

- **Authentication & Authorization**: JWT tokens, RBAC with 4 system roles, session management
- **Reference Data**: 145 countries, 124 currencies, 30 regions, 17 cities, 40+ payment methods, timezones, languages
- **Event Bus**: Memory adapter (dev/test) or Redis adapter (production)
- **Database Management**: Connection pooling, transaction support, migration runner
- **Logging & Config**: Structured logging with slog, YAML configuration per environment

#### Available Modules

- **Posts** (Free): User-generated content with comments, likes, soft delete, and purge policies
- **Profiles** (Free): User profiles with contacts, privacy settings, verification
- **Analytics** (Free): Metrics collection, aggregation, custom reports, dashboards
- **Billing** (Commercial): Subscription management, invoices, payments, trials - production-ready
- **Audit** (Commercial): Immutable audit logs with HMAC signatures for compliance
- **Warehouse** (Planned): Inventory management, product catalog, stock tracking

#### How Modules Work

Each module:

- ✅ Registers itself via `init()` - no manual wiring required
- ✅ Loads own configuration from `config/config.{env}.yaml`
- ✅ Manages own database migrations in separate namespace
- ✅ Defines own RBAC permissions (auto-inserted on startup)
- ✅ Publishes/subscribes to domain events for async communication
- ✅ Can run background workers (cron jobs, queue processors)
- ✅ Provides health checks and graceful shutdown

**Result**: You can build a SaaS platform by simply enabling modules in `config/modules.yaml`. No code changes needed.

### Database

#### PostgreSQL 16 with Modern Patterns

- **UUID v7 Primary Keys**: Time-ordered UUIDs provide 2x faster inserts compared to UUID v4 due to better B-tree locality. Extract timestamps directly from IDs.
- **Raw SQL with sqlx**: No ORM magic. Write SQL, get performance and control. BaseRepository pattern provides common operations (Get, Select, Exec) while allowing custom queries.
- **Soft Delete Pattern**: User content (posts, comments) uses `deleted_at` timestamp. Always filter `WHERE deleted_at IS NULL`. Automated purge system removes old soft-deleted records based on retention policies.
- **Namespace-Based Migrations**: Core migrations run first (`000001_core_*.sql`), then enabled module migrations. Each namespace has independent version tracking in `schema_migrations` table.

#### Reference Data

Promenade includes production-ready reference data:

- **145 countries** with ISO 3166-1 codes (alpha2, alpha3), regions, and geographic data
- **124 currencies** with ISO 4217 codes, symbols, and decimal places
- **30 administrative regions** (states, oblasts, provinces, Länder) with type classification
- **17 major cities** with coordinates, population, capital status, and slugs
- **40+ payment methods** including cards, digital wallets, crypto, BNPL, and local methods
- **IANA timezone database** for accurate time handling worldwide
- **ISO 639 language codes** for multilingual applications

No need to build this yourself - it's already there, properly normalized, and ready to use.

### Testing

**Testing is a first-class citizen in Promenade.** The architecture makes testing natural:

- **275 core tests**: Domain entities (Country, Currency, User, Session, Permission, Role) + use cases (Auth, RBAC, reference data CRUD)
- **125+ module tests**:
  - Posts: 33 tests (83.3% coverage) - PostStatus lifecycle, validation, slug generation
  - Profiles: 21 tests (80.4% coverage) - UserContact, UserProfile, privacy logic
  - Analytics: 11 tests - Metrics validation, aggregation logic
  - Billing: 375 tests (100% coverage) - Complete subscription lifecycle
- **Manual mocks pattern**: No code generation, no magic. Simple inline struct mocks with method implementations.
- **Test execution**: **~20 seconds** for 400+ tests across all layers

**Commands**:

```bash
make test              # Run all tests
make test-core         # Core tests only (275 tests)
make test-modules      # All module tests
make test-coverage     # HTML coverage report
```

The key insight: Clean Architecture makes business logic **independent of frameworks**, which makes it **trivially testable**. No need for complex mocking libraries - just implement interfaces.

## Architecture Highlights

### How Core and Modules Interact

```
Core (Orchestrator)          Modules (Workers)
├── Auth & RBAC              ├── Posts Module
│   └── JWT Manager          │   ├── Entity: UserPost, Comment
│   └── Permission Check     │   ├── UseCase: CreatePost, AddComment
├── Event Bus                │   ├── Repository: PostgreSQL impl
│   └── Memory/Redis         │   └── Handler: HTTP endpoints
├── Database                 │
│   └── Connection Pool      ├── Profiles Module
│   └── Transaction Mgr      │   ├── Entity: UserProfile, Contact
├── Logger                   │   ├── UseCase: UpdateProfile
│   └── Structured slog      │   └── Handler: HTTP endpoints
├── Config                   │
│   └── YAML per env         ├── Analytics Module (Free)
└── Reference Data           │   ├── Entity: Metric, Aggregate
    ├── Countries (145)      │   ├── UseCase: RecordMetric
    ├── Currencies (124)     │   └── Handler: Dashboard API
    ├── Regions (30)         │
    ├── Cities (17)          └── Billing Module (Commercial)
    ├── Payment Methods (40+)    ├── Entity: Plan, Subscription
    ├── Timezones                ├── UseCase: CreateInvoice
    └── Languages                └── Handler: Payment webhook
```

### The Registry Pattern

Modules register themselves with Core via **three registries**:

1. **Module Registry**: Module discovery and lifecycle management
2. **Purge Handler Registry**: Each module registers handlers for soft-deleted entities
3. **Policy Registry**: Each module defines retention policies (e.g., "delete posts after 90 days")

**Core orchestrates purge jobs** by:

- Running on schedule (cron)
- Collecting policies from registry
- Calling registered handlers
- Publishing purge events

**Core has zero knowledge** of what "user_posts" or "post_comments" are. This is true decoupling.

**Key Principle**: Core orchestrates, modules execute. Core knows WHEN to call modules, not HOW they work.

---

---

## Why Choose Promenade?

### For Startups

✅ **Fast Time to Market**: Core infrastructure ready, just add business logic
✅ **Cost Effective**: Free modules for MVP, commercial modules when you scale
✅ **Single Deployment**: No microservices complexity until you need it
✅ **Proven Patterns**: Clean Architecture = easier to hire developers

### Compared to Traditional Monoliths

Traditional monoliths suffer from:

- ❌ Tight coupling between business domains
- ❌ Shared domain entities leading to god objects
- ❌ Single migration stream causing team conflicts
- ❌ Can't deploy/license features independently

**Promenade solves this** with true module independence while keeping operational simplicity of a single process.

### Compared to Microservices

Microservices introduce:

- ❌ Network latency and complexity
- ❌ Distributed transaction challenges
- ❌ Service discovery and orchestration overhead
- ❌ Debugging and testing across services is hard

**Promenade gives you** modular architecture benefits without the operational burden. Need to go distributed later? Modules are already isolated.

### Compared to Other Go Frameworks

Most Go frameworks are either:

- 🔧 Too minimal (Gin, Echo) - no architecture guidance
- 🏗️ Too opinionated (Beego, Buffalo) - locked into patterns
- 📦 Too complex (Go-Kit) - microservice-focused

**Promenade provides**:

- ✅ Clear architectural guidelines (Clean Architecture)
- ✅ Freedom to choose tools (any HTTP router, any DB driver)
- ✅ Plugin system for business logic (modules)
- ✅ Production-ready infrastructure out of the box

### The Module Advantage

**Scenario**: You're building a SaaS platform. You need:

- User management → **Core (built-in)**
- Blog/posts → **Posts module (free)**
- User profiles → **Profiles module (free)**
- Analytics dashboard → **Analytics module (free)**
- Subscription billing → **Billing module (commercial license)**
- Audit logging → **Audit module (commercial license)**

**Traditional approach**: Build everything from scratch, or integrate multiple third-party libraries with incompatible patterns.

**Promenade approach**:

```yaml
# config/modules.yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics
    - billing
    - audit
```

Run migrations, set license keys, done. All modules follow the same patterns, use the same infrastructure, integrate seamlessly.

---

## Real-World Use Cases

### SaaS Platforms

Enable billing module for subscriptions, analytics for metrics, audit for compliance. Core RBAC handles user permissions.

### Content Management

Posts module provides blog/articles, profiles for authors, analytics for engagement metrics. All with soft delete and purge policies.

### E-commerce

Warehouse module (planned) for inventory, billing for orders, analytics for sales tracking. Payment methods reference data included.

### Enterprise Applications

Audit module for compliance logging, RBAC for complex permissions, event bus for async workflows. Multi-language support built-in.

---

## Technology Stack

**Backend**:

- Go 1.25+ (modern, performant, compiled)
- PostgreSQL 16 (ACID, JSON, full-text search)
- Redis 7 (optional, for distributed event bus)

**Architecture**:

- Clean Architecture (Uncle Bob)
- Domain-Driven Design principles
- Event-Driven Architecture
- CQRS patterns where beneficial

**Developer Experience**:

- Makefile for common tasks (`make dev`, `make test`, `make migrate`)
- Hot reload in development
- Swagger/OpenAPI docs auto-generated
- 20+ documentation guides in 3 languages

**Production Ready**:

- Docker/Docker Compose setup
- GitHub Actions CI/CD
- Health checks and metrics
- Graceful shutdown
- Connection pooling
- Rate limiting support

---

✅ **Proven Patterns**: Clean Architecture = easier to hire developers

### For Enterprise

✅ **Maintainable**: Module independence prevents big ball of mud
✅ **Testable**: 400+ tests show how to test your own code
✅ **Auditable**: Commercial audit module for compliance (SOC 2, GDPR)
✅ **Scalable**: Start simple, scale when needed (modules → microservices)

### For Teams

✅ **No Conflicts**: Namespace migrations = parallel development
✅ **Clear Boundaries**: Module boundaries = team boundaries
✅ **Self-Documenting**: Architecture enforces consistency
✅ **Multilingual**: Docs in EN/UK/DE, more languages coming

### For Solo Developers

✅ **Learn Best Practices**: Real-world Clean Architecture example
✅ **Avoid Boilerplate**: Auth, RBAC, migrations already done
✅ **Focus on Business**: Infrastructure handled, you write features
✅ **Copy Patterns**: Every module shows the same structure

---

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Docker & Docker Compose (for PostgreSQL + Redis)
- Make (for build automation)

### Installation

```bash
# 1. Clone repository
git clone https://github.com/basilex/promenade.git
cd promenade

# 2. Install dependencies
make install

# 3. Start infrastructure
make docker-up

# 4. Run migrations (auto-runs on startup, but can run manually)
make migrate

# 5. Start development server (with hot reload)
make dev
```

Server starts on **http://localhost:8081**

### Verify Installation

```bash
# Health check
curl http://localhost:8081/api/v1/health

# Swagger UI (API documentation)
open http://localhost:8081/api/v1/docs/swagger/index.html
```

### Next Steps

1. **Read the docs**: [Documentation](/promenade/docs/) covers architecture, modules, database patterns
2. **Try the API**: Use Swagger UI to test endpoints with default users
3. **Build a module**: Follow [Module Development Guide](/promenade/docs/module-development/)
4. **Run tests**: `make test` to see comprehensive test suite
5. **Deploy**: Docker setup ready for production deployment

---

## Documentation

📖 **20+ Comprehensive Guides**:

- [Architecture Overview](/promenade/docs/architecture/) - System design and principles
- [Database Schema](/promenade/docs/database-schema/) - Complete schema with ER diagrams
- [Module Development](/promenade/docs/module-development/) - Build custom modules
- [Getting Started](/promenade/docs/getting-started/) - Quick start tutorial

🌐 **Multilingual**:

- 🇬🇧 English (complete)
- 🇺🇦 Ukrainian (повна)
- 🇩🇪 German (vollständig)

📝 **API Documentation**:

- [Swagger v1](/api/v1/docs/swagger/index.html) - Current stable API
- [Swagger v2](/api/v2/docs/swagger/index.html) - Next generation API

---

## Community & Support

🐛 **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
💬 **Discussions**: [GitHub Discussions](https://github.com/basilex/promenade/discussions)
📧 **Email**: alexander.vasilenko@gmail.com
⭐ **Star on GitHub**: [github.com/basilex/promenade](https://github.com/basilex/promenade)

---

## License

**MIT License** - Free for personal and commercial use.

Commercial modules (billing, audit, warehouse) require separate license keys.
[Generate license](https://github.com/basilex/promenade#commercial-modules) or contact for enterprise licensing.

---

**Built with ❤️ using Clean Architecture and Go**

Ready to build production-grade APIs? [Get Started](/promenade/docs/getting-started/) or [Explore Features](/promenade/features/)
