package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/services/audit"
	authsvc "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"github.com/turtacn/QuantaID/tests/testutils"
	"go.opentelemetry.io/otel/trace"
	i_audit "github.com/turtacn/QuantaID/internal/audit"
)

func TestAuthenticationFlow(t *testing.T) {
	// Initialize logger and tracer
	logger := utils.NewNoopLogger()
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	// Initialize repositories
	identityService := &testutils.MockIdentityService{}
	sessionRepo := &testutils.MockSessionRepository{}
	tokenRepo := &testutils.MockTokenRepository{}
	auditRepo := &testutils.MockAuditLogRepository{}
	cryptoManager := utils.NewCryptoManager("test-secret-key")

	// Initialize services
	riskEngine := &testutils.MockRiskEngine{}
	authDomain := auth.NewService(
		identityService,
		sessionRepo,
		tokenRepo,
		auditRepo,
		cryptoManager,
		logger,
		riskEngine,
	)
	auditPipeline := i_audit.NewPipeline(logger.(*utils.ZapLogger).Logger, &testutils.MockSink{})
	auditService := audit.NewService(auditPipeline)
	appService := authsvc.NewApplicationService(
		authDomain,
		auditService,
		logger,
		authsvc.Config{},
		tracer,
	)

	// Mock user repository
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

	// Create authentication request
	loginReq := authsvc.LoginRequest{
		Username: "test@example.com",
		Password: "password",
	}

	// Perform authentication
	loginResp, err := appService.Login(context.Background(), loginReq)

	// Assertions
	assert.Nil(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.AccessToken)
}
