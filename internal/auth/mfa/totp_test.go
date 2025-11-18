package mfa

import (
	"context"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/pkg/types"
	"testing"
	"time"
)

func TestTOTPProvider_Challenge(t *testing.T) {
	// Arrange
	provider := NewTOTPProvider()
	user := &types.User{ID: "test-user"}

	// Act
	challenge, err := provider.Challenge(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, types.AuthMethod("totp"), challenge.MFAProvider)
}

func TestTOTPProvider_Verify(t *testing.T) {
	// Arrange
	provider := NewTOTPProvider()
	user := &types.User{ID: "test-user"}
	secret := "JBSWY3DPEHPK3PXP" // Placeholder secret
	validCode, err := totp.GenerateCode(secret, time.Now())
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
	provider := NewTOTPProvider()
	user := &types.User{ID: "test-user"}

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
	provider := NewTOTPProvider()

	// Act
	strength := provider.GetStrength()

	// Assert
	assert.Equal(t, StrengthLevelNormal, strength)
}
