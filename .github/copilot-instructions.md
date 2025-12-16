# Copilot Instructions for Promenade

Essential knowledge for AI coding agents working in this codebase.

## 1. Architecture Overview

**Clean Architecture with explicit layer separation:**

- **Domain** (`internal/domain`): entities and repository interfaces only
- **Use Cases** (`internal/usecase`): business logic, orchestrates repositories
- **Adapters** (`internal/adapter`): HTTP handlers (v1/v2), repository implementations (Postgres)
- **Infrastructure** (`internal/infrastructure`): config, database connections, external tools

**Critical rule**: Dependencies flow inward. Use cases depend on domain interfaces, not concrete implementations. Handlers depend on use case interfaces. Never import adapter packages into use cases or domain.

**Module initialization pattern** (see `internal/adapter/http/v1/router/init_*.go`):

```go
// Each module wires its dependencies: repo → usecase → handler → router
func InitAuthModule(db *sqlx.DB, jwt *jwt.Manager, authMw *middleware.AuthMiddleware) *AuthRouter {
    repo := postgres.NewUserRepository(db)
    sessionRepo := postgres.NewSessionRepository(db)
    usecase := usecase.NewAuthUseCase(repo, sessionRepo, jwt)
    handler := handler.NewAuthHandler(usecase)
    return NewAuthRouter(handler, authMw)
}
```

## 2. Essential Commands

```bash
make dev                    # Start postgres, migrate, run app (primary workflow)
make build                  # Build to bin/promenade (runs swagger-all first)
make test                   # Unit + integration tests
make test-integration       # Integration tests (spins up test DB on port 5433)
make test-unit              # Unit tests only
make migrate-up             # Apply migrations
make migrate-down           # Rollback last migration
make migrate-create NAME=x  # Create new migration pair
make swagger-all            # Generate v1 and v2 Swagger docs
make generate ENTITY=X      # Generate full CRUD boilerplate using templates
make generate-interactive   # Interactive entity generator with prompts
make fmt                    # Format code (go fmt + gofmt -s)
make lint                   # Run golangci-lint
```

**Test infrastructure details**: Integration tests use separate Docker Compose config (`docker/docker-compose.test.yml`) with dedicated test DB on port 5433. Tests run with `make test-integration` which starts DB, runs migrations, executes tests, tears down DB.

## 3. Database & Repository Patterns

**No ORM. Use raw SQL with sqlx.**

All repositories:

1. Embed `*BaseRepository` for common operations (`Get`, `Select`, `Exec`, `NamedExec`)
2. Use `getExecutor(ctx)` to support transactions transparently

Example from `internal/adapter/repository/postgres/user_repository.go`:

```go
type UserRepository struct {
    *BaseRepository
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    var user entity.User
    query := `SELECT id, email, name, ... FROM users WHERE email = $1`
    if err := r.Get(ctx, &user, query, email); err != nil {
        return nil, err
    }
    return &user, nil
}
```

**Transaction pattern** (see `internal/infrastructure/database/transaction.go`):

- `TransactionManager.WithTransaction(ctx, func(ctx) error)` injects `*sqlx.Tx` into context
- `BaseRepository.getExecutor(ctx)` automatically uses transaction if present, else regular DB connection
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
r.Use(middleware.RequestID())  // X-Request-ID
r.Use(middleware.Logger())     // Structured logging
r.Use(middleware.CORS())       // CORS headers
// Per-route: authMiddleware.RequireAuth() for protected endpoints
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

## 8. Configuration & Validation

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

## 9. Common Pitfalls

❌ **Don't**:

- Import adapter packages into use cases or domain (breaks Clean Architecture)
- Suggest ORM frameworks (GORM, Ent, etc.) — this project uses raw SQL by design
- Use UUID v4 or auto-increment IDs — always use `uuidv7.New()`
- Forget to run `make swagger-all` after changing handlers/DTOs
- Skip transactions for multi-step writes

✅ **Do**:

- Follow constructor naming: `NewXxxHandler`, `NewXxxUseCase`, `NewXxxRepository`
- Use BaseRepository methods (`Get`, `Select`, `Exec`) for DB operations
- Check error types with `errors.Is()` for known domain errors (e.g., `usecase.ErrEmailAlreadyExists`)
- Use module initialization pattern (see `init_auth.go`) when adding new modules
- Leverage code generators for boilerplate (`make generate-interactive`)

## 10. Key Files Reference

- `cmd/api/main.go` — Application bootstrap, wiring, server setup
- `Makefile` + `Makefile.test` — All developer workflows
- `internal/adapter/http/v1/router/init_*.go` — Module dependency wiring examples
- `internal/infrastructure/database/transaction.go` — Transaction management
- `internal/adapter/repository/postgres/base_repository.go` — Common DB operations
- `pkg/uuidv7/uuidv7.go` — UUID v7 generation
- `scripts/generate.sh` — Code generation entry point

**Documentation deep-dives**: `docs/TESTING_GUIDE.md`, `docs/UUID_V7_MIGRATION.md`, `docs/VALIDATION.md`
