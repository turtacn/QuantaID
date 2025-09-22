package authorization

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"testing"
)

// --- Mock Domain Services ---

type MockPolicyService struct{ mock.Mock }
func (m *MockPolicyService) Evaluate(ctx context.Context, evalCtx *types.PolicyEvaluationContext) (*types.PolicyDecision, error) {
	args := m.Called(ctx, evalCtx)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*types.PolicyDecision), args.Error(1)
}
func (m *MockPolicyService) CreatePolicy(ctx context.Context, p *types.Policy) error { return nil }


type MockIdentityService struct{ mock.Mock }
func (m *MockIdentityService) GetUser(ctx context.Context, userID string) (*types.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*types.User), args.Error(1)
}
func (m *MockIdentityService) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*types.UserGroup), args.Error(1)
}
func (m *MockIdentityService) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) { return nil, nil }
func (m *MockIdentityService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) { return nil, nil }
func (m *MockIdentityService) AddUserToGroup(ctx context.Context, userID, groupID string) error { return nil }
func (m *MockIdentityService) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error { return nil }

// --- Tests ---

func TestAuthorizationApplicationService_CheckPermission(t *testing.T) {
	mockPolicySvc := new(MockPolicyService)
	mockIdentitySvc := new(MockIdentityService)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	appSvc := NewApplicationService(mockPolicySvc, mockIdentitySvc, logger)
	ctx := context.Background()

	testUser := &types.User{ID: "user1"}
	testGroups := []*types.UserGroup{{ID: "group1"}, {ID: "group2"}}
	req := CheckPermissionRequest{UserID: "user1", Action: "doc:read", ResourceID: "res1"}

	t.Run("Permission Allowed", func(t *testing.T) {
		mockIdentitySvc.On("GetUser", ctx, "user1").Return(testUser, nil).Once()
		mockIdentitySvc.On("GetUserGroups", ctx, "user1").Return(testGroups, nil).Once()

		mockPolicySvc.On("Evaluate", ctx, mock.MatchedBy(func(evalCtx *types.PolicyEvaluationContext) bool {
			return evalCtx.Subject["id"] == "user1" && evalCtx.Action == "doc:read"
		})).Return(&types.PolicyDecision{Allowed: true}, nil).Once()

		allowed, err := appSvc.CheckPermission(ctx, req)
		assert.True(t, allowed)
		assert.Nil(t, err)
		mockIdentitySvc.AssertExpectations(t)
		mockPolicySvc.AssertExpectations(t)
	})

	t.Run("Permission Denied", func(t *testing.T) {
		mockIdentitySvc.On("GetUser", ctx, "user1").Return(testUser, nil).Once()
		mockIdentitySvc.On("GetUserGroups", ctx, "user1").Return(testGroups, nil).Once()
		mockPolicySvc.On("Evaluate", ctx, mock.Anything).Return(&types.PolicyDecision{Allowed: false, Reason: "denied"}, nil).Once()

		allowed, err := appSvc.CheckPermission(ctx, req)
		assert.False(t, allowed)
		assert.NotNil(t, err)
		assert.Equal(t, types.ErrForbidden.Code, err.Code)
		mockIdentitySvc.AssertExpectations(t)
		mockPolicySvc.AssertExpectations(t)
	})
}

//Personal.AI order the ending
