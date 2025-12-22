# Test targets
test: test-core test-modules ## Run all tests (core + all modules)

test-core: ## Run core tests (domain + usecase)
	@echo " Running core tests..."
	@go test -v -race -count=1 ./internal/domain/entity/...
	@go test -v -race -count=1 ./internal/usecase/...

test-modules: test-module-posts test-module-profiles ## Run all module tests

test-module-posts: ## Run posts module tests
	@echo " Running posts module tests..."
	@go test -v -race -count=1 ./internal/modules/posts/domain/entity/...

test-module-profiles: ## Run profiles module tests
	@echo " Running profiles module tests..."
	@go test -v -race -count=1 ./internal/modules/profiles/entity/...

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

.PHONY: test test-core test-modules test-module-posts test-module-profiles test-quick test-coverage test-watch test-verbose test-list
