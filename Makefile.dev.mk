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
	go install github.com/vektra/mockery/v2@latest
	@echo "✅ All tools installed (including mockery)"

mocks: ## Generate all mocks for all modules (uses .mockery.yaml)
	@echo "🔧 Generating all mocks..."
	@which mockery > /dev/null || (echo "❌ mockery not found. Run: make install" && exit 1)
	mockery
	@echo "✅ All mocks generated successfully"

mocks-audit: ## Generate mocks for audit module only
	@echo "🔧 Generating audit module mocks..."
	@which mockery > /dev/null || (echo "❌ mockery not found. Run: make install" && exit 1)
	mockery --name=AuditEventRepository --dir=internal/modules/audit/domain/repository --output=internal/modules/audit/domain/repository/mocks
	@echo "✅ Audit mocks generated"

mocks-posts: ## Generate mocks for posts module only
	@echo "🔧 Generating posts module mocks..."
	@which mockery > /dev/null || (echo "❌ mockery not found. Run: make install" && exit 1)
	mockery --name='.*Repository' --dir=internal/modules/posts/domain/repository --output=internal/modules/posts/domain/repository/mocks
	@echo "✅ Posts mocks generated"

mocks-profiles: ## Generate mocks for profiles module only
	@echo "🔧 Generating profiles module mocks..."
	@which mockery > /dev/null || (echo "❌ mockery not found. Run: make install" && exit 1)
	mockery --name='.*Repository' --dir=internal/modules/profiles/domain/repository --output=internal/modules/profiles/repository/mocks
	@echo "✅ Profiles mocks generated"

mocks-analytics: ## Generate mocks for analytics module only
	@echo "🔧 Generating analytics module mocks..."
	@which mockery > /dev/null || (echo "❌ mockery not found. Run: make install" && exit 1)
	mockery --name='.*Repository' --dir=internal/modules/analytics/domain/repository --output=internal/modules/analytics/usecase/mocks
	@echo "✅ Analytics mocks generated"

mocks-verify: ## Verify all mock files are up to date
	@echo "🔍 Verifying mocks are up to date..."
	@which mockery > /dev/null || (echo "❌ mockery not found. Run: make install" && exit 1)
	@mockery --dry-run || (echo "⚠️  Mocks are out of date. Run: make mocks" && exit 1)
	@echo "✅ All mocks are up to date"

dev: ## Run in development mode
	@echo "Starting development environment..."
	$(DOCKER_COMPOSE) up -d postgres
	@echo "Waiting for database..."
	@sleep 3
	@echo "Creating database if not exists..."
	@docker exec promenade_postgres psql -U system -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$(DB_NAME)'" | grep -q 1 || \
		docker exec promenade_postgres psql -U system -d postgres -c "CREATE DATABASE $(DB_NAME);"
	@make migrate
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

config-show: ## Show current YAML configuration
	@echo "========================================="
	@echo "Configuration Files (YAML)"
	@echo "========================================="
	@echo "Environment: $(ENV) (set ENV=dev|test|prod)"
	@echo ""
	@echo "Active config file: config/app.$(ENV).yaml"
	@echo ""
	@if [ -f "config/app.$(ENV).yaml" ]; then \
		echo "--- App Configuration ---"; \
		grep -E "^[a-z_]+:|^  [a-z_]+:" config/app.$(ENV).yaml | head -20; \
	else \
		echo "Config file not found: config/app.$(ENV).yaml"; \
	fi
	@echo ""
	@echo "--- Database (from YAML) ---"
	@grep -A 8 "^database:" config/app.$(ENV).yaml || echo "Not configured"
	@echo ""
	@echo "--- Server (from YAML) ---"
	@grep -A 5 "^server:" config/app.$(ENV).yaml || echo "Not configured"
	@echo ""
	@echo "To view full config: cat config/app.$(ENV).yaml"
	@echo "========================================="
