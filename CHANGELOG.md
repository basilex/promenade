# Changelog

All notable changes to **Promenade Platform** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased] - 0.2.0-dev

### Added

- Multi-tenancy support (organization isolation, data partitioning)
- Frontend applications (Next.js admin panel, Flutter mobile app)
- Advanced analytics (BI dashboards, custom reports)
- Monobank integration (bank statement import, automatic reconciliation)
- Email/SMS notifications (transactional, marketing campaigns)

### Changed

- N/A

### Fixed

- N/A

### Deprecated

- N/A

### Removed

- N/A

### Security

- N/A

---

## [0.1.0-dev] - 2026-01-22

### Added

- Git hooks for pre-commit and pre-push CI checks (`.githooks/`)
- `make setup-hooks` target to install git hooks
- Enhanced `make test-coverage` with coverage summary and thresholds
- Performance profiling guide ([docs/guides/profiling.md](docs/guides/profiling.md))
- Observability strategy documentation ([docs/guides/observability-strategy.md](docs/guides/observability-strategy.md))
- JSON field validation guide ([docs/guides/json-field-validation.md](docs/guides/json-field-validation.md))
- Migration rollback testing guide ([docs/guides/migration-testing-rollback.md](docs/guides/migration-testing-rollback.md))
- pgx migration documentation and implementation
- Documentation for UI context ([internal/contexts/ui/README.md](internal/contexts/ui/README.md))
- Documentation for Scripting context ([internal/contexts/scripting/README.md](internal/contexts/scripting/README.md))
- HTTP Response Standards section in API documentation guide
- Performance baselines and regression detection in benchmark README

### Changed

- **BREAKING**: Migrated from lib/pq to pgx/v5/stdlib (20-30% performance improvement)
- **BREAKING**: Removed MSSQL support (PostgreSQL 14+ only)
- Updated accounting integration test README with comprehensive test coverage (102+ tests)
- Enhanced API versioning guide with current implementation details
- Integration tests now run sequentially (-p 1) due to advisory locks

### Fixed

- Integration test deadlocks with pgx driver (added advisory locks in CleanAllTables)
- Array scanning for PostgreSQL TEXT[] columns (custom stringSlice scanner)
- SQL type casting for CONCAT operations (explicit int::text casting)

---

## [0.1.0] - 2026-01-22

### Added

#### Core Infrastructure

- **Clean Architecture**: DDD bounded contexts in `internal/contexts/` with strict boundaries
- **Event Bus**: In-memory event bus (`pkg/bus/`) for inter-context communication (67 tests, 377K events/sec)
- **Workspace State Management**: `.promenade.workspace` for DATABASE_DRIVER + ENVIRONMENT switching
- **Makefile Modules**: Split into `Makefile.dev.mk`, `Makefile.test.mk`, `Makefile.prod.mk`

#### Bounded Contexts (11 Total)

**Shared**:

- Country aggregate (ISO 3166-1 alpha-2, currencies, timezones)
- Currency aggregate (ISO 4217, exchange rates, precision)
- Language aggregate (ISO 639-1, localization support)

**Identity**:

- User aggregate (authentication, authorization)
- Contact aggregate (CRM contacts)
- Profile aggregate (user profiles with custom fields)

**Customer Management**:

- Customer aggregate (CRM, invoicing, multi-contact)
- Interaction aggregate (activity tracking, notes, calls)
- Deal aggregate (pipeline management, weighted forecast)

**Order Management**:

- Order aggregate (fulfillment saga, state machine)
- OrderItem aggregate (products, pricing, discounts)
- Fulfillment saga (inventory reservation → payment → shipping)

**Billing**:

- Invoice aggregate (multi-currency, line items)
- Payment aggregate (tracking, reconciliation)
- Subscription aggregate (recurring billing, auto-renewal)

**Accounting**:

- FiscalPeriod aggregate (period management, locking)
- TaxCode aggregate (rates, jurisdictions)
- CostCenter aggregate (departmental tracking)
- Account aggregate (chart of accounts, balance tracking)
- Budget aggregate (planning, variance analysis)
- JournalEntry aggregate (event sourcing, double-entry)
- Reconciliation aggregate (bank reconciliation)

**Warehouse**:

- Product aggregate (inventory management)
- Stock aggregate (location tracking, lot/serial)
- StockMovement aggregate (transfers, adjustments)
- Reservation aggregate (inventory holds for orders)

**Fiscal**:

- FiscalDocument aggregate (invoices, receipts, credit notes)
- FiscalSequence aggregate (sequential numbering with rollover)
- TaxReport aggregate (VAT/GST reporting)

**Banking**:

- BankAccount aggregate (multi-currency accounts)
- BankTransaction aggregate (transaction import, reconciliation)
- BankStatement aggregate (statement parsing, matching)

**UI**:

- Form metadata for dynamic form generation (server-driven UI)
- Field types: text, email, select, checkbox, date, etc.

**Scripting**:

- Lua script execution sandbox (validation, transformation, automation)
- 17 tests, security features (limited stdlib, timeouts)

#### Packages (Reusable)

**Core**:

- `pkg/aggregate`: BaseAggregate with UUID v7, timestamps, Touch()
- `pkg/uuidv7`: Time-sortable UUID v7 implementation (RFC 9562)
- `pkg/database`: sqlx wrapper with transaction propagation
- `pkg/jsonstore`: jsonstore.Field[T] for flexible JSON columns
- `pkg/valueobject`: Email, PhoneNumber, Address, Money, TaxRate
- `pkg/bus`: In-memory event bus (publish-subscribe, 67 tests)
- `pkg/logger`: Structured logging wrapper (zerolog/logrus compatible)
- `pkg/middleware`: Gin middleware (auth, CORS, rate limiting, CSRF)
- `pkg/response`: Standardized HTTP responses
- `pkg/saga`: Saga pattern orchestration (compensation, retries)
- `pkg/scheduler`: Cron-based task scheduler
- `pkg/cache`: Redis caching layer
- `pkg/jwt`: JWT token generation and validation
- `pkg/ref`: Reference data helpers

**Fiscal**:

- `pkg/fiscal`: Fiscal document generation (AT, ES, RO, etc.)

#### Testing

**Four-Tier Testing**:

- **Unit Tests**: 2465+ tests across all contexts (90%+ coverage)
- **Smoke Tests**: 35+ HTTP handler tests (no DB, mirror path in `test/smoke/`)
- **Integration Tests**: 102+ accounting tests, full DB (PostgreSQL/SQLite)
- **Benchmarks**: 12+ benchmarks (CreateInteraction 380μs, ListByCustomer 886μs)

**Test Commands**:

- `make test` - All tests
- `make test-unit` - Unit tests only (fast)
- `make test-smoke` - Smoke tests (HTTP handlers)
- `make test-integration` - Integration tests (DB required)
- `make test-benchmark` - Benchmarks

#### Migrations

**Namespace-Based Migrations**:

- 11 namespaces (core, identity, customer-mgmt, order-mgmt, billing, accounting, warehouse, fiscal, banking, ui, scripting)
- PostgreSQL 14+ support
- Auto-run on app startup
- `make migrate` - Run all migrations
- `make migrate-module MODULE=order-mgmt` - Run specific module
- `make migrate-rollback MODULE=billing STEPS=1` - Rollback (manual)

#### Development Workflow

**Environment Switching**:

- `make switch-postgres-dev` - PostgreSQL + development
- `make switch-postgres-test` - PostgreSQL + test
- `make switch-postgres-prod` - PostgreSQL + production

**Development Commands**:

- `make dev` - Start dev server (Docker + migrations + API)
- `make dev-fresh` - Fresh start with clean database
- `make docker-up` - Start PostgreSQL + Redis
- `make seed` - Seed database with test data

**CI/CD Simulation**:

- `make pre-push` - Full CI checks (lint, tests, build)
- `make ci-check` - CI validation (lint + tests + build)
- `make ci-lint` - Linting only
- `make ci-test` - Tests only
- `make ci-build` - Build only

#### Documentation

**Comprehensive Guides**:

- [README.md](README.md) - Project overview
- [docs/INDEX.md](docs/INDEX.md) - Documentation index
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [docs/guides/security-patterns.md](docs/guides/security-patterns.md) - Security best practices (551 lines)
- [docs/guides/error-handling-patterns.md](docs/guides/error-handling-patterns.md) - Error handling (Gold Standard Pattern)
- [docs/guides/api-documentation.md](docs/guides/api-documentation.md) - Swagger/OpenAPI guide
- [docs/guides/api-versioning.md](docs/guides/api-versioning.md) - Versioning strategy (/api/v1 implemented)
- [docs/guides/local-ci.md](docs/guides/local-ci.md) - Local CI validation
- [test/README.md](test/README.md) - Testing strategy
- [pkg/bus/README.md](pkg/bus/README.md) - Event Bus documentation

**Business Documentation**:

- [docs/business/BUSINESS_OVERVIEW.md](docs/business/BUSINESS_OVERVIEW.md) - Business domain overview
- Translations: DE, ES, FR, JP, PT, UK, ZH

**Concepts**:

- [docs/concepts/bounded-contexts.md](docs/concepts/bounded-contexts.md) - DDD context boundaries
- [docs/concepts/clean-architecture.md](docs/concepts/clean-architecture.md) - Clean Architecture layers
- [docs/concepts/event-driven.md](docs/concepts/event-driven.md) - Event-driven architecture
- [docs/concepts/fulfillment-saga.md](docs/concepts/fulfillment-saga.md) - Order fulfillment saga

### Infrastructure

**Docker**:

- `docker/docker-compose.postgres.dev.yml` - PostgreSQL 16 + Redis 7
- `docker/docker-compose.postgres.test.yml` - Test environment
- `docker/docker-compose.postgres.prod.yml` - Production config

**Configuration**:

- `config/app.postgres-dev.yaml` - Development config
- `config/app.postgres-test.yaml` - Test config
- `config/app.postgres-prod.yaml` - Production config

**Scripts**:

- `scripts/create-migration.sh` - Create new migration
- `scripts/check-links.sh` - Validate documentation links
- `scripts/clean-emojies.py` - Remove emoji from docs (policy enforcement)

---

## Version History

- **[Unreleased]** - Current development (Phase 2 planning)
- **[0.1.0]** - 2026-01-22 - Initial release (Phase 1 complete)

---

## Links

- [Promenade Repository](https://github.com/basilex/promenade)
- [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
- [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
