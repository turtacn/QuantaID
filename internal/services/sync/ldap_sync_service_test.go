package sync_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/services/sync"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// Mocks
type MockLDAPConnector struct {
	mock.Mock
}

func (m *MockLDAPConnector) SyncUsers(ctx context.Context) ([]*types.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*types.User), args.Error(1)
}

func (m *MockLDAPConnector) SearchUsers(ctx context.Context, filter string) ([]*types.User, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*types.User), args.Error(1)
}

type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) RecordAdminAction(ctx context.Context, userID, ip, resource, action, traceID string, details map[string]any) {
	m.Called(ctx, userID, ip, resource, action, traceID, details)
}

func setupTest(t *testing.T) (*sync.LDAPSyncService, *memory.IdentityMemoryRepository, *MockLDAPConnector, *MockAuditService) {
	userRepo := memory.NewIdentityMemoryRepository()
	ldapConnector := new(MockLDAPConnector)
	auditService := new(MockAuditService)
	logger := zap.NewNop()

	config := &sync.LDAPSyncConfig{
		ConflictStrategy: sync.ConflictPreferRemote,
		LifecycleRules: []sync.LifecycleRule{
			{
				SourceAttr:   "hr_status",
				MatchValue:   "terminated",
				TargetStatus: "inactive",
			},
		},
	}

	syncService := sync.NewLDAPSyncService(ldapConnector, userRepo, config, auditService, logger)
	return syncService, userRepo, ldapConnector, auditService
}

func Test_FullSync_CreatesNewUsers(t *testing.T) {
	syncService, userRepo, ldapConnector, auditService := setupTest(t)
	ctx := context.Background()

	ldapUsers := []*types.User{
		{Username: "newuser1", Email: "new1@test.com", Status: types.UserStatusActive},
		{Username: "newuser2", Email: "new2@test.com", Status: types.UserStatusActive},
	}

	ldapConnector.On("SyncUsers", ctx).Return(ldapUsers, nil)
	auditService.On("RecordAdminAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	stats, err := syncService.FullSync(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 2, stats.Created)
	assert.Equal(t, 0, stats.Updated)
	assert.Equal(t, 0, stats.Disabled)
	assert.Equal(t, 0, stats.Errors)

	users, err := userRepo.ListUsers(ctx, identity.PaginationQuery{PageSize: 10, Offset: 0})
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	ldapConnector.AssertExpectations(t)
	auditService.AssertExpectations(t)
}

func Test_FullSync_UpdatesExistingUsers_WithPreferRemote(t *testing.T) {
	syncService, userRepo, ldapConnector, auditService := setupTest(t)
	ctx := context.Background()

	// Pre-populate with an existing user
	existingUser := &types.User{Username: "existinguser", Email: "original@test.com", Status: types.UserStatusActive}
	err := userRepo.CreateUser(ctx, existingUser)
	assert.NoError(t, err)

	ldapUsers := []*types.User{
		{Username: "existinguser", Email: "updated@test.com", Status: types.UserStatusActive},
	}

	ldapConnector.On("SyncUsers", ctx).Return(ldapUsers, nil)
	auditService.On("RecordAdminAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	stats, err := syncService.FullSync(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, stats.Created)
	assert.Equal(t, 1, stats.Updated)
	assert.Equal(t, 0, stats.Disabled)

	updatedUser, err := userRepo.GetUserByUsername(ctx, "existinguser")
	assert.NoError(t, err)
	assert.Equal(t, "updated@test.com", updatedUser.Email)
}

func Test_FullSync_RespectsLifecycleRules(t *testing.T) {
	syncService, userRepo, ldapConnector, auditService := setupTest(t)
	ctx := context.Background()

	ldapUsers := []*types.User{
		{
			Username:   "terminateduser",
			Email:      "terminated@test.com",
			Status:     types.UserStatusActive,
			Attributes: map[string]interface{}{"hr_status": "terminated"},
		},
	}

	ldapConnector.On("SyncUsers", ctx).Return(ldapUsers, nil)
	auditService.On("RecordAdminAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	stats, err := syncService.FullSync(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.Created)

	user, err := userRepo.GetUserByUsername(ctx, "terminateduser")
	assert.NoError(t, err)
	assert.Equal(t, types.UserStatusInactive, user.Status)
}

func Test_FullSync_DisablesUsersNotInLDAP(t *testing.T) {
	syncService, userRepo, ldapConnector, auditService := setupTest(t)
	ctx := context.Background()

	// Pre-populate with a user that will be "deleted" from LDAP
	deletedUser := &types.User{Username: "deleteduser", Email: "deleted@test.com", Status: types.UserStatusActive}
	err := userRepo.CreateUser(ctx, deletedUser)
	assert.NoError(t, err)

	ldapConnector.On("SyncUsers", ctx).Return([]*types.User{}, nil)
	auditService.On("RecordAdminAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	stats, err := syncService.FullSync(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.Disabled)

	user, err := userRepo.GetUserByUsername(ctx, "deleteduser")
	assert.NoError(t, err)
	assert.Equal(t, types.UserStatusInactive, user.Status)
}

func Test_IncrementalSync_OnlySyncsRecentChanges(t *testing.T) {
	syncService, userRepo, ldapConnector, auditService := setupTest(t)
	ctx := context.Background()
	since := time.Now().UTC().Add(-1 * time.Hour)

	// Pre-populate with an existing user that won't be in the search results
	existingUser := &types.User{Username: "existinguser", Email: "original@test.com", Status: types.UserStatusActive}
	err := userRepo.CreateUser(ctx, existingUser)
	assert.NoError(t, err)

	changedUsers := []*types.User{
		{Username: "changeduser", Email: "changed@test.com"},
	}

	filter := fmt.Sprintf("(&(objectClass=inetOrgPerson)(modifyTimestamp>=%s))", since.Format("20060102150405Z"))
	ldapConnector.On("SearchUsers", ctx, filter).Return(changedUsers, nil)
	auditService.On("RecordAdminAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	stats, err := syncService.IncrementalSync(ctx, since)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.TotalRemote)

	ldapConnector.AssertCalled(t, "SearchUsers", ctx, filter)
}
