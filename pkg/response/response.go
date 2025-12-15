package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response with data
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Error returns an error response
func Error(c echo.Context, status int, message string) error {
	return c.JSON(status, ErrorResponse{
		Error: message,
	})
}

// BadRequest returns a 400 Bad Request response
func BadRequest(c echo.Context, message string) error {
	return Error(c, http.StatusBadRequest, message)
}

// Unauthorized returns a 401 Unauthorized response
func Unauthorized(c echo.Context, message string) error {
	return Error(c, http.StatusUnauthorized, message)
}

// NotFound returns a 404 Not Found response
func NotFound(c echo.Context, message string) error {
	return Error(c, http.StatusNotFound, message)
}

// InternalServerError returns a 500 Internal Server Error response
func InternalServerError(c echo.Context, message string) error {
	return Error(c, http.StatusInternalServerError, message)
}

// Success returns a success response
func Success(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, data)
}

// Created returns a 201 Created response
func Created(c echo.Context, data interface{}) error {
	return Success(c, http.StatusCreated, data)
}

// OK returns a 200 OK response
func OK(c echo.Context, data interface{}) error {
	return Success(c, http.StatusOK, data)
}

// NoContent returns a 204 No Content response
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
