package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a JWT token for testing
func createTestToken(userID int64) *jwt.Token {
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"exp":     float64(time.Now().Add(time.Hour).Unix()),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
}

// Test getUserIDFromToken function
func TestGetUserIDFromToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set valid token
	c.Set("user", createTestToken(123))

	userID, err := getUserIDFromToken(c)
	require.NoError(t, err)
	assert.Equal(t, int64(123), userID)
}

func TestGetUserIDFromToken_NoToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No token set

	userID, err := getUserIDFromToken(c)
	assert.Error(t, err)
	assert.Equal(t, int64(0), userID)
}

func TestGetUserIDFromToken_InvalidTokenType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set wrong type
	c.Set("user", "invalid")

	userID, err := getUserIDFromToken(c)
	assert.Error(t, err)
	assert.Equal(t, int64(0), userID)
}

func TestGetUserIDFromToken_InvalidClaimsType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create token with RegisteredClaims instead of MapClaims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{})
	c.Set("user", token)

	userID, err := getUserIDFromToken(c)
	assert.Error(t, err)
	assert.Equal(t, int64(0), userID)
}

func TestGetUserIDFromToken_InvalidUserIDType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create token with wrong user_id type
	claims := jwt.MapClaims{
		"user_id": "not-a-number",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	c.Set("user", token)

	userID, err := getUserIDFromToken(c)
	assert.Error(t, err)
	assert.Equal(t, int64(0), userID)
}

// Test NewTodoHandler constructor
func TestNewTodoHandler(t *testing.T) {
	handler := NewTodoHandler(nil)
	assert.NotNil(t, handler)
	assert.Nil(t, handler.todoService)
}

// Tests for validation paths in handlers without service calls
func TestTodoHandler_Create_ValidationPaths(t *testing.T) {
	handler := NewTodoHandler(nil)

	testCases := []struct {
		name       string
		body       string
		setupToken bool
		wantStatus int
	}{
		{
			name:       "Empty title returns 400",
			body:       `{"title": ""}`,
			setupToken: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Title too long returns 400",
			body:       `{"title": "` + strings.Repeat("a", 256) + `"}`,
			setupToken: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid JSON returns 400",
			body:       `{invalid`,
			setupToken: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "No token returns 401",
			body:       `{"title": "Test"}`,
			setupToken: false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tc.setupToken {
				c.Set("user", createTestToken(1))
			}

			err := handler.Create(c)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

// testIDValidation is a helper to test ID param and token validation for handlers
// that accept an ID parameter (Get, Update, Delete).
func testIDValidation(t *testing.T, method string, handlerFunc func(handler *TodoHandler) func(echo.Context) error, hasBody bool) {
	t.Helper()
	handler := NewTodoHandler(nil)

	testCases := []struct {
		name       string
		idParam    string
		setupToken bool
		wantStatus int
	}{
		{
			name:       "Invalid ID returns 400",
			idParam:    "abc",
			setupToken: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "No token returns 401",
			idParam:    "1",
			setupToken: false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			var req *http.Request
			if hasBody {
				req = httptest.NewRequest(method, "/api/v1/todos/"+tc.idParam, strings.NewReader(`{"title":"Test"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			} else {
				req = httptest.NewRequest(method, "/api/v1/todos/"+tc.idParam, nil)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.idParam)

			if tc.setupToken {
				c.Set("user", createTestToken(1))
			}

			err := handlerFunc(handler)(c)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestTodoHandler_Get_ValidationPaths(t *testing.T) {
	testIDValidation(t, http.MethodGet, func(h *TodoHandler) func(echo.Context) error {
		return h.Get
	}, false)
}

func TestTodoHandler_Delete_ValidationPaths(t *testing.T) {
	testIDValidation(t, http.MethodDelete, func(h *TodoHandler) func(echo.Context) error {
		return h.Delete
	}, false)
}

func TestTodoHandler_Update_ValidationPaths(t *testing.T) {
	handler := NewTodoHandler(nil)

	testCases := []struct {
		name       string
		idParam    string
		body       string
		setupToken bool
		wantStatus int
	}{
		{
			name:       "Invalid ID returns 400",
			idParam:    "abc",
			body:       `{"title": "Test"}`,
			setupToken: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Title too long returns 400",
			idParam:    "1",
			body:       `{"title": "` + strings.Repeat("a", 256) + `"}`,
			setupToken: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid JSON returns 400",
			idParam:    "1",
			body:       `{invalid`,
			setupToken: true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "No token returns 401",
			idParam:    "1",
			body:       `{"title": "Test"}`,
			setupToken: false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/todos/"+tc.idParam, strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.idParam)

			if tc.setupToken {
				c.Set("user", createTestToken(1))
			}

			err := handler.Update(c)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestTodoHandler_List_Unauthorized(t *testing.T) {
	handler := NewTodoHandler(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No token set

	err := handler.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
