package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zareh/go-api-starter/internal/model"
	"github.com/zareh/go-api-starter/internal/repository"
)

var (
	// ErrTodoNotFound is returned when a todo is not found
	ErrTodoNotFound = errors.New("todo not found")
	// ErrUnauthorized is returned when user is not authorized to access a todo
	ErrUnauthorized = errors.New("unauthorized")
)

// TodoService handles todo business logic
type TodoService struct {
	todoRepo *repository.TodoRepository
}

// NewTodoService creates a new TodoService
func NewTodoService(todoRepo *repository.TodoRepository) *TodoService {
	return &TodoService{todoRepo: todoRepo}
}

// Create creates a new todo
func (s *TodoService) Create(ctx context.Context, userID int64, title string) (*model.Todo, error) {
	return s.todoRepo.Create(ctx, userID, title)
}

// GetByID retrieves a todo by ID and verifies ownership
func (s *TodoService) GetByID(ctx context.Context, id, userID int64) (*model.Todo, error) {
	todo, err := s.todoRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTodoNotFound
		}
		return nil, err
	}

	// Verify ownership
	if todo.UserID != userID {
		return nil, ErrUnauthorized
	}

	return todo, nil
}

// ListByUserID retrieves all todos for a user
func (s *TodoService) ListByUserID(ctx context.Context, userID int64) ([]model.Todo, error) {
	todos, err := s.todoRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Return empty slice instead of nil
	if todos == nil {
		todos = []model.Todo{}
	}

	return todos, nil
}

// Update updates a todo
func (s *TodoService) Update(ctx context.Context, id, userID int64, req model.UpdateTodoRequest) (*model.Todo, error) {
	// Verify ownership
	todo, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	return s.todoRepo.Update(ctx, todo.ID, req.Title, req.Completed)
}

// Delete deletes a todo
func (s *TodoService) Delete(ctx context.Context, id, userID int64) error {
	// Verify ownership
	_, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	err = s.todoRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTodoNotFound
		}
		return err
	}

	return nil
}
