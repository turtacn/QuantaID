package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/pkg/types"
)

// PolicyMemoryRepository provides an in-memory implementation of the policy repository.
type PolicyMemoryRepository struct {
	mu       sync.RWMutex
	policies map[string]*types.Policy
}

// NewPolicyMemoryRepository creates a new in-memory policy repository.
func NewPolicyMemoryRepository() *PolicyMemoryRepository {
	return &PolicyMemoryRepository{
		policies: make(map[string]*types.Policy),
	}
}

func (r *PolicyMemoryRepository) CreatePolicy(ctx context.Context, policy *types.Policy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	policy.ID = uuid.New().String()
	r.policies[policy.ID] = policy
	return nil
}

func (r *PolicyMemoryRepository) GetPolicyByID(ctx context.Context, id string) (*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	policy, ok := r.policies[id]
	if !ok {
		return nil, fmt.Errorf("policy with ID '%s' not found", id)
	}
	return policy, nil
}

func (r *PolicyMemoryRepository) UpdatePolicy(ctx context.Context, policy *types.Policy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.policies[policy.ID]; !ok {
		return fmt.Errorf("policy with ID '%s' not found for update", policy.ID)
	}
	r.policies[policy.ID] = policy
	return nil
}

func (r *PolicyMemoryRepository) DeletePolicy(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.policies, id)
	return nil
}

func (r *PolicyMemoryRepository) ListPolicies(ctx context.Context, pq types.PaginationQuery) ([]*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	policies := make([]*types.Policy, 0, len(r.policies))
	for _, policy := range r.policies {
		policies = append(policies, policy)
	}
    // Note: Simple implementation without sorting.
    start := pq.Offset
    end := start + pq.PageSize

    if start > len(policies) {
        return []*types.Policy{}, nil
    }
    if end > len(policies) {
        end = len(policies)
    }

	return policies[start:end], nil
}

func (r *PolicyMemoryRepository) FindPoliciesForSubject(ctx context.Context, subject string) ([]*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var subjectPolicies []*types.Policy
	for _, policy := range r.policies {
		for _, s := range policy.Subjects {
			if s == subject {
				subjectPolicies = append(subjectPolicies, policy)
				break
			}
		}
	}
	return subjectPolicies, nil
}

func (r *PolicyMemoryRepository) FindPoliciesForResource(ctx context.Context, resource string) ([]*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var resourcePolicies []*types.Policy
	for _, policy := range r.policies {
		for _, r := range policy.Resources {
			if r == resource {
				resourcePolicies = append(resourcePolicies, policy)
				break
			}
		}
	}
	return resourcePolicies, nil
}

func (r *PolicyMemoryRepository) FindPoliciesForAction(ctx context.Context, action string) ([]*types.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var actionPolicies []*types.Policy
	for _, policy := range r.policies {
		for _, a := range policy.Actions {
			if a == action {
				actionPolicies = append(actionPolicies, policy)
				break
			}
		}
	}
	return actionPolicies, nil
}
