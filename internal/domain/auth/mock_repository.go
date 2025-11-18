package auth

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
	"time"
)

// MockSessionRepository is a mock implementation of the SessionRepository interface.
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, session *types.UserSession, duration time.Duration) error {
	args := m.Called(ctx, session, duration)
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

// MockTokenRepository is a mock implementation of the TokenRepository interface.
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

// MockAuditLogRepository is a mock implementation of the AuditLogRepository interface.
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) CreateLogEntry(ctx context.Context, entry *types.AuditLog) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockAuditLogRepository) GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	args := m.Called(ctx, userID, pq)
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	args := m.Called(ctx, action, pq)
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}

// MockTokenFamilyRepository is a mock implementation of the TokenFamilyRepository interface.
type MockTokenFamilyRepository struct {
	mock.Mock
}

func (m *MockTokenFamilyRepository) CreateFamily(ctx context.Context, family *TokenFamily) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

func (m *MockTokenFamilyRepository) GetFamilyByToken(ctx context.Context, token string) (*TokenFamily, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*TokenFamily), args.Error(1)
}

func (m *MockTokenFamilyRepository) GetFamilyByID(ctx context.Context, familyID string) (*TokenFamily, error) {
	args := m.Called(ctx, familyID)
	return args.Get(0).(*TokenFamily), args.Error(1)
}

func (m *MockTokenFamilyRepository) UpdateFamily(ctx context.Context, family *TokenFamily) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

func (m *MockTokenFamilyRepository) RevokeFamily(ctx context.Context, familyID string) error {
	args := m.Called(ctx, familyID)
	return args.Error(0)
}
