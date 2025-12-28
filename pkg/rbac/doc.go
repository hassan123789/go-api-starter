// Package rbac provides role-based access control (RBAC) for the API.
//
// This package implements a simple but effective RBAC system with three
// predefined roles (Admin, User, Viewer) and granular permissions.
//
// # Roles
//
// The system supports three roles:
//   - Admin: Full access to all resources and administrative functions
//   - User: Can manage their own resources (todos, profile)
//   - Viewer: Read-only access
//
// # Usage
//
// Basic permission check:
//
//	if rbac.HasPermission(userRole, rbac.PermTodoCreate) {
//	    // Create todo
//	}
//
// Using the enforcer:
//
//	enforcer := rbac.NewEnforcer()
//	if err := enforcer.Enforce(userRole, rbac.PermTodoDelete); err != nil {
//	    return err // Permission denied
//	}
//
// Checking resource ownership:
//
//	if enforcer.CanAccessResource(actorRole, actorID, resourceOwnerID) {
//	    // Allow access
//	}
//
// # Middleware Integration
//
// Use with Echo middleware:
//
//	func RequirePermission(perm rbac.Permission) echo.MiddlewareFunc {
//	    return func(next echo.HandlerFunc) echo.HandlerFunc {
//	        return func(c echo.Context) error {
//	            role := getRoleFromContext(c)
//	            if err := rbac.Enforce(role, perm); err != nil {
//	                return echo.NewHTTPError(http.StatusForbidden, "access denied")
//	            }
//	            return next(c)
//	        }
//	    }
//	}
package rbac
