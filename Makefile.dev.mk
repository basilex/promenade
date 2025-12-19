# Development targets

install: ## Install dependencies and tools
	go mod download
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

dev: ## Run in development mode
	@echo "Starting development environment..."
	$(DOCKER_COMPOSE) up -d postgres
	@echo "Waiting for database..."
	@sleep 3
	@echo "Creating database if not exists..."
	@docker exec promenade_postgres psql -U system -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$(DB_NAME)'" | grep -q 1 || \
		docker exec promenade_postgres psql -U system -d postgres -c "CREATE DATABASE $(DB_NAME);"
	@make migrate-up
	@echo "Starting application..."
	go run cmd/api/*.go

build: swagger-all ## Build application binary
	@echo "Building..."
	go build -o bin/$(APP_NAME) cmd/api/*.go
	@echo "Build complete: bin/$(APP_NAME)"

run: ## Run compiled binary
	./bin/$(APP_NAME)

lint: ## Run Go linter
	golangci-lint run

fmt: ## Format Go code
	go fmt ./...
	gofmt -s -w .

deps-update: ## Update Go dependencies
	go get -u ./...
	go mod tidy

generate: ## Generate entity (usage: make generate ENTITY=Product)
	@if [ -z "$(ENTITY)" ]; then \
		echo "Error: ENTITY is required. Usage: make generate ENTITY=Product"; \
		exit 1; \
	fi
	./scripts/generate.sh entity $(ENTITY)

generate-interactive: ## Interactive entity generator
	./scripts/generate-interactive.sh

gen: generate ## Alias for generate command

config-show: ## Show current configuration values
	@echo "========================================="
	@echo "Current Configuration"
	@echo "========================================="
	@echo "ENVIRONMENT:     $(ENVIRONMENT)"
	@echo "SERVER_PORT:     $(SERVER_PORT)"
	@echo "DB_HOST:         $(DB_HOST)"
	@echo "DB_PORT:         $(DB_PORT)"
	@echo "DB_NAME:         $(DB_NAME)"
	@echo "DB_USER:         $(DB_USER)"
	@echo "DB_SSLMODE:      $(DB_SSLMODE)"
	@echo "JWT_ACCESS_TTL:  $(JWT_ACCESS_TTL)"
	@echo "========================================="
