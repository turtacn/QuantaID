package policy

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"testing"
)

// --- Mock Repository ---

type MockPolicyRepository struct {
	mock.Mock
}
func (m *MockPolicyRepository) CreatePolicy(ctx context.Context, policy *types.Policy) error { return nil }
func (m *MockPolicyRepository) GetPolicyByID(ctx context.Context, id string) (*types.Policy, error) { return nil, nil }
func (m *MockPolicyRepository) UpdatePolicy(ctx context.Context, policy *types.Policy) error { return nil }
func (m *MockPolicyRepository) DeletePolicy(ctx context.Context, id string) error { return nil }
func (m *MockPolicyRepository) ListPolicies(ctx context.Context, pq types.PaginationQuery) ([]*types.Policy, error) { return nil, nil }
func (m *MockPolicyRepository) FindPoliciesForSubject(ctx context.Context, subject string) ([]*types.Policy, error) {
	args := m.Called(ctx, subject)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Policy), args.Error(1)
}
func (m *MockPolicyRepository) FindPoliciesForResource(ctx context.Context, resource string) ([]*types.Policy, error) { return nil, nil }
func (m *MockPolicyRepository) FindPoliciesForAction(ctx context.Context, action string) ([]*types.Policy, error) { return nil, nil }

// --- Tests ---

func TestPolicyService_Evaluate(t *testing.T) {
	mockRepo := new(MockPolicyRepository)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	s := NewService(mockRepo, logger)
	ctx := context.Background()

	allowPolicy := &types.Policy{
		ID: "allow-read",
		Effect: types.EffectAllow,
		Actions: []string{"document:read"},
		Resources: []string{"doc:123"},
		Subjects: []string{"user:u1"},
	}
	denyPolicy := &types.Policy{
		ID: "deny-write",
		Effect: types.EffectDeny,
		Actions: []string{"document:write"},
		Resources: []string{"doc:123"},
		Subjects: []string{"user:u1"},
	}

	t.Run("Allowed by policy", func(t *testing.T) {
		mockRepo.On("FindPoliciesForSubject", ctx, "user:u1").Return([]*types.Policy{allowPolicy}, nil).Once()

		evalCtx := &types.PolicyEvaluationContext{
			Subject: map[string]interface{}{"id": "u1"},
			Action: "document:read",
			Resource: map[string]interface{}{"id": "doc:123"},
		}

		decision, err := s.Evaluate(ctx, evalCtx)
		assert.NoError(t, err)
		assert.True(t, decision.Allowed)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Denied by policy", func(t *testing.T) {
		mockRepo.On("FindPoliciesForSubject", ctx, "user:u1").Return([]*types.Policy{allowPolicy, denyPolicy}, nil).Once()

		evalCtx := &types.PolicyEvaluationContext{
			Subject: map[string]interface{}{"id": "u1"},
			Action: "document:write",
			Resource: map[string]interface{}{"id": "doc:123"},
		}

		decision, err := s.Evaluate(ctx, evalCtx)
		assert.NoError(t, err)
		assert.False(t, decision.Allowed)
		assert.Contains(t, decision.Reason, "denied by policy")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Default deny when no policy matches", func(t *testing.T) {
		mockRepo.On("FindPoliciesForSubject", ctx, "user:u1").Return([]*types.Policy{allowPolicy}, nil).Once()

		evalCtx := &types.PolicyEvaluationContext{
			Subject: map[string]interface{}{"id": "u1"},
			Action: "document:delete",
			Resource: map[string]interface{}{"id": "doc:123"},
		}

		decision, err := s.Evaluate(ctx, evalCtx)
		assert.NoError(t, err)
		assert.False(t, decision.Allowed)
		assert.Equal(t, "No allowing policy found", decision.Reason)
		mockRepo.AssertExpectations(t)
	})
}

//Personal.AI order the ending
