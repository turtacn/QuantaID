package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/auth/mfa"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/notification"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

func TestInitiateRecovery_UserNotFound(t *testing.T) {
	mockRepo := new(identity.MockIdentityRepository)
	mockRepo.On("GetUserByEmail", mock.Anything, "unknown@example.com").Return((*types.User)(nil), types.ErrUserNotFound)

	service := NewRecoveryService(
		mockRepo,
		nil, // OTP provider not needed for this path
		nil,
		nil,
		zap.NewNop(),
	)

	err := service.InitiateRecovery(context.Background(), "unknown@example.com")
	assert.NoError(t, err) // Should not return error
}

func TestVerifyAndReset_Success(t *testing.T) {
	// Setup Mocks
	mockRepo := new(identity.MockIdentityRepository)
	mockRedis := new(redis.MockRedisClient)
	mockNotifierMgr := new(notification.MockNotificationManager)
	// mockNotifier := new(notification.MockNotifier)

	// User
	user := &types.User{
		ID:    "user123",
		Email: "test@example.com",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockRepo.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *types.User) bool {
		// Verify password hash is updated (non-empty)
		// Note: types.User uses 'Password' field, but service might be setting 'PasswordHash'
		// The service code sets 'user.PasswordHash = ...' but types.User has 'Password'.
		// Let's check the service implementation. It sets PasswordHash but the struct has Password.
		// Wait, types.User has Password field which stores the hash.
		return u.ID == "user123" && u.Password != ""
	})).Return(nil)

	// OTP
	mockRedis.On("Get", mock.Anything, "mfa:otp:user123").Return("123456", nil)
	mockRedis.On("Del", mock.Anything, []string{"mfa:otp:user123"}).Return(nil)

	// Session
	mockRedis.On("ZRange", mock.Anything, "user_sessions:user123", int64(0), int64(-1)).Return([]string{"session1"}, nil)
	mockRedis.On("Del", mock.Anything, []string{"session:session1"}).Return(nil)
	mockRedis.On("ZRem", mock.Anything, "user_sessions:user123", []interface{}{"session1"}).Return(int64(1), nil)

	// Metrics needed by SessionManager
	metrics := redis.NewMetrics("test", nil)

	// Dependencies
	cryptoManager := utils.NewCryptoManager("secret")
	otpProvider := mfa.NewOTPProvider(mockRedis, mockNotifierMgr, cryptoManager, mfa.OTPConfig{TTL: time.Minute, Length: 6})
	sessionManager := redis.NewSessionManager(
		mockRedis,
		redis.SessionConfig{},
		zap.NewNop(),
		&redis.UUIDv4Generator{},
		&redis.RealClock{},
		metrics,
	)

	service := NewRecoveryService(
		mockRepo,
		otpProvider,
		cryptoManager,
		sessionManager,
		zap.NewNop(),
	)

	err := service.VerifyAndReset(context.Background(), "test@example.com", "123456", "newpassword")
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}
