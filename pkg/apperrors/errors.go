// Package apperrors provides custom error types for the application.
// These errors support Go 1.13+ error wrapping with errors.Is and errors.As.
package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents application error codes
type ErrorCode string

const (
	// Authentication errors
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	// Validation errors
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"

	// Resource errors
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict      ErrorCode = "CONFLICT"

	// Permission errors
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeAccessDenied ErrorCode = "ACCESS_DENIED"

	// Server errors
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabase ErrorCode = "DATABASE_ERROR"
	ErrCodeTimeout  ErrorCode = "TIMEOUT"
)

// AppError is a custom error type that implements the error interface
// and supports error wrapping for use with errors.Is and errors.As.
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Err        error     `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error for use with errors.Is and errors.As.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is an AppError with the same code.
func (e *AppError) Is(target error) bool {
	var appErr *AppError
	if errors.As(target, &appErr) {
		return e.Code == appErr.Code
	}
	return false
}

// WithDetails returns a copy of the error with additional details.
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		Details:    details,
		HTTPStatus: e.HTTPStatus,
		Err:        e.Err,
	}
}

// Wrap wraps an error with the AppError.
func (e *AppError) Wrap(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		Details:    e.Details,
		HTTPStatus: e.HTTPStatus,
		Err:        err,
	}
}

// Sentinel errors for common cases
var (
	// Authentication errors
	ErrUnauthorized = &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    "authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInvalidToken = &AppError{
		Code:       ErrCodeInvalidToken,
		Message:    "invalid authentication token",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrTokenExpired = &AppError{
		Code:       ErrCodeTokenExpired,
		Message:    "authentication token has expired",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInvalidCredentials = &AppError{
		Code:       ErrCodeInvalidCredentials,
		Message:    "invalid email or password",
		HTTPStatus: http.StatusUnauthorized,
	}

	// Validation errors
	ErrValidation = &AppError{
		Code:       ErrCodeValidation,
		Message:    "validation failed",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidInput = &AppError{
		Code:       ErrCodeInvalidInput,
		Message:    "invalid input provided",
		HTTPStatus: http.StatusBadRequest,
	}

	// Resource errors
	ErrNotFound = &AppError{
		Code:       ErrCodeNotFound,
		Message:    "resource not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrAlreadyExists = &AppError{
		Code:       ErrCodeAlreadyExists,
		Message:    "resource already exists",
		HTTPStatus: http.StatusConflict,
	}

	ErrConflict = &AppError{
		Code:       ErrCodeConflict,
		Message:    "resource conflict",
		HTTPStatus: http.StatusConflict,
	}

	// Permission errors
	ErrForbidden = &AppError{
		Code:       ErrCodeForbidden,
		Message:    "access forbidden",
		HTTPStatus: http.StatusForbidden,
	}

	ErrAccessDenied = &AppError{
		Code:       ErrCodeAccessDenied,
		Message:    "access denied to this resource",
		HTTPStatus: http.StatusForbidden,
	}

	// Server errors
	ErrInternal = &AppError{
		Code:       ErrCodeInternal,
		Message:    "internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrDatabase = &AppError{
		Code:       ErrCodeDatabase,
		Message:    "database error",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrTimeout = &AppError{
		Code:       ErrCodeTimeout,
		Message:    "operation timed out",
		HTTPStatus: http.StatusGatewayTimeout,
	}
)

// New creates a new AppError with the given code, message, and HTTP status.
func New(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// NewNotFound creates a not found error for a specific resource.
func NewNotFound(resource string, id interface{}) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		Details:    fmt.Sprintf("id: %v", id),
		HTTPStatus: http.StatusNotFound,
	}
}

// NewValidation creates a validation error with details.
func NewValidation(field, reason string) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    fmt.Sprintf("validation failed for field '%s'", field),
		Details:    reason,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewAlreadyExists creates an already exists error for a specific resource.
func NewAlreadyExists(resource, field string, value interface{}) *AppError {
	return &AppError{
		Code:       ErrCodeAlreadyExists,
		Message:    fmt.Sprintf("%s already exists", resource),
		Details:    fmt.Sprintf("%s: %v", field, value),
		HTTPStatus: http.StatusConflict,
	}
}

// IsAppError checks if the error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error chain.
func GetAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetHTTPStatus returns the HTTP status code for an error.
// If the error is not an AppError, it returns 500 Internal Server Error.
func GetHTTPStatus(err error) int {
	if appErr, ok := GetAppError(err); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}
