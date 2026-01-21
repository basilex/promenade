# ============================================================================
# Makefile.dev.mk - Development Workflow
# ============================================================================
# Development commands with state management via .promenade.workspace
# See: docs/work-in-progress/WORKFLOW_STATE_MANAGEMENT.md
#
# Note: Go workspace commands (install, build, clean, fmt, lint) moved to
#       Main Makefile as they are database/environment agnostic.
#
# Structure:
#   1. Development Runner - dev, dev-fresh
#   2. Run Command - run (requires workspace context)
#   3. Docker - docker-up, down, logs, ps, clean
#   4. Database Management - db-create, drop, reset, fresh
#   5. Migrations - migrate, migrate-core, etc.
#   6. Seed - seed, seed-shared, seed-identity
#   7. CI Simulation - ci-check, ci-lint, ci-test, ci-build, pre-push
# ============================================================================

.PHONY: dev dev-fresh
.PHONY: run
.PHONY: docker-up docker-down docker-logs docker-ps docker-clean
.PHONY: docker-up-mssql docker-down-mssql
.PHONY: db-create db-drop db-reset db-fresh
.PHONY: migrate migrate-core migrate-all migrate-status migrate-new
.PHONY: seed seed-shared seed-identity
.PHONY: ci-check ci-lint ci-test ci-build pre-push

# ============================================================================
# Development Runner
# ============================================================================

dev: validate-env  ## Run development server (requires ENVIRONMENT=development)
	@if [ "$(ENVIRONMENT)" != "development" ]; then \
		echo "❌ Error: 'make dev' requires ENVIRONMENT=development"; \
		echo "   Current: $(ENVIRONMENT)"; \
		echo ""; \
		echo "💡 Switch to development:"; \
		echo "   make switch-$(DATABASE_DRIVER)-dev"; \
		exit 1; \
	fi
	@echo "🚀 Starting development server ($(DATABASE_DRIVER))..."
	@$(MAKE) docker-up
	@$(MAKE) migrate
	@$(MAKE) run

dev-fresh: validate-env  ## Fresh development start with clean database
	@if [ "$(ENVIRONMENT)" != "development" ]; then \
		echo "❌ Error: 'make dev-fresh' requires ENVIRONMENT=development"; \
		echo "   Current: $(ENVIRONMENT)"; \
		echo ""; \
		echo "💡 Switch to development:"; \
		echo "   make switch-$(DATABASE_DRIVER)-dev"; \
		exit 1; \
	fi
	@echo "🔄 Fresh development setup ($(DATABASE_DRIVER))..."
	@$(MAKE) docker-up
	@$(MAKE) db-fresh
	@$(MAKE) run

# ============================================================================
# Run Command (requires workspace context)
# ============================================================================

run: build validate-env  ## Build and run the application
	@echo "Starting Promenade ($(DATABASE_DRIVER) / $(ENVIRONMENT))..."
	./bin/promenade

# ============================================================================
# Docker (Database-Aware)
# ============================================================================

docker-up: validate-env  ## Start database and Redis containers
	@echo "🐳 Starting $(DATABASE_DRIVER) + Redis containers..."
	@COMPOSE_FILE=docker/docker-compose.$(DATABASE_DRIVER).$(ENVIRONMENT).yml; \
	if [ ! -f "$$COMPOSE_FILE" ]; then \
		echo "❌ Error: $$COMPOSE_FILE not found"; \
		exit 1; \
	fi; \
	$(DOCKER_COMPOSE_DEV) -f $$COMPOSE_FILE up -d
	@echo "⏳ Waiting for database to be ready..."
	@sleep 5
	@echo "✓ $(DATABASE_DRIVER) ready"
	@echo "✓ Redis ready"

docker-down: validate-env  ## Stop database containers
	@echo "🛑 Stopping $(DATABASE_DRIVER) containers..."
	@COMPOSE_FILE=docker/docker-compose.$(DATABASE_DRIVER).$(ENVIRONMENT).yml; \
	if [ -f "$$COMPOSE_FILE" ]; then \
		$(DOCKER_COMPOSE_DEV) -f $$COMPOSE_FILE down; \
	fi

docker-logs: validate-env  ## Show Docker logs
	@COMPOSE_FILE=docker/docker-compose.$(DATABASE_DRIVER).$(ENVIRONMENT).yml; \
	if [ -f "$$COMPOSE_FILE" ]; then \
		$(DOCKER_COMPOSE_DEV) -f $$COMPOSE_FILE logs -f; \
	fi

docker-ps: validate-env  ## Show running containers
	@echo "Running containers:"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

docker-clean: validate-env  ## Remove all containers and volumes (clean slate)
	@echo "⚠️  Removing containers and volumes for $(DATABASE_DRIVER)..."
	@COMPOSE_FILE=docker/docker-compose.$(DATABASE_DRIVER).$(ENVIRONMENT).yml; \
	if [ -f "$$COMPOSE_FILE" ]; then \
		$(DOCKER_COMPOSE_DEV) -f $$COMPOSE_FILE down -v; \
	fi
	@echo "✓ Clean slate ready"

# ============================================================================
# Database Management
# ============================================================================

db-create: validate-env  ## Create database
	@echo "Creating database $(DB_NAME)..."
	@docker exec -i promenade-postgres-dev psql -U system -d postgres -c "CREATE DATABASE $(DB_NAME);" 2>/dev/null || echo "Database already exists"

db-drop: validate-env  ## Drop database (WARNING: destructive!)
	@echo "⚠️  Dropping database $(DB_NAME)..."
	@docker exec -i promenade-postgres-dev psql -U system -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "✓ Database dropped"

db-reset: validate-env  ## Drop and recreate database (WARNING: all data lost!)
	@echo "⚠️  Resetting database ($(DATABASE_DRIVER))..."
	@$(MAKE) db-drop
	@$(MAKE) db-create
	@echo "✓ Database reset complete"

db-fresh: validate-env  ## Fresh database with all migrations
	@$(MAKE) db-reset
	@$(MAKE) migrate
	@echo "✓ Fresh database ready!"

# ============================================================================
# Migrations (Database-Agnostic)
# ============================================================================

migrate: validate-env  ## Run all migrations (auto-detects database driver)
	@echo "Running $(DATABASE_DRIVER) migrations ($(ENVIRONMENT))..."
	@$(MAKE) migrate-core
	@$(MAKE) migrate-shared
	@$(MAKE) migrate-identity
	@$(MAKE) migrate-customer-mgmt
	@$(MAKE) migrate-order-mgmt
	@$(MAKE) migrate-billing
	@$(MAKE) migrate-warehouse
	@$(MAKE) migrate-scripting
	@$(MAKE) migrate-fiscal
	@$(MAKE) migrate-ui
	@echo "✓ All migrations completed for $(DATABASE_DRIVER)"

migrate-core: validate-env  ## Run core migrations (extensions, auth, RBAC)
	@echo "Running core migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=core

migrate-shared: validate-env  ## Run shared context migrations (reference data)
	@echo "Running shared context migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=shared

migrate-identity: validate-env  ## Run identity context migrations
	@echo "Running identity context migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=identity

migrate-customer-mgmt: validate-env  ## Run customer management migrations
	@echo "Running customer management migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=customer-mgmt

migrate-order-mgmt: validate-env  ## Run order management migrations
	@echo "Running order management migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=order-mgmt

migrate-billing: validate-env  ## Run billing context migrations
	@echo "Running billing context migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=billing

migrate-warehouse: validate-env  ## Run warehouse context migrations
	@echo "Running warehouse context migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=warehouse

migrate-scripting: validate-env  ## Run scripting context migrations
	@echo "Running scripting context migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=scripting

migrate-fiscal: validate-env  ## Run fiscal context migrations
	@echo "Running fiscal context migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=fiscal

migrate-ui: validate-env  ## Run UI context migrations
	@echo "Running UI context migrations ($(DATABASE_DRIVER))..."
	@go run cmd/migrate/main.go --cmd=up --namespace=ui

migrate-status: validate-env  ## Show migration status
	@echo "Migration status ($(DATABASE_DRIVER)):"
	@go run cmd/migrate/main.go --cmd=status

migrate-new: validate-env  ## Create new migration (Usage: make migrate-new CONTEXT=core NAME=add_table)
	@if [ -z "$(CONTEXT)" ] || [ -z "$(NAME)" ]; then \
		echo "Error: CONTEXT and NAME required"; \
		echo "Usage: make migrate-new CONTEXT=core NAME=add_table"; \
		echo "       make migrate-new CONTEXT=identity NAME=add_field"; \
		exit 1; \
	fi
	@./scripts/create-migration.sh $(CONTEXT) $(NAME)

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

# ============================================================================
# CI Simulation (Run Locally)
# ============================================================================

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
	@echo "ℹ️  Race detector skipped (hangs with httptest/Redis tests)"
	@echo "💡 To test race conditions: go test -race ./pkg/uuidv7 ./pkg/logger ..."

ci-build:  ## Test build (same as CI)
	@echo "🔨 Testing build..."
	@make build > /dev/null || (echo "❌ Build failed" && exit 1)
	@echo "✅ Build passed"

pre-push: ci-check  ## Alias for ci-check (run before git push)
	@echo ""
	@echo "🎉 All CI checks passed! Safe to push."
