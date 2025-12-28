// Package rbac provides role-based access control for the API.
package rbac

import (
	"errors"
	"slices"
)

// Role represents a user role in the system.
type Role string

// Predefined roles
const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleViewer Role = "viewer"
)

// AllRoles returns all valid roles.
func AllRoles() []Role {
	return []Role{RoleAdmin, RoleUser, RoleViewer}
}

// IsValid checks if the role is valid.
func (r Role) IsValid() bool {
	return slices.Contains(AllRoles(), r)
}

// String returns the string representation of the role.
func (r Role) String() string {
	return string(r)
}

// Permission represents a specific action that can be performed.
type Permission string

// Predefined permissions
const (
	// User permissions
	PermUserCreate Permission = "user:create"
	PermUserRead   Permission = "user:read"
	PermUserUpdate Permission = "user:update"
	PermUserDelete Permission = "user:delete"

	// Todo permissions
	PermTodoCreate Permission = "todo:create"
	PermTodoRead   Permission = "todo:read"
	PermTodoUpdate Permission = "todo:update"
	PermTodoDelete Permission = "todo:delete"

	// Admin permissions
	PermAdminAccess   Permission = "admin:access"
	PermAuditLogRead  Permission = "audit:read"
	PermSystemMetrics Permission = "system:metrics"
)

// rolePermissions maps roles to their allowed permissions.
var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		// Admin has all permissions
		PermUserCreate, PermUserRead, PermUserUpdate, PermUserDelete,
		PermTodoCreate, PermTodoRead, PermTodoUpdate, PermTodoDelete,
		PermAdminAccess, PermAuditLogRead, PermSystemMetrics,
	},
	RoleUser: {
		// Regular users can manage their own todos
		PermUserRead, PermUserUpdate,
		PermTodoCreate, PermTodoRead, PermTodoUpdate, PermTodoDelete,
	},
	RoleViewer: {
		// Viewers can only read
		PermUserRead,
		PermTodoRead,
	},
}

// Errors
var (
	ErrInvalidRole      = errors.New("invalid role")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNoRoleAssigned   = errors.New("no role assigned")
)

// Enforcer handles permission checks.
type Enforcer struct {
	permissions map[Role]map[Permission]struct{}
}

// NewEnforcer creates a new RBAC enforcer.
func NewEnforcer() *Enforcer {
	e := &Enforcer{
		permissions: make(map[Role]map[Permission]struct{}),
	}

	// Build permission lookup maps
	for role, perms := range rolePermissions {
		e.permissions[role] = make(map[Permission]struct{})
		for _, perm := range perms {
			e.permissions[role][perm] = struct{}{}
		}
	}

	return e
}

// HasPermission checks if a role has a specific permission.
func (e *Enforcer) HasPermission(role Role, permission Permission) bool {
	if perms, ok := e.permissions[role]; ok {
		_, has := perms[permission]
		return has
	}
	return false
}

// Enforce checks if a role has permission, returning error if not.
func (e *Enforcer) Enforce(role Role, permission Permission) error {
	if !role.IsValid() {
		return ErrInvalidRole
	}
	if !e.HasPermission(role, permission) {
		return ErrPermissionDenied
	}
	return nil
}

// GetPermissions returns all permissions for a role.
func (e *Enforcer) GetPermissions(role Role) []Permission {
	if perms, ok := rolePermissions[role]; ok {
		return perms
	}
	return nil
}

// CanAccessResource checks if a user can access another user's resource.
// Admin can access any resource, others can only access their own.
func (e *Enforcer) CanAccessResource(actorRole Role, actorID, resourceOwnerID int64) bool {
	// Admins can access anything
	if actorRole == RoleAdmin {
		return true
	}
	// Others can only access their own resources
	return actorID == resourceOwnerID
}

// RoleHierarchy defines role hierarchy for comparison.
var RoleHierarchy = map[Role]int{
	RoleViewer: 0,
	RoleUser:   1,
	RoleAdmin:  2,
}

// IsHigherOrEqual checks if role1 is higher or equal to role2 in hierarchy.
func IsHigherOrEqual(role1, role2 Role) bool {
	return RoleHierarchy[role1] >= RoleHierarchy[role2]
}

// DefaultEnforcer is the default RBAC enforcer instance.
var DefaultEnforcer = NewEnforcer()

// HasPermission is a convenience function using the default enforcer.
func HasPermission(role Role, permission Permission) bool {
	return DefaultEnforcer.HasPermission(role, permission)
}

// Enforce is a convenience function using the default enforcer.
func Enforce(role Role, permission Permission) error {
	return DefaultEnforcer.Enforce(role, permission)
}
