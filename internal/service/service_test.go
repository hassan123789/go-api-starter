package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zareh/go-api-starter/internal/model"
)

func TestUpdateTodoRequest(t *testing.T) {
	title := "Updated Title"
	completed := true

	req := model.UpdateTodoRequest{
		Title:     &title,
		Completed: &completed,
	}

	assert.NotNil(t, req.Title)
	assert.Equal(t, "Updated Title", *req.Title)
	assert.NotNil(t, req.Completed)
	assert.True(t, *req.Completed)
}

func TestTodoListResponse(t *testing.T) {
	todos := []model.Todo{
		{ID: 1, Title: "Task 1"},
		{ID: 2, Title: "Task 2"},
	}

	response := model.TodoListResponse{
		Todos: todos,
		Total: len(todos),
	}

	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Todos, 2)
}

func TestUserToResponse(t *testing.T) {
	user := &model.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	response := user.ToResponse()

	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "test@example.com", response.Email)
}
