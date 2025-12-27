.PHONY: help

# ============================================================================
# Promenade - Modular Makefile System
# ============================================================================
# Main Makefile - Common variables and environment configuration
# Targets are organized in separate modules:
#   - Makefile.dev.mk   → Development workflow (install, build, lint)
#   - Makefile.test.mk  → Testing infrastructure (unit, integration, coverage)
#   - Makefile.prod.mk  → Production/DevOps (docker, migrations, swagger)
# ============================================================================

# Common variables
APP_NAME=promenade
VERSION?=0.1.0
ENV?=dev
DOCKER_IMAGE_TAG=$(VERSION)-$(ENV)

# Docker Compose files per environment
DOCKER_COMPOSE_DEV=docker-compose -f docker/docker-compose.dev.yml
DOCKER_COMPOSE_TEST=docker-compose -f docker/docker-compose.test.yml
DOCKER_COMPOSE_PROD=docker-compose -f docker/docker-compose.prod.yml

# Database connection defaults (can be overridden)
# Note: Application uses config/app.{env}.yaml for runtime configuration
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= promenade
DB_SSLMODE ?= disable

DB_URL=postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATE=migrate -path migrations -database "$(DB_URL)"

# Include modular makefiles
include Makefile.dev.mk
include Makefile.test.mk
include Makefile.prod.mk

# Default target
.DEFAULT_GOAL := help

help:  ## Show this help message
	@echo "================================================================"
	@echo "          Promenade - Available Commands                        "
	@echo "================================================================"
	@echo ""
	@echo "DEV DEVELOPMENT (Makefile.dev.mk)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' Makefile.dev.mk
	@echo ""
	@echo " TESTING (Makefile.test.mk)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' Makefile.test.mk
	@echo ""
	@echo "PROD PRODUCTION (Makefile.prod.mk)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' Makefile.prod.mk
	@echo ""
	@echo "INFO Usage examples:"
	@echo "  make dev              # Start API development server"
	@echo "  make test             # Run all tests"
	@echo "  make docker-run       # Build and run in Docker"
	@echo ""
