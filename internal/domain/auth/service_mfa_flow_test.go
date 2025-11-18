package auth

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"testing"
)

// MockRiskEngine is a mock implementation of the RiskEngine interface for testing.
type MockRiskEngine struct {
	mock.Mock
}

func (m *MockRiskEngine) Evaluate(ctx context.Context, ac AuthContext) (RiskScore, RiskLevel, error) {
	args := m.Called(ctx, ac)
	return args.Get(0).(RiskScore), args.Get(1).(RiskLevel), args.Error(2)
}

// MockPolicyEngine is a mock implementation of the PolicyEngine interface for testing.
type MockPolicyEngine struct {
	mock.Mock
}

func (m *MockPolicyEngine) Decide(level RiskLevel, ac AuthContext) string {
	args := m.Called(level, ac)
	return args.String(0)
}

func TestLoginWithPassword_NoMFAForLowRisk(t *testing.T) {
	// Arrange
	mockIdentityService := new(identity.MockIService)
	mockSessionRepo := new(MockSessionRepository)
	mockTokenRepo := new(MockTokenRepository)
	mockAuditRepo := new(MockAuditLogRepository)
	mockRiskEngine := new(MockRiskEngine)
	mockPolicyEngine := new(MockPolicyEngine)
	mockLogger := new(utils.MockLogger)
	mockCrypto := new(utils.MockCryptoManager)

	service := NewService(mockIdentityService, mockSessionRepo, mockTokenRepo, mockAuditRepo, nil, mockCrypto, mockLogger, mockRiskEngine, mockPolicyEngine, nil)

	user := &types.User{ID: "user1", Username: "test", Password: "hashed_password", Status: types.UserStatusActive}
	mockIdentityService.On("GetUserByUsername", mock.Anything, "test").Return(user, nil)
	mockCrypto.On("CheckPasswordHash", "password", "hashed_password").Return(true)
	mockRiskEngine.On("Evaluate", mock.Anything, mock.Anything).Return(RiskScore(0.2), RiskLevelLow, nil)
	mockPolicyEngine.On("Decide", RiskLevelLow, mock.Anything).Return("ALLOW")
	mockCrypto.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything).Return("access_token", nil)
	mockCrypto.On("GenerateUUID").Return("refresh_token")
	mockSessionRepo.On("CreateSession", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockTokenRepo.On("StoreRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAuditRepo.On("CreateLogEntry", mock.Anything, mock.Anything).Return(nil)

	// Act
	authResult, err := service.LoginWithPassword(context.Background(), AuthnRequest{Username: "test", Password: "password"}, Config{})

	// Assert
	assert.NoError(t, err)
	assert.False(t, authResult.IsMfaRequired)
	assert.NotNil(t, authResult.Token)
}

func TestLoginWithPassword_ReturnMFAChallengeForMediumRisk(t *testing.T) {
	// Arrange
	mockIdentityService := new(identity.MockIService)
	mockRiskEngine := new(MockRiskEngine)
	mockPolicyEngine := new(MockPolicyEngine)
	mockLogger := new(utils.MockLogger)
	mockCrypto := new(utils.MockCryptoManager)

	service := NewService(mockIdentityService, nil, nil, nil, nil, mockCrypto, mockLogger, mockRiskEngine, mockPolicyEngine, nil)

	user := &types.User{ID: "user1", Username: "test", Password: "hashed_password", Status: types.UserStatusActive}
	mockIdentityService.On("GetUserByUsername", mock.Anything, "test").Return(user, nil)
	mockCrypto.On("CheckPasswordHash", "password", "hashed_password").Return(true)
	mockRiskEngine.On("Evaluate", mock.Anything, mock.Anything).Return(RiskScore(0.6), RiskLevelMedium, nil)
	mockPolicyEngine.On("Decide", RiskLevelMedium, mock.Anything).Return("REQUIRE_MFA")

	// Act
	authResult, err := service.LoginWithPassword(context.Background(), AuthnRequest{Username: "test", Password: "password"}, Config{})

	// Assert
	assert.NoError(t, err)
	assert.True(t, authResult.IsMfaRequired)
	assert.NotNil(t, authResult.MFAChallenge)
}

func TestVerifyMFAChallenge_SuccessCreatesSessionAndTokens(t *testing.T) {
	// Arrange
	mockIdentityService := new(identity.MockIService)
	mockSessionRepo := new(MockSessionRepository)
	mockTokenRepo := new(MockTokenRepository)
	mockAuditRepo := new(MockAuditLogRepository)
	mockLogger := new(utils.MockLogger)
	mockCrypto := new(utils.MockCryptoManager)

	service := NewService(mockIdentityService, mockSessionRepo, mockTokenRepo, mockAuditRepo, nil, mockCrypto, mockLogger, nil, nil, nil)

	user := &types.User{ID: "user1", Username: "test"}
	mockIdentityService.On("GetUserByID", mock.Anything, "user1").Return(user, nil)
	mockCrypto.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything).Return("access_token", nil)
	mockCrypto.On("GenerateUUID").Return("refresh_token")
	mockSessionRepo.On("CreateSession", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockTokenRepo.On("StoreRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAuditRepo.On("CreateLogEntry", mock.Anything, mock.Anything).Return(nil)

	// Act
	authResult, err := service.VerifyMFAChallenge(context.Background(), &types.VerifyMFARequest{UserID: "user1", Code: "123456"}, Config{})

	// Assert
	assert.NoError(t, err)
	assert.False(t, authResult.IsMfaRequired)
	assert.NotNil(t, authResult.Token)
}
