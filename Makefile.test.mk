# Test targets
test: test-unit test-integration-all ## Run all tests (unit + integration + event bus)

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v -race -count=1 ./internal/usecase/... ./internal/domain/...

test-integration: ## Run integration tests with real database
	@echo "Running integration tests..."
	@$(MAKE) test-db-start
	@sleep 2
	@go test -v -race -count=1 ./internal/adapter/repository/postgres/...
	@$(MAKE) test-db-stop

test-integration-bus: ## Run event bus integration tests (memory + redis)
	@echo "Running event bus integration tests..."
	@$(MAKE) test-db-start
	@sleep 2
	@echo "\n=== Testing with Memory adapter ==="
	@go test -v -count=1 ./test/integration/event_bus_test.go
	@echo "\n=== Testing with Redis adapter ==="
	@go test -v -count=1 ./test/integration/redis_bus_test.go
	@$(MAKE) test-db-stop

test-integration-all: ## Run all integration tests (DB + Event Bus)
	@echo "Running all integration tests..."
	@$(MAKE) test-db-start
	@sleep 2
	@echo "\n=== Database integration tests ==="
	@go test -v -race -count=1 ./internal/adapter/repository/postgres/...
	@echo "\n=== Event Bus - Memory adapter ==="
	@go test -v -count=1 ./test/integration/event_bus_test.go
	@echo "\n=== Event Bus - Redis adapter ==="
	@go test -v -count=1 ./test/integration/redis_bus_test.go
	@$(MAKE) test-db-stop

test-smoke: ## Run smoke tests (end-to-end critical flows)
	@echo "Running smoke tests..."
	@$(MAKE) test-db-start
	@sleep 2
	@go test -v -count=1 ./test/smoke/...
	@$(MAKE) test-db-stop

test-coverage: ## Generate test coverage report
	@echo "Running tests with coverage..."
	@$(MAKE) test-db-start
	@sleep 2
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./internal/... ./pkg/...
	@$(MAKE) test-db-stop
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-db-start: ## Start test database (PostgreSQL on port 5433) and Redis (port 6380)
	@echo "Starting test database and Redis..."
	@docker-compose -f docker/docker-compose.test.yml up -d postgres redis
	@sleep 3
	@$(MAKE) migrate-test-up

test-db-stop: ## Stop and remove test database
	@echo "Stopping test database..."
	@docker-compose -f docker/docker-compose.test.yml down -v

test-db-logs: ## Show test database logs
	@docker-compose -f docker/docker-compose.test.yml logs -f postgres

migrate-test-up:
	@echo "Running test migrations..."
	@migrate -path migrations -database "postgres://system:passw0rd@localhost:5433/promenade_test?sslmode=disable" up

migrate-test-down:
	@echo "Rolling back test migrations..."
	@migrate -path migrations -database "postgres://system:passw0rd@localhost:5433/promenade_test?sslmode=disable" down

test-watch:
	@echo "Running tests in watch mode..."
	@which gotestsum > /dev/null || go install gotest.tools/gotestsum@latest
	@gotestsum --watch --format testname

.PHONY: test test-unit test-integration test-integration-bus test-integration-all test-smoke test-coverage test-db-start test-db-stop test-db-logs migrate-test-up migrate-test-down test-watch
