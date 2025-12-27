# ============================================================================
# Makefile.test.mk - Testing Infrastructure
# ============================================================================

.PHONY: test test-unit test-integration test-coverage test-db-start test-db-stop

test:  ## Run all tests
	@echo "Running all tests..."
	go test -v -race ./...

test-unit:  ## Run only unit tests (fast, no DB)
	@echo "Running unit tests..."
	go test -v -short ./...

test-integration: test-db-start  ## Run integration tests (requires test DB)
	@echo "Running integration tests..."
	@echo "Test DB: localhost:5433/promenade_test"
	ENVIRONMENT=test go test -v ./internal/contexts/... -run TestIntegration

test-coverage:  ## Generate test coverage report
	@echo "Generating coverage report..."
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report: coverage/coverage.html"

test-db-start:  ## Start test database
	@echo "Starting test database on port 5433..."
	docker-compose -f docker/docker-compose.test.yml up -d
	@echo "Waiting for test database..."
	@sleep 3

test-db-stop:  ## Stop test database
	@echo "Stopping test database..."
	docker-compose -f docker/docker-compose.test.yml down
