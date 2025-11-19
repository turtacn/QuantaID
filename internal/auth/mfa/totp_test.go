package mfa

import (
	"context"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"testing"
	"time"
)

func TestTOTPProvider_Enroll(t *testing.T) {
	// Arrange
	repo := memory.NewMFAFactorMemoryRepository()
	crypto := utils.NewCryptoManager("test-secret")
	provider := NewTOTPProvider(repo, crypto)
	user := &types.User{ID: "550e8400-e29b-41d4-a716-446655440000", Username: "test-user"}

	// Act
	enrollment, err := provider.Enroll(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, enrollment)
	assert.NotEmpty(t, enrollment.Secret)
	assert.NotEmpty(t, enrollment.URL)

	factors, err := repo.GetMFAFactorsByUserID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, factors, 1)
	assert.Equal(t, "totp", factors[0].Type)
}

func TestTOTPProvider_Challenge(t *testing.T) {
	// Arrange
	repo := memory.NewMFAFactorMemoryRepository()
	crypto := utils.NewCryptoManager("test-secret")
	provider := NewTOTPProvider(repo, crypto)
	user := &types.User{ID: "550e8400-e29b-41d4-a716-446655440000"}

	// Act
	challenge, err := provider.Challenge(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, types.AuthMethod("totp"), challenge.MFAProvider)
}

func TestTOTPProvider_Verify(t *testing.T) {
	// Arrange
	repo := memory.NewMFAFactorMemoryRepository()
	crypto := utils.NewCryptoManager("test-secret")
	provider := NewTOTPProvider(repo, crypto)
	user := &types.User{ID: "550e8400-e29b-41d4-a716-446655440000", Username: "test-user"}

	enrollment, err := provider.Enroll(context.Background(), user)
	assert.NoError(t, err)

	validCode, err := totp.GenerateCode(enrollment.Secret, time.Now())
	assert.NoError(t, err)
	invalidCode := "654321"

	// Act
	valid, err := provider.Verify(context.Background(), user, validCode)
	assert.NoError(t, err)

	invalid, err := provider.Verify(context.Background(), user, invalidCode)
	assert.NoError(t, err)

	// Assert
	assert.True(t, valid)
	assert.False(t, invalid)
}

func TestTOTPProvider_ListMethods(t *testing.T) {
	// Arrange
	repo := memory.NewMFAFactorMemoryRepository()
	crypto := utils.NewCryptoManager("test-secret")
	provider := NewTOTPProvider(repo, crypto)
	user := &types.User{ID: "550e8400-e29b-41d4-a716-446655440000", Username: "test-user"}
	_, err := provider.Enroll(context.Background(), user)
	assert.NoError(t, err)
	factors, err := repo.GetMFAFactorsByUserID(context.Background(), user.ID)
	assert.NoError(t, err)
	factors[0].Status = "enrolled"


	// Act
	methods, err := provider.ListMethods(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, methods)
	assert.Len(t, methods, 1)
	assert.Equal(t, "totp", methods[0].Type)
}

func TestTOTPProvider_GetStrength(t *testing.T) {
	// Arrange
	repo := memory.NewMFAFactorMemoryRepository()
	crypto := utils.NewCryptoManager("test-secret")
	provider := NewTOTPProvider(repo, crypto)

	// Act
	strength := provider.GetStrength()

	// Assert
	assert.Equal(t, StrengthLevelNormal, strength)
}
