**Navigation**: [Home](../README.md) > Internal

---

# Internal - Application Core

**Core application code** organized by architectural layers and bounded contexts.

---

## Directory Structure

```
internal/
 CORE.md                  # Core architecture documentation
 contexts/                # Bounded Contexts (DDD domains)
    README.md            # Contexts overview
    shared/              # Shared context (reference data)
    identity/            # Identity context (users, auth)
    customer-mgmt/       # Customer Management context
    billing/             # Billing context (planned)
    order-mgmt/          # Order Management context (planned)
    warehouse/           # Warehouse context (planned)
 infrastructure/          # Cross-cutting infrastructure
     README.md           # Infrastructure overview
     config/             # Configuration management
     database/           # Database connections & transactions
     health/             # Health monitoring
```

---

## Quick Links

### Bounded Contexts

- [Contexts Overview](contexts/README.md) - All bounded contexts catalog
- [Identity Context](contexts/identity/README.md) - User, Contact, Profile, RBAC
- [Shared Context](contexts/shared/README.md) - Reference data (Country, Currency, etc.)
- [Customer Management](contexts/customer-mgmt/README.md) - Customer aggregate

### Infrastructure

- [Infrastructure Overview](infrastructure/README.md) - Cross-cutting concerns
- [Configuration](infrastructure/config/README.md) - YAML config management
- [Database](infrastructure/database/README.md) - PostgreSQL + transactions
- [Health Checks](infrastructure/health/README.md) - Dependency monitoring

---

## Architecture Principles

**Domain-Driven Design (DDD)**:
- Bounded Contexts isolate business domains
- Aggregates enforce business invariants
- Value Objects represent immutable concepts
- Domain Events enable async communication

**Clean Architecture**:
- Domain layer (entities, aggregates)
- Use case layer (business logic)
- Infrastructure layer (adapters)
- Clear dependency rules (inward only)

**Event-Driven**:
- Contexts communicate via Event Bus only
- No direct dependencies between contexts
- Asynchronous, decoupled integration

---

## Related Documentation

- [Main README](../README.md) - Project overview
- [Documentation Index](../docs/INDEX.md) - Complete documentation
- [Clean Architecture Guide](../docs/concepts/clean-architecture.md) - Architecture principles
- [Bounded Contexts Strategy](../docs/concepts/bounded-contexts.md) - Context design
- [Event-Driven Architecture](../docs/concepts/event-driven.md) - Event Bus patterns

---

**Organization**: DDD with Bounded Contexts  
**Status**: 3 contexts production-ready  
**Maintainer**: Promenade Team
