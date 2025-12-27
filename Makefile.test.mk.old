# Test targets
test: test-core test-modules ## Run all tests (core + all modules)

test-all: test test-integration ## Run all tests including integration tests

test-core: ## Run core tests (domain + usecase)
	@echo " Running core tests..."
	@go test -v -race -count=1 ./internal/domain/entity/...
	@go test -v -race -count=1 ./internal/usecase/...

test-modules: test-module-posts test-module-profiles test-module-analytics test-module-notifications test-module-billing ## Run all module tests

test-module-posts: ## Run posts module tests
	@echo " Running posts module tests..."
	@go test -v -race -count=1 ./internal/modules/posts/domain/entity/...

test-module-profiles: ## Run profiles module tests
	@echo " Running profiles module tests..."
	@go test -v -race -count=1 ./internal/modules/profiles/entity/...

test-module-analytics: ## Run analytics module tests
	@echo " Running analytics module tests..."
	@go test -v -race -count=1 ./internal/modules/analytics/domain/entity/...
	@go test -v -race -count=1 ./internal/modules/analytics/usecase/...

test-module-notifications: ## Run notifications module tests (commercial)
	@echo "🧪 Running notifications module tests..."
	@go test -v -race -count=1 ./internal/modules/notifications/domain/entity/...
	@go test -v -race -count=1 ./internal/modules/notifications/usecase/...

test-module-billing: ## Run billing module tests (commercial)
	@echo "🧪 Running billing module tests..."
	@go test -v -race -count=1 ./internal/modules/billing/domain/entity/...
	@go test -v -race -count=1 ./internal/modules/billing/usecase/...

test-quick: ## Quick test run (no race detector, faster)
	@echo " Quick test run..."
	@go test -count=1 ./internal/domain/entity/...
	@go test -count=1 ./internal/usecase/...
	@go test -count=1 ./internal/modules/posts/domain/entity/...
	@go test -count=1 ./internal/modules/profiles/entity/...

test-coverage: ## Generate test coverage report
	@echo " Running tests with coverage..."
	@go test -race -coverprofile=coverage.out -covermode=atomic \
		./internal/domain/entity/... \
		./internal/usecase/... \
		./internal/modules/posts/domain/entity/... \
		./internal/modules/profiles/entity/... \
		./pkg/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo " Coverage report generated: coverage.html"

test-watch: ## Run tests in watch mode
	@echo "  Running tests in watch mode..."
	@which gotestsum > /dev/null || go install gotest.tools/gotestsum@latest
	@gotestsum --watch --format testname

test-verbose: ## Run tests with verbose output
	@echo " Running tests with verbose output..."
	@go test -v -race -count=1 ./internal/domain/entity/...
	@go test -v -race -count=1 ./internal/usecase/...
	@go test -v -race -count=1 ./internal/modules/posts/domain/entity/...
	@go test -v -race -count=1 ./internal/modules/profiles/entity/...

test-list: ## List all test functions
	@echo " Listing all test functions..."
	@go test -list . ./internal/domain/entity/... 2>/dev/null | grep ^Test || true
	@go test -list . ./internal/usecase/... 2>/dev/null | grep ^Test || true
	@go test -list . ./internal/modules/posts/domain/entity/... 2>/dev/null | grep ^Test || true
	@go test -list . ./internal/modules/profiles/entity/... 2>/dev/null | grep ^Test || true

# Integration tests (require test database)
test-integration: test-integration-check ## Run all integration tests
	@echo "🧪 Running integration tests..."
	@go test -v -count=1 -tags=integration ./internal/adapter/repository/postgres/...
	@go test -v -count=1 -tags=integration ./internal/infrastructure/database/...
	@go test -v -count=1 -tags=integration ./internal/modules/posts/adapter/repository/postgres/...
	@go test -v -count=1 -tags=integration ./internal/modules/profiles/adapter/repository/postgres/...
	@go test -v -count=1 -tags=integration ./internal/modules/analytics/adapter/repository/postgres/...
	@echo "[OK] Integration tests completed"

test-integration-core: test-integration-check ## Run core repository integration tests
	@echo "🧪 Running core integration tests..."
	@go test -v -count=1 -tags=integration ./internal/adapter/repository/postgres/...

test-integration-posts: test-integration-check ## Run posts module integration tests
	@echo "🧪 Running posts integration tests..."
	@go test -v -count=1 -tags=integration ./internal/modules/posts/adapter/repository/postgres/...

test-integration-profiles: test-integration-check ## Run profiles module integration tests
	@echo "🧪 Running profiles integration tests..."
	@go test -v -count=1 -tags=integration ./internal/modules/profiles/adapter/repository/postgres/...

test-integration-analytics: test-integration-check ## Run analytics module integration tests
	@echo "🧪 Running analytics integration tests..."
	@go test -v -count=1 -tags=integration ./internal/modules/analytics/adapter/repository/postgres/...

test-integration-notifications: test-integration-check ## Run notifications module integration tests
	@echo "🧪 Running notifications integration tests..."
	@go test -v -count=1 -tags=integration ./internal/modules/notifications/adapter/repository/postgres/...

test-integration-check: ## Check if test database is ready
	@echo " Checking test database connection..."
	@PGPASSWORD=promenade psql -h localhost -p 5432 -U promenade -d promenade_test -c "SELECT 1" > /dev/null 2>&1 || \
		PGPASSWORD=promenade psql -h localhost -p 5433 -U promenade -d promenade_test -c "SELECT 1" > /dev/null 2>&1 || \
		(echo "[ERROR] Test database not available on port 5432 or 5433" && \
		 echo "ℹ️  Options:" && \
		 echo "   1. Run 'make docker-up' (port 5432)" && \
		 echo "   2. Use local PostgreSQL (port 5432)" && \
		 echo "   3. Run test container: docker run -d --name promenade-test-db -e POSTGRES_USER=promenade -e POSTGRES_PASSWORD=promenade -e POSTGRES_DB=promenade_test -p 5433:5432 postgres:16" && \
		 exit 1)
	@echo "[OK] Test database is ready"

test-integration-setup: ## Setup test database (one-time setup)
	@echo " Setting up test database..."
	@PGPASSWORD=promenade psql -h localhost -p 5432 -U promenade -c "CREATE DATABASE promenade_test;" 2>/dev/null || \
		PGPASSWORD=promenade psql -h localhost -p 5433 -U promenade -c "CREATE DATABASE promenade_test;" 2>/dev/null || \
		echo "ℹ️  Test database already exists or connection failed"
	@echo "[OK] Test database setup complete"

test-integration-db-start: ## Start dedicated test database container
	@echo " Starting test database container..." \
	test-integration-db-start test-integration-db-stop
	@docker run -d --name promenade-test-db \
		-e POSTGRES_USER=promenade \
		-e POSTGRES_PASSWORD=promenade \
		-e POSTGRES_DB=promenade_test \
		-p 5433:5432 \
		postgres:16 2>/dev/null || echo "ℹ️  Container already exists"
	@echo "⏳ Waiting for database to be ready..."
	@sleep 3
	@echo "[OK] Test database container started on port 5433"

test-integration-db-stop: ## Stop and remove test database container
	@echo " Stopping test database container..."
	@docker stop promenade-test-db 2>/dev/null || true
	@docker rm promenade-test-db 2>/dev/null || true
	@echo "[OK] Test database container stopped"

# Smoke tests (black-box integration tests against running API)
test-smoke: ## Run smoke tests against running API (requires API on port 8081)
	@echo " Running smoke tests..."
	@chmod +x test/smoke/smoke_test.sh
	@chmod +x test/smoke/tests/*.sh
	@./test/smoke/smoke_test.sh

test-smoke-ci: ## Run smoke tests in CI (start API, test, stop)
	@echo " Running smoke tests in CI mode..."
	@echo "  1. Starting API in background..."
	@make build > /dev/null 2>&1
	@./bin/promenade > /tmp/promenade.log 2>&1 &
	@echo "  2. Waiting for API to be ready..."
	@sleep 5
	@echo "  3. Running smoke tests..."
	@make test-smoke || (pkill promenade && exit 1)
	@echo "  4. Stopping API..."
	@pkill promenade || true
	@echo "[OK] Smoke tests completed in CI mode"

# Stress tests (performance/load tests with wrk)
test-stress: ## Run stress tests against running API (requires wrk and API on port 8081)
	@echo " Running stress tests..."
	@chmod +x test/stress/stress_test.sh
	@./test/stress/stress_test.sh

test-stress-health: ## Run stress test only for health endpoints
	@echo " Running health endpoints stress test..."
	@wrk -t4 -c100 -d30s -s test/stress/scenarios/health.lua http://localhost:8081

.PHONY: test test-all test-core test-modules test-module-posts test-module-profiles test-module-analytics test-module-notifications test-module-billing \
	test-quick test-coverage test-watch test-verbose test-list \
	test-integration test-integration-core test-integration-posts test-integration-notifications test-integration-check test-integration-setup \
	test-smoke test-smoke-ci test-stress test-stress-health
