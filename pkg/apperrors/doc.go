// Package apperrors provides custom error types for structured error handling.
//
// # Overview
//
// This package implements a standardized error system for API applications with:
//   - Predefined error types for common scenarios (NotFound, Validation, etc.)
//   - HTTP status code mapping
//   - Error message customization
//   - Error wrapping support
//
// # Predefined Errors
//
// The package provides several predefined error types:
//
//	ErrNotFound      - Resource not found (404)
//	ErrBadRequest    - Invalid request (400)
//	ErrUnauthorized  - Authentication required (401)
//	ErrForbidden     - Access denied (403)
//	ErrConflict      - Resource conflict (409)
//	ErrInternal      - Internal server error (500)
//	ErrValidation    - Validation failure (422)
//
// # Basic Usage
//
//	func GetUser(id int64) (*User, error) {
//	    user, err := repo.FindByID(id)
//	    if err != nil {
//	        if errors.Is(err, sql.ErrNoRows) {
//	            return nil, apperrors.ErrNotFound.WithMessage("user not found")
//	        }
//	        return nil, apperrors.ErrInternal.Wrap(err)
//	    }
//	    return user, nil
//	}
//
// # Error Handling in Handlers
//
//	func handleError(c echo.Context, err error) error {
//	    var appErr *apperrors.AppError
//	    if errors.As(err, &appErr) {
//	        return c.JSON(appErr.HTTPStatus(), appErr)
//	    }
//	    return c.JSON(500, map[string]string{"error": "internal error"})
//	}
//
// # Custom Errors
//
//	err := apperrors.New(400, "CUSTOM_ERROR", "custom error message")
package apperrors
