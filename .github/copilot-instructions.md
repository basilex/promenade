```instructions
# Promenade Copilot Instructions

See [README.md](README.md) and [docs/INDEX.md](docs/INDEX.md) for full context.

## Architecture & boundaries
- Strict DDD bounded contexts in [internal/contexts](internal/contexts); no cross-context imports. Communicate via Event Bus in [pkg/bus](pkg/bus/README.md).
- Context layout follows clean architecture: aggregate → repository → use case → handler (see [internal/contexts/README.md](internal/contexts/README.md)).
- Dependency wiring lives in [cmd/api/bootstrap.go](cmd/api/bootstrap.go); entry point is [cmd/api/main.go](cmd/api/main.go).
- Modules vs contexts: modules are technical/feature toggles, contexts are domain boundaries (see [internal/contexts/README.md](internal/contexts/README.md)).
- Multi-DB support is intentional (Postgres prod, SQLite dev/test). Migrations must be DB-agnostic (use TEXT for UUID/JSON) and IDs generated in Go (see [migrations/README.md](migrations/README.md)).

## Project-specific conventions
- IDs: always `uuidv7.New()` from [pkg/uuidv7](pkg/uuidv7/README.md); never UUID v4.
- Aggregates embed `aggregate.BaseAggregate` and use `Touch()`/`GetID()` (examples in [pkg/aggregate/README.md](pkg/aggregate/README.md)).
- JSON fields use `jsonstore.Field[T]` from [pkg/jsonstore](pkg/jsonstore/README.md); no manual marshal/unmarshal.
- Repositories embed `BaseRepository` and call `getExecutor(ctx)` for tx propagation (see examples under [internal/contexts](internal/contexts)).
- Use cases return domain error constants from `errors.go`; handlers discriminate with `errors.Is()` and return generic system errors (see [docs/guides/security-patterns.md](docs/guides/security-patterns.md)).

## Critical workflows
- Workspace state in [.promenade.workspace](.promenade.workspace.example). Typical flow: make switch-postgres-dev or make switch-sqlite-dev → make dev / make dev-fresh (see [Makefile](Makefile)).
- Tests are tiered: make test-unit, make test-smoke, make test-integration, make test-benchmark, make test (see [test/README.md](test/README.md)). Smoke tests mirror handler paths under test/smoke/contexts.
- Namespace-based migrations: make migrate, make migrate-module MODULE=..., make migrate-rollback MODULE=... (see [migrations/README.md](migrations/README.md)). Migrations also run on app startup (see [cmd/api/main.go](cmd/api/main.go)).
- Before push: make pre-push (lint + tests + build) in [Makefile.dev.mk](Makefile.dev.mk).

## Integration points & data flow
- Event topics are constants in [pkg/bus/topics.go](pkg/bus/topics.go). Publish after state changes; log publish errors without failing operations (see [pkg/bus/README.md](pkg/bus/README.md)).
- Cross-context communication uses ACLs (Anti-Corruption Layer) instead of direct aggregate/model sharing (see [internal/contexts/README.md](internal/contexts/README.md)).
- Warehouse integration: order events → inventory reservation/commit in [internal/contexts/warehouse/integration](internal/contexts/warehouse/integration).
- Fiscal integration registers order event handlers in [cmd/api/bootstrap.go](cmd/api/bootstrap.go) and [internal/contexts/fiscal/integration](internal/contexts/fiscal/integration).
- Health endpoints live in [internal/infrastructure/health](internal/infrastructure/health) and are wired during bootstrap.
```
