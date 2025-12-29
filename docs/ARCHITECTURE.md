# Architecture Guide

This document explains the key architectural decisions, design patterns, and trade-offs in go-api-starter. It serves as a reference for understanding **why** specific choices were made, not just **what** was implemented.

---

## Table of Contents

1. [Design Philosophy](#design-philosophy)
2. [Layered Architecture](#layered-architecture)
3. [Why These Technologies?](#why-these-technologies)
4. [Design Patterns in Use](#design-patterns-in-use)
5. [Trade-offs and Alternatives](#trade-offs-and-alternatives)
6. [Interview Discussion Points](#interview-discussion-points)

---

## Design Philosophy

### Core Principles

| Principle | Implementation |
|-----------|----------------|
| **Separation of Concerns** | Each layer has a single responsibility. Handlers don't contain business logic; services don't know about HTTP. |
| **Dependency Inversion** | High-level modules depend on abstractions (interfaces), not concrete implementations. |
| **Testability First** | All external dependencies are injected through interfaces, enabling unit tests without databases or network calls. |
| **Explicit over Implicit** | Configuration is explicit (no magic), errors are typed and checkable, dependencies are visible in constructors. |
| **Production Readiness** | Every feature includes observability (metrics, tracing, logging), health checks, and graceful degradation. |

---

## Layered Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Transport Layer                           │
│  • Accepts HTTP/gRPC requests                               │
│  • Handles protocol-specific concerns (headers, status codes)│
│  • Delegates to handlers                                     │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Handler Layer                            │
│  • Request validation and parsing                           │
│  • Response formatting                                       │
│  • Error translation to HTTP/gRPC status codes              │
│  • NO business logic                                         │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                            │
│  • Core business logic                                       │
│  • Orchestrates repository calls                            │
│  • Authorization decisions                                   │
│  • Transaction boundaries                                    │
│  • Protocol-agnostic (doesn't know HTTP vs gRPC)            │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                           │
│  • Data access abstraction                                   │
│  • Database queries (via sqlc)                              │
│  • External service calls                                    │
│  • Caching decisions                                         │
└─────────────────────────────────────────────────────────────┘
```

### Why This Structure?

**Q: Why not just put everything in handlers like many tutorials?**

A: As applications grow, mixing concerns creates several problems:

1. **Testing becomes difficult**: You can't test business logic without spinning up HTTP servers
2. **Code reuse is impossible**: The same logic for REST and gRPC would be duplicated
3. **Changes cascade**: A database schema change affects HTTP response formatting

**Q: Why separate Handler and Service?**

A: Consider adding gRPC support. With this architecture:

- Service layer remains unchanged (100% reuse)
- Only add new gRPC handlers that call the same services
- Both REST and gRPC share identical business logic

**Real example from this project:**

```go
// REST handler calls service
func (h *TodoHandler) GetTodo(c echo.Context) error {
    todo, err := h.todoService.GetByID(ctx, id)
    return c.JSON(http.StatusOK, todo)
}

// gRPC handler calls the SAME service
func (s *TodoServer) GetTodo(ctx context.Context, req *pb.GetTodoRequest) (*pb.Todo, error) {
    todo, err := s.todoService.GetByID(ctx, req.Id)
    return toProto(todo), nil
}
```

---

## Why These Technologies?

### Echo Framework (over Gin, Chi, standard library)

| Criterion | Echo | Gin | Chi | stdlib |
|-----------|------|-----|-----|--------|
| Performance | ★★★★★ | ★★★★★ | ★★★★ | ★★★ |
| Middleware ecosystem | ★★★★★ | ★★★★★ | ★★★★ | ★★ |
| Type-safe context | ★★★★★ | ★★★ | ★★★ | ★★ |
| Learning curve | Easy | Easy | Easy | Medium |
| Extensibility | ★★★★★ | ★★★★ | ★★★★★ | ★★★★★ |

**Decision**: Echo provides the best balance of performance, ergonomics, and built-in features (validation, binding, middleware chain). Its `echo.Context` wrapper is more ergonomic than Gin's, and it has first-class support for custom validators.

**Trade-off accepted**: Slightly more opinionated than Chi/stdlib, which could limit extreme customization. In practice, this hasn't been a limitation.

### PostgreSQL (over MySQL, MongoDB, SQLite)

| Criterion | PostgreSQL | MySQL | MongoDB | SQLite |
|-----------|------------|-------|---------|--------|
| ACID compliance | ★★★★★ | ★★★★ | ★★★ | ★★★★★ |
| JSON support | ★★★★★ | ★★★ | ★★★★★ | ★★★ |
| Concurrent writes | ★★★★★ | ★★★★ | ★★★★★ | ★★ |
| Ecosystem (sqlc, etc.) | ★★★★★ | ★★★★ | ★★★ | ★★★★ |
| Scaling options | ★★★★ | ★★★★★ | ★★★★★ | ★★ |

**Decision**: PostgreSQL offers the best combination of:

1. Strong ACID guarantees for financial/critical data
2. Excellent JSON/JSONB support for flexible schemas
3. Rich feature set (CTEs, window functions, full-text search)
4. Best-in-class tooling with sqlc for type-safe queries

**Trade-off accepted**: Slightly more complex setup than SQLite, less horizontal scalability than MongoDB. For 99% of applications, PostgreSQL's vertical scaling is sufficient.

### sqlc (over GORM, sqlx, raw database/sql)

| Criterion | sqlc | GORM | sqlx | database/sql |
|-----------|------|------|------|--------------|
| Type safety | ★★★★★ | ★★★ | ★★★★ | ★★ |
| Performance | ★★★★★ | ★★★ | ★★★★★ | ★★★★★ |
| SQL visibility | ★★★★★ | ★★ | ★★★★★ | ★★★★★ |
| Learning curve | Easy | Easy | Medium | Hard |
| Compile-time checks | ★★★★★ | ❌ | ❌ | ❌ |

**Decision**: sqlc generates type-safe Go code from SQL queries, providing:

1. Compile-time query validation
2. No runtime reflection
3. Pure SQL (no DSL to learn)
4. Generated code is readable and debuggable

**Trade-off accepted**: Must write SQL instead of Go-like syntax. This is actually preferred for complex queries where ORMs struggle.

### JWT (over Sessions, OAuth2-only)

| Approach | Pros | Cons |
|----------|------|------|
| **JWT** | Stateless, scalable, works with microservices | Token size, revocation complexity |
| Sessions | Simple revocation, smaller payloads | Requires session store, sticky sessions |
| OAuth2-only | Delegates auth, enterprise SSO | External dependency, complexity |

**Decision**: JWT with refresh tokens provides:

1. Stateless authentication (no session store needed)
2. Embedded claims (user ID, role) for authorization
3. Easy integration with microservices
4. Refresh token rotation for security

**Trade-off accepted**: Implemented token blacklisting for logout (adds statefulness). For true statelessness, could use short-lived tokens only.

---

## Design Patterns in Use

### 1. Repository Pattern

**Purpose**: Abstract data access behind interfaces

```go
// Interface defined in service layer
type TodoRepository interface {
    GetByID(ctx context.Context, id int64) (*model.Todo, error)
    Create(ctx context.Context, todo *model.Todo) error
    Update(ctx context.Context, todo *model.Todo) error
    Delete(ctx context.Context, id int64) error
}

// Implementation in repository layer
type PostgresTodoRepository struct {
    db *sql.DB
}
```

**Why?**

- Services can be tested with mock repositories
- Easy to swap implementations (PostgreSQL → Redis → in-memory)
- Clear contract for data operations

### 2. Dependency Injection (Constructor Injection)

**Purpose**: Make dependencies explicit and testable

```go
// Dependencies are passed to constructors, not created inside
func NewTodoService(repo TodoRepository, logger *slog.Logger) *TodoService {
    return &TodoService{
        repo:   repo,
        logger: logger,
    }
}

// In tests, inject mocks
service := NewTodoService(mockRepo, testLogger)
```

**Why?**

- No hidden dependencies
- Easy to mock for testing
- Clear dependency graph

### 3. Functional Options Pattern

**Purpose**: Flexible, extensible configuration

```go
type ServerOption func(*Server)

func WithPort(port int) ServerOption {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(t time.Duration) ServerOption {
    return func(s *Server) {
        s.timeout = t
    }
}

// Usage
server := NewServer(
    WithPort(8080),
    WithTimeout(30*time.Second),
)
```

**Why?**

- Backward-compatible API evolution (add options without breaking callers)
- Self-documenting (option names describe purpose)
- Sensible defaults with selective overrides

### 4. Result Type Pattern (Rust-inspired)

**Purpose**: Explicit error handling without nil checks

```go
type Result[T any] struct {
    value T
    err   error
    ok    bool
}

func Ok[T any](value T) Result[T]
func Err[T any](err error) Result[T]

// Usage
result := doSomething()
if result.IsOk() {
    value := result.Unwrap()
}
```

**Why?**

- Forces handling of both success and error cases
- No nil pointer surprises
- Chainable operations (Map, FlatMap)

### 5. Circuit Breaker Pattern

**Purpose**: Prevent cascade failures in distributed systems

```go
cb := circuitbreaker.New(circuitbreaker.Options{
    MaxFailures:   5,        // Open after 5 failures
    Timeout:       30*time.Second, // Time before half-open
    HalfOpenLimit: 3,        // Requests to try in half-open
})

err := cb.Execute(ctx, func(ctx context.Context) error {
    return callExternalService(ctx)
})
```

**States:**

1. **Closed**: Requests flow normally
2. **Open**: All requests fail immediately (after MaxFailures)
3. **Half-Open**: Limited requests allowed to test recovery

**Why?**

- Prevents thundering herd on failing services
- Gives downstream services time to recover
- Fails fast instead of waiting for timeouts

### 6. Middleware Chain Pattern

**Purpose**: Cross-cutting concerns as composable layers

```go
e := echo.New()
e.Use(
    middleware.RequestID(),      // Add unique ID to every request
    middleware.Logger(),         // Log request/response
    middleware.Recover(),        // Recover from panics
    middleware.RateLimiter(),    // Rate limiting
    middleware.CORS(),           // CORS headers
)
```

**Why?**

- Single Responsibility (each middleware does one thing)
- Composable (add/remove without changing others)
- Reusable across routes

---

## Trade-offs and Alternatives

### Trade-off 1: Code Generation vs Runtime Reflection

| Approach | Used For | Pros | Cons |
|----------|----------|------|------|
| **Code Gen** | sqlc, protobuf | Type-safe, performant | Build step, generated code |
| Reflection | ORMs, JSON | Flexible, less boilerplate | Runtime errors, slower |

**Our choice**: Prefer code generation for critical paths (database, gRPC), accept reflection for less critical parts (JSON serialization).

### Trade-off 2: Monolith vs Microservices

This project is a **modular monolith**:

- Single deployable unit
- Clear package boundaries that could become services
- Shared database (for now)

**Why?**

- Easier to develop and debug
- No network overhead between "services"
- Can extract to microservices when needed (packages are already loosely coupled)

### Trade-off 3: Generic Packages vs Application-Specific

`pkg/` contains generic, reusable packages:

- Could be extracted to separate repositories
- More upfront design effort
- Benefits: reuse across projects, battle-tested

**Why not just inline everything?**

- Encourages thinking about interfaces
- Forces testability
- Creates portfolio of reusable components

---

## Interview Discussion Points

### "Why did you choose Clean Architecture?"

> "I chose Clean Architecture because it enforces separation of concerns and makes the codebase testable. The key insight is that business logic should be independent of frameworks and databases. In this project, the service layer has zero dependencies on Echo or PostgreSQL—it only knows about interfaces. This meant when I added gRPC support, I didn't touch a single line of business logic."

### "How do you handle errors in Go?"

> "Go's error handling philosophy is explicit errors over exceptions. I implemented a custom error types package (`pkg/apperrors`) that wraps errors with additional context while preserving the error chain. Each error type has an associated HTTP status code, so handlers can call `apperrors.GetHTTPStatus(err)` without switch statements. The errors implement `errors.Is` and `errors.As`, so callers can check specific error types."

### "Why use Generics? Isn't Go supposed to be simple?"

> "Generics, added in Go 1.18, solve specific problems well. I use them for:
>
> 1. **Result/Option types**: Avoid nil pointer bugs by making error handling explicit
> 2. **Worker pools**: Type-safe concurrent processing without interface{} casting
> 3. **Cache**: One implementation works for any type
>
> I don't overuse them—most code is still concrete types. Generics are tools, not goals."

### "How would you scale this system?"

> "The architecture supports multiple scaling strategies:
>
> **Horizontal scaling (current)**:
>
> - Stateless JWT auth (no session affinity needed)
> - Database connection pooling
> - K8s HPA based on CPU/memory
>
> **If we hit database limits**:
>
> - Read replicas for query-heavy workloads
> - Connection pooling with PgBouncer
> - Consider caching hot data (already have `pkg/cache`)
>
> **If we need microservices**:
>
> - Packages in `pkg/` are already service-ready
> - gRPC is implemented for internal communication
> - Circuit breakers prevent cascade failures"

### "What would you do differently?"

> "A few things I'd reconsider:
>
> 1. **Event sourcing for audit**: Currently using a simple audit log table. For complex compliance needs, event sourcing would provide complete history replay.
>
> 2. **CQRS for read-heavy workloads**: The current repository pattern treats reads and writes the same. For dashboards or reporting, separate read models would improve performance.
>
> 3. **More comprehensive integration tests**: Unit test coverage is high, but integration tests with real database are minimal. Would add testcontainers for true integration testing."

---

## Summary

This architecture prioritizes:

1. **Maintainability**: Clear structure, explicit dependencies
2. **Testability**: Interface-based design, dependency injection
3. **Extensibility**: Adding features doesn't break existing code
4. **Production readiness**: Observability, resilience, security built-in

The trade-offs accepted are:

- More upfront structure (worth it for team scalability)
- Code generation in build process (worth it for type safety)
- Slightly more verbose than minimal Go (worth it for clarity)
