package identity

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
)

// MockIService is a mock implementation of the IService interface for testing.
type MockIService struct {
	mock.Mock
}

func (m *MockIService) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) {
	args := m.Called(ctx, username, email, password)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIService) GetUser(ctx context.Context, userID string) (*types.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIService) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIService) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*types.UserGroup), args.Error(1)
}

func (m *MockIService) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockIService) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error {
	args := m.Called(ctx, userID, newStatus)
	return args.Error(0)
}
