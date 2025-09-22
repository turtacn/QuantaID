package policy

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// Repository defines the interface for storing and retrieving authorization policies.
type Repository interface {
	CreatePolicy(ctx context.Context, policy *types.Policy) error
	GetPolicyByID(ctx context.Context, id string) (*types.Policy, error)
	UpdatePolicy(ctx context.Context, policy *types.Policy) error
	DeletePolicy(ctx context.Context, id string) error
	ListPolicies(ctx context.Context, pq types.PaginationQuery) ([]*types.Policy, error)
	FindPoliciesForSubject(ctx context.Context, subject string) ([]*types.Policy, error)
	FindPoliciesForResource(ctx context.Context, resource string) ([]*types.Policy, error)
	FindPoliciesForAction(ctx context.Context, action string) ([]*types.Policy, error)
}

// PolicyTemplateRepository defines an interface for managing reusable policy templates.
type PolicyTemplateRepository interface {
	CreateTemplate(ctx context.Context, template *types.Policy) error
	GetTemplate(ctx context.Context, id string) (*types.Policy, error)
	ListTemplates(ctx context.Context) ([]*types.Policy, error)
}

//Personal.AI order the ending
