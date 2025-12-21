# Production/DevOps targets

clean-docker: ## Clean Docker containers and volumes
	$(DOCKER_COMPOSE) down -v
	@echo "✓ Docker cleanup complete"

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

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=create_users"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@echo "Database: $(DB_USER)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)"
	$(MIGRATE) up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	$(MIGRATE) down 1

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required"; \
		exit 1; \
	fi
	$(MIGRATE) force $(VERSION)

migrate-version: ## Show current migration version
	$(MIGRATE) version

migrate-status: ## Show detailed migration status
	@echo "Current migration status:"
	@$(MIGRATE) version

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
