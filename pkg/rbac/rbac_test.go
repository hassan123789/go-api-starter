package rbac_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zareh/go-api-starter/pkg/rbac"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		role     rbac.Role
		expected bool
	}{
		{"admin is valid", rbac.RoleAdmin, true},
		{"user is valid", rbac.RoleUser, true},
		{"viewer is valid", rbac.RoleViewer, true},
		{"invalid role", rbac.Role("superadmin"), false},
		{"empty role", rbac.Role(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.IsValid())
		})
	}
}

func TestEnforcer_HasPermission(t *testing.T) {
	enforcer := rbac.NewEnforcer()

	tests := []struct {
		name       string
		role       rbac.Role
		permission rbac.Permission
		expected   bool
	}{
		// Admin permissions
		{"admin can create users", rbac.RoleAdmin, rbac.PermUserCreate, true},
		{"admin can delete users", rbac.RoleAdmin, rbac.PermUserDelete, true},
		{"admin can access admin", rbac.RoleAdmin, rbac.PermAdminAccess, true},
		{"admin can read audit logs", rbac.RoleAdmin, rbac.PermAuditLogRead, true},

		// User permissions
		{"user can create todos", rbac.RoleUser, rbac.PermTodoCreate, true},
		{"user can read todos", rbac.RoleUser, rbac.PermTodoRead, true},
		{"user can update todos", rbac.RoleUser, rbac.PermTodoUpdate, true},
		{"user can delete todos", rbac.RoleUser, rbac.PermTodoDelete, true},
		{"user cannot create users", rbac.RoleUser, rbac.PermUserCreate, false},
		{"user cannot delete users", rbac.RoleUser, rbac.PermUserDelete, false},
		{"user cannot access admin", rbac.RoleUser, rbac.PermAdminAccess, false},

		// Viewer permissions
		{"viewer can read todos", rbac.RoleViewer, rbac.PermTodoRead, true},
		{"viewer can read users", rbac.RoleViewer, rbac.PermUserRead, true},
		{"viewer cannot create todos", rbac.RoleViewer, rbac.PermTodoCreate, false},
		{"viewer cannot update todos", rbac.RoleViewer, rbac.PermTodoUpdate, false},
		{"viewer cannot delete todos", rbac.RoleViewer, rbac.PermTodoDelete, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enforcer.HasPermission(tt.role, tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnforcer_Enforce(t *testing.T) {
	enforcer := rbac.NewEnforcer()

	tests := []struct {
		name        string
		role        rbac.Role
		permission  rbac.Permission
		expectedErr error
	}{
		{"admin with admin access", rbac.RoleAdmin, rbac.PermAdminAccess, nil},
		{"user without admin access", rbac.RoleUser, rbac.PermAdminAccess, rbac.ErrPermissionDenied},
		{"invalid role", rbac.Role("invalid"), rbac.PermTodoRead, rbac.ErrInvalidRole},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enforcer.Enforce(tt.role, tt.permission)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnforcer_CanAccessResource(t *testing.T) {
	enforcer := rbac.NewEnforcer()

	tests := []struct {
		name            string
		actorRole       rbac.Role
		actorID         int64
		resourceOwnerID int64
		expected        bool
	}{
		{"admin can access any resource", rbac.RoleAdmin, 1, 2, true},
		{"user can access own resource", rbac.RoleUser, 1, 1, true},
		{"user cannot access other resource", rbac.RoleUser, 1, 2, false},
		{"viewer can access own resource", rbac.RoleViewer, 1, 1, true},
		{"viewer cannot access other resource", rbac.RoleViewer, 1, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enforcer.CanAccessResource(tt.actorRole, tt.actorID, tt.resourceOwnerID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsHigherOrEqual(t *testing.T) {
	tests := []struct {
		name     string
		role1    rbac.Role
		role2    rbac.Role
		expected bool
	}{
		{"admin >= admin", rbac.RoleAdmin, rbac.RoleAdmin, true},
		{"admin >= user", rbac.RoleAdmin, rbac.RoleUser, true},
		{"admin >= viewer", rbac.RoleAdmin, rbac.RoleViewer, true},
		{"user >= user", rbac.RoleUser, rbac.RoleUser, true},
		{"user >= viewer", rbac.RoleUser, rbac.RoleViewer, true},
		{"user < admin", rbac.RoleUser, rbac.RoleAdmin, false},
		{"viewer >= viewer", rbac.RoleViewer, rbac.RoleViewer, true},
		{"viewer < user", rbac.RoleViewer, rbac.RoleUser, false},
		{"viewer < admin", rbac.RoleViewer, rbac.RoleAdmin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rbac.IsHigherOrEqual(tt.role1, tt.role2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPermissions(t *testing.T) {
	enforcer := rbac.NewEnforcer()

	// Admin should have the most permissions
	adminPerms := enforcer.GetPermissions(rbac.RoleAdmin)
	assert.Greater(t, len(adminPerms), 0)

	// User should have fewer permissions than admin
	userPerms := enforcer.GetPermissions(rbac.RoleUser)
	assert.Greater(t, len(adminPerms), len(userPerms))

	// Viewer should have the fewest permissions
	viewerPerms := enforcer.GetPermissions(rbac.RoleViewer)
	assert.Greater(t, len(userPerms), len(viewerPerms))
}
