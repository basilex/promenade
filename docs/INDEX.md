# Documentation Index

Complete guide to Promenade CRM Platform architecture and development.

---

## Quick Navigation

### Getting Started

- [Main README](../README.md) - Project overview, quick start, architecture
- [AI Instructions](../.github/copilot-instructions.md) - Essential guide for AI coding agents
- [Testing Guide](../test/README.md) - Three-tier testing strategy
- [Testing Patterns](TESTING_PATTERNS.md) - **Comprehensive testing patterns guide** (unit, smoke, integration)
- [Testing Quick Reference](TESTING_QUICK_REFERENCE.md) - One-page cheat sheet

### Architecture & Design

- [Clean Architecture Summary](CLEAN_ARCHITECTURE_SUMMARY.md) - DDD with Bounded Contexts
- [RBAC Implementation](RBAC.md) - Complete Role-Based Access Control guide
- [Phase 1 Architecture Preparation](PHASE1_ARCHITECTURE_PREPARATION.md) - Migration roadmap
- [Event Bus Documentation](../pkg/bus/README.md) - Central communication hub (Memory/Redis)
- [Event Bus Test Coverage](BUS_TEST_COVERAGE.md) - Test report (67 tests, 100% passing)

### Bounded Contexts

| Context                 | Status     | Documentation                                                  |
| ----------------------- | ---------- | -------------------------------------------------------------- |
| **Shared**              | Production | [README](../internal/contexts/shared/README.md) (~450 lines)   |
| **Identity**            | Production | [README](../internal/contexts/identity/README.md) (~550 lines) |
| **Customer Management** | Planned    | [README](../internal/contexts/customer-mgmt/README.md)         |
| **Order Management**    | Planned    | [README](../internal/contexts/order-mgmt/README.md)            |
| **Billing**             | Planned    | [README](../internal/contexts/billing/README.md)               |
| **Warehouse**           | Planned    | [README](../internal/contexts/warehouse/README.md)             |

### Package Library

- [Package Overview](../pkg/README.md) - All shared packages (~400 lines)
- [Event Bus](../pkg/bus/README.md) - Memory/Redis adapters (~600 lines)
- [UUID v7](../pkg/uuidv7/uuidv7.go) - Time-ordered UUIDs (RFC 9562)
- [Value Objects](../pkg/valueobject/) - Email, Phone, Money, Address
- [Logger](../pkg/logger/logger.go) - Structured logging (slog)
- [Response](../pkg/response/response.go) - Standard HTTP responses

### Development

- [Makefile](../Makefile) - Main commands (`make help`)
- [Makefile.dev.mk](../Makefile.dev.mk) - Development workflow
- [Makefile.test.mk](../Makefile.test.mk) - Testing commands
- [Makefile.prod.mk](../Makefile.prod.mk) - Production/DevOps
- [Docker Setup](../docker/README.md) - PostgreSQL, Redis, test environment

### Database

- [Migrations README](../migrations/README.md) - Namespace-based migration system
- Core migrations: `migrations/core/` - Extensions (UUID v7)
- Shared migrations: `migrations/shared/` - Reference data
- Identity migrations: `migrations/identity/` - Users, contacts

---

## By Task

### I want to...

**Add a new aggregate**:

1. Read [.github/copilot-instructions.md - Adding a New Aggregate](../.github/copilot-instructions.md#adding-a-new-aggregate)
2. Check context structure in [README - Project Structure](../README.md#-project-structure)
3. See examples in `internal/contexts/identity/contact/`

**Write tests**:

1. Read [docs/TESTING_PATTERNS.md](TESTING_PATTERNS.md) - **ЕТАЛОННІ ПАТЕРНИ** (unit/smoke/integration)
2. Check [test/README.md](../test/README.md) - Three-tier testing strategy overview
3. See examples:
   - Unit: `internal/contexts/identity/user/*_test.go`
   - Smoke: `test/smoke/contexts/identity/user/handler_test.go`
   - Integration: `test/integration/contexts/identity/user/repository_test.go`

**Add domain events**:

1. Read [pkg/bus/README.md](../pkg/bus/README.md) - Event Bus documentation
2. Check topic constants in `pkg/bus/topics.go`
3. See publish example in Contact UseCase

**Create migrations**:

1. Run `make migrate-new CONTEXT=identity NAME=add_users_table`
2. Edit generated files in `migrations/identity/`
3. Run `make migrate-identity`

**Add new bounded context**:

1. Read [Clean Architecture Summary](CLEAN_ARCHITECTURE_SUMMARY.md)
2. Copy structure from `internal/contexts/shared/`
3. Create router and register in `cmd/api/main.go`

---

## Documentation Statistics

- **Total guides**: 9 core documents
- **Context READMEs**: 6 bounded contexts
- **Package docs**: 10+ packages with documentation
- **Total lines**: ~15,000+ lines of documentation
- **Code examples**: 100+ working examples
- **Test coverage**: 260+ tests documented

---

## Recently Updated

- **2025-12-28**: Created TESTING_PATTERNS.md - comprehensive testing guide (~800 lines)
- **2025-12-28**: User smoke tests completed (11 tests passing)
- **2025-12-28**: Standardized integration tests (Contact, Profile, User)
- **2025-12-28**: Updated AI instructions (removed module references)
- **2025-12-28**: Fixed Identity Context status (Contact ready, User/Profile planned)
- **2025-12-28**: Created INDEX.md for navigation
- **2025-12-27**: Event Bus README (~600 lines)
- **2025-12-27**: Testing structure migration

---

## Contributing

When adding documentation:

1.  **Update this INDEX.md** - add links to new documents
2.  **Follow existing structure** - use same markdown style
3.  **Include examples** - show, don't just tell
4.  **Keep it accurate** - documentation must match code
5.  **Add date to "Recently Updated"** - track changes

---

**Last updated**: 2025-12-28  
**Maintainer**: Promenade Team
