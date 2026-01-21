# ============================================================================
# Makefile.test.mk - Testing Infrastructure
# ============================================================================
# All test commands are database-agnostic and use .promenade.workspace
# ============================================================================

.PHONY: test-all test test-unit test-integration test-smoke test-benchmark test-benchmark-all test-coverage test-db-start test-db-stop

# ============================================================================
# Testing Runner
# ============================================================================

test-all: validate-env  ## Run all tests (works in any environment, best with test)
	@if [ "$(ENVIRONMENT)" != "test" ]; then \
		echo "  Warning: Running tests in $(ENVIRONMENT) environment"; \
		echo " For best results: make switch-$(DATABASE_DRIVER)-test"; \
		echo ""; \
	fi
	@echo " Running all tests ($(DATABASE_DRIVER) / $(ENVIRONMENT))..."
	@$(MAKE) test

# ============================================================================
# Test Commands
# ============================================================================

test: validate-env  ## Run all tests (uses DATABASE_DRIVER and ENVIRONMENT from workspace)
	@echo "Running all tests ($(DATABASE_DRIVER) / $(ENVIRONMENT))..."
	@echo "Note: Race detector disabled (causes hangs with httptest/Redis tests)"
	@echo " To run with race detector: go test -race ./pkg/aggregate ./pkg/uuidv7 ..."
	go test -v ./pkg/... ./internal/... ./cmd/...

test-unit:  ## Run only unit tests (fast, no DB, no workspace needed)
	@echo "Running unit tests..."
	go test -v -short ./...

test-smoke:  ## Run smoke tests for HTTP handlers (fast, no DB, no workspace needed)
	@echo "Running smoke tests..."
	go test -v ./test/smoke/...

test-integration: validate-env  ## Run integration tests (uses DATABASE_DRIVER from workspace, optional TEST_PATH)
	@if [ -n "$(TEST_PATH)" ]; then \
		INTEGRATION_PATH="$(TEST_PATH)"; \
	else \
		INTEGRATION_PATH="./test/integration/contexts/..."; \
	fi; \
	echo "Running integration tests for $$INTEGRATION_PATH ($(DATABASE_DRIVER))..."; \
	PATH_ARG=$$INTEGRATION_PATH; \
	if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		if [ -z "$$CI" ] && [ -z "$$GITHUB_ACTIONS" ]; then \
			echo "Test DB: PostgreSQL on localhost:5433/promenade_test"; \
			$(MAKE) test-db-start; \
		else \
			echo "CI environment: Using PostgreSQL service on localhost:5432"; \
		fi; \
	fi; \
	echo "Note: Tests run sequentially (-p 1) to prevent foreign key deadlocks"; \
	DB_HOST=$${DB_HOST:-127.0.0.1} DB_PORT=$${DB_PORT:-5433} DB_USER=$${DB_USER:-system} DB_PASSWORD=$${DB_PASSWORD:-passw0rd} DB_NAME=$${DB_NAME:-promenade_test} REDIS_ADDR=$${REDIS_ADDR:-127.0.0.1:6380} go test -v -p 1 $$PATH_ARG; \
	if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		if [ -z "$$CI" ] && [ -z "$$GITHUB_ACTIONS" ]; then \
			$(MAKE) test-db-stop; \
		fi; \
	fi

test-benchmark: validate-env  ## Run benchmark tests (uses DATABASE_DRIVER from workspace)
	@echo "Running benchmark tests ($(DATABASE_DRIVER))..."
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		if [ -z "$$CI" ] && [ -z "$$GITHUB_ACTIONS" ]; then \
			echo "Test DB: PostgreSQL on localhost:5433/promenade_test"; \
			$(MAKE) test-db-start; \
			sleep 3; \
		else \
			echo "CI environment: Using PostgreSQL service on localhost:5432"; \
		fi \
	fi
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5433} DB_USER=$${DB_USER:-system} DB_PASSWORD=$${DB_PASSWORD:-passw0rd} DB_NAME=$${DB_NAME:-promenade_test} REDIS_ADDR=$${REDIS_ADDR:-localhost:6380} go test -bench=. -benchmem -benchtime=5s ./test/benchmark/contexts/...; \
	else \
		go test -bench=. -benchmem -benchtime=5s ./test/benchmark/contexts/...; \
	fi
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		if [ -z "$$CI" ] && [ -z "$$GITHUB_ACTIONS" ]; then \
			$(MAKE) test-db-stop; \
		fi \
	fi

test-benchmark-all: validate-env  ## Run all benchmark tests with extended time
	@echo "Running extended benchmark tests ($(DATABASE_DRIVER))..."
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		if [ -z "$$CI" ] && [ -z "$$GITHUB_ACTIONS" ]; then \
			$(MAKE) test-db-start; \
			sleep 3; \
		fi \
	fi
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5433} DB_USER=$${DB_USER:-system} DB_PASSWORD=$${DB_PASSWORD:-passw0rd} DB_NAME=$${DB_NAME:-promenade_test} REDIS_ADDR=$${REDIS_ADDR:-localhost:6380} go test -bench=. -benchmem -benchtime=10s ./test/benchmark/contexts/...; \
	else \
		go test -bench=. -benchmem -benchtime=10s ./test/benchmark/contexts/...; \
	fi
	@if [ "$(DATABASE_DRIVER)" = "postgres" ]; then \
		if [ -z "$$CI" ] && [ -z "$$GITHUB_ACTIONS" ]; then \
			$(MAKE) test-db-stop; \
		fi \
	fi

test-coverage:  ## Generate test coverage report (no workspace needed)
	@echo "Generating coverage report..."
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report: coverage/coverage.html"

test-db-start:  ## Start test database (driver-aware)
	@if [ -n "$$CI" ] || [ -n "$$GITHUB_ACTIONS" ]; then \
		echo "ℹ  CI/CD environment detected - using existing database service"; \
	else \
		echo "Starting $(DATABASE_DRIVER) test database..."; \
		docker compose -f docker/docker-compose.$(DATABASE_DRIVER).test.yml down -v; \
		docker compose -f docker/docker-compose.$(DATABASE_DRIVER).test.yml up -d; \
		echo "Waiting for test database health..."; \
		ready=0; \
		for i in $$(seq 1 30); do \
			db_status=$$(docker inspect --format='{{.State.Health.Status}}' promenade-$(DATABASE_DRIVER)-test 2>/dev/null); \
			redis_status=$$(docker inspect --format='{{.State.Health.Status}}' promenade-redis-test 2>/dev/null); \
			if [ "$$db_status" = "healthy" ] && [ "$$redis_status" = "healthy" ]; then \
				ready=1; \
				break; \
			fi; \
			sleep 2; \
		done; \
		if [ "$$ready" -ne 1 ]; then \
			echo " Test database not healthy"; \
			exit 1; \
		fi; \
	fi

test-db-stop:  ## Stop test database
	@if [ -n "$$CI" ] || [ -n "$$GITHUB_ACTIONS" ]; then \
		echo "ℹ  CI/CD environment detected - skipping database stop"; \
	else \
		echo "Stopping $(DATABASE_DRIVER) test database..."; \
		docker compose -f docker/docker-compose.$(DATABASE_DRIVER).test.yml down; \
	fi
