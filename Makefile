# Stocky Backend Makefile

.PHONY: help build run test clean docker-build docker-run migrate seed

# Default target
help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application locally"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  migrate     - Run database migrations"
	@echo "  seed        - Seed database with sample data"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run  - Run with Docker Compose"
	@echo "  docker-down - Stop Docker Compose"

# Build the application
build:
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/seed ./cmd/seed

# Run the application
run:
	go run ./cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run database migrations
migrate:
	go run ./cmd/migrate/main.go

# Seed database with sample data
seed:
	go run ./cmd/seed/main.go

# Build Docker image
docker-build:
	docker build -t stocky-backend .

# Run with Docker Compose
docker-run:
	docker-compose up --build -d

# Stop Docker Compose
docker-down:
	docker-compose down

# Run development environment
dev: migrate seed run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Generate mocks (if using mockery)
mocks:
	mockery --all --output=mocks

# Database setup (creates database if not exists)
db-setup:
	createdb assignment || true

# Full setup for new developers
setup: deps db-setup migrate seed
	@echo "Setup complete! Run 'make run' to start the server."