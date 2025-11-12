package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"github.com/turtacn/QuantaID/tests/testutils"
	"go.opentelemetry.io/otel/trace"
)

func TestApplicationService_Login(t *testing.T) {
	identityService := new(testutils.MockIdentityService)
	sessionRepo := new(testutils.MockSessionRepository)
	tokenRepo := new(testutils.MockTokenRepository)
	auditRepo := new(testutils.MockAuditLogRepository)
	cryptoManager := utils.NewCryptoManager("test-secret-key")
	logger := utils.NewNoopLogger()
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	authDomain := auth.NewService(
		identityService,
		sessionRepo,
		tokenRepo,
		auditRepo,
		cryptoManager,
		logger,
	)

	appService := NewApplicationService(
		authDomain,
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

	loginReq := LoginRequest{
		Username: "test@example.com",
		Password: "password",
	}
	loginResp, err := appService.Login(context.Background(), loginReq)
	assert.Nil(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.AccessToken)
}
