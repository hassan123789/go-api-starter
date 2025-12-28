// Package middleware provides HTTP middleware for Echo framework.
package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zareh/go-api-starter/pkg/rbac"
)

// RBACConfig holds configuration for RBAC middleware.
type RBACConfig struct {
	// Enforcer is the RBAC enforcer instance.
	Enforcer *rbac.Enforcer

	// GetUserRole extracts the user's role from the context.
	// Should return the role and user ID.
	GetUserRole func(c echo.Context) (rbac.Role, int64, error)

	// Skipper defines a function to skip middleware.
	Skipper func(c echo.Context) bool
}

// RequirePermission returns middleware that requires a specific permission.
func RequirePermission(cfg RBACConfig, permission rbac.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			role, _, err := cfg.GetUserRole(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}

			if !cfg.Enforcer.HasPermission(role, permission) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(c)
		}
	}
}

// RequireRole returns middleware that requires a minimum role level.
func RequireRole(cfg RBACConfig, minRole rbac.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			role, _, err := cfg.GetUserRole(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}

			if !rbac.IsHigherOrEqual(role, minRole) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient role level")
			}

			return next(c)
		}
	}
}

// RequireResourceAccess returns middleware that checks resource ownership.
func RequireResourceAccess(cfg RBACConfig, resourceType string, getResourceOwnerID func(c echo.Context) (int64, error)) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			role, userID, err := cfg.GetUserRole(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}

			ownerID, err := getResourceOwnerID(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "resource not found")
			}

			if !cfg.Enforcer.CanAccessResource(role, userID, ownerID) {
				return echo.NewHTTPError(http.StatusForbidden, "access denied")
			}

			return next(c)
		}
	}
}

// RequireAdmin is a convenience middleware for admin-only routes.
func RequireAdmin(cfg RBACConfig) echo.MiddlewareFunc {
	return RequireRole(cfg, rbac.RoleAdmin)
}

// RequireAnyPermission returns middleware that requires any of the specified permissions.
func RequireAnyPermission(cfg RBACConfig, permissions ...rbac.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			role, _, err := cfg.GetUserRole(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}

			for _, perm := range permissions {
				if cfg.Enforcer.HasPermission(role, perm) {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

// RequireAllPermissions returns middleware that requires all specified permissions.
func RequireAllPermissions(cfg RBACConfig, permissions ...rbac.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			role, _, err := cfg.GetUserRole(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}

			for _, perm := range permissions {
				if !cfg.Enforcer.HasPermission(role, perm) {
					return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
				}
			}

			return next(c)
		}
	}
}
