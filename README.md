# Promenade Platform

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Tests](https://img.shields.io/badge/Tests-passing-success?style=flat)](test/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![DDD](https://img.shields.io/badge/Architecture-DDD-green.svg)](docs/concepts/clean-architecture.md)

**Enterprise backend platform for customer management, orders, billing, and warehouse operations** — built with **Domain-Driven Design**, **Clean Architecture**, and **Event-Driven patterns**.

> **Modular by Design** — Start with customer management, add orders when needed, integrate warehouse when ready. Each context is independent and can be deployed separately.

---

## 📋 What is Promenade?

Promenade is an **open-source business platform** that replaces multiple SaaS tools with a single, modular backend:

- **CRM** — Customer management, sales pipeline, interactions tracking
- **Order Management** — Order processing, fulfillment workflows, state machines
- **Warehouse** — Inventory tracking, stock movements, product management, multi-location support
- **Billing** — Invoices, payments, subscriptions, recurring billing

**Built for**:

- SaaS companies needing customizable CRM/Order management
- SMB companies replacing legacy ERP systems (1C, Bitrix24)
- Development teams building custom business solutions
- System integrators needing extensible backend platform

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+** installed
- **Docker** and Docker Compose (for PostgreSQL)
- **Make** utility

### Run in 3 Commands

```bash
# 1. Clone and enter directory
git clone https://github.com/basilex/promenade.git
cd promenade

# 2. Start PostgreSQL and create database
make docker-up
make migrate

# 3. Run API server
make dev
```

Server starts at **http://localhost:8081**

- **API Documentation**: http://localhost:8081/api/docs/index.html (Swagger UI)
- **Health Check**: http://localhost:8081/health

**Next Steps**:

- [Complete Quick Start Guide](docs/guides/quick-start.md) — Registration, authentication, first requests
- [API Documentation Guide](docs/guides/api-documentation.md) — Swagger UI, Postman collection
- [Development Workflow](docs/guides/development-workflow.md) — Daily commands, testing, debugging

---

## 📚 Documentation

### For Business Users

- **[Business Overview](docs/business/BUSINESS_OVERVIEW.md)** — Platform capabilities, market opportunity, ROI (8 languages)
- **[Ukraine Market Strategy](docs/roadmap/UKRAINE_MARKET_STRATEGY_2026.md)** — Market analysis, competitive advantages, revenue projections
- **[Implementation Status](docs/business/IMPLEMENTATION_SUMMARY.md)** — Feature completion, development roadmap

### For Developers

- **[Architecture Overview](docs/concepts/clean-architecture.md)** — DDD principles, bounded contexts, clean architecture layers
- **[Quick Start Guide](docs/guides/quick-start.md)** — Get running in 5 minutes
- **[API Documentation](docs/guides/api-documentation.md)** — Swagger UI, Postman collection, common workflows
- **[Development Guide](docs/guides/development-workflow.md)** — Daily commands, testing, debugging
- **[Testing Guide](test/README.md)** — Unit, smoke, integration, benchmark tests
- **[Security Patterns](docs/guides/security-patterns.md)** — Error handling, validation, authentication
- **[Domain Errors](docs/guides/domain-errors.md)** — Type-safe error handling patterns

### Core Concepts

- **[Bounded Contexts](docs/concepts/bounded-contexts.md)** — Domain isolation, cross-context communication
- **[Event-Driven Architecture](docs/concepts/event-driven.md)** — Event Bus, domain events, async workflows
- **[Clean Architecture](docs/concepts/clean-architecture.md)** — Layers, dependencies, testability
- **[Aggregates](pkg/aggregate/README.md)** — Domain modeling, invariants, transactions

### Package Documentation

- **[Event Bus](pkg/bus/README.md)** — Memory and Redis adapters, 377K events/sec
- **[Aggregates](pkg/aggregate/README.md)** — Base aggregate, change tracking
- **[JSON Store](pkg/jsonstore/README.md)** — Type-safe JSON fields for PostgreSQL/SQLite
- **[UUIDv7](pkg/uuidv7/README.md)** — Time-ordered UUIDs, 100 million IDs/sec

### Context Documentation

- **[Identity Context](internal/contexts/identity/README.md)** — Users, contacts, profiles, roles, permissions
- **[Customer Management](internal/contexts/customer-mgmt/README.md)** — Customers, companies, deals, interactions, analytics
- **[Order Management](internal/contexts/order-mgmt/README.md)** — Orders, fulfillment workflows, state machines
- **[Billing](internal/contexts/billing/README.md)** — Invoices, payments, subscriptions
- **[Warehouse](internal/contexts/warehouse/README.md)** — Inventory, products, stock movements, locations
- **[Fiscal Integration](internal/contexts/fiscal/README.md)** — Cash registers, receipts, tax compliance

---

## 🏗️ Architecture

### Domain-Driven Design

Promenade follows **strict DDD** with 8 bounded contexts, each owning its domain model, database schema, and APIs:

```
┌─────────────────────────────────────────────────────────────┐
│                        Event Bus                             │
│          (Async communication between contexts)              │
└─────────────────────────────────────────────────────────────┘
     ▲          ▲         ▲         ▲         ▲         ▲
     │          │         │         │         │         │
┌────┴───┐ ┌───┴────┐ ┌──┴────┐ ┌──┴────┐ ┌──┴─────┐ ┌┴─────┐
│Identity│ │Customer│ │ Order │ │Billing│ │Warehouse│ │Fiscal│
│        │ │  Mgmt  │ │  Mgmt │ │       │ │        │ │      │
└────────┘ └────────┘ └───────┘ └───────┘ └────────┘ └──────┘
```

**Key Principles**:

- **Bounded Context Autonomy** — Each context has own models, database schema, and migrations
- **No Cross-Context Imports** — Contexts communicate only via Event Bus
- **Clean Architecture Layers** — Handler → UseCase → Repository → Aggregate
- **Aggregates as Transaction Boundaries** — Single aggregate per transaction
- **Domain Events** — State changes published to Event Bus for cross-context reactions

**Learn More**: [docs/concepts/bounded-contexts.md](docs/concepts/bounded-contexts.md)

### Technology Stack

- **Language**: Go 1.24 (performance, concurrency, type safety)
- **Database**: PostgreSQL 16 (primary), SQLite 3 (dev/test)
- **HTTP Framework**: Gin (high performance, middleware)
- **Event Bus**: Memory (dev), Redis (production)
- **Authentication**: JWT with RS256 (stateless, secure)
- **API Documentation**: Swagger/OpenAPI 3.0
- **Testing**: Unit, Smoke, Integration, Benchmark (100% passing)

---

## 📦 Project Structure

```
promenade/
├── cmd/                    # Application entry points
│   ├── api/               # HTTP server (bootstrap, routes, shutdown)
│   ├── migrate/           # Migration runner
│   └── seed/              # Test data seeder
├── internal/              # Private application code
│   ├── contexts/          # 8 bounded contexts (DDD)
│   │   ├── identity/      # Users, authentication, authorization
│   │   ├── customer-mgmt/ # Customers, companies, deals, interactions
│   │   ├── order-mgmt/    # Orders, fulfillment, contracts
│   │   ├── billing/       # Invoices, payments, subscriptions
│   │   ├── warehouse/     # Inventory, products, stock, locations
│   │   ├── fiscal/        # Cash registers, receipts, tax compliance
│   │   ├── shared/        # Countries, currencies, languages, timezones
│   │   └── ui/            # UI metadata (forms, views)
│   └── infrastructure/    # Cross-cutting concerns
│       ├── health/        # Health checks
│       ├── http/          # HTTP server, middleware, response helpers
│       └── auth/          # JWT generation, validation
├── pkg/                   # Reusable packages (can be extracted)
│   ├── bus/              # Event Bus (memory, Redis)
│   ├── aggregate/        # Base aggregate, change tracking
│   ├── jsonstore/        # Type-safe JSON fields
│   ├── uuidv7/           # Time-ordered UUIDs
│   ├── middleware/       # HTTP middleware (auth, rate limiting, CORS)
│   └── response/         # Standard HTTP responses
├── migrations/           # Database migrations (per context)
├── test/                 # Integration and smoke tests
├── docs/                 # Documentation
│   ├── business/        # Business overviews (8 languages)
│   ├── concepts/        # Architecture concepts
│   ├── guides/          # Development guides
│   └── reference/       # API reference
└── config/              # Configuration files (dev, test, prod)
```

**Context Structure** (Clean Architecture):

```
internal/contexts/{context}/
├── {aggregate}/               # Domain aggregate (bounded context subdomain)
│   ├── aggregate/            # Domain entities & value objects
│   │   └── {entity}.go       # Entity with business logic
│   ├── repository/           # Repository interface
│   │   └── {entity}_repository.go
│   ├── usecase/              # Use cases (application logic)
│   │   └── {entity}_usecase.go
│   ├── adapter/              # External adapters
│   │   ├── repository/postgres/  # PostgreSQL implementation
│   │   └── http/                 # HTTP handlers
│   ├── dto/                  # Data Transfer Objects
│   └── errors.go             # Domain error constants
└── README.md                 # Context documentation
```

**Learn More**: [internal/contexts/README.md](internal/contexts/README.md)

---

## 🧪 Testing

Promenade has **comprehensive test coverage** across 4 test levels:

| Test Type       | Count | Speed | Purpose                                       |
| --------------- | ----- | ----- | --------------------------------------------- |
| **Unit**        | 200+  | < 1s  | Individual functions, business logic          |
| **Smoke**       | 26    | < 5s  | Basic CRUD per context, fast feedback         |
| **Integration** | 29    | ~30s  | Real database, full workflows                 |
| **Benchmark**   | 15+   | < 10s | Performance measurement, regression detection |

### Run Tests

```bash
# All tests
make test

# By level
make test-unit          # Unit tests only
make test-smoke         # Smoke tests (fast CRUD checks)
make test-integration   # Integration tests (with DB)
make test-benchmark     # Performance benchmarks

# By package
go test ./internal/contexts/customer-mgmt/customer/...
go test ./pkg/bus/...
```

### Test Results (Latest)

- ✅ **Unit Tests**: 200+ passing
- ✅ **Smoke Tests**: 26/26 passing (all contexts)
- ✅ **Integration Tests**: 29/29 passing (100% compilation)
- ✅ **Benchmarks**: 15+ benchmarks, 377K events/sec (Event Bus)

**Learn More**: [test/README.md](test/README.md)

---

## 🔧 Configuration

Promenade uses **YAML configuration** with environment-specific files:

```
config/
├── app.postgres-dev.yaml   # Development (PostgreSQL)
├── app.postgres-prod.yaml  # Production (PostgreSQL)
├── app.postgres-test.yaml  # Testing (PostgreSQL)
├── app.sqlite-dev.yaml     # Development (SQLite)
└── app.sqlite-test.yaml    # Testing (SQLite)
```

### Switch Database

```bash
# PostgreSQL (default)
make switch-postgres-dev
make dev

# SQLite (faster startup, single file)
make switch-sqlite-dev
make dev
```

### Configuration Structure

```yaml
app:
  name: "Promenade Platform"
  version: "1.0.0"

server:
  host: "localhost"
  port: 8081

database:
  driver: "postgres" # or "sqlite3"
  dsn: "host=localhost port=5432 user=postgres password=postgres dbname=promenade_dev sslmode=disable"

jwt:
  private_key_path: "config/keys/jwt-private.pem"
  public_key_path: "config/keys/jwt-public.pem"
  access_token_duration: "15m"
  refresh_token_duration: "7d"

event_bus:
  adapter: "memory" # or "redis"
  max_retries: 3
```

**Learn More**: [config/README.md](config/README.md)

---

## 🛠️ Development

### Daily Commands

```bash
# Start development server (with hot reload)
make dev

# Run migrations
make migrate                    # All contexts
make migrate-context CONTEXT=warehouse  # Single context

# Reset database
make dev-fresh                  # Drop, create, migrate

# Tests
make test                       # All tests
make test-smoke                 # Fast smoke tests
make test-integration           # Integration tests

# Code quality
make lint                       # Run golangci-lint
make fmt                        # Format code

# Build
make build                      # Production binary
make build-dev                  # Development binary

# Docker
make docker-up                  # Start PostgreSQL
make docker-down                # Stop PostgreSQL
```

### Pre-Push Checklist

```bash
# Run before pushing changes
make pre-push

# This runs:
# 1. make lint        (code quality)
# 2. make test        (all tests)
# 3. make build       (verify compilation)
```

**Learn More**: [docs/guides/development-workflow.md](docs/guides/development-workflow.md)

---

## 📖 API Documentation

### Swagger UI

Interactive API documentation with 182+ endpoints:

**http://localhost:8081/api/docs/index.html**

- Try endpoints directly in browser
- Authentication with JWT
- Request/response examples
- Schema validation

### Postman Collection

Pre-built collection with all endpoints, authentication flows, and test scripts:

```bash
# Import in Postman
1. postman/Promenade_API.postman_collection.json  # All endpoints
2. postman/Development.postman_environment.json   # Dev environment
```

**Features**:

- Auto-save JWT tokens after login
- Auto-refresh expired tokens
- Pre-configured workflows
- Response validation scripts

**Learn More**: [docs/guides/api-documentation.md](docs/guides/api-documentation.md), [postman/README.md](postman/README.md)

---

## 🌍 Multi-Language Support

Business documentation available in 8 languages:

| Language     | Document                                                    | Audience          |
| ------------ | ----------------------------------------------------------- | ----------------- |
| 🇬🇧 English   | [Business Overview](docs/business/BUSINESS_OVERVIEW.md)     | Global audience   |
| 🇺🇦 Ukrainian | [Business Overview](docs/business/BUSINESS_OVERVIEW_UK.md)  | Ukrainian market  |
| 🇩🇪 Deutsch   | [Geschäftsübersicht](docs/business/BUSINESS_OVERVIEW_DE.md) | German market     |
| 🇫🇷 Français  | [Aperçu Commercial](docs/business/BUSINESS_OVERVIEW_FR.md)  | French market     |
| 🇪🇸 Español   | [Resumen de Negocio](docs/business/BUSINESS_OVERVIEW_ES.md) | Spanish market    |
| 🇵🇹 Português | [Visão Geral](docs/business/BUSINESS_OVERVIEW_PT.md)        | Portuguese market |
| 🇯🇵 日本語    | [ビジネス概要](docs/business/BUSINESS_OVERVIEW_JP.md)       | Japanese market   |
| 🇨🇳 中文      | [商业概览](docs/business/BUSINESS_OVERVIEW_ZH.md)           | Chinese market    |

---

## 🗺️ Roadmap

### Q1 2026 (Current)

- ✅ DDD refactoring (24 aggregates, 150+ domain errors)
- ✅ Security audit (36 handlers, 417 fixes)
- ✅ Integration tests (29 packages, 100% passing)
- ✅ Event Bus performance (377K events/sec)
- 🔄 Web frontend (React/TypeScript)
- 🔄 Mobile apps (iOS, Android)

### Q2 2026

- Ukrainian fiscal integration (Checkbox, Vchasno.Kasa)
- Tax invoices (XML for State Tax Service)
- Bank integration (Monobank, Privat24)
- HRM module (payroll + Ukrainian taxes)
- Logistics integration (Nova Poshta, Ukrposhta)

### Q3-Q4 2026

- Multi-tenant SaaS mode
- Advanced analytics (CQRS read models)
- Workflow engine (custom business processes)
- Marketplace (plugins, extensions)

**Detailed Roadmap**: [docs/roadmap/](docs/roadmap/)

---

## 🤝 Contributing

We welcome contributions! Please read:

- **[Contributing Guide](CONTRIBUTING.md)** — How to contribute, coding standards
- **[Development Workflow](docs/guides/development-workflow.md)** — Daily commands, testing
- **[Architecture Overview](docs/concepts/clean-architecture.md)** — Design principles

### Quick Start for Contributors

```bash
# 1. Fork repository
# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/promenade.git
cd promenade

# 3. Create feature branch
git checkout -b feature/my-feature

# 4. Make changes
# 5. Run tests
make pre-push

# 6. Commit and push
git commit -m "feat: add my feature"
git push origin feature/my-feature

# 7. Open Pull Request
```

---

## 📄 License

Promenade is open-source software licensed under the **MIT License**.

See [LICENSE](LICENSE) file for details.

---

## 💬 Support

- **Documentation**: https://basilex.github.io/promenade/
- **Issues**: https://github.com/basilex/promenade/issues
- **Email**: alexander.vasilenko@gmail.com
- **Discussions**: https://github.com/basilex/promenade/discussions

---

## 🏆 Status

- **Version**: 1.0.0
- **Status**: Production-ready
- **Go Version**: 1.24+
- **Database**: PostgreSQL 16, SQLite 3
- **Tests**: ✅ 200+ unit, 26 smoke, 29 integration
- **Code Quality**: ✅ 0 lint issues
- **Security**: ✅ Audit complete (Jan 2026)
- **Documentation**: ✅ Comprehensive (8 languages)

---

**Built with ❤️ using Domain-Driven Design and Go**
