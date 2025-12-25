# Makefile Architecture

Modular Makefile system for clean separation of concerns and scalability.

## Structure

```
Makefile             (63 lines)  - Main file: variables, env loading, help
Makefile.dev.mk      (64 lines)  - Development workflow
Makefile.test.mk     (57 lines)  - Testing infrastructure
Makefile.prod.mk     (90 lines)  - Production/DevOps operations

Total:              274 lines
```

## Philosophy

**Main Makefile** - Commons only:

- Environment variables loading (`.env.development`)
- Shared variables (`APP_NAME`, `VERSION`, `DB_URL`, `MIGRATE`)
- Module includes (`include Makefile.*.mk`)
- Grouped help command (shows all modules)

**Makefile.dev.mk** - Developer workflow:

```bash
make install             # Install tools (swag, migrate, golangci-lint)
make dev                 # Start dev server (postgres + migrations + app)
make build               # Build binary (with swagger generation)
make run                 # Run compiled binary
make lint                # Run golangci-lint
make fmt                 # Format code (go fmt + gofmt -s)
make deps-update         # Update dependencies
make config-show         # Show current env configuration
```

**Makefile.test.mk** - Testing:

```bash
make test                # Unit + integration tests
make test-unit           # Unit tests only (domain + usecase)
make test-integration    # Integration tests (real DB on port 5433)
make test-smoke          # Smoke tests (end-to-end critical flows)
make test-coverage       # Generate HTML coverage report
make test-db-start       # Start test database
make test-db-stop        # Stop test database
```

**Makefile.prod.mk** - DevOps operations:

```bash
# Docker
make docker-build        # Build image (VERSION=0.1.0 ENV=dev)
make docker-run          # Build + run containers
make docker-up           # Start services
make docker-down         # Stop services
make docker-logs         # View logs
make docker-restart      # Restart containers
make docker-ps           # Show running containers
make docker-clean        # Remove containers + volumes

# Migrations
make migrate-create      # Create migration (NAME=xxx)
make migrate-up          # Apply migrations
make migrate-down        # Rollback last migration
make migrate-force       # Force version (VERSION=N)
make migrate-version     # Show current version
make migrate-status      # Show status

# Documentation
make swagger-all         # Generate v1 + v2 Swagger docs

# Cleanup
make clean               # Remove artifacts (bin/, docs/, coverage)
```

## Benefits

1. **Modularity** - Each file has single responsibility
2. **Scalability** - Easy to add `Makefile.{stage,ci,deploy}.mk`
3. **Readability** - Clear separation by context
4. **Maintainability** - Small focused files vs 188-line monolith
5. **Team-friendly** - Developers/QA/DevOps see only relevant commands

## Usage Examples

**Development:**

```bash
make help          # See all available commands
make dev           # Start development (most common)
make build         # Build for local testing
make fmt lint      # Format and lint before commit
```

**Testing:**

```bash
make test          # Run full test suite before PR
make test-unit     # Quick feedback loop during development
make test-smoke    # Verify critical flows after changes
```

**DevOps:**

```bash
make docker-run    # Deploy to local Docker
make migrate-up    # Apply database migrations
make swagger-all   # Regenerate API documentation
make clean         # Clean slate before fresh deployment
```

## Adding New Commands

1. Identify context: dev/test/prod
2. Edit appropriate `Makefile.{context}.mk`
3. Add `## Comment` for help display
4. Run `make help` to verify

Example:

```makefile
# In Makefile.dev.mk
watch: ## Watch and reload on file changes
	air -c .air.toml
```

## Variables

All shared variables are in main `Makefile`:

- `APP_NAME` - Application name
- `VERSION` - Build version (default: 0.1.0)
- `ENV` - Environment (dev/test/prod)
- `DB_URL` - PostgreSQL connection string
- `MIGRATE` - Migrate command with DB URL
- `DOCKER_COMPOSE` - Docker Compose command

Override with:

```bash
make docker-build VERSION=1.2.3 ENV=prod
make migrate-up DB_NAME=promenade_staging
```

## Migration from Old Structure

Before (188 lines monolith):

```
Makefile  ← Everything mixed together
```

After (274 lines modular):

```
Makefile          ← Common (63 lines)
Makefile.dev.mk   ← Development (64 lines)
Makefile.test.mk  ← Testing (57 lines)
Makefile.prod.mk  ← Production (90 lines)
```

**No breaking changes** - All commands work exactly as before!
