.PHONY: help migrate-up migrate-down migrate-create migrate-force run build clean

MIGRATION_DIR=./migrations
DB_URL=postgres://postgres:postgres@localhost:5436/payment_db?sslmode=disable

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install required tools
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up: ## Run database migrations up
	@echo "Running payment migrations up..."
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "Payment migrations completed!"

migrate-down: ## Rollback database migrations
	@echo "Rolling back payment migrations..."
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down
	@echo "Rollback completed!"

migrate-force: ## Force migration to specific version (usage: make migrate-force VERSION=1)
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" force $(VERSION)

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_table)
	@migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)

migrate-version: ## Show current migration version
	@migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" version

run: ## Run the service
	@go run main.go

build: ## Build the service
	@go build -o bin/payment-service main.go

clean: ## Clean build artifacts
	@rm -rf bin/
