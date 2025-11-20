package identity

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
)

type MockIdentityRepository struct {
	mock.Mock
}

func (m *MockIdentityRepository) CreateUser(ctx context.Context, user *types.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockIdentityRepository) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIdentityRepository) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIdentityRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIdentityRepository) UpdateUser(ctx context.Context, user *types.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockIdentityRepository) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIdentityRepository) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*types.User), args.Int(1), args.Error(2)
}

func (m *MockIdentityRepository) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error {
	args := m.Called(ctx, userID, newStatus)
	return args.Error(0)
}

func (m *MockIdentityRepository) FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error) {
	args := m.Called(ctx, attribute, value)
	return args.Get(0).([]*types.User), args.Error(1)
}

func (m *MockIdentityRepository) UpsertBatch(ctx context.Context, users []*types.User) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

func (m *MockIdentityRepository) CreateBatch(ctx context.Context, users []*types.User) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

func (m *MockIdentityRepository) UpdateBatch(ctx context.Context, users []*types.User) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

func (m *MockIdentityRepository) DeleteBatch(ctx context.Context, userIDs []string) error {
	args := m.Called(ctx, userIDs)
	return args.Error(0)
}

func (m *MockIdentityRepository) FindUsersBySource(ctx context.Context, sourceID string) ([]*types.User, error) {
	args := m.Called(ctx, sourceID)
	return args.Get(0).([]*types.User), args.Error(1)
}

func (m *MockIdentityRepository) CreateGroup(ctx context.Context, group *types.UserGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockIdentityRepository) GetGroupByID(ctx context.Context, id string) (*types.UserGroup, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.UserGroup), args.Error(1)
}

func (m *MockIdentityRepository) GetGroupByName(ctx context.Context, name string) (*types.UserGroup, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*types.UserGroup), args.Error(1)
}

func (m *MockIdentityRepository) UpdateGroup(ctx context.Context, group *types.UserGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockIdentityRepository) DeleteGroup(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIdentityRepository) ListGroups(ctx context.Context, pq PaginationQuery) ([]*types.UserGroup, error) {
	args := m.Called(ctx, pq)
	return args.Get(0).([]*types.UserGroup), args.Error(1)
}

func (m *MockIdentityRepository) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockIdentityRepository) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockIdentityRepository) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*types.UserGroup), args.Error(1)
}
