# ============================================================================
# Makefile.dev.mk - Development Workflow
# ============================================================================

.PHONY: install build run clean fmt lint dev

install:  ## Install development dependencies
	@echo "Installing development dependencies..."
	go mod download
	go mod tidy

build:  ## Build the application
	@echo "Building Promenade..."
	@mkdir -p bin
	go build -o bin/promenade ./cmd/api

run: build  ## Build and run the application
	@echo "Starting Promenade..."
	./bin/promenade

clean:  ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/

fmt:  ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint:  ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

dev: docker-up migrate run  ## Start full development environment (Docker + migrations + API)

dev-fresh: docker-up db-fresh run  ## Fresh start with clean database
