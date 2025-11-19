package auth

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/auth/adaptive"
	"github.com/turtacn/QuantaID/internal/auth/mfa"
	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"github.com/turtacn/QuantaID/tests/testutils"
	"go.opentelemetry.io/otel/trace"
	"github.com/turtacn/QuantaID/internal/services/audit"
	i_audit "github.com/turtacn/QuantaID/internal/audit"
)

type MockPolicyEngine struct {
	mock.Mock
}

func (m *MockPolicyEngine) Decide(level auth.RiskLevel, ac auth.AuthContext) string {
	args := m.Called(level, ac)
	return args.String(0)
}

func TestApplicationService_Login(t *testing.T) {
	identityService := new(testutils.MockIdentityService)
	sessionRepo := new(testutils.MockSessionRepository)
	tokenRepo := new(testutils.MockTokenRepository)
	auditRepo := new(testutils.MockAuditLogRepository)
	tokenFamilyRepo := new(testutils.MockTokenFamilyRepository)
	cryptoManager := utils.NewCryptoManager("test-secret-key")
	logger := utils.NewNoopLogger()
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	policyEngine := new(MockPolicyEngine)
	redisClient := new(testutils.MockRedisClient)

	riskEngine := adaptive.NewRiskEngine(config.RiskConfig{}, redisClient, logger.(*utils.ZapLogger).Logger)
	mfaManager := &mfa.MFAManager{}
	authDomain := auth.NewService(
		identityService,
		sessionRepo,
		tokenRepo,
		auditRepo,
		tokenFamilyRepo,
		cryptoManager,
		logger,
		riskEngine,
		policyEngine,
		mfaManager,
	)
	auditPipeline := i_audit.NewPipeline(logger.(*utils.ZapLogger).Logger, &testutils.MockSink{})
	auditService := audit.NewService(auditPipeline)
	appService := NewApplicationService(
		authDomain,
		auditService,
		logger,
		Config{},
		tracer,
	)

	// Test case 1: Successful authentication
	hashedPassword, _ := cryptoManager.HashPassword("password")
	user := &types.User{
		ID:       "user-123",
		Password: hashedPassword,
		Status:   types.UserStatusActive,
	}
	identityService.On("GetUserByUsername", mock.Anything, "test@example.com").Return(user, nil)
	sessionRepo.On("CreateSession", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	tokenRepo.On("StoreRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	auditRepo.On("CreateLogEntry", mock.Anything, mock.Anything).Return(nil)
	policyEngine.On("Decide", mock.Anything, mock.Anything).Return("ALLOW")
	tokenFamilyRepo.On("CreateFamily", mock.Anything, mock.Anything).Return(nil)
	redisClient.On("SIsMember", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewBoolResult(false, nil))
	redisClient.On("Get", mock.Anything, mock.Anything).Return(redis.NewStringResult("", redis.Nil))

	loginReq := LoginRequest{
		Username: "test@example.com",
		Password: "password",
	}
	authReq := auth.AuthnRequest{
		Username: loginReq.Username,
		Password: loginReq.Password,
	}
	loginResp, err := appService.LoginWithPassword(context.Background(), authReq, auth.Config{})
	assert.Nil(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.Token.AccessToken)
}
