package mfa

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/notification"
	"github.com/turtacn/QuantaID/pkg/utils"
	go_redis "github.com/redis/go-redis/v9"
)

// MockNotifier mocks the Notifier interface
type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Send(ctx context.Context, msg notification.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockNotifier) Type() string {
	args := m.Called()
	return args.String(0)
}

// MockNotificationManager mocks notification.Manager
type MockNotificationManager struct {
	mock.Mock
}

func (m *MockNotificationManager) GetNotifier(method string) (notification.Notifier, error) {
	args := m.Called(method)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(notification.Notifier), args.Error(1)
}

func TestOTPProvider_Challenge(t *testing.T) {
	mockRedis := new(redis.MockRedisClient)
	mockNotifier := new(MockNotifier)
	mockManager := new(MockNotificationManager)
	cryptoManager := utils.NewCryptoManager("test-secret")
	config := OTPConfig{
		TTL:    5 * time.Minute,
		Length: 6,
	}

	provider := NewOTPProvider(mockRedis, mockManager, cryptoManager, config)

	ctx := context.Background()
	userID := "user123"
	target := "test@example.com"
	method := "email"

	// Mock Manager returning Notifier
	mockManager.On("GetNotifier", method).Return(mockNotifier, nil)

	// Expect Redis SetEx to be called
	mockRedis.On("SetEx", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), config.TTL).Return(go_redis.NewStatusCmd(ctx)).Once()

	// Expect Notifier Send to be called
	mockNotifier.On("Send", ctx, mock.MatchedBy(func(msg notification.Message) bool {
		return msg.Recipient == target && msg.Type == notification.MessageTypeOTP
	})).Return(nil)

	_, err := provider.Challenge(ctx, userID, target, method)
	assert.NoError(t, err)

	mockRedis.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
	mockManager.AssertExpectations(t)
}

func TestOTPProvider_Challenge_NotifierError(t *testing.T) {
	mockRedis := new(redis.MockRedisClient)
	mockManager := new(MockNotificationManager)
	cryptoManager := utils.NewCryptoManager("test-secret")
	config := OTPConfig{
		TTL:    5 * time.Minute,
		Length: 6,
	}

	provider := NewOTPProvider(mockRedis, mockManager, cryptoManager, config)

	ctx := context.Background()
	userID := "user123"
	target := "test@example.com"
	method := "email"

	// Mock Manager returning Error
	mockManager.On("GetNotifier", method).Return(nil, fmt.Errorf("notifier not found"))

	_, err := provider.Challenge(ctx, userID, target, method)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notifier not found")

	mockManager.AssertExpectations(t)
}


func TestOTPProvider_Verify(t *testing.T) {
	mockRedis := new(redis.MockRedisClient)
	mockManager := new(MockNotificationManager)
	cryptoManager := utils.NewCryptoManager("test-secret")
	config := OTPConfig{
		TTL:    5 * time.Minute,
		Length: 6,
	}

	provider := NewOTPProvider(mockRedis, mockManager, cryptoManager, config)

	ctx := context.Background()
	userID := "user123"
	code := "123456"

	// Test success
	mockRedis.On("Get", ctx, "mfa:otp:"+userID).Return(code, nil).Once()
	mockRedis.On("Del", ctx, []string{"mfa:otp:"+userID}).Return(nil).Once()

	valid, err := provider.Verify(ctx, userID, code)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test invalid code
	mockRedis.On("Get", ctx, "mfa:otp:"+userID).Return("654321", nil).Once()

	valid, err = provider.Verify(ctx, userID, code)
	assert.NoError(t, err)
	assert.False(t, valid)

	mockRedis.AssertExpectations(t)
}
