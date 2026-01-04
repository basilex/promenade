# ============================================================================
# Makefile.prod.mk - Production Deployment
# ============================================================================
# Production commands: Docker image builds, registry push, deployment
# All commands are database-agnostic and respect .promenade.workspace
# ============================================================================

.PHONY: prod docker-build docker-push docker-run swagger-generate swagger-all

# ============================================================================
# Production Runner
# ============================================================================

prod: validate-env  ## Run production server (requires ENVIRONMENT=production)
	@if [ "$(ENVIRONMENT)" != "production" ]; then \
		echo "❌ Error: 'make prod' requires ENVIRONMENT=production"; \
		echo "   Current: $(ENVIRONMENT)"; \
		echo ""; \
		echo "💡 Switch to production:"; \
		echo "   make switch-$(DATABASE_DRIVER)-prod"; \
		exit 1; \
	fi
	@echo "⚠️  Starting PRODUCTION server ($(DATABASE_DRIVER))..."
	@echo "⚠️  Make sure you know what you're doing!"
	@$(MAKE) docker-up
	@$(MAKE) migrate
	@$(MAKE) run

# ============================================================================
# Docker Production
# ============================================================================

docker-build: validate-env  ## Build production Docker image (uses DATABASE_DRIVER from workspace)
	@echo "Building production Docker image ($(DATABASE_DRIVER))..."
	docker build -t $(APP_NAME):$(VERSION) -f docker/Dockerfile .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "✓ Built: $(APP_NAME):$(VERSION)"
	@echo "✓ Tagged: $(APP_NAME):latest"

docker-push:  ## Push Docker image to registry (requires DOCKER_REGISTRY)
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "❌ Error: DOCKER_REGISTRY not set"; \
		echo "Usage: make docker-push DOCKER_REGISTRY=your-registry.com"; \
		exit 1; \
	fi
	@echo "Pushing to $(DOCKER_REGISTRY)..."
	docker tag $(APP_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION)
	docker tag $(APP_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(APP_NAME):latest
	docker push $(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(APP_NAME):latest
	@echo "✓ Pushed to registry"

docker-run: validate-env  ## Run production Docker container (uses workspace configuration)
	@echo "Running production container ($(DATABASE_DRIVER) / $(ENVIRONMENT))..."
	docker run --rm \
		-e DATABASE_DRIVER=$(DATABASE_DRIVER) \
		-e ENVIRONMENT=$(ENVIRONMENT) \
		-p 8081:8081 \
		$(APP_NAME):$(VERSION)

# ============================================================================
# API Documentation (Swagger)
# ============================================================================

swagger-generate:  ## Generate Swagger documentation
	@echo "Generating Swagger docs..."
	swag init -g cmd/api/main.go -o docs/swagger
	@echo "✓ Swagger docs generated"

swagger-all: swagger-generate  ## Generate all API documentation
	@echo "✓ All API documentation generated"

