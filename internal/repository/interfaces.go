// Package repository provides interfaces and implementations for data access.
package repository

import (
	"context"

	"github.com/zareh/go-api-starter/internal/model"
)

// UserRepository defines the interface for user data operations.
// This interface enables dependency injection and testability.
type UserRepositoryInterface interface {
	// Create creates a new user with the given email and password hash.
	Create(ctx context.Context, email, passwordHash string) (*model.User, error)

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id int64) (*model.User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*model.User, error)

	// EmailExists checks if a user with the given email already exists.
	EmailExists(ctx context.Context, email string) (bool, error)

	// Update updates an existing user.
	Update(ctx context.Context, user *model.User) error

	// Delete removes a user by their ID.
	Delete(ctx context.Context, id int64) error
}

// TodoRepository defines the interface for todo data operations.
// This interface enables dependency injection and testability.
type TodoRepositoryInterface interface {
	// Create creates a new todo for the given user.
	Create(ctx context.Context, userID int64, title string) (*model.Todo, error)

	// GetByID retrieves a todo by its ID.
	GetByID(ctx context.Context, id int64) (*model.Todo, error)

	// GetByUserID retrieves all todos for a specific user.
	GetByUserID(ctx context.Context, userID int64) ([]model.Todo, error)

	// GetByUserIDWithPagination retrieves todos with pagination support.
	GetByUserIDWithPagination(ctx context.Context, userID int64, limit, offset int) ([]model.Todo, int64, error)

	// Update updates an existing todo.
	Update(ctx context.Context, id int64, title *string, completed *bool) (*model.Todo, error)

	// Delete removes a todo by its ID.
	Delete(ctx context.Context, id int64) error

	// CountByUserID returns the total count of todos for a user.
	CountByUserID(ctx context.Context, userID int64) (int64, error)
}

// Ensure implementations satisfy the interfaces
var _ UserRepositoryInterface = (*UserRepository)(nil)
var _ TodoRepositoryInterface = (*TodoRepository)(nil)
