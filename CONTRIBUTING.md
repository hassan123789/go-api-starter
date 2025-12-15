# Contributing to Go API Starter

First off, thank you for considering contributing to Go API Starter! It's people like you that make this project such a great tool.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)

## Code of Conduct

This project and everyone participating in it is governed by our commitment to maintaining a welcoming and inclusive environment. Please be respectful and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- PostgreSQL 16 (or use Docker)
- Make (optional but recommended)
- golangci-lint

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/go-api-starter.git
   cd go-api-starter
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/hassan123789/go-api-starter.git
   ```

## Development Setup

### Using Docker (Recommended)

```bash
# Start all services
docker-compose up -d

# Run migrations
docker-compose exec app make migrate-up
```

### Local Development

```bash
# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Start PostgreSQL
docker-compose up -d db

# Run the application
make run
```

### Verify Setup

```bash
# Run all tests
make test

# Run linter
make lint

# Check formatting
make fmt
```

## Coding Standards

### Go Code Style

We follow the official Go style guidelines:

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

### Key Principles

1. **Simplicity**: Write simple, readable code
2. **Error Handling**: Always handle errors explicitly
3. **Testing**: Write tests for all new functionality
4. **Documentation**: Document all exported functions and types

### Code Organization

```
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── config/            # Configuration
│   ├── handler/           # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── model/             # Domain models
│   ├── repository/        # Data access layer
│   └── service/           # Business logic
├── pkg/                   # Public reusable packages
│   ├── apperrors/        # Custom error types
│   ├── cache/            # Generic cache
│   ├── circuitbreaker/   # Circuit breaker pattern
│   ├── generic/          # Generic utilities
│   ├── healthcheck/      # Health check utilities
│   ├── metrics/          # Prometheus metrics
│   ├── server/           # HTTP server utilities
│   └── workerpool/       # Worker pool pattern
└── api/                   # API specifications
```

### Naming Conventions

```go
// Package names: lowercase, single word
package handler

// Interfaces: use -er suffix when appropriate
type UserRepository interface {
    FindByID(ctx context.Context, id int64) (*User, error)
}

// Exported functions: PascalCase with clear names
func NewUserService(repo UserRepository) *UserService

// Unexported functions: camelCase
func validateEmail(email string) error

// Constants: PascalCase for exported, camelCase for unexported
const MaxRetries = 3
const defaultTimeout = 30 * time.Second

// Error variables: ErrXxx pattern
var ErrUserNotFound = errors.New("user not found")
```

### Error Handling

```go
// Use custom error types from pkg/apperrors
import "github.com/zareh/go-api-starter/pkg/apperrors"

// Return descriptive errors
if user == nil {
    return apperrors.ErrNotFound.WithMessage("user not found")
}

// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to query user: %w", err)
}

// Always check returned errors
result, err := doSomething()
if err != nil {
    return err
}
```

### Context Usage

```go
// Always accept context as first parameter
func (s *Service) DoWork(ctx context.Context, input Input) (Output, error) {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return Output{}, ctx.Err()
    default:
    }

    // Pass context to downstream calls
    return s.repo.Query(ctx, input)
}
```

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, semicolons, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `build`: Build system or dependency changes
- `ci`: CI/CD changes
- `chore`: Other changes that don't modify src or test files

### Examples

```
feat(auth): add JWT refresh token support

Implement refresh token rotation for improved security.
Tokens are stored in Redis with configurable expiration.

Closes #123
```

```
fix(handler): handle nil pointer in user update

Check for nil user before attempting update to prevent panic.
```

```
docs(readme): add API usage examples

Add curl examples for common API operations.
```

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**:
   ```bash
   make lint
   make test
   make build
   ```

3. **Update documentation** if needed

4. **Add tests** for new functionality

### PR Checklist

- [ ] Code follows the project's coding standards
- [ ] All tests pass locally
- [ ] New tests added for new functionality
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventional commits
- [ ] No merge conflicts with main branch
- [ ] CI pipeline passes

### PR Title Format

Follow the same format as commit messages:
```
feat(scope): brief description
```

### PR Description Template

```markdown
## Description
Brief description of changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe the tests you ran.

## Related Issues
Closes #123
```

## Testing Guidelines

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    t.Parallel() // Enable parallel execution

    t.Run("describes the test case", func(t *testing.T) {
        // Arrange
        input := setupInput()

        // Act
        result, err := FunctionName(input)

        // Assert
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if result != expected {
            t.Errorf("expected %v, got %v", expected, result)
        }
    })
}
```

### Table-Driven Tests

```go
func TestValidateEmail(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {name: "valid email", email: "user@example.com", wantErr: false},
        {name: "missing @", email: "userexample.com", wantErr: true},
        {name: "empty string", email: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Coverage

- Aim for at least 80% code coverage
- Focus on testing business logic and edge cases
- Use `make cover` to check coverage

### Mocking

Use interfaces for dependencies to enable mocking:

```go
// Interface for repository
type UserRepository interface {
    FindByID(ctx context.Context, id int64) (*User, error)
}

// Mock implementation for tests
type mockUserRepo struct {
    findByIDFunc func(ctx context.Context, id int64) (*User, error)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int64) (*User, error) {
    return m.findByIDFunc(ctx, id)
}
```

## Documentation

### Code Documentation

```go
// Package handler provides HTTP request handlers for the API.
package handler

// UserHandler handles HTTP requests for user-related operations.
// It implements the standard CRUD operations for users.
type UserHandler struct {
    service UserService
}

// NewUserHandler creates a new UserHandler with the given service.
// It returns an initialized handler ready for use with an HTTP router.
func NewUserHandler(service UserService) *UserHandler {
    return &UserHandler{service: service}
}

// GetUser retrieves a user by ID.
// It expects an "id" path parameter and returns the user as JSON.
//
// Possible responses:
//   - 200: User found and returned
//   - 404: User not found
//   - 500: Internal server error
func (h *UserHandler) GetUser(c echo.Context) error {
    // Implementation
}
```

### Example Functions

Add example functions for `pkg.go.dev` documentation:

```go
func ExampleNewCache() {
    c := cache.New[string, int](
        cache.WithCapacity[string, int](100),
        cache.WithTTL[string, int](5 * time.Minute),
    )
    defer c.Close()

    c.Set("key", 42)
    if val, ok := c.Get("key"); ok {
        fmt.Println(val)
    }
    // Output: 42
}
```

## Questions?

If you have any questions, please:

1. Check existing [issues](https://github.com/hassan123789/go-api-starter/issues)
2. Create a new issue with the `question` label
3. Start a discussion in the repository

Thank you for contributing! 🎉
