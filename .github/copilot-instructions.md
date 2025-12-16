# Copilot / AI Agent Instructions for Promenade

Short actionable notes to help code-generating agents be productive immediately.

1. Project overview

- Architecture: Clean Architecture (Domain, Usecase, Adapter, Infrastructure).
- Entry point: `cmd/api/main.go` — initializes config, DB, JWT, repos, usecases, handlers, and middleware.
- HTTP adapters: `internal/adapter/http/v1` (and `v2`) contain `handler`, `dto`, and `router` packages.
- Repositories: interfaces in `internal/domain/repository`, implementations in `internal/adapter/repository/postgres` (uses `sqlx`, no ORM).
- Infrastructure: DB and config loaders in `internal/infrastructure/database` and `internal/infrastructure/config`.

2. How to build / run / test (explicit commands)

- Development run: `make dev` — starts Postgres via Docker Compose, runs migrations, then `go run cmd/api/*.go`.
- Build binary: `make build` → `bin/promenade`.
- Docker compose: `make docker-up`, `make docker-down`, `make docker-build`.
- Migrations: `make migrate-up`, `make migrate-down`, `make migrate-create NAME=...`.
- Tests: `make test` (all), `make test-unit`, `make test-integration` (starts postgres via docker-compose).
- Swagger: `make swagger-all` (generates docs to `docs/v1` and `docs/v2`).

3. Common code patterns and conventions (be explicit)

- Versioning: Add new endpoints under `internal/adapter/http/v1` or `v2`. Each version has its own DTOs and handlers.
- Handlers -> Usecases -> Repositories: HTTP handlers call usecases in `internal/usecase`. Usecases depend on repository interfaces from `internal/domain/repository`.
- Repository implementations: `internal/adapter/repository/postgres/*_repository.go` use raw SQL with `sqlx` and explicit transactions via `internal/infrastructure/database/transaction.go`.
- DTOs live next to handlers in `internal/adapter/http/v*/dto` and map to domain entities in `internal/adapter/*/mapper` (or use mappers in new code when needed).
- Middleware: common middleware lives in `internal/adapter/http/shared/middleware` — use `RequestID()`, `Logger()`, `Recovery()`, `CORS()` and `Auth` middleware consistently.
- JWT: `pkg/jwt/jwt.go` provides the JWT manager used in `main.go`.

4. Adding a new API resource (step-by-step)

- Add domain entity in `internal/domain/entity` and repository interface in `internal/domain/repository`.
- Implement repository in `internal/adapter/repository/postgres` using `sqlx` and transaction manager if needed.
- Add usecase in `internal/usecase` and expose constructor `NewXxxUseCase(...)`.
- Add DTOs and handler in `internal/adapter/http/v1` (and v2 if needed). Register handler in the router (see `internal/adapter/http/v1/router/router.go`).
- Add swagger comments in handler or DTOs and run `make swagger-all`.

5. Templates & generators

- The project includes generation scripts: `scripts/generate.sh` and `scripts/generate-interactive.sh` and templates in `scripts/templates/` (e.g. `usecase.tmpl`). Prefer using those to keep files consistent.

6. Environment & config

- Environment files: `.env`, `.env.development`, `.env.local` (local overrides). `Makefile` will include `.env.development` if present.
- Config loader: `internal/infrastructure/config/config.go` — use it for default values and to understand required env vars.

7. Linting / formatting / dependencies

- Format: `make fmt` (runs `go fmt` / `gofmt`).
- Lint: `make lint` (uses `golangci-lint`).
- Dependency management: `go mod` + `make install` or `make deps-update` for updates.

8. Files to inspect for implementation detail (quick links)

- `cmd/api/main.go` — app bootstrap and wiring
- `Makefile` — developer workflows and shortcuts
- `internal/usecase` — business rules
- `internal/adapter/http/v1` — handlers, dto, router
- `internal/adapter/repository/postgres` — SQL implementations
- `internal/infrastructure/database` — connection and transactions
- `pkg/jwt/jwt.go` — JWT manager

9. Do not assume

- There is no ORM; code uses `sqlx` and raw SQL — avoid suggesting ORMs.
- Tests expect Docker Postgres for integration tests — use `make test-integration` or spin up Postgres before running them.

10. When proposing code changes

- Keep layering: handlers → usecases → repo implementations. Avoid cross-layer imports (e.g., usecase should not import adapter packages).
- Follow existing naming conventions: `NewXxxHandler`, `NewXxxUseCase`, `NewXxxRepository` constructors.
- Use existing generator templates when adding common boilerplate.

If anything is unclear or you want more examples (e.g., an example endpoint patch), tell me which area to expand.
