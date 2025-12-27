# ============================================================================
# Makefile.prod.mk - Production Deployment
# ============================================================================
# Production commands: Docker image builds, registry push, deployment
# ============================================================================

.PHONY: docker-build docker-push docker-run docker-prod

# Production Docker image
docker-build:  ## Build production Docker image
	@echo "Building production Docker image..."
	docker build -t $(APP_NAME):$(VERSION) -f docker/Dockerfile .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "✓ Built: $(APP_NAME):$(VERSION)"
	@echo "✓ Tagged: $(APP_NAME):latest"

docker-push:  ## Push Docker image to registry (requires DOCKER_REGISTRY)
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "Error: DOCKER_REGISTRY not set"; \
		echo "Usage: make docker-push DOCKER_REGISTRY=your-registry.com"; \
		exit 1; \
	fi
	@echo "Pushing to $(DOCKER_REGISTRY)..."
	docker tag $(APP_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION)
	docker tag $(APP_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(APP_NAME):latest
	docker push $(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(APP_NAME):latest
	@echo "✓ Pushed to registry"

docker-run:  ## Run production Docker container
	@echo "Running production container..."
	docker run -d \
		--name $(APP_NAME) \
		-p 8080:8080 \
		-e ENVIRONMENT=production \
		--restart unless-stopped \
		$(APP_NAME):$(VERSION)
	@echo "✓ Container started on port 8080"

docker-prod: docker-build docker-run  ## Build and run production container

