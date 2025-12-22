# Development targets

clean: ## Clean all build artifacts and temporary files
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/*
	rm -f coverage.out coverage.html
	@echo " Cleanup complete"

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

build-demos: ## Build all demo applications
	@echo "Building demo applications..."
	@mkdir -p bin
	@for demo in examples/*/; do \
		if [ -f "$$demo/main.go" ]; then \
			demo_name=$$(basename $$demo); \
			echo "  Building $$demo_name..."; \
			go build -o bin/$${demo_name} $$demo/*.go; \
		fi \
	done
	@echo "Demo builds complete in bin/ directory"

run-event-bus-demo: ## Run event bus demo (memory adapter)
	@mkdir -p bin
	@go build -o bin/event_bus_demo examples/event_bus_demo/*.go
	@./bin/event_bus_demo

run-redis-bus-demo: ## Run Redis bus demo (requires Redis running)
	@mkdir -p bin
	@go build -o bin/redis_bus_demo examples/redis_bus_demo/*.go
	@./bin/redis_bus_demo

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
