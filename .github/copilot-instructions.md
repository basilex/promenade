# Promenade Copilot Instructions

See [README.md](README.md) and [docs/INDEX.md](docs/INDEX.md) for full context.

## Architecture & boundaries

- **Strict DDD bounded contexts** in [internal/contexts](internal/contexts); **no cross-context imports**. Communicate via Event Bus in [pkg/bus](pkg/bus/README.md).
- **Context layout** follows clean architecture layers:
  - `aggregate/` — Domain entities (e.g., `Customer`, `Invoice`) with business rules
  - `repository/` — Repository interface (domain layer)
  - `adapter/repository/postgres/` — Repository implementation (infrastructure)
  - `usecase/` — Application business logic (use cases)
  - `adapter/http/` — HTTP handlers (presentation layer)
  - `dto/` — Data transfer objects for HTTP
  - `errors.go` — Domain error constants (validation, not-found, business logic)
- **Dependency wiring** in [cmd/api/bootstrap.go](cmd/api/bootstrap.go); entry point is [cmd/api/main.go](cmd/api/main.go).
- **Modules vs contexts**: modules are technical/feature toggles in `internal/modules/`; contexts are domain boundaries with DDD patterns (see [internal/contexts/README.md](internal/contexts/README.md)).
- **Database support**: PostgreSQL 14+ only. Migrations in [migrations/postgres/](migrations/postgres/).
- **Available contexts**: identity, customer-mgmt, order-mgmt, billing, warehouse, accounting, banking, fiscal, shared, ui, scripting (see [internal/contexts/](internal/contexts/)).

## Project-specific conventions

- **IDs**: always `uuidv7.New()` from [pkg/uuidv7](pkg/uuidv7/README.md); never UUID v4.
- **Aggregates**: embed `aggregate.BaseAggregate`, call `Touch()` on modification (examples in [pkg/aggregate/README.md](pkg/aggregate/README.md)).
  ```go
  type Customer struct {
      aggregate.BaseAggregate
      Name   string
      Email  valueobject.Email
  }
  ```
- **JSON fields**: use `jsonstore.Field[T]` from [pkg/jsonstore](pkg/jsonstore/README.md); handles marshal/unmarshal automatically.
  ```go
  Tags jsonstore.Field[[]string] `db:"tags"`
  c.Tags.Set([]string{"vip", "premium"})
  ```
- **Repositories**: implement `getExecutor(ctx)` for tx propagation (check `database.GetTx(ctx)` to get active transaction or fallback to db).
  ```go
  func (r *CustomerRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
      if tx, ok := database.GetTx(ctx); ok { return tx }
      return r.db
  }
  ```
- **Value objects**: use types from [pkg/valueobject](pkg/valueobject/README.md) for emails, phones, addresses (validation built-in).
- **Error handling**: Use cases return domain error constants from `errors.go`; handlers discriminate with `errors.Is()` and return:
  - **Validation errors** → `response.BadRequest(c, err.Error())` (expose details to user)
  - **Not-found errors** → `response.NotFound(c, "resource not found")` (generic message)
  - **System errors** → `response.InternalServerError(c, "operation failed")` (hide implementation)
  - See [docs/guides/security-patterns.md](docs/guides/security-patterns.md) for the "Gold Standard Pattern."

## Critical workflows

- **Workspace state**: `.promenade.workspace` controls DATABASE_DRIVER + ENVIRONMENT.
  - Initial setup: `make switch-postgres-dev`
  - Switch environments: `make switch-postgres-{env}` (dev, test, prod)
  - Start dev server: `make dev` (hot-reload) or `make dev-fresh` (rebuild migrations)
- **Tests** are four-tiered (see [test/README.md](test/README.md)):
  - Unit tests: in-place next to code (`*_test.go`)
  - Smoke tests: `make test-smoke` (handler-only, no DB, mirror path under `test/smoke/contexts/`)
  - Integration tests: `make test-integration` (full DB, under `test/integration/contexts/`)
  - Benchmarks: `make test-benchmark`
  - Run all: `make test`
- **Migrations**: namespace-based (core, identity, customer-mgmt, etc.) in [migrations/](migrations/).
  - Run all: `make migrate`
  - Run specific: `make migrate-module MODULE=order-mgmt`
  - Rollback: `make migrate-rollback MODULE=billing STEPS=1`
  - Migrations auto-run on app startup (see [cmd/api/main.go](cmd/api/main.go)).
- **Pre-push checks**: `make pre-push` (lint, tests, build) in [Makefile.dev.mk](Makefile.dev.mk).

## Integration points & data flow

- **Event topics**: constants in [pkg/bus/topics.go](pkg/bus/topics.go). Publish after state changes; **log publish errors without failing operations** (event delivery is best-effort).
  ```go
  if err := eventBus.Publish(ctx, bus.TopicCustomerCreated, event); err != nil {
      logger.Warn("failed to publish customer.created event", "error", err)
  }
  ```
- **Cross-context communication**: use **ACLs** (Anti-Corruption Layer) in `integration/` subdirs; never import aggregates from other contexts.
- **Warehouse integration**: order events → inventory reservation/commit in [internal/contexts/warehouse/integration](internal/contexts/warehouse/integration).
- **Fiscal integration**: order event handlers registered in [cmd/api/bootstrap.go](cmd/api/bootstrap.go) and [internal/contexts/fiscal/integration](internal/contexts/fiscal/integration).
- **Health endpoints**: live in [internal/infrastructure/health](internal/infrastructure/health), wired during bootstrap.
