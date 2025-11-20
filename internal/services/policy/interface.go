package policy

import (
	"context"

	"github.com/turtacn/QuantaID/internal/domain/policy"
)

// PolicyService defines the interface for managing policies.
type PolicyService interface {
	CreateRole(ctx context.Context, role *policy.Role) error
	ListRoles(ctx context.Context) ([]*policy.Role, error)
	UpdateRole(ctx context.Context, role *policy.Role) error
	DeleteRole(ctx context.Context, roleID uint) error

	CreatePermission(ctx context.Context, permission *policy.Permission) error
	ListPermissions(ctx context.Context) ([]*policy.Permission, error)
	AddPermissionToRole(ctx context.Context, roleID, permissionID uint) error

	AssignRoleToUser(ctx context.Context, userID string, roleID uint) error
	UnassignRoleFromUser(ctx context.Context, userID string, roleID uint) error
	// ... other service methods
}
