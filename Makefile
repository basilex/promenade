.PHONY: help validate-env workspace
.PHONY: install build clean fmt lint
.PHONY: setup-hooks
.PHONY: clean-emoji clean-emoji-apply check-emoji
.PHONY: switch-postgres-dev switch-postgres-test switch-postgres-prod

# ============================================================================
# Promenade - Modular Makefile System with Workspace State Management
# ============================================================================
# Main Makefile - Workspace configuration and common entry points
# Workspace state: .promenade.workspace (DATABASE_DRIVER + ENVIRONMENT)
#
# This file contains:
#   - Workspace switchers (switch-{driver}-{env})
#   - Environment validation (validate-env, status)
#   - Help system
#
# Modules:
#   - Makefile.dev.mk   → Development workflow (dev, build, docker, migrations)
#   - Makefile.test.mk  → Testing infrastructure (test-all, test, benchmarks)
#   - Makefile.prod.mk  → Production deployment (prod, docker-build, swagger)
#
# Quick Start:
#   1. Configure workspace: make switch-postgres-dev
#   2. Start development: make dev
#   3. Run tests: make switch-postgres-test && make test-all
#
# See: docs/work-in-progress/WORKFLOW_STATE_MANAGEMENT.md
# ============================================================================

# Load workspace configuration (if exists)
-include .promenade.workspace
export

# Common variables
APP_NAME=promenade
VERSION?=0.1.0
DOCKER_IMAGE_TAG=$(VERSION)-$(DATABASE_DRIVER)-$(ENVIRONMENT)

# Docker Compose command (paths are constructed dynamically in commands)
DOCKER_COMPOSE=docker compose

# ============================================================================
# Environment Validation & Status
# ============================================================================

validate-env:  ## Validate .promenade.workspace exists and is configured
	@if [ ! -f .promenade.workspace ]; then \
		echo " .promenade.workspace not found!"; \
		echo ""; \
		echo " Quick start:"; \
		echo "   cp .promenade.workspace.example .promenade.workspace"; \
		echo "   OR"; \
		echo "   make switch-postgres-dev  # PostgreSQL + development"; \
		echo ""; \
		echo " See: docs/work-in-progress/WORKFLOW_STATE_MANAGEMENT.md"; \
		exit 1; \
	fi
	@if [ -z "$(DATABASE_DRIVER)" ]; then \
		echo " DATABASE_DRIVER not set in .promenade.workspace"; \
		exit 1; \
	fi
	@if [ -z "$(ENVIRONMENT)" ]; then \
		echo " ENVIRONMENT not set in .promenade.workspace"; \
		exit 1; \
	fi

workspace:  ## Show current workspace configuration
	@echo "=================================="
	@echo "  Promenade Workspace Status"
	@echo "=================================="
	@if [ -f .promenade.workspace ]; then \
		echo ""; \
		echo " .promenade.workspace:"; \
		cat .promenade.workspace | grep -v '^#' | grep -v '^$$'; \
		echo ""; \
		echo " Configuration loaded"; \
	else \
		echo ""; \
		echo " .promenade.workspace not found"; \
		echo " Run: make switch-postgres-dev OR make switch-sqlite-dev"; \
	fi
	@echo ""

# ============================================================================
# Go Workspace Commands (database/environment agnostic)
# ============================================================================

install:  ## Install development dependencies
	@echo "Installing development dependencies..."
	go mod download
	go mod tidy

build:  ## Build the application
	@echo "Building Promenade..."
	@mkdir -p bin
	go build -o bin/promenade ./cmd/api

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

# ============================================================================
# Git Hooks Setup
# ============================================================================

setup-hooks:  ## Install git hooks (pre-commit, pre-push)
	@echo " Installing git hooks..."
	@if [ ! -d .git ]; then \
		echo " Error: Not a git repository"; \
		exit 1; \
	fi
	@mkdir -p .git/hooks
	@cp .githooks/pre-commit .git/hooks/pre-commit
	@cp .githooks/pre-push .git/hooks/pre-push
	@chmod +x .git/hooks/pre-commit .git/hooks/pre-push
	@echo " Git hooks installed:"
	@echo "   - pre-commit: lint + unit tests"
	@echo "   - pre-push: full CI checks"
	@echo ""
	@echo " Skip hooks with: git commit --no-verify"

# ============================================================================
# Emoji Cleaning (Official Policy Enforcement)
# See: docs/guides/documentation-style-guide.md
# ============================================================================

clean-emoji:  ## Show files with emoji (dry-run, safe)
	@python3 scripts/clean-emojies.py

clean-emoji-apply:  ## Remove emoji from files (MODIFIES files)
	@python3 scripts/clean-emojies.py --apply

check-emoji:  ## Check for emoji violations (CI mode, exit 1 if found)
	@python3 scripts/clean-emojies.py --check

# ============================================================================
# Configuration Switchers (driver-environment)
# Naming: switch-{driver}-{env} matches config files app.{driver}-{env}.yaml
# Usage: Set once, then use runners (dev, test-all, prod)
# ============================================================================

# PostgreSQL configurations
switch-postgres-dev:  ## Switch to: PostgreSQL + development (→ app.postgres-dev.yaml)
	@echo "DATABASE_DRIVER=postgres" > .promenade.workspace
	@echo "ENVIRONMENT=development" >> .promenade.workspace
	@echo " Switched to: postgres + development"
	@echo " Next: make dev"

switch-postgres-test:  ## Switch to: PostgreSQL + test (→ app.postgres-test.yaml)
	@echo "DATABASE_DRIVER=postgres" > .promenade.workspace
	@echo "ENVIRONMENT=test" >> .promenade.workspace
	@echo " Switched to: postgres + test"
	@echo " Next: make test-all"

switch-postgres-prod:  ## Switch to: PostgreSQL + production (→ app.postgres-prod.yaml)
	@echo "DATABASE_DRIVER=postgres" > .promenade.workspace
	@echo "ENVIRONMENT=production" >> .promenade.workspace
	@echo "  Switched to: postgres + PRODUCTION"
	@echo "  Make sure you know what you're doing!"

# Include modular makefiles
include Makefile.dev.mk
include Makefile.test.mk
include Makefile.prod.mk

# Default target
.DEFAULT_GOAL := help

help:  ## Show this help message
	@echo "================================================================"
	@echo "   Promenade - Workspace-Based Makefile System                 "
	@echo "================================================================"
	@echo ""
	@if [ -f .promenade.workspace ]; then \
		echo " Current Workspace:"; \
		cat .promenade.workspace | grep -v '^#' | grep -v '^$$' | sed 's/^/   /'; \
		echo ""; \
	else \
		echo "  No workspace configured!"; \
		echo " Run: make switch-postgres-dev"; \
		echo ""; \
	fi
	@echo " SWITCHERS"
	@awk 'BEGIN {FS = ":.*##"} /^switch-[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}' Makefile
	@echo ""
	@echo " WORKSPACE"
	@awk 'BEGIN {FS = ":.*##"} /^(workspace|validate-env):.*##/ {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}' Makefile
	@echo ""
	@echo "  GO WORKSPACE COMMANDS"
	@awk 'BEGIN {FS = ":.*##"} /^(install|build|clean|fmt|lint):.*##/ {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}' Makefile
	@echo ""
	@echo " EMOJI CLEANING (Official Policy)"
	@awk 'BEGIN {FS = ":.*##"} /^(clean-emoji|clean-emoji-apply|check-emoji):.*##/ {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}' Makefile
	@echo ""
	@echo "  DEVELOPMENT (Makefile.dev.mk)"
	@awk 'BEGIN {FS = ":.*##"} /^(dev|dev-fresh|build|run|install|clean|fmt|lint|docker-up|docker-down|docker-logs|docker-ps|docker-clean|db-create|db-drop|db-reset|db-fresh|migrate|migrate-core|migrate-shared|migrate-identity|migrate-customer-mgmt|migrate-order-mgmt|migrate-status|migrate-new|seed|seed-shared|seed-identity|ci-check|ci-lint|ci-test|ci-build|pre-push):.*##/ {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}' Makefile.dev.mk
	@echo ""
	@echo " TESTING (Makefile.test.mk)"
	@awk 'BEGIN {FS = ":.*##"} /^test[a-zA-Z_-]*:.*##/ {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}' Makefile.test.mk
	@echo ""
	@echo " PRODUCTION (Makefile.prod.mk)"
	@awk 'BEGIN {FS = ":.*##"} /^(prod|docker-build|docker-push|docker-run|swagger-generate|swagger-all):.*##/ {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}' Makefile.prod.mk
	@echo ""
	@echo " Quick Start:"
	@echo "   make switch-postgres-dev  # Configure workspace"
	@echo "   make dev                  # Start development"
	@echo "   make workspace            # Show current config"
	@echo ""
	@echo " Documentation: docs/work-in-progress/WORKFLOW_STATE_MANAGEMENT.md"
	@echo ""
