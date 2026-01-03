# ============================================================================
# Makefile.test.mk - Testing Infrastructure
# ============================================================================

.PHONY: test test-unit test-smoke test-integration test-benchmark test-benchmark-all test-coverage test-db-start test-db-stop

test:  ## Run all tests
	@echo "Running all tests..."
	go test -v -race ./...

test-unit:  ## Run only unit tests (fast, no DB)
	@echo "Running unit tests..."
	go test -v -short ./...

test-smoke:  ## Run smoke tests (mock-based handlers, no DB)
	@echo "Running smoke tests..."
	go test -v ./test/smoke/contexts/...

test-integration:  ## Run integration tests (requires test DB)
	@echo "Running integration tests..."
	@echo "Test DB: localhost:5432/promenade_test (or 5433 if using test-db-start)"
	@echo "Note: Tests run sequentially (-p 1) to prevent foreign key deadlocks"
	DATABASE_DRIVER=postgres ENVIRONMENT=test go test -v -p 1 ./test/integration/contexts/...

test-benchmark:  ## Run benchmark tests (requires test DB)
	@echo "Running benchmark tests..."
	@echo "Test DB: localhost:5433/promenade_test"
	@$(MAKE) test-db-start
	@sleep 3
	go test -bench=. -benchmem -benchtime=5s ./test/benchmark/contexts/...
	@$(MAKE) test-db-stop

test-benchmark-all:  ## Run all benchmark tests with extended time
	@echo "Running extended benchmark tests..."
	@echo "Test DB: localhost:5433/promenade_test"
	@$(MAKE) test-db-start
	@sleep 3
	go test -bench=. -benchmem -benchtime=10s ./test/benchmark/contexts/...
	@$(MAKE) test-db-stop

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
