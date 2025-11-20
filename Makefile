.PHONY: help wire build run test clean migrate docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

wire: ## Generate wire dependency injection
	@echo "🔧 Generating wire dependencies..."
	cd internal/di && wire

build: ## Build the application
	@echo "🔨 Building application..."
	go build -o bin/api cmd/api/main.go

run: ## Run the application
	@echo "🚀 Running application..."
	go run cmd/api/main.go

dev: wire run ## Generate wire and run application

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf bin/
	rm -rf logs/*.log
	rm -f coverage.out

migrate: ## Run database migrations
	@echo "📦 Running migrations..."
	go run cmd/api/main.go migrate

install-wire: ## Install Google Wire
	@echo "📥 Installing Wire..."
	go install github.com/google/wire/cmd/wire@latest

install-deps: ## Install all dependencies
	@echo "📥 Installing dependencies..."
	go mod download
	go mod tidy

docker-up: ## Start docker containers
	@echo "🐳 Starting docker containers..."
	docker-compose up -d

docker-down: ## Stop docker containers
	@echo "🐳 Stopping docker containers..."
	docker-compose down

docker-logs: ## View docker logs
	docker-compose logs -f

lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "🎨 Formatting code..."
	go fmt ./...
	goimports -w .

mod-update: ## Update dependencies
	@echo "📦 Updating dependencies..."
	go get -u ./...
	go mod tidy

setup: install-deps install-wire wire ## Complete setup (install deps + wire generation)
	@echo "✅ Setup complete!"

.DEFAULT_GOAL := help