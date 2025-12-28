package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/zareh/go-api-starter/internal/middleware"
	"github.com/zareh/go-api-starter/pkg/rbac"
)

func setupRBACTest() (*echo.Echo, middleware.RBACConfig) {
	e := echo.New()
	enforcer := rbac.NewEnforcer()

	cfg := middleware.RBACConfig{
		Enforcer: enforcer,
		GetUserRole: func(c echo.Context) (rbac.Role, int64, error) {
			role := c.Request().Header.Get("X-User-Role")
			userIDStr := c.Request().Header.Get("X-User-ID")
			if role == "" {
				return "", 0, errors.New("no role")
			}
			var userID int64 = 1
			if userIDStr == "2" {
				userID = 2
			}
			return rbac.Role(role), userID, nil
		},
	}

	return e, cfg
}

func TestRequirePermission_Allowed(t *testing.T) {
	e, cfg := setupRBACTest()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Role", string(rbac.RoleUser))
	req.Header.Set("X-User-ID", "1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware.RequirePermission(cfg, rbac.PermTodoRead)(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequirePermission_Denied(t *testing.T) {
	e, cfg := setupRBACTest()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Role", string(rbac.RoleViewer))
	req.Header.Set("X-User-ID", "1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware.RequirePermission(cfg, rbac.PermTodoCreate)(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, he.Code)
}

func TestRequirePermission_Unauthorized(t *testing.T) {
	e, cfg := setupRBACTest()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No role header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware.RequirePermission(cfg, rbac.PermTodoRead)(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}

func TestRequireRole_Admin(t *testing.T) {
	e, cfg := setupRBACTest()

	tests := []struct {
		name       string
		userRole   string
		minRole    rbac.Role
		shouldPass bool
	}{
		{"admin accessing admin route", string(rbac.RoleAdmin), rbac.RoleAdmin, true},
		{"user accessing admin route", string(rbac.RoleUser), rbac.RoleAdmin, false},
		{"viewer accessing admin route", string(rbac.RoleViewer), rbac.RoleAdmin, false},
		{"admin accessing user route", string(rbac.RoleAdmin), rbac.RoleUser, true},
		{"user accessing user route", string(rbac.RoleUser), rbac.RoleUser, true},
		{"viewer accessing viewer route", string(rbac.RoleViewer), rbac.RoleViewer, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-User-Role", tt.userRole)
			req.Header.Set("X-User-ID", "1")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := middleware.RequireRole(cfg, tt.minRole)(func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			})

			err := handler(c)
			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	e, cfg := setupRBACTest()

	tests := []struct {
		name       string
		userRole   string
		shouldPass bool
	}{
		{"admin can access", string(rbac.RoleAdmin), true},
		{"user cannot access", string(rbac.RoleUser), false},
		{"viewer cannot access", string(rbac.RoleViewer), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-User-Role", tt.userRole)
			req.Header.Set("X-User-ID", "1")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := middleware.RequireAdmin(cfg)(func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			})

			err := handler(c)
			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRequireResourceAccess(t *testing.T) {
	e, cfg := setupRBACTest()

	getOwnerID := func(c echo.Context) (int64, error) {
		// Resource is owned by user 1
		return 1, nil
	}

	tests := []struct {
		name       string
		userRole   string
		userID     string
		shouldPass bool
	}{
		{"owner can access own resource", string(rbac.RoleUser), "1", true},
		{"non-owner user cannot access", string(rbac.RoleUser), "2", false},
		{"admin can access any resource", string(rbac.RoleAdmin), "2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-User-Role", tt.userRole)
			req.Header.Set("X-User-ID", tt.userID)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := middleware.RequireResourceAccess(cfg, "todo", getOwnerID)(func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			})

			err := handler(c)
			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRequireAnyPermission(t *testing.T) {
	e, cfg := setupRBACTest()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Role", string(rbac.RoleViewer))
	req.Header.Set("X-User-ID", "1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Viewer has TodoRead but not TodoCreate
	handler := middleware.RequireAnyPermission(cfg, rbac.PermTodoCreate, rbac.PermTodoRead)(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	assert.NoError(t, err)
}

func TestRequireAllPermissions(t *testing.T) {
	e, cfg := setupRBACTest()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Role", string(rbac.RoleViewer))
	req.Header.Set("X-User-ID", "1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Viewer has TodoRead but not TodoCreate
	handler := middleware.RequireAllPermissions(cfg, rbac.PermTodoCreate, rbac.PermTodoRead)(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	assert.Error(t, err) // Should fail because viewer doesn't have TodoCreate
}

func TestSkipper(t *testing.T) {
	e, cfg := setupRBACTest()
	cfg.Skipper = func(c echo.Context) bool {
		return c.Path() == "/health"
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	// No role header - would normally fail
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health")

	handler := middleware.RequirePermission(cfg, rbac.PermTodoRead)(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	assert.NoError(t, err) // Should pass because of skipper
}
