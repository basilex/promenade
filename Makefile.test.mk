# ============================================================================
# Makefile.test.mk - Testing Infrastructure
# ============================================================================
# All test commands are database-agnostic and use .promenade.workspace
# ============================================================================

.PHONY: test-all test test-unit test-smoke test-integration test-benchmark test-benchmark-all test-coverage test-db-start test-db-stop

# ============================================================================
# Testing Runner
# ============================================================================

test-all: validate-env  ## Run all tests (works in any environment, best with test)
	@if [ "$(ENVIRONMENT)" != "test" ]; then \
		echo "⚠️  Warning: Running tests in $(ENVIRONMENT) environment"; \
		echo "💡 For best results: make switch-$(DATABASE_DRIVER)-test"; \
		echo ""; \
	fi
	@echo "🧪 Running all tests ($(DATABASE_DRIVER) / $(ENVIRONMENT))..."
	@$(MAKE) test

# ============================================================================
# Test Commands
# ============================================================================

test: validate-env  ## Run all tests (uses DATABASE_DRIVER and ENVIRONMENT from workspace)
	@echo "Running all tests ($(DATABASE_DRIVER) / $(ENVIRONMENT))..."
	go test -v -race ./...

test-unit:  ## Run only unit tests (fast, no DB, no workspace needed)
	@echo "Running unit tests..."
	go test -v -short ./...

test-smoke:  ## Run smoke tests (mock-based handlers, no DB, no workspace needed)
	@echo "Running smoke tests..."
	go test -v ./test/smoke/contexts/...

test-integration: validate-env  ## Run integration tests (uses DATABASE_DRIVER from workspace)
	@echo "Running integration tests ($(DATABASE_DRIVER))..."
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		echo "Test DB: PostgreSQL on localhost:5432/promenade_test"; \
		$(MAKE) test-db-start; \
		sleep 3; \
	else \
		echo "Test DB: SQLite (embedded, no Docker needed)"; \
	fi
	@echo "Note: Tests run sequentially (-p 1) to prevent foreign key deadlocks"
	go test -v -p 1 ./test/integration/contexts/...
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		$(MAKE) test-db-stop; \
	fi

test-benchmark: validate-env  ## Run benchmark tests (uses DATABASE_DRIVER from workspace)
	@echo "Running benchmark tests ($(DATABASE_DRIVER))..."
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		echo "Test DB: PostgreSQL on localhost:5433/promenade_test"; \
		$(MAKE) test-db-start; \
		sleep 3; \
	fi
	go test -bench=. -benchmem -benchtime=5s ./test/benchmark/contexts/...
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		$(MAKE) test-db-stop; \
	fi

test-benchmark-all: validate-env  ## Run all benchmark tests with extended time
	@echo "Running extended benchmark tests ($(DATABASE_DRIVER))..."
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		$(MAKE) test-db-start; \
		sleep 3; \
	fi
	go test -bench=. -benchmem -benchtime=10s ./test/benchmark/contexts/...
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		$(MAKE) test-db-stop; \
	fi

test-coverage:  ## Generate test coverage report (no workspace needed)
	@echo "Generating coverage report..."
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report: coverage/coverage.html"

test-db-start:  ## Start PostgreSQL test database on port 5433
	@if [ -n "$$CI" ] || [ -n "$$GITHUB_ACTIONS" ]; then \
		echo "ℹ️  CI/CD environment detected - using existing PostgreSQL service"; \
	else \
		echo "Starting PostgreSQL test database on port 5433..."; \
		docker compose -f docker/docker-compose.test.yml up -d; \
		echo "Waiting for test database..."; \
		sleep 3; \
	fi

test-db-stop:  ## Stop test database
	@if [ -n "$$CI" ] || [ -n "$$GITHUB_ACTIONS" ]; then \
		echo "ℹ️  CI/CD environment detected - skipping database stop"; \
	else \
		echo "Stopping test database..."; \
		docker compose -f docker/docker-compose.test.yml down; \
	fi
