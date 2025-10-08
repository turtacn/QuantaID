package auth

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"testing"
	"time"
)

// --- Mock Definitions ---

type MockIdentityService struct{ mock.Mock }
func (m *MockIdentityService) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) { return nil, nil }
func (m *MockIdentityService) GetUser(ctx context.Context, userID string) (*types.User, error) { return nil, nil }
func (m *MockIdentityService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*types.User), args.Error(1)
}
func (m *MockIdentityService) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) { return nil, nil }
func (m *MockIdentityService) AddUserToGroup(ctx context.Context, userID, groupID string) error { return nil }
func (m *MockIdentityService) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error { return nil }


type MockTokenRepository struct{ mock.Mock }
func (m *MockTokenRepository) StoreRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error { return m.Called(ctx, token, userID, duration).Error(0) }
func (m *MockTokenRepository) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) { return "", nil }
func (m *MockTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error { return nil }
func (m *MockTokenRepository) AddToDenyList(ctx context.Context, jti string, duration time.Duration) error { return nil }
func (m *MockTokenRepository) IsInDenyList(ctx context.Context, jti string) (bool, error) { return false, nil }

type MockSessionRepository struct{ mock.Mock }
func (m *MockSessionRepository) CreateSession(ctx context.Context, session *types.UserSession, duration time.Duration) error { return m.Called(ctx, session, duration).Error(0) }
func (m *MockSessionRepository) GetSession(ctx context.Context, sessionID string) (*types.UserSession, error) { return nil, nil }
func (m *MockSessionRepository) DeleteSession(ctx context.Context, sessionID string) error { return nil }
func (m *MockSessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*types.UserSession, error) { return nil, nil }

type MockAuditRepository struct{ mock.Mock }
func (m *MockAuditRepository) CreateLogEntry(ctx context.Context, entry *types.AuditLog) error { return m.Called(ctx, mock.Anything).Error(0) }
func (m *MockAuditRepository) GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error) { return nil, nil }
func (m *MockAuditRepository) GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error) { return nil, nil }

// --- Test Cases ---

func TestAuthApplicationService_Login(t *testing.T) {
	mockIdentitySvc := new(MockIdentityService)
	mockTokenRepo := new(MockTokenRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockAuditRepo := new(MockAuditRepository)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	cryptoManager := utils.NewCryptoManager("secret")

	authDomainSvc := auth.NewService(mockIdentitySvc, mockSessionRepo, mockTokenRepo, mockAuditRepo, cryptoManager, logger)
	authAppSvc := NewApplicationService(authDomainSvc, logger, Config{AccessTokenDuration: time.Minute})

	hashedPassword, _ := cryptoManager.HashPassword("correct-password")
	testUser := &types.User{ID: "user-123", Username: "testuser", Password: hashedPassword, Status: types.UserStatusActive}

	t.Run("Successful Login", func(t *testing.T) {
		mockIdentitySvc.On("GetUserByUsername", mock.Anything, "testuser").Return(testUser, nil).Once()
		mockTokenRepo.On("StoreRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mockSessionRepo.On("CreateSession", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditRepo.On("CreateLogEntry", mock.Anything, mock.Anything).Return(nil).Once()

		req := LoginRequest{Username: "testuser", Password: "correct-password"}
		resp, err := authAppSvc.Login(context.Background(), req)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "user-123", resp.User.ID)
		mockIdentitySvc.AssertExpectations(t)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mockIdentitySvc.On("GetUserByUsername", mock.Anything, "testuser").Return(testUser, nil).Once()
		mockAuditRepo.On("CreateLogEntry", mock.Anything, mock.Anything).Return(nil).Once()

		req := LoginRequest{Username: "testuser", Password: "wrong-password"}
		_, err := authAppSvc.Login(context.Background(), req)

		assert.NotNil(t, err)
		assert.Equal(t, types.ErrInvalidCredentials.Code, err.Code)
		mockIdentitySvc.AssertExpectations(t)
	})
}

