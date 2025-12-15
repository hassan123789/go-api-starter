package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/zareh/go-api-starter/internal/model"
)

// TodoRepository handles todo database operations
type TodoRepository struct {
	db *sql.DB
}

// NewTodoRepository creates a new TodoRepository
func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create creates a new todo
func (r *TodoRepository) Create(ctx context.Context, userID int64, title string) (*model.Todo, error) {
	query := `
		INSERT INTO todos (user_id, title, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, completed, created_at, updated_at
	`

	now := time.Now()
	todo := &model.Todo{}

	err := r.db.QueryRowContext(ctx, query, userID, title, false, now, now).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return todo, nil
}

// GetByID retrieves a todo by ID
func (r *TodoRepository) GetByID(ctx context.Context, id int64) (*model.Todo, error) {
	query := `
		SELECT id, user_id, title, completed, created_at, updated_at
		FROM todos
		WHERE id = $1
	`

	todo := &model.Todo{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return todo, nil
}

// GetByUserID retrieves all todos for a user
func (r *TodoRepository) GetByUserID(ctx context.Context, userID int64) ([]model.Todo, error) {
	query := `
		SELECT id, user_id, title, completed, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		var todo model.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Title,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

// Update updates a todo
func (r *TodoRepository) Update(ctx context.Context, id int64, title *string, completed *bool) (*model.Todo, error) {
	// First get the existing todo
	todo, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if title != nil {
		todo.Title = *title
	}
	if completed != nil {
		todo.Completed = *completed
	}
	todo.UpdatedAt = time.Now()

	query := `
		UPDATE todos
		SET title = $1, completed = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, user_id, title, completed, created_at, updated_at
	`

	updatedTodo := &model.Todo{}
	err = r.db.QueryRowContext(ctx, query, todo.Title, todo.Completed, todo.UpdatedAt, id).Scan(
		&updatedTodo.ID,
		&updatedTodo.UserID,
		&updatedTodo.Title,
		&updatedTodo.Completed,
		&updatedTodo.CreatedAt,
		&updatedTodo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return updatedTodo, nil
}

// Delete deletes a todo
func (r *TodoRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM todos WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
