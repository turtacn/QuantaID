package policy

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// Repository defines the persistence interface for authorization policies.
// It outlines the CRUD operations and query methods needed to manage and evaluate policies.
type Repository interface {
	// CreatePolicy saves a new authorization policy to the database.
	CreatePolicy(ctx context.Context, policy *types.Policy) error
	// GetPolicyByID retrieves a policy by its unique ID.
	GetPolicyByID(ctx context.Context, id string) (*types.Policy, error)
	// UpdatePolicy modifies an existing policy.
	UpdatePolicy(ctx context.Context, policy *types.Policy) error
	// DeletePolicy removes a policy from the database by its ID.
	DeletePolicy(ctx context.Context, id string) error
	// ListPolicies retrieves a paginated list of all policies.
	ListPolicies(ctx context.Context, pq types.PaginationQuery) ([]*types.Policy, error)
	// FindPoliciesForSubject retrieves all policies that apply to a given subject (e.g., a user or group).
	FindPoliciesForSubject(ctx context.Context, subject string) ([]*types.Policy, error)
	// FindPoliciesForResource retrieves all policies that apply to a given resource.
	FindPoliciesForResource(ctx context.Context, resource string) ([]*types.Policy, error)
	// FindPoliciesForAction retrieves all policies that are associated with a given action.
	FindPoliciesForAction(ctx context.Context, action string) ([]*types.Policy, error)
}

// PolicyTemplateRepository defines an interface for managing reusable policy templates.
// Templates allow for the easy creation of standardized policies.
type PolicyTemplateRepository interface {
	// CreateTemplate saves a new policy template.
	CreateTemplate(ctx context.Context, template *types.Policy) error
	// GetTemplate retrieves a policy template by its ID.
	GetTemplate(ctx context.Context, id string) (*types.Policy, error)
	// ListTemplates retrieves all available policy templates.
	ListTemplates(ctx context.Context) ([]*types.Policy, error)
}

// RBACRepository defines the interface for managing Roles, Permissions, and their assignments.
type RBACRepository interface {
	// Role management
	CreateRole(ctx context.Context, role *Role) error
	GetRoleByCode(ctx context.Context, code string) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, roleID uint) error
	ListRoles(ctx context.Context) ([]*Role, error)

	// Permission management
	CreatePermission(ctx context.Context, permission *Permission) error
	GetPermission(ctx context.Context, resource, action string) (*Permission, error)
	ListPermissions(ctx context.Context) ([]*Permission, error)

	// Assignment management
	AddPermissionToRole(ctx context.Context, roleID, permissionID uint) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint) error
	AssignRoleToUser(ctx context.Context, userID string, roleID uint) error
	UnassignRoleFromUser(ctx context.Context, userID string, roleID uint) error

	// Query methods
	GetRolesForUser(ctx context.Context, userID string) ([]*Role, error)
	GetPermissionsForRole(ctx context.Context, roleID uint) ([]*Permission, error)
}
