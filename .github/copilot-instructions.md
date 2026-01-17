```instructions
# Promenade Copilot Instructions

Use this as a concise, high-signal guide. See [README.md](README.md) and [docs/INDEX.md](docs/INDEX.md).

## Big picture architecture
- DDD with strict bounded contexts in [internal/contexts](internal/contexts); **no cross-context imports**. Cross-context communication goes through the Event Bus in [pkg/bus](pkg/bus/README.md).
- Contexts are domain-focused; legacy/technical “modules” live elsewhere (see [internal/contexts/README.md](internal/contexts/README.md)).
- Dependency wiring lives in [cmd/api/bootstrap.go](cmd/api/bootstrap.go) (logger → DB → migrations → Redis/cache → JWT → Event Bus → integrations → routers).
- Multi-database support is intentional (PostgreSQL prod, SQLite dev/test). Migrations must be DB-agnostic: use `TEXT` for UUID/JSON, generate IDs in Go (see [migrations/README.md](migrations/README.md)).

## Critical workflows
- Workspace state is in [.promenade.workspace](.promenade.workspace.example). Typical flow: `make switch-postgres-dev` or `make switch-sqlite-dev` → `make dev` / `make dev-fresh` (see [Makefile](Makefile)).
- Tests are four-tier: `make test-unit`, `make test-smoke`, `make test-integration`, `make test` (see [test/README.md](test/README.md) and [test/smoke/README.md](test/smoke/README.md)).
- Namespace-based migrations: `make migrate`, `make migrate-module MODULE=...`, `make migrate-rollback MODULE=...` (see [migrations/README.md](migrations/README.md)).
- Before push: `make pre-push` (lint + tests + build) in [Makefile.dev.mk](Makefile.dev.mk).

## Project-specific conventions (must follow)
- IDs: always `uuidv7.New()` from [pkg/uuidv7](pkg/uuidv7/README.md); never UUID v4.
- Aggregates embed `aggregate.BaseAggregate`; use `Touch()`/`GetID()` (see [pkg/aggregate/README.md](pkg/aggregate/README.md)).
- Use cases return **domain error constants** from `errors.go`; no inline `fmt.Errorf` in `usecase.go`. Handlers map with `errors.Is()` and return generic system errors (see [docs/guides/security-patterns.md](docs/guides/security-patterns.md)).
- Repositories embed `BaseRepository` and use `getExecutor(ctx)` for tx propagation (examples across [internal/contexts](internal/contexts)).
- JSON fields use `jsonstore.Field[T]` from [pkg/jsonstore](pkg/jsonstore/README.md); no manual marshal/unmarshal.

## Integration points & data flows
- Event topics are constants in [pkg/bus/topics.go](pkg/bus/topics.go). Publish after state changes and **log errors without failing** the operation.
- Warehouse integration pattern: order events → inventory reservation/commit in [internal/contexts/warehouse/integration](internal/contexts/warehouse/integration).
- Fiscal integration registers order event handlers in [cmd/api/bootstrap.go](cmd/api/bootstrap.go) and [internal/contexts/fiscal/integration](internal/contexts/fiscal/integration).
- LUA scripting engine and scheduler are production features in [pkg/scripting](pkg/scripting/README.md) and [pkg/scheduler](pkg/scheduler/README.md).

## Examples to copy
- Event-driven flow: [internal/contexts/warehouse/integration](internal/contexts/warehouse/integration).
- Handler error mapping (validation details, generic system errors): any handler under [internal/contexts](internal/contexts).
- Smoke test patterns: [test/smoke/README.md](test/smoke/README.md).

<!-- END -->
```

