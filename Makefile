.PHONY: help run build test clean docker-up docker-down migrate-up migrate-down

# Variables
DOCKER_COMPOSE := docker-compose
GO := go
SERVICES := gateway auth
BIN_DIR := bin

# Ensure bin directory exists
$(shell mkdir -p $(BIN_DIR))

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run:
	@echo "Starting services..."
	@for service in $(SERVICES); do \
		echo "Starting $$service..."; \
		$(GO) run cmd/$$service/main.go & \
	done; \
	wait

run/%: ## Run specific service (e.g., make run/auth)
	$(GO) run cmd/$*/main.go

build: ## Build all services
	@echo "Building services..."
	@for service in $(SERVICES); do \
		echo "  -> $$service"; \
		cd cmd/$$service && $(GO) build -o ../../$(BIN_DIR)/$$service || exit 1; \
		cd ../..; \
	done
	@echo "✅ Build complete! Binaries: $(BIN_DIR)/"

test: ## Run all tests
	$(GO) test -v -race -cover ./...

test/%: ## Run tests for specific package
	$(GO) test -v -race -cover ./$*

clean: ## Clean build artifacts
	rm -rf bin/
	$(GO) clean ./...

docker-up: ## Start Docker services (Postgres, Redis)
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop Docker services
	$(DOCKER_COMPOSE) down

docker-logs: ## View Docker logs
	$(DOCKER_COMPOSE) logs -f

migrate-up: ## Run database migrations
	cd migrations && $(GO) run main.go up

migrate-down: ## Rollback database migrations
	cd migrations && $(GO) run main.go down

migrate-create: ## Create new migration (usage: make migrate-create NAME=add_users_table)
	cd migrations && $(GO) run main.go create $(NAME)

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...
	gofmt -s -w .

mod-tidy: ## Tidy go.mod
	$(GO) mod tidy

SWAGgenerate: ## Generate Swagger docs
	swag init -g cmd/gateway/main.go -o docs
