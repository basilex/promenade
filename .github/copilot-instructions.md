# Promenade AI Agent Instructions

This guide enables AI coding agents to work productively in Promenade. It summarizes architecture, workflows, and conventions unique to this project. For details, see referenced files and docs.

---

## 1. Architecture Overview

- **Clean Architecture**: Four layers—Domain (`internal/domain`), Use Case (`internal/usecase`), Adapter (`internal/adapter`), Infrastructure (`internal/infrastructure`).
- **Dependency Rule**: Inner layers never depend on outer layers. Use cases depend only on domain interfaces. Never import adapter code into use case/domain.
- **Event-Driven**: Domain events (`internal/domain/event`) use `pkg/bus` for async communication. Events embed `bus.BaseEvent` and follow `User{Action}Event` naming. See [pkg/bus/README.md](pkg/bus/README.md).
- **Module Wiring**: Each module wires repo → usecase → handler → router. See [internal/adapter/http/v1/router/init\_\*.go](internal/adapter/http/v1/router/init_auth.go).
- **No ORM**: Use raw SQL with sqlx. All repos embed `*BaseRepository` for common ops. All primary keys are UUID v7 (`pkg/uuidv7.New()`), never v4 or auto-increment.

---

## 2. Developer Workflows

- **Makefile System**: Modular—see [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md). Use `make help` for all commands.
- **Common commands**:
  - `make dev` — Start Postgres, migrate, run app
  - `make test` — Run all tests (unit, integration, smoke)
  - `make test-integration` — Integration tests (test DB on 5433)
  - `make test-unit` — Unit tests only
  - `make build` — Build binary (runs swagger-all first)
  - `make migrate-up` / `make migrate-down` — DB migrations
  - `make generate ENTITY=X` — Generate CRUD boilerplate
  - `make swagger-all` — Generate API docs
  - `make lint` / `make fmt` — Lint and format code
- **Testing**: Integration tests use Docker Compose (`docker/docker-compose.test.yml`). See [test/README.md](test/README.md).
- **Docker**: Use `make docker-up`, `make docker-build`, `make docker-run` for container workflows. See [docker/README.md](docker/README.md).

---

## 3. API & HTTP Conventions

- **Versioning**: v1 and v2 APIs are isolated (handlers, DTOs, routers).
- **Router Structure**: Each module has its own router, registered in [internal/adapter/http/v1/router/router.go](internal/adapter/http/v1/router/router.go).
- **Middleware**: Stack includes recovery, request ID, logger, CORS. Per-route: auth and RBAC via `RequireAuth()` and `RequirePermission()`.
- **Handlers**: Accept use case in constructor. Bind request DTO, call use case, handle errors with domain-specific checks (`errors.Is`). Use `response.Success()`/`response.Error()`.
- **Swagger**: Add comments for API docs. Run `make swagger-all` after handler/DTO changes.

---

## 4. Data, Transactions, and Patterns

- **No ORM**: Use raw SQL with sqlx. All repos embed `*BaseRepository` for `Get`, `Select`, `Exec`.
- **Transactions**: Use `TransactionManager.WithTransaction(ctx, func(ctx) error)`; `getExecutor(ctx)` auto-selects transaction or DB.
- **UUID v7**: All primary keys use time-ordered UUIDs via `pkg/uuidv7.New()`.
- **Soft Delete**: `user_posts` and `post_comments` use `deleted_at` timestamp. **CRITICAL**: Always filter `deleted_at IS NULL` in SELECT queries. See [docs/SOFT_DELETE.md](docs/SOFT_DELETE.md).

---

## 5. Adding Features

- **Preferred**: Use code generators (`make generate ENTITY=X` or `make generate-interactive`).
- **Manual steps**: Add entity, repo interface/impl, usecase, handler, DTO, router, migration. Register in router and main. See [README.md](README.md) and [scripts/templates/].

---

## 6. Testing

- **Integration**: Real Postgres, helpers in [test/helpers/database.go](test/helpers/database.go), [test/helpers/fixtures.go](test/helpers/fixtures.go).
- **Unit**: Mock repos, test business logic only.
- **Smoke**: End-to-end flows in [test/smoke/](test/smoke/).
- **Test DB**: Runs on port 5433, managed by `docker-compose.test.yml`.

---

## 7. Event Bus & Async

- **Event Bus**: In-memory pub/sub (`pkg/bus/memory`), configurable worker pool, buffer, retries. See [pkg/bus/README.md](pkg/bus/README.md).
- **Email Notifications**: Async via event bus. Templates in [templates/email/].

---

## 8. Logging & Context

- **Logger**: Use `logger.FromContext(ctx)` for structured logs with request/user context. See [pkg/logger/].

---

## 9. Authorization & RBAC

- **Permissions**: Format `resource:action` (e.g., `posts:create`). Wildcards supported. Five system roles. See [migrations/000009_create_rbac_tables.up.sql](migrations/000009_create_rbac_tables.up.sql).

---

## 10. Configuration & Validation

- **Env Loading**: Priority—`.env.{env}.local`, `.env.{env}`, `.env.local`, `.env`.
- **Custom Validators**: See [pkg/validator/custom_validators.go](pkg/validator/custom_validators.go).

---

## 11. Key Files & References

- [cmd/api/main.go](cmd/api/main.go) — App bootstrap, wiring
- [Makefile] — All workflows
- [internal/adapter/http/v1/router/init_*.go] — Module wiring
- [internal/infrastructure/database/transaction.go] — Transactions
- [internal/adapter/repository/postgres/base_repository.go] — DB ops
- [test/README.md](test/README.md) — Test structure & helpers
- [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md) — Makefile system

---

**For more, see:**

- [README.md](README.md)
- [docs/](docs/) for guides on testing, UUID v7, validation, and more.
- [pkg/uuidv7/uuidv7.go] — UUID v7
- [scripts/generate.sh] — Code generation

---

## 12. Common Pitfalls

- Never import adapter code into usecase/domain
- Never use ORM (GORM, Ent, etc.)
- Never use UUID v4 or auto-increment
- Always run `make swagger-all` after handler/DTO changes
- Always use transactions for multi-step writes

---

For deep-dives, see [docs/](docs/) for guides on testing, UUID v7, validation, and more.

- `database.GetTx(ctx)` retrieves transaction from context if needed

**UUID v7 for all primary keys**: Use `pkg/uuidv7.New()` for ID generation. Never use UUID v4 or auto-increment. UUIDs are time-ordered for optimal B-tree performance (2x faster inserts than v4).

## 4. HTTP Layer Conventions

**API versioning**: v1 and v2 coexist. Each has isolated handlers, DTOs, routers.

- v1: `internal/adapter/http/v1/{handler,dto,router}`
- v2: `internal/adapter/http/v2/{handler,dto,router,mapper}`

**Router structure**:

- Each module has its own router (e.g., `AuthRouter`, `CountryRouter`)
- Module routers registered in `V1Router.Setup()` (see `internal/adapter/http/v1/router/router.go`)
- Routers are created by `Init*Module()` functions in `cmd/api/main.go`

**Middleware stack** (applied in `cmd/api/main.go`):

```go
r.Use(middleware.Recovery())   // Panic recovery
r.Use(middleware.RequestID())  // X-Request-ID for tracing
r.Use(middleware.Logger())     // Structured logging with slog
r.Use(middleware.CORS())       // CORS headers
// Per-route: authMiddleware.RequireAuth() for protected endpoints
// Per-route: authzMiddleware.RequirePermission("resource:action") for RBAC
```

**Handler pattern**:

- Accept use case in constructor: `func NewAuthHandler(authUC usecase.AuthUseCase) *AuthHandler`
- Bind request DTO: `c.ShouldBindJSON(&req)`
- Call use case, handle errors with domain-specific checks using `errors.Is()`
- Use `response.Success()` or `response.Error()` from `internal/adapter/http/shared/response`
- Add Swagger comments for each endpoint

**Error handling pattern** (see `internal/adapter/http/v1/handler/auth_handler.go`):

```go
user, err := h.authUseCase.Register(ctx, req.Email, req.Name, req.Password)
if err != nil {
    if errors.Is(err, usecase.ErrEmailAlreadyExists) {
        response.Error(c, http.StatusConflict, "email already exists", err)
        return
    }
    response.Error(c, http.StatusInternalServerError, "failed to register user", err)
    return
}
response.Success(c, http.StatusCreated, dto.ToUserResponse(user))
```

**Use case error definitions**: Define domain-specific errors in use case files (e.g., `usecase.ErrEmailAlreadyExists`, `usecase.ErrInvalidCredentials`) rather than using generic `entity.ErrNotFound`

## 5. Adding New Features

**PREFERRED: Use code generators** (`scripts/generate.sh` or `make generate-interactive`):

- Templates in `scripts/templates/`: entity, usecase, repository (interface + impl), handler, dto, migrations
- Generates complete CRUD boilerplate with correct naming and structure
- Example: `make generate ENTITY=Product` creates entity, repo interface/impl, usecase, handler, DTOs, migrations

**Manual step-by-step** (if not using generator):

1. Domain entity: `internal/domain/entity/xxx.go`
2. Repository interface: `internal/domain/repository/xxx_repository.go`
3. Repository implementation: `internal/adapter/repository/postgres/xxx_repository.go` (embed `BaseRepository`)
4. Use case: `internal/usecase/xxx_usecase.go` with `NewXxxUseCase(repo XxxRepository) XxxUseCase`
5. Handler: `internal/adapter/http/v1/handler/xxx_handler.go` with Swagger comments
6. DTOs: `internal/adapter/http/v1/dto/xxx.go`
7. Router: `internal/adapter/http/v1/router/xxx_router.go` and `init_xxx.go`
8. Register in `router.go` and `cmd/api/main.go`
9. Migration: `make migrate-create NAME=create_xxx_table`
10. Regenerate docs: `make swagger-all`

## 6. Testing Conventions

**Integration tests** (`*_test.go` in `internal/adapter/repository/postgres/`):

- Use test helpers from `test/helpers/database.go` and `test/helpers/fixtures.go`
- Tests run against real Postgres on port 5433 (started by `make test-integration`)
- Each test should clean up or use isolated data

**Unit tests** (`*_test.go` in `internal/usecase/` and `internal/domain/`):

- Mock repositories using interfaces
- Test business logic without DB

## 7. Critical Dependencies

- **sqlx**: Query wrapper over database/sql, NOT an ORM
- **Gin**: HTTP framework (production mode set in config)
- **JWT**: `pkg/jwt/jwt.go` manages access/refresh tokens
- **validator/v10**: Request validation with custom validators in `pkg/validator/custom_validators.go`
- **golang-migrate**: Database migrations (install via `make install`)
- **swaggo/swag**: Swagger generation from comments
- **slog**: Structured logging with context fields (request_id, user_id) via `pkg/logger`
- **Event Bus**: In-memory pub/sub (`pkg/bus/memory`) for domain events, configured via `cfg.Bus.*` settings

## 8. Event Bus & Async Communication

**Event-driven architecture** using in-memory event bus (`pkg/bus`):

- Events embed `bus.BaseEvent` with `EventType`, `EventID`, `AggregateID`, `OccurredTime`
- Domain events in `internal/domain/event/` follow naming: `User{Action}Event` (e.g., `UserRegisteredEvent`)
- Use cases publish events: `eventBus.Publish(ctx, bus.TopicUserRegistered, event)`
- Services subscribe to events: `emailService.Subscribe()` listens to user events for notifications

**Email notifications** (`internal/infrastructure/notification/email_service.go`):

- Async via event bus subscriptions (non-blocking)
- Templates in `templates/email/*.html`, fallback templates for tests
- Mock sender for development, swap for SMTP/SendGrid in production
- Started in `main.go`: `emailService.Start(ctx)`, graceful shutdown on exit

**Bus configuration** (`cfg.Bus.*`):

- `WorkerPoolSize`: concurrent event handlers (default: 4)
- `BufferSize`: event queue capacity (default: 100)
- `RetryAttempts`, `RetryDelay`: automatic retry for failed handlers

## 9. Structured Logging & Context

**Logger patterns** (`pkg/logger`):

- Use `logger.FromContext(ctx)` to get logger with request/user context automatically
- Context keys: `RequestIDKey`, `UserIDKey`, `TraceIDKey` extracted from context
- Init in `main.go`: format (json/text), level (debug/info/warn/error), source file trimming
- Middleware adds `request_id` via `RequestID()`, handlers can add `user_id`

**Logging examples**:

```go
// In handlers - use package-level logger
logger.Info("User registered", slog.String("email", user.Email), slog.String("user_id", user.ID.String()))

// With context - automatically includes request_id, user_id if present
log := logger.FromContext(ctx)
log.Error("Failed to save", slog.Any("error", err))
```

## 10. Authorization & RBAC

**Permission-based access control** (`internal/adapter/http/shared/middleware/authorization.go`):

- Format: `resource:action` (e.g., `posts:create`, `users:ban`)
- Wildcard support: `posts:*` (any action on posts), `*:read` (read any resource), `*:*` (superadmin)
- Applied per-route after `authMiddleware.RequireAuth()`

**Middleware methods**:

- `RequirePermission("posts:create")` - single permission check
- `RequireAnyPermission("posts:update", "posts:*")` - at least one permission
- `RequireAllPermissions("posts:read", "comments:read")` - all permissions required

**5 system roles** (see `migrations/000009_create_rbac_tables.up.sql`): Superadmin, Admin, Moderator, Creator, User

## 11. Configuration & Validation

**Environment loading order** (highest priority first):

1. `.env.{environment}.local` (e.g., `.env.development.local`)
2. `.env.{environment}` (e.g., `.env.development`)
3. `.env.local` (gitignored, for local overrides)
4. `.env` (tracked, default values)

Config struct: `internal/infrastructure/config/config.go` with defaults and env var mapping.

**Custom validators** (`pkg/validator/custom_validators.go`):

- Register in `cmd/api/main.go` before starting server
- Available validators: `uppercase`, `alpha`, `alphanum`, `no_special`
- Example usage in DTOs: `Code string \`json:"code" binding:"required,uppercase,len=3"\``

## 12. Common Pitfalls

[X] **Don't**:

- Import adapter packages into use cases or domain (breaks Clean Architecture)
- Suggest ORM frameworks (GORM, Ent, etc.) — this project uses raw SQL by design
- Use UUID v4 or auto-increment IDs — always use `uuidv7.New()`
- Forget to run `make swagger-all` after changing handlers/DTOs
- Skip transactions for multi-step writes
- **FORGET `deleted_at IS NULL` in SELECT queries** — soft-deleted records must be filtered (see [docs/SOFT_DELETE.md](docs/SOFT_DELETE.md))

[+] **Do**:

- Follow constructor naming: `NewXxxHandler`, `NewXxxUseCase`, `NewXxxRepository`
- Use BaseRepository methods (`Get`, `Select`, `Exec`) for DB operations
- Check error types with `errors.Is()` for known domain errors (e.g., `usecase.ErrEmailAlreadyExists`)
- Use module initialization pattern (see `init_auth.go`) when adding new modules
- Leverage code generators for boilerplate (`make generate-interactive`)
- **Always add `WHERE deleted_at IS NULL`** for entities with soft delete support

## 13. Key Files Reference

- `cmd/api/main.go` — Application bootstrap, wiring, server setup
- `Makefile` + `Makefile.test` + `Makefile.prod` — All developer workflows
- `internal/adapter/http/v1/router/init_*.go` — Module dependency wiring examples
- `internal/infrastructure/database/transaction.go` — Transaction management
- `internal/adapter/repository/postgres/base_repository.go` — Common DB operations
- `pkg/uuidv7/uuidv7.go` — UUID v7 generation
- `scripts/generate.sh` — Code generation entry point

**Documentation deep-dives**: `docs/TESTING_GUIDE.md`, `docs/UUID_V7_MIGRATION.md`, `docs/VALIDATION.md`
