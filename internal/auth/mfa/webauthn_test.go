package mfa

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/pkg/types"
)

func TestWebAuthnProvider_Challenge(t *testing.T) {
	// Arrange
	provider := NewWebAuthnProvider()
	user := &types.User{ID: "test-user"}

	// Act
	challenge, err := provider.Challenge(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, challenge)
	assert.Equal(t, types.AuthMethod("webauthn"), challenge.MFAProvider)
}

func TestWebAuthnProvider_Verify(t *testing.T) {
	// Arrange
	provider := NewWebAuthnProvider()
	user := &types.User{ID: "test-user"}
	validResponse := "valid-response"
	invalidResponse := "invalid-response"

	// Act
	valid, err := provider.Verify(context.Background(), user, validResponse)
	assert.NoError(t, err)

	invalid, err := provider.Verify(context.Background(), user, invalidResponse)
	assert.NoError(t, err)

	// Assert
	assert.True(t, valid)
	// This is a placeholder, so it will always return true.
	assert.False(t, invalid)
}

func TestWebAuthnProvider_ListMethods(t *testing.T) {
	// Arrange
	provider := NewWebAuthnProvider()
	user := &types.User{ID: "test-user"}

	// Act
	methods, err := provider.ListMethods(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, methods)
	assert.Len(t, methods, 1)
	assert.Equal(t, "webauthn", methods[0].Type)
}

func TestWebAuthnProvider_GetStrength(t *testing.T) {
	// Arrange
	provider := NewWebAuthnProvider()

	// Act
	strength := provider.GetStrength()

	// Assert
	assert.Equal(t, StrengthLevelStrong, strength)
}
