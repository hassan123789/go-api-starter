.PHONY: build run test lint docker-up docker-down migrate migrate-down sqlc setup help tidy fmt \
       test-coverage test-race test-short bench clean install-tools vet sec docker-build \
       docker-push docker-logs db-shell api-test mock-gen docs

# Project info
PROJECT_NAME := go-api-starter
BINARY_NAME := server
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Versioning
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Docker
DOCKER_IMAGE := $(PROJECT_NAME)
DOCKER_TAG := $(VERSION)

# Directories
BIN_DIR := bin
COVERAGE_DIR := coverage

# Default target
help:
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║              $(PROJECT_NAME) - Makefile Commands                ║"
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "Build & Run:"
	@echo "  build          - Build the application with version info"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  run            - Run the application locally"
	@echo "  clean          - Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run all tests with race detection"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-short     - Run short tests only"
	@echo "  bench          - Run benchmarks"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code with gofmt"
	@echo "  vet            - Run go vet"
	@echo "  sec            - Run security checks with gosec"
	@echo "  check          - Run all code quality checks"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up      - Start Docker containers"
	@echo "  docker-down    - Stop Docker containers"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-push    - Push Docker image to registry"
	@echo "  docker-logs    - View container logs"
	@echo ""
	@echo "Database:"
	@echo "  migrate        - Run database migrations"
	@echo "  migrate-down   - Rollback last migration"
	@echo "  migrate-status - Show migration status"
	@echo "  db-shell       - Open PostgreSQL shell"
	@echo "  sqlc           - Generate sqlc code"
	@echo ""
	@echo "Development:"
	@echo "  setup          - Setup development environment"
	@echo "  install-tools  - Install development tools"
	@echo "  tidy           - Run go mod tidy"
	@echo "  mock-gen       - Generate mocks for testing"
	@echo "  docs           - Generate documentation"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  GIT_COMMIT=$(GIT_COMMIT)"

# === Build Targets ===

# Build the application with version info
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/server
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/server
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/server
	@echo "Build complete for all platforms"

# Run the application
run:
	$(GO) run ./cmd/server

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# === Test Targets ===

# Run all tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -count=1 ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

# Run short tests
test-short:
	$(GO) test -v -short ./...

# Run benchmarks
bench:
	$(GO) test -v -bench=. -benchmem ./...

# === Code Quality Targets ===

# Run linter
lint:
	@echo "Running golangci-lint..."
	golangci-lint run --timeout 5m

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .
	goimports -w .

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run security checks
sec:
	@echo "Running security checks..."
	gosec -quiet ./...

# Run all code quality checks
check: fmt vet lint sec test
	@echo "All checks passed!"

# === Docker Targets ===

# Start Docker containers
docker-up:
	docker-compose up -d

# Stop Docker containers
docker-down:
	docker-compose down

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

# Push Docker image
docker-push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

# View container logs
docker-logs:
	docker-compose logs -f

# === Database Targets ===

# Run database migrations
migrate:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

# Rollback last migration
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

# Show migration status
migrate-status:
	migrate -path db/migrations -database "$(DATABASE_URL)" version

# Open PostgreSQL shell
db-shell:
	docker-compose exec db psql -U postgres -d go_api_starter

# Generate sqlc code
sqlc:
	sqlc generate

# Setup development environment
setup: install-tools
	$(GO) mod download
	@echo "Setup complete! Don't forget to:"
	@echo "  1. Copy .env.example to .env and update values"
	@echo "  2. Run 'make docker-up' to start the database"
	@echo "  3. Run 'make migrate' to run migrations"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	$(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/vektra/mockery/v2@latest
	@echo "Tools installed successfully"

# Run go mod tidy
tidy:
	$(GO) mod tidy

# Generate mocks for testing
mock-gen:
	mockery --all --with-expecter --output ./internal/mocks

# Generate documentation
docs:
	@echo "Generating godoc documentation..."
	@echo "Visit http://localhost:6060/pkg/github.com/zareh/go-api-starter/"
	godoc -http=:6060

# Format code
fmt:
	$(GO) fmt ./...
	gofmt -s -w .
