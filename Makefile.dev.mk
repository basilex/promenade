# ============================================================================
# Makefile.dev.mk - Development Workflow
# ============================================================================
# Local development commands: Docker, Database, Migrations, Running API
# ============================================================================

.PHONY: install build run clean fmt lint
.PHONY: docker-up docker-down docker-logs docker-ps docker-clean
.PHONY: db-create db-drop db-reset db-fresh
.PHONY: migrate migrate-core migrate-identity migrate-status migrate-new
.PHONY: dev dev-fresh
.PHONY: ci-check ci-lint ci-test ci-build pre-push

# Go development
install:  ## Install development dependencies
	@echo "Installing development dependencies..."
	go mod download
	go mod tidy

build:  ## Build the application
	@echo "Building Promenade..."
	@mkdir -p bin
	go build -o bin/promenade ./cmd/api

run: build  ## Build and run the application
	@echo "Starting Promenade..."
	./bin/promenade

clean:  ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/

fmt:  ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint:  ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

# Docker local development
docker-up:  ## Start local Docker containers (PostgreSQL + Redis)
	@echo "Starting Docker containers..."
	$(DOCKER_COMPOSE_DEV) up -d
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	@echo "✓ PostgreSQL ready on localhost:5432"
	@echo "✓ Redis ready on localhost:6379"

docker-down:  ## Stop Docker containers
	@echo "Stopping Docker containers..."
	$(DOCKER_COMPOSE_DEV) down

docker-logs:  ## Show Docker logs
	$(DOCKER_COMPOSE_DEV) logs -f

docker-ps:  ## Show running containers
	$(DOCKER_COMPOSE_DEV) ps

docker-clean:  ## Remove all containers and volumes (clean slate)
	@echo "⚠️  Removing containers and volumes..."
	$(DOCKER_COMPOSE_DEV) down -v
	@echo "✓ Clean slate ready"

# Database management
db-create:  ## Create database
	@echo "Creating database $(DB_NAME)..."
	@docker exec -i promenade_postgres psql -U system -d postgres -c "CREATE DATABASE $(DB_NAME);" 2>/dev/null || echo "Database already exists"

db-drop:  ## Drop database (WARNING: destructive!)
	@echo "⚠️  Dropping database $(DB_NAME)..."
	@docker exec -i promenade_postgres psql -U system -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "✓ Database dropped"

db-reset:  ## Drop and recreate database (WARNING: all data lost!)
	@echo "⚠️  Resetting database $(DB_NAME)..."
	@docker exec -i promenade_postgres psql -U system -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@docker exec -i promenade_postgres psql -U system -d postgres -c "CREATE DATABASE $(DB_NAME);"
	@echo "✓ Database reset complete"

db-fresh: db-reset migrate  ## Fresh database with all migrations
	@echo "✓ Fresh database ready!"

# Migrations - PostgreSQL
migrate-postgres:  ## Run all PostgreSQL migrations (core + all contexts)
	@echo "Running all PostgreSQL migrations..."
	@$(MAKE) migrate-postgres-core
	@$(MAKE) migrate-postgres-shared
	@$(MAKE) migrate-postgres-identity
	@$(MAKE) migrate-postgres-customer-mgmt
	@$(MAKE) migrate-postgres-order-mgmt
	@echo "✓ All PostgreSQL migrations completed"

migrate-postgres-core:  ## Run PostgreSQL core migrations (extensions, auth, RBAC)
	@echo "Running PostgreSQL core migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=core

migrate-postgres-shared:  ## Run PostgreSQL shared context migrations (reference data)
	@echo "Running PostgreSQL shared context migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=shared

migrate-postgres-identity:  ## Run PostgreSQL identity context migrations
	@echo "Running PostgreSQL identity context migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=identity

migrate-postgres-customer-mgmt:  ## Run PostgreSQL customer management migrations
	@echo "Running PostgreSQL customer management migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=customer-mgmt

migrate-postgres-order-mgmt:  ## Run PostgreSQL order management migrations
	@echo "Running PostgreSQL order management migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=order-mgmt

migrate-postgres-status:  ## Show PostgreSQL migration status
	@echo "PostgreSQL migration status:"
	@go run cmd/migrate/main.go --cmd=status

migrate-postgres-new:  ## Create new PostgreSQL migration (Usage: make migrate-postgres-new CONTEXT=core NAME=add_table)
	@if [ -z "$(CONTEXT)" ] || [ -z "$(NAME)" ]; then \
		echo "Error: CONTEXT and NAME required"; \
		echo "Usage: make migrate-postgres-new CONTEXT=core NAME=add_table"; \
		echo "       make migrate-postgres-new CONTEXT=identity NAME=add_field"; \
		exit 1; \
	fi
	@./scripts/create-migration.sh $(CONTEXT) $(NAME)

# Migrations - SQLite
migrate-sqlite:  ## Run all SQLite migrations (core + all contexts)
	@echo "Running all SQLite migrations..."
	@$(MAKE) migrate-sqlite-core
	@$(MAKE) migrate-sqlite-shared
	@$(MAKE) migrate-sqlite-identity
	@$(MAKE) migrate-sqlite-customer-mgmt
	@$(MAKE) migrate-sqlite-order-mgmt
	@echo "✓ All SQLite migrations completed"

migrate-sqlite-core:  ## Run SQLite core migrations (extensions, auth, RBAC)
	@echo "Running SQLite core migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=core

migrate-sqlite-shared:  ## Run SQLite shared context migrations (reference data)
	@echo "Running SQLite shared context migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=shared

migrate-sqlite-identity:  ## Run SQLite identity context migrations
	@echo "Running SQLite identity context migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=identity

migrate-sqlite-customer-mgmt:  ## Run SQLite customer management migrations
	@echo "Running SQLite customer management migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=customer-mgmt

migrate-sqlite-order-mgmt:  ## Run SQLite order management migrations
	@echo "Running SQLite order management migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=order-mgmt

migrate-sqlite-status:  ## Show SQLite migration status
	@echo "SQLite migration status:"
	@go run cmd/migrate/main.go --cmd=status

migrate-sqlite-new:  ## Create new SQLite migration (Usage: make migrate-sqlite-new CONTEXT=core NAME=add_table)
	@if [ -z "$(CONTEXT)" ] || [ -z "$(NAME)" ]; then \
		echo "Error: CONTEXT and NAME required"; \
		echo "Usage: make migrate-sqlite-new CONTEXT=core NAME=add_table"; \
		echo "       make migrate-sqlite-new CONTEXT=identity NAME=add_field"; \
		exit 1; \
	fi
	@./scripts/create-migration.sh $(CONTEXT) $(NAME)

# Migrations - MySQL (placeholder)
migrate-mysql:  ## Run all MySQL migrations (coming soon)
	@echo "⚠️  MySQL support coming soon"
	@exit 1

migrate-mysql-core:  ## Run MySQL core migrations (coming soon)
	@echo "⚠️  MySQL support coming soon"
	@exit 1

# Seed data
seed:  ## Seed all contexts with initial data
	@echo "Seeding database..."
	@go run cmd/seed/main.go --context=all

seed-shared:  ## Seed shared context (reference data)
	@echo "Seeding shared context..."
	@go run cmd/seed/main.go --context=shared

seed-identity:  ## Seed identity context (RBAC)
	@echo "Seeding identity context..."
	@go run cmd/seed/main.go --context=identity

# Development workflows
dev-postgres: docker-up migrate-postgres run  ## PostgreSQL development (Docker + migrations + API)

dev-postgres-fresh: docker-up db-fresh migrate-postgres run  ## PostgreSQL fresh start with clean database

dev-sqlite:  ## SQLite development (embedded, no Docker needed)
	@echo "🚀 Starting Promenade with SQLite..."
	@mkdir -p data
	@DATABASE_DRIVER=sqlite ENVIRONMENT=development ./bin/promenade || (make build && DATABASE_DRIVER=sqlite ENVIRONMENT=development ./bin/promenade)

dev-mysql:  ## MySQL development (planned - coming soon)
	@echo "⚠️  MySQL support coming soon"
	@echo "Usage: make dev-mysql (will use config/app.mysql-dev.yaml)"
	@exit 1

# CI simulation (run locally before push)
ci-check: ci-lint ci-test ci-build  ## Run all CI checks locally (lint + test + build)

ci-lint:  ## Run linters (same as CI)
	@echo "🔍 Running golangci-lint..."
	@golangci-lint run --timeout=5m || (echo "❌ Lint failed" && exit 1)
	@echo "✅ Lint passed"

ci-test:  ## Run all tests (same as CI)
	@echo "🧪 Running unit tests..."
	@make test-unit || (echo "❌ Unit tests failed" && exit 1)
	@echo "✅ Unit tests passed"
	@echo ""
	@echo "🧪 Running integration tests..."
	@make test-integration || (echo "❌ Integration tests failed" && exit 1)
	@echo "✅ Integration tests passed"
	@echo ""
	@echo "🧪 Running race detector..."
	@go test -race ./... > /dev/null 2>&1 || (echo "❌ Race detector failed" && exit 1)
	@echo "✅ Race detector passed"

ci-build:  ## Test build (same as CI)
	@echo "🔨 Testing build..."
	@make build > /dev/null || (echo "❌ Build failed" && exit 1)
	@echo "✅ Build passed"

pre-push: ci-check  ## Alias for ci-check (run before git push)
	@echo ""
	@echo "🎉 All CI checks passed! Safe to push."
