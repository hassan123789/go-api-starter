.PHONY: build run test lint docker-up docker-down migrate migrate-down sqlc setup help tidy fmt

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  docker-up    - Start Docker containers"
	@echo "  docker-down  - Stop Docker containers"
	@echo "  migrate      - Run database migrations"
	@echo "  migrate-down - Rollback last migration"
	@echo "  sqlc         - Generate sqlc code"
	@echo "  setup        - Setup development environment"
	@echo "  tidy         - Run go mod tidy"
	@echo "  fmt          - Format code"

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test -v -race -cover ./...

# Run linter
lint:
	golangci-lint run

# Start Docker containers
docker-up:
	docker-compose up -d

# Stop Docker containers
docker-down:
	docker-compose down

# Run database migrations
migrate:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

# Rollback last migration
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

# Generate sqlc code
sqlc:
	sqlc generate

# Setup development environment
setup:
	go mod download
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Setup complete! Don't forget to:"
	@echo "  1. Copy .env.example to .env and update values"
	@echo "  2. Run 'make docker-up' to start the database"
	@echo "  3. Run 'make migrate' to run migrations"

# Run go mod tidy
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .
