package testutils

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
)

type MockRiskEngine struct {
	mock.Mock
}

func (m *MockRiskEngine) Evaluate(ctx context.Context, ac auth.AuthContext) (auth.RiskScore, auth.RiskLevel, error) {
	args := m.Called(ctx, ac)
	return args.Get(0).(auth.RiskScore), args.Get(1).(auth.RiskLevel), args.Error(2)
}

// MockSink is a simple in-memory sink for testing the pipeline.
type MockSink struct {
	mu     sync.Mutex
	Events []*audit.AuditEvent
	Err    error // Optional error to simulate sink failure
}

func (s *MockSink) Write(event *audit.AuditEvent) error {
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

type MockTokenFamilyRepository struct {
	mock.Mock
}

func (m *MockTokenFamilyRepository) CreateFamily(ctx context.Context, family *auth.TokenFamily) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

func (m *MockTokenFamilyRepository) GetFamilyByToken(ctx context.Context, token string) (*auth.TokenFamily, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenFamily), args.Error(1)
}

func (m *MockTokenFamilyRepository) GetFamilyByID(ctx context.Context, familyID string) (*auth.TokenFamily, error) {
	args := m.Called(ctx, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenFamily), args.Error(1)
}

func (m *MockTokenFamilyRepository) UpdateFamily(ctx context.Context, family *auth.TokenFamily) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

func (m *MockTokenFamilyRepository) RevokeFamily(ctx context.Context, familyID string) error {
	args := m.Called(ctx, familyID)
	return args.Error(0)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockRedisClient) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	args := m.Called(ctx, key, member)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockRedisClient) SCard(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockRedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockRedisClient) ZCard(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRedisClient) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) Client() *redis.Client {
	args := m.Called()
	return args.Get(0).(*redis.Client)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
