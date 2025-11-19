package postgresql

import (
	"context"
	"errors"

	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// PostgresPolicyRepository provides a GORM-backed implementation of the policy.PolicyRepository interface.
type PostgresPolicyRepository struct {
	db *gorm.DB
}

// NewPostgresPolicyRepository creates a new, GORM-backed policy repository.
func NewPostgresPolicyRepository(db *gorm.DB) policy.PolicyRepository {
	return &PostgresPolicyRepository{db: db}
}

// CreatePolicy adds a new policy to the database.
func (r *PostgresPolicyRepository) CreatePolicy(ctx context.Context, policy *types.Policy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

// GetPolicyByID retrieves a policy by its ID from the database.
func (r *PostgresPolicyRepository) GetPolicyByID(ctx context.Context, id string) (*types.Policy, error) {
	var policy types.Policy
	if err := r.db.WithContext(ctx).First(&policy, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrNotFound.WithDetails(map[string]string{"id": id})
		}
		return nil, err
	}
	return &policy, nil
}

// UpdatePolicy updates an existing policy in the database.
func (r *PostgresPolicyRepository) UpdatePolicy(ctx context.Context, policy *types.Policy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

// DeletePolicy removes a policy from the database by its ID.
func (r *PostgresPolicyRepository) DeletePolicy(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.Policy{}, "id = ?", id).Error
}

// ListPolicies returns a paginated list of all policies from the database.
func (r *PostgresPolicyRepository) ListPolicies(ctx context.Context, pq types.PaginationQuery) ([]*types.Policy, error) {
	var policies []*types.Policy
	err := r.db.WithContext(ctx).
		Order("created_at desc").
		Offset(pq.Offset).
		Limit(pq.PageSize).
		Find(&policies).Error
	return policies, err
}

// FindPoliciesForSubject searches the database for all policies that apply to a given subject.
// This implementation uses the ANY operator for PostgreSQL arrays.
func (r *PostgresPolicyRepository) FindPoliciesForSubject(ctx context.Context, subject string) ([]*types.Policy, error) {
	var policies []*types.Policy
	err := r.db.WithContext(ctx).
		Where("? = ANY(subjects) OR '*' = ANY(subjects)", subject).
		Find(&policies).Error
	return policies, err
}
