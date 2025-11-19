package policy

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// PolicyRepository defines the interface for storing and retrieving authorization policies.
type PolicyRepository interface {
	CreatePolicy(ctx context.Context, policy *types.Policy) error
	GetPolicyByID(ctx context.Context, id string) (*types.Policy, error)
	UpdatePolicy(ctx context.Context, policy *types.Policy) error
	DeletePolicy(ctx context.Context, id string) error
	ListPolicies(ctx context.Context, pq types.PaginationQuery) ([]*types.Policy, error)
	FindPoliciesForSubject(ctx context.Context, subject string) ([]*types.Policy, error)
}
