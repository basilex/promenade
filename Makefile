.PHONY: build run test clean fmt vet

# Build the application
build:
	@echo "Building promenade..."
	@go build -o bin/promenade ./cmd/server

# Run the application
run: build
	@echo "Starting promenade server..."
	@./bin/promenade

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -cover

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download

# Run all checks
check: fmt vet test
	@echo "All checks passed!"
