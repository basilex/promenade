# Promenade

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Tests](https://img.shields.io/badge/Tests-passing-success?style=flat)](test/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![DDD](https://img.shields.io/badge/Architecture-DDD-green.svg)](docs/concepts/clean-architecture.md)

**Modern Enterprise Resource Planning (ERP) system built with Domain-Driven Design**

Replace multiple SaaS tools with one modular platform: CRM, Orders, Inventory, Billing, Accounting, and Fiscal compliance — designed for Ukrainian and European SMB market.

---

## 📊 Business Capabilities

### Customer Relationship Management (CRM)

Manage your entire sales pipeline from lead to customer, track interactions, and grow your business.

- **Customer Management** — Companies, contacts, segmentation, lifecycle tracking
- **Deal Pipeline** — Sales opportunities, stages, probability scoring, forecasting
- **Interaction History** — Calls, meetings, emails, notes with full audit trail
- **Sales Analytics** — Revenue reports, conversion metrics, team performance

**🎯 For**: Sales teams, account managers, customer success  
**📚 Documentation**: [Customer Management Context](internal/contexts/customer-mgmt/README.md)

### Order Management & Fulfillment

Process orders efficiently with automated workflows, real-time tracking, and fulfillment orchestration.

- **Order Processing** — Multi-channel orders, complex pricing, promotions
- **Fulfillment Workflows** — Automated state machines: pending → confirmed → processing → fulfilled
- **Saga Orchestration** — Coordinate inventory, billing, and fiscal operations
- **Order Analytics** — Sales metrics, fulfillment KPIs, bottleneck detection

**🎯 For**: E-commerce, B2B distributors, service businesses  
**📚 Documentation**: [Order Management Context](internal/contexts/order-mgmt/README.md)

### Warehouse & Inventory

Track inventory across multiple locations with real-time visibility and automated stock management.

- **Multi-Warehouse** — Unlimited locations with hierarchical organization
- **Product Catalog** — SKU management, categories, variants, pricing
- **Stock Movements** — Receipts, reservations, transfers, adjustments with full audit
- **Low Stock Alerts** — Automated reorder points, stock forecasting

**🎯 For**: Retailers, distributors, manufacturers  
**📚 Documentation**: [Warehouse Context](internal/contexts/warehouse/README.md)

### Billing & Invoicing

Generate invoices, track payments, and manage recurring billing with automatic subscription handling.

- **Invoice Generation** — Automatic from orders, manual creation, templates
- **Payment Tracking** — Multiple payment methods, partial payments, refunds
- **Recurring Billing** — Subscriptions, trials, plan upgrades, automatic renewals
- **Payment Analytics** — Cash flow, aging reports, revenue recognition

**🎯 For**: SaaS companies, subscription businesses, B2B services  
**📚 Documentation**: [Billing Context](internal/contexts/billing/README.md)

### Accounting & Finance

Complete financial management with chart of accounts, journal entries, and financial reporting.

- **Chart of Accounts** — Flexible account hierarchy, multi-currency support
- **Journal Entries** — Double-entry bookkeeping, automatic posting, audit trail
- **Cost Centers & Budgets** — Department tracking, budget planning, variance analysis
- **Financial Reports** — Balance sheet, P&L, cash flow, custom reports

**🎯 For**: Finance teams, accountants, CFOs  
**📚 Documentation**: [Accounting Context](internal/contexts/accounting/README.md)

### Banking Integration

Connect to bank accounts, import transactions, and automate reconciliation.

- **Bank Accounts** — Multiple accounts per organization, multi-currency
- **Transaction Import** — Bank statement parsing, automatic categorization
- **Reconciliation** — Match transactions with invoices and payments
- **Cash Management** — Cash position, forecast, treasury operations

**🎯 For**: Finance teams, treasury management  
**📚 Documentation**: [Banking Context](internal/contexts/banking/README.md)

### Fiscal Compliance (Ukraine/EU)

Built-in integration with Ukrainian tax authorities and ПРРО (fiscal cash registers).

- **Checkbox Integration** — Automatic receipt generation, fiscal reporting
- **ПРРО Support** — Cash register operations, Z-reports, fiscal documents
- **Tax Compliance** — VAT tracking, tax reports, electronic submission
- **Audit Trail** — Complete transaction history for tax audits

**🎯 For**: Ukrainian businesses, tax compliance officers  
**📚 Documentation**: [Fiscal Context](internal/contexts/fiscal/README.md)

### Identity & Access Management

Enterprise-grade user management with role-based access control.

- **User Management** — Users, contacts, profiles, multi-organization support
- **Authentication** — JWT-based, secure password hashing (Argon2id), MFA-ready
- **Authorization** — Role-based access control (RBAC), permissions, API keys
- **Audit Logging** — User actions, security events, compliance tracking

**🎯 For**: IT administrators, security teams  
**📚 Documentation**: [Identity Context](internal/contexts/identity/README.md)

---

## 🎯 Who Is This For?

### 🏢 Small & Medium Businesses (SMB)

Replace expensive ERP systems (SAP, 1C, Bitrix24) with flexible, cost-effective solution:

- **E-commerce stores** — Integrate with online shops, manage inventory, automate fulfillment
- **B2B distributors** — Handle complex pricing, track customer relationships, manage warehouses
- **Service businesses** — Manage projects, track time, invoice clients, recurring billing

**Business Value**: [Market Analysis & ROI](docs/business/BUSINESS_OVERVIEW.md)

### 💻 Software Development Companies

Build custom business solutions on proven foundation:

- **SaaS platforms** — Use as backend for vertical SaaS products
- **Custom ERP** — Extend with industry-specific modules
- **Integration projects** — Connect existing systems via event-driven architecture

**Technical Details**: [Architecture Overview](docs/concepts/clean-architecture.md)

### 🇺🇦 Ukrainian Market Focus

Purpose-built for Ukrainian business requirements:

- **Fiscal compliance** — ПРРО, Checkbox, tax reporting
- **Local banking** — Monobank, PrivatBank integration (roadmap)
- **Language support** — Ukrainian, Russian, English interfaces
- **Market opportunity**: $45M+ TAM, 180K+ potential customers

**Market Strategy**: [Ukraine Market Analysis](docs/roadmap/UKRAINE_MARKET_STRATEGY_2026.md)

---

## 🚀 Quick Start

Get Promenade running in 5 minutes:

### Prerequisites

- **Go 1.24+** — [Install Go](https://go.dev/doc/install)
- **Docker** — [Install Docker](https://docs.docker.com/get-docker/)
- **Make** — Usually pre-installed on macOS/Linux

### 3-Command Start

```bash
# 1. Clone repository
git clone https://github.com/basilex/promenade.git
cd promenade

# 2. Start PostgreSQL and run migrations
make docker-up
make migrate

# 3. Start API server
make dev
```

**🎉 That's it!** Server is running at `http://localhost:8081`

### Verify Installation

```bash
# Health check
curl http://localhost:8081/health

# API documentation (Swagger UI)
open http://localhost:8081/api/docs/index.html
```

### What's Next?

- **[Complete Setup Guide](docs/guides/quick-start.md)** — User registration, authentication, first API calls
- **[API Documentation](docs/guides/api-documentation.md)** — Swagger UI guide, Postman collection
- **[Development Workflow](docs/guides/development-workflow.md)** — Daily commands, testing, debugging
- **[Postman Examples](postman/examples/)** — Ready-to-use API call examples

---

## 🏗️ Architecture

### Domain-Driven Design Foundation

Promenade is built on **strict DDD principles** with 8 bounded contexts that communicate only via events:

```
┌─────────────────────────────────────────────────────────────┐
│                        Event Bus                            │
│            (Async communication layer)                      │
└─────────────────────────────────────────────────────────────┘
     │        │        │        │        │        │        │
     ▼        ▼        ▼        ▼        ▼        ▼        ▼
┌─────────┬─────────┬─────────┬─────────┬─────────┬─────────┬─────────┐
│Identity │Customer │  Order  │ Billing │Warehouse│ Fiscal  │Accounting│
│         │  Mgmt   │  Mgmt   │         │         │         │  & Bank  │
└─────────┴─────────┴─────────┴─────────┴─────────┴─────────┴─────────┘
   Own DB    Own DB    Own DB    Own DB    Own DB    Own DB    Own DB
```

### Core Principles

1. **Bounded Context Autonomy**
   - Each context owns its domain model, database schema, and business logic
   - No shared database tables between contexts
   - Independent deployment and scaling

2. **No Cross-Context Imports**
   - Contexts communicate only through Event Bus
   - No direct Go imports between contexts
   - Uses Anti-Corruption Layer (ACL) for integration

3. **Clean Architecture Layers**
   - **Domain Layer** — Aggregates, entities, value objects, business rules
   - **Application Layer** — Use cases, orchestration, transaction boundaries
   - **Infrastructure Layer** — Database, HTTP, event handlers
   - **Presentation Layer** — DTOs, HTTP handlers, API routes

4. **Event-Driven Integration**
   - Domain events published after state changes
   - Asynchronous processing with guaranteed delivery
   - Saga pattern for distributed transactions

**📚 Learn More**:

- [Bounded Contexts Deep Dive](docs/concepts/bounded-contexts.md)
- [Clean Architecture Explained](docs/concepts/clean-architecture.md)
- [Event-Driven Patterns](docs/concepts/event-driven.md)
- [Fulfillment Saga Pattern](docs/concepts/fulfillment-saga.md)

### Technology Stack

**Backend Core**:

- **Language**: Go 1.24 — Performance, concurrency, type safety
- **Database**: PostgreSQL 16+ — ACID compliance, JSON support
- **HTTP**: Gin framework — High performance, middleware ecosystem
- **Auth**: JWT (RS256) — Stateless, secure, scalable

**Infrastructure**:

- **Event Bus**: In-memory (dev), Redis (production) — 377K events/sec
- **Cache**: Redis — Session storage, rate limiting
- **Migrations**: Custom manager — Multi-namespace, rollback support
- **Monitoring**: Health checks, metrics, structured logging

**Development**:

- **Testing**: 4-tier strategy — Unit, smoke, integration, benchmarks
- **API Docs**: Swagger/OpenAPI 3.0 — Interactive documentation
- **Git Hooks**: Pre-commit linting, pre-push testing
- **CI/CD**: GitHub Actions — Automated testing, deployment

**📚 Learn More**:

- [Technology Decisions (ADRs)](docs/adr/)
- [Database Strategy](docs/guides/database-strategy.md)
- [Event Bus Performance](pkg/bus/README.md)

---

## 📂 Project Structure

```
promenade/
├── cmd/                          # Application entry points
│   ├── api/                      # HTTP API server
│   │   ├── main.go              # Entry point
│   │   ├── bootstrap.go         # Infrastructure setup
│   │   ├── dependencies.go      # Custom DI container
│   │   └── server.go            # HTTP routing
│   ├── migrate/                 # Migration runner
│   └── seed/                    # Test data seeder
│
├── internal/                     # Private application code
│   ├── contexts/                # Bounded contexts (DDD)
│   │   ├── identity/           # Users, auth, permissions
│   │   ├── customer-mgmt/      # CRM, deals, interactions
│   │   ├── order-mgmt/         # Orders, fulfillment
│   │   ├── billing/            # Invoices, payments
│   │   ├── warehouse/          # Inventory, products
│   │   ├── fiscal/             # Cash registers, receipts
│   │   ├── accounting/         # Chart of accounts, journals
│   │   └── banking/            # Bank accounts, reconciliation
│   └── infrastructure/         # Cross-cutting concerns
│       ├── database/           # Connection, transactions
│       ├── auth/               # JWT, password hashing
│       └── http/               # Middleware, responses
│
├── pkg/                         # Reusable packages
│   ├── bus/                    # Event Bus (377K events/sec)
│   ├── aggregate/              # Base aggregate, change tracking
│   ├── jsonstore/              # Type-safe JSON fields
│   ├── uuidv7/                 # Time-ordered UUIDs
│   └── middleware/             # Auth, CORS, rate limiting
│
├── migrations/                  # Database migrations
│   └── postgres/               # PostgreSQL migrations
│       ├── core/               # System tables
│       ├── identity/           # Identity context
│       ├── customer-mgmt/      # Customer management
│       └── .../                # Other contexts
│
├── test/                        # Test suites
│   ├── integration/            # Full workflow tests
│   ├── smoke/                  # Fast CRUD checks
│   └── benchmark/              # Performance tests
│
├── docs/                        # Documentation
│   ├── business/               # Business overviews (8 languages)
│   ├── concepts/               # Architecture concepts
│   ├── guides/                 # Development guides
│   ├── deployment/             # Kubernetes, AWS, Docker
│   └── adr/                    # Architecture Decision Records
│
└── config/                      # Configuration files
    ├── app.postgres-dev.yaml   # PostgreSQL development
    ├── app.postgres-prod.yaml  # PostgreSQL production
    └── .../                     # Other configs
```

**Each Bounded Context** follows clean architecture:

```
internal/contexts/{context}/
├── {aggregate}/              # Domain subdomain
│   ├── aggregate/           # Domain entities & business logic
│   ├── repository/          # Repository interface
│   ├── usecase/             # Application use cases
│   ├── adapter/
│   │   ├── repository/
│   │   │   └── postgres/    # PostgreSQL implementation
│   │   └── http/            # HTTP handlers
│   ├── dto/                 # Data Transfer Objects
│   └── errors.go            # Domain error constants
└── README.md                 # Context documentation
```

**📚 Learn More**: [Context Structure Guide](internal/contexts/README.md)

---

## 🧪 Testing Strategy

Promenade uses **4-tier testing approach** for comprehensive quality assurance:

### Test Pyramid

| Level           | Count | Speed | Database  | Purpose                           |
| --------------- | ----- | ----- | --------- | --------------------------------- |
| **Unit**        | 200+  | < 1s  | ❌ Mock   | Business logic, pure functions    |
| **Smoke**       | 26    | < 5s  | ✅ Real   | Basic CRUD, fast feedback loop    |
| **Integration** | 29    | ~30s  | ✅ Real   | Full workflows, cross-aggregate   |
| **Benchmark**   | 15+   | < 10s | ⚡ Memory | Performance, regression detection |

### Run Tests

```bash
# All tests (recommended before commit)
make test

# Individual test levels
make test-unit           # Unit tests only (fast)
make test-smoke          # Smoke tests (basic CRUD)
make test-integration    # Integration tests (full workflows)
make test-benchmark      # Performance benchmarks

# Pre-push validation (lint + all tests + build)
make pre-push

# Specific package
go test ./internal/contexts/customer-mgmt/...
go test ./pkg/bus/... -bench=.
```

### Current Test Status

✅ **All tests passing**

- **Unit**: 200+ tests, 100% passing
- **Smoke**: 26/26 contexts, < 5s runtime
- **Integration**: 29/29 tests, full database workflows
- **Benchmarks**: Event Bus 377K events/sec, UUID 100M/sec

### Test Best Practices

- **Unit tests** — Next to source files (`*_test.go`)
- **Smoke tests** — Mirror structure in `test/smoke/contexts/`
- **Integration tests** — Group by context in `test/integration/contexts/`
- **Coverage target** — 80%+ for critical business logic

**📚 Learn More**: [Complete Testing Guide](test/README.md)

---

## 📚 Documentation

### Business Documentation

**For Decision Makers & Stakeholders**

- **[Business Overview](docs/business/BUSINESS_OVERVIEW.md)** — Platform capabilities, market opportunity, ROI
  - Available in: 🇬🇧 English, 🇺🇦 Ukrainian, 🇪🇸 Spanish, 🇩🇪 German, 🇫🇷 French, 🇵🇹 Portuguese, 🇨🇳 Chinese, 🇯🇵 Japanese
- **[Ukraine Market Strategy](docs/roadmap/UKRAINE_MARKET_STRATEGY_2026.md)** — Market analysis, TAM $45M+, GTM strategy
- **[Implementation Status](docs/business/IMPLEMENTATION_SUMMARY.md)** — Feature completion, development roadmap
- **[Deployment Options](docs/deployment/)** — Kubernetes, AWS, Docker Compose, monitoring

### Developer Documentation

**Getting Started**

- **[Quick Start Guide](docs/guides/quick-start.md)** — 5-minute setup, first API calls
- **[API Documentation](docs/guides/api-documentation.md)** — Swagger UI, Postman collection
- **[Development Workflow](docs/guides/development-workflow.md)** — Daily commands, debugging, troubleshooting
- **[Contributing Guide](CONTRIBUTING.md)** — Code standards, PR process, git hooks

**Architecture & Patterns**

- **[Clean Architecture](docs/concepts/clean-architecture.md)** — Layers, dependencies, testability
- **[Bounded Contexts](docs/concepts/bounded-contexts.md)** — DDD principles, context isolation
- **[Event-Driven Architecture](docs/concepts/event-driven.md)** — Event Bus, domain events, sagas
- **[Fulfillment Saga](docs/concepts/fulfillment-saga.md)** — Distributed transaction orchestration

**Development Guides**

- **[Security Patterns](docs/guides/security-patterns.md)** — Authentication, authorization, error handling
- **[Database Strategy](docs/guides/database-strategy.md)** — PostgreSQL patterns, migrations, transactions
- **[Error Handling](docs/guides/error-handling-patterns.md)** — Domain errors, HTTP responses, logging
- **[JSON Fields](docs/guides/json-field-validation.md)** — Type-safe JSONB handling
- **[Profiling](docs/guides/profiling.md)** — CPU/memory profiling, pprof, optimization

**Architecture Decision Records**

- **[ADR-0001: Use UUID v7](docs/adr/adr-0001-use-uuid-v7.md)** — Time-ordered UUIDs for better DB performance
- **[ADR-0002: Use sqlx over GORM](docs/adr/adr-0002-use-sqlx-over-gorm.md)** — Control, performance, maintainability
- **[ADR-0003: Use Redis for Caching](docs/adr/adr-0003-use-redis-for-caching.md)** — Distributed cache strategy
- **[ADR Template](docs/adr/adr-template.md)** — Template for new ADRs

### Bounded Context Documentation

Each context has detailed README with domain model, use cases, and API examples:

- **[Identity](internal/contexts/identity/README.md)** — Users, authentication, RBAC
- **[Customer Management](internal/contexts/customer-mgmt/README.md)** — CRM, deals, interactions
- **[Order Management](internal/contexts/order-mgmt/README.md)** — Orders, fulfillment workflows
- **[Billing](internal/contexts/billing/README.md)** — Invoices, payments, subscriptions
- **[Warehouse](internal/contexts/warehouse/README.md)** — Inventory, products, stock movements
- **[Accounting](internal/contexts/accounting/README.md)** — Chart of accounts, journal entries
- **[Banking](internal/contexts/banking/README.md)** — Bank accounts, reconciliation
- **[Fiscal](internal/contexts/fiscal/README.md)** — ПРРО, Checkbox integration

### Package Documentation

Reusable packages with usage examples and benchmarks:

- **[Event Bus](pkg/bus/README.md)** — In-memory & Redis, 377K events/sec
- **[Aggregates](pkg/aggregate/README.md)** — Base aggregate, change tracking, Touch()
- **[JSON Store](pkg/jsonstore/README.md)** — Type-safe JSONB fields
- **[UUIDv7](pkg/uuidv7/README.md)** — Time-ordered IDs, 100M/sec generation
- **[Value Objects](pkg/valueobject/README.md)** — Email, phone, money, address

### API Examples

- **[Postman Collection](postman/)** — Import-ready API collection
- **[Customer Management Examples](postman/examples/customer-mgmt-examples.md)** — CRM workflows
- **[Order Management Examples](postman/examples/order-mgmt-examples.md)** — Order processing
- **[Billing Examples](postman/examples/billing-examples.md)** — Invoice & payment flows

---

## 🤝 Contributing

We welcome contributions! Promenade follows strict code standards to maintain quality.

### Before You Start

1. Read [CONTRIBUTING.md](CONTRIBUTING.md) — Code standards, PR process
2. Set up git hooks: `make setup-hooks` — Automatic linting and testing
3. Review [Architecture Decision Records](docs/adr/) — Understand design choices

### Development Workflow

```bash
# 1. Create feature branch
git checkout -b feature/your-feature-name

# 2. Make changes and test
make test              # Run all tests
make lint              # Check code style

# 3. Pre-push validation (automatic via git hook)
make pre-push          # Lint + tests + build

# 4. Commit and push
git add .
git commit -m "feat: add amazing feature"
git push origin feature/your-feature-name
```

### Code Standards

- **Go** — Follow [Effective Go](https://go.dev/doc/effective_go) and project conventions
- **DDD** — Respect bounded context boundaries, no cross-context imports
- **Testing** — Add tests for new features (smoke + integration)
- **Documentation** — Update README and context docs for public APIs
- **Commits** — Follow [Conventional Commits](https://www.conventionalcommits.org/)

### What to Contribute

**High-Priority Areas**:

- 🔌 **Integrations** — Payment gateways, shipping providers, accounting systems
- 🌍 **Localization** — UI translations, fiscal compliance for other countries
- 📊 **Analytics** — Reports, dashboards, data visualization
- 🔒 **Security** — Audit, penetration testing, security hardening
- 📱 **Clients** — SDKs for popular languages (Python, PHP, JavaScript)

**Good First Issues**: Check [GitHub Issues](https://github.com/basilex/promenade/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)

---

## 📋 Roadmap

### ✅ Completed (v0.1.0 - Current)

- ✅ 8 bounded contexts with clean architecture
- ✅ Customer Management (CRM)
- ✅ Order Management with fulfillment workflows
- ✅ Warehouse & Inventory
- ✅ Billing & Invoicing
- ✅ Accounting (Chart of Accounts, Journal Entries)
- ✅ Banking Integration
- ✅ Fiscal Compliance (ПРРО, Checkbox)
- ✅ Event-driven architecture (377K events/sec)
- ✅ Comprehensive test coverage (200+ tests)
- ✅ API documentation (Swagger/OpenAPI)

### 🔄 In Progress (v0.2.0 - Q1 2026)

- 🔄 **Multi-tenancy** — Organization isolation, data partitioning
- 🔄 **Frontend** — Next.js admin panel, Flutter mobile app
- 🔄 **Advanced Analytics** — BI dashboards, custom reports
- 🔄 **Monobank Integration** — Bank statement import, automatic reconciliation
- 🔄 **Email/SMS** — Transactional notifications, marketing campaigns

### 🎯 Planned (v0.3.0+ - 2026)

- 📅 **CRM Enhancements** — Marketing automation, lead scoring, pipeline analytics
- 📅 **Advanced Billing** — Usage-based pricing, revenue recognition, dunning
- 📅 **Procurement** — Purchase orders, supplier management, receiving
- 📅 **Manufacturing** — Bill of materials, production orders, shop floor tracking
- 📅 **Project Management** — Projects, tasks, time tracking, resource planning
- 📅 **HR Module** — Employees, payroll, attendance, leave management
- 📅 **Mobile Apps** — iOS/Android native apps with offline support
- 📅 **AI/ML** — Demand forecasting, churn prediction, smart recommendations

**📚 Full Roadmap**: [Project Roadmap](docs/roadmap/)

---

## 📄 License

**MIT License** — See [LICENSE](LICENSE) file for details.

Free to use for commercial and non-commercial projects. Attribution appreciated but not required.

---

## 🙏 Acknowledgments

Built with love using:

- [Go](https://go.dev) — Google's systems programming language
- [PostgreSQL](https://postgresql.org) — World's most advanced open source database
- [Gin](https://gin-gonic.com) — High-performance HTTP framework
- [Redis](https://redis.io) — In-memory data structure store
- [Swagger](https://swagger.io) — API documentation and testing

Special thanks to the DDD community and all contributors!

---

## 📞 Support & Community

- 📧 **Email**: support@promenade.dev
- 💬 **Discord**: [Join our community](https://discord.gg/promenade)
- 🐛 **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- 📖 **Docs**: [Documentation Site](https://promenade.dev/docs)
- 🐦 **Twitter**: [@PromenadeERP](https://twitter.com/PromenadeERP)

---

<div align="center">

**⭐ Star us on GitHub** — it helps!

[Documentation](https://promenade.dev/docs) • [API Reference](http://localhost:8081/api/docs) • [Contributing](CONTRIBUTING.md)

Made with ❤️ for Ukrainian businesses

</div>
