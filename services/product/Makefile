.PHONY: help migrate-up migrate-down migrate-create migrate-force run build test clean

MIGRATION_DIR=./migrations
DB_URL=postgres://postgres:postgres@localhost:5434/product_db?sslmode=disable
DB_HOST=localhost
DB_PORT=5434
DB_USER=postgres
DB_NAME=product_db

help: ## Display this help screen
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install required tools
	@echo "Installing migrate tool..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed successfully!"

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "Migrations completed!"

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down
	@echo "Rollback completed!"

migrate-force: ## Force migration to specific version (usage: make migrate-force VERSION=1)
	@echo "Forcing migration to version $(VERSION)..."
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" force $(VERSION)

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_table)
	@echo "Creating migration: $(NAME)..."
	@migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)
	@echo "Migration files created!"

migrate-version: ## Show current migration version
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" version

run: ## Run the service
	@echo "Starting product service..."
	@go run main.go

build: ## Build the service
	@echo "Building product service..."
	@go build -o bin/product-service main.go
	@echo "Build completed! Binary: bin/product-service"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Tests completed! Coverage report: coverage.html"

test-short: ## Run short tests
	@go test -v -short ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean completed!"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t product-service:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8082:8082 --env-file .env product-service:latest
