package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/zareh/go-api-starter/internal/model"
	"github.com/zareh/go-api-starter/internal/service"
)

// TodoHandler handles todo endpoints
type TodoHandler struct {
	todoService *service.TodoService
}

// NewTodoHandler creates a new TodoHandler
func NewTodoHandler(todoService *service.TodoService) *TodoHandler {
	return &TodoHandler{todoService: todoService}
}

// getUserIDFromToken extracts user ID from JWT token
func getUserIDFromToken(c echo.Context) (int64, error) {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := int64(claims["user_id"].(float64))
	return userID, nil
}

// List returns all todos for the authenticated user
// GET /api/v1/todos
func (h *TodoHandler) List(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid token",
		})
	}

	todos, err := h.todoService.ListByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch todos",
		})
	}

	return c.JSON(http.StatusOK, model.TodoListResponse{
		Todos: todos,
		Total: len(todos),
	})
}

// Create creates a new todo
// POST /api/v1/todos
func (h *TodoHandler) Create(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid token",
		})
	}

	var req model.CreateTodoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Validation
	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "title is required",
		})
	}

	if len(req.Title) > 255 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "title must be at most 255 characters",
		})
	}

	todo, err := h.todoService.Create(c.Request().Context(), userID, req.Title)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create todo",
		})
	}

	return c.JSON(http.StatusCreated, todo)
}

// Get returns a single todo
// GET /api/v1/todos/:id
func (h *TodoHandler) Get(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid token",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid id",
		})
	}

	todo, err := h.todoService.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "todo not found",
			})
		}
		if errors.Is(err, service.ErrUnauthorized) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "todo not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch todo",
		})
	}

	return c.JSON(http.StatusOK, todo)
}

// Update updates a todo
// PUT /api/v1/todos/:id
func (h *TodoHandler) Update(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid token",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid id",
		})
	}

	var req model.UpdateTodoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Validation
	if req.Title != nil && len(*req.Title) > 255 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "title must be at most 255 characters",
		})
	}

	todo, err := h.todoService.Update(c.Request().Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "todo not found",
			})
		}
		if errors.Is(err, service.ErrUnauthorized) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "todo not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to update todo",
		})
	}

	return c.JSON(http.StatusOK, todo)
}

// Delete deletes a todo
// DELETE /api/v1/todos/:id
func (h *TodoHandler) Delete(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid token",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid id",
		})
	}

	err = h.todoService.Delete(c.Request().Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "todo not found",
			})
		}
		if errors.Is(err, service.ErrUnauthorized) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "todo not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to delete todo",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
