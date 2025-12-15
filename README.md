# Go API Starter 🚀

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![CI](https://github.com/hassan123789/go-api-starter/actions/workflows/ci.yml/badge.svg)](https://github.com/hassan123789/go-api-starter/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hassan123789/go-api-starter)](https://goreportcard.com/report/github.com/hassan123789/go-api-starter)
[![codecov](https://codecov.io/gh/hassan123789/go-api-starter/branch/main/graph/badge.svg)](https://codecov.io/gh/hassan123789/go-api-starter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Production-ready Go REST API starter template** featuring JWT authentication, clean architecture, and modern Go patterns.

A comprehensive TODO management API built with Go best practices, designed to demonstrate production-grade code patterns including:

- 🔐 **JWT Authentication** - Secure token-based auth with refresh tokens
- 🏗️ **Clean Architecture** - Handler → Service → Repository layering
- 🛡️ **Type-Safe Error Handling** - Custom errors with `errors.Is/As` support
- ⚡ **Generics** - Go 1.18+ generic utilities (Result, Option, functional helpers)
- 🔄 **Circuit Breaker** - Resilient external service calls
- 🧵 **Worker Pool** - Concurrent task processing with generics
- 📊 **Structured Logging** - Production-ready logging with `log/slog`
- 🩺 **Health Checks** - Kubernetes-ready liveness/readiness probes
- 🚦 **Rate Limiting** - Token bucket algorithm implementation
- 📝 **Context Utilities** - Type-safe context value handling

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              HTTP Layer                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  RequestID  │→ │   Logger    │→ │   Recover   │→ │ RateLimiter │        │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘        │
│         ↓                                                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                        Echo Router                                   │    │
│  │  /health          /api/v1/users      /api/v1/todos                  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
         ↓                    ↓                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                             Handler Layer                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                         │
│  │HealthHandler│  │ AuthHandler │  │ TodoHandler │                         │
│  └─────────────┘  └─────────────┘  └─────────────┘                         │
│        │                 │                 │                                 │
│        │         Request Validation        │                                 │
│        │         Response Formatting       │                                 │
└─────────────────────────────────────────────────────────────────────────────┘
         ↓                    ↓                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                             Service Layer                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                         │
│  │ HealthService│  │ AuthService │  │ TodoService │                         │
│  └─────────────┘  └─────────────┘  └─────────────┘                         │
│        │                 │                 │                                 │
│        │         Business Logic            │                                 │
│        │         JWT Generation            │                                 │
│        │         Password Hashing          │                                 │
└─────────────────────────────────────────────────────────────────────────────┘
         ↓                    ↓                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Repository Layer                                   │
│  ┌──────────────────────┐  ┌──────────────────────┐                        │
│  │   UserRepository     │  │   TodoRepository     │                        │
│  │   (interface)        │  │   (interface)        │                        │
│  └──────────────────────┘  └──────────────────────┘                        │
│              │                        │                                      │
│              ↓                        ↓                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                        PostgreSQL                                    │    │
│  │  ┌───────────┐    ┌───────────┐                                     │    │
│  │  │   users   │───→│   todos   │                                     │    │
│  │  └───────────┘    └───────────┘                                     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 📁 Project Structure

```
go-api-starter/
├── cmd/
│   └── server/              # Application entry point
│       └── main.go          # Graceful shutdown, DI setup
│
├── internal/                # Private application code
│   ├── config/              # Configuration management
│   ├── handler/             # HTTP handlers (controllers)
│   │   ├── auth_handler.go
│   │   ├── todo_handler.go
│   │   └── handler_test.go
│   ├── middleware/          # Custom middleware
│   │   └── middleware.go    # RequestID, Logger, RateLimiter
│   ├── model/               # Domain models
│   ├── repository/          # Data access layer
│   │   ├── interfaces.go    # Repository interfaces
│   │   ├── user_repository.go
│   │   └── todo_repository.go
│   └── service/             # Business logic layer
│
├── pkg/                     # Public reusable packages
│   ├── apperrors/           # Custom error types
│   │   ├── errors.go        # AppError, ErrorCode, helpers
│   │   └── errors_test.go
│   ├── circuitbreaker/      # Circuit breaker pattern
│   │   ├── circuitbreaker.go
│   │   └── circuitbreaker_test.go
│   ├── ctxutil/             # Context utilities
│   │   ├── ctxutil.go       # Type-safe context values
│   │   └── ctxutil_test.go
│   ├── generic/             # Generic utilities
│   │   ├── generic.go       # Result, Option, Filter, Map, etc.
│   │   └── generic_test.go
│   ├── healthcheck/         # Health check system
│   │   ├── healthcheck.go
│   │   └── healthcheck_test.go
│   ├── server/              # Server with functional options
│   │   └── server.go
│   ├── workerpool/          # Worker pool for concurrency
│   │   ├── workerpool.go
│   │   └── workerpool_test.go
│   ├── response/            # Standard API responses
│   └── validator/           # Input validation
│
├── db/
│   ├── migrations/          # SQL migrations
│   └── queries/             # sqlc queries
│
├── .github/
│   └── workflows/
│       └── ci.yml           # GitHub Actions CI/CD
│
├── .golangci.yml            # Linter configuration (50+ linters)
├── Dockerfile               # Multi-stage build
├── docker-compose.yml       # Local development
├── Makefile                 # Development commands
└── README.md
```

## 🛠️ Tech Stack

| Category | Technology | Purpose |
|----------|------------|---------|
| **Language** | Go 1.22+ | Core language |
| **Framework** | Echo v4 | HTTP routing, middleware |
| **Database** | PostgreSQL 16 | Data persistence |
| **Auth** | golang-jwt/jwt/v5 | JWT token handling |
| **Logging** | log/slog | Structured logging |
| **Container** | Docker | Containerization |
| **CI/CD** | GitHub Actions | Automated testing |
| **Linting** | golangci-lint | Code quality (50+ linters) |

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make (optional)

### Installation

```bash
# Clone the repository
git clone https://github.com/hassan123789/go-api-starter.git
cd go-api-starter

# Copy environment file
cp .env.example .env

# Install development tools
make setup

# Start database
make docker-up

# Run migrations
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/go_api_starter?sslmode=disable"
make migrate

# Start the server
make run
```

### Using Docker Compose (Full Stack)

```bash
docker-compose up -d
```

The API will be available at `http://localhost:8080`.

## 🔌 API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/health` | Health check | ❌ |
| `GET` | `/livez` | Liveness probe | ❌ |
| `GET` | `/readyz` | Readiness probe | ❌ |
| `POST` | `/api/v1/users` | Register user | ❌ |
| `POST` | `/api/v1/auth/login` | Login | ❌ |
| `GET` | `/api/v1/todos` | List todos | ✅ |
| `POST` | `/api/v1/todos` | Create todo | ✅ |
| `GET` | `/api/v1/todos/:id` | Get todo | ✅ |
| `PUT` | `/api/v1/todos/:id` | Update todo | ✅ |
| `DELETE` | `/api/v1/todos/:id` | Delete todo | ✅ |

### Example Usage

```bash
# Register a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "securepass123"}'

# Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "securepass123"}' | jq -r '.token')

# Create a todo
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title": "Learn Go patterns"}'

# List todos
curl http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer $TOKEN"
```

## 📦 Featured Packages

### `pkg/apperrors` - Custom Error Types

```go
import "github.com/hassan123789/go-api-starter/pkg/apperrors"

// Create typed errors
err := apperrors.NewNotFound("todo", todoID)
err := apperrors.NewValidation("email", "invalid format")

// Check error types
if errors.Is(err, apperrors.ErrNotFound) {
    // Handle not found
}

// Get HTTP status
status := apperrors.GetHTTPStatus(err) // 404
```

### `pkg/generic` - Generic Utilities

```go
import "github.com/hassan123789/go-api-starter/pkg/generic"

// Result type (Rust-like)
result := generic.Ok(42)
if result.IsOk() {
    value := result.Unwrap()
}

// Option type
opt := generic.Some("value")
value := opt.UnwrapOr("default")

// Functional helpers
numbers := []int{1, 2, 3, 4, 5}
evens := generic.Filter(numbers, func(n int) bool { return n%2 == 0 })
doubled := generic.MapSlice(numbers, func(n int) int { return n * 2 })
sum := generic.Reduce(numbers, 0, func(acc, n int) int { return acc + n })
```

### `pkg/circuitbreaker` - Circuit Breaker

```go
import "github.com/hassan123789/go-api-starter/pkg/circuitbreaker"

cb := circuitbreaker.New(circuitbreaker.Options{
    MaxFailures: 5,
    Timeout:     30 * time.Second,
})

err := cb.Execute(ctx, func(ctx context.Context) error {
    return callExternalService(ctx)
})

// With fallback
err = cb.ExecuteWithFallback(ctx, mainFn, fallbackFn)
```

### `pkg/workerpool` - Worker Pool

```go
import "github.com/hassan123789/go-api-starter/pkg/workerpool"

// Process items concurrently
results, errors := workerpool.Process(ctx, 10, items, func(ctx context.Context, item Item) (Result, error) {
    return processItem(ctx, item)
})

// Pipeline processing
pipeline := workerpool.NewPipeline[Data]().
    AddStage(validate).
    AddStage(transform).
    AddStage(enrich)

result, err := pipeline.Execute(ctx, input)
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage report
make test-coverage

# Run benchmarks
make bench

# Run short tests only
make test-short
```

## 📝 Development Commands

```bash
make help           # Show all available commands

# Build & Run
make build          # Build with version info
make build-all      # Build for multiple platforms
make run            # Run locally
make clean          # Remove build artifacts

# Code Quality
make lint           # Run golangci-lint (50+ linters)
make fmt            # Format code
make vet            # Run go vet
make sec            # Run security checks
make check          # Run all checks

# Docker
make docker-up      # Start containers
make docker-down    # Stop containers
make docker-build   # Build image
make docker-logs    # View logs

# Database
make migrate        # Run migrations
make migrate-down   # Rollback migration
make db-shell       # Open psql shell
```

## 🔧 Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection URL | - |
| `JWT_SECRET` | JWT signing key | - |
| `JWT_EXPIRY` | Token expiry (hours) | `24` |

## 📊 Database Schema

```
┌──────────────────┐       ┌──────────────────┐
│      users       │       │      todos       │
├──────────────────┤       ├──────────────────┤
│ id (PK)          │───┐   │ id (PK)          │
│ email (UNIQUE)   │   │   │ user_id (FK)     │←─┘
│ password_hash    │   │   │ title            │
│ created_at       │   └──→│ completed        │
│ updated_at       │       │ created_at       │
└──────────────────┘       │ updated_at       │
                           └──────────────────┘
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

**⭐ If you find this project useful, please give it a star!**
