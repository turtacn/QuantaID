package postgresql

import (
	"context"

	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// PostgresPolicyRepository provides a GORM-based implementation of the policy-related repositories.
type PostgresPolicyRepository struct {
	db *gorm.DB
}

// NewPostgresPolicyRepository creates a new PostgreSQL policy repository.
func NewPostgresPolicyRepository(db *gorm.DB) *PostgresPolicyRepository {
	return &PostgresPolicyRepository{db: db}
}

// --- PolicyRepository Implementation ---

// CreatePolicy adds a new policy to the database.
func (r *PostgresPolicyRepository) CreatePolicy(ctx context.Context, policy *types.Policy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

// GetPolicyByID retrieves a policy by its ID from the database.
func (r *PostgresPolicyRepository) GetPolicyByID(ctx context.Context, id string) (*types.Policy, error) {
	var policy types.Policy
	err := r.db.WithContext(ctx).First(&policy, "id = ?", id).Error
	return &policy, err
}

// UpdatePolicy updates an existing policy in the database.
func (r *PostgresPolicyRepository) UpdatePolicy(ctx context.Context, policy *types.Policy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

// DeletePolicy removes a policy from the database.
func (r *PostgresPolicyRepository) DeletePolicy(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.Policy{}, "id = ?", id).Error
}

// ListPolicies returns a paginated list of all policies from the database.
func (r *PostgresPolicyRepository) ListPolicies(ctx context.Context, pq types.PaginationQuery) ([]*types.Policy, error) {
	var policies []*types.Policy
	err := r.db.WithContext(ctx).Offset(pq.Offset).Limit(pq.PageSize).Find(&policies).Error
	return policies, err
}

// FindPoliciesForSubject finds all policies that apply to a given subject.
func (r *PostgresPolicyRepository) FindPoliciesForSubject(ctx context.Context, subject string) ([]*types.Policy, error) {
	var policies []*types.Policy
	// This query uses the `?` operator, which checks if a string exists in a JSON array.
	// This is a PostgreSQL-specific feature.
	err := r.db.WithContext(ctx).Where("subjects @> ?", `{"`+subject+`"}`).Find(&policies).Error
	return policies, err
}

// FindPoliciesForResource finds all policies that apply to a given resource.
func (r *PostgresPolicyRepository) FindPoliciesForResource(ctx context.Context, resource string) ([]*types.Policy, error) {
	var policies []*types.Policy
	err := r.db.WithContext(ctx).Where("resources @> ?", `{"`+resource+`"}`).Find(&policies).Error
	return policies, err
}

// FindPoliciesForAction finds all policies that apply to a given action.
func (r *PostgresPolicyRepository) FindPoliciesForAction(ctx context.Context, action string) ([]*types.Policy, error) {
	var policies []*types.Policy
	err := r.db.WithContext(ctx).Where("actions @> ?", `{"`+action+`"}`).Find(&policies).Error
	return policies, err
}