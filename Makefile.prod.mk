# ============================================================================
# Makefile.prod.mk - Production & DevOps
# ============================================================================

.PHONY: docker-build docker-up docker-down docker-logs docker-ps docker-clean
.PHONY: migrate-core migrate-identity migrate-context migrate-status migrate-create-core migrate-create-context

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

# Migration commands (context-based, no modules)
migrate-core:  ## Run core migrations (uuid, auth, RBAC, reference data)
	@echo "Running core migrations..."
	@go run cmd/migrate/main.go --namespace core

migrate-identity:  ## Run Identity context migrations (users, contacts)
	@echo "Running Identity context migrations..."
	@go run cmd/migrate/main.go --namespace identity

migrate-context:  ## Run specific context migrations (Usage: make migrate-context CONTEXT=customer-mgmt)
	@if [ -z "$(CONTEXT)" ]; then \
		echo "Error: CONTEXT is required. Usage: make migrate-context CONTEXT=customer-mgmt"; \
		exit 1; \
	fi
	@echo "Running $(CONTEXT) context migrations..."
	@go run cmd/migrate/main.go --namespace $(CONTEXT)

migrate-status:  ## Show migration status
	@echo "Migration status:"
	@go run cmd/migrate/main.go --status

migrate-create-core:  ## Create new core migration (Usage: make migrate-create-core NAME=add_new_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create-core NAME=add_new_table"; \
		exit 1; \
	fi
	@echo "Creating core migration: $(NAME)"
	@./scripts/create-migration.sh core $(NAME)

migrate-create-context:  ## Create new context migration (Usage: make migrate-create-context CONTEXT=identity NAME=add_contacts)
	@if [ -z "$(CONTEXT)" ] || [ -z "$(NAME)" ]; then \
		echo "Error: CONTEXT and NAME are required."; \
		echo "Usage: make migrate-create-context CONTEXT=identity NAME=add_contacts"; \
		exit 1; \
	fi
	@echo "Creating $(CONTEXT) migration: $(NAME)"
	@./scripts/create-migration.sh $(CONTEXT) $(NAME)
