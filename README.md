# Go API Starter 🚀

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![CI](https://github.com/hassan123789/go-api-starter/actions/workflows/ci.yml/badge.svg)](https://github.com/hassan123789/go-api-starter/actions/workflows/ci.yml)
[![Security](https://github.com/hassan123789/go-api-starter/actions/workflows/security.yml/badge.svg)](https://github.com/hassan123789/go-api-starter/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hassan123789/go-api-starter)](https://goreportcard.com/report/github.com/hassan123789/go-api-starter)
[![codecov](https://codecov.io/gh/hassan123789/go-api-starter/branch/main/graph/badge.svg)](https://codecov.io/gh/hassan123789/go-api-starter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Production-ready Go REST &amp; gRPC API starter template** with JWT authentication, Kubernetes deployment, Clean Architecture, and a Rust CLI client.

---

## ✨ Key Features

| Category | Features |
|----------|----------|
| **Authentication** | JWT with refresh tokens, bcrypt password hashing, RBAC (Admin/User/Viewer) |
| **Architecture** | Clean Architecture (Handler → Service → Repository), Dependency Injection |
| **API** | REST (Echo v4) + gRPC with streaming, OpenAPI 3.1 specification |
| **Resilience** | Circuit breaker, retry with exponential backoff, rate limiting, graceful shutdown |
| **Observability** | OpenTelemetry tracing (Jaeger), Prometheus metrics, structured logging (slog), audit logs |
| **Infrastructure** | Docker, Kubernetes manifests with Kustomize (dev/prod), GitHub Actions CI/CD |
| **Developer Tools** | Rust CLI client, Dev Container, Taskfile, pre-commit hooks, sqlc |
| **Go Patterns** | Generics (Result/Option types), Worker Pool, Context utilities, Type-safe errors |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Transport Layer                                    │
│  ┌───────────────────────────────┐  ┌───────────────────────────────┐       │
│  │         REST (Echo)           │  │       gRPC (grpc-go)          │       │
│  │  Port 8080                    │  │  Port 9090                    │       │
│  │  /api/v1/todos, /health       │  │  TodoService (streaming)      │       │
│  └───────────────────────────────┘  └───────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Middleware Chain                                  │
│  RequestID → Logger → Recovery → RateLimiter → Auth → RBAC                  │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Handler Layer                                     │
│  Request validation, response formatting, error mapping to HTTP/gRPC codes  │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Service Layer                                     │
│  Business logic, JWT generation, password hashing, authorization checks      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Repository Layer                                   │
│  Data access interfaces, PostgreSQL implementation (sqlc generated)         │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                             PostgreSQL                                       │
│  users (id, email, password_hash, role) ←──→ todos (id, user_id, title...)  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
go-api-starter/
├── cmd/server/              # Application entry point with graceful shutdown
├── internal/
│   ├── config/              # Configuration management (env, validation)
│   ├── handler/             # HTTP handlers (REST controllers)
│   ├── grpc/                # gRPC server and service implementations
│   ├── middleware/          # Custom middleware (auth, logging, rate limit)
│   ├── model/               # Domain models
│   ├── repository/          # Data access layer (interfaces + implementations)
│   └── service/             # Business logic layer
├── pkg/                     # Reusable packages (see below)
│   ├── apperrors/           # Custom error types with Is/As support
│   ├── cache/               # Generic in-memory cache with TTL/LRU
│   ├── circuitbreaker/      # Circuit breaker pattern
│   ├── generic/             # Result/Option types, functional helpers
│   ├── healthcheck/         # K8s-ready health check system
│   ├── metrics/             # Prometheus metrics
│   ├── rbac/                # Role-based access control
│   ├── resilience/          # Retry, timeout, fallback patterns
│   ├── retry/               # Exponential backoff with jitter
│   ├── token/               # JWT token generation/validation
│   ├── tracing/             # OpenTelemetry integration
│   └── workerpool/          # Concurrent task processing
├── api/grpc/                # Protocol Buffers definitions
├── gen/go/                  # Generated gRPC code (buf)
├── deploy/
│   └── k8s/                 # Kubernetes manifests
│       ├── base/            # Base manifests (Deployment, Service, HPA, etc.)
│       └── overlays/        # Environment-specific (development, production)
├── tools/
│   └── todo-cli/            # Rust CLI client for the API
├── db/
│   ├── migrations/          # SQL migrations
│   └── queries/             # sqlc query definitions
├── docs/
│   ├── adr/                 # Architecture Decision Records
│   └── openapi.yaml         # OpenAPI 3.1 specification
└── .github/workflows/       # CI/CD pipelines
```

---

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Docker &amp; Docker Compose
- (Optional) [Task](https://taskfile.dev/) for development commands

### Option 1: Docker Compose (Recommended)

```bash
# Clone and start all services
git clone https://github.com/hassan123789/go-api-starter.git
cd go-api-starter
docker-compose up -d

# API available at http://localhost:8080
# gRPC available at localhost:9090
```

### Option 2: Local Development

```bash
# Clone repository
git clone https://github.com/hassan123789/go-api-starter.git
cd go-api-starter

# Setup environment
cp .env.example .env

# Start PostgreSQL
docker-compose up -d postgres

# Run migrations
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/go_api_starter?sslmode=disable"
task migrate  # or: make migrate

# Start server (with hot reload)
task run      # or: make run
```

### Option 3: Dev Container (VS Code)

1. Open the project in VS Code
2. Click "Reopen in Container" when prompted
3. All tools and dependencies are pre-configured

---

## 🔌 API Reference

### REST Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/health` | Full health check with dependencies | ❌ |
| `GET` | `/livez` | Kubernetes liveness probe | ❌ |
| `GET` | `/readyz` | Kubernetes readiness probe | ❌ |
| `GET` | `/metrics` | Prometheus metrics | ❌ |
| `POST` | `/api/v1/users` | Register new user | ❌ |
| `POST` | `/api/v1/auth/login` | Login, returns JWT | ❌ |
| `POST` | `/api/v1/auth/refresh` | Refresh access token | ✅ |
| `GET` | `/api/v1/todos` | List todos (paginated) | ✅ |
| `POST` | `/api/v1/todos` | Create todo | ✅ |
| `GET` | `/api/v1/todos/:id` | Get todo by ID | ✅ |
| `PUT` | `/api/v1/todos/:id` | Update todo | ✅ |
| `DELETE` | `/api/v1/todos/:id` | Delete todo | ✅ |

### gRPC Service

```protobuf
service TodoService {
  rpc CreateTodo(CreateTodoRequest) returns (Todo);
  rpc GetTodo(GetTodoRequest) returns (Todo);
  rpc ListTodos(ListTodosRequest) returns (ListTodosResponse);
  rpc UpdateTodo(UpdateTodoRequest) returns (Todo);
  rpc DeleteTodo(DeleteTodoRequest) returns (google.protobuf.Empty);
  rpc StreamTodos(StreamTodosRequest) returns (stream TodoEvent);  // Real-time updates
}
```

### Example Usage

```bash
# Register a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "securepass123"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "securepass123"}' | jq -r '.token')

# Create a todo
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title": "Learn Go patterns", "description": "Study clean architecture"}'

# List todos
curl http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🦀 Rust CLI Client

A native CLI tool for interacting with the API:

```bash
# Build the CLI
cd tools/todo-cli
cargo build --release

# Configure API endpoint
./target/release/todo-cli config set-url http://localhost:8080

# Authenticate
./target/release/todo-cli auth login

# Manage todos
./target/release/todo-cli list
./target/release/todo-cli create --title "New task" --description "Details"
./target/release/todo-cli done 1
./target/release/todo-cli delete 1

# JSON output for scripting
./target/release/todo-cli list --format json
```

---

## ☸️ Kubernetes Deployment

Production-ready Kubernetes manifests with Kustomize:

```bash
# Preview development deployment
kubectl kustomize deploy/k8s/overlays/development

# Preview production deployment
kubectl kustomize deploy/k8s/overlays/production

# Apply to cluster
kubectl apply -k deploy/k8s/overlays/production
```

### Included Resources

| Resource | Description |
|----------|-------------|
| Deployment | Rolling updates, resource limits, security context (non-root) |
| Service | ClusterIP for internal communication |
| HPA | Auto-scaling (3-20 replicas based on CPU/memory) |
| PDB | Pod disruption budget (minAvailable: 2) |
| Ingress | With cert-manager TLS annotations |
| NetworkPolicy | Restrict traffic to namespace |
| ConfigMap | Environment configuration |
| Secret | Sensitive data (JWT secret, DB credentials) |

---

## 📦 Reusable Packages (`pkg/`)

### Error Handling (`pkg/apperrors`)

```go
import "github.com/hassan123789/go-api-starter/pkg/apperrors"

// Create typed errors
err := apperrors.NewNotFound("todo", todoID)
err := apperrors.NewValidation("email", "invalid format")
err := apperrors.NewUnauthorized("invalid credentials")

// Check error types (works with errors.Is)
if errors.Is(err, apperrors.ErrNotFound) {
    // Handle not found
}

// Automatic HTTP status mapping
status := apperrors.GetHTTPStatus(err) // 404, 400, 401, etc.
```

### Generic Utilities (`pkg/generic`)

```go
import "github.com/hassan123789/go-api-starter/pkg/generic"

// Result type (Rust-like error handling)
result := generic.Ok(fetchData())
if result.IsOk() {
    data := result.Unwrap()
}
errorResult := generic.Err[int](errors.New("failed"))

// Option type
opt := generic.Some("value")
value := opt.UnwrapOr("default")

// Functional helpers
numbers := []int{1, 2, 3, 4, 5}
evens := generic.Filter(numbers, func(n int) bool { return n%2 == 0 })
doubled := generic.MapSlice(numbers, func(n int) int { return n * 2 })
sum := generic.Reduce(numbers, 0, func(acc, n int) int { return acc + n })
```

### Circuit Breaker (`pkg/circuitbreaker`)

```go
import "github.com/hassan123789/go-api-starter/pkg/circuitbreaker"

cb := circuitbreaker.New(circuitbreaker.Options{
    MaxFailures:   5,
    Timeout:       30 * time.Second,
    HalfOpenLimit: 3,
})

err := cb.Execute(ctx, func(ctx context.Context) error {
    return callExternalService(ctx)
})

// With fallback
result, err := cb.ExecuteWithFallback(ctx, primaryFn, fallbackFn)
```

### Worker Pool (`pkg/workerpool`)

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

---

## 🛠️ Development

### Available Commands

```bash
# Using Taskfile (recommended)
task            # Show all available tasks
task build      # Build binary
task test       # Run tests
task lint       # Run golangci-lint (50+ linters)
task check      # Run all checks (lint, test, vet)
task docker:up  # Start Docker services
task proto      # Regenerate gRPC code

# Using Makefile (alternative)
make help       # Show all commands
make build      # Build binary
make test       # Run tests with coverage
make lint       # Run linter
```

### Testing

```bash
# Run all tests
task test

# Run with coverage report
task test:coverage

# Run specific package tests
go test -v ./pkg/circuitbreaker/...

# Run benchmarks
task bench
```

### Code Quality

```bash
# Lint with 50+ rules
task lint

# Format code
task fmt

# Security scan
task sec

# Pre-commit hooks (auto-run on commit)
pre-commit install
pre-commit run --all-files
```

---

## 🔧 Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `GRPC_PORT` | gRPC server port | `9090` |
| `GRPC_ENABLED` | Enable gRPC server | `true` |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `JWT_SECRET` | JWT signing key | - |
| `JWT_EXPIRY` | Access token expiry | `24h` |
| `JWT_REFRESH_EXPIRY` | Refresh token expiry | `168h` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `OTEL_EXPORTER_JAEGER_ENDPOINT` | Jaeger collector URL | - |

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [API Specification](docs/openapi.yaml) | OpenAPI 3.1 specification |
| [Architecture Decisions](docs/adr/) | ADRs explaining design choices |
| [Architecture Guide](docs/ARCHITECTURE.md) | Deep dive into design decisions |
| [Contributing Guide](CONTRIBUTING.md) | How to contribute |
| [Security Policy](SECURITY.md) | Reporting vulnerabilities |
| [Kubernetes Guide](deploy/k8s/README.md) | Deployment instructions |

### Architecture Decision Records (ADR)

| ADR | Title |
|-----|-------|
| [001](docs/adr/001-use-clean-architecture.md) | Adopting Clean Architecture |
| [002](docs/adr/002-choose-echo-framework.md) | Choosing Echo Framework |
| [003](docs/adr/003-jwt-authentication-strategy.md) | JWT Authentication Strategy |
| [004](docs/adr/004-error-handling-approach.md) | Error Handling Design |
| [005](docs/adr/0005-opentelemetry-tracing.md) | OpenTelemetry Tracing |
| [006](docs/adr/0006-rbac-strategy.md) | RBAC Strategy |
| [007](docs/adr/0007-resilience-patterns.md) | Resilience Patterns |
| [008](docs/adr/0008-audit-logging.md) | Audit Logging |

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all checks pass (`task check`)
5. Commit with conventional commits (`git commit -m 'feat: add amazing feature'`)
6. Push to your branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

**⭐ If this project helps you, please consider giving it a star!**
