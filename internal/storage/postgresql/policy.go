package postgresql

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"golang.org/x/exp/slices"
	"sync"
)

// InMemoryPolicyRepository provides an in-memory implementation of the policy repository.
// NOTE: Despite the package name 'postgresql', this is an IN-MEMORY implementation,
// likely used for testing or simple, non-persistent deployments. It uses a map
// with a mutex for thread-safe operations.
type InMemoryPolicyRepository struct {
	mu       sync.RWMutex
	policies map[string]*types.Policy
}

// NewInMemoryPolicyRepository creates a new, empty in-memory policy repository.
func NewInMemoryPolicyRepository() *InMemoryPolicyRepository {
	return &InMemoryPolicyRepository{
		policies: make(map[string]*types.Policy),
	}
}

// CreatePolicy adds a new policy to the in-memory store.
func (r *InMemoryPolicyRepository) CreatePolicy(ctx context.Context, policy *types.Policy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.policies[policy.ID]; exists {
		return types.ErrConflict.WithDetails(map[string]string{"id": policy.ID})
	}
	r.policies[policy.ID] = policy
	return nil
}

// GetPolicyByID retrieves a policy by its ID from the in-memory store.
func (r *InMemoryPolicyRepository) GetPolicyByID(ctx context.Context, id string) (*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	policy, exists := r.policies[id]
	if !exists {
		return nil, types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	return policy, nil
}

// UpdatePolicy updates an existing policy in the in-memory store.
func (r *InMemoryPolicyRepository) UpdatePolicy(ctx context.Context, policy *types.Policy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.policies[policy.ID]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": policy.ID})
	}
	r.policies[policy.ID] = policy
	return nil
}

// DeletePolicy removes a policy from the in-memory store by its ID.
func (r *InMemoryPolicyRepository) DeletePolicy(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.policies[id]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	delete(r.policies, id)
	return nil
}

// ListPolicies returns a paginated list of all policies from the in-memory store.
func (r *InMemoryPolicyRepository) ListPolicies(ctx context.Context, pq types.PaginationQuery) ([]*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	policies := make([]*types.Policy, 0, len(r.policies))
	for _, p := range r.policies {
		policies = append(policies, p)
	}
	start, end := pq.Offset, pq.Offset+pq.PageSize
	if start > len(policies) { return []*types.Policy{}, nil }
	if end > len(policies) { end = len(policies) }
	return policies[start:end], nil
}

// FindPoliciesForSubject searches the in-memory store for all policies that apply to a given subject.
// It supports wildcards for the subject field.
func (r *InMemoryPolicyRepository) FindPoliciesForSubject(ctx context.Context, subject string) ([]*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var foundPolicies []*types.Policy
	for _, p := range r.policies {
		if slices.Contains(p.Subjects, subject) || slices.Contains(p.Subjects, "*") {
			foundPolicies = append(foundPolicies, p)
		}
	}
	return foundPolicies, nil
}

// FindPoliciesForResource searches the in-memory store for all policies that apply to a given resource.
// It supports wildcards for the resource field.
func (r *InMemoryPolicyRepository) FindPoliciesForResource(ctx context.Context, resource string) ([]*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var foundPolicies []*types.Policy
	for _, p := range r.policies {
		if slices.Contains(p.Resources, resource) || slices.Contains(p.Resources, "*") {
			foundPolicies = append(foundPolicies, p)
		}
	}
	return foundPolicies, nil
}

// FindPoliciesForAction searches the in-memory store for all policies that apply to a given action.
// It supports wildcards for the action field.
func (r *InMemoryPolicyRepository) FindPoliciesForAction(ctx context.Context, action string) ([]*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var foundPolicies []*types.Policy
	for _, p := range r.policies {
		if slices.Contains(p.Actions, action) || slices.Contains(p.Actions, "*") {
			foundPolicies = append(foundPolicies, p)
		}
	}
	return foundPolicies, nil
}
