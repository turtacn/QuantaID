package testutils

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/auth/adaptive"
	"github.com/turtacn/QuantaID/pkg/types"
)

type MockRiskEngine struct {
	mock.Mock
}

func (m *MockRiskEngine) Evaluate(ctx context.Context, event *adaptive.AuthEvent) (*adaptive.RiskScore, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adaptive.RiskScore), args.Error(1)
}

// MockSink is a simple in-memory sink for testing the pipeline.
type MockSink struct {
	mu     sync.Mutex
	Events []*audit.AuditEvent
	Err    error // Optional error to simulate sink failure
}

func (s *MockSink) Write(ctx context.Context, event *audit.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Err != nil {
		return s.Err
	}
	s.Events = append(s.Events, event)
	return nil
}

func (s *MockSink) Close() error { return nil }

type MockMFARepository struct {
	mock.Mock
}

func (m *MockMFARepository) CreateFactor(ctx context.Context, factor *types.MFAFactor) error {
	args := m.Called(ctx, factor)
	return args.Error(0)
}

func (m *MockMFARepository) GetFactor(ctx context.Context, factorID uuid.UUID) (*types.MFAFactor, error) {
	args := m.Called(ctx, factorID)
	return args.Get(0).(*types.MFAFactor), args.Error(1)
}

func (m *MockMFARepository) GetUserFactors(ctx context.Context, userID uuid.UUID) ([]*types.MFAFactor, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*types.MFAFactor), args.Error(1)
}

func (m *MockMFARepository) UpdateFactor(ctx context.Context, factor *types.MFAFactor) error {
	args := m.Called(ctx, factor)
	return args.Error(0)
}

func (m *MockMFARepository) DeleteFactor(ctx context.Context, factorID uuid.UUID) error {
	args := m.Called(ctx, factorID)
	return args.Error(0)
}

func (m *MockMFARepository) CreateVerificationLog(ctx context.Context, log *types.MFAVerificationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockMFARepository) CreateMFAVerificationLog(ctx context.Context, log *types.MFAVerificationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

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

type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) StoreRefreshToken(ctx context.Context, token string, userID string, ttl time.Duration) error {
	args := m.Called(ctx, token, userID, ttl)
	return args.Error(0)
}

func (m *MockTokenRepository) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockTokenRepository) ValidateRefreshToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) AddToDenyList(ctx context.Context, jti string, ttl time.Duration) error {
	args := m.Called(ctx, jti, ttl)
	return args.Error(0)
}

func (m *MockTokenRepository) IsInDenyList(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) CreateLogEntry(ctx context.Context, logEntry *types.AuditLog) error {
	args := m.Called(ctx, logEntry)
	return args.Error(0)
}

func (m *MockAuditLogRepository) GetLogsForUser(ctx context.Context, userID string, pagination types.PaginationQuery) ([]*types.AuditLog, error) {
	args := m.Called(ctx, userID, pagination)
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetLogsByAction(ctx context.Context, action string, pagination types.PaginationQuery) ([]*types.AuditLog, error) {
	args := m.Called(ctx, action, pagination)
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}
