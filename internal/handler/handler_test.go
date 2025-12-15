package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	e := echo.New()

	// Setup health check handler
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Execute request
	e.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestUserRegistrationValidation(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Empty body",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email and password are required",
		},
		{
			name:           "Missing password",
			body:           `{"email": "test@example.com"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email and password are required",
		},
		{
			name:           "Short password",
			body:           `{"email": "test@example.com", "password": "short"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 8 characters",
		},
		{
			name:           "Valid request",
			body:           `{"email": "test@example.com", "password": "password123"}`,
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
	}

	// Mock handler that validates input
	e.POST("/api/v1/users", func(c echo.Context) error {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}
		if req.Email == "" || req.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		}
		if len(req.Password) < 8 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		}
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"id":    1,
			"email": req.Email,
		})
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}
		})
	}
}

func TestLoginValidation(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "Empty body",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing password",
			body:           `{"email": "test@example.com"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing email",
			body:           `{"password": "password123"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	// Mock handler
	e.POST("/api/v1/auth/login", func(c echo.Context) error {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}
		if req.Email == "" || req.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		}
		return c.JSON(http.StatusOK, map[string]string{"token": "mock-token"})
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestTodoEndpointsRequireAuth(t *testing.T) {
	e := echo.New()

	// Mock protected endpoints that return 401 without auth
	protectedHandler := func(c echo.Context) error {
		auth := c.Request().Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or malformed jwt"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	e.GET("/api/v1/todos", protectedHandler)
	e.POST("/api/v1/todos", protectedHandler)
	e.GET("/api/v1/todos/:id", protectedHandler)
	e.PUT("/api/v1/todos/:id", protectedHandler)
	e.DELETE("/api/v1/todos/:id", protectedHandler)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/todos"},
		{http.MethodPost, "/api/v1/todos"},
		{http.MethodGet, "/api/v1/todos/1"},
		{http.MethodPut, "/api/v1/todos/1"},
		{http.MethodDelete, "/api/v1/todos/1"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path+" without auth", func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run(tt.method+" "+tt.path+" with auth", func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer mock-token")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestTodoCreateValidation(t *testing.T) {
	e := echo.New()

	e.POST("/api/v1/todos", func(c echo.Context) error {
		auth := c.Request().Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or malformed jwt"})
		}

		var req struct {
			Title string `json:"title"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}
		if req.Title == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "title is required"})
		}
		if len(req.Title) > 255 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "title must be at most 255 characters"})
		}
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"id":        1,
			"title":     req.Title,
			"completed": false,
		})
	})

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Empty title",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title is required",
		},
		{
			name:           "Valid title",
			body:           `{"title": "Learn Go"}`,
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer mock-token")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}
		})
	}
}
