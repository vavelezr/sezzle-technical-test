BACKEND_DIR  := backend
FRONTEND_DIR := frontend

.PHONY: install test coverage lint up down

install: ## Install frontend dependencies
	cd $(FRONTEND_DIR) && npm ci

test: ## Run backend and frontend unit tests
	$(MAKE) -C $(BACKEND_DIR) test
	cd $(FRONTEND_DIR) && npm test

coverage: ## Generate coverage reports for both layers
	$(MAKE) -C $(BACKEND_DIR) coverage
	cd $(FRONTEND_DIR) && npm run test:coverage

lint: ## Lint and type-check both layers
	$(MAKE) -C $(BACKEND_DIR) lint
	cd $(FRONTEND_DIR) && npm run lint && npm run typecheck && npm run format:check

up: ## Build and start the full stack with Docker Compose
	docker compose up --build -d

down: ## Stop the stack and remove containers
	docker compose down
