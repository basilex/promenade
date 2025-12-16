.PHONY: help install dev build test clean docker swagger migrate

# Load environment variables from .env.development if exists
ifneq (,$(wildcard ./.env.development))
    include .env.development
    export
endif

# Include test targets
include Makefile.test

# Variables
APP_NAME=promenade
DOCKER_COMPOSE=docker-compose -f docker/docker-compose.yml

# Database connection string (из переменных окружения или defaults)
DB_USER ?= postgres
DB_PASSWORD ? = postgres
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= promenade
DB_SSLMODE ?= disable

DB_URL=postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATE=migrate -path migrations -database "$(DB_URL)"

help:  ## Show this help message
	@echo "🚀 Promenade - Available Commands"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

install: ## Install dependencies and tools
	go mod download
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

dev: ## Run in development mode
	@echo "Starting development environment..."
	$(DOCKER_COMPOSE) up -d postgres
	@echo "Waiting for database..."
	@sleep 3
	@make migrate-up
	@echo "Starting application..."
	go run cmd/api/*.go

build: swagger-all ## Build application binary
	@echo "Building..."
	go build -o bin/$(APP_NAME) cmd/api/*.go
	@echo "Build complete:  bin/$(APP_NAME)"

run: ## Run compiled binary
	./bin/$(APP_NAME)

# Testing targets are in Makefile.test

clean: ## Clean build artifacts and containers
	rm -rf bin/
	rm -rf docs/v1/ docs/v2/
	rm -f coverage.out coverage.html
	$(DOCKER_COMPOSE) down -v

docker-up: ## Start all Docker services
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop all Docker services
	$(DOCKER_COMPOSE) down

docker-logs: ## View Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-build: ## Build Docker image
	docker build -f docker/Dockerfile -t $(APP_NAME):latest .

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required.  Usage: make migrate-create NAME=create_users"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@echo "Database:  $(DB_USER)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)"
	$(MIGRATE) up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	$(MIGRATE) down 1

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error:  VERSION is required"; \
		exit 1; \
	fi
	$(MIGRATE) force $(VERSION)

migrate-version: ## Show current migration version
	$(MIGRATE) version

migrate-status: ## Show detailed migration status
	@echo "Current migration status:"
	@$(MIGRATE) version

swagger-v1: ## Generate Swagger docs for API v1
	swag init -g cmd/api/main.go \
		--instanceName v1 \
		--parseDependency \
		--parseInternal \
		--dir . \
		--exclude "*_test.go,test,scripts,docs,bin,migrations,docker" \
		-o docs/v1

swagger-v2: ## Generate Swagger docs for API v2
	swag init -g cmd/api/main.go \
		--instanceName v2 \
		--parseDependency \
		--parseInternal \
		--dir . \
		--exclude "*_test.go,test,scripts,docs,bin,migrations,docker,internal/adapter/http/v1" \
		-o docs/v2

swagger-all: swagger-v1 swagger-v2 ## Generate all Swagger documentation
	@echo "Swagger documentation generated"

lint: ## Run Go linter
	golangci-lint run

fmt: ## Format Go code
	go fmt ./... 
	gofmt -s -w . 

deps-update: ## Update Go dependencies
	go get -u ./... 
	go mod tidy

generate: ## Generate entity (usage: make generate ENTITY=Product)
	@if [ -z "$(ENTITY)" ]; then \
		echo "Error:  ENTITY is required.  Usage: make generate ENTITY=Product"; \
		exit 1; \
	fi
	./scripts/generate.sh entity $(ENTITY)

generate-interactive: ## Interactive entity generator
	./scripts/generate-interactive.sh

gen: generate ## Alias for generate command

config-show: ## Show current configuration values
	@echo "========================================="
	@echo "Current Configuration"
	@echo "========================================="
	@echo "ENVIRONMENT:     $(ENVIRONMENT)"
	@echo "SERVER_PORT:     $(SERVER_PORT)"
	@echo "DB_HOST:         $(DB_HOST)"
	@echo "DB_PORT:         $(DB_PORT)"
	@echo "DB_NAME:         $(DB_NAME)"
	@echo "DB_USER:         $(DB_USER)"
	@echo "DB_SSLMODE:      $(DB_SSLMODE)"
	@echo "JWT_ACCESS_TTL:  $(JWT_ACCESS_TTL)"
	@echo "========================================="

. DEFAULT_GOAL := help
