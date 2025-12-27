# ============================================================================
# Makefile.prod.mk - Production & DevOps
# ============================================================================

.PHONY: docker-build docker-up docker-down docker-logs docker-ps docker-clean
.PHONY: db-create db-drop db-reset db-fresh
.PHONY: migrate migrate-core migrate-identity migrate-status migrate-new

# Docker commands
docker-build:  ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(DOCKER_IMAGE_TAG) -f docker/Dockerfile .

docker-up:  ## Start Docker containers (PostgreSQL)
	@echo "Starting Docker containers..."
	$(DOCKER_COMPOSE) up -d
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	@echo "PostgreSQL ready on localhost:5432"

docker-down:  ## Stop Docker containers
	@echo "Stopping Docker containers..."
	$(DOCKER_COMPOSE) down

docker-logs:  ## Show Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-ps:  ## Show running containers
	$(DOCKER_COMPOSE) ps

docker-clean:  ## Remove all containers and volumes
	@echo "Removing containers and volumes..."
	$(DOCKER_COMPOSE) down -v
	@echo "Removing Docker images..."
	docker rmi $(APP_NAME):$(DOCKER_IMAGE_TAG) 2>/dev/null || true

# Database management commands
db-create:  ## Create database (development)
	@echo "Creating database $(DB_NAME)..."
	@docker exec -i promenade_postgres psql -U system -d postgres -c "CREATE DATABASE $(DB_NAME);" 2>/dev/null || echo "Database already exists"

db-drop:  ## Drop database (WARNING: destructive!)
	@echo "⚠️  Dropping database $(DB_NAME)..."
	@docker exec -i promenade_postgres psql -U system -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "✓ Database dropped"

db-reset:  ## Drop and recreate database (WARNING: all data will be lost!)
	@echo "⚠️  Resetting database $(DB_NAME)..."
	@docker exec -i promenade_postgres psql -U system -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@docker exec -i promenade_postgres psql -U system -d postgres -c "CREATE DATABASE $(DB_NAME);"
	@echo "✓ Database reset complete"

db-fresh: db-reset migrate  ## Fresh database with all migrations
	@echo "✓ Fresh database ready with all migrations!"

# Migration commands
migrate:  ## Run all migrations (core + identity)
	@echo "Running all migrations..."
	@$(MAKE) migrate-core
	@$(MAKE) migrate-identity
	@echo "✓ All migrations completed"

migrate-core:  ## Run core migrations only (uuid, auth, RBAC)
	@echo "Running core migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=core

migrate-identity:  ## Run identity context migrations only
	@echo "Running identity context migrations..."
	@go run cmd/migrate/main.go --cmd=up --namespace=identity

migrate-status:  ## Show migration status for all contexts
	@echo "Migration status:"
	@go run cmd/migrate/main.go --cmd=status

migrate-new:  ## Create new migration (Usage: make migrate-new CONTEXT=core NAME=add_table)
	@if [ -z "$(CONTEXT)" ] || [ -z "$(NAME)" ]; then \
		echo "Error: CONTEXT and NAME are required."; \
		echo "Usage: make migrate-new CONTEXT=core NAME=add_new_table"; \
		echo "       make migrate-new CONTEXT=identity NAME=add_contacts"; \
		exit 1; \
	fi
	@echo "Creating $(CONTEXT) migration: $(NAME)"
	@./scripts/create-migration.sh $(CONTEXT) $(NAME)
