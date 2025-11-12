package testutils

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *types.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockIdentityService is a mock implementation of the IdentityService interface
type MockIdentityService struct {
	mock.Mock
}

func (m *MockIdentityService) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) {
	args := m.Called(ctx, username, email, password)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIdentityService) GetUser(ctx context.Context, userID string) (*types.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIdentityService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIdentityService) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*types.UserGroup), args.Error(1)
}

func (m *MockIdentityService) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockIdentityService) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error {
	args := m.Called(ctx, userID, newStatus)
	return args.Error(0)
}

// MockSessionRepository is a mock implementation of the SessionRepository interface
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, session *types.UserSession, ttl time.Duration) error {
	args := m.Called(ctx, session, ttl)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSession(ctx context.Context, sessionID string) (*types.UserSession, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*types.UserSession), args.Error(1)
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*types.UserSession, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*types.UserSession), args.Error(1)
}

// MockTokenRepository is a mock implementation of the TokenRepository interface
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) StoreRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error {
	args := m.Called(ctx, token, userID, duration)
	return args.Error(0)
}

func (m *MockTokenRepository) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) AddToDenyList(ctx context.Context, jti string, duration time.Duration) error {
	args := m.Called(ctx, jti, duration)
	return args.Error(0)
}

func (m *MockTokenRepository) IsInDenyList(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

// MockAuditLogRepository is a mock implementation of the AuditLogRepository interface
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) CreateLogEntry(ctx context.Context, logEntry *types.AuditLog) error {
	args := m.Called(ctx, logEntry)
	return args.Error(0)
}

func (m *MockAuditLogRepository) GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	args := m.Called(ctx, action, pq)
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	args := m.Called(ctx, userID, pq)
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}
