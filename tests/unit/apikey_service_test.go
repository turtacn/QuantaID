package unit

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/apikey"
	"github.com/turtacn/QuantaID/internal/services/platform"
	"github.com/turtacn/QuantaID/pkg/types"
	"golang.org/x/crypto/bcrypt"
)

// MockRepository is a mock implementation of apikey.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, key *apikey.APIKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*apikey.APIKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*apikey.APIKey), args.Error(1)
}

func (m *MockRepository) GetByPrefix(ctx context.Context, prefix string) ([]*apikey.APIKey, error) {
	args := m.Called(ctx, prefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*apikey.APIKey), args.Error(1)
}

func (m *MockRepository) GetByKeyID(ctx context.Context, keyID string) (*apikey.APIKey, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*apikey.APIKey), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, key *apikey.APIKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListByAppID(ctx context.Context, appID string) ([]*apikey.APIKey, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*apikey.APIKey), args.Error(1)
}

func (m *MockRepository) GetRateLimitPolicy(ctx context.Context, appID string) (*apikey.RateLimitPolicy, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*apikey.RateLimitPolicy), args.Error(1)
}

func TestAPIKeyService_GenerateKey(t *testing.T) {
	mockRepo := new(MockRepository)
	service := platform.NewAPIKeyService(mockRepo)

	ctx := context.Background()
	appID := "test-app"
	prefix := "qid_test_"
	scopes := []string{"read", "write"}

	// When Create is called, we return nil (success)
	// mock.MatchedBy allows us to match any *apikey.APIKey and ensures we don't need exact object matching
	mockRepo.On("Create", ctx, mock.MatchedBy(func(k *apikey.APIKey) bool {
		// KeyID should be 16 chars
		return k.AppID == appID && k.Prefix == prefix && len(k.KeyID) == 16
	})).Return(nil)

	plaintext, key, err := service.GenerateKey(ctx, appID, prefix, scopes, time.Hour)

	if assert.NoError(t, err) {
		assert.NotEmpty(t, plaintext)
		assert.True(t, strings.HasPrefix(plaintext, prefix))
		assert.NotNil(t, key)
		if key != nil {
			assert.Equal(t, appID, key.AppID)
			assert.Equal(t, prefix, key.Prefix)
			assert.Len(t, key.KeyID, 16)
			// Verify hash
			err = bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(plaintext))
			assert.NoError(t, err)
		}
	}

	mockRepo.AssertExpectations(t)
}

func TestAPIKeyService_ValidateKey(t *testing.T) {
	ctx := context.Background()
	prefix := "qid_test_"
	// Use 48 chars random suffix (16 keyID + 32 secret)
	keyID := "0000000000000000" // 16 chars
	secret := "11111111111111111111111111111111" // 32 chars
	plaintext := prefix + keyID + secret
	hash, _ := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)

	validKey := &apikey.APIKey{
		ID:      "1",
		AppID:   "app-1",
		KeyID:   keyID,
		KeyHash: string(hash),
		Prefix:  prefix,
		Revoked: false,
	}

	t.Run("ValidKey", func(t *testing.T) {
		m := new(MockRepository)
		s := platform.NewAPIKeyService(m)
		m.On("GetByKeyID", ctx, keyID).Return(validKey, nil).Once()

		key, err := s.ValidateKey(ctx, plaintext)
		assert.NoError(t, err)
		if assert.NotNil(t, key) {
			assert.Equal(t, validKey.ID, key.ID)
		}
		m.AssertExpectations(t)
	})

	t.Run("InvalidKey_WrongHash", func(t *testing.T) {
		m := new(MockRepository)
		s := platform.NewAPIKeyService(m)
		// Same KeyID, different secret -> wrong hash
		wrongPlaintext := prefix + keyID + "22222222222222222222222222222222"
		m.On("GetByKeyID", ctx, keyID).Return(validKey, nil).Once()

		_, err := s.ValidateKey(ctx, wrongPlaintext)
		assert.ErrorIs(t, err, types.ErrInvalidToken)
		m.AssertExpectations(t)
	})

	t.Run("RevokedKey", func(t *testing.T) {
		m := new(MockRepository)
		s := platform.NewAPIKeyService(m)
		revokedKey := *validKey
		revokedKey.Revoked = true
		m.On("GetByKeyID", ctx, keyID).Return(&revokedKey, nil).Once()

		_, err := s.ValidateKey(ctx, plaintext)
		assert.ErrorIs(t, err, types.ErrInvalidToken)
		m.AssertExpectations(t)
	})

	t.Run("ExpiredKey", func(t *testing.T) {
		m := new(MockRepository)
		s := platform.NewAPIKeyService(m)
		expiredKey := *validKey
		yesterday := time.Now().Add(-24 * time.Hour)
		expiredKey.ExpiresAt = &yesterday
		m.On("GetByKeyID", ctx, keyID).Return(&expiredKey, nil).Once()

		_, err := s.ValidateKey(ctx, plaintext)
		assert.ErrorIs(t, err, types.ErrInvalidToken)
		m.AssertExpectations(t)
	})
}
