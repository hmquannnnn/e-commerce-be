.PHONY: help migrate-all migrate-down-all run install-tools

USER_DB_URL    = postgres://postgres:postgres@localhost:5433/user_db?sslmode=disable
PRODUCT_DB_URL = postgres://postgres:postgres@localhost:5434/product_db?sslmode=disable
ORDER_DB_URL   = postgres://postgres:postgres@localhost:5435/order_db?sslmode=disable
PAYMENT_DB_URL = postgres://postgres:postgres@localhost:5436/payment_db?sslmode=disable

MIGRATE = $(HOME)/go/bin/migrate

help: ## Display this help screen
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install golang-migrate tool
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-all: ## Run migrations up for all services
	@echo "▶ [user]    migrating..."
	@$(MIGRATE) -path services/user/migrations    -database "$(USER_DB_URL)"    up
	@echo "▶ [product] migrating..."
	@$(MIGRATE) -path services/product/migrations -database "$(PRODUCT_DB_URL)" up
	@echo "▶ [order]   migrating..."
	@$(MIGRATE) -path services/order/migrations   -database "$(ORDER_DB_URL)"   up
	@echo "▶ [payment] migrating..."
	@$(MIGRATE) -path services/payment/migrations -database "$(PAYMENT_DB_URL)" up
	@echo "✓ All migrations completed!"

migrate-down-all: ## Rollback all migrations for all services
	@echo "▶ [payment] rolling back..."
	@$(MIGRATE) -path services/payment/migrations -database "$(PAYMENT_DB_URL)" down
	@echo "▶ [order]   rolling back..."
	@$(MIGRATE) -path services/order/migrations   -database "$(ORDER_DB_URL)"   down
	@echo "▶ [product] rolling back..."
	@$(MIGRATE) -path services/product/migrations -database "$(PRODUCT_DB_URL)" down
	@echo "▶ [user]    rolling back..."
	@$(MIGRATE) -path services/user/migrations    -database "$(USER_DB_URL)"    down
	@echo "✓ All rollbacks completed!"

run: ## Start all services
	@bash scripts/run-all.sh
