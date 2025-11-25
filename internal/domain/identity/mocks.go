package identity

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
)

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

func (m *MockIService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIService) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	args := m.Called(ctx, id)
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

func (m *MockIService) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*types.User), args.Int(1), args.Error(2)
}

func (m *MockIService) UpdateUser(ctx context.Context, user *types.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockIService) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockIService) GetUserByExternalID(ctx context.Context, externalID, sourceID string) (*types.User, error) {
	args := m.Called(ctx, externalID, sourceID)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIService) CreateGroup(ctx context.Context, group *types.UserGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockIService) GetGroup(ctx context.Context, groupID string) (*types.UserGroup, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).(*types.UserGroup), args.Error(1)
}

func (m *MockIService) UpdateGroup(ctx context.Context, group *types.UserGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockIService) DeleteGroup(ctx context.Context, groupID string) error {
	args := m.Called(ctx, groupID)
	return args.Error(0)
}

func (m *MockIService) ListGroups(ctx context.Context, offset, limit int) ([]*types.UserGroup, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*types.UserGroup), args.Error(1)
}

func (m *MockIService) GetUserRepo() UserRepository {
	return nil
}
