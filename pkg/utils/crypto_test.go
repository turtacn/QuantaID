package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCryptoManager_PasswordHashing(t *testing.T) {
	crypto := NewCryptoManager("secret")
	password := "my-strong-password-123"

	hashedPassword, err := crypto.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	// Check correct password
	assert.True(t, crypto.CheckPasswordHash(password, hashedPassword), "Correct password should be valid")

	// Check incorrect password
	assert.False(t, crypto.CheckPasswordHash("wrong-password", hashedPassword), "Incorrect password should be invalid")
}

func TestCryptoManager_JWT(t *testing.T) {
	crypto := NewCryptoManager("a-very-secure-secret-key")
	userID := "user-abc-123"

	t.Run("Generate and Validate successfully", func(t *testing.T) {
		token, err := crypto.GenerateJWT(userID, 15*time.Minute, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := crypto.ValidateJWT(token)
		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims["sub"])
		assert.Equal(t, "QuantaID", claims["iss"])
	})

	t.Run("Validation fails for expired token", func(t *testing.T) {
		// Generate a token that expired 1 second ago
		token, err := crypto.GenerateJWT(userID, -1*time.Second, nil)
		require.NoError(t, err)

		_, err = crypto.ValidateJWT(token)
		assert.Error(t, err, "Expired token should fail validation")
	})

	t.Run("Validation fails for token with wrong secret", func(t *testing.T) {
		token, err := crypto.GenerateJWT(userID, 15*time.Minute, nil)
		require.NoError(t, err)

		wrongCrypto := NewCryptoManager("a-different-secret-key")
		_, err = wrongCrypto.ValidateJWT(token)
		assert.Error(t, err, "Token with wrong secret should fail validation")
	})
}

//Personal.AI order the ending
