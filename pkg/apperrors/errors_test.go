package apperrors_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/zareh/go-api-starter/pkg/apperrors"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *apperrors.AppError
		contains string
	}{
		{
			name:     "error without wrapped error",
			err:      apperrors.ErrNotFound,
			contains: "NOT_FOUND",
		},
		{
			name:     "error with wrapped error",
			err:      apperrors.ErrDatabase.Wrap(errors.New("connection failed")),
			contains: "connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			if errMsg == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := apperrors.ErrDatabase.Wrap(originalErr)

	unwrapped := errors.Unwrap(wrappedErr)
	if unwrapped != originalErr {
		t.Errorf("expected %v, got %v", originalErr, unwrapped)
	}
}

func TestAppError_Is(t *testing.T) {
	err := apperrors.ErrNotFound.WithDetails("user not found")

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Error("expected errors.Is to return true for same error code")
	}

	if errors.Is(err, apperrors.ErrUnauthorized) {
		t.Error("expected errors.Is to return false for different error code")
	}
}

func TestAppError_As(t *testing.T) {
	originalErr := apperrors.ErrNotFound.WithDetails("test details")
	wrappedErr := apperrors.ErrInternal.Wrap(originalErr)

	var appErr *apperrors.AppError
	if !errors.As(wrappedErr, &appErr) {
		t.Error("expected errors.As to extract AppError")
	}

	if appErr.Code != apperrors.ErrCodeInternal {
		t.Errorf("expected code %s, got %s", apperrors.ErrCodeInternal, appErr.Code)
	}
}

func TestNewNotFound(t *testing.T) {
	err := apperrors.NewNotFound("user", 123)

	if err.Code != apperrors.ErrCodeNotFound {
		t.Errorf("expected code %s, got %s", apperrors.ErrCodeNotFound, err.Code)
	}

	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}
}

func TestNewValidation(t *testing.T) {
	err := apperrors.NewValidation("email", "invalid format")

	if err.Code != apperrors.ErrCodeValidation {
		t.Errorf("expected code %s, got %s", apperrors.ErrCodeValidation, err.Code)
	}

	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "app error",
			err:            apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "wrapped app error",
			err:            apperrors.ErrUnauthorized.Wrap(errors.New("token invalid")),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "standard error",
			err:            errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := apperrors.GetHTTPStatus(tt.err)
			if status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}
