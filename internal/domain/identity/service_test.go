package identity

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"testing"
)

// --- Mock Repositories ---

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *types.User) error { return m.Called(ctx, user).Error(0) }
func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*types.User, error) { args := m.Called(ctx, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*types.User), args.Error(1) }
func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*types.User, error) { args := m.Called(ctx, username); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*types.User), args.Error(1) }
func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) { args := m.Called(ctx, email); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*types.User), args.Error(1) }
func (m *MockUserRepository) UpdateUser(ctx context.Context, user *types.User) error { return m.Called(ctx, user).Error(0) }
func (m *MockUserRepository) DeleteUser(ctx context.Context, id string) error { return m.Called(ctx, id).Error(0) }
func (m *MockUserRepository) ListUsers(ctx context.Context, pq PaginationQuery) ([]*types.User, error) { return nil, nil }
func (m *MockUserRepository) FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error) { return nil, nil }

type MockGroupRepository struct {
	mock.Mock
}
func (m *MockGroupRepository) CreateGroup(ctx context.Context, group *types.UserGroup) error { return nil }
func (m *MockGroupRepository) GetGroupByID(ctx context.Context, id string) (*types.UserGroup, error) { args := m.Called(ctx, id); if args.Get(0) == nil { return nil, args.Error(1) }; return args.Get(0).(*types.UserGroup), args.Error(1) }
func (m *MockGroupRepository) GetGroupByName(ctx context.Context, name string) (*types.UserGroup, error) { return nil, nil }
func (m *MockGroupRepository) UpdateGroup(ctx context.Context, group *types.UserGroup) error { return nil }
func (m *MockGroupRepository) DeleteGroup(ctx context.Context, id string) error { return nil }
func (m *MockGroupRepository) ListGroups(ctx context.Context, pq PaginationQuery) ([]*types.UserGroup, error) { return nil, nil }
func (m *MockGroupRepository) AddUserToGroup(ctx context.Context, userID, groupID string) error { return m.Called(ctx, userID, groupID).Error(0) }
func (m *MockGroupRepository) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error { return nil }
func (m *MockGroupRepository) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) { return nil, nil }


// --- Tests ---

func TestIdentityService_CreateUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockGroupRepo := new(MockGroupRepository)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	crypto := utils.NewCryptoManager("secret")

	s := NewService(mockUserRepo, mockGroupRepo, crypto, logger)

	ctx := context.Background()

	t.Run("Successful Creation", func(t *testing.T) {
		mockUserRepo.On("GetUserByUsername", ctx, "newuser").Return(nil, types.ErrNotFound).Once()
		mockUserRepo.On("GetUserByEmail", ctx, "new@example.com").Return(nil, types.ErrNotFound).Once()
		mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*types.User")).Return(nil).Once()

		user, err := s.CreateUser(ctx, "newuser", "new@example.com", "password123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "newuser", user.Username)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Username Conflict", func(t *testing.T) {
		existingUser := &types.User{ID: "1", Username: "existinguser"}
		mockUserRepo.On("GetUserByUsername", ctx, "existinguser").Return(existingUser, nil).Once()

		_, err := s.CreateUser(ctx, "existinguser", "new@example.com", "password123")

		assert.Error(t, err)
		assert.Equal(t, types.ErrConflict.Code, err.(*types.Error).Code)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestIdentityService_AddUserToGroup(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockGroupRepo := new(MockGroupRepository)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	crypto := utils.NewCryptoManager("secret")

	s := NewService(mockUserRepo, mockGroupRepo, crypto, logger)
	ctx := context.Background()

	testUser := &types.User{ID: "user1"}
	testGroup := &types.UserGroup{ID: "group1"}

	mockUserRepo.On("GetUserByID", ctx, "user1").Return(testUser, nil)
	mockGroupRepo.On("GetGroupByID", ctx, "group1").Return(testGroup, nil)
	mockGroupRepo.On("AddUserToGroup", ctx, "user1", "group1").Return(nil)

	err := s.AddUserToGroup(ctx, "user1", "group1")

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
}

//Personal.AI order the ending
