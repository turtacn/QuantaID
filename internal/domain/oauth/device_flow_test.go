package oauth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
)

// MockDeviceCodeRepository is a mock implementation of DeviceCodeRepository.
type MockDeviceCodeRepository struct {
	mock.Mock
}

func (m *MockDeviceCodeRepository) Create(ctx context.Context, record *DeviceCodeRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockDeviceCodeRepository) GetByDeviceCode(ctx context.Context, deviceCode string) (*DeviceCodeRecord, error) {
	args := m.Called(ctx, deviceCode)
	record, _ := args.Get(0).(*DeviceCodeRecord)
	return record, args.Error(1)
}

func (m *MockDeviceCodeRepository) GetByUserCode(ctx context.Context, userCode string) (*DeviceCodeRecord, error) {
	args := m.Called(ctx, userCode)
	record, _ := args.Get(0).(*DeviceCodeRecord)
	return record, args.Error(1)
}

func (m *MockDeviceCodeRepository) UpdateStatus(ctx context.Context, deviceCode, status string) error {
	args := m.Called(ctx, deviceCode, status)
	return args.Error(0)
}

func (m *MockDeviceCodeRepository) MarkUsed(ctx context.Context, deviceCode string) error {
	args := m.Called(ctx, deviceCode)
	return args.Error(0)
}

func (m *MockDeviceCodeRepository) Update(ctx context.Context, deviceCode string, updates map[string]interface{}) error {
	args := m.Called(ctx, deviceCode, updates)
	return args.Error(0)
}

// MockTokenService is a mock implementation of TokenService.
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) IssueTokens(ctx context.Context, req types.TokenRequest) (*types.Token, error) {
	args := m.Called(ctx, req)
	token, _ := args.Get(0).(*types.Token)
	return token, args.Error(1)
}

func TestDeviceFlow_GenerateCode(t *testing.T) {
	repo := new(MockDeviceCodeRepository)
	config := DeviceFlowConfig{
		DeviceCodeLength: 32,
		UserCodeLength:   8,
		UserCodeCharset:  "BCDFGHJKLMNPQRSTVWXYZ23456789",
		ExpiresIn:        15 * time.Minute,
		PollingInterval:  5,
		VerificationURI:  "https://example.com/device",
	}
	handler := NewDeviceFlowHandler(repo, nil, config)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*oauth.DeviceCodeRecord")).Return(nil)

	resp, err := handler.HandleDeviceAuthorizationRequest(context.Background(), "client_id", "openid profile")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.DeviceCode)
	assert.NotEmpty(t, resp.UserCode)
}
