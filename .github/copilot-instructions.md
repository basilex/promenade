# Promenade AI Agent Instructions

This guide enables AI coding agents to work productively in Promenade. It summarizes architecture, workflows, and conventions unique to this project. For details, see referenced files and docs.

---

## 1. Architecture Overview

- **Clean Architecture**: Four layers—Domain (`internal/domain`), Use Case (`internal/usecase`), Adapter (`internal/adapter`), Infrastructure (`internal/infrastructure`).
- **Dependency Rule**: Inner layers never depend on outer layers. Use cases depend only on domain interfaces. Never import adapter code into use case/domain.
- **Plugin Architecture**: Business modules (`internal/modules/*`) are **completely independent vertical slices**:
  - **CRITICAL**: Modules MUST NOT import core internal packages (`internal/domain`, `internal/usecase`, `internal/adapter`). Only `pkg/*` imports allowed.
  - **Module = Domain Area**: One module can contain multiple related entities (e.g., `posts` module includes posts + comments + likes)
  - Auto-registration via `init()` in `register.go`, enabled/disabled via `config/modules.yaml`
  - Each module has own domain/entity/repository/usecase/adapter/handler/purge structure
  - Modules implement `pkg/module.Module` interface with lifecycle hooks (Initialize, Start, Stop, HealthCheck)
  - Examples: `posts` (posts+comments+likes), `profiles` (profiles+contacts), `warehouse` (commercial). See [internal/modules/README.md](internal/modules/README.md) and [docs/MODULE_INDEPENDENCE.md](docs/MODULE_INDEPENDENCE.md).
- **Core as Orchestrator**: Core provides registry systems (purge, modules, permissions) and delegates to modules via interfaces. Core never knows HOW modules work, only WHEN to call them.
- **Event-Driven**: Domain events (`internal/domain/event`) use `pkg/bus` with **dual adapters**:
  - **Memory Adapter** (`pkg/bus/memory`): In-memory Pub/Sub for dev/test/single-instance (fast, zero dependencies)
  - **Redis Adapter** (`pkg/bus/redis`): Distributed Pub/Sub for production multi-instance deployments (persistent, scalable)
  - Factory pattern with graceful fallback (Redis → Memory if Redis unavailable)
  - Events embed `bus.BaseEvent` and follow `User{Action}Event` naming. See [pkg/bus/README.md](pkg/bus/README.md).
- **Core vs Modules**: Core (`internal/domain`, `internal/usecase`) provides auth, RBAC, events, audit, reference data (countries, currencies, regions, cities, payment methods, timezones, languages) - always enabled. Modules add optional business features. See [internal/CORE.md](internal/CORE.md).
- **Module Wiring**: Core modules wire repo → usecase → handler → router in `init_*.go` files (e.g., [internal/adapter/http/v1/router/init_auth.go](internal/adapter/http/v1/router/init_auth.go)). Business modules self-wire in their `Initialize()` method.
- **No ORM**: Use raw SQL with sqlx. All repos embed `*BaseRepository` for `Get`, `Select`, `Exec`. All primary keys are UUID v7 (`pkg/uuidv7.New()`), never v4 or auto-increment.
- **Namespace-Based Migrations**: Each module has independent migration history (`migrations/{namespace}/NNNNNN_*.sql`). Core migrations run first, then enabled modules. See [docs/MIGRATION_ARCHITECTURE.md](docs/MIGRATION_ARCHITECTURE.md).
- **Automated Purge**: Modules register purge handlers via `pkg/purge.DefaultRegistry`. Cron scheduler (`internal/infrastructure/scheduler`) auto-purges soft-deleted records based on retention policies.

---

## 2. Developer Workflows

- **Makefile System**: Modular (3 files) — see [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md). Use `make help` for all commands.
- **Essential commands**:
  - `make dev` — Start Postgres, run migrations, start app in dev mode
  - `make test` — Run all tests (core + modules, 120+ tests, < 1 minute)
  - `make test-core` — Core tests only (domain + usecase)
  - `make test-modules` — All module tests
  - `make test-module-posts` / `make test-module-profiles` — Individual module tests
  - `make test-coverage` — Generate HTML coverage report
  - `make build` — Build binary (runs `swagger-all` first)
  - `make migrate-up` / `make migrate-down` — DB migrations
  - `make migrate-create MODULE=posts NAME=xxx` — Create module migration
  - `make migrate-create-core NAME=xxx` — Create core migration
  - `make swagger-all` — Generate API docs for v1 and v2
  - `make lint` / `make fmt` — Lint and format code
  - `make docker-up` / `make docker-down` — Manage Docker Compose services
- **Testing**: Tests are organized per-component (tests live alongside code). Core tests in `internal/domain/entity/*_test.go` and `internal/usecase/*_test.go`. Module tests in each module's directory (e.g., `internal/modules/posts/domain/entity/*_test.go`). Test helpers in [test/helpers/](test/helpers/). See [test/README.md](test/README.md) and [docs/TESTING_GUIDE.md](docs/TESTING_GUIDE.md).
- **Docker**: Use `make docker-build`, `make docker-run` for container workflows. See [docker/README.md](docker/README.md).

---

## 3. API & HTTP Conventions

- **Versioning**: v1 and v2 APIs are isolated (handlers, DTOs, routers). Each version has separate Swagger docs.
- **Router Structure**: Each module has its own router, registered in [internal/adapter/http/v1/router/router.go](internal/adapter/http/v1/router/router.go).
- **Middleware**: Stack includes recovery, request ID, logger, CORS. Per-route: auth and RBAC via `RequireAuth()` and `RequirePermission()`.
- **Handlers**: Accept use case in constructor. Bind request DTO, call use case, handle errors with domain-specific checks (`errors.Is`). Use `response.Success()`/`response.Error()` from `pkg/response`.
- **Response format**: Consistent JSON structure via `pkg/response`:
  - Success: `{"status":"success","data":{...}}`
  - Error: `{"status":"error","error":{"code":"VALIDATION_ERROR","message":"..."}}`
  - Pagination: `{"status":"success","data":[...],"pagination":{"total":100,"page":1,"page_size":20}}`
- **Swagger**: Add comments for API docs. Run `make swagger-all` after handler/DTO changes.

---

## 4. Data, Transactions, and Patterns

- **No ORM**: Use raw SQL with sqlx. All repos embed `*BaseRepository` for `Get`, `Select`, `Exec`.
- **Transactions**: Use `TransactionManager.WithTransaction(ctx, func(ctx) error)`; `getExecutor(ctx)` auto-selects transaction or DB.
- **UUID v7**: All primary keys use time-ordered UUIDs via `pkg/uuidv7.New()`. Never use `uuid.New()` (v4) or database auto-increment.
- **Soft Delete**: `user_posts` and `post_comments` use `deleted_at` timestamp. **CRITICAL**: Always filter `deleted_at IS NULL` in SELECT queries. See [docs/SOFT_DELETE.md](docs/SOFT_DELETE.md).
- **BaseRepository Pattern**: All repos in `internal/adapter/repository/postgres/*_repository.go` embed `*BaseRepository` which provides:

  - `Get(ctx, dest, query, args...)` - Single row
  - `Select(ctx, dest, query, args...)` - Multiple rows
  - `Exec(ctx, query, args...)` - No return
  - `getExecutor(ctx)` - Auto-selects transaction or DB connection from context

  Example:

  ```go
  func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
      var user entity.User
      query := `SELECT * FROM users WHERE id = $1`
      if err := r.Get(ctx, &user, query, id); err != nil {
          return nil, err
      }
      return &user, nil
  }
  ```

---

## 5. Adding Features

- **Manual steps**: Add entity, repo interface/impl, usecase, handler, DTO, router, migration. Register in router and main. See [README.md](README.md).
- **New Module Pattern**:

  1. Create `internal/modules/{name}/` with structure: `domain/entity/`, `domain/repository/`, `usecase/`, `adapter/http/handler/`, `adapter/repository/postgres/`
  2. Implement `pkg/module.Module` interface in `module.go`
  3. Create `register.go` with `init()` function calling `module.DefaultRegistry.Register()`
  4. Import module in `cmd/api/main.go`: `_ "github.com/basilex/promenade/internal/modules/{name}"`
  5. Create migrations in `migrations/{name}/000001_*.sql`
  6. Add config in `config/modules.yaml`

  Example `register.go`:

  ```go
  package mymodule

  import (
      "log/slog"
      "github.com/basilex/promenade/pkg/module"
  )

  func init() {
      mod := New()
      if err := module.DefaultRegistry.Register(mod); err != nil {
          slog.Error("Failed to register mymodule", "error", err)
      }
  }
  ```

---

## 6. Testing

- **Test Organization**: Tests live alongside code they test (`*_test.go` files in same directory as source). Core tests in `internal/domain/entity/` and `internal/usecase/`. Module tests in each module's directory.
- **Test Types**:
  - **Unit**: Test business logic in isolation. Mock repos using generated mocks (e.g., `usecase/mocks/`).
  - **Core Tests**: Domain entities + use case business logic (`make test-core`)
  - **Module Tests**: Each module's entities and use cases (`make test-modules` or `make test-module-posts`)
- **Test Helpers**: Available in [test/helpers/](test/helpers/) for database setup, fixtures, and common test utilities.
- **Coverage**: 120+ total tests, all passing in < 1 minute. Use `make test-coverage` for HTML report.
- **Running Tests**: Use `make test` for all tests, `make test-core` for core only, `make test-module-{name}` for specific modules.

---

## 7. Event Bus & Async

- **Dual Bus Adapters**: Factory pattern with graceful fallback
  - **Memory** (`pkg/bus/memory`): In-memory Pub/Sub for dev/test, zero dependencies, fast
  - **Redis** (`pkg/bus/redis`): Distributed Pub/Sub for production multi-instance, persistent, scalable
  - Config: Set `bus.adapter: memory` or `redis` in `config/app.{env}.yaml`, Redis auto-falls back to memory if unavailable
  - Health checks and reconnection logic built-in for Redis adapter
- **Bus configuration** (`cfg.Bus.*`): WorkerPoolSize (default: 4), BufferSize (default: 100), RetryAttempts/RetryDelay
- **Event patterns**: Events embed `bus.BaseEvent`, published via `eventBus.Publish(ctx, bus.TopicUserRegistered, event)`
- **Email Notifications**: Async via event bus subscriptions. Templates in `templates/email/`. Started in [cmd/api/main.go](cmd/api/main.go) with graceful shutdown.
- **Testing**: Both adapters have integration tests. See [docs/REDIS_BUS_TESTING.md](docs/REDIS_BUS_TESTING.md) for Redis-specific testing patterns.

---

## 8. Logging & Context

- **Logger**: Use `logger.FromContext(ctx)` for structured logs with request/user context. Never use global logger directly.
- **Context Propagation**: Context carries transaction state, logger, request ID, and user info. Always pass `ctx` through call chain.
- **Transaction Context**: `database.GetTx(ctx)` retrieves active transaction from context. `getExecutor(ctx)` in repos auto-selects transaction or DB connection.

---

## 9. Authorization & RBAC

- **Permissions**: Format `resource:action` (e.g., `posts:create`). Wildcards supported. Five system roles. See [migrations/000009_create_rbac_tables.up.sql](migrations/000009_create_rbac_tables.up.sql).

---

## 10. Configuration & Validation

- **YAML-Based Config**: Primary configuration via `config/app.{env}.yaml` (dev/test/prod). Core loads `app.{env}.yaml` based on `ENVIRONMENT` variable (defaults to "development").
- **Environment Variable Overrides**: Sensitive values (DB_PASSWORD, JWT_SECRET) can override YAML settings via `applyEnvOverrides()`.
- **Module Config**: Modules load their own config from `internal/modules/{name}/config/config.{env}.yaml`. Core does NOT load module configs - modules are autonomous.
- **Config Loading**: `config.Load()` → auto-detects environment → loads `config/app.{env}.yaml` → applies env overrides.
- **Custom Validators**: See [pkg/validator/custom_validators.go](pkg/validator/custom_validators.go).

---

, module loading

- [Makefile] + [Makefile.dev.mk] + [Makefile.test.mk] + [Makefile.prod.mk] — All workflows
- [internal/adapter/http/v1/router/init_*.go] — Core module wiring examples (auth, users, RBAC)
- [internal/infrastructure/config/yaml_config.go] — YAML config loader with env overrides
- [internal/infrastructure/database/transaction.go] — Transaction management
- [internal/adapter/repository/postgres/base_repository.go] — Base repo with Get/Select/Exec
- [internal/infrastructure/scheduler/scheduler.go] — Automated purge scheduler (cron-based)
- [pkg/module/module.go] — Module interface and registry
- [pkg/purge/registry.go] — Purge policy registry
- [internal/infrastructure/database/transaction.go] — Transactions
- [internal/adapter/repository/postgres/base_repository.go] — DB ops
- [internal/infrastructure/scheduler/scheduler.go] — Automated purge scheduler (cron-based)
- [test/README.md](test/README.md) — Test structure & helpers
- [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md) — Makefile system

---

## 12. Critical Gotchas & Debugging

### Common Pitfalls

1. **UUID v4 vs v7**: NEVER use `uuid.New()` (v4). Always use `pkg/uuidv7.New()` for time-ordered UUIDs.
2. **Soft Delete Filtering**: Always add `WHERE deleted_at IS NULL` to SELECT queries on soft-deleted tables.
3. **Module Dependencies**: Modules importing `internal/domain` or `internal/usecase` break independence. Only import `pkg/*`.
4. **Transaction Context**: Always pass `ctx` through call chain. `getExecutor(ctx)` in repos will fail without it.
5. **Logger Context**: Use `logger.FromContext(ctx)`, never global logger, to preserve request/user context.
6. **Migration Namespaces**: Core migrations (`migrations/core/`) MUST run before module migrations. Wrong namespace breaks history.
   Verify `config/app.dev.yaml` database settings.

- **Test Failures**: Run `make test-core` first (fast), then `make test-modules`. Tests are isolated - check `*_test.go` files in same directory as failing code.
- **Migration Errors**: Check `schema_migrations` table for dirty flag. Use `make migrate-down` then `make migrate-up`. Verify namespace is correct (`core/`, `posts/`, `profiles/`).
- **Module Not Loading**: Verify import in `cmd/api/main.go` and enabled in `config/modules.yaml`. Check `init()` registration in module's `register.go`.
- **Event Bus Issues**: Memory adapter is default. For Redis, set `bus.adapter: redis` in `config/app.{env}.yaml` and verify Redis is running on port 6379.
- **API 404s**: Run `make swagger-all` to regenerate routes. Check handler registration in module's router setup.
- **Config Issues**: Check `ENVIRONMENT` variable (development/test/production). Use `make config-show ENV=dev` to view loaded configte-up`. Verify namespace is correct (`core/`, `posts/`, `profiles/`).
- **Module Not Loading**: Verify import in `cmd/api/main.go` and enabled in `config/modules.yaml`. Check `init()` registration in module's `register.go`.
- **Event Bus Issues**: Memory adapter is default. For Redis, set `BUS_ADAPTER=redis` in `.env` and verify Redis is running on port 6379.
- **API 404s**: Run `make swagger-all` to regenerate routes. Check handler registration in module's router setup.

### Performance Patterns

- **Batch Operations**: Use `COPY` or bulk inserts for > 100 rows. See purge handlers for examples.
- **N+1 Queries**: Use `SELECT ... WHERE id IN (...)` with sqlx `IN` query expansion.
- **Transaction Scope**: Keep transactions short. Don't call external APIs inside `WithTransaction()`.
- **Context Timeouts**: Set explicit timeouts for long operations: `ctx, cancel := context.WithTimeout(ctx, 30*time.Second)`.

---

**For more, see:**

- [README.md](README.md)
- [docs/](docs/) for guides on testing, UUID v7, validation, and more.
- [pkg/uuidv7/uuidv7.go] — UUID v7 implementation
- [scripts/create-migration.sh] — Migration creation helper
