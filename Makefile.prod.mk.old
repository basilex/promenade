# Production/DevOps targets

clean-docker: ## Clean Docker containers and volumes
	$(DOCKER_COMPOSE) down -v
	@echo " Docker cleanup complete"

docker-up: ## Start all Docker services
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop all Docker services
	$(DOCKER_COMPOSE) down

docker-logs: ## View Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-build: ## Build Docker image (usage: make docker-build VERSION=0.1.0 ENV=dev)
	docker build -f docker/Dockerfile -t $(APP_NAME):$(DOCKER_IMAGE_TAG) -t $(APP_NAME):latest .
	@echo "Built image: $(APP_NAME):$(DOCKER_IMAGE_TAG)"

docker-run: docker-build docker-up ## Build and run Docker containers
	@echo "[+] Promenade is running in Docker!"
	@echo "-> API Health: http://localhost:8080/api/v1/health"
	@echo "-> Swagger v1: http://localhost:8080/api/v1/docs/swagger/index.html"
	@echo "-> Swagger v2: http://localhost:8080/api/v2/docs/swagger/index.html"
	@echo ""
	@echo "View logs: make docker-logs"
	@echo "Stop: make docker-down"

docker-restart: ## Restart Docker containers
	$(DOCKER_COMPOSE) restart

docker-ps: ## Show running Docker containers
	docker ps --filter "name=promenade"

docker-clean: ## Remove containers and volumes (clean slate)
	$(DOCKER_COMPOSE) down -v
	@echo "[+] All containers and volumes removed"

# ============================================================================
# LEGACY MIGRATIONS (deprecated - use namespace-based migrations below)
# ============================================================================

migrate-legacy-up: ## [DEPRECATED] Run database migrations (legacy)
	@echo "  WARNING: Using legacy migration system"
	@echo "Database: $(DB_USER)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)"
	$(MIGRATE) up

migrate-legacy-down: ## [DEPRECATED] Rollback last migration (legacy)
	@echo "  WARNING: Using legacy migration system"
	$(MIGRATE) down 1

# ============================================================================
# NAMESPACE-BASED MIGRATIONS (current system)
# ============================================================================

migrate: ## Run all migrations (core + enabled modules)
	@echo "Running namespace-based migrations..."
	@go run cmd/migrate/main.go -cmd=up -all

migrate-core: ## Run core migrations only
	@echo "Migrating core..."
	@go run cmd/migrate/main.go -cmd=up -namespace=core

migrate-module: ## Run migrations for specific module (usage: make migrate-module MODULE=posts)
	@if [ -z "$(MODULE)" ]; then \
		echo "Error: MODULE is required. Usage: make migrate-module MODULE=posts"; \
		exit 1; \
	fi
	@echo "Migrating module: $(MODULE)"
	@go run cmd/migrate/main.go -cmd=up -namespace=$(MODULE)

migrate-rollback: ## Rollback migrations for module (usage: make migrate-rollback MODULE=posts STEPS=1)
	@if [ -z "$(MODULE)" ]; then \
		echo "Error: MODULE is required"; \
		exit 1; \
	fi
	@STEPS_VAL=$${STEPS:-1}; \
	echo "Rolling back $(MODULE) by $$STEPS_VAL steps..."; \
	go run cmd/migrate/main.go -cmd=down -namespace=$(MODULE) -steps=$$STEPS_VAL

migrate-status: ## Show migration status for all namespaces
	@go run cmd/migrate/main.go -cmd=status

migrate-version: ## Show current version for namespace (usage: make migrate-version MODULE=posts)
	@if [ -z "$(MODULE)" ]; then \
		echo "Error: MODULE is required"; \
		exit 1; \
	fi
	@go run cmd/migrate/main.go -cmd=version -namespace=$(MODULE)

migrate-create: ## Create new migration (usage: make migrate-create MODULE=posts NAME=add_views)
	@if [ -z "$(MODULE)" ] || [ -z "$(NAME)" ]; then \
		echo "Error: MODULE and NAME are required."; \
		echo "Usage: make migrate-create MODULE=posts NAME=add_views"; \
		exit 1; \
	fi
	@./scripts/create-migration.sh $(MODULE) $(NAME)

migrate-create-core: ## Create core migration (usage: make migrate-create-core NAME=add_audit)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required."; \
		echo "Usage: make migrate-create-core NAME=add_audit"; \
		exit 1; \
	fi
	@./scripts/create-migration.sh core $(NAME)

swagger-v1: ## Generate Swagger docs for API v1
	swag init -g cmd/api/main.go \
		--dir . \
		--parseInternal \
		--parseDependency \
		--instanceName v1 \
		--exclude "*_test.go,test,scripts,docs,bin,migrations,docker" \
		-o docs/v1

swagger-v2: ## Generate Swagger docs for API v2
	swag init -g cmd/api/main.go \
		--dir . \
		--parseInternal \
		--parseDependency \
		--instanceName v2 \
		--exclude "*_test.go,test,scripts,docs,bin,migrations,docker,internal/adapter/http/v1" \
		-o docs/v2

swagger-all: swagger-v1 swagger-v2 ## Generate all Swagger documentation
	@echo "Swagger documentation generated"

# Stress/Load Testing
stress-test: ## Run basic stress tests with wrk
	@echo " Running stress tests..."
	@which wrk > /dev/null || (echo "[ERROR] wrk not found. Install: brew install wrk" && exit 1)
	@chmod +x test/stress/stress_test.sh
	@./test/stress/stress_test.sh

stress-health: ## Stress test health endpoint (10k+ RPS expected)
	@echo " Stress testing health endpoint..."
	@wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health

stress-auth: ## Stress test authentication (500+ RPS expected)
	@echo " Stress testing authentication..."
	@wrk -t4 -c50 -d30s -s test/stress/scenarios/auth.lua http://localhost:8081

stress-heavy: ## Heavy stress test (find limits)
	@echo " Running heavy stress test..."
	@echo "⚠️  This will push the API to its limits"
	@wrk -t8 -c500 -d60s http://localhost:8081/api/v1/health

stress-install: ## Install wrk (macOS only)
	@echo " Installing wrk..."
	@brew install wrk
	@wrk --version
