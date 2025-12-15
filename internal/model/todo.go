package model

import "time"

// Todo represents a todo item
type Todo struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateTodoRequest represents the request body for creating a todo
type CreateTodoRequest struct {
	Title string `json:"title" validate:"required,max=255"`
}

// UpdateTodoRequest represents the request body for updating a todo
type UpdateTodoRequest struct {
	Title     *string `json:"title,omitempty" validate:"omitempty,max=255"`
	Completed *bool   `json:"completed,omitempty"`
}

// TodoListResponse represents the response body for a list of todos
type TodoListResponse struct {
	Todos []Todo `json:"todos"`
	Total int    `json:"total"`
}
