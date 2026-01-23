.PHONY: help run build test clean install dev

help:
	@echo "Available commands:"
	@echo "  make run        - Run the file service"
	@echo "  make build      - Build the file service"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make install    - Install dependencies"
	@echo "  make dev        - Run in development mode with hot reload"

run:
	@echo "Starting file service..."
	@go run main.go

build:
	@echo "Building file service..."
	@go build -o bin/file-service main.go

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

dev:
	@echo "Running in development mode..."
	@air || go run main.go
