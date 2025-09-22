package identity

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"testing"
)

// --- Mock Domain Service ---

type MockIdentityDomainService struct {
	mock.Mock
}

func (m *MockIdentityDomainService) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) {
	args := m.Called(ctx, username, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}
func (m *MockIdentityDomainService) GetUser(ctx context.Context, userID string) (*types.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}
func (m *MockIdentityDomainService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) { return nil, nil }
func (m *MockIdentityDomainService) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) { return nil, nil }
func (m *MockIdentityDomainService) AddUserToGroup(ctx context.Context, userID, groupID string) error { return nil }
func (m *MockIdentityDomainService) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error { return nil }


// --- Tests ---

func TestIdentityApplicationService_CreateUser(t *testing.T) {
	mockDomainSvc := new(MockIdentityDomainService)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	appSvc := NewApplicationService(mockDomainSvc, logger)
	ctx := context.Background()

	t.Run("Successful Creation", func(t *testing.T) {
		req := CreateUserRequest{Username: "test", Email: "test@test.com", Password: "password"}

		// Note: we return a user with a password hash to test that it gets stripped
		returnedUser := &types.User{ID: "123", Username: "test", Password: "hashed_password"}

		mockDomainSvc.On("CreateUser", ctx, req.Username, req.Email, req.Password).Return(returnedUser, nil).Once()

		user, err := appSvc.CreateUser(ctx, req)

		assert.Nil(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "123", user.ID)
		assert.Empty(t, user.Password, "Password should be stripped from the response")
		mockDomainSvc.AssertExpectations(t)
	})

	t.Run("Domain Service Error", func(t *testing.T) {
		req := CreateUserRequest{Username: "test", Email: "test@test.com", Password: "password"}

		mockDomainSvc.On("CreateUser", ctx, req.Username, req.Email, req.Password).Return(nil, types.ErrConflict).Once()

		_, err := appSvc.CreateUser(ctx, req)

		assert.NotNil(t, err)
		assert.Equal(t, types.ErrConflict.Code, err.Code)
		mockDomainSvc.AssertExpectations(t)
	})
}

//Personal.AI order the ending
