.PHONY: help migrate-up migrate-down migrate-create migrate-force seed seed-locations run build test clean

# Variables
MIGRATION_DIR=./migrations
SEED_DIR=./seeds
DB_URL=postgres://postgres:postgres@localhost:5433/user_db?sslmode=disable
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_NAME=user_db

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

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_users_table)
	@echo "Creating migration: $(NAME)..."
	@migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)
	@echo "Migration files created!"

migrate-version: ## Show current migration version
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" version

seed: ## Run all database seeds
	@echo "Running all seeds..."
	@docker exec -i ecommerce-postgres-user psql -U $(DB_USER) -d $(DB_NAME) < $(SEED_DIR)/01_vietnam_locations.sql
	@echo "Seeds completed!"

seed-locations: ## Seed Vietnam location data (cities, districts, wards)
	@echo "Seeding Vietnam locations..."
	@docker exec -i ecommerce-postgres-user psql -U $(DB_USER) -d $(DB_NAME) < $(SEED_DIR)/01_vietnam_locations.sql
	@echo "Vietnam locations seeded!"

seed-clean: ## Clean all seeded data
	@echo "Cleaning seeded data..."
	@docker exec ecommerce-postgres-user psql -U $(DB_USER) -d $(DB_NAME) -c "TRUNCATE TABLE wards, districts, cities RESTART IDENTITY CASCADE;"
	@echo "Seeded data cleaned!"

run: ## Run the service
	@echo "Starting user service..."
	@go run main.go

build: ## Build the service
	@echo "Building user service..."
	@go build -o bin/user-service main.go
	@echo "Build completed! Binary: bin/user-service"

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

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t user-service:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 50051:50051 --env-file .env user-service:latest

