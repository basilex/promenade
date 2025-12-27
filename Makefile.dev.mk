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
	$(DOCKER_COMPOSE) up -d
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	@echo "✓ PostgreSQL ready on localhost:5432"
	@echo "✓ Redis ready on localhost:6379"

docker-down:  ## Stop Docker containers
	@echo "Stopping Docker containers..."
	$(DOCKER_COMPOSE) down

docker-logs:  ## Show Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-ps:  ## Show running containers
	$(DOCKER_COMPOSE) ps

docker-clean:  ## Remove all containers and volumes (clean slate)
	@echo "⚠️  Removing containers and volumes..."
	$(DOCKER_COMPOSE) down -v
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

# Migrations
migrate:  ## Run all migrations (core + identity)
	@echo "Running all migrations..."
	@$(MAKE) migrate-core
	@$(MAKE) migrate-identity
	@echo "✓ All migrations completed"

migrate-core:  ## Run core migrations (extensions, auth, RBAC)
	@echo "Running core migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=core

migrate-identity:  ## Run identity context migrations
	@echo "Running identity context migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=identity

migrate-status:  ## Show migration status
	@echo "Migration status:"
	@go run cmd/migrate/main.go --cmd=status

migrate-new:  ## Create new migration (Usage: make migrate-new CONTEXT=core NAME=add_table)
	@if [ -z "$(CONTEXT)" ] || [ -z "$(NAME)" ]; then \
		echo "Error: CONTEXT and NAME required"; \
		echo "Usage: make migrate-new CONTEXT=core NAME=add_table"; \
		echo "       make migrate-new CONTEXT=identity NAME=add_field"; \
		exit 1; \
	fi
	@./scripts/create-migration.sh $(CONTEXT) $(NAME)

# Development workflows
dev: docker-up migrate run  ## Full dev environment (Docker + migrations + API)

dev-fresh: docker-up db-fresh run  ## Fresh start with clean database

